package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	db "github.com/lindehoff/Budget-Assist/internal/db"
)

// OpenAIService implements the AIService interface using OpenAI's API.
type OpenAIService struct {
	rateLimiter *RateLimiter // 8 bytes (pointer)
	client      *http.Client // 8 bytes (pointer)
	config      AIConfig     // struct
	retryConfig RetryConfig  // struct
}

// NewOpenAIService returns a new instance of OpenAIService.
func NewOpenAIService(config AIConfig) AIService {
	return &OpenAIService{
		client: &http.Client{
			Timeout: config.RequestTimeout,
		},
		config:      config,
		rateLimiter: NewRateLimiter(10, 30), // 10 requests per second, burst of 30
		retryConfig: DefaultRetryConfig,
	}
}

// AnalyzeTransaction analyzes a transaction using OpenAI's API.
func (s *OpenAIService) AnalyzeTransaction(ctx context.Context, tx *db.Transaction) (*Analysis, error) {
	template, ok := promptTemplates["categorize"]
	if !ok {
		return nil, fmt.Errorf("categorization template not found")
	}

	data := struct {
		Description string
		Amount      string
		Date        string
		Rules       []string
		Examples    []Example
	}{
		Description: tx.Description,
		Amount:      fmt.Sprintf("%.2f %s", tx.Amount, tx.Currency),
		Date:        tx.TransactionDate.Format(time.RFC3339),
		Rules:       template.Rules,
		Examples:    template.Examples,
	}

	prompt, err := template.Execute(data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate prompt: %w", err)
	}

	requestPayload := map[string]any{
		"model": "gpt-4",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a financial transaction analyzer. Analyze transactions and provide categories with confidence scores.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.3,
	}

	var result Analysis
	err = s.doRequestWithRetry(ctx, requestPayload, &result, "/v1/chat/completions")
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// ExtractDocument extracts data from a document using OpenAI's API.
func (s *OpenAIService) ExtractDocument(ctx context.Context, doc *Document) (*Extraction, error) {
	template, ok := promptTemplates["extract"]
	if !ok {
		return nil, fmt.Errorf("extraction template not found")
	}

	data := struct {
		Content string
	}{
		Content: string(doc.Content),
	}

	prompt, err := template.Execute(data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate prompt: %w", err)
	}

	requestPayload := map[string]any{
		"model": "gpt-4",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a document information extractor. Extract and structure financial information from documents.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.2,
	}

	var result Extraction
	err = s.doRequestWithRetry(ctx, requestPayload, &result, "/v1/chat/completions")
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// SuggestCategories suggests categories for a given description using OpenAI's API.
func (s *OpenAIService) SuggestCategories(ctx context.Context, desc string) ([]CategoryMatch, error) {
	requestPayload := map[string]any{
		"model": "gpt-4",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a financial transaction categorizer. Suggest relevant categories with confidence scores.",
			},
			{
				"role":    "user",
				"content": fmt.Sprintf("Suggest categories for this transaction: %s", desc),
			},
		},
		"temperature": 0.3,
	}

	var result []CategoryMatch
	err := s.doRequestWithRetry(ctx, requestPayload, &result, "/v1/chat/completions")
	if err != nil {
		return nil, err
	}

	return result, nil
}

// doRequestWithRetry performs an HTTP request with retry and rate limiting
func (s *OpenAIService) doRequestWithRetry(ctx context.Context, payload any, out any, endpoint string) error {
	op := func() error {
		if err := s.rateLimiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limiter wait failed: %w", err)
		}
		return s.doRequest(ctx, payload, out, endpoint)
	}

	return retryWithBackoff(ctx, s.retryConfig, op)
}

// doRequest is a helper method to perform an HTTP POST request to the OpenAI API.
func (s *OpenAIService) doRequest(ctx context.Context, payload any, out any, endpoint string) error {
	url := s.config.BaseURL + endpoint
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	if err = json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
