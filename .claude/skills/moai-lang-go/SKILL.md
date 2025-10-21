---
name: moai-lang-go
description: Go best practices with go test, golint, gofmt, and standard library utilization
allowed-tools:
  - Read
  - Bash
tier: 3
auto-load: "true"
---

# Go Expert

## What it does

Provides Go-specific expertise for TDD development, including go test framework, golint/staticcheck, gofmt formatting, and effective standard library usage.

## When to use

- “Writing Go tests”, “How to use go tests”, “Go standard library”
- Automatically invoked when working with Go projects
- Go SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **go test**: Built-in testing framework
- **Table-driven tests**: Structured test cases
- **testify/assert**: Optional assertion library
- Test coverage ≥85% with `go test -cover`

**Code Quality**:
- **gofmt**: Automatic code formatting
- **golint**: Go linter (deprecated, use staticcheck)
- **staticcheck**: Advanced static analysis
- **go vet**: Built-in error detection

**Standard Library**:
- Use standard library first before external dependencies
- **net/http**: HTTP server/client
- **encoding/json**: JSON marshaling
- **context**: Context propagation

**Go Patterns**:
- Interfaces for abstraction (small interfaces)
- Error handling with explicit returns
- Defer for cleanup
- Goroutines and channels for concurrency

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Exported names start with capital letters
- Error handling: `if err != nil { return err }`
- Avoid naked returns in large functions

## Examples

### Example 1: TDD with table-driven tests
User: "/alfred:2-run PARSE-001"
Claude: (creates RED test with table-driven approach, GREEN implementation, REFACTOR)

### Example 2: Coverage check
User: "go test check coverage"
Claude: (runs go test -cover ./... and reports coverage)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Go-specific review)
- alfred-performance-optimizer (Go profiling)
