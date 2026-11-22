---
id: SPEC-20261121-debug-line-tables-only
title: "Dev and test builds keep line tables for this crate and no debug info for dependencies; the build folder records the before and after"
status: draft
revision: 2
created: 2026-11-21
---

# Dev and test builds keep line tables for this crate and no debug info for dependencies; the build folder records the before and after

## Intent

리더 배차(사용자 지시, 2026-11-21T03:47Z): dev/test 프로필의 디버그 정보를 `line-tables-only` 로 줄이되 백트레이스는 유지해요. 2026-11-22 리더 배차가 한 가지를 더했어요: 우리 크레이트는 `line-tables-only`, 의존성은 디버그 정보 없음. 근거는 보고서 `test-load-redesign-2026-11-21.md` 3.2 — 한 워크트리 `target` 9.8 GiB 중 `incremental` 7.6 GiB, `.o` 78,489개(macOS `split-debuginfo=unpacked` 라 디버그 정보가 `.o` 에 쌓여요), 테스트마다 수만 파일을 쓰고 지워 fseventsd 가 CPU 99.6% 를 먹었어요. `line-tables-only` 는 함수·파일·줄 번호를 남겨 이 크레이트의 백트레이스를 유지하고(Rust 참고 문서: "minimum amount of debug info needed to implement the backtrace feature"), 변수·타입 정보만 빼요. 의존성은 우리가 고치지 않는 코드라서 디버그 정보를 빼요: 의존성 안의 프레임은 함수 이름만 남고 파일·줄은 사라져요. 도구가 1.94.1 고정(`rust-toolchain.toml`)이라 `debug = "line-tables-only"`(1.71 안정)를 쓸 수 있어요. 출하 빌드(release·dist)는 그대로 두는 제품 무변화 변경이에요.

첫 초안(revision 1, 다른 레인, 착지 안 됨)의 실측은 incremental 을 켠 조건이었어요. 지금 레인은 `CARGO_INCREMENTAL=0` 으로 빌드하므로 revision 2 는 그 조건으로 before·after 를 다시 재요.

## Non-goals

- release·dist 프로필은 건드리지 않아요(출하 바이너리·사용자 설치 무변화).
- `split-debuginfo`, incremental 사용, 프로필 상속 구조는 바꾸지 않아요 — `debug` 값만 정해요.
- 빌드 폴더 정리 정책은 이 SPEC 의 영역이 아니에요. 측정은 비어 있는 새 빌드 폴더 둘에서 해요.
- sccache 의 효과는 재지 않아요(측정은 sccache 없이).

## Requirements

- REQ-1: The file Cargo.toml shall set the debug value of the dev profile and the test profile to "line-tables-only".
- REQ-2: The file Cargo.toml shall set the debug value of every dependency in the dev profile and the test profile to false.
- REQ-3: A panic backtrace from a binary built with the dev debuginfo shall keep the file names and line numbers of this crate's frames.
- REQ-4: The change shall leave the dist profile settings and the sha2 dev-package override untouched.

## Acceptance

- AC-1 (REQ-1, REQ-2): `Cargo.toml` 의 `[profile.dev]`·`[profile.test]` 가 `debug = "line-tables-only"` 를, `[profile.dev.package."*"]`·`[profile.test.package."*"]` 가 `debug = false` 를 가지고, 이 저장소의 verbose 빌드에서 `moai_cli` 의 rustc 줄은 `-C debuginfo=line-tables-only` 를, 의존성(예: `serde_json`)의 줄은 디버그 정보 플래그 없이 나와요.
  verify: `cargo test --locked --offline --test profile_debug_info`
- AC-2 (REQ-3): `line-tables-only` 로 컴파일한 프로그램의 `RUST_BACKTRACE=1` 출력이 panic 위치의 파일 이름과 줄 번호를 말해요.
  verify: 이 SPEC Result 의 백트레이스 실측 기록(명령·출력 행 포함)
- AC-3 (REQ-4): `[profile.dist]` 의 `lto = "thin"`·`inherits = "release"` 와 `[profile.dev.package.sha2]` 의 `opt-level = 3` 이 그대로이고 `[profile.release]` 표는 생겨나지 않아요.
  verify: `cargo test --locked --offline --test profile_debug_info`
- AC-4 (REQ-1, REQ-2): 이 SPEC Result 가 before 와 after 의 빌드 폴더 크기·파일 수·`.o` 수, cold lib 시험 빌드 시간과 warm 재빌드 시간을 같은 명령·같은 조건으로 기록해요.
  verify: 이 SPEC Result 의 before/after 표

## Scope

- write: `.moai/specs/SPEC-20261121-debug-line-tables-only/spec.md`, `Cargo.toml`, `tests/profile_debug_info.rs`(새), `docs/worktree-strategy-2026-11-10.md`(실측 문장과 WT-10 행)
- read-only: `rust-toolchain.toml`, `tests/cargo_metadata_test.py`
- forbidden: 공유 `target/` 삭제, release·dist 프로필 변경, 실제 HOME, 네트워크

## Plan

