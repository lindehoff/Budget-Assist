package db

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// MockStore is a mock implementation of the Store interface for testing
type MockStore struct {
	prompts           map[string]*Prompt
	categories        map[uint]*Category
	subcategories     map[uint]*Subcategory
	categoryTypes     map[uint]*CategoryType
	transactions      map[uint]*Transaction
	tags              map[string]*Tag
	categoryTypeNames map[string]*CategoryType
	nextID            uint
	mu                sync.RWMutex
	lastID            uint
}

// NewMockStore creates a new MockStore instance
func NewMockStore() *MockStore {
	return &MockStore{
		prompts:           make(map[string]*Prompt),
		categories:        make(map[uint]*Category),
		subcategories:     make(map[uint]*Subcategory),
		categoryTypes:     make(map[uint]*CategoryType),
		transactions:      make(map[uint]*Transaction),
		tags:              make(map[string]*Tag),
		categoryTypeNames: make(map[string]*CategoryType),
		nextID:            1,
		lastID:            0,
	}
}

// CreateCategoryType implements Store
func (s *MockStore) CreateCategoryType(ctx context.Context, categoryType *CategoryType) error {
	if categoryType == nil {
		return fmt.Errorf("category type cannot be nil")
	}
	categoryType.ID = s.nextID
	s.nextID++
	s.categoryTypes[categoryType.ID] = categoryType
	s.categoryTypeNames[categoryType.Name] = categoryType
	return nil
}

// UpdateCategoryType implements Store
func (s *MockStore) UpdateCategoryType(ctx context.Context, categoryType *CategoryType) error {
	if categoryType == nil {
		return fmt.Errorf("category type cannot be nil")
	}
	if _, exists := s.categoryTypes[categoryType.ID]; !exists {
		return ErrNotFound
	}
	s.categoryTypes[categoryType.ID] = categoryType
	s.categoryTypeNames[categoryType.Name] = categoryType
	return nil
}

// GetCategoryTypeByID implements Store
func (s *MockStore) GetCategoryTypeByID(ctx context.Context, id uint) (*CategoryType, error) {
	categoryType, exists := s.categoryTypes[id]
	if !exists {
		return nil, ErrNotFound
	}
	return categoryType, nil
}

// ListCategoryTypes implements Store
func (s *MockStore) ListCategoryTypes(ctx context.Context) ([]CategoryType, error) {
	types := make([]CategoryType, 0, len(s.categoryTypes))
	for _, t := range s.categoryTypes {
		types = append(types, *t)
	}
	return types, nil
}

