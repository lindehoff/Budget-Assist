package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

// TransactionFilter defines filters for listing transactions
type TransactionFilter struct {
	CategoryID    *uint
	SubcategoryID *uint
	StartDate     *time.Time
	EndDate       *time.Time
}

// Store defines the interface for database operations
type Store interface {
	// Category type operations
	CreateCategoryType(ctx context.Context, categoryType *CategoryType) error
	UpdateCategoryType(ctx context.Context, categoryType *CategoryType) error
	GetCategoryTypeByID(ctx context.Context, id uint) (*CategoryType, error)
	ListCategoryTypes(ctx context.Context) ([]CategoryType, error)

	// Category operations
	CreateCategory(ctx context.Context, category *Category) error
	UpdateCategory(ctx context.Context, category *Category) error
	GetCategoryByID(ctx context.Context, id uint) (*Category, error)
	ListCategories(ctx context.Context, typeID *uint) ([]Category, error)
	DeleteCategory(ctx context.Context, id uint) error

	// Subcategory operations
	CreateSubcategory(ctx context.Context, subcategory *Subcategory) error
	UpdateSubcategory(ctx context.Context, subcategory *Subcategory) error
	GetSubcategoryByID(ctx context.Context, id uint) (*Subcategory, error)
	ListSubcategories(ctx context.Context) ([]Subcategory, error)
	DeleteSubcategory(ctx context.Context, id uint) error

	// Category-Subcategory relationship operations
	CreateCategorySubcategory(ctx context.Context, link *CategorySubcategory) error
	DeleteCategorySubcategory(ctx context.Context, categoryID, subcategoryID uint) error

	// Translation operations
	CreateTranslation(ctx context.Context, translation *Translation) error
	UpdateTranslation(ctx context.Context, translation *Translation) error
	GetTranslationByID(ctx context.Context, id uint) (*Translation, error)
	ListTranslations(ctx context.Context, entityType string, entityID uint) ([]Translation, error)

	// Transaction operations
	CreateTransaction(ctx context.Context, transaction *Transaction) error
	UpdateTransaction(ctx context.Context, transaction *Transaction) error
	GetTransactionByID(ctx context.Context, id uint) (*Transaction, error)
	ListTransactions(ctx context.Context, filter *TransactionFilter) ([]Transaction, error)
	DeleteTransaction(ctx context.Context, id uint) error

	// Prompt operations
	CreatePrompt(ctx context.Context, prompt *Prompt) error
	UpdatePrompt(ctx context.Context, prompt *Prompt) error
	GetPromptByID(ctx context.Context, id uint) (*Prompt, error)
	GetPromptByType(ctx context.Context, promptType string) (*Prompt, error)
	ListPrompts(ctx context.Context) ([]Prompt, error)
	DeletePrompt(ctx context.Context, id uint) error

	// Close closes the database connection
	Close() error

	// New method for GetTranslations
	GetTranslations(ctx context.Context, entityID uint, entityType string) ([]Translation, error)
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

// CreateCategoryType creates a new category type in the database
func (s *SQLStore) CreateCategoryType(ctx context.Context, categoryType *CategoryType) error {
	result := s.db.WithContext(ctx).Create(categoryType)
	if result.Error != nil {
		return fmt.Errorf("failed to create category type: %w", result.Error)
	}
	return nil
}

// UpdateCategoryType updates an existing category type
func (s *SQLStore) UpdateCategoryType(ctx context.Context, categoryType *CategoryType) error {
	result := s.db.WithContext(ctx).Save(categoryType)
	if result.Error != nil {
		return fmt.Errorf("failed to update category type: %w", result.Error)
	}
	return nil
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

// ListCategoryTypes returns all category types
func (s *SQLStore) ListCategoryTypes(ctx context.Context) ([]CategoryType, error) {
	var categoryTypes []CategoryType
	result := s.db.WithContext(ctx).Find(&categoryTypes)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list category types: %w", result.Error)
	}
	return categoryTypes, nil
}

// CreateCategory creates a new category in the database
func (s *SQLStore) CreateCategory(ctx context.Context, category *Category) error {
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

// DeleteCategory deletes a category from the database
func (s *SQLStore) DeleteCategory(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Delete(&Category{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete category: %w", result.Error)
	}
	return nil
}

// CreateSubcategory creates a new subcategory in the database
func (s *SQLStore) CreateSubcategory(ctx context.Context, subcategory *Subcategory) error {
	// Validate required fields
	if len(subcategory.Translations) == 0 {
		return fmt.Errorf("at least one translation is required")
	}

	result := s.db.WithContext(ctx).Create(subcategory)
	if result.Error != nil {
		return fmt.Errorf("failed to create subcategory: %w", result.Error)
	}
	return nil
}

// UpdateSubcategory updates an existing subcategory
func (s *SQLStore) UpdateSubcategory(ctx context.Context, subcategory *Subcategory) error {
	result := s.db.WithContext(ctx).Save(subcategory)
	if result.Error != nil {
		return fmt.Errorf("failed to update subcategory: %w", result.Error)
	}
	return nil
}

// GetSubcategoryByID retrieves a subcategory by its ID
func (s *SQLStore) GetSubcategoryByID(ctx context.Context, id uint) (*Subcategory, error) {
	var subcategory Subcategory
	result := s.db.WithContext(ctx).First(&subcategory, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get subcategory: %w", result.Error)
	}
	return &subcategory, nil
}

// ListSubcategories returns all subcategories
func (s *SQLStore) ListSubcategories(ctx context.Context) ([]Subcategory, error) {
	var subcategories []Subcategory
	result := s.db.WithContext(ctx).Find(&subcategories)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list subcategories: %w", result.Error)
	}
	return subcategories, nil
}

// DeleteSubcategory deletes a subcategory from the database
func (s *SQLStore) DeleteSubcategory(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Delete(&Subcategory{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete subcategory: %w", result.Error)
	}
	return nil
}

// CreateCategorySubcategory creates a new category-subcategory relationship in the database
func (s *SQLStore) CreateCategorySubcategory(ctx context.Context, link *CategorySubcategory) error {
	if err := s.db.WithContext(ctx).Create(link).Error; err != nil {
		return fmt.Errorf("failed to create category-subcategory relationship: %w", err)
	}
	return nil
}

// DeleteCategorySubcategory deletes a category-subcategory relationship from the database
func (s *SQLStore) DeleteCategorySubcategory(ctx context.Context, categoryID, subcategoryID uint) error {
	result := s.db.WithContext(ctx).Delete(&CategorySubcategory{}, "category_id = ? AND subcategory_id = ?", categoryID, subcategoryID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete category-subcategory relationship: %w", result.Error)
	}
	return nil
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

// UpdateTranslation updates an existing translation in the database
func (s *SQLStore) UpdateTranslation(ctx context.Context, translation *Translation) error {
	result := s.db.WithContext(ctx).Save(translation)
	if result.Error != nil {
		return fmt.Errorf("failed to update translation: %w", result.Error)
	}
	return nil
}

// GetTranslationByID retrieves a translation by its ID
func (s *SQLStore) GetTranslationByID(ctx context.Context, id uint) (*Translation, error) {
	var translation Translation
	result := s.db.WithContext(ctx).First(&translation, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get translation: %w", result.Error)
	}
	return &translation, nil
}

// ListTranslations retrieves all translations for a given entity
func (s *SQLStore) ListTranslations(ctx context.Context, entityType string, entityID uint) ([]Translation, error) {
	var translations []Translation
	result := s.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Find(&translations)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get translations: %w", result.Error)
	}
	return translations, nil
}

// CreateTransaction creates a new transaction in the database
func (s *SQLStore) CreateTransaction(ctx context.Context, transaction *Transaction) error {
	if err := s.db.WithContext(ctx).Create(transaction).Error; err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}
	return nil
}

// UpdateTransaction updates an existing transaction in the database
func (s *SQLStore) UpdateTransaction(ctx context.Context, transaction *Transaction) error {
	result := s.db.WithContext(ctx).Save(transaction)
	if result.Error != nil {
		return fmt.Errorf("failed to update transaction: %w", result.Error)
	}
	return nil
}

// GetTransactionByID retrieves a transaction by its ID
func (s *SQLStore) GetTransactionByID(ctx context.Context, id uint) (*Transaction, error) {
	var transaction Transaction
	result := s.db.WithContext(ctx).First(&transaction, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get transaction: %w", result.Error)
	}
	return &transaction, nil
}

// ListTransactions retrieves all transactions, optionally filtered by filter
func (s *SQLStore) ListTransactions(ctx context.Context, filter *TransactionFilter) ([]Transaction, error) {
	var transactions []Transaction
	query := s.db.WithContext(ctx)
	if filter != nil {
		if filter.CategoryID != nil {
			query = query.Where("category_id = ?", *filter.CategoryID)
		}
		if filter.SubcategoryID != nil {
			query = query.Where("subcategory_id = ?", *filter.SubcategoryID)
		}
		if filter.StartDate != nil {
			query = query.Where("created_at >= ?", *filter.StartDate)
		}
		if filter.EndDate != nil {
			query = query.Where("created_at <= ?", *filter.EndDate)
		}
	}
	result := query.Find(&transactions)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", result.Error)
	}
	return transactions, nil
}

// DeleteTransaction deletes a transaction from the database
func (s *SQLStore) DeleteTransaction(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Delete(&Transaction{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete transaction: %w", result.Error)
	}
	return nil
}

// CreatePrompt creates a new prompt template in the database
func (s *SQLStore) CreatePrompt(ctx context.Context, prompt *Prompt) error {
	if err := s.db.WithContext(ctx).Create(prompt).Error; err != nil {
		return fmt.Errorf("failed to create prompt: %w", err)
	}
	return nil
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

// GetPromptByID retrieves a prompt template by its ID
func (s *SQLStore) GetPromptByID(ctx context.Context, id uint) (*Prompt, error) {
	var prompt Prompt
	result := s.db.WithContext(ctx).First(&prompt, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get prompt: %w", result.Error)
	}
	return &prompt, nil
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

// ListPrompts returns all prompt templates
func (s *SQLStore) ListPrompts(ctx context.Context) ([]Prompt, error) {
	var prompts []Prompt
	result := s.db.WithContext(ctx).Find(&prompts)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list prompts: %w", result.Error)
	}
	return prompts, nil
}

// DeletePrompt deletes a prompt template from the database
func (s *SQLStore) DeletePrompt(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Delete(&Prompt{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete prompt: %w", result.Error)
	}
	return nil
}

// Close closes the database connection
func (s *SQLStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}
	return sqlDB.Close()
}

// GetTranslations retrieves translations for a given entity
func (s *SQLStore) GetTranslations(ctx context.Context, entityID uint, entityType string) ([]Translation, error) {
	var translations []Translation
	result := s.db.WithContext(ctx).Where("entity_id = ? AND entity_type = ?", entityID, entityType).Find(&translations)
	return translations, result.Error
}
