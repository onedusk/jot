# Contributing to Jot

Thank you for your interest in contributing to Jot! This document provides guidelines for contributing to the project.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/jot.git`
3. Create a feature branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Run tests: `go test ./...`
6. Commit your changes
7. Push to your fork
8. Create a Pull Request

## Code Style

- Follow standard Go formatting (`gofmt`)
- Run `go vet` to catch common issues
- Keep functions focused and concise
- Add comments for exported functions and types
- Write idiomatic Go code

## Testing

- Write tests for new features
- Ensure existing tests pass
- Aim for good test coverage
- Use table-driven tests where appropriate

## Commit Messages

Use clear, descriptive commit messages:

```
type: brief description

Longer explanation if needed

Closes #issue-number
```

Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`

## Pull Request Process

1. Update documentation for any changed functionality
2. Add tests for new features
3. Ensure CI passes
4. Request review from maintainers
5. Address review feedback
6. Squash commits if requested

## Questions?

Open an issue for discussion or reach out to the maintainers.
