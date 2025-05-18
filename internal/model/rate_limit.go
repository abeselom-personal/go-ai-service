package models

import (
	"github.com/google/uuid"
)

type RateLimit struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ModuleName  string    `gorm:"index;not null"`
	Provider    string    `gorm:"index;not null"`
	MaxRequests int       `gorm:"not null"`
	PerSeconds  int       `gorm:"not null"`
}
