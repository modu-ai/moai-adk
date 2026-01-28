---
paths:
  - "**/*.rs"
  - "**/Cargo.toml"
  - "**/Cargo.lock"
---

# Rust Rules

Version: Rust 1.92+

## Tooling

- Linting: clippy (800+ lints)
- Formatting: rustfmt
- Testing: cargo test with coverage
- Package management: cargo

## Clippy Configuration

```toml
# clippy.toml or Cargo.toml
[lints.clippy]
pedantic = "warn"
nursery = "warn"
# restriction lints: enable case-by-case only
```

## CI Integration

```bash
# Recommended CI command - fail on warnings
cargo clippy --all-targets --all-features -- -D warnings
```

## Best Practices (2026)

- Make code Clippy-warning free before commit
- Use `thiserror` for library errors, `anyhow` for applications
- Use tokio as the async runtime
- Use `Result<T, E>` for all fallible operations
- Minimize unsafe blocks - document when necessary

## Lint Categories

- **correctness**: Common programming errors (always fix)
- **perf**: Performance issues (review carefully)
- **style**: Readability and consistency
- **complexity**: Overly complex code
- **pedantic**: More nuanced suggestions

## Error Handling Pattern

```rust
#[derive(Debug, thiserror::Error)]
pub enum AppError {
    #[error("Database error: {0}")]
    Database(#[from] sqlx::Error),
    #[error("Not found: {0}")]
    NotFound(String),
}
```

## MoAI Integration

- Use Skill("moai-lang-rust") for detailed patterns
- Follow TRUST 5 quality gates
