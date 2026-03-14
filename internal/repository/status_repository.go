package repository

import (
	"job-tracker/internal/entity"

	"gorm.io/gorm"
)

type StatusRepository interface {
	FindAll() ([]entity.Status, error)
}

type statusRepository struct {
	db *gorm.DB
}

func NewStatusRepository(db *gorm.DB) StatusRepository {
	return &statusRepository{db: db}
}

func (r *statusRepository) FindAll() ([]entity.Status, error) {
	var statuses []entity.Status
	err := r.db.Order("id asc").Find(&statuses).Error
	return statuses, err
}
