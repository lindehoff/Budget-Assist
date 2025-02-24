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
func NewOpenAIService(config Config, store db.Store, logger *slog.Logger) *OpenAIService {
	return &OpenAIService{
		rateLimiter: NewRateLimiter(10, 20), // 10 RPS with burst of 20
		client: &http.Client{
			Timeout: config.RequestTimeout * time.Second,
		},
		config:      config,
		retryConfig: DefaultRetryConfig,
		promptMgr:   NewPromptManager(store, logger),
	}
}

// AnalyzeTransaction analyzes a transaction using OpenAI's API.
func (s *OpenAIService) AnalyzeTransaction(ctx context.Context, tx *db.Transaction) (*Analysis, error) {
	template, err := s.promptMgr.GetPrompt(ctx, TransactionAnalysisPrompt)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, &OperationError{
				Operation: "AnalyzeTransaction",
				Err:       ctx.Err(),
			}
		}
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
		Date:        tx.TransactionDate.Format("2006-01-02"),
		Rules:       template.Rules,
		Examples:    template.Examples,
	}

	prompt, err := template.Execute(data)
	if err != nil {
		return nil, &OperationError{
			Operation: "AnalyzeTransaction",
			Err:       fmt.Errorf("failed to execute template: %w", err),
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
		if ctx.Err() == context.DeadlineExceeded {
			return nil, &OperationError{
				Operation: "AnalyzeTransaction",
				Err:       ctx.Err(),
			}
		}
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
	template, err := s.promptMgr.GetPrompt(ctx, TransactionCategorizationPrompt)
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
	maxRetries := 3
	backoff := 1 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		if err := s.makeRequest(ctx, payload, out, endpoint); err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return &OperationError{
					Operation: "doRequestWithRetry",
					Err:       ctx.Err(),
				}
			}

			// If it's a rate limit error and we have retries left, wait and try again
			if _, isRateLimit := err.(*RateLimitError); isRateLimit && attempt < maxRetries-1 {
				select {
				case <-ctx.Done():
					return &OperationError{
						Operation: "doRequestWithRetry",
						Err:       ctx.Err(),
					}
				case <-time.After(backoff * time.Duration(attempt+1)):
					continue
				}
			}

			return err
		}

		return nil
	}

	return &OperationError{
		Operation: "doRequestWithRetry",
		Err:       fmt.Errorf("max retries exceeded"),
	}
}

// makeRequest performs a single HTTP request to the OpenAI API
func (s *OpenAIService) makeRequest(ctx context.Context, payload any, out any, endpoint string) error {
	if err := s.rateLimiter.Wait(ctx); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &OperationError{
				Operation: "makeRequest",
				Err:       ctx.Err(),
			}
		}
		return &OperationError{
			Operation: "makeRequest",
			Err:       fmt.Errorf("rate limit wait failed: %w", err),
		}
	}

	url := s.config.BaseURL + endpoint
	data, err := json.Marshal(payload)
	if err != nil {
		return &OperationError{
			Operation: "makeRequest",
			Err:       fmt.Errorf("failed to marshal request: %w", err),
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return &OperationError{
			Operation: "makeRequest",
			Err:       fmt.Errorf("failed to create request: %w", err),
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &OperationError{
				Operation: "makeRequest",
				Err:       ctx.Err(),
			}
		}
		return &OperationError{
			Operation: "makeRequest",
			Err:       fmt.Errorf("request failed: %w", err),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &OperationError{
			Operation: "makeRequest",
			Err:       fmt.Errorf("failed to read response body: %w", err),
		}
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return &RateLimitError{
			Message:    "rate limit exceeded",
			StatusCode: resp.StatusCode,
		}
	}

	if resp.StatusCode != http.StatusOK {
		return &OpenAIError{
			Operation:  "makeRequest",
			Message:    string(body),
			StatusCode: resp.StatusCode,
		}
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return &OperationError{
			Operation: "makeRequest",
			Err:       fmt.Errorf("failed to decode response content: %w", err),
		}
	}

	if len(response.Choices) == 0 {
		return &OperationError{
			Operation: "makeRequest",
			Err:       ErrNoChoices,
		}
	}

	content := response.Choices[0].Message.Content
	if content == "" {
		return &OperationError{
			Operation: "makeRequest",
			Err:       ErrEmptyContent,
		}
	}

	if err := json.Unmarshal([]byte(content), out); err != nil {
		return &OperationError{
			Operation: "makeRequest",
			Err:       fmt.Errorf("failed to decode response content: %w", err),
		}
	}

	return nil
}
