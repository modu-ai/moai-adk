---
name: moai-lang-rust
description: Rust best practices with cargo test, clippy, rustfmt, and ownership/borrow
  checker mastery
allowed-tools:
- Read
- Bash
---

# Rust Expert

## What it does

Provides Rust-specific expertise for TDD development, including cargo test, clippy linting, rustfmt formatting, and ownership/borrow checker compliance.

## When to use

- "Rust 테스트 작성", "cargo test 사용법", "소유권 규칙"
- Automatically invoked when working with Rust projects
- Rust SPEC implementation (`/alfred:2-build`)

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

## Examples

### Example 1: TDD with cargo test
User: "/alfred:2-build PARSER-001"
Claude: (creates RED test, GREEN implementation with Result<T, E>, REFACTOR with lifetimes)

### Example 2: Clippy check
User: "clippy 린트 실행"
Claude: (runs cargo clippy -- -D warnings and reports issues)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Rust-specific review)
- alfred-performance-optimizer (Rust benchmarking)
