---
name: moai-lang-go
version: 4.0.0
updated: '2025-11-19'
status: stable
stability: stable
description: Enterprise Go for systems and network programming Go 1.25.4, Fiber v3,
allowed-tools:
- Read
- Bash
- WebSearch
- WebFetch
---



# Go Systems Development — Enterprise  

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

---

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

---

## Level 3: Advanced Features

### gRPC Services

**Protocol Buffer Definition**:
```protobuf
// user.proto
syntax = "proto3";

package user;

option go_package = "./pb";

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
}

message GetUserRequest {
  int32 id = 1;
}

message GetUserResponse {
  int32 id = 1;
  string name = 2;
  string email = 3;
}

message ListUsersRequest {}
message ListUsersResponse {
  repeated User users = 1;
}
```

**Go gRPC Server**:
```go
type server struct {
    pb.UnimplementedUserServiceServer
}

func (s *server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    // Database lookup logic
    return &pb.GetUserResponse{
        Id:    req.Id,
        Name:  "John Doe",
        Email: "john@example.com",
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

### Advanced Concurrency Patterns

**Fan-Out/Fan-In**:
```go
func fanIn(input1, input2 <-chan string) <-chan string {
    output := make(chan string)
    
    go func() {
        defer close(output)
        for {
            select {
            case s := <-input1:
                output <- s
            case s := <-input2:
                output <- s
            case <-time.After(time.Second):
                return
            }
        }
    }()
    
    return output
}

func main() {
    ch1 := make(chan string)
    ch2 := make(chan string)
    
    // Send data to channels
    go func() {
        for i := 0; i < 5; i++ {
            ch1 <- fmt.Sprintf("Channel 1: %d", i)
            time.Sleep(100 * time.Millisecond)
        }
        close(ch1)
    }()
    
    go func() {
        for i := 0; i < 5; i++ {
            ch2 <- fmt.Sprintf("Channel 2: %d", i)
            time.Sleep(150 * time.Millisecond)
        }
        close(ch2)
    }()
    
    // Receive from combined channel
    for msg := range fanIn(ch1, ch2) {
        fmt.Println(msg)
    }
}
```

### Testing with Testify

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
)

type UserTestSuite struct {
    suite.Suite
    user *User
}

func (suite *UserTestSuite) SetupTest() {
    suite.user = &User{ID: 1, Name: "John", Email: "john@example.com"}
}

func (suite *UserTestSuite) TestUserCreation() {
    assert.Equal(suite.T(), 1, suite.user.ID)
    assert.Equal(suite.T(), "John", suite.user.Name)
    assert.Equal(suite.T(), "john@example.com", suite.user.Email)
}

func TestUserTestSuite(t *testing.T) {
    suite.Run(t, new(UserTestSuite))
}
```

---

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

---

**Version**: 4.0.0 Enterprise  
**Last Updated**: 2025-11-13  
**Status**: Production Ready
