package repository

import (
	"context"
	"dan-ai/internal/embedding/entity"
	"gorm.io/gorm"
)

type Repository interface {
	ListEnabledProfiles(ctx context.Context) ([]entity.EmbeddingProfile, error)
	GetProfileByName(ctx context.Context, name string) (*entity.EmbeddingProfile, error)
}

type postgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) ListEnabledProfiles(ctx context.Context) ([]entity.EmbeddingProfile, error) {
	var profiles []entity.EmbeddingProfile
	err := r.db.WithContext(ctx).Where("enabled = ?", true).Find(&profiles).Error
	return profiles, err
}

func (r *postgresRepository) GetProfileByName(ctx context.Context, name string) (*entity.EmbeddingProfile, error) {
	var profile entity.EmbeddingProfile
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}
