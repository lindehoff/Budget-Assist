package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func closeTestDB(db interface{}) error {
	switch d := db.(type) {
	case *gorm.DB:
		sqlDB, err := d.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	case *sql.DB:
		return d.Close()
	default:
		return fmt.Errorf("unsupported database type: %T", db)
	}
}

func setupTestDB(t *testing.T) (context.Context, *gorm.DB) {
	t.Helper()

	// Create test context with logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	ctx := WithLogger(context.Background(), logger)

	tempDir, err := os.MkdirTemp("", "budgetassist-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	config := &Config{
		DBPath: tempDir + "/test.db",
	}

	db, err := Initialize(config, logger)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	t.Cleanup(func() {
		if err := closeTestDB(db); err != nil {
			t.Errorf("failed to close test database: %v", err)
		}
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("failed to remove temp dir: %v", err)
		}
	})

	return ctx, db
}

// Test_Successfully_create_and_retrieve_entities tests the creation and retrieval of entities
func Test_Successfully_create_and_retrieve_entities(t *testing.T) {
	t.Parallel()
	ctx, db := setupTestDB(t)

	tests := []struct {
		name         string
		categoryType *CategoryType
		want         *CategoryType
	}{
		{
			name: "Successfully create valid category type",
			categoryType: &CategoryType{
				Name:        "Test Category Type",
				Description: "Test description",
				IsMultiple:  true,
			},
		},
		{
			name: "Successfully create another category type",
			categoryType: &CategoryType{
				Name:        "Another Category Type",
				Description: "Another description",
				IsMultiple:  false,
			},
		},
		{
			name: "Successfully create third category type",
			categoryType: &CategoryType{
				Name:        "Third Category Type",
				Description: "Third description",
				IsMultiple:  true,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create category type
			if err := db.WithContext(ctx).Create(tt.categoryType).Error; err != nil {
				t.Fatalf("failed to create category type: %v", err)
			}
			if tt.categoryType.ID == 0 {
				t.Fatal("category type ID should not be 0 after creation")
			}

			// Verify the category type was created correctly
			var got CategoryType
			if err := db.WithContext(ctx).First(&got, tt.categoryType.ID).Error; err != nil {
				t.Fatalf("failed to retrieve category type: %v", err)
			}

			if got.Name != tt.categoryType.Name {
				t.Errorf("Name = %v, want %v", got.Name, tt.categoryType.Name)
			}
			if got.Description != tt.categoryType.Description {
				t.Errorf("Description = %v, want %v", got.Description, tt.categoryType.Description)
			}
			if got.IsMultiple != tt.categoryType.IsMultiple {
				t.Errorf("IsMultiple = %v, want %v", got.IsMultiple, tt.categoryType.IsMultiple)
			}
		})
	}
}

// Test_Transaction_error_invalid_currency tests that an error is returned when creating a transaction with an invalid currency
func Test_Transaction_error_invalid_currency(t *testing.T) {
	t.Parallel()
	ctx, db := setupTestDB(t)

	transaction := &Transaction{
		Date:            time.Now(),
		TransactionDate: time.Now(),
		Amount:          decimal.NewFromInt(100),
		Description:     "Test transaction",
		Currency:        "INVALID",
	}

	err := db.WithContext(ctx).Create(transaction).Error
	if err == nil {
		t.Error("expected error for invalid currency, got nil")
	}
	if err != nil && err.Error() != "invalid currency: INVALID" {
		t.Errorf("expected error message 'invalid currency: INVALID', got %v", err)
	}
}

func TestStore_GetTransactions(t *testing.T) {
	ctx, db := setupTestDB(t)

	// Create test data
	categoryType := &CategoryType{Name: "Test"}
	if err := db.WithContext(ctx).Create(categoryType).Error; err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	category := &Category{TypeID: categoryType.ID, Name: "Test Category"}
	if err := db.WithContext(ctx).Create(category).Error; err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	tx := &Transaction{
		Amount:      decimal.NewFromFloat(100.00),
		Currency:    CurrencySEK,
		CategoryID:  &category.ID,
		Description: "Test transaction",
	}
	if err := db.WithContext(ctx).Create(tx).Error; err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Test transaction retrieval
	var got []Transaction
	if err := db.WithContext(ctx).Find(&got).Error; err != nil {
		t.Fatalf("GetTransactions() error: %v", err)
	}

	if len(got) != 1 {
		t.Errorf("GetTransactions() got %d transactions, want 1", len(got))
	}
	if !got[0].Amount.Equal(tx.Amount) {
		t.Errorf("GetTransactions() amount = %v, want %v", got[0].Amount, tx.Amount)
	}
}
