package repository

import (
	"fmt"
	"time"

	"job-tracker/internal/entity"

	"gorm.io/gorm"
)

type ApplicationAnalysisDLQRepository interface {
	Insert(dlq *entity.ApplicationAnalysisDLQ) error
	FindAllByUserID(userID int) ([]entity.ApplicationAnalysisDLQ, error)
	FindByUUID(uuid string) (*entity.ApplicationAnalysisDLQ, error)
	Delete(userID int, uuid string) error
	Requeue(rtype string, userID int, dlqUUID string) error
}

type applicationAnalysisDLQRepository struct {
	db *gorm.DB
}

func NewApplicationAnalysisDLQRepository(db *gorm.DB) ApplicationAnalysisDLQRepository {
	return &applicationAnalysisDLQRepository{db: db}
}

func (r *applicationAnalysisDLQRepository) Insert(dlq *entity.ApplicationAnalysisDLQ) error {
	return r.db.Create(dlq).Error
}

func (r *applicationAnalysisDLQRepository) FindAllByUserID(userID int) ([]entity.ApplicationAnalysisDLQ, error) {
	var dlqs []entity.ApplicationAnalysisDLQ
	err := r.db.Where("user_id = ? AND deleted_at IS NULL", userID).Find(&dlqs).Error
	return dlqs, err
}

func (r *applicationAnalysisDLQRepository) FindByUUID(uuid string) (*entity.ApplicationAnalysisDLQ, error) {
	var dlq entity.ApplicationAnalysisDLQ
	err := r.db.Where("uuid = ?", uuid).First(&dlq).Error
	return &dlq, err
}

func (r *applicationAnalysisDLQRepository) Delete(userID int, uuid string) error {
	now := time.Now()
	result := r.db.Model(&entity.ApplicationAnalysisDLQ{}).
		Where("uuid = ? AND user_id = ? AND deleted_at IS NULL", uuid, userID).
		Update("deleted_at", &now)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *applicationAnalysisDLQRepository) Requeue(rtype string, userID int, dlqUUID string) error {
	if rtype == "single" {
		return r.db.Transaction(func(tx *gorm.DB) error {
			var dlq entity.ApplicationAnalysisDLQ
			if err := tx.Where("uuid = ?", dlqUUID).First(&dlq).Error; err != nil {
				return err
			}
			if dlq.UserID != userID {
				return fmt.Errorf("forbidden")
			}

			var job entity.ApplicationAnalysisJob
			if err := tx.Where("uuid = ?", dlq.JobUUID.String()).First(&job).Error; err != nil {
				return err
			}
			if err := tx.Model(&job).Updates(map[string]interface{}{
				"status":        "pending",
				"retry_count":   0,
				"next_retry_at": nil,
				"error_msg":     nil,
			}).Error; err != nil {
				return err
			}

			now := time.Now()
			return tx.Model(&entity.ApplicationAnalysisDLQ{}).
				Where("uuid = ?", dlqUUID).
				Update("deleted_at", &now).Error
		})
	}

	if rtype == "bulk" {
		return r.db.Transaction(func(tx *gorm.DB) error {
			var dlqs []entity.ApplicationAnalysisDLQ
			if err := tx.Where("user_id = ? AND deleted_at IS NULL", userID).Find(&dlqs).Error; err != nil {
				return err
			}
			if len(dlqs) == 0 {
				return nil
			}

			jobUUIDs := make([]string, len(dlqs))
			for i, d := range dlqs {
				jobUUIDs[i] = d.JobUUID.String()
			}

			if err := tx.Model(&entity.ApplicationAnalysisJob{}).
				Where("uuid IN ?", jobUUIDs).
				Updates(map[string]interface{}{
					"status":        "pending",
					"retry_count":   0,
					"next_retry_at": nil,
					"error_msg":     nil,
				}).Error; err != nil {
				return err
			}

			now := time.Now()
			return tx.Model(&entity.ApplicationAnalysisDLQ{}).
				Where("user_id = ? AND deleted_at IS NULL", userID).
				Update("deleted_at", &now).Error
		})
	}

	return nil
}
