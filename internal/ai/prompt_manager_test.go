package ai

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

func TestPromptManager_GetPrompt(t *testing.T) {
	logger := slog.Default()
	store := newMockStore()
	pm := NewPromptManager(store, logger)

	// Add a test template
	template := &PromptTemplate{
		Type:         TransactionCategorizationPrompt,
		Name:         "Test Template",
		SystemPrompt: "System prompt",
		UserPrompt:   "User prompt",
		Version:      "1.0.0",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := pm.UpdatePrompt(context.Background(), template); err != nil {
		t.Fatalf("Failed to update template: %v", err)
	}

	tests := []struct {
		name       string
		promptType PromptType
		wantErr    bool
	}{
		{
			name:       "Successfully_get_existing_prompt",
			promptType: TransactionCategorizationPrompt,
			wantErr:    false,
		},
		{
			name:       "Error_get_non_existent_prompt",
			promptType: "non-existent",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pm.GetPrompt(context.Background(), tt.promptType)
			if (err != nil) != tt.wantErr {
				t.Errorf("PromptManager.GetPrompt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("PromptManager.GetPrompt() = nil, want non-nil")
			}
		})
	}
}

func TestPromptManager_UpdatePrompt(t *testing.T) {
	logger := slog.Default()
	store := newMockStore()
	pm := NewPromptManager(store, logger)

	tests := []struct {
		name     string
		template *PromptTemplate
		wantErr  bool
	}{
		{
			name: "Successfully_update_prompt",
			template: &PromptTemplate{
				Type:         TransactionCategorizationPrompt,
				Name:         "Test Template",
				SystemPrompt: "System prompt",
				UserPrompt:   "User prompt",
				Version:      "1.0.0",
				IsActive:     true,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			wantErr: false,
		},
		{
			name:     "Error_update_nil_prompt",
			template: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := pm.UpdatePrompt(context.Background(), tt.template); (err != nil) != tt.wantErr {
				t.Errorf("PromptManager.UpdatePrompt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPromptManager_ListPrompts(t *testing.T) {
	logger := slog.Default()
	store := newMockStore()
	pm := NewPromptManager(store, logger)

	// Add some test templates
	templates := []*PromptTemplate{
		{
			Type:         TransactionCategorizationPrompt,
			Name:         "Template 1",
			SystemPrompt: "System prompt 1",
			UserPrompt:   "User prompt 1",
			Version:      "1.0.0",
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			Type:         "test-type-2",
			Name:         "Template 2",
			SystemPrompt: "System prompt 2",
			UserPrompt:   "User prompt 2",
			Version:      "1.0.0",
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	for _, template := range templates {
		if err := pm.UpdatePrompt(context.Background(), template); err != nil {
			t.Fatalf("Failed to update template: %v", err)
		}
	}

	got := pm.ListPrompts(context.Background())
	if len(got) != len(templates) {
		t.Errorf("PromptManager.ListPrompts() = %v, want %v", len(got), len(templates))
	}
}

func TestPromptManager_DeactivatePrompt(t *testing.T) {
	logger := slog.Default()
	store := newMockStore()
	pm := NewPromptManager(store, logger)

	// Add a test template
	template := &PromptTemplate{
		Type:         TransactionCategorizationPrompt,
		Name:         "Test Template",
		SystemPrompt: "System prompt",
		UserPrompt:   "User prompt",
		Version:      "1.0.0",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := pm.UpdatePrompt(context.Background(), template); err != nil {
		t.Fatalf("Failed to update template: %v", err)
	}

	tests := []struct {
		name       string
		promptType PromptType
		wantErr    bool
	}{
		{
			name:       "Successfully_deactivate_prompt",
			promptType: TransactionCategorizationPrompt,
			wantErr:    false,
		},
		{
			name:       "Error_deactivate_non_existent_prompt",
			promptType: "non-existent",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := pm.DeactivatePrompt(context.Background(), tt.promptType); (err != nil) != tt.wantErr {
				t.Errorf("PromptManager.DeactivatePrompt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIncrementVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "Successfully_increment_patch",
			version: "1.0.0",
			want:    "1.0.1",
		},
		{
			name:    "Successfully_increment_minor",
			version: "1.0.9",
			want:    "1.1.0",
		},
		{
			name:    "Successfully_increment_major",
			version: "1.9.9",
			want:    "2.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := incrementVersion(tt.version)
			if got != tt.want {
				t.Errorf("incrementVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
