package category

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"log/slog"

	"github.com/lindehoff/Budget-Assist/internal/ai"
	"github.com/lindehoff/Budget-Assist/internal/db"
)

// mockStore implements db.Store interface for testing
type mockStore struct {
	categories   map[uint]*db.Category
	translations map[uint][]db.Translation
	nextID       uint
}

func newMockStore() db.Store {
	return &mockStore{
		categories:   make(map[uint]*db.Category),
		translations: make(map[uint][]db.Translation),
		nextID:       1,
	}
}

func (m *mockStore) CreateCategory(_ context.Context, category *db.Category) error {
	category.ID = m.nextID
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()
	m.categories[category.ID] = category
	m.nextID++
	return nil
}

func (m *mockStore) UpdateCategory(_ context.Context, category *db.Category) error {
	if _, exists := m.categories[category.ID]; !exists {
		return db.ErrNotFound
	}
	category.UpdatedAt = time.Now()
	m.categories[category.ID] = category
	return nil
}

func (m *mockStore) GetCategoryByID(_ context.Context, id uint) (*db.Category, error) {
	if category, exists := m.categories[id]; exists {
		return category, nil
	}
	return nil, db.ErrNotFound
}

func (m *mockStore) ListCategories(_ context.Context, typeID *uint) ([]db.Category, error) {
	var categories []db.Category
	for _, cat := range m.categories {
		if typeID == nil || cat.TypeID == *typeID {
			categories = append(categories, *cat)
		}
	}
	return categories, nil
}

func (m *mockStore) CreateTranslation(_ context.Context, translation *db.Translation) error {
	m.translations[translation.EntityID] = append(m.translations[translation.EntityID], *translation)
	return nil
}

func (m *mockStore) GetTranslations(_ context.Context, entityID uint, entityType string) ([]db.Translation, error) {
	return m.translations[entityID], nil
}

func (m *mockStore) DeleteCategory(_ context.Context, id uint) error {
	if _, exists := m.categories[id]; !exists {
		return db.ErrNotFound
	}
	delete(m.categories, id)
	return nil
}

func (m *mockStore) GetCategoryTypeByID(_ context.Context, id uint) (*db.CategoryType, error) {
	// For testing, we'll return a simple category type
	if id == 1 {
		return &db.CategoryType{
			ID:          1,
			Name:        "Expense",
			Description: "Expense category type",
		}, nil
	}
	return nil, db.ErrNotFound
}

type mockAIService struct{}

func (m *mockAIService) AnalyzeTransaction(_ context.Context, _ *db.Transaction) (*ai.Analysis, error) {
	return &ai.Analysis{
		Remarks: "Test analysis",
		Score:   0.95,
	}, nil
}

func (m *mockAIService) ExtractDocument(_ context.Context, _ *ai.Document) (*ai.Extraction, error) {
	return &ai.Extraction{
		Content: "Test extraction",
	}, nil
}

func (m *mockAIService) SuggestCategories(_ context.Context, _ string) ([]ai.CategoryMatch, error) {
	return []ai.CategoryMatch{
		{
			Category:   "expenses.food",
			Confidence: 0.95,
		},
		{
			Category:   "expenses.groceries",
			Confidence: 0.85,
		},
	}, nil
}

// createTestManager creates a new manager with mock dependencies for testing
func createTestManager(t *testing.T) (*Manager, db.Store) {
	t.Helper()
	store := newMockStore()
	aiService := &mockAIService{}
	logger := slog.Default()
	return NewManager(store, aiService, logger), store
}

// createTestCategory creates a test category with the given details
func createTestCategory(t *testing.T, store db.Store, name string, typeID uint) *db.Category {
	t.Helper()
	category := &db.Category{
		Name:        name,
		Description: "Test description",
		TypeID:      typeID,
		IsActive:    true,
	}
	if err := store.CreateCategory(context.Background(), category); err != nil {
		t.Fatalf("failed to create test category: %v", err)
	}
	return category
}

