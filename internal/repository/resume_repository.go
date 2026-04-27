package repository

import (
	"job-tracker/internal/entity"

	"gorm.io/gorm"
)

type ResumeRepository interface {
	Create(resume *entity.Resume) error
	FindAllByUserID(userID int) ([]entity.Resume, error)
	FindByID(id int) (*entity.Resume, error)
	FindByUUID(uuid string) (*entity.Resume, error)
	Update(resume *entity.Resume) error
	Delete(id int) error
}

type resumeRepository struct {
	db *gorm.DB
}

func NewResumeRepository(db *gorm.DB) ResumeRepository {
	return &resumeRepository{db: db}
}

func (r *resumeRepository) FindByUUID(uuid string) (*entity.Resume, error) {
	var resume entity.Resume
	if err := r.db.First(&resume, "uuid = ?", uuid).Error; err != nil {
		return nil, err
	}
	return &resume, nil
}

func (r *resumeRepository) Create(resume *entity.Resume) error {
	return r.db.Create(resume).Error
}

func (r *resumeRepository) FindAllByUserID(userID int) ([]entity.Resume, error) {
	var resumes []entity.Resume
	err := r.db.Where("user_id = ?", userID).Find(&resumes).Error
	return resumes, err
}

func (r *resumeRepository) FindByID(id int) (*entity.Resume, error) {
	var resume entity.Resume
	if err := r.db.First(&resume, id).Error; err != nil {
		return nil, err
	}
	return &resume, nil
}

func (r *resumeRepository) Update(resume *entity.Resume) error {
	return r.db.Save(resume).Error
}

func (r *resumeRepository) Delete(id int) error {
	return r.db.Delete(&entity.Resume{}, id).Error
}
