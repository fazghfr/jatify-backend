package service

import (
	"fmt"
	"job-tracker/internal/entity"
	"job-tracker/internal/repository"
)

type DLQService struct {
	DLQRepo repository.ResumeAnalysisDLQRepository
}

func NewDLQService(DLQRepo repository.ResumeAnalysisDLQRepository) *DLQService {
	return &DLQService{DLQRepo: DLQRepo}
}

func (s *DLQService) ListByUser(userid int) ([]entity.ResumeAnalysisDLQ, error) {
	return s.DLQRepo.FindAllByUserID(userid)
}

func (s *DLQService) Requeue(Rtype string, userid int, dlqUUID string) (error) {
	if Rtype != "single" && Rtype != "bulk" {
		return fmt.Errorf("invalid request type: %s", Rtype)
	}

	return s.DLQRepo.Requeue(Rtype, userid, dlqUUID)
}
