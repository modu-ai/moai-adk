# Systems Programming Reference

Complete reference for Go 1.23 and Rust 1.91 systems programming.

---

## Go 1.23 Complete Reference

### Language Features

Range Over Integers (New):
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

Improved Generics:
```go
// Generic constraints with type inference
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
```

### Web Framework: Fiber v3

Installation:
```bash
go get github.com/gofiber/fiber/v3
```

Complete API Setup:
```go
package main

import (
    "github.com/gofiber/fiber/v3"
    "github.com/gofiber/fiber/v3/middleware/cors"
    "github.com/gofiber/fiber/v3/middleware/logger"
    "github.com/gofiber/fiber/v3/middleware/recover"
)

func main() {
    app := fiber.New(fiber.Config{
        ErrorHandler: customErrorHandler,
        Prefork:      true, // Enable prefork for performance
    })

    // Middleware
    app.Use(recover.New())
    app.Use(logger.New())
    app.Use(cors.New(cors.Config{
        AllowOrigins: []string{"*"},
        AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
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
    if e, ok := err.(*fiber.Error); ok {
        code = e.Code
    }
    return c.Status(code).JSON(fiber.Map{"error": err.Error()})
}
```

### Web Framework: Echo 4.13

Installation:
```bash
go get github.com/labstack/echo/v4
```

Complete API Setup:
```go
package main

import (
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
)

func main() {
    e := echo.New()

    // Middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Use(middleware.CORS())
    e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))

    // Routes
    api := e.Group("/api/v1")
    api.GET("/users", listUsers)
    api.GET("/users/:id", getUser)
    api.POST("/users", createUser)

    // Custom validator
    e.Validator = &CustomValidator{validator: validator.New()}

    e.Logger.Fatal(e.Start(":3000"))
}
```

### Web Framework: Chi

Installation:
```bash
go get github.com/go-chi/chi/v5
```

Complete API Setup:
```go
package main

import (
    "net/http"
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

func main() {
    r := chi.NewRouter()

    // Middleware
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.RealIP)
    r.Use(middleware.Timeout(60 * time.Second))

    // Routes
    r.Route("/api/v1", func(r chi.Router) {
        r.Route("/users", func(r chi.Router) {
            r.Get("/", listUsers)
            r.Post("/", createUser)
            r.Route("/{id}", func(r chi.Router) {
                r.Use(userCtx)
                r.Get("/", getUser)
                r.Put("/", updateUser)
                r.Delete("/", deleteUser)
            })
        })
    })

    http.ListenAndServe(":3000", r)
}
```

### ORM: GORM 1.25

Installation:
```bash
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres
```

Model Definition:
```go
type User struct {
    gorm.Model
    Name      string    `gorm:"uniqueIndex;not null"`
    Email     string    `gorm:"uniqueIndex;not null"`
    Age       int       `gorm:"default:0"`
    Birthday  time.Time `gorm:"type:date"`
    Posts     []Post    `gorm:"foreignKey:AuthorID"`
    Profile   Profile   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

type Post struct {
    gorm.Model
    Title    string `gorm:"size:255;not null"`
    Content  string `gorm:"type:text"`
    AuthorID uint
    Tags     []Tag  `gorm:"many2many:post_tags"`
}
```

Query Patterns:
```go
// Preloading associations
var user User
db.Preload("Posts", func(db *gorm.DB) *gorm.DB {
    return db.Order("posts.created_at DESC").Limit(10)
}).Preload("Profile").First(&user, 1)

// Transaction
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
```

### Type-Safe SQL: sqlc 1.26

Configuration (sqlc.yaml):
```yaml
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
        emit_json_tags: true
        emit_prepared_queries: true
```

Schema and Queries:
```sql
-- schema.sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- query.sql
-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: CreateUser :one
INSERT INTO users (name, email) VALUES ($1, $2) RETURNING *;

-- name: UpdateUser :one
UPDATE users SET name = $2, email = $3 WHERE id = $1 RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;
```

### PostgreSQL Driver: pgx

Connection Pool:
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

    return pgxpool.NewWithConfig(ctx, config)
}
```

### Concurrency Patterns

Errgroup for Structured Concurrency:
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

Semaphore for Rate Limiting:
```go
import "golang.org/x/sync/semaphore"

var sem = semaphore.NewWeighted(10) // Max 10 concurrent operations

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

### Testing with testify

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
)

