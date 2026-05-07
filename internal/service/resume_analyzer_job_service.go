package service

import (
	"context"
	"job-tracker/internal/entity"
	"job-tracker/internal/repository"
	"strings"

	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
)

type ResumeAnalyzerJobService interface {
	Enqueue(ctx context.Context, userID int, resumeUUID string) (entity.ResumeAnalysisJob, error)
	GetResult(ctx context.Context, userID int, jobUUID string) (entity.ResumeAnalysisJob, error)
	ListByResumeID(ctx context.Context, userID int, resumeUUID string) ([]entity.ResumeAnalysisJob, error)
}

type resumeAnalyzerJobService struct {
	jobrepo    repository.ResumeAnalysisJobRepository
	resumerepo repository.ResumeRepository
	jobCh      chan int
}

func NewResumeAnalyzerJobService(
	jobrepo repository.ResumeAnalysisJobRepository,
	resumerepo repository.ResumeRepository,
	jobCh chan int,
) ResumeAnalyzerJobService {
	return &resumeAnalyzerJobService{jobrepo: jobrepo, resumerepo: resumerepo, jobCh: jobCh}
}

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

func (s *resumeAnalyzerJobService) Enqueue(_ context.Context, userID int, resumeUUID string) (entity.ResumeAnalysisJob, error) {
	resume, err := s.resumerepo.FindByUUID(resumeUUID)
	if err != nil {
		return entity.ResumeAnalysisJob{}, err
	}
	if resume.UserID != userID {
		return entity.ResumeAnalysisJob{}, ErrForbidden
	}

	job := &entity.ResumeAnalysisJob{
		UUID:     uuid.New(),
		ResumeID: resume.ID,
		UserID:   userID,
		Status:   "pending",
	}
	if err := s.jobrepo.Create(job); err != nil {
		return entity.ResumeAnalysisJob{}, err
	}

	// non-blocking: if channel is full, worker picks it up via DB poll
	select {
	case s.jobCh <- job.ID:
	default:
	}

	return *job, nil
}

func (s *resumeAnalyzerJobService) GetResult(_ context.Context, userID int, jobUUID string) (entity.ResumeAnalysisJob, error) {
	job, err := s.jobrepo.FindByUUID(jobUUID)
	if err != nil {
		return entity.ResumeAnalysisJob{}, err
	}
	if job.UserID != userID {
		return entity.ResumeAnalysisJob{}, ErrForbidden
	}
	return *job, nil
}

func (s *resumeAnalyzerJobService) ListByResumeID(_ context.Context, userID int, resumeUUID string) ([]entity.ResumeAnalysisJob, error) {
	resume, err := s.resumerepo.FindByUUID(resumeUUID)
	if err != nil {
		return nil, err
	}
	if resume.UserID != userID {
		return nil, ErrForbidden
	}
	return s.jobrepo.FindAllByResumeIDAndUserID(resume.ID, userID)
}
