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
	var found CategoryType
	result := db.WithContext(ctx).First(&found, tc.categoryType.ID)
	if result.Error != nil {
		t.Fatalf("failed to retrieve category type: %v", result.Error)
	}

	var translations []Translation
	result = db.WithContext(ctx).Where("entity_type = ? AND entity_id = ?", EntityTypeCategoryType, found.ID).Find(&translations)
	if result.Error != nil {
		t.Fatalf("failed to retrieve translations: %v", result.Error)
	}
	found.Translations = translations

	// Test translation methods
	if found.GetTranslation(LangSV) != tc.categoryTypeTransl.Name {
		t.Errorf("expected Swedish translation %q, got %q", tc.categoryTypeTransl.Name, found.GetTranslation(LangSV))
	}
	if found.GetTranslation(LangEN) != tc.categoryTypeTransl.Name {
		t.Errorf("expected English name %q, got %q", tc.categoryTypeTransl.Name, found.GetTranslation(LangEN))
	}

	// Test Transaction Retrieval with Relations
	var foundTransaction Transaction
	result = db.WithContext(ctx).Preload("Category").Preload("Subcategory").First(&foundTransaction, tc.transaction.ID)
	if result.Error != nil {
		t.Fatalf("failed to retrieve transaction: %v", result.Error)
	}

	// Verify transaction details
	if !foundTransaction.Amount.Equal(tc.transaction.Amount) {
		t.Errorf("expected amount %s, got %s", tc.transaction.Amount.String(), foundTransaction.Amount.String())
	}
	if foundTransaction.Currency != tc.transaction.Currency {
		t.Errorf("expected currency %v, got %v", tc.transaction.Currency, foundTransaction.Currency)
	}
	if foundTransaction.Category == nil {
		t.Error("expected category to be loaded")
	}
	if foundTransaction.Subcategory == nil {
		t.Error("expected subcategory to be loaded")
	}
}

