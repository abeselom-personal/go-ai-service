// internal/repositories/prompt_cache_repository.go
package repositories

import (
	"context"

	"github.com/abeselom-personal/go-ai-service/internal/models"
	"gorm.io/gorm"
)

type PromptCacheRepository interface {
	Create(ctx context.Context, cache *models.PromptCache) error
	FindByHash(ctx context.Context, hash string) (*models.PromptCache, error)
	DeleteExpired(ctx context.Context) error
}

type promptCacheRepo struct {
	db *gorm.DB
}

func NewPromptCacheRepository(db *gorm.DB) PromptCacheRepository {
	return &promptCacheRepo{db: db}
}

func (r *promptCacheRepo) Create(ctx context.Context, cache *models.PromptCache) error {
	return r.db.WithContext(ctx).Create(cache).Error
}

func (r *promptCacheRepo) FindByHash(ctx context.Context, hash string) (*models.PromptCache, error) {
	var cache models.PromptCache
	err := r.db.WithContext(ctx).Where("hash = ?", hash).First(&cache).Error
	return &cache, err
}

func (r *promptCacheRepo) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("expires_at < NOW()").Delete(&models.PromptCache{}).Error
}
