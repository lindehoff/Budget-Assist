package docprocess

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

// createTestPDF creates a minimal valid PDF file for testing
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

// createCorruptedPDF creates an invalid PDF file for testing
func createCorruptedPDF(t *testing.T) []byte {
	t.Helper()
	validPDF := createTestPDF(t)
	// Corrupt the PDF by modifying some bytes in the middle
	if len(validPDF) > 100 {
		copy(validPDF[50:], []byte("corrupted data"))
	}
	return validPDF
}

func TestPDFProcessor_Type(t *testing.T) {
	processor := NewPDFProcessor(nil)
	if got := processor.Type(); got != TypePDF {
		t.Errorf("PDFProcessor.Type() = %v, want %v", got, TypePDF)
	}
}

func TestPDFProcessor_Validate(t *testing.T) {
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
			processor := NewPDFProcessor(nil)
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
	ctx := context.Background()

	type testCase struct {
		name        string
		errStage    ProcessingStage
		filename    string
		input       []byte
		wantErr     bool
		checkResult func(*testing.T, *ProcessingResult)
	}

	tests := []testCase{
		{
			name:     "Process_error_empty_file",
			input:    []byte{},
			filename: "empty.pdf",
			wantErr:  true,
			errStage: StageExtraction,
		},
		{
			name:     "Process_error_invalid_pdf_content",
			input:    []byte("not a PDF"),
			filename: "invalid.pdf",
			wantErr:  true,
			errStage: StageExtraction,
		},
		{
			name:     "Successfully_process_valid_pdf",
			input:    createTestPDF(t),
			filename: "valid.pdf",
			wantErr:  false,
			checkResult: func(t *testing.T, result *ProcessingResult) {
				if result.Metadata["filename"] != "valid.pdf" {
					t.Errorf("ProcessingResult.Metadata[filename] = %v, want %v", result.Metadata["filename"], "valid.pdf")
				}
				if result.Metadata["content_type"] != "application/pdf" {
					t.Errorf("ProcessingResult.Metadata[content_type] = %v, want application/pdf", result.Metadata["content_type"])
				}
				if result.Metadata["page_count"].(int) != 1 {
					t.Errorf("ProcessingResult.Metadata[page_count] = %v, want 1", result.Metadata["page_count"])
				}
				if result.ProcessedAt.IsZero() {
					t.Error("ProcessingResult.ProcessedAt is zero")
				}
			},
		},
		{
			name:     "Process_error_corrupted_pdf",
			input:    createCorruptedPDF(t),
			filename: "corrupted.pdf",
			wantErr:  true,
			errStage: StageExtraction,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewPDFProcessor(nil)
			result, err := processor.Process(ctx, bytes.NewReader(tt.input), tt.filename)

			if (err != nil) != tt.wantErr {
				t.Errorf("PDFProcessor.Process() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				var procErr *ProcessingError
				if !errors.As(err, &procErr) {
					t.Errorf("PDFProcessor.Process() error is not ProcessingError, got %T", err)
					return
				}
				if procErr.Stage != tt.errStage {
					t.Errorf("ProcessingError.Stage = %v, want %v", procErr.Stage, tt.errStage)
				}
				return
			}

			if result == nil {
				t.Error("PDFProcessor.Process() returned nil result for success case")
				return
			}

			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func TestDefaultProcessorFactory(t *testing.T) {
	t.Run("Successfully_get_supported_types", func(t *testing.T) {
		factory := NewDefaultProcessorFactory(nil)
		types := factory.SupportedTypes()

		found := false
		for _, t := range types {
			if t == TypePDF {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("SupportedTypes() = %v, want to include %v", types, TypePDF)
		}
	})

	t.Run("Successfully_create_pdf_processor", func(t *testing.T) {
		factory := NewDefaultProcessorFactory(nil)
		processor, err := factory.CreateProcessor(TypePDF)

		if err != nil {
			t.Errorf("CreateProcessor(%v) unexpected error: %v", TypePDF, err)
			return
		}

		if processor == nil {
			t.Error("CreateProcessor(TypePDF) returned nil processor")
			return
		}

		if processor.Type() != TypePDF {
			t.Errorf("processor.Type() = %v, want %v", processor.Type(), TypePDF)
		}
	})

	t.Run("Create_error_unsupported_type", func(t *testing.T) {
		factory := NewDefaultProcessorFactory(nil)
		processor, err := factory.CreateProcessor("unsupported")

		if err == nil {
			t.Error("CreateProcessor(unsupported) expected error, got nil")
		}

		if processor != nil {
			t.Errorf("CreateProcessor(unsupported) = %v, want nil", processor)
		}
	})
}

func TestPDFProcessor_Integration(t *testing.T) {
	ctx := context.Background()
	processor := NewPDFProcessor(nil)

	// Create a test PDF with known content
	pdfData := createTestPDF(t)

	// Test validation
	if err := processor.Validate(bytes.NewReader(pdfData)); err != nil {
		t.Errorf("Validate() error = %v", err)
		return
	}

	// Test processing
	result, err := processor.Process(ctx, bytes.NewReader(pdfData), "test.pdf")
	if err != nil {
		t.Errorf("Process() error = %v", err)
		return
	}

	// Verify the processing result
	if result.ProcessedAt.IsZero() {
		t.Error("ProcessingResult.ProcessedAt is zero")
	}

	expectedMetadata := map[string]any{
		"filename":     "test.pdf",
		"content_type": "application/pdf",
		"page_count":   1,
	}

	for key, want := range expectedMetadata {
		if got := result.Metadata[key]; got != want {
			t.Errorf("ProcessingResult.Metadata[%q] = %v, want %v", key, got, want)
		}
	}

	// Verify text extraction
	if textLen := result.Metadata["text_length"].(int); textLen == 0 {
		t.Error("ProcessingResult.Metadata[text_length] is 0, expected some text to be extracted")
	}
}
