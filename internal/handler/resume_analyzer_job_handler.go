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
	resumeID := c.Param("id")

	job, err := h.svc.AnalyzeResume(userID, resumeID)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			util.Forbidden(c, err.Error())
			return
		}
		util.InternalError(c, err.Error())
		return
	}

	util.Created(c, job)
}
