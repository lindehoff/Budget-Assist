package api

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name       string
		apiError   APIError
		wantString string
	}{
		{
			name: "GET request error",
			apiError: APIError{
				Method:     "GET",
				Path:       "/api/v1/categories",
				StatusCode: http.StatusNotFound,
				Err:        fmt.Errorf("category not found"),
			},
			wantString: "GET /api/v1/categories request failed (status 404): category not found",
		},
		{
			name: "POST request error",
			apiError: APIError{
				Method:     "POST",
				Path:       "/api/v1/categories",
				StatusCode: http.StatusBadRequest,
				Err:        fmt.Errorf("invalid category data"),
			},
			wantString: "POST /api/v1/categories request failed (status 400): invalid category data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.apiError.Error(); got != tt.wantString {
				t.Errorf("APIError.Error() = %v, want %v", got, tt.wantString)
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
			name: "string field validation",
			validationErr: ValidationError{
				Field:   "name",
				Value:   "",
				Rule:    "required",
				Message: "field cannot be empty",
			},
			wantString: `validation failed for field "name" with value  (required): field cannot be empty`,
		},
		{
			name: "numeric field validation",
			validationErr: ValidationError{
				Field:   "age",
				Value:   -1,
				Rule:    "min",
				Message: "must be positive",
			},
			wantString: `validation failed for field "age" with value -1 (min): must be positive`,
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

func TestNewAPIError(t *testing.T) {
	err := fmt.Errorf("test error")
	apiErr := NewAPIError("GET", "/test", http.StatusBadRequest, err)

	var typedErr APIError
	if !errors.As(apiErr, &typedErr) {
		t.Error("NewAPIError() did not return an APIError type")
	}

	if typedErr.Method != "GET" {
		t.Errorf("NewAPIError() Method = %v, want %v", typedErr.Method, "GET")
	}
	if typedErr.Path != "/test" {
		t.Errorf("NewAPIError() Path = %v, want %v", typedErr.Path, "/test")
	}
	if typedErr.StatusCode != http.StatusBadRequest {
		t.Errorf("NewAPIError() StatusCode = %v, want %v", typedErr.StatusCode, http.StatusBadRequest)
	}
	if typedErr.Err != err {
		t.Errorf("NewAPIError() Err = %v, want %v", typedErr.Err, err)
	}
}

func TestNewValidationError(t *testing.T) {
	valErr := NewValidationError("email", "invalid@email", "format", "invalid email format")

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
	if typedErr.Rule != "format" {
		t.Errorf("NewValidationError() Rule = %v, want %v", typedErr.Rule, "format")
	}
	if typedErr.Message != "invalid email format" {
		t.Errorf("NewValidationError() Message = %v, want %v", typedErr.Message, "invalid email format")
	}
}

func TestStatusErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantError  error
	}{
		{"BadRequest", http.StatusBadRequest, ErrInvalidRequest},
		{"Unauthorized", http.StatusUnauthorized, ErrUnauthorized},
		{"Forbidden", http.StatusForbidden, ErrForbidden},
		{"NotFound", http.StatusNotFound, ErrNotFound},
		{"MethodNotAllowed", http.StatusMethodNotAllowed, ErrMethodNotAllowed},
		{"Conflict", http.StatusConflict, ErrConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StatusErrors[tt.statusCode]; got != tt.wantError {
				t.Errorf("StatusErrors[%v] = %v, want %v", tt.statusCode, got, tt.wantError)
			}
		})
	}
}
