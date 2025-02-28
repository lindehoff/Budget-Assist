package db

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gorm.io/gorm"
)

func Test_Successfully_import_default_categories(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for the test
	tempDir := t.TempDir()

	// Get current directory before changing it
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	// Change working directory to the temp dir
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	// Ensure we restore the directory at the end
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	// Create a test database with proper migrations
	store, db := createTestStore(t)
	defer func() {
		if err := store.Close(); err != nil {
			t.Errorf("failed to close store: %v", err)
		}
	}()

	// Clean up any existing data
	cleanupDatabase(t, db)

	// Create the categories.json file
	filePath := filepath.Join("categories.json")
	if err := os.WriteFile(filePath, []byte(`{
		"categoryTypes": [
			{
				"name": "Test Type",
				"description": "Test Description",
				"isMultiple": true
			}
		],
		"categories": [
			{
				"name": "Test Category",
				"description": "Test Category Description",
				"typeId": 1,
				"subcategories": [
					{
						"name": "Test Subcategory",
						"description": "Test Subcategory Description"
					}
				]
			}
		]
	}`), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Run import
	err = ImportDefaultCategories(context.Background(), db)
	if err != nil {
		t.Errorf("ImportDefaultCategories() error = %v", err)
	}

	// Verify that the category type was created
	var categoryType CategoryType
	if err := db.Where("name = ?", "Test Type").First(&categoryType).Error; err != nil {
		t.Errorf("failed to find category type: %v", err)
	}

	// Verify that the category was created
	var category Category
	if err := db.Where("name = ?", "Test Category").First(&category).Error; err != nil {
		t.Errorf("failed to find category: %v", err)
	}

	// Verify that the subcategory was created
	var subcategory Subcategory
	if err := db.Where("name = ?", "Test Subcategory").First(&subcategory).Error; err != nil {
		t.Errorf("failed to find subcategory: %v", err)
	}

	// Verify that the subcategory is linked to the category
	var link CategorySubcategory
	if err := db.Where("category_id = ? AND subcategory_id = ?", category.ID, subcategory.ID).First(&link).Error; err != nil {
		t.Errorf("failed to find category-subcategory link: %v", err)
	}
}

func cleanupDatabase(t *testing.T, db *gorm.DB) {
	// Clean up any existing data
	if err := db.Exec("DELETE FROM category_subcategories").Error; err != nil {
		t.Fatalf("failed to clean up category_subcategories: %v", err)
	}
	if err := db.Exec("DELETE FROM subcategory_tags").Error; err != nil {
		t.Fatalf("failed to clean up subcategory_tags: %v", err)
	}
	if err := db.Exec("DELETE FROM categories").Error; err != nil {
		t.Fatalf("failed to clean up categories: %v", err)
	}
	if err := db.Exec("DELETE FROM subcategories").Error; err != nil {
		t.Fatalf("failed to clean up subcategories: %v", err)
	}
	if err := db.Exec("DELETE FROM category_types").Error; err != nil {
		t.Fatalf("failed to clean up category_types: %v", err)
	}
	if err := db.Exec("DELETE FROM tags").Error; err != nil {
		t.Fatalf("failed to clean up tags: %v", err)
	}
}

func Test_Import_error_missing_categories_file(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for the test
	tempDir := t.TempDir()

	// Get current directory before changing it
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	// Change working directory to the temp dir
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	// Ensure we restore the directory at the end
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	// Create a test database with proper migrations
	store, db := createTestStore(t)
	defer func() {
		if err := store.Close(); err != nil {
			t.Errorf("failed to close store: %v", err)
		}
	}()

	// Clean up any existing data
	cleanupDatabase(t, db)

	// Run import without creating the categories.json file
	err = ImportDefaultCategories(context.Background(), db)

	if err == nil {
		t.Fatal("ImportDefaultCategories() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "failed to read categories file: no file found in any location") {
		t.Errorf("ImportDefaultCategories() error = %v, want error containing 'failed to read categories file: no file found in any location'", err)
	}
}

func Test_Import_error_invalid_json(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for the test
	tempDir := t.TempDir()

	// Get current directory before changing it
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	// Change working directory to the temp dir
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	// Ensure we restore the directory at the end
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	// Create a test database with proper migrations
	store, db := createTestStore(t)
	defer func() {
		if err := store.Close(); err != nil {
			t.Errorf("failed to close store: %v", err)
		}
	}()

	// Clean up any existing data
	cleanupDatabase(t, db)

	// Create an invalid categories.json file
	filePath := filepath.Join("categories.json")
	if err := os.WriteFile(filePath, []byte(`{
		"categoryTypes": [
			{
				"name": "Test Type"
				"description": "Test Description",
				"isMultiple": true
			}
		]
	}`), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Run import
	err = ImportDefaultCategories(context.Background(), db)

	if err == nil {
		t.Error("ImportDefaultCategories() error = nil, want error")
		return
	}
	if !strings.Contains(err.Error(), "failed to parse categories file") {
		t.Errorf("ImportDefaultCategories() error = %v, want error containing 'failed to parse categories file'", err)
	}
}

func Test_Import_error_create_category_type(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for the test
	tempDir := t.TempDir()

	// Get current directory before changing it
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	// Change working directory to the temp dir
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	// Ensure we restore the directory at the end
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	// Create a test database with proper migrations
	store, db := createTestStore(t)
	defer func() {
		if err := store.Close(); err != nil {
			t.Errorf("failed to close store: %v", err)
		}
	}()

	// Clean up any existing data
	cleanupDatabase(t, db)

	// Create a category type that will cause a unique constraint violation
	categoryType := &CategoryType{
		Name:        "Invalid Type",
		Description: "Test Description",
		IsMultiple:  true,
	}
	if err := db.Create(categoryType).Error; err != nil {
		t.Fatalf("failed to create category type: %v", err)
	}

	// Create the categories.json file with a duplicate category type
	filePath := filepath.Join("categories.json")
	if err := os.WriteFile(filePath, []byte(`{
		"categoryTypes": [
			{
				"name": "Invalid Type",
				"description": "Different Description",
				"isMultiple": false
			}
		],
		"categories": []
	}`), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Run import
	err = ImportDefaultCategories(context.Background(), db)

	if err == nil {
		t.Error("ImportDefaultCategories() error = nil, want error")
		return
	}
	if !strings.Contains(err.Error(), "failed to create category type") {
		t.Errorf("ImportDefaultCategories() error = %v, want error containing 'failed to create category type'", err)
	}
}

func Test_Import_error_create_subcategory(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for the test
	tempDir := t.TempDir()

	// Get current directory before changing it
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	// Change working directory to the temp dir
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	// Ensure we restore the directory at the end
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	// Create a test database with proper migrations
	store, db := createTestStore(t)
	defer func() {
		if err := store.Close(); err != nil {
			t.Errorf("failed to close store: %v", err)
		}
	}()

	// Clean up any existing data
	cleanupDatabase(t, db)

	// Create a subcategory that will cause a unique constraint violation
	subcategory := &Subcategory{
		Name:        "Invalid Subcategory",
		Description: "Test Description",
		IsSystem:    true,
		IsActive:    true,
	}
	if err := db.Create(subcategory).Error; err != nil {
		t.Fatalf("failed to create subcategory: %v", err)
	}

	// Create the categories.json file with a duplicate subcategory
	filePath := filepath.Join("categories.json")
	if err := os.WriteFile(filePath, []byte(`{
		"categoryTypes": [
			{
				"name": "Test Type",
				"description": "Test Description",
				"isMultiple": true
			}
		],
		"categories": [
			{
				"name": "Test Category",
				"description": "Test Category Description",
				"typeId": 1,
				"subcategories": [
					{
						"name": "Invalid Subcategory",
						"description": "Different Description"
					}
				]
			}
		]
	}`), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Run import
	err = ImportDefaultCategories(context.Background(), db)

	if err == nil {
		t.Error("ImportDefaultCategories() error = nil, want error")
		return
	}
	if !strings.Contains(err.Error(), "failed to create subcategory") {
		t.Errorf("ImportDefaultCategories() error = %v, want error containing 'failed to create subcategory'", err)
	}
}

func Test_Import_error_create_category(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for the test
	tempDir := t.TempDir()

	// Get current directory before changing it
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	// Change working directory to the temp dir
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	// Ensure we restore the directory at the end
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	// Create a test database with proper migrations
	store, db := createTestStore(t)
	defer func() {
		if err := store.Close(); err != nil {
			t.Errorf("failed to close store: %v", err)
		}
	}()

	// Clean up any existing data
	cleanupDatabase(t, db)

	// Create a category that will cause a unique constraint violation
	category := &Category{
		Name:        "Test Category",
		Description: "Test Description",
		TypeID:      1,
		IsActive:    true,
	}
	if err := db.Create(category).Error; err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	// Create the categories.json file with a duplicate category
	filePath := filepath.Join("categories.json")
	if err := os.WriteFile(filePath, []byte(`{
		"categoryTypes": [
			{
				"name": "Test Type",
				"description": "Test Description",
				"isMultiple": true
			}
		],
		"categories": [
			{
				"name": "Test Category",
				"description": "Different Description",
				"typeId": 1,
				"subcategories": [
					{
						"name": "Test Subcategory",
						"description": "Test Subcategory Description"
					}
				]
			}
		]
	}`), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Run import
	err = ImportDefaultCategories(context.Background(), db)

	if err == nil {
		t.Error("ImportDefaultCategories() error = nil, want error")
		return
	}
	if !strings.Contains(err.Error(), "failed to create category") {
		t.Errorf("ImportDefaultCategories() error = %v, want error containing 'failed to create category'", err)
	}
}
