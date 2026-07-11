// internal/technology/service/service.go
package service

import (
	"context"
	"portfolio-ai/internal/technology/entity"
	"portfolio-ai/internal/technology/repository"
	"portfolio-ai/pkg/ulid"
)

// Service defines the interface for Technology business operations.
type Service interface {
	List(ctx context.Context, page, limit int) ([]*entity.Technology, int64, error)
	Get(ctx context.Context, id string) (*entity.Technology, error)
	Create(ctx context.Context, tech *entity.Technology) error
	Update(ctx context.Context, tech *entity.Technology) error
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo repository.Repository
}

// NewService creates a new Service instance.
func NewService(repo repository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, page, limit int) ([]*entity.Technology, int64, error) {
	return s.repo.List(ctx, page, limit)
}

func (s *service) Get(ctx context.Context, id string) (*entity.Technology, error) {
	return s.repo.Get(ctx, id)
}

func (s *service) Create(ctx context.Context, tech *entity.Technology) error {
	if tech.ID == "" {
		tech.ID = ulid.New()
	}
	return s.repo.Create(ctx, tech)
}

func (s *service) Update(ctx context.Context, tech *entity.Technology) error {
	return s.repo.Update(ctx, tech)
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
