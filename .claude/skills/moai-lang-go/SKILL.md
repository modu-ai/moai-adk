---
name: moai-lang-go
version: 2.0.0
created: 2025-11-06
updated: 2025-11-06
status: active
description: "Go best practices with modern cloud-native development, performance optimization, and concurrent programming for 2025"
keywords: [go, golang, backend, microservices, cloud-native, performance, concurrency, devops]
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - WebFetch
  - WebSearch
---

# Go Development Mastery

**Modern Go Development with 2025 Best Practices**

> Comprehensive Go development guidance covering cloud-native services, high-performance concurrent programming, microservices architecture, and production-ready applications using the latest tools and methodologies.

## What It Does

- **Cloud-Native Services**: Kubernetes, Docker, and container-native applications
- **High-Performance APIs**: Fast HTTP servers with minimal memory footprint
- **Concurrent Programming**: Goroutines, channels, and advanced patterns
- **Microservices Architecture**: Service discovery, load balancing, distributed systems
- **Testing & Quality**: Comprehensive testing, benchmarking, and quality assurance
- **Database Integration**: SQL, NoSQL, caching, and data pipelines
- **DevOps Integration**: CI/CD, monitoring, observability, and deployment
- **Enterprise Patterns**: Clean architecture, domain-driven design, scalability

## When to Use

### Perfect Scenarios
- **Building microservices and distributed systems**
- **High-performance network services and APIs**
- **Cloud-native applications and DevOps tools**
- **Concurrent and parallel processing applications**
- **System programming and infrastructure tools**
- **CLI applications and automation scripts**
- **Real-time data processing and streaming**

### Common Triggers
- "Create Go microservice"
- "Set up Go API server"
- "Go concurrent programming"
- "Go best practices"
- "Optimize Go performance"
- "Deploy Go application"

## Tool Version Matrix (2025-11-06)

### Core Go
- **Go**: 1.25.x (latest) / 1.23.x (LTS)
- **Gin**: 1.10.x - HTTP web framework
- **Echo**: 4.12.x - High-performance web framework
- **Chi**: 5.0.x - Lightweight router
- **Fiber**: 3.x - Express.js-inspired framework

### Database & Storage
- **GORM**: 2.0.x - ORM for Go
- **sqlx**: 1.4.x - SQL extensions
- **pgx**: 5.6.x - PostgreSQL driver
- **Redis**: 9.x - Redis client
- **MongoDB**: 2.x - MongoDB driver

### Testing Tools
- **Testify**: 1.9.x - Testing toolkit
- **GoMock**: 1.6.x - Mock generation
- **gomock**: Built-in mocking framework
- **goleak**: 1.3.x - Goroutine leak detection

### Development Tools
- **Air**: 1.53.x - Live reload
- **Gin-swagger**: 1.6.x - API documentation
- **Zap**: 1.27.x - Structured logging
- **Viper**: 1.19.x - Configuration management
- **golangci-lint**: 1.62.x - Linter aggregator

### Observability
- **Prometheus**: 4.x client - Metrics
- **Jaeger**: 2.x client - Distributed tracing
- **OpenTelemetry**: 1.30.x - Observability framework
- **pprof**: Built-in profiling

## Ecosystem Overview

### Project Setup (2025 Best Practice)

```bash
# Modern Go project with modules
go mod init github.com/your-org/your-project

# Initialize project structure
mkdir -p {cmd/server,internal/{handler,service,repository},pkg/{utils,middleware},configs,deployments/{docker,k8s},scripts,test}

# Install essential tools
go install github.com/cosmtrek/air@latest
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install go.uber.org/mock/mockgen@latest

# Generate code documentation
swag init -g cmd/server/main.go
```

### Modern Project Structure

```
my-go-project/
├── go.mod
├── go.sum
├── Makefile                  # Build commands and utilities
├── README.md
├── .github/
│   └── workflows/            # CI/CD pipelines
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/                # Private application code
│   ├── handler/             # HTTP handlers
│   ├── service/             # Business logic
│   ├── repository/          # Data access layer
│   ├── model/               # Domain models
│   └── middleware/          # HTTP middleware
├── pkg/                     # Public library code
│   ├── utils/               # Utility functions
│   ├── config/              # Configuration
│   └── logger/              # Logging utilities
├── configs/                 # Configuration files
├── deployments/
│   ├── docker/              # Docker configurations
│   └── k8s/                 # Kubernetes manifests
├── scripts/                 # Build and utility scripts
├── test/                    # Integration and e2e tests
├── docs/                    # API documentation
└── migrations/              # Database migrations
```

## Modern Go Patterns

### Context-First Design

```go
// handler/user.go
package handler

import (
    "context"
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/your-org/your-project/internal/model"
    "github.com/your-org/your-project/internal/service"
)

type UserHandler struct {
    userService service.UserService
    logger      Logger
}

func NewUserHandler(userService service.UserService, logger Logger) *UserHandler {
    return &UserHandler{
        userService: userService,
        logger:      logger,
    }
}

// CreateUser handles user creation requests
func (h *UserHandler) CreateUser(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
    defer cancel()
    
    var req model.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.Error("Invalid request body", "error", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    user, err := h.userService.CreateUser(ctx, req)
    if err != nil {
        h.logger.Error("Failed to create user", "error", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        return
    }
    
    c.JSON(http.StatusCreated, user)
}

// GetUser retrieves a user by ID
func (h *UserHandler) GetUser(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()
    
    userID := c.Param("id")
    user, err := h.userService.GetUser(ctx, userID)
    if err != nil {
        if err == service.ErrUserNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
            return
        }
        
        h.logger.Error("Failed to get user", "userID", userID, "error", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        return
    }
    
    c.JSON(http.StatusOK, user)
}
```

