package repository

import (
	"job-tracker/internal/entity"

	"gorm.io/gorm"
)

type JobRepository interface {
	Create(job *entity.Job) error
	FindAllByUserID(userID int) ([]entity.Job, error)
	FindByID(id int) (*entity.Job, error)
	FindByCompanyPositionUserID(userID int, company, position string) (*entity.Job, error)
	Update(job *entity.Job) error
	Delete(id int) error
}

type jobRepository struct {
	db *gorm.DB
}

func NewJobRepository(db *gorm.DB) JobRepository {
	return &jobRepository{db: db}
}

func (r *jobRepository) Create(job *entity.Job) error {
	return r.db.Create(job).Error
}

func (r *jobRepository) FindAllByUserID(userID int) ([]entity.Job, error) {
	var jobs []entity.Job
	err := r.db.Where("user_id = ?", userID).Find(&jobs).Error
	return jobs, err
}

func (r *jobRepository) FindByID(id int) (*entity.Job, error) {
	var job entity.Job
	if err := r.db.First(&job, id).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *jobRepository) FindByCompanyPositionUserID(userID int, company, position string) (*entity.Job, error) {
	var job entity.Job
	err := r.db.Where("user_id = ? AND company = ? AND position = ?", userID, company, position).
		First(&job).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *jobRepository) Update(job *entity.Job) error {
	return r.db.Save(job).Error
}

func (r *jobRepository) Delete(id int) error {
	return r.db.Delete(&entity.Job{}, id).Error
}
