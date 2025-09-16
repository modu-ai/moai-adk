# Rust 규칙(요약)

- 도구: rustfmt, clippy, cargo-nextest
- 에러: anyhow/thiserror, Result 반환, panic 금지(초기화 제외)
- 동시성: Send/Sync 주의, tokio runtime 규칙 준수
- 소유권: borrowing 우선, clone 최소화, lifetimes 단순화
