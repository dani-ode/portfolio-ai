// internal/experience/repository/postgres.go
package repository

import (
	"context"
	"portfolio-ai/internal/experience/entity"

	"gorm.io/gorm"
)

// Repository defines the interface for Experience database operations.
type Repository interface {
	List(ctx context.Context, page, limit int) ([]*entity.Experience, int64, error)
	Get(ctx context.Context, id string) (*entity.Experience, error)
	Create(ctx context.Context, experience *entity.Experience) error
	Update(ctx context.Context, experience *entity.Experience) error
	Delete(ctx context.Context, id string) error
}

type postgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a new Repository implementation using GORM.
func NewPostgresRepository(db *gorm.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) List(ctx context.Context, page, limit int) ([]*entity.Experience, int64, error) {
	var experiences []*entity.Experience
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Experience{})
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

	if err := query.Order("display_order ASC, start_date DESC").Find(&experiences).Error; err != nil {
		return nil, 0, err
	}

	return experiences, total, nil
}

func (r *postgresRepository) Get(ctx context.Context, id string) (*entity.Experience, error) {
	var experience entity.Experience
	if err := r.db.WithContext(ctx).First(&experience, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &experience, nil
}

func (r *postgresRepository) Create(ctx context.Context, experience *entity.Experience) error {
	return r.db.WithContext(ctx).Create(experience).Error
}

func (r *postgresRepository) Update(ctx context.Context, experience *entity.Experience) error {
	return r.db.WithContext(ctx).Save(experience).Error
}

func (r *postgresRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Experience{}, "id = ?", id).Error
}
