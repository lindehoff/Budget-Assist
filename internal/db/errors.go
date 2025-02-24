package db

import (
	"errors"
	"fmt"
)

// DatabaseOperationError represents database operation errors
type DatabaseOperationError struct {
	Err       error
	Operation string
	Entity    string
}

func (e DatabaseOperationError) Error() string {
	return fmt.Sprintf("database %s operation failed for entity %q: %v", e.Operation, e.Entity, e.Err)
}

// Common errors
var (
	ErrInvalidCategoryType = fmt.Errorf("invalid category type")
	ErrDuplicateEntry      = fmt.Errorf("duplicate entry")
	ErrTranslationMissing  = fmt.Errorf("required translation missing")
	ErrInvalidLanguage     = fmt.Errorf("invalid language code")
	ErrNotFound            = errors.New("record not found")
)
