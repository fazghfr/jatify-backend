package handler

import (
	"errors"

	"job-tracker/internal/dto"
	"job-tracker/internal/middleware"
	"job-tracker/internal/service"
	"job-tracker/internal/util"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, err.Error())
		return
	}

	user, err := h.svc.Register(&req)
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			util.Conflict(c, err.Error())
			return
		}
		util.InternalError(c, err.Error())
		return
	}

	util.Created(c, dto.ToAuthUserResponse(user))
}

// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, err.Error())
		return
	}

	token, user, err := h.svc.Login(&req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCreds) {
			util.Unauthorized(c, err.Error())
			return
		}
		util.InternalError(c, err.Error())
		return
	}

	util.OK(c, dto.LoginResponse{
		Token: token,
		User:  dto.ToAuthUserResponse(user),
	})
}

// POST /api/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	util.OK(c, gin.H{"message": "logged out successfully"})
}

// GET /api/user/profile
func (h *AuthHandler) Profile(c *gin.Context) {
	userID := c.GetInt(middleware.UserIDKey)

	user, err := h.svc.GetProfile(userID)
	if err != nil {
		util.NotFound(c, "user not found")
		return
	}

	util.OK(c, dto.ToAuthUserResponse(user))
}
