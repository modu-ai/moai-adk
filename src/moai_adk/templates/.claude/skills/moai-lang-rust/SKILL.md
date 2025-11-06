---
name: moai-lang-rust
version: 2.0.0
created: 2025-11-06
updated: 2025-11-06
status: active
description: "Rust best practices with systems programming, performance-critical applications, and memory-safe patterns for 2025"
keywords: [rust, systems, performance, memory-safety, concurrency, web-services, embedded, blockchain]
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - WebFetch
  - WebSearch
---

# Rust Development Mastery

**Modern Rust Development with 2025 Best Practices**

> Comprehensive Rust development guidance covering systems programming, high-performance applications, web services, embedded development, and production-safe code using the latest tools and methodologies.

## What It Does

- **Systems Programming**: OS-level utilities, embedded systems, low-level optimizations
- **High-Performance Web Services**: Actix-web, Axum, and async frameworks
- **Memory-Safe Concurrency**: Fearless concurrency with ownership and borrowing
- **CLI Applications**: Fast, safe command-line tools and utilities
- **Blockchain & Crypto**: Smart contracts, DeFi protocols, cryptographic applications
- **Game Development**: Game engines, real-time systems, graphics programming
- **DevOps & Infrastructure**: Monitoring tools, containerized applications, cloud services

## When to Use

### Perfect Scenarios
- **Performance-critical applications requiring zero-cost abstractions**
- **Systems programming and embedded development**
- **Web services requiring maximum throughput and security**
- **CLI tools and developer utilities**
- **Blockchain and cryptocurrency applications**
- **Game engines and real-time systems**
- **Applications requiring memory safety without garbage collection**

### Common Triggers
- "Create Rust web service"
- "Build Rust CLI tool"
- "Rust performance optimization"
- "Rust memory safety"
- "Systems programming with Rust"
- "Async Rust patterns"

## Tool Version Matrix (2025-11-06)

### Core Rust
- **Rust**: 1.91.x (latest) / 1.89.x (LTS)
- **Cargo**: Built-in package manager and build tool
- **Rustup**: Rust version manager
- **Rust-analyzer**: Advanced LSP support

### Web Frameworks
- **Actix-web**: 4.9.x - High-performance web framework
- **Axum**: 0.8.x - Web framework from tokio team
- **Rocket**: 0.6.x - Web framework with macros
- **Poem**: 3.1.x - Modular web framework
- **Warp**: 0.4.x - Composable web server

### Async Runtime
- **Tokio**: 1.42.x - Asynchronous runtime
- **Async-std**: 1.13.x - Async standard library
- **Smol**: 2.0.x - Compact async runtime

### Database & Storage
- **SQLx**: 0.8.x - Async SQL toolkit
- **Diesel**: 2.2.x - ORM and query builder
- **Serde**: 1.0.x - Serialization framework
- **Redis**: 0.26.x - Redis client

### Testing Tools
- **Tokio-test**: 0.4.x - Async testing utilities
- **Mockall**: 0.13.x - Mocking framework
- **Cucumber-rs**: 0.22.x - BDD testing
- **Proptest**: 1.5.x - Property-based testing

### Development Tools
- **Clippy**: Built-in linter
- **rustfmt**: Built-in formatter
- **cargo-watch**: 8.5.x - File watcher
- **cargo-audit**: 0.21.x - Security audit

## Ecosystem Overview

### Project Setup (2025 Best Practice)

```bash
# Modern Rust project with workspace structure
cargo new my-rust-project --bin
cd my-rust-project

# Create workspace structure
mkdir -p crates/{common,server,client}
mkdir -p src/{bin,lib}
mkdir -p tests/{integration,unit}
mkdir -p benches
mkdir -p examples

# Initialize workspace
echo '[workspace]
members = [
    "crates/common",
    "crates/server", 
    "crates/client",
]' > Cargo.toml

# Add workspace crates
cargo new --lib crates/common
cargo new --lib crates/server
cargo new --lib crates/client

# Install development tools
cargo install cargo-watch
cargo install cargo-audit
cargo install cargo-expand
cargo install cargo-flamegraph
```

### Modern Project Structure

```
my-rust-project/
├── Cargo.toml              # Workspace configuration
├── Cargo.lock              # Lock file
├── rust-toolchain.toml     # Toolchain specification
├── .clippy.toml           # Clippy configuration
├── .rustfmt.toml          # Rustfmt configuration
├── deny.toml              # Dependency deny list
├── crates/
│   ├── common/            # Shared utilities
│   │   ├── Cargo.toml
│   │   └── src/
│   ├── server/            # Server application
│   │   ├── Cargo.toml
│   │   └── src/
│   └── client/            # Client library
│       ├── Cargo.toml
│       └── src/
├── src/
│   ├── bin/               # Binary entry points
│   └── lib/               # Main library code
├── tests/
│   ├── integration/       # Integration tests
│   └── unit/              # Unit tests
├── benches/               # Performance benchmarks
├── examples/              # Example usage
├── docs/                  # Documentation
└── scripts/               # Build and utility scripts
```

## Modern Rust Patterns

### Error Handling with `thiserror` and `anyhow`

```rust
// src/error.rs
use thiserror::Error;

#[derive(Error, Debug)]
pub enum AppError {
    #[error("Database error: {0}")]
    Database(#[from] sqlx::Error),
    
    #[error("Validation error: {message}")]
    Validation { message: String },
    
    #[error("Not found: {entity} with id {id}")]
    NotFound { entity: String, id: String },
    
    #[error("Authentication failed")]
    Authentication,
    
    #[error("Authorization denied for {action}")]
    Authorization { action: String },
    
    #[error("Internal server error: {0}")]
    Internal(#[from] anyhow::Error),
}

pub type Result<T> = std::result::Result<T, AppError>;

// Usage in application code
use anyhow::Context;

pub async fn get_user(id: &str) -> Result<User> {
    let user = sqlx::query_as!(
        User,
        "SELECT * FROM users WHERE id = ?",
        id
    )
    .fetch_one(&pool)
    .await
    .with_context(|| format!("Failed to fetch user with id: {}", id))?;
    
    Ok(user)
}
```

### Async Programming with Tokio

