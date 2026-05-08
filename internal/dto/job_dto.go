package dto

import "job-tracker/internal/entity"

type CreateJobRequest struct {
	Company     string `json:"company" binding:"required"`
	Position    string `json:"position" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type UpdateJobRequest struct {
	Company     *string `json:"company"`
	Position    *string `json:"position"`
	Description *string `json:"description"`
}

type PaginatedJobsResponse struct {
	Data       []entity.Job `json:"data"`
	Pagination Pagination   `json:"pagination"`
}
