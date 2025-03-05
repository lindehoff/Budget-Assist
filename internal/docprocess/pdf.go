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
	return p.ProcessWithOptions(ctx, file, filename, ProcessOptions{})
}

// ProcessWithOptions processes a PDF document with additional options
func (p *PDFProcessor) ProcessWithOptions(ctx context.Context, file io.Reader, filename string, opts ProcessOptions) (*ProcessingResult, error) {
	// Extract text from PDF
	text, err := p.extractTextFromPDF(ctx, file)
	if err != nil {
		return nil, err
	}

	p.logger.Info("extracted text content",
		"filename", filename,
		"text_length", len(text))

	// Extract transactions using AI service
	extraction, err := p.extractDocumentWithAI(ctx, text, opts)
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

// extractDocumentWithAI extracts document information using AI
func (p *PDFProcessor) extractDocumentWithAI(ctx context.Context, text string, opts ProcessOptions) (*ai.Extraction, error) {
	if p.aiService == nil {
		return nil, fmt.Errorf("AI service is not configured")
	}

	doc := &ai.Document{
		Content:         []byte(text),
		Type:            "bill", // Default to bill type
		RuntimeInsights: opts.RuntimeInsights,
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

	// If we have transactions from the AI extraction, process them
	if len(extraction.Transactions) > 0 {
		p.logger.Debug("Processing transactions from AI extraction",
			"transactions_count", len(extraction.Transactions))

		return p.processAITransactions(extraction.Transactions, extraction.Date)
	}

	// Fallback to the old method if no transactions in the extraction
	p.logger.Debug("No transactions in AI extraction, using fallback method")
	return p.processFallbackTransaction(extraction)
}

// processAITransactions processes transactions from AI extraction
func (p *PDFProcessor) processAITransactions(txDataList []map[string]interface{}, fallbackDate string) []Transaction {
	var transactions []Transaction

	for _, txData := range txDataList {
		tx, ok := p.createTransactionFromMap(txData, fallbackDate)
		if ok {
			transactions = append(transactions, tx)
		}
	}

	return transactions
}

// createTransactionFromMap creates a transaction from a map of data
func (p *PDFProcessor) createTransactionFromMap(txData map[string]interface{}, fallbackDate string) (Transaction, bool) {
	// Extract description
	description := p.extractDescription(txData)
	if description == "" {
		p.logger.Debug("Skipping transaction with empty description")
		return Transaction{}, false
	}

	// Extract amount
	amount := p.extractAmount(txData, description)
	if amount.IsZero() {
		p.logger.Warn("Skipping transaction with zero amount", "description", description)
		return Transaction{}, false
	}

	// Extract date
	date := p.extractDate(txData, fallbackDate, description)
	if date.IsZero() {
		p.logger.Warn("Skipping transaction due to invalid date",
			"description", description,
			"amount", amount)
		return Transaction{}, false
	}

	// Extract category and subcategory
	category, subcategory := p.extractCategories(txData)

	// Create transaction
	return Transaction{
		Description: description,
		Amount:      amount,
		Date:        date,
		RawData:     txData,
		Category:    category,
		SubCategory: subcategory,
		Source:      "pdf",
	}, true
}

// extractDescription extracts the description from transaction data
func (p *PDFProcessor) extractDescription(txData map[string]interface{}) string {
	if desc, ok := txData["description"].(string); ok && desc != "" {
		return desc
	}
	if desc, ok := txData["beskrivning"].(string); ok && desc != "" {
		return desc
	}
	return ""
}

// extractAmount extracts the amount from transaction data
func (p *PDFProcessor) extractAmount(txData map[string]interface{}, description string) decimal.Decimal {
	// Try English field name first
	amount := p.parseAmount(txData, "amount", description)

	// If no amount found, try Swedish field name
	if amount.IsZero() {
		amount = p.parseAmount(txData, "belopp", description)
	}

	return amount
}

// parseAmount parses an amount from a specific field in the transaction data
func (p *PDFProcessor) parseAmount(txData map[string]interface{}, fieldName string, description string) decimal.Decimal {
	switch v := txData[fieldName].(type) {
	case float64:
		return decimal.NewFromFloat(v)
	case int:
		return decimal.NewFromInt(int64(v))
	case string:
		amount, err := decimal.NewFromString(v)
		if err != nil {
			p.logger.Warn("Invalid amount in transaction",
				"description", description,
				"field", fieldName,
				"value", v,
				"error", err)
			return decimal.Zero
		}
		return amount
	}
	return decimal.Zero
}

// extractDate extracts the date from transaction data
func (p *PDFProcessor) extractDate(txData map[string]interface{}, fallbackDate string, description string) time.Time {
	// Try English field name first
	date := p.parseDate(txData, "date", description)
	if !date.IsZero() {
		return date
	}

	// Try Swedish field name
	date = p.parseDate(txData, "datum", description)
	if !date.IsZero() {
		return date
	}

	// If date is still empty, use the extraction date
	if fallbackDate != "" {
		date, err := time.Parse("2006-01-02", fallbackDate)
		if err != nil {
			p.logger.Warn("Invalid date format in extraction",
				"date", fallbackDate,
				"error", err)
			return time.Time{}
		}
		return date
	}

	return time.Time{}
}

// parseDate parses a date from a specific field in the transaction data
func (p *PDFProcessor) parseDate(txData map[string]interface{}, fieldName string, description string) time.Time {
	if dateStr, ok := txData[fieldName].(string); ok && dateStr != "" {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			p.logger.Warn("Invalid date format in transaction",
				"description", description,
				"field", fieldName,
				"date", dateStr,
				"error", err)
			return time.Time{}
		}
		return date
	}
	return time.Time{}
}

// extractCategories extracts category and subcategory from transaction data
func (p *PDFProcessor) extractCategories(txData map[string]interface{}) (string, string) {
	// Extract category
	category, _ := txData["category"].(string)
	if category == "" {
		category, _ = txData["kategori"].(string)
	}

	// Extract subcategory
	subcategory, _ := txData["subcategory"].(string)
	if subcategory == "" {
		subcategory, _ = txData["underkategori"].(string)
	}

	return category, subcategory
}

// processFallbackTransaction processes a single transaction from extraction data
func (p *PDFProcessor) processFallbackTransaction(extraction *ai.Extraction) []Transaction {
	var transactions []Transaction

	// Create a single transaction from the extraction
	if extraction.Description == "" || extraction.Amount == 0 {
		return transactions
	}

	// Parse the date
	date, err := time.Parse("2006-01-02", extraction.Date)
	if err != nil {
		p.logger.Warn("invalid date format in extraction",
			"date", extraction.Date,
			"error", err)
		return transactions
	}

	if date.IsZero() {
		return transactions
	}

	// Convert extraction to map for RawData
	rawData := p.extractionToMap(extraction)

	transactions = append(transactions, Transaction{
		Description: extraction.Description,
		Amount:      decimal.NewFromFloat(extraction.Amount),
		Date:        date,
		RawData:     rawData,
		Category:    extraction.Category,
		SubCategory: extraction.Subcategory,
		Source:      "pdf",
	})

	return transactions
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
