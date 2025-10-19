---
name: moai-lang-rust
description: Rust best practices with cargo test, clippy, rustfmt, and ownership/borrow checker mastery
allowed-tools:
  - Read
  - Bash
tier: 3
auto-load: "true"
---

# Rust Expert

## What it does

Provides Rust-specific expertise for TDD development, including cargo test, clippy linting, rustfmt formatting, and ownership/borrow checker compliance.

## When to use

- "Rust 테스트 작성", "cargo test 사용법", "소유권 규칙", "시스템 프로그래밍", "웹어셈블리", "성능 최적화"
- "WebAssembly", "임베디드 시스템", "네트워크 프로그래밍", "CLI 도구", "데이터베이스"
- "tokio", "Actix", "Axum", "Tauri", "Bevy"
- Automatically invoked when working with Rust projects
- Rust SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **cargo test**: Built-in test framework
- **proptest**: Property-based testing
- **criterion**: Benchmarking
- Test coverage with `cargo tarpaulin` or `cargo llvm-cov`

**Code Quality**:
- **clippy**: Rust linter with 500+ lint rules
- **rustfmt**: Automatic code formatting
- **cargo check**: Fast compilation check
- **cargo audit**: Security vulnerability scanning

**Memory Safety**:
- **Ownership**: One owner per value
- **Borrowing**: Immutable (&T) or mutable (&mut T) references
- **Lifetimes**: Explicit lifetime annotations when needed
- **Move semantics**: Understanding Copy vs Clone

**Rust Patterns**:
- Result<T, E> for error handling (no exceptions)
- Option<T> for nullable values
- Traits for polymorphism
- Match expressions for exhaustive handling

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Prefer immutable bindings (let vs let mut)
- Use iterators over manual loops
- Avoid `unwrap()` in production code, use proper error handling

## Modern Rust (1.75+)

**Recommended Version**: Rust 1.75+ for production, 1.70+ for stable features

**Modern Features**:
- **async/await** (stable): Asynchronous programming
- **Trait objects** (dyn Trait): Dynamic dispatch
- **const generics** (stable): Compile-time arrays
- **GAT (Generic Associated Types)** (stable): Flexible type associations
- **min_specialization** (nightly): Specialized trait implementations
- **let-else** (1.80+): Pattern matching with else branch

**Version Check**:
```bash
rustc --version  # Check Rust version
cargo --version
rustup update stable  # Update to latest stable
```

## Package Management Commands

### Using cargo (Built-in - Recommended)
```bash
# Initialize project
cargo new my_project
cargo new --lib my_library

# Add dependencies
cargo add tokio
cargo add --dev proptest criterion

# Add specific version
cargo add serde@1.0

# Remove dependencies
cargo remove tokio

# Update dependencies
cargo update
cargo update -p tokio

# Show dependency tree
cargo tree
cargo tree --duplicate

# Build and run
cargo build
cargo build --release  # Optimized build
cargo run
cargo run --release

# Test
cargo test
cargo test -- --nocapture
cargo test --release

# Benchmarks
cargo bench

# Format
cargo fmt
cargo fmt -- --check

# Lint
cargo clippy
cargo clippy -- -D warnings

# Security audit
cargo audit
cargo audit fix

# Generate documentation
cargo doc --open
```

### Project Configuration
```bash
# Cargo.toml examples
[package]
name = "my_app"
version = "0.1.0"
edition = "2021"

[dependencies]
tokio = { version = "1", features = ["full"] }
serde = { version = "1.0", features = ["derive"] }

[dev-dependencies]
criterion = "0.5"

[profile.release]
opt-level = 3
lto = true  # Link-time optimization
```

## Examples

### Example 1: TDD with cargo test
User: "/alfred:2-run PARSER-001"
Claude: (creates RED test, GREEN implementation with Result<T, E>, REFACTOR with lifetimes)

### Example 2: Clippy check
User: "clippy 린트 실행"
Claude: (runs cargo clippy -- -D warnings and reports issues)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Rust-specific review)
- alfred-performance-optimizer (Rust benchmarking)
