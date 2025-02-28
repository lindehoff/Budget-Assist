package category

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"log/slog"

	"github.com/lindehoff/Budget-Assist/internal/ai"
	"github.com/lindehoff/Budget-Assist/internal/db"
)

type mockAIService struct{}

func (m *mockAIService) AnalyzeTransaction(ctx context.Context, _ *db.Transaction, _ ai.AnalysisOptions) (*ai.Analysis, error) {
	return &ai.Analysis{
		Remarks: "Test analysis",
		Score:   0.95,
	}, nil
}

func (m *mockAIService) ExtractDocument(ctx context.Context, _ *ai.Document) (*ai.Extraction, error) {
	return &ai.Extraction{
		Content: "Test extraction",
	}, nil
}

func (m *mockAIService) SuggestCategories(ctx context.Context, _ string) ([]ai.CategoryMatch, error) {
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

type mockAIServiceWithError struct {
	mockAIService
}

func (m *mockAIServiceWithError) AnalyzeTransaction(ctx context.Context, _ *db.Transaction, _ ai.AnalysisOptions) (*ai.Analysis, error) {
	return nil, fmt.Errorf("AI service error")
}

func (m *mockAIServiceWithError) SuggestCategories(ctx context.Context, description string) ([]ai.CategoryMatch, error) {
	return nil, fmt.Errorf("AI service error")
}

// createTestManager creates a new manager with mock dependencies for testing
func createTestManager() (*Manager, db.Store) {
	store := ai.NewMockStore()
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

func TestManager_CreateCategory(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateCategoryRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: CreateCategoryRequest{
				TypeID:             1,
				InstanceIdentifier: "test-category",
				Name:               "Test Category",
				Description:        "Test Description",
				Type:               "expense",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			req: CreateCategoryRequest{
				TypeID:             1,
				InstanceIdentifier: "test-category",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := ai.NewMockStore()
			aiService := &mockAIService{}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			m := NewManager(store, aiService, logger)

			got, err := m.CreateCategory(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCategory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("CreateCategory() got = nil, want non-nil")
			}
			if !tt.wantErr {
				if got.Name != tt.req.Name {
					t.Errorf("CreateCategory() got name = %v, want %v", got.Name, tt.req.Name)
				}
			}
		})
	}
}

func TestManager_UpdateCategory(t *testing.T) {
	store := ai.NewMockStore()
	aiService := &mockAIService{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	m := NewManager(store, aiService, logger)

	// Create a test category first
	createReq := CreateCategoryRequest{
		Name:        "Original",
		Description: "Original Description",
		TypeID:      1,
		Type:        "expense",
	}
	existingCategory, err := m.CreateCategory(context.Background(), createReq)
	if err != nil {
		t.Fatalf("failed to create test category: %v", err)
	}

	tests := []struct {
		name    string
		id      uint
		req     UpdateCategoryRequest
		wantErr bool
	}{
		{
			name: "valid request",
			id:   existingCategory.ID,
			req: UpdateCategoryRequest{
				Name:               "Updated Category",
				Description:        "Updated Description",
				InstanceIdentifier: "updated-category",
			},
			wantErr: false,
		},
		{
			name:    "category not found",
			id:      999,
			req:     UpdateCategoryRequest{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := m.UpdateCategory(context.Background(), tt.id, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateCategory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("UpdateCategory() got = nil, want non-nil")
			}
			if !tt.wantErr {
				if got.Name != tt.req.Name {
					t.Errorf("UpdateCategory() got name = %v, want %v", got.Name, tt.req.Name)
				}
			}
		})
	}
}

func TestCreateCategory(t *testing.T) {
	// TODO: Replace context.Background() with proper context handling to test timeouts
	// and cancellation in a future improvement. This should include:
	// - Testing with context timeout
	// - Testing with context cancellation
	// - Testing with parent context values
	ctx := context.TODO()

	manager, _ := createTestManager()

	tests := []struct {
		name        string
		req         CreateCategoryRequest
		wantErr     bool
		expectedErr error
	}{
		{
			name: "valid request",
			req: CreateCategoryRequest{
				Name:        "Food",
				Description: "Food expenses",
				TypeID:      1,
				Type:        "expense",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category, err := manager.CreateCategory(ctx, tt.req)

			// Validate error cases with detailed messages
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCategory(%+v) error = %v, wantErr = %v",
					tt.req, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				var categoryErr CategoryError
				if !errors.As(err, &categoryErr) {
					t.Errorf("CreateCategory(%+v) error type = %T, want CategoryError\nGot error: %v",
						tt.req, err, err)
					return
				}
				if tt.expectedErr != nil && categoryErr.Err.Error() != tt.expectedErr.Error() {
					t.Errorf("CreateCategory(%+v) error message:\ngot:  %v\nwant: %v",
						tt.req, categoryErr.Err, tt.expectedErr)
				}
				return
			}

			// Validate success cases with detailed field comparison
			if category == nil {
				t.Fatalf("CreateCategory(%+v) returned nil category when no error expected", tt.req)
			}

			if category.Name != tt.req.Name {
				t.Errorf("CreateCategory(%+v) category name:\ngot:  %v\nwant: %v",
					tt.req, category.Name, tt.req.Name)
			}
			if category.Description != tt.req.Description {
				t.Errorf("CreateCategory(%+v) category description:\ngot:  %v\nwant: %v",
					tt.req, category.Description, tt.req.Description)
			}
			if category.TypeID != tt.req.TypeID {
				t.Errorf("CreateCategory(%+v) category typeID:\ngot:  %v\nwant: %v",
					tt.req, category.TypeID, tt.req.TypeID)
			}
			if !category.IsActive {
				t.Errorf("CreateCategory(%+v) category isActive = false, want true", tt.req)
			}
		})
	}
}

func TestUpdateCategory(t *testing.T) {
	// TODO: Replace context.Background() with proper context handling to test timeouts
	// and cancellation in a future improvement. This should include:
	// - Testing with context timeout
	// - Testing with context cancellation
	// - Testing with parent context values
	ctx := context.TODO()

	manager, store := createTestManager()
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
			category, err := manager.UpdateCategory(ctx, tt.id, tt.req)

			// Validate error cases with detailed messages
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateCategory(id=%d, %+v) error = %v, wantErr = %v",
					tt.id, tt.req, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				var categoryErr CategoryError
				if !errors.As(err, &categoryErr) {
					t.Errorf("UpdateCategory(id=%d, %+v) error type = %T, want CategoryError\nGot error: %v",
						tt.id, tt.req, err, err)
					return
				}
				if tt.expectedErr != nil && !errors.Is(categoryErr.Err, tt.expectedErr) {
					t.Errorf("UpdateCategory(id=%d, %+v) error:\ngot:  %v\nwant: %v",
						tt.id, tt.req, categoryErr.Err, tt.expectedErr)
				}
				return
			}

			// Validate success cases with detailed field comparison
			if category == nil {
				t.Fatalf("UpdateCategory(id=%d, %+v) returned nil category when no error expected",
					tt.id, tt.req)
				return
			}

			if tt.req.Name != "" && category.Name != tt.req.Name {
				t.Errorf("UpdateCategory(id=%d, %+v) category name:\ngot:  %v\nwant: %v",
					tt.id, tt.req, category.Name, tt.req.Name)
			}
			if tt.req.Description != "" && category.Description != tt.req.Description {
				t.Errorf("UpdateCategory(id=%d, %+v) category description:\ngot:  %v\nwant: %v",
					tt.id, tt.req, category.Description, tt.req.Description)
			}
			if tt.req.IsActive != nil && category.IsActive != *tt.req.IsActive {
				t.Errorf("UpdateCategory(id=%d, %+v) category isActive:\ngot:  %v\nwant: %v",
					tt.id, tt.req, category.IsActive, *tt.req.IsActive)
			}
		})
	}
}

