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
	Type               string
	TypeID             uint
	InstanceIdentifier string
	Subcategories      []string // List of subcategory names to link
}

// CreateSubcategoryRequest represents the data needed to create a new subcategory
type CreateSubcategoryRequest struct {
	Name               string
	Description        string
	InstanceIdentifier string
	Categories         []string // List of category names to link
	Tags               []string // List of tags to attach
	IsSystem           bool     // Whether this is a system subcategory
}

// UpdateCategoryRequest represents the data needed to update a category
type UpdateCategoryRequest struct {
	Name                string
	Description         string
	IsActive            *bool
	InstanceIdentifier  string
	AddSubcategories    []string // Subcategories to add by name
	RemoveSubcategories []string // Subcategories to remove by name
}

// UpdateSubcategoryRequest represents the data needed to update a subcategory
type UpdateSubcategoryRequest struct {
	Name               string
	Description        string
	IsActive           *bool
	InstanceIdentifier string
	AddCategories      []string // Categories to add by name
	RemoveCategories   []string // Categories to remove by name
	AddTags            []string // Tags to add
	RemoveTags         []string // Tags to remove
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

	category := &db.Category{
		Name:               req.Name,
		Description:        req.Description,
		TypeID:             req.TypeID,
		Type:               req.Type,
		InstanceIdentifier: req.InstanceIdentifier,
		IsActive:           true,
	}

	if err := m.store.CreateCategory(ctx, category); err != nil {
		return nil, CategoryError{
			Operation: "create",
			Category:  req.Name,
			Err:       err,
		}
	}

	// Link subcategories if provided
	for _, subcatName := range req.Subcategories {
		subcat, err := m.store.GetSubcategoryByName(ctx, subcatName)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				// Create new subcategory if it doesn't exist
				subcat = &db.Subcategory{
					Name:        subcatName,
					Description: subcatName,
					IsSystem:    false,
					IsActive:    true,
				}
				if err := m.store.CreateSubcategory(ctx, subcat); err != nil {
					return nil, CategoryError{
						Operation: "create_subcategory",
						Category:  subcatName,
						Err:       err,
					}
				}
			} else {
				return nil, CategoryError{
					Operation: "get_subcategory",
					Category:  subcatName,
					Err:       err,
				}
			}
		}

		// Link the subcategory to the category
		link := &db.CategorySubcategory{
			CategoryID:    category.ID,
			SubcategoryID: subcat.ID,
			IsActive:      true,
		}
		if err := m.store.CreateCategorySubcategory(ctx, link); err != nil {
			return nil, CategoryError{
				Operation: "link_subcategory",
				Category:  req.Name,
				Err:       err,
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
		Name:        req.Name,
		Description: req.Description,
		IsSystem:    req.IsSystem,
		IsActive:    true,
	}

	// Create subcategory first
	if err := m.store.CreateSubcategory(ctx, subcategory); err != nil {
		return nil, CategoryError{
			Operation: "create_subcategory",
			Category:  req.Name,
			Err:       err,
		}
	}

	// Create and link tags
	for _, tagName := range req.Tags {
		tag, err := m.getOrCreateTag(ctx, tagName)
		if err != nil {
			m.logger.Error("failed to create/get tag",
				"tag_name", tagName,
				"error", err)
			continue
		}

		if err := m.store.LinkSubcategoryTag(ctx, subcategory.ID, tag.ID); err != nil {
			m.logger.Error("failed to link tag",
				"subcategory_id", subcategory.ID,
				"tag_id", tag.ID,
				"error", err)
		}
	}

	// Link categories if provided
	if len(req.Categories) > 0 {
		for _, catName := range req.Categories {
			// Find category by name
			cat, err := m.store.GetCategoryByName(ctx, catName)
			if err != nil {
				m.logger.Error("failed to find category",
					"name", catName,
					"error", err)
				continue
			}

			link := &db.CategorySubcategory{
				CategoryID:    cat.ID,
				SubcategoryID: subcategory.ID,
				IsActive:      true,
			}
			if err := m.store.CreateCategorySubcategory(ctx, link); err != nil {
				m.logger.Error("failed to link category",
					"category_name", catName,
					"subcategory_id", subcategory.ID,
					"error", err)
			}
		}
	}

	return subcategory, nil
}

