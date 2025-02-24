package db

import (
	"context"
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

// Store defines the interface for database operations
type Store interface {
	CreateCategory(ctx context.Context, category *Category) error
	UpdateCategory(ctx context.Context, category *Category) error
	GetCategoryByID(ctx context.Context, id uint) (*Category, error)
	ListCategories(ctx context.Context, typeID *uint) ([]Category, error)
	GetCategoryTypeByID(ctx context.Context, id uint) (*CategoryType, error)
	CreateTranslation(ctx context.Context, translation *Translation) error
	GetTranslations(ctx context.Context, entityID uint, entityType string) ([]Translation, error)
	DeleteCategory(ctx context.Context, id uint) error

	// Prompt-related methods
	GetPromptByType(ctx context.Context, promptType string) (*Prompt, error)
	UpdatePrompt(ctx context.Context, prompt *Prompt) error
	ListPrompts(ctx context.Context) ([]Prompt, error)
}

// SQLStore implements Store interface using GORM
type SQLStore struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewStore creates a new database store
func NewStore(db *gorm.DB, logger *slog.Logger) Store {
	return &SQLStore{
		db:     db,
		logger: logger,
	}
}

// CreateCategory creates a new category in the database
func (s *SQLStore) CreateCategory(ctx context.Context, category *Category) error {
	// Validate required fields
	if category.Name == "" {
		return fmt.Errorf("category name is required")
	}
	if category.TypeID == 0 {
		return fmt.Errorf("category type ID is required")
	}

	result := s.db.WithContext(ctx).Create(category)
	if result.Error != nil {
		return fmt.Errorf("failed to create category: %w", result.Error)
	}
	return nil
}

// UpdateCategory updates an existing category
func (s *SQLStore) UpdateCategory(ctx context.Context, category *Category) error {
	result := s.db.WithContext(ctx).Save(category)
	if result.Error != nil {
		return fmt.Errorf("failed to update category: %w", result.Error)
	}
	return nil
}

// GetCategoryByID retrieves a category by its ID
func (s *SQLStore) GetCategoryByID(ctx context.Context, id uint) (*Category, error) {
	var category Category
	result := s.db.WithContext(ctx).First(&category, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get category: %w", result.Error)
	}
	return &category, nil
}

// ListCategories returns all categories, optionally filtered by type
func (s *SQLStore) ListCategories(ctx context.Context, typeID *uint) ([]Category, error) {
	var categories []Category
	query := s.db.WithContext(ctx)
	if typeID != nil {
		query = query.Where("type_id = ?", *typeID)
	}
	result := query.Find(&categories)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list categories: %w", result.Error)
	}
	return categories, nil
}

// GetCategoryTypeByID retrieves a category type by its ID
func (s *SQLStore) GetCategoryTypeByID(ctx context.Context, id uint) (*CategoryType, error) {
	var categoryType CategoryType
	result := s.db.WithContext(ctx).First(&categoryType, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get category type: %w", result.Error)
	}
	return &categoryType, nil
}

// CreateTranslation creates a new translation in the database
func (s *SQLStore) CreateTranslation(ctx context.Context, translation *Translation) error {
	// Validate required fields
	if translation.Name == "" {
		return fmt.Errorf("translation name is required")
	}
	if translation.EntityID == 0 {
		return fmt.Errorf("entity ID is required")
	}
	if translation.EntityType == "" {
		return fmt.Errorf("entity type is required")
	}
	if translation.LanguageCode == "" {
		return fmt.Errorf("language code is required")
	}

	result := s.db.WithContext(ctx).Create(translation)
	if result.Error != nil {
		return fmt.Errorf("failed to create translation: %w", result.Error)
	}
	return nil
}

// DeleteCategory deletes a category from the database
func (s *SQLStore) DeleteCategory(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Delete(&Category{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete category: %w", result.Error)
	}
	return nil
}

// GetTranslations retrieves all translations for a given entity
func (s *SQLStore) GetTranslations(ctx context.Context, entityID uint, entityType string) ([]Translation, error) {
	var translations []Translation
	result := s.db.WithContext(ctx).
		Where("entity_id = ? AND entity_type = ?", entityID, entityType).
		Find(&translations)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get translations: %w", result.Error)
	}
	return translations, nil
}

// GetPromptByType retrieves a prompt template by its type
func (s *SQLStore) GetPromptByType(ctx context.Context, promptType string) (*Prompt, error) {
	var prompt Prompt
	result := s.db.WithContext(ctx).
		Where("type = ? AND is_active = ?", promptType, true).
		First(&prompt)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("prompt template not found for type: %s", promptType)
		}
		return nil, fmt.Errorf("failed to get prompt: %w", result.Error)
	}
	return &prompt, nil
}

// UpdatePrompt updates or creates a prompt template
func (s *SQLStore) UpdatePrompt(ctx context.Context, prompt *Prompt) error {
	// Check if prompt exists
	var existing Prompt
	result := s.db.WithContext(ctx).
		Where("type = ?", prompt.Type).
		First(&existing)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create new prompt
			result = s.db.WithContext(ctx).Create(prompt)
			if result.Error != nil {
				return fmt.Errorf("failed to create prompt: %w", result.Error)
			}
			return nil
		}
		return fmt.Errorf("failed to check existing prompt: %w", result.Error)
	}

	// Update existing prompt
	prompt.ID = existing.ID
	result = s.db.WithContext(ctx).Save(prompt)
	if result.Error != nil {
		return fmt.Errorf("failed to update prompt: %w", result.Error)
	}
	return nil
}

// ListPrompts returns all prompt templates
func (s *SQLStore) ListPrompts(ctx context.Context) ([]Prompt, error) {
	var prompts []Prompt
	result := s.db.WithContext(ctx).Find(&prompts)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list prompts: %w", result.Error)
	}
	return prompts, nil
}
