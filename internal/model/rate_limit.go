package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RateLimit struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	ModuleName  string    `gorm:"index;not null"`
	Provider    string    `gorm:"index;not null"`
	MaxRequests int       `gorm:"not null"`
	PerSeconds  int       `gorm:"not null"`
}

func (r *RateLimit) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
