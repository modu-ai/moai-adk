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
- REQ-4: When a caller asks heavy-slot for a slot, it shall take only a free slot of its own that is left over once every live waiter that arrived earlier has been given, in arrival order, the lowest free slot it can use · failure: none — a dead waiter's unlocked file is not counted, and a waiting file that names no `--max` can use any slot.
- REQ-5: When a process that inherited the holder's open slot file description asks for a slot in the same folder, heavy-slot shall reuse that slot without queueing · failure: none — any other descriptor, path or closed number queues as before.
- REQ-6: When check affected starts a cargo step and the caller set no CARGO_BUILD_JOBS, it shall set CARGO_BUILD_JOBS to the larger of 2 and the logical CPU count divided by the effective slot count · failure: none — a caller's value passes through unchanged.

## Acceptance

- AC-1 (REQ-1, REQ-3): 가짜 cargo 와 임시 자리 폴더로 한 번 실행하면 자리를 잡은 기록이 한 번이고, 시험 호출은 `nextest run` 한 번이며 `cargo test` 호출은 없어요 · verify: `cargo test --locked --offline --test check_affected_one_slot one_run_takes_one_slot_and_one_nextest_call`
- AC-2 (REQ-3): `cargo nextest --version` 이 실패하는 가짜 cargo 에서는 `cargo test --lib --bin moai-cli --test …` 한 번만 불러요. 두 호출 모두 `--bins` 를 쓰지 않아요 · verify: `cargo test --locked --offline --test check_affected_one_slot without_nextest_one_cargo_test_call`
- AC-3 (REQ-2): 자리가 거절되면 요약에 cargo 단계마다 그 종료 코드가 있고(`test=` 포함) cargo 는 한 번도 불리지 않아요 · verify: `cargo test --locked --offline --test check_affected_one_slot a_refused_slot_marks_every_cargo_step_and_runs_none`
- AC-4 (REQ-6): 호출자가 값을 주지 않으면 cargo 가 받은 `CARGO_BUILD_JOBS` 가 규칙의 값이고, 호출자가 주면 그 값이 그대로 가요 · verify: `cargo test --locked --offline --test check_affected_one_slot cargo_build_jobs_defaults_from_the_slot_count_and_a_caller_value_wins`
- AC-5 (REQ-4): 먼저 와서 살아 있는 기다림이 있으면 새로 온 쪽은 빈 자리가 있어도 기다리고, 둘이 기다리면 도착 순서로 시작해요 · verify: `cargo test --locked --offline --test heavy_slot_fifo_cli a_newcomer_waits_behind_an_earlier_live_waiter` and `cargo test --locked --offline --test heavy_slot_fifo_cli two_waiters_start_in_arrival_order`
- AC-7 (REQ-4): `--max` 가 섞이면 앞선 기다림은 자기가 쓸 수 있는 자리만 막아요: slot 0 이 잡혀 있을 때 `--max 1` 기다림 뒤에 온 `--max 2` 는 slot 1 을 바로 잡고, slot 0 이 풀리면 먼저 온 `--max 1` 이 나중 온 `--max 1` 보다 먼저 시작해요 · verify: `cargo test --locked --offline --test heavy_slot_fifo_cli a_wider_newcomer_takes_the_slot_a_narrower_earlier_waiter_cannot_use` and `cargo test --locked --offline --test heavy_slot_fifo_cli an_earlier_waiter_counts_only_for_the_slots_it_can_use`
- AC-6 (REQ-5): 자리 하나짜리 `heavy-slot-acquire` 안에서 check affected 를 돌려도 자기 자신을 기다리지 않고 끝나요 · verify: `cargo test --locked --offline --test check_affected_one_slot a_check_inside_a_one_slot_acquire_does_not_wait_on_itself`
- AC-8 (REQ-5): 쥔 쪽의 설명자를 물려받으면 바로 재사용하고, 다른 파일의 설명자·닫힌 번호·`slot-<숫자>` 가 아닌 이름·다른 폴더의 경로·스스로 새로 연 slot 파일은 재사용하지 않고 줄을 서요 · verify: `cargo test --locked --offline --test heavy_slot_fifo_cli reused`

## Scope

