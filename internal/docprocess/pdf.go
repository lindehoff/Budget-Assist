package docprocess

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

// PDFProcessor implements DocumentProcessor for PDF files
type PDFProcessor struct {
	logger *slog.Logger
}

// NewPDFProcessor creates a new PDFProcessor instance
func NewPDFProcessor(logger *slog.Logger) *PDFProcessor {
	if logger == nil {
		logger = slog.Default()
	}
	return &PDFProcessor{
		logger: logger,
	}
}

// Type returns the document type this processor handles
func (p *PDFProcessor) Type() DocumentType {
	return TypePDF
}

// Validate checks if the provided file is a valid PDF
func (p *PDFProcessor) Validate(file io.Reader) error {
	// Convert io.Reader to []byte for pdfcpu
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(file); err != nil {
		return &ProcessingError{
			Stage:    StageValidation,
			Document: "unknown",
			Err:      fmt.Errorf("failed to read PDF: %w", err),
		}
	}

	// Validate PDF
	conf := pdfcpu.NewDefaultConfiguration()
	if err := api.Validate(bytes.NewReader(buf.Bytes()), conf); err != nil {
		return &ProcessingError{
			Stage:    StageValidation,
			Document: "unknown",
			Err:      fmt.Errorf("invalid PDF format: %w", err),
		}
	}

	return nil
}

// Process extracts text and potential transaction data from a PDF document
func (p *PDFProcessor) Process(ctx context.Context, file io.Reader, filename string) (*ProcessingResult, error) {
	// Convert io.Reader to []byte for pdfcpu
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(file); err != nil {
		return nil, &ProcessingError{
			Stage:    StageExtraction,
			Document: filename,
			Err:      fmt.Errorf("failed to read PDF: %w", err),
		}
	}

	// Get page count
	conf := pdfcpu.NewDefaultConfiguration()
	pageCount, err := api.PageCount(bytes.NewReader(buf.Bytes()), conf)
	if err != nil {
		return nil, &ProcessingError{
			Stage:    StageExtraction,
			Document: filename,
			Err:      fmt.Errorf("failed to get page count: %w", err),
		}
	}

	// Create a temporary directory for content extraction
	tempDir, err := os.MkdirTemp("", "pdf-extract-*")
	if err != nil {
		return nil, &ProcessingError{
			Stage:    StageExtraction,
			Document: filename,
			Err:      fmt.Errorf("failed to create temp directory: %w", err),
		}
	}
	defer os.RemoveAll(tempDir)

	// Extract content to temporary directory
	extractErr := api.ExtractContent(bytes.NewReader(buf.Bytes()), tempDir, "content", nil, conf)
	if extractErr != nil {
		return nil, &ProcessingError{
			Stage:    StageExtraction,
			Document: filename,
			Err:      fmt.Errorf("failed to extract content: %w", extractErr),
		}
	}

	// Read extracted content
	textBuf := new(bytes.Buffer)
	contentFiles, err := filepath.Glob(filepath.Join(tempDir, "*.txt"))
	if err != nil {
		return nil, &ProcessingError{
			Stage:    StageExtraction,
			Document: filename,
			Err:      fmt.Errorf("failed to find extracted content: %w", err),
		}
	}

	for _, file := range contentFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			p.logger.Warn("failed to read content file",
				"file", file,
				"error", err)
			continue
		}
		textBuf.Write(content)
	}

	// Create processing result with metadata
	result := &ProcessingResult{
		Transactions: make([]Transaction, 0),
		Metadata: map[string]any{
			"filename":     filename,
			"content_type": "application/pdf",
			"text_length":  textBuf.Len(),
			"page_count":   pageCount,
		},
		Warnings:    make([]string, 0),
		ProcessedAt: time.Now(),
	}

	// TODO: Implement transaction extraction logic
	// This will involve:
	// 1. Pattern matching for transaction data
	// 2. Date and amount extraction
	// 3. Category inference
	// 4. Data normalization

	p.logger.Info("processed PDF document",
		"filename", filename,
		"text_length", textBuf.Len(),
		"page_count", pageCount,
		"transactions_found", len(result.Transactions))

	return result, nil
}

// DefaultProcessorFactory implements ProcessorFactory
type DefaultProcessorFactory struct {
	logger *slog.Logger
}

// NewDefaultProcessorFactory creates a new processor factory
func NewDefaultProcessorFactory(logger *slog.Logger) *DefaultProcessorFactory {
	return &DefaultProcessorFactory{
		logger: logger,
	}
}

// CreateProcessor returns a processor for the given document type
func (f *DefaultProcessorFactory) CreateProcessor(docType DocumentType) (DocumentProcessor, error) {
	switch docType {
	case TypePDF:
		return NewPDFProcessor(f.logger), nil
	default:
		return nil, fmt.Errorf("unsupported document type: %s", docType)
	}
}

// SupportedTypes returns a list of supported document types
func (f *DefaultProcessorFactory) SupportedTypes() []DocumentType {
	return []DocumentType{TypePDF}
}
