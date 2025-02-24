# Iteration 3: OpenAI Integration

## Current Focus
Implementing AI capabilities using OpenAI's API for intelligent transaction categorization and document information extraction, with support for multiple use cases and configurable prompts.

## Tasks Breakdown

### 1. OpenAI Service Client ✅
- [x] Create AI service interface
  ```go
  type AIService interface {
      AnalyzeTransaction(ctx context.Context, tx *Transaction) (*Analysis, error)
      ExtractDocument(ctx context.Context, doc *Document) (*Extraction, error)
      SuggestCategories(ctx context.Context, desc string) ([]CategoryMatch, error)
  }
  ```
- [x] Implement OpenAI client
  ```go
  type OpenAIService struct {
      client      *openai.Client  // OpenAI's official Go client
      config      AIConfig        // Configuration struct
      rateLimiter *RateLimiter   // Rate limiter for API calls
  }

  type AIConfig struct {
      Model               string        // e.g., "gpt-4-turbo-preview"
      APIKey             string        // OpenAI API key
      MaxTokens          int           // Maximum tokens per request
      Temperature        float32       // Response randomness (0-1)
      RequestTimeout     time.Duration // Timeout for API calls
      RetryAttempts     int           // Number of retry attempts
  }
  ```
- [x] Add retry and fallback logic
  - Implemented exponential backoff
  - Added configurable retry settings
  - Created error type detection
- [x] Set up API key management
  - Using environment variables
  - Added secure configuration handling
- [x] Create request rate limiting
  - Implemented token bucket algorithm
  - Added configurable limits
  - Created burst handling

### 2. Transaction Categorization with OpenAI ✅
- [x] Implement prompt engineering
  ```go
  type PromptTemplate struct {
      SystemPrompt    string   // Context and instructions for the model
      UserPrompt      string   // Transaction-specific prompt
      FewShotExamples []Example // Example transactions for better accuracy
  }
  ```
- [x] Create category matching logic
  - Structured JSON response parsing
  - Confidence score extraction
  - Category validation
- [x] Add confidence scoring
  - Using OpenAI's confidence scores
  - Added validation thresholds
- [x] Create category suggestions
  - Multiple category support
  - Confidence-based ranking

### 3. Category Management System ✅
- [x] Implement category structure
  ```go
  type Category struct {
      ID                 uint
      TypeID            uint
      Name              string
      Description       string
      IsActive          bool
      InstanceIdentifier string
      Translations      []Translation
  }

  type Manager struct {
      store     db.Store
      aiService ai.Service
      logger    *slog.Logger
  }
  ```
- [x] Create category validation
  - Input validation for required fields
  - Type ID validation
  - Name uniqueness checks
- [x] Add category translations
  - Multi-language support
  - Translation management
  - Language code validation
- [x] Implement category-specific error handling
  ```go
  type CategoryError struct {
      Operation string
      Category  string
      Err       error
  }
  ```
- [x] Add comprehensive testing
  - Table-driven tests
  - Error case validation
  - Translation testing
  - Mock implementations

### 4. Prompt Management System ✅
- [x] Implement prompt types
  ```go
  type PromptTemplate struct {
      Type         PromptType
      SystemPrompt string
      UserPrompt   string
      Categories   []Category
      Rules        []Rule
      Version      string
  }
  ```
- [x] Create prompt versioning
- [x] Add prompt testing
- [x] Implement prompt validation
- [x] Create example management

### 5. CLI Tool Development 🔄
- [ ] Create category management commands
- [ ] Add prompt management interface
- [ ] Implement testing commands
- [ ] Add configuration management
- [ ] Create import/export functionality

### 6. Response Management 🔄
- [ ] Design response caching
- [ ] Implement response validation
- [ ] Add cost tracking
- [ ] Create usage analytics
- [ ] Implement rate limit monitoring

### 7. Document Analysis Pipeline 🔄
- [x] Implement PDF processing with pdfcpu
  ```go
  type PDFProcessor struct {
      logger        *slog.Logger
      aiService     AIService
      storage       Storage
      validator     Validator
      conf          *model.Configuration
  }
  ```
- [x] Create text extraction service
  - [x] PDF text extraction using pdfcpu
  - [x] Basic validation
  - [x] Error handling
- [ ] Enhance document analysis
  - [ ] Document type detection
  - [ ] Smart prompt selection
  - [ ] Multi-page handling
  - [ ] OCR integration for scanned documents
- [ ] Implement OpenAI integration
  - [ ] Document-specific prompts
  - [ ] Response parsing
  - [ ] Transaction extraction
  - [ ] Confidence scoring
- [ ] Add validation pipeline
  - [ ] Amount validation
  - [ ] Date validation
  - [ ] Category validation
  - [ ] Confidence thresholds
- [ ] Enhance storage integration
  - [ ] Document storage
  - [ ] Extraction results
  - [ ] Transaction metadata

### 8. Document Processing CLI 🔄
- [ ] Add document processing commands
  ```go
  type DocumentCommands struct {
      Process  func(filepath string) error
      List     func() []ProcessedDocument
      Export   func(id string) error
      Status   func(id string) *ProcessingStatus
  }
  ```
