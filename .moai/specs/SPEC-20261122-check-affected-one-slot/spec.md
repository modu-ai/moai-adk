---
id: SPEC-20261122-check-affected-one-slot
title: "check affected holds one heavy slot for the whole cargo part, runs one clippy and one test call, and heavy-slot serves waiters in arrival order"
status: draft
revision: 1
created: 2026-11-22
---

# check affected holds one heavy slot for the whole cargo part, runs one clippy and one test call, and heavy-slot serves waiters in arrival order

## Intent

사용자가 승인한 재설계(2026-11-22, `~/.moai/reports/moai-adk/agentic-redesign-2026-11-22.md`, 시험 검토 F1·F2, 빌드 검토 2·3)의 첫 속도 작업이에요. 지금 `moai check affected` 는 무거운 단계(clippy, test_lib, test_bins, test_targets, build_cli)마다 `heavy-slot-acquire` 를 새로 불러 한 번 실행에 4~5번 줄을 서요(기다림 198~1,879초). 시험은 `cargo test` 세 번으로 바이너리를 하나씩 돌려요. 무거운 자리의 줄은 먼저 확인한 쪽이 가져가서 도착 순서가 지켜지지 않아요(`src/heavy_slot.rs` 의 `take_slot` 이 빈 자리를 먼저 잡고, 기다리는 쪽은 200 ms 마다 다시 봐요). `CARGO_BUILD_JOBS` 기본값도 없어요.

이 SPEC 은 (1) 첫 cargo 단계에서 자리 하나를 잡아 마지막 cargo 단계까지 쥐고, (2) 시험을 한 번의 호출(`cargo nextest run`, 없으면 `cargo test`)로 돌리고, (3) heavy-slot 의 줄을 도착 순서로 만들고, (4) 이미 자리를 쥔 프로세스 안에서 다시 부르면 그 자리를 이어 쓰게 하고, (5) `CARGO_BUILD_JOBS` 기본값을 정해요.

## Non-goals

- 오래된 기다림 파일 정리와 `heavy-slot-acquire` 자체의 `CARGO_BUILD_JOBS` 기본값은 다른 작업이에요.
- nextest 의 `--test-threads` 상한은 다른 작업이에요.
- Python 쪽 heavy-slot 구현은 바꾸지 않아요(그 구현은 따로 없어져요).
- `build_cli` 규칙과 clippy 호출 내용은 바꾸지 않아요.

## Requirements

- REQ-1: When check affected reaches its first cargo step, it shall take one heavy slot and hold it until its last cargo step ends · failure: none — REQ-2 handles a refused slot.
- REQ-2 (REQ-1): IF the heavy slot is refused or the machine is too busy, THEN check affected shall record the slot's exit code for every cargo step and start no cargo.
- REQ-3: When check affected runs tests, it shall run the library tests, the `moai-cli` binary tests and the selected target tests in one call labelled `test`, with `cargo nextest run` when `cargo nextest --version` succeeds and `cargo test` otherwise · failure: none — a failed call is its exit code under `test`.
- REQ-4: When a caller asks heavy-slot for a slot, it shall start only after every live waiter that arrived earlier has started or left · failure: none — a dead waiter's unlocked file is not counted.
- REQ-5: When a process that inherited an open held slot file asks for a slot in the same folder, heavy-slot shall reuse that slot without queueing · failure: none — without the inherited file it queues as before.
- REQ-6: When check affected starts a cargo step and the caller set no CARGO_BUILD_JOBS, it shall set CARGO_BUILD_JOBS to the larger of 2 and the logical CPU count divided by the effective slot count · failure: none — a caller's value passes through unchanged.

## Acceptance

- AC-1 (REQ-1, REQ-3): 가짜 cargo 와 임시 자리 폴더로 한 번 실행하면 자리를 잡은 기록이 한 번이고, 시험 호출은 `nextest run` 한 번이며 `cargo test` 호출은 없어요 · verify: `cargo test --locked --offline --test check_affected_one_slot one_run_takes_one_slot_and_one_nextest_call`
- AC-2 (REQ-3): `cargo nextest --version` 이 실패하는 가짜 cargo 에서는 `cargo test --lib --bin moai-cli --test …` 한 번만 불러요. 두 호출 모두 `--bins` 를 쓰지 않아요 · verify: `cargo test --locked --offline --test check_affected_one_slot without_nextest_one_cargo_test_call`
- AC-3 (REQ-2): 자리가 거절되면 요약에 cargo 단계마다 그 종료 코드가 있고(`test=` 포함) cargo 는 한 번도 불리지 않아요 · verify: `cargo test --locked --offline --test check_affected_one_slot a_refused_slot_marks_every_cargo_step_and_runs_none`
- AC-4 (REQ-6): 호출자가 값을 주지 않으면 cargo 가 받은 `CARGO_BUILD_JOBS` 가 규칙의 값이고, 호출자가 주면 그 값이 그대로 가요 · verify: `cargo test --locked --offline --test check_affected_one_slot cargo_build_jobs_defaults_from_the_slot_count_and_a_caller_value_wins`
- AC-5 (REQ-4): 먼저 와서 살아 있는 기다림이 있으면 새로 온 쪽은 빈 자리가 있어도 기다리고, 둘이 기다리면 도착 순서로 시작해요 · verify: `cargo test --locked --offline --test heavy_slot_fifo_cli a_newcomer_waits_behind_an_earlier_live_waiter` and `cargo test --locked --offline --test heavy_slot_fifo_cli two_waiters_start_in_arrival_order`
- AC-6 (REQ-5): 자리 하나짜리 `heavy-slot-acquire` 안에서 check affected 를 돌려도 자기 자신을 기다리지 않고 끝나요 · verify: `cargo test --locked --offline --test check_affected_one_slot a_check_inside_a_one_slot_acquire_does_not_wait_on_itself`

