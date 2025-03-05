package docprocess

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lindehoff/Budget-Assist/internal/ai"
	"github.com/lindehoff/Budget-Assist/internal/db"
)

// mockAIService implements ai.Service for testing
type mockAIService struct {
	extractDocumentFunc          func(ctx context.Context, doc *ai.Document) (*ai.Extraction, error)
	analyzeTransactionFunc       func(ctx context.Context, tx *db.Transaction, opts ai.AnalysisOptions) (*ai.Analysis, error)
	suggestCategoriesFunc        func(ctx context.Context, description string) ([]ai.CategoryMatch, error)
	batchAnalyzeTransactionsFunc func(ctx context.Context, transactions []*db.Transaction, opts ai.AnalysisOptions) ([]*ai.Analysis, error)
}

func (m *mockAIService) ExtractDocument(ctx context.Context, doc *ai.Document) (*ai.Extraction, error) {
	if m.extractDocumentFunc != nil {
		return m.extractDocumentFunc(ctx, doc)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockAIService) AnalyzeTransaction(ctx context.Context, tx *db.Transaction, opts ai.AnalysisOptions) (*ai.Analysis, error) {
	if m.analyzeTransactionFunc != nil {
		return m.analyzeTransactionFunc(ctx, tx, opts)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockAIService) SuggestCategories(ctx context.Context, description string) ([]ai.CategoryMatch, error) {
	if m.suggestCategoriesFunc != nil {
		return m.suggestCategoriesFunc(ctx, description)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockAIService) BatchAnalyzeTransactions(ctx context.Context, transactions []*db.Transaction, opts ai.AnalysisOptions) ([]*ai.Analysis, error) {
	if m.batchAnalyzeTransactionsFunc != nil {
		return m.batchAnalyzeTransactionsFunc(ctx, transactions, opts)
	}
	return nil, fmt.Errorf("not implemented")
}

// createTestPDF creates a simple PDF file for testing
func createTestPDF(t *testing.T) []byte {
	t.Helper()

	// For testing purposes, we'll use a pre-generated minimal valid PDF
	// This is a minimal valid PDF file that contains just enough to be valid
	return []byte(`%PDF-1.7
1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj
2 0 obj<</Type/Pages/Count 1/Kids[3 0 R]>>endobj
3 0 obj<</Type/Page/Parent 2 0 R/MediaBox[0 0 612 792]/Contents 4 0 R>>endobj
4 0 obj<</Length 50>>stream
BT /F1 12 Tf 72 712 Td (Test PDF Document) Tj ET
endstream endobj
xref
0 5
0000000000 65535 f
0000000010 00000 n
0000000053 00000 n
0000000102 00000 n
0000000177 00000 n
trailer<</Size 5/Root 1 0 R>>
startxref
277
%%EOF`)
}

// createCorruptedPDF creates a corrupted PDF file for testing
func createCorruptedPDF(t *testing.T) []byte {
	t.Helper()

	// Create a corrupted PDF by removing essential parts
	return []byte(`%PDF-1.7
1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj
2 0 obj<</Type/Pages/Count 1/Kids[3 0 R]>>endobj
3 0 obj<</Type/Page/Parent 2 0 R/MediaBox[0 0 612 792]/Contents 4 0 R>>endobj
4 0 obj<</Length 50>>stream
BT /F1 12 Tf 72 712 Td (Test PDF Document) Tj ET
endstream endobj
xref
0 5
0000000000 65535 f
0000000010 00000 n
0000000053 00000 n
0000000102 00000 n
trailer<</Size 5/Root 1 0 R>>
%%EOF`)
}

func TestPDFProcessor_Type(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	aiService := &mockAIService{}
	processor := NewPDFProcessor(logger, aiService)
	if got := processor.Type(); got != TypePDF {
		t.Errorf("PDFProcessor.Type() = %v, want %v", got, TypePDF)
	}
}

func TestPDFProcessor_Validate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	aiService := &mockAIService{}

	type testCase struct {
		name     string
		errStage ProcessingStage
		input    []byte
		wantErr  bool
	}

	tests := []testCase{
		{
			name:    "Successfully_validate_valid_pdf",
			input:   createTestPDF(t),
			wantErr: false,
		},
		{
			name:     "Validate_error_empty_file",
			input:    []byte{},
			wantErr:  true,
			errStage: StageValidation,
		},
		{
			name:     "Validate_error_invalid_pdf_content",
			input:    []byte("not a PDF"),
			wantErr:  true,
			errStage: StageValidation,
		},
		{
			name:     "Validate_error_corrupted_pdf",
			input:    createCorruptedPDF(t),
			wantErr:  true,
			errStage: StageValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewPDFProcessor(logger, aiService)
			err := processor.Validate(bytes.NewReader(tt.input))

			if (err != nil) != tt.wantErr {
				t.Errorf("PDFProcessor.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				var procErr *ProcessingError
				if !errors.As(err, &procErr) {
					t.Errorf("PDFProcessor.Validate() error is not ProcessingError, got %T", err)
					return
				}
				if procErr.Stage != tt.errStage {
					t.Errorf("ProcessingError.Stage = %v, want %v", procErr.Stage, tt.errStage)
				}
			}
		})
	}
}

func TestPDFProcessor_Process(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Create a mock AI service
	aiService := &mockAIService{
		extractDocumentFunc: func(ctx context.Context, doc *ai.Document) (*ai.Extraction, error) {
			return &ai.Extraction{
				Date:        "2023-01-01",
				Amount:      100.0,
				Currency:    "SEK",
				Description: "Test transaction",
			}, nil
		},
	}

	// Create a test PDF file
	pdfContent := createTestPDF(t)

	tests := []struct {
		name     string
		file     io.Reader
		filename string
		wantErr  bool
	}{
		{
			name:     "Successfully_process_valid_pdf",
			file:     bytes.NewReader(pdfContent),
			filename: "test.pdf",
			wantErr:  false,
		},
		{
			name:     "Process_error_empty_pdf",
			file:     bytes.NewReader([]byte{}),
			filename: "empty.pdf",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a processor with a mocked extractTextFromPDF method
			p := NewPDFProcessor(logger, aiService)

			// Replace the extractTextFromPDF method with a mock
			originalExtractTextFromPDF := p.extractTextFromPDF
			p.extractTextFromPDF = func(ctx context.Context, file io.Reader) (string, error) {
				if tt.wantErr {
					return "", fmt.Errorf("mock error")
				}
				return "This is a test PDF content", nil
			}

			// Restore the original method after the test
			defer func() {
				p.extractTextFromPDF = originalExtractTextFromPDF
			}()

			result, err := p.Process(context.Background(), tt.file, tt.filename)
			if tt.wantErr {
				if err == nil {
					t.Errorf("PDFProcessor.Process() error = nil, wantErr %v", tt.wantErr)
				}
				if result != nil {
					t.Errorf("PDFProcessor.Process() result = %v, want nil", result)
				}
			} else {
				if err != nil {
					t.Errorf("PDFProcessor.Process() error = %v, wantErr %v", err, tt.wantErr)
				}
				if result == nil {
					t.Errorf("PDFProcessor.Process() result = nil, want non-nil")
					return
				}
				if len(result.Transactions) != 1 {
					t.Errorf("PDFProcessor.Process() result.Transactions length = %d, want 1", len(result.Transactions))
					return
				}
				if result.Transactions[0].Description != "Test transaction" {
					t.Errorf("PDFProcessor.Process() result.Transactions[0].Description = %s, want %s",
						result.Transactions[0].Description, "Test transaction")
				}
			}
		})
	}
}

func TestPDFProcessor_Process_NoAIService(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	p := NewPDFProcessor(logger, nil)

	// Create a test PDF file
	pdfContent := createTestPDF(t)

	// Mock the extractTextFromPDF method
	p.extractTextFromPDF = func(ctx context.Context, file io.Reader) (string, error) {
		return "This is a test PDF content", nil
	}

	// Test with nil AI service
	result, err := p.Process(context.Background(), bytes.NewReader(pdfContent), "test.pdf")
	if err == nil {
		t.Errorf("PDFProcessor.Process() error = nil, want error")
	}
	if result != nil {
		t.Errorf("PDFProcessor.Process() result = %v, want nil", result)
	}
	if err != nil && !strings.Contains(err.Error(), "AI service is not configured") {
		t.Errorf("PDFProcessor.Process() error = %v, want error containing 'AI service is not configured'", err)
	}
}

func TestPDFProcessor_CanProcess(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	aiService := &mockAIService{}

	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{
			name:     "Successfully_identify_pdf_file",
			filename: "test.pdf",
			want:     true,
		},
		{
			name:     "Successfully_identify_non_pdf_file",
			filename: "test.txt",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPDFProcessor(logger, aiService)
			got := p.CanProcess(tt.filename)
			if got != tt.want {
				t.Errorf("PDFProcessor.CanProcess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultProcessorFactory_NewDefaultProcessorFactory(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	aiService := &mockAIService{}

	factory := NewDefaultProcessorFactory(logger, aiService)
	if factory == nil {
		t.Fatalf("NewDefaultProcessorFactory() = nil, want non-nil")
		return
	}

	if factory.logger != logger {
		t.Errorf("factory.logger = %v, want %v", factory.logger, logger)
	}
	if factory.aiService != aiService {
		t.Errorf("factory.aiService = %v, want %v", factory.aiService, aiService)
	}
}

func TestDefaultProcessorFactory_CreateProcessor(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	aiService := &mockAIService{}
	factory := NewDefaultProcessorFactory(logger, aiService)

	tests := []struct {
		name    string
		docType DocumentType
		wantErr bool
	}{
		{
			name:    "Successfully_create_pdf_processor",
			docType: TypePDF,
			wantErr: false,
		},
		{
			name:    "CreateProcessor_error_unsupported_type",
			docType: "unsupported",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor, err := factory.CreateProcessor(tt.docType)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DefaultProcessorFactory.CreateProcessor() error = nil, wantErr %v", tt.wantErr)
				}
				if processor != nil {
					t.Errorf("DefaultProcessorFactory.CreateProcessor() processor = %v, want nil", processor)
				}
				if err != nil && !strings.Contains(err.Error(), "unsupported document type") {
					t.Errorf("DefaultProcessorFactory.CreateProcessor() error = %v, want error containing 'unsupported document type'", err)
				}
			} else {
				if err != nil {
					t.Errorf("DefaultProcessorFactory.CreateProcessor() error = %v, wantErr %v", err, tt.wantErr)
				}
				if processor == nil {
					t.Errorf("DefaultProcessorFactory.CreateProcessor() processor = nil, want non-nil")
					return
				}

				// Check if it's the right type of processor
				_, ok := processor.(*PDFProcessor)
				if !ok {
					t.Errorf("DefaultProcessorFactory.CreateProcessor() processor type = %T, want *PDFProcessor", processor)
				}
			}
		})
	}
}

func TestDefaultProcessorFactory_SupportedTypes(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	aiService := &mockAIService{}
	factory := NewDefaultProcessorFactory(logger, aiService)

	types := factory.SupportedTypes()
	if len(types) != 1 {
		t.Errorf("DefaultProcessorFactory.SupportedTypes() length = %d, want 1", len(types))
	}
	if len(types) > 0 && types[0] != TypePDF {
		t.Errorf("DefaultProcessorFactory.SupportedTypes()[0] = %v, want %v", types[0], TypePDF)
	}
}

func TestPDFProcessor_extractDocumentWithAI(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name           string
		aiResponse     *ai.Extraction
		aiError        error
		expectedResult *ai.Extraction
		wantErr        bool
	}{
		{
			name: "Successfully_extract_document_with_ai",
			aiResponse: &ai.Extraction{
				Date:        "2023-01-01",
				Amount:      100.0,
				Currency:    "SEK",
				Description: "Test transaction",
			},
			aiError: nil,
			expectedResult: &ai.Extraction{
				Date:        "2023-01-01",
				Amount:      100.0,
				Currency:    "SEK",
				Description: "Test transaction",
			},
			wantErr: false,
		},
		{
			name:           "ExtractDocumentWithAI_error_ai_service_error",
			aiResponse:     nil,
			aiError:        fmt.Errorf("AI service error"),
			expectedResult: nil,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock AI service with the desired behavior
			aiService := &mockAIService{
				extractDocumentFunc: func(ctx context.Context, doc *ai.Document) (*ai.Extraction, error) {
					return tt.aiResponse, tt.aiError
				},
			}

			processor := NewPDFProcessor(logger, aiService)

			// Call the method
			result, err := processor.extractDocumentWithAI(context.Background(), "test content", ProcessOptions{})

			if tt.wantErr {
				if err == nil {
					t.Errorf("PDFProcessor.extractDocumentWithAI() error = nil, wantErr %v", tt.wantErr)
				}
				if result != nil {
					t.Errorf("PDFProcessor.extractDocumentWithAI() result = %v, want nil", result)
				}
			} else {
				if err != nil {
					t.Errorf("PDFProcessor.extractDocumentWithAI() error = %v, wantErr %v", err, tt.wantErr)
				}
				if result == nil {
					t.Errorf("PDFProcessor.extractDocumentWithAI() result = nil, want non-nil")
					return
				}
				if result.Date != tt.expectedResult.Date {
					t.Errorf("PDFProcessor.extractDocumentWithAI() result.Date = %v, want %v",
						result.Date, tt.expectedResult.Date)
				}
				if result.Amount != tt.expectedResult.Amount {
					t.Errorf("PDFProcessor.extractDocumentWithAI() result.Amount = %v, want %v",
						result.Amount, tt.expectedResult.Amount)
				}
				if result.Currency != tt.expectedResult.Currency {
					t.Errorf("PDFProcessor.extractDocumentWithAI() result.Currency = %v, want %v",
						result.Currency, tt.expectedResult.Currency)
				}
				if result.Description != tt.expectedResult.Description {
					t.Errorf("PDFProcessor.extractDocumentWithAI() result.Description = %v, want %v",
						result.Description, tt.expectedResult.Description)
				}
			}
		})
	}
}

