# Rust Programming Examples

Production-ready code examples for Rust 1.91.

---

## REST API Examples

### Complete Axum API

```rust
// src/main.rs
use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    response::{IntoResponse, Response},
    routing::{delete, get, post, put},
    Json, Router,
};
use serde::{Deserialize, Serialize};
use sqlx::{postgres::PgPoolOptions, PgPool};
use std::net::SocketAddd;
use thiserror::Error;
use tokio::signal;
use tower_http::{cors::CorsLayer, trace::TraceLayer};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

#[derive(Clone)]
struct AppState {
    db: PgPool,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Tracing
    tracing_subscriber::registry()
        .with(tracing_subscriber::EnvFilter::new(
            std::env::var("RUST_LOG").unwrap_or_else(|_| "info".into()),
        ))
        .with(tracing_subscriber::fmt::layer())
        .init();

    // Database
    let database_url = std::env::var("DATABASE_URL")
        .unwrap_or_else(|_| "postgres://localhost/myapp".into());

    let pool = PgPoolOptions::new()
        .max_connections(25)
        .connect(&database_url)
        .await?;

    let state = AppState { db: pool };

    // Router
    let app = Router::new()
        .route("/health", get(health_check))
        .nest("/api/v1", api_routes())
        .layer(TraceLayer::new_for_http())
        .layer(CorsLayer::permissive())
        .with_state(state);

    // Server
    let addd = SocketAddd::from(([0, 0, 0, 0], 3000));
    tracing::info!("listening on {}", addd);

    let listener = tokio::net::TcpListener::bind(addd).await?;
    axum::serve(listener, app)
        .with_graceful_shutdown(shutdown_signal())
        .await?;

    Ok(())
}

fn api_routes() -> Router<AppState> {
    Router::new()
        .route("/users", get(list_users).post(create_user))
        .route(
            "/users/:id",
            get(get_user).put(update_user).delete(delete_user),
        )
}

async fn shutdown_signal() {
    let ctrl_c = async {
        signal::ctrl_c().await.expect("failed to install Ctrl+C handler");
    };

    #[cfg(unix)]
    let terminate = async {
        signal::unix::signal(signal::unix::SignalKind::terminate())
            .expect("failed to install signal handler")
            .recv()
            .await;
    };

    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();

    tokio::select! {
        _ = ctrl_c => {},
        _ = terminate => {},
    }

    tracing::info!("signal received, starting graceful shutdown");
}

// Models
#[derive(Debug, Serialize, Deserialize, sqlx::FromRow)]
struct User {
    id: i64,
    name: String,
    email: String,
    created_at: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Deserialize)]
struct CreateUserRequest {
    name: String,
    email: String,
}

#[derive(Debug, Deserialize)]
struct ListParams {
    limit: Option<i64>,
    offset: Option<i64>,
}

// Error handling
#[derive(Error, Debug)]
enum AppError {
    #[error("database error: {0}")]
    Database(#[from] sqlx::Error),
    #[error("not found: {0}")]
    NotFound(String),
    #[error("bad request: {0}")]
    BadRequest(String),
}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let (status, message) = match &self {
            AppError::NotFound(msg) => (StatusCode::NOT_FOUND, msg.clone()),
            AppError::BadRequest(msg) => (StatusCode::BAD_REQUEST, msg.clone()),
            AppError::Database(_) => {
                tracing::error!("Database error: {:?}", self);
                (StatusCode::INTERNAL_SERVER_ERROR, "Internal error".into())
            }
        };

        (status, Json(serde_json::json!({"error": message}))).into_response()
    }
}

// Handlers
async fn health_check() -> Json<serde_json::Value> {
    Json(serde_json::json!({"status": "ok"}))
}

async fn list_users(
    State(state): State<AppState>,
    Query(params): Query<ListParams>,
) -> Result<Json<Vec<User>>, AppError> {
    let limit = params.limit.unwrap_or(10);
    let offset = params.offset.unwrap_or(0);

    let users = sqlx::query_as!(
        User,
        r#"
        SELECT id, name, email, created_at
        FROM users
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
        "#,
        limit,
        offset
    )
    .fetch_all(&state.db)
    .await?;

    Ok(Json(users))
}

async fn get_user(
    State(state): State<AppState>,
    Path(id): Path<i64>,
) -> Result<Json<User>, AppError> {
    let user = sqlx::query_as!(
        User,
        "SELECT id, name, email, created_at FROM users WHERE id = $1",
        id
    )
    .fetch_optional(&state.db)
    .await?
    .ok_or_else(|| AppError::NotFound(format!("User {} not found", id)))?;

    Ok(Json(user))
}

async fn create_user(
    State(state): State<AppState>,
    Json(req): Json<CreateUserRequest>,
) -> Result<(StatusCode, Json<User>), AppError> {
    if req.name.len() < 2 {
        return Err(AppError::BadRequest("Name must be at least 2 characters".into()));
    }

    let user = sqlx::query_as!(
        User,
        r#"
        INSERT INTO users (name, email)
        VALUES ($1, $2)
        RETURNING id, name, email, created_at
        "#,
        req.name,
        req.email
    )
    .fetch_one(&state.db)
    .await?;

    Ok((StatusCode::CREATED, Json(user)))
}

async fn update_user(
    State(state): State<AppState>,
    Path(id): Path<i64>,
    Json(req): Json<CreateUserRequest>,
) -> Result<Json<User>, AppError> {
    let user = sqlx::query_as!(
        User,
        r#"
        UPDATE users
        SET name = $2, email = $3
        WHERE id = $1
        RETURNING id, name, email, created_at
        "#,
        id,
        req.name,
        req.email
    )
    .fetch_optional(&state.db)
    .await?
    .ok_or_else(|| AppError::NotFound(format!("User {} not found", id)))?;

    Ok(Json(user))
}

async fn delete_user(
    State(state): State<AppState>,
    Path(id): Path<i64>,
) -> Result<StatusCode, AppError> {
    let result = sqlx::query!("DELETE FROM users WHERE id = $1", id)
        .execute(&state.db)
        .await?;

    if result.rows_affected() == 0 {
        return Err(AppError::NotFound(format!("User {} not found", id)));
    }

    Ok(StatusCode::NO_CONTENT)
}
```

