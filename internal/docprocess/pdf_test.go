package docprocess

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

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
		// TODO: Add test with valid PDF once we have test files
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
	// TODO: Use context.TODO() until we implement proper context handling
	ctx := context.TODO() // TODO: Implement proper context handling with timeouts and cancellation

	type testCase struct {
		checkResult func(*testing.T, *ProcessingResult)
		name        string
		errStage    ProcessingStage
		filename    string
		input       []byte
		wantErr     bool
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
		// TODO: Add test with valid PDF once we have test files
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

			if result.ProcessedAt.IsZero() {
				t.Error("ProcessingResult.ProcessedAt is zero")
			}

			if result.Metadata == nil {
				t.Error("ProcessingResult.Metadata is nil")
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

// TODO: Add integration tests with real PDF files
func TestPDFProcessor_Integration(t *testing.T) {
	t.Skip("TODO: Implement integration tests with real PDF files")
}