func TestPDFProcessor_convertExtractionToTransactions(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	aiService := &mockAIService{}
	processor := NewPDFProcessor(logger, aiService)

	tests := []struct {
		name       string
		extraction *ai.Extraction
		wantCount  int
	}{
		{
			name: "Successfully_convert_single_transaction",
			extraction: &ai.Extraction{
				Date:        "2023-01-01",
				Amount:      100.0,
				Currency:    "SEK",
				Description: "Test transaction",
				Transactions: []map[string]interface{}{
					{
						"date":        "2023-01-01",
						"amount":      100.0,
						"description": "Test transaction",
					},
				},
			},
			wantCount: 1,
		},
		{
			name: "Successfully_convert_multiple_transactions",
			extraction: &ai.Extraction{
				Date:        "2023-01-01",
				Amount:      300.0,
				Currency:    "SEK",
				Description: "Test transaction",
				Transactions: []map[string]interface{}{
					{
						"date":        "2023-01-01",
						"amount":      100.0,
						"description": "Transaction 1",
					},
					{
						"date":        "2023-01-01",
						"amount":      100.0,
						"description": "Transaction 2",
					},
					{
						"date":        "2023-01-01",
						"amount":      100.0,
						"description": "Transaction 3",
					},
				},
			},
			wantCount: 3,
		},
		{
			name: "Successfully_convert_with_fallback",
			extraction: &ai.Extraction{
				Date:         "2023-01-01",
				Amount:       100.0,
				Currency:     "SEK",
				Description:  "Test transaction",
				Transactions: []map[string]interface{}{},
			},
			wantCount: 1,
		},
		{
			name:       "Successfully_handle_nil_extraction",
			extraction: nil,
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the method
			transactions := processor.convertExtractionToTransactions(tt.extraction)

			if len(transactions) != tt.wantCount {
				t.Errorf("PDFProcessor.convertExtractionToTransactions() length = %d, want %d",
					len(transactions), tt.wantCount)
				return
			}

			// Check transaction fields
			if tt.wantCount > 0 && tt.extraction != nil {
				// For transactions from the Transactions field
				if len(tt.extraction.Transactions) > 0 {
					for i, tx := range transactions {
						if i < len(tt.extraction.Transactions) {
							expectedDesc, _ := tt.extraction.Transactions[i]["description"].(string)
							if tx.Description != expectedDesc {
								t.Errorf("Transaction[%d].Description = %s, want %s",
									i, tx.Description, expectedDesc)
							}
						}
					}
				} else if tt.wantCount == 1 {
					// For fallback single transaction
					if transactions[0].Description != tt.extraction.Description {
						t.Errorf("Transaction.Description = %s, want %s",
							transactions[0].Description, tt.extraction.Description)
					}
				}
			}
		})
	}
}

