package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/lindehoff/Budget-Assist/internal/db"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

type mockRoundTripper struct {
	response *http.Response
	err      error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func setupMockClient(statusCode int, body interface{}) *http.Client {
	var responseBody []byte

	// Marshal the body directly, regardless of type
	responseBody, _ = json.Marshal(body)

	response := &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}
	return &http.Client{
		Transport: &mockRoundTripper{response: response},
	}
}

func Test_OpenAIService_AnalyzeTransaction(t *testing.T) {
	tests := []struct {
		name           string
		transaction    *db.Transaction
		opts           AnalysisOptions
		mockResponse   interface{}
		expectedResult *Analysis
		expectedError  error
	}{
		{
			name: "successful analysis",
			transaction: &db.Transaction{
				Description:     "Grocery shopping at Walmart",
				Amount:          decimal.NewFromFloat(100.50),
				Currency:        "USD",
				TransactionDate: time.Now(),
			},
			opts: AnalysisOptions{
				DocumentType: "bill",
			},
			mockResponse: &ChatCompletionResponse{
				Choices: []Choice{
					{
						Message: Message{
							Content: `{"category": "Groceries", "subcategory": "Supermarket", "confidence": 0.95}`,
						},
					},
				},
			},
			expectedResult: &Analysis{
				Category:    "Groceries",
				Subcategory: "Supermarket",
				Confidence:  0.95,
			},
			expectedError: nil,
		},
		{
			name: "Analyze_error_unsupported_document_type",
			transaction: &db.Transaction{
				Description:     "Test transaction",
				Amount:          decimal.NewFromFloat(100.50),
				Currency:        "USD",
				TransactionDate: time.Now(),
			},
			opts: AnalysisOptions{
				DocumentType: "unsupported",
			},
			expectedError: fmt.Errorf("AnalyzeTransaction operation failed: unsupported document type: unsupported"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock store
			store := db.NewMockStore()

			// Create a test prompt
			prompt := &db.Prompt{
				Type:         db.BillAnalysisPrompt,
				Name:         "Test Prompt",
				SystemPrompt: "System prompt",
				UserPrompt:   "User prompt for {{.Description}}",
				Version:      "1.0",
				IsActive:     true,
			}
			if err := store.CreatePrompt(context.Background(), prompt); err != nil {
				t.Fatalf("failed to create test prompt: %v", err)
			}

			// Create the service with mocked HTTP client
			client := setupMockClient(http.StatusOK, tt.mockResponse)
			service := NewOpenAIService(Config{
				BaseURL:        "https://api.openai.com",
				APIKey:         "test-key",
				RequestTimeout: 30 * time.Second,
			}, store, slog.Default())
			service.client = client

			// Call the method
			analysis, err := service.AnalyzeTransaction(context.Background(), tt.transaction, tt.opts)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, analysis)
			}
		})
	}
}

func Test_OpenAIService_ExtractDocument(t *testing.T) {
	tests := []struct {
		name           string
		document       *Document
		mockResponse   interface{}
		expectedResult *Extraction
		expectedError  error
	}{
		{
			name: "successful extraction",
			document: &Document{
				Content: []byte("Receipt from Walmart\nDate: 2024-03-20\nAmount: $50.25\nGroceries"),
				Type:    "receipt",
			},
			mockResponse: &ChatCompletionResponse{
				Choices: []Choice{
					{
						Message: Message{
							Content: `{
								"date": "2024-03-20",
								"amount": 50.25,
								"currency": "USD",
								"description": "Receipt from Walmart",
								"category": "Groceries",
								"subcategory": "Supermarket"
							}`,
						},
					},
				},
			},
			expectedResult: &Extraction{
				Date:        "2024-03-20",
				Amount:      50.25,
				Currency:    "USD",
				Description: "Receipt from Walmart",
				Category:    "Groceries",
				Subcategory: "Supermarket",
			},
			expectedError: nil,
		},
		{
			name: "Extract_error_empty_document",
			document: &Document{
				Content: []byte{},
				Type:    "receipt",
			},
			expectedError: fmt.Errorf("ExtractDocument operation failed: empty document content"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock store
			store := db.NewMockStore()

			// Create a test prompt
			prompt := &db.Prompt{
				Type:         db.BillAnalysisPrompt,
				Name:         "Test Prompt",
				SystemPrompt: "System prompt",
				UserPrompt:   "User prompt for {{.Content}}",
				Version:      "1.0",
				IsActive:     true,
			}
			if err := store.CreatePrompt(context.Background(), prompt); err != nil {
				t.Fatalf("failed to create test prompt: %v", err)
			}

			// Create the service with mocked HTTP client
			client := setupMockClient(http.StatusOK, tt.mockResponse)
			service := NewOpenAIService(Config{
				BaseURL:        "https://api.openai.com",
				APIKey:         "test-key",
				RequestTimeout: 30 * time.Second,
			}, store, slog.Default())
			service.client = client

			// Call the method
			extraction, err := service.ExtractDocument(context.Background(), tt.document)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, extraction)
			}
		})
	}
}

func Test_OpenAIService_SuggestCategories(t *testing.T) {
	tests := []struct {
		name           string
		description    string
		mockResponse   interface{}
		expectedResult []CategoryMatch
		expectedError  error
	}{
		{
			name:        "successful suggestion",
			description: "Netflix subscription",
			mockResponse: &ChatCompletionResponse{
				Choices: []Choice{
					{
						Message: Message{
							Content: `[{"category": "Entertainment", "confidence": 0.98}]`,
						},
					},
				},
			},
			expectedResult: []CategoryMatch{
				{
					Category:   "Entertainment",
					Confidence: 0.98,
				},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock store
			store := db.NewMockStore()

			// Create a test prompt
			prompt := &db.Prompt{
				Type:         db.TransactionCategorizationPrompt,
				Name:         "Test Prompt",
				SystemPrompt: "System prompt",
				UserPrompt:   "User prompt for {{.Description}}",
				Version:      "1.0",
				IsActive:     true,
			}
			if err := store.CreatePrompt(context.Background(), prompt); err != nil {
				t.Fatalf("failed to create test prompt: %v", err)
			}

			// Create the service with mocked HTTP client
			client := setupMockClient(http.StatusOK, tt.mockResponse)
			service := NewOpenAIService(Config{
				BaseURL:        "https://api.openai.com",
				APIKey:         "test-key",
				RequestTimeout: 30 * time.Second,
			}, store, slog.Default())
			service.client = client

			// Call the method
			matches, err := service.SuggestCategories(context.Background(), tt.description)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, matches)
			}
		})
	}
}
