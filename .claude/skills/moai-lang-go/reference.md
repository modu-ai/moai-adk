# Go CLI Reference

Quick reference for Go 1.24, golangci-lint 1.64.7, gofmt, go test, and standard library tools.

---

## Tool Versions (2025-10-22)

| Tool | Version | Release Date | Purpose |
|------|---------|--------------|---------|
| **Go** | 1.24 | 2025-02-11 | Runtime & Language |
| **golangci-lint** | 1.64.7 | 2025-03-15 | Linter aggregator |
| **gofmt** | 1.24 | 2025-02-11 | Code formatter |
| **go vet** | 1.24 | 2025-02-11 | Static analyzer |
| **staticcheck** | 2025.1 | 2025-01-20 | Advanced linter |
| **govulncheck** | 1.1.3 | 2025-01-10 | Vulnerability scanner |

---

## Go 1.24

### Installation

```bash
# macOS (Homebrew)
brew install go@1.24

# Linux (via official binary)
wget https://go.dev/dl/go1.24.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.linux-amd64.tar.gz

# Using gvm (Go Version Manager)
gvm install go1.24
gvm use go1.24 --default

# Verify installation
go version
# Expected: go version go1.24 linux/amd64
```

### Environment Setup

```bash
# Set GOPATH (optional, defaults to $HOME/go)
export GOPATH=$HOME/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin

# Enable Go modules (default in Go 1.16+)
export GO111MODULE=on

# Set GOPROXY for faster downloads
export GOPROXY=https://proxy.golang.org,direct

# Verify environment
go env
```

### Key Go 1.24 Features

**1. Tool Dependencies Management**

```bash
# Add tool as module dependency (new in Go 1.24)
go get -tool github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.7

# Run tool via go tool command
go tool golangci-lint run

# List all registered tools
go tool -list

# Benefits:
# - Tools versioned with go.mod
# - Consistent across team members
# - No manual PATH management
```

**2. Testing Improvements: `b.Loop()`**

```go
// New benchmark pattern with automatic timer management
func BenchmarkProcessData(b *testing.B) {
    data := generateTestData(1000)

    for b.Loop() {  // Replaces: for i := 0; i < b.N; i++
        ProcessData(data)
    }
}

// Benefits:
// - Automatic timer pause/resume
// - More accurate measurements
// - Cleaner syntax
```

**3. `testing/synctest` Package**

```go
import "testing/synctest"

// Faster testing of time-dependent code without mocks
func TestWithTime(t *testing.T) {
    synctest.Run(func() {
        start := time.Now()
        time.Sleep(1 * time.Hour)
        elapsed := time.Since(start)
        // Test completes instantly, not after 1 hour
    })
}
```

---

## Project Setup

### Initialize Module

```bash
# Create new module
go mod init example.com/myapp

# Download dependencies
go mod download

# Tidy dependencies (remove unused, add missing)
go mod tidy

# Vendor dependencies locally
go mod vendor

# Verify dependencies
go mod verify
```

### Project Structure (Standard Layout)

```
myapp/
├── cmd/
│   └── myapp/
│       └── main.go          # Application entry point
├── internal/                # Private application code
│   ├── handlers/
│   ├── services/
│   └── storage/
├── pkg/                     # Public library code
│   └── models/
├── tests/                   # Additional test files
├── configs/                 # Configuration files
├── go.mod                   # Module dependencies
├── go.sum                   # Dependency checksums
└── README.md
```

---

## go test

### Basic Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific test function
go test -run TestUserValidation

# Run tests with regex pattern
go test -run 'TestUser.*'

# Run tests in specific package
go test ./internal/services

# Show only failures
go test -v ./... | grep -E '(FAIL|PASS)'
```

### Test Coverage

```bash
# Run tests with coverage
go test -cover ./...

# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# View coverage as function summary
go tool cover -func=coverage.out

# Coverage by package
go test -coverprofile=coverage.out -covermode=atomic ./...

# Enforce minimum coverage threshold
go test -cover ./... | awk '/coverage:/ {if ($2+0 < 85) exit 1}'
```

### Table-Driven Tests

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive numbers", 2, 3, 5},
        {"negative numbers", -2, -3, -5},
        {"mixed numbers", -2, 3, 1},
        {"zeros", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
            }
        })
    }
}
```

### Benchmarking

```bash
# Run all benchmarks
go test -bench=.

# Run specific benchmark
go test -bench=BenchmarkAdd

# Run with memory statistics
go test -bench=. -benchmem

# Run for specific duration
go test -bench=. -benchtime=10s

# Run with CPU profiling
go test -bench=. -cpuprofile=cpu.prof

# Run with memory profiling
go test -bench=. -memprofile=mem.prof
```

