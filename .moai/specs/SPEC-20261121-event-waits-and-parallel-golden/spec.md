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
- AC-4 (REQ-4, REQ-5, REQ-6, REQ-7): 작은 가짜 바이너리와 가짜 케이스로 `--jobs 4` 와 `--jobs 1` 의 출력(줄 순서)과 종료 코드가 같고, 다른 골든은 두 방식 모두 `FAIL <case> <variant>` 로 1 을 내고, 색상·입력 경계 검사는 한 번만 돌아요 · verify: `python3 -m unittest tests.statusline_golden_parallel_test`
- AC-5 (REQ-8, REQ-9): `--jobs` 가 없으면 `MOAI_SPAWN_TEST_CAP`(없으면 4)을 쓰고, 0·음수·글자는 아무것도 돌리기 전에 거절돼요 · verify: `python3 -m unittest tests.statusline_golden_parallel_test`
- AC-6 (REQ-10): 골든 시험이 실제 바이너리로 통과해요 · verify: `cargo test --locked --offline --test statusline_golden`
- AC-7 (REQ-11): Result 에 전후 시간 표가 있어요 · verify: `grep -n "전후 시간" .moai/specs/SPEC-20261121-event-waits-and-parallel-golden/spec.md`
- AC-8 (REQ-12): 바뀐 시험 파일에서 지워진 줄 가운데 `#[test]` 나 `assert` 가 든 줄이 없어요 · verify: `git diff b4681b3a6f..HEAD -- tests/go_state_classes_cli.rs tests/index_cli.rs tests/update_moved_receipts.rs tests/update_version_folders.rs | grep -E '^-' | grep -vE '^---' | grep -cE '#\[test\]|assert'` 가 0 을 찍어요(없으면 grep 이 1 로 끝나요)

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

조각 1(네 파일의 고정 잠을 시계 대기로)과 조각 2(골든 하네스의 병렬 실행)까지 끝났어요. 남은 것은 `update_command_cli.rs` 의 잠 한 곳뿐이고, 그 파일을 쥔 다른 카드가 돌려주면 해요. 그래서 이 Result 는 부분이에요(AC-1~8 모두 met, 다만 REQ-3 의 다섯째 파일과 AC-2 목록이 네 파일이에요).

### 이번에 한 것

- `tests/clock_wait/mod.rs`(새): `wait_for_next_second()` — 시계의 Unix 초가 바뀔 때까지 5 ms 간격으로 읽고, 3초 안에 안 바뀌면 `wait_for_next_second: the clock stayed at Unix second <n> for 3s` 로 시험을 실패시켜요.
- 네 파일(`go_state_classes_cli`, `index_cli`, `update_moved_receipts`, `update_version_folders`)의 `from_millis(1100)` 을 그 함수 호출로 바꿨어요. 단언은 하나도 바꾸지 않았고(시험 개수 71·23·8·11 그대로) 쓰지 않게 된 `thread::sleep`·`Duration` 가져오기만 지웠어요.
- `tests/no_fixed_second_sleeps.rs`(새, 시험 2개): 도우미를 실제·가짜 시계로 시험하고, 목록의 파일이 1100 ms 잠을 갖지 않고 시계 대기를 부르는지 파일 내용으로 확인해요.

### AC 결과

- AC-1 met: `cargo test --locked --offline --test no_fixed_second_sleeps` → 2 passed (실제 시계 대기가 1.1초 미만, 가짜 시계 통과, 멈춘 시계·되돌린 시계는 이름이 든 메시지로 실패).
- AC-2 met(목록은 네 파일, 다섯째는 조각 2): 같은 시험 `the_listed_files_hold_no_fixed_1100_ms_sleep` 통과.
- AC-3 met(네 파일): `cargo test --locked --offline --test go_state_classes_cli --test index_cli --test update_moved_receipts --test update_version_folders` → 71·23·8·11 passed, 실패 0. `update_command_cli` 는 조각 2.
- AC-4 met(조각 2): `python3 -m unittest tests.statusline_golden_parallel_test` → 8 OK. `--jobs 4` 와 `--jobs 1` 의 출력 줄과 순서·종료 코드가 같고, 바뀐 골든 하나는 두 방식 모두 `FAIL case-3 ansi: golden differs` 한 줄에 `checked 23 golden outputs`·종료 1, 색상·입력 경계 계약은 `--jobs` 가 1·4·16 일 때 모두 한 번씩(2회 호출)만 돌아요. 동시에 도는 수의 최대값도 `--jobs 1` 이면 1, `--jobs 4` 이면 4 로 관측했어요.
- AC-5 met(조각 2): 같은 시험 — `--jobs` 가 없으면 환경값 `MOAI_SPAWN_TEST_CAP=2` 면 2, 없으면 4, `--jobs 3` 이 환경값보다 앞서요. `--jobs=0`·`abc`·`-1`·`2.5`·빈 값과 `MOAI_SPAWN_TEST_CAP=0` 은 종료 2 와 값 출처를 적은 메시지로 거절되고 어떤 케이스도 돌리지 않아요(호출 수 0).
- AC-6 met(조각 2): `cargo test --locked --offline --test statusline_golden` → 1 passed, 3회 모두(아래 표). 시험은 `--jobs` 를 주지 않아 하네스 기본값(환경값 또는 4)이 쓰여요.
- AC-7 met(부분): 아래 전후 시간 표.
- AC-8 met(조각 1): 위 명령이 0 을 찍었어요. 지워진 줄은 모두 8줄(고정 잠 4줄, 쓰지 않게 된 가져오기 4줄)이고 `#[test]`·`assert` 줄은 없어요. 시험 개수도 그대로예요(71·23·8·11). 이 저장소 tip 에는 시험 약화 감지기가 아직 없어서 그 검사는 돌리지 않았어요.

