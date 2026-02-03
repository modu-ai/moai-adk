---
paths:
  - "**/*.rs"
  - "**/Cargo.toml"
  - "**/Cargo.lock"
---

# Rust Rules

Version: Rust 1.92+ (2024 edition)

## Tooling

- Build: Cargo
- Linting: clippy
- Formatting: rustfmt
- Testing: cargo test
- Coverage: cargo-llvm-cov >= 85%

## MUST

- Use Result and Option for error handling
- Implement proper error types with thiserror
- Use async/await with tokio for I/O
- Document public items with /// comments
- Use clippy with pedantic warnings
- Prefer references over cloning

## MUST NOT

- Use unwrap() in production code
- Use unsafe without clear justification
- Ignore clippy warnings without allow attribute
- Clone large data structures unnecessarily
- Use panic! for recoverable errors
- Leave TODO comments in production

## File Conventions

- Tests in same file with #[cfg(test)] module
- Integration tests in tests/ directory
- Use snake_case for modules and functions
- Use PascalCase for types and traits
- Use SCREAMING_CASE for constants

## Testing

- Use #[test] for unit tests
- Use proptest for property-based testing
- Use mockall for mocking traits
- Use test fixtures with rstest

## MoAI Integration

- Use Skill("moai-lang-rust") for detailed patterns
- Follow TRUST 5 quality gates
