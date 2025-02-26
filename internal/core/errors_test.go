package core

import (
	"errors"
	"fmt"
	"testing"
)

func TestOperationError_Error(t *testing.T) {
	tests := []struct {
		name       string
		opError    OperationError
		wantString string
	}{
		{
			name: "Successfully_format_error_with_resource",
			opError: OperationError{
				Operation: "create",
				Resource:  "user",
				Err:       fmt.Errorf("database connection failed"),
			},
			wantString: `create operation failed for "user": database connection failed`,
		},
		{
			name: "Successfully_format_error_without_resource",
			opError: OperationError{
				Operation: "validate",
				Err:       fmt.Errorf("invalid data"),
			},
			wantString: "validate operation failed: invalid data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.opError.Error(); got != tt.wantString {
				t.Errorf("OperationError.Error() = %v, want %v", got, tt.wantString)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name          string
		validationErr ValidationError
		wantString    string
	}{
		{
			name: "Successfully_format_string_validation_error",
			validationErr: ValidationError{
				Field:   "username",
				Value:   "",
				Message: "cannot be empty",
			},
			wantString: `validation failed for field "username" with value : cannot be empty`,
		},
		{
			name: "Successfully_format_numeric_validation_error",
			validationErr: ValidationError{
				Field:   "age",
				Value:   -1,
				Message: "must be positive",
			},
			wantString: `validation failed for field "age" with value -1: must be positive`,
		},
		{
			name: "Successfully_format_complex_value_validation_error",
			validationErr: ValidationError{
				Field:   "settings",
				Value:   map[string]bool{"enabled": false},
				Message: "invalid configuration",
			},
			wantString: `validation failed for field "settings" with value map[enabled:false]: invalid configuration`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.validationErr.Error(); got != tt.wantString {
				t.Errorf("ValidationError.Error() = %v, want %v", got, tt.wantString)
			}
		})
	}
}

func TestNewOperationError(t *testing.T) {
	err := fmt.Errorf("test error")
	opErr := NewOperationError("create", err)

	var typedErr OperationError
	if !errors.As(opErr, &typedErr) {
		t.Error("NewOperationError() did not return an OperationError type")
	}

	if typedErr.Operation != "create" {
		t.Errorf("NewOperationError() Operation = %v, want %v", typedErr.Operation, "create")
	}
	if typedErr.Resource != "" {
		t.Errorf("NewOperationError() Resource = %v, want empty string", typedErr.Resource)
	}
	if typedErr.Err != err {
		t.Errorf("NewOperationError() Err = %v, want %v", typedErr.Err, err)
	}
}

func TestNewResourceOperationError(t *testing.T) {
	err := fmt.Errorf("test error")
	opErr := NewResourceOperationError("update", "user", err)

	var typedErr OperationError
	if !errors.As(opErr, &typedErr) {
		t.Error("NewResourceOperationError() did not return an OperationError type")
	}

	if typedErr.Operation != "update" {
		t.Errorf("NewResourceOperationError() Operation = %v, want %v", typedErr.Operation, "update")
	}
	if typedErr.Resource != "user" {
		t.Errorf("NewResourceOperationError() Resource = %v, want %v", typedErr.Resource, "user")
	}
	if typedErr.Err != err {
		t.Errorf("NewResourceOperationError() Err = %v, want %v", typedErr.Err, err)
	}
}

func TestNewValidationError(t *testing.T) {
	valErr := NewValidationError("email", "invalid@email", "invalid format")

	var typedErr ValidationError
	if !errors.As(valErr, &typedErr) {
		t.Error("NewValidationError() did not return a ValidationError type")
	}

	if typedErr.Field != "email" {
		t.Errorf("NewValidationError() Field = %v, want %v", typedErr.Field, "email")
	}
	if typedErr.Value != "invalid@email" {
		t.Errorf("NewValidationError() Value = %v, want %v", typedErr.Value, "invalid@email")
	}
	if typedErr.Message != "invalid format" {
		t.Errorf("NewValidationError() Message = %v, want %v", typedErr.Message, "invalid format")
	}
}

func TestCommonErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantText string
	}{
		{
			name:     "Successfully_return_InvalidInput_error",
			err:      ErrInvalidInput,
			wantText: "invalid input provided",
		},
		{
			name:     "Successfully_return_NotFound_error",
			err:      ErrNotFound,
			wantText: "resource not found",
		},
		{
			name:     "Successfully_return_AlreadyExists_error",
			err:      ErrAlreadyExists,
			wantText: "resource already exists",
		},
		{
			name:     "Successfully_return_InvalidOperation_error",
			err:      ErrInvalidOperation,
			wantText: "invalid operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.wantText {
				t.Errorf("Error() = %q, want %q", tt.err.Error(), tt.wantText)
			}
		})
	}
}
