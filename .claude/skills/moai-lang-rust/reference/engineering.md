# Rust Engineering Patterns

## Role Definition

Senior Rust engineer with 10+ years of systems programming experience specializing in Rust's ownership model, async programming with tokio, trait-based design, and performance optimization.

## Core Workflow

1. **Analyze ownership** - Design lifetime relationships and borrowing patterns
2. **Design traits** - Create trait hierarchies with generics and associated types
3. **Implement safely** - Write idiomatic Rust with minimal unsafe code
4. **Handle errors** - Use Result/Option with ? operator and custom error types
5. **Test thoroughly** - Unit tests, integration tests, property testing, benchmarks

## MUST DO

- Use ownership and borrowing for memory safety
- Minimize unsafe code (document all unsafe blocks)
- Use type system for compile-time guarantees
- Handle all errors explicitly (Result/Option)
- Add comprehensive documentation with examples
- Run clippy and fix all warnings
- Use cargo fmt for consistent formatting
- Write tests including doctests

## MUST NOT DO

- Use unwrap() in production code (prefer expect() with messages)
- Create memory leaks or dangling pointers
- Use unsafe without documenting safety invariants
- Ignore clippy warnings
- Mix blocking and async code incorrectly
- Skip error handling
- Use String when &str suffices
- Clone unnecessarily (use borrowing)

## Ownership and Borrowing

### Ownership Rules

1. Each value has an owner
2. There can only be one owner at a time
3. When the owner goes out of scope, the value is dropped

### Borrowing

- **Immutable borrow**: `&T` - Multiple references allowed, no modification
- **Mutable borrow**: `&mut T` - Single reference only, allows modification
- **Borrow checker**: Enforces that references are always valid

### Lifetimes

Lifetimes ensure references remain valid. Use meaningful lifetime names like `'src`, `'ctx` instead of just `'a`.

```rust
struct Context<'a> {
    data: &'a str,
}

impl<'a> Context<'a> {
    fn new(data: &'a str) -> Self {
        Context { data }
    }
}
```

### Smart Pointers

- **Box<T>** - Heap allocation with single ownership
- **Rc<T>** - Reference counting for multiple ownership
- **Arc<T>** - Thread-safe reference counting
- **RefCell<T>** - Runtime borrow checking for interior mutability
- **Mutex<T>** - Thread-safe mutual exclusion

## Traits

### Trait Design

Create small, focused traits with single responsibility. Use associated types for flexibility. Implement traits for external types with newtype pattern.

### Common Traits

- **Iterator** - Sequential access to elements
- **IntoIterator** - Conversion to iterator
- **From/Into** - Type conversions
- **AsRef/AsMut** - Cheap reference conversions
- **Borrow/BorrowMut** - Generic borrowing
- **ToOwned** - Create owned data from borrowed
- **Display/Debug** - String representation

### Derive Macros

Use derive macros for common trait implementations:
- `Clone` - Duplicate value
- `Copy` - Bitwise copy (for types without Drop)
- `Debug` - Debug formatting
- `PartialEq/Eq` - Equality comparisons
- `PartialOrd/Ord` - Ordering comparisons
- `Hash` - Hashable values
- `Default` - Default value

## Error Handling

### Result and Option

Use `Result<T, E>` for operations that can fail. Use `Option<T>` for optional values.

```rust
fn divide(a: f64, b: f64) -> Result<f64, String> {
    if b == 0.0 {
        Err(String::from("Division by zero"))
    } else {
        Ok(a / b)
    }
}

fn get_first(items: &[i32]) -> Option<&i32> {
    items.first()
}
```

### Error Propagation

Use `?` operator for error propagation:

```rust
fn read_file(path: &str) -> Result<String, io::Error> {
    let content = fs::read_to_string(path)?;
    Ok(content)
}
```

### Custom Error Types

Use `thiserror` for custom error types:

```rust
use thiserror::Error;

#[derive(Error, Debug)]
pub enum MyError {
    #[error("IO error: {0}")]
    Io(#[from] io::Error),

    #[error("Parse error: {0}")]
    Parse(String),

    #[error("Not found: {0}")]
    NotFound(String),
}
```

## Async Programming

### Async/Await

Use `async` functions for asynchronous operations. Use `.await` to suspend execution.

```rust
async fn fetch_data(url: &str) -> Result<String, Error> {
    let response = reqwest::get(url).await?;
    let text = response.text().await?;
    Ok(text)
}
```

### Tokio Runtime

Use tokio for async runtime:

```rust
#[tokio::main]
async fn main() -> Result<(), Error> {
    let data = fetch_data("https://example.com").await?;
    println!("{}", data);
    Ok(())
}
```

### Concurrency

Spawn tasks with `tokio::spawn`. Use channels for communication:

```rust
use tokio::sync::mpsc;

#[tokio::main]
async fn main() {
    let (tx, mut rx) = mpsc::channel(100);

    tokio::spawn(async move {
        tx.send("Hello").await.unwrap();
    });

    while let Some(message) = rx.recv().await {
        println!("{}", message);
    }
}
```

## Testing

### Unit Tests

Write tests in the same module with `#[cfg(test)]`:

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_add() {
        assert_eq!(add(2, 3), 5);
    }

    #[test]
    fn test_add_negative() {
        assert_eq!(add(-1, -2), -3);
    }
}
```

### Integration Tests

Place integration tests in `tests/` directory:

```rust
// tests/integration_test.rs
use my_crate;

#[test]
fn test_integration() {
    let result = my_crate::public_function();
    assert_eq!(result, expected);
}
```

### Doctests

Include examples in documentation that run as tests:

```rust
/// Adds two numbers together.
///
/// # Examples
///
/// ```
/// use my_crate::add;
/// assert_eq!(add(2, 3), 5);
/// ```
pub fn add(a: i32, b: i32) -> i32 {
    a + b
}
```

## Performance Optimization

### Zero-Cost Abstractions

Rust provides zero-cost abstractions - use iterators instead of loops:

```rust
// Prefer this
let sum: i32 = numbers.iter().sum();

// Over this
let mut sum = 0;
for n in &numbers {
    sum += n;
}
```

### Avoid Allocations

Use references instead of cloning. Use `Cow<T>` for conditional cloning.

### Profile with Criterion

Use criterion for statistical benchmarks:

```rust
use criterion::{black_box, criterion_group, criterion_main, Criterion};

fn fibonacci(n: u64) -> u64 {
    match n {
        0 => 1,
        1 => 1,
        _ => fibonacci(n - 1) + fibonacci(n - 2),
    }
}

fn criterion_benchmark(c: &mut Criterion) {
    c.bench_function("fib 20", |b| b.iter(|| fibonacci(black_box(20))));
}

criterion_group!(benches, criterion_benchmark);
criterion_main!(benches);
```

## Knowledge Reference

Rust 2021, Cargo, ownership/borrowing, lifetimes, traits, generics, async/await, tokio, Result/Option, thiserror/anyhow, serde, clippy, rustfmt, cargo-test, criterion benchmarks, MIRI, unsafe Rust
