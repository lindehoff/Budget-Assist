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
func (s *OpenAIService) AnalyzeTransaction(ctx context.Context, tx *db.Transaction, opts AnalysisOptions) (*Analysis, error) {
	var promptType db.PromptType
	switch opts.DocumentType {
	case "bill":
		promptType = db.BillAnalysisPrompt
	case "receipt":
		promptType = db.ReceiptAnalysisPrompt
	case "bank_statement":
		promptType = db.BankStatementAnalysisPrompt
	default:
		return nil, &OperationError{
			Operation: "AnalyzeTransaction",
			Err:       fmt.Errorf("unsupported document type: %s", opts.DocumentType),
		}
	}

	template, err := s.promptMgr.GetPrompt(ctx, promptType)
	if err != nil {
		return nil, &OperationError{
			Operation: "AnalyzeTransaction",
			Err:       err,
		}
	}

	data := struct {
		Description     string
		Amount          string
		Date            string
		RuntimeInsights string
	}{
		Description:     tx.Description,
		Amount:          fmt.Sprintf("%s %s", tx.Amount.String(), tx.Currency),
		Date:            tx.TransactionDate.Format("2006-01-02"),
		RuntimeInsights: opts.RuntimeInsights,
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

	template, err := s.promptMgr.GetPrompt(ctx, db.BillAnalysisPrompt)
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
	template, err := s.promptMgr.GetPrompt(ctx, db.TransactionCategorizationPrompt)
	if err != nil {
		return nil, &OperationError{
			Operation: "SuggestCategories",
			Err:       err,
		}
	}

	data := struct {
		Description string
	}{
		Description: desc,
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
		return nil, &OperationError{
			Operation: "SuggestCategories",
			Err:       err,
		}
	}

	return result, nil
}

func (s *OpenAIService) doRequestWithRetry(ctx context.Context, requestPayload map[string]any, result interface{}, endpoint string) error {
	operation := func() error {
		if err := s.rateLimiter.Wait(ctx); err != nil {
			return err
		}

		requestBody, err := json.Marshal(requestPayload)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.config.BaseURL+endpoint, bytes.NewReader(requestBody))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+s.config.APIKey)

		resp, err := s.client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			if resp.StatusCode == http.StatusTooManyRequests {
				return &RateLimitError{
					Message:    string(body),
					StatusCode: resp.StatusCode,
				}
			}
			return &OpenAIError{
				Operation:  "API request",
				Message:    string(body),
				StatusCode: resp.StatusCode,
			}
		}

		var response ChatCompletionResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}

		if len(response.Choices) == 0 {
			return ErrNoChoices
		}

		content := response.Choices[0].Message.Content
		if content == "" {
			return ErrEmptyContent
		}

		if err := json.Unmarshal([]byte(content), result); err != nil {
			return fmt.Errorf("failed to unmarshal content: %w", err)
		}

		return nil
	}

	return retryWithBackoff(ctx, s.retryConfig, operation)
}
