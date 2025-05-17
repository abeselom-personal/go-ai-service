// internal/repositories/rate_limit_repository.go
package repositories

import (
	"context"

	"github.com/abeselom-personal/go-ai-service/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RateLimitRepository interface {
	Increment(ctx context.Context, userID, providerID uuid.UUID) (*models.RateLimit, error)
	Get(ctx context.Context, userID, providerID uuid.UUID) (*models.RateLimit, error)
	Reset(ctx context.Context, userID, providerID uuid.UUID) error
}

type rateLimitRepo struct {
	db *gorm.DB
}

func NewRateLimitRepository(db *gorm.DB) RateLimitRepository {
	return &rateLimitRepo{db: db}
}

func (r *rateLimitRepo) Increment(ctx context.Context, userID, providerID uuid.UUID) (*models.RateLimit, error) {
	var limit models.RateLimit
	err := r.db.WithContext(ctx).Where(models.RateLimit{
		UserID:       userID,
		AIProviderID: providerID,
	}).Assign(map[string]any{
		"request_count":     gorm.Expr("request_count + 1"),
		"last_request_time": gorm.Expr("NOW()"),
	}).FirstOrCreate(&limit).Error
	return &limit, err
}

func (r *rateLimitRepo) Get(ctx context.Context, userID, providerID uuid.UUID) (*models.RateLimit, error) {
	var limit models.RateLimit
	err := r.db.WithContext(ctx).Where("user_id = ? AND ai_provider_id = ?", userID, providerID).First(&limit).Error
	return &limit, err
}

func (r *rateLimitRepo) Reset(ctx context.Context, userID, providerID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.RateLimit{}).
		Where("user_id = ? AND ai_provider_id = ?", userID, providerID).
		Updates(map[string]any{
			"request_count":     0,
			"last_request_time": gorm.Expr("NOW()"),
		}).Error
}
