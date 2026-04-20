# Rust 1.91 Reference

Complete reference for Rust 1.91 systems programming.

---

## Language Features

### Async Traits (Stable)

```rust
// No more async-trait crate needed
trait AsyncRepository {
    async fn get(&self, id: i64) -> Result<User, Error>;
    async fn create(&self, user: CreateUser) -> Result<User, Error>;
    async fn update(&self, id: i64, user: UpdateUser) -> Result<User, Error>;
    async fn delete(&self, id: i64) -> Result<(), Error>;
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

    async fn update(&self, id: i64, user: UpdateUser) -> Result<User, Error> {
        sqlx::query_as!(User,
            "UPDATE users SET name = $2, email = $3 WHERE id = $1 RETURNING *",
            id, user.name, user.email)
            .fetch_one(&self.pool)
            .await
    }

    async fn delete(&self, id: i64) -> Result<(), Error> {
        sqlx::query!("DELETE FROM users WHERE id = $1", id)
            .execute(&self.pool)
            .await?;
        Ok(())
    }
}
```

### Const Generics

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

impl<T: Default + Copy, const SIZE: usize> Buffer<T, SIZE> {
    fn new() -> Self {
        Self {
            data: [T::default(); SIZE],
            len: 0,
        }
    }

    fn push(&mut self, item: T) -> Result<(), &'static str> {
        if self.len >= SIZE {
            return Err("Buffer full");
        }
        self.data[self.len] = item;
        self.len += 1;
        Ok(())
    }
}
```

### Let-Else Pattern

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

// Complex pattern matching
fn process_response(response: Option<Response>) -> Result<Data, Error> {
    let Some(Response { status: 200, body: Some(data), .. }) = response else {
        return Err(Error::InvalidResponse);
    };

    Ok(parse_data(data)?)
}
```

---

## Web Framework: Axum 0.8

### Complete API Setup

```rust
use axum::{
    extract::{Path, State, Query, Json as JsonExtract},
    http::StatusCode,
    response::{IntoResponse, Response, Json},
    routing::{get, post, put, delete},
    Router,
};
use tower_http::cors::{CorsLayer, Any};
use tower_http::trace::TraceLayer;
use tower_http::timeout::TimeoutLayer;
use std::time::Duration;

#[derive(Clone)]
struct AppState {
    db: PgPool,
    redis: RedisPool,
}

pub fn create_app(state: AppState) -> Router {
    Router::new()
        .route("/health", get(health_check))
        .nest("/api/v1", api_routes())
        .layer(TraceLayer::new_for_http())
        .layer(TimeoutLayer::new(Duration::from_secs(30)))
        .layer(CorsLayer::new()
            .allow_origin(Any)
            .allow_methods(Any)
            .allow_headers(Any))
        .with_state(state)
}

fn api_routes() -> Router<AppState> {
    Router::new()
        .route("/users", get(list_users).post(create_user))
        .route("/users/:id", get(get_user).put(update_user).delete(delete_user))
        .route("/posts", get(list_posts).post(create_post))
        .route("/posts/:id", get(get_post).put(update_post).delete(delete_post))
}
```

### Extractors

```rust
use axum::extract::{Path, Query, State, Json, Extension, ConnectInfo};

// Path parameters
async fn get_user(Path(id): Path<i64>) -> impl IntoResponse { ... }

// Multiple path parameters
async fn get_post_comment(
    Path((post_id, comment_id)): Path<(i64, i64)>
) -> impl IntoResponse { ... }

// Query parameters
#[derive(Deserialize)]
struct ListParams {
    limit: Option<i64>,
    offset: Option<i64>,
    sort: Option<String>,
}

async fn list_users(Query(params): Query<ListParams>) -> impl IntoResponse { ... }

// JSON body
async fn create_user(Json(req): Json<CreateUserRequest>) -> impl IntoResponse { ... }

// State injection
async fn handler(State(state): State<AppState>) -> impl IntoResponse { ... }

// Combined extractors
async fn complex_handler(
    State(state): State<AppState>,
    Path(id): Path<i64>,
    Query(params): Query<ListParams>,
    Json(body): Json<UpdateRequest>,
) -> Result<Json<Response>, AppError> { ... }
```

### Middleware

```rust
use axum::middleware::{self, Next};
use axum::http::Request;

async fn auth_middleware<B>(
    State(state): State<AppState>,
    request: Request<B>,
    next: Next<B>,
) -> Result<Response, AppError> {
    let token = request
        .headers()
        .get("Authorization")
        .and_then(|v| v.to_str().ok())
        .and_then(|v| v.strip_prefix("Bearer "))
        .ok_or(AppError::Unauthorized)?;

    let claims = verify_token(token, &state.jwt_secret)?;

    let mut request = request;
    request.extensions_mut().insert(claims);

    Ok(next.run(request).await)
}

// Apply middleware
let protected_routes = Router::new()
    .route("/users/me", get(get_current_user))
    .layer(middleware::from_fn_with_state(state.clone(), auth_middleware));
```

---

## Async Runtime: Tokio 1.48

### Task Management

