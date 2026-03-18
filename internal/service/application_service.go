package service

import (
	"errors"

	"job-tracker/internal/dto"
	"job-tracker/internal/entity"
	"job-tracker/internal/repository"

	"github.com/google/uuid"
)

var ErrForbidden = errors.New("access denied")

type ApplicationService interface {
	Create(userID int, req *dto.CreateApplicationRequest) (*entity.Application, error)
	GetAll(userID int) ([]entity.Application, error)
	GetByID(userID, id int) (*entity.Application, error)
	Update(userID, id int, req *dto.UpdateApplicationRequest) (*entity.Application, error)
	Delete(userID, id int) error
}

type applicationService struct {
	appRepo     repository.ApplicationRepository
	historyRepo repository.StatusHistoryRepository
}

func NewApplicationService(
	appRepo repository.ApplicationRepository,
	historyRepo repository.StatusHistoryRepository,
) ApplicationService {
	return &applicationService{appRepo: appRepo, historyRepo: historyRepo}
}

func (s *applicationService) Create(userID int, req *dto.CreateApplicationRequest) (*entity.Application, error) {
	app := &entity.Application{
		ResumeID: req.ResumeID,
		JobID:    req.JobID,
		Text:     req.Text,
		StatusID: req.StatusID,
		UserID:   userID,
		UUID:     uuid.New(),
	}

	if err := s.appRepo.Create(app); err != nil {
		return nil, err
	}

	_ = s.historyRepo.Create(&entity.StatusHistory{
		ApplicationID: app.ID,
		StatusID:      app.StatusID,
	})

	return app, nil
}

func (s *applicationService) GetAll(userID int) ([]entity.Application, error) {
	return s.appRepo.FindAllByUserID(userID)
}

func (s *applicationService) GetByID(userID, id int) (*entity.Application, error) {
	app, err := s.appRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if app.UserID != userID {
		return nil, ErrForbidden
	}
	return app, nil
}

func (s *applicationService) Update(userID, id int, req *dto.UpdateApplicationRequest) (*entity.Application, error) {
	app, err := s.appRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if app.UserID != userID {
		return nil, ErrForbidden
	}

	prevStatusID := app.StatusID

	if req.ResumeID != nil {
		app.ResumeID = req.ResumeID
	}
	if req.JobID != nil {
		app.JobID = *req.JobID
	}
	if req.Text != nil {
		app.Text = *req.Text
	}
	if req.StatusID != nil {
		app.StatusID = *req.StatusID
	}

	if err := s.appRepo.Update(app); err != nil {
		return nil, err
	}

	if req.StatusID != nil && *req.StatusID != prevStatusID {
		_ = s.historyRepo.Create(&entity.StatusHistory{
			ApplicationID: app.ID,
			StatusID:      app.StatusID,
		})
	}

	return app, nil
}

func (s *applicationService) Delete(userID, id int) error {
	app, err := s.appRepo.FindByID(id)
	if err != nil {
		return err
	}
	if app.UserID != userID {
		return ErrForbidden
	}
	return s.appRepo.Delete(id)
}
