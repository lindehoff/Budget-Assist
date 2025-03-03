package ai

import (
	"fmt"
)

// Common errors
var (
	ErrEmptyDocument    = fmt.Errorf("empty document content")
	ErrNoChoices        = fmt.Errorf("no choices in OpenAI response")
	ErrEmptyContent     = fmt.Errorf("empty content in OpenAI response")
	ErrTemplateNotFound = fmt.Errorf("template not found")
	ErrInvalidOperation = fmt.Errorf("invalid operation")
)

// OperationError represents an error that occurred during an operation
type OperationError struct {
	Err       error
	Operation string
	Resource  string
}

func (e *OperationError) Error() string {
	if e.Resource != "" {
		return fmt.Sprintf("%s operation failed for %q: %v", e.Operation, e.Resource, e.Err)
	}
	return fmt.Sprintf("%s operation failed: %v", e.Operation, e.Err)
}

// OpenAIError represents an error from the OpenAI API
type OpenAIError struct {
	Operation  string
	Message    string
	StatusCode int
}

func (e *OpenAIError) Error() string {
	return fmt.Sprintf("OpenAI API error during %s operation (status %d): %s", e.Operation, e.StatusCode, e.Message)
}

// RateLimitError represents a rate limit error from the OpenAI API
type RateLimitError struct {
	Message    string
	StatusCode int
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded (status %d): %s", e.StatusCode, e.Message)
}
