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

type ApplicationHandler struct {
	svc service.ApplicationService
}

func NewApplicationHandler(svc service.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{svc: svc}
}

// POST /api/applications
func (h *ApplicationHandler) Create(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)

	var req dto.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, err.Error())
		return
	}

	app, err := h.svc.Create(userID, &req)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}

	util.Created(c, app)
}

// GET /api/applications
func (h *ApplicationHandler) GetAll(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)

	apps, err := h.svc.GetAll(userID)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}

	util.OK(c, apps)
}

// GET /api/applications/:id
func (h *ApplicationHandler) GetByID(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.BadRequest(c, "invalid id")
		return
	}

	app, err := h.svc.GetByID(userID, id)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			util.Forbidden(c, "access denied")
			return
		}
		util.NotFound(c, "application not found")
		return
	}

	util.OK(c, app)
}

// PUT /api/applications/:id
func (h *ApplicationHandler) Update(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		util.BadRequest(c, "invalid id")
		return
	}

	var req dto.UpdateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, err.Error())
		return
	}

	app, err := h.svc.Update(userID, id, &req)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			util.Forbidden(c, "access denied")
			return
		}
		util.NotFound(c, "application not found")
		return
	}

	util.OK(c, app)
}

// DELETE /api/applications/:id
func (h *ApplicationHandler) Delete(c *gin.Context) {
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
		util.NotFound(c, "application not found")
		return
	}

	util.OK(c, gin.H{"deleted": true})
}