```rust
// src/services/user_service.rs
use tokio::sync::{RwLock, broadcast};
use std::sync::Arc;
use uuid::Uuid;

pub struct UserService {
    users: Arc<RwLock<std::collections::HashMap<Uuid, User>>>,
    event_tx: broadcast::Sender<UserEvent>,
}

impl UserService {
    pub fn new() -> Self {
        let (event_tx, _) = broadcast::channel(1000);
        Self {
            users: Arc::new(RwLock::new(std::collections::HashMap::new())),
            event_tx,
        }
    }
    
    pub async fn create_user(&self, user_data: CreateUserRequest) -> Result<User> {
        let user_id = Uuid::new_v4();
        let user = User {
            id: user_id,
            email: user_data.email.clone(),
            name: user_data.name,
            created_at: chrono::Utc::now(),
        };
        
        // Insert into user store
        {
            let mut users = self.users.write().await;
            users.insert(user_id, user.clone());
        }
        
        // Emit event
        let _ = self.event_tx.send(UserEvent::Created(user.clone()));
        
        Ok(user)
    }
    
    pub async fn get_user(&self, user_id: Uuid) -> Option<User> {
        let users = self.users.read().await;
        users.get(&user_id).cloned()
    }
    
    pub async fn list_users(&self) -> Vec<User> {
        let users = self.users.read().await;
        users.values().cloned().collect()
    }
    
    pub fn subscribe_events(&self) -> broadcast::Receiver<UserEvent> {
        self.event_tx.subscribe()
    }
}

#[derive(Debug, Clone)]
pub enum UserEvent {
    Created(User),
    Updated(User),
    Deleted { id: Uuid },
}

// Concurrent processing example
pub async fn process_users_concurrently(
    user_service: Arc<UserService>,
    users: Vec<User>,
) -> Vec<Result<ProcessedUser>> {
    let semaphore = Arc::new(tokio::sync::Semaphore::new(10)); // Limit concurrency
    
    let tasks: Vec<_> = users
        .into_iter()
        .map(|user| {
            let user_service = user_service.clone();
            let semaphore = semaphore.clone();
            
            tokio::spawn(async move {
                let _permit = semaphore.acquire().await.unwrap();
                process_single_user(&user_service, user).await
            })
        })
        .collect();
    
    let results = futures::future::join_all(tasks).await;
    results.into_iter().map(|r| r.unwrap()).collect()
}
```

### Web Service with Axum

```rust
// src/main.rs
use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use serde::{Deserialize, Serialize};
use std::net::SocketAddr;
use tower::ServiceBuilder;
use tower_http::{
    cors::{Any, CorsLayer},
    trace::TraceLayer,
};
use tracing::{info, Level};
use tracing_subscriber;

mod error;
mod handlers;
mod models;
mod services;
mod middleware;

use error::AppError;
use handlers::{create_user, get_user, health_check};
use services::UserService;

#[derive(Clone)]
pub struct AppState {
    pub user_service: Arc<UserService>,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Initialize tracing
    tracing_subscriber::fmt()
        .with_max_level(Level::INFO)
        .init();
    
    // Initialize services
    let user_service = Arc::new(UserService::new());
    let app_state = AppState { user_service };
    
    // Build application
    let app = Router::new()
        // Routes
        .route("/health", get(health_check))
        .route("/users", post(create_user))
        .route("/users/:id", get(get_user))
        // Middleware
        .layer(
            ServiceBuilder::new()
                .layer(TraceLayer::new_for_http())
                .layer(CorsLayer::new().allow_origin(Any))
                .layer(axum::middleware::from_fn(middleware::request_logging)),
        )
        .with_state(app_state);
    
    // Run server
    let addr = SocketAddr::from(([0, 0, 0, 0], 8080));
    info!("Server listening on {}", addr);
    
    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await?;
    
    Ok(())
}

// src/handlers/user.rs
use axum::{extract::{Path, State}, Json};
use uuid::Uuid;
use crate::{error::AppError, models::User, services::UserService, AppState};

pub async fn create_user(
    State(state): State<AppState>,
    Json(payload): Json<CreateUserRequest>,
) -> Result<Json<User>, AppError> {
    let user = state.user_service.create_user(payload).await?;
    Ok(Json(user))
}

pub async fn get_user(
    State(state): State<AppState>,
    Path(user_id): Path<Uuid>,
) -> Result<Json<User>, AppError> {
    let user = state
        .user_service
        .get_user(user_id)
        .await
        .ok_or(AppError::NotFound {
            entity: "User".to_string(),
            id: user_id.to_string(),
        })?;
    
    Ok(Json(user))
}

pub async fn health_check() -> impl IntoResponse {
    Json(serde_json::json!({
        "status": "healthy",
        "timestamp": chrono::Utc::now(),
    }))
}
```

## Performance Optimization

### Memory Management and Zero-Copy Patterns

