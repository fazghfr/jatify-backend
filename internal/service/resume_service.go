package service

import (
	"job-tracker/internal/dto"
	"job-tracker/internal/entity"
	"job-tracker/internal/repository"

	"github.com/google/uuid"
)

type ResumeService interface {
	Create(userID int, req *dto.CreateResumeRequest) (*entity.Resume, error)
	GetAll(userID int) ([]entity.Resume, error)
	GetByID(userID, id int) (*entity.Resume, error)
	Update(userID, id int, req *dto.UpdateResumeRequest) (*entity.Resume, error)
	Delete(userID, id int) error
}

type resumeService struct {
	repo repository.ResumeRepository
}

func NewResumeService(repo repository.ResumeRepository) ResumeService {
	return &resumeService{repo: repo}
}

func (s *resumeService) Create(userID int, req *dto.CreateResumeRequest) (*entity.Resume, error) {
	resume := &entity.Resume{
		Filepath: req.Filepath,
		UserID:   userID,
		UUID:     uuid.New(),
	}
	if err := s.repo.Create(resume); err != nil {
		return nil, err
	}
	return resume, nil
}

func (s *resumeService) GetAll(userID int) ([]entity.Resume, error) {
	return s.repo.FindAllByUserID(userID)
}

func (s *resumeService) GetByID(userID, id int) (*entity.Resume, error) {
	resume, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if resume.UserID != userID {
		return nil, ErrForbidden
	}
	return resume, nil
}

func (s *resumeService) Update(userID, id int, req *dto.UpdateResumeRequest) (*entity.Resume, error) {
	resume, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if resume.UserID != userID {
		return nil, ErrForbidden
	}

	if req.Filepath != nil {
		resume.Filepath = *req.Filepath
	}

	if err := s.repo.Update(resume); err != nil {
		return nil, err
	}
	return resume, nil
}

func (s *resumeService) Delete(userID, id int) error {
	resume, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if resume.UserID != userID {
		return ErrForbidden
	}
	return s.repo.Delete(id)
}
