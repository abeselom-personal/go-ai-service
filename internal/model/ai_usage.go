package models

import (
	"time"

	"github.com/google/uuid"
)

type AIUsageLog struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ModuleName string    `gorm:"index;not null"`
	Provider   string    `gorm:"index;not null"`
	PromptHash string    `gorm:"index;not null"`
	Request    string    `gorm:"type:text;not null"`
	Response   string    `gorm:"type:text;not null"`
	UsedAt     time.Time `gorm:"autoCreateTime"`
}
