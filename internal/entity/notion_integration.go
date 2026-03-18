package entity

import "time"

type NotionIntegration struct {
	ID            uint   `gorm:"primaryKey"`
	UserID        int    `gorm:"uniqueIndex;not null"`
	AccessToken   string `gorm:"not null"`
	WorkspaceID   string
	WorkspaceName string
	BotID         string
	DatabaseID    string
	LastSyncAt    *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
