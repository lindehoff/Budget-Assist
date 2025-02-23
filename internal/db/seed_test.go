package db

import (
	"testing"
)

func Test_Successfully_seed_predefined_categories(t *testing.T) {
	// Test setup
	ctx, db, cleanup := setupTestDB(t)
	defer cleanup()

	// Execute seeding
	if err := SeedPredefinedCategories(ctx, db); err != nil {
		t.Fatalf("failed to seed predefined categories: %v", err)
	}

	// Define test cases
	testCases := []struct {
		name          string
		categoryType  string
		isMultiple    bool
		translation   string
		langCode      string
		subcategories []string
	}{
		{
			name:         "Vehicle category type with translations and subcategories",
			categoryType: "Vehicle",
			isMultiple:   true,
			translation:  "Fordon",
			langCode:     LangSV,
			subcategories: []string{
				"Tax",
				"Inspection",
				"Insurance",
				"Tires",
				"Repairs and Service",
				"Parking",
				"Other",
				"Loan Interest",
				"Loan Amortization",
			},
		},
		{
			name:         "Income category type with translations",
			categoryType: "Income",
			isMultiple:   true,
			translation:  "Inkomster",
			langCode:     LangSV,
			subcategories: []string{
				"Salary",
				"Sales",
				"Gift",
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Verify category type
			var categoryType CategoryType
			result := db.WithContext(ctx).Where("name = ?", tc.categoryType).First(&categoryType)
			if result.Error != nil {
				t.Fatalf("failed to find category type %s: %v", tc.categoryType, result.Error)
			}
			if categoryType.IsMultiple != tc.isMultiple {
				t.Errorf("expected IsMultiple=%v for %s, got %v", tc.isMultiple, tc.categoryType, categoryType.IsMultiple)
			}

			// Verify translation
			var translations []Translation
			result = db.WithContext(ctx).Where("entity_type = ? AND entity_id = ? AND language_code = ?",
				EntityTypeCategoryType, categoryType.ID, tc.langCode).Find(&translations)
			if result.Error != nil {
				t.Errorf("failed to find translations: %v", result.Error)
			}
			if len(translations) != 1 {
				t.Errorf("expected 1 translation, got %d", len(translations))
			}
			if translations[0].Name != tc.translation {
				t.Errorf("expected translation %q, got %q", tc.translation, translations[0].Name)
			}

			// Verify subcategories
			var subcategories []Subcategory
			result = db.WithContext(ctx).Where("category_type_id = ?", categoryType.ID).Find(&subcategories)
			if result.Error != nil {
				t.Errorf("failed to find subcategories: %v", result.Error)
			}
			if len(subcategories) != len(tc.subcategories) {
				t.Errorf("expected %d subcategories, got %d", len(tc.subcategories), len(subcategories))
			}

			// Verify each subcategory exists
			for _, expectedSub := range tc.subcategories {
				found := false
				for _, sub := range subcategories {
					if sub.Name == expectedSub {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("subcategory %q not found", expectedSub)
				}
			}
		})
	}
}

func Test_Seed_error_invalid_category_type(t *testing.T) {
	ctx, db, cleanup := setupTestDB(t)
	defer cleanup()

	// Temporarily replace predefinedCategories with invalid data
	original := predefinedCategories
	predefinedCategories = []CategoryTypeData{
		{
			Name:        "", // Invalid empty name
			IsMultiple:  true,
			Description: "Invalid category",
		},
	}
	defer func() { predefinedCategories = original }()

	err := SeedPredefinedCategories(ctx, db)
	if err == nil {
		t.Error("expected error for invalid category type, got nil")
		return
	}

	// Verify error type
	dbErr, ok := err.(*DatabaseOperationError)
	if !ok {
		t.Errorf("expected DatabaseOperationError, got %T", err)
		return
	}
	if dbErr.Operation != "validate" || dbErr.Entity != "category_type" {
		t.Errorf("expected validate operation on category_type, got %s on %s", dbErr.Operation, dbErr.Entity)
	}
}

func Test_Seed_error_invalid_subcategory(t *testing.T) {
	ctx, db, cleanup := setupTestDB(t)
	defer cleanup()

	// Temporarily replace predefinedCategories with invalid subcategory data
	original := predefinedCategories
	predefinedCategories = []CategoryTypeData{
		{
			Name:        "Test Category",
			IsMultiple:  true,
			Description: "Test category with invalid subcategory",
			Translations: map[string]string{
				LangSV: "Test Kategori",
			},
			Subcategories: []SubcategoryData{
				{
					Name: "", // Invalid empty name
				},
			},
		},
	}
	defer func() { predefinedCategories = original }()

	err := SeedPredefinedCategories(ctx, db)
	if err == nil {
		t.Error("expected error for invalid subcategory, got nil")
		return
	}

	// Verify error type
	dbErr, ok := err.(*DatabaseOperationError)
	if !ok {
		t.Errorf("expected DatabaseOperationError, got %T", err)
		return
	}
	if dbErr.Operation != "validate" || dbErr.Entity != "subcategory" {
		t.Errorf("expected validate operation on subcategory, got %s on %s", dbErr.Operation, dbErr.Entity)
	}
}

func Test_Successfully_seed_idempotent_operation(t *testing.T) {
	ctx, db, cleanup := setupTestDB(t)
	defer cleanup()

	// First seeding
	if err := SeedPredefinedCategories(ctx, db); err != nil {
		t.Fatalf("first seeding failed: %v", err)
	}

	// Count categories after first seeding
	var count1 int64
	if err := db.WithContext(ctx).Model(&CategoryType{}).Count(&count1).Error; err != nil {
		t.Fatalf("failed to count categories after first seeding: %v", err)
	}

	// Second seeding
	if err := SeedPredefinedCategories(ctx, db); err != nil {
		t.Fatalf("second seeding failed: %v", err)
	}

	// Count categories after second seeding
	var count2 int64
	if err := db.WithContext(ctx).Model(&CategoryType{}).Count(&count2).Error; err != nil {
		t.Fatalf("failed to count categories after second seeding: %v", err)
	}

	// Verify counts are the same
	if count1 != count2 {
		t.Errorf("expected same number of categories after second seeding, got %d != %d", count1, count2)
	}
}
