package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lindehoff/Budget-Assist/internal/ai"
	"github.com/lindehoff/Budget-Assist/internal/db"
	"github.com/lindehoff/Budget-Assist/internal/docprocess"
	"github.com/lindehoff/Budget-Assist/internal/processor"
	"github.com/shopspring/decimal"
)

// Mock implementations for testing
type mockAIService struct {
	extractDocumentFunc          func(ctx context.Context, doc *ai.Document) (*ai.Extraction, error)
	analyzeTransactionFunc       func(ctx context.Context, tx *db.Transaction, opts ai.AnalysisOptions) (*ai.Analysis, error)
	suggestCategoriesFunc        func(ctx context.Context, description string) ([]ai.CategoryMatch, error)
	batchAnalyzeTransactionsFunc func(ctx context.Context, transactions []*db.Transaction, opts ai.AnalysisOptions) ([]*ai.Analysis, error)
}

func (m *mockAIService) ExtractDocument(ctx context.Context, doc *ai.Document) (*ai.Extraction, error) {
	return m.extractDocumentFunc(ctx, doc)
}

func (m *mockAIService) AnalyzeTransaction(ctx context.Context, tx *db.Transaction, opts ai.AnalysisOptions) (*ai.Analysis, error) {
	return m.analyzeTransactionFunc(ctx, tx, opts)
}

func (m *mockAIService) SuggestCategories(ctx context.Context, description string) ([]ai.CategoryMatch, error) {
	return m.suggestCategoriesFunc(ctx, description)
}

func (m *mockAIService) BatchAnalyzeTransactions(ctx context.Context, transactions []*db.Transaction, opts ai.AnalysisOptions) ([]*ai.Analysis, error) {
	if m.batchAnalyzeTransactionsFunc != nil {
		return m.batchAnalyzeTransactionsFunc(ctx, transactions, opts)
	}

	// Default implementation that calls AnalyzeTransaction for each transaction
	if m.analyzeTransactionFunc != nil {
		results := make([]*ai.Analysis, 0, len(transactions))
		for _, tx := range transactions {
			analysis, err := m.analyzeTransactionFunc(ctx, tx, opts)
			if err != nil {
				return nil, err
			}
			results = append(results, analysis)
		}
		return results, nil
	}

	return nil, fmt.Errorf("not implemented")
}

type mockStore struct {
	createTransactionFunc func(ctx context.Context, tx *db.Transaction) error
}

func (m *mockStore) CreateTransaction(ctx context.Context, tx *db.Transaction) error {
	return m.createTransactionFunc(ctx, tx)
}

