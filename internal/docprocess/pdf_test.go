package docprocess

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"

	"github.com/lindehoff/Budget-Assist/internal/ai"
	"github.com/lindehoff/Budget-Assist/internal/db"
	"github.com/stretchr/testify/assert"
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
			name:     "Successfully_validate_empty_file_returns_error",
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
			name:    "Successfully_validate_valid_pdf",
			input:   createTestPDF(t),
			wantErr: false,
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
			name:     "valid PDF",
			file:     bytes.NewReader(pdfContent),
			filename: "test.pdf",
			wantErr:  false,
		},
		{
			name:     "empty PDF",
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
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result.Transactions, 1)
				assert.Equal(t, "Test transaction", result.Transactions[0].Description)
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
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "AI service is not configured")
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
			name:     "PDF file",
			filename: "test.pdf",
			want:     true,
		},
		{
			name:     "non-PDF file",
			filename: "test.txt",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPDFProcessor(logger, aiService)
			got := p.CanProcess(tt.filename)
			assert.Equal(t, tt.want, got)
		})
	}
}
