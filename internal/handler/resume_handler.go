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

type ResumeHandler struct {
	svc service.ResumeService
}

func NewResumeHandler(svc service.ResumeService) *ResumeHandler {
	return &ResumeHandler{svc: svc}
}

// POST /api/resumes
func (h *ResumeHandler) Create(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)

	var req dto.CreateResumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, err.Error())
		return
	}

	resume, err := h.svc.Create(userID, &req)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}

	util.Created(c, resume)
}

// GET /api/resumes
func (h *ResumeHandler) GetAll(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)

	resumes, err := h.svc.GetAll(userID)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}

	util.OK(c, resumes)
}

// GET /api/resumes/:id
func (h *ResumeHandler) GetByID(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.BadRequest(c, "invalid id")
		return
	}

	resume, err := h.svc.GetByID(userID, id)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			util.Forbidden(c, "access denied")
			return
		}
		util.NotFound(c, "resume not found")
		return
	}

	util.OK(c, resume)
}

// PUT /api/resumes/:id
func (h *ResumeHandler) Update(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.BadRequest(c, "invalid id")
		return
	}

	var req dto.UpdateResumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, err.Error())
		return
	}

	resume, err := h.svc.Update(userID, id, &req)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			util.Forbidden(c, "access denied")
			return
		}
		util.NotFound(c, "resume not found")
		return
	}

	util.OK(c, resume)
}

// DELETE /api/resumes/:id
func (h *ResumeHandler) Delete(c *gin.Context) {
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
		util.NotFound(c, "resume not found")
		return
	}

	util.OK(c, gin.H{"deleted": true})
}
