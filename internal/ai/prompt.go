package ai

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/lindehoff/Budget-Assist/internal/db"
)

// PromptTemplate defines a template for generating AI prompts
type PromptTemplate struct {
	Type         db.PromptType `json:"type"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	SystemPrompt string        `json:"system_prompt"`
	UserPrompt   string        `json:"user_prompt"`
	Version      string        `json:"version"`
	IsActive     bool          `json:"is_active"`
}

// Validate checks if the prompt template is valid
func (pt *PromptTemplate) Validate() error {
	if pt.Type == "" {
		return fmt.Errorf("prompt type is required")
	}
	if pt.SystemPrompt == "" {
		return fmt.Errorf("system prompt is required")
	}
	if pt.UserPrompt == "" {
		return fmt.Errorf("user prompt is required")
	}
	if pt.Version == "" {
		return fmt.Errorf("version is required")
	}
	return nil
}

// Execute generates the final prompt string with the given data
func (pt *PromptTemplate) Execute(data any) (string, error) {
	// First validate the template
	if err := pt.Validate(); err != nil {
		return "", fmt.Errorf("invalid template: %w", err)
	}

	// Parse and execute the system prompt
	systemTmpl, err := template.New("system").Parse(pt.SystemPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to parse system prompt template: %w", err)
	}

	var systemBuf bytes.Buffer
	if err := systemTmpl.Execute(&systemBuf, data); err != nil {
		return "", fmt.Errorf("failed to execute system prompt: %w", err)
	}

	// Parse and execute the user prompt
	userTmpl, err := template.New("user").Parse(pt.UserPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to parse user prompt template: %w", err)
	}

	var userBuf bytes.Buffer
	if err := userTmpl.Execute(&userBuf, data); err != nil {
		return "", fmt.Errorf("failed to execute user prompt: %w", err)
	}

	// Combine the prompts
	return fmt.Sprintf("System: %s\n\nUser: %s", systemBuf.String(), userBuf.String()), nil
}
