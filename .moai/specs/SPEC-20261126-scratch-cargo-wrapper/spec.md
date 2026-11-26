---
id: SPEC-20261126-scratch-cargo-wrapper
title: "Scratch crate tests start cargo without the outer compiler wrapper"
status: draft
revision: 2
created: 2026-11-26
---

# Scratch crate tests start cargo without the outer compiler wrapper

## Intent

열째 묶음 넓은 검사(2026-11-26)에서 `test_home_guard` 6개와 `run_verify` 의 `real_runner_receipt` 가 `sccache: error: path must be shorter than SUN_LEN` 으로 실패했어요. 시험이 만든 scratch crate 의 cargo 가 바깥의 `RUSTC_WRAPPER` 를 물려받아서, 결과가 시험이 보는 동작이 아니라 바깥 캐시 서버의 상태에 따라 갈렸어요. 시험이 시작하는 cargo 는 컴파일러 wrapper 를 쓰지 않게 하고, 그 규칙을 한 곳에 두고 검사가 지키게 해요. 시험이 확인하는 동작은 바꾸지 않아요.

## Requirements

- REQ-1: When a test starts cargo to build a scratch crate, directly or through a `moai` command it runs, the test shall start it without `RUSTC_WRAPPER`, `CARGO_BUILD_RUSTC_WRAPPER`, `RUSTC_WORKSPACE_WRAPPER` and `CARGO_BUILD_RUSTC_WORKSPACE_WRAPPER` · failure: none — removing a variable that is not set changes nothing.
- REQ-2: When a test file under `tests/` starts cargo directly, the file shall use the one helper in `tests/common/cargo.rs`, which names those variables.
- REQ-3 (REQ-2): IF a test file outside the helper starts cargo directly, THEN the scan test shall fail and name the file and the line.
- REQ-4: The change shall keep every existing assertion of the tests it edits.

## Non-goals

- 바깥 빌드(시험 바이너리를 만드는 cargo)의 wrapper 사용은 그대로예요. sccache 를 쓰는 습관도 바꾸지 않아요.
- 제품 코드(`src/**`)와 `moai` 명령의 환경 처리는 바꾸지 않아요.
- `~/.cargo/config.toml` 로 켠 wrapper 를 끄는 일.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 변수는 빈 값으로 두지 않고 지워요. | 카드 기준 2 의 말("wrapper 변수를 지우는 것")과 이미 있는 시험 세 파일의 방식이 같아요. | 설정 파일 wrapper 까지 막아야 하면 개정으로 빈 값 방식을 검토해요. |
| D-2 | 없는 wrapper 로 다시 실행하는 시험을 `test_home_guard`, `run_verify`, `check_affected_touch` 에 하나씩 둬요. | 일회성 실행 증거만으로는 회귀를 못 막고, `run-verify` 와 `check affected` 를 거친 간접 시작은 스캔이 못 봐요. | 시험 세 개를 지워요. |

## Acceptance

- AC-1 (REQ-1): `test_home_guard`, `run_verify`, `check_affected_touch` 에 각각 시험 하나를 더해요. 그 시험은 같은 시험 바이너리를 wrapper 변수가 없는 프로그램을 가리키는 환경으로 다시 실행해 scratch crate 를 빌드하는 시험 하나가 통과하는 것을 확인해요 · verify: `cargo nextest run --test test_home_guard --test run_verify --test check_affected_touch -E 'test(/missing_outer_wrapper/)'`
- AC-2 (REQ-2, REQ-3): 도우미가 네 변수를 지우는 것을 시험이 보고, `tests/` 의 직접 cargo 시작이 도우미 밖에 있으면 스캔 시험이 파일과 줄을 말하며 실패해요(지어낸 위반 문장으로 확인) · verify: `cargo nextest run --test test_scratch_cargo`
- AC-3 (REQ-4): 고친 시험 파일과 scan registry 시험이 통과해요 · verify: `cargo nextest run --test test_home_guard --test run_verify --test audit_prep_cli --test audit_slot_cli --test worktree_slot_cli --test check_affected_touch --test repo_scans_table`
- AC-4 (REQ-1, REQ-2, REQ-3, REQ-4): SPEC 검사 오류가 없어요 · verify: `moai spec check SPEC-20261126-scratch-cargo-wrapper`

## Scope

- write: `.moai/specs/SPEC-20261126-scratch-cargo-wrapper/**`, `tests/common/cargo.rs`(새 파일), `tests/common/mod.rs`, `tests/test_scratch_cargo.rs`(새 파일), `tests/test_home_guard.rs`, `tests/run_verify.rs`, `tests/audit_prep_cli.rs`, `tests/audit_slot_cli.rs`, `tests/worktree_slot_cli.rs`, `tests/check_affected_touch.rs`(개정 2), `tools/repo_scans.json`(스캔 한 항목, line lease), `tests/areas.toml`(새 시험 파일 이름에 줄이 필요할 때만, line lease)
- write 하지 않는 것: `src/**`(제품 동작은 그대로)

## Plan

