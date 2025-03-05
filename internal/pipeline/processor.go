package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/lindehoff/Budget-Assist/internal/ai"
	"github.com/lindehoff/Budget-Assist/internal/db"
	"github.com/lindehoff/Budget-Assist/internal/docprocess"
	"github.com/lindehoff/Budget-Assist/internal/processor"
)

// ProcessOptions contains runtime options for document processing
type ProcessOptions struct {
	// DocumentType specifies the type of document being processed (e.g., "receipt", "bank_statement", "bill")
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

	// Log transaction details before storing
	p.logger.Info("Extracted transactions summary:", "count", len(transactions))
	for i, tx := range transactions {
		// Get category and subcategory names
		var categoryName, subcategoryName string
		if tx.CategoryID != nil {
			category, err := p.store.GetCategoryByID(ctx, *tx.CategoryID)
			if err == nil && category != nil {
				categoryName = category.Name
			}
		}
		if tx.SubcategoryID != nil {
			subcategory, err := p.store.GetSubcategoryByID(ctx, *tx.SubcategoryID)
			if err == nil && subcategory != nil {
				subcategoryName = subcategory.Name
			}
		}

		p.logger.Info(fmt.Sprintf("Transaction #%d:", i+1),
			"description", tx.Description,
			"amount", tx.Amount,
			"date", tx.Date.Format("2006-01-02"),
			"category", categoryName,
			"subcategory", subcategoryName,
			"category_id", tx.CategoryID,
			"subcategory_id", tx.SubcategoryID)
	}

	// Store transactions in database
	for _, tx := range transactions {
		if err := p.store.CreateTransaction(ctx, &tx); err != nil {
			p.logger.Error("failed to store transaction", "error", err)
			continue
		}

		// Log successful creation with ID
		p.logger.Info("Transaction created successfully",
			"id", tx.ID,
			"description", tx.Description,
			"amount", tx.Amount,
			"date", tx.Date.Format("2006-01-02"))
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

	// Process PDF document with options
	docOpts := docprocess.ProcessOptions{
		RuntimeInsights: opts.TransactionInsights,
	}
	result, err := p.docProcessor.ProcessWithOptions(ctx, file, filepath.Base(path), docOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to process PDF: %w", err)
	}

	// Convert to database transactions
	dbTransactions := make([]*db.Transaction, 0, len(result.Transactions))
	transactions := make([]db.Transaction, 0, len(result.Transactions))

	for _, tx := range result.Transactions {
		// Convert raw data to JSON string
		rawData, err := json.Marshal(tx.RawData)
		if err != nil {
			p.logger.Error("failed to marshal raw data", "error", err)
			continue
		}

		// Create transaction record with insights
		dbTx := &db.Transaction{
			Description:     tx.Description,
			Amount:          tx.Amount,
			Date:            time.Now(), // Set to current time (import time)
			TransactionDate: tx.Date,    // Keep the actual transaction date
			RawData:         string(rawData),
			Currency:        db.CurrencySEK, // Default to SEK
		}

		// Extract category IDs from raw data if available
		p.setCategoryIDsFromRawData(dbTx)

		dbTransactions = append(dbTransactions, dbTx)
	}

	// Skip analysis if no transactions found
	if len(dbTransactions) == 0 {
		p.logger.Warn("No transactions found in document", "path", path)
		return transactions, nil
	}

	// Analyze all transactions at once for categorization
	p.logger.Info("Analyzing transactions for categorization",
		"transaction_count", len(dbTransactions),
		"document_type", opts.DocumentType)

	analyses, err := p.aiService.BatchAnalyzeTransactions(ctx, dbTransactions, ai.AnalysisOptions{
		DocumentType:    opts.DocumentType,
		RuntimeInsights: opts.TransactionInsights + "\n" + opts.CategoryInsights,
	})

	if err != nil {
		p.logger.Error("failed to analyze transactions", "error", err)
		// Continue without categories
		for _, dbTx := range dbTransactions {
			transactions = append(transactions, *dbTx)
		}
		return transactions, nil
	}

	// Apply analyses to transactions
	for i, dbTx := range dbTransactions {
		if i < len(analyses) {
			// Convert analysis to JSON string
			aiAnalysis, err := json.Marshal(analyses[i])
			if err != nil {
				p.logger.Error("failed to marshal AI analysis", "error", err)
				continue
			}

			dbTx.AIAnalysis = string(aiAnalysis)

			// Set category and subcategory IDs from analysis
			if analyses[i].CategoryID > 0 {
				// Safe conversion since we've checked the value is positive
				categoryID := uint(analyses[i].CategoryID) // #nosec G115
				dbTx.CategoryID = &categoryID
				p.logger.Debug("Setting category ID from AI analysis",
					"transaction_description", dbTx.Description,
					"category_id", categoryID,
					"category", analyses[i].Category)
			}

			if analyses[i].SubcategoryID > 0 {
				// Safe conversion since we've checked the value is positive
				subcategoryID := uint(analyses[i].SubcategoryID) // #nosec G115
				dbTx.SubcategoryID = &subcategoryID
				p.logger.Debug("Setting subcategory ID from AI analysis",
					"transaction_description", dbTx.Description,
					"subcategory_id", subcategoryID,
					"subcategory", analyses[i].Subcategory)
			}
		}

		transactions = append(transactions, *dbTx)
	}

	return transactions, nil
}

// setCategoryIDsFromRawData extracts category and subcategory IDs from raw data
func (p *Pipeline) setCategoryIDsFromRawData(tx *db.Transaction) {
	if tx.RawData == "" {
		return
	}

	var rawData map[string]interface{}
	if err := json.Unmarshal([]byte(tx.RawData), &rawData); err != nil {
		p.logger.Error("failed to unmarshal raw data", "error", err)
		return
	}

	// Extract category_id
	if categoryIDVal, ok := rawData["category_id"]; ok {
		var categoryID uint
		switch v := categoryIDVal.(type) {
		case float64:
			if v > 0 {
				categoryID = uint(v)
				tx.CategoryID = &categoryID
				p.logger.Debug("Setting category ID from raw data",
					"transaction_description", tx.Description,
					"category_id", categoryID)
			}
		case int:
			if v > 0 {
				categoryID = uint(v)
				tx.CategoryID = &categoryID
				p.logger.Debug("Setting category ID from raw data",
					"transaction_description", tx.Description,
					"category_id", categoryID)
			}
		case string:
			if id, err := strconv.ParseUint(v, 10, 32); err == nil && id > 0 {
				categoryID = uint(id)
				tx.CategoryID = &categoryID
				p.logger.Debug("Setting category ID from raw data",
					"transaction_description", tx.Description,
					"category_id", categoryID)
			}
		}
	}

	// Extract subcategory_id
	if subcategoryIDVal, ok := rawData["subcategory_id"]; ok {
		var subcategoryID uint
		switch v := subcategoryIDVal.(type) {
		case float64:
			if v > 0 {
				subcategoryID = uint(v)
				tx.SubcategoryID = &subcategoryID
				p.logger.Debug("Setting subcategory ID from raw data",
					"transaction_description", tx.Description,
					"subcategory_id", subcategoryID)
			}
		case int:
			if v > 0 {
				subcategoryID = uint(v)
				tx.SubcategoryID = &subcategoryID
				p.logger.Debug("Setting subcategory ID from raw data",
					"transaction_description", tx.Description,
					"subcategory_id", subcategoryID)
			}
		case string:
			if id, err := strconv.ParseUint(v, 10, 32); err == nil && id > 0 {
				subcategoryID = uint(id)
				tx.SubcategoryID = &subcategoryID
				p.logger.Debug("Setting subcategory ID from raw data",
					"transaction_description", tx.Description,
					"subcategory_id", subcategoryID)
			}
		}
	}
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
			Date:            time.Now(), // Set to current time (import time)
			TransactionDate: tx.Date,    // Keep the actual transaction date
			RawData:         string(rawData),
			Currency:        db.CurrencySEK, // Default to SEK
		}

		// Extract category IDs from raw data if available
		p.setCategoryIDsFromRawData(&dbTx)

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

					// Set category and subcategory IDs from analysis
					if analysis.CategoryID > 0 {
						// Safe conversion since we've checked the value is positive
						categoryID := uint(analysis.CategoryID)
						dbTx.CategoryID = &categoryID
						p.logger.Debug("Setting category ID from AI analysis",
							"transaction_description", dbTx.Description,
							"category_id", categoryID,
							"category", analysis.Category)
					}

					if analysis.SubcategoryID > 0 {
						// Safe conversion since we've checked the value is positive
						subcategoryID := uint(analysis.SubcategoryID)
						dbTx.SubcategoryID = &subcategoryID
						p.logger.Debug("Setting subcategory ID from AI analysis",
							"transaction_description", dbTx.Description,
							"subcategory_id", subcategoryID,
							"subcategory", analysis.Subcategory)
					}
				}
			}
		}

		transactions = append(transactions, dbTx)
	}

	return transactions, nil
}
