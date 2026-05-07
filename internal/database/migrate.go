package database

import (
	"job-tracker/internal/entity"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.Status{},
		&entity.User{},
		&entity.Job{},
		&entity.Resume{},
		&entity.Application{},
		&entity.StatusHistory{},
		&entity.NotionIntegration{},
		&entity.ResumeAnalysisJob{},
		&entity.ResumeAnalysisDLQ{},
	)
}
