package db

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// mockCategoryImporter is a mock implementation of CategoryImporter for testing
type mockCategoryImporter struct {
	categoryTypes []*CategoryType
	categories    []*Category
	subcategories []*Subcategory
	translations  []*Translation
	shouldFail    bool
	failOperation string
}

func (m *mockCategoryImporter) CreateCategory(ctx context.Context, name, description string, typeID uint, translations map[string]TranslationData, subcategoryIDs []uint) (*Category, error) {
	if m.shouldFail && m.failOperation == "CreateCategory" {
		return nil, fmt.Errorf("mock create category error")
	}
	cat := &Category{
		Name:        name,
		Description: description,
		TypeID:      typeID,
		IsActive:    true,
	}
	nextID := len(m.categories) + 1
	if nextID <= 0 || uint(nextID) > math.MaxUint {
		return nil, fmt.Errorf("integer overflow when generating category ID")
	}
	cat.ID = uint(nextID)
	m.categories = append(m.categories, cat)

	// Create translations
	for lang, trans := range translations {
		translation := &Translation{
			EntityID:     cat.ID,
			EntityType:   string(EntityTypeCategory),
			LanguageCode: lang,
			Name:         trans.Name,
			Description:  trans.Description,
		}
		if err := m.CreateTranslation(ctx, translation); err != nil {
			return nil, err
		}
	}

	return cat, nil
}

func (m *mockCategoryImporter) CreateSubcategory(ctx context.Context, name, description string, isSystem bool, translations map[string]TranslationData) (*Subcategory, error) {
	if m.shouldFail && m.failOperation == "CreateSubcategory" {
		return nil, fmt.Errorf("mock create subcategory error")
	}
	subcat := &Subcategory{
		Name:        translations[LangEN].Name, // Use English name as base name
		Description: description,
		IsSystem:    isSystem,
		IsActive:    true,
	}
	nextID := len(m.subcategories) + 1
	if nextID <= 0 || uint(nextID) > math.MaxUint {
		return nil, fmt.Errorf("integer overflow when generating subcategory ID")
	}
	subcat.ID = uint(nextID)
	m.subcategories = append(m.subcategories, subcat)

	// Create translations
	for lang, trans := range translations {
		translation := &Translation{
			EntityID:     subcat.ID,
			EntityType:   string(EntityTypeSubcategory),
			LanguageCode: lang,
			Name:         trans.Name,
			Description:  trans.Description,
		}
		if err := m.CreateTranslation(ctx, translation); err != nil {
			return nil, err
		}
	}

	return subcat, nil
}

func (m *mockCategoryImporter) CreateCategoryType(ctx context.Context, categoryType *CategoryType) error {
	if m.shouldFail && m.failOperation == "CreateCategoryType" {
		return fmt.Errorf("mock create category type error")
	}
	nextID := len(m.categoryTypes) + 1
	if nextID <= 0 || uint(nextID) > math.MaxUint {
		return fmt.Errorf("integer overflow when generating category type ID")
	}
	categoryType.ID = uint(nextID)
	m.categoryTypes = append(m.categoryTypes, categoryType)
	return nil
}

func (m *mockCategoryImporter) CreateTranslation(ctx context.Context, translation *Translation) error {
	if m.shouldFail && m.failOperation == "CreateTranslation" {
		return fmt.Errorf("mock create translation error")
	}
	m.translations = append(m.translations, translation)
	return nil
}

func createTestCategoriesFile(t *testing.T, data *DefaultCategoriesData) string {
	t.Helper()

	// Create a temporary directory for the test
	tempDir := t.TempDir()
	defaultsDir := filepath.Join(tempDir, "defaults")
	if err := os.MkdirAll(defaultsDir, 0755); err != nil {
		t.Fatalf("failed to create defaults directory: %v", err)
	}

	// Create the categories.json file
	filePath := filepath.Join(defaultsDir, "categories.json")
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Change working directory to the temp dir
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	return tempDir
}

