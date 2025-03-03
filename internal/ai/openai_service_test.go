package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/lindehoff/Budget-Assist/internal/db"
	"github.com/shopspring/decimal"
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
			name: "Successfully_analyze_transaction",
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
				if err == nil {
					t.Errorf("OpenAIService.AnalyzeTransaction() error = nil, want error")
					return
				}
				if err.Error() != tt.expectedError.Error() {
					t.Errorf("OpenAIService.AnalyzeTransaction() error = %v, want %v", err, tt.expectedError)
					return
				}
			} else {
				if err != nil {
					t.Errorf("OpenAIService.AnalyzeTransaction() error = %v, want nil", err)
					return
				}
				if analysis == nil {
					t.Errorf("OpenAIService.AnalyzeTransaction() analysis = nil, want non-nil")
					return
				}
				if analysis.Category != tt.expectedResult.Category {
					t.Errorf("OpenAIService.AnalyzeTransaction() analysis.Category = %v, want %v",
						analysis.Category, tt.expectedResult.Category)
				}
				if analysis.Subcategory != tt.expectedResult.Subcategory {
					t.Errorf("OpenAIService.AnalyzeTransaction() analysis.Subcategory = %v, want %v",
						analysis.Subcategory, tt.expectedResult.Subcategory)
				}
				if analysis.Confidence != tt.expectedResult.Confidence {
					t.Errorf("OpenAIService.AnalyzeTransaction() analysis.Confidence = %v, want %v",
						analysis.Confidence, tt.expectedResult.Confidence)
				}
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
			name: "Successfully_extract_document",
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
				if err == nil {
					t.Errorf("OpenAIService.ExtractDocument() error = nil, want error")
					return
				}
				if err.Error() != tt.expectedError.Error() {
					t.Errorf("OpenAIService.ExtractDocument() error = %v, want %v", err, tt.expectedError)
					return
				}
			} else {
				if err != nil {
					t.Errorf("OpenAIService.ExtractDocument() error = %v, want nil", err)
					return
				}
				if extraction == nil {
					t.Errorf("OpenAIService.ExtractDocument() extraction = nil, want non-nil")
					return
				}
				if extraction.Date != tt.expectedResult.Date {
					t.Errorf("OpenAIService.ExtractDocument() extraction.Date = %v, want %v",
						extraction.Date, tt.expectedResult.Date)
				}
				if extraction.Amount != tt.expectedResult.Amount {
					t.Errorf("OpenAIService.ExtractDocument() extraction.Amount = %v, want %v",
						extraction.Amount, tt.expectedResult.Amount)
				}
				if extraction.Currency != tt.expectedResult.Currency {
					t.Errorf("OpenAIService.ExtractDocument() extraction.Currency = %v, want %v",
						extraction.Currency, tt.expectedResult.Currency)
				}
				if extraction.Description != tt.expectedResult.Description {
					t.Errorf("OpenAIService.ExtractDocument() extraction.Description = %v, want %v",
						extraction.Description, tt.expectedResult.Description)
				}
				if extraction.Category != tt.expectedResult.Category {
					t.Errorf("OpenAIService.ExtractDocument() extraction.Category = %v, want %v",
						extraction.Category, tt.expectedResult.Category)
				}
				if extraction.Subcategory != tt.expectedResult.Subcategory {
					t.Errorf("OpenAIService.ExtractDocument() extraction.Subcategory = %v, want %v",
						extraction.Subcategory, tt.expectedResult.Subcategory)
				}
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
			name:        "Successfully_suggest_categories",
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
				if err == nil {
					t.Errorf("OpenAIService.SuggestCategories() error = nil, want error")
					return
				}
				if err.Error() != tt.expectedError.Error() {
					t.Errorf("OpenAIService.SuggestCategories() error = %v, want %v", err, tt.expectedError)
					return
				}
			} else {
				if err != nil {
					t.Errorf("OpenAIService.SuggestCategories() error = %v, want nil", err)
					return
				}
				if len(matches) != len(tt.expectedResult) {
					t.Errorf("OpenAIService.SuggestCategories() matches length = %d, want %d",
						len(matches), len(tt.expectedResult))
					return
				}
				for i, match := range matches {
					if match.Category != tt.expectedResult[i].Category {
						t.Errorf("OpenAIService.SuggestCategories() matches[%d].Category = %v, want %v",
							i, match.Category, tt.expectedResult[i].Category)
					}
					if match.Confidence != tt.expectedResult[i].Confidence {
						t.Errorf("OpenAIService.SuggestCategories() matches[%d].Confidence = %v, want %v",
							i, match.Confidence, tt.expectedResult[i].Confidence)
					}
				}
			}
		})
	}
}

