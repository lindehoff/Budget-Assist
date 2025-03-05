package ai

import (
	"context"
	"time"

	"github.com/lindehoff/Budget-Assist/internal/db"
)

// Config represents the configuration for the OpenAI service
type Config struct {
	BaseURL        string
	APIKey         string
	Model          string
	RequestTimeout time.Duration
	MaxRetries     int
}

// Document represents a document to be analyzed
type Document struct {
	Content         []byte
	Type            string
	RuntimeInsights string
}

// AnalysisOptions represents options for transaction analysis
type AnalysisOptions struct {
	DocumentType    string
	RuntimeInsights string
}

// CategoryMatch represents a suggested category with confidence
type CategoryMatch struct {
	Category   string                 `json:"category"`
	Confidence float64                `json:"confidence"`
	Raw        map[string]interface{} `json:"-"`
}

// Analysis represents the result of analyzing a transaction
type Analysis struct {
	Category      string  `json:"category"`
	Subcategory   string  `json:"subcategory"`
	CategoryID    int     `json:"category_id"`
	SubcategoryID int     `json:"subcategory_id"`
	Confidence    float64 `json:"confidence"`
}

// Extraction represents the result of extracting information from a document
type Extraction struct {
	Date         string                   `json:"date"`
	Amount       float64                  `json:"amount"`
	Currency     string                   `json:"currency"`
	Description  string                   `json:"description"`
	Category     string                   `json:"category"`
	Subcategory  string                   `json:"subcategory"`
	Content      string                   `json:"content"`
	Transactions []map[string]interface{} `json:"transactions"`
}

// Service defines the interface for AI services
type Service interface {
	AnalyzeTransaction(ctx context.Context, tx *db.Transaction, opts AnalysisOptions) (*Analysis, error)
	ExtractDocument(ctx context.Context, doc *Document) (*Extraction, error)
	SuggestCategories(ctx context.Context, description string) ([]CategoryMatch, error)
	BatchAnalyzeTransactions(ctx context.Context, transactions []*db.Transaction, opts AnalysisOptions) ([]*Analysis, error)
}

// ModelExample represents a training example for the AI model
type ModelExample struct {
	Input  string  `json:"input"`
	Output string  `json:"output"`
	Score  float64 `json:"score,omitempty"`
}

// ChatCompletionResponse represents the response from OpenAI's chat completion API
type ChatCompletionResponse struct {
	Choices []Choice `json:"choices"`
}

// Choice represents a choice in the OpenAI API response
type Choice struct {
	Message Message `json:"message"`
}

// Message represents a message in the OpenAI API response
type Message struct {
	Content string `json:"content"`
}
