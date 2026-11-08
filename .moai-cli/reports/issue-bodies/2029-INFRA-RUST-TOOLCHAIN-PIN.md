<!-- moai-template:moai-bugfix@0.3.0; 2026-11-08 -->
```yaml
logical_key: INFRA-RUST-TOOLCHAIN-PIN
milestone: 2026.11.0-rc.1
parent_key: EP-09
work_kind: bugfix
related: [00-E, 08-D]
claim: {state: unclaimed}
evidence_status: not_run
```
# [RC1][결함] 로컬(1.94.1)과 rc-ci 러너(1.99) Rust 툴체인 불일치 — `rust-toolchain.toml` 고정

## 재현(2026-11-08, PR #2027 rc-ci run 42465855778)
로컬 `cargo clippy -D warnings` 0인데 러너(Rust 1.99 clippy)가 `clippy::unnecessary_sort_by`(tests/product_cli.rs:1147)로 실패. 로컬 `rustc 1.94.1`, `clippy 0.1.94`. 레인·감사자의 "clippy 0"이 CI 통과를 보장하지 못함(agent-2 #2020 감사에서 비차단 위험으로 예고).

## 리더 결정(RC1 임시, 2026-11-08 18:1x; 출처 docs/decisions-2026-11-08.md)
`rust-toolchain.toml`로 **로컬에 이미 있는 1.94.1**을 고정한다(모든 세션이 재설치 없이 즉시 일치; 러너는 rustup이 파일을 읽어 그 버전을 설치). 최신 stable로의 상향은 08-D(release gate)에서 결정.

## 수정 방향
- 저장소 루트에 `rust-toolchain.toml`(`[toolchain] channel = "1.94.1"`, `components = ["clippy", "rustfmt"]`, `profile = "minimal"`).
- `rc-ci.yml`: 툴체인 설치가 rustup 기본 동작으로 파일을 따르는지 확인(필요 시 `rustup show` step을 summary에 추가, 액션 추가 없음). 러너의 `rustc -V` 출력이 1.94.1임을 로그로 확인.
- `docs/rust-product-path.md`·AGENTS.local.md 검사 절에 "툴체인은 rust-toolchain.toml 고정" 1줄(AGENTS.local.md는 리더 PR).
- 공식 문서 확인·기록: rustup overrides(`rust-toolchain.toml`) https://rust-lang.github.io/rustup/overrides.html, GitHub 러너의 rustup 포함 여부(runner-images).

## 인수 기준
- AC-TC-1: 이 PR의 rc-ci 로그에 `rustc 1.94.1` 출력, clippy·test 통과 · 결과: not_run
- AC-TC-2: 로컬 `rustup show active-toolchain`이 1.94.1(파일 적용) · 결과: not_run
- AC-TC-3: 검사 묶음 통과, 독립 AUDIT PASS → 통합 · 결과: not_run

## 범위
- 변경 가능: `rust-toolchain.toml`(신규), `.github/workflows/rc-ci.yml`(00-E 1차 소유였으나 통합됨 — 이 Issue는 툴체인 확인 step만), `docs/rust-product-path.md`
- 변경 금지: `Cargo.lock` 의존성, src/

## HOLD
- 최신 stable 상향 시점 · 상태: HOLD — 08-D
