package entity

import "time"

type EmbeddingProfile struct {
	ID                  string    `gorm:"type:char(26);primaryKey"`
	Name                string    `gorm:"type:text;uniqueIndex;not null"`
	Provider            string    `gorm:"type:text;not null"`
	Model               string    `gorm:"type:text;not null"`
	Dimension           int       `gorm:"type:int;not null"`
	MetricType          string    `gorm:"type:text;not null;default:'COSINE'"`
	KnowledgeCollection string    `gorm:"type:text;not null"`
	VisitorCollection   string    `gorm:"type:text;not null"`
	Enabled             bool      `gorm:"type:boolean;default:false"`
	CreatedAt           time.Time `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt           time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (EmbeddingProfile) TableName() string {
	return "embedding_profiles"
}
