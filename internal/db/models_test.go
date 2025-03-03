package db

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func Test_Transaction_FormatAmount(t *testing.T) {
	tests := []struct {
		name        string
		transaction Transaction
		want        string
	}{
		{
			name: "Successfully_format_SEK_amount",
			transaction: Transaction{
				Amount:   decimal.NewFromFloat(100.50),
				Currency: CurrencySEK,
			},
			want: "100.50 SEK",
		},
		{
			name: "Successfully_format_EUR_amount",
			transaction: Transaction{
				Amount:   decimal.NewFromFloat(-50.25),
				Currency: CurrencyEUR,
			},
			want: "-50.25 EUR",
		},
		{
			name: "Successfully_format_USD_amount",
			transaction: Transaction{
				Amount:   decimal.NewFromFloat(1000),
				Currency: CurrencyUSD,
			},
			want: "1000.00 USD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.transaction.FormatAmount()
			if got != tt.want {
				t.Errorf("FormatAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Transaction_BeforeCreate(t *testing.T) {
	tests := []struct {
		name        string
		transaction Transaction
		wantErr     string
	}{
		{
			name: "Successfully_validate_SEK_currency",
			transaction: Transaction{
				Date:            time.Now(),
				TransactionDate: time.Now(),
				Amount:          decimal.NewFromFloat(100),
				Description:     "Test transaction",
				Currency:        CurrencySEK,
			},
		},
		{
			name: "Successfully_validate_EUR_currency",
			transaction: Transaction{
				Date:            time.Now(),
				TransactionDate: time.Now(),
				Amount:          decimal.NewFromFloat(100),
				Description:     "Test transaction",
				Currency:        CurrencyEUR,
			},
		},
		{
			name: "Successfully_validate_USD_currency",
			transaction: Transaction{
				Date:            time.Now(),
				TransactionDate: time.Now(),
				Amount:          decimal.NewFromFloat(100),
				Description:     "Test transaction",
				Currency:        CurrencyUSD,
			},
		},
		{
			name: "Validate_error_invalid_currency",
			transaction: Transaction{
				Date:            time.Now(),
				TransactionDate: time.Now(),
				Amount:          decimal.NewFromFloat(100),
				Description:     "Test transaction",
				Currency:        "INVALID",
			},
			wantErr: "invalid currency: INVALID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.transaction.BeforeCreate(&gorm.DB{})
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("BeforeCreate() error = nil, wantErr %q", tt.wantErr)
					return
				}
				if err.Error() != tt.wantErr {
					t.Errorf("BeforeCreate() error = %v, wantErr %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("BeforeCreate() error = %v, wantErr nil", err)
			}
		})
	}
}
