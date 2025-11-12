---
name: moai-lang-go
description: Enterprise Go for systems and network programming: Go 1.25.4, Fiber v3, gRPC, context patterns, goroutine orchestration, standard library mastery; activates for REST APIs, microservices, concurrent systems, backend infrastructure, and performance-critical code.
allowed-tools:
  - Read
  - Bash
  - WebSearch
  - WebFetch
---

# Go Systems Development — Enterprise v4.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Version** | **4.0.0 Enterprise** |
| **Created** | 2025-11-12 |
| **Updated** | 2025-11-12 |
| **Lines** | ~950 lines |
| **Size** | ~30KB |
| **Tier** | **3 (Professional)** |
| **Allowed tools** | Read, Bash, WebSearch, WebFetch |
| **Auto-load** | Go REST APIs, goroutine patterns, gRPC services, microservices |
| **Trigger cues** | Go, Golang, Fiber, gRPC, goroutine, context, microservice, REST API, concurrent |

## Technology Stack (November 2025 Stable)

### Core Language
- **Go 1.25.4** (Released November 5, 2025)
  - Compiler improvements
  - Runtime enhancements
  - Standard library updates
  - Maintenance until August 2026

### Web Frameworks
- **Fiber v3.x** (Express-like framework)
  - High performance
  - Middleware system
  - Request routing
  - Built-in recovery

- **Echo 4.13.x** (Scalable framework)
  - Route groups
  - Middleware support
  - Custom context

- **Chi 5.x** (Composable router)
  - Lightweight
  - Middleware chains
  - Subrouters

### RPC & Communication
- **gRPC 1.67.x** (Service communication)
  - Protocol Buffers
  - Streaming support
  - Load balancing

- **Protobuf 3.21.x** (Message serialization)
  - Backward compatibility
  - Language support

### Database & ORM
- **sqlc 1.26.x** (Type-safe SQL)
  - Code generation
  - Query validation
  - PostgreSQL-first

- **pgx 5.7.x** (PostgreSQL driver)
  - Async support
  - Connection pooling
  - Prepared statements

### Testing & Tooling
- **testify 1.9.x** (Testing assertions)
  - Assert helpers
  - Mock framework
  - Suite support

- **Go standard library**
  - testing package
  - context package
  - net package

## Level 1: Fundamentals (High Freedom)

### 1. Go 1.25 Features

Go 1.25 provides stability and performance improvements:

**Basic HTTP Server**:
```go
package main

import (
    "net/http"
    "log"
    "fmt"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprintf(w, `{"message": "Hello, World!"}`)
}

func main() {
    http.HandleFunc("/hello", helloHandler)
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

**Context Management**:
```go
package main

import (
    "context"
    "time"
    "fmt"
)

func expensiveOperation(ctx context.Context) string {
    select {
    case <-time.After(5 * time.Second):
        return "Operation completed"
    case <-ctx.Done():
        return fmt.Sprintf("Cancelled: %v", ctx.Err())
    }
}

func main() {
    // Timeout context
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    
    result := expensiveOperation(ctx)
    fmt.Println(result) // Output: Cancelled: context deadline exceeded
}
```

### 2. Fiber Framework (v3)

Fiber is Express.js-inspired Go framework:

**Route Handlers**:
```go
package main

import (
    "github.com/gofiber/fiber/v3"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Email string `json:"email"`
}

func main() {
    app := fiber.New()
    
    // GET request
    app.Get("/users", func(c fiber.Ctx) error {
        users := []User{
            {ID: 1, Name: "John", Email: "john@example.com"},
            {ID: 2, Name: "Jane", Email: "jane@example.com"},
        }
        return c.JSON(users)
    })
    
    // GET with parameter
    app.Get("/users/:id", func(c fiber.Ctx) error {
        id := c.Params("id")
        return c.JSON(fiber.Map{"id": id})
    })
    
    // POST request
    app.Post("/users", func(c fiber.Ctx) error {
        var user User
        if err := c.BodyParser(&user); err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
        }
        
        // Save user to database
        return c.Status(fiber.StatusCreated).JSON(user)
    })
    
    app.Listen(":3000")
}
```

**Middleware**:
```go
package main

import (
    "github.com/gofiber/fiber/v3"
    "log"
    "time"
)

func main() {
    app := fiber.New()
    
    // Logging middleware
    app.Use(func(c fiber.Ctx) error {
        start := time.Now()
        err := c.Next()
        duration := time.Since(start)
        
        log.Printf("%s %s - %dms", c.Method(), c.Path(), duration.Milliseconds())
        return err
    })
    
    // Authentication middleware
    app.Use(func(c fiber.Ctx) error {
        token := c.Get("Authorization")
        if token == "" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing token"})
        }
        return c.Next()
    })
    
    app.Get("/protected", func(c fiber.Ctx) error {
        return c.JSON(fiber.Map{"message": "Access granted"})
    })
    
    app.Listen(":3000")
}
```

### 3. Goroutine Patterns

Goroutines provide lightweight concurrency:

**Basic Goroutines**:
```go
package main

import (
    "fmt"
    "time"
)

