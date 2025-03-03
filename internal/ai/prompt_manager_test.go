package ai

import (
	"context"
	"log/slog"
	"testing"

	"github.com/lindehoff/Budget-Assist/internal/db"
)

func createTestPromptManager(t *testing.T) (*PromptManager, *db.MockStore) {
	t.Helper()
	store := db.NewMockStore()
	logger := slog.Default()
	return NewPromptManager(store, logger), store
}

func createTestPromptTemplate(promptType db.PromptType) *PromptTemplate {
	return &PromptTemplate{
		Type:         promptType,
		Name:         "Test Prompt",
		Description:  "Test Description",
		SystemPrompt: "You are a helpful assistant",
		UserPrompt:   "Please help with: {{.Content}}",
		Version:      "1.0.0",
		IsActive:     true,
	}
}

func Test_prompt_manager_get_prompt(t *testing.T) {
	tests := []struct {
		name       string
		setupStore func(*db.MockStore)
		promptType db.PromptType
		want       *PromptTemplate
		wantErr    string
	}{
		{
			name: "Successfully_get_existing_prompt",
			setupStore: func(store *db.MockStore) {
				template := createTestPromptTemplate(db.BillAnalysisPrompt)
				dbPrompt := &db.Prompt{
					Type:         template.Type,
					Name:         template.Name,
					Description:  template.Description,
					SystemPrompt: template.SystemPrompt,
					UserPrompt:   template.UserPrompt,
					Version:      template.Version,
					IsActive:     template.IsActive,
				}
				_ = store.CreatePrompt(context.TODO(), dbPrompt)
			},
			promptType: db.BillAnalysisPrompt,
			want:       createTestPromptTemplate(db.BillAnalysisPrompt),
		},
		{
			name:       "Get_error_nonexistent_prompt",
			setupStore: func(store *db.MockStore) {},
			promptType: db.BillAnalysisPrompt,
			wantErr:    "prompt template not found for type: bill_analysis",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm, store := createTestPromptManager(t)
			tt.setupStore(store)

			got, err := pm.GetPrompt(context.TODO(), tt.promptType)
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

			if got.Type != tt.want.Type ||
				got.Name != tt.want.Name ||
				got.Description != tt.want.Description ||
				got.SystemPrompt != tt.want.SystemPrompt ||
				got.UserPrompt != tt.want.UserPrompt ||
				got.Version != tt.want.Version ||
				got.IsActive != tt.want.IsActive {
				t.Errorf("Expected prompt %+v, got %+v", tt.want, got)
			}
		})
	}
}

func Test_prompt_manager_update_prompt(t *testing.T) {
	tests := []struct {
		name       string
		setupStore func(*db.MockStore)
		template   *PromptTemplate
		wantErr    string
	}{
		{
			name:       "Successfully_update_new_prompt",
			setupStore: func(store *db.MockStore) {},
			template:   createTestPromptTemplate(db.BillAnalysisPrompt),
		},
		{
			name: "Successfully_update_existing_prompt",
			setupStore: func(store *db.MockStore) {
				template := createTestPromptTemplate(db.BillAnalysisPrompt)
				dbPrompt := &db.Prompt{
					Type:         template.Type,
					Name:         template.Name,
					Description:  template.Description,
					SystemPrompt: template.SystemPrompt,
					UserPrompt:   template.UserPrompt,
					Version:      template.Version,
					IsActive:     template.IsActive,
				}
				_ = store.CreatePrompt(context.TODO(), dbPrompt)
			},
			template: &PromptTemplate{
				Type:         db.BillAnalysisPrompt,
				Name:         "Updated Prompt",
				Description:  "Updated Description",
				SystemPrompt: "Updated system prompt",
				UserPrompt:   "Updated user prompt",
				Version:      "1.0.1",
				IsActive:     true,
			},
		},
		{
			name:       "Update_error_nil_template",
			setupStore: func(store *db.MockStore) {},
			template:   nil,
			wantErr:    "template cannot be nil",
		},
		{
			name:       "Update_error_invalid_template",
			setupStore: func(store *db.MockStore) {},
			template: &PromptTemplate{
				Name: "Invalid Template",
				// Missing required fields
			},
			wantErr: "prompt type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm, store := createTestPromptManager(t)
			tt.setupStore(store)

			err := pm.UpdatePrompt(context.TODO(), tt.template)
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

			// Verify the prompt was updated in the store
			if tt.template != nil {
				dbPrompt, err := store.GetPromptByType(context.TODO(), string(tt.template.Type))
				if err != nil {
					t.Errorf("Failed to get prompt from store: %v", err)
					return
				}

				if dbPrompt.Type != tt.template.Type ||
					dbPrompt.Name != tt.template.Name ||
					dbPrompt.Description != tt.template.Description ||
					dbPrompt.SystemPrompt != tt.template.SystemPrompt ||
					dbPrompt.UserPrompt != tt.template.UserPrompt ||
					dbPrompt.Version != tt.template.Version ||
					dbPrompt.IsActive != tt.template.IsActive {
					t.Errorf("Expected prompt %+v, got %+v", tt.template, dbPrompt)
				}
			}
		})
	}
}