// Mock
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) GetByID(id int64) (*User, error) {
    args := m.Called(id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

// Table-driven tests
func TestUserService(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateUserInput
        want    *User
        wantErr bool
    }{
        {"valid user", CreateUserInput{Name: "John"}, &User{Name: "John"}, false},
        {"empty name", CreateUserInput{Name: ""}, nil, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := svc.Create(tt.input)
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.want.Name, got.Name)
        })
    }
}
```

### HTTP Test Patterns

```go
import (
    "net/http/httptest"
    "github.com/gofiber/fiber/v3"
)

func TestGetUser(t *testing.T) {
    app := fiber.New()
    app.Get("/users/:id", getUser)

    req := httptest.NewRequest("GET", "/users/1", nil)
    req.Header.Set("Content-Type", "application/json")

    resp, err := app.Test(req)
    require.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)

    var user User
    json.NewDecoder(resp.Body).Decode(&user)
    assert.Equal(t, int64(1), user.ID)
}
```

### CLI: Cobra with Viper

```go
import (
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: "Application description",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        return initConfig()
    },
}

func init() {
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
    rootCmd.PersistentFlags().String("database-url", "", "database connection string")

    viper.BindPFlag("database.url", rootCmd.PersistentFlags().Lookup("database-url"))
    viper.SetEnvPrefix("MYAPP")
    viper.AutomaticEnv()
}

func initConfig() error {
    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        viper.AddConfigPath(".")
        viper.SetConfigName("config")
    }
    return viper.ReadInConfig()
}
```

---

## Rust 1.91 Complete Reference

### Language Features

Async Traits (Stable):
```rust
// No more async-trait crate needed
trait AsyncRepository {
    async fn get(&self, id: i64) -> Result<User, Error>;
    async fn create(&self, user: CreateUser) -> Result<User, Error>;
}

impl AsyncRepository for PostgresRepository {
    async fn get(&self, id: i64) -> Result<User, Error> {
        sqlx::query_as!(User, "SELECT * FROM users WHERE id = $1", id)
            .fetch_one(&self.pool)
            .await
    }

    async fn create(&self, user: CreateUser) -> Result<User, Error> {
        sqlx::query_as!(User,
            "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING *",
            user.name, user.email)
            .fetch_one(&self.pool)
            .await
    }
}
```

Const Generics:
```rust
// Compile-time sized arrays
fn process_batch<const N: usize>(items: [Item; N]) -> [Result<Output, Error>; N] {
    items.map(|item| process_item(item))
}

// Generic buffer with compile-time size
struct Buffer<T, const SIZE: usize> {
    data: [T; SIZE],
    len: usize,
}
```

Let-Else Pattern:
```rust
fn get_user(id: Option<i64>) -> Result<User, Error> {
    let Some(id) = id else {
        return Err(Error::MissingId);
    };

    let Ok(user) = repository.find(id) else {
        return Err(Error::NotFound);
    };

    Ok(user)
}
```

### Web Framework: Axum 0.8

Installation:
```toml
[dependencies]
axum = "0.8"
tokio = { version = "1.48", features = ["full"] }
tower = "0.5"
tower-http = { version = "0.6", features = ["cors", "trace"] }
```

Complete API Setup:
```rust
use axum::{
    extract::{Path, State, Query},
    http::StatusCode,
    response::{IntoResponse, Json},
    routing::{get, post, put, delete},
    Router,
};
use tower_http::cors::{CorsLayer, Any};
use tower_http::trace::TraceLayer;

#[derive(Clone)]
struct AppState {
    db: PgPool,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    tracing_subscriber::init();

    let pool = PgPoolOptions::new()
        .max_connections(25)
        .connect(&std::env::var("DATABASE_URL")?).await?;

    let state = AppState { db: pool };

    let app = Router::new()
        .route("/api/v1/users", get(list_users).post(create_user))
        .route("/api/v1/users/:id", get(get_user).put(update_user).delete(delete_user))
        .layer(TraceLayer::new_for_http())
        .layer(CorsLayer::new().allow_origin(Any).allow_methods(Any))
        .with_state(state);

    let listener = tokio::net::TcpListener::bind("0.0.0.0:3000").await?;
    axum::serve(listener, app).await?;
    Ok(())
}

