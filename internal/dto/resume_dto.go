package dto

import "job-tracker/internal/entity"

type UpdateResumeRequest struct {
	Filepath *string `json:"filepath"`
}

type PaginatedResumeResponse struct {
	Data       []entity.Resume		`json:"data"`
	Pagination Pagination           `json:"pagination"`
}
