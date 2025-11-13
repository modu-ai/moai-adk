---
name: "moai-lang-rust"
description: "Enterprise Rust with ownership model and safety guarantees: Rust 1.91.1, Tokio 1.48, async/await, macro system, error handling, memory safety patterns; activates for systems programming, performance-critical code, concurrent applications, and safety-first development."
allowed-tools: 
version: "4.0.0"
status: stable
---

# Rust Systems Programming — Enterprise v4.0

## Technology Stack (November 2025 Stable)

### Core Language
- **Rust 1.91.1** (Latest stable, November 2025)
  - Ownership and borrowing system
  - Zero-cost abstractions
  - Memory safety without GC
  - Performance optimization

### Async Runtime
- **Tokio 1.48.x** (Production async runtime)
  - Async I/O
  - Task scheduling
  - Synchronization primitives
  - Macro utilities

- **async-std 1.13.x** (Alternative runtime)
  - Compatible API
  - Task-based execution

### Web & Network
- **Axum 0.8.x** (Web framework)
  - Composable handlers
  - Router support
  - Type-safe extractors

- **Rocket 0.5.x** (Developer-friendly framework)
  - Macro-driven API
  - Type-safe routing

- **Warp 0.3.x** (Filter-based framework)
  - Composable filters
  - High performance

### Serialization
- **serde 1.0.x** (Serialization framework)
  - Derive macros
  - Custom implementations
  - Format support

- **serde_json 1.0.x** (JSON support)

### Macros & Code Generation
- **proc-macro 1.1.x** (Procedural macros)
- **syn 2.x** (Parser for Rust code)
- **quote 1.x** (Code generation)

### Testing & Profiling
- **cargo test** (Built-in testing)
- **proptest 1.5.x** (Property-based testing)
- **criterion 0.5.x** (Benchmarking)

---

## Level 1: Quick Reference

### Rust 1.91 Ownership System

**Ownership Basics**:
```rust
fn main() {
    let s1 = String::from("hello");
    let s2 = s1; // Move: s1 no longer valid
    
    // println!("{}", s1); // Compile error!
    println!("{}", s2); // OK
    
    let s3 = String::from("world");
    let s4 = &s3; // Borrow: s3 still valid
    let s5 = &s3; // Multiple immutable borrows OK
    
    println!("{} {}", s4, s5);
    println!("{}", s3); // Still valid
}
```

**Mutable References**:
```rust
fn change_string(s: &mut String) {
    s.push_str(" world");
}

fn main() {
    let mut s = String::from("hello");
    change_string(&mut s);
    println!("{}", s); // "hello world"
    
    // Can't have immutable references while mutable borrow exists
    let r1 = &mut s;
    // let r2 = &s; // Compile error!
    r1.push_str("!");
    println!("{}", r1);
}
```

**Lifetimes**:
```rust
fn longest<'a>(x: &'a str, y: &'a str) -> &'a str {
    if x.len() > y.len() {
        x
    } else {
        y
    }
}

fn main() {
    let s1 = String::from("hello");
    let s2 = "world";
    let result = longest(&s1, s2);
    println!("{}", result);
}
```

### Tokio Async Runtime

**Basic Async Tasks**:
```rust
use tokio::task;
use tokio::time::{sleep, Duration};

#[tokio::main]
async fn main() {
    let handle = task::spawn(async {
        sleep(Duration::from_secs(1)).await;
        println!("Task completed!");
    });
    
    // Wait for task
    handle.await.unwrap();
}
```

**Error Handling**:
```rust
use std::fs;
use std::io;

fn read_file(path: &str) -> Result<String, io::Error> {
    fs::read_to_string(path)
}

fn main() {
    match read_file("data.txt") {
        Ok(contents) => println!("{}", contents),
        Err(e) => eprintln!("Error: {}", e),
    }
    
    // Shorthand
    let result = read_file("data.txt").expect("Failed to read");
}
```

---

## Level 2: Core Implementation

### Custom Error Types

```rust
use std::fmt;

#[derive(Debug)]
enum ParseError {
    InvalidFormat,
    OutOfRange,
}

impl fmt::Display for ParseError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            ParseError::InvalidFormat => write!(f, "Invalid format"),
            ParseError::OutOfRange => write!(f, "Out of range"),
        }
    }
}

impl std::error::Error for ParseError {}

fn parse_number(s: &str) -> Result<i32, ParseError> {
    let num = s.parse::<i32>()
        .map_err(|_| ParseError::InvalidFormat)?;
    
    if num < 0 || num > 100 {
        return Err(ParseError::OutOfRange);
    }
    
    Ok(num)
}
```

### Procedural Macros

