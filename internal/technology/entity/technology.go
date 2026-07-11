// internal/technology/entity/technology.go
package entity

import "time"

type Technology struct {
	ID          string    `gorm:"type:char(26);primaryKey"`
	Name        string    `gorm:"type:text;uniqueIndex;not null"`
	Category    string    `gorm:"type:text"`
	Icon        string    `gorm:"type:text"`
	Color       string    `gorm:"type:text"`
	OfficialURL string    `gorm:"type:text"`
	Logo        string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

// TableName overrides the table name used by Technology to technologies.
func (Technology) TableName() string {
	return "technologies"
}
