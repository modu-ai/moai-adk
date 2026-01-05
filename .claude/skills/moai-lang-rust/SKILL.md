---
name: moai-lang-rust
description: Rust 1.92+ development specialist covering Axum, Tokio, SQLx, and memory-safe systems programming. Use when building high-performance, memory-safe applications or WebAssembly.
version: 1.1.0
updated: 2025-12-30
status: active
---

## Quick Reference (30 seconds)

Rust 1.92+ Development Specialist with deep patterns for high-performance, memory-safe applications.

Auto-Triggers: `.rs`, `Cargo.toml`, async/await, Tokio, Axum, SQLx, serde, lifetimes, traits

Core Use Cases:
- High-performance REST APIs and microservices
- Memory-safe concurrent systems
- CLI tools and system utilities
- WebAssembly applications
- Low-latency networking services

Quick Patterns:

Axum REST API:
```rust
let app = Router::new()
    .route("/api/users/:id", get(get_user))
    .with_state(app_state);
let listener = tokio::net::TcpListener::bind("0.0.0.0:3000").await?;
axum::serve(listener, app).await?;
```

Async Handler with SQLx:
```rust
async fn get_user(
    State(state): State<AppState>,
    Path(id): Path<i64>,
) -> Result<Json<User>, AppError> {
    let user = sqlx::query_as!(User, "SELECT * FROM users WHERE id = $1", id)
        .fetch_optional(&state.db).await?
        .ok_or(AppError::NotFound)?;
    Ok(Json(user))
}
```

---

## Implementation Guide (5 minutes)

### Rust 1.92 Features

Modern Rust Features:
- Rust 2024 Edition available (released with Rust 1.85)
- Async traits in stable (no more async-trait crate needed)
- Const generics for compile-time array sizing
- let-else for pattern matching with early return
- Improved borrow checker with polonius

Async Traits (Stable):
```rust
trait AsyncRepository {
    async fn get(&self, id: i64) -> Result<User, Error>;
    async fn create(&self, user: CreateUser) -> Result<User, Error>;
}

impl AsyncRepository for PostgresRepository {
    async fn get(&self, id: i64) -> Result<User, Error> {
        sqlx::query_as!(User, "SELECT * FROM users WHERE id = $1", id)
            .fetch_one(&self.pool).await
    }
}
```

Let-Else Pattern:
```rust
fn get_user(id: Option<i64>) -> Result<User, Error> {
    let Some(id) = id else { return Err(Error::MissingId); };
    let Ok(user) = repository.find(id) else { return Err(Error::NotFound); };
    Ok(user)
}
```

### Web Framework: Axum 0.8

Installation:
```toml
[dependencies]
axum = "0.8"
tokio = { version = "1.48", features = ["full"] }
tower-http = { version = "0.6", features = ["cors", "trace"] }
```

Complete API Setup:
```rust
use axum::{extract::{Path, State, Query}, routing::{get, post}, Router, Json};
use tower_http::cors::CorsLayer;

#[derive(Clone)]
struct AppState { db: PgPool }

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let pool = PgPoolOptions::new()
        .max_connections(25)
        .connect(&std::env::var("DATABASE_URL")?).await?;

    let app = Router::new()
        .route("/api/v1/users", get(list_users).post(create_user))
        .route("/api/v1/users/:id", get(get_user))
        .layer(CorsLayer::permissive())
        .with_state(AppState { db: pool });

    let listener = tokio::net::TcpListener::bind("0.0.0.0:3000").await?;
    axum::serve(listener, app).await?;
    Ok(())
}
```

Handler Patterns:
```rust
async fn list_users(
    State(state): State<AppState>,
    Query(params): Query<ListParams>,
) -> Result<Json<Vec<User>>, AppError> {
    let users = sqlx::query_as!(User,
        "SELECT * FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2",
        params.limit.unwrap_or(10), params.offset.unwrap_or(0))
        .fetch_all(&state.db).await?;
    Ok(Json(users))
}
```

### Async Runtime: Tokio 1.48

Task Spawning and Channels:
```rust
use tokio::sync::mpsc;

async fn worker_pool() {
    let (tx, mut rx) = mpsc::channel::<Job>(100);
    for _ in 0..4 {
        tokio::spawn(async move {
            while let Some(job) = rx.recv().await { process_job(job).await; }
        });
    }
}

async fn timeout_operation() -> Result<Data, Error> {
    tokio::select! {
        result = fetch_data() => result,
        _ = tokio::time::sleep(Duration::from_secs(5)) => Err(Error::Timeout),
    }
}
```

### Database: SQLx 0.8

