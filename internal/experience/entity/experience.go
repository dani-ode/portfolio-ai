// internal/experience/entity/experience.go
package entity

import "time"

type Experience struct {
	ID             string     `gorm:"type:char(26);primaryKey"`
	Company        string     `gorm:"type:text;not null"`
	Position       string     `gorm:"type:text;not null"`
	EmploymentType string     `gorm:"type:text"`
	StartDate      *time.Time `gorm:"type:date"`
	EndDate        *time.Time `gorm:"type:date"`
	CurrentJob     bool       `gorm:"default:false"`
	Location       string     `gorm:"type:text"`
	Description    string     `gorm:"type:text"`
	DisplayOrder   int32      `gorm:"default:0"`
	CompanyLogo    string     `gorm:"type:text"`
	Skills         []string   `gorm:"serializer:json;type:jsonb"`
	RemoteType     string     `gorm:"type:text"`
	CreatedAt      time.Time  `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt      time.Time  `gorm:"type:timestamptz;not null;default:now()"`
}

// TableName overrides the table name used by Experience to experiences.
func (Experience) TableName() string {
	return "experiences"
}