func Test_Successfully_import_default_categories(t *testing.T) {
	t.Parallel()

	testData := &DefaultCategoriesData{
		CategoryTypes: []struct {
			Name         string                     `json:"name"`
			Description  string                     `json:"description"`
			IsMultiple   bool                       `json:"isMultiple"`
			Translations map[string]TranslationData `json:"translations"`
		}{
			{
				Name:        "Test Type",
				Description: "Test Description",
				IsMultiple:  true,
				Translations: map[string]TranslationData{
					LangEN: {Name: "Test Type EN", Description: "Test Description EN"},
					"sv":   {Name: "Test Type SV", Description: "Test Description SV"},
				},
			},
		},
		Categories: []struct {
			Name          string                     `json:"name"`
			Description   string                     `json:"description"`
			TypeID        uint                       `json:"typeId"`
			Translations  map[string]TranslationData `json:"translations"`
			Subcategories []struct {
				Name         string                     `json:"name"`
				Description  string                     `json:"description"`
				Translations map[string]TranslationData `json:"translations"`
			} `json:"subcategories"`
		}{
			{
				Name:        "Test Category",
				Description: "Test Category Description",
				TypeID:      1,
				Translations: map[string]TranslationData{
					LangEN: {Name: "Test Category EN", Description: "Test Category Description EN"},
					"sv":   {Name: "Test Category SV", Description: "Test Category Description SV"},
				},
				Subcategories: []struct {
					Name         string                     `json:"name"`
					Description  string                     `json:"description"`
					Translations map[string]TranslationData `json:"translations"`
				}{
					{
						Name:        "Test Subcategory",
						Description: "Test Subcategory Description",
						Translations: map[string]TranslationData{
							LangEN: {Name: "Test Subcategory EN", Description: "Test Subcategory Description EN"},
							"sv":   {Name: "Test Subcategory SV", Description: "Test Subcategory Description SV"},
						},
					},
				},
			},
		},
	}

	// Create test categories file
	tempDir := createTestCategoriesFile(t, testData)
	defer func() {
		if err := os.Chdir(tempDir); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	// Create mock importer
	importer := &mockCategoryImporter{}

	// Run import
	if err := ImportDefaultCategories(context.Background(), nil, importer); err != nil {
		t.Fatalf("ImportDefaultCategories() error = %v", err)
	}

	// Verify category types were created
	if len(importer.categoryTypes) != 1 {
		t.Errorf("expected 1 category type, got %d", len(importer.categoryTypes))
	}
	if ct := importer.categoryTypes[0]; ct.Name != "Test Type" {
		t.Errorf("category type name = %q, want %q", ct.Name, "Test Type")
	}

	// Verify translations were created
	expectedTranslations := 6 // 2 for category type, 2 for category, 2 for subcategory
	if len(importer.translations) != expectedTranslations {
		t.Errorf("expected %d translations, got %d", expectedTranslations, len(importer.translations))
		for i, trans := range importer.translations {
			t.Logf("Translation %d: EntityType=%s, LanguageCode=%s, Name=%s", i+1, trans.EntityType, trans.LanguageCode, trans.Name)
		}
	}

	// Count translations by entity type
	translationsByType := make(map[string]int)
	for _, trans := range importer.translations {
		translationsByType[trans.EntityType]++
	}
	for entityType, count := range translationsByType {
		if count != 2 {
			t.Errorf("expected 2 translations for %s, got %d", entityType, count)
		}
	}

	// Verify subcategories were created
	if len(importer.subcategories) != 1 {
		t.Errorf("expected 1 subcategory, got %d", len(importer.subcategories))
	} else if sc := importer.subcategories[0]; sc.Name != "Test Subcategory EN" {
		t.Errorf("subcategory name = %q, want %q", sc.Name, "Test Subcategory EN")
	}

	// Verify categories were created
	if len(importer.categories) != 1 {
		t.Errorf("expected 1 category, got %d", len(importer.categories))
	} else if c := importer.categories[0]; c.Name != "Test Category" {
		t.Errorf("category name = %q, want %q", c.Name, "Test Category")
	}
}

func Test_Import_error_missing_categories_file(t *testing.T) {
	t.Parallel()

	// Create a temporary directory without categories.json
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(tempDir); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	importer := &mockCategoryImporter{}
	err := ImportDefaultCategories(context.Background(), nil, importer)
	if err == nil {
		t.Error("expected error for missing categories.json, got nil")
	}
}

func Test_Import_error_invalid_json(t *testing.T) {
	t.Parallel()

	// Create a temporary directory
	tempDir := t.TempDir()
	defaultsDir := filepath.Join(tempDir, "defaults")
	if err := os.MkdirAll(defaultsDir, 0755); err != nil {
		t.Fatalf("failed to create defaults directory: %v", err)
	}

	// Create invalid JSON file
	filePath := filepath.Join(defaultsDir, "categories.json")
	if err := os.WriteFile(filePath, []byte("{invalid json"), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(tempDir); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	importer := &mockCategoryImporter{}
	err := ImportDefaultCategories(context.Background(), nil, importer)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func Test_Import_error_create_category_type(t *testing.T) {
	t.Parallel()

	testData := &DefaultCategoriesData{
		CategoryTypes: []struct {
			Name         string                     `json:"name"`
			Description  string                     `json:"description"`
			IsMultiple   bool                       `json:"isMultiple"`
			Translations map[string]TranslationData `json:"translations"`
		}{
			{
				Name:        "Test Type",
				Description: "Test Description",
				IsMultiple:  true,
				Translations: map[string]TranslationData{
					LangEN: {Name: "Test Type EN", Description: "Test Description EN"},
				},
			},
		},
	}

	tempDir := createTestCategoriesFile(t, testData)
	defer func() {
		if err := os.Chdir(tempDir); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	importer := &mockCategoryImporter{
		shouldFail:    true,
		failOperation: "CreateCategoryType",
	}

	err := ImportDefaultCategories(context.Background(), nil, importer)
	if err == nil {
		t.Error("expected error for create category type failure, got nil")
	}
}

func Test_Import_error_create_translation(t *testing.T) {
	t.Parallel()

	testData := &DefaultCategoriesData{
		CategoryTypes: []struct {
			Name         string                     `json:"name"`
			Description  string                     `json:"description"`
			IsMultiple   bool                       `json:"isMultiple"`
			Translations map[string]TranslationData `json:"translations"`
		}{
			{
				Name:        "Test Type",
				Description: "Test Description",
				IsMultiple:  true,
				Translations: map[string]TranslationData{
					LangEN: {Name: "Test Type EN", Description: "Test Description EN"},
				},
			},
		},
	}

	tempDir := createTestCategoriesFile(t, testData)
	defer func() {
		if err := os.Chdir(tempDir); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	importer := &mockCategoryImporter{
		shouldFail:    true,
		failOperation: "CreateTranslation",
	}

	err := ImportDefaultCategories(context.Background(), nil, importer)
	if err == nil {
		t.Error("expected error for create translation failure, got nil")
	}
}

func Test_Import_error_create_subcategory(t *testing.T) {
	t.Parallel()

	testData := &DefaultCategoriesData{
		Categories: []struct {
			Name          string                     `json:"name"`
			Description   string                     `json:"description"`
			TypeID        uint                       `json:"typeId"`
			Translations  map[string]TranslationData `json:"translations"`
			Subcategories []struct {
				Name         string                     `json:"name"`
				Description  string                     `json:"description"`
				Translations map[string]TranslationData `json:"translations"`
			} `json:"subcategories"`
		}{
			{
				Name:   "Test Category",
				TypeID: 1,
				Subcategories: []struct {
					Name         string                     `json:"name"`
					Description  string                     `json:"description"`
					Translations map[string]TranslationData `json:"translations"`
				}{
					{
						Name: "Test Subcategory",
						Translations: map[string]TranslationData{
							LangEN: {Name: "Test Subcategory EN", Description: "Test Description EN"},
						},
					},
				},
			},
		},
	}

	tempDir := createTestCategoriesFile(t, testData)
	defer func() {
		if err := os.Chdir(tempDir); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	importer := &mockCategoryImporter{
		shouldFail:    true,
		failOperation: "CreateSubcategory",
	}

	err := ImportDefaultCategories(context.Background(), nil, importer)
	if err == nil {
		t.Error("expected error for create subcategory failure, got nil")
	}
}

func Test_Import_error_create_category(t *testing.T) {
	t.Parallel()

	testData := &DefaultCategoriesData{
		Categories: []struct {
			Name          string                     `json:"name"`
			Description   string                     `json:"description"`
			TypeID        uint                       `json:"typeId"`
			Translations  map[string]TranslationData `json:"translations"`
			Subcategories []struct {
				Name         string                     `json:"name"`
				Description  string                     `json:"description"`
				Translations map[string]TranslationData `json:"translations"`
			} `json:"subcategories"`
		}{
			{
				Name:   "Test Category",
				TypeID: 1,
				Translations: map[string]TranslationData{
					LangEN: {Name: "Test Category EN", Description: "Test Description EN"},
				},
			},
		},
	}

	tempDir := createTestCategoriesFile(t, testData)
	defer func() {
		if err := os.Chdir(tempDir); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	importer := &mockCategoryImporter{
		shouldFail:    true,
		failOperation: "CreateCategory",
	}

	err := ImportDefaultCategories(context.Background(), nil, importer)
	if err == nil {
		t.Error("expected error for create category failure, got nil")
	}
}
