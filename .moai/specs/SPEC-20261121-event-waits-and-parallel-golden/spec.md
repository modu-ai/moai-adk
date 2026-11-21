---
id: SPEC-20261121-event-waits-and-parallel-golden
title: "statusline_golden runs its cases in parallel and tests wait on the clock instead of sleeping"
status: draft
revision: 1
created: 2026-11-21
---

# statusline_golden runs its cases in parallel and tests wait on the clock instead of sleeping

## Intent

무거운 테스트 전수 조사(2026-11-21, tip `b4681b3a6f`)에서 두 가지 낭비가 나왔어요. 첫째, `statusline_golden` 은 `#[test]` 하나가 파이썬 도구를 띄워 골든 케이스를 차례로 돌려서 평균 159초(최대 236초)가 걸려요. 케이스마다 61 MiB 짜리 바이너리를 임시 폴더에 복사하고 샌드박스로 띄우는데, 서로 독립이라 병렬로 돌릴 수 있어요. 둘째, 영수증·백업·색인의 시각이 초 단위라서 "다음 초가 될 때까지" 테스트가 1100 ms 를 고정으로 자는 곳이 다섯 군데 있어요. 같은 초에 두 번 설치하면 제품이 막으므로 이 기다림 자체는 필요하지만, 고정 시간 잠은 평균 0.5초면 될 일을 매번 1.1초 쓰게 하고 부하가 높을 때는 시간 한도 테스트와 겹쳐요. 이 SPEC 은 골든 하네스에 상한이 있는 병렬 실행(`--jobs`)을 더하고, 다섯 군데의 고정 잠을 "시계의 초가 바뀔 때까지 기다리기"(이벤트 대기)로 바꿔요. 검사 내용(바이트 비교, 단언)은 그대로 둬요. 모든 사용자에게 같은 말이 통해요(제품 코드는 안 바뀌고 테스트와 도구만 바뀌어요).

## Non-goals

- 제품 코드(영수증·백업·색인의 시각 처리)를 바꾸는 일: 하지 않아요. 같은 초에 두 번 설치하면 막는 것은 제품의 약속이에요.
- 골든 파일 내용, 케이스 목록, 비교 규칙을 바꾸는 일: 하지 않아요. 같은 바이트를 비교해요.
- 시간 한도가 있는 다른 테스트(500·600·1500 ms 잠)와 `atlas_three_arm`·`doc_reference_boundaries` 의 느림: 이번 조각이 아니에요.
- `tests/common/mod.rs` 를 건드리는 일: 하지 않아요. 모든 시험 crate 가 그 파일을 포함해서 바꾸면 거의 모든 시험이 다시 돌아요.
- 테스트 러너를 nextest 로 옮기는 일: 별도 SPEC 이 맡아요.

## Requirements

