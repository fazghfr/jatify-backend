package service

import (
	"context"
	"job-tracker/internal/entity"
	"job-tracker/internal/openrouter"
	"job-tracker/internal/repository"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
)

type ResumeAnalyzerJobService interface {
	AnalyzeResume(userID int, resumeUUID string) (entity.ResumeAnalysisJob, error)
}

type resumeAnalyzerJobService struct {
	jobrepo    *repository.ResumeAnalyzerJobRepository
	resumerepo repository.ResumeRepository
	orClient   openrouter.ORClient
}

func NewResumeAnalyzerJobService(
	jobrepo *repository.ResumeAnalyzerJobRepository,
	resumerepo repository.ResumeRepository,
	orClient openrouter.ORClient,
) ResumeAnalyzerJobService {
	return &resumeAnalyzerJobService{jobrepo: jobrepo, resumerepo: resumerepo, orClient: orClient}
}

// Utils function on the service layer
func ExtractTextFromPDF(filepath string) (string, error) {
	file, reader, err := pdf.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var buf strings.Builder
	for i := 1; i <= reader.NumPage(); i++ {
		p := reader.Page(i)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			return "", err
		}
		buf.WriteString(text)
	}
	return buf.String(), nil
}

func (s *resumeAnalyzerJobService) AnalyzeResume(userID int, resumeUUID string) (entity.ResumeAnalysisJob, error) {
	// 1 : find resume by resume ID verified by the userID
	// 2 : create the job
	// 3 : update the job status as processing
	// 4 : pdf text extraction
	// 5 : system prompt
	// 6 : actual LLM call using openrouter client
	// 7 : update job as completed
	// 8 : return resume analysis job entity

	resume, err := s.resumerepo.FindByUUID(resumeUUID)
	if err != nil {
		return entity.ResumeAnalysisJob{}, err
	}
	if resume.UserID != userID {
		return entity.ResumeAnalysisJob{}, ErrForbidden
	}

	// job creation
	jobuuid := uuid.New()
	resumeid := resume.ID
	job := &entity.ResumeAnalysisJob{
		UUID:     jobuuid,
		ResumeID: resumeid,
		UserID:   userID,
		Status:   "processing",
	}
	if err := s.jobrepo.Create(job); err != nil {
		return entity.ResumeAnalysisJob{}, err
	}

	// text extraction
	text, err := ExtractTextFromPDF(resume.Filepath)
	if err != nil {
		return entity.ResumeAnalysisJob{}, err
	}

	// LLM call (60-second timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	resultJSON, err := s.orClient.AnalyzeResume(ctx, text)
	if err != nil {
		return entity.ResumeAnalysisJob{}, err
	}
	job.Status = "done"
	job.ResultJSON = &resultJSON
	if err := s.jobrepo.Update(job); err != nil {
		return entity.ResumeAnalysisJob{}, err
	}

	return *job, nil
}