func Test_OpenAIService_handleErrorResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       []byte
		wantErr    error
	}{
		{
			name:       "HandleErrorResponse_error_rate_limit",
			statusCode: http.StatusTooManyRequests,
			body:       []byte("rate limit exceeded"),
			wantErr:    &RateLimitError{Message: "rate limit exceeded", StatusCode: http.StatusTooManyRequests},
		},
		{
			name:       "HandleErrorResponse_error_api_error",
			statusCode: http.StatusBadRequest,
			body:       []byte("invalid request"),
			wantErr:    &OpenAIError{Operation: "API request", Message: "invalid request", StatusCode: http.StatusBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the service
			service := NewOpenAIService(Config{
				BaseURL:        "https://api.openai.com",
				APIKey:         "test-key",
				RequestTimeout: 30 * time.Second,
			}, nil, slog.Default())

			// Create a mock response
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(bytes.NewReader(tt.body)),
			}

			// Call the method
			err := service.handleErrorResponse(resp, tt.body)
			if err == nil {
				t.Errorf("OpenAIService.handleErrorResponse() error = nil, want error")
				return
			}

			// Check the error type
			switch tt.statusCode {
			case http.StatusTooManyRequests:
				rateLimitErr, ok := err.(*RateLimitError)
				if !ok {
					t.Errorf("OpenAIService.handleErrorResponse() error type = %T, want *RateLimitError", err)
					return
				}
				if rateLimitErr.StatusCode != tt.statusCode {
					t.Errorf("RateLimitError.StatusCode = %v, want %v", rateLimitErr.StatusCode, tt.statusCode)
				}
				if rateLimitErr.Message != string(tt.body) {
					t.Errorf("RateLimitError.Message = %v, want %v", rateLimitErr.Message, string(tt.body))
				}
			default:
				apiErr, ok := err.(*OpenAIError)
				if !ok {
					t.Errorf("OpenAIService.handleErrorResponse() error type = %T, want *OpenAIError", err)
					return
				}
				if apiErr.StatusCode != tt.statusCode {
					t.Errorf("OpenAIError.StatusCode = %v, want %v", apiErr.StatusCode, tt.statusCode)
				}
				if apiErr.Message != string(tt.body) {
					t.Errorf("OpenAIError.Message = %v, want %v", apiErr.Message, string(tt.body))
				}
				if apiErr.Operation != "API request" {
					t.Errorf("OpenAIError.Operation = %v, want %v", apiErr.Operation, "API request")
				}
			}
		})
	}
}

