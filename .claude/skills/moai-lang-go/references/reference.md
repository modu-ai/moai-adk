# Go 1.23+ Complete Reference

Comprehensive reference for Go development with modern frameworks and patterns.

---

## Go 1.23 Language Features

### Range Over Integers

```go
// Iterate 0 to n-1
for i := range 10 {
    fmt.Println(i) // 0, 1, 2, ..., 9
}

// Traditional range still works
for i, v := range slice {
    fmt.Printf("%d: %v\n", i, v)
}
```

### Improved Generics

```go
// Generic Map function
func Map[T, U any](slice []T, fn func(T) U) []U {
    result := make([]U, len(slice))
    for i, v := range slice {
        result[i] = fn(v)
    }
    return result
}

// Type constraint with comparable
func Contains[T comparable](slice []T, item T) bool {
    for _, v := range slice {
        if v == item {
            return true
        }
    }
    return false
}

// Multiple type constraints
type Number interface {
    int | int32 | int64 | float32 | float64
}

func Sum[T Number](slice []T) T {
    var sum T
    for _, v := range slice {
        sum += v
    }
    return sum
}
```

### Error Handling Patterns

```go
// Sentinel errors
var (
    ErrNotFound = errors.New("not found")
    ErrInvalid  = errors.New("invalid")
)

// Custom error types
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Error wrapping
func fetchUser(id int64) (*User, error) {
    user, err := db.FindByID(id)
    if err != nil {
        return nil, fmt.Errorf("fetch user %d: %w", id, err)
    }
    return user, nil
}

// Error checking
if errors.Is(err, ErrNotFound) {
    // Handle not found
}

var validErr ValidationError
if errors.As(err, &validErr) {
    // Handle validation error
}
```

---

## Web Framework: Fiber v3

### Installation

```bash
go get github.com/gofiber/fiber/v3
go get github.com/gofiber/fiber/v3/middleware/cors
go get github.com/gofiber/fiber/v3/middleware/logger
go get github.com/gofiber/fiber/v3/middleware/recover
go get github.com/gofiber/fiber/v3/middleware/limiter
```

### Complete Application Structure

```go
package main

import (
    "time"
    "github.com/gofiber/fiber/v3"
    "github.com/gofiber/fiber/v3/middleware/cors"
    "github.com/gofiber/fiber/v3/middleware/logger"
    "github.com/gofiber/fiber/v3/middleware/recover"
    "github.com/gofiber/fiber/v3/middleware/limiter"
)

func main() {
    app := fiber.New(fiber.Config{
        ErrorHandler:  customErrorHandler,
        Prefork:       true,
        ReadTimeout:   10 * time.Second,
        WriteTimeout:  10 * time.Second,
        IdleTimeout:   120 * time.Second,
    })

    // Middleware stack
    app.Use(recover.New())
    app.Use(logger.New(logger.Config{
        Format: "[${time}] ${status} - ${method} ${path}\n",
    }))
    app.Use(cors.New(cors.Config{
        AllowOrigins: []string{"*"},
        AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
    }))
    app.Use(limiter.New(limiter.Config{
        Max:        100,
        Expiration: time.Minute,
    }))

    // Routes
    api := app.Group("/api/v1")
    api.Get("/users", listUsers)
    api.Get("/users/:id", getUser)
    api.Post("/users", createUser)
    api.Put("/users/:id", updateUser)
    api.Delete("/users/:id", deleteUser)

    app.Listen(":3000")
}

func customErrorHandler(c fiber.Ctx, err error) error {
    code := fiber.StatusInternalServerError
    message := "Internal Server Error"

    if e, ok := err.(*fiber.Error); ok {
        code = e.Code
        message = e.Message
    }

    return c.Status(code).JSON(fiber.Map{"error": message})
}
```

### Request Handling

```go
// Path parameters
func getUser(c fiber.Ctx) error {
    id, err := c.ParamsInt("id")
    if err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "Invalid ID")
    }
    return c.JSON(fiber.Map{"id": id})
}

// Query parameters
func listUsers(c fiber.Ctx) error {
    limit := c.QueryInt("limit", 10)
    offset := c.QueryInt("offset", 0)
    sort := c.Query("sort", "created_at")
    return c.JSON(fiber.Map{"limit": limit, "offset": offset})
}

// Request body
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

func createUser(c fiber.Ctx) error {
    var req CreateUserRequest
    if err := c.BodyParser(&req); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "Invalid body")
    }
    return c.Status(fiber.StatusCreated).JSON(req)
}

// Headers
func authenticated(c fiber.Ctx) error {
    token := c.Get("Authorization")
    if token == "" {
        return fiber.NewError(fiber.StatusUnauthorized, "Missing token")
    }
    return c.Next()
}
```

---

## Web Framework: Gin

### Installation

