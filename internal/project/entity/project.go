// internal/project/entity/project.go

package entity

import "time"

type Project struct {
	ID               string     `gorm:"type:char(26);primaryKey"`
	Slug             string     `gorm:"type:text;uniqueIndex;not null"`
	Title            string     `gorm:"type:text;not null"`
	Summary          string     `gorm:"type:text"`
	Description      string     `gorm:"type:text"`
	Architecture     string     `gorm:"type:text"`
	RepositoryURL    string     `gorm:"type:text"`
	DemoURL          string     `gorm:"type:text"`
	Thumbnail        string     `gorm:"type:text"`
	Featured         bool       `gorm:"default:false"`
	Status           string     `gorm:"type:text;default:'Draft'"`
	GithubStars      int32      `gorm:"default:0"`
	GithubLastCommit *time.Time `gorm:"type:timestamptz"`
	ReadTime         int32      `gorm:"default:0"`
	CreatedAt        time.Time  `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt        time.Time  `gorm:"type:timestamptz;not null;default:now()"`
}

// TableName overrides the table name used by Project to projects.
func (Project) TableName() string {
	return "projects"
}
