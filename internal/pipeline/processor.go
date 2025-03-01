package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/lindehoff/Budget-Assist/internal/ai"
	"github.com/lindehoff/Budget-Assist/internal/db"
	"github.com/lindehoff/Budget-Assist/internal/docprocess"
	"github.com/lindehoff/Budget-Assist/internal/processor"
)

// ProcessOptions contains runtime options for document processing
type ProcessOptions struct {
	// DocumentType specifies the type of document being processed (e.g., "receipt", "bank_statement", "invoice")
	DocumentType string
	// TransactionInsights provides additional context about the transactions in the document
	TransactionInsights string
	// CategoryInsights provides hints or context for transaction categorization
	CategoryInsights string
}

// ProcessingResult represents the result of processing a document
type ProcessingResult struct {
	FilePath          string
	TransactionsFound int
	Error             error
}

// Pipeline handles the document processing workflow
type Pipeline struct {
	docProcessor *docprocess.PDFProcessor
	csvProcessor *processor.SEBProcessor
	aiService    ai.Service
	store        db.Store
	logger       *slog.Logger
}

// NewPipeline creates a new processing pipeline
func NewPipeline(dp *docprocess.PDFProcessor, cp *processor.SEBProcessor, ai ai.Service, store db.Store, logger *slog.Logger) *Pipeline {
	return &Pipeline{
		docProcessor: dp,
		csvProcessor: cp,
		aiService:    ai,
		store:        store,
		logger:       logger,
	}
}

// ProcessDocuments processes all documents in the given path with the specified options
func (p *Pipeline) ProcessDocuments(ctx context.Context, path string, opts ProcessOptions) ([]ProcessingResult, error) {
	var results []ProcessingResult

	// Check if path is a file or directory
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to access path: %w", err)
	}

	if fileInfo.IsDir() {
		// Process directory
		if err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				result, err := p.processFile(ctx, path, opts)
				if err != nil {
					p.logger.Error("failed to process file", "path", path, "error", err)
					results = append(results, ProcessingResult{
						FilePath: path,
						Error:    err,
					})
				} else {
					results = append(results, result)
				}
			}
			return nil
		}); err != nil {
			return nil, fmt.Errorf("failed to walk directory: %w", err)
		}
	} else {
		// Process single file
		result, err := p.processFile(ctx, path, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to process file: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// processFile handles processing of a single file
func (p *Pipeline) processFile(ctx context.Context, path string, opts ProcessOptions) (ProcessingResult, error) {
	ext := strings.ToLower(filepath.Ext(path))

	var transactions []db.Transaction
	var err error

	switch ext {
	case ".pdf":
		transactions, err = p.processPDF(ctx, path, opts)
	case ".csv":
		transactions, err = p.processCSV(ctx, path, opts)
	default:
		return ProcessingResult{FilePath: path}, fmt.Errorf("unsupported file type: %s", ext)
	}

	if err != nil {
		return ProcessingResult{FilePath: path}, err
	}

	// Store transactions in database
	for _, tx := range transactions {
		if err := p.store.CreateTransaction(ctx, &tx); err != nil {
			p.logger.Error("failed to store transaction", "error", err)
			continue
		}
	}

	return ProcessingResult{
		FilePath:          path,
		TransactionsFound: len(transactions),
	}, nil
}

// processPDF handles PDF document processing
func (p *Pipeline) processPDF(ctx context.Context, path string, opts ProcessOptions) ([]db.Transaction, error) {
	// Extract text from PDF
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer file.Close()

	// Process PDF document
	result, err := p.docProcessor.Process(ctx, file, filepath.Base(path))
	if err != nil {
		return nil, fmt.Errorf("failed to process PDF: %w", err)
	}

	// Convert to database transactions
	transactions := make([]db.Transaction, 0, len(result.Transactions))
	for _, tx := range result.Transactions {
		// Convert raw data to JSON string
		rawData, err := json.Marshal(tx.RawData)
		if err != nil {
			p.logger.Error("failed to marshal raw data", "error", err)
			continue
		}

		// Create transaction record with insights
		dbTx := db.Transaction{
			Description:     tx.Description,
			Amount:          tx.Amount,
			TransactionDate: tx.Date,
			RawData:         string(rawData),
			Currency:        db.CurrencySEK, // Default to SEK
		}

		// Analyze transaction for categorization with insights
		analysis, err := p.aiService.AnalyzeTransaction(ctx, &dbTx, ai.AnalysisOptions{
			DocumentType:    opts.DocumentType,
			RuntimeInsights: opts.TransactionInsights + "\n" + opts.CategoryInsights,
		})
		if err != nil {
			p.logger.Error("failed to analyze transaction", "error", err)
			continue
		}

		// Convert analysis to JSON string
		aiAnalysis, err := json.Marshal(analysis)
		if err != nil {
			p.logger.Error("failed to marshal AI analysis", "error", err)
			continue
		}

		dbTx.AIAnalysis = string(aiAnalysis)
		transactions = append(transactions, dbTx)
	}

	return transactions, nil
}

// processCSV handles CSV document processing
func (p *Pipeline) processCSV(ctx context.Context, path string, opts ProcessOptions) ([]db.Transaction, error) {
	// Process CSV using SEB processor
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV: %w", err)
	}
	defer file.Close()

	rawTransactions, err := p.csvProcessor.ProcessDocument(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("failed to process CSV: %w", err)
	}

	// Convert and categorize transactions
	transactions := make([]db.Transaction, 0, len(rawTransactions))
	for _, tx := range rawTransactions {
		// Convert raw data to JSON string
		rawData, err := json.Marshal(tx.RawData)
		if err != nil {
			p.logger.Error("failed to marshal raw data", "error", err)
			continue
		}

		dbTx := db.Transaction{
			Description:     tx.Description,
			Amount:          tx.Amount,
			TransactionDate: tx.Date,
			RawData:         string(rawData),
			Currency:        db.CurrencySEK, // Default to SEK
		}

		// Only analyze with AI if service is available
		if p.aiService != nil {
			analysis, err := p.aiService.AnalyzeTransaction(ctx, &dbTx, ai.AnalysisOptions{
				DocumentType:    opts.DocumentType,
				RuntimeInsights: opts.TransactionInsights + "\n" + opts.CategoryInsights,
			})
			if err != nil {
				p.logger.Error("failed to analyze transaction", "error", err)
			} else {
				// Convert analysis to JSON string
				aiAnalysis, err := json.Marshal(analysis)
				if err != nil {
					p.logger.Error("failed to marshal AI analysis", "error", err)
				} else {
					dbTx.AIAnalysis = string(aiAnalysis)
				}
			}
		}

		transactions = append(transactions, dbTx)
	}

	return transactions, nil
}