```rust
// src/processing/data_processor.rs
use bytes::{Bytes, BytesMut, BufMut};
use std::io::{self, Read};

pub struct DataProcessor {
    buffer: BytesMut,
}

impl DataProcessor {
    pub fn new() -> Self {
        Self {
            buffer: BytesMut::with_capacity(8192),
        }
    }
    
    // Zero-copy data processing
    pub fn process_stream<R: Read>(&mut self, mut reader: R) -> io::Result<ProcessedData> {
        self.buffer.clear();
        
        // Read data into buffer
        let bytes_read = reader.read(&mut self.buffer)?;
        if bytes_read == 0 {
            return Ok(ProcessedData::empty());
        }
        
        // Process without copying
        let data = self.buffer.split().freeze();
        self.process_bytes(data)
    }
    
    fn process_bytes(&self, data: Bytes) -> io::Result<ProcessedData> {
        let mut records = Vec::new();
        
        // Process data without allocations
        for line in data.split(|&b| b == b'\n') {
            if line.is_empty() {
                continue;
            }
            
            let record = self.parse_record(line)?;
            records.push(record);
        }
        
        Ok(ProcessedData { records })
    }
    
    fn parse_record(&self, line: &[u8]) -> io::Result<Record> {
        // Parse using slices instead of String allocations
        let mut parts = line.split(|&b| b == b',');
        
        let id_parts = parts.next().ok_or(io::ErrorKind::InvalidData)?;
        let name_parts = parts.next().ok_or(io::ErrorKind::InvalidData)?;
        
        Ok(Record {
            id: std::str::from_utf8(id_parts)
                .map_err(|_| io::Error::new(io::ErrorKind::InvalidData, "Invalid UTF-8"))?
                .to_string(),
            name: std::str::from_utf8(name_parts)
                .map_err(|_| io::Error::new(io::ErrorKind::InvalidData, "Invalid UTF-8"))?
                .to_string(),
        })
    }
}

// SIMD-optimized processing
#[cfg(target_arch = "x86_64")]
use std::arch::x86_64::*;

pub fn fast_checksum(data: &[u8]) -> u64 {
    #[target_feature(enable = "avx2")]
    unsafe {
        if is_x86_feature_detected!("avx2") {
            checksum_avx2(data)
        } else {
            checksum_scalar(data)
        }
    }
}

#[target_feature(enable = "avx2")]
unsafe fn checksum_avx2(data: &[u8]) -> u64 {
    let chunks = data.chunks_exact(32);
    let remainder = chunks.remainder();
    
    let mut sum = _mm256_setzero_si256();
    
    for chunk in chunks {
        let chunk = _mm256_loadu_si256(chunk.as_ptr() as *const __m256i);
        sum = _mm256_add_epi64(sum, _mm256_sad_epu8(chunk, _mm256_setzero_si256()));
    }
    
    let mut result = _mm256_extract_epi128(sum, 1) as u64 + _mm256_extract_epi128(sum, 0) as u64;
    
    // Process remainder
    for &byte in remainder {
        result += byte as u64;
    }
    
    result
}

fn checksum_scalar(data: &[u8]) -> u64 {
    data.iter().map(|&b| b as u64).sum()
}
```

### Efficient Data Structures

```rust
// src/datastructures/high_performance_map.rs
use std::collections::HashMap;
use std::hash::{Hash, Hasher};

// Custom hash map optimized for string keys
pub struct StringKeyMap<V> {
    inner: HashMap<Box<str>, V>,
    key_pool: Vec<Box<str>>,
}

impl<V> StringKeyMap<V> {
    pub fn new() -> Self {
        Self {
            inner: HashMap::new(),
            key_pool: Vec::new(),
        }
    }
    
    pub fn insert(&mut self, key: &str, value: V) -> Option<V> {
        // Intern the key to avoid allocations for duplicate keys
        let interned_key = self.intern_key(key);
        self.inner.insert(interned_key, value)
    }
    
    pub fn get(&self, key: &str) -> Option<&V> {
        self.inner.get(key)
    }
    
    pub fn get_mut(&mut self, key: &str) -> Option<&mut V> {
        self.inner.get_mut(key)
    }
    
    pub fn remove(&mut self, key: &str) -> Option<V> {
        self.inner.remove(key)
    }
    
    fn intern_key(&mut self, key: &str) -> Box<str> {
        // Simple string interning - in production, use a proper interner
        for interned in &self.key_pool {
            if interned.as_ref() == key {
                return interned.clone();
            }
        }
        
        let interned: Box<str> = key.into();
        self.key_pool.push(interned.clone());
        interned
    }
}

// High-performance ring buffer
pub struct RingBuffer<T> {
    buffer: Vec<Option<T>>,
    head: usize,
    tail: usize,
    size: usize,
    capacity: usize,
}

impl<T> RingBuffer<T> {
    pub fn new(capacity: usize) -> Self {
        Self {
            buffer: vec![None; capacity],
            head: 0,
            tail: 0,
            size: 0,
            capacity,
        }
    }
    
    pub fn push(&mut self, item: T) -> bool {
        if self.size >= self.capacity {
            return false; // Buffer full
        }
        
        self.buffer[self.tail] = Some(item);
        self.tail = (self.tail + 1) % self.capacity;
        self.size += 1;
        true
    }
    
    pub fn pop(&mut self) -> Option<T> {
        if self.size == 0 {
            return None;
        }
        
        let item = self.buffer[self.head].take();
        self.head = (self.head + 1) % self.capacity;
        self.size -= 1;
        item
    }
    
    pub fn is_empty(&self) -> bool {
        self.size == 0
    }
    
    pub fn is_full(&self) -> bool {
        self.size == self.capacity
    }
    
    pub fn len(&self) -> usize {
        self.size
    }
    
    pub fn capacity(&self) -> usize {
        self.capacity
    }
}
```

## Testing Strategies

### Comprehensive Unit Testing

```rust
// tests/unit/user_service_test.rs
use tokio_test;
use uuid::Uuid;
use crate::services::{UserService, CreateUserRequest};
use crate::models::User;

#[tokio::test]
async fn test_create_user() {
    let service = UserService::new();
    let request = CreateUserRequest {
        email: "test@example.com".to_string(),
        name: "Test User".to_string(),
    };
    
    let user = service.create_user(request).await.unwrap();
    
    assert!(!user.id.to_string().is_empty());
    assert_eq!(user.email, "test@example.com");
    assert_eq!(user.name, "Test User");
    assert!(user.created_at <= chrono::Utc::now());
}

#[tokio::test]
async fn test_get_user() {
    let service = UserService::new();
    let request = CreateUserRequest {
        email: "test@example.com".to_string(),
        name: "Test User".to_string(),
    };
    
    let created_user = service.create_user(request).await.unwrap();
    let retrieved_user = service.get_user(created_user.id).await.unwrap();
    
    assert_eq!(created_user.id, retrieved_user.id);
    assert_eq!(created_user.email, retrieved_user.email);
    assert_eq!(created_user.name, retrieved_user.name);
}

#[tokio::test]
async fn test_get_nonexistent_user() {
    let service = UserService::new();
    let nonexistent_id = Uuid::new_v4();
    
    let user = service.get_user(nonexistent_id).await;
    assert!(user.is_none());
}

#[tokio::test]
async fn test_create_user_duplicate_email() {
    let service = UserService::new();
    let request = CreateUserRequest {
        email: "duplicate@example.com".to_string(),
        name: "User 1".to_string(),
    };
    
    // Create first user
    service.create_user(request.clone()).await.unwrap();
    
    // Try to create second user with same email
    let result = service.create_user(request).await;
    assert!(result.is_err());
}

#[tokio::test]
async fn test_user_events() {
    let service = UserService::new();
    let mut event_rx = service.subscribe_events();
    
    let request = CreateUserRequest {
        email: "events@example.com".to_string(),
        name: "Event User".to_string(),
    };
    
    let created_user = service.create_user(request).await.unwrap();
    
    // Should receive UserCreated event
    let event = event_rx.recv().await.unwrap();
    match event {
        UserEvent::Created(user) => {
            assert_eq!(user.id, created_user.id);
            assert_eq!(user.email, created_user.email);
        }
        _ => panic!("Expected UserCreated event"),
    }
}

// Property-based testing with proptest
use proptest::prelude::*;

proptest! {
    #[test]
    fn test_user_email_validation(email in "[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}") {
        prop_assert!(is_valid_email(&email));
    }
    
    #[test]
    fn test_ring_buffer(
        items in prop::collection::vec(any::<i32>(), 1..100),
        capacity in 1usize..=100
    ) {
        let mut buffer = RingBuffer::new(capacity);
        
        for &item in &items {
            if !buffer.push(item) {
                break; // Buffer full
            }
        }
        
        let popped_count = std::cmp::min(items.len(), capacity);
        for (i, &expected_item) in items.iter().take(popped_count).enumerate() {
            let popped = buffer.pop().unwrap();
            prop_assert_eq!(popped, expected_item);
        }
        
        prop_assert!(buffer.is_empty());
    }
}

fn is_valid_email(email: &str) -> bool {
    email.contains('@') && email.contains('.')
}
```

