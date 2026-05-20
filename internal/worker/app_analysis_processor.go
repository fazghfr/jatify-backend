package worker

import (
	"context"
	"fmt"

	"job-tracker/internal/entity"
	"job-tracker/internal/openrouter"
	"job-tracker/internal/repository"
	"job-tracker/internal/service"
)

type AppAnalysisDeps struct {
	AppRepo  repository.ApplicationRepository
	JobRepo  repository.ApplicationAnalysisJobRepository
	DLQRepo  repository.ApplicationAnalysisDLQRepository
	ORClient openrouter.ORClient
}

func ProcessAppAnalysisJob(ctx context.Context, job entity.ApplicationAnalysisJob, deps AppAnalysisDeps) (string, error) {
	app, err := deps.AppRepo.FindByID(job.ApplicationID)
	if err != nil {
		return "", fmt.Errorf("load application: %w", err)
	}

	if app.ResumeID == nil {
		return "", fmt.Errorf("application has no resume attached")
	}
	if app.Job == nil || app.Job.Description == "" {
		return "", fmt.Errorf("application job has no description")
	}

	resumeText, err := service.ExtractTextFromPDF(app.Resume.Filepath)
	if err != nil {
		return "", fmt.Errorf("extract PDF: %w", err)
	}

	result, err := deps.ORClient.AnalyzeJobMatch(ctx, resumeText, app.Job.Description)
	if err != nil {
		return "", fmt.Errorf("openrouter: %w", err)
	}
	return result, nil
}
