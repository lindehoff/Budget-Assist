package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/lindehoff/Budget-Assist/internal/db"
)

// PromptManager handles storage and management of prompt templates
type PromptManager struct {
	templates map[PromptType]*PromptTemplate
	mu        sync.RWMutex
	logger    *slog.Logger
	store     db.Store
}

// NewPromptManager creates a new prompt manager
func NewPromptManager(store db.Store, logger *slog.Logger) *PromptManager {
	return &PromptManager{
		templates: make(map[PromptType]*PromptTemplate),
		logger:    logger,
		store:     store,
	}
}

// GetPrompt retrieves a prompt template by type
func (pm *PromptManager) GetPrompt(ctx context.Context, promptType PromptType) (*PromptTemplate, error) {
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
		return nil, fmt.Errorf("prompt template not found for type: %s: %w", promptType, err)
	}
	if dbPrompt == nil {
		return nil, fmt.Errorf("prompt template not found for type: %s", promptType)
	}

	template := &PromptTemplate{
		Type:         PromptType(dbPrompt.Type),
		Name:         dbPrompt.Name,
		SystemPrompt: dbPrompt.SystemPrompt,
		UserPrompt:   dbPrompt.UserPrompt,
		Version:      dbPrompt.Version,
		IsActive:     dbPrompt.IsActive,
		CreatedAt:    dbPrompt.CreatedAt,
		UpdatedAt:    dbPrompt.UpdatedAt,
	}

	// Unmarshal examples and rules if they exist
	if dbPrompt.Examples != "" {
		if err := json.Unmarshal([]byte(dbPrompt.Examples), &template.Examples); err != nil {
			pm.logger.Error("failed to unmarshal examples",
				"type", promptType,
				"error", err)
		}
	}
	if dbPrompt.Rules != "" {
		if err := json.Unmarshal([]byte(dbPrompt.Rules), &template.Rules); err != nil {
			pm.logger.Error("failed to unmarshal rules",
				"type", promptType,
				"error", err)
		}
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
	if template.Type == "" {
		return fmt.Errorf("prompt type is required")
	}
	if template.Name == "" {
		return fmt.Errorf("prompt name is required")
	}
	if template.SystemPrompt == "" {
		return fmt.Errorf("system prompt is required")
	}
	if template.UserPrompt == "" {
		return fmt.Errorf("user prompt is required")
	}

	// Convert examples to JSON if present
	var examplesJSON string
	if len(template.Examples) > 0 {
		data, err := json.Marshal(template.Examples)
		if err != nil {
			return fmt.Errorf("failed to marshal examples: %w", err)
		}
		examplesJSON = string(data)
	}

	// Convert rules to JSON if present
	var rulesJSON string
	if len(template.Rules) > 0 {
		data, err := json.Marshal(template.Rules)
		if err != nil {
			return fmt.Errorf("failed to marshal rules: %w", err)
		}
		rulesJSON = string(data)
	}

	// Create or update the prompt in the database
	dbPrompt := &db.Prompt{
		Type:         string(template.Type),
		Name:         template.Name,
		SystemPrompt: template.SystemPrompt,
		UserPrompt:   template.UserPrompt,
		Examples:     examplesJSON,
		Rules:        rulesJSON,
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

// ListPrompts returns all active prompt templates
func (pm *PromptManager) ListPrompts(ctx context.Context) []*PromptTemplate {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Get from database
	dbPrompts, err := pm.store.ListPrompts(ctx)
	if err != nil {
		pm.logger.Error("failed to list prompts from database",
			"error", err)
		return nil
	}

	templates := make([]*PromptTemplate, 0, len(dbPrompts))
	for _, dbPrompt := range dbPrompts {
		if !dbPrompt.IsActive {
			continue
		}

		template := &PromptTemplate{
			Type:         PromptType(dbPrompt.Type),
			Name:         dbPrompt.Name,
			SystemPrompt: dbPrompt.SystemPrompt,
			UserPrompt:   dbPrompt.UserPrompt,
			Version:      dbPrompt.Version,
			IsActive:     dbPrompt.IsActive,
			CreatedAt:    dbPrompt.CreatedAt,
			UpdatedAt:    dbPrompt.UpdatedAt,
		}

		// Unmarshal examples and rules if they exist
		if dbPrompt.Examples != "" {
			if err := json.Unmarshal([]byte(dbPrompt.Examples), &template.Examples); err != nil {
				pm.logger.Error("failed to unmarshal examples",
					"type", dbPrompt.Type,
					"error", err)
			}
		}
		if dbPrompt.Rules != "" {
			if err := json.Unmarshal([]byte(dbPrompt.Rules), &template.Rules); err != nil {
				pm.logger.Error("failed to unmarshal rules",
					"type", dbPrompt.Type,
					"error", err)
			}
		}

		templates = append(templates, template)
		pm.templates[template.Type] = template
	}

	return templates
}

// DeactivatePrompt deactivates a prompt template
func (pm *PromptManager) DeactivatePrompt(ctx context.Context, promptType PromptType) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Get from database
	dbPrompt, err := pm.store.GetPromptByType(ctx, string(promptType))
	if err != nil {
		return fmt.Errorf("prompt template not found: %s", promptType)
	}
	if dbPrompt == nil {
		return fmt.Errorf("prompt template not found: %s", promptType)
	}

	dbPrompt.IsActive = false
	dbPrompt.UpdatedAt = time.Now()

	if err := pm.store.UpdatePrompt(ctx, dbPrompt); err != nil {
		return fmt.Errorf("failed to update prompt in database: %w", err)
	}

	// Update cache
	if template, ok := pm.templates[promptType]; ok {
		template.IsActive = false
		template.UpdatedAt = dbPrompt.UpdatedAt
	}

	pm.logger.Info("prompt template deactivated",
		"type", promptType,
		"version", dbPrompt.Version)

	return nil
}

// incrementVersion increments the version number (e.g., "1.0.9" -> "1.1.0" for minor, "1.9.9" -> "2.0.0" for major)
func incrementVersion(version string) string {
	var major, minor, patch int
	if _, err := fmt.Sscanf(version, "%d.%d.%d", &major, &minor, &patch); err != nil {
		// If we can't parse the version, return a default increment
		return "1.0.0"
	}

	// If patch is 9, increment minor and reset patch
	if patch == 9 {
		patch = 0
		// If minor is 9, increment major and reset minor
		if minor == 9 {
			major++
			minor = 0
		} else {
			minor++
		}
	} else {
		patch++
	}

	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}
