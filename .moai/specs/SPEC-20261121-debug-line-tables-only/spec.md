---
id: SPEC-20261121-debug-line-tables-only
title: "Dev and test profiles use line-tables-only debug info; the build folder records the before and after"
status: draft
revision: 1
created: 2026-11-21
---

# Dev and test profiles use line-tables-only debug info; the build folder records the before and after

## Intent

리더 배차(사용자 지시, 2026-11-21T03:47Z): dev/test 프로필의 디버그 정보를 `line-tables-only` 로 줄이되 백트레이스는 유지해요. 근거는 보고서 `test-load-redesign-2026-11-21.md` 3.2 — 한 워크트리 `target` 9.8 GiB 중 `incremental` 7.6 GiB, `.o` 78,489개(macOS `split-debuginfo=unpacked` 라 디버그 정보가 `.o` 에 누적), 테스트마다 수만 파일을 쓰고 지워 fseventsd 가 CPU 99.6% 를 먹어요. `line-tables-only` 는 파일·줄 번호를 남겨 백트레이스 품질을 유지하고(Rust 참고 문서: "minimum amount of debug info needed to implement the backtrace feature"), 변수·타입 정보만 생략해요. 이 워크트리 before 실측(2026-11-21T03:51Z): `target/debug` 8.6 GiB, `incremental` 6.3 GiB(파일 14,395), 파일 총 38,434(`.o` 30,767), warm lib 재빌드(`touch src/lib.rs` → `cargo test --lib --no-run`) 33.50 s. 도구가 1.94.1 고정(`rust-toolchain.toml`)이라 `debug = "line-tables-only"`(1.71 안정)를 쓸 수 있어요. 출하 빌드(release·dist)는 그대로 두는 제품 무변화 변경이에요.

## Non-goals

- release·dist 프로필은 건드리지 않아요(출하 바이너리·사용자 설치 무변화).
- `split-debuginfo`, incremental 사용, 프로파일 상속 구조는 바꾸지 않아요 — `debug` 값만 정해요.
- 빌드 폴더 정리 정책(무엇을 언제 지울지)은 이 SPEC 의 영역이 아니에요(보고서의 별도 제안). 측정 조건을 맞추기 위해 이 워크트리 `incremental` 캐시를 한 번 지운 것만 기록해요.

## Requirements

- REQ-1: The file Cargo.toml shall set the debug value of the dev profile and the test profile to "line-tables-only".
- REQ-2: A panic backtrace from a binary built with the dev debuginfo shall keep file names and line numbers.
- REQ-3: The change shall leave the dist profile settings and the sha2 dev-package override untouched.

## Acceptance

- AC-1 (REQ-1): `Cargo.toml` 의 `[profile.dev]` 와 `[profile.test]` 가 모두 `debug = "line-tables-only"` 를 가지고, 장난감 프로젝트가 아니라 이 저장소의 verbose 빌드 두 개(`cargo build --bin moai-cli -v` 와 `cargo test --lib --no-run -v`)가 실제로 `-C debuginfo=line-tables-only` 를 rustc 에 넘겨요.
  verify: `cargo test --locked --offline --test profile_debug_info`
- AC-2 (REQ-2): `line-tables-only` 로 컴파일한 프로그램의 `RUST_BACKTRACE=1` 출력이 panic 위치의 파일 이름과 줄 번호를 말해요(실측 기록: rustc 직접 실측 + 같은 플래그가 verbose 빌드에 나타나는 AC-1 조합).
  verify: 이 SPEC Result 의 백트레이스 실측 기록(명령·출력 행 포함)
- AC-3 (REQ-3): `[profile.dist]` 의 `lto = "thin"`·`inherits = "release"` 와 `[profile.dev.package.sha2]` 의 `opt-level = 3` 이 그대로이고 `[profile.release]` 표는 생겨나지 않아요.
  verify: `cargo test --locked --offline --test profile_debug_info`
- AC-4 (REQ-1): 이 SPEC Result 가 before(2026-11-21T03:51Z, 위 Intent 값)와 after 의 크기(incremental 포함)·파일 수·lib 재빌드 시간(cold 1회·warm 1회)을 명령과 함께 기록해요.
  verify: 이 SPEC Result 의 after 표

## Scope

- write: `.moai/specs/SPEC-20261121-debug-line-tables-only/spec.md`, `Cargo.toml`, `tests/profile_debug_info.rs`(새)
- read-only: `rust-toolchain.toml`, `docs/`, `tests/cargo_metadata_test.py`
- forbidden: `target/` 전체 삭제, release·dist 프로필 변경, 실제 HOME, 네트워크

## Plan

1. RED: `tests/profile_debug_info.rs` — `Cargo.toml` 을 읽어 `[profile.dev]`·`[profile.test]` 의 `debug` 값이 `"line-tables-only"` 인지, `[profile.dist]`·sha2 오버라이드가 그대로인지 단언. 변경 전이므로 첫 단언들이 실패하는지 눈으로 봐요.
2. GREEN: `Cargo.toml` 에 `[profile.dev]` 와 `[profile.test]` 표를 더해요(`[profile.dev]` 는 기존 `[profile.dev.package.sha2]` 바로 앞에, 같은 그룹으로).
3. 실측: verbose 빌드 두 개에서 `-C debuginfo=line-tables-only` 확인 → 빈 폴더의 panic probe(rustc `-C debuginfo=line-tables-only`)에서 `RUST_BACKTRACE=1` 파일:줄 확인 → 이 워크트리 `target/debug/incremental` 한 번 정리 후 cold `cargo test --lib --no-run` 시간, warm 재측정, 크기·파일 수 재기록.
4. lib 전체 시험 + fmt + clippy, 문서에 프로필 표기가 있으면 갱신.