func TestPDFProcessor_extractionToMap(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	aiService := &mockAIService{}
	processor := NewPDFProcessor(logger, aiService)

	tests := []struct {
		name       string
		extraction *ai.Extraction
		wantEmpty  bool
	}{
		{
			name: "Successfully_convert_extraction_to_map",
			extraction: &ai.Extraction{
				Date:        "2023-01-01",
				Amount:      100.0,
				Currency:    "SEK",
				Description: "Test transaction",
			},
			wantEmpty: false,
		},
		{
			name:       "Successfully_handle_nil_extraction",
			extraction: nil,
			wantEmpty:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the method
			result := processor.extractionToMap(tt.extraction)

			if tt.wantEmpty {
				if len(result) != 0 {
					t.Errorf("PDFProcessor.extractionToMap() result length = %d, want 0", len(result))
				}
			} else {
				if len(result) == 0 {
					t.Errorf("PDFProcessor.extractionToMap() result is empty, want non-empty")
					return
				}

				// Check if the map contains the expected fields
				if tt.extraction != nil {
					if result["date"] != tt.extraction.Date {
						t.Errorf("PDFProcessor.extractionToMap() result[\"date\"] = %v, want %v",
							result["date"], tt.extraction.Date)
					}
					if result["amount"] != tt.extraction.Amount {
						t.Errorf("PDFProcessor.extractionToMap() result[\"amount\"] = %v, want %v",
							result["amount"], tt.extraction.Amount)
					}
					if result["currency"] != tt.extraction.Currency {
						t.Errorf("PDFProcessor.extractionToMap() result[\"currency\"] = %v, want %v",
							result["currency"], tt.extraction.Currency)
					}
					if result["description"] != tt.extraction.Description {
						t.Errorf("PDFProcessor.extractionToMap() result[\"description\"] = %v, want %v",
							result["description"], tt.extraction.Description)
					}
				}
			}
		})
	}
}

