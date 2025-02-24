package category

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lindehoff/Budget-Assist/internal/ai"
	"github.com/lindehoff/Budget-Assist/internal/db"
)

// Manager handles category operations and synchronization between DB and AI models
type Manager struct {
	store     db.Store
	aiService ai.Service
	logger    *slog.Logger
}

// CategoryError represents category-specific errors
type CategoryError struct {
	Operation string
	Category  string
	Err       error
}

func (e CategoryError) Error() string {
	return fmt.Sprintf("category operation %q failed for %q: %v", e.Operation, e.Category, e.Err)
}

// NewManager creates a new category manager
func NewManager(store db.Store, aiService ai.Service, logger *slog.Logger) *Manager {
	return &Manager{
		store:     store,
		aiService: aiService,
		logger:    logger,
	}
}

// CreateCategory creates a new category with the given details
func (m *Manager) CreateCategory(ctx context.Context, req CreateCategoryRequest) (*db.Category, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, CategoryError{
			Operation: "create",
			Category:  req.Name,
			Err:       err,
		}
	}

	// Create category in DB
	category := &db.Category{
		Name:               req.Name,
		Description:        req.Description,
		TypeID:             req.TypeID,
		IsActive:           true,
		InstanceIdentifier: req.InstanceIdentifier,
	}

	if err := m.store.CreateCategory(ctx, category); err != nil {
		return nil, CategoryError{
			Operation: "create",
			Category:  req.Name,
			Err:       err,
		}
	}

	// Create translations if provided
	if len(req.Translations) > 0 {
		if err := m.createTranslations(ctx, category.ID, req.Translations); err != nil {
			m.logger.Error("failed to create translations",
				"category", category.Name,
				"error", err)
		}
	}

	return category, nil
}

// UpdateCategory updates an existing category
func (m *Manager) UpdateCategory(ctx context.Context, id uint, req UpdateCategoryRequest) (*db.Category, error) {
	category, err := m.store.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, CategoryError{
			Operation: "update",
			Category:  fmt.Sprintf("id=%d", id),
			Err:       err,
		}
	}

	// Update fields if provided
	if req.Name != "" {
		category.Name = req.Name
	}
	if req.Description != "" {
		category.Description = req.Description
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if err := m.store.UpdateCategory(ctx, category); err != nil {
		return nil, CategoryError{
			Operation: "update",
			Category:  category.Name,
			Err:       err,
		}
	}

	return category, nil
}

// GetCategoryByID retrieves a category by its ID
func (m *Manager) GetCategoryByID(ctx context.Context, id uint) (*db.Category, error) {
	category, err := m.store.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, CategoryError{
			Operation: "get",
			Category:  fmt.Sprintf("id=%d", id),
			Err:       err,
		}
	}
	return category, nil
}

// ListCategories returns all categories, optionally filtered by type
func (m *Manager) ListCategories(ctx context.Context, typeID *uint) ([]db.Category, error) {
	categories, err := m.store.ListCategories(ctx, typeID)
	if err != nil {
		return nil, CategoryError{
			Operation: "list",
			Category:  "all",
			Err:       err,
		}
	}
	return categories, nil
}

// SuggestCategory suggests a category for a transaction description
func (m *Manager) SuggestCategory(ctx context.Context, description string) ([]CategorySuggestion, error) {
	matches, err := m.aiService.SuggestCategories(ctx, description)
	if err != nil {
		return nil, CategoryError{
			Operation: "suggest",
			Category:  description,
			Err:       err,
		}
	}

	suggestions := make([]CategorySuggestion, 0, len(matches))
	for _, match := range matches {
		suggestions = append(suggestions, CategorySuggestion{
			CategoryPath: match.Category,
			Confidence:   match.Confidence,
		})
	}

	return suggestions, nil
}

// CreateCategoryRequest represents the data needed to create a new category
type CreateCategoryRequest struct {
	Name               string
	Description        string
	TypeID             uint
	InstanceIdentifier string
	Translations       map[string]TranslationData
}

// TranslationData represents translation information
type TranslationData struct {
	Name        string
	Description string
}

// UpdateCategoryRequest represents the data needed to update a category
type UpdateCategoryRequest struct {
	Name        string
	Description string
	IsActive    *bool
}

// CategorySuggestion represents an AI-suggested category with confidence
type CategorySuggestion struct {
	CategoryPath string
	Confidence   float64
}

func (r *CreateCategoryRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if r.TypeID == 0 {
		return fmt.Errorf("type ID is required")
	}
	return nil
}

func (m *Manager) createTranslations(ctx context.Context, categoryID uint, translations map[string]TranslationData) error {
	for langCode, data := range translations {
		translation := &db.Translation{
			EntityID:     categoryID,
			EntityType:   "category",
			LanguageCode: langCode,
			Name:         data.Name,
			Description:  data.Description,
		}
		if err := m.store.CreateTranslation(ctx, translation); err != nil {
			return fmt.Errorf("failed to create translation for language %s: %w", langCode, err)
		}
	}
	return nil
}