1. RED: `tests/profile_debug_info.rs` 가 `Cargo.toml` 의 네 표와 dist·sha2 를 단언해요. 변경 전 끝점에서 처음 두 단언이 실패해요.
2. before 실측: 끝점의 `Cargo.toml` 로 비어 있는 새 빌드 폴더에서 cold `cargo test --lib --no-run`, `touch src/lib.rs` 뒤 warm 같은 명령, 크기·파일 수·`.o` 수(`CARGO_INCREMENTAL=0`, sccache 없음, heavy slot 안).
3. GREEN: `Cargo.toml` 에 네 표를 더해요(`[profile.dev.package.sha2]` 앞, 같은 그룹).
4. after 실측: 다른 새 빌드 폴더에서 같은 명령·같은 조건. verbose 빌드에서 플래그 확인, 빈 폴더 panic probe 로 백트레이스 확인.
5. check affected 한 번, 문서 문장 갱신.

## Risks

- `[profile.test]` 명시가 내장 상속(dev)을 깨면 test 바이너리의 debuginfo가 달라져요: AC-1 의 verbose 실측이 막아요.
- 의존성 프레임에 파일·줄이 없어져요: 의존성 안의 버그는 함수 이름까지만 보여요. 되돌림은 `package."*"` 표 두 개를 지우는 것이에요.
- 구형 rustc 로 소스 빌드하는 사용자는 `debug = "line-tables-only"` 를 거절할 수 있어요: 도구 고정이 1.94.1 이라 로컬·CI 같고, 출하 빌드는 dist(release 상속)라 무관해요.
- 측정 시점의 기계 부하가 다르면 시간이 흔들려요: 표가 측정 때의 부하를 함께 적어요.

## Result

실측일 2026-11-22(UTC). 조건은 모두 같아요: `CARGO_INCREMENTAL=0`, sccache 없음(`RUSTC_WRAPPER` 해제), heavy slot 안, rustc 1.94.1, 명령 `cargo test --locked --offline --no-run --lib`, 비어 있는 새 빌드 폴더. before 는 끝점 `90f1683f76` 의 트리(`git archive`), after 는 이 변경의 트리예요.

**A/B 번갈아 실측** (09:52:05Z–09:57:54Z, 같은 시간대에 before(B)·after(A)를 번갈아 빌드해 부하 차이를 나눠 가짐):

| 측정 | before (B) | after (A) | 차이 |
|---|---|---|---|
| cold 빌드(초, 2회) | 40, 46 (평균 43.0) | 39, 34 (평균 36.5) | −15% |
| warm 재빌드(초, `touch src/lib.rs` 뒤, 3회) | 36, 42, 29 (평균 35.7) | 27, 28, 26 (평균 27.0) | −24% |
| 빌드 폴더 크기 | 388,072 KiB | 293,780 KiB | −24% |
| 파일 수 | 1,011 | 860 | −151 |
| `.o` 파일 수 | 166 | 17 | −149 |

측정 순서는 cold B1·A1·A2·B2, warm B1·A1·A2·B2·B3·A3 이고, 각 빌드의 5분 부하는 15.7–26.6 이었어요.

**한 번씩 잰 값** (before 09:21:10Z 부하 30.98, after 09:41:34Z 부하 14.74): 크기 404,460 → 294,792 KiB, 파일 1,011 → 860, `.o` 166 → 17, `debug/deps` 370,924 → 266,516 KiB, lib 시험 바이너리 49,825,656 → 46,465,016 바이트. 이때 시간은 cold 30 → 36 초, warm 24 → 29 초로 반대로 나왔어요. 두 측정 사이에 다른 레인의 검사 두 개가 함께 돌았기 때문에, 시간 비교는 위 A/B 표를 기준으로 삼아요.

AC 별 결과:

- AC-1 met: `[profile.dev]`·`[profile.test]` 에 `debug = "line-tables-only"`, `[profile.dev.package."*"]`·`[profile.test.package."*"]` 에 `debug = false` 가 있어요. verbose 빌드(`cargo test --no-run --lib -v`)의 `moai_cli` rustc 줄은 `-C debuginfo=line-tables-only --test` 를 가지고, `cargo build -v -p serde_json` 의 `serde_json`·`itoa` rustc 줄에는 `-C debuginfo` 플래그가 없어요. `cargo test --locked --offline --test profile_debug_info`: 2 passed. 같은 시험을 끝점의 `Cargo.toml` 에 대고 돌리면 `dev_and_test_profiles_pin_line_tables_only` 가 `tests/profile_debug_info.rs:31` 에서 실패해요(RED).
- AC-2 met: `rustc -C debuginfo=line-tables-only -C split-debuginfo=unpacked -o probe probe.rs` 로 만든 두 줄짜리 panic 프로그램을 `RUST_BACKTRACE=1` 로 돌리면 `panicked at probe.rs:1:14`, 프레임 `1: inner at ./probe.rs:1:14`, `2: main at ./probe.rs:2:13` 이 나와요.
- AC-3 met: `[profile.dist]` 의 `inherits = "release"`·`lto = "thin"`, `[profile.dev.package.sha2]` 의 `opt-level = 3` 이 그대로이고 `[profile.release]` 표는 없어요(같은 시험의 `dist_profile_and_sha2_override_stay_untouched` 통과).
- AC-4 met: 위 두 표예요.

측정에 쓴 빌드 폴더는 지우지 않고 남겨 두었어요: `~/.moai/build/moai-adk-debuginfo-before`, `~/.moai/build/moai-adk-debuginfo-after`, 그리고 세션 임시 폴더의 `ab/`.
