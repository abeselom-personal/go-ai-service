package repository

import (
	"context"
	"errors"

	models "github.com/abeselom-personal/go-ai-service/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrModelNotFound  = errors.New("model not found")
	ErrDuplicateModel = errors.New("model name must be unique per provider")
)

// ModelRepository defines the model persistence interface
type ModelRepository interface {
	Create(ctx context.Context, model *models.Model) error
	GetByID(ctx context.Context, id string) (*models.Model, error)
	GetByProviderAndName(ctx context.Context, providerID, name string) (*models.Model, error)
	Update(ctx context.Context, model *models.Model) error
	Delete(ctx context.Context, id string) error
	ListByProvider(ctx context.Context, providerID string) ([]models.Model, error)
}

type modelRepository struct {
	db *gorm.DB
}

// NewModelRepository creates a new model repository instance
func NewModelRepository(db *gorm.DB) ModelRepository {
	return &modelRepository{db: db}
}

func (r *modelRepository) Create(ctx context.Context, model *models.Model) error {
	if model.ProviderID == uuid.Nil {
		return errors.New("invalid provider ID")
	}

	err := r.db.WithContext(ctx).Create(model).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrDuplicateModel
	}
	return err
}

func (r *modelRepository) GetByID(ctx context.Context, id string) (*models.Model, error) {
	var model models.Model
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrModelNotFound
	}
	return &model, err
}

func (r *modelRepository) GetByProviderAndName(ctx context.Context, providerID, name string) (*models.Model, error) {
	var model models.Model
	err := r.db.WithContext(ctx).
		Where("provider_id = ? AND name = ?", providerID, name).
		First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrModelNotFound
	}
	return &model, err
}

func (r *modelRepository) Update(ctx context.Context, model *models.Model) error {
	if model.ID == uuid.Nil {
		return errors.New("invalid model ID")
	}

	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	result := tx.Model(&models.Model{}).
		Where("id = ?", model.ID).
		Updates(model)

	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return ErrModelNotFound
	}

	return tx.Commit().Error
}

func (r *modelRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&models.Model{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrModelNotFound
	}
	return nil
}

func (r *modelRepository) ListByProvider(ctx context.Context, providerID string) ([]models.Model, error) {
	var models []models.Model
	err := r.db.WithContext(ctx).
		Where("provider_id = ?", providerID).
		Find(&models).Error
	return models, err
}