func Test_OpenAIService_parseResponse(t *testing.T) {
	tests := []struct {
		name        string
		body        []byte
		result      interface{}
		expectError bool
		errorType   error
	}{
		{
			name: "Successfully_parse_response",
			body: []byte(`{
				"choices": [
					{
						"message": {
							"content": "{\"category\": \"Groceries\", \"subcategory\": \"Supermarket\"}"
						}
					}
				]
			}`),
			result: &struct {
				Category    string `json:"category"`
				Subcategory string `json:"subcategory"`
			}{},
			expectError: false,
		},
		{
			name:        "ParseResponse_error_invalid_json",
			body:        []byte(`{invalid json}`),
			result:      &struct{}{},
			expectError: true,
		},
		{
			name:        "ParseResponse_error_no_choices",
			body:        []byte(`{"choices": []}`),
			result:      &struct{}{},
			expectError: true,
			errorType:   ErrNoChoices,
		},
		{
			name: "ParseResponse_error_empty_content",
			body: []byte(`{
				"choices": [
					{
						"message": {
							"content": ""
						}
					}
				]
			}`),
			result:      &struct{}{},
			expectError: true,
			errorType:   ErrEmptyContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the service
			service := NewOpenAIService(Config{
				BaseURL:        "https://api.openai.com",
				APIKey:         "test-key",
				RequestTimeout: 30 * time.Second,
			}, nil, slog.Default())

			// Call the method
			err := service.parseResponse(tt.body, tt.result)

			if tt.expectError {
				if err == nil {
					t.Errorf("OpenAIService.parseResponse() error = nil, want error")
					return
				}
				if tt.errorType != nil && err != tt.errorType {
					t.Errorf("OpenAIService.parseResponse() error = %v, want %v", err, tt.errorType)
				}
			} else {
				if err != nil {
					t.Errorf("OpenAIService.parseResponse() error = %v, want nil", err)
					return
				}
				// For successful case, check the parsed result
				if result, ok := tt.result.(*struct {
					Category    string `json:"category"`
					Subcategory string `json:"subcategory"`
				}); ok {
					if result.Category != "Groceries" {
						t.Errorf("Parsed result.Category = %v, want %v", result.Category, "Groceries")
					}
					if result.Subcategory != "Supermarket" {
						t.Errorf("Parsed result.Subcategory = %v, want %v", result.Subcategory, "Supermarket")
					}
				}
			}
		})
	}
}

func Test_OpenAIService_extractJSONContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "Successfully_extract_plain_json",
			content:  `{"category": "Groceries"}`,
			expected: `{"category": "Groceries"}`,
		},
		{
			name:     "Successfully_extract_json_from_markdown_code_block",
			content:  "```json\n{\"category\": \"Groceries\"}\n```",
			expected: `{"category": "Groceries"}`,
		},
		{
			name:     "Successfully_extract_json_from_simple_code_block",
			content:  "```\n{\"category\": \"Groceries\"}\n```",
			expected: `{"category": "Groceries"}`,
		},
		{
			name:     "Successfully_extract_json_from_swedish_text",
			content:  "Här är kategorin: [{\"category\": \"Groceries\"}]",
			expected: "[{\"category\": \"Groceries\"}]",
		},
		{
			name:     "Successfully_extract_json_from_swedish_text_with_for",
			content:  "För denna transaktion: [{\"category\": \"Groceries\"}]",
			expected: "[{\"category\": \"Groceries\"}]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the service
			service := NewOpenAIService(Config{
				BaseURL:        "https://api.openai.com",
				APIKey:         "test-key",
				RequestTimeout: 30 * time.Second,
			}, nil, slog.Default())

			// Call the method
			result := service.extractJSONContent(tt.content)
			if result != tt.expected {
				t.Errorf("OpenAIService.extractJSONContent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func Test_OpenAIService_processTransactionsResponse(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		docContent []byte
		expected   *Extraction
		expectErr  bool
	}{
		{
			name: "Successfully_process_transactions_with_total",
			content: `[
				{"beskrivning": "Item 1", "belopp": 100.50},
				{"beskrivning": "Item 2", "belopp": 200.75},
				{"beskrivning": "Total", "belopp": 301.25, "metadata": {"date": "2024-03-20"}}
			]`,
			docContent: []byte("Test document content"),
			expected: &Extraction{
				Currency:    "SEK",
				Amount:      301.25,
				Description: "Item 1 (100.50 SEK), Item 2 (200.75 SEK), Total (301.25 SEK)",
				Content:     "Test document content",
			},
			expectErr: false,
		},
		{
			name: "Successfully_process_transactions_without_total",
			content: `[
				{"beskrivning": "Item 1", "belopp": 100.50},
				{"beskrivning": "Item 2", "belopp": 200.75}
			]`,
			docContent: []byte("Test document content"),
			expected: &Extraction{
				Currency:    "SEK",
				Amount:      100.50,
				Description: "Item 1 (100.50 SEK), Item 2 (200.75 SEK)",
				Content:     "Test document content",
			},
			expectErr: false,
		},
		{
			name:       "Successfully_handle_invalid_json",
			content:    `{invalid json}`,
			docContent: []byte("Test document content"),
			expected: &Extraction{
				Currency: "SEK",
				Content:  "Test document content",
			},
			expectErr: false,
		},
		{
			name:       "ProcessTransactionsResponse_error_empty_transactions",
			content:    `[]`,
			docContent: []byte("Test document content"),
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the service
			service := NewOpenAIService(Config{
				BaseURL:        "https://api.openai.com",
				APIKey:         "test-key",
				RequestTimeout: 30 * time.Second,
			}, nil, slog.Default())

			// Call the method
			result, err := service.processTransactionsResponse(tt.content, tt.docContent)

			if tt.expectErr {
				if err == nil {
					t.Errorf("OpenAIService.processTransactionsResponse() error = nil, want error")
				}
			} else {
				if err != nil {
					t.Errorf("OpenAIService.processTransactionsResponse() error = %v, want nil", err)
					return
				}

				// For successful case, check the extraction
				if tt.expected != nil {
					if result.Currency != tt.expected.Currency {
						t.Errorf("OpenAIService.processTransactionsResponse() result.Currency = %v, want %v",
							result.Currency, tt.expected.Currency)
					}
					if result.Content != tt.expected.Content {
						t.Errorf("OpenAIService.processTransactionsResponse() result.Content = %v, want %v",
							result.Content, tt.expected.Content)
					}

					// For valid transactions, check the description contains all items
					if !strings.Contains(tt.content, "invalid") && !strings.Contains(tt.content, "[]") {
						if !strings.Contains(result.Description, "Item 1") {
							t.Errorf("OpenAIService.processTransactionsResponse() result.Description = %v, should contain 'Item 1'",
								result.Description)
						}
						if !strings.Contains(result.Description, "Item 2") {
							t.Errorf("OpenAIService.processTransactionsResponse() result.Description = %v, should contain 'Item 2'",
								result.Description)
						}
					}
				}
			}
		})
	}
}

