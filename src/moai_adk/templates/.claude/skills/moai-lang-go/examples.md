# Go 1.25 Code Examples

Production-ready examples for Go systems and network programming with Go 1.25.4, Fiber v3, gRPC, context patterns, and goroutine orchestration.

---

## Example 1: Basic HTTP Server with Fiber

### Simple REST API
```go
package main

import "github.com/gofiber/fiber/v3"

func main() {
    app := fiber.New()

    // GET handler
    app.Get("/hello", func(c fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "message": "Hello, World!",
        })
    })

    // POST handler
    app.Post("/users", func(c fiber.Ctx) error {
        type User struct {
            Name  string `json:"name"`
            Email string `json:"email"`
        }

        var user User
        if err := c.BodyParser(&user); err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
                "error": err.Error(),
            })
        }

        return c.Status(fiber.StatusCreated).JSON(user)
    })

    app.Listen(":8080")
}
```

### Route Parameters & Query Strings
```go
// Route with parameters
app.Get("/users/:id", func(c fiber.Ctx) error {
    id := c.Params("id")
    return c.JSON(fiber.Map{"id": id})
})

// Query string
app.Get("/search", func(c fiber.Ctx) error {
    query := c.Query("q")
    limit := c.Query("limit", "10")
    return c.JSON(fiber.Map{
        "query": query,
        "limit": limit,
    })
})
```

---

## Example 2: Context & Cancellation

### Timeout Context
```go
package main

import (
    "context"
    "fmt"
    "time"
)

func fetch(ctx context.Context, url string) string {
    // Simulate work with context
    select {
    case <-time.After(2 * time.Second):
        return fmt.Sprintf("Fetched: %s", url)
    case <-ctx.Done():
        return fmt.Sprintf("Cancelled: %v", ctx.Err())
    }
}

func main() {
    // Create timeout context
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()

    result := fetch(ctx, "http://example.com")
    fmt.Println(result) // Output: Cancelled: context deadline exceeded
}
```

### Context with Values
```go
type userKey struct{}

func setUserInContext(ctx context.Context, userID string) context.Context {
    return context.WithValue(ctx, userKey{}, userID)
}

func getUserFromContext(ctx context.Context) string {
    userID, ok := ctx.Value(userKey{}).(string)
    if !ok {
        return ""
    }
    return userID
}

func main() {
    ctx := setUserInContext(context.Background(), "user123")
    userID := getUserFromContext(ctx)
    fmt.Println(userID) // Output: user123
}
```

---

## Example 3: Goroutines & Channels

### Worker Pool Pattern
```go
package main

import (
    "fmt"
    "sync"
)

func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
    defer wg.Done()

    for job := range jobs {
        fmt.Printf("Worker %d processing job %d\n", id, job)
        results <- job * 2
    }
}

func main() {
    jobs := make(chan int, 100)
    results := make(chan int, 100)
    var wg sync.WaitGroup

    // Start 3 workers
    for w := 1; w <= 3; w++ {
        wg.Add(1)
        go worker(w, jobs, results, &wg)
    }

    // Send 9 jobs
    for j := 1; j <= 9; j++ {
        jobs <- j
    }
    close(jobs)

    // Wait for workers to complete
    wg.Wait()

    // Collect results
    close(results)
    for result := range results {
        fmt.Println(result)
    }
}
```

### Fan-Out / Fan-In
```go
func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        for _, n := range nums {
            out <- n
        }
        close(out)
    }()
    return out
}

func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            out <- n * n
        }
        close(out)
    }()
    return out
}

func merge(channels ...<-chan int) <-chan int {
    var wg sync.WaitGroup
    out := make(chan int)

    multiplexer := func(c <-chan int) {
        defer wg.Done()
        for n := range c {
            out <- n
        }
    }

    wg.Add(len(channels))
    for _, c := range channels {
        go multiplexer(c)
    }

    go func() {
        wg.Wait()
        close(out)
    }()
    return out
}
```

---

## Example 4: Type-Safe SQL with sqlc

### Database Schema
```sql
-- users.sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- queries.sql
-- name: GetUser :one
SELECT id, name, email, created_at FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT id, name, email, created_at FROM users ORDER BY id;

-- name: CreateUser :one
INSERT INTO users (name, email) VALUES ($1, $2)
RETURNING id, name, email, created_at;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;
```

