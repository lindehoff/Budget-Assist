# OpenAI Integration Design

## Overview
The system leverages OpenAI's GPT models for intelligent transaction categorization and document information extraction, providing high-accuracy financial data processing. The system supports multiple use cases including transaction categorization and document analysis, with configurable prompts and rules.

## OpenAI Service Interface

```go
type AIService interface {
    AnalyzeTransaction(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error)
    ExtractDocument(ctx context.Context, doc DocumentData) (*ExtractionResult, error)
    SuggestCategories(ctx context.Context, description string) ([]CategorySuggestion, error)
    UpdatePrompt(ctx context.Context, promptType PromptType, content string) error
    UpdateRules(ctx context.Context, rules []Rule) error
}

type AnalysisRequest struct {
    Description      string
    Amount          float64
    Date            time.Time
    Metadata        map[string]interface{}
    UserPreferences UserPrefs
    PromptType      PromptType // Specific prompt type for this analysis
}

type AnalysisResponse struct {
    Category       CategoryMatch
    Confidence     float64
    Reasoning      string
    Alternatives   []Alternative
}

type CategoryMatch struct {
    MainCategory    string
    SubCategory     string
    CategoryPath    string // e.g., "expenses.housing.rent"
    ValidCategory   bool   // Indicates if it matches our predefined categories
}

// Document extraction specific types
type DocumentData struct {
    Content     string
    Type        DocumentType // e.g., Invoice, Receipt, Statement
    Metadata    map[string]interface{}
    PromptType  PromptType
}

type ExtractionResult struct {
    Items       []TransactionItem
    Total       float64
    Confidence  float64
    Metadata    map[string]interface{}
}
```

## Category Management

### Category Structure
```go
type Category struct {
    ID          string
    Path        string      // Hierarchical path (e.g., "expenses.housing.rent")
    Name        string      // Display name
    Description string
    Rules       []Rule      // Category-specific rules
    Examples    []Example   // Example transactions
}

type Rule struct {
    Type        RuleType    // e.g., Pattern, Amount, Date
    Pattern     string      // Regex pattern or rule expression
    Confidence  float64     // Base confidence for this rule
    Description string
}

// CLI commands for category management
type CategoryCommands struct {
    List        func() []Category
    Add         func(Category) error
    Update      func(Category) error
    Delete      func(string) error
    Export      func() ([]byte, error)
    Import      func([]byte) error
}
```

## Prompt Management

### Prompt Types
```go
type PromptType string

const (
    TransactionCategorizationPrompt PromptType = "transaction_categorization"
    InvoiceExtractionPrompt        PromptType = "invoice_extraction"
    ReceiptExtractionPrompt        PromptType = "receipt_extraction"
    StatementExtractionPrompt      PromptType = "statement_extraction"
)

type PromptTemplate struct {
    Type         PromptType
    SystemPrompt string
    UserPrompt   string
    Examples     []Example
    Categories   []Category  // Available categories for this prompt
    Rules        []Rule     // Prompt-specific rules
    Version      string
}

// CLI commands for prompt management
type PromptCommands struct {
    List     func() []PromptTemplate
    Get      func(PromptType) (*PromptTemplate, error)
    Update   func(PromptType, *PromptTemplate) error
    Test     func(PromptType, string) (*AnalysisResponse, error)
}
```

### Example Prompts

```go
const (
    TransactionSystemPrompt = `You are a financial transaction analyzer specializing in Swedish budget categories.
Your task is to analyze transactions and categorize them according to the following strict category hierarchy:

{{range .Categories}}
- {{.Path}}: {{.Description}}
{{- range .Rules}}
  - Rule: {{.Description}}
{{- end}}
{{end}}

Only use categories from this list. Respond in JSON format with category path, confidence score (0-1), and reasoning.`

    ReceiptExtractionPrompt = `Analyze this receipt and extract individual items.
For each item identify:
1. Description
2. Amount
3. Category (using our predefined categories)
4. Quantity if available

Available categories:
{{range .Categories}}
- {{.Path}}: {{.Description}}
{{end}}

Respond with a JSON array of items and metadata.`
)
```

### Response Format
```json
{
  "category_path": "expenses.transport.fuel",
  "confidence": 0.95,
  "reasoning": "Transaction at PREEM, a well-known Swedish gas station chain",
  "alternatives": [
    {
      "category_path": "expenses.misc.retail",
      "confidence": 0.05,
      "reasoning": "Could be in-store purchase"
    }
  ]
}
```

## Processing Pipeline

### Document Processing
1. Text preprocessing
2. OpenAI API call with appropriate prompt
3. Response parsing and validation
4. Confidence threshold checking
5. Human review if needed

### Transaction Analysis
1. Description normalization
2. OpenAI categorization
3. Response validation
4. Cache management
5. Fallback handling

## Document Processing Pipeline

