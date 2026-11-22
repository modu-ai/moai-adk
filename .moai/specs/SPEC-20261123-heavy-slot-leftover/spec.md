---
id: SPEC-20261123-heavy-slot-leftover
title: "heavy-slot-acquire caps CARGO_BUILD_JOBS for every command it starts, heavy-slot-status prints the cap, and a caller removes the waiting files of dead callers"
status: draft
revision: 1
created: 2026-11-23
---

# heavy-slot-acquire caps CARGO_BUILD_JOBS for every command it starts, heavy-slot-status prints the cap, and a caller removes the waiting files of dead callers

## Intent

`SPEC-20261122-check-affected-one-slot` 이 이미 착지해서, 자리를 쥔 프로세스 안에서 다시 부르면 그 자리를 이어 쓰고(REQ-5), check affected 는 자리 하나로 cargo 단계를 모두 돌리고 `CARGO_BUILD_JOBS` 기본값을 정해요(REQ-6). 그 SPEC 의 Non-goals 가 남긴 두 가지가 아직 열려 있어요.

첫째, `heavy-slot-acquire -- cargo test` 처럼 check affected 를 거치지 않는 호출은 `CARGO_BUILD_JOBS` 를 받지 못해요. 자리 수는 지켜지지만 자리 하나 안의 rustc 가 CPU 전부를 쓸 수 있어서, 자리 3개에 16코어면 한 번에 최대 48개까지 늘어요. 둘째, 기다리다가 죽은 호출자(SIGKILL, 전원 끊김)의 기다림 파일 `waiting/<나노초>-<pid>-<max>` 이 그대로 남아요. 2026-11-23 에 실제 폴더를 재 보니 113개가 있었고 113개 모두 어떤 프로세스도 열고 있지 않았어요(`lsof -t` 가 빈 값). 줄 계산은 이 파일들을 세지 않지만, 폴더는 계속 커지고 `status` 도 이를 알려 주지 않아요.

이 SPEC 은 (1) `heavy-slot-acquire` 가 시작하는 명령 모두에 `CARGO_BUILD_JOBS` 상한을 주고, (2) `heavy-slot-status` 가 그 상한과 남은 죽은 기다림 파일 수를 보여 주고, (3) 자리를 부르는 쪽이 죽은 호출자의 기다림 파일을 지우게 해요.

## Non-goals

- `src/check_affected.rs` 는 바꾸지 않아요. 그 파일의 `default_jobs` 를 그대로 불러 쓰고, check affected 는 이미 `CARGO_BUILD_JOBS` 를 정해서 넘기므로 감싼 쪽이 그 값을 지켜요.
- 자리 파일(`slot-<숫자>`)은 지우지도 고치지도 않아요. 자리는 운영체제가 풀어 줘요.
- `status` 는 쓰지 않아요(`writes: false` 계약 그대로). 지우기는 자리를 부르는 쪽만 해요.
- Python 쪽 heavy-slot 구현은 바꾸지 않아요.
- nextest 의 `--test-threads` 상한과 자리 수의 기기 권장값은 다른 작업이에요.

## Requirements

- REQ-1: When heavy-slot-acquire starts a command and the caller's environment holds no CARGO_BUILD_JOBS, it shall set CARGO_BUILD_JOBS for that command to the larger of 2 and the logical CPU count divided by the effective slot count · failure: none — a caller's value, also the one an outer holder passed down, stays unchanged.
- REQ-2: When heavy-slot-status runs, it shall print the CARGO_BUILD_JOBS cap with the slot count and the CPU count it comes from, and the number of stale waiting files · failure: none — a limits file that cannot be used prints a null cap.
- REQ-3: When a caller asks heavy-slot for a slot, it shall remove every waiting file of another caller that no process holds locked and whose recorded pid is not alive · failure: none — a locked file, a live pid, a file that is not a plain file and a name that is not `<digits>-<digits>` or `<digits>-<digits>-<digits>` stay.
- REQ-4: When heavy-slot-status runs, it shall count the stale waiting files REQ-3 would remove without removing any · failure: none — a folder that cannot be read counts zero.

## Acceptance

