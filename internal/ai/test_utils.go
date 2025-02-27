package ai

import (
	"context"
	"sync"

	db "github.com/lindehoff/Budget-Assist/internal/db"
)

// MockStore implements a mock database store for testing
type MockStore struct {
	categories    []db.Category
	subcategories []db.Subcategory
	prompts       map[string]*db.Prompt
	err           error
	nextID        uint
	mu            sync.Mutex
	Categories    []*db.Category
}

func NewMockStore() *MockStore {
	return &MockStore{
		categories:    make([]db.Category, 0),
		subcategories: make([]db.Subcategory, 0),
		prompts:       make(map[string]*db.Prompt),
		nextID:        1,
	}
}

func (m *MockStore) CreateCategory(ctx context.Context, category *db.Category) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return m.err
	}

	category.ID = m.nextID
	m.nextID++
	m.categories = append(m.categories, *category)
	return nil
}

func (m *MockStore) UpdateCategory(ctx context.Context, category *db.Category) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return m.err
	}

	for i := range m.categories {
		if m.categories[i].ID == category.ID {
			m.categories[i] = *category
			return nil
		}
	}
	return db.ErrNotFound
}

func (m *MockStore) GetCategoryByID(ctx context.Context, id uint) (*db.Category, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return nil, m.err
	}

	for i := range m.categories {
		if m.categories[i].ID == id {
			return &m.categories[i], nil
		}
	}
	return nil, db.ErrNotFound
}

func (m *MockStore) ListCategories(ctx context.Context, typeID *uint) ([]db.Category, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return nil, m.err
	}

	if typeID == nil {
		return m.categories, nil
	}

	var filtered []db.Category
	for _, cat := range m.categories {
		if cat.TypeID == *typeID {
			filtered = append(filtered, cat)
		}
	}
	return filtered, nil
}

func (m *MockStore) GetCategoryTypeByID(ctx context.Context, id uint) (*db.CategoryType, error) {
	return nil, nil
}

func (m *MockStore) CreateTranslation(ctx context.Context, translation *db.Translation) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if translation.EntityType == string(db.EntityTypeCategory) {
		// Find the category in the slice
		var category *db.Category
		for i := range m.categories {
			if m.categories[i].ID == translation.EntityID {
				category = &m.categories[i]
				break
			}
		}
		if category == nil {
			return db.ErrNotFound
		}

		// Update or add translation
		found := false
		for i, t := range category.Translations {
			if t.LanguageCode == translation.LanguageCode {
				category.Translations[i] = *translation
				found = true
				break
			}
		}
		if !found {
			category.Translations = append(category.Translations, *translation)
		}
	}
	return nil
}

func (m *MockStore) GetTranslations(ctx context.Context, entityID uint, entityType string) ([]db.Translation, error) {
	return nil, nil
}

func (m *MockStore) DeleteCategory(ctx context.Context, id uint) error {
	return nil
}

func (m *MockStore) GetPromptByType(ctx context.Context, promptType string) (*db.Prompt, error) {
	if prompt, ok := m.prompts[promptType]; ok {
		return prompt, nil
	}
	return nil, nil
}

func (m *MockStore) UpdatePrompt(ctx context.Context, prompt *db.Prompt) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.prompts[prompt.Type] = prompt
	return nil
}

func (m *MockStore) ListPrompts(ctx context.Context) ([]db.Prompt, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	prompts := make([]db.Prompt, 0, len(m.prompts))
	for _, p := range m.prompts {
		prompts = append(prompts, *p)
	}
	return prompts, nil
}

func (m *MockStore) StoreTransaction(ctx context.Context, tx *db.Transaction) error {
	return nil
}

func (m *MockStore) Close() error {
	return nil
}

func (m *MockStore) CreateSubcategory(ctx context.Context, subcategory *db.Subcategory) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return m.err
	}

	subcategory.ID = m.nextID
	m.nextID++
	m.subcategories = append(m.subcategories, *subcategory)
	return nil
}

