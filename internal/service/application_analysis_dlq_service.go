package service

import (
	"fmt"

	"job-tracker/internal/entity"
	"job-tracker/internal/repository"
)

type ApplicationAnalysisDLQService struct {
	dlqRepo repository.ApplicationAnalysisDLQRepository
}

func NewApplicationAnalysisDLQService(dlqRepo repository.ApplicationAnalysisDLQRepository) *ApplicationAnalysisDLQService {
	return &ApplicationAnalysisDLQService{dlqRepo: dlqRepo}
}

func (s *ApplicationAnalysisDLQService) ListByUser(userID int) ([]entity.ApplicationAnalysisDLQ, error) {
	return s.dlqRepo.FindAllByUserID(userID)
}

func (s *ApplicationAnalysisDLQService) Requeue(rtype string, userID int, dlqUUID string) error {
	if rtype != "single" && rtype != "bulk" {
		return fmt.Errorf("invalid request type: %s", rtype)
	}
	return s.dlqRepo.Requeue(rtype, userID, dlqUUID)
}

func (s *ApplicationAnalysisDLQService) Delete(userID int, dlqUUID string) error {
	return s.dlqRepo.Delete(userID, dlqUUID)
}
