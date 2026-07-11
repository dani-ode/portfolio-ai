// internal/skill/entity/skill.go
package entity

import (
	"time"
	techEntity "portfolio-ai/internal/technology/entity"
)

type Skill struct {
	ID           string                `gorm:"type:char(26);primaryKey"`
	DisplayOrder int32                 `gorm:"default:0"`
	TechnologyID string                `gorm:"type:char(26);not null"`
	Technology   techEntity.Technology `gorm:"foreignKey:TechnologyID"`
	Level        string                `gorm:"type:text"`
	Years        float32               `gorm:"type:numeric(4,1);default:0.0"`
	Favorite     bool                  `gorm:"default:false"`
	CreatedAt    time.Time             `gorm:"type:timestamptz;not null;default:now()"`
}

// TableName overrides the table name used by Skill to skills.
func (Skill) TableName() string {
	return "skills"
}
