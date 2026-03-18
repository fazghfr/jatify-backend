package handler

import (
	"errors"

	"job-tracker/internal/dto"
	"job-tracker/internal/middleware"
	"job-tracker/internal/service"
	"job-tracker/internal/util"

	"github.com/gin-gonic/gin"
)

type NotionHandler struct {
	notionSvc service.NotionService
}

func NewNotionHandler(notionSvc service.NotionService) *NotionHandler {
	return &NotionHandler{notionSvc: notionSvc}
}

// Connect GET /api/notion/connect
func (h *NotionHandler) Connect(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)
	url, err := h.notionSvc.Connect(userID)
	if err != nil {
		util.InternalError(c, "failed to build Notion OAuth URL")
		return
	}
	util.OK(c, gin.H{"url": url})
}

// Callback GET /api/notion/callback
func (h *NotionHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		util.BadRequest(c, "missing code or state query parameter")
		return
	}

	integration, err := h.notionSvc.Callback(code, state)
	if err != nil {
		util.BadRequest(c, err.Error())
		return
	}

	util.OK(c, gin.H{
		"workspace_id":   integration.WorkspaceID,
		"workspace_name": integration.WorkspaceName,
		"connected":      true,
	})
}

// Status GET /api/notion/status
func (h *NotionHandler) Status(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)
	status, err := h.notionSvc.Status(userID)
	if err != nil {
		util.InternalError(c, "failed to retrieve Notion status")
		return
	}
	util.OK(c, status)
}

// Configure POST /api/notion/configure
func (h *NotionHandler) Configure(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)

	var req dto.NotionConfigureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, err.Error())
		return
	}

	if err := h.notionSvc.Configure(userID, req.DatabaseID); err != nil {
		if errors.Is(err, service.ErrNotionNotConnected) {
			util.BadRequest(c, "Notion integration not connected")
			return
		}
		util.InternalError(c, "failed to configure Notion database")
		return
	}

	util.OK(c, gin.H{"message": "database configured successfully"})
}

// ListDatabases GET /api/notion/databases
func (h *NotionHandler) ListDatabases(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)

	databases, err := h.notionSvc.ListDatabases(userID)
	if err != nil {
		if errors.Is(err, service.ErrNotionNotConnected) {
			util.BadRequest(c, "Notion integration not connected")
			return
		}
		util.InternalError(c, "failed to list Notion databases")
		return
	}

	util.OK(c, databases)
}

// Sync POST /api/notion/sync
func (h *NotionHandler) Sync(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)

	syncResult, err := h.notionSvc.Sync(userID)
	if err != nil {
		if errors.Is(err, service.ErrNotionNotConnected) {
			util.BadRequest(c, "Notion integration not connected")
			return
		}
		if errors.Is(err, service.ErrNotionDatabaseNotSet) {
			util.BadRequest(c, "Notion database not configured")
			return
		}
		util.InternalError(c, "sync failed")
		return
	}

	util.OK(c, syncResult)
}

// Disconnect DELETE /api/notion/disconnect
func (h *NotionHandler) Disconnect(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)

	if err := h.notionSvc.Disconnect(userID); err != nil {
		util.InternalError(c, "failed to disconnect Notion integration")
		return
	}

	util.OK(c, gin.H{"message": "Notion integration disconnected"})
}