### Integration Testing with TestContainers

```rust
// tests/integration/api_test.rs
use axum_test::TestServer;
use serde_json::json;
use sqlx::PgPool;
use std::sync::Arc;
use tempfile::TempDir;
use tokio::time::{timeout, Duration};

use my_app::{create_app, AppState};
use my_app::services::{UserService, DatabaseUserService};

struct TestApp {
    server: TestServer,
    pool: PgPool,
    _temp_dir: TempDir,
}

async fn setup_test_app() -> TestApp {
    // Create temporary directory for test database
    let temp_dir = TempDir::new().unwrap();
    
    // Setup test database
    let database_url = format!("sqlite:{}", temp_dir.path().join("test.db").display());
    let pool = sqlx::SqlitePool::connect(&database_url).await.unwrap();
    
    // Run migrations
    sqlx::migrate!("./migrations").run(&pool).await.unwrap();
    
    // Create services
    let db_service = DatabaseUserService::new(pool.clone());
    let user_service = Arc::new(UserService::new(Box::new(db_service)));
    
    let app_state = AppState { user_service };
    let app = create_app(app_state);
    
    let server = TestServer::new(app).unwrap();
    
    TestApp {
        server,
        pool,
        _temp_dir: temp_dir,
    }
}

#[tokio::test]
async fn test_user_lifecycle() {
    let test_app = setup_test_app().await;
    
    // Create user
    let create_response = test_app
        .server
        .post("/users")
        .json(&json!({
            "email": "integration@example.com",
            "name": "Integration User"
        }))
        .await;
    
    create_response.assert_status_ok();
    
    let user: User = create_response.json();
    assert_eq!(user.email, "integration@example.com");
    assert_eq!(user.name, "Integration User");
    
    // Get user
    let get_response = test_app
        .server
        .get(&format!("/users/{}", user.id))
        .await;
    
    get_response.assert_status_ok();
    
    let retrieved_user: User = get_response.json();
    assert_eq!(retrieved_user.id, user.id);
    assert_eq!(retrieved_user.email, user.email);
    
    // List users
    let list_response = test_app.server.get("/users").await;
    list_response.assert_status_ok();
    
    let users: Vec<User> = list_response.json();
    assert_eq!(users.len(), 1);
    assert_eq!(users[0].id, user.id);
}

#[tokio::test]
async fn test_concurrent_user_creation() {
    let test_app = setup_test_app().await;
    
    let mut handles = vec![];
    
    for i in 0..10 {
        let server = test_app.server.clone();
        let handle = tokio::spawn(async move {
            server
                .post("/users")
                .json(&json!({
                    "email": format!("user{}@example.com", i),
                    "name": format!("User {}", i)
                }))
                .await
        });
        handles.push(handle);
    }
    
    let responses = futures::future::join_all(handles).await;
    
    for response in responses {
        let response = response.unwrap();
        response.assert_status_ok();
    }
    
    // Verify all users were created
    let list_response = test_app.server.get("/users").await;
    list_response.assert_status_ok();
    
    let users: Vec<User> = list_response.json();
    assert_eq!(users.len(), 10);
}

#[tokio::test]
async fn test_request_timeout() {
    let test_app = setup_test_app().await;
    
    // Simulate slow operation
    let response = test_app
        .server
        .post("/slow-endpoint")
        .add_header("x-delay", "5000") // 5 second delay
        .await;
    
    // Should timeout before response
    let result = timeout(Duration::from_secs(1), response).await;
    assert!(result.is_err());
}
```

## Security Best Practices

### Input Validation and Sanitization

