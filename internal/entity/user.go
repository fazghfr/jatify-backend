package entity

import "github.com/google/uuid"

type User struct {
	ID       int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Username string    `gorm:"type:varchar(64);not null" json:"username"`
	Name     string    `gorm:"type:varchar(64);not null" json:"name"`
	Phone    string    `gorm:"type:varchar(64);not null" json:"phone"`
	Email    string    `gorm:"type:varchar(128);not null" json:"email"`
	Password string    `gorm:"type:varchar(64);not null" json:"-"`
	UUID     uuid.UUID `gorm:"type:varchar(36);not null" json:"uuid"`

	Jobs         []Job         `gorm:"foreignKey:UserID" json:"jobs,omitempty"`
	Resumes      []Resume      `gorm:"foreignKey:UserID" json:"resumes,omitempty"`
	Applications []Application `gorm:"foreignKey:UserID" json:"applications,omitempty"`
}
