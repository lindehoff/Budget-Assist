package ai

import (
	"context"
	"fmt"

	db "github.com/lindehoff/Budget-Assist/internal/db"
)

type MockStore struct {
	prompts    map[string]*db.Prompt
	categories map[uint]*db.Category
	nextID     uint
}

func NewMockStore() *MockStore {
	return &MockStore{
		prompts:    make(map[string]*db.Prompt),
		categories: make(map[uint]*db.Category),
		nextID:     1,
	}
}

func (m *MockStore) CreateCategory(ctx context.Context, category *db.Category) error {
	if category.Name == "" {
		return db.DatabaseOperationError{
			Operation: "create",
			Entity:    "category",
			Err:       fmt.Errorf("category name is required"),
		}
	}
	category.ID = m.nextID
	m.nextID++
	m.categories[category.ID] = category
	return nil
}

func (m *MockStore) UpdateCategory(ctx context.Context, category *db.Category) error {
	if _, exists := m.categories[category.ID]; !exists {
		return db.ErrNotFound
	}
	m.categories[category.ID] = category
	return nil
}

func (m *MockStore) GetCategoryByID(ctx context.Context, id uint) (*db.Category, error) {
	category, exists := m.categories[id]
	if !exists {
		return nil, db.ErrNotFound
	}
	return category, nil
}

func (m *MockStore) ListCategories(ctx context.Context, typeID *uint) ([]db.Category, error) {
	var categories []db.Category
	for _, category := range m.categories {
		if typeID == nil || category.TypeID == *typeID {
			categories = append(categories, *category)
		}
	}
	return categories, nil
}

func (m *MockStore) GetCategoryTypeByID(ctx context.Context, id uint) (*db.CategoryType, error) {
	return nil, nil
}

func (m *MockStore) CreateTranslation(ctx context.Context, translation *db.Translation) error {
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
	m.prompts[prompt.Type] = prompt
	return nil
}

func (m *MockStore) ListPrompts(ctx context.Context) ([]db.Prompt, error) {
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
	return nil
}

func (m *MockStore) UpdateSubcategory(ctx context.Context, subcategory *db.Subcategory) error {
	return nil
}

func (m *MockStore) GetSubcategoryByID(ctx context.Context, id uint) (*db.Subcategory, error) {
	return nil, nil
}

func (m *MockStore) ListSubcategories(ctx context.Context) ([]db.Subcategory, error) {
	return nil, nil
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