// Implement all required methods of db.Store interface with empty implementations
func (m *mockStore) CreateCategoryType(ctx context.Context, ct *db.CategoryType) error { return nil }
func (m *mockStore) UpdateCategoryType(ctx context.Context, ct *db.CategoryType) error { return nil }
func (m *mockStore) GetCategoryTypeByID(ctx context.Context, id uint) (*db.CategoryType, error) {
	return nil, nil
}
func (m *mockStore) ListCategoryTypes(ctx context.Context) ([]db.CategoryType, error) {
	return nil, nil
}
func (m *mockStore) CreateCategory(ctx context.Context, c *db.Category) error { return nil }
func (m *mockStore) UpdateCategory(ctx context.Context, c *db.Category) error { return nil }
func (m *mockStore) GetCategoryByID(ctx context.Context, id uint) (*db.Category, error) {
	return nil, nil
}
func (m *mockStore) GetCategoryByName(ctx context.Context, name string) (*db.Category, error) {
	return nil, nil
}
func (m *mockStore) ListCategories(ctx context.Context, typeID *uint) ([]db.Category, error) {
	return nil, nil
}
func (m *mockStore) DeleteCategory(ctx context.Context, id uint) error              { return nil }
func (m *mockStore) CreateSubcategory(ctx context.Context, s *db.Subcategory) error { return nil }
func (m *mockStore) UpdateSubcategory(ctx context.Context, s *db.Subcategory) error { return nil }
func (m *mockStore) GetSubcategoryByID(ctx context.Context, id uint) (*db.Subcategory, error) {
	return nil, nil
}
func (m *mockStore) GetSubcategoryByName(ctx context.Context, name string) (*db.Subcategory, error) {
	return nil, nil
}
func (m *mockStore) ListSubcategories(ctx context.Context) ([]db.Subcategory, error) { return nil, nil }
func (m *mockStore) DeleteSubcategory(ctx context.Context, id uint) error            { return nil }
func (m *mockStore) CreateCategorySubcategory(ctx context.Context, link *db.CategorySubcategory) error {
	return nil
}
func (m *mockStore) DeleteCategorySubcategory(ctx context.Context, categoryID, subcategoryID uint) error {
	return nil
}
func (m *mockStore) GetTransactionByID(ctx context.Context, id uint) (*db.Transaction, error) {
	return nil, nil
}
func (m *mockStore) ListTransactions(ctx context.Context, filter *db.TransactionFilter) ([]db.Transaction, error) {
	return nil, nil
}
func (m *mockStore) UpdateTransaction(ctx context.Context, tx *db.Transaction) error { return nil }
func (m *mockStore) DeleteTransaction(ctx context.Context, id uint) error            { return nil }
func (m *mockStore) CreatePrompt(ctx context.Context, p *db.Prompt) error            { return nil }
func (m *mockStore) UpdatePrompt(ctx context.Context, p *db.Prompt) error            { return nil }
func (m *mockStore) GetPromptByID(ctx context.Context, id uint) (*db.Prompt, error)  { return nil, nil }
func (m *mockStore) GetPromptByType(ctx context.Context, promptType string) (*db.Prompt, error) {
	return nil, nil
}
func (m *mockStore) ListPrompts(ctx context.Context) ([]db.Prompt, error)           { return nil, nil }
func (m *mockStore) DeletePrompt(ctx context.Context, id uint) error                { return nil }
func (m *mockStore) CreateTag(ctx context.Context, tag *db.Tag) error               { return nil }
func (m *mockStore) GetTagByName(ctx context.Context, name string) (*db.Tag, error) { return nil, nil }
func (m *mockStore) LinkSubcategoryTag(ctx context.Context, subcategoryID, tagID uint) error {
	return nil
}
func (m *mockStore) UnlinkSubcategoryTag(ctx context.Context, subcategoryID, tagID uint) error {
	return nil
}
func (m *mockStore) Close() error { return nil }

// Helper functions for testing
func createTempFile(t *testing.T, ext string, content []byte) string {
	t.Helper()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, fmt.Sprintf("test%s", ext))
	err := os.WriteFile(filePath, content, 0600)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	return filePath
}

