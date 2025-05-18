package repository

import (
	"context"
	"errors"

	models "github.com/abeselom-personal/go-ai-service/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrProviderNotFound  = errors.New("provider not found")
	ErrDuplicateProvider = errors.New("provider already exists")
)

// ProviderRepository defines the provider persistence interface
type ProviderRepository interface {
	Create(ctx context.Context, provider *models.Provider) error
	GetByID(ctx context.Context, id string) (*models.Provider, error)
	GetByName(ctx context.Context, name string) (*models.Provider, error)
	Update(ctx context.Context, provider *models.Provider) error
	Delete(ctx context.Context, id string) error
	ListAll(ctx context.Context) ([]models.Provider, error)
	WithTransaction(ctx context.Context, fn func(context.Context) error) error
}

type providerRepository struct {
	db *gorm.DB
}

// NewProviderRepository creates a new provider repository instance
func NewProviderRepository(db *gorm.DB) ProviderRepository {
	return &providerRepository{db: db}
}

func (r *providerRepository) Create(ctx context.Context, provider *models.Provider) error {
	if provider.Name == "" {
		return errors.New("provider name cannot be empty")
	}
	err := r.db.WithContext(ctx).Create(provider).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrDuplicateProvider
	}
	return err
}

func (r *providerRepository) GetByID(ctx context.Context, id string) (*models.Provider, error) {
	var provider models.Provider
	err := r.db.WithContext(ctx).
		Preload("Models").
		Where("id = ?", uuid.MustParse(id)).
		First(&provider).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProviderNotFound
	}
	return &provider, err
}

func (r *providerRepository) GetByName(ctx context.Context, name string) (*models.Provider, error) {
	var provider models.Provider
	err := r.db.WithContext(ctx).
		Preload("Models").
		Where("name = ?", name).
		First(&provider).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProviderNotFound
	}
	return &provider, err
}

func (r *providerRepository) Update(ctx context.Context, provider *models.Provider) error {
	if provider.ID == uuid.Nil {
		return errors.New("invalid provider ID")
	}

	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	result := tx.Model(&models.Provider{}).
		Where("id = ?", provider.ID).
		Updates(provider)

	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return ErrProviderNotFound
	}

	return tx.Commit().Error
}

func (r *providerRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&models.Provider{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrProviderNotFound
	}
	return nil
}

func (r *providerRepository) ListAll(ctx context.Context) ([]models.Provider, error) {
	var providers []models.Provider
	err := r.db.WithContext(ctx).
		Preload("Models").
		Find(&providers).Error
	return providers, err
}