### Service Layer with Error Handling

```go
// service/user.go
package service

import (
    "context"
    "errors"
    "fmt"
    
    "github.com/your-org/your-project/internal/model"
    "github.com/your-org/your-project/internal/repository"
)

var (
    ErrUserNotFound    = errors.New("user not found")
    ErrUserExists      = errors.New("user already exists")
    ErrInvalidInput    = errors.New("invalid input")
)

type UserService interface {
    CreateUser(ctx context.Context, req model.CreateUserRequest) (*model.User, error)
    GetUser(ctx context.Context, id string) (*model.User, error)
    UpdateUser(ctx context.Context, id string, req model.UpdateUserRequest) (*model.User, error)
    DeleteUser(ctx context.Context, id string) error
    ListUsers(ctx context.Context, filter model.UserFilter) ([]*model.User, int, error)
}

type userService struct {
    userRepo repository.UserRepository
    logger   Logger
}

func NewUserService(userRepo repository.UserRepository, logger Logger) UserService {
    return &userService{
        userRepo: userRepo,
        logger:   logger,
    }
}

func (s *userService) CreateUser(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
    // Validate input
    if err := req.Validate(); err != nil {
        s.logger.Warn("Invalid user creation request", "error", err)
        return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
    }
    
    // Check if user already exists
    existing, err := s.userRepo.GetByEmail(ctx, req.Email)
    if err == nil && existing != nil {
        s.logger.Warn("User already exists", "email", req.Email)
        return nil, ErrUserExists
    }
    
    // Create user
    user := &model.User{
        ID:        generateID(),
        Email:     req.Email,
        Name:      req.Name,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    
    if err := s.userRepo.Create(ctx, user); err != nil {
        s.logger.Error("Failed to create user", "error", err)
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    s.logger.Info("User created successfully", "userID", user.ID)
    return user, nil
}

func (s *userService) GetUser(ctx context.Context, id string) (*model.User, error) {
    user, err := s.userRepo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, repository.ErrNotFound) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    
    return user, nil
}
```

### Repository Pattern with SQLx

