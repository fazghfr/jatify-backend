package handler

import (
	"errors"
	"strconv"

	"job-tracker/internal/dto"
	"job-tracker/internal/middleware"
	"job-tracker/internal/service"
	"job-tracker/internal/util"

	"github.com/gin-gonic/gin"
)

type JobHandler struct {
	svc service.JobService
}

func NewJobHandler(svc service.JobService) *JobHandler {
	return &JobHandler{svc: svc}
}

// POST /api/jobs
func (h *JobHandler) Create(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)

	var req dto.CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, err.Error())
		return
	}

	job, err := h.svc.Create(userID, &req)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}

	util.Created(c, job)
}

// GET /api/jobs
func (h *JobHandler) GetAll(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)

	jobs, err := h.svc.GetAll(userID)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}

	util.OK(c, jobs)
}

// GET /api/jobs/:id
func (h *JobHandler) GetByID(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.BadRequest(c, "invalid id")
		return
	}

	job, err := h.svc.GetByID(userID, id)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			util.Forbidden(c, "access denied")
			return
		}
		util.NotFound(c, "job not found")
		return
	}

	util.OK(c, job)
}

// PUT /api/jobs/:id
func (h *JobHandler) Update(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.BadRequest(c, "invalid id")
		return
	}

	var req dto.UpdateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, err.Error())
		return
	}

	job, err := h.svc.Update(userID, id, &req)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			util.Forbidden(c, "access denied")
			return
		}
		util.NotFound(c, "job not found")
		return
	}

	util.OK(c, job)
}

// DELETE /api/jobs/:id
func (h *JobHandler) Delete(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.BadRequest(c, "invalid id")
		return
	}

	if err := h.svc.Delete(userID, id); err != nil {
		if errors.Is(err, service.ErrForbidden) {
			util.Forbidden(c, "access denied")
			return
		}
		util.NotFound(c, "job not found")
		return
	}

	util.OK(c, gin.H{"deleted": true})
}