func Test_Successfully_create_and_retrieve_entities(t *testing.T) {
	ctx, db, cleanup := setupTestDB(t)
	defer cleanup()

	tests := []struct {
		name               string
		categoryType       CategoryType
		categoryTypeTransl Translation
		category           Category
		subcategory        Subcategory
		wantErr            bool
	}{
		{
			name:         "Successfully_create_valid_category_type",
			categoryType: CategoryType{},
			categoryTypeTransl: Translation{
				EntityType:   string(EntityTypeCategoryType),
				LanguageCode: LangEN,
				Name:         "Test Type",
				Description:  "Test type description",
			},
			category: Category{
				IsActive: true,
			},
			subcategory: Subcategory{
				IsActive: true,
			},
			wantErr: false,
		},
		{
			name:         "Successfully_create_another_category_type",
			categoryType: CategoryType{},
			categoryTypeTransl: Translation{
				EntityType:   string(EntityTypeCategoryType),
				LanguageCode: LangEN,
				Name:         "Another Type",
				Description:  "Another type description",
			},
			category: Category{
				IsActive: true,
			},
			subcategory: Subcategory{
				IsActive: true,
			},
			wantErr: false,
		},
		{
			name:         "Successfully_create_third_category_type",
			categoryType: CategoryType{},
			categoryTypeTransl: Translation{
				EntityType:   string(EntityTypeCategoryType),
				LanguageCode: LangEN,
				Name:         "Third Type",
				Description:  "Third type description",
			},
			category: Category{
				IsActive: true,
			},
			subcategory: Subcategory{
				IsActive: true,
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create CategoryType
			result := db.WithContext(ctx).Create(&tc.categoryType)
			if result.Error != nil {
				t.Fatalf("failed to create category type: %v", result.Error)
			}
			if tc.categoryType.ID == 0 {
				t.Fatal("category type ID not set")
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

			// Create Subcategory
			tc.subcategory.CategoryTypeID = tc.categoryType.ID
			result = db.WithContext(ctx).Create(&tc.subcategory)
			if result.Error != nil {
				t.Fatalf("failed to create subcategory: %v", result.Error)
			}

			// Verify CategoryType and its translations
			var found CategoryType
			result = db.WithContext(ctx).First(&found, tc.categoryType.ID)
			if result.Error != nil {
				t.Fatalf("failed to retrieve category type: %v", result.Error)
			}

			var translations []Translation
			result = db.WithContext(ctx).Where("entity_type = ? AND entity_id = ?", EntityTypeCategoryType, found.ID).Find(&translations)
			if result.Error != nil {
				t.Fatalf("failed to retrieve translations: %v", result.Error)
			}

			if len(translations) != 1 {
				t.Errorf("expected 1 translation, got %d", len(translations))
			}

			if found.GetTranslation(LangEN) != tc.categoryTypeTransl.Name {
				t.Errorf("expected English name %q, got %q", tc.categoryTypeTransl.Name, found.GetTranslation(LangEN))
			}
		})
	}
}

func Test_Transaction_error_invalid_currency(t *testing.T) {
	ctx, db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a basic category structure
	categoryType := &CategoryType{}
	if err := db.WithContext(ctx).Create(categoryType).Error; err != nil {
		t.Fatalf("failed to create test category type: %v", err)
	}

	// Add translation for the category type
	translation := &Translation{
		EntityID:     categoryType.ID,
		EntityType:   string(EntityTypeCategoryType),
		LanguageCode: LangEN,
		Name:         "Test",
		Description:  "Test type description",
	}
	if err := db.WithContext(ctx).Create(translation).Error; err != nil {
		t.Fatalf("failed to create test translation: %v", err)
	}

	category := &Category{
		TypeID:   categoryType.ID,
		IsActive: true,
	}
	result := db.WithContext(ctx).Create(category)
	if result.Error != nil {
		t.Fatalf("failed to create category: %v", result.Error)
	}

	// Try to create a transaction with invalid currency
	tx := &Transaction{
		Amount:          decimal.NewFromFloat(100.00),
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

func TestStore_GetTransactions(t *testing.T) {
	ctx, db, cleanup := setupTestDB(t)
	defer cleanup()

	// Use context consistently
	categoryType := &CategoryType{Name: "Test"}
	if err := db.WithContext(ctx).Create(categoryType).Error; err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	category := &Category{TypeID: categoryType.ID, Name: "Test Category"}
	if err := db.WithContext(ctx).Create(category).Error; err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Create transaction with context
	tx := &Transaction{
		Amount:      decimal.NewFromFloat(100.00),
		Currency:    CurrencySEK,
		CategoryID:  &category.ID,
		Description: "Test transaction",
	}
	if err := db.WithContext(ctx).Create(tx).Error; err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Retrieve transactions with context
	var got []Transaction
	if err := db.WithContext(ctx).Find(&got).Error; err != nil {
		t.Fatalf("GetTransactions() error: %v", err)
	}

	// Validate results
	if len(got) != 1 || !got[0].Amount.Equal(tx.Amount) {
		t.Errorf("GetTransactions() unexpected results: %+v", got)
	}
}

func Test_Successfully_create_valid_category_type(t *testing.T) {
	ctx, db, cleanup := setupTestDB(t)
	defer cleanup()

	categoryType := CategoryType{}
	categoryTypeTransl := Translation{
		EntityType:   string(EntityTypeCategoryType),
		LanguageCode: LangSV,
		Name:         "Test Type SV",
		Description:  "Test type description SV",
	}
	category := Category{
		IsActive: true,
	}
	subcategory := Subcategory{
		IsActive: true,
	}

	result := db.WithContext(ctx).Create(&categoryType)
	if result.Error != nil {
		t.Fatalf("failed to create category type: %v", result.Error)
	}

	result = db.WithContext(ctx).Create(&categoryTypeTransl)
	if result.Error != nil {
		t.Fatalf("failed to create category type translation: %v", result.Error)
	}

	category.TypeID = categoryType.ID
	result = db.WithContext(ctx).Create(&category)
	if result.Error != nil {
		t.Fatalf("failed to create category: %v", result.Error)
	}

	subcategory.CategoryTypeID = categoryType.ID
	result = db.WithContext(ctx).Create(&subcategory)
	if result.Error != nil {
		t.Fatalf("failed to create subcategory: %v", result.Error)
	}

	transaction := Transaction{
		Amount:        decimal.NewFromFloat(100.00),
		Currency:      CurrencySEK,
		Date:          time.Now(),
		Description:   "Test transaction",
		CategoryID:    &category.ID,
		SubcategoryID: &subcategory.ID,
	}
	result = db.WithContext(ctx).Create(&transaction)
	if result.Error != nil {
		t.Fatalf("failed to create transaction: %v", result.Error)
	}

	var foundCategoryType CategoryType
	result = db.WithContext(ctx).First(&foundCategoryType, categoryType.ID)
	if result.Error != nil {
		t.Fatalf("failed to retrieve category type: %v", result.Error)
	}

	var foundTranslation Translation
	result = db.WithContext(ctx).Where("entity_type = ? AND entity_id = ?", EntityTypeCategoryType, foundCategoryType.ID).First(&foundTranslation)
	if result.Error != nil {
		t.Fatalf("failed to retrieve category type translation: %v", result.Error)
	}

	if foundTranslation.Name != categoryTypeTransl.Name {
		t.Errorf("expected category type translation name %q, got %q", categoryTypeTransl.Name, foundTranslation.Name)
	}

	var foundTransaction Transaction
	result = db.WithContext(ctx).Preload("Category").Preload("Subcategory").First(&foundTransaction, transaction.ID)
	if result.Error != nil {
		t.Fatalf("failed to retrieve transaction: %v", result.Error)
	}

	if !foundTransaction.Amount.Equal(transaction.Amount) {
		t.Errorf("expected transaction amount %s, got %s", transaction.Amount.String(), foundTransaction.Amount.String())
	}
	if foundTransaction.Currency != transaction.Currency {
		t.Errorf("expected transaction currency %v, got %v", transaction.Currency, foundTransaction.Currency)
	}
	if foundTransaction.Category == nil {
		t.Error("expected category to be loaded")
	}
	if foundTransaction.Subcategory == nil {
		t.Error("expected subcategory to be loaded")
	}
}

func Test_Successfully_create_another_category_type(t *testing.T) {
	ctx, db, cleanup := setupTestDB(t)
	defer cleanup()

	categoryType := CategoryType{}
	categoryTypeTransl := Translation{
		EntityType:   string(EntityTypeCategoryType),
		LanguageCode: LangSV,
		Name:         "Another Type SV",
		Description:  "Another type description SV",
	}
	category := Category{
		IsActive: true,
	}
	subcategory := Subcategory{
		IsActive: true,
	}

	result := db.WithContext(ctx).Create(&categoryType)
	if result.Error != nil {
		t.Fatalf("failed to create category type: %v", result.Error)
	}

	result = db.WithContext(ctx).Create(&categoryTypeTransl)
	if result.Error != nil {
		t.Fatalf("failed to create category type translation: %v", result.Error)
	}

	category.TypeID = categoryType.ID
	result = db.WithContext(ctx).Create(&category)
	if result.Error != nil {
		t.Fatalf("failed to create category: %v", result.Error)
	}

	subcategory.CategoryTypeID = categoryType.ID
	result = db.WithContext(ctx).Create(&subcategory)
	if result.Error != nil {
		t.Fatalf("failed to create subcategory: %v", result.Error)
	}

	transaction := Transaction{
		Amount:        decimal.NewFromFloat(100.00),
		Currency:      CurrencySEK,
		Date:          time.Now(),
		Description:   "Test transaction",
		CategoryID:    &category.ID,
		SubcategoryID: &subcategory.ID,
	}
	result = db.WithContext(ctx).Create(&transaction)
	if result.Error != nil {
		t.Fatalf("failed to create transaction: %v", result.Error)
	}

	var foundCategoryType CategoryType
	result = db.WithContext(ctx).First(&foundCategoryType, categoryType.ID)
	if result.Error != nil {
		t.Fatalf("failed to retrieve category type: %v", result.Error)
	}

	var foundTranslation Translation
	result = db.WithContext(ctx).Where("entity_type = ? AND entity_id = ?", EntityTypeCategoryType, foundCategoryType.ID).First(&foundTranslation)
	if result.Error != nil {
		t.Fatalf("failed to retrieve category type translation: %v", result.Error)
	}

	if foundTranslation.Name != categoryTypeTransl.Name {
		t.Errorf("expected category type translation name %q, got %q", categoryTypeTransl.Name, foundTranslation.Name)
	}

	var foundTransaction Transaction
	result = db.WithContext(ctx).Preload("Category").Preload("Subcategory").First(&foundTransaction, transaction.ID)
	if result.Error != nil {
		t.Fatalf("failed to retrieve transaction: %v", result.Error)
	}

	if !foundTransaction.Amount.Equal(transaction.Amount) {
		t.Errorf("expected transaction amount %s, got %s", transaction.Amount.String(), foundTransaction.Amount.String())
	}
	if foundTransaction.Currency != transaction.Currency {
		t.Errorf("expected transaction currency %v, got %v", transaction.Currency, foundTransaction.Currency)
	}
	if foundTransaction.Category == nil {
		t.Error("expected category to be loaded")
	}
	if foundTransaction.Subcategory == nil {
		t.Error("expected subcategory to be loaded")
	}
}

func Test_Successfully_create_third_category_type(t *testing.T) {
	ctx, db, cleanup := setupTestDB(t)
	defer cleanup()

	categoryType := CategoryType{}
	categoryTypeTransl := Translation{
		EntityType:   string(EntityTypeCategoryType),
		LanguageCode: LangSV,
		Name:         "Third Type SV",
		Description:  "Third type description SV",
	}
	category := Category{
		IsActive: true,
	}
	subcategory := Subcategory{
		IsActive: true,
	}

	result := db.WithContext(ctx).Create(&categoryType)
	if result.Error != nil {
		t.Fatalf("failed to create category type: %v", result.Error)
	}

	result = db.WithContext(ctx).Create(&categoryTypeTransl)
	if result.Error != nil {
		t.Fatalf("failed to create category type translation: %v", result.Error)
	}

	category.TypeID = categoryType.ID
	result = db.WithContext(ctx).Create(&category)
	if result.Error != nil {
		t.Fatalf("failed to create category: %v", result.Error)
	}

	subcategory.CategoryTypeID = categoryType.ID
	result = db.WithContext(ctx).Create(&subcategory)
	if result.Error != nil {
		t.Fatalf("failed to create subcategory: %v", result.Error)
	}

	transaction := Transaction{
		Amount:        decimal.NewFromFloat(100.00),
		Currency:      CurrencySEK,
		Date:          time.Now(),
		Description:   "Test transaction",
		CategoryID:    &category.ID,
		SubcategoryID: &subcategory.ID,
	}
	result = db.WithContext(ctx).Create(&transaction)
	if result.Error != nil {
		t.Fatalf("failed to create transaction: %v", result.Error)
	}

	var foundCategoryType CategoryType
	result = db.WithContext(ctx).First(&foundCategoryType, categoryType.ID)
	if result.Error != nil {
		t.Fatalf("failed to retrieve category type: %v", result.Error)
	}

	var foundTranslation Translation
	result = db.WithContext(ctx).Where("entity_type = ? AND entity_id = ?", EntityTypeCategoryType, foundCategoryType.ID).First(&foundTranslation)
	if result.Error != nil {
		t.Fatalf("failed to retrieve category type translation: %v", result.Error)
	}

	if foundTranslation.Name != categoryTypeTransl.Name {
		t.Errorf("expected category type translation name %q, got %q", categoryTypeTransl.Name, foundTranslation.Name)
	}

	var foundTransaction Transaction
	result = db.WithContext(ctx).Preload("Category").Preload("Subcategory").First(&foundTransaction, transaction.ID)
	if result.Error != nil {
		t.Fatalf("failed to retrieve transaction: %v", result.Error)
	}

	if !foundTransaction.Amount.Equal(transaction.Amount) {
		t.Errorf("expected transaction amount %s, got %s", transaction.Amount.String(), foundTransaction.Amount.String())
	}
	if foundTransaction.Currency != transaction.Currency {
		t.Errorf("expected transaction currency %v, got %v", transaction.Currency, foundTransaction.Currency)
	}
	if foundTransaction.Category == nil {
		t.Error("expected category to be loaded")
	}
	if foundTransaction.Subcategory == nil {
		t.Error("expected subcategory to be loaded")
	}
}
