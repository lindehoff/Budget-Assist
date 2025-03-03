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
		if err := ImportDefaultCategories(ctx, db); err != nil {
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
