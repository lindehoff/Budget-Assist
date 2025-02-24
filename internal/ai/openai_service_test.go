package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
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
	mockStore := newMockStore()

	// Initialize the mock store with test prompts
	mockStore.prompts[string(TransactionAnalysisPrompt)] = &db.Prompt{
		Type:         string(TransactionAnalysisPrompt),
		Name:         "Test Transaction Prompt",
		SystemPrompt: "You are a helpful assistant",
		UserPrompt:   "Analyze this transaction",
		Examples:     `[]`,
		Rules:        `[]`,
		Version:      "1.0.0",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	mockStore.prompts[string(DocumentExtractionPrompt)] = &db.Prompt{
		Type:         string(DocumentExtractionPrompt),
		Name:         "Test Document Prompt",
		SystemPrompt: "You are a helpful assistant",
		UserPrompt:   "Extract content from this document",
		Examples:     `[]`,
		Rules:        `[]`,
		Version:      "1.0.0",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	mockStore.prompts[string(TransactionCategorizationPrompt)] = &db.Prompt{
		Type:         string(TransactionCategorizationPrompt),
		Name:         "Test Categorization Prompt",
		SystemPrompt: "You are a helpful assistant",
		UserPrompt:   "Suggest categories for this transaction",
		Examples:     `[]`,
		Rules:        `[]`,
		Version:      "1.0.0",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return NewOpenAIService(Config{
		BaseURL:        server.URL,
		APIKey:         "test-api-key",
		RequestTimeout: 5 * time.Second,
	}, mockStore, logger)
}

func TestOpenAIService_Successfully_analyze_transaction(t *testing.T) {
	ctx := context.TODO()

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
			t.Cleanup(func() {
				server.Close()
			})

			service := setupTestService(t, server)

			got, err := service.AnalyzeTransaction(ctx, tt.transaction)
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
		{
			name: "Successfully_extract_empty_content",
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
							Content: `{"content":""}`,
						},
					},
				},
			},
			document: &Document{
				Content: []byte("Empty content"),
			},
			wantContent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupMockServer(t, tt.response, http.StatusOK, map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-api-key",
			})
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
								Content: `{"score":0.95,"remarks":"Test"}`,
							},
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(response); err != nil {
					t.Fatal(err)
				}
			}))
			t.Cleanup(func() {
				server.Close()
			})

			service := setupTestService(t, server)

			ctx, cancel := context.WithTimeout(context.Background(), tt.contextTimeout)
			defer cancel()

			tx := &db.Transaction{
				Amount:          100,
				TransactionDate: time.Now(),
				Description:     "Test transaction",
				Currency:        "USD",
			}

			_, err := service.AnalyzeTransaction(ctx, tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("AnalyzeTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.wantErrType != nil {
				var opErr *OperationError
				if !errors.As(err, &opErr) {
					t.Errorf("AnalyzeTransaction() error type = %T, want %T", err, &OperationError{})
					return
				}

				// Check if any error in the chain is a context.DeadlineExceeded
				var found bool
				for e := opErr.Err; e != nil; {
					if errors.Is(e, tt.wantErrType) {
						found = true
						break
					}
					if unwrapped, ok := e.(interface{ Unwrap() error }); ok {
						e = unwrapped.Unwrap()
					} else {
						break
					}
				}

				if !found {
					t.Errorf("AnalyzeTransaction() error chain does not contain %v", tt.wantErrType)
				}
			}
		})
	}
}

func TestOpenAIService_Concurrent_requests(t *testing.T) {
	// Create a channel to track concurrent requests
	requestCount := 0
	var mu sync.Mutex
	maxConcurrent := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		if requestCount > maxConcurrent {
			maxConcurrent = requestCount
		}
		mu.Unlock()

		// Simulate processing time
		time.Sleep(50 * time.Millisecond)

		mu.Lock()
		requestCount--
		mu.Unlock()

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
						Content: `{"score":0.95,"remarks":"Test"}`,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatal(err)
		}
	}))
	t.Cleanup(func() {
		server.Close()
	})

	service := setupTestService(t, server)

	// Number of concurrent requests to make
	numRequests := 10
	var wg sync.WaitGroup
	wg.Add(numRequests)

	// Channel to collect errors
	errCh := make(chan error, numRequests)

	// Launch concurrent requests
	for i := 0; i < numRequests; i++ {
		go func(i int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			tx := &db.Transaction{
				Amount:          float64(100 + i),
				TransactionDate: time.Now(),
				Description:     fmt.Sprintf("Test transaction %d", i),
				Currency:        "USD",
			}

			_, err := service.AnalyzeTransaction(ctx, tx)
			if err != nil {
				errCh <- fmt.Errorf("request %d failed: %w", i, err)
			}
		}(i)
	}

	// Wait for all requests to complete
	wg.Wait()
	close(errCh)

	// Check for any errors
	errors := make([]error, 0, numRequests)
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		t.Errorf("Got %d errors from concurrent requests:", len(errors))
		for _, err := range errors {
			t.Errorf("  %v", err)
		}
	}

	// Verify that requests were actually concurrent
	if maxConcurrent < 2 {
		t.Errorf("Expected concurrent requests, but max concurrent requests was %d", maxConcurrent)
	}
}
