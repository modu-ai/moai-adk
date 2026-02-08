---
name: moai-lang-rust
description: >
  Rust 1.92+ development specialist covering Axum, Tokio, SQLx, and memory-safe systems programming. Use when building high-performance, memory-safe applications or WebAssembly.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.2.0"
  category: "language"
  status: "active"
  updated: "2026-01-11"
  modularized: "false"
  tags: "language, rust, axum, tokio, sqlx, serde, wasm, cargo"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["Rust", "Axum", "Tokio", "SQLx", "serde", ".rs", "Cargo.toml", "async", "await", "lifetime", "trait", "ownership", "borrowing", "performance", "optimization", "clippy", "memory safety"]
  languages: ["rust"]
---

## Quick Reference (30 seconds)

Rust 1.92+ Development Specialist with deep patterns for high-performance, memory-safe applications.

Auto-Triggers: `.rs`, `Cargo.toml`, async/await, Tokio, Axum, SQLx, serde, lifetimes, traits

Core Use Cases:

- High-performance REST APIs and microservices
- Memory-safe concurrent systems
- CLI tools and system utilities
- WebAssembly applications
- Low-latency networking services

Quick Patterns:

Axum REST API: Create Router with route macro chaining path and handler. Add with_state for shared state. Bind TcpListener with tokio::net and serve with axum::serve.

Async Handler with SQLx: Define async handler function taking State extractor for AppState and Path extractor for id. Use sqlx::query_as! macro with SQL string and bind parameters. Call fetch_optional on pool, await, and use ok_or for error conversion. Return Json wrapped result.

---

## Implementation Guide (5 minutes)

### Rust 1.92 Features

Modern Rust Features:

- Rust 2024 Edition available (released with Rust 1.85)
- Async traits in stable (no more async-trait crate needed)
- Const generics for compile-time array sizing
- let-else for pattern matching with early return
- Improved borrow checker with polonius

Async Traits (Stable): Define trait with async fn signatures. Implement trait for concrete types with async fn implementations. Call sqlx macros directly in trait methods.

Let-Else Pattern: Use let Some(value) = option else with return for early exit. Chain multiple let-else statements for sequential validation. Return error types in else blocks.

### Web Framework: Axum 0.8

Installation: In Cargo.toml dependencies section, add axum version 0.8, tokio version 1.48 with full features, and tower-http version 0.6 with cors and trace features.

Complete API Setup: Import extractors from axum::extract and routing macros. Define Clone-derive AppState struct holding PgPool. In tokio::main async main, create pool with PgPoolOptions setting max_connections and connecting with DATABASE_URL from env. Build Router with route chains for paths and handlers, add CorsLayer, and call with_state. Bind TcpListener and call axum::serve.

Handler Patterns: Define async handlers taking State, Path, and Query extractors with appropriate types. Use sqlx::query_as! for type-safe queries with positional binds. Return Result with Json success and AppError failure.

### Async Runtime: Tokio 1.48

Task Spawning and Channels: Create mpsc channel with capacity. Spawn worker tasks with tokio::spawn that receive from channel in loop. For timeouts, use tokio::select! macro with operation branch and sleep branch, returning error on timeout.

### Database: SQLx 0.8

Type-Safe Queries: Derive sqlx::FromRow on structs for automatic mapping. Use query_as! macro for compile-time checked queries. Call fetch_one or fetch_optional on pool. For transactions, call pool.begin, execute queries on transaction reference, and call tx.commit.

### Serialization: Serde 1.0

Derive Serialize and Deserialize on structs. Use serde attribute with rename_all for case conversion. Use rename attribute for field-specific naming. Use skip_serializing_if with Option::is_none. Use default attribute for default values.

### Error Handling

thiserror: Derive Error on enum with error attribute for display messages. Use from attribute for automatic conversion from source errors. Implement IntoResponse by matching on variants and returning status code with Json body containing error message.

### CLI Development: clap

Derive Parser on main Cli struct with command attributes for name, version, about. Use arg attribute for global flags. Derive Subcommand on enum for commands. Match on command in main to dispatch logic.

### Testing Patterns

Create test module with cfg(test) attribute. Define tokio::test async functions. Call setup helpers, invoke functions under test, and use assert! macros for verification.

---

## Advanced Patterns

For comprehensive coverage including:

- Advanced ownership patterns, lifetimes, and smart pointers
- Trait design and generic programming
- Performance optimization strategies and profiling
- Engineering best practices and coding guidelines
- Async patterns and concurrency

See: [reference/engineering.md](reference/engineering.md) for ownership and traits, [reference/performance.md](reference/performance.md) for optimization, [reference/guidelines.md](reference/guidelines.md) for coding standards

### Performance Optimization

Release Build: In Cargo.toml profile.release section, enable lto, set codegen-units to 1, set panic to abort, and enable strip.

### Deployment

Minimal Container: Use multi-stage Dockerfile. First stage uses rust alpine image, copies Cargo files, creates dummy main for dependency caching, builds release, copies source, touches main.rs for rebuild, builds final release. Second stage uses alpine, copies binary from builder, exposes port, and sets CMD.

### Concurrency

Rate-Limited Operations: Create Arc-wrapped Semaphore with max permits. Map over items spawning tasks that acquire permit, process, and return result. Use futures::future::join_all to collect results.

---

## Context7 Integration

Library Documentation Access:

- `/rust-lang/rust` - Rust language and stdlib
- `/tokio-rs/tokio` - Tokio async runtime
- `/tokio-rs/axum` - Axum web framework
- `/launchbadge/sqlx` - SQLx async SQL
- `/serde-rs/serde` - Serialization framework
- `/dtolnay/thiserror` - Error derive
- `/clap-rs/clap` - CLI parser

---

## Works Well With

- `moai-lang-go` - Go systems programming patterns
- `moai-domain-backend` - REST API architecture and microservices patterns
- `moai-foundation-quality` - Security hardening for Rust applications
- `moai-workflow-testing` - Test-driven development workflows

---

## Troubleshooting

Common Issues:

- Cargo errors: Run cargo clean followed by cargo build
- Version check: Run rustc --version and cargo --version
- Dependency issues: Run cargo update and cargo tree
- Compile-time SQL check: Run cargo sqlx prepare

Performance Characteristics:

- Startup Time: 50-100ms
- Memory Usage: 5-20MB base
- Throughput: 100k-200k req/s
- Latency: p99 less than 5ms
- Container Size: 5-15MB (alpine)

---

## Additional Resources

See reference/engineering.md for advanced ownership patterns and trait design.

See reference/performance.md for optimization strategies and profiling techniques.

See reference/guidelines.md for Rust coding standards and best practices.

---

Last Updated: 2026-01-11
Version: 1.2.0
