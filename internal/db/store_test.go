package db

import (
	"context"
	"errors"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func createTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	// Auto migrate all required tables
	if err := db.AutoMigrate(
		&CategoryType{},
		&Category{},
		&Subcategory{},
		&CategorySubcategory{},
		&Tag{},
		&Transaction{},
		&Prompt{},
	); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	return db
}

func createTestStore(t *testing.T) (Store, *gorm.DB) {
	t.Helper()
	db := createTestDB(t)
	return NewStore(db, nil), db
}

func createTestCategory(t *testing.T, store Store, name string, typeID uint) *Category {
	t.Helper()
	category := &Category{
		Name:        name,
		Description: "Test description",
		TypeID:      typeID,
		IsActive:    true,
	}
	err := store.CreateCategory(context.Background(), category)
	if err != nil {
		t.Fatalf("failed to create test category: %v", err)
	}
	return category
}

func createTestCategoryType(t *testing.T, db *gorm.DB, name string) *CategoryType {
	t.Helper()
	categoryType := &CategoryType{
		Name:        name,
		Description: "Test type description",
		IsMultiple:  false,
	}
	result := db.Create(categoryType)
	if result.Error != nil {
		t.Fatalf("failed to create test category type: %v", result.Error)
	}
	return categoryType
}

func TestSQLStore_CreateCategory(t *testing.T) {
	store, db := createTestStore(t)

	// Create a test category type first
	categoryType := createTestCategoryType(t, db, "Test Type")

	tests := []struct {
		name     string
		category *Category
		wantErr  bool
	}{
		{
			name: "Successfully_create_valid_category",
			category: &Category{
				Name:        "Test Category",
				Description: "Test Description",
				TypeID:      categoryType.ID,
				IsActive:    true,
			},
			wantErr: false,
		},
		{
			name: "Error_create_category_with_empty_type_id",
			category: &Category{
				Name:     "Test Category",
				IsActive: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.CreateCategory(context.Background(), tt.category)
			if (err != nil) != tt.wantErr {
				t.Errorf("SQLStore.CreateCategory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.category.ID == 0 {
				t.Error("SQLStore.CreateCategory() category ID not set")
			}
		})
	}
}

func TestSQLStore_GetCategoryByID(t *testing.T) {
	store, _ := createTestStore(t)
	existingCategory := createTestCategory(t, store, "Test Category", 1)

	tests := []struct {
		name    string
		id      uint
		want    *Category
		wantErr error
	}{
		{
			name: "Successfully_get_existing_category",
			id:   existingCategory.ID,
			want: existingCategory,
		},
		{
			name:    "Error_get_non_existent_category",
			id:      999,
			wantErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetCategoryByID(context.Background(), tt.id)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("SQLStore.GetCategoryByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != nil {
				return
			}
			if got.ID != tt.want.ID || got.Name != tt.want.Name {
				t.Errorf("SQLStore.GetCategoryByID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLStore_ListCategories(t *testing.T) {
	store, db := createTestStore(t)

	// Create test category types
	type1 := createTestCategoryType(t, db, "Type 1")
	type2 := createTestCategoryType(t, db, "Type 2")

	// Create test categories
	_ = createTestCategory(t, store, "Category 1", type1.ID)
	_ = createTestCategory(t, store, "Category 2", type1.ID)
	_ = createTestCategory(t, store, "Category 3", type2.ID)

	tests := []struct {
		name           string
		typeID         *uint
		expectedCount  int
		expectedTypeID uint
		wantErr        bool
	}{
		{
			name:          "Successfully_list_all_categories",
			typeID:        nil,
			expectedCount: 3,
			wantErr:       false,
		},
		{
			name:           "Successfully_list_categories_by_type",
			typeID:         &type1.ID,
			expectedCount:  2,
			expectedTypeID: type1.ID,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			categories, err := store.ListCategories(context.Background(), tt.typeID)
			if (err != nil) != tt.wantErr {
				t.Errorf("SQLStore.ListCategories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(categories) != tt.expectedCount {
				t.Errorf("SQLStore.ListCategories() got %d categories, want %d", len(categories), tt.expectedCount)
			}

			if tt.typeID != nil {
				for _, category := range categories {
					if category.TypeID != tt.expectedTypeID {
						t.Errorf("SQLStore.ListCategories() got category with type ID %d, want %d", category.TypeID, tt.expectedTypeID)
					}
				}
			}
		})
	}
}

func TestSQLStore_DeleteCategory(t *testing.T) {
	store, db := createTestStore(t)

	// Create a test category type first
	categoryType := createTestCategoryType(t, db, "Test Type")

	// Create a test category
	existingCategory := createTestCategory(t, store, "Test Category", categoryType.ID)

	tests := []struct {
		name    string
		id      uint
		wantErr error
	}{
		{
			name: "Successfully_delete_existing_category",
			id:   existingCategory.ID,
		},
		{
			name:    "Error_delete_non_existent_category",
			id:      999,
			wantErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.DeleteCategory(context.Background(), tt.id)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("SQLStore.DeleteCategory() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr == nil {
				// Verify category was deleted
				_, err := store.GetCategoryByID(context.Background(), tt.id)
				if !errors.Is(err, ErrNotFound) {
					t.Errorf("SQLStore.DeleteCategory() category still exists after deletion")
				}
			}
		})
	}
}