**Custom Derive Macro**:
```rust
// Cargo.toml
[lib]
proc-macro = true

// lib.rs
use proc_macro::TokenStream;
use quote::quote;
use syn::{parse_macro_input, DeriveInput};

#[proc_macro_derive(MyDerive)]
pub fn my_derive(input: TokenStream) -> TokenStream {
    let input = parse_macro_input!(input as DeriveInput);
    let name = &input.ident;
    
    let expanded = quote! {
        impl #name {
            fn describe() -> &'static str {
                stringify!(#name)
            }
        }
    };
    
    TokenStream::from(expanded)
}

// Usage
#[derive(MyDerive)]
struct MyStruct;

fn main() {
    println!("{}", MyStruct::describe());
}
```

### Async with Tokio Channels

**MPSC Channel**:
```rust
use tokio::sync::mpsc;
use tokio::task;

#[tokio::main]
async fn main() {
    let (tx, mut rx) = mpsc::channel(32);
    
    task::spawn(async move {
        for i in 0..10 {
            tx.send(i).await.ok();
        }
    });
    
    while let Some(value) = rx.recv().await {
        println!("Received: {}", value);
    }
}
```

### Web Server with Axum

```rust
use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::Json,
    routing::{get, post},
    Router,
};
use serde::{Deserialize, Serialize};
use std::sync::Arc;

#[derive(Serialize, Deserialize)]
struct User {
    id: u32,
    name: String,
}

#[derive(Clone)]
struct AppState {
    users: Arc<Vec<User>>,
}

async fn get_user(Path(id): Path<u32>, State(state): State<AppState>) -> Result<Json<User>, StatusCode> {
    state.users
        .iter()
        .find(|user| user.id == id)
        .cloned()
        .ok_or(StatusCode::NOT_FOUND)
        .map(Json)
}

async fn create_user(
    State(state): State<AppState>,
    Json(user): Json<User>,
) -> Result<Json<User>, StatusCode> {
    // In a real app, you'd add to database
    Ok(Json(user))
}

#[tokio::main]
async fn main() {
    let state = AppState {
        users: Arc::new(vec![
            User { id: 1, name: "Alice".to_string() },
            User { id: 2, name: "Bob".to_string() },
        ]),
    };

    let app = Router::new()
        .route("/users/:id", get(get_user))
        .route("/users", post(create_user))
        .with_state(state);

    let listener = tokio::net::TcpListener::bind("0.0.0.0:3000").await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
```

---

## Level 3: Advanced Features

### Testing Rust Code

**Unit Tests**:
```rust
#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_addition() {
        assert_eq!(2 + 2, 4);
    }
    
    #[test]
    fn test_error_handling() {
        match parse_number("50") {
            Ok(num) => assert_eq!(num, 50),
            Err(_) => panic!("Should not error"),
        }
        
        assert!(parse_number("150").is_err());
        assert!(parse_number("abc").is_err());
    }
}

// Run with: cargo test
```

**Async Tests**:
```rust
#[tokio::test]
async fn test_async_operation() {
    let result = async_function().await;
    assert_eq!(result, expected);
}

#[tokio::test]
async fn test_web_server() {
    let app = create_test_app();
    let response = app
        .oneshot(Request::builder()
            .uri("/users/1")
            .body(Body::empty())
            .unwrap())
        .await
        .unwrap();
    
    assert_eq!(response.status(), StatusCode::OK);
}
```

### Performance Optimization

**Cargo.toml**:
```toml
[profile.release]
opt-level = 3
lto = true
codegen-units = 1
strip = true
panic = "abort"
```

**Benchmarking with Criterion**:
```rust
use criterion::{black_box, criterion_group, criterion_main, Criterion};

fn fibonacci(n: u64) -> u64 {
    match n {
        0 => 1,
        1 => 1,
        n => fibonacci(n - 1) + fibonacci(n - 2),
    }
}

fn criterion_benchmark(c: &mut Criterion) {
    c.bench_function("fib 20", |b| b.iter(|| fibonacci(black_box(20))));
}

criterion_group!(benches, criterion_benchmark);
criterion_main!(benches);
```

---

## Level 4: Production Deployment

### Docker Deployment

```dockerfile
FROM rust:1.91 as builder
WORKDIR /app
COPY . .
RUN cargo build --release

FROM debian:bookworm-slim
COPY --from=builder /app/target/release/app /usr/local/bin/
CMD ["app"]
```

### Production Best Practices

1. **Embrace the borrow checker**
2. **Use Result for error handling**
3. **Leverage type system for correctness**
4. **Test thoroughly with #[test]**
5. **Optimize with release profile**
6. **Use Tokio for async I/O**
7. **Implement proper error types**
8. **Profile with perf tools**
9. **Document unsafe code**
10. **Keep dependencies minimal**

### Related Skills
- `Skill("moai-essentials-perf")` for performance optimization
- `Skill("moai-security-backend")` for security patterns
- `Skill("moai-domain-cli-tool")` for CLI development

---

**Version**: 4.0.0 Enterprise  
**Last Updated**: 2025-11-13  
**Status**: Production Ready