func worker(id int, jobs <-chan int, results chan<- int) {
    for job := range jobs {
        fmt.Printf("Worker %d processing job %d\n", id, job)
        time.Sleep(1 * time.Second)
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
    
    // Send jobs
    for j := 1; j <= 9; j++ {
        jobs <- j
    }
    close(jobs)
    
    // Collect results
    for a := 1; a <= 9; a++ {
        <-results
    }
}
```

**WaitGroup for Synchronization**:
```go
package main

import (
    "fmt"
    "sync"
    "time"
)

func main() {
    var wg sync.WaitGroup
    
    for i := 1; i <= 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            fmt.Printf("Goroutine %d starting\n", id)
            time.Sleep(1 * time.Second)
            fmt.Printf("Goroutine %d done\n", id)
        }(i)
    }
    
    fmt.Println("Waiting for goroutines...")
    wg.Wait()
    fmt.Println("All goroutines completed")
}
```

## Level 2: Advanced Patterns (Medium Freedom)

### 1. gRPC Services

gRPC enables high-performance RPC:

**Protocol Buffer Definition**:
```protobuf
// user.proto
syntax = "proto3";

package user;

message User {
    int32 id = 1;
    string name = 2;
    string email = 3;
}

message GetUserRequest {
    int32 id = 1;
}

message ListUsersRequest {
    int32 limit = 1;
    int32 offset = 2;
}

message ListUsersResponse {
    repeated User users = 1;
    int32 total = 2;
}

service UserService {
    rpc GetUser(GetUserRequest) returns (User);
    rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
    rpc CreateUser(User) returns (User);
}
```

**gRPC Server Implementation**:
```go
package main

import (
    "context"
    "log"
    "net"
    "google.golang.org/grpc"
    pb "your-module/gen/proto/user"
)

type server struct {
    pb.UnimplementedUserServiceServer
}

func (s *server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    // Database lookup
    return &pb.User{
        Id:    req.Id,
        Name:  "John Doe",
        Email: "john@example.com",
    }, nil
}

func (s *server) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
    users := []*pb.User{
        {Id: 1, Name: "John", Email: "john@example.com"},
        {Id: 2, Name: "Jane", Email: "jane@example.com"},
    }
    
    return &pb.ListUsersResponse{
        Users: users,
        Total: 2,
    }, nil
}

func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }
    
    s := grpc.NewServer()
    pb.RegisterUserServiceServer(s, &server{})
    
    log.Println("gRPC server listening on :50051")
    if err := s.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}
```

### 2. Database with sqlc

sqlc generates type-safe database code:

**SQL Queries**:
```sql
-- queries/user.sql

-- name: GetUserByID :one
SELECT id, name, email FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT id, name, email FROM users
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: CreateUser :one
INSERT INTO users (name, email)
VALUES ($1, $2)
RETURNING id, name, email;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;
```

**Using Generated Code**:
```go
package main

import (
    "context"
    "log"
    "your-module/db"
    "github.com/jackc/pgx/v5"
)

func main() {
    conn, err := pgx.Connect(context.Background(), "postgres://user:password@localhost/mydb")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close(context.Background())
    
    queries := db.New(conn)
    
    // Get user
    user, err := queries.GetUserByID(context.Background(), 1)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("User: %+v", user)
    
    // Create user
    newUser, err := queries.CreateUser(context.Background(), db.CreateUserParams{
        Name:  "Alice",
        Email: "alice@example.com",
    })
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Created: %+v", newUser)
}
```

### 3. Testing with testify

testify provides assertion helpers:

**Unit Tests**:
```go
package main

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func Add(a, b int) int {
    return a + b
}

func TestAdd(t *testing.T) {
    // Basic assertions
    assert.Equal(t, 5, Add(2, 3))
    assert.NotEqual(t, 5, Add(2, 2))
    
    // Require (fails immediately)
    require.True(t, Add(1, 1) == 2)
}

func TestAddMultiple(t *testing.T) {
    tests := []struct {
        a, b, expected int
    }{
        {1, 1, 2},
        {0, 0, 0},
        {-1, 1, 0},
    }
    
    for _, test := range tests {
        assert.Equal(t, test.expected, Add(test.a, test.b))
    }
}
```

**HTTP Tests**:
```go
package main

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "github.com/stretchr/testify/assert"
)

func TestGetUser(t *testing.T) {
    req, _ := http.NewRequest("GET", "/users/1", nil)
    w := httptest.NewRecorder()
    
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"id": 1, "name": "John"}`))
    })
    
    handler.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Body.String(), "John")
}
```

## Level 3: Production Deployment (Low Freedom, Expert Only)

### 1. Docker Deployment

**Multi-stage Build**:
```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/app .
EXPOSE 8080

CMD ["./app"]
```

### 2. Graceful Shutdown

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    server := &http.Server{
        Addr:    ":8080",
        Handler: http.DefaultServeMux,
    }
    
    // Handle shutdown signals
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        log.Println("Shutdown signal received")
        
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        
        if err := server.Shutdown(ctx); err != nil {
            log.Fatalf("Shutdown error: %v", err)
        }
    }()
    
    log.Println("Server starting on :8080")
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatalf("Server error: %v", err)
    }
    
    log.Println("Server stopped")
}
```

## Auto-Load Triggers

This Skill activates when you:
- Work with Go projects and REST APIs
- Need to implement gRPC services
- Handle goroutine orchestration
- Use context for cancellation
- Implement microservices
- Debug concurrent systems

## Best Practices Summary

1. Always use context for goroutine cancellation
2. Implement proper error handling with wrapping
3. Use goroutine pools for resource management
4. Validate input early in handlers
5. Test concurrent code thoroughly
6. Monitor goroutine leaks
7. Use type-safe database code with sqlc
8. Implement proper logging
9. Graceful shutdown for long-running services
10. Profile performance-critical code

