package ai

import (
	"context"
	"time"

	db "github.com/lindehoff/Budget-Assist/internal/db"
)

// Document represents a document to be processed
type Document struct {
	Content []byte
}

// Extraction represents extracted information from a document
type Extraction struct {
	Content string `json:"content"`
}

// CategoryMatch represents a category suggestion with confidence score
type CategoryMatch struct {
	Category   string  `json:"category"`
	Confidence float64 `json:"confidence"`
}

// Analysis represents the result of analyzing a transaction
type Analysis struct {
	Remarks string  `json:"remarks"`
	Score   float64 `json:"score"`
}

// AnalysisOptions contains runtime options for transaction analysis
type AnalysisOptions struct {
	// DocumentType specifies the type of document being processed
	DocumentType string
	// RuntimeInsights provides additional context about the transactions
	RuntimeInsights string
}

// Service defines the interface for AI operations
type Service interface {
	// AnalyzeTransaction analyzes a transaction using the specified prompt type
	AnalyzeTransaction(ctx context.Context, tx *db.Transaction, opts AnalysisOptions) (*Analysis, error)

	// ExtractDocument extracts information from a document using the appropriate prompt type
	ExtractDocument(ctx context.Context, doc *Document) (*Extraction, error)

	// SuggestCategories suggests categories for a transaction description
	SuggestCategories(ctx context.Context, desc string) ([]CategoryMatch, error)
}

// Config holds configuration for the AI service
type Config struct {
	BaseURL        string
	APIKey         string
	RequestTimeout time.Duration
	MaxRetries     int
}

// ModelExample represents a training example for the AI model
type ModelExample struct {
	Input  string  `json:"input"`
	Output string  `json:"output"`
	Score  float64 `json:"score,omitempty"`
}

// OpenAIResponse represents the raw response from OpenAI API
type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}
