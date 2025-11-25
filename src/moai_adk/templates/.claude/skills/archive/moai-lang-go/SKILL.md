---
name: moai-lang-go
description: Enterprise Go 1.23+ for systems and network programming, Fiber v3,
version: 1.0.0
modularized: false
tags:
  - go
  - programming-language
  - enterprise
  - development
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: lang, go, moai  


## Quick Reference (30 seconds)

# Go Systems Development — Enterprise

## Level 1: Quick Reference

### Go Fundamentals

**Variables & Types**:
```go
// Type declarations
var name string = "John"
var age int = 30
price := 19.99  // Type inference

// Structs
type User struct {
    ID    int
    Name  string
    Email string
}

// Interfaces
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

**Functions & Error Handling**:
```go
// Basic function
func Greet(name string) string {
    return "Hello, " + name
}

// Multiple return values
func Divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

// Error handling
result, err := Divide(10, 0)
if err != nil {
    log.Fatal(err)
}
```

### HTTP Server with Fiber

**Quick REST API**:
```go
package main

import "github.com/gofiber/fiber/v3"

func main() {
    app := fiber.New()

    // GET handler
    app.Get("/users", func(c fiber.Ctx) error {
        return c.JSON(fiber.Map{"users": []string{"John", "Jane"}})
    })

    // POST handler
    app.Post("/users", func(c fiber.Ctx) error {
        type User struct {
            Name  string `json:"name"`
            Email string `json:"email"`
        }

        var user User
        if err := c.BodyParser(&user); err != nil {
            return c.Status(fiber.StatusBadRequest).SendString(err.Error())
        }

        return c.Status(fiber.StatusCreated).JSON(user)
    })

    // Route parameters
    app.Get("/users/:id", func(c fiber.Ctx) error {
        id := c.Params("id")
        return c.SendString("User: " + id)
    })

    app.Listen(":3000")
}
```

### Context & Cancellation

**Timeout Context**:
```go
package main

import "context"

ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()

// Use context for operations
select {
case <-ctx.Done():
    fmt.Println("Operation cancelled:", ctx.Err())
case <-time.After(5 * time.Second):
    fmt.Println("Operation completed")
}
```

**Context with Values**:
```go
ctx := context.WithValue(context.Background(), "user_id", "123")

// Retrieve value
userID, ok := ctx.Value("user_id").(string)
if ok {
    fmt.Println("User:", userID)
}
```

### Goroutines & Channels

**Basic Concurrency**:
```go
// Start goroutine
go func() {
    fmt.Println("Running concurrently")
}()

// Channels
ch := make(chan int)
go func() {
    ch <- 42
}()
value := <-ch

// Close channel
close(ch)
```

**Worker Pool Pattern**:
```go
func worker(id int, jobs <-chan int, results chan<- int) {
    for job := range jobs {
        results <- job * 2
    }
}

func main() {
    jobs := make(chan int, 100)
    results := make(chan int, 100)

    // Start 3 workers
    for w := 1; w <= 3; w++ {
        go worker(w, jobs, results)
    }

    // Send jobs and collect results
    for j := 1; j <= 9; j++ {
        jobs <- j
    }
    close(jobs)
}
```



## Core Implementation

## Technology Stack (November 2025 Stable)

### Language & Runtime
- **Go 1.23.4** (August 2024, compiler & runtime improvements)
- **Unix/Linux first** with Windows/macOS support
- **Garbage collection** with concurrent sweeper

### Web Frameworks
- **Fiber v3.x** (Express.js-inspired, high performance)
- **Echo 4.13.x** (Scalable, middleware-rich)
- **Chi 5.x** (Lightweight, composable)

### Concurrency & RPC
- **goroutines** (lightweight threads, stdlib)
- **channels** (typed message passing)
- **gRPC 1.67** (Protocol buffers, streaming)
- **Protobuf 3.21** (Message serialization)

### Data Access
- **sqlc 1.26** (Type-safe SQL code generation)
- **pgx 5.7** (PostgreSQL driver with pooling)
- **context** (Request-scoped data, timeouts)

### Testing & Quality
- **testing** (stdlib testing package)
- **testify 1.9** (Assertions, mocking, suites)
- **benchmarking** (Built-in performance testing)


## Level 2: Core Implementation

### Type-Safe SQL with sqlc

**Queries**:
```sql
-- queries.sql
-- name: GetUser :one
SELECT id, name, email FROM users WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (name, email) VALUES ($1, $2)
RETURNING id, name, email;