async fn list_users(
    State(state): State<AppState>,
    Query(params): Query<ListParams>,
) -> Result<Json<Vec<User>>, AppError> {
    let users = sqlx::query_as!(User,
        "SELECT * FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2",
        params.limit.unwrap_or(10),
        params.offset.unwrap_or(0))
        .fetch_all(&state.db)
        .await?;
    Ok(Json(users))
}
```

### Web Framework: Actix-web 4

Installation:
```toml
[dependencies]
actix-web = "4"
actix-rt = "2"
```

Complete API Setup:
```rust
use actix_web::{web, App, HttpServer, HttpResponse, middleware};

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let pool = PgPoolOptions::new()
        .max_connections(25)
        .connect(&std::env::var("DATABASE_URL").unwrap()).await.unwrap();

    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(pool.clone()))
            .wrap(middleware::Logger::default())
            .service(
                web::scope("/api/v1")
                    .route("/users", web::get().to(list_users))
                    .route("/users", web::post().to(create_user))
                    .route("/users/{id}", web::get().to(get_user))
            )
    })
    .bind("0.0.0.0:3000")?
    .run()
    .await
}
```

### Async Runtime: Tokio 1.48

Task Spawning and Channels:
```rust
use tokio::sync::{mpsc, oneshot};

async fn worker_pool() {
    let (tx, mut rx) = mpsc::channel::<Job>(100);

    // Spawn workers
    for _ in 0..4 {
        let mut rx = rx.clone();
        tokio::spawn(async move {
            while let Some(job) = rx.recv().await {
                process_job(job).await;
            }
        });
    }

    // Send jobs
    for job in jobs {
        tx.send(job).await.unwrap();
    }
}

// Select for multiple futures
async fn timeout_operation() -> Result<Data, Error> {
    tokio::select! {
        result = fetch_data() => result,
        _ = tokio::time::sleep(Duration::from_secs(5)) => {
            Err(Error::Timeout)
        }
    }
}
```

### Serialization: Serde 1.0

```rust
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct User {
    id: i64,
    #[serde(rename = "userName")]
    name: String,
    email: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    profile_url: Option<String>,
    #[serde(default)]
    is_active: bool,
}

// Custom serialization
#[derive(Serialize, Deserialize)]
struct ApiResponse<T> {
    #[serde(flatten)]
    data: T,
    #[serde(with = "chrono::serde::ts_seconds")]
    timestamp: DateTime<Utc>,
}
```

### Database: SQLx 0.8

Type-Safe Queries:
```rust
use sqlx::{PgPool, FromRow, postgres::PgPoolOptions};

#[derive(Debug, FromRow)]
struct User {
    id: i64,
    name: String,
    email: String,
    created_at: chrono::DateTime<chrono::Utc>,
}

async fn user_operations(pool: &PgPool) -> Result<(), sqlx::Error> {
    // Compile-time checked query
    let user = sqlx::query_as!(User,
        "SELECT id, name, email, created_at FROM users WHERE id = $1",
        1i64)
        .fetch_one(pool)
        .await?;

    // Transaction
    let mut tx = pool.begin().await?;

    sqlx::query!("INSERT INTO users (name, email) VALUES ($1, $2)", "John", "john@example.com")
        .execute(&mut *tx)
        .await?;

    sqlx::query!("INSERT INTO profiles (user_id, bio) VALUES ($1, $2)", 1, "Developer")
        .execute(&mut *tx)
        .await?;

    tx.commit().await?;
    Ok(())
}
```

### Database: Diesel 2.0

Schema and Models:
```rust
// schema.rs (generated by diesel)
diesel::table! {
    users (id) {
        id -> Int8,
        name -> Varchar,
        email -> Varchar,
        created_at -> Timestamp,
    }
}

// models.rs
#[derive(Queryable, Selectable)]
#[diesel(table_name = users)]
struct User {
    id: i64,
    name: String,
    email: String,
    created_at: chrono::NaiveDateTime,
}

#[derive(Insertable)]
#[diesel(table_name = users)]
struct NewUser<'a> {
    name: &'a str,
    email: &'a str,
}
```

### Error Handling

thiserror for Library Errors:
```rust
use thiserror::Error;

#[derive(Error, Debug)]
pub enum AppError {
    #[error("database error: {0}")]
    Database(#[from] sqlx::Error),

    #[error("validation error: {field} - {message}")]
    Validation { field: String, message: String },

    #[error("not found: {0}")]
    NotFound(String),

    #[error("unauthorized")]
    Unauthorized,
}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let (status, message) = match &self {
            AppError::NotFound(_) => (StatusCode::NOT_FOUND, self.to_string()),
            AppError::Validation { .. } => (StatusCode::BAD_REQUEST, self.to_string()),
            AppError::Unauthorized => (StatusCode::UNAUTHORIZED, self.to_string()),
            AppError::Database(_) => (StatusCode::INTERNAL_SERVER_ERROR, "Internal error".into()),
        };
        (status, Json(json!({"error": message}))).into_response()
    }
}
```

anyhow for Application Errors:
```rust
use anyhow::{Context, Result, bail};

