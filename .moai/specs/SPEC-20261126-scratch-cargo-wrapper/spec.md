---
id: SPEC-20261126-scratch-cargo-wrapper
title: "Scratch crate tests start cargo without the outer compiler wrapper"
status: draft
revision: 1
created: 2026-11-26
---

# Scratch crate tests start cargo without the outer compiler wrapper

## Intent

열째 묶음 넓은 검사(2026-11-26)에서 `test_home_guard` 6개와 `run_verify` 의 `real_runner_receipt` 가 `sccache: error: path must be shorter than SUN_LEN` 으로 실패했어요. 시험이 만든 scratch crate 의 cargo 가 바깥의 `RUSTC_WRAPPER` 를 물려받아서, 결과가 시험이 보는 동작이 아니라 바깥 캐시 서버의 상태에 따라 갈렸어요. 시험이 시작하는 cargo 는 컴파일러 wrapper 를 쓰지 않게 하고, 그 규칙을 한 곳에 두고 검사가 지키게 해요. 시험이 확인하는 동작은 바꾸지 않아요.

## Requirements

- REQ-1: When a test starts cargo to build a scratch crate, directly or through a `moai` command it runs, the test shall start it without `RUSTC_WRAPPER`, `CARGO_BUILD_RUSTC_WRAPPER`, `RUSTC_WORKSPACE_WRAPPER` and `CARGO_BUILD_RUSTC_WORKSPACE_WRAPPER` · failure: none — removing a variable that is not set changes nothing.
- REQ-2: When a test file under `tests/` starts cargo directly, the file shall use the one helper in `tests/common/cargo.rs`, which names those variables.
- REQ-3 (REQ-2): IF a test file outside the helper starts cargo directly and the scan does not list it as an exception with its reason, THEN the scan test shall fail and name the file and the line.
- REQ-4: The change shall keep every existing assertion of the tests it edits.

## Non-goals

- 바깥 빌드(시험 바이너리를 만드는 cargo)의 wrapper 사용은 그대로예요. sccache 를 쓰는 습관도 바꾸지 않아요.
- 제품 코드(`src/**`)와 `moai` 명령의 환경 처리는 바꾸지 않아요.
- `~/.cargo/config.toml` 로 켠 wrapper 를 끄는 일.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 변수는 빈 값으로 두지 않고 지워요. | 카드 기준 2 의 말("wrapper 변수를 지우는 것")과 이미 있는 시험 세 파일의 방식이 같아요. | 설정 파일 wrapper 까지 막아야 하면 개정으로 빈 값 방식을 검토해요. |
| D-2 | 없는 wrapper 로 다시 실행하는 시험을 `test_home_guard` 와 `run_verify` 에 하나씩 둬요. | 일회성 실행 증거만으로는 회귀를 못 막고, `run-verify` 를 거친 간접 시작은 스캔이 못 봐요. | 시험 두 개를 지워요. |

## Acceptance

- AC-1 (REQ-1): `test_home_guard` 와 `run_verify` 에 각각 시험 하나를 더해요. 그 시험은 같은 시험 바이너리를 `RUSTC_WRAPPER` 와 `CARGO_BUILD_RUSTC_WRAPPER` 가 없는 프로그램을 가리키는 환경으로 다시 실행해 scratch crate 를 빌드하는 시험 하나가 통과하는 것을 확인해요 · verify: `cargo nextest run --test test_home_guard --test run_verify -E 'test(/missing_outer_wrapper/)'`
- AC-2 (REQ-2, REQ-3): 도우미가 네 변수를 지우는 것을 시험이 보고, `tests/` 의 직접 cargo 시작이 도우미 밖에 있으면 스캔 시험이 파일과 줄을 말하며 실패해요(지어낸 위반 문장으로 확인) · verify: `cargo nextest run --test test_scratch_cargo`
- AC-3 (REQ-4): 고친 시험 파일과 scan registry 시험이 통과해요 · verify: `cargo nextest run --test test_home_guard --test run_verify --test audit_prep_cli --test audit_slot_cli --test worktree_slot_cli --test repo_scans_table`
- AC-4 (REQ-1, REQ-2, REQ-3, REQ-4): SPEC 검사 오류가 없어요 · verify: `moai spec check SPEC-20261126-scratch-cargo-wrapper`

## Scope

- write: `.moai/specs/SPEC-20261126-scratch-cargo-wrapper/**`, `tests/common/cargo.rs`(새 파일), `tests/common/mod.rs`, `tests/test_scratch_cargo.rs`(새 파일), `tests/test_home_guard.rs`, `tests/run_verify.rs`, `tests/audit_prep_cli.rs`, `tests/audit_slot_cli.rs`, `tests/worktree_slot_cli.rs`, `tools/repo_scans.json`(스캔 한 항목, line lease), `tests/areas.toml`(새 시험 파일 이름에 줄이 필요할 때만, line lease)
- write 하지 않는 것: `src/**`(제품 동작은 그대로), `tests/check_affected_*.rs`(다른 카드가 고치는 중)

## Plan

1. 지금 코드의 시험 바이너리를 없는 wrapper 환경에서 실행해 실패를 재현해요(빨강).
2. `tests/common/cargo.rs` 에 도우미를 두고 직접 시작하는 시험과 `run-verify` 를 부르는 시험이 쓰게 해요.
3. 없는 wrapper 로 다시 실행하는 시험 두 개와 스캔 시험, registry 항목을 더해요.
4. 같은 재현 실행을 다시 해 통과를 확인하고 `check affected --card` 를 한 번 돌려요.

## Risks

- `tests/check_affected_touch.rs` 는 도구 체인의 cargo 를 직접 시작하고 `RUSTC_WRAPPER` 만 지워요. 그 파일은 다른 카드가 고치는 중이라 스캔의 예외 목록에 이유와 함께 두고, 옮기는 일은 그 카드가 착지한 뒤의 몫이에요.
- 스캔은 `Command::new(…cargo…)` 와 `CARGO` 변수 읽기만 봐요. 셸 문자열이나 `moai` 명령을 거친 간접 시작은 보지 못해요. `run-verify` 의 길은 AC-1 의 시험이 지켜요.
- 환경 변수를 지우는 것은 `~/.cargo/config.toml` 의 `build.rustc-wrapper` 는 끄지 못해요(이 기계에는 그 파일이 없어요). 설정 파일로 wrapper 를 켠 기계는 이 변경의 범위 밖이에요.

## Result

아직 없어요.
