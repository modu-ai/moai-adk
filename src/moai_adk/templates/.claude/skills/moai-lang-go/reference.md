# Go 1.25 API Reference

Complete reference for Go systems development with November 2025 tool versions and best practices.

---

## Tool Versions (November 2025)

| Tool | Version | Release | Support |
|------|---------|---------|---------|
| **Go** | 1.25.4 | Nov 2025 | Aug 2026 |
| **golangci-lint** | 1.62.x | 2025 | Active |
| **gotestsum** | 1.12.x | 2025 | Active |
| **govulncheck** | Latest | 2025 | Active |
| **protoc** | 3.21.x | 2025 | Active |

---

## Go Type System

### Basic Types
```go
// Primitives
var name string = "John"
var age int = 30
var price float64 = 19.99
var active bool = true

// Arrays
var numbers [5]int
var names []string  // slice

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

// Function types
type HandlerFunc func(http.ResponseWriter, *http.Request)
```

### Pointers & References
```go
// Pointer declaration
var ptr *User

// Taking address
user := User{ID: 1, Name: "John"}
ptr = &user

// Dereferencing
fmt.Println(ptr.Name)      // Equivalent to (*ptr).Name

// Function parameters
func Update(u *User) {      // Pass by reference
    u.Name = "Updated"
}

func Read(u User) User {    // Pass by value
    return u
}
```

---

## Fiber Web Framework API

### Route Handlers
```go
// Basic handler
app.Get("/path", func(c fiber.Ctx) error {
    return c.SendString("Hello")
})

// JSON response
app.Get("/json", func(c fiber.Ctx) error {
    return c.JSON(fiber.Map{"key": "value"})
})

// Path parameters
app.Get("/users/:id", func(c fiber.Ctx) error {
    id := c.Params("id")
    return c.SendString(id)
})

// Query parameters
app.Get("/search", func(c fiber.Ctx) error {
    q := c.Query("q")
    return c.SendString(q)
})

// Request body
app.Post("/data", func(c fiber.Ctx) error {
    var data map[string]interface{}
    if err := c.BodyParser(&data); err != nil {
        return err
    }
    return c.JSON(data)
})

// HTTP status codes
app.Get("/status", func(c fiber.Ctx) error {
    return c.Status(fiber.StatusOK).SendString("OK")
})
```

### Middleware
```go
// Global middleware
app.Use(middleware.Logger())
app.Use(middleware.Recover())

// Route-specific middleware
app.Get("/protected", authMiddleware, handler)

// Custom middleware
func customMiddleware(c fiber.Ctx) error {
    fmt.Println("Before handler")
    err := c.Next()
    fmt.Println("After handler")
    return err
}
```

---

## Context Package

### Creating Context
```go
import "context"

// Background context
ctx := context.Background()

// TODO context
ctx := context.TODO()

// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// With deadline
deadline := time.Now().Add(10 * time.Second)
ctx, cancel := context.WithDeadline(context.Background(), deadline)
defer cancel()

// With value
ctx = context.WithValue(ctx, "user_id", "123")
```

### Using Context
```go
// Check if context is done
select {
case <-ctx.Done():
    fmt.Println("Cancelled:", ctx.Err())
case <-time.After(1 * time.Second):
    fmt.Println("Done")
}

// Get value from context
userID := ctx.Value("user_id")

// Type assertion
if uid, ok := userID.(string); ok {
    fmt.Println(uid)
}
```

---

## Goroutines & Channels

### Goroutine Creation
```go
// Basic goroutine
go func() {
    fmt.Println("Running concurrently")
}()

// Wait for completion
var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    fmt.Println("Task")
}()
wg.Wait()
```

### Channel Operations
```go
// Create channel
ch := make(chan int)
ch := make(chan int, 10)  // buffered

// Send value
ch <- 42

// Receive value
value := <-ch

// Close channel
close(ch)

// Range over channel
for value := range ch {
    fmt.Println(value)
}

// Non-blocking receive
select {
case value := <-ch:
    fmt.Println(value)
case <-time.After(1 * time.Second):
    fmt.Println("Timeout")
}
```

---

## Standard Library Essentials

