---
name: moai-lang-systems
description: Systems programming specialist for Go 1.23 and Rust 1.91. Use when building high-performance microservices, CLI tools, concurrent systems, or memory-safe applications requiring low-latency execution.
version: 1.0.0
category: language
tags:
  - systems-programming
  - go
  - rust
  - concurrency
  - performance
updated: 2025-12-07
status: active
---

## Quick Reference (30 seconds)

Systems Programming Expert for Go 1.23 and Rust 1.91 with deep patterns for high-performance applications.

Auto-Triggers: `.go`, `.rs`, `go.mod`, `Cargo.toml`, goroutines, channels, async/await, Tokio, Axum, Fiber

Core Use Cases:
- High-performance microservices and REST APIs
- CLI tools and system utilities
- Concurrent and parallel processing systems
- Memory-safe low-latency applications
- Cloud-native containerized services

Quick Patterns:

Go Fiber API:
```go
app := fiber.New()
app.Get("/api/users/:id", func(c fiber.Ctx) error {
    id := c.Params("id")
    return c.JSON(fiber.Map{"id": id, "status": "active"})
})
app.Listen(":3000")
```

Rust Axum API:
```rust
let app = Router::new()
    .route("/api/users/:id", get(get_user))
    .with_state(app_state);
let listener = tokio::net::TcpListener::bind("0.0.0.0:3000").await?;
axum::serve(listener, app).await?;
```

---

## Implementation Guide (5 minutes)

### Go 1.23 Features and Patterns

New Language Features:
- Range over integers: `for i := range 10 { fmt.Println(i) }`
- Profile-Guided Optimization (PGO) 2.0 for production builds
- Improved generics with better type inference
- Enhanced toolchain with go telemetry and versioning

Concurrency Model:
```go
// Goroutines with errgroup for structured concurrency
g, ctx := errgroup.WithContext(context.Background())

g.Go(func() error {
    return processUsers(ctx)
})
g.Go(func() error {
    return processOrders(ctx)
})

if err := g.Wait(); err != nil {
    log.Fatal(err)
}
```

Channel Patterns:
```go
// Worker pool pattern
func workerPool(jobs <-chan Job, results chan<- Result, numWorkers int) {
    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobs {
                results <- processJob(job)
            }
        }()
    }
    wg.Wait()
    close(results)
}
```

Context Best Practices:
```go
func handler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    result, err := fetchData(ctx)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            http.Error(w, "timeout", http.StatusGatewayTimeout)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(result)
}
```

### Rust 1.91 Features and Patterns

Modern Rust Features:
- Async traits in stable (no more async-trait crate needed)
- Const generics for compile-time array sizing
- let-else for pattern matching with early return
- Improved borrow checker with polonius

Async Pattern with Tokio:
```rust
#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let pool = PgPoolOptions::new()
        .max_connections(5)
        .connect("postgres://localhost/db").await?;

    let app = Router::new()
        .route("/users", get(list_users).post(create_user))
        .route("/users/:id", get(get_user).delete(delete_user))
        .with_state(pool);

    let listener = tokio::net::TcpListener::bind("0.0.0.0:3000").await?;
    axum::serve(listener, app).await?;
    Ok(())
}
```

Error Handling with thiserror and anyhow:
```rust
use thiserror::Error;

#[derive(Error, Debug)]
pub enum AppError {
    #[error("database error: {0}")]
    Database(#[from] sqlx::Error),
    #[error("validation error: {0}")]
    Validation(String),
    #[error("not found: {resource} with id {id}")]
    NotFound { resource: &'static str, id: i64 },
}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let (status, message) = match &self {
            AppError::NotFound { .. } => (StatusCode::NOT_FOUND, self.to_string()),
            AppError::Validation(_) => (StatusCode::BAD_REQUEST, self.to_string()),
            AppError::Database(_) => (StatusCode::INTERNAL_SERVER_ERROR, "Internal error".into()),
        };
        (status, Json(json!({"error": message}))).into_response()
    }
}
```

### Web Framework Comparison

Go Frameworks:
- Fiber v3: Fastest, Express-like syntax, zero allocation router
- Echo 4.13: Feature-rich, middleware ecosystem, OpenAPI support
- Chi: Lightweight, composable, stdlib compatible

Rust Frameworks:
- Axum 0.8: Tokio-native, type-safe extractors, tower middleware
- Actix-web 4: Actor model, high throughput, mature ecosystem

### Database Integration

Go with GORM 1.25:
```go
type User struct {
    gorm.Model
    Name  string `gorm:"uniqueIndex"`
    Email string `gorm:"uniqueIndex"`
    Posts []Post `gorm:"foreignKey:AuthorID"`
}

// Auto-migrate and query
db.AutoMigrate(&User{})
var user User
db.Preload("Posts").First(&user, "name = ?", "john")
```

Go with sqlc 1.26 (Type-safe SQL):
```sql
-- query.sql
-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (name, email) VALUES ($1, $2) RETURNING *;
```

Rust with SQLx 0.8:
```rust
#[derive(Debug, sqlx::FromRow)]
struct User {
    id: i64,
    name: String,
    email: String,
}

async fn get_user(pool: &PgPool, id: i64) -> Result<User, sqlx::Error> {
    sqlx::query_as!(User, "SELECT id, name, email FROM users WHERE id = $1", id)
        .fetch_one(pool)
        .await
}
```

