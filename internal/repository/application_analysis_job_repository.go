package repository

import (
	"time"

	"job-tracker/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ApplicationAnalysisJobRepository interface {
	Create(job *entity.ApplicationAnalysisJob) error
	FindByUUID(uuid string) (*entity.ApplicationAnalysisJob, error)
	FindLatestByApplicationID(applicationID int) (*entity.ApplicationAnalysisJob, error)
	FindAllByApplicationID(applicationID int) ([]entity.ApplicationAnalysisJob, error)
	HasActiveJob(applicationID int) (bool, error)
	ClaimNext() (*entity.ApplicationAnalysisJob, error)
	MarkDone(job *entity.ApplicationAnalysisJob) error
	MarkFailed(job *entity.ApplicationAnalysisJob, errMsg string) error
	ResetStale() error
}

type applicationAnalysisJobRepository struct {
	db *gorm.DB
}

func NewApplicationAnalysisJobRepository(db *gorm.DB) ApplicationAnalysisJobRepository {
	return &applicationAnalysisJobRepository{db: db}
}

func (r *applicationAnalysisJobRepository) Create(job *entity.ApplicationAnalysisJob) error {
	return r.db.Create(job).Error
}

func (r *applicationAnalysisJobRepository) FindByUUID(id string) (*entity.ApplicationAnalysisJob, error) {
	var job entity.ApplicationAnalysisJob
	if err := r.db.Where("uuid = ?", id).First(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *applicationAnalysisJobRepository) FindLatestByApplicationID(applicationID int) (*entity.ApplicationAnalysisJob, error) {
	var job entity.ApplicationAnalysisJob
	err := r.db.Where("application_id = ?", applicationID).
		Order("created_at DESC").
		First(&job).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *applicationAnalysisJobRepository) FindAllByApplicationID(applicationID int) ([]entity.ApplicationAnalysisJob, error) {
	var jobs []entity.ApplicationAnalysisJob
	err := r.db.Where("application_id = ?", applicationID).
		Order("created_at DESC").
		Find(&jobs).Error
	return jobs, err
}

func (r *applicationAnalysisJobRepository) HasActiveJob(applicationID int) (bool, error) {
	var count int64
	err := r.db.Model(&entity.ApplicationAnalysisJob{}).
		Where("application_id = ? AND status IN ?", applicationID, []string{"pending", "processing"}).
		Count(&count).Error
	return count > 0, err
}

func (r *applicationAnalysisJobRepository) ClaimNext() (*entity.ApplicationAnalysisJob, error) {
	var job entity.ApplicationAnalysisJob
	err := r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Raw(
			`SELECT * FROM application_analysis_jobs
			WHERE status = 'pending'
			AND (next_retry_at IS NULL OR next_retry_at <= NOW())
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
		return tx.Model(&job).Update("status", "processing").Error
	})
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *applicationAnalysisJobRepository) MarkDone(job *entity.ApplicationAnalysisJob) error {
	job.Status = "done"
	now := time.Now()
	job.TimeFinished = &now
	return r.db.Save(job).Error
}

func (r *applicationAnalysisJobRepository) MarkFailed(job *entity.ApplicationAnalysisJob, errMsg string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		applyAppAnalysisRetryLogic(job, errMsg, time.Now())
		if err := tx.Save(job).Error; err != nil {
			return err
		}
		if job.Status == "failed" {
			now := time.Now()
			dlq := &entity.ApplicationAnalysisDLQ{
				UUID:     uuid.New(),
				JobUUID:  job.UUID,
				UserID:   job.UserID,
				ErrorMsg: *job.ErrorMsg,
				FailedAt: &now,
			}
			if err := tx.Create(dlq).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *applicationAnalysisJobRepository) ResetStale() error {
	return r.db.Model(&entity.ApplicationAnalysisJob{}).
		Where("status = ?", "processing").
		Updates(map[string]interface{}{
			"status":        "pending",
			"next_retry_at": nil,
		}).Error
}

func applyAppAnalysisRetryLogic(job *entity.ApplicationAnalysisJob, errMsg string, now time.Time) {
	delay := time.Second * 30 * (1 << job.RetryCount)
	if job.RetryCount+1 < job.MaxRetries {
		job.RetryCount++
		job.Status = "pending"
		retryAt := now.Add(delay)
		job.NextRetryAt = &retryAt
	} else {
		job.Status = "failed"
		job.ErrorMsg = &errMsg
		job.NextRetryAt = nil
		job.TimeFinished = &now
	}
}