func TestSuggestCategory(t *testing.T) {
	// TODO: Replace context.Background() with proper context handling to test timeouts
	// and cancellation in a future improvement. This should include:
	// - Testing with context timeout
	// - Testing with context cancellation
	// - Testing with parent context values
	ctx := context.TODO()

	manager, _ := createTestManager()

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
			suggestions, err := manager.SuggestCategory(ctx, tt.description)

			// Validate error cases with descriptive messages
			if (err != nil) != tt.wantErr {
				t.Errorf("SuggestCategory(%q) error = %v, wantErr = %v",
					tt.description, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Validate success cases with detailed output
			if got := len(suggestions); got != tt.expectedCount {
				t.Errorf("SuggestCategory(%q) returned %d suggestions, want %d\nGot suggestions: %+v",
					tt.description, got, tt.expectedCount, suggestions)
				return
			}

			// Validate suggestions with detailed comparison
			for i, suggestion := range suggestions {
				if suggestion.CategoryPath != tt.expectedPaths[i] {
					t.Errorf("SuggestCategory(%q) category[%d]:\ngot:  %v\nwant: %v",
						tt.description, i, suggestion.CategoryPath, tt.expectedPaths[i])
					return
				}
				if suggestion.Confidence < tt.minConfidence {
					t.Errorf("SuggestCategory(%q) confidence[%d] = %.2f, want >= %.2f",
						tt.description, i, suggestion.Confidence, tt.minConfidence)
					return
				}
			}
		})
	}
}

