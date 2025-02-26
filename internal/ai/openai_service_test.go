package ai

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	db "github.com/lindehoff/Budget-Assist/internal/db"
	"github.com/shopspring/decimal"
)

func setupMockServer(t *testing.T, response OpenAIResponse, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-api-key" {
			t.Errorf("Authorization header = %q, want Bearer test-api-key", got)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatal(err)
		}
	}))
}

func setupTestService(t *testing.T, server *httptest.Server) *OpenAIService {
	store := NewMockStore()

	// Create prompt templates
	prompts := []struct {
		promptType PromptType
		version    string
	}{
		{TransactionAnalysisPrompt, "1.0.0"},
		{DocumentExtractionPrompt, "1.0.0"},
		{TransactionCategorizationPrompt, "1.0.0"},
	}

	for _, p := range prompts {
		prompt := &db.Prompt{
			Type:         string(p.promptType),
			Version:      p.version,
			SystemPrompt: "You are a helpful assistant.",
			UserPrompt:   "Please help me with this task.",
			IsActive:     true,
		}
		if err := store.CreatePrompt(context.Background(), prompt); err != nil {
			t.Fatalf("failed to create prompt template: %v", err)
		}
	}

	logger := slog.Default()
	config := Config{
		BaseURL:        server.URL,
		APIKey:         "test-api-key",
		RequestTimeout: 5,
	}
	return NewOpenAIService(config, store, logger)
}

func TestOpenAIService_AnalyzeTransaction(t *testing.T) {
	ctx := context.TODO()

	type testCase struct {
		name        string
		transaction *db.Transaction
		options     AnalysisOptions
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
							Content: `{"remarks":"Food purchase","score":0.9}`,
						},
					},
				},
			},
			transaction: &db.Transaction{
				Amount:          decimal.RequireFromString("100.50"),
				TransactionDate: time.Date(2024, 2, 24, 12, 0, 0, 0, time.UTC),
				Description:     "COOP Grocery Store",
				Currency:        db.CurrencySEK,
			},
			options: AnalysisOptions{
				DocumentType:     "receipt",
				TransactionHints: "Regular grocery shopping",
				CategoryHints:    "Food and household items",
			},
			wantScore:   0.9,
			wantRemarks: "Food purchase",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupMockServer(t, tt.response, http.StatusOK)
			t.Cleanup(func() {
				server.Close()
			})

			service := setupTestService(t, server)

			got, err := service.AnalyzeTransaction(ctx, tt.transaction, tt.options)
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

func TestOpenAIService_ExtractDocument(t *testing.T) {
	ctx := context.TODO()

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
			server := setupMockServer(t, tt.response, http.StatusOK)
			t.Cleanup(func() {
				server.Close()
			})

			service := setupTestService(t, server)

			got, err := service.ExtractDocument(ctx, tt.document)
			if err != nil {
				t.Fatalf("ExtractDocument(%q) error = %v", string(tt.document.Content), err)
				return
			}

			if got.Content != tt.wantContent {
				t.Errorf("ExtractDocument(%q) content:\ngot:  %q\nwant: %q",
					string(tt.document.Content), got.Content, tt.wantContent)
				return
			}
		})
	}
}

