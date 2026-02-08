---
name: moai-lang-go
description: >
  Go 1.23+ development specialist covering Fiber, Gin, GORM, and concurrent programming patterns. Use when building high-performance microservices, CLI tools, or cloud-native applications.
license: Apache-2.0
compatibility: Designed for Claude Code
user-invocable: false
metadata:
  version: "1.1.0"
  category: "language"
  status: "active"
  updated: "2026-01-11"
  modularized: "false"
  tags: "go, golang, fiber, gin, concurrency, microservices"
  context7-libraries: "/gofiber/fiber, /gin-gonic/gin, /go-gorm/gorm"
  related-skills: "moai-lang-rust, moai-domain-backend"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["Go", "Golang", "Fiber", "Gin", "GORM", "Echo", "Chi", ".go", "go.mod", "goroutine", "channel", "generics", "concurrent", "testing", "benchmark", "fuzzing", "microservices", "gRPC"]
  languages: ["go", "golang"]
---

## Quick Reference (30 seconds)

Go 1.23+ Development Expert for high-performance backend systems and CLI applications.

Auto-Triggers: Files with .go extension, go.mod, go.sum, goroutines, channels, Fiber, Gin, GORM, Echo, Chi

Core Use Cases:

- High-performance REST APIs and microservices
- Concurrent and parallel processing systems
- CLI tools and system utilities
- Cloud-native containerized services

Quick Patterns:

Fiber API Pattern:

Create app by calling fiber.New function. Define a get route at api/users/:id with handler function that takes fiber.Ctx and returns error. In the handler, call c.JSON with fiber.Map containing id from c.Params. Call app.Listen on port 3000.

Gin API Pattern:

Create r by calling gin.Default function. Define a GET route at api/users/:id with handler function taking gin.Context pointer. In handler, call c.JSON with status 200 and gin.H containing id from c.Param. Call r.Run on port 3000.

Goroutine with Error Handling:

Create g and ctx by calling errgroup.WithContext with context.Background. Call g.Go with function that returns processUsers with ctx. Call g.Go with function that returns processOrders with ctx. If err from g.Wait is not nil, call log.Fatal with error.

---

## Implementation Guide (5 minutes)

### Go 1.23 Language Features

New Features:

- Range over integers using for i range 10 syntax and print i
- Profile-Guided Optimization PGO 2.0
- Improved generics with better type inference

Generics Pattern:

Create generic Map function with type parameters T and U as any. Accept slice of T and function from T to U. Create result slice of U with same length. Iterate range slice setting result elements to function applied to values. Return result.

### Web Framework Fiber v3

Create app with fiber.New passing fiber.Config with ErrorHandler and Prefork true. Use recover.New, logger.New, and cors.New middleware. Create api group at api/v1 path. Define routes for listUsers, getUser with id parameter, createUser, updateUser with id, and deleteUser with id. Call app.Listen on port 3000.

### Web Framework Gin

Create r with gin.Default. Use cors.Default middleware. Create api group at api/v1 path. Define GET for users calling listUsers, GET for users/:id calling getUser, POST for users calling createUser. Call r.Run on port 3000.

Request Binding Pattern:

Define CreateUserRequest struct with Name and Email fields. Add json tags and binding tags for required, min length 2, and required email validation. In createUser handler, declare req variable, call c.ShouldBindJSON with pointer. If error, call c.JSON with 400 status and error. Otherwise call c.JSON with 201 and response data.

### Web Framework Echo

Create e with echo.New. Use middleware.Logger, middleware.Recover, and middleware.CORS. Create api group at api/v1 path. Define GET for users and POST for users. Call e.Logger.Fatal with e.Start on port 3000.

### Web Framework Chi

Create r with chi.NewRouter. Use middleware.Logger and middleware.Recoverer. Call r.Route with api/v1 path and function. Inside, call r.Route with users path. Define Get for list, Post for create, Get with id parameter for single user. Call http.ListenAndServe on port 3000 with r.

### ORM GORM 1.25

Model Definition:

Define User struct embedding gorm.Model. Add Name with uniqueIndex and not null tags, Email with uniqueIndex and not null, and Posts slice with foreignKey AuthorID tag.

Query Patterns:

Call db.Preload with Posts and function that orders by created_at desc and limits to 10, then First with user and id 1. For transactions, call db.Transaction with function taking tx pointer. Inside, create user and profile, returning any errors.

### Type-Safe SQL with sqlc

