package ai

import (
	"bytes"
	"text/template"
	"time"
)

// Example represents a training example for prompt templates
type Example struct {
	Input    string
	Expected string
}

// PromptTemplate defines a template for generating AI prompts
type PromptTemplate struct {
	Name     string
	Template string
	Examples []Example
	Rules    []string
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

// promptTemplates contains our predefined templates for different AI tasks
var promptTemplates = map[string]PromptTemplate{
	"categorize": {
		Name: "Transaction Categorization",
		Template: `Analyze the following transaction and categorize it:
Description: {{.Description}}
Amount: {{.Amount}}
Date: {{.Date}}

Consider these rules:
{{range .Rules}}
- {{.}}{{end}}

Examples of similar transactions:
{{range .Examples}}
Input: {{.Input}}
Category: {{.Expected}}
{{end}}

Provide the most likely category and a confidence score (0-1).`,
		Rules: []string{
			"Focus on the description's key terms",
			"Consider the transaction amount for context",
			"Use parent categories when unsure of specifics",
			"Maintain consistency with similar transactions",
		},
	},
	"extract": {
		Name: "Document Information Extraction",
		Template: `Extract key information from the following document:
{{.Content}}

Focus on:
- Transaction dates
- Amount values
- Account numbers
- Transaction types
- Descriptions

Format the output as structured data.`,
		Rules: []string{
			"Extract all date formats consistently",
			"Identify and standardize amount formats",
			"Preserve original transaction descriptions",
			"Flag any uncertain extractions",
		},
	},
}

// NewPromptTemplate creates a new prompt template with the given name and content
func NewPromptTemplate(name, content string) *PromptTemplate {
	return &PromptTemplate{
		Name:     name,
		Template: content,
		Examples: make([]Example, 0),
		Rules:    make([]string, 0),
	}
}

// AddExample adds a new example to the prompt template
func (pt *PromptTemplate) AddExample(input, expected string) {
	pt.Examples = append(pt.Examples, Example{
		Input:    input,
		Expected: expected,
	})
}

// AddRule adds a new rule to the prompt template
func (pt *PromptTemplate) AddRule(rule string) {
	pt.Rules = append(pt.Rules, rule)
}

// Execute generates the final prompt string with the given data
func (pt *PromptTemplate) Execute(data any) (string, error) {
	tmpl, err := template.New(pt.Name).Parse(pt.Template)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
