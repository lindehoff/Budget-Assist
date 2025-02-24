package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	db "github.com/lindehoff/Budget-Assist/internal/db"
)

// Common errors
var (
	ErrEmptyDocument    = fmt.Errorf("empty document content")
	ErrNoChoices        = fmt.Errorf("no choices in OpenAI response")
	ErrEmptyContent     = fmt.Errorf("empty content in OpenAI response")
	ErrTemplateNotFound = fmt.Errorf("template not found")
	ErrInvalidOperation = fmt.Errorf("invalid operation")
)

// OperationError represents an error that occurred during an operation
type OperationError struct {
	Err       error
	Operation string
	Resource  string
}

func (e *OperationError) Error() string {
	if e.Resource != "" {
		return fmt.Sprintf("%s operation failed for %q: %v", e.Operation, e.Resource, e.Err)
	}
	return fmt.Sprintf("%s operation failed: %v", e.Operation, e.Err)
}

// OpenAIError represents an error from the OpenAI API
type OpenAIError struct {
	Operation  string
	Message    string
	StatusCode int
}

func (e *OpenAIError) Error() string {
	return fmt.Sprintf("OpenAI API error during %s operation (status %d): %s", e.Operation, e.StatusCode, e.Message)
}

// RateLimitError represents a rate limit error from the OpenAI API
type RateLimitError struct {
	Message    string
	StatusCode int
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded (status %d): %s", e.StatusCode, e.Message)
}

// OpenAIService implements the Service interface using OpenAI's API.
type OpenAIService struct {
	rateLimiter *RateLimiter
	client      *http.Client
	config      Config
	retryConfig RetryConfig
	promptMgr   *PromptManager
}

// NewOpenAIService returns a new instance of OpenAIService.
func NewOpenAIService(config Config, logger *slog.Logger) Service {
	return &OpenAIService{
		client: &http.Client{
			Timeout: config.RequestTimeout,
		},
		config:      config,
		rateLimiter: NewRateLimiter(10, 30), // 10 requests per second, burst of 30
		retryConfig: DefaultRetryConfig,
		promptMgr:   NewPromptManager(logger),
	}
}

// AnalyzeTransaction analyzes a transaction using OpenAI's API.
func (s *OpenAIService) AnalyzeTransaction(ctx context.Context, tx *db.Transaction) (*Analysis, error) {
	template, err := s.promptMgr.GetPrompt(ctx, TransactionAnalysisPrompt)
	if err != nil {
		return nil, &OperationError{
			Operation: "AnalyzeTransaction",
			Err:       err,
		}
	}

	data := struct {
		Description string
		Amount      string
		Date        string
		Rules       []Rule
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
		return nil, &OperationError{
			Operation: "AnalyzeTransaction",
			Err:       fmt.Errorf("failed to generate prompt: %w", err),
		}
	}

	requestPayload := map[string]any{
		"model": "gpt-4",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": template.SystemPrompt,
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
		return nil, &OperationError{
			Operation: "AnalyzeTransaction",
			Err:       err,
		}
	}

	return &result, nil
}

// ExtractDocument extracts data from a document using OpenAI's API.
func (s *OpenAIService) ExtractDocument(ctx context.Context, doc *Document) (*Extraction, error) {
	if len(doc.Content) == 0 {
		return nil, &OperationError{
			Operation: "ExtractDocument",
			Err:       ErrEmptyDocument,
		}
	}

	template, err := s.promptMgr.GetPrompt(ctx, DocumentExtractionPrompt)
	if err != nil {
		return nil, &OperationError{
			Operation: "ExtractDocument",
			Err:       err,
		}
	}

	data := struct {
		Content string
	}{
		Content: string(doc.Content),
	}

	prompt, err := template.Execute(data)
	if err != nil {
		return nil, &OperationError{
			Operation: "ExtractDocument",
			Err:       fmt.Errorf("failed to generate prompt: %w", err),
		}
	}

	requestPayload := map[string]any{
		"model": "gpt-4",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": template.SystemPrompt,
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
		return nil, &OperationError{
			Operation: "ExtractDocument",
			Err:       err,
		}
	}

	return &result, nil
}

// SuggestCategories suggests categories for a given description using OpenAI's API.
func (s *OpenAIService) SuggestCategories(ctx context.Context, desc string) ([]CategoryMatch, error) {
	template, err := s.promptMgr.GetPrompt(ctx, CategorySuggestionPrompt)
	if err != nil {
		return nil, &OperationError{
			Operation: "SuggestCategories",
			Err:       err,
		}
	}

	data := struct {
		Description string
		Categories  []Category
	}{
		Description: desc,
		Categories:  template.Categories,
	}

	prompt, err := template.Execute(data)
	if err != nil {
		return nil, &OperationError{
			Operation: "SuggestCategories",
			Err:       fmt.Errorf("failed to generate prompt: %w", err),
		}
	}

	requestPayload := map[string]any{
		"model": "gpt-4",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": template.SystemPrompt,
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.3,
	}

	var result []CategoryMatch
	err = s.doRequestWithRetry(ctx, requestPayload, &result, "/v1/chat/completions")
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
		return &OperationError{
			Operation: "doRequest",
			Err:       fmt.Errorf("failed to marshal request payload: %w", err),
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return &OperationError{
			Operation: "doRequest",
			Err:       fmt.Errorf("failed to create request: %w", err),
		}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return &OperationError{
			Operation: "doRequest",
			Err:       fmt.Errorf("request failed: %w", err),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &OperationError{
			Operation: "doRequest",
			Err:       fmt.Errorf("failed to read response body: %w", err),
		}
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return &RateLimitError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	if resp.StatusCode != http.StatusOK {
		return &OpenAIError{
			Operation:  "doRequest",
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	// First decode into a raw OpenAI response to get the content
	var openAIResp OpenAIResponse
	if err = json.Unmarshal(body, &openAIResp); err != nil {
		return &OperationError{
			Operation: "doRequest",
			Err:       fmt.Errorf("failed to decode OpenAI response: %w", err),
		}
	}

	if len(openAIResp.Choices) == 0 {
		return ErrNoChoices
	}

	content := openAIResp.Choices[0].Message.Content
	if content == "" {
		return ErrEmptyContent
	}

	// Now decode the content into the target type
	if err = json.Unmarshal([]byte(content), out); err != nil {
		return &OperationError{
			Operation: "doRequest",
			Err:       fmt.Errorf("failed to decode response content: %w", err),
		}
	}

	return nil
}