func TestCreateCategory(t *testing.T) {
	manager, _ := createTestManager(t)

	tests := []struct {
		name        string
		req         CreateCategoryRequest
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Successfully_create_valid_category",
			req: CreateCategoryRequest{
				Name:        "Food",
				Description: "Food expenses",
				TypeID:      1,
			},
			wantErr: false,
		},
		{
			name: "Error_create_category_with_empty_name",
			req: CreateCategoryRequest{
				Description: "Test category",
				TypeID:      1,
			},
			wantErr:     true,
			expectedErr: fmt.Errorf("name is required"),
		},
		{
			name: "Error_create_category_with_zero_type_id",
			req: CreateCategoryRequest{
				Name:        "Test",
				Description: "Test category",
			},
			wantErr:     true,
			expectedErr: fmt.Errorf("type ID is required"),
		},
		{
			name: "Successfully_create_category_with_translations",
			req: CreateCategoryRequest{
				Name:        "Food",
				Description: "Food expenses",
				TypeID:      1,
				Translations: map[string]TranslationData{
					"sv": {
						Name:        "Mat",
						Description: "Matkostnader",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category, err := manager.CreateCategory(context.Background(), tt.req)

			// Validate error cases
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCategory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				var categoryErr CategoryError
				if !errors.As(err, &categoryErr) {
					t.Errorf("CreateCategory() error type = %T, want CategoryError", err)
					return
				}
				if tt.expectedErr != nil && categoryErr.Err.Error() != tt.expectedErr.Error() {
					t.Errorf("CreateCategory() error = %v, expectedErr %v", categoryErr.Err, tt.expectedErr)
				}
				return
			}

			// Validate success cases
			if category == nil {
				t.Fatal("CreateCategory() returned nil category when no error expected")
			}

			// Validate category fields
			if category.Name != tt.req.Name {
				t.Errorf("CreateCategory() category name = %v, want %v", category.Name, tt.req.Name)
			}
			if category.Description != tt.req.Description {
				t.Errorf("CreateCategory() category description = %v, want %v", category.Description, tt.req.Description)
			}
			if category.TypeID != tt.req.TypeID {
				t.Errorf("CreateCategory() category typeID = %v, want %v", category.TypeID, tt.req.TypeID)
			}
			if !category.IsActive {
				t.Error("CreateCategory() category should be active by default")
			}
		})
	}
}

func TestUpdateCategory(t *testing.T) {
	manager, store := createTestManager(t)
	existingCategory := createTestCategory(t, store, "Original", 1)

	tests := []struct {
		name        string
		id          uint
		req         UpdateCategoryRequest
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Successfully_update_existing_category",
			id:   existingCategory.ID,
			req: UpdateCategoryRequest{
				Name:        "Updated",
				Description: "Updated description",
			},
			wantErr: false,
		},
		{
			name: "Successfully_update_category_active_status",
			id:   existingCategory.ID,
			req: UpdateCategoryRequest{
				IsActive: func() *bool { b := false; return &b }(),
			},
			wantErr: false,
		},
		{
			name: "Error_update_non_existent_category",
			id:   999,
			req: UpdateCategoryRequest{
				Name: "Test",
			},
			wantErr:     true,
			expectedErr: db.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category, err := manager.UpdateCategory(context.Background(), tt.id, tt.req)

			// Validate error cases
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateCategory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				var categoryErr CategoryError
				if !errors.As(err, &categoryErr) {
					t.Errorf("UpdateCategory() error type = %T, want CategoryError", err)
					return
				}
				if tt.expectedErr != nil && !errors.Is(categoryErr.Err, tt.expectedErr) {
					t.Errorf("UpdateCategory() error = %v, expectedErr %v", categoryErr.Err, tt.expectedErr)
				}
				return
			}

			// Validate success cases
			if category == nil {
				t.Fatal("UpdateCategory() returned nil category when no error expected")
			}

			// Validate updated fields
			if tt.req.Name != "" && category.Name != tt.req.Name {
				t.Errorf("UpdateCategory() category name = %v, want %v", category.Name, tt.req.Name)
			}
			if tt.req.Description != "" && category.Description != tt.req.Description {
				t.Errorf("UpdateCategory() category description = %v, want %v", category.Description, tt.req.Description)
			}
			if tt.req.IsActive != nil && category.IsActive != *tt.req.IsActive {
				t.Errorf("UpdateCategory() category isActive = %v, want %v", category.IsActive, *tt.req.IsActive)
			}
		})
	}
}

func TestSuggestCategory(t *testing.T) {
	manager, _ := createTestManager(t)

	tests := []struct {
		name          string
		description   string
		wantErr       bool
		expectedCount int
		expectedPaths []string
		minConfidence float64
	}{
		{
			name:          "Successfully_suggest_categories_for_grocery_shopping",
			description:   "Grocery shopping at ICA",
			wantErr:       false,
			expectedCount: 2,
			expectedPaths: []string{"expenses.food", "expenses.groceries"},
			minConfidence: 0.8,
		},
		{
			name:          "Successfully_handle_empty_description",
			description:   "",
			wantErr:       false,
			expectedCount: 2, // Our mock always returns 2 suggestions
			expectedPaths: []string{"expenses.food", "expenses.groceries"},
			minConfidence: 0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions, err := manager.SuggestCategory(context.Background(), tt.description)

			// Validate error cases
			if (err != nil) != tt.wantErr {
				t.Errorf("SuggestCategory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Validate success cases
			if len(suggestions) != tt.expectedCount {
				t.Errorf("SuggestCategory() got %d suggestions, want %d", len(suggestions), tt.expectedCount)
				return
			}

			// Validate suggestions
			for i, suggestion := range suggestions {
				if suggestion.CategoryPath != tt.expectedPaths[i] {
					t.Errorf("SuggestCategory() category[%d] = %v, want %v", i, suggestion.CategoryPath, tt.expectedPaths[i])
				}
				if suggestion.Confidence < tt.minConfidence {
					t.Errorf("SuggestCategory() confidence[%d] = %v, want >= %v", i, suggestion.Confidence, tt.minConfidence)
				}
			}
		})
	}
}
