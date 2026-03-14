package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Application struct {
	ID        int            `gorm:"primaryKey;autoIncrement" json:"id"`
	ResumeID  int            `gorm:"not null" json:"resume_id"`
	JobID     int            `gorm:"not null" json:"job_id"`
	Text      string         `gorm:"type:text;not null" json:"text"`
	CreatedAt time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	StatusID  int            `gorm:"not null" json:"status_id"`
	UserID    int            `gorm:"not null" json:"user_id"`
	UUID      uuid.UUID      `gorm:"type:varchar(36);not null" json:"uuid"`

	User          *User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Job           *Job            `gorm:"foreignKey:JobID" json:"job,omitempty"`
	Resume        *Resume         `gorm:"foreignKey:ResumeID" json:"resume,omitempty"`
	Status        *Status         `gorm:"foreignKey:StatusID" json:"status,omitempty"`
	StatusHistory []StatusHistory `gorm:"foreignKey:ApplicationID" json:"status_history,omitempty"`
}
