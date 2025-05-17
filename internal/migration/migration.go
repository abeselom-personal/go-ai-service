package migrations

import (
	"github.com/abeselom-personal/go-ai-service/internal/models"
	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) error {
	// Enable required extensions
	db.Exec("CREATE INDEX IF NOT EXISTS idx_prompts_tenant ON prompts (tenant_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_cache_expiry ON prompt_caches (expires_at)")

	// Partial indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_default_providers ON ai_providers (is_default) WHERE is_default = true")
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\"").Error; err != nil {
		return err
	}
	// Auto migrate models
	return db.AutoMigrate(
		&models.AIProvider{},
		&models.Prompt{},
		&models.PromptCache{},
		&models.User{},
		&models.RateLimit{},
	)
}