```go
// repository/user.go
package repository

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq" // PostgreSQL driver
    "github.com/your-org/your-project/internal/model"
)

var (
    ErrNotFound = errors.New("record not found")
    ErrDuplicate = errors.New("duplicate record")
)

type UserRepository interface {
    Create(ctx context.Context, user *model.User) error
    GetByID(ctx context.Context, id string) (*model.User, error)
    GetByEmail(ctx context.Context, email string) (*model.User, error)
    Update(ctx context.Context, user *model.User) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter UserFilter) ([]*model.User, int, error)
}

type userRepository struct {
    db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
    query := `
        INSERT INTO users (id, email, name, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (email) DO NOTHING
    `
    
    result, err := r.db.ExecContext(ctx, query, user.ID, user.Email, user.Name, user.CreatedAt, user.UpdatedAt)
    if err != nil {
        return fmt.Errorf("failed to insert user: %w", err)
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    
    if rowsAffected == 0 {
        return ErrDuplicate
    }
    
    return nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
    query := `
        SELECT id, email, name, created_at, updated_at
        FROM users
        WHERE id = $1
    `
    
    var user model.User
    err := r.db.GetContext(ctx, &user, query, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("failed to get user by ID: %w", err)
    }
    
    return &user, nil
}

func (r *userRepository) List(ctx context.Context, filter UserFilter) ([]*model.User, int, error) {
    // Build dynamic query
    baseQuery := `
        SELECT id, email, name, created_at, updated_at
        FROM users
        WHERE 1=1
    `
    
    countQuery := `
        SELECT COUNT(*)
        FROM users
        WHERE 1=1
    `
    
    var args []interface{}
    conditions := []string{}
    
    if filter.Email != "" {
        conditions = append(conditions, "email = $"+fmt.Sprintf("%d", len(args)+1))
        args = append(args, filter.Email)
    }
    
    if len(conditions) > 0 {
        whereClause := " AND " + fmt.Sprintf(" AND ", conditions)
        baseQuery += whereClause
        countQuery += whereClause
    }
    
    // Get total count
    var total int
    err := r.db.GetContext(ctx, &total, countQuery, args...)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to count users: %w", err)
    }
    
    // Add pagination
    if filter.Limit > 0 {
        baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
        args = append(args, filter.Limit, filter.Offset)
    }
    
    var users []*model.User
    err = r.db.SelectContext(ctx, &users, baseQuery, args...)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to list users: %w", err)
    }
    
    return users, total, nil
}
```

## Concurrent Programming

### Worker Pool Pattern

```go
// pkg/worker/pool.go
package worker

import (
    "context"
    "sync"
    "time"
)

type Task[T any] interface {
    Execute(ctx context.Context) (T, error)
}

type Result[T any] struct {
    Value T
    Error error
}

type WorkerPool[T any] struct {
    workers   int
    taskQueue chan Task[T]
    wg        sync.WaitGroup
    ctx       context.Context
    cancel    context.CancelFunc
}

func NewWorkerPool[T any](workers int) *WorkerPool[T] {
    ctx, cancel := context.WithCancel(context.Background())
    
    return &WorkerPool[T]{
        workers:   workers,
        taskQueue: make(chan Task[T], workers*2),
        ctx:       ctx,
        cancel:    cancel,
    }
}

func (p *WorkerPool[T]) Start() {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go p.worker(i)
    }
}

func (p *WorkerPool[T]) worker(id int) {
    defer p.wg.Done()
    
    for {
        select {
        case task := <-p.taskQueue:
            if task == nil {
                return
            }
            
            result, err := task.Execute(p.ctx)
            if err != nil {
                // Handle error (could use error channel or logging)
                continue
            }
            
            // Process result
            
        case <-p.ctx.Done():
            return
        }
    }
}

func (p *WorkerPool[T]) Submit(task Task[T]) {
    select {
    case p.taskQueue <- task:
    case <-p.ctx.Done():
        return
    }
}

func (p *WorkerPool[T]) Stop() {
    close(p.taskQueue)
    p.cancel()
    p.wg.Wait()
}

// Example usage
type ImageProcessingTask struct {
    ImagePath string
}

func (t *ImageProcessingTask) Execute(ctx context.Context) (string, error) {
    // Simulate image processing
    time.Sleep(100 * time.Millisecond)
    return "processed_" + t.ImagePath, nil
}

func ProcessImages(imagePaths []string) []string {
    pool := NewWorkerPool[string](10)
    pool.Start()
    defer pool.Stop()
    
    var results []string
    var mu sync.Mutex
    
    for _, path := range imagePaths {
        task := &ImageProcessingTask{ImagePath: path}
        
        go func(t *ImageProcessingTask) {
            result, err := t.Execute(context.Background())
            if err == nil {
                mu.Lock()
                results = append(results, result)
                mu.Unlock()
            }
        }(task)
    }
    
    return results
}
```

### Fan-In/Fan-Out Pattern

```go
// pkg/concurrent/fan.go
package concurrent

import (
    "context"
    "sync"
)

// FanIn merges multiple input channels into a single output channel
func FanIn[T any](ctx context.Context, channels ...<-chan T) <-chan T {
    var wg sync.WaitGroup
    out := make(chan T)
    
    output := func(c <-chan T) {
        defer wg.Done()
        for {
            select {
            case v, ok := <-c:
                if !ok {
                    return
                }
                select {
                case out <- v:
                case <-ctx.Done():
                    return
                }
            case <-ctx.Done():
                return
            }
        }
    }
    
    wg.Add(len(channels))
    for _, c := range channels {
        go output(c)
    }
    
    go func() {
        wg.Wait()
        close(out)
    }()
    
    return out
}

// FanOut distributes input across multiple output channels
func FanOut[T any](ctx context.Context, in <-chan T, n int) []<-chan T {
    outs := make([]chan T, n)
    for i := 0; i < n; i++ {
        outs[i] = make(chan T)
    }
    
    distribute := func() {
        defer func() {
            for _, out := range outs {
                close(out)
            }
        }()
        
        for {
            select {
            case v, ok := <-in:
                if !ok {
                    return
                }
                
                for _, out := range outs {
                    select {
                    case out <- v:
                    case <-ctx.Done():
                        return
                    }
                }
                
            case <-ctx.Done():
                return
            }
        }
    }
    
    go distribute()
    return outs
}

// Pipeline combines fan-out and fan-in for parallel processing
func Pipeline[T, R any](
    ctx context.Context,
    input []T,
    workers int,
    process func(context.Context, T) (R, error),
) (<-chan R, <-chan error) {
    out := make(chan R, workers)
    errChan := make(chan error, workers)
    
    // Create input channel
    in := make(chan T, len(input))
    for _, item := range input {
        in <- item
    }
    close(in)
    
    // Fan out to workers
    channels := FanOut(ctx, in, workers)
    
    // Process each item
    var wg sync.WaitGroup
    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan T) {
            defer wg.Done()
            for item := range c {
                result, err := process(ctx, item)
                if err != nil {
                    select {
                    case errChan <- err:
                    case <-ctx.Done():
                        return
                    }
                    continue
                }
                
                select {
                case out <- result:
                case <-ctx.Done():
                    return
                }
            }
        }(ch)
    }
    
    go func() {
        wg.Wait()
        close(out)
        close(errChan)
    }()
    
    return out, errChan
}
```

## Performance Optimization

### Memory Pooling

```go
// pkg/pool/buffer.go
package pool

import (
    "sync"
)

var (
    bufferPool = sync.Pool{
        New: func() interface{} {
            return make([]byte, 0, 1024) // Pre-allocate 1KB
        },
    }
)

// GetBuffer returns a buffer from the pool
func GetBuffer() []byte {
    return bufferPool.Get().([]byte)
}

// PutBuffer returns a buffer to the pool
func PutBuffer(buf []byte) {
    if cap(buf) <= 64*1024 { // Don't pool large buffers
        bufferPool.Put(buf[:0]) // Reset length but keep capacity
    }
}

// Usage example
func ProcessData(data []byte) ([]byte, error) {
    buf := GetBuffer()
    defer PutBuffer(buf)
    
    // Process data using the pooled buffer
    buf = append(buf, data...)
    // ... processing logic ...
    
    result := make([]byte, len(buf))
    copy(result, buf)
    
    return result, nil
}
```

### String Building Optimization

```go
// pkg/utils/strings.go
package utils

import (
    "strings"
    "sync"
)

var stringBuilderPool = sync.Pool{
    New: func() interface{} {
        return &strings.Builder{}
    },
}

// StringBuilderPool provides a thread-safe string builder pool
func GetStringBuilder() *strings.Builder {
    sb := stringBuilderPool.Get().(*strings.Builder)
    sb.Reset()
    return sb
}

func PutStringBuilder(sb *strings.Builder) {
    if sb.Cap() <= 4096 { // Don't pool large builders
        stringBuilderPool.Put(sb)
    }
}

// FastJoin efficiently joins strings using a pooled builder
func FastJoin(strs ...string) string {
    if len(strs) == 0 {
        return ""
    }
    if len(strs) == 1 {
        return strs[0]
    }
    
    sb := GetStringBuilder()
    defer PutStringBuilder(sb)
    
    for _, s := range strs {
        sb.WriteString(s)
    }
    
    return sb.String()
}

// Usage example
func BuildMessage(parts []string) string {
    sb := GetStringBuilder()
    defer PutStringBuilder(sb)
    
    for i, part := range parts {
        if i > 0 {
            sb.WriteString(" ")
        }
        sb.WriteString(part)
    }
    
    return sb.String()
}
```

## Testing Strategies

### Comprehensive Unit Testing

```go
// service/user_test.go
package service

import (
    "context"
    "errors"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
    "github.com/your-org/your-project/internal/model"
    "github.com/your-org/your-project/internal/repository"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
    args := m.Called(ctx, email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.User), args.Error(1)
}

func TestUserService_CreateUser_Success(t *testing.T) {
    // Arrange
    mockRepo := new(MockUserRepository)
    mockLogger := &MockLogger{} // Assume you have this
    service := NewUserService(mockRepo, mockLogger)
    
    req := model.CreateUserRequest{
        Email: "test@example.com",
        Name:  "Test User",
    }
    
    expectedUser := &model.User{
        ID:    "user-123",
        Email: req.Email,
        Name:  req.Name,
    }
    
    mockRepo.On("GetByEmail", mock.Anything, req.Email).Return(nil, repository.ErrNotFound)
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)
    
    // Act
    ctx := context.Background()
    user, err := service.CreateUser(ctx, req)
    
    // Assert
    require.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, req.Email, user.Email)
    assert.Equal(t, req.Name, user.Name)
    mockRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_EmailExists(t *testing.T) {
    // Arrange
    mockRepo := new(MockUserRepository)
    mockLogger := &MockLogger{}
    service := NewUserService(mockRepo, mockLogger)
    
    req := model.CreateUserRequest{
        Email: "existing@example.com",
        Name:  "Test User",
    }
    
    existingUser := &model.User{
        ID:    "existing-123",
        Email: req.Email,
    }
    
    mockRepo.On("GetByEmail", mock.Anything, req.Email).Return(existingUser, nil)
    
    // Act
    ctx := context.Background()
    user, err := service.CreateUser(ctx, req)
    
    // Assert
    require.Error(t, err)
    assert.Nil(t, user)
    assert.Equal(t, ErrUserExists, err)
    mockRepo.AssertExpectations(t)
}

func TestUserService_GetUser_NotFound(t *testing.T) {
    // Arrange
    mockRepo := new(MockUserRepository)
    mockLogger := &MockLogger{}
    service := NewUserService(mockRepo, mockLogger)
    
    userID := "nonexistent"
    mockRepo.On("GetByID", mock.Anything, userID).Return(nil, repository.ErrNotFound)
    
    // Act
    ctx := context.Background()
    user, err := service.GetUser(ctx, userID)
    
    // Assert
    require.Error(t, err)
    assert.Nil(t, user)
    assert.Equal(t, ErrUserNotFound, err)
    mockRepo.AssertExpectations(t)
}

// Benchmark example
func BenchmarkUserService_CreateUser(b *testing.B) {
    mockRepo := new(MockUserRepository)
    mockLogger := &MockLogger{}
    service := NewUserService(mockRepo, mockLogger)
    
    req := model.CreateUserRequest{
        Email: "test@example.com",
        Name:  "Test User",
    }
    
    mockRepo.On("GetByEmail", mock.Anything, req.Email).Return(nil, repository.ErrNotFound)
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)
    
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := service.CreateUser(ctx, req)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Integration Testing with Testcontainers

```go
// test/integration/user_test.go
package integration

import (
    "context"
    "fmt"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
    
    "github.com/your-org/your-project/internal/model"
    "github.com/your-org/your-project/internal/repository"
    "github.com/your-org/your-project/internal/service"
    "github.com/your-org/your-project/pkg/config"
)

func TestUserService_Integration(t *testing.T) {
    // Setup test database container
    ctx := context.Background()
    
    req := testcontainers.ContainerRequest{
        Image:        "postgres:16",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_DB":       "testdb",
            "POSTGRES_USER":     "test",
            "POSTGRES_PASSWORD": "test",
        },
        WaitingFor: wait.ForLog("database system is ready to accept connections"),
    }
    
    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    require.NoError(t, err)
    defer container.Terminate(ctx)
    
    // Get database connection info
    host, err := container.Host(ctx)
    require.NoError(t, err)
    
    port, err := container.MappedPort(ctx, "5432")
    require.NoError(t, err)
    
    // Connect to database
    dbConfig := config.Database{
        Host:     host,
        Port:     port.Int(),
        User:     "test",
        Password: "test",
        DBName:   "testdb",
        SSLMode:  "disable",
    }
    
    db, err := setupTestDB(dbConfig)
    require.NoError(t, err)
    defer db.Close()
    
    // Run migrations
    err = runMigrations(db)
    require.NoError(t, err)
    
    // Setup service
    userRepo := repository.NewUserRepository(db)
    userService := service.NewUserService(userRepo, &MockLogger{})
    
    // Test user creation
    t.Run("Create and Get User", func(t *testing.T) {
        req := model.CreateUserRequest{
            Email: "integration@example.com",
            Name:  "Integration Test User",
        }
        
        // Create user
        user, err := userService.CreateUser(ctx, req)
        require.NoError(t, err)
        assert.NotEmpty(t, user.ID)
        assert.Equal(t, req.Email, user.Email)
        assert.Equal(t, req.Name, user.Name)
        assert.False(t, user.CreatedAt.IsZero())
        
        // Get user
        retrievedUser, err := userService.GetUser(ctx, user.ID)
        require.NoError(t, err)
        assert.Equal(t, user.ID, retrievedUser.ID)
        assert.Equal(t, user.Email, retrievedUser.Email)
        assert.Equal(t, user.Name, retrievedUser.Name)
    })
    
    t.Run("List Users", func(t *testing.T) {
        // Create multiple users
        for i := 0; i < 5; i++ {
            req := model.CreateUserRequest{
                Email: fmt.Sprintf("user%d@example.com", i),
                Name:  fmt.Sprintf("User %d", i),
            }
            _, err := userService.CreateUser(ctx, req)
            require.NoError(t, err)
        }
        
        // List users
        users, total, err := userService.ListUsers(ctx, model.UserFilter{
            Limit:  10,
            Offset: 0,
        })
        
        require.NoError(t, err)
        assert.Equal(t, 6, total) // 5 new users + 1 from previous test
        assert.Equal(t, 6, len(users))
    })
}

