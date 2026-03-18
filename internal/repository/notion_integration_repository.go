package repository

import (
	"job-tracker/internal/entity"

	"gorm.io/gorm"
)

type NotionIntegrationRepository interface {
	Upsert(integration *entity.NotionIntegration) error
	FindByUserID(userID int) (*entity.NotionIntegration, error)
	DeleteByUserID(userID int) error
}

type notionIntegrationRepository struct {
	db *gorm.DB
}

func NewNotionIntegrationRepository(db *gorm.DB) NotionIntegrationRepository {
	return &notionIntegrationRepository{db: db}
}

func (r *notionIntegrationRepository) Upsert(integration *entity.NotionIntegration) error {
	return r.db.Save(integration).Error
}

func (r *notionIntegrationRepository) FindByUserID(userID int) (*entity.NotionIntegration, error) {
	var integration entity.NotionIntegration
	if err := r.db.Where("user_id = ?", userID).First(&integration).Error; err != nil {
		return nil, err
	}
	return &integration, nil
}

func (r *notionIntegrationRepository) DeleteByUserID(userID int) error {
	return r.db.Where("user_id = ?", userID).Delete(&entity.NotionIntegration{}).Error
}
