package entity

import (
	"time"

	"github.com/google/uuid"
)

type ApplicationAnalysisDLQ struct {
	ID          int        `json:"id"           gorm:"primaryKey"`
	UUID        uuid.UUID  `json:"uuid"         gorm:"uniqueIndex;not null"`
	JobUUID     uuid.UUID  `json:"job_uuid"     gorm:"not null"`
	UserID      int        `json:"user_id"      gorm:"not null"`
	ErrorMsg    string     `json:"error_msg"    gorm:"not null"`
	FailureType *string    `json:"failure_type"`
	FailedAt    *time.Time `json:"failed_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}