### HTTP Server
```go
import "net/http"

// Simple server
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello"))
})
http.ListenAndServe(":8080", nil)

// Custom server
server := &http.Server{
    Addr:         ":8080",
    Handler:      handler,
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 15 * time.Second,
}
server.ListenAndServe()
```

### JSON Encoding
```go
import "encoding/json"

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// Marshal (Go → JSON)
user := User{ID: 1, Name: "John", Email: "john@example.com"}
jsonBytes, _ := json.Marshal(user)

// Unmarshal (JSON → Go)
var u User
json.Unmarshal(jsonBytes, &u)

// Pretty print
prettyJSON, _ := json.MarshalIndent(user, "", "  ")
fmt.Println(string(prettyJSON))
```

### File I/O
```go
import "os"

// Read file
data, _ := os.ReadFile("file.txt")
fmt.Println(string(data))

// Write file
os.WriteFile("file.txt", []byte("content"), 0644)

// Create file
file, _ := os.Create("new.txt")
defer file.Close()
file.WriteString("Hello")
```

---

## Database Access (sqlc + pgx)

### Connection Setup
```go
import "github.com/jackc/pgx/v5/pgxpool"

pool, _ := pgxpool.New(context.Background(),
    "postgresql://user:pass@localhost/db")
defer pool.Close()

db := New(pool)
```

### CRUD Operations
```go
// Create
user, _ := db.CreateUser(ctx, CreateUserParams{
    Name:  "John",
    Email: "john@example.com",
})

// Read
user, _ := db.GetUser(ctx, 1)

// Update
db.UpdateUser(ctx, UpdateUserParams{
    ID:    1,
    Name:  "Jane",
})

// Delete
db.DeleteUser(ctx, 1)

// List
users, _ := db.ListUsers(ctx)
```

---

## Testing

### Unit Tests
```go
import "testing"

func TestAdd(t *testing.T) {
    result := Add(2, 3)
    expected := 5

    if result != expected {
        t.Errorf("Expected %d, got %d", expected, result)
    }
}

func TestAddParallel(t *testing.T) {
    t.Run("positive", func(t *testing.T) {
        if Add(2, 3) != 5 {
            t.Fail()
        }
    })
}
```

### Table-Driven Tests
```go
func TestCalculate(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"2+3", 2, 3, 5},
        {"0+0", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if Calculate(tt.a, tt.b) != tt.expected {
                t.Fail()
            }
        })
    }
}
```

### Benchmarks
```go
func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Add(2, 3)
    }
}

// Run: go test -bench=.
```

---

## Common Commands

```bash
# Build and run
go run main.go
go build -o myapp main.go
go install ./...

# Testing
go test ./...
go test -v ./...
go test -cover ./...
go test -bench=.
go test -benchmem

# Format & lint
gofmt -w .
go vet ./...
golangci-lint run

# Module management
go mod init github.com/user/project
go mod tidy
go mod download
go get -u github.com/package@latest

# Database
sqlc generate
```

---

## Error Handling Best Practices

```go
// Check errors immediately
result, err := operation()
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Type assertion for custom errors
if ve, ok := err.(ValidationError); ok {
    fmt.Println("Validation failed:", ve.Message)
}

// Sentinel errors
var ErrNotFound = errors.New("not found")

if err == ErrNotFound {
    fmt.Println("Item not found")
}
```

---

## Performance Tips

1. **Use buffered channels** to prevent goroutine blocking
2. **Avoid unnecessary allocations** in hot paths
3. **Use sync.Pool** for frequently allocated objects
4. **Profile with pprof** before optimization
5. **Use context for timeouts** in concurrent operations
6. **Implement connection pooling** for databases
7. **Use goroutine limits** to prevent resource exhaustion
8. **Cache frequently used values** appropriately

---

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Deadlock | Ensure channels are closed, use buffered channels |
| Memory leak | Defer close channels, use context cancellation |
| Race condition | Use sync.Mutex or channels, run with -race flag |
| Module not found | Run `go mod tidy`, verify go.mod |
| Connection refused | Check server is running, verify port |
| Timeout errors | Increase timeout, check network |
| Type mismatch | Check JSON tags, use json.RawMessage for flexibility |

