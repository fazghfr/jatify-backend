package worker

import (
	"context"
	"job-tracker/internal/entity"
	"job-tracker/internal/openrouter"
	"job-tracker/internal/repository"
	"job-tracker/internal/service"
)

type ProcessorDeps struct {
	ResumeRepo repository.ResumeRepository
	JobRepo    repository.ResumeAnalysisJobRepository
	ORClient   openrouter.ORClient
}

func ProcessJob(ctx context.Context, job entity.ResumeAnalysisJob, deps ProcessorDeps) (string, error) {
	// subprocess : resume parsing
	resume, err := deps.ResumeRepo.FindByID(job.ResumeID)
	if err != nil {
		return "", err
	}
	resumeText, err := service.ExtractTextFromPDF(resume.Filepath)
	if err != nil {
		return "", err
	}

	// subprocess : prompt injection and processing
	response, err := deps.ORClient.AnalyzeResume(ctx, resumeText)
	if err != nil {
		return "", err
	}
	return response, nil
}