## Scope

- write: `.moai/specs/SPEC-20261122-check-affected-one-slot/**`, `src/check_affected.rs`, `src/ops/check.rs`, `tests/check_affected_cli.rs`, `tests/check_affected_one_slot.rs`(새), `tests/heavy_slot_fifo_cli.rs`(새), `docs/design-review/appendix/operations/check.md`, `src/heavy_slot.rs`(리더 말로 허락), `tests/heavy_slot_cli.rs`(끝 줄 시험 하나, 리더 결정 D-6)
- read-only: `src/main.rs`, `tools/heavy_slot.py`, `.cargo/config.toml`
- forbidden: 실제 HOME, 상태 폴더 이름을 글자로 쓴 시험 기대값, 새 Python 시험

## Plan

1. 조각 1 (`src/heavy_slot.rs`, `tests/heavy_slot_fifo_cli.rs`): RED 시험 AC-5 → `hold`/`Held` 나누기(`acquire` 서명은 그대로), 도착 순서 규칙(`waiting/<나노초>-<pid>` 파일, 살아 있는 앞선 기다림의 수가 빈 자리 수보다 적을 때만 시작), 물려받은 자리 표시 `MOAI_HEAVY_SLOT_FD=<fd>:<경로>`(장치·inode 가 같을 때만 재사용), `resolve`.
2. 조각 2 (`src/check_affected.rs`, `src/ops/check.rs`, 시험, `check.md`): RED 시험 AC-1~AC-4, AC-6 → `steps(…, runner)` 에 `test` 단계 하나, nextest 확인, `run()` 에서 `hold` 한 번(기다림 줄은 `<out>/heavy-slot.log` 로, stdout 은 JSON 답), cargo 단계는 그 자리를 쥔 채 직접 실행하고 python 전에 놓기, 거절 코드를 모든 cargo 단계에, `CARGO_BUILD_JOBS` 기본값 → 기존 라벨 시험·문서·`src/ops/check.rs` 설명 갱신.

## Risks

- Python 쪽 구현은 여전히 줄을 새치기해요(따로 없어질 예정).
- 기다림 파일 이름이 벽시계 기준이라 시계가 뒤로 가면 순서가 어긋날 수 있어요.
- nextest 가 `.cargo/config.toml` 의 `[env]`(RUST_MIN_STACK)를 따르는지 첫 실제 실행에서 확인해요.
- 레인들은 새로 설치한 뒤에야 바뀐 동작을 봐요.
- 시험 바이너리를 합치는 작업이 새 시험 파일 위치를 옮길 수 있어요.
- AC-5 의 두 기다림 순서 시험은 오늘 동작에서 반쯤만 실패해요(동전 던지기); 앞선 기다림 시험은 매번 실패해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 시험 단계 라벨은 `test` 하나예요. | 세 라벨을 읽는 곳은 이 작업의 시험·문서뿐이에요(계획 조사). | 카드 열림 |
| D-2 | 자리는 마지막 cargo 단계 뒤, python 전에 놓아요. | python 단계는 cargo 를 부르지 않아요. | 카드 열림 |
| D-3 | nextest 가 없으면 `cargo test` 한 번으로 돌려요. | 시험 검토 F1-b·F2 의 대안. | 카드 열림 |
| D-4 | 호출자가 준 `CARGO_BUILD_JOBS` 가 이겨요. | 레인이 리더 규칙으로 값을 내보내요. | 카드 열림 |
| D-6 | 이 가지가 `SPEC-20261122-heavy-slot-exit-line` 의 REQ-1 도 실어요: 시작한 명령이 끝나면 일반 경로와 물려받은 자리 재사용 경로 모두 `heavy-slot: <name> exited <code>; …` 줄을 성공 때도 찍어요. 그 SPEC 의 RED 시험(`tests/heavy_slot_cli.rs` `a_started_command_prints_its_exit_line_on_success_and_on_failure`)도 같은 범위에 있어요. | 리더 결정 2026-11-22(선택지 b): 이 작업이 그 함수를 다시 써요. 그 SPEC 과 문서 줄은 이 가지가 착지한 뒤 따로 착지해요. | 카드 열림 |
| D-5 | 시험 호출은 `--bins` 대신 `--bin moai-cli` 를 골라요. | 명시한 `--bins` 는 `test = false` 인 바이너리의 시험 하네스까지 지어요(같은 28개 시험이 `moai`·`moai-cli` 두 번, 리더 결정 2026-11-22). | 카드 열림 |

## Result

(status draft — written when the slices land)
