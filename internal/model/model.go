// models/model.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Model struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	ProviderID uuid.UUID `gorm:"type:uuid;index"`
	Name       string    `gorm:"type:varchar(255);uniqueIndex:idx_provider_model"`
	Parameters string    `gorm:"type:jsonb"`
	Config     string    `gorm:"type:jsonb"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (m *Model) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
