---
name: my-harness-quality
description: >
  Quality/Testing domain knowledge for moai-adk-go covering Go testing patterns, linting,
  coverage targets, race detection, and LSP quality gates.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Grep, Glob, Bash
user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-05-14"
  modularized: "false"
  tags: "quality, testing, linting, coverage, golangci-lint, go vet, race detection, LSP"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["go test", "lint", "vet", "coverage", "golangci-lint", "quality gate", "LSP", "race", "t.TempDir", "flaky", "test isolation"]
  agents:
    - "my-harness-quality-specialist"
    - "manager-quality"
  phases:
    - "run"
    - "sync"
  languages:
    - "go"
---

# Quality/Testing Domain Knowledge

Domain-specific knowledge for Go testing, linting, and quality gates in moai-adk-go. Supplements `manager-quality` with project-specific patterns.

## Quick Reference

### Quality Targets

| Metric | Command | Target | Phase |
|--------|---------|--------|-------|
| Vet | `go vet ./...` | Zero errors | All |
| Lint | `golangci-lint run` | Zero errors | All |
| Test | `go test ./...` | All pass | All |
| Race | `go test -race ./...` | Zero races | Run |
| Coverage (package) | `go test -cover ./...` | 85%+ | Sync |
| Coverage (critical) | `go test -cover ./internal/cli/...` | 90%+ | Sync |
| LSP errors (run) | LSP diagnostics | Zero errors | Run |
| LSP warnings (sync) | LSP diagnostics | Max 10 | Sync |

### Critical Test Isolation Rules

1. **Always use `t.TempDir()`** -- creates dirs under `/tmp` with automatic cleanup
2. **`filepath.Join` trap**: `filepath.Join("/a/b", "/var/folders/x")` = `/a/b/var/folders/x` (WRONG). Use `filepath.Abs()` for user-supplied paths.
3. **OTEL env vars**: Never use `t.Setenv("OTEL_*")` in parallel tests -- causes data races from global state
4. **Full suite after any fix**: `go test ./...` (not just the fixed package)
5. **Disable caching for flaky tests**: `go test -count=1 ./...`

### Test Execution Commands

```bash
# Full test suite
go test ./...

# With race detection
go test -race ./...

# With coverage
go test -cover ./...

# Specific test
go test -run TestFunctionName ./internal/cli/

# Disable caching (flaky test debugging)
go test -count=1 ./...

# Verbose output
go test -v ./internal/template/...
```

## Implementation Guide

### Table-Driven Test Pattern

Standard Go convention for parameterized tests:

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "input", "expected", false},
        {"empty input", "", "", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Function(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Function() error = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("Function() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Test Isolation Best Practices

```go
func TestWithTempDir(t *testing.T) {
    tempDir := t.TempDir()  // Auto-cleanup, under /tmp

    // For paths that may be absolute, always resolve:
    absPath, err := filepath.Abs(somePath)
    if err != nil {
        t.Fatal(err)
    }
    // NEVER: filepath.Join(cwd, absPath) when absPath starts with /
}
```

### LSP Quality Gate Thresholds

Quality gates auto-detect project language and run the appropriate toolchain:

| Phase | LSP Errors | LSP Type Errors | Lint Errors | Warnings |
|-------|-----------|----------------|-------------|----------|
| plan | Baseline captured | Baseline captured | Baseline captured | Baseline captured |
| run | Zero required | Zero required | Zero required | - |
| sync | Zero required | Zero required | Zero required | Max 10 |

### Quality Gate Commands

```bash
# Pre-commit quality gate (parallel)
go vet ./... && golangci-lint run && go test ./...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Common Quality Issues

| Issue | Cause | Fix |
|-------|-------|-----|
| `t.TempDir()` path resolution | `filepath.Join` with absolute path | Use `filepath.Abs()` |
| OTEL data race in parallel tests | Global state from env vars | Use fake/no-op exporter |
| Test modifies project root | No temp dir isolation | Always use `t.TempDir()` |
| Flaky test on CI | Race condition or timing | Use `go test -race -count=1` |

## Cross-References

- CLAUDE.local.md Section 6: Testing Guidelines
- CLAUDE.md Section 6: Quality Gates (TRUST 5 framework)
- `moai-foundation-quality` skill: General quality orchestration
- `moai-ref-testing-pyramid` skill: Test pyramid strategy
- `.moai/config/sections/quality.yaml`: Quality configuration
