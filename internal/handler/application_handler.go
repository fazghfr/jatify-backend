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

	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	resp, err := h.svc.GetPage(userID, page, pageSize)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}

	util.OK(c, resp)
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
