package category

import (
	"context"
	"testing"

	"log/slog"

	"github.com/lindehoff/Budget-Assist/internal/ai"
	"github.com/lindehoff/Budget-Assist/internal/db"
)

func createTestManager(t *testing.T) (*Manager, *db.MockStore) {
	t.Helper()
	store := db.NewMockStore()
	logger := slog.Default()
	aiService := &mockAIService{}
	return NewManager(store, aiService, logger), store
}

type mockAIService struct {
	ai.Service
}

func (m *mockAIService) SuggestCategories(ctx context.Context, description string) ([]ai.CategoryMatch, error) {
	return []ai.CategoryMatch{
		{
			Category:   "Test Category/Test Subcategory",
			Confidence: 0.9,
		},
	}, nil
}

func Test_CreateCategory(t *testing.T) {
	tests := []struct {
		name         string
		req          CreateCategoryRequest
		setupStore   func(store *db.MockStore)
		wantErr      string
		validateFunc func(t *testing.T, category *db.Category, store *db.MockStore)
	}{
		{
			name: "Successfully_create_category_with_subcategories",
			req: CreateCategoryRequest{
				Name:        "Test Category",
				Description: "Test Description",
				Type:        "Test Type",
				TypeID:      1,
				Subcategories: []string{
					"Test Subcategory",
				},
			},
			setupStore: func(store *db.MockStore) {
				// Create a category type
				_ = store.CreateCategoryType(context.Background(), &db.CategoryType{
					Name:        "Test Type",
					Description: "Test Type Description",
					IsMultiple:  true,
				})
			},
			validateFunc: func(t *testing.T, category *db.Category, store *db.MockStore) {
				if category == nil {
					t.Error("expected category to be created, got nil")
					return
				}
				if category.Name != "Test Category" {
					t.Errorf("expected category name %q, got %q", "Test Category", category.Name)
				}
				if category.Description != "Test Description" {
					t.Errorf("expected category description %q, got %q", "Test Description", category.Description)
				}
				if category.Type != "Test Type" {
					t.Errorf("expected category type %q, got %q", "Test Type", category.Type)
				}
				if !category.IsActive {
					t.Error("expected category to be active")
				}
				if len(category.Subcategories) != 1 {
					t.Errorf("expected 1 subcategory, got %d", len(category.Subcategories))
					return
				}
			},
		},
		{
			name: "Create_error_missing_name",
			req: CreateCategoryRequest{
				Description: "Test Description",
				Type:        "Test Type",
				TypeID:      1,
			},
			wantErr: "category operation \"create\" failed for \"\": name is required",
		},
		{
			name: "Create_error_missing_type_id",
			req: CreateCategoryRequest{
				Name:        "Test Category",
				Description: "Test Description",
				Type:        "Test Type",
			},
			wantErr: "category operation \"create\" failed for \"Test Category\": type_id is required",
		},
		{
			name: "Successfully_create_category_without_subcategories",
			req: CreateCategoryRequest{
				Name:        "Test Category",
				Description: "Test Description",
				Type:        "Test Type",
				TypeID:      1,
			},
			setupStore: func(store *db.MockStore) {
				// Create a category type
				_ = store.CreateCategoryType(context.Background(), &db.CategoryType{
					Name:        "Test Type",
					Description: "Test Type Description",
					IsMultiple:  true,
				})
			},
			validateFunc: func(t *testing.T, category *db.Category, store *db.MockStore) {
				if category == nil {
					t.Error("expected category to be created, got nil")
					return
				}
				if category.Name != "Test Category" {
					t.Errorf("expected category name %q, got %q", "Test Category", category.Name)
				}
				if len(category.Subcategories) != 0 {
					t.Errorf("expected no subcategories, got %d", len(category.Subcategories))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, store := createTestManager(t)

			if tt.setupStore != nil {
				tt.setupStore(store)
			}

			category, err := manager.CreateCategory(context.Background(), tt.req)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.wantErr)
					return
				}
				if err.Error() != tt.wantErr {
					t.Errorf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, category, store)
			}
		})
	}
}