```bash
go get -u github.com/gin-gonic/gin
go get -u github.com/gin-contrib/cors
go get -u github.com/gin-contrib/zap
```

### Complete Application Structure

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
)

func main() {
    r := gin.Default()

    // CORS configuration
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
    }))

    // Routes
    api := r.Group("/api/v1")
    {
        users := api.Group("/users")
        {
            users.GET("", listUsers)
            users.GET("/:id", getUser)
            users.POST("", createUser)
            users.PUT("/:id", updateUser)
            users.DELETE("/:id", deleteUser)
        }
    }

    r.Run(":3000")
}
```

### Request Handling

```go
// Path parameters
func getUser(c *gin.Context) {
    id := c.Param("id")
    c.JSON(http.StatusOK, gin.H{"id": id})
}

// Query parameters
func listUsers(c *gin.Context) {
    limit := c.DefaultQuery("limit", "10")
    offset := c.DefaultQuery("offset", "0")
    c.JSON(http.StatusOK, gin.H{"limit": limit, "offset": offset})
}

// Request binding with validation
type CreateUserRequest struct {
    Name  string `json:"name" binding:"required,min=2,max=100"`
    Email string `json:"email" binding:"required,email"`
    Age   int    `json:"age" binding:"gte=0,lte=150"`
}

func createUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, req)
}

// Custom validation
type UpdateUserRequest struct {
    Name  string `json:"name" binding:"omitempty,min=2"`
    Email string `json:"email" binding:"omitempty,email"`
}
```

### Middleware

```go
// Authentication middleware
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Missing authorization token",
            })
            return
        }
        // Validate token
        c.Set("user_id", 123)
        c.Next()
    }
}

// Logging middleware
func LoggingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        duration := time.Since(start)
        log.Printf("%s %s %d %v", c.Request.Method, c.Request.URL.Path,
            c.Writer.Status(), duration)
    }
}

// Usage
api.Use(AuthMiddleware())
```

---

## ORM: GORM 1.25

### Installation

```bash
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres
go get -u gorm.io/driver/mysql
go get -u gorm.io/driver/sqlite
```

### Connection Setup

```go
import (
    "gorm.io/gorm"
    "gorm.io/driver/postgres"
)

func NewDB(dsn string) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        PrepareStmt: true,
        Logger:      logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        return nil, err
    }

    sqlDB, _ := db.DB()
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)

    return db, nil
}
```

### Model Definitions

```go
type User struct {
    gorm.Model
    Name      string    `gorm:"uniqueIndex;not null;size:255"`
    Email     string    `gorm:"uniqueIndex;not null;size:255"`
    Age       int       `gorm:"default:0;check:age >= 0"`
    Birthday  time.Time `gorm:"type:date"`
    Profile   Profile   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
    Posts     []Post    `gorm:"foreignKey:AuthorID"`
    Roles     []Role    `gorm:"many2many:user_roles"`
}

type Profile struct {
    gorm.Model
    UserID uint   `gorm:"uniqueIndex"`
    Bio    string `gorm:"type:text"`
    Avatar string
}

type Post struct {
    gorm.Model
    Title    string `gorm:"size:255;not null"`
    Content  string `gorm:"type:text"`
    AuthorID uint   `gorm:"index"`
    Tags     []Tag  `gorm:"many2many:post_tags"`
}

type Role struct {
    gorm.Model
    Name        string `gorm:"uniqueIndex;size:100"`
    Permissions []Permission `gorm:"many2many:role_permissions"`
}
```

### Query Patterns

```go
// Basic queries
var user User
db.First(&user, 1)
db.First(&user, "email = ?", "john@example.com")

var users []User
db.Find(&users)
db.Where("age > ?", 18).Find(&users)
db.Where("name LIKE ?", "%John%").Find(&users)

// Preloading associations
db.Preload("Posts").First(&user, 1)
db.Preload("Posts", func(db *gorm.DB) *gorm.DB {
    return db.Order("posts.created_at DESC").Limit(10)
}).Preload("Profile").First(&user, 1)

// Pagination
var users []User
db.Limit(10).Offset(20).Find(&users)

// Ordering
db.Order("created_at DESC").Find(&users)

// Select specific fields
db.Select("id", "name", "email").Find(&users)

// Transactions
db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&user).Error; err != nil {
        return err
    }
    if err := tx.Create(&profile).Error; err != nil {
        return err
    }
    return nil
})

// Batch operations
db.CreateInBatches(users, 100)

// Raw SQL
var result struct {
    Count int
}
db.Raw("SELECT COUNT(*) as count FROM users WHERE age > ?", 18).Scan(&result)
```

---

## PostgreSQL Driver: pgx

### Connection Pool

```go
import "github.com/jackc/pgx/v5/pgxpool"

func NewPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
    config, err := pgxpool.ParseConfig(connString)
    if err != nil {
        return nil, err
    }

    config.MaxConns = 25
    config.MinConns = 5
    config.MaxConnLifetime = time.Hour
    config.MaxConnIdleTime = 30 * time.Minute
    config.HealthCheckPeriod = time.Minute

    return pgxpool.NewWithConfig(ctx, config)
}
```

### Query Patterns

```go
// Single row
var user User
err := pool.QueryRow(ctx,
    "SELECT id, name, email FROM users WHERE id = $1", id).
    Scan(&user.ID, &user.Name, &user.Email)

// Multiple rows
rows, err := pool.Query(ctx,
    "SELECT id, name, email FROM users ORDER BY created_at DESC LIMIT $1", 10)
defer rows.Close()

var users []User
for rows.Next() {
    var u User
    if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
        return nil, err
    }
    users = append(users, u)
}

// Execute
result, err := pool.Exec(ctx,
    "UPDATE users SET name = $2 WHERE id = $1", id, name)
rowsAffected := result.RowsAffected()

// Transaction
tx, err := pool.Begin(ctx)
defer tx.Rollback(ctx)

_, err = tx.Exec(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)", name, email)
if err != nil {
    return err
}

err = tx.Commit(ctx)
```

---

## Concurrency Patterns

### Errgroup for Structured Concurrency

```go
import "golang.org/x/sync/errgroup"

func fetchAllData(ctx context.Context) (*AllData, error) {
    g, ctx := errgroup.WithContext(ctx)

    var users []User
    var orders []Order
    var products []Product

    g.Go(func() error {
        var err error
        users, err = fetchUsers(ctx)
        return err
    })

    g.Go(func() error {
        var err error
        orders, err = fetchOrders(ctx)
        return err
    })

    g.Go(func() error {
        var err error
        products, err = fetchProducts(ctx)
        return err
    })

    if err := g.Wait(); err != nil {
        return nil, err
    }

    return &AllData{Users: users, Orders: orders, Products: products}, nil
}
```

### Semaphore for Rate Limiting

```go
import "golang.org/x/sync/semaphore"

var sem = semaphore.NewWeighted(10)

func processWithLimit(ctx context.Context, items []Item) error {
    g, ctx := errgroup.WithContext(ctx)

    for _, item := range items {
        item := item
        g.Go(func() error {
            if err := sem.Acquire(ctx, 1); err != nil {
                return err
            }
            defer sem.Release(1)
            return processItem(ctx, item)
        })
    }

    return g.Wait()
}
```

### Worker Pool

```go
func workerPool(ctx context.Context, jobs <-chan Job, numWorkers int) <-chan Result {
    results := make(chan Result, 100)

    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            for job := range jobs {
                select {
                case <-ctx.Done():
                    return
                default:
                    results <- processJob(job)
                }
            }
        }(i)
    }

    go func() {
        wg.Wait()
        close(results)
    }()

    return results
}
```

---

## CLI: Cobra with Viper

### Complete CLI Structure

```go
import (
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
    Use:   "myctl",
    Short: "My CLI tool",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        return initConfig()
    },
}

func init() {
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
    rootCmd.PersistentFlags().String("database-url", "", "database connection")

    viper.BindPFlag("database.url", rootCmd.PersistentFlags().Lookup("database-url"))
    viper.SetEnvPrefix("MYCTL")
    viper.AutomaticEnv()
}

func initConfig() error {
    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        home, _ := os.UserHomeDir()
        viper.AddConfigPath(home)
        viper.SetConfigName(".myctl")
    }
    return viper.ReadInConfig()
}
```

---

## Context7 Library Mappings

### Core Language and Tools

- `/golang/go` - Go language and stdlib
- `/golang/tools` - Go tools (gopls, goimports)

### Web Frameworks

- `/gofiber/fiber` - Fiber v3 web framework
- `/gin-gonic/gin` - Gin web framework
- `/labstack/echo` - Echo 4.13 web framework
- `/go-chi/chi` - Chi router

### Database

- `/go-gorm/gorm` - GORM ORM
- `/sqlc-dev/sqlc` - Type-safe SQL generator
- `/jackc/pgx` - PostgreSQL driver
- `/jmoiron/sqlx` - SQL extensions

### Testing

- `/stretchr/testify` - Testing toolkit
- `/golang/mock` - Mocking framework

### CLI

- `/spf13/cobra` - CLI framework
- `/spf13/viper` - Configuration

### Concurrency

- `/golang/sync` - Sync primitives (errgroup, semaphore)

---

## Performance Characteristics

### Startup Time

- Fast: 10-50ms typical startup

### Memory Usage

- Low: 10-50MB base memory footprint

### Throughput

- High: 50k-100k requests/second typical

### Latency

- Low: p99 less than 10ms for most APIs

### Container Image Size

- 10-20MB with scratch base image

---

Last Updated: 2025-12-07
Version: 1.0.0