### Race Detection

```bash
# Run tests with race detector
go test -race ./...

# Build with race detector
go build -race

# Run application with race detector
go run -race main.go
```

### Test Helpers

```go
// test_helpers.go
package myapp

import "testing"

// Helper function that calls t.Helper() to report errors at call site
func assertEqual(t *testing.T, got, want interface{}) {
    t.Helper()
    if got != want {
        t.Errorf("got %v, want %v", got, want)
    }
}

// Setup function
func setupTest(t *testing.T) func() {
    t.Helper()
    // Setup code
    return func() {
        // Teardown code
    }
}

// Usage
func TestSomething(t *testing.T) {
    teardown := setupTest(t)
    defer teardown()

    assertEqual(t, Add(2, 3), 5)
}
```

---

## golangci-lint 1.64.7

### Installation

```bash
# Using Go 1.24 tool management
go get -tool github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.7

# Via Homebrew
brew install golangci-lint

# Via script
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.64.7

# Verify installation
golangci-lint --version
# Expected: golangci-lint has version 1.64.7
```

### Configuration (`.golangci.yml`)

```yaml
run:
  timeout: 5m
  tests: true
  build-tags:
    - integration

linters:
  enable:
    - gofmt
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - structcheck
    - varcheck
    - ineffassign
    - deadcode
    - typecheck
    - gosec
    - gocyclo
    - gocognit
    - goconst
    - misspell
    - revive

linters-settings:
  gocyclo:
    min-complexity: 10

  gocognit:
    min-complexity: 15

  govet:
    check-shadowing: true

  errcheck:
    check-type-assertions: true
    check-blank: true

  staticcheck:
    checks: ["all"]

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0
```

### Common Commands

```bash
# Run all enabled linters
golangci-lint run

# Run on specific directory
golangci-lint run ./internal/...

# Run specific linters only
golangci-lint run --enable=gofmt,govet

# Disable specific linters
golangci-lint run --disable=deadcode

# Auto-fix issues (when supported)
golangci-lint run --fix

# Show only new issues (for CI/CD)
golangci-lint run --new

# Generate configuration file
golangci-lint config init
```

### Pre-configured Linter Profiles

```bash
# Run with bug-prone checks
golangci-lint run --preset=bugs

# Run with style checks
golangci-lint run --preset=style

# Run with performance checks
golangci-lint run --preset=performance

# Run all presets
golangci-lint run --preset=bugs,style,performance
```

---

## Code Formatting

### gofmt

```bash
# Format all Go files in current directory
gofmt -w .

# Format with simplification (-s flag)
gofmt -s -w .

# Show diff without modifying files
gofmt -d .

# Format specific file
gofmt -w main.go

# Check if files are formatted (exit code 1 if not)
gofmt -l . | wc -l
```

### goimports

```bash
# Install goimports
go install golang.org/x/tools/cmd/goimports@latest

# Format and organize imports
goimports -w .

# Show diff
goimports -d .
```

---

## go vet

### Static Analysis

```bash
# Run go vet on all packages
go vet ./...

# Run on specific package
go vet ./internal/services

# Run with shadow checker
go vet -vettool=$(which shadow) ./...

# Common checks:
# - Unreachable code
# - Invalid printf formats
# - Suspicious constructs
# - Common mistakes
```

---

## Build & Run

### Building

```bash
# Build current package
go build

# Build with output name
go build -o myapp

# Build for different OS/architecture
GOOS=linux GOARCH=amd64 go build -o myapp-linux
GOOS=windows GOARCH=amd64 go build -o myapp.exe
GOOS=darwin GOARCH=arm64 go build -o myapp-mac

# Build with optimizations
go build -ldflags="-s -w" -o myapp  # Strip debug info

# Build with version info
go build -ldflags="-X main.version=1.0.0" -o myapp

# Show build info
go version -m myapp
```

### Running

```bash
# Run main package
go run main.go

# Run with arguments
go run main.go --config=config.yaml

# Run with build tags
go run -tags=integration main.go

# Run with race detector
go run -race main.go
```

---

## Dependency Management

### go mod Commands

```bash
# Initialize module
go mod init example.com/myapp

# Add dependency
go get github.com/gin-gonic/gin@v1.9.1

# Add dependency (latest)
go get github.com/gin-gonic/gin@latest

# Update dependency
go get -u github.com/gin-gonic/gin

# Update all dependencies
go get -u ./...

# Remove unused dependencies
go mod tidy

# Download dependencies
go mod download

# Vendor dependencies
go mod vendor

# Verify dependencies
go mod verify

# Show dependency graph
go mod graph

# Explain why dependency is needed
go mod why github.com/gin-gonic/gin

# List all modules
go list -m all

# List direct dependencies only
go list -m -json all | jq 'select(.Main != true)'
```

