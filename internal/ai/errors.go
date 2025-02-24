package ai

import "fmt"

// Error represents a base error type for the AI package
type Error struct {
	Err     error
	Code    string
	Message string
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// StatusError represents an error with an HTTP status code
type StatusError struct {
	Err        error
	Code       string
	Message    string
	StatusCode int
}

func (e *StatusError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s (status %d): %s: %v", e.Code, e.StatusCode, e.Message, e.Err)
	}
	return fmt.Sprintf("%s (status %d): %s", e.Code, e.StatusCode, e.Message)
}

// CategoryError represents an error related to category operations
type CategoryError struct {
	Err      error
	Category string
	Message  string
}

func (e *CategoryError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("category %q error: %s: %v", e.Category, e.Message, e.Err)
	}
	return fmt.Sprintf("category %q error: %s", e.Category, e.Message)
}

// Common errors
var (
	ErrLowConfidence    = fmt.Errorf("prediction confidence below threshold")
	ErrInvalidModel     = fmt.Errorf("invalid model specified")
	ErrModelUnavailable = fmt.Errorf("AI model currently unavailable")
	ErrInvalidInput     = fmt.Errorf("invalid input for AI processing")
)

// NewError creates a new Error instance with the given code and message
func NewError(code, message string, err error) *Error {
	return &Error{
		Err:     err,
		Code:    code,
		Message: message,
	}
}

// NewStatusError creates a new StatusError instance with the given code, message, and status code
func NewStatusError(code string, statusCode int, message string, err error) *StatusError {
	return &StatusError{
		Err:        err,
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// NewCategoryError creates a new CategoryError instance with the given category and message
func NewCategoryError(category, message string, err error) *CategoryError {
	return &CategoryError{
		Err:      err,
		Category: category,
		Message:  message,
	}
}