-- name: ListUsers :many
SELECT id, name, email FROM users ORDER BY id;
```

**Usage**:
```go
db := New(pool)
ctx := context.Background()

// Create user
user, _ := db.CreateUser(ctx, CreateUserParams{
    Name:  "John",
    Email: "john@example.com",
})

// Get user
user, _ := db.GetUser(ctx, 1)

// List users
users, _ := db.ListUsers(ctx)
```

### Middleware with Fiber

```go
app.Use(func(c fiber.Ctx) error {
    fmt.Println("Before handler")
    err := c.Next()
    fmt.Println("After handler")
    return err
})

app.Get("/protected", AuthMiddleware, func(c fiber.Ctx) error {
    return c.SendString("Protected route")
})
```

### Advanced Error Handling

```go
type AppError struct {
    Code    int
    Message string
    Details string
}

func (e *AppError) Error() string {
    return fmt.Sprintf("Error %d: %s - %s", e.Code, e.Message, e.Details)
}

func NewAppError(code int, message, details string) *AppError {
    return &AppError{
        Code:    code,
        Message: message,
        Details: details,
    }
}

// Usage in handler
app.Get("/error", func(c fiber.Ctx) error {
    err := NewAppError(500, "Internal Error", "Database connection failed")
    return c.Status(err.Code).JSON(err)
})
```


## Level 4: Production Deployment

### Production Best Practices

1. **Use context for timeouts** in concurrent operations
2. **Handle errors immediately** with meaningful messages
3. **Use type-safe SQL** with sqlc, not raw queries
4. **Implement connection pooling** for databases
5. **Use middleware for cross-cutting concerns** (logging, auth)
6. **Goroutines should be bounded** to prevent resource exhaustion
7. **Close channels explicitly** to signal completion
8. **Use sync.WaitGroup** for goroutine synchronization
9. **Profile before optimization** with pprof
10. **Deploy with graceful shutdown** handling

### Docker Deployment

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 3000
CMD ["./main"]
```

### Graceful Shutdown

```go
func main() {
    app := fiber.New()
    
    // Setup routes...
    
    // Graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    
    go func() {
        sigchan := make(chan os.Signal, 1)
        signal.Notify(sigchan, os.Interrupt)
        <-sigchan
        cancel()
    }()
    
    go func() {
        <-ctx.Done()
        log.Println("Shutting down server...")
        app.Shutdown()
    }()
    
    app.Listen(":3000")
}
```

### Related Skills
- `Skill("moai-domain-cli-tool")` for CLI development
- `Skill("moai-essentials-perf")` for performance optimization
- `Skill("moai-security-backend")` for security patterns




## Changelog

- **v4.1.0** (2025-11-22): Updated to Go 1.23.4, removed Go 1.19 patterns, added Range over integers, PGO 2.0, workspace improvements, Go 1.24 preview features
- **v4.0.0** (2025-11-13): Previous major update

---

**Version**: 4.1.0 Enterprise  
**Last Updated**: 2025-11-22  
**Status**: Production Ready




## Context7 Integration

### Related Libraries & Tools
- [Go](/golang/go): Modern systems programming language
- [Gin](/gin-gonic/gin): High-performance HTTP framework
- [Echo](/labstack/echo): Minimalist web framework

### Official Documentation
- [Documentation](https://go.dev/doc/)
- [API Reference](https://pkg.go.dev/std)

### Version-Specific Guides
Latest stable version: 1.21
- [Release Notes](https://go.dev/doc/devel/release)
- [Migration Guide](https://go.dev/doc/go1.21)
