package repository

import (
	"job-tracker/internal/entity"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ResumeAnalysisJobRepository interface {
	Create(job *entity.ResumeAnalysisJob) error
	FindByUUID(uuid string) (*entity.ResumeAnalysisJob, error)
	ClaimNext() (*entity.ResumeAnalysisJob, error)
	MarkDone(job *entity.ResumeAnalysisJob) error
	MarkFailed(job *entity.ResumeAnalysisJob, errmsg string) error
	UpdateStatus(job *entity.ResumeAnalysisJob, newstatus string) error
	UpdateStatusByUUIDs(uuids []string, status string) error
	ResetStale() error
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
				AND  (next_retry_at IS NULL OR next_retry_at <= NOW())
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
	now := time.Now()
	job.TimeFinished = &now
	return r.db.Save(job).Error
}

func (r *ResumeAnalyzerJobRepository) ResetStale() error {
	err := r.db.Model(&entity.ResumeAnalysisJob{}).
    Where("status = ?", "processing").
    Updates(map[string]interface{}{
                "status":        "pending",
                "next_retry_at": nil,
            }).Error
	return err
}

// applyRetryLogic mutates job with the next retry state or marks it permanently failed.
// extracted so it can be unit tested
func applyRetryLogic(job *entity.ResumeAnalysisJob, errmsg string, now time.Time) {
	delay := time.Second * 30 * (1 << job.RetryCount)
	if job.RetryCount+1 <= job.MaxRetries {
		job.RetryCount++
		job.Status = "pending"
		retryAt := now.Add(delay)
		job.NextRetryAt = &retryAt
	} else {
		job.Status = "failed"
		job.ErrorMsg = &errmsg
		job.NextRetryAt = nil
		job.TimeFinished = &now
	}
}

func (r *ResumeAnalyzerJobRepository) MarkFailed(job *entity.ResumeAnalysisJob, errmsg string) error {
	return r.db.Transaction(
		func(tx *gorm.DB) error {
			applyRetryLogic(job, errmsg, time.Now())
			if err := tx.Save(&job).Error; err != nil {
				return err
			}

			// push to dlq
			now := time.Now()
			if job.Status == "failed" {
				newDlq := &entity.ResumeAnalysisDLQ{
					UUID:     uuid.New(),
					JobUUID:  job.UUID,
					ErrorMsg: *job.ErrorMsg,
					FailedAt: &now,
					FailureType: nil,
				}
				if err := tx.Save(newDlq).Error; err != nil {
					return err
				}
			}
			return nil
		},
	)
}

// THIS FUNCTION IS INTENDED FOR REQUEUEING LOGIC ONLY.
// refer to dlq_service@requeue function
func (r *ResumeAnalyzerJobRepository) UpdateStatus(job *entity.ResumeAnalysisJob, newstatus string) error {
	job.Status = newstatus
	return r.db.Save(job).Error
}

func (r *ResumeAnalyzerJobRepository) UpdateStatusByUUIDs(uuids []string, status string) error {
	return r.db.Model(&entity.ResumeAnalysisJob{}).
		Where("uuid IN ?", uuids).
		Update("status", status).Error
}
