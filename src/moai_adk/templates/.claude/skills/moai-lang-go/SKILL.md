---

name: moai-lang-go
description: Go best practices with go test, golint, gofmt, and standard library utilization. Use when writing or reviewing Go code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Go Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Go code discussions, framework guidance, or file extensions such as .go. |
| Tier | 3 |

## What it does

Provides Go-specific expertise for TDD development, including go test framework, golint/staticcheck, gofmt formatting, and effective standard library usage.

## When to use

- Engages when the conversation references Go work, frameworks, or files like .go.
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
```bash
go test ./... && golangci-lint run
```

## Inputs
- Language-specific source directories (e.g. `src/`, `app/`).
- Language-specific build/test configuration files (e.g. `package.json`, `pyproject.toml`, `go.mod`).
- Relevant test suites and sample data.

## Outputs
- Test/lint execution plan tailored to the selected language.
- List of key language idioms and review checkpoints.

## Failure Modes
- When the language runtime or package manager is not installed.
- When the main language cannot be determined in a multilingual project.

## Dependencies
- Access to the project file is required using the Read/Grep tool.
- When used with `Skill("moai-foundation-langs")`, it is easy to share cross-language conventions.

## References
- The Go Authors. "Effective Go." https://go.dev/doc/effective_go (accessed 2025-03-29).
- GolangCI. "golangci-lint Documentation." https://golangci-lint.run/usage/quick-start/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Input/output/failure response/reference information for each language has been specified.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Go-specific review)
- alfred-performance-optimizer (Go profiling)

## Best Practices
- Enable automatic validation by matching your linter with the language's official style guide.
- Fix test/build pipelines with reproducible commands in CI.
