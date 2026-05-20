package entity

import (
	"time"

	"github.com/google/uuid"
)

type ApplicationAnalysisJob struct {
	ID            int        `json:"id"             gorm:"primaryKey"`
	UUID          uuid.UUID  `json:"uuid"           gorm:"uniqueIndex;not null"`
	ApplicationID int        `json:"application_id" gorm:"not null"`
	UserID        int        `json:"user_id"        gorm:"not null"`
	Status        string     `json:"status"         gorm:"not null;default:'pending'"`
	ResultJSON    string     `json:"result_json"`
	RetryCount    int        `json:"retry_count"    gorm:"default:0"`
	MaxRetries    int        `json:"max_retries"    gorm:"default:3"`
	NextRetryAt   *time.Time `json:"next_retry_at"`
	TimeFinished  *time.Time `json:"time_finished"`
	ErrorMsg      *string    `json:"error_msg"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type AppAnalysisResult struct {
	Score       int      `json:"score"`
	Strengths   []string `json:"strengths"`
	SkillsMatch []string `json:"skills_match"`
	Gaps        []string `json:"gaps"`
	Suggestions []string `json:"suggestions"`
}
