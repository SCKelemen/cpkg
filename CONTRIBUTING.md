# Contributing to cpkg

Thank you for your interest in contributing to cpkg! This document provides guidelines and instructions for contributing.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/cpkg.git`
3. Create a branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Test your changes: `go test ./...`
6. Commit your changes: `git commit -m "Add feature: description"`
7. Push to your fork: `git push origin feature/your-feature-name`
8. Open a Pull Request

## Development Setup

```bash
# Clone the repository
git clone https://github.com/SCKelemen/cpkg.git
cd cpkg

# Build the binary
go build -o cpkg .

# Run tests
go test ./...

# Run integration tests (requires git and internet)
go test -tags=integration ./...
```

## Code Style

- Follow Go conventions and use `gofmt`
- Run `go vet ./...` before committing
- Keep functions focused and small
- Add comments for exported functions and types
- Write tests for new functionality

## Commit Messages

- Use clear, descriptive commit messages
- Start with a verb in imperative mood (e.g., "Add", "Fix", "Update")
- Reference issues when applicable: "Fix #123: description"

## Pull Request Process

1. Ensure all tests pass
2. Update documentation if needed
3. Add tests for new features
4. Keep PRs focused on a single feature or fix
5. Update CHANGELOG.md if applicable (if we add one)

## Reporting Issues

When reporting issues, please include:
- Description of the problem
- Steps to reproduce
- Expected behavior
- Actual behavior
- Environment (OS, Go version, etc.)
- Relevant error messages or logs

## Feature Requests

Feature requests are welcome! Please open an issue with:
- Clear description of the feature
- Use case and motivation
- Proposed implementation approach (if you have one)

## Questions?

Feel free to open an issue for questions or discussions about the project.

