package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"gorm.io/gorm"
)

type testCase struct {
	name               string
	categoryType       CategoryType
	category           Category
	subcategory        Subcategory
	transaction        Transaction
	categoryTypeTransl Translation
	categoryTransl     Translation
	subcategoryTransl  Translation
}

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

func setupTestDB(t *testing.T) (context.Context, *gorm.DB, func()) {
	// Create test context with logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	ctx := WithLogger(context.Background(), logger)

	tempDir, err := os.MkdirTemp("", "budgetassist-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	config := &Config{
		DBPath: tempDir + "/test.db",
	}

	db, err := Initialize(config)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to initialize test database: %v", err)
	}

	cleanup := func() {
		if err := closeTestDB(db); err != nil {
			t.Errorf("failed to close database: %v", err)
		}
		os.RemoveAll(tempDir)
	}

	return ctx, db, cleanup
}

func createAndValidateEntities(t *testing.T, ctx context.Context, db *gorm.DB, tc *testCase) {
	// Create CategoryType
	result := db.WithContext(ctx).Create(&tc.categoryType)
	if result.Error != nil {
		t.Fatalf("failed to create category type: %v", result.Error)
	}
	if tc.categoryType.ID == 0 {
		t.Fatal("category type ID should not be 0 after creation")
	}

	// Create CategoryType Translation
	tc.categoryTypeTransl.EntityID = tc.categoryType.ID
	result = db.WithContext(ctx).Create(&tc.categoryTypeTransl)
	if result.Error != nil {
		t.Fatalf("failed to create category type translation: %v", result.Error)
	}

	// Create Category
	tc.category.TypeID = tc.categoryType.ID
	result = db.WithContext(ctx).Create(&tc.category)
	if result.Error != nil {
		t.Fatalf("failed to create category: %v", result.Error)
	}
	if tc.category.ID == 0 {
		t.Fatal("category ID should not be 0 after creation")
	}

	// Create Category Translation
	tc.categoryTransl.EntityID = tc.category.ID
	result = db.WithContext(ctx).Create(&tc.categoryTransl)
	if result.Error != nil {
		t.Fatalf("failed to create category translation: %v", result.Error)
	}

	// Create Subcategory
	tc.subcategory.CategoryTypeID = tc.categoryType.ID
	result = db.WithContext(ctx).Create(&tc.subcategory)
	if result.Error != nil {
		t.Fatalf("failed to create subcategory: %v", result.Error)
	}
	if tc.subcategory.ID == 0 {
		t.Fatal("subcategory ID should not be 0 after creation")
	}

	// Create Subcategory Translation
	tc.subcategoryTransl.EntityID = tc.subcategory.ID
	result = db.WithContext(ctx).Create(&tc.subcategoryTransl)
	if result.Error != nil {
		t.Fatalf("failed to create subcategory translation: %v", result.Error)
	}

	// Create Transaction
	tc.transaction.CategoryID = &tc.category.ID
	tc.transaction.SubcategoryID = &tc.subcategory.ID
	result = db.WithContext(ctx).Create(&tc.transaction)
	if result.Error != nil {
		t.Fatalf("failed to create transaction: %v", result.Error)
	}
	if tc.transaction.ID == 0 {
		t.Fatal("transaction ID should not be 0 after creation")
	}
}

func validateEntities(t *testing.T, ctx context.Context, db *gorm.DB, tc *testCase) {
	// Test Translation Retrieval
	var foundCategoryType CategoryType
	result := db.WithContext(ctx).First(&foundCategoryType, tc.categoryType.ID)
	if result.Error != nil {
		t.Fatalf("failed to retrieve category type: %v", result.Error)
	}

	var translations []Translation
	result = db.WithContext(ctx).Where("entity_type = ? AND entity_id = ?", EntityTypeCategoryType, foundCategoryType.ID).Find(&translations)
	if result.Error != nil {
		t.Fatalf("failed to retrieve translations: %v", result.Error)
	}
	foundCategoryType.Translations = translations

	// Test translation methods
	if foundCategoryType.GetTranslation(LangSV) != tc.categoryTypeTransl.Name {
		t.Errorf("expected Swedish translation %q, got %q", tc.categoryTypeTransl.Name, foundCategoryType.GetTranslation(LangSV))
	}
	if foundCategoryType.GetTranslation(LangEN) != tc.categoryType.Name {
		t.Errorf("expected English name %q, got %q", tc.categoryType.Name, foundCategoryType.GetTranslation(LangEN))
	}

	// Test Transaction Retrieval with Relations
	var found Transaction
	result = db.WithContext(ctx).Preload("Category").Preload("Subcategory").First(&found, tc.transaction.ID)
	if result.Error != nil {
		t.Fatalf("failed to retrieve transaction: %v", result.Error)
	}

	// Verify transaction details
	if found.Amount != tc.transaction.Amount {
		t.Errorf("expected amount %v, got %v", tc.transaction.Amount, found.Amount)
	}
	if found.Currency != tc.transaction.Currency {
		t.Errorf("expected currency %v, got %v", tc.transaction.Currency, found.Currency)
	}
	if found.Category == nil {
		t.Error("expected category to be loaded")
	}
	if found.Subcategory == nil {
		t.Error("expected subcategory to be loaded")
	}
}

