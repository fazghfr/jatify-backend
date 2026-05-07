package dto

import "job-tracker/internal/entity"

type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type PaginatedApplicationsResponse struct {
	Data       []entity.Application `json:"data"`
	Pagination Pagination           `json:"pagination"`
}

// UserID is intentionally omitted — it comes from the JWT context, not the request body.
type CreateApplicationRequest struct {
	ResumeID *int   `json:"resume_id"`
	JobID    int    `json:"job_id" binding:"required"`
	Text     string `json:"text" binding:"required"`
	StatusID int    `json:"status_id" binding:"required"`
}

type UpdateApplicationRequest struct {
	ResumeID *int    `json:"resume_id"`
	JobID    *int    `json:"job_id"`
	Text     *string `json:"text"`
	StatusID *int    `json:"status_id"`
}
