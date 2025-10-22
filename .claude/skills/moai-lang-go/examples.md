# Go 1.24 Code Examples

Production-ready examples for modern Go development with go test, table-driven tests, golangci-lint 1.64.7, gofmt, and Go 1.24 features.

---

## Example 1: Table-Driven Tests with Subtests

### Test File: `user_service_test.go`

```go
package user

import (
    "testing"
)

// @TEST:USER-001 | SPEC: SPEC-USER-001.md | CODE: user_service.go
func TestUserValidation(t *testing.T) {
    tests := []struct {
        name    string
        user    User
        wantErr bool
    }{
        {
            name: "valid user",
            user: User{
                Name:  "Alice",
                Email: "alice@example.com",
                Age:   25,
            },
            wantErr: false,
        },
        {
            name: "empty name",
            user: User{
                Name:  "",
                Email: "test@example.com",
                Age:   30,
            },
            wantErr: true,
        },
        {
            name: "invalid email",
            user: User{
                Name:  "Bob",
                Email: "invalid-email",
                Age:   35,
            },
            wantErr: true,
        },
        {
            name: "negative age",
            user: User{
                Name:  "Charlie",
                Email: "charlie@example.com",
                Age:   -5,
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateUser(tt.user)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateUser() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Implementation: `user_service.go`

```go
package user

import (
    "errors"
    "regexp"
)

// @CODE:USER-001 | SPEC: SPEC-USER-001.md | TEST: user_service_test.go
type User struct {
    Name  string
    Email string
    Age   int
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func ValidateUser(u User) error {
    if u.Name == "" {
        return errors.New("name cannot be empty")
    }
    if !emailRegex.MatchString(u.Email) {
        return errors.New("invalid email format")
    }
    if u.Age < 0 {
        return errors.New("age must be non-negative")
    }
    return nil
}
```

**Key Features**:
- ✅ Table-driven test pattern with subtests
- ✅ Descriptive test names for better failure reporting
- ✅ Anonymous struct for inline test case definition
- ✅ `t.Run()` for isolated subtest execution
- ✅ Clear error messages with context

**Run Commands**:
```bash
go test -v ./...                    # Run all tests with verbose output
go test -run TestUserValidation     # Run specific test
go test -cover                      # Show coverage
go test -coverprofile=coverage.out  # Generate coverage profile
go tool cover -html=coverage.out    # View coverage in browser
```

---

## Example 2: TDD Workflow with HTTP Handler

### RED Phase: Write Failing Test First

```go
// handlers/user_handler_test.go
package handlers

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
)

// @TEST:API-001 | SPEC: SPEC-API-001.md | CODE: user_handler.go
func TestCreateUserHandler(t *testing.T) {
    tests := []struct {
        name         string
        payload      string
        expectedCode int
        expectedBody string
    }{
        {
            name:         "valid user creation",
            payload:      `{"name":"Alice","email":"alice@example.com","age":25}`,
            expectedCode: http.StatusCreated,
            expectedBody: `"id":`,
        },
        {
            name:         "invalid JSON",
            payload:      `{"name":"Bob"`,
            expectedCode: http.StatusBadRequest,
            expectedBody: "invalid request payload",
        },
        {
            name:         "missing required field",
            payload:      `{"email":"charlie@example.com"}`,
            expectedCode: http.StatusBadRequest,
            expectedBody: "name cannot be empty",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(tt.payload))
            req.Header.Set("Content-Type", "application/json")
            w := httptest.NewRecorder()

            CreateUserHandler(w, req)

            if w.Code != tt.expectedCode {
                t.Errorf("expected status %d, got %d", tt.expectedCode, w.Code)
            }

            if !strings.Contains(w.Body.String(), tt.expectedBody) {
                t.Errorf("expected body to contain %q, got %q", tt.expectedBody, w.Body.String())
            }
        })
    }
}
```

### GREEN Phase: Implement Minimum Code to Pass

```go
// handlers/user_handler.go
package handlers

import (
    "encoding/json"
    "net/http"
    "user/models"
)

// @CODE:API-001 | SPEC: SPEC-API-001.md | TEST: user_handler_test.go
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    var user models.User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "invalid request payload", http.StatusBadRequest)
        return
    }

    if err := user.Validate(); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    user.ID = generateID() // Simple ID generation

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}

func generateID() int {
    // Placeholder implementation
    return 1
}
```

### REFACTOR Phase: Improve Code Quality

```go
// handlers/user_handler.go (refactored)
package handlers

import (
    "encoding/json"
    "net/http"
    "user/models"
    "user/services"
)

type UserService interface {
    Create(user models.User) (*models.User, error)
}

type UserHandler struct {
    service UserService
}

func NewUserHandler(service UserService) *UserHandler {
    return &UserHandler{service: service}
}

