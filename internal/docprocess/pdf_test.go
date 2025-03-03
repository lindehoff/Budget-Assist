package docprocess

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/lindehoff/Budget-Assist/internal/ai"
	"github.com/lindehoff/Budget-Assist/internal/db"
	"github.com/shopspring/decimal"
)

// mockAIService implements ai.Service for testing
type mockAIService struct {
	extractDocumentFunc func(ctx context.Context, doc *ai.Document) (*ai.Extraction, error)
}

func (m *mockAIService) ExtractDocument(ctx context.Context, doc *ai.Document) (*ai.Extraction, error) {
	if m.extractDocumentFunc != nil {
		return m.extractDocumentFunc(ctx, doc)
	}
	return nil, nil
}

func (m *mockAIService) AnalyzeTransaction(ctx context.Context, tx *db.Transaction, opts ai.AnalysisOptions) (*ai.Analysis, error) {
	return nil, nil
}

func (m *mockAIService) SuggestCategories(ctx context.Context, description string) ([]ai.CategoryMatch, error) {
	return nil, nil
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
		t.Errorf("NewDefaultProcessorFactory() = nil, want non-nil")
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
			result, err := processor.extractDocumentWithAI(context.Background(), "test content")

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
			},
			wantCount: 1,
		},
		{
			name: "Successfully_convert_multiple_transactions",
			extraction: &ai.Extraction{
				Date:        "2023-01-01",
				Amount:      300.0,
				Currency:    "SEK",
				Description: "Transaction 1, Transaction 2, Transaction 3",
			},
			wantCount: 3,
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
				// For single transaction case
				if tt.wantCount == 1 && !strings.Contains(tt.extraction.Description, ",") {
					if transactions[0].Description != tt.extraction.Description {
						t.Errorf("PDFProcessor.convertExtractionToTransactions() transaction.Description = %v, want %v",
							transactions[0].Description, tt.extraction.Description)
					}
					if !transactions[0].Amount.Equal(decimal.NewFromFloat(tt.extraction.Amount)) {
						t.Errorf("PDFProcessor.convertExtractionToTransactions() transaction.Amount = %v, want %v",
							transactions[0].Amount, decimal.NewFromFloat(tt.extraction.Amount))
					}
				}
			}
		})
	}
}

func TestPDFProcessor_parseTransactionFromPart(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	aiService := &mockAIService{}
	processor := NewPDFProcessor(logger, aiService)

	extraction := &ai.Extraction{
		Date:        "2023-01-01",
		Amount:      100.0,
		Currency:    "SEK",
		Description: "Test transaction",
	}

	tests := []struct {
		name   string
		part   string
		wantTx bool
	}{
		{
			name:   "Successfully_parse_transaction_with_amount",
			part:   "Test transaction (100.0 SEK)",
			wantTx: true,
		},
		{
			name:   "Successfully_parse_transaction_without_amount",
			part:   "Test transaction",
			wantTx: true,
		},
		{
			name:   "Successfully_handle_empty_part",
			part:   "",
			wantTx: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the method
			tx, ok := processor.parseTransactionFromPart(tt.part, extraction)

			if ok != tt.wantTx {
				t.Errorf("PDFProcessor.parseTransactionFromPart() ok = %v, want %v", ok, tt.wantTx)
				return
			}

			if tt.wantTx {
				if tx.Description == "" {
					t.Errorf("PDFProcessor.parseTransactionFromPart() tx.Description is empty, want non-empty")
				}

				// If the part contains an amount in SEK format
				if strings.Contains(tt.part, " (") && strings.HasSuffix(tt.part, " SEK)") {
					if tx.Amount.Equal(decimal.NewFromFloat(0)) {
						t.Errorf("PDFProcessor.parseTransactionFromPart() tx.Amount = 0, want non-zero")
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
