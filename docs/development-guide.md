# Development Guide

## Prerequisites

- Go 1.24.0 or later
- Node.js 20.11.0 or later
- Docker (for local development)
- Git

## Getting Started

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/Budget-Assist.git
   cd Budget-Assist
   ```

2. Install dependencies:
   ```bash
   # Install Go dependencies
   go mod download

   # Install Node.js dependencies
   npm install
   ```

3. Set up your development environment:
   ```bash
   # Install development tools
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   go install github.com/swaggo/swag/cmd/swag@latest
   ```

## Development Workflow

### Code Structure

```
.
├── cmd/                    # Command-line entry points
│   └── budgetassist/      # Main application
├── internal/              # Private application code
│   ├── api/              # API handlers
│   ├── core/             # Core business logic
│   ├── db/               # Database access
│   └── ai/               # AI integration
├── pkg/                  # Public libraries
├── web/                  # Frontend application
└── docs/                 # Documentation
```

### Running the Application

1. Start the development server:
   ```bash
   make dev
   ```

2. Run tests:
   ```bash
   make test
   ```

3. Run linters:
   ```bash
   make lint
   ```

### Making Changes

1. Create a new branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes following our coding standards
3. Run tests and linters
4. Commit using conventional commits format
5. Push and create a pull request

### Testing

- Write unit tests for all new functionality
- Use table-driven tests where appropriate
- Ensure race condition checking with `-race` flag
- Aim for high test coverage

### Code Quality

- Follow Go coding standards (see `docs/Golang standards and best practices.md`)
- Use `golangci-lint` for code quality checks
- Document all exported functions and types
- Keep functions focused and small

## Debugging

1. Use the built-in Go debugger:
   ```bash
   go debug ./cmd/budgetassist
   ```

2. Log levels:
   - DEBUG: Development details
   - INFO: General information
   - WARN: Warning conditions
   - ERROR: Error conditions

## Common Issues

1. **Database Connection Issues**
   - Check connection string in config
   - Ensure database is running
   - Verify migrations are up to date

2. **Build Errors**
   - Run `go mod tidy`
   - Check Go version compatibility
   - Clear module cache if needed

## Getting Help

- Check existing documentation
- Review closed issues on GitHub
- Ask in the team chat
- Create a new issue if needed 