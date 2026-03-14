package repository

import (
	"job-tracker/internal/entity"

	"gorm.io/gorm"
)

type ApplicationRepository interface {
	Create(app *entity.Application) error
	FindAllByUserID(userID int) ([]entity.Application, error)
	FindByID(id int) (*entity.Application, error)
	Update(app *entity.Application) error
	Delete(id int) error
}

type applicationRepository struct {
	db *gorm.DB
}

func NewApplicationRepository(db *gorm.DB) ApplicationRepository {
	return &applicationRepository{db: db}
}

func (r *applicationRepository) Create(app *entity.Application) error {
	return r.db.Create(app).Error
}

func (r *applicationRepository) FindAllByUserID(userID int) ([]entity.Application, error) {
	var apps []entity.Application
	err := r.db.Where("user_id = ?", userID).
		Preload("Status").Preload("Job").
		Find(&apps).Error
	return apps, err
}

func (r *applicationRepository) FindByID(id int) (*entity.Application, error) {
	var app entity.Application
	err := r.db.Preload("Status").Preload("Job").Preload("Resume").Preload("StatusHistory").
		First(&app, id).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *applicationRepository) Update(app *entity.Application) error {
	return r.db.Save(app).Error
}

func (r *applicationRepository) Delete(id int) error {
	return r.db.Delete(&entity.Application{}, id).Error
}
