---
name: moai-lang-go
description: Go 1.23+ systems programming with Fiber v3, gRPC, sqlc, concurrency patterns
allowed-tools: [Read, Bash, WebFetch]
---

## Quick Reference (30 seconds)

**Primary Focus**: Systems programming, network services, concurrent applications  
**Best For**: Microservices, CLI tools, API servers, distributed systems  
**Key Libraries**: Fiber v3, gRPC, sqlc, pgx v5  
**Auto-triggers**: Go development, concurrency, performance optimization

**Version Matrix (2025-11-22)**:
- Go: 1.23.4 (stable), 1.24 (beta)
- Fiber: v3.x (web framework)
- sqlc: 1.26 (type-safe SQL)
- gRPC: 1.67 (RPC framework)

---

## What It Does

Enterprise Go development with modern patterns for building high-performance systems, network services, and concurrent applications.

**Core Capabilities**:
- ✅ Systems programming with Go 1.23+
- ✅ Web services with Fiber v3
- ✅ Type-safe database access with sqlc
- ✅ gRPC microservices
- ✅ Advanced concurrency patterns
- ✅ Performance profiling and optimization

---

## When to Use

**Automatic Triggers**:
- Go code development and review
- Microservice architecture design
- CLI tool implementation
- Concurrent system design
- High-performance API servers

**Manual Invocation**:
```
Skill("moai-lang-go")
```

---

## Three-Level Learning Path

### Level 1: Fundamentals
**See examples.md for 15 practical examples**:
- REST API with Fiber v3
- Worker pools with channels
- Type-safe SQL with sqlc
- gRPC services
- Context management
- Error handling patterns
- Testing strategies

### Level 2: Advanced Patterns
**See modules/advanced-patterns.md**:
- Generics and type parameters
- Pipeline and fan-out/fan-in patterns
- Reflection and metaprogramming
- Functional options pattern
- Code generation with go:generate
- Advanced error handling with stack traces

### Level 3: Production Optimization
**See modules/optimization.md**:
- Benchmarking and profiling
- Memory optimization (pre-allocation, pooling)
- Concurrency optimization
- Algorithm optimization
- Compiler optimization techniques
- Network and database optimization

---

## Best Practices

### DO ✅

**Concurrency**:
```go
// Use context for cancellation and timeouts
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Bound goroutines with worker pools
for w := 0; w < numWorkers; w++ {
    go worker(jobs, results)
}

// Close channels to signal completion
close(jobs)
```

**Error Handling**:
```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to process user %d: %w", userID, err)
}

// Check sentinel errors with errors.Is()
if errors.Is(err, ErrNotFound) {
    return NotFoundResponse()
}
```

**Database**:
```go
// Use sqlc for type-safe queries
user, err := queries.GetUser(ctx, userID)

// Use connection pooling
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
```

### DON'T ❌

**Anti-Patterns**:
```go
// DON'T: Leak goroutines without cleanup
go func() {
    for {
        // Infinite loop with no exit condition
    }
}()

// DON'T: Ignore errors
result, _ := riskyOperation() // Never do this

// DON'T: Use panic for control flow
if err != nil {
    panic(err) // Use error returns instead
}

// DON'T: Forget to close channels
ch := make(chan int)
go func() {
    ch <- 42
    // Missing close(ch) causes goroutine leak
}()

// DON'T: Use raw SQL strings
db.Query("SELECT * FROM users WHERE id = " + userID) // SQL injection risk
```

---

## Tool Versions (2025-11-22)

**Core**:
- Go: 1.23.4 (August 2024)
- go modules: Built-in dependency management

**Web & API**:
- Fiber: v3.x (Express-inspired framework)
- Echo: 4.13.x (Lightweight framework)
- gRPC: 1.67 (RPC framework)

**Database**:
- sqlc: 1.26 (Type-safe SQL code generator)
- pgx: v5.7 (PostgreSQL driver with pooling)
- GORM: v1.25 (ORM alternative)

**Testing**:
- testing: stdlib (built-in test framework)
- testify: v1.9 (Assertions and mocking)
- mockgen: Latest (Mock generation)

**Profiling**:
- pprof: Built-in (CPU/memory profiling)
- trace: Built-in (Execution tracing)
- Delve: v1.22 (Debugger)

---

## Installation & Setup

**Go Installation**:
```bash
# macOS
brew install go

# Linux
wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz

# Verify
go version  # go version go1.23.4 darwin/amd64
```

**Project Setup**:
```bash
# Initialize module
go mod init github.com/username/project

# Install dependencies
go get github.com/gofiber/fiber/v3
go get github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go get google.golang.org/grpc

# Run project
go run main.go

# Build binary
go build -o app main.go

# Run tests
go test ./...
go test -cover
```

**IDE Setup**:
```bash
# Install gopls (Language Server)
go install golang.org/x/tools/gopls@latest

# Install tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

---

## Works Well With

**Related Skills**:
- `moai-domain-backend` - Backend architecture patterns
- `moai-essentials-perf` - Performance profiling
- `moai-security-backend` - Security best practices
- `moai-domain-cli-tool` - CLI application development
- `moai-context7-lang-integration` - Latest Go documentation

**Complementary Technologies**:
- PostgreSQL with pgx
- Redis for caching
- Docker for containerization
- Kubernetes for orchestration

---

## Learn More

**Documentation**:
- **Examples**: [examples.md](examples.md) - 15 practical examples
- **Advanced**: [modules/advanced-patterns.md](modules/advanced-patterns.md) - Generics, concurrency, reflection
- **Performance**: [modules/optimization.md](modules/optimization.md) - Profiling and optimization
- **API Reference**: [reference.md](reference.md) - CLI commands and tools

**Progressive Learning**:
1. Start with **examples.md** for hands-on practice
2. Explore **modules/advanced-patterns.md** for enterprise patterns
3. Master **modules/optimization.md** for production performance

---

## Changelog

- **v4.1.0** (2025-11-22): Complete modularization with examples.md (15 examples), advanced-patterns.md (450 lines), optimization.md (390 lines)
- **v4.0.0** (2025-11-13): Updated to Go 1.23.4, added Fiber v3, sqlc 1.26

---

## Context7 Integration

### Related Libraries & Tools
- [Go](/golang/go): Modern systems programming language with concurrency primitives
- [Gin](/gin-gonic/gin): High-performance HTTP web framework
- [Echo](/labstack/echo): Minimalist web framework with middleware
- [Fiber](/gofiber/fiber): Express-inspired web framework (v3)
- [sqlc](/sqlc-dev/sqlc): Generate type-safe Go from SQL

### Official Documentation
- [Go Documentation](https://go.dev/doc/)
- [Go Standard Library](https://pkg.go.dev/std)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Blog](https://go.dev/blog/)

### Version-Specific Guides
Latest stable: Go 1.23.4 (August 2024)
- [Go 1.23 Release Notes](https://go.dev/doc/go1.23)
- [Go 1.24 Beta Features](https://tip.golang.org/doc/go1.24)
- [Generics Tutorial](https://go.dev/doc/tutorial/generics)
- [Concurrency Patterns](https://go.dev/blog/pipelines)

---

**Version**: 4.1.0  
**Status**: Production Ready  
**Last Updated**: 2025-11-22
