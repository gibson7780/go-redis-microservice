package stats

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type service struct {
	repo Repository
}

type Repository interface {
	Create(example *ExampleCreate) (*Example, error)
	GetByID(id uuid.UUID) (*Example, error)
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateStats(ctx context.Context, req *CreateExampleRequest) (*Example, error) {
	// validation and error handling
	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Name field is required")
	}

	// format to fit model for db tags
	createExample := &ExampleCreate{
		Name: req.Name,
	}
	example, err := s.repo.Create(createExample)

	if err != nil {
		return nil, err
	}

	// publish rabbit mq message after succesfuly creating
	_, err = json.Marshal(example)

	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return &Example{
		ID:   example.ID,
		Name: example.Name,
	}, nil
}

func (s *service) GetStats(ctx context.Context, id uuid.UUID) (*Example, error) {
	example, err := s.repo.GetByID(id)

	if err != nil {
		return nil, err
	}

	// format to fit grpc structure
	return &Example{
		ID:   example.ID,
		Name: example.Name,
	}, nil
}
