package entity

import "github.com/google/uuid"

type Resume struct {
	ID       int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Filepath string    `gorm:"type:text;not null" json:"filepath"`
	UserID   int       `gorm:"not null" json:"user_id"`
	UUID     uuid.UUID `gorm:"type:varchar(36);not null" json:"uuid"`

	User         *User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Applications []Application `gorm:"foreignKey:ResumeID" json:"applications,omitempty"`
}
