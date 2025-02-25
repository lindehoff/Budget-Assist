# OpenAI Integration Design

## Overview
The system leverages OpenAI's GPT models for PDF document processing and transaction categorization. The CLI tool provides functionality to extract text from PDFs (invoices and receipts), analyze the content to identify transactions, and categorize these transactions using OpenAI's intelligent processing capabilities. The system also includes robust category and prompt management to allow users to customize and improve the AI's performance over time.

## Core Components

### OpenAI Service Interface

```go
type AIService interface {
    ExtractDocument(ctx context.Context, doc DocumentData) (*ExtractionResult, error)
    AnalyzeTransaction(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error)
    UpdatePrompt(ctx context.Context, promptType PromptType, content string) error
    AddExample(ctx context.Context, promptType PromptType, example Example) error
}

type DocumentData struct {
    Content     string
    Type        DocumentType // e.g., Invoice, Receipt
    Metadata    map[string]interface{}
    PromptType  PromptType
}

type ExtractionResult struct {
    Items       []TransactionItem
    Total       float64
    Confidence  float64
    Metadata    map[string]interface{}
}

type AnalysisRequest struct {
    Description      string
    Amount          float64
    Date            time.Time
    Metadata        map[string]interface{}
    PromptType      PromptType
}

type AnalysisResponse struct {
    Category       CategoryMatch
    Confidence     float64
    Reasoning      string
}

type CategoryMatch struct {
    MainCategory    string
    SubCategory     string
    CategoryPath    string // e.g., "expenses.housing.rent"
    ValidCategory   bool   // Indicates if it matches our predefined categories
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
    Examples    []Example   // Example transactions
}

type Example struct {
    Description string
    Amount     float64
    Date       time.Time
    Category   string
    Metadata   map[string]interface{}
}

// CLI commands for category management
type CategoryCommands struct {
    List        func() []Category
    Add         func(Category) error
    Update      func(Category) error
    Delete      func(string) error
    Export      func() ([]byte, error)
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
    BankStatementPrompt            PromptType = "bank_statement"
    CSVTransactionPrompt           PromptType = "csv_transaction"
)

type PromptTemplate struct {
    Type         PromptType
    SystemPrompt string
    UserPrompt   string
    Examples     []Example
    Categories   []Category  // Available categories for this prompt
    Version      string
}

// CLI commands for prompt management
type PromptCommands struct {
    List     func() []PromptTemplate
    Get      func(PromptType) (*PromptTemplate, error)
    Update   func(PromptType, *PromptTemplate) error
    AddExample func(PromptType, Example) error
}
```

### Example Prompts

```go
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

Respond with a JSON array of transactions.`

    CSVTransactionPrompt = `Analyze these CSV transactions and extract the relevant information.
For each transaction identify:
1. Date
2. Amount
3. Description/Payee
4. Type (income/expense)

The CSV content is:
{{.Content}}

Respond with a JSON array of transactions.`

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

Respond with a structured JSON including all extracted information.`

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

The receipt text is:
{{.Content}}

Respond with a structured JSON including all extracted information.`

    TransactionCategorizationPrompt = `You are a financial transaction analyzer. Your task is to analyze transactions and categorize them according to the following category hierarchy:

{{range .Categories}}
- {{.Path}}: {{.Description}}
{{end}}

{{if .Examples}}
Here are some example transactions that show how to categorize similar items:
{{range .Examples}}
Description: {{.Description}}
Amount: {{.Amount}}
Category: {{.Category}}
{{end}}
{{end}}

Transaction details:
Description: {{.Description}}
Amount: {{.Amount}}
Date: {{.Date}}

Only use categories from the provided list. Respond with a JSON containing:
1. category_path: The full path of the chosen category
2. main_category: The top-level category
3. sub_category: The specific sub-category
4. confidence: Your confidence in this categorization (0-1)
5. reasoning: A brief explanation of why this category was chosen`
)
```

## Document Processing Pipeline

### Document Types and Processing
```go
type DocumentType string

const (
    PDFDocument    DocumentType = "pdf"
    TextDocument   DocumentType = "text"
    CSVDocument    DocumentType = "csv"
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
}

type ProcessedDocument struct {
    ID              string
    OriginalDoc     *Document
    ExtractedText   string
    ProcessedAt     time.Time
    ProcessingMeta  map[string]interface{}
}

type ExtractedItem struct {
    Description     string
    Amount         float64
    Date           time.Time
    Category       CategoryMatch
    Confidence     float64
    Page           int
    Metadata       map[string]interface{}
}
```

### PDF Processing Pipeline
```go
type PDFProcessor struct {
    logger        *slog.Logger
    aiService     AIService
    conf          *Configuration
}

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
```

## CLI Integration

```go
type CLI struct {
    Process   ProcessCommands
    Category  CategoryCommands
    Prompt    PromptCommands
}

type ProcessCommands struct {
    ExtractPDF func(filepath string) (*ProcessingResult, error)
    Categorize func(transactions []Transaction) ([]CategorizedTransaction, error)
}
```

## Configuration

```go
type OpenAIConfig struct {
    Model           string        // e.g., "gpt-4-turbo-preview"
    APIKey          string
    MaxTokens       int
    Temperature     float32
    DefaultPrompts  map[PromptType]*PromptTemplate
    Categories      []Category
}

type PDFConfig struct {
    ExtractorConfig *Configuration  // pdfcpu configuration
    AIConfig        *OpenAIConfig   // OpenAI configuration
}
```

## Error Handling

```go
type OpenAIError struct {
    StatusCode  int
    Type       string    // error type from OpenAI
    Message    string
    RequestID  string
}

func (e *OpenAIError) Error() string {
    return fmt.Sprintf("OpenAI API error (%s): %s", e.Type, e.Message)
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