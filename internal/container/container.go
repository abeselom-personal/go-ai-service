// internal/container/container.go
package container

import (
	"github.com/abeselom-personal/go-ai-service/internal/repositories"
	"gorm.io/gorm"
)

type Container struct {
	DB              *gorm.DB
	AIProviderRepo  repositories.AIProviderRepository
	PromptRepo      repositories.PromptRepository
	PromptCacheRepo repositories.PromptCacheRepository
	UserRepo        repositories.UserRepository
	RateLimitRepo   repositories.RateLimitRepository
}

// Initialize with actual GORM v2 instance
func NewContainer(db *gorm.DB) *Container {
	return &Container{
		DB:              db,
		AIProviderRepo:  repositories.NewAIProviderRepository(db),
		PromptRepo:      repositories.NewPromptRepository(db),
		PromptCacheRepo: repositories.NewPromptCacheRepository(db),
		UserRepo:        repositories.NewUserRepository(db),
		RateLimitRepo:   repositories.NewRateLimitRepository(db),
	}
}