func Test_CreateSubcategory(t *testing.T) {
	tests := []struct {
		name         string
		req          CreateSubcategoryRequest
		setupStore   func(store *db.MockStore)
		wantErr      string
		validateFunc func(t *testing.T, subcategory *db.Subcategory, store *db.MockStore)
	}{
		{
			name: "Successfully_create_subcategory_with_tags_and_categories",
			req: CreateSubcategoryRequest{
				Name:        "Test Subcategory",
				Description: "Test Description",
				Categories:  []string{"Test Category"},
				Tags:        []string{"tag1", "tag2"},
				IsSystem:    true,
			},
			setupStore: func(store *db.MockStore) {
				// Create a category type
				_ = store.CreateCategoryType(context.Background(), &db.CategoryType{
					Name:        "Test Type",
					Description: "Test Type Description",
					IsMultiple:  true,
				})
				// Create a category
				_ = store.CreateCategory(context.Background(), &db.Category{
					Name:        "Test Category",
					Description: "Test Category Description",
					TypeID:      1,
					Type:        "Test Type",
					IsActive:    true,
				})
			},
			validateFunc: func(t *testing.T, subcategory *db.Subcategory, store *db.MockStore) {
				if subcategory == nil {
					t.Error("expected subcategory to be created, got nil")
					return
				}
				if subcategory.Name != "Test Subcategory" {
					t.Errorf("expected subcategory name %q, got %q", "Test Subcategory", subcategory.Name)
				}
				if subcategory.Description != "Test Description" {
					t.Errorf("expected subcategory description %q, got %q", "Test Description", subcategory.Description)
				}
				if !subcategory.IsSystem {
					t.Error("expected subcategory to be system")
				}
				if !subcategory.IsActive {
					t.Error("expected subcategory to be active")
				}
				if len(subcategory.Tags) != 2 {
					t.Errorf("expected 2 tags, got %d", len(subcategory.Tags))
				}
			},
		},
		{
			name: "Create_error_missing_name",
			req: CreateSubcategoryRequest{
				Description: "Test Description",
				IsSystem:    true,
			},
			wantErr: "category operation \"create_subcategory\" failed for \"\": name is required",
		},
		{
			name: "Successfully_create_subcategory_without_tags",
			req: CreateSubcategoryRequest{
				Name:        "Test Subcategory",
				Description: "Test Description",
				IsSystem:    true,
			},
			validateFunc: func(t *testing.T, subcategory *db.Subcategory, store *db.MockStore) {
				if subcategory == nil {
					t.Error("expected subcategory to be created, got nil")
					return
				}
				if subcategory.Name != "Test Subcategory" {
					t.Errorf("expected subcategory name %q, got %q", "Test Subcategory", subcategory.Name)
				}
				if len(subcategory.Tags) != 0 {
					t.Errorf("expected no tags, got %d", len(subcategory.Tags))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, store := createTestManager(t)

			if tt.setupStore != nil {
				tt.setupStore(store)
			}

			subcategory, err := manager.CreateSubcategory(context.Background(), tt.req)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.wantErr)
					return
				}
				if err.Error() != tt.wantErr {
					t.Errorf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, subcategory, store)
			}
		})
	}
}

