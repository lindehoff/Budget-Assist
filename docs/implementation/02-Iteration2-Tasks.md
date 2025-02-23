# Iteration 2: Document Processing

## Current Focus
Implementing core document processing capabilities for handling various financial document formats.

## Tasks Breakdown

### 1. PDF Processing
- [ ] Set up PDF text extraction library
  ```go
  // Evaluate and implement one of:
  // - pdfcpu
  // - unidoc
  // - pdftotext wrapper
  ```
- [ ] Create PDF processor interface
- [ ] Implement text extraction
- [ ] Add layout analysis
- [ ] Create test suite with sample documents

### 2. CSV Processing
- [ ] Define CSV processor interface
- [ ] Implement bank-specific parsers
  - [ ] Swedbank format
  - [ ] Nordea format
  - [ ] Generic CSV format
- [ ] Add header detection
- [ ] Create column mapping system
- [ ] Implement validation rules

### 3. Transaction Model
- [ ] Implement core transaction struct
  ```go
  type Transaction struct {
      ID              int64
      Amount          decimal.Decimal
      Date            time.Time
      Description     string
      Category        *Category
      RawData         map[string]interface{}
      Source          string
      ProcessedAt     time.Time
  }
  ```
- [ ] Add validation rules
- [ ] Create data normalizers
- [ ] Implement storage layer
- [ ] Write model tests

### 4. File Detection System
- [ ] Create file type detector
- [ ] Implement MIME type checking
- [ ] Add content analysis
- [ ] Create processor router
- [ ] Add new file type handlers

### 5. Error Handling & Logging
- [ ] Set up structured logging
- [ ] Implement error types
  ```go
  type ProcessingError struct {
      Stage     string
      Filename  string
      Cause     error
      Raw       []byte
  }
  ```
- [ ] Add error recovery
- [ ] Create error reporting
- [ ] Implement retry logic

## Integration Points
- Database schema from Iteration 1
- CLI commands from Iteration 1
- Preparing for AI integration in Iteration 3

## Review Checklist
- [ ] All processors working
- [ ] Error handling complete
- [ ] Tests passing
- [ ] Documentation updated
- [ ] Performance metrics collected
- [ ] Security review done

## Success Criteria
1. Successfully process PDF bank statements
2. Handle multiple CSV formats
3. Accurate transaction extraction
4. Robust error handling
5. Test coverage > 80%

## Notes
- Keep processing modular for future formats
- Document format specifications
- Consider performance with large files 