Create sqlc.yaml with version 2, sql section with postgresql engine, queries and schema paths, and go generation settings for package name, output directory, and pgx v5 sql_package.

In query.sql file, add name GetUser as one returning all columns where id matches parameter. Add name CreateUser as one inserting name and email values and returning all columns.

### Concurrency Patterns

Errgroup Pattern:

Create g and ctx with errgroup.WithContext. Call g.Go for fetchUsers that assigns to users variable. Call g.Go for fetchOrders that assigns to orders variable. If g.Wait returns error, return nil and error.

Worker Pool Pattern:

Define workerPool function taking jobs receive-only channel, results send-only channel, and n worker count. Create WaitGroup. Loop n times, incrementing WaitGroup and spawning goroutine that defers Done, ranges over jobs, and sends processJob result to results. Wait then close results.

Context with Timeout:

Create ctx and cancel with context.WithTimeout for 5 seconds. Defer cancel call. Call fetchData with ctx. If error is context.DeadlineExceeded, respond with timeout and StatusGatewayTimeout.

### Testing Patterns

Table-Driven Tests:

Define tests slice with struct containing name string, input CreateUserInput, and wantErr bool. Add test cases for valid input and empty name. Range over tests calling t.Run with name and test function. Call service Create, check if wantErr is true and require.Error.

HTTP Testing:

Create app with fiber.New. Add GET route for users/:id calling getUser. Create request with httptest.NewRequest for GET at users/1. Call app.Test with request to get response. Assert 200 status code.

### CLI Cobra with Viper

Define rootCmd as cobra.Command pointer with Use and Short fields. In init function, add PersistentFlags StringVar for cfgFile. Call viper.BindPFlag with config and lookup. Set viper.SetEnvPrefix to MYAPP and call viper.AutomaticEnv.

---

## Advanced Patterns

For comprehensive coverage including:

- Advanced concurrency patterns (worker pools, rate limiting, errgroup)
- Generics and type constraints
- Interface design and composition
- Comprehensive testing patterns (TDD, table-driven, benchmarks, fuzzing)
- Performance optimization and profiling

See: [reference/advanced.md](reference/advanced.md) for advanced patterns, [reference/testing.md](reference/testing.md) for testing patterns

### Performance Optimization

PGO Build:

Run application with GODEBUG pgo enabled and cpuprofile output. Build with go build using pgo flag pointing to profile file.

Object Pooling:

Create bufferPool as sync.Pool with New function returning 4096 byte slice. Get buffer with type assertion, defer Put to return to pool.

### Container Deployment 10-20MB

Multi-stage Dockerfile: First stage uses golang:1.23-alpine as builder, sets WORKDIR, copies go.mod and go.sum, runs go mod download, copies source, builds with CGO_ENABLED 0 and ldflags for stripped binary. Second stage uses scratch, copies binary, sets ENTRYPOINT.

### Graceful Shutdown

Spawn goroutine calling app.Listen. Create quit channel for os.Signal with buffer 1. Call signal.Notify for SIGINT and SIGTERM. Receive from quit then call app.Shutdown.

---

## Context7 Libraries

- golang/go for Go language and stdlib
- gofiber/fiber for Fiber web framework
- gin-gonic/gin for Gin web framework
- labstack/echo for Echo web framework
- go-chi/chi for Chi router
- go-gorm/gorm for GORM ORM
- sqlc-dev/sqlc for type-safe SQL
- jackc/pgx for PostgreSQL driver
- spf13/cobra for CLI framework
- spf13/viper for configuration
- stretchr/testify for testing toolkit

---

## Works Well With

- moai-domain-backend for REST API architecture and microservices
- moai-lang-rust for systems programming companion
- moai-quality-security for security hardening
- moai-essentials-debug for performance profiling
- moai-workflow-ddd for domain-driven development

---

## Troubleshooting

Common Issues:

- Module errors: Run go mod tidy and go mod verify
- Version check: Run go version and go env GOVERSION
- Build issues: Run go clean -cache and go build -v

Performance Diagnostics:

- CPU profiling: Run go test -cpuprofile cpu.prof -bench .
- Memory profiling: Run go test -memprofile mem.prof -bench .
- Race detection: Run go test -race ./...

---

## Additional Resources

See reference/advanced.md for advanced concurrency patterns, generics, and interface design.

See reference/testing.md for comprehensive testing patterns including TDD, benchmarks, and fuzzing.

---

Last Updated: 2026-01-11
Version: 1.1.0
