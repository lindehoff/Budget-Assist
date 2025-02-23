package docprocess

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/shopspring/decimal"
)

// ProcessingStage represents different stages in document processing
type ProcessingStage string

const (
	StageValidation    ProcessingStage = "validation"
	StageExtraction    ProcessingStage = "extraction"
	StageNormalization ProcessingStage = "normalization"
	StageAnalysis      ProcessingStage = "analysis"
)

// ProcessingError represents an error that occurred during document processing
type ProcessingError struct {
	Err      error
	Stage    ProcessingStage
	Document string
}

func (e *ProcessingError) Error() string {
	return fmt.Sprintf("%s failed for document %q: %v", e.Stage, e.Document, e.Err)
}

func (e *ProcessingError) Unwrap() error {
	return e.Err
}

// DocumentType represents supported document types
type DocumentType string

const (
	TypePDF  DocumentType = "pdf"
	TypeCSV  DocumentType = "csv"
	TypeXLSX DocumentType = "xlsx"
)

// ProcessingResult contains the extracted data and metadata from a document
type ProcessingResult struct {
	ProcessedAt  time.Time
	Transactions []Transaction
	Metadata     map[string]any
	Warnings     []string
}

// Transaction represents a financial transaction extracted from a document
type Transaction struct {
	Date        time.Time
	Amount      decimal.Decimal
	RawData     map[string]any
	Description string
	Category    string
	SubCategory string
	Source      string
}

// DocumentProcessor defines the interface for processing different types of documents
type DocumentProcessor interface {
	// Process processes a document and returns extracted transactions
	Process(ctx context.Context, file io.Reader, filename string) (*ProcessingResult, error)

	// Validate validates if the document can be processed
	Validate(file io.Reader) error

	// Type returns the type of documents this processor can handle
	Type() DocumentType
}

// ProcessorFactory creates document processors based on document type
type ProcessorFactory interface {
	// CreateProcessor returns a processor for the given document type
	CreateProcessor(docType DocumentType) (DocumentProcessor, error)

	// SupportedTypes returns a list of supported document types
	SupportedTypes() []DocumentType
}
