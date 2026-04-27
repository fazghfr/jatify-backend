package repository

import (
	"job-tracker/internal/entity"
	"time"

	"gorm.io/gorm"
)

type ResumeAnalysisJobRepository interface {
	Create(job *entity.ResumeAnalysisJob) error
	FindByUUID(uuid string) (*entity.ResumeAnalysisJob, error)
	ClaimNext() (*entity.ResumeAnalysisJob, error)
	MarkDone(job *entity.ResumeAnalysisJob) error
	MarkFailed(job *entity.ResumeAnalysisJob, errmsg string) error
}

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

func (r *ResumeAnalyzerJobRepository) FindByUUID(uuid string) (*entity.ResumeAnalysisJob, error) {
	// placeholder temp
	job := &entity.ResumeAnalysisJob{}
	if err := r.db.Where("uuid = ?", uuid).First(job).Error; err != nil {
		return nil, err
	}
	return job, nil
}

func (r *ResumeAnalyzerJobRepository) ClaimNext() (*entity.ResumeAnalysisJob, error) {
	// using select for update skip locked
	var job entity.ResumeAnalysisJob
	err := r.db.Transaction(
		func(ctx *gorm.DB) error {
			res := ctx.Raw(
				`SELECT * FROM resume_analysis_jobs
				WHERE status = 'pending'
				ORDER BY created_at
				LIMIT 1
				FOR UPDATE SKIP LOCKED`,
			).Scan(&job)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
			return ctx.Model(&job).Update("status", "processing").Error
		})
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *ResumeAnalyzerJobRepository) MarkDone(job *entity.ResumeAnalysisJob) error {
	job.Status = "done"
	job.TimeFinished = time.Now()
	return r.db.Save(job).Error
}

func (r *ResumeAnalyzerJobRepository) MarkFailed(job *entity.ResumeAnalysisJob, errmsg string) error {
	// skipping the retry logic for mvp day 1
	// TODO: retry logic
	job.Status = "failed"
	job.ErrorMsg = errmsg
	job.TimeFinished = time.Now()
	return r.db.Save(&job).Error
}
