package handler

import (
	"errors"
	"job-tracker/internal/middleware"
	"job-tracker/internal/service"
	"job-tracker/internal/util"

	"github.com/gin-gonic/gin"
)

type ResumeAnalyzerJobHandler struct {
	svc service.ResumeAnalyzerJobService
}

func NewResumeAnalyzerJobHandler(svc service.ResumeAnalyzerJobService) *ResumeAnalyzerJobHandler {
	return &ResumeAnalyzerJobHandler{svc: svc}
}

func (h *ResumeAnalyzerJobHandler) Analyze(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)
	resumeUUID := c.Param("id")

	job, err := h.svc.Enqueue(c.Request.Context(), userID, resumeUUID)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			util.Forbidden(c, err.Error())
			return
		}
		util.InternalError(c, err.Error())
		return
	}

	util.Accepted(c, job)
}

func (h *ResumeAnalyzerJobHandler) GetResult(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)
	jobUUID := c.Param("jobid")

	job, err := h.svc.GetResult(c.Request.Context(), userID, jobUUID)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			util.Forbidden(c, err.Error())
			return
		}
		util.NotFound(c, err.Error())
		return
	}

	util.OK(c, job)
}

func (h *ResumeAnalyzerJobHandler) ListAnalyses(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)
	resumeUUID := c.Param("id")

	jobs, err := h.svc.ListByResumeID(c.Request.Context(), userID, resumeUUID)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			util.Forbidden(c, err.Error())
			return
		}
		util.NotFound(c, err.Error())
		return
	}

	util.OK(c, jobs)
}