### 전후 시간 (각 3회, `cargo test --test <대상>` 의 "finished in", 초)

측정 도구: `target/notes/measure277.py`(회마다 시각·기계 부하·git head 기록). 전: head `e0d5e47a6e`(시험 코드는 `b4681b3a6f` 와 같음), 2026-11-21T03:33Z~03:46Z, 부하(1분) 46~108. 후: head `7e252e216b`, 04:21Z~04:23Z, 부하(1분) 105~123 으로 전보다 더 바빴어요.

| 대상 | 전(3회) | 전 중앙값 | 후(3회) | 후 중앙값 |
| --- | --- | --- | --- | --- |
| go_state_classes_cli | 13.71, 14.23, 14.74 | 14.23 | 12.51, 13.48, 13.89 | 13.48 |
| index_cli | 5.10, 7.12, 20.82 | 7.12 | 5.74, 6.77, 11.41 | 6.77 |
| update_moved_receipts | 13.28, 13.95, 18.14 | 13.95 | 5.82, 6.62, 6.79 | 6.62 |
| update_version_folders | 11.72, 12.73, 18.04 | 12.73 | 8.70, 8.92, 9.11 | 8.92 |
| update_command_cli (다섯째 잠, 아직) | 19.17, 21.71, 24.46 | 21.71 | 아직 | 아직 |
| statusline_golden | 144.27, 155.89, 185.48 | 155.89 | 32.32, 29.58, 32.93 | 32.32 |

`statusline_golden` 의 후 측정은 커밋된 head `ed6cb08c6d`, 2026-11-21T04:35Z~04:37Z, 부하(1분) 36~80 에서 쟀고, 전은 부하 46~108 이었어요. 같은 하네스를 커밋 전 작업본(head `43bf7b2b7d` 위에 고친 파일)으로 먼저 잰 값(42.11, 42.02, 39.48초, 부하 35~76)도 같은 방향이었어요. 목표(보고서 제안) 40초 이하를 커밋된 head 의 3회 모두 만족해요.

읽는 법: 부하가 달라서 이 숫자는 정확한 비교가 아니에요. 줄어든 몫은 잠을 부른 횟수와 같아요(`update_moved_receipts` 는 `install` 이 11번이라 한 번에 평균 0.6초씩, `update_version_folders` 는 18번 호출). `go_state_classes_cli` 의 잠은 4번뿐이라 차이가 부하 흔들림 안에 있어요. 부하가 더 높았는데도 네 대상 모두 중앙값이 줄었어요.

### 돌리지 않은 것

- `tests/update_command_cli.rs` 의 `confirm()` 잠 한 곳과 그 후 측정: 다른 카드가 그 파일을 쥐고 있어서(2026-11-21T04:36Z 의 `card show` 로 확인) 아직 안 했어요. 그래서 REQ-3 의 "다섯 파일" 중 네 파일만 닫혔고, AC-2 의 시험 목록도 네 파일이에요. 파일이 풀리면 한 줄 교체와 시험 목록 한 줄·측정이 남아요.
- 후 측정은 부하가 흔들리는 때 한 번(3회씩)이라 같은 부하에서 다시 재지 않았어요.
- 병렬 하네스를 macOS 가 아닌 곳(샌드박스 없는 Linux)에서는 돌리지 않았어요(이 Mac 에서만).

### 검사 (head `7e252e216b` 기준, 전부 줄이지 않고)

diff-check·fmt·ops-rustfmt·clippy(`--all-targets -D warnings`)·`--lib`·`--bins`·영역 시험 31개·저장소 스캔 시험 4개·macOS 시험 6개 exit 0, `tools/repo_scans.json` 검사 41개 중 실패 0, 파이썬 시험 129개 OK, trailer-check exit 0. 뮤턴트 5개 모두 잡힘(`tests/clock_wait/mod.rs` 3개, 두 파일의 시계 대기 제거 2개 — 후자는 파일 내용 시험이 잡아요).
