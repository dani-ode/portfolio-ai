// internal/project/service/service.go
package service

import (
	"context"
	"portfolio-ai/internal/project/entity"
	"portfolio-ai/internal/project/repository"
	"portfolio-ai/pkg/ulid"
)

// Service defines the interface for Project business operations.
type Service interface {
	List(ctx context.Context, page, limit int) ([]*entity.Project, int64, error)
	Get(ctx context.Context, id string) (*entity.Project, error)
	Create(ctx context.Context, project *entity.Project) error
	Update(ctx context.Context, project *entity.Project) error
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo repository.Repository
}

// NewService creates a new Service instance.
func NewService(repo repository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, page, limit int) ([]*entity.Project, int64, error) {
	return s.repo.List(ctx, page, limit)
}

func (s *service) Get(ctx context.Context, id string) (*entity.Project, error) {
	return s.repo.Get(ctx, id)
}

func (s *service) Create(ctx context.Context, project *entity.Project) error {
	if project.ID == "" {
		project.ID = ulid.New()
	}
	return s.repo.Create(ctx, project)
}

func (s *service) Update(ctx context.Context, project *entity.Project) error {
	return s.repo.Update(ctx, project)
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