func TestOpenAIService_SuggestCategories(t *testing.T) {
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
			want: []CategoryMatch{
				{Category: "Food", Confidence: 0.9},
				{Category: "Groceries", Confidence: 0.8},
			},
			description: "Grocery shopping at ICA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupMockServer(t, tt.response, http.StatusOK)
			defer server.Close()

			service := setupTestService(t, server)

			got, err := service.SuggestCategories(context.TODO(), tt.description)
			if err != nil {
				t.Fatalf("SuggestCategories() unexpected error: %v", err)
			}

			if len(got) != len(tt.want) {
				t.Fatalf("SuggestCategories() returned %d suggestions, want %d", len(got), len(tt.want))
			}

			for i, want := range tt.want {
				if got[i].Category != want.Category {
					t.Errorf("SuggestCategories() category[%d] = %v, want %v", i, got[i].Category, want.Category)
				}
				if got[i].Confidence != want.Confidence {
					t.Errorf("SuggestCategories() confidence[%d] = %v, want %v", i, got[i].Confidence, want.Confidence)
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
			server := setupMockServer(t, tt.response, tt.statusCode)
			defer server.Close()

			service := setupTestService(t, server)

			opts := AnalysisOptions{
				DocumentType: "receipt",
			}

			_, err := service.AnalyzeTransaction(context.TODO(), &db.Transaction{
				Description: "Test transaction",
				Amount:      decimal.RequireFromString("100.00"),
				Currency:    "SEK",
			}, opts)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("error = %v, want %v", err, tt.wantErrMsg)
			}
		})
	}
}

func TestOpenAIService_Error_rate_limit_with_retry(t *testing.T) {
	type response struct {
		statusCode int
		response   interface{}
	}

	responses := []response{
		{
			statusCode: http.StatusTooManyRequests,
			response:   nil,
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
							Content: `{"score":0.8,"remarks":"This appears to be a food-related expense."}`,
						},
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("expected Authorization header with Bearer test-key")
		}

		response := responses[0]
		responses = responses[1:]

		w.WriteHeader(response.statusCode)
		if response.response != nil {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(response.response); err != nil {
				t.Fatal(err)
			}
		}
	}))
	defer server.Close()

	service := setupTestService(t, server)
	service.config.APIKey = "test-key"

	opts := AnalysisOptions{
		DocumentType: "receipt",
	}

	_, err := service.AnalyzeTransaction(context.TODO(), &db.Transaction{
		Description: "Test transaction",
		Amount:      decimal.RequireFromString("100.00"),
		Currency:    "SEK",
	}, opts)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var opErr *OperationError
	if !errors.As(err, &opErr) {
		t.Errorf("error = %v, want OperationError", err)
		return
	}

	var rateLimitErr *RateLimitError
	if !errors.As(opErr.Err, &rateLimitErr) {
		t.Errorf("inner error = %v, want RateLimitError", opErr.Err)
		return
	}

	if rateLimitErr.StatusCode != http.StatusTooManyRequests {
		t.Errorf("status code = %d, want %d", rateLimitErr.StatusCode, http.StatusTooManyRequests)
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

func TestOpenAIService_Context_timeout(t *testing.T) {
	type testCase struct {
		name           string
		contextTimeout time.Duration
		serverDelay    time.Duration
		wantErr        bool
		wantErrType    error
	}

	tests := []testCase{
		{
			name:           "Error_context_timeout_exceeded",
			contextTimeout: 50 * time.Millisecond,
			serverDelay:    100 * time.Millisecond,
			wantErr:        true,
			wantErrType:    context.DeadlineExceeded,
		},
		{
			name:           "Successfully_complete_before_timeout",
			contextTimeout: 100 * time.Millisecond,
			serverDelay:    50 * time.Millisecond,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(tt.serverDelay)
				response := OpenAIResponse{
					Choices: []struct {
						Message struct {
							Content string `json:"content"`
						} `json:"message"`
					}{
						{
							Message: struct {
								Content string `json:"content"`
							}{
								Content: `{"score":0.8,"remarks":"This appears to be a food-related expense."}`,
							},
						},
					},
				}
				if err := json.NewEncoder(w).Encode(response); err != nil {
					t.Fatal(err)
				}
			}))
			defer server.Close()

			service := setupTestService(t, server)

			ctx, cancel := context.WithTimeout(context.Background(), tt.contextTimeout)
			defer cancel()

			_, err := service.AnalyzeTransaction(ctx, &db.Transaction{
				Description: "Test transaction",
				Amount:      decimal.RequireFromString("100.00"),
				Currency:    "SEK",
			}, AnalysisOptions{
				DocumentType: "receipt",
			})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErrType) {
					t.Errorf("error = %v, want %v", err, tt.wantErrType)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
