package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AIUsageLog struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	ModuleName string    `gorm:"index;not null"`
	Provider   string    `gorm:"index;not null"`
	PromptHash string    `gorm:"index;not null"`
	Request    string    `gorm:"type:text;not null"`
	Response   string    `gorm:"type:text;not null"`
	UsedAt     time.Time `gorm:"autoCreateTime"`
}

func (a *AIUsageLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
