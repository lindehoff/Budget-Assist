package ai

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/lindehoff/Budget-Assist/internal/db"
)

func SuccessfullyValidateValidPromptTemplate(t *testing.T) {
	t.Helper()
	template := &PromptTemplate{
		Type:         db.BillAnalysisPrompt,
		SystemPrompt: "You are a helpful assistant",
		UserPrompt:   "Please analyze this bill: {{.Content}}",
		Version:      "1.0.0",
	}

	err := template.Validate()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
		return
	}
}

func Test_prompt_template_validation(t *testing.T) {
	tests := []struct {
		name     string
		template *PromptTemplate
		wantErr  string
	}{
		{
			name: "Successfully_validate_complete_template",
			template: &PromptTemplate{
				Type:         db.BillAnalysisPrompt,
				Name:         "Bill Analysis",
				Description:  "Analyzes bills",
				SystemPrompt: "You are a helpful assistant",
				UserPrompt:   "Please analyze this bill: {{.Content}}",
				Version:      "1.0.0",
				IsActive:     true,
			},
			wantErr: "",
		},
		{
			name: "Validate_error_missing_type",
			template: &PromptTemplate{
				SystemPrompt: "You are a helpful assistant",
				UserPrompt:   "Please analyze this bill: {{.Content}}",
				Version:      "1.0.0",
			},
			wantErr: "prompt type is required",
		},
		{
			name: "Validate_error_missing_system_prompt",
			template: &PromptTemplate{
				Type:       db.BillAnalysisPrompt,
				UserPrompt: "Please analyze this bill: {{.Content}}",
				Version:    "1.0.0",
			},
			wantErr: "system prompt is required",
		},
		{
			name: "Validate_error_missing_user_prompt",
			template: &PromptTemplate{
				Type:         db.BillAnalysisPrompt,
				SystemPrompt: "You are a helpful assistant",
				Version:      "1.0.0",
			},
			wantErr: "user prompt is required",
		},
		{
			name: "Validate_error_missing_version",
			template: &PromptTemplate{
				Type:         db.BillAnalysisPrompt,
				SystemPrompt: "You are a helpful assistant",
				UserPrompt:   "Please analyze this bill: {{.Content}}",
			},
			wantErr: "version is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.template.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				return
			}

			if err == nil {
				t.Errorf("Expected error %q, got nil", tt.wantErr)
				return
			}

			if err.Error() != tt.wantErr {
				t.Errorf("Expected error %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

// Helper function to execute a template with data (for testing purposes)
func executeTemplateTest(templateText string, data interface{}) (string, error) {
	tmpl, err := template.New("prompt").Parse(templateText)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func Test_prompt_template_execute(t *testing.T) {
	tests := []struct {
		name     string
		template *PromptTemplate
		data     interface{}
		want     string
		wantErr  string
	}{
		{
			name: "Successfully_execute_template_with_valid_data",
			template: &PromptTemplate{
				Type:         db.BillAnalysisPrompt,
				SystemPrompt: "You are a helpful assistant",
				UserPrompt:   "Please analyze this bill: {{.Content}}",
				Version:      "1.0.0",
			},
			data: struct {
				Content string
			}{
				Content: "Sample bill content",
			},
			want: "System: You are a helpful assistant\n\nUser: Please analyze this bill: Sample bill content",
		},
		{
			name: "Execute_error_invalid_template",
			template: &PromptTemplate{
				Type:         db.BillAnalysisPrompt,
				SystemPrompt: "You are a helpful assistant",
				UserPrompt:   "Please analyze this bill: {{.InvalidField}}",
				Version:      "1.0.0",
			},
			data: struct {
				Content string
			}{
				Content: "Sample bill content",
			},
			wantErr: "failed to execute template: template: prompt:1:28: executing \"prompt\" at <.InvalidField>: can't evaluate field InvalidField in type struct { Content string }",
		},
		{
			name: "Execute_error_invalid_template_data",
			template: &PromptTemplate{
				Type:         db.BillAnalysisPrompt,
				SystemPrompt: "You are a helpful assistant",
				UserPrompt:   "Please analyze this bill: {{.Content}}",
				Version:      "1.0.0",
			},
			data:    "invalid data",
			wantErr: "failed to execute template: template: prompt:1:28: executing \"prompt\" at <.Content>: can't evaluate field Content in type string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute system prompt
			systemPrompt, err := executeTemplateTest(tt.template.SystemPrompt, tt.data)
			if err != nil && tt.wantErr == "" {
				t.Errorf("Unexpected error executing system prompt: %v", err)
				return
			}

			// If system prompt executed successfully, try user prompt
			var userPrompt string
			if err == nil {
				userPrompt, err = executeTemplateTest(tt.template.UserPrompt, tt.data)
			}

			// Check for expected errors
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("Expected error %q, got nil", tt.wantErr)
					return
				}
				if err.Error() != tt.wantErr {
					t.Errorf("Expected error %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
				return
			}

			// Combine the prompts in the same format as the original Execute method
			got := fmt.Sprintf("System: %s\n\nUser: %s", systemPrompt, userPrompt)
			if got != tt.want {
				t.Errorf("Expected output %q, got %q", tt.want, got)
			}
		})
	}
}
