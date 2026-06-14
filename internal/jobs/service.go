package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type service struct {
	db          *sqlx.DB
	repo        Repository
	redisClient redis.UniversalClient
}

type Repository interface {
	Create(tx *sqlx.Tx, job *Job) (*Job, error)
	GetJobs() (*[]Job, error)
	SetStatus(job *Job) error
	CreateIdem(tx *sqlx.Tx, data *IdemData) (*IdemData, error)
	UpdateIdem(tx *sqlx.Tx, data *IdemData) (*IdemData, error)
}

func NewService(db *sqlx.DB, repo Repository, redisClient redis.UniversalClient) *service {
	return &service{
		db:          db,
		repo:        repo,
		redisClient: redisClient,
	}
}

func (s *service) Create(ctx context.Context, userID uuid.UUID, key string, payload *JobsCreateRequest) (*JobsCreateResponse, error) {
	id, err := s.redisClient.Get(ctx, fmt.Sprintf("idem:result:%s", key)).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	parseId, err := uuid.Parse(id)
	if err == nil {
		return &JobsCreateResponse{
			ID: parseId,
			// Type:    result.Type,
			// Status:  result.Status,
			// Payload: result.Payload,
		}, nil
	}
	// dataNX, err := json.Marshal(payload)
	// if err != nil {
	// 	slog.Info("redis Marshal error", "err", err)
	// 	return nil, err
	// }
	locked, err := s.redisClient.SetNX(ctx, fmt.Sprintf("idem:lock:%s", key), "in_flight", 30*time.Second).Result()
	if err != nil {
		slog.Info("redis error", "err", err)
		return nil, err
	}
	if !locked {
		slog.Info("key is exist", "key", key)
		return nil, errors.New("duplicate request")
	}

	// transation
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	// save idem key
	idemPayload := &IdemData{
		Status: "in_flight",
		Lease:  time.Now().Add(30 * time.Second),
		Key:    key,
		UserID: userID,
	}
	_, err = s.repo.CreateIdem(tx, idemPayload)
	if err != nil {
		return nil, err
	}

	job := &Job{
		Type:    payload.Type,
		Payload: payload.Payload,
	}
	result, err := s.repo.Create(tx, job)
	if err != nil {
		return nil, err
	}
	marshalData, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	// update idem
	updateIdemPayload := &IdemData{
		Key:      key,
		Status:   "complete",
		Response: marshalData,
	}
	_, err = s.repo.UpdateIdem(tx, updateIdemPayload)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.New("create job transation error")
	}
	data, _ := json.Marshal(result)
	s.redisClient.Set(ctx, fmt.Sprintf("idem:result:%s", key), data, 24*time.Hour)

	response := &JobsCreateResponse{
		ID: result.ID,
		// Type:    result.Type,
		// Status:  result.Status,
		// Payload: result.Payload,
	}

	return response, nil
}

func (s *service) GetJobs(ctx context.Context) (*[]Job, error) {
	jobs, err := s.repo.GetJobs()
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

func (s *service) SetStatus(ctx context.Context, job *Job) error {
	err := s.repo.SetStatus(job)
	if err != nil {
		return err
	}

	return nil
}
