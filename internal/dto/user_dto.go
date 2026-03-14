package dto

import (
	"job-tracker/internal/entity"

	"github.com/google/uuid"
)

type UserResumeDTO struct {
	UserUUID uuid.UUID       `json:"user_uuid"`
	UserID   int             `json:"user_id"`
	Resume   []entity.Resume `json:"resumes"`
}

type UserApplicationDTO struct {
	UserUUID    uuid.UUID            `json:"user_uuid"`
	UserID      int                  `json:"user_id"`
	Application []entity.Application `json:"applications"`
}

type UserJobDTO struct {
	UserUUID uuid.UUID    `json:"user_uuid"`
	UserID   int          `json:"user_id"`
	Job      []entity.Job `json:"jobs"`
}
