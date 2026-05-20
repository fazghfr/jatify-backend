package handler

import (
	"errors"
	"strings"

	"job-tracker/internal/middleware"
	"job-tracker/internal/service"
	"job-tracker/internal/util"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ApplicationAnalysisDLQHandler struct {
	svc *service.ApplicationAnalysisDLQService
}

func NewApplicationAnalysisDLQHandler(svc *service.ApplicationAnalysisDLQService) *ApplicationAnalysisDLQHandler {
	return &ApplicationAnalysisDLQHandler{svc: svc}
}

// GET /api/app-analysis-dlq
func (h *ApplicationAnalysisDLQHandler) List(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)
	dlqs, err := h.svc.ListByUser(userID)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.OK(c, dlqs)
}

// POST /api/app-analysis-dlq/:uuid/requeue
func (h *ApplicationAnalysisDLQHandler) RequeueSingle(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)
	dlqUUID := c.Param("uuid")

	if err := h.svc.Requeue("single", userID, dlqUUID); err != nil {
		h.mapRequeueError(c, err)
		return
	}

	util.Accepted(c, gin.H{"requeued": true, "uuid": dlqUUID})
}

// POST /api/app-analysis-dlq/requeue
func (h *ApplicationAnalysisDLQHandler) RequeueBulk(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)

	if err := h.svc.Requeue("bulk", userID, ""); err != nil {
		h.mapRequeueError(c, err)
		return
	}

	util.Accepted(c, gin.H{"requeued": true, "scope": "bulk"})
}

// DELETE /api/app-analysis-dlq/:uuid
func (h *ApplicationAnalysisDLQHandler) Delete(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)
	dlqUUID := c.Param("uuid")

	if err := h.svc.Delete(userID, dlqUUID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			util.NotFound(c, "dlq entry not found")
			return
		}
		util.InternalError(c, err.Error())
		return
	}

	util.NoContent(c)
}

func (h *ApplicationAnalysisDLQHandler) mapRequeueError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		util.NotFound(c, "dlq entry not found")
		return
	}
	msg := err.Error()
	if strings.Contains(msg, "forbidden") {
		util.Forbidden(c, "access denied")
		return
	}
	if strings.Contains(msg, "invalid request type") {
		util.BadRequest(c, msg)
		return
	}
	util.InternalError(c, msg)
}