Type-Safe Queries:
```rust
#[derive(Debug, sqlx::FromRow)]
struct User { id: i64, name: String, email: String }

async fn user_operations(pool: &PgPool) -> Result<(), sqlx::Error> {
    let user = sqlx::query_as!(User,
        "SELECT id, name, email FROM users WHERE id = $1", 1i64)
        .fetch_one(pool).await?;

    let mut tx = pool.begin().await?;
    sqlx::query!("INSERT INTO users (name, email) VALUES ($1, $2)", "John", "john@example.com")
        .execute(&mut *tx).await?;
    tx.commit().await?;
    Ok(())
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
    #[serde(skip_serializing_if = "Option::is_none")]
    profile_url: Option<String>,
    #[serde(default)]
    is_active: bool,
}
```

### Error Handling

thiserror:
```rust
use thiserror::Error;

#[derive(Error, Debug)]
pub enum AppError {
    #[error("database error: {0}")]
    Database(#[from] sqlx::Error),
    #[error("not found: {0}")]
    NotFound(String),
    #[error("unauthorized")]
    Unauthorized,
}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let (status, message) = match &self {
            AppError::NotFound(_) => (StatusCode::NOT_FOUND, self.to_string()),
            AppError::Unauthorized => (StatusCode::UNAUTHORIZED, self.to_string()),
            AppError::Database(_) => (StatusCode::INTERNAL_SERVER_ERROR, "Internal error".into()),
        };
        (status, Json(json!({"error": message}))).into_response()
    }
}
```

### CLI Development: clap

```rust
use clap::{Parser, Subcommand};

#[derive(Parser)]
#[command(name = "myapp", version, about)]
struct Cli {
    #[arg(short, long, global = true)]
    config: Option<PathBuf>,
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    Serve { #[arg(short, long, default_value = "3000")] port: u16 },
    Migrate,
}

fn main() {
    let cli = Cli::parse();
    match cli.command {
        Commands::Serve { port } => serve(port),
        Commands::Migrate => migrate(),
    }
}
```

### Testing Patterns

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
}
```

---

## Advanced Patterns

### Performance Optimization

Release Build:
```toml
[profile.release]
lto = true
codegen-units = 1
panic = "abort"
strip = true
```

### Deployment

Minimal Container (5-15MB):
```dockerfile
FROM rust:1.92-alpine AS builder
WORKDIR /app
COPY Cargo.toml Cargo.lock ./
RUN mkdir src && echo "fn main(){}" > src/main.rs && cargo build --release
COPY src ./src
RUN touch src/main.rs && cargo build --release

FROM alpine:latest
COPY --from=builder /app/target/release/app /app
EXPOSE 3000
CMD ["/app"]
```

### Concurrency

Rate-Limited Operations:
```rust
use tokio::sync::Semaphore;

async fn rate_limited(items: Vec<String>, max: usize) -> Vec<String> {
    let sem = std::sync::Arc::new(Semaphore::new(max));
    let handles: Vec<_> = items.into_iter().map(|item| {
        let sem = sem.clone();
        tokio::spawn(async move {
            let _permit = sem.acquire().await.unwrap();
            process_item(item).await
        })
    }).collect();
    futures::future::join_all(handles).await.into_iter().filter_map(|r| r.ok()).collect()
}
```

---

## Context7 Integration

Library Documentation Access:
- `/rust-lang/rust` - Rust language and stdlib
- `/tokio-rs/tokio` - Tokio async runtime
- `/tokio-rs/axum` - Axum web framework
- `/launchbadge/sqlx` - SQLx async SQL
- `/serde-rs/serde` - Serialization framework
- `/dtolnay/thiserror` - Error derive
- `/clap-rs/clap` - CLI parser

---

## Works Well With

- `moai-lang-go` - Go systems programming patterns
- `moai-domain-backend` - REST API architecture and microservices patterns
- `moai-foundation-quality` - Security hardening for Rust applications
- `moai-workflow-testing` - Test-driven development workflows

---

## Troubleshooting

Common Issues:
- Cargo errors: `cargo clean && cargo build`
- Version check: `rustc --version && cargo --version`
- Dependency issues: `cargo update && cargo tree`
- Compile-time SQL check: `cargo sqlx prepare`

Performance Characteristics:
- Startup Time: 50-100ms
- Memory Usage: 5-20MB base
- Throughput: 100k-200k req/s
- Latency: p99 less than 5ms
- Container Size: 5-15MB (alpine)

---

## Additional Resources

See [reference.md](reference.md) for complete language reference and Context7 library mappings.

See [examples.md](examples.md) for production-ready code examples.

---

Last Updated: 2025-12-30
Version: 1.1.0
