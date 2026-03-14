package handler

import (
	"job-tracker/internal/repository"
	"job-tracker/internal/util"

	"github.com/gin-gonic/gin"
)

type StatusHandler struct {
	repo repository.StatusRepository
}

func NewStatusHandler(repo repository.StatusRepository) *StatusHandler {
	return &StatusHandler{repo: repo}
}

// GET /api/statuses
func (h *StatusHandler) GetAll(c *gin.Context) {
	statuses, err := h.repo.FindAll()
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.OK(c, statuses)
}
