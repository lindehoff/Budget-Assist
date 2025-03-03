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
	pm.logger.Debug("Retrieving prompt template", "type", promptType)

	// Try to get from cache first with read lock
	pm.mu.RLock()
	if template, ok := pm.templates[promptType]; ok && template.IsActive {
		pm.mu.RUnlock()
		pm.logger.Debug("Prompt template found in cache",
			"type", promptType,
			"version", template.Version)
		return template, nil
	}
	pm.mu.RUnlock()

	pm.logger.Debug("Prompt template not found in cache, retrieving from database", "type", promptType)

	// Get from database
	dbPrompt, err := pm.store.GetPromptByType(ctx, string(promptType))
	if err != nil {
		pm.logger.Error("Failed to retrieve prompt template from database",
			"error", err,
			"type", promptType)
		return nil, err
	}
	if dbPrompt == nil {
		pm.logger.Warn("Prompt template not found in database", "type", promptType)
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
	if cachedTemplate, ok := pm.templates[promptType]; ok && cachedTemplate.IsActive {
		pm.logger.Debug("Prompt template was cached by another goroutine",
			"type", promptType,
			"version", cachedTemplate.Version)
		return cachedTemplate, nil
	}

	pm.templates[promptType] = template
	pm.logger.Info("Prompt template retrieved and cached",
		"type", promptType,
		"name", template.Name,
		"version", template.Version)
	return template, nil
}

// UpdatePrompt updates or creates a prompt template
func (pm *PromptManager) UpdatePrompt(ctx context.Context, template *PromptTemplate) error {
	if template == nil {
		pm.logger.Error("Attempted to update nil template")
		return fmt.Errorf("template cannot be nil")
	}

	pm.logger.Debug("Updating prompt template",
		"type", template.Type,
		"name", template.Name,
		"version", template.Version)

	// Basic validation
	if err := template.Validate(); err != nil {
		pm.logger.Warn("Prompt template validation failed",
			"error", err,
			"type", template.Type)
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
		pm.logger.Error("Failed to update prompt template in database",
			"error", err,
			"type", template.Type)
		return fmt.Errorf("failed to update prompt in database: %w", err)
	}

	// Update the cache
	pm.mu.Lock()
	pm.templates[template.Type] = template
	pm.mu.Unlock()

	pm.logger.Info("Prompt template updated successfully",
		"type", template.Type,
		"name", template.Name,
		"version", template.Version)
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

// DeletePrompt removes a prompt template from the cache and marks it as inactive
func (pm *PromptManager) DeletePrompt(ctx context.Context, promptType db.PromptType) error {
	pm.logger.Debug("Deleting prompt template", "type", promptType)

	// Remove from cache
	pm.mu.Lock()
	delete(pm.templates, promptType)
	pm.mu.Unlock()

	// Get the prompt from the database
	dbPrompt, err := pm.store.GetPromptByType(ctx, string(promptType))
	if err != nil {
		pm.logger.Error("Failed to retrieve prompt template for deletion",
			"error", err,
			"type", promptType)
		return fmt.Errorf("failed to retrieve prompt for deletion: %w", err)
	}

	if dbPrompt == nil {
		pm.logger.Warn("Prompt template not found for deletion", "type", promptType)
		return nil
	}

	// Mark as inactive
	dbPrompt.IsActive = false
	if err := pm.store.UpdatePrompt(ctx, dbPrompt); err != nil {
		pm.logger.Error("Failed to mark prompt template as inactive",
			"error", err,
			"type", promptType)
		return fmt.Errorf("failed to mark prompt as inactive: %w", err)
	}

	pm.logger.Info("Prompt template deleted successfully", "type", promptType)
	return nil
}
