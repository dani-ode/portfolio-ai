// internal/project/repository/postgres.go
package repository

import (
	"context"
	"portfolio-ai/internal/project/entity"

	"gorm.io/gorm"
)

// Repository defines the interface for Project database operations.
type Repository interface {
	List(ctx context.Context, page, limit int) ([]*entity.Project, int64, error)
	Get(ctx context.Context, id string) (*entity.Project, error)
	Create(ctx context.Context, project *entity.Project) error
	Update(ctx context.Context, project *entity.Project) error
	Delete(ctx context.Context, id string) error
}

type postgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a new Repository implementation using GORM.
func NewPostgresRepository(db *gorm.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) List(ctx context.Context, page, limit int) ([]*entity.Project, int64, error) {
	var projects []*entity.Project
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Project{})
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

	if err := query.Order("created_at DESC").Find(&projects).Error; err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

func (r *postgresRepository) Get(ctx context.Context, id string) (*entity.Project, error) {
	var project entity.Project
	if err := r.db.WithContext(ctx).First(&project, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *postgresRepository) Create(ctx context.Context, project *entity.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *postgresRepository) Update(ctx context.Context, project *entity.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

func (r *postgresRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Project{}, "id = ?", id).Error
}