```rust
// src/security/validation.rs
use regex::Regex;
use std::collections::HashSet;
use once_cell::sync::Lazy;

static EMAIL_REGEX: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$").unwrap()
});

static PASSWORD_REGEX: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$").unwrap()
});

static SQL_INJECTION_PATTERNS: Lazy<Vec<Regex>> = Lazy::new(|| {
    vec![
        Regex::new(r"(?i)(union|select|insert|update|delete|drop|create|alter)").unwrap(),
        Regex::new(r"(?i)(script|javascript|vbscript|onload|onerror)").unwrap(),
        Regex::new(r"(--|;|\/\*|\*\/)").unwrap(),
    ]
});

pub struct Validator;

impl Validator {
    pub fn validate_email(email: &str) -> Result<(), ValidationError> {
        if email.len() > 254 {
            return Err(ValidationError::EmailTooLong);
        }
        
        if !EMAIL_REGEX.is_match(email) {
            return Err(ValidationError::InvalidEmailFormat);
        }
        
        Ok(())
    }
    
    pub fn validate_password(password: &str) -> Result<(), ValidationError> {
        if password.len() < 8 {
            return Err(ValidationError::PasswordTooShort);
        }
        
        if password.len() > 128 {
            return Err(ValidationError::PasswordTooLong);
        }
        
        if !PASSWORD_REGEX.is_match(password) {
            return Err(ValidationError::WeakPassword);
        }
        
        // Check for common weak passwords
        if Self::is_common_password(password) {
            return Err(ValidationError::CommonPassword);
        }
        
        Ok(())
    }
    
    pub fn sanitize_input(input: &str) -> String {
        let mut sanitized = String::with_capacity(input.len());
        
        for ch in input.chars() {
            match ch {
                // Allow safe characters
                'a'..='z' | 'A'..='Z' | '0'..='9' | ' ' | '-' | '_' | '.' | '@' => {
                    sanitized.push(ch);
                }
                // Replace dangerous characters
                '<' => sanitized.push_str("&lt;"),
                '>' => sanitized.push_str("&gt;"),
                '&' => sanitized.push_str("&amp;"),
                '"' => sanitized.push_str("&quot;"),
                '\'' => sanitized.push_str("&#x27;"),
                _ => {
                    // Skip other characters
                }
            }
        }
        
        sanitized.trim().to_string()
    }
    
    pub fn detect_sql_injection(input: &str) -> bool {
        let lowercase_input = input.to_lowercase();
        
        for pattern in SQL_INJECTION_PATTERNS.iter() {
            if pattern.is_match(&lowercase_input) {
                return true;
            }
        }
        
        false
    }
    
    fn is_common_password(password: &str) -> bool {
        static COMMON_PASSWORDS: Lazy<HashSet<&'static str>> = Lazy::new(|| {
            vec![
                "password", "123456", "123456789", "12345678", "12345",
                "1234567", "1234567890", "qwerty", "abc123", "password123",
            ].into_iter().collect()
        });
        
        COMMON_PASSWORDS.contains(password.to_lowercase().as_str())
    }
}

#[derive(Debug, thiserror::Error)]
pub enum ValidationError {
    #[error("Email is too long (max 254 characters)")]
    EmailTooLong,
    
    #[error("Invalid email format")]
    InvalidEmailFormat,
    
    #[error("Password is too short (min 8 characters)")]
    PasswordTooShort,
    
    #[error("Password is too long (max 128 characters)")]
    PasswordTooLong,
    
    #[error("Password is too weak")]
    WeakPassword,
    
    #[error("Password is too common")]
    CommonPassword,
}

// Rate limiting
use std::collections::HashMap;
use std::time::{Duration, Instant};
use tokio::sync::RwLock;

pub struct RateLimiter {
    requests: Arc<RwLock<HashMap<String, Vec<Instant>>>>,
    max_requests: usize,
    window: Duration,
}

impl RateLimiter {
    pub fn new(max_requests: usize, window: Duration) -> Self {
        Self {
            requests: Arc::new(RwLock::new(HashMap::new())),
            max_requests,
            window,
        }
    }
    
    pub async fn is_allowed(&self, key: &str) -> bool {
        let now = Instant::now();
        let mut requests = self.requests.write().await;
        
        let user_requests = requests.entry(key.to_string()).or_insert_with(Vec::new);
        
        // Remove old requests outside the window
        user_requests.retain(|&timestamp| now.duration_since(timestamp) < self.window);
        
        if user_requests.len() >= self.max_requests {
            return false;
        }
        
        user_requests.push(now);
        true
    }
}
```

### Authentication and Authorization

```rust
// src/security/auth.rs
use argon2::{password_hash::{rand_core::OsRng, PasswordHash, PasswordHasher, PasswordVerifier, SaltString}, Argon2};
use jsonwebtoken::{decode, encode, DecodingKey, EncodingKey, Header, Validation};
use serde::{Deserialize, Serialize};
use std::time::{Duration, SystemTime, UNIX_EPOCH};
use uuid::Uuid;

#[derive(Debug, Serialize, Deserialize)]
pub struct Claims {
    pub sub: String, // Subject (user ID)
    pub email: String,
    pub role: String,
    pub exp: usize, // Expiration time
    pub iat: usize, // Issued at
}

pub struct AuthService {
    jwt_secret: String,
    token_expiry: Duration,
}

impl AuthService {
    pub fn new(jwt_secret: String, token_expiry: Duration) -> Self {
        Self {
            jwt_secret,
            token_expiry,
        }
    }
    
    pub fn hash_password(password: &str) -> Result<String, AuthError> {
        let salt = SaltString::generate(&mut OsRng);
        let argon2 = Argon2::default();
        
        let password_hash = argon2
            .hash_password(password.as_bytes(), &salt)
            .map_err(|e| AuthError::PasswordHashing(e.to_string()))?;
        
        Ok(password_hash.to_string())
    }
    
    pub fn verify_password(password: &str, hash: &str) -> Result<bool, AuthError> {
        let parsed_hash = PasswordHash::new(hash)
            .map_err(|e| AuthError::InvalidHash(e.to_string()))?;
        
        let argon2 = Argon2::default();
        
        match argon2.verify_password(password.as_bytes(), &parsed_hash) {
            Ok(_) => Ok(true),
            Err(argon2::password_hash::Error::Password) => Ok(false),
            Err(e) => Err(AuthError::PasswordHashing(e.to_string())),
        }
    }
    
    pub fn generate_token(&self, user_id: &str, email: &str, role: &str) -> Result<String, AuthError> {
        let now = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .map_err(|e| AuthError::TokenGeneration(e.to_string()))?;
        
        let claims = Claims {
            sub: user_id.to_string(),
            email: email.to_string(),
            role: role.to_string(),
            exp: (now + self.token_expiry).as_secs() as usize,
            iat: now.as_secs() as usize,
        };
        
        let token = encode(
            &Header::default(),
            &claims,
            &EncodingKey::from_secret(self.jwt_secret.as_bytes()),
        )
        .map_err(|e| AuthError::TokenGeneration(e.to_string()))?;
        
        Ok(token)
    }
    
    pub fn validate_token(&self, token: &str) -> Result<Claims, AuthError> {
        let validation = Validation::default();
        let decoded = decode::<Claims>(
            token,
            &DecodingKey::from_secret(self.jwt_secret.as_bytes()),
            &validation,
        )
        .map_err(|_| AuthError::InvalidToken)?;
        
        Ok(decoded.claims)
    }
    
    pub fn refresh_token(&self, claims: &Claims) -> Result<String, AuthError> {
        self.generate_token(&claims.sub, &claims.email, &claims.role)
    }
}

#[derive(Debug, thiserror::Error)]
pub enum AuthError {
    #[error("Password hashing failed: {0}")]
    PasswordHashing(String),
    
    #[error("Invalid password hash: {0}")]
    InvalidHash(String),
    
    #[error("Token generation failed: {0}")]
    TokenGeneration(String),
    
    #[error("Invalid token")]
    InvalidToken,
    
    #[error("Token expired")]
    TokenExpired,
}

// Authorization middleware for axum
use axum::{
    extract::{Request, State},
    http::{header::AUTHORIZATION, StatusCode},
    middleware::Next,
    response::Response,
};
use std::sync::Arc;

pub async fn auth_middleware(
    State(auth_service): State<Arc<AuthService>>,
    mut request: Request,
    next: Next,
) -> Result<Response, StatusCode> {
    let auth_header = request
        .headers()
        .get(AUTHORIZATION)
        .and_then(|header| header.to_str().ok())
        .and_then(|header| header.strip_prefix("Bearer "));
    
    let token = match auth_header {
        Some(token) => token,
        None => return Err(StatusCode::UNAUTHORIZED),
    };
    
    let claims = match auth_service.validate_token(token) {
        Ok(claims) => claims,
        Err(_) => return Err(StatusCode::UNAUTHORIZED),
    };
    
    // Add claims to request extensions
    request.extensions_mut().insert(claims);
    
    Ok(next.run(request).await)
}

pub fn require_role(required_role: &'static str) -> impl Fn(Request, Next) -> Result<Response, StatusCode> {
    move |request: Request, next: Next| {
        let claims = request.extensions().get::<Claims>();
        
        match claims {
            Some(claims) if claims.role == required_role => Ok(next.run(request)),
            Some(_) => Err(StatusCode::FORBIDDEN),
            None => Err(StatusCode::UNAUTHORIZED),
        }
    }
}
```

