package ai

import (
	"strings"
	"testing"
	"time"
)

func TestPromptTemplate_Execute(t *testing.T) {
	tests := []struct {
		name        string
		template    *PromptTemplate
		data        interface{}
		wantContain []string
		wantErr     bool
	}{
		{
			name: "Successfully_execute_transaction_categorization_template",
			template: &PromptTemplate{
				Type:         TransactionCategorizationPrompt,
				Name:         "Test Template",
				SystemPrompt: "System prompt",
				UserPrompt: `Analyze the following transaction:
Description: {{.Description}}
Amount: {{.Amount}}
Date: {{.Date}}`,
				Examples: []Example{
					{
						Input:     "Grocery shopping",
						Expected:  "Food",
						CreatedAt: time.Now(),
						Score:     0.95,
					},
				},
				Rules: []Rule{
					{
						Description: "Consider the amount",
						Pattern:     "amount > 0",
						Weight:      1.0,
					},
				},
				Version:   "1.0.0",
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			data: struct {
				Description string
				Amount      string
				Date        string
			}{
				Description: "COOP Grocery Store",
				Amount:      "150.00 SEK",
				Date:        "2024-02-24T15:04:05Z",
			},
			wantContain: []string{
				"COOP Grocery Store",
				"150.00 SEK",
				"2024-02-24T15:04:05Z",
			},
			wantErr: false,
		},
		{
			name: "Error_execute_template_with_invalid_syntax",
			template: &PromptTemplate{
				Type:         TransactionCategorizationPrompt,
				Name:         "Invalid Template",
				SystemPrompt: "System prompt",
				UserPrompt:   "Invalid {{.Missing}",
				Version:      "1.0.0",
				IsActive:     true,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			data:        struct{}{},
			wantContain: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.template.Execute(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			for _, want := range tt.wantContain {
				if !strings.Contains(got, want) {
					t.Errorf("Execute() result does not contain %q", want)
				}
			}
		})
	}
}

func TestNewPromptTemplate(t *testing.T) {
	tests := []struct {
		name       string
		promptType PromptType
		tplName    string
	}{
		{
			name:       "Successfully_create_new_prompt_template",
			promptType: TransactionCategorizationPrompt,
			tplName:    "Test Template",
		},
		{
			name:       "Successfully_create_template_with_empty_name",
			promptType: DocumentExtractionPrompt,
			tplName:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt := NewPromptTemplate(tt.promptType, tt.tplName)
			if pt == nil {
				t.Fatal("NewPromptTemplate() returned nil")
			}
			if pt.Name != tt.tplName {
				t.Errorf("NewPromptTemplate().Name = %v, want %v", pt.Name, tt.tplName)
			}
			if pt.Type != tt.promptType {
				t.Errorf("NewPromptTemplate().Type = %v, want %v", pt.Type, tt.promptType)
			}
			if len(pt.Examples) != 0 {
				t.Error("NewPromptTemplate().Examples should be empty")
			}
			if len(pt.Rules) != 0 {
				t.Error("NewPromptTemplate().Rules should be empty")
			}
		})
	}
}

func TestPromptTemplate_AddExample(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		score    float64
	}{
		{
			name:     "Successfully_add_example_to_template",
			input:    "test input",
			expected: "test output",
			score:    0.95,
		},
		{
			name:     "Successfully_add_example_with_empty_values",
			input:    "",
			expected: "",
			score:    0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt := NewPromptTemplate(TransactionCategorizationPrompt, "Test")
			pt.AddExample(tt.input, tt.expected, tt.score)

			if len(pt.Examples) != 1 {
				t.Errorf("AddExample() resulted in %d examples, want 1", len(pt.Examples))
				return
			}

			example := pt.Examples[0]
			if example.Input != tt.input {
				t.Errorf("AddExample() example.Input = %v, want %v", example.Input, tt.input)
			}
			if example.Expected != tt.expected {
				t.Errorf("AddExample() example.Expected = %v, want %v", example.Expected, tt.expected)
			}
			if example.Score != tt.score {
				t.Errorf("AddExample() example.Score = %v, want %v", example.Score, tt.score)
			}
		})
	}
}

func TestPromptTemplate_AddRule(t *testing.T) {
	tests := []struct {
		name        string
		description string
		pattern     string
		weight      float64
	}{
		{
			name:        "Successfully_add_rule_to_template",
			description: "test rule",
			pattern:     "test pattern",
			weight:      1.0,
		},
		{
			name:        "Successfully_add_rule_with_empty_values",
			description: "",
			pattern:     "",
			weight:      0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt := NewPromptTemplate(TransactionCategorizationPrompt, "Test")
			pt.AddRule(tt.description, tt.pattern, tt.weight)

			if len(pt.Rules) != 1 {
				t.Errorf("AddRule() resulted in %d rules, want 1", len(pt.Rules))
				return
			}

			rule := pt.Rules[0]
			if rule.Description != tt.description {
				t.Errorf("AddRule() rule.Description = %v, want %v", rule.Description, tt.description)
			}
			if rule.Pattern != tt.pattern {
				t.Errorf("AddRule() rule.Pattern = %v, want %v", rule.Pattern, tt.pattern)
			}
			if rule.Weight != tt.weight {
				t.Errorf("AddRule() rule.Weight = %v, want %v", rule.Weight, tt.weight)
			}
		})
	}
}

func TestCategory_Validation(t *testing.T) {
	tests := []struct {
		name     string
		category Category
		wantErr  bool
	}{
		{
			name: "Successfully_validate_complete_category",
			category: Category{
				Name:        "Food",
				Description: "Food and groceries",
				Keywords:    []string{"grocery", "restaurant"},
				Parent:      "Expenses",
			},
			wantErr: false,
		},
		{
			name: "Error_validate_category_with_empty_name",
			category: Category{
				Description: "Test description",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Add validation logic to Category struct
			hasError := tt.category.Name == ""
			if hasError != tt.wantErr {
				t.Errorf("Category validation error = %v, wantErr %v", hasError, tt.wantErr)
			}
		})
	}
}

func TestTrainingExample_Validation(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		example TrainingExample
		wantErr bool
	}{
		{
			name: "Successfully_validate_complete_example",
			example: TrainingExample{
				Category: Category{
					Name:     "Food",
					Keywords: []string{"grocery"},
				},
				Input:       "Grocery shopping at COOP",
				ValidatedBy: "user123",
				CreatedAt:   now,
				Confidence:  0.95,
			},
			wantErr: false,
		},
		{
			name: "Error_validate_example_with_invalid_confidence",
			example: TrainingExample{
				Category: Category{
					Name: "Food",
				},
				Input:       "Test input",
				ValidatedBy: "user123",
				CreatedAt:   now,
				Confidence:  1.5, // Should be between 0 and 1
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Add validation logic to TrainingExample struct
			hasError := tt.example.Confidence > 1.0 || tt.example.Confidence < 0.0
			if hasError != tt.wantErr {
				t.Errorf("TrainingExample validation error = %v, wantErr %v", hasError, tt.wantErr)
			}
		})
	}
}
