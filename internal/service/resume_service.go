package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"job-tracker/internal/dto"
	"job-tracker/internal/entity"
	"job-tracker/internal/repository"

	"github.com/google/uuid"
)

type ResumeService interface {
	Create(userID int, file io.ReadSeeker, originalName string) (*entity.Resume, error)
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

func (s *resumeService) Create(userID int, file io.ReadSeeker, originalName string) (*entity.Resume, error) {
	id := uuid.New()
	ext := filepath.Ext(originalName)
	dir := fmt.Sprintf("uploads/%d", userID)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	savePath := fmt.Sprintf("%s/%s%s", dir, id.String(), ext)

	dst, err := os.Create(savePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(savePath)
		return nil, err
	}

	resume := &entity.Resume{
		Filepath: savePath,
		UserID:   userID,
		UUID:     id,
	}
	if err := s.repo.Create(resume); err != nil {
		os.Remove(savePath)
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
