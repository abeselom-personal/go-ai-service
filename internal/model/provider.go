// models/provider.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Provider struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name      string    `gorm:"type:varchar(255);uniqueIndex"`
	BaseURL   string    `gorm:"type:varchar(512)"`
	APIKey    string    `gorm:"type:text"`
	Default   bool
	CreatedAt time.Time
	UpdatedAt time.Time
	Models    []Model `gorm:"foreignKey:ProviderID;constraint:OnDelete:CASCADE"`
}

func (p *Provider) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
