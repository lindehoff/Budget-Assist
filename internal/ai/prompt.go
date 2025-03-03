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

// ExecuteTemplate executes a template with the provided data and returns the result
func ExecuteTemplate(templateText string, data interface{}) (string, error) {
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