func TestCategoryError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      CategoryError
		expected string
	}{
		{
			name: "Format_error_message_correctly",
			err: CategoryError{
				Operation: "create",
				Category:  "Food",
				Err:       fmt.Errorf("validation failed"),
			},
			expected: `category operation "create" failed for "Food": validation failed`,
		},
		{
			name: "Handle_empty_fields",
			err: CategoryError{
				Operation: "",
				Category:  "",
				Err:       fmt.Errorf("unknown error"),
			},
			expected: `category operation "" failed for "": unknown error`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("CategoryError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetCategoryByID(t *testing.T) {
	manager, store := createTestManager()
	existingCategory := createTestCategory(t, store, "Test", 1)

	tests := []struct {
		name        string
		id          uint
		wantErr     bool
		expectedErr error
	}{
		{
			name:    "Successfully_get_existing_category",
			id:      existingCategory.ID,
			wantErr: false,
		},
		{
			name:        "GetCategoryByID_error_non_existent_category",
			id:          999,
			wantErr:     true,
			expectedErr: db.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()

			category, err := manager.GetCategoryByID(ctx, tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetCategoryByID() error = %v, wantErr %v (input: %d)", err, tt.wantErr, tt.id)
				return
			}
			if tt.wantErr {
				var categoryErr CategoryError
				if !errors.As(err, &categoryErr) {
					t.Errorf("GetCategoryByID() error type = %T, want CategoryError (input: %d)", err, tt.id)
					return
				}
				if tt.expectedErr != nil && !errors.Is(categoryErr.Err, tt.expectedErr) {
					t.Errorf("GetCategoryByID() error = %v, expectedErr %v (input: %d)",
						categoryErr.Err, tt.expectedErr, tt.id)
				}
				return
			}

			if category == nil {
				t.Fatal("GetCategoryByID() returned nil category when no error expected")
			}

			if category.ID != tt.id {
				t.Errorf("GetCategoryByID() category ID = %v, want %v (input: %d)",
					category.ID, tt.id, tt.id)
				return
			}
		})
	}
}

func TestListCategories(t *testing.T) {
	manager, store := createTestManager()

	// Create test categories
	_ = createTestCategory(t, store, "Food", 1)
	_ = createTestCategory(t, store, "Transport", 1)
	_ = createTestCategory(t, store, "Housing", 2)

	tests := []struct {
		name           string
		typeID         *uint
		expectedCount  int
		expectedTypeID uint
		wantErr        bool
	}{
		{
			name:           "Successfully_list_all_categories",
			typeID:         nil,
			expectedCount:  3,
			expectedTypeID: 0,
			wantErr:        false,
		},
		{
			name:           "Successfully_filter_by_type",
			typeID:         func() *uint { id := uint(1); return &id }(),
			expectedCount:  2,
			expectedTypeID: 1,
			wantErr:        false,
		},
		{
			name:           "Successfully_handle_empty_type",
			typeID:         func() *uint { id := uint(3); return &id }(),
			expectedCount:  0,
			expectedTypeID: 3,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			categories, err := manager.ListCategories(context.Background(), tt.typeID)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListCategories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(categories) != tt.expectedCount {
					t.Errorf("ListCategories() got %d categories, want %d", len(categories), tt.expectedCount)
				}

				if tt.typeID != nil {
					for _, cat := range categories {
						if cat.TypeID != tt.expectedTypeID {
							t.Errorf("ListCategories() category typeID = %v, want %v", cat.TypeID, tt.expectedTypeID)
						}
					}
				}
			}
		})
	}
}

// mockStoreWithErrors implements db.Store interface for testing error cases
type mockStoreWithErrors struct {
	db.Store
	shouldError bool
}

func (m *mockStoreWithErrors) CreateCategory(ctx context.Context, _ *db.Category) error {
	if m.shouldError {
		return fmt.Errorf("database error")
	}
	return nil
}

