package docprocess

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/lindehoff/Budget-Assist/internal/ai"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/shopspring/decimal"
)

// PDFProcessor processes PDF documents
type PDFProcessor struct {
	logger    *slog.Logger
	aiService ai.Service

	// Function to extract text from PDF, can be replaced in tests
	extractTextFromPDF func(ctx context.Context, file io.Reader) (string, error)
}

// NewPDFProcessor creates a new PDF processor
func NewPDFProcessor(logger *slog.Logger, aiService ai.Service) *PDFProcessor {
	p := &PDFProcessor{
		logger:    logger,
		aiService: aiService,
	}

	// Set the default implementation
	p.extractTextFromPDF = p.defaultExtractTextFromPDF

	return p
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
	conf := model.NewDefaultConfiguration()
	if err := api.Validate(bytes.NewReader(buf.Bytes()), conf); err != nil {
		return &ProcessingError{
			Stage:    StageValidation,
			Document: "unknown",
			Err:      fmt.Errorf("invalid PDF format: %w", err),
		}
	}

	return nil
}

// Process processes a PDF file and extracts transactions
func (p *PDFProcessor) Process(ctx context.Context, file io.Reader, filename string) (*ProcessingResult, error) {
	// Extract text from PDF
	text, err := p.extractTextFromPDF(ctx, file)
	if err != nil {
		return nil, err
	}

	p.logger.Info("extracted text content",
		"filename", filename,
		"text_length", len(text))

	// Extract transactions using AI service
	extraction, err := p.extractDocumentWithAI(ctx, text)
	if err != nil {
		return nil, err
	}

	// Convert AI extraction to transactions
	transactions := p.convertExtractionToTransactions(extraction)

	return &ProcessingResult{
		Transactions: transactions,
		Metadata: map[string]any{
			"filename":     filename,
			"content_type": "application/pdf",
			"text_length":  len(text),
		},
		Warnings:    make([]string, 0),
		ProcessedAt: time.Now(),
	}, nil
}

// defaultExtractTextFromPDF extracts text content from a PDF file
func (p *PDFProcessor) defaultExtractTextFromPDF(ctx context.Context, file io.Reader) (string, error) {
	// Create a temporary file to store the PDF content
	tempFile, err := os.CreateTemp("", "pdf-*.pdf")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Copy the file content to the temp file
	if _, err := io.Copy(tempFile, file); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	// Check if pdftotext is installed
	if _, err := exec.LookPath("pdftotext"); err != nil {
		return "", fmt.Errorf("pdftotext is not installed. Please install poppler-utils")
	}

	// Extract text from PDF using pdftotext
	args := []string{
		"-layout",       // Maintain original layout
		"-nopgbrk",      // Don't insert page breaks
		tempFile.Name(), // Input file
		"-",             // Output to stdout
	}
	cmd := exec.CommandContext(ctx, "pdftotext", args...)

	var buf bytes.Buffer
	cmd.Stdout = &buf

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to extract text from PDF: %w", err)
	}

	text := buf.String()
	if text == "" {
		return "", fmt.Errorf("no text content found in PDF")
	}

	return text, nil
}

// extractDocumentWithAI uses the AI service to extract information from the document
func (p *PDFProcessor) extractDocumentWithAI(ctx context.Context, text string) (*ai.Extraction, error) {
	if p.aiService == nil {
		return nil, fmt.Errorf("AI service is not configured")
	}

	doc := &ai.Document{
		Content: []byte(text),
		Type:    "bill", // Default to bill type
	}

	extraction, err := p.aiService.ExtractDocument(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to extract document: %w", err)
	}

	return extraction, nil
}

// convertExtractionToTransactions converts AI extraction to transactions
func (p *PDFProcessor) convertExtractionToTransactions(extraction *ai.Extraction) []Transaction {
	var transactions []Transaction
	if extraction == nil {
		return transactions
	}

	// Parse the description to get individual transactions
	parts := strings.Split(extraction.Description, ", ")
	for _, part := range parts {
		// Skip empty parts
		if part == "" {
			continue
		}

		tx, ok := p.parseTransactionFromPart(part, extraction)
		if ok {
			transactions = append(transactions, tx)
		}
	}

	return transactions
}

// parseTransactionFromPart parses a transaction from a part of the description
func (p *PDFProcessor) parseTransactionFromPart(part string, extraction *ai.Extraction) (Transaction, bool) {
	// Check for empty part
	if part == "" {
		return Transaction{}, false
	}

	// Extract amount from description (e.g., "Bredband 600/50 (629.00 SEK)")
	var amount decimal.Decimal
	var description string

	if strings.Contains(part, " (") && strings.HasSuffix(part, " SEK)") {
		descParts := strings.Split(part, " (")
		description = descParts[0]
		amountStr := strings.TrimSuffix(descParts[1], " SEK)")
		if amt, err := decimal.NewFromString(amountStr); err == nil {
			amount = amt
		} else {
			amount = decimal.NewFromFloat(extraction.Amount)
		}
	} else {
		description = part
		amount = decimal.NewFromFloat(extraction.Amount)
	}

	// Parse the date
	var date time.Time
	var err error
	if extraction.Date != "" {
		date, err = time.Parse("2006-01-02", extraction.Date)
		if err != nil {
			p.logger.Error("failed to parse date",
				"date", extraction.Date,
				"error", err)
			return Transaction{}, false
		}
	}

	if date.IsZero() {
		p.logger.Warn("skipping transaction due to invalid date",
			"description", description,
			"amount", amount)
		return Transaction{}, false
	}

	// Convert extraction to map for RawData
	rawData := p.extractionToMap(extraction)

	return Transaction{
		Description: description,
		Amount:      amount,
		Date:        date,
		RawData:     rawData,
		Category:    extraction.Category,
		SubCategory: extraction.Subcategory,
		Source:      "pdf",
	}, true
}

// extractionToMap converts an extraction to a map for RawData
func (p *PDFProcessor) extractionToMap(extraction *ai.Extraction) map[string]any {
	rawData := make(map[string]any)
	if data, err := json.Marshal(extraction); err == nil {
		if err := json.Unmarshal(data, &rawData); err != nil {
			p.logger.Error("failed to convert extraction to map",
				"error", err)
		}
	}
	return rawData
}

// CanProcess returns true if the file is a PDF
func (p *PDFProcessor) CanProcess(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".pdf"
}

// DefaultProcessorFactory implements ProcessorFactory
type DefaultProcessorFactory struct {
	logger    *slog.Logger
	aiService ai.Service
}

// NewDefaultProcessorFactory creates a new processor factory
func NewDefaultProcessorFactory(logger *slog.Logger, aiService ai.Service) *DefaultProcessorFactory {
	return &DefaultProcessorFactory{
		logger:    logger,
		aiService: aiService,
	}
}

// CreateProcessor returns a processor for the given document type
func (f *DefaultProcessorFactory) CreateProcessor(docType DocumentType) (DocumentProcessor, error) {
	switch docType {
	case TypePDF:
		return NewPDFProcessor(f.logger, f.aiService), nil
	default:
		return nil, fmt.Errorf("unsupported document type: %s", docType)
	}
}

// SupportedTypes returns a list of supported document types
func (f *DefaultProcessorFactory) SupportedTypes() []DocumentType {
	return []DocumentType{TypePDF}
}
