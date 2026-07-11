// internal/certificate/entity/certificate.go
package entity

import "time"

type Certificate struct {
	ID             string     `gorm:"type:char(26);primaryKey"`
	Title          string     `gorm:"type:text;not null"`
	Issuer         string     `gorm:"type:text;not null"`
	IssueDate      *time.Time `gorm:"type:date"`
	ExpirationDate *time.Time `gorm:"type:date"`
	CredentialID   string     `gorm:"type:text"`
	CredentialURL  string     `gorm:"type:text"`
	Thumbnail      string     `gorm:"type:text"`
	Skills         []string   `gorm:"serializer:json;type:jsonb"`
	IssuerLogo     string     `gorm:"type:text"`
	CreatedAt      time.Time  `gorm:"type:timestamptz;not null;default:now()"`
}

// TableName overrides the table name used by Certificate to certificates.
func (Certificate) TableName() string {
	return "certificates"
}