- REQ-1: When a test needs a later clock second than the one an earlier command ran in, the test shall wait until the Unix second changes instead of sleeping a fixed 1100 ms. · failure: none
- REQ-2: IF the clock second does not change within three seconds, THEN the wait shall stop and fail the test with a message that names the wait. · failure: none
- REQ-3: The five test files that slept 1100 ms shall hold no `from_millis(1100)` call. · failure: none
- REQ-4: When the golden harness is given `--jobs N`, it shall run its (case, variant) checks on up to N workers. · failure: none
- REQ-5: When the checks run in parallel, the harness shall print one result line per check in the order of the case manifest. · failure: none
- REQ-6: When a golden differs, the parallel run shall print the same `FAIL <case> <variant>` line and exit 1 as the serial run. · failure: none
- REQ-7: The harness shall run the native color contract and the stdin boundary checks once per run, whatever `--jobs` is. · failure: none
- REQ-8: When `--jobs` is not given, the harness shall use the spawn cap `MOAI_SPAWN_TEST_CAP`, and 4 when it is unset. · failure: none
- REQ-9: IF `--jobs` is below 1 or not a whole number, THEN the harness shall refuse before it runs any check. · failure: none
- REQ-10: The golden test shall run the harness with its default worker count, keeping its assertion on the exit status and the output. · failure: none
- REQ-11: The Result shall record the time of each changed test before and after, with the clock, the machine load and the git head of each measurement. · failure: none
- REQ-12: The change shall delete no test and reduce no assertion. · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-2): 시계 대기 도우미가 초가 바뀌면 끝나고 1.1초를 넘지 않으며, 시계가 안 바뀌면 이름이 든 메시지로 실패해요 · verify: `cargo test --locked --offline --test no_fixed_second_sleeps the_clock_wait_ends_when_the_second_changes_and_fails_when_it_does_not`
- AC-2 (REQ-3): 목록에 든 파일에 `from_millis(1100)` 이 없고 시계 대기를 불러요(목록은 조각 1에서 네 파일, 조각 2에서 `update_command_cli.rs` 가 더해져 다섯 파일) · verify: `cargo test --locked --offline --test no_fixed_second_sleeps the_listed_files_hold_no_fixed_1100_ms_sleep`
- AC-3 (REQ-1): 바꾼 시험이 모두 같은 단언으로 통과해요 · verify: `cargo test --locked --offline --test go_state_classes_cli --test index_cli --test update_command_cli --test update_moved_receipts --test update_version_folders`
- AC-4 (REQ-4, REQ-5, REQ-6, REQ-7): 작은 가짜 바이너리와 가짜 케이스로 `--jobs 4` 와 `--jobs 1` 의 출력(줄 순서)과 종료 코드가 같고, 다른 골든은 두 방식 모두 `FAIL <case> <variant>` 로 1 을 내고, 색상·입력 경계 검사는 한 번만 돌아요 · verify: `python3 -I -m unittest tests.statusline_golden_parallel_test`
- AC-5 (REQ-8, REQ-9): `--jobs` 가 없으면 `MOAI_SPAWN_TEST_CAP`(없으면 4)을 쓰고, 0·음수·글자는 아무것도 돌리기 전에 거절돼요 · verify: `python3 -I -m unittest tests.statusline_golden_parallel_test`
- AC-6 (REQ-10): 골든 시험이 실제 바이너리로 통과해요 · verify: `cargo test --locked --offline --test statusline_golden`
- AC-7 (REQ-11): Result 에 전후 시간 표가 있어요 · verify: `grep -n "전후 시간" .moai/specs/SPEC-20261121-event-waits-and-parallel-golden/spec.md`
- AC-8 (REQ-12): 바뀐 시험 파일에서 `#[test]` 개수가 같고 `weakened_in` 검사가 약화를 찾지 않아요 · verify: `moai-cli gate-check --base b4681b3a6f`

## Scope

- write: 이 SPEC, `tests/go_state_classes_cli.rs`, `tests/index_cli.rs`, `tests/update_moved_receipts.rs`, `tests/update_version_folders.rs`, `tests/clock_wait/mod.rs`(새), `tests/no_fixed_second_sleeps.rs`(새)
- 기다려요(2026-11-21 예약 읽음): `tests/update_command_cli.rs` 는 다른 카드가 쥐고 있고, `tools/statusline_golden.py` 도 다른 카드가 쥐고 있어요(그 카드의 임대는 만료됐지만 반납되지 않았어요). 골든 병렬화(조각 2: `tools/statusline_golden.py`, `tests/statusline_golden.rs`, `tests/statusline_golden_parallel_test.py`(새))와 `update_command_cli.rs` 의 잠 한 곳은 그 파일들이 풀린 뒤에 해요.
- read-only: `tests/fixtures/statusline-claude/`, `tests/common/`
- forbidden: `tests/common/mod.rs`, 골든 파일과 `cases.json`, 제품 코드

## Plan

