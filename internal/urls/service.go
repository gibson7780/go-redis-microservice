package urls

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	commonconstants "github.com/gibson7780/go-project/common/constants"
	"github.com/gibson7780/go-project/internal/stats"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type StatsService interface {
	CreateStat(ctx context.Context, state *stats.CreateStatRequest) error
	BatchStats(map[string]int64) error
}
type service struct {
	db           *sqlx.DB
	redisClient  redis.UniversalClient
	repo         Repository
	statsService StatsService
}

type Repository interface {
	CreateUrl(urlPayload *Url) (*Url, error)
	GetUrl(key string) (*Url, error)
}

func NewService(db *sqlx.DB, redisClient redis.UniversalClient, repo Repository, statsService StatsService) Service {
	s := &service{db: db, repo: repo, redisClient: redisClient, statsService: statsService}

	return s
}

func generateCode() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 10)

	for i := range result {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}

		result[i] = charset[randomIndex.Int64()]
	}

	return string(result), nil

}
func (s *service) CreateUrl(ctx context.Context, req *CreateUrlRequest, idemKey string) (*CreateUrlResponse, error) {

	_, err := s.redisClient.Ping(ctx).Result()
	if err != nil {
		slog.Error("redis ping failed", "err", err)
	}
	id := uuid.New()
	code, err := generateCode()
	// validation and error handling
	// redis check
	// 前端idempotency key 保證不重整的情況可以拿到同一個避免每次都創新的
	result, err := s.redisClient.Get(ctx, idemKey).Result()
	var response = &CreateUrlResponse{}

	if err == redis.Nil { // miss
		// create and set redis
		response, err = s.Create(ctx, id, code, req)
		if err != nil {
			return nil, err
		}
		response.ShortUrl = fmt.Sprintf("http://localhost:7001/%s", response.Code)
		data, err := json.Marshal(response)
		if err != nil {
			return nil, err
		}
		// set code
		err = s.redisClient.Set(ctx, response.Code, data, s.generateTTLJitter()).Err()
		if err != nil {
			return nil, err
		}
		// set idempotency key
		err = s.redisClient.Set(ctx, idemKey, data, s.generateTTLJitter()).Err()
		if err != nil {
			return nil, err
		}
		// set count
		_, err = s.redisClient.HGet(ctx, "clicks", response.Code).Result()
		if err == redis.Nil {
			s.redisClient.HSet(ctx, "clicks", response.Code, 0)
		} else if err != nil {
			return nil, err
		} else {
			s.redisClient.HSet(ctx, "clicks", response.Code, 0)
		}
		response.ShortUrl = fmt.Sprintf("http://localhost:7001/%s", response.Code)
		return response, nil
	} else if err != nil { // err
		slog.Error("redis get error", "err", err)
		return nil, err
	} else { //有的話回傳value
		json.Unmarshal([]byte(result), &response)
		return response, nil
	}
	// return nil, nil
}

func (s *service) Create(ctx context.Context, id uuid.UUID, code string, req *CreateUrlRequest) (*CreateUrlResponse, error) {

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = tx.Rollback()
	}()
	CreateUrl := &Url{
		ID:        id,
		OriginUrl: req.Url,
		Code:      code,
	}
	// create url
	response, err := s.repo.CreateUrl(CreateUrl)
	// create state relation
	statesPayload := &stats.CreateStatRequest{
		ID: id,
	}
	err = s.statsService.CreateStat(ctx, statesPayload)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &CreateUrlResponse{
		ID:        response.ID,
		Code:      response.Code,
		OriginUrl: response.OriginUrl,
		ShortUrl:  response.ShortUrl,
	}, nil
}

func (s *service) GetUrl(ctx context.Context, code string) (*GetUrlResponse, error) {
	// redis
	// check tombstone
	_, err := s.redisClient.HGet(ctx, "tombstone", code).Result()
	if err == nil {
		return nil, commonconstants.ErrNotFound
	}
	// count add
	_, err = s.redisClient.HIncrBy(ctx, "clicks", code, 1).Result()
	if err != nil {
		return nil, err
	}
	// get redirect url
	urlResult, err := s.redisClient.Get(ctx, code).Result()

	if err != nil { // get db
		res, err := s.repo.GetUrl(code) // format to fit grpc structure
		if err == commonconstants.ErrNotFound {
			// tombstone
			s.redisClient.HSet(ctx, "tombstone", code, true)
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		return &GetUrlResponse{
			ID:        res.ID,
			Code:      res.Code,
			OriginUrl: res.OriginUrl,
			ShortUrl:  fmt.Sprintf(`localhost:3000/%s`, res.Code),
			CreatedAt: res.CreatedAt,
			UpdatedAt: res.UpdatedAt,
		}, nil
	} else {
		var urlData Url
		json.Unmarshal([]byte(urlResult), &urlData)
		return &GetUrlResponse{
			ID:        urlData.ID,
			Code:      urlData.Code,
			OriginUrl: urlData.OriginUrl,
			ShortUrl:  fmt.Sprintf(`localhost:3000/%s`, urlData.Code),
			CreatedAt: urlData.CreatedAt,
			UpdatedAt: urlData.UpdatedAt,
		}, nil
	}

}

func (s *service) generateTTLJitter() time.Duration {
	randomNum, _ := rand.Int(rand.Reader, big.NewInt(30)) // 1- 30天
	jitter := 24 * time.Duration(randomNum.Int64()) * time.Hour
	return jitter
}
