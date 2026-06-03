package stats

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type service struct {
	repo Repository
}

type Repository interface {
	CreateStat(id uuid.UUID) error
	GetStat(id uuid.UUID) (*Stat, error)
	BatchStats(data map[string]int64) error
}

func NewService(repo Repository) *service {
	return &service{repo: repo}
}

func (s *service) CreateStat(ctx context.Context, req *CreateStatRequest) error {
	// validation and error handling
	if req.ID.String() == "" {
		return status.Errorf(codes.InvalidArgument, "Name field is required")
	}

	// format to fit model for db tags
	err := s.repo.CreateStat(req.ID)

	if err != nil {
		return err
	}

	return nil
}

func (s *service) GetStat(ctx context.Context, id uuid.UUID) (*GetStatResponse, error) {
	stat, err := s.repo.GetStat(id)

	if err != nil {
		return nil, err
	}

	// format to fit grpc structure
	return &GetStatResponse{
		ID:        stat.ID,
		Count:     stat.Count,
		UpdatedAt: stat.UpdatedAt,
		CreatedAt: stat.CreatedAt,
	}, nil
}

func (s *service) BatchStats(data map[string]int64) error {
	err := s.repo.BatchStats(data)
	return err
}