// Helper functions
func setupTestDB(cfg config.Database) (*sqlx.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
    )
    
    db, err := sqlx.Connect(context.Background(), "postgres", dsn)
    if err != nil {
        return nil, err
    }
    
    // Wait for database to be ready
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    for {
        err := db.PingContext(ctx)
        if err == nil {
            break
        }
        
        select {
        case <-time.After(1 * time.Second):
            continue
        case <-ctx.Done():
            return nil, ctx.Err()
        }
    }
    
    return db, nil
}
```

## Security Best Practices

### Input Validation and Sanitization

```go
// pkg/validation/validator.go
package validation

import (
    "regexp"
    "strings"
    "unicode"
    
    "github.com/go-playground/validator/v10"
)

var (
    validate = validator.New()
    
    // Email regex pattern
    emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    
    // Strong password regex
    passwordRegex = regexp.MustCompile(`^.{8,}$`) // At least 8 characters
)

// ValidationResult represents validation result
type ValidationResult struct {
    Valid  bool
    Errors map[string]string
}

// New creates a new validator
func New() *Validator {
    return &Validator{}
}

// Validator handles input validation
type Validator struct{}

// ValidateStruct validates a struct using validator tags
func (v *Validator) ValidateStruct(s interface{}) ValidationResult {
    err := validate.Struct(s)
    if err == nil {
        return ValidationResult{Valid: true}
    }
    
    errors := make(map[string]string)
    for _, err := range err.(validator.ValidationErrors) {
        field := err.Field()
        switch err.Tag() {
        case "required":
            errors[field] = field + " is required"
        case "email":
            errors[field] = field + " must be a valid email address"
        case "min":
            errors[field] = field + " must be at least " + err.Param() + " characters"
        case "max":
            errors[field] = field + " must be at most " + err.Param() + " characters"
        default:
            errors[field] = field + " is invalid"
        }
    }
    
    return ValidationResult{Valid: false, Errors: errors}
}

