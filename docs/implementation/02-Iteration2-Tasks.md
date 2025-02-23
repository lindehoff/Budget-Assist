# Iteration 2: Document Processing

## Current Focus
Implementing core document processing capabilities for handling various financial document formats.

## Tasks Breakdown

### 1. PDF Processing ✅
- [x] Set up PDF text extraction library
  ```go
  // Using pdfcpu for PDF processing
  import (
    "github.com/pdfcpu/pdfcpu/pkg/api"
    "github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
  )
  ```
- [x] Create PDF processor interface
- [x] Implement text extraction
- [x] Add layout analysis
- [x] Create test suite with sample documents
  - Basic test structure implemented
  - TODO: Add real PDF samples for integration tests

### 2. CSV Processing 🔄
- [ ] Define CSV processor interface
  ```go
  // Will implement DocumentProcessor interface
  type CSVProcessor struct {
      logger *slog.Logger
  }
  ```
- [ ] Implement bank-specific parsers
  - [ ] Swedbank format
  - [ ] Nordea format
  - [ ] Generic CSV format
- [ ] Add header detection
- [ ] Create column mapping system
- [ ] Implement validation rules

### 3. Transaction Model ✅
- [x] Implement core transaction struct
  ```go
  type Transaction struct {
      Date        time.Time
      Amount      decimal.Decimal
      RawData     map[string]any
      Description string
      Category    string
      SubCategory string
      Source      string
  }
  ```
- [x] Add validation rules
- [x] Create data normalizers
- [x] Implement storage layer
- [x] Write model tests

### 4. File Detection System 🔄
- [x] Create file type detector
  ```go
  type DocumentType string

  const (
      TypePDF  DocumentType = "pdf"
      TypeCSV  DocumentType = "csv"
      TypeXLSX DocumentType = "xlsx"
  )
  ```
- [ ] Implement MIME type checking
- [x] Add content analysis (basic)
- [x] Create processor router
  ```go
  type ProcessorFactory interface {
      CreateProcessor(docType DocumentType) (DocumentProcessor, error)
      SupportedTypes() []DocumentType
  }
  ```
- [ ] Add new file type handlers

### 5. Error Handling & Logging ✅
- [x] Set up structured logging
  ```go
  type ProcessingError struct {
      Err      error
      Stage    ProcessingStage
      Document string
  }
  ```
- [x] Implement error types
  ```go
  type ProcessingStage string

  const (
      StageValidation    ProcessingStage = "validation"
      StageExtraction    ProcessingStage = "extraction"
      StageNormalization ProcessingStage = "normalization"
      StageAnalysis      ProcessingStage = "analysis"
  )
  ```
- [x] Add error recovery
- [x] Create error reporting
- [ ] Implement retry logic

## Integration Points
- [x] Database schema from Iteration 1
- [x] CLI commands from Iteration 1
- [x] Preparing for AI integration in Iteration 3

## Review Checklist
- [x] PDF processor working
- [ ] CSV processor working
- [x] Error handling complete
- [x] Tests passing
- [x] Documentation updated
- [ ] Performance metrics collected
- [ ] Security review done

## Success Criteria
1. [x] Successfully process PDF bank statements (basic implementation)
2. [ ] Handle multiple CSV formats
3. [ ] Accurate transaction extraction
4. [x] Robust error handling
5. [x] Test coverage > 80%

## Notes
- Keep processing modular for future formats ✅
- Document format specifications 🔄
- Consider performance with large files ✅
  - Using buffered reading
  - Proper cleanup of temporary files
  - Structured error handling with early returns 