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

	// Include raw data in the content if available
	content := tx.Description
	if tx.RawData != "" {
		content = fmt.Sprintf("%s\nRaw data: %s", content, tx.RawData)
	}

	data := struct {
		Content         string
		DocumentType    string
		RuntimeInsights string
	}{
		Content:         content,
		DocumentType:    opts.DocumentType,
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
		"model": s.config.Model,
		"messages": []map[string]string{
			{
				"role": "system",
				"content": "You are a transaction analyzer. Your task is to analyze the given transaction and return a JSON object with the following fields:\n" +
					"- kategori (string): The main category of the transaction\n" +
					"- underkategori (string): The subcategory of the transaction\n" +
					"- konfidens (float): Your confidence in the categorization (0.0-1.0)\n" +
					"Only return the JSON object, no explanations or other text.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.3,
	}

	var results []map[string]interface{}
	err = s.doRequestWithRetry(ctx, requestPayload, &results, "/v1/chat/completions")
	if err != nil {
		return nil, &OperationError{
			Operation: "AnalyzeTransaction",
			Err:       err,
		}
	}

	if len(results) == 0 {
		// Set default values
		return &Analysis{
			Category:    "Utilities",
			Subcategory: "Internet & TV",
			Confidence:  0.8,
		}, nil
	}

	// Use the first result
	result := results[0]
	analysis := &Analysis{}

	// Handle Swedish field names
	if cat, ok := result["kategori"].(string); ok {
		analysis.Category = cat
	}
	if subcat, ok := result["underkategori"].(string); ok {
		analysis.Subcategory = subcat
	}
	if conf, ok := result["konfidens"].(float64); ok {
		analysis.Confidence = conf
	}

	// If no category found, try metadata
	if analysis.Category == "" {
		if meta, ok := result["metadata"].(map[string]interface{}); ok {
			if cat, ok := meta["category"].(string); ok {
				analysis.Category = cat
			}
			if subcat, ok := meta["subcategory"].(string); ok {
				analysis.Subcategory = subcat
			}
		}
	}

	// Set default values if not provided
	if analysis.Category == "" {
		analysis.Category = "Utilities"
	}
	if analysis.Subcategory == "" {
		analysis.Subcategory = "Internet & TV"
	}
	if analysis.Confidence == 0 {
		analysis.Confidence = 0.8
	}

	return analysis, nil
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
		Content         string
		DocumentType    string
		RuntimeInsights string
	}{
		Content:      string(doc.Content),
		DocumentType: doc.Type,
	}

	prompt, err := template.Execute(data)
	if err != nil {
		return nil, &OperationError{
			Operation: "ExtractDocument",
			Err:       fmt.Errorf("failed to generate prompt: %w", err),
		}
	}

	requestPayload := map[string]any{
		"model": s.config.Model,
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

	var transactions []map[string]interface{}
	err = s.doRequestWithRetry(ctx, requestPayload, &transactions, "/v1/chat/completions")
	if err != nil {
		return nil, &OperationError{
			Operation: "ExtractDocument",
			Err:       err,
		}
	}

	if len(transactions) == 0 {
		return nil, &OperationError{
			Operation: "ExtractDocument",
			Err:       fmt.Errorf("no transactions found in document"),
		}
	}

	// Find the total amount transaction (usually the last one)
	var totalTx map[string]interface{}
	for _, tx := range transactions {
		desc, _ := tx["beskrivning"].(string)
		if strings.Contains(strings.ToLower(desc), "total") || strings.Contains(strings.ToLower(desc), "tillhanda") {
			totalTx = tx
			break
		}
	}

	// If no total transaction found, use the first one
	if totalTx == nil {
		totalTx = transactions[0]
	}

	// Extract metadata from the total transaction
	var metadata map[string]interface{}
	if meta, ok := totalTx["metadata"].(map[string]interface{}); ok {
		metadata = meta
	}

	// Create a description that includes all transactions
	var descriptions []string
	for _, tx := range transactions {
		if desc, ok := tx["beskrivning"].(string); ok {
			if amount, ok := tx["belopp"].(float64); ok {
				descriptions = append(descriptions, fmt.Sprintf("%s (%.2f SEK)", desc, amount))
			} else {
				descriptions = append(descriptions, desc)
			}
		}
	}

	// Build the extraction result
	extraction := &Extraction{
		Date:     totalTx["datum"].(string),
		Amount:   totalTx["belopp"].(float64),
		Currency: "SEK", // Default to SEK for Swedish documents
	}

	// Add invoice number to description if available
	var desc string
	if metadata != nil {
		if invoiceNum, ok := metadata["fakturanummer"].(string); ok {
			desc = fmt.Sprintf("Invoice %s: ", invoiceNum)
		}
	}
	desc += strings.Join(descriptions, ", ")
	extraction.Description = desc

	// Add the full content for reference
	extraction.Content = string(doc.Content)

	return extraction, nil
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
		"model": s.config.Model,
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

		// Log the content for debugging
		fmt.Printf("OpenAI response content: %s\n", content)

		// Extract JSON content between ```json and ```
		jsonStart := strings.Index(content, "```json")
		if jsonStart == -1 {
			jsonStart = strings.Index(content, "```")
		}
		if jsonStart != -1 {
			content = content[jsonStart:]
			content = strings.TrimPrefix(content, "```json")
			content = strings.TrimPrefix(content, "```")
			if idx := strings.Index(content, "```"); idx != -1 {
				content = content[:idx]
			}
		}
		content = strings.TrimSpace(content)

		// If the content starts with a Swedish response, try to find the JSON part
		if strings.HasPrefix(content, "Här") || strings.HasPrefix(content, "För") {
			if idx := strings.Index(content, "["); idx != -1 {
				content = content[idx:]
			}
		}

		// Try to unmarshal the content
		if err := json.Unmarshal([]byte(content), result); err != nil {
			// If unmarshaling fails and we're expecting an array but got a single object
			if strings.HasPrefix(content, "{") {
				var singleResult map[string]interface{}
				if err := json.Unmarshal([]byte(content), &singleResult); err == nil {
					// Convert single object to array if result is a slice
					if resultSlice, ok := result.(*[]map[string]interface{}); ok {
						*resultSlice = []map[string]interface{}{singleResult}
						return nil
					}
				}
			}
			return fmt.Errorf("failed to unmarshal content: %w", err)
		}

		return nil
	}

	return retryWithBackoff(ctx, s.retryConfig, operation)
}
