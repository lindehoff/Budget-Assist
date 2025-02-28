package db

import (
	"context"
	"os"
	"strings"
	"testing"

	"gorm.io/gorm"
)

func Test_ImportDefaultCategories(t *testing.T) {
	tests := []struct {
		name          string
		jsonContent   string
		setupFunc     func(t *testing.T, db *gorm.DB)
		expectedError string
		validateFunc  func(t *testing.T, db *gorm.DB)
	}{
		{
			name: "Successfully_import_default_categories",
			jsonContent: `{
				"categoryTypes": [
					{
						"name": "Test Type",
						"description": "Test Description",
						"isMultiple": true
					}
				],
				"subcategories": [
					{
						"name": "Test Subcategory",
						"description": "Test Subcategory Description",
						"tags": ["tag1", "tag2"]
					}
				],
				"categories": [
					{
						"name": "Test Category",
						"description": "Test Category Description",
						"type": "Test Type",
						"subcategories": ["Test Subcategory"]
					}
				]
			}`,
			validateFunc: func(t *testing.T, db *gorm.DB) {
				// Verify category type
				var categoryType CategoryType
				if err := db.Where("name = ?", "Test Type").First(&categoryType).Error; err != nil {
					t.Errorf("failed to find category type: %v", err)
					return
				}

				// Verify category
				var category Category
				if err := db.Where("name = ?", "Test Category").First(&category).Error; err != nil {
					t.Errorf("failed to find category: %v", err)
					return
				}

				// Verify subcategory
				var subcategory Subcategory
				if err := db.Where("name = ?", "Test Subcategory").First(&subcategory).Error; err != nil {
					t.Errorf("failed to find subcategory: %v", err)
					return
				}

				// Verify category-subcategory link
				var link CategorySubcategory
				if err := db.Where("category_id = ? AND subcategory_id = ?", category.ID, subcategory.ID).First(&link).Error; err != nil {
					t.Errorf("failed to find category-subcategory link: %v", err)
					return
				}

				// Verify tags
				var tags []Tag
				if err := db.Model(&subcategory).Association("Tags").Find(&tags); err != nil {
					t.Errorf("failed to find tags: %v", err)
					return
				}
				if len(tags) != 2 {
					t.Errorf("expected 2 tags, got %d", len(tags))
				}
			},
		},
		{
			name: "Import_error_missing_categories_file",
			// No jsonContent means no file will be created
			expectedError: "failed to read categories file: no file found in any location",
		},
		{
			name: "Import_error_invalid_json",
			jsonContent: `{
				"categoryTypes": [
					{
						"name": "Test Type"
						"description": "Test Description",
					}
				]
			}`,
			expectedError: "failed to parse categories file",
		},
		{
			name: "Import_error_duplicate_category_type",
			jsonContent: `{
				"categoryTypes": [
					{
						"name": "Duplicate Type",
						"description": "Test Description",
						"isMultiple": true
					}
				],
				"subcategories": [],
				"categories": []
			}`,
			setupFunc: func(t *testing.T, db *gorm.DB) {
				categoryType := &CategoryType{
					Name:        "Duplicate Type",
					Description: "Original Description",
					IsMultiple:  false,
				}
				if err := db.Create(categoryType).Error; err != nil {
					t.Fatalf("failed to create category type: %v", err)
				}
			},
			expectedError: "failed to create category type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for each test
			tempDir := t.TempDir()
			origDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("failed to get current directory: %v", err)
			}

			// Change to temp directory and ensure we change back
			if err := os.Chdir(tempDir); err != nil {
				t.Fatalf("failed to change to temp directory: %v", err)
			}
			t.Cleanup(func() {
				if err := os.Chdir(origDir); err != nil {
					t.Errorf("failed to restore working directory: %v", err)
				}
			})

			// Create test database
			store, db := createTestStore(t)
			t.Cleanup(func() {
				if err := store.Close(); err != nil {
					t.Errorf("failed to close store: %v", err)
				}
			})

			// Clean database
			cleanupDatabase(t, db)

			// Run setup if provided
			if tt.setupFunc != nil {
				tt.setupFunc(t, db)
			}

			// Create categories.json if content provided
			if tt.jsonContent != "" {
				if err := os.WriteFile("categories.json", []byte(tt.jsonContent), 0600); err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}
			}

			// Run the import
			err = ImportDefaultCategories(context.TODO(), db)

			// Validate error if expected
			if tt.expectedError != "" {
				if err == nil {
					t.Error("expected error but got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing %q, got %v", tt.expectedError, err)
					return
				}
				return
			}

			// If no error expected, validate success
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Run validation if provided
			if tt.validateFunc != nil {
				tt.validateFunc(t, db)
			}
		})
	}
}

func cleanupDatabase(t *testing.T, db *gorm.DB) {
	tables := []string{
		"category_subcategories",
		"subcategory_tags",
		"categories",
		"subcategories",
		"category_types",
		"tags",
	}

	for _, table := range tables {
		if err := db.Exec("DELETE FROM " + table).Error; err != nil {
			t.Fatalf("failed to clean up %s: %v", table, err)
		}
	}
}
