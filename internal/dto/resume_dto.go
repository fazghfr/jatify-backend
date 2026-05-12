package dto

import "job-tracker/internal/entity"

type UpdateResumeRequest struct {
	Name     *string `json:"name"`
	Filepath *string `json:"filepath"`
}

type PaginatedResumeResponse struct {
	Data       []entity.Resume		`json:"data"`
	Pagination Pagination           `json:"pagination"`
}
