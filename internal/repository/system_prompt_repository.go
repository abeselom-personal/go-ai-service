// internal/repository/system_prompt_repo.go
package repository

import (
	"context"

	models "github.com/abeselom-personal/go-ai-service/internal/model"
	"gorm.io/gorm"
)

type SystemPromptRepo struct {
	db *gorm.DB
}

func NewSystemPromptRepo(db *gorm.DB) *SystemPromptRepo {
	return &SystemPromptRepo{db}
}

func (r *SystemPromptRepo) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, contextTxKey, tx)
		return fn(txCtx)
	})
}

func (r *SystemPromptRepo) Create(ctx context.Context, sp *models.SystemPrompt) error {
	return getDB(ctx, r.db).WithContext(ctx).Create(sp).Error
}

func (r *SystemPromptRepo) GetByHash(ctx context.Context, hash string) (*models.SystemPrompt, error) {
	var sp models.SystemPrompt
	err := getDB(ctx, r.db).WithContext(ctx).Where("prompt_hash = ?", hash).First(&sp).Error
	return &sp, err
}

func (r *SystemPromptRepo) Update(ctx context.Context, sp *models.SystemPrompt) error {
	return getDB(ctx, r.db).WithContext(ctx).Save(sp).Error
}

func (r *SystemPromptRepo) Delete(ctx context.Context, id string) error {
	return getDB(ctx, r.db).WithContext(ctx).Delete(&models.SystemPrompt{}, "id = ?", id).Error
}

func (r *SystemPromptRepo) List(ctx context.Context) ([]models.SystemPrompt, error) {
	var prompts []models.SystemPrompt
	err := getDB(ctx, r.db).WithContext(ctx).Find(&prompts).Error
	return prompts, err
}
