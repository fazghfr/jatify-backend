package entity

import (
	"time"

	"github.com/google/uuid"
)

type ResumeAnalysisDLQ struct {
	ID          int       `gorm:"primaryKey"`
	UUID        uuid.UUID `gorm:"uniqueIndex;not null"`
	JobUUID     uuid.UUID `gorm:"not null"`
	ErrorMsg    string    `gorm:"not null"`
	FailureType *string
	FailedAt    *time.Time `gorm:"not null"`
	DeletedAt	*time.Time
}
