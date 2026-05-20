package handler

import (
	"errors"

	"job-tracker/internal/middleware"
	"job-tracker/internal/service"
	"job-tracker/internal/util"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ApplicationAnalysisHandler struct {
	svc service.ApplicationAnalysisJobService
}

func NewApplicationAnalysisHandler(svc service.ApplicationAnalysisJobService) *ApplicationAnalysisHandler {
	return &ApplicationAnalysisHandler{svc: svc}
}

// POST /api/applications/:uuid/analysis
func (h *ApplicationAnalysisHandler) Enqueue(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)
	appUUID := c.Param("id")

	job, err := h.svc.Enqueue(c.Request.Context(), userID, appUUID)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			util.Forbidden(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrActiveJobExists) {
			util.Conflict(c, err.Error())
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			util.NotFound(c, "application not found")
			return
		}
		util.InternalError(c, err.Error())
		return
	}

	util.Accepted(c, gin.H{"job_uuid": job.UUID, "status": job.Status})
}

// GET /api/applications/:uuid/analysis
func (h *ApplicationAnalysisHandler) GetLatest(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)
	appUUID := c.Param("id")

	job, err := h.svc.GetLatest(c.Request.Context(), userID, appUUID)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			util.Forbidden(c, err.Error())
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			util.NotFound(c, "no analysis found")
			return
		}
		util.InternalError(c, err.Error())
		return
	}

	util.OK(c, job)
}

// GET /api/applications/:uuid/analysis/history
func (h *ApplicationAnalysisHandler) GetHistory(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)
	appUUID := c.Param("id")

	jobs, err := h.svc.GetHistory(c.Request.Context(), userID, appUUID)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			util.Forbidden(c, err.Error())
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			util.NotFound(c, "application not found")
			return
		}
		util.InternalError(c, err.Error())
		return
	}

	util.OK(c, jobs)
}