// getOrCreateTag gets an existing tag or creates a new one
func (m *Manager) getOrCreateTag(ctx context.Context, name string) (*db.Tag, error) {
	tag, err := m.store.GetTagByName(ctx, name)
	if err == nil {
		return tag, nil
	}

	// Create new tag if not found
	tag = &db.Tag{
		Name: name,
	}
	if err := m.store.CreateTag(ctx, tag); err != nil {
		return nil, err
	}
	return tag, nil
}

// updateBasicFields updates the basic fields of a category
func (m *Manager) updateBasicFields(category *db.Category, req UpdateCategoryRequest) {
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
}

// addSubcategories adds new subcategories to a category
func (m *Manager) addSubcategories(ctx context.Context, category *db.Category, subcatNames []string) {
	for _, subcatName := range subcatNames {
		subcat, err := m.getOrCreateSubcategory(ctx, subcatName)
		if err != nil {
			m.logger.Error("failed to get/create subcategory",
				"name", subcatName,
				"error", err)
			continue
		}

		link := &db.CategorySubcategory{
			CategoryID:    category.ID,
			SubcategoryID: subcat.ID,
			IsActive:      true,
		}
		if err := m.store.CreateCategorySubcategory(ctx, link); err != nil {
			m.logger.Error("failed to add subcategory link",
				"category_id", category.ID,
				"subcategory_name", subcatName,
				"error", err)
		}
	}
}

// removeSubcategories removes subcategories from a category
func (m *Manager) removeSubcategories(ctx context.Context, category *db.Category, subcatNames []string) {
	for _, subcatName := range subcatNames {
		subcat, err := m.store.GetSubcategoryByName(ctx, subcatName)
		if err != nil {
			m.logger.Error("failed to find subcategory",
				"name", subcatName,
				"error", err)
			continue
		}

		if err := m.store.DeleteCategorySubcategory(ctx, category.ID, subcat.ID); err != nil {
			m.logger.Error("failed to remove subcategory link",
				"category_id", category.ID,
				"subcategory_name", subcatName,
				"error", err)
		}
	}
}

// getOrCreateSubcategory gets an existing subcategory or creates a new one
func (m *Manager) getOrCreateSubcategory(ctx context.Context, name string) (*db.Subcategory, error) {
	subcat, err := m.store.GetSubcategoryByName(ctx, name)
	if err == nil {
		return subcat, nil
	}

	if !errors.Is(err, db.ErrNotFound) {
		return nil, err
	}

	// Create new subcategory if it doesn't exist
	subcat = &db.Subcategory{
		Name:        name,
		Description: name,
		IsSystem:    false,
		IsActive:    true,
	}
	if err := m.store.CreateSubcategory(ctx, subcat); err != nil {
		return nil, err
	}
	return subcat, nil
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
	m.updateBasicFields(category, req)

	// Update category-subcategory relationships
	if len(req.AddSubcategories) > 0 {
		m.addSubcategories(ctx, category, req.AddSubcategories)
	}

	if len(req.RemoveSubcategories) > 0 {
		m.removeSubcategories(ctx, category, req.RemoveSubcategories)
	}

	// Save the updated category
	if err := m.store.UpdateCategory(ctx, category); err != nil {
		return nil, CategoryError{
			Operation: "update",
			Category:  fmt.Sprintf("id=%d", id),
			Err:       fmt.Errorf("failed to update category: %w", err),
		}
	}

	// Fetch the updated category to get the latest state
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

// Validate validates the create category request
func (r *CreateCategoryRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if r.TypeID == 0 {
		return fmt.Errorf("type_id is required")
	}
	return nil
}

// Validate validates the create subcategory request
func (r *CreateSubcategoryRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

// GetStore returns the underlying store
func (m *Manager) GetStore() db.Store {
	return m.store
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
