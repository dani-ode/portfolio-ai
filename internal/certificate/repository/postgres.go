// internal/certificate/repository/postgres.go
package repository

import (
	"context"
	"portfolio-ai/internal/certificate/entity"

	"gorm.io/gorm"
)

// Repository defines the interface for Certificate database operations.
type Repository interface {
	List(ctx context.Context, page, limit int) ([]*entity.Certificate, int64, error)
	Get(ctx context.Context, id string) (*entity.Certificate, error)
	Create(ctx context.Context, cert *entity.Certificate) error
	Update(ctx context.Context, cert *entity.Certificate) error
	Delete(ctx context.Context, id string) error
}

type postgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a new Repository implementation using GORM.
func NewPostgresRepository(db *gorm.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) List(ctx context.Context, page, limit int) ([]*entity.Certificate, int64, error) {
	var certs []*entity.Certificate
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Certificate{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		offset := (page - 1) * limit
		if offset < 0 {
			offset = 0
		}
		query = query.Offset(offset).Limit(limit)
	}

	if err := query.Order("issue_date DESC").Find(&certs).Error; err != nil {
		return nil, 0, err
	}

	return certs, total, nil
}

func (r *postgresRepository) Get(ctx context.Context, id string) (*entity.Certificate, error) {
	var cert entity.Certificate
	if err := r.db.WithContext(ctx).First(&cert, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &cert, nil
}

func (r *postgresRepository) Create(ctx context.Context, cert *entity.Certificate) error {
	return r.db.WithContext(ctx).Create(cert).Error
}

func (r *postgresRepository) Update(ctx context.Context, cert *entity.Certificate) error {
	return r.db.WithContext(ctx).Save(cert).Error
}

func (r *postgresRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Certificate{}, "id = ?", id).Error
}