---

## Security & Vulnerability Scanning

### govulncheck

```bash
# Install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# Scan for vulnerabilities
govulncheck ./...

# Scan with JSON output
govulncheck -json ./...

# Scan specific package
govulncheck ./internal/services
```

### gosec (Security Linter)

```bash
# Install gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run security scan
gosec ./...

# Generate report in JSON
gosec -fmt=json -out=results.json ./...

# Run with specific rules
gosec -include=G101,G102 ./...
```

---

## Performance Profiling

### CPU Profiling

```bash
# Generate CPU profile during test
go test -cpuprofile=cpu.prof -bench=.

# Analyze CPU profile
go tool pprof cpu.prof

# Interactive commands in pprof:
# (pprof) top10       # Show top 10 functions
# (pprof) list main   # Show source code for main
# (pprof) web         # Open in browser (requires graphviz)
```

### Memory Profiling

```bash
# Generate memory profile
go test -memprofile=mem.prof -bench=.

# Analyze memory profile
go tool pprof mem.prof

# Show allocations
go tool pprof -alloc_space mem.prof
```

### HTTP Profiling

```go
import (
    _ "net/http/pprof"
    "net/http"
)

func main() {
    go func() {
        http.ListenAndServe("localhost:6060", nil)
    }()

    // Your application code
}
```

```bash
# Analyze running application
go tool pprof http://localhost:6060/debug/pprof/profile  # CPU
go tool pprof http://localhost:6060/debug/pprof/heap     # Memory
go tool pprof http://localhost:6060/debug/pprof/goroutine  # Goroutines
```

---

## Code Generation

### go generate

```bash
# Run code generation
go generate ./...

# Example directive in Go file:
//go:generate mockgen -source=user_repository.go -destination=mock_user_repository.go
```

### Protocol Buffers

```bash
# Install protoc-gen-go
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Generate Go code from .proto files
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/*.proto
```

---

## CI/CD Integration

### GitHub Actions

```yaml
name: Go CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.24']

    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: ${{ matrix.go-version }}

    - name: Install dependencies
      run: go mod download

    - name: Run golangci-lint
      uses: golangci/golangci-lint-action@v6
      with:
        version: v1.64.7

    - name: Run tests
      run: go test -v -race -coverprofile=coverage.out ./...

    - name: Check coverage
      run: |
        coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        if (( $(echo "$coverage < 85" | bc -l) )); then
          echo "Coverage $coverage% is below 85%"
          exit 1
        fi

    - name: Upload coverage
      uses: codecov/codecov-action@v4
      with:
        files: ./coverage.out
```

---

## TRUST 5 Principles for Go

### T - Test First (Coverage ≥85%)

```bash
# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Check coverage percentage
go tool cover -func=coverage.out | grep total
```

### R - Readable

```bash
# Format code
gofmt -s -w .

# Run linters
golangci-lint run
```

### U - Unified (Type Safety)

```go
// Leverage Go's strong typing
type UserID int64
type Email string

// Compile-time safety
func FindUser(id UserID) (*User, error) {
    // Implementation
}
```

### S - Secured

```bash
# Run security scanner
gosec ./...

# Check for vulnerabilities
govulncheck ./...

# Run with security-focused linters
golangci-lint run --enable=gosec,govet
```

### T - Trackable

```go
// @TAG integration in code
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: auth_service_test.go
package auth

func Authenticate(token string) (*User, error) {
    // Implementation
}
```

---

## Quick Reference Commands

```bash
# Development
go mod tidy                        # Clean dependencies
go fmt ./...                       # Format code
go vet ./...                       # Static analysis
golangci-lint run                  # Lint code

# Testing
go test ./...                      # Run all tests
go test -v ./...                   # Verbose output
go test -cover ./...               # With coverage
go test -race ./...                # With race detector
go test -bench=.                   # Run benchmarks

# Building
go build                           # Build current package
go build -o myapp                  # Build with name
go install                         # Build and install

# Dependencies
go get package@version             # Add dependency
go get -u ./...                    # Update all
go mod tidy                        # Remove unused

# Security
govulncheck ./...                  # Scan vulnerabilities
gosec ./...                        # Security analysis
```

---

**Version**: 1.0.0 (2025-10-22)
**Updated**: Latest tool versions verified 2025-10-22
**Framework**: MoAI-ADK Go Language Skill
**Status**: Production-ready
