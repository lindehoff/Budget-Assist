package core

import "fmt"

// OperationError represents a generic operation error in the core package
type OperationError struct {
	Err       error
	Operation string
	Resource  string
}

func (e OperationError) Error() string {
	if e.Resource != "" {
		return fmt.Sprintf("%s operation failed for %q: %v", e.Operation, e.Resource, e.Err)
	}
	return fmt.Sprintf("%s operation failed: %v", e.Operation, e.Err)
}

// ValidationError represents validation failures
type ValidationError struct {
	Field   string
	Value   any
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field %q with value %v: %s", e.Field, e.Value, e.Message)
}

// Common errors
var (
	ErrInvalidInput     = fmt.Errorf("invalid input provided")
	ErrNotFound         = fmt.Errorf("resource not found")
	ErrAlreadyExists    = fmt.Errorf("resource already exists")
	ErrInvalidOperation = fmt.Errorf("invalid operation")
)

// Error constructors
func NewOperationError(operation string, err error) error {
	return OperationError{
		Operation: operation,
		Err:       err,
	}
}

func NewResourceOperationError(operation, resource string, err error) error {
	return OperationError{
		Operation: operation,
		Resource:  resource,
		Err:       err,
	}
}

func NewValidationError(field string, value any, message string) error {
	return ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}
