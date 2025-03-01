package ai

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/lindehoff/Budget-Assist/internal/db"
)

// PromptManager handles storage and management of prompt templates
type PromptManager struct {
	templates map[db.PromptType]*PromptTemplate
	mu        sync.RWMutex
	logger    *slog.Logger
	store     db.Store
}

// NewPromptManager creates a new prompt manager
func NewPromptManager(store db.Store, logger *slog.Logger) *PromptManager {
	return &PromptManager{
		templates: make(map[db.PromptType]*PromptTemplate),
		logger:    logger,
		store:     store,
	}
}

// GetPrompt retrieves a prompt template by type
func (pm *PromptManager) GetPrompt(ctx context.Context, promptType db.PromptType) (*PromptTemplate, error) {
	// Try to get from cache first with read lock
	pm.mu.RLock()
	if template, ok := pm.templates[promptType]; ok && template.IsActive {
		pm.mu.RUnlock()
		return template, nil
	}
	pm.mu.RUnlock()

	// Get from database
	dbPrompt, err := pm.store.GetPromptByType(ctx, string(promptType))
	if err != nil {
		return nil, err
	}
	if dbPrompt == nil {
		return nil, fmt.Errorf("prompt template not found for type: %s", promptType)
	}

	template := &PromptTemplate{
		Type:         dbPrompt.Type,
		Name:         dbPrompt.Name,
		Description:  dbPrompt.Description,
		SystemPrompt: dbPrompt.SystemPrompt,
		UserPrompt:   dbPrompt.UserPrompt,
		Version:      dbPrompt.Version,
		IsActive:     dbPrompt.IsActive,
	}

	// Cache the template with write lock
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check cache again in case another goroutine has already cached it
	if template, ok := pm.templates[promptType]; ok && template.IsActive {
		return template, nil
	}

	pm.templates[promptType] = template
	return template, nil
}

// UpdatePrompt updates or creates a prompt template
func (pm *PromptManager) UpdatePrompt(ctx context.Context, template *PromptTemplate) error {
	if template == nil {
		return fmt.Errorf("template cannot be nil")
	}

	// Basic validation
	if err := template.Validate(); err != nil {
		return err
	}

	// Create or update the prompt in the database
	dbPrompt := &db.Prompt{
		Type:         template.Type,
		Name:         template.Name,
		Description:  template.Description,
		SystemPrompt: template.SystemPrompt,
		UserPrompt:   template.UserPrompt,
		Version:      template.Version,
		IsActive:     template.IsActive,
	}

	if err := pm.store.UpdatePrompt(ctx, dbPrompt); err != nil {
		return fmt.Errorf("failed to update prompt in database: %w", err)
	}

	// Update the cache
	pm.mu.Lock()
	pm.templates[template.Type] = template
	pm.mu.Unlock()

	return nil
}

// ListPrompts returns all prompt templates
func (pm *PromptManager) ListPrompts(ctx context.Context) ([]*PromptTemplate, error) {
	dbPrompts, err := pm.store.ListPrompts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list prompts: %w", err)
	}

	templates := make([]*PromptTemplate, len(dbPrompts))
	for i, dbPrompt := range dbPrompts {
		templates[i] = &PromptTemplate{
			Type:         dbPrompt.Type,
			Name:         dbPrompt.Name,
			Description:  dbPrompt.Description,
			SystemPrompt: dbPrompt.SystemPrompt,
			UserPrompt:   dbPrompt.UserPrompt,
			Version:      dbPrompt.Version,
			IsActive:     dbPrompt.IsActive,
		}
	}

	return templates, nil
}
