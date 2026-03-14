package dto

type CreateResumeRequest struct {
	Filepath string `json:"filepath" binding:"required"`
}

type UpdateResumeRequest struct {
	Filepath *string `json:"filepath"`
}
