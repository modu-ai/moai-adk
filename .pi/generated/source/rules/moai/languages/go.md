---
paths: "**/*.go,**/go.mod,**/go.sum"
---

# Go Development Guide

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

## Go 1.23 Features

- Range over integers: for i := range 10
- Profile-Guided Optimization (PGO) 2.0
- Improved generics with better type inference

## Web Frameworks

### Fiber v3

Create app with fiber.New, configure ErrorHandler and Prefork. Use recover, logger, cors middleware. Create API group at /api/v1, define CRUD routes. Call app.Listen on port 3000.

### Gin

Create router with gin.Default, use cors.Default middleware. Create API group, define routes with request binding (ShouldBindJSON with validation tags).

### Echo

Create with echo.New, use Logger, Recover, CORS middleware. Define route groups and start with e.Start.

### Chi

Create router with chi.NewRouter, use Logger and Recoverer middleware. Define route groups with r.Route pattern.

## ORM: GORM 1.25

Model definition: Embed gorm.Model, use struct tags (uniqueIndex, not null, foreignKey). Query with Preload, First, Where. Use db.Transaction for atomic operations.

## Type-Safe SQL: sqlc

Configure sqlc.yaml with version 2, PostgreSQL engine, pgx v5. Write SQL queries with -- name annotations (:one, :many, :exec).

## Concurrency Patterns

- **Errgroup**: errgroup.WithContext for parallel tasks with error propagation
- **Worker Pool**: Channel-based with WaitGroup for bounded concurrency
- **Context Timeout**: context.WithTimeout for deadline-aware operations
- **Graceful Shutdown**: Signal channel (SIGINT/SIGTERM) + app.Shutdown

## Testing

- Table-driven tests are preferred
- Use testify/assert or go-cmp for assertions
- Mock external dependencies with interfaces
- Use t.Parallel() for independent tests
- HTTP testing: httptest.NewRequest + app.Test for framework handlers
- CPU profiling: go test -cpuprofile cpu.prof -bench .
- Memory profiling: go test -memprofile mem.prof -bench .
- Race detection: go test -race ./...
- Fuzzing: func FuzzXxx(f *testing.F) with f.Add seed values

## CLI: Cobra + Viper

Define rootCmd as cobra.Command with Use, Short fields. In init, add PersistentFlags. Bind with viper.BindPFlag, set viper.SetEnvPrefix and viper.AutomaticEnv.

## Performance Optimization

- PGO Build: go build -pgo=cpu.prof
- Object Pooling: sync.Pool with New function, Get/Put pattern
- Container deployment: Multi-stage Dockerfile (golang:1.23-alpine → scratch), CGO_ENABLED=0, stripped binary (10-20MB)

## Troubleshooting

- Module errors: go mod tidy && go mod verify
- Version check: go version && go env GOVERSION
- Build issues: go clean -cache && go build -v

## Context7 Libraries

golang/go, gofiber/fiber, gin-gonic/gin, labstack/echo, go-chi/chi, go-gorm/gorm, sqlc-dev/sqlc, jackc/pgx, spf13/cobra, spf13/viper, stretchr/testify
