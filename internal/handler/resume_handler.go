package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"job-tracker/internal/middleware"
	"job-tracker/internal/service"
	"job-tracker/internal/util"

	"github.com/gin-gonic/gin"
	"job-tracker/internal/dto"
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

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		util.BadRequest(c, "file is required")
		return
	}
	defer file.Close()

	if header.Size > 1*1024*1024 {
		util.BadRequest(c, "file too large (max 1 MB)")
		return
	}

	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	mime := http.DetectContentType(buf[:n])
	if mime != "application/pdf" {
		util.BadRequest(c, "only PDF files are accepted")
		return
	}
	file.Seek(0, io.SeekStart)

	resume, err := h.svc.Create(userID, file, header.Filename)
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
