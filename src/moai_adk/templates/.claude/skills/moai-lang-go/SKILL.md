---
name: moai-lang-go
description: Go 1.23+ development specialist covering Fiber, Gin, GORM, and concurrent programming patterns. Use when building high-performance microservices, CLI tools, or cloud-native applications.
version: 1.0.0
category: language
tags: [go, golang, fiber, gin, concurrency, microservices]
context7-libraries: [/gofiber/fiber, /gin-gonic/gin, /go-gorm/gorm]
related-skills: [moai-lang-rust, moai-domain-backend]
updated: 2025-12-07
status: active
---

## Quick Reference (30 seconds)

Go 1.23+ Development Expert for high-performance backend systems and CLI applications.

Auto-Triggers: `.go`, `go.mod`, `go.sum`, goroutines, channels, Fiber, Gin, GORM, Echo, Chi

Core Use Cases:
- High-performance REST APIs and microservices
- Concurrent and parallel processing systems
- CLI tools and system utilities
- Cloud-native containerized services

Quick Patterns:

Fiber API:
```go
app := fiber.New()
app.Get("/api/users/:id", func(c fiber.Ctx) error {
    return c.JSON(fiber.Map{"id": c.Params("id")})
})
app.Listen(":3000")
```

Gin API:
```go
r := gin.Default()
r.GET("/api/users/:id", func(c *gin.Context) {
    c.JSON(200, gin.H{"id": c.Param("id")})
})
r.Run(":3000")
```

Goroutine with Error Handling:
```go
g, ctx := errgroup.WithContext(context.Background())
g.Go(func() error { return processUsers(ctx) })
g.Go(func() error { return processOrders(ctx) })
if err := g.Wait(); err != nil { log.Fatal(err) }
```

---

## Implementation Guide (5 minutes)

### Go 1.23 Language Features

New Features:
- Range over integers: `for i := range 10 { fmt.Println(i) }`
- Profile-Guided Optimization (PGO) 2.0
- Improved generics with better type inference

Generics:
```go
func Map[T, U any](slice []T, fn func(T) U) []U {
    result := make([]U, len(slice))
    for i, v := range slice { result[i] = fn(v) }
    return result
}
```

### Web Framework: Fiber v3

```go
app := fiber.New(fiber.Config{ErrorHandler: customErrorHandler, Prefork: true})
app.Use(recover.New())
app.Use(logger.New())
app.Use(cors.New())

api := app.Group("/api/v1")
api.Get("/users", listUsers)
api.Get("/users/:id", getUser)
api.Post("/users", createUser)
api.Put("/users/:id", updateUser)
api.Delete("/users/:id", deleteUser)
app.Listen(":3000")
```

### Web Framework: Gin

```go
r := gin.Default()
r.Use(cors.Default())

api := r.Group("/api/v1")
api.GET("/users", listUsers)
api.GET("/users/:id", getUser)
api.POST("/users", createUser)
r.Run(":3000")
```

Request Binding:
```go
type CreateUserRequest struct {
    Name  string `json:"name" binding:"required,min=2"`
    Email string `json:"email" binding:"required,email"`
}

func createUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    c.JSON(201, gin.H{"id": 1, "name": req.Name})
}
```

### Web Framework: Echo

```go
e := echo.New()
e.Use(middleware.Logger())
e.Use(middleware.Recover())
e.Use(middleware.CORS())

api := e.Group("/api/v1")
api.GET("/users", listUsers)
api.POST("/users", createUser)
e.Logger.Fatal(e.Start(":3000"))
```

### Web Framework: Chi

```go
r := chi.NewRouter()
r.Use(middleware.Logger)
r.Use(middleware.Recoverer)

r.Route("/api/v1", func(r chi.Router) {
    r.Route("/users", func(r chi.Router) {
        r.Get("/", listUsers)
        r.Post("/", createUser)
        r.Get("/{id}", getUser)
    })
})
http.ListenAndServe(":3000", r)
```

### ORM: GORM 1.25

Model Definition:
```go
type User struct {
    gorm.Model
    Name    string  `gorm:"uniqueIndex;not null"`
    Email   string  `gorm:"uniqueIndex;not null"`
    Posts   []Post  `gorm:"foreignKey:AuthorID"`
}
```

