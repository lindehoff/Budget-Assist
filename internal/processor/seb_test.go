package processor

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestSEBProcessor_Successfully_process_valid_transactions(t *testing.T) {
	type testCase struct {
		name  string        // string, 16 bytes
		input string        // string, 16 bytes
		want  []Transaction // slice of pointers, 8-byte aligned
	}
	tests := []testCase{
		{
			name: "Successfully_process_multiple_transactions",
			input: `Bokföringsdatum;Valutadatum;Verifikationsnummer;Text;Belopp;Saldo
2025-02-24;2025-02-22;5490990004;56130086210 ;-1000.000;2814.160
2025-02-10;2025-02-10;5490990004;56130086210 ;-2000.000;3814.160`,
			want: []Transaction{
				{
					Date:        time.Date(2025, 2, 24, 0, 0, 0, 0, time.UTC),
					Amount:      decimal.NewFromFloat(-1000.000),
					Description: "56130086210",
					Reference:   "5490990004",
					Source:      "SEB",
					RawData: map[string]any{
						"ValueDate": "2025-02-22",
						"Balance":   "2814.160",
					},
				},
				{
					Date:        time.Date(2025, 2, 10, 0, 0, 0, 0, time.UTC),
					Amount:      decimal.NewFromFloat(-2000.000),
					Description: "56130086210",
					Reference:   "5490990004",
					Source:      "SEB",
					RawData: map[string]any{
						"ValueDate": "2025-02-10",
						"Balance":   "3814.160",
					},
				},
			},
		},
		{
			name: "Successfully_process_single_transaction",
			input: `Bokföringsdatum;Valutadatum;Verifikationsnummer;Text;Belopp;Saldo
2025-02-24;2025-02-22;5490990004;56130086210 ;-1000.000;2814.160`,
			want: []Transaction{
				{
					Date:        time.Date(2025, 2, 24, 0, 0, 0, 0, time.UTC),
					Amount:      decimal.NewFromFloat(-1000.000),
					Description: "56130086210",
					Reference:   "5490990004",
					Source:      "SEB",
					RawData: map[string]any{
						"ValueDate": "2025-02-22",
						"Balance":   "2814.160",
					},
				},
			},
		},
		{
			name: "Successfully_process_file_with_bom",
			input: "\uFEFFBokföringsdatum;Valutadatum;Verifikationsnummer;Text;Belopp;Saldo\n" +
				"2025-02-24;2025-02-22;5490990004;56130086210 ;-1000.000;2814.160",
			want: []Transaction{
				{
					Date:        time.Date(2025, 2, 24, 0, 0, 0, 0, time.UTC),
					Amount:      decimal.NewFromFloat(-1000.000),
					Description: "56130086210",
					Reference:   "5490990004",
					Source:      "SEB",
					RawData: map[string]any{
						"ValueDate": "2025-02-22",
						"Balance":   "2814.160",
					},
				},
			},
		},
	}

	// TODO: Replace context.TODO() with proper context handling for timeouts and cancellation
	ctx := context.TODO()
	logger := slog.Default()
	processor := NewSEBProcessor(logger)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Process the test data
			got, err := processor.ProcessDocument(ctx, strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Validate transaction count
			if len(got) != len(tt.want) {
				t.Fatalf("expected %d transactions, got %d", len(tt.want), len(got))
			}

			// Validate each transaction
			for i, want := range tt.want {
				// Validate Date
				if !want.Date.Equal(got[i].Date) {
					t.Errorf("transaction[%d].Date = %v, want %v", i, got[i].Date, want.Date)
					return
				}

				// Validate Amount
				if !want.Amount.Equal(got[i].Amount) {
					t.Errorf("transaction[%d].Amount = %v, want %v", i, got[i].Amount, want.Amount)
					return
				}

				// Validate Description
				if want.Description != got[i].Description {
					t.Errorf("transaction[%d].Description = %q, want %q", i, got[i].Description, want.Description)
					return
				}

				// Validate Reference
				if want.Reference != got[i].Reference {
					t.Errorf("transaction[%d].Reference = %q, want %q", i, got[i].Reference, want.Reference)
					return
				}

				// Validate Source
				if want.Source != got[i].Source {
					t.Errorf("transaction[%d].Source = %q, want %q", i, got[i].Source, want.Source)
					return
				}

				// Validate RawData
				for k, v := range want.RawData {
					if got[i].RawData[k] != v {
						t.Errorf("transaction[%d].RawData[%q] = %v, want %v", i, k, got[i].RawData[k], v)
						return
					}
				}
			}
		})
	}
}

func TestSEBProcessor_error_validation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name: "Process_error_invalid_header_format",
			input: `Invalid;Header;Format
2025-02-24;2025-02-22;5490990004;56130086210;-1000.000;2814.160`,
			wantErr: "validate_header failed",
		},
		{
			name: "Process_error_invalid_date_format",
			input: `Bokföringsdatum;Valutadatum;Verifikationsnummer;Text;Belopp;Saldo
invalid-date;2025-02-22;5490990004;56130086210;-1000.000;2814.160`,
			wantErr: "no valid transactions found",
		},
		{
			name: "Process_error_invalid_amount_format",
			input: `Bokföringsdatum;Valutadatum;Verifikationsnummer;Text;Belopp;Saldo
2025-02-24;2025-02-22;5490990004;56130086210;invalid;2814.160`,
			wantErr: "no valid transactions found",
		},
		{
			name: "Process_error_missing_fields",
			input: `Bokföringsdatum;Valutadatum;Verifikationsnummer;Text;Belopp;Saldo
2025-02-24;2025-02-22;5490990004;56130086210;-1000.000`,
			wantErr: "read_record failed",
		},
	}

	// TODO: Replace context.TODO() with proper context handling for timeouts and cancellation
	ctx := context.TODO()
	logger := slog.Default()
	processor := NewSEBProcessor(logger)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Process the test data
			_, err := processor.ProcessDocument(ctx, strings.NewReader(tt.input))

			// Validate error presence
			if err == nil {
				t.Fatal("expected error, got nil")
				return
			}

			// Validate error message
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
				return
			}
		})
	}
}

// Helper function to create a reader that fails
type failingReader struct {
	err error
}

func (f failingReader) Read(p []byte) (n int, err error) {
	return 0, f.err
}

func TestSEBProcessor_error_io_failures(t *testing.T) {
	tests := []struct {
		name    string
		reader  io.Reader
		wantErr string
	}{
		{
			name:    "Process_error_read_failure",
			reader:  failingReader{err: fmt.Errorf("simulated read error")},
			wantErr: "read_header failed",
		},
		{
			name:    "Process_error_empty_input",
			reader:  strings.NewReader(""),
			wantErr: "read_header failed",
		},
	}

	// TODO: Replace context.TODO() with proper context handling for timeouts and cancellation
	ctx := context.TODO()
	logger := slog.Default()
	processor := NewSEBProcessor(logger)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Process the test data
			_, err := processor.ProcessDocument(ctx, tt.reader)

			// Validate error presence
			if err == nil {
				t.Fatal("expected error, got nil")
				return
			}

			// Validate error message
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
				return
			}
		})
	}
}
