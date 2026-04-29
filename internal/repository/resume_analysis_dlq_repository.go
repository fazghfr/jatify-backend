package repository

import (
	"fmt"
	"job-tracker/internal/entity"
	"time"

	"gorm.io/gorm"
)

type ResumeAnalysisDLQRepository interface {
	Insert(dlq *entity.ResumeAnalysisDLQ) error
	FindAllByUserID(userID int) ([]entity.ResumeAnalysisDLQ, error)
	FindByUUID(uuid string) (*entity.ResumeAnalysisDLQ, error)
	Delete(uuid string) error
	Requeue(Rtype string, userid int, dlqUUID string) error
}

type ResumeAnalysisDLQRepo struct {
	db *gorm.DB
}

func NewResumeAnalysisDLQRepository(db *gorm.DB) *ResumeAnalysisDLQRepo {
	return &ResumeAnalysisDLQRepo{db: db}
}

func (r *ResumeAnalysisDLQRepo) Insert(dlq *entity.ResumeAnalysisDLQ) error {
	return r.db.Create(dlq).Error
}

func (r *ResumeAnalysisDLQRepo) FindAllByUserID(userID int) ([]entity.ResumeAnalysisDLQ, error) {
	var dlqs []entity.ResumeAnalysisDLQ
	result := r.db.
		Joins("JOIN resume_analysis_jobs ON resume_analysis_jobs.uuid = resume_analysis_dlqs.job_uuid").
		Where("resume_analysis_jobs.user_id = ? AND resume_analysis_dlqs.deleted_at IS NULL", userID).
		Find(&dlqs)
	return dlqs, result.Error
}

func (r *ResumeAnalysisDLQRepo) FindByUUID(uuid string) (*entity.ResumeAnalysisDLQ, error) {
	var dlq entity.ResumeAnalysisDLQ
	result := r.db.Where("uuid = ?", uuid).First(&dlq)
	return &dlq, result.Error
}

func (r *ResumeAnalysisDLQRepo) Delete(uuid string) error {
	now := time.Now()
	result := r.db.Where("uuid = ?", uuid).Update("deleted_at", &now)
	return result.Error
}

func (r *ResumeAnalysisDLQRepo) Requeue(Rtype string, userid int, dlqUUID string) error {
	if Rtype == "single" {
		return r.db.Transaction(func(tx *gorm.DB) error {
			var dlqObj entity.ResumeAnalysisDLQ
			if err := tx.Where("uuid = ?", dlqUUID).First(&dlqObj).Error; err != nil {
				return err
			}

			var job entity.ResumeAnalysisJob
			if err := tx.Where("uuid = ?", dlqObj.JobUUID.String()).First(&job).Error; err != nil {
				return err
			}
			if job.UserID != userid {
				return fmt.Errorf("forbidden")
			}

			if err := tx.Model(&job).Update("status", "pending").Error; err != nil {
				return err
			}

			now := time.Now()
			return tx.Model(&entity.ResumeAnalysisDLQ{}).
				Where("uuid = ?", dlqUUID).
				Update("deleted_at", &now).Error
		})
	}

	if Rtype == "bulk" {
		return r.db.Transaction(func(tx *gorm.DB) error {
			var userJobs []entity.ResumeAnalysisJob
			if err := tx.Where("user_id = ?", userid).Select("uuid").Find(&userJobs).Error; err != nil {
				return err
			}

			if len(userJobs) == 0 {
				return nil
			}

			userJobUUIDs := make([]string, len(userJobs))
			for i, job := range userJobs {
				userJobUUIDs[i] = job.UUID.String()
			}

			var dlqObjs []entity.ResumeAnalysisDLQ
			if err := tx.Where("job_uuid IN ? AND deleted_at IS NULL", userJobUUIDs).Find(&dlqObjs).Error; err != nil {
				return err
			}

			if len(dlqObjs) == 0 {
				return nil
			}

			jobUUIDs := make([]string, len(dlqObjs))
			for i, dlqObj := range dlqObjs {
				jobUUIDs[i] = dlqObj.JobUUID.String()
			}

			if err := tx.Model(&entity.ResumeAnalysisJob{}).
				Where("uuid IN ?", jobUUIDs).
				Update("status", "pending").Error; err != nil {
				return err
			}

			now := time.Now()
			return tx.Model(&entity.ResumeAnalysisDLQ{}).
				Where("job_uuid IN ?", jobUUIDs).
				Update("deleted_at", &now).Error
		})
	}

	return nil
}
