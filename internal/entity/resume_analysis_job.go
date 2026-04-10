package entity

import (
	"time"

	"github.com/google/uuid"
)

type ResumeAnalysisJob struct {
	ID         int       `json:"id"          gorm:"primaryKey"`
	UUID       uuid.UUID `json:"uuid"        gorm:"uniqueIndex;not null"`
	ResumeID   int       `json:"resume_id"   gorm:"not null"`
	UserID     int       `json:"user_id"     gorm:"not null"`
	Status     string    `json:"status"      gorm:"not null;default:'pending'"`
	ResultJSON *string   `json:"result_json"`
	ErrorMsg   *string   `json:"error_msg"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
