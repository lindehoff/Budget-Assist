package docprocess

import (
	"errors"
	"fmt"
	"testing"
)

func TestProcessingError_Error(t *testing.T) {
	tests := []struct {
		name       string
		err        ProcessingError
		wantString string
	}{
		{
			name: "Successfully_format_validation_error",
			err: ProcessingError{
				Stage:    StageValidation,
				Document: "test.pdf",
				Err:      fmt.Errorf("invalid format"),
			},
			wantString: "validation failed for document \"test.pdf\": invalid format",
		},
		{
			name: "Successfully_format_extraction_error",
			err: ProcessingError{
				Stage:    StageExtraction,
				Document: "test.pdf",
				Err:      fmt.Errorf("extraction failed"),
			},
			wantString: "extraction failed for document \"test.pdf\": extraction failed",
		},
		{
			name: "Successfully_format_normalization_error",
			err: ProcessingError{
				Stage:    StageNormalization,
				Document: "test.pdf",
				Err:      fmt.Errorf("normalization failed"),
			},
			wantString: "normalization failed for document \"test.pdf\": normalization failed",
		},
		{
			name: "Successfully_format_analysis_error",
			err: ProcessingError{
				Stage:    StageAnalysis,
				Document: "test.pdf",
				Err:      fmt.Errorf("analysis failed"),
			},
			wantString: "analysis failed for document \"test.pdf\": analysis failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantString {
				t.Errorf("ProcessingError.Error() = %v, want %v", got, tt.wantString)
			}
		})
	}
}

func TestProcessingError_Unwrap(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr error
	}{
		{
			name: "Successfully_unwrap_nested_error",
			err: &ProcessingError{
				Stage:    StageValidation,
				Document: "test.pdf",
				Err:      fmt.Errorf("original error"),
			},
			wantErr: fmt.Errorf("original error"),
		},
		{
			name: "Successfully_unwrap_nil_error",
			err: &ProcessingError{
				Stage:    StageValidation,
				Document: "test.pdf",
				Err:      nil,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var procErr *ProcessingError
			if !errors.As(tt.err, &procErr) {
				t.Fatal("error is not a ProcessingError")
			}

			unwrapped := procErr.Unwrap()
			if tt.wantErr == nil {
				if unwrapped != nil {
					t.Errorf("ProcessingError.Unwrap() = %v, want nil", unwrapped)
				}
			} else {
				if unwrapped == nil {
					t.Error("ProcessingError.Unwrap() = nil, want error")
				} else if unwrapped.Error() != tt.wantErr.Error() {
					t.Errorf("ProcessingError.Unwrap().Error() = %v, want %v", unwrapped, tt.wantErr)
				}
			}
		})
	}
}

func TestDocumentType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		docType  DocumentType
		wantType string
	}{
		{"PDF_type", TypePDF, "pdf"},
		{"CSV_type", TypeCSV, "csv"},
		{"XLSX_type", TypeXLSX, "xlsx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.docType) != tt.wantType {
				t.Errorf("DocumentType %v = %v, want %v", tt.name, tt.docType, tt.wantType)
			}
		})
	}
}

func TestProcessingStage_Constants(t *testing.T) {
	tests := []struct {
		name      string
		stage     ProcessingStage
		wantStage string
	}{
		{"Validation_stage", StageValidation, "validation"},
		{"Extraction_stage", StageExtraction, "extraction"},
		{"Normalization_stage", StageNormalization, "normalization"},
		{"Analysis_stage", StageAnalysis, "analysis"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.stage) != tt.wantStage {
				t.Errorf("ProcessingStage %v = %v, want %v", tt.name, tt.stage, tt.wantStage)
			}
		})
	}
}