// @CODE:API-001 | SPEC: SPEC-API-001.md | TEST: user_handler_test.go
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var user models.User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        respondWithError(w, http.StatusBadRequest, "invalid request payload")
        return
    }

    created, err := h.service.Create(user)
    if err != nil {
        respondWithError(w, http.StatusBadRequest, err.Error())
        return
    }

    respondWithJSON(w, http.StatusCreated, created)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    response, _ := json.Marshal(payload)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
    respondWithJSON(w, code, map[string]string{"error": message})
}
```

**Refactoring Improvements**:
- ✅ Dependency injection via interface
- ✅ Extracted helper functions for DRY principle
- ✅ Struct-based handler for better testability
- ✅ Consistent error response format

---

## Example 3: Interface-Based Design with Mocks

### Interface Definition

```go
// storage/user_repository.go
package storage

import "user/models"

// @CODE:REPO-001 | SPEC: SPEC-REPO-001.md | TEST: user_repository_test.go
type UserRepository interface {
    Create(user models.User) (*models.User, error)
    FindByID(id int) (*models.User, error)
    FindByEmail(email string) (*models.User, error)
    Update(user models.User) error
    Delete(id int) error
}
```

### Production Implementation

```go
// storage/postgres_repository.go
package storage

import (
    "database/sql"
    "user/models"
)

type PostgresUserRepository struct {
    db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
    return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(user models.User) (*models.User, error) {
    query := `INSERT INTO users (name, email, age) VALUES ($1, $2, $3) RETURNING id`
    err := r.db.QueryRow(query, user.Name, user.Email, user.Age).Scan(&user.ID)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *PostgresUserRepository) FindByID(id int) (*models.User, error) {
    var user models.User
    query := `SELECT id, name, email, age FROM users WHERE id = $1`
    err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Name, &user.Email, &user.Age)
    if err != nil {
        return nil, err
    }
    return &user, nil
}
```

### Mock Implementation for Testing

```go
// storage/mock_repository.go
package storage

import "user/models"

type MockUserRepository struct {
    CreateFunc    func(user models.User) (*models.User, error)
    FindByIDFunc  func(id int) (*models.User, error)
    FindByEmailFunc func(email string) (*models.User, error)
}

func (m *MockUserRepository) Create(user models.User) (*models.User, error) {
    if m.CreateFunc != nil {
        return m.CreateFunc(user)
    }
    return &user, nil
}

func (m *MockUserRepository) FindByID(id int) (*models.User, error) {
    if m.FindByIDFunc != nil {
        return m.FindByIDFunc(id)
    }
    return nil, nil
}

func (m *MockUserRepository) FindByEmail(email string) (*models.User, error) {
    if m.FindByEmailFunc != nil {
        return m.FindByEmailFunc(email)
    }
    return nil, nil
}
```

### Testing with Mock

```go
// services/user_service_test.go
package services

import (
    "errors"
    "testing"
    "user/models"
    "user/storage"
)

