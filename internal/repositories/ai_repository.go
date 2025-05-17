package repositories

import (
	"github.com/abeselom-personal/go-ai-service/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AIProviderRepository interface {
	Create(provider *models.AIProvider) error
	FindByID(id uuid.UUID) (*models.AIProvider, error)
	Update(provider *models.AIProvider) error
	Delete(id uuid.UUID) error
	SetAsDefault(id uuid.UUID) error
}

type aiProviderRepo struct {
	db *gorm.DB
}

func NewAIProviderRepository(db *gorm.DB) AIProviderRepository {
	return &aiProviderRepo{db: db}
}

func (r *aiProviderRepo) Create(provider *models.AIProvider) error {
	return r.db.Create(provider).Error
}

func (r *aiProviderRepo) FindByID(id uuid.UUID) (*models.AIProvider, error) {
	var provider models.AIProvider
	err := r.db.Where("id = ?", id).First(&provider).Error
	return &provider, err
}

func (r *aiProviderRepo) Update(provider *models.AIProvider) error {
	return r.db.Save(provider).Error
}

func (r *aiProviderRepo) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.AIProvider{}).Error
}

func (r *aiProviderRepo) SetAsDefault(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.AIProvider{}).Update("is_default", false).Error; err != nil {
			return err
		}
		return tx.Model(&models.AIProvider{}).Where("id = ?", id).Update("is_default", true).Error
	})
}