- [ ] Implement batch processing
- [ ] Add progress tracking
- [ ] Create export functionality
- [ ] Add validation commands

### 9. Storage Integration 🔄
- [ ] Implement document storage
  ```go
  type Storage interface {
      StoreDocument(ctx context.Context, doc *ProcessedDocument) error
      StoreExtractionResult(ctx context.Context, result *ExtractionResult) error
      StoreTransactions(ctx context.Context, transactions []Transaction) error
  }
  ```
- [ ] Add transaction storage
- [ ] Create metadata storage
- [ ] Implement batch operations
- [ ] Add data validation

## Integration Points
- [x] Transaction processing from Iteration 2
- [x] Database models from Iteration 1
- [ ] CLI integration
- [ ] Document processing pipeline
- [ ] Storage integration
- [ ] Preparing for API integration in Iteration 4

## Review Checklist
- [x] OpenAI service operational
- [x] Basic categorization working
- [x] Category management implemented
- [ ] CLI tool functional
- [ ] Prompt management working
- [ ] Response caching implemented
- [ ] Fallbacks tested
- [x] Documentation updated
- [ ] Performance metrics collected
- [ ] PDF processing implemented
- [ ] Document extraction working
- [ ] Storage integration complete
- [ ] Batch processing functional
- [ ] Export functionality working

## Success Criteria
1. [ ] Categorization accuracy > 90% (using GPT-4)
2. [x] Response time < 1 second (with caching)
3. [ ] Fallback success rate > 95%
4. [ ] Cost per transaction < $0.01
5. [x] Test coverage > 80%
6. [ ] CLI commands implemented and tested
7. [x] Category management working
8. [ ] Prompt management operational
9. [ ] PDF processing accuracy > 85%
10. [ ] Document processing time < 30 seconds
11. [ ] Storage integration tested
12. [ ] Batch processing working

## Technical Considerations

### OpenAI Configuration ✅
```go
type OpenAIConfig struct {
    Model               string        // GPT model to use
    APIKey             string        // OpenAI API key
    OrgID              string        // Optional organization ID
    MaxTokens          int           // Token limit per request
    Temperature        float32       // Response randomness
    RequestTimeout     time.Duration // API timeout
    RetryConfig       RetryConfig   // Retry settings
}
```

### Error Handling ✅
```go
type OpenAIError struct {
    StatusCode   int
    Message      string
    Type        string    // error type from OpenAI
    RetryCount   int
    RequestID    string
}
```

### Document Processing Configuration ✅
```go
type PDFConfig struct {
    ExtractorConfig *model.Configuration  // pdfcpu configuration
    AIConfig        *OpenAIConfig         // OpenAI configuration
    StorageConfig   *StorageConfig        // Storage configuration
    ValidationRules []ValidationRule      // Validation rules
}
```

### Monitoring 🔄
- [x] Response times
- [x] Token usage
- [x] API costs
- [x] Error rates
- [x] Cache hit rates
- [ ] Document processing times
- [ ] Extraction accuracy
- [ ] Storage performance
- [ ] Batch processing metrics
- [ ] PDF extraction success rate
- [ ] OCR accuracy metrics
- [ ] Storage performance

### Category Management ✅
```go
// Store interface for database operations
type Store interface {
    CreateCategory(ctx context.Context, category *Category) error
    UpdateCategory(ctx context.Context, category *Category) error
    GetCategoryByID(ctx context.Context, id uint) (*Category, error)
    ListCategories(ctx context.Context, typeID *uint) ([]Category, error)
    GetCategoryTypeByID(ctx context.Context, id uint) (*CategoryType, error)
    CreateTranslation(ctx context.Context, translation *Translation) error
    GetTranslations(ctx context.Context, entityID uint, entityType string) ([]Translation, error)
    DeleteCategory(ctx context.Context, id uint) error
}

// Request/Response types
type CreateCategoryRequest struct {
    Name               string
    Description        string
    TypeID            uint
    InstanceIdentifier string
    Translations      map[string]TranslationData
}

type CategorySuggestion struct {
    CategoryPath string
    Confidence  float64
}
```

### Testing Standards ✅
- Removed testify dependency
- Using standard library testing
- Table-driven tests with descriptive names
- Proper error validation with errors.As and errors.Is
- Comprehensive test coverage
- Mock implementations for external dependencies
- Helper functions for test setup
- Clear validation messages

## Notes
- [x] Use OpenAI's official Go client
- [x] Monitor API costs carefully
- [x] Implement proper rate limiting
- [x] Consider privacy implications
- [ ] Plan for model updates
- [ ] Document CLI usage
- [ ] Create example prompts
- [ ] Test with various document types
- [ ] Test with various PDF formats
- [ ] Implement proper error handling for PDF processing
- [ ] Consider OCR requirements
- [ ] Plan for large document handling
- [ ] Document storage requirements
- [x] PDF extraction using pdfcpu implemented
- [ ] Add OCR support for scanned documents
- [ ] Implement smart prompt selection
- [ ] Add multi-page document handling
- [ ] Create document type detection
- [ ] Enhance error handling for PDF processing 