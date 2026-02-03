---
paths:
  - "**/*.go"
  - "**/go.mod"
  - "**/go.sum"
---

# Go Rules

Version: Go 1.23+

## Tooling

- Linting: golangci-lint
- Formatting: gofmt, goimports
- Testing: go test with coverage >= 85%
- Package management: go modules

## MUST

- Use context.Context as first parameter for functions that may block
- Handle all errors explicitly with proper error wrapping
- Use errgroup for concurrent operations with error handling
- Run golangci-lint before commit
- Use defer for cleanup operations
- Document exported functions and types

## MUST NOT

- Ignore errors with blank identifier (_)
- Use panic for normal error handling
- Use init() for complex initialization logic
- Import packages without alias when names conflict
- Use global variables for state management
- Embed credentials or secrets in code

## File Conventions

- *_test.go for test files
- internal/ for private packages
- cmd/ for main entry points
- pkg/ for public reusable libraries
- Use snake_case for file names

## Testing

- Table-driven tests are preferred
- Use testify/assert or go-cmp for assertions
- Mock external dependencies with interfaces
- Use t.Parallel() for independent tests

## MoAI Integration

- Use Skill("moai-lang-go") for detailed patterns
- Follow TRUST 5 quality gates
