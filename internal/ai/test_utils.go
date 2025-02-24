package ai

import (
	"context"

	db "github.com/lindehoff/Budget-Assist/internal/db"
)

type mockStore struct {
	prompts map[string]*db.Prompt
}

func newMockStore() *mockStore {
	return &mockStore{
		prompts: make(map[string]*db.Prompt),
	}
}

func (m *mockStore) CreateCategory(ctx context.Context, category *db.Category) error {
	return nil
}

func (m *mockStore) UpdateCategory(ctx context.Context, category *db.Category) error {
	return nil
}

func (m *mockStore) GetCategoryByID(ctx context.Context, id uint) (*db.Category, error) {
	return nil, nil
}

func (m *mockStore) ListCategories(ctx context.Context, typeID *uint) ([]db.Category, error) {
	return nil, nil
}

func (m *mockStore) GetCategoryTypeByID(ctx context.Context, id uint) (*db.CategoryType, error) {
	return nil, nil
}

func (m *mockStore) CreateTranslation(ctx context.Context, translation *db.Translation) error {
	return nil
}

func (m *mockStore) GetTranslations(ctx context.Context, entityID uint, entityType string) ([]db.Translation, error) {
	return nil, nil
}

func (m *mockStore) DeleteCategory(ctx context.Context, id uint) error {
	return nil
}

func (m *mockStore) GetPromptByType(ctx context.Context, promptType string) (*db.Prompt, error) {
	if prompt, ok := m.prompts[promptType]; ok {
		return prompt, nil
	}
	return nil, nil
}

func (m *mockStore) UpdatePrompt(ctx context.Context, prompt *db.Prompt) error {
	m.prompts[prompt.Type] = prompt
	return nil
}

func (m *mockStore) ListPrompts(ctx context.Context) ([]db.Prompt, error) {
	prompts := make([]db.Prompt, 0, len(m.prompts))
	for _, p := range m.prompts {
		prompts = append(prompts, *p)
	}
	return prompts, nil
}
