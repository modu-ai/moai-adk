# Rust Advanced Patterns — Traits, Async, Macros

_Last updated: 2025-11-22_

## Traits & Polymorphism

### Trait Objects vs Generics

```rust
// ✅ Static dispatch (generic) - monomorphization
fn process_generic<T: Iterator>(iter: T) {
    for item in iter {
        println!("{:?}", item);
    }
}

// ❌ Dynamic dispatch (trait object) - runtime overhead
fn process_dynamic(iter: &mut dyn Iterator<Item=i32>) {
    // Requires static lifetime in many cases
}

// Trait with associated types
trait Database {
    type Connection;
    type Error;

    fn connect(&self, url: &str) -> Result<Self::Connection, Self::Error>;
}

struct PostgreSQL;

impl Database for PostgreSQL {
    type Connection = PgConnection;
    type Error = SqlError;

    fn connect(&self, url: &str) -> Result<PgConnection, SqlError> {
        // Implementation
    }
}

// Usage with type inference
fn get_connection<T: Database>(db: &T) -> Result<T::Connection, T::Error> {
    db.connect("postgresql://localhost/db")
}
```

## Async/Tokio Patterns

### Tokio Task Management

```rust
use tokio::task;
use tokio::sync::Semaphore;
use std::sync::Arc;

// Spawn multiple tasks with semaphore for concurrency limit
async fn fetch_with_limit(urls: Vec<&str>) -> Vec<String> {
    let semaphore = Arc::new(Semaphore::new(3));  // Max 3 concurrent
    let mut handles = vec![];

    for url in urls {
        let permit = semaphore.acquire_owned().await.unwrap();
        let handle = task::spawn(async move {
            let _guard = permit;  // Hold permit while fetching
            fetch_url(url).await
        });
        handles.push(handle);
    }

    let mut results = vec![];
    for handle in handles {
        if let Ok(result) = handle.await {
            results.push(result);
        }
    }
    results
}

// Select multiple futures with first-to-complete
async fn race_requests(url1: &str, url2: &str) -> String {
    tokio::select! {
        result1 = fetch_url(url1) => {
            println!("URL1 won!");
            result1
        }
        result2 = fetch_url(url2) => {
            println!("URL2 won!");
            result2
        }
    }
}

// Timeout handling
async fn fetch_with_timeout(url: &str) -> Result<String, Box<dyn std::error::Error>> {
    tokio::time::timeout(
        tokio::time::Duration::from_secs(5),
        fetch_url(url)
    )
    .await?
    .map_err(|e| Box::new(e) as Box<dyn std::error::Error>)
}
```

### Channels for Message Passing

```rust
use tokio::sync::mpsc;

async fn producer_consumer() {
    let (tx, mut rx) = mpsc::channel(10);

    // Producer task
    tokio::spawn(async move {
        for i in 0..5 {
            tx.send(format!("Item {}", i)).await.ok();
        }
    });

    // Consumer
    while let Some(item) = rx.recv().await {
        println!("Received: {}", item);
    }
}

// Broadcast channel (one-to-many)
async fn broadcast_example() {
    let (tx, _rx) = tokio::sync::broadcast::channel(16);

    let tx1 = tx.clone();
    tokio::spawn(async move {
        tx1.send("Message 1").ok();
    });

    let tx2 = tx.clone();
    tokio::spawn(async move {
        tx2.send("Message 2").ok();
    });
}
```

## Macro System

### Derive Macros

```rust
// Using derive attribute from serde
#[derive(serde::Serialize, serde::Deserialize, Debug)]
struct User {
    id: u64,
    name: String,
    email: String,
}

// Custom derive macro (procedural)
#[derive(Builder)]
struct Config {
    name: String,
    #[builder(required = true)]
    host: String,
    port: Option<u16>,
}

// Usage
let config = ConfigBuilder::default()
    .host("localhost")
    .port(Some(8080))
    .build()
    .unwrap();
```

### Declarative Macros

