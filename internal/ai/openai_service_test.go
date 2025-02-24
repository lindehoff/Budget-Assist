package ai

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	db "github.com/lindehoff/Budget-Assist/internal/db"
)

func setupMockServer(t *testing.T, response OpenAIResponse, statusCode int, wantHeaders map[string]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for key, want := range wantHeaders {
			if got := r.Header.Get(key); got != want {
				t.Errorf("header %q = %q, want %q", key, got, want)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatal(err)
		}
	}))
}

func setupTestService(t *testing.T, server *httptest.Server) Service {
	t.Helper()
	logger := slog.Default()
	service := NewOpenAIService(Config{
		BaseURL:        server.URL,
		APIKey:         "test-api-key",
		RequestTimeout: 5 * time.Second,
	}, logger)

	// Initialize test templates
	s := service.(*OpenAIService)

	// Template for transaction analysis
	analysisTemplate := &PromptTemplate{
		Type:         TransactionAnalysisPrompt,
		Name:         "Transaction Analysis",
		SystemPrompt: "You are a financial transaction analyzer. Analyze transactions and provide categories with confidence scores.",
		UserPrompt:   "Analyze the following transaction:\nDescription: {{.Description}}\nAmount: {{.Amount}}\nDate: {{.Date}}",
		Version:      "1.0.0",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := s.promptMgr.UpdatePrompt(context.Background(), analysisTemplate); err != nil {
		t.Fatalf("Failed to update analysis template: %v", err)
	}

	// Template for category suggestions
	suggestionTemplate := &PromptTemplate{
		Type:         CategorySuggestionPrompt,
		Name:         "Category Suggestions",
		SystemPrompt: "You are a financial transaction analyzer. Suggest categories for transactions based on their descriptions.",
		UserPrompt:   "Suggest categories for the following transaction:\nDescription: {{.Description}}\nCategories:\n{{range .Categories}}- {{.Name}}: {{.Description}}\n{{end}}",
		Version:      "1.0.0",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := s.promptMgr.UpdatePrompt(context.Background(), suggestionTemplate); err != nil {
		t.Fatalf("Failed to update suggestion template: %v", err)
	}

	// Template for document extraction
	docTemplate := &PromptTemplate{
		Type:         DocumentExtractionPrompt,
		Name:         "Document Extraction",
		SystemPrompt: "You are a document information extractor. Extract and structure financial information from documents.",
		UserPrompt:   "Extract key information from the following document:\n{{.Content}}",
		Version:      "1.0.0",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := s.promptMgr.UpdatePrompt(context.Background(), docTemplate); err != nil {
		t.Fatalf("Failed to update document template: %v", err)
	}

	return service
}

func TestOpenAIService_Successfully_analyze_transaction(t *testing.T) {
	type testCase struct {
		name        string
		transaction *db.Transaction
		response    OpenAIResponse
		wantScore   float64
		wantRemarks string
	}

	tests := []testCase{
		{
			name: "Successfully_analyze_food_category",
			response: OpenAIResponse{
				Choices: []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				}{
					{
						Message: struct {
							Content string `json:"content"`
						}{
							Content: `{"score":0.95,"remarks":"Food category"}`,
						},
					},
				},
			},
			transaction: &db.Transaction{
				Amount:          100.50,
				TransactionDate: time.Date(2024, 2, 24, 12, 0, 0, 0, time.UTC),
				Description:     "COOP Grocery Store",
				Currency:        "SEK",
			},
			wantScore:   0.95,
			wantRemarks: "Food category",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupMockServer(t, tt.response, http.StatusOK, map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-api-key",
			})
			defer server.Close()

			service := setupTestService(t, server)

			got, err := service.AnalyzeTransaction(context.TODO(), tt.transaction)
			if err != nil {
				t.Fatalf("AnalyzeTransaction() error = %v", err)
				return
			}

			if got.Score != tt.wantScore {
				t.Errorf("Score = %v, want %v", got.Score, tt.wantScore)
				return
			}
			if got.Remarks != tt.wantRemarks {
				t.Errorf("Remarks = %v, want %v", got.Remarks, tt.wantRemarks)
				return
			}
		})
	}
}

func TestOpenAIService_Successfully_extract_document(t *testing.T) {
	type testCase struct {
		name        string
		document    *Document
		response    OpenAIResponse
		wantContent string
	}

	tests := []testCase{
		{
			name: "Successfully_extract_pdf_content",
			response: OpenAIResponse{
				Choices: []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				}{
					{
						Message: struct {
							Content string `json:"content"`
						}{
							Content: `{"content":"Sample document content"}`,
						},
					},
				},
			},
			document: &Document{
				Content: []byte("Sample PDF content"),
			},
			wantContent: "Sample document content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupMockServer(t, tt.response, http.StatusOK, map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-api-key",
			})
			defer server.Close()

			service := setupTestService(t, server)

			got, err := service.ExtractDocument(context.TODO(), tt.document)
			if err != nil {
				t.Fatalf("ExtractDocument() error = %v", err)
				return
			}

			if got.Content != tt.wantContent {
				t.Errorf("Content = %v, want %v", got.Content, tt.wantContent)
				return
			}
		})
	}
}