## Integration Patterns

### gRPC Service Implementation

```rust
// src/grpc/user_service.rs
use tonic::{transport::Server, Request, Response, Status, Code};
use tonic_health::server::HealthReporter;

pub mod user {
    tonic::include_proto!("user");
}

use user::{
    user_service_server::{UserService, UserServiceServer},
    CreateUserRequest, CreateUserResponse, 
    GetUserRequest, GetUserResponse,
    User as GrpcUser,
};

use crate::models::User;
use crate::services::UserService as AppUserService;
use crate::error::AppError;

pub struct GrpcUserService {
    app_service: Arc<AppUserService>,
}

impl GrpcUserService {
    pub fn new(app_service: Arc<AppUserService>) -> Self {
        Self { app_service }
    }
}

#[tonic::async_trait]
impl UserService for GrpcUserService {
    async fn create_user(
        &self,
        request: Request<CreateUserRequest>,
    ) -> Result<Response<CreateUserResponse>, Status> {
        let req = request.into_inner();
        
        let create_req = crate::services::CreateUserRequest {
            email: req.email,
            name: req.name,
        };
        
        match self.app_service.create_user(create_req).await {
            Ok(user) => {
                let grpc_user = GrpcUser {
                    id: user.id.to_string(),
                    email: user.email,
                    name: user.name,
                    created_at: Some(user.created_at.into()),
                };
                
                let response = CreateUserResponse { user: Some(grpc_user) };
                Ok(Response::new(response))
            }
            Err(e) => {
                let status = match e {
                    AppError::Validation { .. } => Status::invalid_argument(e.to_string()),
                    AppError::DuplicateEmail => Status::already_exists(e.to_string()),
                    _ => Status::internal(e.to_string()),
                };
                Err(status)
            }
        }
    }
    
    async fn get_user(
        &self,
        request: Request<GetUserRequest>,
    ) -> Result<Response<GetUserResponse>, Status> {
        let req = request.into_inner();
        
        let user_id = Uuid::parse_str(&req.id)
            .map_err(|_| Status::invalid_argument("Invalid user ID"))?;
        
        match self.app_service.get_user(user_id).await {
            Some(user) => {
                let grpc_user = GrpcUser {
                    id: user.id.to_string(),
                    email: user.email,
                    name: user.name,
                    created_at: Some(user.created_at.into()),
                };
                
                let response = GetUserResponse { user: Some(grpc_user) };
                Ok(Response::new(response))
            }
            None => Err(Status::not_found("User not found")),
        }
    }
}

// Server setup
pub async fn run_grpc_server(
    addr: std::net::SocketAddr,
    user_service: Arc<AppUserService>,
) -> Result<(), Box<dyn std::error::Error>> {
    let grpc_service = GrpcUserService::new(user_service);
    
    let (mut health_reporter, health_service) = tonic_health::server::health_reporter();
    health_reporter.set_serving::<UserServiceServer<GrpcUserService>>().await;
    
    Server::builder()
        .add_service(health_service)
        .add_service(UserServiceServer::new(grpc_service))
        .serve(addr)
        .await?;
    
    Ok(())
}
```

### Message Queue Integration