// mockExtractTextFromPDF is a test helper that creates a mock implementation of extractTextFromPDF
func mockExtractTextFromPDF(t *testing.T, expectedContent string, expectedErr error) func(ctx context.Context, file io.Reader) (string, error) {
	t.Helper()
	return func(ctx context.Context, file io.Reader) (string, error) {
		if expectedErr != nil {
			return "", expectedErr
		}
		return expectedContent, nil
	}
}

func TestPDFProcessor_defaultExtractTextFromPDF(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	aiService := &mockAIService{}
	processor := NewPDFProcessor(logger, aiService)

	// Create a temporary directory for test files
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		setupMock      bool
		mockContent    string
		mockErr        error
		pdfContent     []byte
		wantErr        bool
		expectedErrMsg string
	}{
		{
			name:        "Successfully_extract_text_from_valid_pdf",
			setupMock:   true,
			mockContent: "This is extracted text from a PDF",
			mockErr:     nil,
			pdfContent:  createTestPDF(t),
			wantErr:     false,
		},
		{
			name:           "Extract_error_empty_pdf",
			setupMock:      true,
			mockContent:    "",
			mockErr:        fmt.Errorf("failed to extract text from PDF: empty file"),
			pdfContent:     []byte{},
			wantErr:        true,
			expectedErrMsg: "empty file",
		},
		{
			name:           "Extract_error_invalid_pdf",
			setupMock:      true,
			mockContent:    "",
			mockErr:        fmt.Errorf("failed to extract text from PDF: invalid format"),
			pdfContent:     []byte("not a PDF"),
			wantErr:        true,
			expectedErrMsg: "invalid format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary PDF file
			pdfPath := filepath.Join(tempDir, fmt.Sprintf("%s.pdf", tt.name))
			err := os.WriteFile(pdfPath, tt.pdfContent, 0600)
			if err != nil {
				t.Fatalf("Failed to create test PDF file: %v", err)
			}

			// Open the file
			file, err := os.Open(pdfPath)
			if err != nil {
				t.Fatalf("Failed to open test PDF file: %v", err)
			}
			defer file.Close()

			// Save the original method
			originalMethod := processor.extractTextFromPDF

			// Setup mock if needed
			if tt.setupMock {
				// Use our mock
				processor.extractTextFromPDF = mockExtractTextFromPDF(t, tt.mockContent, tt.mockErr)
			}

			// Call the method
			text, err := processor.extractTextFromPDF(context.Background(), file)

			// Restore the original method
			processor.extractTextFromPDF = originalMethod

			if tt.wantErr {
				if err == nil {
					t.Errorf("PDFProcessor.extractTextFromPDF() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.expectedErrMsg != "" && !strings.Contains(err.Error(), tt.expectedErrMsg) {
					t.Errorf("PDFProcessor.extractTextFromPDF() error = %v, want error containing %q",
						err, tt.expectedErrMsg)
				}
			} else {
				if err != nil {
					t.Errorf("PDFProcessor.extractTextFromPDF() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if text == "" {
					t.Errorf("PDFProcessor.extractTextFromPDF() text is empty, want non-empty")
				}
				if text != tt.mockContent {
					t.Errorf("PDFProcessor.extractTextFromPDF() text = %q, want %q", text, tt.mockContent)
				}
			}
		})
	}
}

