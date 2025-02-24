package db

import (
	"gorm.io/gorm"
)

// runMigrations performs all necessary database migrations
func runMigrations(db *gorm.DB) error {
	// Auto-migrate all models
	return db.AutoMigrate(
		&CategoryType{},
		&Category{},
		&Subcategory{},
		&Transaction{},
		&Tag{},
		&Budget{},
		&Report{},
		&Translation{},
		&Prompt{},
	)
}
