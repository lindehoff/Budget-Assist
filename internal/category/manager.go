package category

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/lindehoff/Budget-Assist/internal/ai"
	"github.com/lindehoff/Budget-Assist/internal/db"
)

// Manager handles category operations
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

// CreateCategoryRequest represents the data needed to create a new category
type CreateCategoryRequest struct {
	Name               string
	Description        string
	TypeID             uint
	InstanceIdentifier string
	Translations       map[string]TranslationData
	Subcategories      []uint // List of subcategory IDs to link
}

// CreateSubcategoryRequest represents the data needed to create a new subcategory
type CreateSubcategoryRequest struct {
	Name               string
	Description        string
	InstanceIdentifier string
	Translations       map[string]TranslationData
	Categories         []uint // List of category IDs to link
	IsSystem           bool   // Whether this is a system subcategory
}

// TranslationData represents translation information
type TranslationData struct {
	Name        string
	Description string
}

// UpdateCategoryRequest represents the data needed to update a category
type UpdateCategoryRequest struct {
	Name                string
	Description         string
	IsActive            *bool
	InstanceIdentifier  string
	Translations        map[string]TranslationData
	AddSubcategories    []uint // Subcategories to add
	RemoveSubcategories []uint // Subcategories to remove
}

// UpdateSubcategoryRequest represents the data needed to update a subcategory
type UpdateSubcategoryRequest struct {
	Name               string
	Description        string
	IsActive           *bool
	InstanceIdentifier string
	Translations       map[string]TranslationData
	AddCategories      []uint // Categories to add
	RemoveCategories   []uint // Categories to remove
}

// CreateCategory creates a new category
func (m *Manager) CreateCategory(ctx context.Context, req CreateCategoryRequest) (*db.Category, error) {
	if err := req.Validate(); err != nil {
		return nil, CategoryError{
			Operation: "create",
			Category:  req.Name,
			Err:       err,
		}
	}

	// Create the category
	category := &db.Category{
		TypeID:             req.TypeID,
		InstanceIdentifier: req.InstanceIdentifier,
		IsActive:           true,
	}

	// Create translations
	for langCode, data := range req.Translations {
		translation := &db.Translation{
			EntityType:   string(db.EntityTypeCategory),
			LanguageCode: langCode,
			Name:         data.Name,
			Description:  data.Description,
		}
		category.Translations = append(category.Translations, *translation)
	}

	// Add default English translation if not provided
	if _, exists := req.Translations[db.LangEN]; !exists {
		translation := &db.Translation{
			EntityType:   string(db.EntityTypeCategory),
			LanguageCode: db.LangEN,
			Name:         req.Name,
			Description:  req.Description,
		}
		category.Translations = append(category.Translations, *translation)
	}

	// Create category first
	if err := m.store.CreateCategory(ctx, category); err != nil {
		return nil, CategoryError{
			Operation: "create",
			Category:  req.Name,
			Err:       err,
		}
	}

	// Link subcategories if provided
	if len(req.Subcategories) > 0 {
		for _, subcatID := range req.Subcategories {
			link := &db.CategorySubcategory{
				CategoryID:    category.ID,
				SubcategoryID: subcatID,
				IsActive:      true,
			}
			if err := m.store.CreateCategorySubcategory(ctx, link); err != nil {
				m.logger.Error("failed to link subcategory",
					"category_id", category.ID,
					"subcategory_id", subcatID,
					"error", err)
			}
		}
	}

	return category, nil
}

// CreateSubcategory creates a new subcategory
func (m *Manager) CreateSubcategory(ctx context.Context, req CreateSubcategoryRequest) (*db.Subcategory, error) {
	if err := req.Validate(); err != nil {
		return nil, CategoryError{
			Operation: "create_subcategory",
			Category:  req.Name,
			Err:       err,
		}
	}

	// Create the subcategory
	subcategory := &db.Subcategory{
		InstanceIdentifier: req.InstanceIdentifier,
		IsActive:           true,
	}

	// Create translations
	for langCode, data := range req.Translations {
		translation := &db.Translation{
			EntityType:   "subcategory",
			LanguageCode: langCode,
			Name:         data.Name,
			Description:  data.Description,
		}
		subcategory.Translations = append(subcategory.Translations, *translation)
	}

	// Create subcategory first
	if err := m.store.CreateSubcategory(ctx, subcategory); err != nil {
		return nil, CategoryError{
			Operation: "create_subcategory",
			Category:  req.Name,
			Err:       err,
		}
	}

	// Link categories if provided
	if len(req.Categories) > 0 {
		for _, catID := range req.Categories {
			link := &db.CategorySubcategory{
				CategoryID:    catID,
				SubcategoryID: subcategory.ID,
				IsActive:      true,
			}
			if err := m.store.CreateCategorySubcategory(ctx, link); err != nil {
				m.logger.Error("failed to link category",
					"category_id", catID,
					"subcategory_id", subcategory.ID,
					"error", err)
			}
		}
	}

	return subcategory, nil
}

