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
				Name: "Test Template",
				Template: `Analyze the following transaction:
Description: {{.Description}}
Amount: {{.Amount}}
Date: {{.Date}}`,
				Examples: []Example{
					{Input: "Grocery shopping", Expected: "Food"},
				},
				Rules: []string{"Consider the amount"},
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
				Name:     "Invalid Template",
				Template: "Invalid {{.Missing}",
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
		name    string
		tplName string
		content string
	}{
		{
			name:    "Successfully_create_new_prompt_template",
			tplName: "Test Template",
			content: "Test content",
		},
		{
			name:    "Successfully_create_template_with_empty_content",
			tplName: "Empty Template",
			content: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt := NewPromptTemplate(tt.tplName, tt.content)
			if pt == nil {
				t.Fatal("NewPromptTemplate() returned nil")
			}
			if pt.Name != tt.tplName {
				t.Errorf("NewPromptTemplate().Name = %v, want %v", pt.Name, tt.tplName)
			}
			if pt.Template != tt.content {
				t.Errorf("NewPromptTemplate().Template = %v, want %v", pt.Template, tt.content)
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
	}{
		{
			name:     "Successfully_add_example_to_template",
			input:    "test input",
			expected: "test output",
		},
		{
			name:     "Successfully_add_example_with_empty_values",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt := NewPromptTemplate("Test", "content")
			pt.AddExample(tt.input, tt.expected)

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
		})
	}
}

func TestPromptTemplate_AddRule(t *testing.T) {
	tests := []struct {
		name string
		rule string
	}{
		{
			name: "Successfully_add_rule_to_template",
			rule: "test rule",
		},
		{
			name: "Successfully_add_empty_rule",
			rule: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt := NewPromptTemplate("Test", "content")
			pt.AddRule(tt.rule)

			if len(pt.Rules) != 1 {
				t.Errorf("AddRule() resulted in %d rules, want 1", len(pt.Rules))
				return
			}

			if pt.Rules[0] != tt.rule {
				t.Errorf("AddRule() rule = %v, want %v", pt.Rules[0], tt.rule)
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
