---
name: "moai-lang-rust"
description: "Enterprise Rust with ownership model and safety guarantees: Rust 1.91.1, Tokio 1.48, async/await, macro system, error handling, memory safety patterns; activates for systems programming, performance-critical code, concurrent applications, and safety-first development."
allowed-tools: 
version: "4.0.0"
status: stable
---

# Rust Systems Programming — Enterprise v4.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Version** | **4.0.0 Enterprise** |
| **Created** | 2025-11-12 |
| **Updated** | 2025-11-12 |
| **Lines** | ~950 lines |
| **Size** | ~30KB |
| **Tier** | **3 (Professional)** |
| **Allowed tools** | Read, Bash, WebSearch, WebFetch |
| **Auto-load** | Rust systems programming, async/await, Tokio patterns, safety |
| **Trigger cues** | Rust, Tokio, async, ownership, safety, systems, performance, macro |

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

## Level 1: Fundamentals (High Freedom)

### 1. Rust 1.91 Ownership System

Rust's ownership model ensures memory safety:

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

### 2. Tokio Async Runtime

Tokio provides production-grade async execution:

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

**Async HTTP with Tokio**:
```rust
use reqwest::Client;
use tokio::task;

#[tokio::main]
async fn fetch_urls(urls: Vec<&str>) {
    let client = Client::new();
    let mut handles = vec![];
    
    for url in urls {
        let client = client.clone();
        let handle = task::spawn(async move {
            match client.get(url).send().await {
                Ok(resp) => println!("Status: {}", resp.status()),
                Err(e) => println!("Error: {}", e),
            }
        });
        handles.push(handle);
    }
    
    for handle in handles {
        handle.await.ok();
    }
}
```

### 3. Error Handling

Rust's Result type enforces error handling:

**Using Result**:
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

**Custom Error Types**:
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

## Level 2: Advanced Patterns (Medium Freedom)

### 1. Procedural Macros

Procedural macros generate code at compile time:

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

### 2. Async with Tokio Channels

Tokio channels for task communication:

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

### 3. Testing Rust Code

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
    fn test_should_fail() {
        assert_eq!(1 + 1, 2);
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
```

## Level 3: Production Deployment (Low Freedom, Expert Only)

### 1. Release Optimization

Optimize Rust binaries for production:

**Cargo.toml**:
```toml
[profile.release]
opt-level = 3
lto = true
codegen-units = 1
strip = true
```

### 2. Docker Deployment

```dockerfile
FROM rust:1.91 as builder
WORKDIR /app
COPY . .
RUN cargo build --release

FROM debian:bookworm-slim
COPY --from=builder /app/target/release/app /usr/local/bin/
CMD ["app"]
```

## Auto-Load Triggers

This Skill activates when you:
- Work with Rust and memory safety
- Implement async operations with Tokio
- Need error handling patterns
- Create performance-critical code
- Develop systems programming
- Use procedural macros

## Best Practices

1. Embrace the borrow checker
2. Use Result for error handling
3. Leverage type system for correctness
4. Test thoroughly with #[test]
5. Optimize with release profile
6. Use Tokio for async I/O
7. Implement proper error types
8. Profile with perf tools
9. Document unsafe code
10. Keep dependencies minimal