func Test_Successfully_create_and_retrieve_entities(t *testing.T) {
	ctx, db, cleanup := setupTestDB(t)
	defer cleanup()

	testCases := []testCase{
		{
			name: "Vehicle with translations and transaction",
			categoryType: CategoryType{
				Name:        "Vehicle",
				IsMultiple:  true,
				Description: "Vehicle related expenses",
			},
			categoryTypeTransl: Translation{
				EntityType:   string(EntityTypeCategoryType),
				LanguageCode: LangSV,
				Name:         "Fordon",
			},
			category: Category{
				Name:               "Car",
				InstanceIdentifier: "Vehicle: ABC123",
				IsActive:           true,
			},
			categoryTransl: Translation{
				EntityType:   string(EntityTypeCategory),
				LanguageCode: LangSV,
				Name:         "Bil",
			},
			subcategory: Subcategory{
				Name:        "Fuel",
				Description: "Fuel expenses",
				IsSystem:    true,
			},
			subcategoryTransl: Translation{
				EntityType:   string(EntityTypeSubcategory),
				LanguageCode: LangSV,
				Name:         "Bränsle",
			},
			transaction: Transaction{
				Amount:          150.00,
				Currency:        CurrencySEK,
				TransactionDate: time.Now(),
				Description:     "Fuel purchase",
				RawData:         "Original transaction data",
				AIAnalysis:      "AI-generated insights about the transaction",
			},
		},
		{
			name: "Fixed costs with translations and transaction",
			categoryType: CategoryType{
				Name:        "Fixed Costs",
				IsMultiple:  false,
				Description: "Regular fixed expenses",
			},
			categoryTypeTransl: Translation{
				EntityType:   string(EntityTypeCategoryType),
				LanguageCode: LangSV,
				Name:         "Fasta kostnader",
			},
			category: Category{
				Name:     "Media",
				IsActive: true,
			},
			categoryTransl: Translation{
				EntityType:   string(EntityTypeCategory),
				LanguageCode: LangSV,
				Name:         "Medier",
			},
			subcategory: Subcategory{
				Name:        "Internet",
				Description: "Internet subscription",
				IsSystem:    true,
			},
			subcategoryTransl: Translation{
				EntityType:   string(EntityTypeSubcategory),
				LanguageCode: LangSV,
				Name:         "Internet",
			},
			transaction: Transaction{
				Amount:          299.00,
				Currency:        CurrencyEUR,
				TransactionDate: time.Now(),
				Description:     "Monthly internet subscription",
				RawData:         "Original transaction data",
				AIAnalysis:      "AI-generated insights about the transaction",
			},
		},
		{
			name: "Income with translations and transaction",
			categoryType: CategoryType{
				Name:        "Income",
				IsMultiple:  true,
				Description: "Income sources",
			},
			categoryTypeTransl: Translation{
				EntityType:   string(EntityTypeCategoryType),
				LanguageCode: LangSV,
				Name:         "Inkomster",
			},
			category: Category{
				Name:               "Primary Job",
				InstanceIdentifier: "Income: Salary",
				IsActive:           true,
			},
			categoryTransl: Translation{
				EntityType:   string(EntityTypeCategory),
				LanguageCode: LangSV,
				Name:         "Huvudarbete",
			},
			subcategory: Subcategory{
				Name:        "Salary",
				Description: "Monthly salary",
				IsSystem:    true,
			},
			subcategoryTransl: Translation{
				EntityType:   string(EntityTypeSubcategory),
				LanguageCode: LangSV,
				Name:         "Lön",
			},
			transaction: Transaction{
				Amount:          25000.00,
				Currency:        CurrencyUSD,
				TransactionDate: time.Now(),
				Description:     "Monthly salary payment",
				RawData:         "Original transaction data",
				AIAnalysis:      "AI-generated insights about the transaction",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			createAndValidateEntities(t, ctx, db, &tc)
			validateEntities(t, ctx, db, &tc)
		})
	}
}

func Test_Transaction_error_invalid_currency(t *testing.T) {
	ctx, db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a basic category structure
	categoryType := &CategoryType{
		Name:        "Test",
		IsMultiple:  false,
		Description: "Test category",
	}
	result := db.WithContext(ctx).Create(categoryType)
	if result.Error != nil {
		t.Fatalf("failed to create category type: %v", result.Error)
	}

	category := &Category{
		TypeID:   categoryType.ID,
		Name:     "Test Category",
		IsActive: true,
	}
	result = db.WithContext(ctx).Create(category)
	if result.Error != nil {
		t.Fatalf("failed to create category: %v", result.Error)
	}

	// Try to create a transaction with invalid currency
	tx := &Transaction{
		Amount:          100.00,
		Currency:        "INVALID",
		TransactionDate: time.Now(),
		Description:     "Test transaction",
		CategoryID:      &category.ID,
	}
	result = db.WithContext(ctx).Create(tx)

	// Should fail due to currency check constraint
	if result.Error == nil {
		t.Error("expected error for invalid currency, got nil")
		return
	}
}
