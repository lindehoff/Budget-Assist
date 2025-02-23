package db

import (
	"context"
	"log/slog"
)

// ContextKey is a type for context keys
type ContextKey string

const (
	// LoggerKey is the context key for the logger
	LoggerKey ContextKey = "logger"
)

// WithLogger adds a logger to the context
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, logger)
}

// LoggerFromContext retrieves the logger from the context
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(LoggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// DBOperation represents a database operation with context
type DBOperation struct {
	ctx    context.Context
	logger *slog.Logger
}

// NewDBOperation creates a new DBOperation with context
func NewDBOperation(ctx context.Context) *DBOperation {
	return &DBOperation{
		ctx:    ctx,
		logger: LoggerFromContext(ctx),
	}
}