```rust
use tokio::task::{JoinHandle, JoinSet};

// Spawn tasks
let handle: JoinHandle<i32> = tokio::spawn(async {
    // async work
    42
});

// JoinSet for multiple tasks
let mut set = JoinSet::new();
for i in 0..10 {
    set.spawn(async move {
        process(i).await
    });
}

while let Some(result) = set.join_next().await {
    match result {
        Ok(value) => println!("Task completed: {:?}", value),
        Err(e) => eprintln!("Task failed: {:?}", e),
    }
}
```

### Channels

```rust
use tokio::sync::{mpsc, oneshot, broadcast, watch};

// Multi-producer, single-consumer
let (tx, mut rx) = mpsc::channel::<Message>(100);

tokio::spawn(async move {
    while let Some(msg) = rx.recv().await {
        process(msg).await;
    }
});

tx.send(Message::new()).await?;

// One-shot for single response
let (tx, rx) = oneshot::channel::<Response>();

tokio::spawn(async move {
    let response = compute().await;
    let _ = tx.send(response);
});

let response = rx.await?;

// Broadcast for multiple consumers
let (tx, _) = broadcast::channel::<Event>(100);
let mut rx1 = tx.subscribe();
let mut rx2 = tx.subscribe();

// Watch for single-value updates
let (tx, rx) = watch::channel(Config::default());
```

### Select

```rust
use tokio::time::{timeout, sleep, Duration};

// Select multiple futures
async fn with_timeout() -> Result<Data, Error> {
    tokio::select! {
        result = fetch_data() => result,
        _ = sleep(Duration::from_secs(5)) => Err(Error::Timeout),
    }
}

// Biased selection
tokio::select! {
    biased;
    _ = shutdown_signal() => {
        println!("Shutting down");
    }
    result = server.serve() => {
        println!("Server stopped: {:?}", result);
    }
}

// Timeout helper
let result = timeout(Duration::from_secs(10), async_operation()).await??;
```

---

## Database: SQLx 0.8

### Connection Pool

```rust
use sqlx::{PgPool, postgres::PgPoolOptions};

async fn create_pool() -> Result<PgPool, sqlx::Error> {
    PgPoolOptions::new()
        .max_connections(25)
        .min_connections(5)
        .acquire_timeout(Duration::from_secs(5))
        .idle_timeout(Duration::from_secs(600))
        .max_lifetime(Duration::from_secs(1800))
        .connect(&std::env::var("DATABASE_URL")?)
        .await
}
```

### Compile-Time Checked Queries

```rust
// Requires DATABASE_URL at compile time
let user = sqlx::query_as!(User,
    r#"
    SELECT id, name, email, created_at
    FROM users
    WHERE id = $1
    "#,
    id
)
.fetch_one(&pool)
.await?;

// Dynamic query
let users = sqlx::query_as::<_, User>(
    "SELECT * FROM users WHERE name LIKE $1"
)
.bind(format!("%{}%", search))
.fetch_all(&pool)
.await?;
```

### Transactions

```rust
async fn transfer_funds(pool: &PgPool, from: i64, to: i64, amount: i64) -> Result<(), Error> {
    let mut tx = pool.begin().await?;

    sqlx::query!(
        "UPDATE accounts SET balance = balance - $1 WHERE id = $2",
        amount, from
    )
    .execute(&mut *tx)
    .await?;

    sqlx::query!(
        "UPDATE accounts SET balance = balance + $1 WHERE id = $2",
        amount, to
    )
    .execute(&mut *tx)
    .await?;

    tx.commit().await?;
    Ok(())
}
```

---

## Error Handling

### thiserror

```rust
use thiserror::Error;

#[derive(Error, Debug)]
pub enum AppError {
    #[error("database error: {0}")]
    Database(#[from] sqlx::Error),

    #[error("validation error: {field} - {message}")]
    Validation { field: String, message: String },

    #[error("not found: {resource} with id {id}")]
    NotFound { resource: &'static str, id: i64 },

    #[error("unauthorized: {0}")]
    Unauthorized(String),

    #[error("internal error")]
    Internal(#[source] anyhow::Error),
}
```

### Axum Error Response

```rust
impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let (status, error_code, message) = match &self {
            AppError::NotFound { .. } => (
                StatusCode::NOT_FOUND,
                "NOT_FOUND",
                self.to_string()
            ),
            AppError::Validation { .. } => (
                StatusCode::BAD_REQUEST,
                "VALIDATION_ERROR",
                self.to_string()
            ),
            AppError::Unauthorized(_) => (
                StatusCode::UNAUTHORIZED,
                "UNAUTHORIZED",
                self.to_string()
            ),
            AppError::Database(_) | AppError::Internal(_) => {
                tracing::error!("Internal error: {:?}", self);
                (
                    StatusCode::INTERNAL_SERVER_ERROR,
                    "INTERNAL_ERROR",
                    "Internal server error".to_string()
                )
            }
        };

        (status, Json(json!({
            "error": {
                "code": error_code,
                "message": message
            }
        }))).into_response()
    }
}
```

---

## Context7 Library Mappings

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

- Startup Time: 50-100ms
- Memory Usage: 5-20MB base
- Throughput: 100k-200k req/s
- Latency: p99 less than 5ms
- Container Image Size: 5-15MB (alpine base)

---

Last Updated: 2025-12-07
Version: 1.0.0