---

## CLI Tool Examples

### Complete CLI with clap

```rust
// src/main.rs
use clap::{Args, Parser, Subcommand};
use std::path::PathBuf;

#[derive(Parser)]
#[command(name = "myctl")]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(short, long, global = true)]
    config: Option<PathBuf>,

    #[arg(short, long, global = true)]
    verbose: bool,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// Start the API server
    Serve(ServeArgs),
    /// Run database migrations
    Migrate(MigrateArgs),
    /// Manage users
    User(UserArgs),
}

#[derive(Args)]
struct ServeArgs {
    #[arg(short, long, default_value = "3000")]
    port: u16,

    #[arg(long)]
    workers: Option<usize>,
}

#[derive(Args)]
struct MigrateArgs {
    #[command(subcommand)]
    command: MigrateCommands,
}

#[derive(Subcommand)]
enum MigrateCommands {
    /// Run all pending migrations
    Up,
    /// Rollback migrations
    Down {
        #[arg(short = 'n', long, default_value = "1")]
        steps: u32,
    },
}

#[derive(Args)]
struct UserArgs {
    #[command(subcommand)]
    command: UserCommands,
}

#[derive(Subcommand)]
enum UserCommands {
    /// List all users
    List {
        #[arg(short, long, default_value = "10")]
        limit: u32,
    },
    /// Create a new user
    Create {
        name: String,
        email: String,
    },
    /// Delete a user
    Delete {
        id: i64,
        #[arg(short, long)]
        force: bool,
    },
}

fn main() {
    let cli = Cli::parse();

    if cli.verbose {
        if let Some(config) = &cli.config {
            eprintln!("Using config: {:?}", config);
        }
    }

    match cli.command {
        Commands::Serve(args) => {
            println!("Starting server on port {}...", args.port);
            if let Some(workers) = args.workers {
                println!("Using {} workers", workers);
            }
        }
        Commands::Migrate(args) => match args.command {
            MigrateCommands::Up => {
                println!("Running migrations...");
            }
            MigrateCommands::Down { steps } => {
                println!("Rolling back {} migrations...", steps);
            }
        },
        Commands::User(args) => match args.command {
            UserCommands::List { limit } => {
                println!("Listing {} users...", limit);
            }
            UserCommands::Create { name, email } => {
                println!("Creating user: {} <{}>", name, email);
            }
            UserCommands::Delete { id, force } => {
                if force {
                    println!("Force deleting user {}...", id);
                } else {
                    println!("Deleting user {}...", id);
                }
            }
        },
    }
}
```

---

## Concurrency Examples

### Async Concurrency Patterns

