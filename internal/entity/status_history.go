package entity

import "time"

type StatusHistory struct {
	ID            int       `gorm:"primaryKey;autoIncrement" json:"id"`
	ApplicationID int       `gorm:"not null" json:"application_id"`
	StatusID      int       `gorm:"not null" json:"status_id"`
	CreatedAt     time.Time `gorm:"not null" json:"created_at"`

	Application *Application `gorm:"foreignKey:ApplicationID" json:"application,omitempty"`
	Status      *Status     `gorm:"foreignKey:StatusID" json:"status,omitempty"`
}
