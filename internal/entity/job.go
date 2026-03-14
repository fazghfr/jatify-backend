package entity

import (
	"time"

	"github.com/google/uuid"
)

type Job struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Company     string    `gorm:"type:varchar(256);not null" json:"company"`
	Position    string    `gorm:"type:varchar(256);not null" json:"position"`
	Description string    `gorm:"type:text;not null" json:"description"`
	CreatedAt   time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null" json:"updated_at"`
	UserID      int       `gorm:"not null" json:"user_id"`
	UUID        uuid.UUID `gorm:"type:varchar(255);not null" json:"uuid"`

	User         *User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Applications []Application `gorm:"foreignKey:JobID" json:"applications,omitempty"`
}
