package processor

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// SEBFormat represents the column structure of SEB CSV files
type SEBFormat struct {
	BookingDate int // Bokföringsdatum
	ValueDate   int // Valutadatum
	Reference   int // Verifikationsnummer
	Description int // Text
	Amount      int // Belopp
	Balance     int // Saldo
}

// Default SEB CSV format column indices
var defaultSEBFormat = SEBFormat{
	BookingDate: 0,
	ValueDate:   1,
	Reference:   2,
	Description: 3,
	Amount:      4,
	Balance:     5,
}

// SEBProcessor implements the DocumentProcessor interface for SEB bank statements
type SEBProcessor struct {
	logger *slog.Logger
	format SEBFormat
}

// NewSEBProcessor creates a new SEB CSV processor
func NewSEBProcessor(logger *slog.Logger) *SEBProcessor {
	return &SEBProcessor{
		logger: logger,
		format: defaultSEBFormat,
	}
}

// ProcessDocument implements the DocumentProcessor interface for SEB CSV files
func (p *SEBProcessor) ProcessDocument(ctx context.Context, reader io.Reader) ([]Transaction, error) {
	csvReader := csv.NewReader(reader)
	csvReader.Comma = ';' // SEB uses semicolon as delimiter
	csvReader.TrimLeadingSpace = true

	// Read and validate header
	header, err := csvReader.Read()
	if err != nil {
		return nil, &ProcessingError{
			Operation: "read_header",
			Err:       err,
		}
	}

	if err := p.validateHeader(header); err != nil {
		return nil, err
	}

	var transactions []Transaction
	lineNum := 1 // Start after header

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, &ProcessingError{
				Operation: "read_record",
				Err:       err,
				Line:      lineNum,
			}
		}

		trans, err := p.parseTransaction(record, lineNum)
		if err != nil {
			// Log the error but continue processing other records
			p.logger.Warn("failed to parse transaction",
				"line", lineNum,
				"error", err,
				"raw_data", record)
			lineNum++
			continue
		}

		transactions = append(transactions, trans)
		lineNum++
	}

	if len(transactions) == 0 {
		return nil, &ProcessingError{
			Operation: "process_document",
			Err:       fmt.Errorf("no valid transactions found in document"),
		}
	}

	return transactions, nil
}

func (p *SEBProcessor) validateHeader(header []string) error {
	expectedHeaders := []string{
		"Bokföringsdatum",
		"Valutadatum",
		"Verifikationsnummer",
		"Text",
		"Belopp",
		"Saldo",
	}

	if len(header) != len(expectedHeaders) {
		return &ProcessingError{
			Operation: "validate_header",
			Err:       fmt.Errorf("invalid number of columns: got %d, want %d", len(header), len(expectedHeaders)),
		}
	}

	for i, expected := range expectedHeaders {
		if header[i] != expected {
			return &ProcessingError{
				Operation: "validate_header",
				Err:       fmt.Errorf("invalid header at column %d: got %s, want %s", i+1, header[i], expected),
			}
		}
	}

	return nil
}

func (p *SEBProcessor) parseTransaction(record []string, lineNum int) (Transaction, error) {
	if len(record) != 6 {
		return Transaction{}, &ProcessingError{
			Operation: "parse_transaction",
			Err:       fmt.Errorf("invalid number of fields: got %d, want 6", len(record)),
			Line:      lineNum,
		}
	}

	// Parse booking date
	bookingDate, err := time.Parse("2006-01-02", record[p.format.BookingDate])
	if err != nil {
		return Transaction{}, &ProcessingError{
			Operation: "parse_date",
			Err:       err,
			Line:      lineNum,
		}
	}

	// Parse amount (replace comma with dot for decimal parsing)
	amountStr := strings.Replace(record[p.format.Amount], ",", ".", 1)
	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return Transaction{}, &ProcessingError{
			Operation: "parse_amount",
			Err:       err,
			Line:      lineNum,
		}
	}

	return Transaction{
		Date:        bookingDate,
		Amount:      amount,
		Description: strings.TrimSpace(record[p.format.Description]),
		Reference:   record[p.format.Reference],
		RawData: map[string]any{
			"ValueDate": record[p.format.ValueDate],
			"Balance":   record[p.format.Balance],
		},
		Source: "SEB",
	}, nil
}

// ProcessingError represents an error during CSV processing
type ProcessingError struct {
	Err       error  // 8-byte pointer
	Operation string // 16 bytes
	Line      int    // 8 bytes
}

func (e *ProcessingError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("%s failed at line %d: %v", e.Operation, e.Line, e.Err)
	}
	return fmt.Sprintf("%s failed: %v", e.Operation, e.Err)
}

func (e *ProcessingError) Unwrap() error {
	return e.Err
}