1. 조각 1(이번): 시계 대기 도우미와 그 시험(`tests/no_fixed_second_sleeps.rs`), 풀려 있는 네 파일의 잠 교체, 전 시간 측정 기록.
2. 조각 2(파일이 풀리면): `update_command_cli.rs` 의 잠 교체, 골든 하네스 `--jobs`(ThreadPoolExecutor, 입력 순서대로 출력, 한 번만 도는 계약 검사), 골든 시험 연결, 파이썬 시험, 후 시간 측정과 Result.

## Risks

- 초가 바뀌기를 기다리는 방식은 시계가 뒤로 가면(NTP 조정) 끝나지 않을 수 있어요. 그래서 3초 한도를 두고 이름이 든 메시지로 실패해요(REQ-2).
- 파이썬 도구가 케이스를 병렬로 돌리면 임시 폴더와 61 MiB 복사가 동시에 일어나 디스크와 프로세스 수가 늘어요. 상한 기본값을 C-227 의 시험 스레드 상한(4)에 맞춰서 한 번에 4개까지만이에요(D-3).
- 병렬이면 같은 출력이라도 줄 순서가 섞일 수 있어서, 작업은 병렬로 하고 출력은 매니페스트 순서로 모아서 찍어요(REQ-5). 이 순서가 깨지면 골든 비교가 아니라 로그만 달라 보이므로 시험이 줄 순서를 고정해요.
- 측정은 기계 부하에 따라 달라서 숫자만으로는 비교할 수 없어요. 그래서 측정마다 시각·부하·head 를 같이 적고 각 3회를 넘게 재요(REQ-11).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 시각을 주입하지 않고 시계의 초가 바뀔 때까지 기다려요(조건 대기, 3초 한도) | 시각 주입 `--now-unix` 는 상태줄만 있어요. 영수증·백업·색인은 제품이 시스템 시계를 직접 읽고, `update.rs` 는 같은 초에 끝난 설치 둘을 막아요. 그래서 시험이 지켜야 하는 것은 "다음 설치는 더 늦은 초"이고, 기다림이 그 조건 자체예요. 설치 직후 호출 시점의 초가 영수증의 초보다 늦거나 같으므로, 그 초가 지나면 다음 설치는 반드시 더 늦은 초예요 | 고정 잠으로 되돌리기 |
| D-2 | 도우미는 새 폴더 `tests/clock_wait/mod.rs` 한 곳에 두고 쓰는 파일이 `mod clock_wait;` 로 불러요. `tests/common/` 에는 넣지 않아요 | `tests/common/` 아래를 바꾸면 `mod common;` 을 가진 거의 모든 시험이 다시 돌아요(`check affected` 는 `tests/<폴더>/…` 변경에 그 폴더를 `mod` 로 부르는 시험을 골라요, `src/check_affected.rs` `select`). 별도 폴더면 그 폴더를 부르는 시험만 골라져요. 시험 `no_fixed_second_sleeps.rs` 가 도우미 한 벌을 실제·가짜 시계로 직접 시험해요 | `common` 으로 모으기 |
| D-3 | 골든 병렬 작업 수의 기본값은 `MOAI_SPAWN_TEST_CAP`(없으면 4)이고 `--jobs N` 으로 바꿔요 | 시험 스레드 상한을 4 로 두는 결정(스폰 시험 세마포어)과 맞춰요. 한 시험 바이너리가 프로세스를 마음대로 늘리지 않게 하려는 거예요 | 상한 값 변경 |
| D-4 | 병렬은 파이썬 하네스 안에서 해요(ThreadPoolExecutor). 러너가 `--case` 를 나눠 여러 프로세스를 띄우는 방식은 쓰지 않아요 | `main` 이 매번 색상·입력 경계 검사를 이어서 돌려서, 나눠 띄우면 그 검사가 프로세스 수만큼 되풀이돼요(REQ-7) | 서로 다른 `--case` 로 나눠 띄우기 |

## Result

(아직 구현 전이에요.)