### Document Types and Processing
```go
type DocumentType string

const (
    PDFDocument    DocumentType = "pdf"
    ImageDocument  DocumentType = "image"
    TextDocument   DocumentType = "text"
)

type Document struct {
    ID          string
    Type        DocumentType
    Content     []byte
    ContentType string
    Metadata    map[string]interface{}
}

type DocumentProcessor interface {
    Process(ctx context.Context, doc *Document) (*ProcessedDocument, error)
    Extract(ctx context.Context, doc *ProcessedDocument) (*ExtractionResult, error)
    Store(ctx context.Context, result *ExtractionResult) error
}

type ProcessedDocument struct {
    ID              string
    OriginalDoc     *Document
    ExtractedText   string
    Pages           []Page
    ProcessedAt     time.Time
    ProcessingMeta  map[string]interface{}
}

type Page struct {
    Number      int
    Text        string
    Confidence  float64
    Items       []ExtractedItem
}

type ExtractedItem struct {
    Description     string
    Amount         float64
    Date           time.Time
    Category       CategoryMatch
    Confidence     float64
    Location       TextLocation    // Position in document
    Page           int
    Metadata       map[string]interface{}
}
```

### PDF Processing Pipeline
```go
// Using pdfcpu for PDF processing
type PDFProcessor struct {
    logger        *slog.Logger
    aiService     AIService
    storage       Storage
    validator     Validator
    conf          *model.Configuration
}

// Process extracts text and analyzes content using OpenAI
func (p *PDFProcessor) Process(ctx context.Context, file io.Reader, filename string) (*ProcessingResult, error) {
    // 1. Extract text using pdfcpu
    extractedText, pageCount, err := p.extractText(file)
    if err != nil {
        return nil, &ProcessingError{
            Stage:    StageExtraction,
            Document: filename,
            Err:      err,
        }
    }

    // 2. Send to OpenAI for analysis
    doc := &DocumentData{
        Content:    extractedText,
        Type:      TypePDF,
        Metadata: map[string]interface{}{
            "filename":   filename,
            "page_count": pageCount,
        },
        PromptType: determinePromptType(extractedText),
    }

    result, err := p.aiService.ExtractDocument(ctx, doc)
    if err != nil {
        return nil, &ProcessingError{
            Stage:    StageAnalysis,
            Document: filename,
            Err:      err,
        }
    }

    // 3. Process and validate results
    transactions, err := p.processExtraction(result)
    if err != nil {
        return nil, &ProcessingError{
            Stage:    StageNormalization,
            Document: filename,
            Err:      err,
        }
    }

    // 4. Store results
    if err := p.storage.StoreTransactions(ctx, transactions); err != nil {
        return nil, &ProcessingError{
            Stage:    "storage",
            Document: filename,
            Err:      err,
        }
    }

    return &ProcessingResult{
        Transactions: transactions,
        Metadata: map[string]interface{}{
            "filename":     filename,
            "page_count":   pageCount,
            "text_length":  len(extractedText),
            "processed_at": time.Now(),
        },
    }, nil
}

// Prompt templates for different document types
const (
    BankStatementPrompt = `Analyze this bank statement and extract all transactions.
For each transaction identify:
1. Date (in the format YYYY-MM-DD)
2. Amount (with currency if available)
3. Description/Payee
4. Reference number (if available)
5. Transaction type

The statement text is:
{{.Content}}

Respond with a JSON array of transactions, including confidence scores for each field.
Each transaction should include its location in the text (page number and line number if available).`

    InvoicePrompt = `Analyze this invoice and extract:
1. Invoice number
2. Date
3. Due date
4. Total amount
5. Individual line items with:
   - Description
   - Quantity
   - Unit price
   - Total price
   - VAT (if applicable)

The invoice text is:
{{.Content}}

Respond with a structured JSON including all extracted information and confidence scores.`

    ReceiptPrompt = `Analyze this receipt and extract:
1. Store/Merchant name
2. Date and time
3. Total amount
4. Payment method
5. Individual items with:
   - Description
   - Quantity
   - Unit price
   - Total price
   - Category

The receipt text is:
{{.Content}}

Respond with a structured JSON including all extracted information and confidence scores.`
)

// Configuration for PDF processing
type PDFConfig struct {
    ExtractorConfig *model.Configuration  // pdfcpu configuration
    AIConfig        *OpenAIConfig         // OpenAI configuration
    StorageConfig   *StorageConfig        // Storage configuration
    ValidationRules []ValidationRule      // Validation rules
}

// Validation rules for extracted data
type ValidationRule struct {
    Field     string                    // Field to validate
    Validator func(interface{}) bool    // Validation function
    Required  bool                      // Whether the field is required
    MinScore  float64                  // Minimum confidence score
}
```

