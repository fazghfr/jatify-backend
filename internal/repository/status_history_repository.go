package repository

import (
	"job-tracker/internal/entity"

	"gorm.io/gorm"
)

type StatusHistoryRepository interface {
	Create(h *entity.StatusHistory) error
}

type statusHistoryRepository struct {
	db *gorm.DB
}

func NewStatusHistoryRepository(db *gorm.DB) StatusHistoryRepository {
	return &statusHistoryRepository{db: db}
}

func (r *statusHistoryRepository) Create(h *entity.StatusHistory) error {
	return r.db.Create(h).Error
}