```rust
// src/messaging/kafka_producer.rs
use rdkafka::{
    config::ClientConfig,
    producer::{FutureProducer, FutureRecord},
    Client as KafkaClient,
};
use serde::{Deserialize, Serialize};
use std::time::Duration;
use tokio::time::timeout;

#[derive(Debug, Serialize, Deserialize)]
pub struct Event {
    pub id: String,
    pub event_type: String,
    pub data: serde_json::Value,
    pub timestamp: chrono::DateTime<chrono::Utc>,
    pub version: u32,
}

pub struct EventProducer {
    producer: FutureProducer,
}

impl EventProducer {
    pub fn new(brokers: &str) -> Result<Self, Box<dyn std::error::Error>> {
        let producer: FutureProducer = ClientConfig::new()
            .set("bootstrap.servers", brokers)
            .set("message.timeout.ms", "5000")
            .set("delivery.timeout.ms", "10000")
            .set("request.timeout.ms", "5000")
            .create()?;
        
        Ok(Self { producer })
    }
    
    pub async fn publish_event(&self, topic: &str, event: Event) -> Result<(), Box<dyn std::error::Error>> {
        let event_json = serde_json::to_vec(&event)?;
        let key = Some(event.id.as_bytes());
        
        let record = FutureRecord::to(topic)
            .key(key)
            .payload(&event_json)
            .headers(
                rdkafka::message::Headers::new()
                    .add("event_type", &event.event_type)
                    .add("version", &event.version.to_string())
            );
        
        let delivery_future = self.producer.send(record, Duration::from_secs(0));
        
        match timeout(Duration::from_secs(5), delivery_future).await {
            Ok(Ok((partition, offset))) => {
                tracing::info!(
                    "Event published to partition {} at offset {}",
                    partition,
                    offset
                );
                Ok(())
            }
            Ok(Err((kafka_error, _))) => {
                tracing::error!("Failed to publish event: {}", kafka_error);
                Err(Box::new(kafka_error))
            }
            Err(_) => {
                tracing::error!("Timeout while publishing event");
                Err("Timeout while publishing event".into())
            }
        }
    }
}

// Event consumer
use rdkafka::{
    config::ConsumerConfig,
    consumer::{StreamConsumer, Consumer},
    message::BorrowedMessage,
};

pub struct EventConsumer {
    consumer: StreamConsumer,
}

impl EventConsumer {
    pub fn new(brokers: &str, group_id: &str) -> Result<Self, Box<dyn std::error::Error>> {
        let consumer: StreamConsumer = ClientConfig::new()
            .set("group.id", group_id)
            .set("bootstrap.servers", brokers)
            .set("enable.auto.commit", "false")
            .set("auto.offset.reset", "earliest")
            .create()?;
        
        Ok(Self { consumer })
    }
    
    pub async fn subscribe(&self, topics: &[&str]) -> Result<(), Box<dyn std::error::Error>> {
        self.consumer.subscribe(topics)?;
        Ok(())
    }
    
    pub async fn start_consuming<F, Fut>(&mut self, handler: F) 
    where
        F: Fn(Event) -> Fut + Send + Sync + 'static,
        Fut: std::future::Future<Output = Result<(), Box<dyn std::error::Error>>> + Send,
    {
        while let Some(message_result = self.consumer.recv().await {
            match message_result {
                Ok(message) => {
                    match self.handle_message(&message, &handler).await {
                        Ok(_) => {
                            // Commit message offset
                            let _ = self.consumer.commit_message(&message, rdkafka::consumer::CommitMode::Async).await;
                        }
                        Err(e) => {
                            tracing::error!("Error handling message: {}", e);
                        }
                    }
                }
                Err(e) => {
                    tracing::error!("Error receiving message: {}", e);
                }
            }
        }
    }
    
    async fn handle_message<F, Fut>(
        &self,
        message: &BorrowedMessage<'_>,
        handler: &F,
    ) -> Result<(), Box<dyn std::error::Error>>
    where
        F: Fn(Event) -> Fut,
        Fut: std::future::Future<Output = Result<(), Box<dyn std::error::Error>>>,
    {
        let payload = message.payload().ok_or("Empty message payload")?;
        
        let event: Event = serde_json::from_slice(payload)?;
        
        handler(event).await
    }
}
```

## Modern Development Workflow

### Configuration Management

```rust
// src/config/mod.rs
use serde::{Deserialize, Serialize};
use std::path::Path;

#[derive(Debug, Clone, Deserialize)]
pub struct Config {
    pub server: ServerConfig,
    pub database: DatabaseConfig,
    pub auth: AuthConfig,
    pub kafka: KafkaConfig,
    pub logging: LoggingConfig,
}

#[derive(Debug, Clone, Deserialize)]
pub struct ServerConfig {
    pub host: String,
    pub port: u16,
    pub workers: usize,
}

#[derive(Debug, Clone, Deserialize)]
pub struct DatabaseConfig {
    pub url: String,
    pub max_connections: u32,
    pub min_connections: u32,
    pub connect_timeout: u64,
    pub acquire_timeout: u64,
}

#[derive(Debug, Clone, Deserialize)]
pub struct AuthConfig {
    pub jwt_secret: String,
    pub token_expiry_hours: u64,
    pub bcrypt_cost: u32,
}

#[derive(Debug, Clone, Deserialize)]
pub struct KafkaConfig {
    pub brokers: Vec<String>,
    pub group_id: String,
    pub topic_prefix: String,
}

#[derive(Debug, Clone, Deserialize)]
pub struct LoggingConfig {
    pub level: String,
    pub format: String,
    pub output: String,
}

impl Config {
    pub fn load() -> Result<Self, Box<dyn std::error::Error>> {
        // Try environment variable first
        if let Ok(config_path) = std::env::var("CONFIG_PATH") {
            return Self::load_from_file(&config_path);
        }
        
        // Try default locations
        let default_paths = vec![
            "config.toml",
            "config.yaml",
            "/etc/my-app/config.toml",
            "/etc/my-app/config.yaml",
        ];
        
        for path in default_paths {
            if Path::new(path).exists() {
                return Self::load_from_file(path);
            }
        }
        
        // Fallback to environment variables
        Self::load_from_env()
    }
    
    fn load_from_file(path: &str) -> Result<Self, Box<dyn std::error::Error>> {
        let content = std::fs::read_to_string(path)?;
        
        if path.ends_with(".toml") {
            let config: Config = toml::from_str(&content)?;
            Ok(config)
        } else if path.ends_with(".yaml") || path.ends_with(".yml") {
            let config: Config = serde_yaml::from_str(&content)?;
            Ok(config)
        } else {
            Err("Unsupported config file format".into())
        }
    }
    
    fn load_from_env() -> Result<Self, Box<dyn std::error::Error>> {
        Ok(Config {
            server: ServerConfig {
                host: std::env::var("SERVER_HOST").unwrap_or_else(|_| "0.0.0.0".to_string()),
                port: std::env::var("SERVER_PORT")
                    .unwrap_or_else(|_| "8080".to_string())
                    .parse()?,
                workers: std::env::var("SERVER_WORKERS")
                    .unwrap_or_else(|_| "4".to_string())
                    .parse()?,
            },
            database: DatabaseConfig {
                url: std::env::var("DATABASE_URL")?,
                max_connections: std::env::var("DATABASE_MAX_CONNECTIONS")
                    .unwrap_or_else(|_| "10".to_string())
                    .parse()?,
                min_connections: std::env::var("DATABASE_MIN_CONNECTIONS")
                    .unwrap_or_else(|_| "1".to_string())
                    .parse()?,
                connect_timeout: std::env::var("DATABASE_CONNECT_TIMEOUT")
                    .unwrap_or_else(|_| "30".to_string())
                    .parse()?,
                acquire_timeout: std::env::var("DATABASE_ACQUIRE_TIMEOUT")
                    .unwrap_or_else(|_| "30".to_string())
                    .parse()?,
            },
            auth: AuthConfig {
                jwt_secret: std::env::var("JWT_SECRET")?,
                token_expiry_hours: std::env::var("TOKEN_EXPIRY_HOURS")
                    .unwrap_or_else(|_| "24".to_string())
                    .parse()?,
                bcrypt_cost: std::env::var("BCRYPT_COST")
                    .unwrap_or_else(|_| "12".to_string())
                    .parse()?,
            },
            kafka: KafkaConfig {
                brokers: std::env::var("KAFKA_BROKERS")
                    .unwrap_or_else(|_| "localhost:9092".to_string())
                    .split(',')
                    .map(|s| s.trim().to_string())
                    .collect(),
                group_id: std::env::var("KAFKA_GROUP_ID")
                    .unwrap_or_else(|_| "my-app".to_string()),
                topic_prefix: std::env::var("KAFKA_TOPIC_PREFIX")
                    .unwrap_or_else(|_| "my-app".to_string()),
            },
            logging: LoggingConfig {
                level: std::env::var("LOG_LEVEL").unwrap_or_else(|_| "info".to_string()),
                format: std::env::var("LOG_FORMAT").unwrap_or_else(|_| "json".to_string()),
                output: std::env::var("LOG_OUTPUT").unwrap_or_else(|_| "stdout".to_string()),
            },
        })
    }
}
```

