package processor

import (
	"context"
	"io"
	"time"

	"github.com/shopspring/decimal"
)

// Transaction represents a financial transaction from any source
type Transaction struct {
	RawData     map[string]any  // 8-byte pointer
	Amount      decimal.Decimal // 8-byte pointer
	Date        time.Time       // 24 bytes
	Description string          // 16 bytes
	Reference   string          // 16 bytes
	Category    string          // 16 bytes
	SubCategory string          // 16 bytes
	Source      string          // 16 bytes
}

// DocumentProcessor defines the interface for processing different types of financial documents
type DocumentProcessor interface {
	// ProcessDocument processes a document and returns a slice of transactions
	ProcessDocument(ctx context.Context, reader io.Reader) ([]Transaction, error)
}
