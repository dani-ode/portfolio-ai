// internal/skill/service/service.go
package service

import (
	"context"
	"portfolio-ai/internal/skill/entity"
	"portfolio-ai/internal/skill/repository"
	"portfolio-ai/pkg/ulid"
)

// Service defines the interface for Skill business operations.
type Service interface {
	List(ctx context.Context, page, limit int) ([]*entity.Skill, int64, error)
	Get(ctx context.Context, id string) (*entity.Skill, error)
	Create(ctx context.Context, skill *entity.Skill) error
	Update(ctx context.Context, skill *entity.Skill) error
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo repository.Repository
}

// NewService creates a new Service instance.
func NewService(repo repository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, page, limit int) ([]*entity.Skill, int64, error) {
	return s.repo.List(ctx, page, limit)
}

func (s *service) Get(ctx context.Context, id string) (*entity.Skill, error) {
	return s.repo.Get(ctx, id)
}

func (s *service) Create(ctx context.Context, skill *entity.Skill) error {
	if skill.ID == "" {
		skill.ID = ulid.New()
	}
	return s.repo.Create(ctx, skill)
}

func (s *service) Update(ctx context.Context, skill *entity.Skill) error {
	return s.repo.Update(ctx, skill)
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