func Test_UpdateCategory(t *testing.T) {
	tests := []struct {
		name         string
		req          UpdateCategoryRequest
		setupStore   func(store *db.MockStore) uint
		wantErr      string
		validateFunc func(t *testing.T, category *db.Category, store *db.MockStore)
	}{
		{
			name: "Successfully_update_category_all_fields",
			req: UpdateCategoryRequest{
				Name:               "Updated Category",
				Description:        "Updated Description",
				InstanceIdentifier: "updated-instance",
				AddSubcategories:   []string{"New Subcategory"},
				IsActive:           boolPtr(false),
			},
			setupStore: func(store *db.MockStore) uint {
				// Create a category type
				categoryType := &db.CategoryType{
					Name:        "Test Type",
					Description: "Test Type Description",
					IsMultiple:  true,
				}
				if err := store.CreateCategoryType(context.Background(), categoryType); err != nil {
					t.Fatalf("failed to create category type: %v", err)
				}

				// Create initial category
				category := &db.Category{
					Name:        "Test Category",
					Description: "Test Description",
					TypeID:      categoryType.ID,
					Type:        "Test Type",
					IsActive:    true,
				}
				if err := store.CreateCategory(context.Background(), category); err != nil {
					t.Fatalf("failed to create category: %v", err)
				}
				return category.ID
			},
			validateFunc: func(t *testing.T, category *db.Category, store *db.MockStore) {
				if category == nil {
					t.Error("expected category to be updated, got nil")
					return
				}
				if category.Name != "Updated Category" {
					t.Errorf("expected category name %q, got %q", "Updated Category", category.Name)
				}
				if category.Description != "Updated Description" {
					t.Errorf("expected category description %q, got %q", "Updated Description", category.Description)
				}
				if category.InstanceIdentifier != "updated-instance" {
					t.Errorf("expected instance identifier %q, got %q", "updated-instance", category.InstanceIdentifier)
				}
				if category.IsActive {
					t.Error("expected category to be inactive")
				}
				if len(category.Subcategories) != 1 {
					t.Errorf("expected 1 subcategory, got %d", len(category.Subcategories))
				}
			},
		},
		{
			name: "Update_error_category_not_found",
			req: UpdateCategoryRequest{
				Name: "Updated Category",
			},
			setupStore: func(store *db.MockStore) uint {
				return 999
			},
			wantErr: "category operation \"update\" failed for \"id=999\": record not found",
		},
		{
			name: "Successfully_update_category_partial_fields",
			req: UpdateCategoryRequest{
				Name:        "Updated Category",
				Description: "Updated Description",
			},
			setupStore: func(store *db.MockStore) uint {
				// Create a category type
				categoryType := &db.CategoryType{
					Name:        "Test Type",
					Description: "Test Type Description",
					IsMultiple:  true,
				}
				if err := store.CreateCategoryType(context.Background(), categoryType); err != nil {
					t.Fatalf("failed to create category type: %v", err)
				}

				// Create initial category
				category := &db.Category{
					Name:        "Test Category",
					Description: "Test Description",
					TypeID:      categoryType.ID,
					Type:        "Test Type",
					IsActive:    true,
				}
				if err := store.CreateCategory(context.Background(), category); err != nil {
					t.Fatalf("failed to create category: %v", err)
				}
				return category.ID
			},
			validateFunc: func(t *testing.T, category *db.Category, store *db.MockStore) {
				if category == nil {
					t.Error("expected category to be updated, got nil")
					return
				}
				if category.Name != "Updated Category" {
					t.Errorf("expected category name %q, got %q", "Updated Category", category.Name)
				}
				if category.Description != "Updated Description" {
					t.Errorf("expected category description %q, got %q", "Updated Description", category.Description)
				}
				if !category.IsActive {
					t.Error("expected category to still be active")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, store := createTestManager(t)

			var categoryID uint
			if tt.setupStore != nil {
				categoryID = tt.setupStore(store)
			}

			category, err := manager.UpdateCategory(context.Background(), categoryID, tt.req)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.wantErr)
					return
				}
				if err.Error() != tt.wantErr {
					t.Errorf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, category, store)
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func Test_SuggestCategory(t *testing.T) {
	tests := []struct {
		name         string
		description  string
		setupStore   func(store *db.MockStore)
		wantErr      string
		validateFunc func(t *testing.T, suggestions []CategorySuggestion)
	}{
		{
			name:        "Successfully_suggest_categories",
			description: "Test transaction description",
			validateFunc: func(t *testing.T, suggestions []CategorySuggestion) {
				if len(suggestions) != 1 {
					t.Errorf("expected 1 suggestion, got %d", len(suggestions))
					return
				}
				suggestion := suggestions[0]
				if suggestion.CategoryPath != "Test Category/Test Subcategory" {
					t.Errorf("expected category path %q, got %q", "Test Category/Test Subcategory", suggestion.CategoryPath)
				}
				if suggestion.Confidence != 0.9 {
					t.Errorf("expected confidence %f, got %f", 0.9, suggestion.Confidence)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, store := createTestManager(t)

			if tt.setupStore != nil {
				tt.setupStore(store)
			}

			suggestions, err := manager.SuggestCategory(context.Background(), tt.description)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.wantErr)
					return
				}
				if err.Error() != tt.wantErr {
					t.Errorf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, suggestions)
			}
		})
	}
}