### CLI Development

Go with Cobra:
```go
var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: "A CLI application",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Hello from CLI")
    },
}

func init() {
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
    viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
}
```

Rust with clap:
```rust
use clap::Parser;

#[derive(Parser)]
#[command(name = "myapp", about = "A CLI application")]
struct Cli {
    #[arg(short, long)]
    config: Option<PathBuf>,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    Run { name: String },
    Build { #[arg(short, long)] release: bool },
}
```

### Testing Patterns

Go Testing:
```go
func TestUserService_Create(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        svc := NewUserService(mockDB)
        user, err := svc.Create(context.Background(), CreateUserInput{
            Name: "John", Email: "john@example.com",
        })
        require.NoError(t, err)
        assert.Equal(t, "John", user.Name)
    })

    t.Run("duplicate email", func(t *testing.T) {
        svc := NewUserService(mockDB)
        _, err := svc.Create(context.Background(), CreateUserInput{
            Name: "John", Email: "existing@example.com",
        })
        assert.ErrorIs(t, err, ErrDuplicateEmail)
    })
}
```

Rust Testing:
```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_create_user() {
        let pool = setup_test_db().await;
        let result = create_user(&pool, "John", "john@example.com").await;
        assert!(result.is_ok());
        assert_eq!(result.unwrap().name, "John");
    }

    #[tokio::test]
    async fn test_duplicate_email() {
        let pool = setup_test_db().await;
        create_user(&pool, "John", "john@example.com").await.unwrap();
        let result = create_user(&pool, "Jane", "john@example.com").await;
        assert!(matches!(result, Err(AppError::Validation(_))));
    }
}
```

---

## Advanced Patterns

### Performance Optimization

Go PGO Build:
```bash
# Collect production profile
GODEBUG=pgo=1 ./myapp -cpuprofile=default.pgo

# Build with PGO
go build -pgo=default.pgo -o myapp
```

Rust Release Optimization:
```toml
# Cargo.toml
[profile.release]
lto = true
codegen-units = 1
panic = "abort"
strip = true
opt-level = 3
```

### Memory Management

Go Memory Optimization:
```go
// Object pooling for high-frequency allocations
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 4096)
    },
}

func processRequest() {
    buf := bufferPool.Get().([]byte)
    defer bufferPool.Put(buf)
    // Use buffer
}
```

Rust Zero-Copy Patterns:
```rust
use bytes::Bytes;

async fn handle_body(body: Bytes) -> Result<Response, AppError> {
    // Bytes provides zero-copy slicing
    let data: Request = serde_json::from_slice(&body)?;
    Ok(Json(process(data)).into_response())
}
```

### Deployment Patterns

Go Container (10-20MB):
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o main .

FROM scratch
COPY --from=builder /app/main /main
EXPOSE 3000
ENTRYPOINT ["/main"]
```

Rust Container (5-15MB):
```dockerfile
FROM rust:1.91-alpine AS builder
WORKDIR /app
COPY Cargo.toml Cargo.lock ./
RUN mkdir src && echo "fn main(){}" > src/main.rs
RUN cargo build --release
COPY src ./src
RUN touch src/main.rs && cargo build --release

FROM alpine:latest
COPY --from=builder /app/target/release/app /app
EXPOSE 3000
CMD ["/app"]
```

---

## Context7 Integration

Library Documentation Access:
```python
# Go libraries
/golang/go                    # Go language and stdlib
/gofiber/fiber               # Fiber web framework
/labstack/echo               # Echo web framework
/go-chi/chi                  # Chi router
/go-gorm/gorm                # GORM ORM
/sqlc-dev/sqlc               # Type-safe SQL
/jackc/pgx                   # PostgreSQL driver
/spf13/cobra                 # CLI framework
/spf13/viper                 # Configuration

# Rust libraries
/rust-lang/rust              # Rust language
/tokio-rs/tokio              # Async runtime
/tokio-rs/axum               # Web framework
/actix/actix-web             # Actor-based web
/serde-rs/serde              # Serialization
/launchbadge/sqlx            # Async SQL
/diesel-rs/diesel            # ORM
/clap-rs/clap                # CLI parser
```

---

## Works Well With

- `moai-domain-backend` - REST API architecture and microservices patterns
- `moai-quality-security` - Security hardening for Go and Rust applications
- `moai-essentials-debug` - Performance profiling and debugging
- `moai-foundation-trust` - TRUST 5 quality validation
- `moai-workflow-tdd` - Test-driven development workflows

---

## Troubleshooting

Go Issues:
- Module errors: `go mod tidy && go mod verify`
- Version check: `go version && go env GOVERSION`
- Build issues: `go clean -cache && go build -v`

Rust Issues:
- Cargo errors: `cargo clean && cargo build`
- Version check: `rustc --version && cargo --version`
- Dependency issues: `cargo update && cargo tree`

---

## Additional Resources

See [reference.md](reference.md) for complete language reference, performance characteristics, and Context7 library mappings.

See [examples.md](examples.md) for production-ready code examples including REST APIs, gRPC services, CLI tools, and deployment configurations.

---

Last Updated: 2025-12-07
Version: 1.0.0
