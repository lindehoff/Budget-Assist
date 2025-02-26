# OpenAI Integration Design

## Overview
The system leverages OpenAI's GPT models for document processing and transaction categorization. The CLI tool provides functionality to:
1. Extract text from documents (PDFs and images)
2. Analyze the content to identify transactions based on document type (bill, receipt, bank statement)
3. Categorize transactions using predefined categories and runtime user insights
4. Allow users to provide context-specific insights during processing

## Core Components

### OpenAI Service Interface
```go
type AIService interface {
    ExtractTransactions(ctx context.Context, text string, opts ExtractionOptions) ([]Transaction, error)
    CategorizeTransaction(ctx context.Context, tx Transaction, insights string) (*CategoryMatch, error)
    GetPrompt(ctx context.Context, promptType PromptType) (*Prompt, error)
}

type ExtractionOptions struct {
    DocumentType string       // "bill", "receipt", "bankstatement"
    Insights    string       // Runtime user insights for extraction
    Metadata    map[string]any
}

type Transaction struct {
    Date        time.Time
    Amount      decimal.Decimal
    Description string
    RawText     string
    Category    *CategoryMatch
    Metadata    map[string]any
}

type CategoryMatch struct {
    MainCategory string
    SubCategory  string
    Confidence   float64
    Reasoning    string
}
```

## Prompt Management

### Prompt Types and Structure
```go
type PromptType string

const (
    BillAnalysisPrompt          PromptType = "bill_analysis"
    ReceiptAnalysisPrompt       PromptType = "receipt_analysis"
    BankStatementAnalysisPrompt PromptType = "bank_statement_analysis"
    TransactionCategorizationPrompt PromptType = "transaction_categorization"
)

type Prompt struct {
    Type           PromptType
    BasePrompt     string    // The core prompt template
    Examples       []Example // Example inputs and outputs
    Categories     []Category // Available categories for this prompt
    Version        string
}

// Note: User insights are not stored in prompts, they are provided at runtime
```

### Example Base Prompts with Runtime Insights

```go
const (
    BillAnalysisBasePrompt = `Analyze this bill and extract all transactions.
For each transaction identify:
1. Date (in YYYY-MM-DD format)
2. Amount
3. Description
4. Any additional metadata (reference numbers, invoice numbers, etc)

Additional context from user for this specific document:
{{.RuntimeInsights}}

The bill text is:
{{.Content}}

Respond with a JSON array of transactions.`

    ReceiptAnalysisBasePrompt = `Analyze this receipt and extract all transactions.
For each transaction identify:
1. Date (in YYYY-MM-DD format)
2. Amount
3. Description/Item
4. Quantity (if applicable)
5. Unit price (if applicable)

Additional context from user for this specific document:
{{.RuntimeInsights}}

The receipt text is:
{{.Content}}

Respond with a JSON array of transactions.`

    TransactionCategorizationBasePrompt = `Categorize this transaction according to our predefined categories:

{{range .Categories}}
- {{.Path}}: {{.Description}}
{{end}}

Additional categorization rules for this specific transaction:
{{.RuntimeInsights}}

Transaction details:
Description: {{.Description}}
Amount: {{.Amount}}
Date: {{.Date}}

Respond with a JSON containing:
1. main_category: The top-level category
2. sub_category: The specific sub-category
3. confidence: Your confidence in this categorization (0-1)
4. reasoning: A brief explanation of why this category was chosen`
)
```

## Processing Pipeline

### Document Analysis Flow
1. User provides document(s) with optional insights
2. System determines document type and extracts text
3. Base prompt is loaded from database
4. Runtime insights are merged into prompt template
5. AI analyzes text with enhanced prompt
6. Results are validated and processed
7. Each transaction is categorized using category insights
8. Results are stored in database

### Example Processing with Insights
```go
func (p *Pipeline) processDocument(ctx context.Context, doc Document, opts ProcessOptions) error {
    // 1. Load base prompt
    prompt, err := p.aiService.GetPrompt(ctx, getPromptType(opts.DocumentType))
    if err != nil {
        return err
    }

    // 2. Extract text
    text, err := p.textExtractor.Extract(ctx, doc.Path)
    if err != nil {
        return err
    }

    // 3. Extract transactions with user insights
    transactions, err := p.aiService.ExtractTransactions(ctx, text, ExtractionOptions{
        DocumentType: opts.DocumentType,
        Insights:    opts.TransactionInsights,
    })
    if err != nil {
        return err
    }

    // 4. Categorize each transaction with category insights
    for _, tx := range transactions {
        category, err := p.aiService.CategorizeTransaction(ctx, tx, opts.CategoryInsights)
        if err != nil {
            return err
        }
        tx.Category = category
    }

    // 5. Store results
    return p.store.SaveTransactions(ctx, transactions)
}
```

## CLI Commands for Prompt Management

```go
type PromptManager interface {
    // Get current prompt
    Get(ctx context.Context, promptType PromptType) (*Prompt, error)
    
    // Update base prompt
    Update(ctx context.Context, promptType PromptType, basePrompt string) error
    
    // Add example
    AddExample(ctx context.Context, promptType PromptType, example Example) error
    
    // Add user insight
    AddInsight(ctx context.Context, promptType PromptType, insight string) error
    
    // List all prompts
    List(ctx context.Context) ([]Prompt, error)
    
    // Export prompts for backup/sharing
    Export(ctx context.Context) ([]byte, error)
    
    // Import prompts
    Import(ctx context.Context, data []byte) error
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