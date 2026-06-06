package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type service struct {
	repo        Repository
	redisClient redis.UniversalClient
}

type Repository interface {
	Create(job *Job) (*Job, error)
}

func NewService(repo Repository, redisClient redis.UniversalClient) *service {
	return &service{
		repo:        repo,
		redisClient: redisClient,
	}
}

func (s *service) Create(ctx context.Context, key string, payload *JobsCreateRequest) (*JobsCreateResponse, error) {
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
	dataNX, err := json.Marshal(payload)
	if err != nil {
		slog.Info("redis Marshal error", "err", err)
		return nil, err
	}
	locked, err := s.redisClient.SetNX(ctx, fmt.Sprintf("idem:lock:%s", key), dataNX, 30*time.Minute).Result()
	if err != nil {
		slog.Info("redis error", "err", err)
		return nil, err
	}
	if !locked {
		slog.Info("key is exist", "key", key)
		return nil, errors.New("duplicate request")
	}

	job := &Job{
		Type:    payload.Type,
		Payload: payload.Payload,
	}
	result, err := s.repo.Create(job)
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(result.ID)
	s.redisClient.Set(ctx, fmt.Sprintf("idem:result:%s", key), data, 24*time.Hour)

	response := &JobsCreateResponse{
		ID: result.ID,
		// Type:    result.Type,
		// Status:  result.Status,
		// Payload: result.Payload,
	}

	return response, nil
}