func Test_prompt_manager_list_prompts(t *testing.T) {
	tests := []struct {
		name       string
		setupStore func(*db.MockStore)
		want       []*PromptTemplate
		wantErr    string
	}{
		{
			name: "Successfully_list_multiple_prompts",
			setupStore: func(store *db.MockStore) {
				templates := []struct {
					Type db.PromptType
					Name string
				}{
					{db.BillAnalysisPrompt, "Bill Analysis"},
					{db.ReceiptAnalysisPrompt, "Receipt Analysis"},
					{db.TransactionCategorizationPrompt, "Transaction Categorization"},
				}

				for _, tmpl := range templates {
					template := createTestPromptTemplate(tmpl.Type)
					template.Name = tmpl.Name
					dbPrompt := &db.Prompt{
						Type:         template.Type,
						Name:         template.Name,
						Description:  template.Description,
						SystemPrompt: template.SystemPrompt,
						UserPrompt:   template.UserPrompt,
						Version:      template.Version,
						IsActive:     template.IsActive,
					}
					_ = store.CreatePrompt(context.TODO(), dbPrompt)
				}
			},
			want: []*PromptTemplate{
				func() *PromptTemplate {
					t := createTestPromptTemplate(db.BillAnalysisPrompt)
					t.Name = "Bill Analysis"
					return t
				}(),
				func() *PromptTemplate {
					t := createTestPromptTemplate(db.ReceiptAnalysisPrompt)
					t.Name = "Receipt Analysis"
					return t
				}(),
				func() *PromptTemplate {
					t := createTestPromptTemplate(db.TransactionCategorizationPrompt)
					t.Name = "Transaction Categorization"
					return t
				}(),
			},
		},
		{
			name:       "Successfully_list_empty_prompts",
			setupStore: func(store *db.MockStore) {},
			want:       []*PromptTemplate{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm, store := createTestPromptManager(t)
			tt.setupStore(store)

			got, err := pm.ListPrompts(context.TODO())
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

			if len(got) != len(tt.want) {
				t.Errorf("Expected %d prompts, got %d", len(tt.want), len(got))
				return
			}

			for i, want := range tt.want {
				if got[i].Type != want.Type ||
					got[i].Name != want.Name ||
					got[i].Description != want.Description ||
					got[i].SystemPrompt != want.SystemPrompt ||
					got[i].UserPrompt != want.UserPrompt ||
					got[i].Version != want.Version ||
					got[i].IsActive != want.IsActive {
					t.Errorf("Prompt at index %d:\nExpected: %+v\nGot: %+v", i, want, got[i])
				}
			}
		})
	}
}
