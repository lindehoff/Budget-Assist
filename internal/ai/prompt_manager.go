package ai

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// PromptManager handles storage and management of prompt templates
type PromptManager struct {
	templates map[PromptType]*PromptTemplate
	mu        sync.RWMutex
	logger    *slog.Logger
}

// NewPromptManager creates a new prompt manager
func NewPromptManager(logger *slog.Logger) *PromptManager {
	return &PromptManager{
		templates: make(map[PromptType]*PromptTemplate),
		logger:    logger,
	}
}

// GetPrompt retrieves a prompt template by type
func (pm *PromptManager) GetPrompt(ctx context.Context, promptType PromptType) (*PromptTemplate, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	template, ok := pm.templates[promptType]
	if !ok {
		return nil, fmt.Errorf("prompt template not found for type: %s", promptType)
	}

	if !template.IsActive {
		return nil, fmt.Errorf("prompt template is not active: %s", promptType)
	}

	return template, nil
}

// UpdatePrompt updates or creates a prompt template
func (pm *PromptManager) UpdatePrompt(ctx context.Context, template *PromptTemplate) error {
	if err := template.Validate(); err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// If updating existing template, increment version
	if existing, ok := pm.templates[template.Type]; ok {
		template.Version = incrementVersion(existing.Version)
	}

	template.UpdatedAt = time.Now()
	pm.templates[template.Type] = template

	pm.logger.Info("prompt template updated",
		"type", template.Type,
		"version", template.Version)

	return nil
}

// ListPrompts returns all active prompt templates
func (pm *PromptManager) ListPrompts(ctx context.Context) []*PromptTemplate {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	templates := make([]*PromptTemplate, 0, len(pm.templates))
	for _, tpl := range pm.templates {
		if tpl.IsActive {
			templates = append(templates, tpl)
		}
	}

	return templates
}

// DeactivatePrompt deactivates a prompt template
func (pm *PromptManager) DeactivatePrompt(ctx context.Context, promptType PromptType) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	template, ok := pm.templates[promptType]
	if !ok {
		return fmt.Errorf("prompt template not found: %s", promptType)
	}

	template.IsActive = false
	template.UpdatedAt = time.Now()

	pm.logger.Info("prompt template deactivated",
		"type", promptType,
		"version", template.Version)

	return nil
}

// TestPrompt tests a prompt template with sample data
func (pm *PromptManager) TestPrompt(ctx context.Context, promptType PromptType, data interface{}) (string, error) {
	template, err := pm.GetPrompt(ctx, promptType)
	if err != nil {
		return "", err
	}

	result, err := template.Execute(data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return result, nil
}

// incrementVersion increments the version number (e.g., "1.0.0" -> "1.0.1")
func incrementVersion(version string) string {
	var major, minor, patch int
	if _, err := fmt.Sscanf(version, "%d.%d.%d", &major, &minor, &patch); err != nil {
		// If we can't parse the version, return a default increment
		return "1.0.0"
	}
	return fmt.Sprintf("%d.%d.%d", major, minor, patch+1)
}
