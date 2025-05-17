package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type AIProvider struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string    `gorm:"not null"`
	APIUrl    string    `gorm:"not null"`
	APIKey    string    `gorm:"not null"`
	Model     string    `gorm:"not null"`
	IsDefault bool      `gorm:"default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Prompt struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name         string    `gorm:"not null"`
	Content      string    `gorm:"not null"`
	TenantID     uuid.UUID
	AIProviderID uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PromptCache struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Hash         string    `gorm:"uniqueIndex;not null"`
	Request      JSONB     `gorm:"type:jsonb;not null"`
	Response     JSONB     `gorm:"type:jsonb;not null"`
	AIProviderID uuid.UUID
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string
	Email     string `gorm:"unique"`
	CreatedAt time.Time
}

type RateLimit struct {
	UserID          uuid.UUID `gorm:"type:uuid;primary_key"`
	AIProviderID    uuid.UUID `gorm:"type:uuid;primary_key"`
	RequestCount    int       `gorm:"default:0"`
	LastRequestTime time.Time
}

type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &j)
}
