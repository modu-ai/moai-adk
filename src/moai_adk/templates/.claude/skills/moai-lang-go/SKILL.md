---
name: "moai-lang-go"
version: "4.0.0"
status: stable
description: Enterprise Go for systems and network programming Go 1.25.4, Fiber v3, gRPC, context patterns, goroutine orchestration, standard library mastery; activates for REST APIs, microservices, concurrent systems, backend infrastructure, and performance-critical code.
allowed-tools:
  - Read
  - Bash
  - WebSearch
  - WebFetch
---

# Go Systems Development — Enterprise v4.0

## Quick Summary

**Primary Focus**: Go 1.25 with Fiber, context patterns, goroutine orchestration, and gRPC
**Best For**: REST APIs, microservices, concurrent systems, backend infrastructure
**Key Libraries**: Fiber v3, context (stdlib), sqlc, pgx, gRPC
**Auto-triggers**: Go, Golang, Fiber, gRPC, goroutine, context, microservice, REST API

| Version | Release | Support |
|---------|---------|---------|
| Go 1.25.4 | Nov 2025 | Aug 2026 |
| Fiber 3.x | 2025 | Active |
| sqlc 1.26 | 2025 | Active |
| pgx 5.7 | 2025 | Active |

---

## Three-Level Learning Path

### Level 1: Fundamentals (Read examples.md)

Core Go concepts and patterns:
- **Go 1.25 Core**: Variables, functions, interfaces, error handling
- **HTTP Servers**: Basic handlers, JSON responses, routing
- **Context**: Timeouts, cancellation, value passing
- **Goroutines & Channels**: Concurrent patterns, synchronization
- **Examples**: See `examples.md` for full code samples

### Level 2: Advanced Patterns (See reference.md)

Production-ready enterprise patterns:
- **Fiber Framework**: Middleware, route groups, advanced routing
- **Advanced Concurrency**: Worker pools, fan-out/fan-in, select patterns
- **Database Access**: sqlc, pgx, connection pooling
- **Error Handling**: Custom error types, error wrapping
- **Pattern Reference**: See `reference.md` for API details and best practices

### Level 3: Production Deployment (Consult security/performance skills)

Enterprise deployment and optimization:
- **gRPC Services**: Protocol buffers, service definitions, streaming
- **Docker Deployment**: Multi-stage builds, container optimization
- **Performance**: Profiling, benchmarking, optimization techniques
- **Monitoring**: Logging, metrics, observability
- **Details**: Skill("moai-essentials-perf"), Skill("moai-security-backend")

---

## Technology Stack (November 2025 Stable)

### Language & Runtime
- **Go 1.25.4** (November 2025, compiler & runtime improvements)
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

---

## Go Fundamentals

### Variables & Types
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

### Functions & Error Handling
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

---

## HTTP Server with Fiber

### Quick REST API
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

### Middleware
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

---

## Context & Cancellation

### Timeout Context
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

### Context with Values
```go
ctx := context.WithValue(context.Background(), "user_id", "123")

// Retrieve value
userID, ok := ctx.Value("user_id").(string)
if ok {
    fmt.Println("User:", userID)
}
```

---

## Goroutines & Channels

### Basic Concurrency
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

### Worker Pool Pattern
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

---

## Type-Safe SQL with sqlc

### Queries
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

### Usage
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

---

## Production Best Practices

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

---

## Learn More

- **Examples**: See `examples.md` for Fiber, context, goroutines, and database patterns
- **Reference**: See `reference.md` for API details, tool versions, and troubleshooting
- **Go Official**: https://golang.org/
- **Fiber Docs**: https://docs.gofiber.io/
- **sqlc**: https://sqlc.dev/
- **gRPC**: https://grpc.io/

---

**Skills**: Skill("moai-essentials-debug"), Skill("moai-essentials-perf"), Skill("moai-security-backend")
**Auto-loads**: Go projects mentioning Fiber, gRPC, goroutine, context, microservice, REST API