// ValidateEmail validates email format
func (v *Validator) ValidateEmail(email string) bool {
    email = strings.TrimSpace(strings.ToLower(email))
    return emailRegex.MatchString(email) && len(email) <= 254
}

// ValidatePassword validates password strength
func (v *Validator) ValidatePassword(password string) ValidationResult {
    errors := make(map[string]string)
    
    if len(password) < 8 {
        errors["length"] = "Password must be at least 8 characters long"
    }
    
    if !containsUppercase(password) {
        errors["uppercase"] = "Password must contain at least one uppercase letter"
    }
    
    if !containsLowercase(password) {
        errors["lowercase"] = "Password must contain at least one lowercase letter"
    }
    
    if !containsDigit(password) {
        errors["digit"] = "Password must contain at least one digit"
    }
    
    if !containsSpecialChar(password) {
        errors["special"] = "Password must contain at least one special character"
    }
    
    return ValidationResult{
        Valid:  len(errors) == 0,
        Errors: errors,
    }
}

// SanitizeInput sanitizes user input
func (v *Validator) SanitizeInput(input string) string {
    // Remove potentially dangerous characters
    sanitized := strings.ReplaceAll(input, "<", "&lt;")
    sanitized = strings.ReplaceAll(sanitized, ">", "&gt;")
    sanitized = strings.ReplaceAll(sanitized, "&", "&amp;")
    sanitized = strings.ReplaceAll(sanitized, "\"", "&quot;")
    sanitized = strings.ReplaceAll(sanitized, "'", "&#x27;")
    
    // Trim whitespace
    return strings.TrimSpace(sanitized)
}