func Test_OpenAIService_handleUnmarshalError(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		result       interface{}
		unmarshalErr error
		expectError  bool
	}{
		{
			name:         "Successfully_convert_single_object_to_array",
			content:      `{"category": "Groceries", "confidence": 0.95}`,
			result:       &[]map[string]interface{}{},
			unmarshalErr: fmt.Errorf("json: cannot unmarshal object into Go value of type []map[string]interface {}"),
			expectError:  false,
		},
		{
			name:         "HandleUnmarshalError_error_non_convertible_result",
			content:      `{"category": "Groceries", "confidence": 0.95}`,
			result:       &struct{}{},
			unmarshalErr: fmt.Errorf("json: cannot unmarshal object into Go value of type struct {}"),
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the service
			service := NewOpenAIService(Config{
				BaseURL:        "https://api.openai.com",
				APIKey:         "test-key",
				RequestTimeout: 30 * time.Second,
			}, nil, slog.Default())

			// Call the method
			err := service.handleUnmarshalError(tt.content, tt.result, tt.unmarshalErr)

			if tt.expectError {
				if err == nil {
					t.Errorf("OpenAIService.handleUnmarshalError() error = nil, want error")
				}
			} else {
				if err != nil {
					t.Errorf("OpenAIService.handleUnmarshalError() error = %v, want nil", err)
					return
				}

				// For successful case with array conversion
				if resultArr, ok := tt.result.(*[]map[string]interface{}); ok {
					if len(*resultArr) != 1 {
						t.Errorf("OpenAIService.handleUnmarshalError() resultArr length = %d, want 1", len(*resultArr))
						return
					}
					if (*resultArr)[0]["category"] != "Groceries" {
						t.Errorf("OpenAIService.handleUnmarshalError() resultArr[0][\"category\"] = %v, want %v",
							(*resultArr)[0]["category"], "Groceries")
					}
					if (*resultArr)[0]["confidence"] != 0.95 {
						t.Errorf("OpenAIService.handleUnmarshalError() resultArr[0][\"confidence\"] = %v, want %v",
							(*resultArr)[0]["confidence"], 0.95)
					}
				}
			}
		})
	}
}
