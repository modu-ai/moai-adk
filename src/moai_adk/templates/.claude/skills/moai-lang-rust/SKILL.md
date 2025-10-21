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
- 언어별 소스 디렉터리(e.g. `src/`, `app/`).
- 언어별 빌드/테스트 설정 파일(예: `package.json`, `pyproject.toml`, `go.mod`).
- 관련 테스트 스위트 및 샘플 데이터.

## Outputs
- 선택된 언어에 맞춘 테스트/린트 실행 계획.
- 주요 언어 관용구와 리뷰 체크포인트 목록.

## Failure Modes
- 언어 런타임이나 패키지 매니저가 설치되지 않았을 때.
- 다중 언어 프로젝트에서 주 언어를 판별하지 못했을 때.

## Dependencies
- Read/Grep 도구로 프로젝트 파일 접근이 필요합니다.
- `Skill("moai-foundation-langs")`와 함께 사용하면 교차 언어 규약 공유가 용이합니다.

## References
- Rust Project Developers. "The Rust Programming Language." https://doc.rust-lang.org/book/ (accessed 2025-03-29).
- Rust Project Developers. "Clippy." https://doc.rust-lang.org/clippy/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: 언어별 입력·출력·실패 대응·참조 정보를 명세했습니다.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Rust-specific review)
- alfred-performance-optimizer (Rust benchmarking)

## Best Practices
- 언어 공식 스타일 가이드와 린터를 일치시켜 자동 검증을 활성화하세요.
- CI에서 재현 가능한 명령으로 테스트/빌드 파이프라인을 고정합니다.