func TestOpenAIService_Successfully_suggest_categories(t *testing.T) {
	type testCase struct {
		name        string
		response    OpenAIResponse
		want        []CategoryMatch
		description string
	}

	tests := []testCase{
		{
			name: "Successfully_suggest_multiple_categories",
			response: OpenAIResponse{
				Choices: []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				}{
					{
						Message: struct {
							Content string `json:"content"`
						}{
							Content: `[{"category":"Food","confidence":0.9},{"category":"Groceries","confidence":0.8}]`,
						},
					},
				},
			},
			description: "COOP Grocery Store",
			want: []CategoryMatch{
				{Category: "Food", Confidence: 0.9},
				{Category: "Groceries", Confidence: 0.8},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupMockServer(t, tt.response, http.StatusOK, map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-api-key",
			})
			defer server.Close()

			service := setupTestService(t, server)

			got, err := service.SuggestCategories(context.TODO(), tt.description)
			if err != nil {
				t.Fatalf("SuggestCategories() error = %v", err)
				return
			}

			if len(got) != len(tt.want) {
				t.Fatalf("got %d categories, want %d", len(got), len(tt.want))
				return
			}

			for i, want := range tt.want {
				if got[i].Category != want.Category {
					t.Errorf("Category[%d] = %v, want %v", i, got[i].Category, want.Category)
					return
				}
				if got[i].Confidence != want.Confidence {
					t.Errorf("Confidence[%d] = %v, want %v", i, got[i].Confidence, want.Confidence)
					return
				}
			}
		})
	}
}

func TestOpenAIService_Error_invalid_response(t *testing.T) {
	type testCase struct {
		name       string
		response   OpenAIResponse
		statusCode int
		wantErrMsg string
	}

	tests := []testCase{
		{
			name: "Error_invalid_json_content",
			response: OpenAIResponse{
				Choices: []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				}{
					{
						Message: struct {
							Content string `json:"content"`
						}{
							Content: `invalid json`,
						},
					},
				},
			},
			statusCode: http.StatusOK,
			wantErrMsg: "failed to decode response content",
		},
		{
			name: "Error_empty_content",
			response: OpenAIResponse{
				Choices: []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				}{
					{
						Message: struct {
							Content string `json:"content"`
						}{
							Content: "",
						},
					},
				},
			},
			statusCode: http.StatusOK,
			wantErrMsg: "empty content in OpenAI response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupMockServer(t, tt.response, tt.statusCode, map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-api-key",
			})
			defer server.Close()

			service := setupTestService(t, server)

			tx := &db.Transaction{
				Amount:          100.50,
				TransactionDate: time.Now(),
				Description:     "COOP Grocery Store",
				Currency:        "SEK",
			}

			_, err := service.AnalyzeTransaction(context.TODO(), tx)
			if err == nil {
				t.Fatal("expected error, got nil")
				return
			}

			if !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("error = %v, want to contain %v", err, tt.wantErrMsg)
				return
			}
		})
	}
}

func TestOpenAIService_Error_rate_limit_with_retry(t *testing.T) {
	type response struct {
		statusCode int
		response   OpenAIResponse
	}

	responses := []response{
		{
			statusCode: http.StatusTooManyRequests,
			response: OpenAIResponse{
				Choices: []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				}{
					{
						Message: struct {
							Content string `json:"content"`
						}{
							Content: `{"error":"rate limit exceeded"}`,
						},
					},
				},
			},
		},
		{
			statusCode: http.StatusOK,
			response: OpenAIResponse{
				Choices: []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				}{
					{
						Message: struct {
							Content string `json:"content"`
						}{
							Content: `[{"category":"Food","confidence":0.9}]`,
						},
					},
				},
			},
		},
	}

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if callCount >= len(responses) {
			t.Fatal("too many requests")
			return
		}

		resp := responses[callCount]
		w.WriteHeader(resp.statusCode)
		if err := json.NewEncoder(w).Encode(resp.response); err != nil {
			t.Fatal(err)
			return
		}
		callCount++
	}))
	defer server.Close()

	service := setupTestService(t, server)

	categories, err := service.SuggestCategories(context.TODO(), "test")
	if err != nil {
		t.Fatalf("SuggestCategories() error = %v", err)
		return
	}

	if len(categories) != 1 {
		t.Errorf("got %d categories, want 1", len(categories))
		return
	}

	if callCount != 2 {
		t.Errorf("callCount = %d, want 2 (should have retried once)", callCount)
		return
	}
}

func TestOpenAIService_Error_empty_document(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("server should not be called for empty document")
	}))
	defer server.Close()

	service := setupTestService(t, server)

	doc := &Document{
		Content: []byte{},
	}

	_, err := service.ExtractDocument(context.TODO(), doc)
	if err == nil {
		t.Fatal("expected error for empty document, got nil")
		return
	}

	wantErr := "empty document content"
	if !strings.Contains(err.Error(), wantErr) {
		t.Errorf("error = %v, want to contain %v", err, wantErr)
		return
	}
}