### Storage Integration
```go
type Storage interface {
    StoreDocument(ctx context.Context, doc *ProcessedDocument) error
    StoreExtractionResult(ctx context.Context, result *ExtractionResult) error
    StoreTransactions(ctx context.Context, transactions []Transaction) error
}

type Transaction struct {
    ID              string
    DocumentID      string
    Description     string
    Amount          float64
    Date            time.Time
    Category        CategoryMatch
    Confidence      float64
    Source          DocumentType
    Location        TextLocation
    RawText         string
    ProcessingMeta  map[string]interface{}
}
```

### OpenAI Prompts for Document Processing
```go
const (
    PDFExtractionPrompt = `Extract all transactions from this document.
For each transaction, identify:
1. Description/Payee
2. Amount
3. Date
4. Any additional metadata (reference numbers, categories, etc.)

The document text is as follows:
{{.ExtractedText}}

Respond with a JSON array of transactions, including confidence scores for each field.
Each transaction should include its location in the text (page number and approximate position).`

    ReceiptAnalysisPrompt = `Analyze this receipt and extract:
1. Store/Merchant information
2. Date and time
3. Total amount
4. Individual items with:
   - Description
   - Amount
   - Quantity
   - Category (using provided categories)
   - Any discounts or special pricing

Receipt text:
{{.ExtractedText}}

Available categories:
{{range .Categories}}
- {{.Path}}: {{.Description}}
{{end}}

Respond with a structured JSON including all extracted information and confidence scores.`
)
```

### Document Processing Configuration
```go
type ProcessingConfig struct {
    OCRConfig       OCRConfig
    AIConfig        OpenAIConfig
    ValidationRules []ValidationRule
    StorageConfig   StorageConfig
}

type ValidationRule struct {
    Type      string
    Threshold float64
    Required  bool
    Validator func(interface{}) bool
}

type StorageConfig struct {
    DatabaseURL     string
    BatchSize       int
    RetryConfig    RetryConfig
    CacheConfig    CacheConfig
}
```

## Performance Optimization

### Caching Strategy
```go
type Cache interface {
    Get(key string) (*CacheEntry, error)
    Set(key string, entry *CacheEntry) error
    Invalidate(pattern string) error
}

type CacheEntry struct {
    Response     string
    Timestamp    time.Time
    ModelVersion string
    TTL          time.Duration
}
```

### Rate Limiting
```go
type RateLimiter struct {
    TokenBucket *rate.Limiter
    Costs      map[string]int // Token costs per operation
    MaxRetries int
}
```

## Error Handling

### OpenAI Errors
```go
type OpenAIError struct {
    StatusCode  int
    Type       string    // error type from OpenAI
    Message    string
    RequestID  string
    Retryable  bool
}

func (e *OpenAIError) Error() string {
    return fmt.Sprintf("OpenAI API error (%s): %s", e.Type, e.Message)
}
```

### Fallback Strategy
1. Cache lookup
2. Pattern matching
3. Historical data
4. Manual review queue
5. Default categorization

## Monitoring & Analytics

### Metrics
- API response times
- Token usage per request
- Cost per transaction
- Cache hit rates
- Error frequencies
- Categorization accuracy

### Cost Management
```go
type CostTracker struct {
    TokensUsed    int64
    RequestCount  int64
    TotalCost    float64
    LastSync     time.Time
}
```

## Security Considerations

1. API Key Management
   - Environment variables
   - Secure key rotation
   - Access logging

2. Data Privacy
   - PII detection
   - Data minimization
   - Audit logging

3. Request Validation
   - Input sanitization
   - Token limits
   - Response validation

## CLI Integration

```go
type CLI struct {
    Categories CategoryCommands
    Prompts    PromptCommands
    Rules      RuleCommands
    Test       TestCommands
}

type RuleCommands struct {
    Add      func(Rule) error
    Update   func(Rule) error
    Delete   func(string) error
    Test     func(Rule, string) (bool, float64)
}

type TestCommands struct {
    TestPrompt    func(PromptType, string) (*AnalysisResponse, error)
    TestRule      func(Rule, string) (bool, float64)
    TestCategory  func(string, Category) (*CategoryMatch, error)
    BatchTest     func([]string, PromptType) (*BatchTestResult, error)
}
```

## Configuration

```go
type OpenAIConfig struct {
    Model           string        // e.g., "gpt-4-turbo-preview"
    APIKey          string
    OrgID           string        // Optional
    MaxTokens       int
    Temperature     float32
    RetryConfig     RetryConfig
    CacheConfig     CacheConfig
    CostLimits      CostLimits
    DefaultPrompts  map[PromptType]*PromptTemplate
    Categories      []Category
}

// Configuration can be updated via CLI or API
type ConfigCommands struct {
    UpdatePrompt    func(PromptType, *PromptTemplate) error
    UpdateCategory  func(Category) error
    UpdateRules     func([]Rule) error
    Export          func() ([]byte, error)
    Import          func([]byte) error
}
``` 