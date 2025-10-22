# Rust 1.84+ CLI Reference — Tool Command Matrix

**Framework**: Rust 1.84+ + Cargo + Clippy + rustfmt

---

## Rust Compiler Commands

| Command | Purpose | Example |
|---------|---------|---------|
| `rustc --version` | Check Rust version | `rustc --version` → `rustc 1.84.0` |
| `rustc file.rs` | Compile Rust file | `rustc main.rs` |
| `rustc -O file.rs` | Compile with optimizations | `rustc -O main.rs` |
| `rustc --edition 2024 file.rs` | Compile with Rust 2024 edition | `rustc --edition 2024 main.rs` |
| `rustc --explain E0308` | Explain error code | `rustc --explain E0308` |
| `rustc --help` | Show compiler help | `rustc --help` |

---

## Cargo Commands (Rust 1.84+)

| Command | Purpose | Example |
|---------|---------|---------|
| `cargo --version` | Check Cargo version | `cargo --version` → `cargo 1.84.0` |
| `cargo new project` | Create new binary project | `cargo new my_app` |
| `cargo new --lib project` | Create new library project | `cargo new --lib my_lib` |
| `cargo init` | Initialize project in current dir | `cargo init` |
| `cargo build` | Build project | `cargo build` |
| `cargo build --release` | Build optimized release | `cargo build --release` |
| `cargo run` | Run project | `cargo run` |
| `cargo run --release` | Run optimized release | `cargo run --release` |
| `cargo test` | Run tests | `cargo test` |
| `cargo test --release` | Run tests with optimizations | `cargo test --release` |
| `cargo test test_name` | Run specific test | `cargo test test_add` |
| `cargo test -- --nocapture` | Show println output | `cargo test -- --nocapture` |
| `cargo test -- --test-threads=1` | Run tests serially | `cargo test -- --test-threads=1` |
| `cargo bench` | Run benchmarks | `cargo bench` |
| `cargo doc` | Generate documentation | `cargo doc` |
| `cargo doc --open` | Generate and open docs | `cargo doc --open` |
| `cargo clean` | Clean build artifacts | `cargo clean` |
| `cargo check` | Check without building | `cargo check` |
| `cargo update` | Update dependencies | `cargo update` |
| `cargo tree` | Show dependency tree | `cargo tree` |
| `cargo add crate` | Add dependency | `cargo add serde` |
| `cargo remove crate` | Remove dependency | `cargo remove serde` |

**Cargo.toml MSRV Support (Rust 1.84)**:
```toml
[package]
name = "my_app"
version = "0.1.0"
edition = "2024"
rust-version = "1.84"  # Minimum Supported Rust Version

[dependencies]
serde = { version = "1.0", features = ["derive"] }
```

---

## Testing Framework (Built-in)

| Command | Purpose | Example |
|---------|---------|---------|
| `cargo test` | Run all tests | `cargo test` |
| `cargo test test_name` | Run specific test | `cargo test test_add` |
| `cargo test --lib` | Run library tests only | `cargo test --lib` |
| `cargo test --doc` | Run doc tests | `cargo test --doc` |
| `cargo test -- --ignored` | Run ignored tests | `cargo test -- --ignored` |
| `cargo test -- --show-output` | Show test output | `cargo test -- --show-output` |

---

## Linting (Clippy)

| Command | Purpose | Example |
|---------|---------|---------|
| `cargo clippy` | Run Clippy lints | `cargo clippy` |
| `cargo clippy --fix` | Auto-fix Clippy suggestions | `cargo clippy --fix` |
| `cargo clippy -- -D warnings` | Treat warnings as errors | `cargo clippy -- -D warnings` |
| `cargo clippy -- -W clippy::pedantic` | Enable pedantic lints | `cargo clippy -- -W clippy::pedantic` |

**Clippy Configuration (Cargo.toml)**:
```toml
[lints.clippy]
pedantic = "warn"
nursery = "warn"
unwrap_used = "deny"
expect_used = "deny"

[lints.rust]
unsafe_code = "forbid"
```

**New in Rust 1.84**: The `double_negations` lint (previously `clippy::double_neg`) is now built into rustc.

---

## Formatting (rustfmt)

| Command | Purpose | Example |
|---------|---------|---------|
| `cargo fmt` | Format all files | `cargo fmt` |
| `cargo fmt -- --check` | Check if formatting needed | `cargo fmt -- --check` |
| `rustfmt file.rs` | Format specific file | `rustfmt src/main.rs` |

**rustfmt.toml Configuration**:
```toml
edition = "2024"
max_width = 100
hard_tabs = false
tab_spaces = 4
newline_style = "Unix"
reorder_imports = true
```

---

## Code Coverage (cargo-tarpaulin / cargo-llvm-cov)

### cargo-tarpaulin

| Command | Purpose | Example |
|---------|---------|---------|
| `cargo install cargo-tarpaulin` | Install tarpaulin | `cargo install cargo-tarpaulin` |
| `cargo tarpaulin` | Generate coverage | `cargo tarpaulin` |
| `cargo tarpaulin --out Html` | Generate HTML report | `cargo tarpaulin --out Html` |

### cargo-llvm-cov

| Command | Purpose | Example |
|---------|---------|---------|
| `cargo install cargo-llvm-cov` | Install llvm-cov | `cargo install cargo-llvm-cov` |
| `cargo llvm-cov` | Generate coverage | `cargo llvm-cov` |
| `cargo llvm-cov --html` | Generate HTML report | `cargo llvm-cov --html` |
| `cargo llvm-cov --open` | Generate and open report | `cargo llvm-cov --open` |

---

## Combined Workflow (Quality Gate)

**Before Commit** (all must pass):

```bash
#!/bin/bash
set -e

echo "Running Rust quality gate checks..."

# 1. Check formatting
echo "1. Checking formatting..."
cargo fmt -- --check

# 2. Run Clippy
echo "2. Running Clippy..."
cargo clippy -- -D warnings

# 3. Build project
echo "3. Building..."
cargo build --release

# 4. Run tests
echo "4. Running tests..."
cargo test

# 5. Check coverage
echo "5. Checking coverage..."
if command -v cargo-tarpaulin &> /dev/null; then
  cargo tarpaulin --out Stdout
fi

echo "✅ All quality gates passed!"
```

---

## TRUST 5 Principles Integration

### T - Test First (Built-in)
```bash
cargo test
```

### R - Readable (rustfmt)
```bash
cargo fmt
cargo fmt -- --check
```

### U - Unified Types (Rust Type System)
```bash
cargo check   # Compile-time type checking
cargo clippy  # Additional type safety
```

### S - Security
```bash
cargo audit  # Install: cargo install cargo-audit
```

### T - Trackable (@TAG)
```bash
rg '@(CODE|TEST|SPEC):' -n src/ tests/ --type rust
```

---

**Version**: 0.1.0
**Created**: 2025-10-22
**Framework**: Rust 1.84+ CLI Tools Reference
