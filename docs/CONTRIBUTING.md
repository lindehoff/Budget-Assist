# Contributing to Budget-Assist

Thank you for your interest in contributing to Budget-Assist! This document provides guidelines and instructions for contributing to the project.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct. Please read it before contributing.

## How Can I Contribute?

### Reporting Bugs

1. Check the [issue tracker](https://github.com/yourusername/Budget-Assist/issues) to avoid duplicates
2. If no existing issue exists, create a new one with:
   - Clear title and description
   - Steps to reproduce
   - Expected vs actual behavior
   - Screenshots if applicable
   - System information

### Suggesting Enhancements

1. Check existing issues and discussions
2. Create a new issue with:
   - Clear use case
   - Expected benefits
   - Potential implementation approach
   - Any potential drawbacks

### Pull Requests

1. Fork the repository
2. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. Make your changes following our coding standards
4. Write or update tests
5. Update documentation
6. Commit your changes using [conventional commits](https://www.conventionalcommits.org/):
   ```bash
   git commit -m "feat: add new budget analysis feature"
   ```
7. Push to your fork
8. Create a Pull Request

## Development Process

### Setting Up Development Environment

1. Follow the [Installation Guide](./installation.md)
2. Set up pre-commit hooks:
   ```bash
   pre-commit install
   ```

### Coding Standards

1. Follow [Go standards](./Golang%20standards%20and%20best%20practices.md)
2. Use meaningful variable and function names
3. Write clear comments and documentation
4. Keep functions focused and small
5. Use proper error handling

### Testing

1. Write tests for new functionality
2. Update existing tests when changing behavior
3. Ensure all tests pass:
   ```bash
   make test
   ```
4. Check test coverage:
   ```bash
   make coverage
   ```

### Documentation

1. Update relevant documentation
2. Document new features
3. Update API documentation if needed
4. Keep README.md current

## Review Process

### Before Submitting

1. Run all tests
2. Run linters:
   ```bash
   make lint
   ```
3. Update documentation
4. Squash related commits
5. Rebase on main branch

### Code Review

1. All code must be reviewed
2. Address review comments
3. Keep discussions focused
4. Be respectful and constructive

## Project Structure

```
.
├── cmd/                    # Command-line entry points
├── internal/              # Private application code
├── pkg/                  # Public libraries
├── web/                  # Frontend application
├── docs/                 # Documentation
└── tests/               # Test suites
```

## Communication

- GitHub Issues: Bug reports and feature requests
- Discussions: General questions and ideas
- Pull Requests: Code review discussions

## Release Process

1. Version numbers follow [semantic versioning](https://semver.org/)
2. Changes are documented in CHANGELOG.md
3. Releases are tagged and published to GitHub Releases

## Getting Help

1. Check the documentation
2. Search existing issues
3. Ask in GitHub Discussions
4. Create a new issue

## Recognition

Contributors are recognized in:
- CONTRIBUTORS.md file
- Release notes
- Project documentation

Thank you for contributing to Budget-Assist! 