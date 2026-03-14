package database

import (
	"job-tracker/internal/entity"

	"gorm.io/gorm"
)

var defaultStatuses = []entity.Status{
	{ID: 1, Text: "Applied"},
	{ID: 2, Text: "HR Interview"},
	{ID: 3, Text: "User Interview"},
	{ID: 4, Text: "Ghosted"},
	{ID: 5, Text: "Offer"},
	{ID: 6, Text: "Rejected"},
	{ID: 7, Text: "Offer Accepted"},
	{ID: 8, Text: "Offer Turned Down"},
}

func Seed(db *gorm.DB) error {
	for _, s := range defaultStatuses {
		if err := db.FirstOrCreate(&s, entity.Status{ID: s.ID}).Error; err != nil {
			return err
		}
	}
	return nil
}
