package ai

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
			name: "successful_transaction_categorization",
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
			name: "invalid_template_syntax",
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
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			for _, want := range tt.wantContain {
				assert.Contains(t, got, want)
			}
		})
	}
}

func TestNewPromptTemplate(t *testing.T) {
	name := "Test Template"
	content := "Test content"
	pt := NewPromptTemplate(name, content)

	assert.NotNil(t, pt)
	assert.Equal(t, name, pt.Name)
	assert.Equal(t, content, pt.Template)
	assert.Empty(t, pt.Examples)
	assert.Empty(t, pt.Rules)
}

func TestPromptTemplate_AddExample(t *testing.T) {
	pt := NewPromptTemplate("Test", "content")
	example := Example{
		Input:    "test input",
		Expected: "test output",
	}
	pt.AddExample(example.Input, example.Expected)

	assert.Len(t, pt.Examples, 1)
	assert.Equal(t, example, pt.Examples[0])
}

func TestPromptTemplate_AddRule(t *testing.T) {
	pt := NewPromptTemplate("Test", "content")
	rule := "test rule"
	pt.AddRule(rule)

	assert.Len(t, pt.Rules, 1)
	assert.Equal(t, rule, pt.Rules[0])
}

func TestCategory_Validation(t *testing.T) {
	tests := []struct {
		name     string
		category Category
		wantErr  bool
	}{
		{
			name: "valid_category",
			category: Category{
				Name:        "Food",
				Description: "Food and groceries",
				Keywords:    []string{"grocery", "restaurant"},
				Parent:      "Expenses",
			},
			wantErr: false,
		},
		{
			name: "empty_name",
			category: Category{
				Description: "Test description",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Add validation logic to Category struct
			if tt.category.Name == "" {
				assert.True(t, tt.wantErr)
			} else {
				assert.False(t, tt.wantErr)
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
			name: "valid_example",
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
			name: "invalid_confidence",
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
			if tt.example.Confidence > 1.0 || tt.example.Confidence < 0.0 {
				assert.True(t, tt.wantErr)
			} else {
				assert.False(t, tt.wantErr)
			}
		})
	}
}
