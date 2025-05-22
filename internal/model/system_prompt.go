// models/system_prompt.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SystemPrompt struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ModuleName   string    `gorm:"index;not null"`
	ModelName    string    `gorm:"index;not null"`
	Provider     string    `gorm:"index;not null"`
	SystemPrompt string    `gorm:"type:text;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}
