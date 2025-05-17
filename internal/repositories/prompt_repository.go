// internal/repositories/prompt_repository.go
package repositories

import (
	"context"

	"github.com/abeselom-personal/go-ai-service/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PromptRepository interface {
	Create(ctx context.Context, prompt *models.Prompt) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Prompt, error)
	FindByTenant(ctx context.Context, tenantID uuid.UUID) ([]models.Prompt, error)
	Update(ctx context.Context, prompt *models.Prompt) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type promptRepo struct {
	db *gorm.DB
}

func NewPromptRepository(db *gorm.DB) PromptRepository {
	return &promptRepo{db: db}
}

func (r *promptRepo) Create(ctx context.Context, prompt *models.Prompt) error {
	return r.db.WithContext(ctx).Create(prompt).Error
}

func (r *promptRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Prompt, error) {
	var prompt models.Prompt
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&prompt).Error
	return &prompt, err
}

func (r *promptRepo) FindByTenant(ctx context.Context, tenantID uuid.UUID) ([]models.Prompt, error) {
	var prompts []models.Prompt
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&prompts).Error
	return prompts, err
}

func (r *promptRepo) Update(ctx context.Context, prompt *models.Prompt) error {
	return r.db.WithContext(ctx).Save(prompt).Error
}

func (r *promptRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Prompt{}).Error
}