- AC-1 (REQ-1): 임시 자리 폴더에서 `heavy-slot-acquire` 로 환경을 찍는 명령을 부르면 `CARGO_BUILD_JOBS` 가 규칙의 값이고, 호출자가 값을 주면 그 값이 그대로 가고, 자리 안에서 다시 부른 안쪽 명령도 바깥 값을 그대로 받아요 · verify: `cargo test --locked --offline --test heavy_slot_jobs_cleanup_cli acquire_caps_cargo_build_jobs_and_a_given_value_wins`
- AC-2 (REQ-2): `heavy-slot-status` 가 `build_jobs_cap` 에 `slots`, `cpus`, `jobs` 를 싣고, 쓸 수 없는 limits 파일이면 `null` 이에요 · verify: `cargo test --locked --offline --test heavy_slot_jobs_cleanup_cli status_prints_the_jobs_cap`
- AC-3 (REQ-3): 죽은 pid 의 잠기지 않은 기다림 파일은 다음 호출에서 지워지고, 잠긴 파일·살아 있는 pid 의 파일·심볼릭 링크·폴더·이름이 다른 파일·호출자 자신의 파일은 남아요 · verify: `cargo test --locked --offline --test heavy_slot_jobs_cleanup_cli a_caller_removes_only_dead_unlocked_waiting_files`
- AC-4 (REQ-4): `heavy-slot-status` 의 `stale_waiting` 가 지울 파일 수와 같고 파일은 하나도 지우지 않아요 · verify: `cargo test --locked --offline --test heavy_slot_jobs_cleanup_cli status_counts_stale_waiting_files_and_removes_none`

## Scope

- write: `.moai/specs/SPEC-20261123-heavy-slot-leftover/**`, `src/heavy_slot.rs`, `tests/heavy_slot_jobs_cleanup_cli.rs`(새)
- read-only: `src/check_affected.rs`(다른 카드가 쥐고 있어요), `src/main.rs`, `tests/heavy_slot_cli.rs`, `tests/heavy_slot_fifo_cli.rs`, `tests/check_affected_one_slot.rs`
- forbidden: 실제 HOME, 실제 기기의 자리 폴더, 새 Python 시험

## Plan

1. RED: `tests/heavy_slot_jobs_cleanup_cli.rs` 의 시험 4개가 오늘 팁에서 실패하는 것을 확인해요(상한 없음, `build_jobs_cap` 없음, 죽은 파일이 남음, `stale_waiting` 없음).
2. GREEN: `src/heavy_slot.rs` 의 `run_holding` 이 `CARGO_BUILD_JOBS` 를 정하고(`check_affected::default_jobs` 재사용), `hold` 가 죽은 기다림 파일을 지우고, `status` 가 상한과 `stale_waiting` 을 싣게 해요.
3. 같은 시험과 `heavy_slot_cli`·`heavy_slot_fifo_cli`·`check_affected_one_slot` 을 nextest 로 돌리고, 마지막에 bare `check affected` 를 한 번 돌려 자리 수와 시간을 재요.

## Risks

- 살아 있는 기다림이 파일을 만들고 잠그기 직전의 짧은 틈에 그 파일을 지울 수 있어요. 기다림은 파일이 사라진 것을 보면 같은 이름으로 다시 만들어요(`hold` 의 기존 동작). 호출자 자신의 이름과 살아 있는 pid 의 이름은 지우지 않아요.
- pid 가 다른 프로세스에 다시 쓰이면 죽은 호출자의 파일이 남을 수 있어요. 지우지 않는 쪽으로만 틀려요.
- 레인들은 새로 설치한 뒤에야 바뀐 동작을 봐요.
- 지우기는 파일을 지우는 동작이라 그 위험한 경로(링크, 폴더, 잠긴 파일, 낯선 이름)를 임시 폴더에서 직접 돌려 READY 에 실어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `heavy-slot-status` 는 죽은 기다림 파일을 세기만 하고 지우지 않아요. 지우기는 자리를 부르는 쪽이 해요. | `status` 의 문서와 `writes: false` 계약이 읽기 전용이에요. 카드 본문은 "status/acquire" 라고 적었어요. | 카드 열림 |
| D-2 | 상한 공식은 `check_affected::default_jobs(cpus, slots)` 를 그대로 불러 써요. | 한 곳의 공식이어야 check affected 와 감싼 호출의 값이 같아요. `src/check_affected.rs` 는 읽기만 해요. | 카드 열림 |
| D-3 | 죽은 기다림은 "누구도 잠그지 않음" 과 "이름의 pid 가 죽음" 을 둘 다 볼 때만 지워요. | 잠금만 보면 만들고 잠그기 직전의 살아 있는 기다림과 겹쳐요. pid 만 보면 pid 재사용에 속아요. | 카드 열림 |

## Result

(status draft — written when the slices land)
