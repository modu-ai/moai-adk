---
paths:
  - "**/*.rs"
  - "**/Cargo.toml"
  - "**/Cargo.lock"
---

# Rust Rules

Version: Rust 1.92+

## Tooling

- Linting: clippy
- Formatting: rustfmt
- Testing: cargo test with coverage
- Package management: cargo

## Preferred Patterns

- Use Result<T, E> for error handling
- Use tokio for async runtime
- Minimize unsafe blocks

## MoAI Integration

- Use Skill("moai-lang-rust") for detailed patterns
- Follow TRUST 5 quality gates