async fn process_request(id: i64) -> Result<Response> {
    let user = repository.find(id)
        .await
        .context("Failed to fetch user")?;

    if !user.is_active {
        bail!("User {} is not active", id);
    }

    Ok(Response::new(user))
}
```

### CLI: clap

```rust
use clap::{Parser, Subcommand, Args};

#[derive(Parser)]
#[command(name = "myapp", version, about)]
struct Cli {
    #[arg(short, long, global = true)]
    config: Option<PathBuf>,

    #[arg(short, long, global = true, default_value = "info")]
    log_level: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    Serve(ServeArgs),
    Migrate(MigrateArgs),
}

#[derive(Args)]
struct ServeArgs {
    #[arg(short, long, default_value = "3000")]
    port: u16,

    #[arg(long)]
    workers: Option<usize>,
}

fn main() {
    let cli = Cli::parse();
    match cli.command {
        Commands::Serve(args) => serve(args),
        Commands::Migrate(args) => migrate(args),
    }
}
```

### Testing

```rust
#[cfg(test)]
mod tests {
    use super::*;
    use axum::http::StatusCode;
    use axum_test::TestServer;

    async fn setup_test_server() -> TestServer {
        let pool = PgPoolOptions::new()
            .connect("postgres://test:test@localhost/test")
            .await
            .unwrap();

        let app = create_app(pool);
        TestServer::new(app).unwrap()
    }

    #[tokio::test]
    async fn test_create_user() {
        let server = setup_test_server().await;

        let response = server
            .post("/api/v1/users")
            .json(&json!({"name": "John", "email": "john@example.com"}))
            .await;

        assert_eq!(response.status_code(), StatusCode::CREATED);

        let user: User = response.json();
        assert_eq!(user.name, "John");
    }

    #[tokio::test]
    async fn test_get_user_not_found() {
        let server = setup_test_server().await;

        let response = server.get("/api/v1/users/99999").await;

        assert_eq!(response.status_code(), StatusCode::NOT_FOUND);
    }
}
```

---

## Context7 Library Mappings

### Go Libraries

Core Language and Tools:
- `/golang/go` - Go language, stdlib, toolchain
- `/golang/tools` - Go tools (gopls, goimports)

Web Frameworks:
- `/gofiber/fiber` - Fiber v3 web framework
- `/labstack/echo` - Echo 4.13 web framework
- `/go-chi/chi` - Chi router
- `/gin-gonic/gin` - Gin web framework

Database:
- `/go-gorm/gorm` - GORM ORM
- `/sqlc-dev/sqlc` - Type-safe SQL generator
- `/jackc/pgx` - PostgreSQL driver
- `/jmoiron/sqlx` - SQL extensions

Testing:
- `/stretchr/testify` - Testing toolkit
- `/golang/mock` - Mocking framework

CLI:
- `/spf13/cobra` - CLI framework
- `/spf13/viper` - Configuration

Concurrency:
- `/golang/sync` - Sync primitives (errgroup, semaphore)

### Rust Libraries

Core Language:
- `/rust-lang/rust` - Rust language and stdlib
- `/rust-lang/cargo` - Package manager

Async Runtime:
- `/tokio-rs/tokio` - Tokio async runtime
- `/async-rs/async-std` - async-std runtime

Web Frameworks:
- `/tokio-rs/axum` - Axum web framework
- `/actix/actix-web` - Actix-web framework

Serialization:
- `/serde-rs/serde` - Serialization framework
- `/serde-rs/json` - JSON serialization

Database:
- `/launchbadge/sqlx` - SQLx async SQL
- `/diesel-rs/diesel` - Diesel ORM

Error Handling:
- `/dtolnay/thiserror` - Error derive
- `/dtolnay/anyhow` - Error handling

CLI:
- `/clap-rs/clap` - CLI parser

---

## Performance Characteristics

### Startup Time
- Go: Fast (10-50ms)
- Rust: Medium (50-100ms)

### Memory Usage
- Go: Low (10-50MB base)
- Rust: Very Low (5-20MB base)

### Throughput
- Go: High (50k-100k req/s)
- Rust: Very High (100k-200k req/s)

### Latency
- Go: Low (p99 < 10ms)
- Rust: Very Low (p99 < 5ms)

### Container Image Size
- Go: 10-20MB (scratch base)
- Rust: 5-15MB (alpine base)

---

Last Updated: 2025-12-07
Version: 1.0.0
