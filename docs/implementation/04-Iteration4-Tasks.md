# Iteration 4: Web Server & API

## Current Focus
Creating a robust web server and REST API for accessing budget data and services.

## Tasks Breakdown

### 1. Web Server Setup
- [ ] Initialize web framework (Echo)
  ```go
  type Server struct {
      router     *echo.Echo
      db         *gorm.DB
      ai         AIService
      logger     *slog.Logger
      metrics    MetricsCollector
      config     ServerConfig
  }
  ```
- [ ] Configure middleware
  - [ ] CORS
  - [ ] Request logging
  - [ ] Panic recovery
  - [ ] Request ID tracking
- [ ] Set up health checks
- [ ] Configure TLS
- [ ] Implement graceful shutdown

### 2. Core API Endpoints
- [ ] Implement transaction endpoints
  ```go
  // Transaction routes
  POST   /api/v1/transactions
  GET    /api/v1/transactions
  GET    /api/v1/transactions/:id
  PATCH  /api/v1/transactions/:id
  DELETE /api/v1/transactions/:id
  ```
- [ ] Add category management
  ```go
  // Category routes
  GET    /api/v1/categories
  POST   /api/v1/categories
  PATCH  /api/v1/categories/:id
  GET    /api/v1/categories/:id/transactions
  ```
- [ ] Create analysis endpoints
  ```go
  // Analysis routes
  GET    /api/v1/analysis/monthly
  GET    /api/v1/analysis/category
  GET    /api/v1/analysis/trends
  ```
- [ ] Implement search functionality
- [ ] Add bulk operations support

### 3. Authentication System
- [ ] Implement JWT authentication
  ```go
  type Claims struct {
      UserID    int64     `json:"uid"`
      Role      string    `json:"role"`
      IssuedAt  time.Time `json:"iat"`
      ExpiresAt time.Time `json:"exp"`
  }
  ```
- [ ] Add refresh token logic
- [ ] Create user management
- [ ] Implement role-based access
- [ ] Add session management

### 4. Rate Limiting
- [ ] Implement rate limiter middleware
  ```go
  type RateLimiter struct {
      Store      RedisStore
      WindowSize time.Duration
      MaxRequest int
      KeyFunc    func(*echo.Context) string
  }
  ```
- [ ] Add per-endpoint limits
- [ ] Create burst handling
- [ ] Implement user quotas
- [ ] Add rate limit headers

### 5. API Documentation
- [ ] Set up Swagger/OpenAPI
- [ ] Document all endpoints
- [ ] Add request/response examples
- [ ] Create API usage guide
- [ ] Document error responses

## Integration Points
- AI service from Iteration 3
- Transaction processing from Iteration 2
- Database models from Iteration 1

## Review Checklist
- [ ] All endpoints tested
- [ ] Authentication working
- [ ] Rate limiting effective
- [ ] Documentation complete
- [ ] Security review passed
- [ ] Performance tested

## Success Criteria
1. API response time < 200ms
2. Authentication working
3. Rate limiting effective
4. All core endpoints implemented
5. API documentation complete

## Technical Considerations

### Request Validation
```go
type TransactionRequest struct {
    Amount      decimal.Decimal `json:"amount" validate:"required"`
    Date        time.Time      `json:"date" validate:"required"`
    Description string         `json:"description" validate:"required"`
    CategoryID  int64         `json:"category_id" validate:"required"`
}
```

### Response Formats
```go
type APIResponse struct {
    Status   string      `json:"status"`
    Data     interface{} `json:"data,omitempty"`
    Error    *APIError   `json:"error,omitempty"`
    Metadata *Metadata   `json:"metadata,omitempty"`
}
```

### Monitoring
- Request latency
- Error rates
- Authentication failures
- Rate limit hits
- Active sessions

## Notes
- Follow REST best practices
- Implement proper versioning
- Document all error codes
- Consider API backwards compatibility
- Plan for future scaling 

# Iteration 4: Enhanced CLI Document Processing