// UpdateCategory updates an existing category
func (m *Manager) UpdateCategory(ctx context.Context, id uint, req UpdateCategoryRequest) (*db.Category, error) {
	category, err := m.store.GetCategoryByID(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, CategoryError{
				Operation: "update",
				Category:  fmt.Sprintf("id=%d", id),
				Err:       db.ErrNotFound,
			}
		}
		return nil, CategoryError{
			Operation: "update",
			Category:  fmt.Sprintf("id=%d", id),
			Err:       fmt.Errorf("failed to get category: %w", err),
		}
	}

	// Update basic fields
	if req.Name != "" {
		category.Name = req.Name
	}
	if req.Description != "" {
		category.Description = req.Description
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}
	if req.InstanceIdentifier != "" {
		category.InstanceIdentifier = req.InstanceIdentifier
	}

	// Update translations if provided
	if len(req.Translations) > 0 {
		for langCode, data := range req.Translations {
			translation := &db.Translation{
				EntityID:     category.ID,
				EntityType:   string(db.EntityTypeCategory),
				LanguageCode: langCode,
				Name:         data.Name,
				Description:  data.Description,
			}
			if err := m.store.CreateTranslation(ctx, translation); err != nil {
				m.logger.Error("failed to update translation",
					"category_id", category.ID,
					"language", langCode,
					"error", err)
			}
		}
	}

	// Update default English translation if name/description provided
	if req.Name != "" || req.Description != "" {
		translation := &db.Translation{
			EntityID:     category.ID,
			EntityType:   string(db.EntityTypeCategory),
			LanguageCode: db.LangEN,
			Name:         req.Name,
			Description:  req.Description,
		}
		if err := m.store.CreateTranslation(ctx, translation); err != nil {
			m.logger.Error("failed to update default translation",
				"category_id", category.ID,
				"error", err)
		}
	}

	// Update category-subcategory relationships
	if len(req.AddSubcategories) > 0 {
		for _, subcatID := range req.AddSubcategories {
			link := &db.CategorySubcategory{
				CategoryID:    category.ID,
				SubcategoryID: subcatID,
				IsActive:      true,
			}
			if err := m.store.CreateCategorySubcategory(ctx, link); err != nil {
				m.logger.Error("failed to add subcategory link",
					"category_id", category.ID,
					"subcategory_id", subcatID,
					"error", err)
			}
		}
	}

	if len(req.RemoveSubcategories) > 0 {
		for _, subcatID := range req.RemoveSubcategories {
			if err := m.store.DeleteCategorySubcategory(ctx, category.ID, subcatID); err != nil {
				m.logger.Error("failed to remove subcategory link",
					"category_id", category.ID,
					"subcategory_id", subcatID,
					"error", err)
			}
		}
	}

	// Save the updated category
	if err := m.store.UpdateCategory(ctx, category); err != nil {
		return nil, CategoryError{
			Operation: "update",
			Category:  fmt.Sprintf("id=%d", id),
			Err:       fmt.Errorf("failed to update category: %w", err),
		}
	}

	// Fetch the updated category to get the latest translations
	return m.GetCategoryByID(ctx, id)
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

// GetSubcategoryByID retrieves a subcategory by its ID
func (m *Manager) GetSubcategoryByID(ctx context.Context, id uint) (*db.Subcategory, error) {
	subcategory, err := m.store.GetSubcategoryByID(ctx, id)
	if err != nil {
		return nil, CategoryError{
			Operation: "get_subcategory",
			Category:  fmt.Sprintf("id=%d", id),
			Err:       err,
		}
	}
	return subcategory, nil
}

// ListSubcategories returns all subcategories
func (m *Manager) ListSubcategories(ctx context.Context) ([]db.Subcategory, error) {
	subcategories, err := m.store.ListSubcategories(ctx)
	if err != nil {
		return nil, CategoryError{
			Operation: "list_subcategories",
			Category:  "all",
			Err:       err,
		}
	}
	return subcategories, nil
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

func (r *CreateSubcategoryRequest) Validate() error {
	if len(r.Translations) == 0 {
		return fmt.Errorf("at least one translation is required")
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

// CreateTranslation creates a new translation for an entity
func (m *Manager) CreateTranslation(ctx context.Context, translation *db.Translation) error {
	if err := m.store.CreateTranslation(ctx, translation); err != nil {
		return CategoryError{
			Operation: "create_translation",
			Category:  fmt.Sprintf("%s_%d", translation.EntityType, translation.EntityID),
			Err:       err,
		}
	}
	return nil
}

// CreateCategoryType creates a new category type
func (m *Manager) CreateCategoryType(ctx context.Context, categoryType *db.CategoryType) error {
	if err := m.store.CreateCategoryType(ctx, categoryType); err != nil {
		return CategoryError{
			Operation: "create_category_type",
			Category:  categoryType.Name,
			Err:       err,
		}
	}
	return nil
}