// CreateCategory implements Store
func (s *MockStore) CreateCategory(ctx context.Context, category *Category) error {
	if category == nil {
		return fmt.Errorf("category cannot be nil")
	}
	if category.TypeID == 0 {
		return fmt.Errorf("type_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	category.ID = s.nextID
	s.nextID++
	s.categories[category.ID] = category
	return nil
}

// UpdateCategory implements Store
func (s *MockStore) UpdateCategory(ctx context.Context, category *Category) error {
	if category == nil {
		return fmt.Errorf("category cannot be nil")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.categories[category.ID]; !exists {
		return ErrNotFound
	}
	s.categories[category.ID] = category
	return nil
}

// GetCategoryByID implements Store
func (s *MockStore) GetCategoryByID(ctx context.Context, id uint) (*Category, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	category, exists := s.categories[id]
	if !exists {
		return nil, ErrNotFound
	}
	return category, nil
}

// GetCategoryByName implements Store
func (s *MockStore) GetCategoryByName(ctx context.Context, name string) (*Category, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, category := range s.categories {
		if category.Name == name {
			return category, nil
		}
	}
	return nil, ErrNotFound
}

// ListCategories implements Store
func (s *MockStore) ListCategories(ctx context.Context, typeID *uint) ([]Category, error) {
	categories := make([]Category, 0, len(s.categories))
	for _, category := range s.categories {
		if typeID == nil || category.TypeID == *typeID {
			categories = append(categories, *category)
		}
	}
	return categories, nil
}

// DeleteCategory implements Store
func (s *MockStore) DeleteCategory(ctx context.Context, id uint) error {
	if _, exists := s.categories[id]; !exists {
		return ErrNotFound
	}
	delete(s.categories, id)
	return nil
}

// CreateSubcategory implements Store
func (s *MockStore) CreateSubcategory(ctx context.Context, subcategory *Subcategory) error {
	if subcategory == nil {
		return fmt.Errorf("subcategory cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate ID
	subcategory.ID = s.nextID
	s.nextID++

	// Store the subcategory
	s.subcategories[subcategory.ID] = subcategory
	return nil
}

// UpdateSubcategory implements Store
func (s *MockStore) UpdateSubcategory(ctx context.Context, subcategory *Subcategory) error {
	if subcategory == nil {
		return fmt.Errorf("subcategory cannot be nil")
	}
	if _, exists := s.subcategories[subcategory.ID]; !exists {
		return ErrNotFound
	}
	s.subcategories[subcategory.ID] = subcategory
	return nil
}

// GetSubcategoryByID implements Store
func (s *MockStore) GetSubcategoryByID(ctx context.Context, id uint) (*Subcategory, error) {
	subcategory, exists := s.subcategories[id]
	if !exists {
		return nil, ErrNotFound
	}
	return subcategory, nil
}

// GetSubcategoryByName implements Store
func (s *MockStore) GetSubcategoryByName(ctx context.Context, name string) (*Subcategory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, subcategory := range s.subcategories {
		if subcategory.Name == name {
			return subcategory, nil
		}
	}
	return nil, ErrNotFound
}

// ListSubcategories implements Store
func (s *MockStore) ListSubcategories(ctx context.Context) ([]Subcategory, error) {
	subcategories := make([]Subcategory, 0, len(s.subcategories))
	for _, subcategory := range s.subcategories {
		subcategories = append(subcategories, *subcategory)
	}
	return subcategories, nil
}

// DeleteSubcategory implements Store
func (s *MockStore) DeleteSubcategory(ctx context.Context, id uint) error {
	if _, exists := s.subcategories[id]; !exists {
		return ErrNotFound
	}
	delete(s.subcategories, id)
	return nil
}

// CreateCategorySubcategory implements Store
func (s *MockStore) CreateCategorySubcategory(ctx context.Context, link *CategorySubcategory) error {
	if link == nil {
		return fmt.Errorf("link cannot be nil")
	}
	category, exists := s.categories[link.CategoryID]
	if !exists {
		return ErrNotFound
	}
	category.Subcategories = append(category.Subcategories, *link)
	return nil
}

// DeleteCategorySubcategory implements Store
func (s *MockStore) DeleteCategorySubcategory(ctx context.Context, categoryID, subcategoryID uint) error {
	category, exists := s.categories[categoryID]
	if !exists {
		return ErrNotFound
	}
	for i, link := range category.Subcategories {
		if link.SubcategoryID == subcategoryID {
			category.Subcategories = append(category.Subcategories[:i], category.Subcategories[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

// CreateTransaction implements Store
func (s *MockStore) CreateTransaction(ctx context.Context, transaction *Transaction) error {
	if transaction == nil {
		return fmt.Errorf("transaction cannot be nil")
	}
	transaction.ID = s.nextID
	s.nextID++
	s.transactions[transaction.ID] = transaction
	return nil
}

// UpdateTransaction implements Store
func (s *MockStore) UpdateTransaction(ctx context.Context, transaction *Transaction) error {
	if transaction == nil {
		return fmt.Errorf("transaction cannot be nil")
	}
	if _, exists := s.transactions[transaction.ID]; !exists {
		return ErrNotFound
	}
	s.transactions[transaction.ID] = transaction
	return nil
}

// GetTransactionByID implements Store
func (s *MockStore) GetTransactionByID(ctx context.Context, id uint) (*Transaction, error) {
	transaction, exists := s.transactions[id]
	if !exists {
		return nil, ErrNotFound
	}
	return transaction, nil
}

// ListTransactions implements Store
func (s *MockStore) ListTransactions(ctx context.Context, filter *TransactionFilter) ([]Transaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var transactions []Transaction
	for _, tx := range s.transactions {
		matches := true

		if filter != nil {
			if filter.CategoryID != nil && (tx.CategoryID == nil || *tx.CategoryID != *filter.CategoryID) {
				matches = false
			}
			if filter.SubcategoryID != nil && (tx.SubcategoryID == nil || *tx.SubcategoryID != *filter.SubcategoryID) {
				matches = false
			}
			if filter.StartDate != nil && tx.Date.Before(*filter.StartDate) {
				matches = false
			}
			if filter.EndDate != nil && tx.Date.After(*filter.EndDate) {
				matches = false
			}
		}

		if matches {
			transactions = append(transactions, *tx)
		}
	}

	return transactions, nil
}

// DeleteTransaction implements Store
func (s *MockStore) DeleteTransaction(ctx context.Context, id uint) error {
	if _, exists := s.transactions[id]; !exists {
		return ErrNotFound
	}
	delete(s.transactions, id)
	return nil
}

// CreatePrompt implements Store
func (s *MockStore) CreatePrompt(ctx context.Context, prompt *Prompt) error {
	if prompt == nil {
		return fmt.Errorf("prompt cannot be nil")
	}
	s.prompts[string(prompt.Type)] = prompt
	return nil
}

// UpdatePrompt implements Store
func (s *MockStore) UpdatePrompt(ctx context.Context, prompt *Prompt) error {
	if prompt == nil {
		return fmt.Errorf("prompt cannot be nil")
	}
	s.prompts[string(prompt.Type)] = prompt
	return nil
}

// GetPromptByID implements Store
func (s *MockStore) GetPromptByID(ctx context.Context, id uint) (*Prompt, error) {
	for _, prompt := range s.prompts {
		if prompt.ID == id {
			return prompt, nil
		}
	}
	return nil, ErrNotFound
}

// GetPromptByType implements Store
func (s *MockStore) GetPromptByType(ctx context.Context, promptType string) (*Prompt, error) {
	prompt, ok := s.prompts[promptType]
	if !ok {
		return nil, fmt.Errorf("prompt template not found for type: %s", promptType)
	}
	return prompt, nil
}

// ListPrompts implements Store
func (s *MockStore) ListPrompts(ctx context.Context) ([]Prompt, error) {
	prompts := make([]Prompt, 0, len(s.prompts))
	// First, collect all types to sort them
	types := make([]string, 0, len(s.prompts))
	for promptType := range s.prompts {
		types = append(types, promptType)
	}
	// Sort types to ensure deterministic order
	sort.Strings(types)
	// Add prompts in sorted order
	for _, promptType := range types {
		prompts = append(prompts, *s.prompts[promptType])
	}
	return prompts, nil
}

// DeletePrompt implements Store
func (s *MockStore) DeletePrompt(ctx context.Context, id uint) error {
	for promptType, prompt := range s.prompts {
		if prompt.ID == id {
			delete(s.prompts, promptType)
			return nil
		}
	}
	return ErrNotFound
}

// CreateTag implements Store
func (s *MockStore) CreateTag(ctx context.Context, tag *Tag) error {
	if tag == nil {
		return fmt.Errorf("tag cannot be nil")
	}
	s.tags[tag.Name] = tag
	return nil
}

// GetTagByName implements Store
func (s *MockStore) GetTagByName(ctx context.Context, name string) (*Tag, error) {
	tag, exists := s.tags[name]
	if !exists {
		return nil, ErrNotFound
	}
	return tag, nil
}

// LinkSubcategoryTag implements Store
func (s *MockStore) LinkSubcategoryTag(ctx context.Context, subcategoryID, tagID uint) error {
	subcategory, exists := s.subcategories[subcategoryID]
	if !exists {
		return ErrNotFound
	}
	for _, tag := range s.tags {
		if tag.ID == tagID {
			subcategory.Tags = append(subcategory.Tags, *tag)
			return nil
		}
	}
	return ErrNotFound
}

// UnlinkSubcategoryTag implements Store
func (s *MockStore) UnlinkSubcategoryTag(ctx context.Context, subcategoryID, tagID uint) error {
	subcategory, exists := s.subcategories[subcategoryID]
	if !exists {
		return ErrNotFound
	}
	for i, tag := range subcategory.Tags {
		if tag.ID == tagID {
			subcategory.Tags = append(subcategory.Tags[:i], subcategory.Tags[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

// Close implements Store
func (s *MockStore) Close() error {
	return nil
}

// Reset clears all stored data
func (s *MockStore) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.categories = make(map[uint]*Category)
	s.categoryTypes = make(map[uint]*CategoryType)
	s.subcategories = make(map[uint]*Subcategory)
	s.categoryTypeNames = make(map[string]*CategoryType)
	s.nextID = 1
	s.lastID = 0
}