func TestUserService_Create(t *testing.T) {
    tests := []struct {
        name      string
        user      models.User
        mockSetup func(m *storage.MockUserRepository)
        wantErr   bool
    }{
        {
            name: "successful creation",
            user: models.User{Name: "Alice", Email: "alice@example.com", Age: 25},
            mockSetup: func(m *storage.MockUserRepository) {
                m.CreateFunc = func(u models.User) (*models.User, error) {
                    u.ID = 1
                    return &u, nil
                }
            },
            wantErr: false,
        },
        {
            name: "repository error",
            user: models.User{Name: "Bob", Email: "bob@example.com", Age: 30},
            mockSetup: func(m *storage.MockUserRepository) {
                m.CreateFunc = func(u models.User) (*models.User, error) {
                    return nil, errors.New("database error")
                }
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := &storage.MockUserRepository{}
            tt.mockSetup(mockRepo)

            service := NewUserService(mockRepo)
            _, err := service.Create(tt.user)

            if (err != nil) != tt.wantErr {
                t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Key Features**:
- ✅ Interface-based design for loose coupling
- ✅ Manual mock with function fields for flexibility
- ✅ Dependency injection for testability
- ✅ Clean separation of concerns

---

## Example 4: Goroutines and Concurrency

### Concurrent HTTP Requests

```go
// services/api_aggregator.go
package services

import (
    "context"
    "fmt"
    "sync"
    "time"
)

// @CODE:API-002 | SPEC: SPEC-API-002.md | TEST: api_aggregator_test.go
type Result struct {
    Source string
    Data   interface{}
    Error  error
}

func FetchFromMultipleSources(ctx context.Context, urls []string) []Result {
    results := make([]Result, len(urls))
    var wg sync.WaitGroup

    for i, url := range urls {
        wg.Add(1)
        go func(index int, source string) {
            defer wg.Done()
            results[index] = fetchWithTimeout(ctx, source)
        }(i, url)
    }

    wg.Wait()
    return results
}

func fetchWithTimeout(ctx context.Context, url string) Result {
    type response struct {
        data interface{}
        err  error
    }

    ch := make(chan response, 1)

    go func() {
        // Simulate HTTP request
        time.Sleep(100 * time.Millisecond)
        ch <- response{data: fmt.Sprintf("data from %s", url), err: nil}
    }()

    select {
    case res := <-ch:
        return Result{Source: url, Data: res.data, Error: res.err}
    case <-ctx.Done():
        return Result{Source: url, Error: ctx.Err()}
    }
}
```

### Testing Concurrent Code

```go
// services/api_aggregator_test.go
package services

import (
    "context"
    "testing"
    "time"
)

func TestFetchFromMultipleSources(t *testing.T) {
    tests := []struct {
        name        string
        urls        []string
        timeout     time.Duration
        expectError bool
    }{
        {
            name:        "successful fetch from multiple sources",
            urls:        []string{"https://api1.com", "https://api2.com", "https://api3.com"},
            timeout:     5 * time.Second,
            expectError: false,
        },
        {
            name:        "timeout context",
            urls:        []string{"https://slow-api.com"},
            timeout:     10 * time.Millisecond,
            expectError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
            defer cancel()

            results := FetchFromMultipleSources(ctx, tt.urls)

            if len(results) != len(tt.urls) {
                t.Errorf("expected %d results, got %d", len(tt.urls), len(results))
            }

            for _, result := range results {
                hasError := result.Error != nil
                if hasError != tt.expectError {
                    t.Errorf("expected error: %v, got error: %v", tt.expectError, hasError)
                }
            }
        })
    }
}
```

**Key Features**:
- ✅ `sync.WaitGroup` for goroutine coordination
- ✅ Context for cancellation and timeout
- ✅ Channel-based communication
- ✅ Safe concurrent access to shared data

---

## Example 5: Benchmarking

### Benchmark Test

```go
// utils/string_utils_test.go
package utils

import (
    "strings"
    "testing"
)

func BenchmarkStringConcatenation(b *testing.B) {
    b.Run("using + operator", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            result := ""
            for j := 0; j < 100; j++ {
                result += "a"
            }
        }
    })

    b.Run("using strings.Builder", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            var builder strings.Builder
            for j := 0; j < 100; j++ {
                builder.WriteString("a")
            }
            _ = builder.String()
        }
    })
}

func BenchmarkMapAccess(b *testing.B) {
    data := make(map[string]int, 1000)
    for i := 0; i < 1000; i++ {
        data[fmt.Sprintf("key%d", i)] = i
    }

    b.ResetTimer() // Exclude setup time from benchmark

    for i := 0; i < b.N; i++ {
        _ = data["key500"]
    }
}
```

**Run Benchmarks**:
```bash
go test -bench=.                        # Run all benchmarks
go test -bench=BenchmarkStringConcat    # Run specific benchmark
go test -bench=. -benchmem              # Include memory allocation stats
go test -bench=. -benchtime=10s         # Run for 10 seconds
go test -bench=. -cpuprofile=cpu.prof   # Generate CPU profile
```

---

## Example 6: Go 1.24 Features

### Tool Dependencies (Go 1.24+)

```bash
# Add development tool as module dependency
go get -tool github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.7

# Run tool via go tool
go tool golangci-lint run

# List all tools
go tool -list
```

### Testing with `b.Loop()` (Go 1.24+)

```go
// Modern benchmark pattern with b.Loop()
func BenchmarkProcessData(b *testing.B) {
    data := generateTestData(1000)

    for b.Loop() {
        ProcessData(data)
    }
}
```

**Benefits of `b.Loop()`**:
- ✅ Automatic timer management
- ✅ More accurate measurements
- ✅ Cleaner syntax than `for i := 0; i < b.N; i++`

---

## CI/CD Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/go.yml
name: Go CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

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

    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v4
      with:
        files: ./coverage.out
        flags: unittests

    - name: Check test coverage
      run: |
        coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        if (( $(echo "$coverage < 85" | bc -l) )); then
          echo "Coverage $coverage% is below 85%"
          exit 1
        fi
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
# Format code with gofmt
gofmt -s -w .

# Run golangci-lint
golangci-lint run
```

### U - Unified (Type Safety)

```go
// Leverage Go's strong type system
type UserID int64
type Email string

func FindUser(id UserID) (*User, error) {
    // Compile-time type safety
}
```

### S - Secured

```bash
# Run security scanner
go tool golangci-lint run --enable gosec

# Check for vulnerabilities
go list -json -m all | nancy sleuth
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

**Version**: 1.0.0 (2025-10-22)
**Updated**: Latest tool versions verified 2025-10-22
**Framework**: MoAI-ADK Go Language Skill
**Status**: Production-ready
