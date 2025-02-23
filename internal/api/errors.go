package api

import (
	"fmt"
	"net/http"
)

// APIError represents API-level errors
type APIError struct {
	Err        error
	Path       string
	Method     string
	StatusCode int
}

func (e APIError) Error() string {
	return fmt.Sprintf("%s %s request failed (status %d): %v",
		e.Method, e.Path, e.StatusCode, e.Err)
}

// ValidationError represents API request validation errors
type ValidationError struct {
	Field   string
	Value   any
	Rule    string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field %q with value %v (%s): %s",
		e.Field, e.Value, e.Rule, e.Message)
}

// Common errors
var (
	ErrInvalidRequest   = fmt.Errorf("invalid request")
	ErrUnauthorized     = fmt.Errorf("unauthorized access")
	ErrForbidden        = fmt.Errorf("forbidden access")
	ErrNotFound         = fmt.Errorf("resource not found")
	ErrMethodNotAllowed = fmt.Errorf("method not allowed")
	ErrConflict         = fmt.Errorf("resource conflict")
)

// Error constructors
func NewAPIError(method, path string, statusCode int, err error) error {
	return APIError{
		Method:     method,
		Path:       path,
		StatusCode: statusCode,
		Err:        err,
	}
}

func NewValidationError(field string, value any, rule, message string) error {
	return ValidationError{
		Field:   field,
		Value:   value,
		Rule:    rule,
		Message: message,
	}
}

// HTTP status code mapping
var (
	StatusErrors = map[int]error{
		http.StatusBadRequest:       ErrInvalidRequest,
		http.StatusUnauthorized:     ErrUnauthorized,
		http.StatusForbidden:        ErrForbidden,
		http.StatusNotFound:         ErrNotFound,
		http.StatusMethodNotAllowed: ErrMethodNotAllowed,
		http.StatusConflict:         ErrConflict,
	}
)
