# Rust Coding Guidelines

## Naming (Rust-Specific)

| Rule | Guideline |
|------|-----------|
| No `get_` prefix | `fn name()` not `fn get_name()` |
| Iterator convention | `iter()` / `iter_mut()` / `into_iter()` |
| Conversion naming | `as_` (cheap &), `to_` (expensive), `into_` (ownership) |
| Static var prefix | `G_CONFIG` for `static`, no prefix for `const` |

## Data Types

| Rule | Guideline |
|------|-----------|
| Use newtypes | `struct Email(String)` for domain semantics |
| Prefer slice patterns | `if let [first, .., last] = slice` |
| Pre-allocate | `Vec::with_capacity()`, `String::with_capacity()` |
| Avoid Vec abuse | Use arrays for fixed sizes |

## Strings

| Rule | Guideline |
|------|-----------|
| Prefer bytes | `s.bytes()` over `s.chars()` when ASCII |
| Use `Cow<str>` | When might modify borrowed data |
| Use `format!` | Over string concatenation with `+` |
| Avoid nested iteration | `contains()` on string is O(n*m) |

## Error Handling

| Rule | Guideline |
|------|-----------|
| Use `?` propagation | Not `try!()` macro |
| `expect()` over `unwrap()` | When value guaranteed |
| Assertions for invariants | `assert!` at function entry |

## Memory

| Rule | Guideline |
|------|-----------|
| Meaningful lifetimes | `'src`, `'ctx` not just `'a` |
| `try_borrow()` for RefCell | Avoid panic |
| Shadowing for transformation | `let x = x.parse()?` |

## Concurrency

| Rule | Guideline |
|------|-----------|
| Identify lock ordering | Prevent deadlocks |
| Atomics for primitives | Not Mutex for bool/usize |
| Choose memory order carefully | Relaxed/Acquire/Release/SeqCst |

## Async

| Rule | Guideline |
|------|-----------|
| Sync for CPU-bound | Async is for I/O |
| Don't hold locks across await | Use scoped guards |

## Macros

| Rule | Guideline |
|------|-----------|
| Avoid unless necessary | Prefer functions/generics |
| Follow Rust syntax | Macro input should look like Rust |

## Deprecated → Better

| Deprecated | Better | Since |
|------------|--------|-------|
| `lazy_static!` | `std::sync::OnceLock` | 1.70 |
| `once_cell::Lazy` | `std::sync::LazyLock` | 1.80 |
| `std::sync::mpsc` | `crossbeam::channel` | - |
| `std::sync::Mutex` | `parking_lot::Mutex` | - |
| `failure`/`error-chain` | `thiserror`/`anyhow` | - |
| `try!()` | `?` operator | 2018 |

## Quick Reference

```
Naming: snake_case (fn/var), CamelCase (type), SCREAMING_CASE (const)
Format: rustfmt (just use it)
Docs: /// for public items, //! for module docs
Lint: #![warn(clippy::all)]
```

## Clippy Lints

Enable useful clippy lints:

```rust
#![warn(clippy::all)]
#![warn(clippy::pedantic)]
#![warn(clippy::cargo)]
```

Common clippy warnings to fix:
- `clippy::unwrap_used` - Use proper error handling
- `clippy::panic` - Use Result instead
- `clippy::clone_on_copy` - Unnecessary clone
- `clippy::needless_lifetimes` - Elidable lifetimes

## Documentation

### Public Items

Document all public items with `///`:

```rust
/// Adds two numbers together.
///
/// # Examples
///
/// ```
/// let result = add(2, 3);
/// assert_eq!(result, 5);
/// ```
///
/// # Panics
///
/// This function will panic if the addition overflows.
///
/// # Errors
///
/// This function never returns an error.
///
/// # Safety
///
/// This function is safe to call.
pub fn add(a: i32, b: i32) -> i32 {
    a + b
}
```

### Module Documentation

Use `//!` for module-level docs:

```rust
//! # My Crate
//!
//! This crate provides functionality for...
```

## Unsafe Rust

### Guidelines

- Minimize unsafe code
- Document safety invariants
- Prefer safe abstractions
- Use `unsafe` blocks, not functions

### Example

```rust
/// # Safety
///
/// The pointer must be non-null and properly aligned.
/// The memory must be valid for reads of `T`.
unsafe fn read_raw<T>(ptr: *const T) -> T {
    ptr.read()
}
```

## Testing

### Unit Tests

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_add() {
        assert_eq!(add(2, 3), 5);
    }
}
```

### Integration Tests

Place in `tests/` directory:

```rust
// tests/integration_test.rs
use my_crate;

#[test]
fn test_public_api() {
    assert_eq!(my_crate::public_function(), expected);
}
```

## Cargo Workspace

Organize large projects with workspaces:

```toml
# Cargo.toml
[workspace]
members = [
    "crate1",
    "crate2",
    "crate3",
]
```

## Dependencies

### Semantic Versioning

- Specify version ranges carefully
- Use `~` for compatible updates
- Lock important dependencies

```toml
[dependencies]
serde = "1.0"      # Allows 1.x.x
tokio = { version = "1.0", features = ["full"] }
```

### Feature Flags

Use feature flags for optional functionality:

```toml
[features]
default = ["std"]
std = []
async = ["tokio"]
```

Remember: These guidelines focus on Rust-specific rules that aren't obvious. Claude knows standard programming conventions well.