### Cargo Workspace Configuration

```toml
# Cargo.toml (workspace root)
[workspace]
members = [
    "crates/common",
    "crates/server", 
    "crates/client",
    "crates/cli",
]

[workspace.package]
version = "0.1.0"
edition = "2021"
authors = ["Your Name <your.email@example.com>"]
license = "MIT OR Apache-2.0"
repository = "https://github.com/yourorg/your-project"

[workspace.dependencies]
# Async runtime
tokio = { version = "1.42", features = ["full"] }
tokio-util = "0.7"

# Web framework
axum = "0.8"
tower = "0.5"
tower-http = "0.6"

# Database
sqlx = { version = "0.8", features = ["runtime-tokio-rustls", "postgres", "chrono", "uuid"] }
diesel = { version = "2.2", features = ["postgres", "uuid", "chrono"] }

# Serialization
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"

# Error handling
thiserror = "1.0"
anyhow = "1.0"

# Logging
tracing = "0.1"
tracing-subscriber = { version = "0.3", features = ["json"] }

# Security
argon2 = "0.5"
jsonwebtoken = "9.3"
uuid = { version = "1.11", features = ["v4", "serde"] }

# Configuration
config = "0.14"
serde_yaml = "0.9"

# Testing
tokio-test = "0.4"
mockall = "0.13"

[profile.release]
lto = true
codegen-units = 1
panic = "abort"

[profile.dev]
debug = true
```

### GitHub Actions CI/CD

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  CARGO_TERM_COLOR: always

jobs:
  test:
    name: Test Suite
    runs-on: ubuntu-latest
    
    strategy:
      matrix:
        rust:
          - stable
          - beta
          - nightly
        features:
          - ""
          - "kafka,database"
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Install Rust
      uses: dtolnay/rust-toolchain@master
      with:
        toolchain: ${{ matrix.rust }}
        components: rustfmt, clippy
        
    - name: Cache dependencies
      uses: actions/cache@v4
      with:
        path: |
          ~/.cargo/registry
          ~/.cargo/git
          target
        key: ${{ runner.os }}-cargo-${{ hashFiles('**/Cargo.lock') }}
        
    - name: Install system dependencies
      run: |
        sudo apt-get update
        sudo apt-get install -y postgresql-client libpq-dev
        
    - name: Check formatting
      run: cargo fmt --all -- --check
      
    - name: Run clippy
      run: cargo clippy --all-targets --all-features -- -D warnings
      
    - name: Run tests
      run: cargo test --workspace --features ${{ matrix.features }}
      
    - name: Run doc tests
      run: cargo test --doc --features ${{ matrix.features }}

  coverage:
    name: Code Coverage
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Install Rust
      uses: dtolnay/rust-toolchain@stable
      with:
        components: llvm-tools-preview
        
    - name: Install cargo-llvm-cov
      run: cargo install cargo-llvm-cov
      
    - name: Install system dependencies
      run: |
        sudo apt-get update
        sudo apt-get install -y postgresql-client libpq-dev
        
    - name: Generate code coverage
      run: cargo llvm-cov --workspace --lcov --output-path lcov.info
      
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v4
      with:
        files: lcov.info
        fail_ci_if_error: true

  security:
    name: Security Audit
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Install Rust
      uses: dtolnay/rust-toolchain@stable
      
    - name: Install cargo-audit
      run: cargo install cargo-audit
      
    - name: Security audit
      run: cargo audit
      
    - name: Install cargo-deny
      run: cargo install cargo-deny
      
    - name: Check dependencies
      run: cargo deny check

  build:
    name: Build Release
    runs-on: ubuntu-latest
    
    needs: [test, security]
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Install Rust
      uses: dtolnay/rust-toolchain@stable
      
    - name: Cache dependencies
      uses: actions/cache@v4
      with:
        path: |
          ~/.cargo/registry
          ~/.cargo/git
          target
        key: ${{ runner.os }}-cargo-release-${{ hashFiles('**/Cargo.lock') }}
        
    - name: Build release
      run: cargo build --workspace --release
      
    - name: Upload binary artifacts
      uses: actions/upload-artifact@v4
      with:
        name: release-binaries
        path: target/release/
```

---

**Created by**: MoAI Language Skill Factory  
**Last Updated**: 2025-11-06  
**Version**: 2.0.0  
**Rust Target**: 1.91+ with latest language features  

This skill provides comprehensive Rust development guidance with 2025 best practices, covering everything from basic memory safety to advanced systems programming and performance-critical applications.