- write: `.moai/specs/SPEC-20261122-check-affected-one-slot/**`, `src/check_affected.rs`, `src/ops/check.rs`, `tests/check_affected_cli.rs`, `tests/check_affected_one_slot.rs`(새), `tests/heavy_slot_fifo_cli.rs`(새), `docs/design-review/appendix/operations/check.md`, `src/heavy_slot.rs`(리더 말로 허락), `tests/heavy_slot_cli.rs`(끝 줄 시험 하나, 리더 결정 D-6)
- read-only: `src/main.rs`, `tools/heavy_slot.py`, `.cargo/config.toml`
- forbidden: 실제 HOME, 상태 폴더 이름을 글자로 쓴 시험 기대값, 새 Python 시험

## Plan

1. 조각 1 (`src/heavy_slot.rs`, `tests/heavy_slot_fifo_cli.rs`): RED 시험 AC-5 → `hold`/`Held` 나누기(`acquire` 서명은 그대로), 도착 순서 규칙(`waiting/<나노초>-<pid>-<max>` 파일, 앞선 살아 있는 기다림마다 도착 순서로 그 기다림이 쓸 수 있는 가장 낮은 빈 자리를 내주고 남은 자기 빈 자리만 잡기), 물려받은 자리 표시 `MOAI_HEAVY_SLOT_FD=<fd>:<경로>`(같은 폴더의 `slot-<숫자>`, 장치·inode 가 같고, 누군가 쥐고 있고, 그 설명자에 대한 비차단 배타 `flock` 이 성공할 때만 재사용), `resolve`.
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
| D-7 | `heavy-slot: slot <i> taken by <name>` 줄(과 새 `reuses slot <i> it inherited` 줄)은 stdout 이 아니라 stderr 로 가요. 기다림 줄은 그대로 stdout 이에요. `src/main.rs` 의 `heavy-slot-acquire` 도움말 문장("on failure")은 이 가지가 바꾸지 않고 끝 줄 SPEC 쪽에 남겨요. | 끝 줄 시험이 감싼 명령의 stdout 이 그 명령의 것 그대로이기를 요구해요. 저장소에서 그 줄을 stdout 에서 읽는 곳은 없어요(`rg -n 'taken by'` 로 확인). `src/main.rs` 는 이 작업의 읽기 전용이에요. | 카드 열림 |
| D-5 | 시험 호출은 `--bins` 대신 `--bin moai-cli` 를 골라요. | 명시한 `--bins` 는 `test = false` 인 바이너리의 시험 하네스까지 지어요(같은 28개 시험이 `moai`·`moai-cli` 두 번, 리더 결정 2026-11-22). | 카드 열림 |
| D-8 | 기다림 파일 이름 끝에 그 기다림의 `--max` 를 적고(`<나노초>-<pid>-<max>`), 자리는 지금처럼 가장 낮은 것부터 잡아요. 앞선 기다림이 넓은(`--max` 가 큰) 쪽이면 가장 낮은 빈 자리를 가져가서, 그 뒤의 좁은 쪽이 높은 빈 자리를 못 쓰는 짧은 틈은 남아요. | 감사 B1: `--max` 가 다른 호출자가 한 폴더를 같이 써요(4, 5, 3). 이름에 적으면 파일이 생기는 순간 값이 함께 보여요. 가장 높은 자리부터 잡으면 그 틈도 없어지지만 혼자 온 호출자의 자리 번호와 Python 쌍둥이의 순서가 바뀌어요. | 카드 열림 |
| D-9 | 물려받은 자리는 같은 폴더의 `slot-<숫자>`, 같은 장치·inode, 지금 누가 쥐고 있음(새로 연 파일의 공유 시도 잠금 실패), 그 설명자에 대한 비차단 배타 `flock` 성공을 모두 볼 때만 재사용해요. 쥐고 있는지를 먼저 보고 `flock` 은 나중에 해서, 확인이 빈 자리를 잡는 일은 없어요. | 감사 B2: inode 만 보면 스스로 새로 연 slot 파일도 통과했어요. `flock` 은 열린 파일 설명에 붙으니 쥔 쪽이 물려준 설명자만 성공해요. inode 비교는 남겨요: 빼면 잠기지 않은 다른 파일의 설명자가 `flock` 을 통과해요. | 카드 열림 |

## Result

(status draft — written when the slices land)