```rust
// DSL for building query conditions
macro_rules! query {
    ($($col:ident == $val:expr),+ $(,)?) => {
        vec![
            $(stringify!($col).to_string() + "='" + $val + "'"),+
        ].join(" AND ")
    };
}

let where_clause = query!(
    age == "30",
    status == "active"
);
// Expands to: "age='30' AND status='active'"
```

## Concurrency Primitives

### Mutexes & RwLocks

```rust
use tokio::sync::Mutex;
use std::sync::RwLock;

// Async mutex (preferred for async code)
async fn async_mutex_example() {
    let data = Mutex::new(vec![1, 2, 3]);
    let mut locked = data.lock().await;
    locked.push(4);
}

// Sync RwLock for read-heavy scenarios
fn rwlock_example() {
    let data = RwLock::new(vec![1, 2, 3]);

    // Multiple readers
    let r1 = data.read().unwrap();
    let r2 = data.read().unwrap();
    println!("{:?} {:?}", *r1, *r2);

    // Single writer (blocks readers)
    let mut w = data.write().unwrap();
    w.push(4);
}

// Once for one-time initialization
use std::sync::Once;

static INIT: Once = Once::new();
static mut CONFIG: Option<String> = None;

fn initialize_config() {
    INIT.call_once(|| unsafe {
        CONFIG = Some(String::from("initialized"));
    });
}

// Atomic operations for lock-free coordination
use std::sync::atomic::{AtomicUsize, Ordering};
use std::sync::Arc;

async fn atomic_counter() {
    let count = Arc::new(AtomicUsize::new(0));
    let mut handles = vec![];

    for _ in 0..10 {
        let count_clone = count.clone();
        handles.push(tokio::spawn(async move {
            count_clone.fetch_add(1, Ordering::Relaxed);
        }));
    }

    for handle in handles {
        handle.await.ok();
    }

    println!("Count: {}", count.load(Ordering::Relaxed));
}
```

## Error Handling Patterns

### Custom Error Types

```rust
use std::fmt;
use std::error::Error;

#[derive(Debug)]
enum ApiError {
    NetworkError(String),
    ParseError(String),
    NotFound(u64),
    ServerError { code: u16, message: String },
}

impl fmt::Display for ApiError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            ApiError::NetworkError(msg) => write!(f, "Network error: {}", msg),
            ApiError::ParseError(msg) => write!(f, "Parse error: {}", msg),
            ApiError::NotFound(id) => write!(f, "Not found: {}", id),
            ApiError::ServerError { code, message } => {
                write!(f, "Server error {}: {}", code, message)
            }
        }
    }
}

impl Error for ApiError {}

// Error conversion for ? operator
impl From<std::io::Error> for ApiError {
    fn from(err: std::io::Error) -> Self {
        ApiError::NetworkError(err.to_string())
    }
}

// Usage
fn fetch_data(id: u64) -> Result<String, ApiError> {
    if id == 0 {
        return Err(ApiError::NotFound(id));
    }
    Ok(format!("Data for {}", id))
}

async fn process() -> Result<(), ApiError> {
    let data = fetch_data(1)?;  // ? operator with custom error
    Ok(())
}
```

## Unsafe Code Patterns

### Foreign Function Interface

```rust
extern "C" {
    fn malloc(size: usize) -> *mut u8;
    fn free(ptr: *mut u8);
}

unsafe fn allocate_c_buffer(size: usize) -> *mut u8 {
    malloc(size)
}

unsafe fn deallocate(ptr: *mut u8) {
    free(ptr);
}

// Safer wrapper
struct CBuffer {
    ptr: *mut u8,
    size: usize,
}

impl Drop for CBuffer {
    fn drop(&mut self) {
        unsafe { free(self.ptr) };
    }
}

impl CBuffer {
    fn new(size: usize) -> Self {
        unsafe {
            CBuffer {
                ptr: malloc(size),
                size,
            }
        }
    }
}
```

---

**Last Updated**: 2025-11-22
**Related**: moai-lang-rust/SKILL.md, modules/optimization.md

