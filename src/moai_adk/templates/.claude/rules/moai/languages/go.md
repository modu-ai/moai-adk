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

## Preferred Patterns

- Use context for cancellation
- Use errgroup for concurrent operations
- Handle errors explicitly

## MoAI Integration

- Use Skill("moai-lang-go") for detailed patterns
- Follow TRUST 5 quality gates
