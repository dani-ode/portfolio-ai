// internal/experience/service/service.go
package service

import (
	"context"
	"portfolio-ai/internal/experience/entity"
	"portfolio-ai/internal/experience/repository"
	"portfolio-ai/pkg/ulid"
)

// Service defines the interface for Experience business operations.
type Service interface {
	List(ctx context.Context, page, limit int) ([]*entity.Experience, int64, error)
	Get(ctx context.Context, id string) (*entity.Experience, error)
	Create(ctx context.Context, experience *entity.Experience) error
	Update(ctx context.Context, experience *entity.Experience) error
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo repository.Repository
}

// NewService creates a new Service instance.
func NewService(repo repository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, page, limit int) ([]*entity.Experience, int64, error) {
	return s.repo.List(ctx, page, limit)
}

func (s *service) Get(ctx context.Context, id string) (*entity.Experience, error) {
	return s.repo.Get(ctx, id)
}

func (s *service) Create(ctx context.Context, experience *entity.Experience) error {
	if experience.ID == "" {
		experience.ID = ulid.New()
	}
	return s.repo.Create(ctx, experience)
}

func (s *service) Update(ctx context.Context, experience *entity.Experience) error {
	return s.repo.Update(ctx, experience)
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
