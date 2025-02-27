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

	// Auto migrate the schema
	if err := db.AutoMigrate(&Category{}, &CategoryType{}, &Translation{}); err != nil {
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
		TypeID:   typeID,
		IsActive: true,
	}
	err := store.CreateCategory(context.Background(), category)
	if err != nil {
		t.Fatalf("failed to create test category: %v", err)
	}

	// Add translation for the category
	translation := &Translation{
		EntityID:     category.ID,
		EntityType:   "category",
		LanguageCode: LangEN,
		Name:         name,
		Description:  "Test description",
	}
	err = store.CreateTranslation(context.Background(), translation)
	if err != nil {
		t.Fatalf("failed to create test translation: %v", err)
	}
	return category
}

func createTestCategoryType(t *testing.T, db *gorm.DB, name string) *CategoryType {
	t.Helper()
	categoryType := &CategoryType{}
	result := db.Create(categoryType)
	if result.Error != nil {
		t.Fatalf("failed to create test category type: %v", result.Error)
	}

	// Add translation for the category type
	translation := &Translation{
		EntityID:     categoryType.ID,
		EntityType:   "category_type",
		LanguageCode: LangEN,
		Name:         name,
		Description:  "Test type description",
	}
	result = db.Create(translation)
	if result.Error != nil {
		t.Fatalf("failed to create test translation: %v", result.Error)
	}
	return categoryType
}

func TestSQLStore_CreateCategory(t *testing.T) {
	store, _ := createTestStore(t)

	tests := []struct {
		name     string
		category *Category
		wantErr  bool
	}{
		{
			name: "Successfully_create_valid_category",
			category: &Category{
				TypeID:   1,
				IsActive: true,
			},
			wantErr: false,
		},
		{
			name: "Error_create_category_with_empty_type_id",
			category: &Category{
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
			if got.ID != tt.want.ID || got.GetName(LangEN) != tt.want.GetName(LangEN) {
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

	// Create test categories with translations
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
			name:           "Successfully_filter_by_type",
			typeID:         func() *uint { id := type1.ID; return &id }(),
			expectedCount:  2,
			expectedTypeID: type1.ID,
			wantErr:        false,
		},
		{
			name:           "Successfully_handle_empty_type",
			typeID:         func() *uint { id := uint(999); return &id }(),
			expectedCount:  0,
			expectedTypeID: 999,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			categories, err := store.ListCategories(context.Background(), tt.typeID)
			if err != nil {
				t.Errorf("SQLStore.ListCategories() error = %v", err)
				return
			}

			if len(categories) != tt.expectedCount {
				t.Errorf("SQLStore.ListCategories() got %d categories, want %d", len(categories), tt.expectedCount)
			}

			if tt.typeID != nil {
				for _, cat := range categories {
					if cat.TypeID != *tt.typeID {
						t.Errorf("SQLStore.ListCategories() category typeID = %v, want %v", cat.TypeID, *tt.typeID)
					}
				}
			}
		})
	}
}

func TestSQLStore_CreateTranslation(t *testing.T) {
	store, _ := createTestStore(t)
	category := createTestCategory(t, store, "Test Category", 1)

	tests := []struct {
		name        string
		translation *Translation
		wantErr     bool
	}{
		{
			name: "Successfully_create_valid_translation",
			translation: &Translation{
				EntityID:     category.ID,
				EntityType:   "category",
				LanguageCode: "sv",
				Name:         "Test Name SV",
				Description:  "Test Description SV",
			},
			wantErr: false,
		},
		{
			name: "Error_create_translation_without_language_code",
			translation: &Translation{
				EntityID:   category.ID,
				EntityType: "category",
				Name:       "Test Name",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.CreateTranslation(context.Background(), tt.translation)
			if (err != nil) != tt.wantErr {
				t.Errorf("SQLStore.CreateTranslation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.translation.ID == 0 {
				t.Error("SQLStore.CreateTranslation() translation ID not set")
			}
		})
	}
}

func TestSQLStore_GetTranslations(t *testing.T) {
	store, _ := createTestStore(t)
	category := createTestCategory(t, store, "Test Category", 1)

	// Create test translations
	translations := []Translation{
		{
			EntityID:     category.ID,
			EntityType:   "category",
			LanguageCode: "sv",
			Name:         "Swedish Name",
			Description:  "Swedish Description",
		},
		{
			EntityID:     category.ID,
			EntityType:   "category",
			LanguageCode: "en",
			Name:         "English Name",
			Description:  "English Description",
		},
	}

	for _, tr := range translations {
		if err := store.CreateTranslation(context.Background(), &tr); err != nil {
			t.Fatalf("failed to create test translation: %v", err)
		}
	}

	tests := []struct {
		name          string
		entityID      uint
		entityType    string
		expectedCount int
		wantErr       bool
	}{
		{
			name:          "Successfully_get_translations",
			entityID:      category.ID,
			entityType:    "category",
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name:          "Successfully_handle_no_translations",
			entityID:      999,
			entityType:    "category",
			expectedCount: 0,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetTranslations(context.Background(), tt.entityID, tt.entityType)
			if (err != nil) != tt.wantErr {
				t.Errorf("SQLStore.GetTranslations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.expectedCount {
				t.Errorf("SQLStore.GetTranslations() got %d translations, want %d", len(got), tt.expectedCount)
			}
		})
	}
}

func TestSQLStore_DeleteCategory(t *testing.T) {
	store, _ := createTestStore(t)
	category := createTestCategory(t, store, "Test Category", 1)

	tests := []struct {
		name    string
		id      uint
		wantErr bool
	}{
		{
			name:    "Successfully_delete_existing_category",
			id:      category.ID,
			wantErr: false,
		},
		{
			name:    "Successfully_handle_non_existent_category",
			id:      999,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.DeleteCategory(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("SQLStore.DeleteCategory() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify deletion
			if !tt.wantErr {
				_, err := store.GetCategoryByID(context.Background(), tt.id)
				if !errors.Is(err, ErrNotFound) {
					t.Errorf("SQLStore.DeleteCategory() category still exists after deletion")
				}
			}
		})
	}
}