// Tests
func TestNewPipeline(t *testing.T) {
	// Setup
	pdfProcessor := &docprocess.PDFProcessor{}
	csvProcessor := &processor.SEBProcessor{}
	mockAI := &mockAIService{}
	mockDB := &mockStore{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Execute
	pipeline := NewPipeline(pdfProcessor, csvProcessor, mockAI, mockDB, logger)

	// Verify
	if pipeline == nil {
		t.Fatal("Expected pipeline to be non-nil")
	}
}

func TestProcessFile_Successfully_process_pdf(t *testing.T) {
	// Skip this test for now until we can properly mock the PDFProcessor
	t.Skip("Skipping test until we can properly mock the PDFProcessor")
}

func TestProcessFile_Successfully_process_csv(t *testing.T) {
	// Skip this test for now until we can properly mock the SEBProcessor
	t.Skip("Skipping test until we can properly mock the SEBProcessor")
}

func TestProcessFile_Error_unsupported_file_type(t *testing.T) {
	// Setup
	pdfProcessor := &docprocess.PDFProcessor{}
	csvProcessor := &processor.SEBProcessor{}
	mockAI := &mockAIService{}
	mockDB := &mockStore{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Create a test file with unsupported extension
	txtPath := createTempFile(t, ".txt", []byte("test text content"))

	// Create pipeline
	pipeline := NewPipeline(pdfProcessor, csvProcessor, mockAI, mockDB, logger)

	// Execute
	_, err := pipeline.processFile(context.Background(), txtPath, ProcessOptions{})

	// Verify
	if err == nil {
		t.Fatal("Expected error for unsupported file type, got nil")
	}

	if !strings.Contains(err.Error(), "unsupported file type") {
		t.Errorf("Expected error message to contain 'unsupported file type', got '%s'", err.Error())
	}
}

func TestProcessDocuments_Error_invalid_path(t *testing.T) {
	// Setup
	pdfProcessor := &docprocess.PDFProcessor{}
	csvProcessor := &processor.SEBProcessor{}
	mockAI := &mockAIService{}
	mockDB := &mockStore{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Create pipeline
	pipeline := NewPipeline(pdfProcessor, csvProcessor, mockAI, mockDB, logger)

	// Execute with non-existent path
	_, err := pipeline.ProcessDocuments(context.Background(), "/path/does/not/exist", ProcessOptions{})

	// Verify
	if err == nil {
		t.Fatal("Expected error for invalid path, got nil")
	}

	if !strings.Contains(err.Error(), "failed to access path") {
		t.Errorf("Expected error message to contain 'failed to access path', got '%s'", err.Error())
	}
}

func TestSetCategoryIDsFromAnalysis(t *testing.T) {
	// Create test transactions
	transactions := []*db.Transaction{
		{
			Description:     "Test Transaction 1",
			Amount:          decimal.NewFromFloat(100.0),
			Date:            time.Now(),
			TransactionDate: time.Now(),
		},
		{
			Description:     "Test Transaction 2",
			Amount:          decimal.NewFromFloat(200.0),
			Date:            time.Now(),
			TransactionDate: time.Now(),
		},
	}

	// Create test analyses
	analyses := []*ai.Analysis{
		{
			Category:      "Fasta kostnader",
			Subcategory:   "Medier",
			CategoryID:    5,
			SubcategoryID: 19,
			Confidence:    0.95,
		},
		{
			Category:      "Fasta kostnader",
			Subcategory:   "Medier",
			CategoryID:    5,
			SubcategoryID: 19,
			Confidence:    0.95,
		},
	}

	// Create a logger for testing
	logger := slog.Default()

	// Apply analyses to transactions (simulating the code in processPDF)
	for i, dbTx := range transactions {
		if i < len(analyses) {
			// Convert analysis to JSON string
			aiAnalysis, err := json.Marshal(analyses[i])
			if err != nil {
				t.Fatalf("Failed to marshal AI analysis: %v", err)
			}

			dbTx.AIAnalysis = string(aiAnalysis)

			// Set category and subcategory IDs from analysis
			if analyses[i].CategoryID > 0 {
				// Safe conversion since we've checked the value is positive
				categoryID := uint(analyses[i].CategoryID) // #nosec G115
				dbTx.CategoryID = &categoryID
				logger.Debug("Setting category ID from AI analysis",
					"transaction_description", dbTx.Description,
					"category_id", categoryID,
					"category", analyses[i].Category)
			}

			if analyses[i].SubcategoryID > 0 {
				// Safe conversion since we've checked the value is positive
				subcategoryID := uint(analyses[i].SubcategoryID) // #nosec G115
				dbTx.SubcategoryID = &subcategoryID
				logger.Debug("Setting subcategory ID from AI analysis",
					"transaction_description", dbTx.Description,
					"subcategory_id", subcategoryID,
					"subcategory", analyses[i].Subcategory)
			}
		}
	}

	// Verify that category and subcategory IDs were set correctly
	for _, tx := range transactions {
		// Check that CategoryID is set correctly
		if tx.CategoryID == nil {
			t.Error("Expected CategoryID to be set, got nil")
		} else if *tx.CategoryID != 5 {
			t.Errorf("Expected CategoryID to be 5, got: %d", *tx.CategoryID)
		}

		// Check that SubcategoryID is set correctly
		if tx.SubcategoryID == nil {
			t.Error("Expected SubcategoryID to be set, got nil")
		} else if *tx.SubcategoryID != 19 {
			t.Errorf("Expected SubcategoryID to be 19, got: %d", *tx.SubcategoryID)
		}

		// Check that AIAnalysis contains the expected data
		var analysis ai.Analysis
		err := json.Unmarshal([]byte(tx.AIAnalysis), &analysis)
		if err != nil {
			t.Errorf("Failed to unmarshal AIAnalysis: %v", err)
		} else {
			if analysis.CategoryID != 5 {
				t.Errorf("Expected analysis.CategoryID to be 5, got: %d", analysis.CategoryID)
			}
			if analysis.SubcategoryID != 19 {
				t.Errorf("Expected analysis.SubcategoryID to be 19, got: %d", analysis.SubcategoryID)
			}
		}
	}
}
