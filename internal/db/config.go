package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"log/slog"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Config holds database configuration
type Config struct {
	DBPath string
	// Add flags to control importing of default data
	ImportDefaultCategories bool
	ImportDefaultPrompts    bool
}

// Initialize sets up the database connection and runs migrations
func Initialize(cfg *Config, logger *slog.Logger) (*gorm.DB, error) {
	// Configure GORM logger
	gormConfig := &gorm.Config{
		Logger: gormlogger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			gormlogger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  gormlogger.Error,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
	}

	// Ensure the directory exists
	dbDir := filepath.Dir(cfg.DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	db, err := gorm.Open(sqlite.Open(cfg.DBPath), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Use global logger context
	ctx := WithLogger(context.Background(), logger)

	// Import default categories if requested
	if cfg.ImportDefaultCategories {
		importer := &categoryImporter{db: db}
		if err := ImportDefaultCategories(ctx, db, importer); err != nil {
			return nil, fmt.Errorf("failed to import default categories: %w", err)
		}
	}

	// Import default prompts if requested
	if cfg.ImportDefaultPrompts {
		if err := ImportDefaultPrompts(ctx, db); err != nil {
			return nil, fmt.Errorf("failed to import default prompts: %w", err)
		}
	}

	return db, nil
}

// Close closes the database connection
func Close(db *gorm.DB) error {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("failed to get database instance: %w", err)
		}
		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("failed to close database connection: %w", err)
		}
	}
	return nil
}

// categoryImporter implements the CategoryImporter interface
type categoryImporter struct {
	db *gorm.DB
}

func (i *categoryImporter) CreateCategory(ctx context.Context, name, description string, typeID uint, translations map[string]TranslationData, subcategoryIDs []uint) (*Category, error) {
	category := &Category{
		Name:        name,
		Description: description,
		TypeID:      typeID,
		IsActive:    true,
	}

	if err := i.db.WithContext(ctx).Create(category).Error; err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	// Create translations
	for lang, trans := range translations {
		translation := &Translation{
			EntityID:     category.ID,
			EntityType:   string(EntityTypeCategory),
			LanguageCode: lang,
			Name:         trans.Name,
			Description:  trans.Description,
		}
		if err := i.db.WithContext(ctx).Create(translation).Error; err != nil {
			return nil, fmt.Errorf("failed to create translation: %w", err)
		}
	}

	// Link subcategories
	for _, subcatID := range subcategoryIDs {
		link := &CategorySubcategory{
			CategoryID:    category.ID,
			SubcategoryID: subcatID,
			IsActive:      true,
		}
		if err := i.db.WithContext(ctx).Create(link).Error; err != nil {
			return nil, fmt.Errorf("failed to link subcategory: %w", err)
		}
	}

	return category, nil
}

func (i *categoryImporter) CreateSubcategory(ctx context.Context, name, description string, isSystem bool, translations map[string]TranslationData) (*Subcategory, error) {
	subcategory := &Subcategory{
		Name:        name,
		Description: description,
		IsSystem:    isSystem,
		IsActive:    true,
	}

	if err := i.db.WithContext(ctx).Create(subcategory).Error; err != nil {
		return nil, fmt.Errorf("failed to create subcategory: %w", err)
	}

	// Create translations
	for lang, trans := range translations {
		translation := &Translation{
			EntityID:     subcategory.ID,
			EntityType:   string(EntityTypeSubcategory),
			LanguageCode: lang,
			Name:         trans.Name,
			Description:  trans.Description,
		}
		if err := i.db.WithContext(ctx).Create(translation).Error; err != nil {
			return nil, fmt.Errorf("failed to create translation: %w", err)
		}
	}

	return subcategory, nil
}

func (i *categoryImporter) CreateCategoryType(ctx context.Context, categoryType *CategoryType) error {
	return i.db.WithContext(ctx).Create(categoryType).Error
}

func (i *categoryImporter) CreateTranslation(ctx context.Context, translation *Translation) error {
	return i.db.WithContext(ctx).Create(translation).Error
}