1. 지금 코드의 시험 바이너리를 없는 wrapper 환경에서 실행해 실패를 재현해요(빨강).
2. `tests/common/cargo.rs` 에 도우미를 두고 직접 시작하는 시험과 `run-verify` 를 부르는 시험이 쓰게 해요.
3. 없는 wrapper 로 다시 실행하는 시험 두 개와 스캔 시험, registry 항목을 더해요.
4. 같은 재현 실행을 다시 해 통과를 확인하고 `check affected --card` 를 한 번 돌려요.

## Risks

- 스캔은 `Command::new(…cargo…)` 와 `CARGO` 변수 읽기만 봐요. 한 줄씩 읽어서, 프로그램 이름을 변수에 담아 넘기는 모양이나 여러 줄에 걸친 호출, 셸 문자열, `moai` 명령을 거친 간접 시작은 보지 못해요. `run-verify` 와 `check_affected_touch` 의 `check affected` 길은 AC-1 의 시험이 지켜요.
- 환경 변수를 지우는 것은 `~/.cargo/config.toml` 의 `build.rustc-wrapper` 는 끄지 못해요(이 기계에는 그 파일이 없어요). 설정 파일로 wrapper 를 켠 기계는 이 변경의 범위 밖이에요.

## Result

코드 head 는 `19369fc04b`(base `91dd561a5a`)예요. 모든 실행은 2026-11-26 에 macOS, debug 빌드로 heavy slot 안에서 했어요.

첫 판(`86344f8ecd`)은 카드 검사 영수증이 PASS(490 대상, 8268 통과)였지만 감사가 FAIL 이었어요(카드 노트 10115). `tests/check_affected_touch.rs` 가 도우미 밖에서 scratch crate 를 빌드했고 `RUSTC_WRAPPER` 만 지워서 REQ-1 과 REQ-2 가 충족되지 않았어요. 그 파일은 그때 다른 카드가 잡고 있어 스캔의 예외로 두었고, 그 카드가 착지한 뒤 `19369fc04b` 가 두 시작(도구 체인의 cargo, `check affected` 가 시작하는 cargo)을 도우미로 옮기고 스캔의 예외 목록을 지웠어요.

수리의 빨강(`86344f8ecd` 에 새 시험과 예외 삭제만 더한 작업 트리): `cargo nextest run --test check_affected_touch --test test_scratch_cargo` 8 실행, 6 통과·2 실패, exit 100. 새 시험 안의 빌드가 `could not execute process` (없는 wrapper)로 실패했고, 스캔 시험이 `tests/check_affected_touch.rs:307` 을 말했어요. 고친 뒤 `check_affected_touch`, `test_scratch_cargo`, `home_isolation_scan`, `repo_scans_table`, `state_folder_literal_scan`, `next_flow_skill_scan`, `source_manifest_facts` 153 통과, exit 0. 바꾼 두 대상의 `cargo clippy ... -D warnings` 와 `cargo fmt --all -- --check` 도 exit 0.

처음의 재현(고치기 전 바이너리, `RUSTC_WRAPPER` 와 `CARGO_BUILD_RUSTC_WRAPPER` 를 없는 프로그램으로 둠): `test_home_guard` 48 통과·6 실패, `run_verify` 12 통과·1 실패(`real_runner_receipt`). 같은 바이너리를 wrapper 없이 돌리면 54·13 모두 통과였어요. 실패 수는 사고 때(6개와 1개)와 같아요.

`19369fc04b` 의 코드에서 Acceptance 의 verify 명령을 적힌 그대로 실행했어요.

- AC-1: met — `cargo nextest run --test test_home_guard --test run_verify --test check_affected_touch -E 'test(/missing_outer_wrapper/)'` 3 통과, exit 0.
- AC-2: met — `cargo nextest run --test test_scratch_cargo` 3 통과, exit 0.
- AC-3: met — `cargo nextest run --test test_home_guard --test run_verify --test audit_prep_cli --test audit_slot_cli --test worktree_slot_cli --test check_affected_touch --test repo_scans_table` 158 통과, exit 0.
- AC-4: met — `moai spec check SPEC-20261126-scratch-cargo-wrapper` errors 0, warnings 0.

첫 판의 바이너리를 없는 wrapper 환경에서 직접 실행(`aa78987a6f` 와 같은 내용의 작업 트리): `test_home_guard` 55, `run_verify` 14, `audit_slot_cli` 19, `worktree_slot_cli` 19 모두 통과. `audit_prep_cli` 는 이 직접 실행에서 10개가 실패했는데, 임시 HOME 에서 rustup 프록시가 toolchain 을 못 고른 오류이고(`rustup could not choose a version of cargo to run`) 출력에 없는 wrapper 경로는 0건이에요. nextest 에서는 AC-3 에 포함돼 통과해요.

카드 검사 결과는 READY 노트의 영수증에 있어요.

not_run: Windows·Linux, `~/.cargo/config.toml` 의 wrapper, 전체 `cargo test`, sccache 서버를 실제로 끄고 긴 `TMPDIR` 에서 재현하는 일(없는 wrapper 로 대신했어요), 지금 통합 tip 과 합친 트리의 테스트.
