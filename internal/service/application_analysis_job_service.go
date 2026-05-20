package service

import (
	"context"
	"fmt"

	"job-tracker/internal/entity"
	"job-tracker/internal/repository"

	"github.com/google/uuid"
)

var ErrActiveJobExists = fmt.Errorf("an active analysis job already exists for this application")

type AnalysisEnqueuer interface {
	AutoEnqueue(ctx context.Context, userID int, applicationID int) error
}

type ApplicationAnalysisJobService interface {
	Enqueue(ctx context.Context, userID int, applicationUUID string) (*entity.ApplicationAnalysisJob, error)
	GetLatest(ctx context.Context, userID int, applicationUUID string) (*entity.ApplicationAnalysisJob, error)
	GetHistory(ctx context.Context, userID int, applicationUUID string) ([]entity.ApplicationAnalysisJob, error)
	AutoEnqueue(ctx context.Context, userID int, applicationID int) error
}

type applicationAnalysisJobService struct {
	jobRepo repository.ApplicationAnalysisJobRepository
	appRepo repository.ApplicationRepository
	jobCh   chan int
}

func NewApplicationAnalysisJobService(
	jobRepo repository.ApplicationAnalysisJobRepository,
	appRepo repository.ApplicationRepository,
	jobCh chan int,
) ApplicationAnalysisJobService {
	return &applicationAnalysisJobService{jobRepo: jobRepo, appRepo: appRepo, jobCh: jobCh}
}

func (s *applicationAnalysisJobService) Enqueue(_ context.Context, userID int, applicationUUID string) (*entity.ApplicationAnalysisJob, error) {
	app, err := s.appRepo.FindByUUID(applicationUUID)
	if err != nil {
		return nil, err
	}
	if app.UserID != userID {
		return nil, ErrForbidden
	}

	active, err := s.jobRepo.HasActiveJob(app.ID)
	if err != nil {
		return nil, err
	}
	if active {
		return nil, ErrActiveJobExists
	}

	job := &entity.ApplicationAnalysisJob{
		UUID:          uuid.New(),
		ApplicationID: app.ID,
		UserID:        userID,
		Status:        "pending",
	}
	if err := s.jobRepo.Create(job); err != nil {
		return nil, err
	}

	select {
	case s.jobCh <- job.ID:
	default:
	}

	return job, nil
}

func (s *applicationAnalysisJobService) GetLatest(_ context.Context, userID int, applicationUUID string) (*entity.ApplicationAnalysisJob, error) {
	app, err := s.appRepo.FindByUUID(applicationUUID)
	if err != nil {
		return nil, err
	}
	if app.UserID != userID {
		return nil, ErrForbidden
	}

	return s.jobRepo.FindLatestByApplicationID(app.ID)
}

func (s *applicationAnalysisJobService) GetHistory(_ context.Context, userID int, applicationUUID string) ([]entity.ApplicationAnalysisJob, error) {
	app, err := s.appRepo.FindByUUID(applicationUUID)
	if err != nil {
		return nil, err
	}
	if app.UserID != userID {
		return nil, ErrForbidden
	}

	return s.jobRepo.FindAllByApplicationID(app.ID)
}

func (s *applicationAnalysisJobService) AutoEnqueue(ctx context.Context, userID int, applicationID int) error {
	app, err := s.appRepo.FindByID(applicationID)
	if err != nil {
		return err
	}
	if app.ResumeID == nil {
		return nil
	}
	if app.Job == nil || app.Job.Description == "" {
		return nil
	}

	active, err := s.jobRepo.HasActiveJob(applicationID)
	if err != nil {
		return err
	}
	if active {
		return nil
	}

	job := &entity.ApplicationAnalysisJob{
		UUID:          uuid.New(),
		ApplicationID: applicationID,
		UserID:        userID,
		Status:        "pending",
	}
	if err := s.jobRepo.Create(job); err != nil {
		return err
	}

	select {
	case s.jobCh <- job.ID:
	default:
	}

	return nil
}
