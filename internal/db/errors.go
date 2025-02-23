package db

import "fmt"

// DatabaseOperationError represents an error that occurs during database operations
type DatabaseOperationError struct {
	Operation string
	Entity    string
	Err       error
}

func (e DatabaseOperationError) Error() string {
	return fmt.Sprintf("database %s failed for %s: %v", e.Operation, e.Entity, e.Err)
}

// Common errors
var (
	ErrInvalidCategoryType = fmt.Errorf("invalid category type")
	ErrDuplicateEntry      = fmt.Errorf("duplicate entry")
	ErrTranslationMissing  = fmt.Errorf("required translation missing")
	ErrInvalidLanguage     = fmt.Errorf("invalid language code")
)
