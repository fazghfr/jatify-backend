package repository

import (
	"job-tracker/internal/entity"

	"gorm.io/gorm"
)

type ResumeAnalyzerJobRepository struct {
	db *gorm.DB
}

func NewResumeAnalyzerJobRepository(db *gorm.DB) *ResumeAnalyzerJobRepository {
	return &ResumeAnalyzerJobRepository{db: db}
}

func (r *ResumeAnalyzerJobRepository) Create(job *entity.ResumeAnalysisJob) error {
	return r.db.Create(job).Error
}

func (r *ResumeAnalyzerJobRepository) Update(job *entity.ResumeAnalysisJob) error {
	return r.db.Save(job).Error
}
