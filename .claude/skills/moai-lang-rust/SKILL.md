---

name: moai-lang-rust
description: Rust best practices with cargo test, clippy, rustfmt, and ownership/borrow checker mastery. Use when writing or reviewing Rust code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Rust Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Rust code discussions, framework guidance, or file extensions such as .rs. |
| Tier | 3 |

## What it does

Provides Rust-specific expertise for TDD development, including cargo test, clippy linting, rustfmt formatting, and ownership/borrow checker compliance.

## When to use

- Engages when the conversation references Rust work, frameworks, or files like .rs.
- “Writing Rust tests”, “How to use cargo tests”, “Ownership rules”
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

## Examples
```bash
cargo test && cargo clippy -- -D warnings
```

## Inputs
- Language-specific source directories (e.g. `src/`, `app/`).
- Language-specific build/test configuration files (e.g. `package.json`, `pyproject.toml`, `go.mod`).
- Relevant test suites and sample data.

## Outputs
- Test/lint execution plan tailored to the selected language.
- List of key language idioms and review checkpoints.

## Failure Modes
- When the language runtime or package manager is not installed.
- When the main language cannot be determined in a multilingual project.

## Dependencies
- Access to the project file is required using the Read/Grep tool.
- When used with `Skill("moai-foundation-langs")`, it is easy to share cross-language conventions.

## References
- Rust Project Developers. "The Rust Programming Language." https://doc.rust-lang.org/book/ (accessed 2025-03-29).
- Rust Project Developers. "Clippy." https://doc.rust-lang.org/clippy/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Input/output/failure response/reference information for each language has been specified.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Rust-specific review)
- alfred-performance-optimizer (Rust benchmarking)

## Best Practices
- Enable automatic validation by matching your linter with the language's official style guide.
- Fix test/build pipelines with reproducible commands in CI.
