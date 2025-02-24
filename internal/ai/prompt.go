package ai

import (
	"bytes"
	"fmt"
	"text/template"
	"time"
)

// PromptType represents different types of prompts
type PromptType string

const (
	TransactionCategorizationPrompt PromptType = "transaction_categorization"
	TransactionAnalysisPrompt       PromptType = "transaction_analysis"
	CategorySuggestionPrompt        PromptType = "category_suggestion"
	InvoiceExtractionPrompt         PromptType = "invoice_extraction"
	ReceiptExtractionPrompt         PromptType = "receipt_extraction"
	StatementExtractionPrompt       PromptType = "statement_extraction"
	DocumentExtractionPrompt        PromptType = "document_extraction"
)

// Example represents a training example for prompt templates
type Example struct {
	Input     string    `json:"input"`
	Expected  string    `json:"expected"`
	CreatedAt time.Time `json:"created_at"`
	Score     float64   `json:"score,omitempty"`
}

// Rule represents a prompt rule
type Rule struct {
	Description string  `json:"description"`
	Pattern     string  `json:"pattern,omitempty"`
	Weight      float64 `json:"weight,omitempty"`
}

// PromptTemplate defines a template for generating AI prompts
type PromptTemplate struct {
	Type         PromptType `json:"type"`
	Name         string     `json:"name"`
	SystemPrompt string     `json:"system_prompt"`
	UserPrompt   string     `json:"user_prompt"`
	Examples     []Example  `json:"examples"`
	Categories   []Category `json:"categories"`
	Rules        []Rule     `json:"rules"`
	Version      string     `json:"version"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	IsActive     bool       `json:"is_active"`
}

// Category represents a transaction category with metadata
//
//nolint:govet // Current layout is logical and maintainable
type Category struct {
	Keywords    []string // 24 bytes (slice)
	Parent      string   // 16 bytes
	Name        string   // 16 bytes
	Description string   // 16 bytes
}

// TrainingExample represents a validated categorization example
type TrainingExample struct {
	Category    Category  // struct
	CreatedAt   time.Time // 24 bytes
	Input       string    // 16 bytes
	ValidatedBy string    // 16 bytes
	Confidence  float64   // 8 bytes
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

// NewPromptTemplate creates a new prompt template with the given type and name
func NewPromptTemplate(promptType PromptType, name string) *PromptTemplate {
	now := time.Now()
	return &PromptTemplate{
		Type:      promptType,
		Name:      name,
		Version:   "1.0.0",
		CreatedAt: now,
		UpdatedAt: now,
		IsActive:  true,
		Examples:  make([]Example, 0),
		Rules:     make([]Rule, 0),
	}
}

// AddExample adds a new example to the prompt template
func (pt *PromptTemplate) AddExample(input, expected string, score float64) {
	pt.Examples = append(pt.Examples, Example{
		Input:     input,
		Expected:  expected,
		CreatedAt: time.Now(),
		Score:     score,
	})
	pt.UpdatedAt = time.Now()
}

// AddRule adds a new rule to the prompt template
func (pt *PromptTemplate) AddRule(description, pattern string, weight float64) {
	pt.Rules = append(pt.Rules, Rule{
		Description: description,
		Pattern:     pattern,
		Weight:      weight,
	})
	pt.UpdatedAt = time.Now()
}

// Execute generates the final prompt string with the given data
func (pt *PromptTemplate) Execute(data any) (string, error) {
	// First validate the template
	if err := pt.Validate(); err != nil {
		return "", fmt.Errorf("invalid template: %w", err)
	}

	// Create a combined template with system and user prompts
	combinedTemplate := fmt.Sprintf("%s\n\n%s", pt.SystemPrompt, pt.UserPrompt)
	tmpl, err := template.New(pt.Name).Parse(combinedTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