Query Patterns:
```go
db.Preload("Posts", func(db *gorm.DB) *gorm.DB {
    return db.Order("created_at DESC").Limit(10)
}).First(&user, 1)

db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&user).Error; err != nil { return err }
    return tx.Create(&profile).Error
})
```

### Type-Safe SQL: sqlc

```yaml
# sqlc.yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "query.sql"
    schema: "schema.sql"
    gen:
      go:
        package: "db"
        out: "internal/db"
        sql_package: "pgx/v5"
```

```sql
-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (name, email) VALUES ($1, $2) RETURNING *;
```

### Concurrency Patterns

Errgroup:
```go
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { users, err = fetchUsers(ctx); return err })
g.Go(func() error { orders, err = fetchOrders(ctx); return err })
if err := g.Wait(); err != nil { return nil, err }
```

Worker Pool:
```go
func workerPool(jobs <-chan Job, results chan<- Result, n int) {
    var wg sync.WaitGroup
    for i := 0; i < n; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobs { results <- processJob(job) }
        }()
    }
    wg.Wait()
    close(results)
}
```

Context with Timeout:
```go
ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
defer cancel()
result, err := fetchData(ctx)
if errors.Is(err, context.DeadlineExceeded) {
    http.Error(w, "timeout", http.StatusGatewayTimeout)
}
```

### Testing Patterns

Table-Driven Tests:
```go
tests := []struct {
    name    string
    input   CreateUserInput
    wantErr bool
}{
    {"valid", CreateUserInput{Name: "John"}, false},
    {"empty", CreateUserInput{Name: ""}, true},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        _, err := svc.Create(tt.input)
        if tt.wantErr { require.Error(t, err) }
    })
}
```

HTTP Testing:
```go
app := fiber.New()
app.Get("/users/:id", getUser)
req := httptest.NewRequest("GET", "/users/1", nil)
resp, _ := app.Test(req)
assert.Equal(t, 200, resp.StatusCode)
```

### CLI: Cobra with Viper

```go
var rootCmd = &cobra.Command{Use: "myapp", Short: "Description"}

func init() {
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
    viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
    viper.SetEnvPrefix("MYAPP")
    viper.AutomaticEnv()
}
```

---

## Advanced Patterns

### Performance Optimization

PGO Build:
```bash
GODEBUG=pgo=1 ./myapp -cpuprofile=default.pgo
go build -pgo=default.pgo -o myapp
```

Object Pooling:
```go
var bufferPool = sync.Pool{
    New: func() interface{} { return make([]byte, 4096) },
}
buf := bufferPool.Get().([]byte)
defer bufferPool.Put(buf)
```

### Container Deployment (10-20MB)

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o main .

FROM scratch
COPY --from=builder /app/main /main
ENTRYPOINT ["/main"]
```

### Graceful Shutdown

```go
go func() { app.Listen(":3000") }()
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit
app.Shutdown()
```

---

## Context7 Libraries

```
/golang/go        - Go language and stdlib
/gofiber/fiber    - Fiber web framework
/gin-gonic/gin    - Gin web framework
/labstack/echo    - Echo web framework
/go-chi/chi       - Chi router
/go-gorm/gorm     - GORM ORM
/sqlc-dev/sqlc    - Type-safe SQL
/jackc/pgx        - PostgreSQL driver
/spf13/cobra      - CLI framework
/spf13/viper      - Configuration
/stretchr/testify - Testing toolkit
```

---

## Works Well With

- `moai-domain-backend` - REST API architecture and microservices
- `moai-lang-rust` - Systems programming companion
- `moai-quality-security` - Security hardening
- `moai-essentials-debug` - Performance profiling
- `moai-workflow-tdd` - Test-driven development

---

## Troubleshooting

Common Issues:
- Module errors: `go mod tidy && go mod verify`
- Version check: `go version && go env GOVERSION`
- Build issues: `go clean -cache && go build -v`

Performance Diagnostics:
- CPU profiling: `go test -cpuprofile=cpu.prof -bench=.`
- Memory profiling: `go test -memprofile=mem.prof -bench=.`
- Race detection: `go test -race ./...`

---

## Additional Resources

See [reference.md](reference.md) for complete framework reference, advanced patterns, and Context7 library mappings.

See [examples.md](examples.md) for production-ready code including REST APIs, CLI tools, and deployment configurations.

---

Last Updated: 2025-12-07
Version: 1.0.0