func (m *MockStore) UpdateSubcategory(ctx context.Context, subcategory *db.Subcategory) error {
	return nil
}

func (m *MockStore) GetSubcategoryByID(ctx context.Context, id uint) (*db.Subcategory, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return nil, m.err
	}

	for i := range m.subcategories {
		if m.subcategories[i].ID == id {
			return &m.subcategories[i], nil
		}
	}
	return nil, db.ErrNotFound
}

func (m *MockStore) ListSubcategories(ctx context.Context) ([]db.Subcategory, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return nil, m.err
	}

	return m.subcategories, nil
}

func (m *MockStore) DeleteSubcategory(ctx context.Context, id uint) error {
	return nil
}

func (m *MockStore) CreateCategorySubcategory(ctx context.Context, link *db.CategorySubcategory) error {
	return nil
}

func (m *MockStore) DeleteCategorySubcategory(ctx context.Context, categoryID, subcategoryID uint) error {
	return nil
}

func (m *MockStore) UpdateTranslation(ctx context.Context, translation *db.Translation) error {
	return nil
}

func (m *MockStore) GetTranslationByID(ctx context.Context, id uint) (*db.Translation, error) {
	return nil, nil
}

func (m *MockStore) ListTranslations(ctx context.Context, entityType string, entityID uint) ([]db.Translation, error) {
	return nil, nil
}

func (m *MockStore) CreateTransaction(ctx context.Context, transaction *db.Transaction) error {
	return nil
}

func (m *MockStore) UpdateTransaction(ctx context.Context, transaction *db.Transaction) error {
	return nil
}

func (m *MockStore) GetTransactionByID(ctx context.Context, id uint) (*db.Transaction, error) {
	return nil, nil
}

func (m *MockStore) ListTransactions(ctx context.Context, filter *db.TransactionFilter) ([]db.Transaction, error) {
	return nil, nil
}

func (m *MockStore) DeleteTransaction(ctx context.Context, id uint) error {
	return nil
}

func (m *MockStore) CreatePrompt(ctx context.Context, prompt *db.Prompt) error {
	if m.prompts == nil {
		m.prompts = make(map[string]*db.Prompt)
	}
	m.prompts[prompt.Type] = prompt
	return nil
}

func (m *MockStore) GetPromptByID(ctx context.Context, id uint) (*db.Prompt, error) {
	return nil, nil
}

func (m *MockStore) DeletePrompt(ctx context.Context, id uint) error {
	return nil
}

func (m *MockStore) GetPromptByTypeAndVersion(ctx context.Context, promptType, version string) (*db.Prompt, error) {
	if m.prompts == nil {
		m.prompts = make(map[string]*db.Prompt)
	}
	prompt, exists := m.prompts[promptType]
	if !exists {
		return nil, db.ErrNotFound
	}
	if prompt.Version != version {
		return nil, db.ErrNotFound
	}
	return prompt, nil
}

// CreateCategoryType implements db.Store
func (m *MockStore) CreateCategoryType(ctx context.Context, categoryType *db.CategoryType) error {
	return nil
}

// UpdateCategoryType implements db.Store
func (m *MockStore) UpdateCategoryType(ctx context.Context, categoryType *db.CategoryType) error {
	return nil
}

// ListCategoryTypes implements db.Store
func (m *MockStore) ListCategoryTypes(ctx context.Context) ([]db.CategoryType, error) {
	return nil, nil
}

// FindSubcategoriesByTag implements db.Store
func (m *MockStore) FindSubcategoriesByTag(ctx context.Context, tag string) ([]db.Subcategory, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return nil, m.err
	}

	var matchingSubcategories []db.Subcategory
	for _, sub := range m.subcategories {
		if sub.HasTag(tag) {
			matchingSubcategories = append(matchingSubcategories, sub)
		}
	}
	return matchingSubcategories, nil
}
