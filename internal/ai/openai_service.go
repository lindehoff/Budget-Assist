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
	logger      *slog.Logger
	store       db.Store
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
		logger:      logger,
		store:       store,
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
					"- main_category: The main category of the transaction\n" +
					"- sub_category: The subcategory of the transaction\n" +
					"- confidence: Your confidence in the categorization (0.0-1.0)\n" +
					"- metadata: Any additional metadata extracted from the transaction\n",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.3,
	}

	s.logger.Debug("Sending transaction analysis request",
		"model", s.config.Model,
		"content_length", len(prompt))

	var results []map[string]interface{}
	err = s.doRequestWithRetry(ctx, requestPayload, &results, "/v1/chat/completions")
	if err != nil {
		s.logger.Error("OpenAI API request failed", "error", err)
		return nil, &OperationError{
			Operation: "AnalyzeTransaction",
			Err:       fmt.Errorf("failed to make API request: %w", err),
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

	s.logger.Debug("Received OpenAI API response successfully")

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

// SuggestCategories suggests categories for a transaction description
func (s *OpenAIService) SuggestCategories(ctx context.Context, desc string) ([]CategoryMatch, error) {
	// Get the prompt template
	template, err := s.promptMgr.GetPrompt(ctx, db.TransactionCategorizationPrompt)
	if err != nil {
		return nil, &OperationError{
			Operation: "SuggestCategories",
			Err:       err,
		}
	}

	// Get category information
	categoryInfos, err := s.getCategoryInfos(ctx)
	if err != nil {
		return nil, err
	}

	// Generate the prompt
	prompt, err := s.generateCategoryPrompt(template, desc, categoryInfos)
	if err != nil {
		return nil, err
	}

	// Make the API request
	rawResults, err := s.makeCategoryRequest(ctx, template.SystemPrompt, prompt)
	if err != nil {
		return nil, err
	}

	// Process the results
	return s.processCategoryResults(rawResults), nil
}

// getCategoryInfos retrieves and processes category information from the database
func (s *OpenAIService) getCategoryInfos(ctx context.Context) ([]CategoryInfo, error) {
	// Get all available categories
	categories, err := s.store.ListCategories(ctx, nil)
	if err != nil {
		return nil, &OperationError{
			Operation: "SuggestCategories",
			Err:       fmt.Errorf("failed to list categories: %w", err),
		}
	}

	// Log categories for debugging
	if s.logger != nil {
		s.logger.Debug("Retrieved categories", "count", len(categories))
		for _, cat := range categories {
			s.logger.Debug("Category details",
				"name", cat.Name,
				"id", cat.ID,
				"description", cat.Description,
				"type", cat.Type,
				"active", cat.IsActive,
				"subcategories_count", len(cat.Subcategories))
		}
	}

	// Get all available subcategories
	subcategories, err := s.store.ListSubcategories(ctx)
	if err != nil {
		return nil, &OperationError{
			Operation: "SuggestCategories",
			Err:       fmt.Errorf("failed to list subcategories: %w", err),
		}
	}

	// Log subcategories for debugging
	if s.logger != nil {
		s.logger.Debug("Retrieved subcategories", "count", len(subcategories))
		for _, subcat := range subcategories {
			s.logger.Debug("Subcategory details",
				"name", subcat.Name,
				"id", subcat.ID,
				"description", subcat.Description,
				"active", subcat.IsActive,
				"system", subcat.IsSystem,
				"tags_count", len(subcat.Tags))
		}
	}

	// Build category paths for the prompt
	var categoryInfos []CategoryInfo

	for _, cat := range categories {
		// Skip inactive categories
		if !cat.IsActive {
			continue
		}
		for _, subcat := range subcategories {
			// Check if this subcategory is linked to this category
			isLinked := false
			for _, link := range cat.Subcategories {
				if link.SubcategoryID == subcat.ID && link.IsActive {
					isLinked = true
					break
				}
			}
			if isLinked {
				path := fmt.Sprintf("%s/%s", cat.Name, subcat.Name)
				desc := fmt.Sprintf("%s - %s", cat.Description, subcat.Description)
				categoryInfos = append(categoryInfos, CategoryInfo{
					Path:        path,
					Description: desc,
				})
			}
		}
	}

	// Log available categories for debugging
	if s.logger != nil {
		s.logger.Debug("Available category paths", "count", len(categoryInfos))
		for i, cat := range categoryInfos {
			s.logger.Debug("Category path",
				"index", i+1,
				"path", cat.Path,
				"description", cat.Description)
		}
	}

	return categoryInfos, nil
}

// CategoryInfo represents a category path and description
type CategoryInfo struct {
	Path        string
	Description string
}

// generateCategoryPrompt generates the prompt for category suggestion
func (s *OpenAIService) generateCategoryPrompt(template *PromptTemplate, desc string, categoryInfos []CategoryInfo) (string, error) {
	data := struct {
		Description string
		Categories  []CategoryInfo
	}{
		Description: desc,
		Categories:  categoryInfos,
	}

	prompt, err := template.Execute(data)
	if err != nil {
		return "", &OperationError{
			Operation: "SuggestCategories",
			Err:       fmt.Errorf("failed to generate prompt: %w", err),
		}
	}

	// Debug logging for prompt
	if s.logger != nil {
		s.logger.Debug("Generated prompt",
			"description", desc,
			"system_prompt", template.SystemPrompt,
			"user_prompt", prompt,
			"available_categories", len(categoryInfos))
	}

	return prompt, nil
}

// makeCategoryRequest makes the API request for category suggestion
func (s *OpenAIService) makeCategoryRequest(ctx context.Context, systemPrompt string, prompt string) ([]map[string]interface{}, error) {
	requestPayload := map[string]any{
		"model": s.config.Model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.3,
	}

	// First unmarshal into raw map to handle different response formats
	var rawResults []map[string]interface{}
	err := s.doRequestWithRetry(ctx, requestPayload, &rawResults, "/v1/chat/completions")
	if err != nil {
		return nil, &OperationError{
			Operation: "SuggestCategories",
			Err:       err,
		}
	}

	return rawResults, nil
}

// processCategoryResults processes the raw results from the API
func (s *OpenAIService) processCategoryResults(rawResults []map[string]interface{}) []CategoryMatch {
	// Convert response format
	matches := make([]CategoryMatch, 0, len(rawResults))
	for _, raw := range rawResults {
		// Try both English and Swedish field names
		var category string
		if cat, ok := raw["main_category"].(string); ok {
			category = cat
		} else if cat, ok := raw["kategori"].(string); ok {
			category = cat
		}

		var confidence float64
		if conf, ok := raw["confidence"].(float64); ok {
			confidence = conf
		} else if conf, ok := raw["konfidens"].(float64); ok {
			confidence = conf
		}

		if category != "" {
			matches = append(matches, CategoryMatch{
				Category:   category,
				Confidence: confidence,
				Raw:        raw,
			})
		}
	}

	return matches
}

// doRequestWithRetry sends a request to the OpenAI API with retry logic
func (s *OpenAIService) doRequestWithRetry(ctx context.Context, requestPayload map[string]any, result interface{}, endpoint string) error {
	operation := func() error {
		if err := s.rateLimiter.Wait(ctx); err != nil {
			s.logger.Error("Rate limiter wait failed", "error", err)
			return err
		}

		// Marshal the request payload
		requestBody, err := json.Marshal(requestPayload)
		if err != nil {
			s.logger.Error("Failed to marshal request", "error", err)
			return fmt.Errorf("failed to marshal request: %w", err)
		}

		// Create the request
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.config.BaseURL+endpoint, bytes.NewReader(requestBody))
		if err != nil {
			s.logger.Error("Failed to create request", "error", err, "url", s.config.BaseURL+endpoint)
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+s.config.APIKey)

		s.logger.Debug("Sending API request",
			"endpoint", endpoint,
			"model", requestPayload["model"],
			"request_size", len(requestBody))

		// Send the request
		resp, err := s.client.Do(req)
		if err != nil {
			s.logger.Error("Failed to send request", "error", err, "endpoint", endpoint)
			return fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close() // Ensure body is closed

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			s.logger.Error("Failed to read response", "error", err)
			return fmt.Errorf("failed to read response: %w", err)
		}

		s.logger.Debug("Received API response",
			"status_code", resp.StatusCode,
			"content_length", len(body))

		// Handle non-200 responses
		if resp.StatusCode != http.StatusOK {
			return s.handleErrorResponse(resp, body)
		}

		// Parse the response
		return s.parseResponse(body, result)
	}

	return retryWithBackoff(ctx, s.retryConfig, operation)
}

// handleErrorResponse processes non-200 responses from the API
func (s *OpenAIService) handleErrorResponse(resp *http.Response, body []byte) error {
	s.logger.Error("API request failed",
		"status_code", resp.StatusCode,
		"response", string(body))

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

// parseResponse processes the API response and extracts the result
func (s *OpenAIService) parseResponse(body []byte, result interface{}) error {
	var response ChatCompletionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		s.logger.Error("Failed to unmarshal response", "error", err)
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Choices) == 0 {
		s.logger.Error("No choices returned in response")
		return ErrNoChoices
	}

	content := response.Choices[0].Message.Content
	if content == "" {
		s.logger.Error("Empty content returned in response")
		return ErrEmptyContent
	}

	// Log the content for debugging
	s.logger.Debug("Extracted response content",
		"content_length", len(content),
		"choices_count", len(response.Choices))

	// Extract and process the content
	processedContent := s.extractJSONContent(content)

	// Try to unmarshal the content
	if err := json.Unmarshal([]byte(processedContent), result); err != nil {
		return s.handleUnmarshalError(processedContent, result, err)
	}

	s.logger.Debug("Successfully processed API response and unmarshaled result")
	return nil
}

// extractJSONContent extracts JSON content from the response
func (s *OpenAIService) extractJSONContent(content string) string {
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

	return content
}

// handleUnmarshalError handles errors when unmarshaling the response content
func (s *OpenAIService) handleUnmarshalError(content string, result interface{}, err error) error {
	s.logger.Warn("Failed to unmarshal content",
		"error", err,
		"content", content)

	// If unmarshaling fails and we're expecting an array but got a single object
	if strings.HasPrefix(content, "{") {
		var singleResult map[string]interface{}
		if err := json.Unmarshal([]byte(content), &singleResult); err == nil {
			// Convert single object to array if result is a slice
			if resultSlice, ok := result.(*[]map[string]interface{}); ok {
				*resultSlice = []map[string]interface{}{singleResult}
				s.logger.Debug("Converted single object to array")
				return nil
			}
		}
	}
	return fmt.Errorf("failed to unmarshal content: %w", err)
}
