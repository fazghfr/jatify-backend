package dto

import "job-tracker/internal/entity"

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Phone    string `json:"phone"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthUserResponse struct {
	ID       int    `json:"id"`
	UUID     string `json:"uuid"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}

type LoginResponse struct {
	Token string           `json:"token"`
	User  AuthUserResponse `json:"user"`
}

func ToAuthUserResponse(u *entity.User) AuthUserResponse {
	return AuthUserResponse{
		ID:       u.ID,
		UUID:     u.UUID.String(),
		Username: u.Username,
		Name:     u.Name,
		Email:    u.Email,
	}
}