### Generated Code Usage
```go
package main

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
)

func main() {
    // Setup database connection
    pool, _ := pgxpool.New(context.Background(), "postgres://...")
    defer pool.Close()

    db := New(pool)
    ctx := context.Background()

    // Create user
    user, _ := db.CreateUser(ctx, CreateUserParams{
        Name:  "John Doe",
        Email: "john@example.com",
    })

    // Get user
    fetchedUser, _ := db.GetUser(ctx, user.ID)

    // List users
    users, _ := db.ListUsers(ctx)

    // Delete user
    db.DeleteUser(ctx, user.ID)
}
```

---

## Example 5: Middleware & Error Handling

### Custom Middleware
```go
package main

import (
    "fmt"
    "github.com/gofiber/fiber/v3"
)

// Logger middleware
func LoggerMiddleware(c fiber.Ctx) error {
    fmt.Printf("[%s] %s %s\n", c.Method(), c.Path(), c.IP())
    return c.Next()
}

// Authentication middleware
func AuthMiddleware(c fiber.Ctx) error {
    token := c.Get("Authorization")
    if token == "" {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Missing token",
        })
    }

    // Validate token
    if !isValidToken(token) {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Invalid token",
        })
    }

    return c.Next()
}

func main() {
    app := fiber.New()

    // Register middleware
    app.Use(LoggerMiddleware)

    // Protected routes
    app.Post("/protected", AuthMiddleware, func(c fiber.Ctx) error {
        return c.JSON(fiber.Map{"status": "ok"})
    })

    app.Listen(":8080")
}

func isValidToken(token string) bool {
    return len(token) > 0 // Simplified
}
```

---

## Example 6: Struct & Methods

### Struct Definition
```go
package main

import (
    "fmt"
    "time"
)

type User struct {
    ID        int
    Name      string
    Email     string
    CreatedAt time.Time
}

// Method receiver (value receiver)
func (u User) String() string {
    return fmt.Sprintf("User{ID: %d, Name: %s, Email: %s}", u.ID, u.Name, u.Email)
}

// Method receiver (pointer receiver - for mutations)
func (u *User) UpdateEmail(newEmail string) {
    u.Email = newEmail
}

// Interface
type Repository interface {
    GetUser(id int) (*User, error)
    CreateUser(user *User) error
}

// Implementation
type PostgresRepository struct {
    // connection details
}

func (r *PostgresRepository) GetUser(id int) (*User, error) {
    // Query database
    return &User{ID: id}, nil
}

func (r *PostgresRepository) CreateUser(user *User) error {
    // Insert into database
    return nil
}
```

---

## Example 7: Error Handling

### Custom Error Types
```go
package main

import (
    "errors"
    "fmt"
)

// Custom error type
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

// Function that returns error
func validateEmail(email string) error {
    if email == "" {
        return ValidationError{
            Field:   "email",
            Message: "email cannot be empty",
        }
    }
    return nil
}

func main() {
    err := validateEmail("")

    // Type assertion
    if ve, ok := err.(ValidationError); ok {
        fmt.Printf("Validation failed: %s\n", ve.Message)
    }

    // Error comparison
    if errors.Is(err, ValidationError{}) {
        fmt.Println("Is validation error")
    }
}
```

---

## Example 8: Testing

### Unit Tests
```go
package math

import "testing"

func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive", 2, 3, 5},
        {"negative", -2, -3, -5},
        {"mixed", 2, -3, -1},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("Add(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
            }
        })
    }
}

// Benchmark
func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Add(2, 3)
    }
}
```

### Table-Driven Tests
```go
func TestCalculate(t *testing.T) {
    cases := []struct {
        input    string
        expected int
        err      bool
    }{
        {"1+2", 3, false},
        {"5*5", 25, false},
        {"invalid", 0, true},
    }

    for _, c := range cases {
        result, err := Calculate(c.input)
        if (err != nil) != c.err {
            t.Errorf("Calculate(%q) error = %v, wantErr %v", c.input, err, c.err)
        }
        if result != c.expected {
            t.Errorf("Calculate(%q) = %d, want %d", c.input, result, c.expected)
        }
    }
}
```

---

**Learn More**: See `reference.md` for API details, tool versions, and troubleshooting guides.