// Helper functions
func containsUppercase(s string) bool {
    for _, r := range s {
        if unicode.IsUpper(r) {
            return true
        }
    }
    return false
}

func containsLowercase(s string) bool {
    for _, r := range s {
        if unicode.IsLower(r) {
            return true
        }
    }
    return false
}

func containsDigit(s string) bool {
    for _, r := range s {
        if unicode.IsDigit(r) {
            return true
        }
    }
    return false
}

func containsSpecialChar(s string) bool {
    for _, r := range s {
        if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
            return true
        }
    }
    return false
}
```

### Authentication and Authorization

```go
// pkg/auth/jwt.go
package auth

import (
    "fmt"
    "time"
    
    "github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims
type Claims struct {
    UserID   string   `json:"user_id"`
    Email    string   `json:"email"`
    Role     string   `json:"role"`
    jwt.RegisteredClaims
}

// JWTManager handles JWT token operations
type JWTManager struct {
    secretKey  []byte
    expiration time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, expiration time.Duration) *JWTManager {
    return &JWTManager{
        secretKey:  []byte(secretKey),
        expiration: expiration,
    }
}

// GenerateToken generates a new JWT token
func (m *JWTManager) GenerateToken(userID, email, role string) (string, error) {
    claims := &Claims{
        UserID: userID,
        Email:  email,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.expiration)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    "your-app",
            Subject:   userID,
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(m.secretKey)
}

// ValidateToken validates and parses a JWT token
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return m.secretKey, nil
    })
    
    if err != nil {
        return nil, err
    }
    
    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, fmt.Errorf("invalid token")
    }
    
    return claims, nil
}

// RefreshToken generates a new token with extended expiration
func (m *JWTManager) RefreshToken(claims *Claims) (string, error) {
    claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(m.expiration))
    claims.IssuedAt = jwt.NewNumericDate(time.Now())
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(m.secretKey)
}

