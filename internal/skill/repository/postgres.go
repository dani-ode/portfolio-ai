// internal/skill/repository/postgres.go
package repository

import (
	"context"
	"portfolio-ai/internal/skill/entity"

	"gorm.io/gorm"
)

// Repository defines the interface for Skill database operations.
type Repository interface {
	List(ctx context.Context, page, limit int) ([]*entity.Skill, int64, error)
	Get(ctx context.Context, id string) (*entity.Skill, error)
	Create(ctx context.Context, skill *entity.Skill) error
	Update(ctx context.Context, skill *entity.Skill) error
	Delete(ctx context.Context, id string) error
}

type postgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a new Repository implementation using GORM.
func NewPostgresRepository(db *gorm.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) List(ctx context.Context, page, limit int) ([]*entity.Skill, int64, error) {
	var skills []*entity.Skill
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Skill{})
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

	if err := query.Preload("Technology").Order("display_order ASC").Find(&skills).Error; err != nil {
		return nil, 0, err
	}

	return skills, total, nil
}

func (r *postgresRepository) Get(ctx context.Context, id string) (*entity.Skill, error) {
	var skill entity.Skill
	if err := r.db.WithContext(ctx).Preload("Technology").First(&skill, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &skill, nil
}

func (r *postgresRepository) Create(ctx context.Context, skill *entity.Skill) error {
	if err := r.db.WithContext(ctx).Create(skill).Error; err != nil {
		return err
	}
	// Reload to fetch the Technology relation
	return r.db.WithContext(ctx).Preload("Technology").First(skill, "id = ?", skill.ID).Error
}

func (r *postgresRepository) Update(ctx context.Context, skill *entity.Skill) error {
	if err := r.db.WithContext(ctx).Save(skill).Error; err != nil {
		return err
	}
	// Reload to fetch the Technology relation
	return r.db.WithContext(ctx).Preload("Technology").First(skill, "id = ?", skill.ID).Error
}

func (r *postgresRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Skill{}, "id = ?", id).Error
}
