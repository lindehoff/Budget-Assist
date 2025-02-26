package db

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config holds database configuration
type Config struct {
	DBPath string
	// Add a flag to control seeding
	SeedPredefined bool
}

// Initialize sets up the database connection and runs migrations
func Initialize(cfg *Config) (*gorm.DB, error) {
	// Get log level from global logger
	logLevel := logger.Error // Default to only errors
	if slog.Default().Handler().Enabled(context.Background(), slog.LevelDebug) {
		logLevel = logger.Info // Show SQL only in debug mode
	}

	// Ensure the directory exists
	dbDir := filepath.Dir(cfg.DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Configure GORM logger using global settings
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Open database connection
	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Seed predefined data if requested
	if cfg.SeedPredefined {
		// Use global logger context
		ctx := WithLogger(context.Background(), slog.Default())
		if err := SeedPredefinedCategories(ctx, db); err != nil {
			return nil, fmt.Errorf("failed to seed predefined categories: %w", err)
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