// TestPDFProcessor_defaultExtractTextFromPDF_NoTool tests the case where pdftotext is not installed
func TestPDFProcessor_defaultExtractTextFromPDF_NoTool(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	aiService := &mockAIService{}
	processor := NewPDFProcessor(logger, aiService)

	// Create a custom implementation of defaultExtractTextFromPDF that simulates pdftotext not being installed
	customExtractTextFromPDF := func(ctx context.Context, file io.Reader) (string, error) {
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

		// Simulate pdftotext not being installed
		return "", fmt.Errorf("pdftotext is not installed. Please install poppler-utils")
	}

	// Save the original method and replace it with our custom implementation
	originalMethod := processor.extractTextFromPDF
	processor.extractTextFromPDF = customExtractTextFromPDF
	defer func() { processor.extractTextFromPDF = originalMethod }()

	// Create a test PDF file
	tempDir := t.TempDir()
	pdfPath := filepath.Join(tempDir, "test.pdf")
	err := os.WriteFile(pdfPath, createTestPDF(t), 0600)
	if err != nil {
		t.Fatalf("Failed to create test PDF file: %v", err)
	}

	// Open the file
	file, err := os.Open(pdfPath)
	if err != nil {
		t.Fatalf("Failed to open test PDF file: %v", err)
	}
	defer file.Close()

	// Call the method
	_, err = processor.extractTextFromPDF(context.Background(), file)

	// Check that we got an error
	if err == nil {
		t.Errorf("PDFProcessor.extractTextFromPDF() error = nil, want error about pdftotext not installed")
	} else if !strings.Contains(err.Error(), "pdftotext is not installed") {
		t.Errorf("PDFProcessor.extractTextFromPDF() error = %v, want error containing 'pdftotext is not installed'", err)
	}
}
