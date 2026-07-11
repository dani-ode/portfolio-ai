// internal/technology/repository/postgres.go
package repository

import (
	"context"
	"portfolio-ai/internal/technology/entity"

	"gorm.io/gorm"
)

// Repository defines the interface for Technology database operations.
type Repository interface {
	List(ctx context.Context, page, limit int) ([]*entity.Technology, int64, error)
	Get(ctx context.Context, id string) (*entity.Technology, error)
	Create(ctx context.Context, tech *entity.Technology) error
	Update(ctx context.Context, tech *entity.Technology) error
	Delete(ctx context.Context, id string) error
}

type postgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a new Repository implementation using GORM.
func NewPostgresRepository(db *gorm.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) List(ctx context.Context, page, limit int) ([]*entity.Technology, int64, error) {
	var techs []*entity.Technology
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Technology{})
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

	if err := query.Order("category ASC, name ASC").Find(&techs).Error; err != nil {
		return nil, 0, err
	}

	return techs, total, nil
}

func (r *postgresRepository) Get(ctx context.Context, id string) (*entity.Technology, error) {
	var tech entity.Technology
	if err := r.db.WithContext(ctx).First(&tech, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &tech, nil
}

func (r *postgresRepository) Create(ctx context.Context, tech *entity.Technology) error {
	return r.db.WithContext(ctx).Create(tech).Error
}

func (r *postgresRepository) Update(ctx context.Context, tech *entity.Technology) error {
	return r.db.WithContext(ctx).Save(tech).Error
}

func (r *postgresRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Technology{}, "id = ?", id).Error
}
