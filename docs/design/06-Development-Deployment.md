# Development & Deployment

## Development Environment

### Local Setup
```bash
# Required tools
- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- SQLite3
- golangci-lint
- swag (Swagger docs)

# Development database
docker-compose up -d db

# Run development server
make dev
```

### Project Structure
```
.
├── cmd/                    # Command-line entry points
│   ├── budgetassist/      # Main application
│   └── tools/             # Development tools
├── internal/              # Private application code
│   ├── api/              # API handlers
│   ├── core/             # Core business logic
│   ├── db/               # Database access
│   └── ai/               # AI integration
├── pkg/                   # Public libraries
├── web/                   # Frontend application
├── docs/                  # Documentation
└── scripts/              # Build and deployment scripts
```

## Build System

### Makefile Targets
```makefile
.PHONY: build test lint docker

build: generate
    go build -o bin/budgetassist ./cmd/budgetassist

test:
    go test -race -cover ./...

lint:
    golangci-lint run

generate:
    go generate ./...
    swag init -g cmd/budgetassist/main.go
```

### Docker Builds
```dockerfile
# Multi-stage build
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN make build

FROM alpine:3.18
COPY --from=builder /app/bin/budgetassist /usr/local/bin/
USER nobody:nobody
ENTRYPOINT ["budgetassist"]
```

## Testing Strategy

### Test Categories
```go
// Unit Tests
func TestTransactionValidation(t *testing.T) {
    cases := []struct{
        name string
        input Transaction
        wantErr bool
    }{
        // test cases...
    }
}

// Integration Tests
func TestDatabaseOperations(t *testing.T) {
    if testing.Short() {
        t.Skip()
    }
    // test with real db
}

// End-to-End Tests
func TestAPIEndpoints(t *testing.T) {
    // Start test server
    // Make real HTTP calls
}
```

### Test Data Management
```go
type TestFixture struct {
    DB          *gorm.DB
    Categories  []Category
    Transactions []Transaction
}

func setupTestData(t *testing.T) *TestFixture {
    // Create test database
    // Load fixtures
    // Return setup
}
```

## Continuous Integration

### GitHub Actions Workflow
```yaml
name: CI Pipeline

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: make test

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: golangci/golangci-lint-action@v3

  build:
    needs: [test, lint]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: make build
      - run: make docker
```

## Deployment

### Environment Configuration
```yaml
# config/production.yaml
database:
  url: ${DATABASE_URL}
  max_connections: 100
  idle_timeout: 300s

ai_service:
  url: ${AI_SERVICE_URL}
  api_key: ${AI_API_KEY}
  timeout: 30s

web_server:
  port: 8080
  read_timeout: 5s
  write_timeout: 10s
```

### Database Migrations
```go
type Migration struct {
    ID        string
    Timestamp time.Time
    Up        func(*gorm.DB) error
    Down      func(*gorm.DB) error
}

var migrations = []Migration{
    {
        ID: "001_initial_schema",
        Up: func(db *gorm.DB) error {
            // Create initial tables
        },
    },
}
```

### Monitoring & Observability

#### Prometheus Metrics
```go
var (
    transactionProcessed = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "transactions_processed_total",
            Help: "Number of processed transactions",
        },
        []string{"status", "category"},
    )
)
```

#### Logging Configuration
```go
func setupLogging() *slog.Logger {
    return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level:     slog.LevelInfo,
        AddSource: true,
    }))
}
```

## Release Process

### Version Management
```go
var (
    Version   = "dev"
    CommitSHA = "unknown"
    BuildTime = "unknown"
)
```

### Release Checklist
1. Update CHANGELOG.md
2. Run full test suite
3. Tag release version
4. Build production artifacts
5. Deploy to staging
6. Smoke tests
7. Production deployment

### Rollback Procedure
```bash
# Quick rollback script
#!/bin/bash
VERSION=$1

# Verify version exists
docker pull budgetassist:$VERSION

# Switch traffic
kubectl rollout undo deployment/budgetassist

# Verify health
./scripts/health-check.sh
``` 