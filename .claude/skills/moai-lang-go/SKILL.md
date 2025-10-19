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

- "Go 테스트 작성", "go test 사용법", "Go 표준 라이브러리", "마이크로서비스", "네트워크 프로그래밍", "시스템 도구"
- "gRPC", "REST API", "CLI 도구", "Docker", "Kubernetes", "웹 서버"
- "gin", "Echo", "Beego", "Fiber"
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

## Modern Go (1.21+)

**Recommended Version**: Go 1.21+ for production, 1.20+ for legacy support

**Modern Features**:
- **Range over integers** (1.22+): `for i := range 10`
- **Range over functions** (1.22+): `for v := range generator()`
- **Clear built-in** (1.21+): `clear(map)`
- **Iterators** (1.22+): Custom iteration patterns
- **slog package** (1.21+): Structured logging
- **unsafe.String/SliceData** (1.20+): Safe unsafe operations

**Version Check**:
```bash
go version  # Check Go version
go env GOVERSION
```

## Package Management Commands

### Using go mod (Built-in - Recommended)
```bash
# Initialize module
go mod init github.com/user/project

# Add dependencies
go get github.com/gin-gonic/gin
go get github.com/stretchr/testify

# Tidy dependencies (remove unused, add missing)
go mod tidy

# Upgrade dependencies
go get -u ./...
go get -u github.com/gin-gonic/gin@latest

# Show dependency graph
go mod graph
go mod why

# Vendor dependencies
go mod vendor

# Run tests
go test ./...
go test -v ./...
go test -cover ./...

# Download cached modules
go mod download
```

### Common Development Commands
```bash
# Format code
go fmt ./...
gofmt -w .

# Lint
golangci-lint run ./...
staticcheck ./...

# Build
go build -o myapp
go build -ldflags "-s -w" -o myapp

# Run directly
go run main.go
go run ./cmd/cli
```

## Examples

### Example 1: TDD with table-driven tests
User: "/alfred:2-run PARSE-001"
Claude: (creates RED test with table-driven approach, GREEN implementation, REFACTOR)

### Example 2: Coverage check
User: "go test 커버리지 확인"
Claude: (runs go test -cover ./... and reports coverage)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Go-specific review)
- alfred-performance-optimizer (Go profiling)
