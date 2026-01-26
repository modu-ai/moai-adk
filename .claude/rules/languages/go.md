---
paths:
  - "**/*.go"
  - "**/go.mod"
  - "**/go.sum"
---

# Go Rules

Version: Go 1.23+

## Tooling

- Linting: golangci-lint v2.8.0+ (recommended)
- Formatting: gofmt, goimports
- Testing: go test with coverage >= 85%
- Package management: go modules

## golangci-lint Configuration

```yaml
# .golangci.yml
linters:
  enable:
    - godoc-lint      # documentation quality
    - modernize       # stringscut, unsafefuncs analyzers
    - prealloc        # slice capacity pre-allocation
    - gosec           # security rules (G116+)
    - golines         # line length enforcement

linters-settings:
  golines:
    max-len: 120
  gomoddirectives:
    toolchain:
      pattern: 'go1\.23\.\d+$'
```

## Best Practices (2026)

- Enable as many linters as feasible for maximum code quality
- Use `prealloc` rule: provide capacity to `make([]T, 0, cap)`
- Use context for cancellation and timeouts
- Use errgroup for concurrent operations with error handling
- Handle errors explicitly - never ignore with `_`

## Performance

- Pre-allocate slice capacity to minimize allocations
- Use sync.Pool for frequently allocated objects
- Profile with pprof before optimizing

## Concurrency Patterns

- Use `errgroup.WithContext()` for parallel operations
- Use `sync.Once` for one-time initialization
- Prefer channels for communication, mutexes for state

## MoAI Integration

- Use Skill("moai-lang-go") for detailed patterns
- Follow TRUST 5 quality gates
