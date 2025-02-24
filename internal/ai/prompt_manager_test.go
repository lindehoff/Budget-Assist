package ai

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

func TestPromptManager_GetPrompt(t *testing.T) {
	logger := slog.Default()
	pm := NewPromptManager(logger)

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
				t.Errorf("GetPrompt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("GetPrompt() returned nil, want non-nil")
			}
		})
	}
}

func TestPromptManager_UpdatePrompt(t *testing.T) {
	logger := slog.Default()
	pm := NewPromptManager(logger)

	tests := []struct {
		name     string
		template *PromptTemplate
		wantErr  bool
	}{
		{
			name: "Successfully_update_new_prompt",
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
			name: "Error_update_invalid_prompt",
			template: &PromptTemplate{
				Name:     "Invalid Template",
				IsActive: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pm.UpdatePrompt(context.Background(), tt.template)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdatePrompt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPromptManager_ListPrompts(t *testing.T) {
	logger := slog.Default()
	pm := NewPromptManager(logger)

	// Add test templates
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
			Type:         ReceiptExtractionPrompt,
			Name:         "Template 2",
			SystemPrompt: "System prompt 2",
			UserPrompt:   "User prompt 2",
			Version:      "1.0.0",
			IsActive:     false,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	for _, tpl := range templates {
		if err := pm.UpdatePrompt(context.Background(), tpl); err != nil {
			t.Fatalf("Failed to update template: %v", err)
		}
	}

	got := pm.ListPrompts(context.Background())
	if len(got) != 1 {
		t.Errorf("ListPrompts() returned %d templates, want 1 (only active)", len(got))
	}
}

func TestPromptManager_DeactivatePrompt(t *testing.T) {
	logger := slog.Default()
	pm := NewPromptManager(logger)

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
			name:       "Successfully_deactivate_existing_prompt",
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
			err := pm.DeactivatePrompt(context.Background(), tt.promptType)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeactivatePrompt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPromptManager_TestPrompt(t *testing.T) {
	logger := slog.Default()
	pm := NewPromptManager(logger)

	// Add a test template
	template := &PromptTemplate{
		Type:         TransactionCategorizationPrompt,
		Name:         "Test Template",
		SystemPrompt: "System prompt for {{.Name}}",
		UserPrompt:   "User prompt with {{.Value}}",
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
		data       interface{}
		wantErr    bool
	}{
		{
			name:       "Successfully_test_prompt",
			promptType: TransactionCategorizationPrompt,
			data: struct {
				Name  string
				Value int
			}{
				Name:  "Test",
				Value: 42,
			},
			wantErr: false,
		},
		{
			name:       "Error_test_non_existent_prompt",
			promptType: "non-existent",
			data:       nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := pm.TestPrompt(context.Background(), tt.promptType, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestPrompt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == "" {
				t.Error("TestPrompt() returned empty string, want non-empty")
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
			name:    "Successfully_increment_patch_version",
			version: "1.0.0",
			want:    "1.0.1",
		},
		{
			name:    "Successfully_increment_from_higher_version",
			version: "2.3.5",
			want:    "2.3.6",
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
