package dto

// UserID is intentionally omitted — it comes from the JWT context, not the request body.
type CreateApplicationRequest struct {
	ResumeID int    `json:"resume_id" binding:"required"`
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
