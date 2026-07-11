// internal/certificate/service/service.go
package service

import (
	"context"
	"portfolio-ai/internal/certificate/entity"
	"portfolio-ai/internal/certificate/repository"
	"portfolio-ai/pkg/ulid"
)

// Service defines the interface for Certificate business operations.
type Service interface {
	List(ctx context.Context, page, limit int) ([]*entity.Certificate, int64, error)
	Get(ctx context.Context, id string) (*entity.Certificate, error)
	Create(ctx context.Context, cert *entity.Certificate) error
	Update(ctx context.Context, cert *entity.Certificate) error
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo repository.Repository
}

// NewService creates a new Service instance.
func NewService(repo repository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, page, limit int) ([]*entity.Certificate, int64, error) {
	return s.repo.List(ctx, page, limit)
}

func (s *service) Get(ctx context.Context, id string) (*entity.Certificate, error) {
	return s.repo.Get(ctx, id)
}

func (s *service) Create(ctx context.Context, cert *entity.Certificate) error {
	if cert.ID == "" {
		cert.ID = ulid.New()
	}
	return s.repo.Create(ctx, cert)
}

func (s *service) Update(ctx context.Context, cert *entity.Certificate) error {
	return s.repo.Update(ctx, cert)
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
