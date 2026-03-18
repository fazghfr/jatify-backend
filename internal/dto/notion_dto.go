package dto

import "time"

type NotionConfigureRequest struct {
	DatabaseID string `json:"database_id" binding:"required"`
}

type NotionStatusResponse struct {
	Connected     bool       `json:"connected"`
	WorkspaceID   string     `json:"workspace_id,omitempty"`
	WorkspaceName string     `json:"workspace_name,omitempty"`
	DatabaseID    string     `json:"database_id,omitempty"`
	LastSyncAt    *time.Time `json:"last_sync_at,omitempty"`
}

type NotionDatabaseItem struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type NotionSyncResult struct {
	Created int      `json:"created"`
	Updated int      `json:"updated"`
	Skipped int      `json:"skipped"`
	Errors  []string `json:"errors"`
}