// Middleware for authentication
func (m *JWTManager) AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Authorization header required", http.StatusUnauthorized)
            return
        }
        
        tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
        claims, err := m.ValidateToken(tokenString)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        
        // Add claims to context
        ctx := context.WithValue(r.Context(), "user_claims", claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Role-based authorization middleware
func (m *JWTManager) RequireRole(requiredRole string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims, ok := r.Context().Value("user_claims").(*Claims)
            if !ok {
                http.Error(w, "User claims not found", http.StatusUnauthorized)
                return
            }
            
            if claims.Role != requiredRole {
                http.Error(w, "Insufficient permissions", http.StatusForbidden)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

## Integration Patterns

### gRPC Service Implementation

```go
// api/proto/user.proto
syntax = "proto3";

package user.v1;

option go_package = "github.com/your-org/your-project/api/user/v1;userv1";

service UserService {
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse);
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse);
}

message User {
  string id = 1;
  string email = 2;
  string name = 3;
  google.protobuf.Timestamp created_at = 4;
  google.protobuf.Timestamp updated_at = 5;
}

message CreateUserRequest {
  string email = 1;
  string name = 2;
}

message CreateUserResponse {
  User user = 1;
}

// server/grpc/user_service.go
package grpc

import (
    "context"
    
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/protobuf/types/known/timestamppb"
    
    "github.com/your-org/your-project/api/user/v1"
    "github.com/your-org/your-project/internal/service"
)

type GRPCUserServer struct {
    userv1.UnimplementedUserServiceServer
    userService service.UserService
}

func NewGRPCUserServer(userService service.UserService) *GRPCUserServer {
    return &GRPCUserServer{
        userService: userService,
    }
}

func (s *GRPCUserServer) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
    // Convert protobuf to domain model
    createReq := model.CreateUserRequest{
        Email: req.GetEmail(),
        Name:  req.GetName(),
    }
    
    // Call service layer
    user, err := s.userService.CreateUser(ctx, createReq)
    if err != nil {
        if errors.Is(err, service.ErrUserExists) {
            return nil, status.Error(codes.AlreadyExists, "user already exists")
        }
        if errors.Is(err, service.ErrInvalidInput) {
            return nil, status.Error(codes.InvalidArgument, err.Error())
        }
        return nil, status.Error(codes.Internal, "internal server error")
    }
    
    // Convert domain model to protobuf
    return &userv1.CreateUserResponse{
        User: s.userToProto(user),
    }, nil
}

func (s *GRPCUserServer) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
    user, err := s.userService.GetUser(ctx, req.GetId())
    if err != nil {
        if errors.Is(err, service.ErrUserNotFound) {
            return nil, status.Error(codes.NotFound, "user not found")
        }
        return nil, status.Error(codes.Internal, "internal server error")
    }
    
    return &userv1.GetUserResponse{
        User: s.userToProto(user),
    }, nil
}

func (s *GRPCUserServer) userToProto(user *model.User) *userv1.User {
    return &userv1.User{
        Id:        user.ID,
        Email:     user.Email,
        Name:      user.Name,
        CreatedAt: timestamppb.New(user.CreatedAt),
        UpdatedAt: timestamppb.New(user.UpdatedAt),
    }
}
```

### Message Queue Integration

```go
// pkg/messaging/nats.go
package messaging

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/nats-io/nats.go"
)

// Message represents a message payload
type Message struct {
    ID        string                 `json:"id"`
    Type      string                 `json:"type"`
    Data      map[string]interface{} `json:"data"`
    Timestamp time.Time              `json:"timestamp"`
}

// MessageHandler handles incoming messages
type MessageHandler func(ctx context.Context, msg *Message) error

// MessageBroker defines message broker interface
type MessageBroker interface {
    Publish(ctx context.Context, subject string, msg *Message) error
    Subscribe(ctx context.Context, subject string, handler MessageHandler) error
    Close() error
}

// NATSMessageBroker implements MessageBroker using NATS
type NATSMessageBroker struct {
    conn *nats.Conn
    js   nats.JetStreamContext
}

func NewNATSMessageBroker(url string) (*NATSMessageBroker, error) {
    conn, err := nats.Connect(url,
        nats.ReconnectWait(2*time.Second),
        nats.MaxReconnects(5),
        nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
            fmt.Printf("NATS disconnected: %v\n", err)
        }),
        nats.ReconnectHandler(func(nc *nats.Conn) {
            fmt.Printf("NATS reconnected to %v\n", nc.ConnectedUrl())
        }),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to connect to NATS: %w", err)
    }
    
    js, err := conn.JetStream()
    if err != nil {
        return nil, fmt.Errorf("failed to get JetStream context: %w", err)
    }
    
    return &NATSMessageBroker{
        conn: conn,
        js:   js,
    }, nil
}

func (b *NATSMessageBroker) Publish(ctx context.Context, subject string, msg *Message) error {
    data, err := json.Marshal(msg)
    if err != nil {
        return fmt.Errorf("failed to marshal message: %w", err)
    }
    
    // Use JetStream for persistent messaging
    _, err = b.js.Publish(ctx, subject, data)
    if err != nil {
        return fmt.Errorf("failed to publish message: %w", err)
    }
    
    return nil
}

func (b *NATSMessageBroker) Subscribe(ctx context.Context, subject string, handler MessageHandler) error {
    // Create stream if it doesn't exist
    streamConfig := &nats.StreamConfig{
        Name:        "events",
        Subjects:    []string{subject},
        Retention:   nats.WorkQueuePolicy,
        MaxBytes:    1024 * 1024 * 1024, // 1GB
        Storage:     nats.FileStorage,
        Replicas:    1,
    }
    
    _, err := b.js.AddStream(streamConfig)
    if err != nil && !errors.Is(err, nats.ErrStreamNameAlreadyInUse) {
        return fmt.Errorf("failed to create stream: %w", err)
    }
    
    // Create consumer
    consumerConfig := &nats.ConsumerConfig{
        Durable:   "worker",
        AckPolicy: nats.AckExplicitPolicy,
    }
    
    _, err = b.js.AddConsumer("events", consumerConfig)
    if err != nil && !errors.Is(err, nats.ErrConsumerAlreadyExists) {
        return fmt.Errorf("failed to create consumer: %w", err)
    }
    
    // Subscribe to messages
    _, err = b.js.Subscribe(subject, func(msg *nats.Msg) {
        var message Message
        if err := json.Unmarshal(msg.Data, &message); err != nil {
            fmt.Printf("Failed to unmarshal message: %v\n", err)
            msg.Nak()
            return
        }
        
        if err := handler(ctx, &message); err != nil {
            fmt.Printf("Handler failed: %v\n", err)
            msg.Nak()
            return
        }
        
        msg.Ack()
    }, nats.Durable("worker"))
    
    if err != nil {
        return fmt.Errorf("failed to subscribe: %w", err)
    }
    
    return nil
}

func (b *NATSMessageBroker) Close() error {
    b.conn.Close()
    return nil
}

// Usage example
func setupEventHandlers(broker MessageBroker) {
    // User created event handler
    broker.Subscribe(context.Background(), "user.created", func(ctx context.Context, msg *Message) error {
        fmt.Printf("User created: %v\n", msg)
        // Send welcome email, update analytics, etc.
        return nil
    })
    
    // User updated event handler
    broker.Subscribe(context.Background(), "user.updated", func(ctx context.Context, msg *Message) error {
        fmt.Printf("User updated: %v\n", msg)
        // Update search index, notify other services, etc.
        return nil
    })
}
```

## Modern Development Workflow

### Configuration Management

```go
// pkg/config/config.go
package config

import (
    "fmt"
    "os"
    "strconv"
    "time"
    
    "github.com/spf13/viper"
)

// Config holds application configuration
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Redis    RedisConfig    `mapstructure:"redis"`
    Auth     AuthConfig     `mapstructure:"auth"`
    Logging  LoggingConfig  `mapstructure:"logging"`
    Monitoring MonitoringConfig `mapstructure:"monitoring"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
    Host         string        `mapstructure:"host"`
    Port         int           `mapstructure:"port"`
    ReadTimeout  time.Duration `mapstructure:"read_timeout"`
    WriteTimeout time.Duration `mapstructure:"write_timeout"`
    GracefulShutdownTimeout time.Duration `mapstructure:"graceful_shutdown_timeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
    Host            string        `mapstructure:"host"`
    Port            int           `mapstructure:"port"`
    User            string        `mapstructure:"user"`
    Password        string        `mapstructure:"password"`
    DBName          string        `mapstructure:"db_name"`
    SSLMode         string        `mapstructure:"ssl_mode"`
    MaxOpenConns    int           `mapstructure:"max_open_conns"`
    MaxIdleConns    int           `mapstructure:"max_idle_conns"`
    ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
    // Set default values
    setDefaults()
    
    // Load from config file
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("./configs")
    viper.AddConfigPath(".")
    
    // Load environment variables
    viper.AutomaticEnv()
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    
    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, fmt.Errorf("failed to read config file: %w", err)
        }
        // Config file not found is OK, we'll use defaults and env vars
    }
    
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    
    // Validate configuration
    if err := validate(&config); err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }
    
    return &config, nil
}

func setDefaults() {
    viper.SetDefault("server.host", "0.0.0.0")
    viper.SetDefault("server.port", 8080)
    viper.SetDefault("server.read_timeout", 30*time.Second)
    viper.SetDefault("server.write_timeout", 30*time.Second)
    viper.SetDefault("server.graceful_shutdown_timeout", 30*time.Second)
    
    viper.SetDefault("database.host", "localhost")
    viper.SetDefault("database.port", 5432)
    viper.SetDefault("database.ssl_mode", "disable")
    viper.SetDefault("database.max_open_conns", 25)
    viper.SetDefault("database.max_idle_conns", 25)
    viper.SetDefault("database.conn_max_lifetime", 5*time.Minute)
    
    viper.SetDefault("redis.host", "localhost")
    viper.SetDefault("redis.port", 6379)
    viper.SetDefault("redis.db", 0)
    
    viper.SetDefault("auth.jwt.expiration", 24*time.Hour)
    
    viper.SetDefault("logging.level", "info")
    viper.SetDefault("logging.format", "json")
}

func validate(config *Config) error {
    if config.Server.Port <= 0 || config.Server.Port > 65535 {
        return fmt.Errorf("invalid server port: %d", config.Server.Port)
    }
    
    if config.Database.Host == "" {
        return fmt.Errorf("database host is required")
    }
    
    if config.Database.User == "" {
        return fmt.Errorf("database user is required")
    }
    
    return nil
}

// GetDSN returns database connection string
func (c *DatabaseConfig) GetDSN() string {
    return fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
    )
}

// GetRedisAddr returns Redis connection address
func (c *RedisConfig) GetAddr() string {
    return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetServerAddr returns server address
func (c *ServerConfig) GetAddr() string {
    return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
```

### Makefile for Development

```makefile
# Makefile
.PHONY: help build run test clean lint format docker-build docker-run

# Variables
APP_NAME := your-app
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Docker variables
DOCKER_REGISTRY := your-registry
DOCKER_TAG := $(DOCKER_REGISTRY)/$(APP_NAME):$(VERSION)

# Help
help:
	@echo "Available commands:"
	@echo "  build      Build the application"
	@echo "  run        Run the application"
	@echo "  test       Run tests"
	@echo "  test-coverage Run tests with coverage"
	@echo "  lint       Run linter"
	@echo "  format     Format code"
	@echo "  clean      Clean build artifacts"
	@echo "  deps       Install dependencies"
	@echo "  migrate    Run database migrations"
	@echo "  docker-build Build Docker image"
	@echo "  docker-run  Run Docker container"
	@echo "  generate   Generate code (mocks, protobuf, etc.)"

# Build
build:
	go build $(LDFLAGS) -o bin/$(APP_NAME) cmd/server/main.go

# Run
run:
	go run $(LDFLAGS) cmd/server/main.go

# Development mode with hot reload
dev:
	air

# Tests
test:
	go test -v ./...

test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-bench:
	go test -bench=. -benchmem ./...

# Linting
lint:
	golangci-lint run

# Formatting
format:
	go fmt ./...
	goimports -w .

# Dependencies
deps:
	go mod download
	go mod tidy

# Generate code
generate:
	go generate ./...
	mockgen -source=internal/service/user.go -destination=test/mocks/user_service_mock.go
	protoc --go_out=. --go-grpc_out=. api/proto/user.proto

# Database
migrate-up:
	migrate -path migrations -database "$(shell grep -A5 'database:' configs/config.yaml | tail -n1 | cut -d' ' -f2)" up

migrate-down:
	migrate -path migrations -database "$(shell grep -A5 'database:' configs/config.yaml | tail -n1 | cut -d' ' -f2)" down

# Clean
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Docker
docker-build:
	docker build -t $(DOCKER_TAG) .

docker-run:
	docker run -p 8080:8080 $(DOCKER_TAG)

docker-push:
	docker push $(DOCKER_TAG)

# Install tools
install-tools:
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golang/mock/mockgen@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

---

**Created by**: MoAI Language Skill Factory  
**Last Updated**: 2025-11-06  
**Version**: 2.0.0  
**Go Target**: 1.25+ with latest language features  

This skill provides comprehensive Go development guidance with 2025 best practices, covering everything from basic concurrent programming to advanced cloud-native patterns and enterprise-grade applications.