```rust
use std::time::Duration;
use tokio::sync::{mpsc, Semaphore};
use tokio::time::sleep;

#[derive(Debug)]
struct Job {
    id: u32,
    data: String,
}

#[derive(Debug)]
struct Result {
    job_id: u32,
    data: String,
}

// Worker pool with channels
async fn worker_pool(mut rx: mpsc::Receiver<Job>, num_workers: usize) -> Vec<Result> {
    let (result_tx, mut result_rx) = mpsc::channel::<Result>(100);

    // Spawn workers
    for _ in 0..num_workers {
        let result_tx = result_tx.clone();
        tokio::spawn(async move {
            while let Some(job) = rx.recv().await {
                let result = process_job(job).await;
                let _ = result_tx.send(result).await;
            }
        });
    }
    drop(result_tx);

    // Collect results
    let mut results = Vec::new();
    while let Some(result) = result_rx.recv().await {
        results.push(result);
    }
    results
}

async fn process_job(job: Job) -> Result {
    sleep(Duration::from_millis(100)).await;
    Result {
        job_id: job.id,
        data: format!("Processed: {}", job.data),
    }
}

// Rate-limited operations with semaphore
async fn rate_limited_operations(items: Vec<String>, max_concurrent: usize) -> Vec<String> {
    let semaphore = std::sync::Arc::new(Semaphore::new(max_concurrent));
    let mut handles = Vec::new();

    for item in items {
        let sem = semaphore.clone();
        handles.push(tokio::spawn(async move {
            let _permit = sem.acquire().await.unwrap();
            process_item(item).await
        }));
    }

    let mut results = Vec::new();
    for handle in handles {
        if let Ok(result) = handle.await {
            results.push(result);
        }
    }
    results
}

async fn process_item(item: String) -> String {
    sleep(Duration::from_millis(100)).await;
    format!("Processed: {}", item)
}

// Select for multiple futures
async fn timeout_or_result() -> Option<String> {
    tokio::select! {
        result = fetch_data() => Some(result),
        _ = sleep(Duration::from_secs(5)) => {
            eprintln!("Timeout");
            None
        }
    }
}

async fn fetch_data() -> String {
    sleep(Duration::from_secs(1)).await;
    "Data fetched".to_string()
}
```

---

## Testing Examples

### Integration Tests

```rust
// tests/integration_test.rs
use axum::{body::Body, http::Request};
use sqlx::PgPool;
use tower::ServiceExt;

async fn setup_test_db() -> PgPool {
    let database_url = std::env::var("TEST_DATABASE_URL")
        .unwrap_or_else(|_| "postgres://test:test@localhost/test".into());

    let pool = PgPool::connect(&database_url).await.unwrap();

    // Run migrations
    sqlx::query(
        r#"
        CREATE TABLE IF NOT EXISTS users (
            id BIGSERIAL PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            email VARCHAR(255) UNIQUE NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )
        "#,
    )
    .execute(&pool)
    .await
    .unwrap();

    // Clean up
    sqlx::query("DELETE FROM users").execute(&pool).await.unwrap();

    pool
}

#[tokio::test]
async fn test_create_and_get_user() {
    let pool = setup_test_db().await;
    let app = create_app(pool.clone());

    // Create user
    let response = app
        .clone()
        .oneshot(
            Request::builder()
                .method("POST")
                .uri("/api/v1/users")
                .header("Content-Type", "application/json")
                .body(Body::from(r#"{"name": "John Doe", "email": "john@example.com"}"#))
                .unwrap(),
        )
        .await
        .unwrap();

    assert_eq!(response.status(), StatusCode::CREATED);

    let body = hyper::body::to_bytes(response.into_body()).await.unwrap();
    let created: User = serde_json::from_slice(&body).unwrap();
    assert_eq!(created.name, "John Doe");

    // Get user
    let response = app
        .oneshot(
            Request::builder()
                .uri(format!("/api/v1/users/{}", created.id))
                .body(Body::empty())
                .unwrap(),
        )
        .await
        .unwrap();

    assert_eq!(response.status(), StatusCode::OK);

    let body = hyper::body::to_bytes(response.into_body()).await.unwrap();
    let fetched: User = serde_json::from_slice(&body).unwrap();
    assert_eq!(fetched.id, created.id);
}

#[tokio::test]
async fn test_get_nonexistent_user() {
    let pool = setup_test_db().await;
    let app = create_app(pool);

    let response = app
        .oneshot(
            Request::builder()
                .uri("/api/v1/users/99999")
                .body(Body::empty())
                .unwrap(),
        )
        .await
        .unwrap();

    assert_eq!(response.status(), StatusCode::NOT_FOUND);
}
```

---

## Deployment Configuration

### Dockerfile

```dockerfile
FROM rust:1.91-alpine AS builder
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache musl-dev

# Cache dependencies
COPY Cargo.toml Cargo.lock ./
RUN mkdir src && echo "fn main(){}" > src/main.rs
RUN cargo build --release

# Build application
COPY src ./src
RUN touch src/main.rs && cargo build --release

# Runtime image
FROM alpine:latest
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/target/release/app /app

EXPOSE 3000
CMD ["/app"]
```

### Cargo.toml

```toml
[package]
name = "myapp"
version = "0.1.0"
edition = "2021"

[dependencies]
axum = "0.8"
tokio = { version = "1.48", features = ["full"] }
tower = "0.5"
tower-http = { version = "0.6", features = ["cors", "trace"] }
sqlx = { version = "0.8", features = ["runtime-tokio", "postgres", "chrono"] }
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
thiserror = "2.0"
anyhow = "1.0"
chrono = { version = "0.4", features = ["serde"] }
tracing = "0.1"
tracing-subscriber = { version = "0.3", features = ["env-filter"] }
clap = { version = "4.5", features = ["derive"] }

[profile.release]
lto = true
codegen-units = 1
panic = "abort"
strip = true
opt-level = 3
```

---

Last Updated: 2025-12-07
Version: 1.0.0
