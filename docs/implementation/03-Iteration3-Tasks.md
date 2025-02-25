# Iteration 3: OpenAI Integration

## Current Focus
Implementing AI capabilities using OpenAI's API for PDF document processing and transaction categorization through a CLI tool. The focus is on extracting transactions from PDFs, bank statements, and CSV files, then categorizing them appropriately, while maintaining robust category and prompt management capabilities.

## Tasks Breakdown

### 1. OpenAI Service Client ✅
- [x] Create AI service interface
  ```go
  type AIService interface {
      ExtractDocument(ctx context.Context, doc *Document) (*Extraction, error)
      AnalyzeTransaction(ctx context.Context, tx *Transaction) (*Analysis, error)
      UpdatePrompt(ctx context.Context, promptType PromptType, content string) error
      AddExample(ctx context.Context, promptType PromptType, example Example) error
  }
  ```
- [x] Implement OpenAI client
  ```go
  type OpenAIService struct {
      client      *openai.Client  // OpenAI's official Go client
      config      AIConfig        // Configuration struct
  }

  type AIConfig struct {
      Model               string        // e.g., "gpt-4-turbo-preview"
      APIKey             string        // OpenAI API key
      MaxTokens          int           // Maximum tokens per request
      Temperature        float32       // Response randomness (0-1)
      RequestTimeout     time.Duration // Timeout for API calls
  }
  ```
- [x] Add error handling
  - Implemented error type detection
  - Created error type definitions
- [x] Set up API key management
  - Using environment variables
  - Added secure configuration handling

### 2. Transaction Categorization with OpenAI ✅
- [x] Implement prompt engineering
  ```go
  type PromptTemplate struct {
      SystemPrompt    string   // Context and instructions for the model
      UserPrompt      string   // Transaction-specific prompt
      Examples        []Example // Example transactions for better accuracy
  }
  ```
- [x] Create category matching logic
  - Structured JSON response parsing
  - Category validation
  - Main and sub-category extraction
- [x] Add confidence scoring
  - Using OpenAI's confidence scores
  - Added validation thresholds
- [x] Implement transaction source handling
  - PDF document extraction
  - Bank statement processing
  - CSV file processing

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
      Examples     []Example
      Categories   []Category
      Version      string
      IsActive     bool
  }
  ```
- [x] Create prompt versioning
  - [x] Version tracking
  - [x] Auto-increment version on updates
- [x] Add prompt testing
  - [x] Comprehensive test coverage
  - [x] Error case validation
- [x] Implement prompt validation
  - [x] Input validation
  - [x] Active/inactive state handling
- [x] Create example management
  - [x] Example storage and retrieval
  - [x] Example validation
  - [x] Proper error handling
- [x] Implement default prompts
  - [x] Transaction categorization prompt
  - [x] Invoice extraction prompt
  - [x] Receipt extraction prompt
  - [x] Bank statement prompt
  - [x] CSV transaction prompt

### 5. CLI Tool Development 🔄
- [x] Create category management commands
  - [x] List categories with table/JSON output
  - [x] Add new categories with validation
  - [x] Update existing categories
  - [x] Soft delete with confirmation
- [ ] Add document processing commands
  - [ ] Process PDF command
  - [ ] Process bank statement command
  - [ ] Process CSV command
  - [ ] Extract transactions command
  - [ ] Categorize transactions command
- [ ] Add prompt management commands
  - [ ] List prompts
  - [ ] Update prompts
  - [ ] Add examples to prompts
- [ ] Implement progress tracking
  - [ ] Processing status display
  - [ ] Error reporting
- [ ] Add configuration management
  - [ ] OpenAI configuration
  - [ ] PDF processor configuration

### 6. Document Analysis Pipeline 🔄
- [x] Implement PDF processing with pdfcpu
  ```go
  type PDFProcessor struct {
      logger        *slog.Logger
      aiService     AIService
      conf          *Configuration
  }
  ```
- [x] Create text extraction service
  - [x] PDF text extraction using pdfcpu
  - [x] CSV parsing and validation
  - [x] Bank statement parsing
  - [x] Basic validation
  - [x] Error handling
- [ ] Enhance document analysis
  - [ ] Document type detection
  - [ ] Smart prompt selection
  - [ ] Multi-page handling
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

## Integration Points
- [x] Transaction processing from Iteration 2
- [x] Database models from Iteration 1
- [ ] CLI integration
- [ ] Document processing pipeline

## Review Checklist
- [x] OpenAI service operational
- [x] Basic categorization working
- [x] Category management implemented
- [x] Prompt management working
- [ ] CLI tool functional
- [ ] PDF processing implemented
- [ ] Document extraction working
- [ ] Bank statement processing working
- [ ] CSV processing working
- [x] Documentation updated
- [x] Test coverage > 80%

## Success Criteria
1. [ ] Categorization accuracy > 90% (using GPT-4)
2. [ ] PDF processing accuracy > 85%
3. [ ] Bank statement processing accuracy > 95%
4. [ ] CSV processing accuracy > 95%
5. [ ] Document processing time < 30 seconds
6. [ ] CLI commands implemented and tested
7. [x] Category management working
8. [x] Prompt management working
9. [x] Test coverage > 80%

## Technical Considerations

### OpenAI Configuration ✅
```go
type OpenAIConfig struct {
    Model               string        // GPT model to use
    APIKey             string        // OpenAI API key
    MaxTokens          int           // Token limit per request
    Temperature        float32       // Response randomness
    RequestTimeout     time.Duration // API timeout
}
```

### Error Handling ✅
```go
type OpenAIError struct {
    StatusCode   int
    Message      string
    Type        string    // error type from OpenAI
    RequestID    string
}
```

### Document Processing Configuration ✅
```go
type PDFConfig struct {
    ExtractorConfig *Configuration  // pdfcpu configuration
    AIConfig        *OpenAIConfig   // OpenAI configuration
}
```

### Monitoring 🔄
- [x] Response times
- [x] Error rates
- [ ] Document processing times
- [ ] Extraction accuracy
- [ ] PDF extraction success rate

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
- [x] Consider privacy implications
- [ ] Document CLI usage
- [x] Create example prompts
- [ ] Test with various document types
- [ ] Test with various PDF formats
- [ ] Implement proper error handling for PDF processing
- [ ] Plan for large document handling 