## Risks

- `[profile.test]` 명시가 내장 상속(dev)을 깨면 test 바이너리의 debuginfo가 달라져요: AC-1 의 `cargo test --lib --no-run -v` 실측이 막아요.
- 구형 rustc 로 소스 빌드하는 사용자는 `debug = "line-tables-only"` 를 unstable 로 거절할 수 있어요: 도구 고정이 1.94.1 이라 로컬·CI 동일하고, 출하 빌드는 dist(release 상속)라 무관해요.
- incremental 캐시 정리 후 이 워크트리 첫 빌드가 한 번 느려요: cold 측정이 그 비용을 기록해요.
- `line-tables-only` 로 변수 값이 보이지 않아 디버깅 정보가 줄어요: 백트레이스(함수·파일·줄)는 유지가 요구의 본질이라 REQ-2 로 고정해요.

## Result

- 구현: `Cargo.toml` 에 `[profile.dev]` 와 `[profile.test]` 가 `debug = "line-tables-only"` 를 가지고(`[profile.dev]` 는 기존 `[profile.dev.package.sha2]` 바로 앞에, 같은 그룹으로), `docs/worktree-strategy-2026-11-10.md` 의 미측정 문장(이전 159행)과 WT-10 행을 실측으로 갱신해요(갱신도 이 커밋에 들어가요).
- AC-1: met — `tests/profile_debug_info.rs` 의 `dev_and_test_profiles_pin_line_tables_only` 가 두 표를 고정하고, verbose 실측 두 개가 실제 rustc 플래그를 말해요: `cargo build --bin moai-cli -v`(2026-11-21T04:04Z)에서 lib·bin 모두 `-C debuginfo=line-tables-only -C split-debuginfo=unpacked`, `touch src/lib.rs; cargo test --lib --no-run -v`(03:59Z cold, 04:0xZ warm 재측정)에서 같은 플래그.
- AC-2: met — probe 실측: `rustc -C debuginfo=line-tables-only` 로 컴파일한 프로그램의 `RUST_BACKTRACE=1` 출력이 panic 위치를 `probe.rs:1:15` 와 `probe.rs:2:13` 로 말했어요 — 파일 이름과 줄 번호 유지.
- AC-3: met — `dist_profile_and_sha2_override_stay_untouched` 가 `[profile.dist]` 의 `inherits = "release"`·`lto = "thin"`, sha2 오버라이드 `opt-level = 3`, `[profile.release]` 표 부재를 고정해요.
- AC-4: met — 아래 before/after 기록(같은 워크트리, heavy slot 안, before 2026-11-21T03:51Z·after 03:59–04:06Z).
- 실측 표:

  | 항목 | before | after |
  | --- | --- | --- |
  | `target/debug` 총 | 8.6 GiB | 5.5 GiB |
  | `incremental` | 6.3 GiB / 14,395 파일 | 1.4 GiB / 1,301 파일 |
  | 파일 총(`.o`) | 38,434(30,767) | 29,336(24,602) |
  | warm lib 재빌드(`touch src/lib.rs` → `cargo test --lib --no-run`) | 33.50 s(user 39.37) | 13.68 s(user 5.94) |
  | cold lib 재빌드(정리된 캐시) | 재지 않음 | 41.01 s |
  | lib unittest 바이너리 1:1 | 46 MiB | 44 MiB |

- 해석: macOS `split-debuginfo=unpacked` 라 절감의 본체는 `.o` 와 incremental 캐시예요(incremental −78%, incremental 파일 −91%). 실행 바이너리 자체는 소폭만 줄어요(46→44 MiB — 디버그 정보가 `.o` 쪽에 나뉘어 있음). before warm은 부하 79·user 39.37 s, after warm은 더 낮은 부하에서 측정되어 시간 절감(−59%)이 다소 유리하게 읽힐 수 있어요; 컴파일 CPU 자체가 39.37→5.94 s 로 준 것도 같은 원인(incremental 부담 감소)이에요. cold 41.01 s 는 정리된 캐시에서의 첫 빌드 비용이고, 캐시 정리(이 워크트리 `target/debug/incremental` 만, 7.0 GiB)를 측정 조건으로 했어요.
- 검사: lib 전체 1541 passed 0(126.03 s), profile_debug_info 2 passed 0, fmt --check 클린, clippy --all-targets -D warnings 통과(Finished, 오류 없음), git diff --check 클린, built-binary spec check 0 오류 0 경고.
- 도구: rustc/cargo 1.94.1(`rust-toolchain.toml` 고정) — `debug = "line-tables-only"` 는 1.71 안정이라 호환 위험 없어요.
- 돌리지 않은 것: CI, release/dist 빌드 실측, 다른 워크트리의 정리, sccache(WT-10 의 남은 절반, 별도 카드).
