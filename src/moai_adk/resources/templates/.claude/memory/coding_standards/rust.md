# Rust 규칙

## ✅ 필수
- `rustfmt`, `clippy -D warnings`, `cargo nextest` 로 포맷/정적 분석/테스트 수행
- 에러는 `anyhow`/`thiserror`로 구성, `Result` 반환을 기본으로 하고 panic은 초기화 단계에서만 허용
- `Send/Sync` 제약을 명확히 하고 tokio runtime 정책(멀티/커런트) 준수
- 소유권은 borrowing 우선, 불필요한 `clone` 제거, lifetime 단순화

## 👍 권장
- feature flag 로 기능 분기, workspaces 로 모듈화, `cargo deny` 로 의존성 점검
- tracing + structured log(`tracing_subscriber`), metrics(exporter) 통합
- 테스트: property 기반(quickcheck/proptest), fuzzing(cargo fuzz), integration 테스트 폴더 분리

## 🚀 확장/고급
- async trait/`async function in trait` 적용 시 `async_trait` 또는 GAT 활용
- WASM/embedded 타깃 고려 시 `no_std`/feature gating 전략 문서화
- Unsafe 블록은 justification 주석 필수, Miri/sanitizer로 런타임 검증