## Overview
This iteration focuses on enhancing the CLI document processing functionality to support multiple document types and runtime user insights. The goal is to create a flexible document processing pipeline that can handle various financial documents while allowing users to provide context-specific insights during processing.

## Tasks

### 1. Core Infrastructure Updates

#### 1.1 Process Command Updates
- [ ] Update `process.go` to support new document types flag
- [ ] Add transaction insights flag for document-specific rules
- [ ] Add category insights flag for categorization rules
- [ ] Add support for directory processing
- [ ] Implement progress reporting

Example:
```go
var processCmd = &cobra.Command{
    Use:   "process [path]",
    Short: "Process documents for transaction extraction",
    Args:  cobra.ExactArgs(1),
}

func init() {
    processCmd.Flags().String("doc-type", "", "Document type (bankstatement, bill, receipt)")
    processCmd.Flags().String("transaction-insights", "", "Additional context for transaction extraction")
    processCmd.Flags().String("category-insights", "", "Additional context for categorization")
    processCmd.MarkFlagRequired("doc-type")
}
```

#### 1.2 Processing Pipeline
- [ ] Create new `ProcessOptions` struct with insights fields
- [ ] Update pipeline to handle both files and directories
- [ ] Implement file type detection
- [ ] Add validation for document types and insights

### 2. Document Processors

#### 2.1 Text Extraction Layer
- [ ] Create `TextExtractor` interface
- [ ] Implement PDF text extraction
- [ ] Add image text extraction (OCR)
- [ ] Add text cleanup utilities

#### 2.2 CSV Processing
- [ ] Update SEB processor for new pipeline
- [ ] Add validation for CSV format
- [ ] Implement direct transaction mapping
- [ ] Add error handling

#### 2.3 Document Analysis
- [ ] Create transaction extraction service
- [ ] Add support for different document types
- [ ] Implement metadata extraction
- [ ] Add validation for extracted data

### 3. AI Integration

#### 3.1 Prompt Enhancement
- [ ] Update prompt templates to include runtime insights
- [ ] Add validation for insight format
- [ ] Implement insight merging with base prompts
- [ ] Add confidence scoring

#### 3.2 Transaction Analysis
- [ ] Update AI service interface for insights
- [ ] Implement document-specific analysis
- [ ] Add transaction validation
- [ ] Implement error handling

#### 3.3 Categorization
- [ ] Update categorization with insights
- [ ] Implement batch categorization
- [ ] Add category validation
- [ ] Add confidence thresholds

### 4. Testing and Documentation

#### 4.1 Unit Tests
- [ ] Add tests for process command
- [ ] Add tests for pipeline
- [ ] Add tests for text extraction
- [ ] Add tests for AI integration

#### 4.2 Integration Tests
- [ ] Add end-to-end tests
- [ ] Test directory processing
- [ ] Test different document types
- [ ] Test with various insights

#### 4.3 Documentation
- [ ] Update CLI documentation
- [ ] Add examples for each document type
- [ ] Document insight format and usage
- [ ] Add troubleshooting guide

## Dependencies
- Tesseract OCR for image processing
- OpenAI API for text analysis
- SQLite for data storage
- Cobra for CLI framework

## Timeline
- Process Command Updates: 1 day
- Pipeline Implementation: 2 days
- Document Processors: 3 days
- AI Integration: 2 days
- Testing and Documentation: 2 days

Total: 10 working days

## Example Usage

```bash
# Process a single receipt with insights
budget-assist process receipt.pdf \
  --doc-type receipt \
  --transaction-insights "Items starting with 'REA' are discounts" \
  --category-insights "If store is 'ICA' categorize as Groceries"

# Process a directory of bank statements
budget-assist process ./statements \
  --doc-type bankstatement \
  --category-insights "Transactions from 'SWISH' are transfers"
```

## Risks and Mitigation
1. **OCR Quality**: Test with various image types, add preprocessing
2. **AI Costs**: Implement caching and rate limiting
3. **Performance**: Add batch processing for large directories
4. **Insight Validation**: Add clear validation and error messages 