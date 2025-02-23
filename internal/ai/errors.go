package ai

import "fmt"

// AIError represents errors from AI operations
type AIError struct {
	Err        error
	Operation  string
	Model      string
	StatusCode int
}

func (e AIError) Error() string {
	if e.StatusCode != 0 {
		return fmt.Sprintf("AI %s operation failed with model %q (status %d): %v",
			e.Operation, e.Model, e.StatusCode, e.Err)
	}
	return fmt.Sprintf("AI %s operation failed with model %q: %v",
		e.Operation, e.Model, e.Err)
}

// CategoryError represents category prediction errors
type CategoryError struct {
	Err         error
	Transaction string
	Confidence  float64
}

func (e CategoryError) Error() string {
	return fmt.Sprintf("category prediction failed for transaction %q (confidence: %.2f): %v",
		e.Transaction, e.Confidence, e.Err)
}

// Common errors
var (
	ErrLowConfidence    = fmt.Errorf("prediction confidence below threshold")
	ErrInvalidModel     = fmt.Errorf("invalid model specified")
	ErrModelUnavailable = fmt.Errorf("AI model currently unavailable")
	ErrInvalidInput     = fmt.Errorf("invalid input for AI processing")
)

// Error constructors
func NewAIError(operation, model string, err error) error {
	return AIError{
		Operation: operation,
		Model:     model,
		Err:       err,
	}
}

func NewAIStatusError(operation, model string, statusCode int, err error) error {
	return AIError{
		Operation:  operation,
		Model:      model,
		StatusCode: statusCode,
		Err:        err,
	}
}

func NewCategoryError(transaction string, confidence float64, err error) error {
	return CategoryError{
		Transaction: transaction,
		Confidence:  confidence,
		Err:         err,
	}
}
