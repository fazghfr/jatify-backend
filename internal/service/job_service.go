package service

import (
	"job-tracker/internal/dto"
	"job-tracker/internal/entity"
	"job-tracker/internal/repository"

	"github.com/google/uuid"
)

type JobService interface {
	Create(userID int, req *dto.CreateJobRequest) (*entity.Job, error)
	GetAll(userID int) ([]entity.Job, error)
	GetByID(userID, id int) (*entity.Job, error)
	Update(userID, id int, req *dto.UpdateJobRequest) (*entity.Job, error)
	Delete(userID, id int) error
}

type jobService struct {
	repo repository.JobRepository
}

func NewJobService(repo repository.JobRepository) JobService {
	return &jobService{repo: repo}
}

func (s *jobService) Create(userID int, req *dto.CreateJobRequest) (*entity.Job, error) {
	job := &entity.Job{
		Company:     req.Company,
		Position:    req.Position,
		Description: req.Description,
		UserID:      userID,
		UUID:        uuid.New(),
	}
	if err := s.repo.Create(job); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *jobService) GetAll(userID int) ([]entity.Job, error) {
	return s.repo.FindAllByUserID(userID)
}

func (s *jobService) GetByID(userID, id int) (*entity.Job, error) {
	job, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if job.UserID != userID {
		return nil, ErrForbidden
	}
	return job, nil
}

func (s *jobService) Update(userID, id int, req *dto.UpdateJobRequest) (*entity.Job, error) {
	job, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if job.UserID != userID {
		return nil, ErrForbidden
	}

	if req.Company != nil {
		job.Company = *req.Company
	}
	if req.Position != nil {
		job.Position = *req.Position
	}
	if req.Description != nil {
		job.Description = *req.Description
	}

	if err := s.repo.Update(job); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *jobService) Delete(userID, id int) error {
	job, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if job.UserID != userID {
		return ErrForbidden
	}
	return s.repo.Delete(id)
}