func (m *mockStoreWithErrors) CreateTranslation(ctx context.Context, _ *db.Translation) error {
	if m.shouldError {
		return fmt.Errorf("database error")
	}
	return nil
}

func TestCreateTranslations(t *testing.T) {
	tests := []struct {
		name         string
		translations map[string]TranslationData
		shouldError  bool
		wantErr      bool
	}{
		{
			name: "Successfully_create_translations",
			translations: map[string]TranslationData{
				"sv": {Name: "Mat", Description: "Matkostnader"},
				"en": {Name: "Food", Description: "Food expenses"},
			},
			shouldError: false,
			wantErr:     false,
		},
		{
			name: "Error_creating_translation",
			translations: map[string]TranslationData{
				"sv": {Name: "Mat", Description: "Matkostnader"},
			},
			shouldError: true,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockStoreWithErrors{Store: ai.NewMockStore(), shouldError: tt.shouldError}
			manager := NewManager(store, &mockAIService{}, slog.Default())

			err := manager.createTranslations(context.Background(), 1, tt.translations)

			if (err != nil) != tt.wantErr {
				t.Errorf("createTranslations() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSuggestCategory_Error(t *testing.T) {
	manager := NewManager(ai.NewMockStore(), &mockAIServiceWithError{}, slog.Default())

	_, err := manager.SuggestCategory(context.Background(), "test description")

	if err == nil {
		t.Error("SuggestCategory() expected error, got nil")
		return
	}

	var categoryErr CategoryError
	if !errors.As(err, &categoryErr) {
		t.Errorf("SuggestCategory() error type = %T, want CategoryError", err)
		return
	}

	if categoryErr.Operation != "suggest" {
		t.Errorf("SuggestCategory() error operation = %v, want 'suggest'", categoryErr.Operation)
	}
}

// TestConcurrent_category_operations tests the manager's behavior under concurrent load
func TestConcurrent_category_operations(t *testing.T) {
	manager, _ := createTestManager()

	numOperations := 10
	var wg sync.WaitGroup
	wg.Add(numOperations * 2) // For both create and update operations

	categories := make([]*db.Category, numOperations)
	var categoriesMu sync.RWMutex

	errCh := make(chan error, numOperations*2)

	// Create categories concurrently
	for i := 0; i < numOperations; i++ {
		go func(i int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			req := CreateCategoryRequest{
				Name:        fmt.Sprintf("Category %d", i),
				Description: fmt.Sprintf("Test category %d", i),
				TypeID:      1,
				Type:        "expense",
			}

			category, err := manager.CreateCategory(ctx, req)
			if err != nil {
				errCh <- fmt.Errorf("create operation %d failed: %w", i, err)
				return
			}

			categoriesMu.Lock()
			categories[i] = category
			categoriesMu.Unlock()
		}(i)
	}

	// Update categories concurrently
	for i := 0; i < numOperations; i++ {
		go func(i int) {
			defer wg.Done()

			// Wait a bit to ensure category is created
			time.Sleep(10 * time.Millisecond)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			// Try updating until category is available
			var lastErr error
			for retries := 0; retries < 3; retries++ {
				categoriesMu.RLock()
				category := categories[i]
				categoriesMu.RUnlock()

				if category == nil {
					time.Sleep(10 * time.Millisecond)
					continue
				}

				req := UpdateCategoryRequest{
					Name:        fmt.Sprintf("Updated Category %d", i),
					Description: fmt.Sprintf("Updated description %d", i),
				}

				_, err := manager.UpdateCategory(ctx, category.ID, req)
				if err == nil {
					return
				}
				lastErr = err
				time.Sleep(10 * time.Millisecond)
			}
			errCh <- fmt.Errorf("update operation %d failed after retries: %w", i, lastErr)
		}(i)
	}

	wg.Wait()
	close(errCh)

	// Check for any errors
	for err := range errCh {
		t.Error(err)
	}

	// Verify final state
	for i := 0; i < numOperations; i++ {
		categoriesMu.RLock()
		category := categories[i]
		categoriesMu.RUnlock()

		if category == nil {
			t.Errorf("Category %d was not created", i)
			continue
		}

		expectedName := fmt.Sprintf("Updated Category %d", i)
		if category.Name != expectedName {
			t.Errorf("Category %d name = %q, want %q", i, category.Name, expectedName)
		}
	}
}
