---
id: SPEC-20261121-nextest-local-checks
title: "check affected runs its tests through nextest when it is installed"
status: draft
revision: 1
created: 2026-11-21
---

# check affected runs its tests through nextest when it is installed

## Intent

`moai check affected` 는 시험 단계(`--lib`, `--bins`, 고른 `--test` 대상)를 `cargo test` 로 돌려요. 그러면 시험 바이너리를 차례로 돌리되 바이너리 안에서는 코어 수만큼 스레드가 돌고, 프로세스를 띄우는 시험이 한꺼번에 몰리고, 멈춘 시험은 직계 자식만 죽어서 고아 프로세스가 남고, 한 번 실패했다가 다시 돌려 통과한 시험은 사람이 알아채지 못하면 그냥 지나가요. CI 는 이미 고정 버전의 `cargo-nextest` 를 쓰는데, 로컬 검사는 쓰지 않아요.

이 변경으로 `cargo-nextest` 가 있으면 시험 단계를 `cargo nextest run` 으로 돌려요. 설정(`.config/nextest.toml`)이 프로세스를 띄우는 시험의 동시 수를 묶고, 오래 걸리는 시험을 프로세스 그룹째 끝내고, 다시 돌려 통과한 시험을 FLAKY 로 표시해요. 요약 줄은 FLAKY 개수를 따로 보여 줘요. `cargo-nextest` 가 없거나 오래된 버전이면 지금과 똑같이 `cargo test` 로 돌려요.

이 명령은 어떤 MoAI 프로젝트에서도 같은 규칙으로 돌아야 해요. 통합 가지 이름이나 카드 번호를 코드에 두지 않아요. 설치 자체(`cargo install cargo-nextest --locked`)는 관리자가 한 번 하는 일이라 이 변경에 없어요.

## Non-goals

- `cargo-nextest` 를 설치하거나 버전을 올리지 않아요. 없으면 `cargo test` 로 돌아가요.
- 시험 코드나 시험 대상 선택 규칙을 바꾸지 않아요. 무엇을 돌릴지는 그대로고, 어떻게 돌릴지만 바꿔요.
- 실패했다가 통과한 시험을 실패로 만들지 않아요. `flaky-result = "fail"` 은 쓰지 않고, 결과(`result=PASS|FAIL`)는 단계 종료 코드로만 정해요.
- 스레드 상한의 기본값과 머신별 설정 파일(`limits.json`)은 만들지 않아요. 상한을 읽는 함수 하나만 두고, 값은 환경 변수 `RUST_TEST_THREADS` 와 기본 4 예요.
- 시험 바이너리 개수를 줄이거나 시험 이름을 바꾸지 않아요.
- 실행 시간의 이득을 이 SPEC 에서 주장하지 않아요. 이 Mac 에 `cargo-nextest` 가 설치돼 있지 않아서 실제 nextest 로 돌려 본 결과는 없어요(Result 의 돌리지 않은 것).

## Requirements

- REQ-1: When `cargo nextest --version` answers with version 0.9.85 or later, the runner shall run the library, binary and selected-target test steps through `cargo nextest run` · failure: none — an answer that cannot be read counts as not installed.
- REQ-2: When nextest is not installed or is older than 0.9.85, the runner shall run those test steps through `cargo test` exactly as before · failure: none — the summary line keeps the words it has today.
- REQ-3: When the runner uses nextest, each of its test steps shall pass `--test-threads` with the value of the environment variable `RUST_TEST_THREADS` when that is a positive whole number and with 4 otherwise · failure: none
- REQ-4: When the runner uses nextest, each of its test steps shall pass `--no-fail-fast` and `--no-tests=pass` · failure: none — a binary without tests is no failure and a failed test stops nothing.
- REQ-5: The project's nextest profile shall run the tests of the binaries that start the product binary through one test group of at most two threads · failure: none.
- REQ-6: The project's nextest profile shall end a test that runs past five periods of 60 seconds by stopping its whole process group · failure: none — nextest signals the group, then waits its grace period.
- REQ-7: The project's nextest profile shall retry a failed test twice so that nextest labels a test that passed on a retry FLAKY · failure: none.
- REQ-8: When the runner uses nextest, the summary line shall show `runner=nextest` and `flaky=<count>` before `result=`, the count being the tests that lines beginning with `FLAKY` in the logs of its test steps name once each · failure: none — a run without nextest shows neither word.
- REQ-9: The result of a run shall stay decided by the exit status of every step alone, whatever the FLAKY count is · failure: none.
- REQ-10: When `--list` is given, the command shall name the runner it would use and show the nextest commands it would run · failure: none — the version query is the only process it starts and it writes nothing.

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-4): 가짜 `cargo` 가 `nextest --version` 에 `cargo-nextest 0.9.151` 을 답하면 `test_lib`·`test_bins`·`test_targets` 가 `cargo nextest run --locked --offline --no-fail-fast --no-tests=pass --test-threads 4 …` 로 불리고 `RUST_TEST_THREADS=3` 이면 3, `abc`·`0`·빈 값이면 4 예요 · verify: `cargo test --locked --offline --test check_affected_cli the_test_steps_run_through_nextest_when_it_answers_a_version`
- AC-2 (REQ-2): 가짜 `cargo` 가 아무 답이 없거나(지금의 가짜 도구), `0.9.84`·`garbage` 를 답하거나, 시작할 수 없으면 시험 단계는 `cargo test --locked --offline --no-fail-fast …` 이고 요약 줄에 `runner=`·`flaky=` 가 없어요 · verify: `cargo test --locked --offline --test check_affected_cli without_a_usable_nextest_the_steps_and_the_summary_are_as_before`
- AC-3 (REQ-8, REQ-9): nextest 단계 로그에 `FLAKY 2/3 …` 가 두 줄 있으면 요약 줄이 `… python=0 runner=nextest flaky=2 result=PASS` 로 끝나고, 단계 하나가 실패하면 같은 줄이 `flaky=2 result=FAIL` 이에요. 로그에 `FLAKY` 가 없으면 `flaky=0` 이에요 · verify: `cargo test --locked --offline --test check_affected_cli flaky_tests_are_counted_in_the_summary_and_never_change_the_result`
- AC-4 (REQ-10): `--list` 의 답에 `runner` 가 `nextest` 또는 `cargo` 로 있고 단계 명령이 그 러너의 것이며, 저장소 상태를 바꾸지 않아요 · verify: `cargo test --locked --offline --test check_affected_cli list_names_the_runner_and_changes_nothing`
- AC-5 (REQ-5, REQ-6, REQ-7): `.config/nextest.toml` 을 읽으면 프로세스를 띄우는 바이너리(`home_*`, `update_*_cli`, `statusline_golden`)가 최대 2 스레드의 시험 그룹에 있고, 기본 프로필의 `slow-timeout` 이 `period = "60s"`·`terminate-after = 5`, `retries = 2` 이며, 기존 `ac_command_lint` 의 `threads-required` 규칙이 그대로예요 · verify: `cargo test --locked --offline --test nextest_local_checks_cli the_profile_groups_the_spawning_binaries_ends_slow_tests_and_retries_twice`
- AC-6 (REQ-1, REQ-2): 버전 읽기는 `cargo-nextest 0.9.151 (…)` 같은 한 줄에서 세 숫자를 읽고, 0.9.85 는 되고 0.9.84·`0.8.200`·낱말만 있는 줄·빈 줄은 안 돼요 · verify: `cargo test --locked --offline --lib check_affected::tests::the_nextest_version_floor_is_read_from_the_version_line`
- AC-7: 문서 한 줄(`docs/design-review/appendix/operations/check.md` 의 요약 줄 설명)이 새 낱말 둘을 말해요 · verify: `grep -n "runner=nextest" docs/design-review/appendix/operations/check.md`

## Scope

- write: `.moai/specs/SPEC-20261121-nextest-local-checks/**`, `.config/nextest.toml`, `src/check_affected.rs`, `tests/check_affected_cli.rs`, `tests/nextest_local_checks_cli.rs`(새), `docs/design-review/appendix/operations/check.md`(요약 줄 설명 한 줄)
- read-only: `src/heavy_slot.rs`, `.github/workflows/ci.yml`·`candidate-target.yml`(CI 가 쓰는 nextest 버전과 설정 이름), `tools/repo_scans.json`
- forbidden: 카드 번호를 SPEC id·가지·커밋·PR 에 쓰는 것, 코드에 통합 가지 이름 박기, 시험을 대신 통과로 쓰는 것, 실제 HOME, 다른 레인의 파일, `cargo-nextest` 설치

## Plan

1. 이 SPEC.
2. 설정: `.config/nextest.toml` 에 시험 그룹, `slow-timeout`, `retries` 를 더하고 `ac_command_lint` 규칙은 그대로 둬요. 설정 시험(AC-5)을 먼저 써서 빨갛게 확인해요.
3. 러너: `src/check_affected.rs` 에 버전 읽기, 러너 고르기, 시험 단계 세 개의 명령 바꾸기, 요약 낱말 둘을 더해요. 가짜 `cargo` 시험(AC-1~4, AC-6)을 먼저 써서 빨갛게 확인해요.
4. 문서 한 줄(AC-7).
5. READY 전 검사 전부, 변이 시험(≤5), Result.

## Risks

- **실제 nextest 로 돌려 보지 못했어요.** 이 Mac 에는 설치돼 있지 않아요. 가짜 `cargo` 로 우리가 부르는 인자와 요약을 시험하고 설정 키는 공식 문서(nexte.st 의 retries·slow-tests·test-groups·running 쪽, 2026-11-21 읽음)와 대조했어요. 설치한 뒤 한 번 실제로 돌려 `FLAKY`·시간 초과 표시가 문서와 같은지 보는 일이 남아요(관리자 단계).
- **FLAKY 개수는 로그 글자에 기대요.** nextest 가 `FLAKY 2/3 [ … ] <바이너리> <시험>` 줄을 찍는다는 문서의 예를 읽어요. 형식이 바뀌면 개수가 0 으로 읽혀서 FLAKY 를 놓쳐요: 그때는 로그 자체가 남아 있어요(`target/check-affected-<head>/` 의 단계 로그).
- **시간 한도가 부하 때문에 멀쩡한 시험을 끝낼 수 있어요.** 기본 한도는 300초(60초 × 5)고 `statusline_golden` 처럼 오래 걸리는 시험은 따로 더 길게 줘요. 그래도 기계가 아주 바쁘면 시간 초과가 나올 수 있고, 그 경우 nextest 가 TIMEOUT 으로 실패시켜요(숨기지 않아요).
- **`--no-tests=pass` 는 0.9.85 부터 쓸 수 있어요.** 그 아래 버전은 `cargo test` 로 돌아가는 바닥으로 정했어요(REQ-2). CI 는 0.9.151 이에요.
- **시험 스레드 상한을 정하는 변경이 아직 합쳐지지 않았어요.** 이 변경은 상한을 읽는 함수 하나(환경 변수, 기본 4)만 두고, 그 카드가 합쳐질 때 같은 함수가 머신별 값을 읽게 바뀌어요. 같은 파일을 두 카드가 고쳐서 병합 때 충돌할 수 있어요: 함수는 작고 이름이 분명해서 합치기 쉽게 했어요.
- **진짜로 실패하는 시험은 세 번 돌아요.** 재시도가 둘이라 실패가 그대로인 시험은 시간이 세 배 걸려요. 고장 난 시험이 많은 때는 검사가 그만큼 길어져요: 불안정한 시험을 드러내는 값이라고 보고 받아들였어요.
- **재시도는 시험이 실제로 불안정할 때 그 사실을 가려 줄 수 있어요.** 그래서 FLAKY 를 요약 줄에 따로 세고 READY 노트에 적게 했어요(결과를 PASS 로만 두지 않아요). 개수가 0 이 아니면 감사가 이름을 보게 돼요.
- doctest 는 지금도 `check affected` 의 단계에 없어요(`--lib` 와 `--bins` 와 `--test` 만 돌아요). 이 변경도 그 선을 그대로 둬요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | nextest 를 쓸지는 `cargo nextest --version` 한 번으로 정하고, 읽지 못하면 없는 것으로 본다 — 별도 설정 스위치를 두지 않는다. | 카드 본문(있으면 쓰고 없으면 `cargo test`), 기존 시험이 쓰는 가짜 `cargo` 는 답이 없어 자동으로 옛 길을 탄다 | 카드 열림 |
| D-2 | 요약 낱말은 nextest 를 쓸 때만 `runner=nextest flaky=<n>` 으로 `result=` 바로 앞에 더한다 — 옛 길의 요약은 글자 그대로 둔다. | 요약 줄을 읽는 리더·감사 도구(`result=PASS` 첫 줄 검사)가 깨지지 않게 | 카드 열림 |
| D-3 | FLAKY 는 결과를 바꾸지 않고 개수만 보인다 — `flaky-result = "fail"` 을 쓰지 않는다. | 카드 본문의 "FLAKY 로 기록하고 PASS 에 숨기지 않는다": 숨기지 않는 것은 요약 줄의 개수와 READY 노트이고, 결과는 종료 코드로 정한다는 기존 규칙(REQ-11)을 지킨다 | 카드 열림 |
| D-4 | 버전 바닥은 0.9.85 로 한다 — `--no-tests` 의 기본이 fail 로 바뀐 버전이라, 시험이 없는 `--bins` 가 실패하지 않게 `--no-tests=pass` 를 쓸 수 있는 첫 버전 | nexte.st/docs/running (2026-11-21 읽음) | 카드 열림 |
| D-5 | 스레드 상한은 `RUST_TEST_THREADS` 가 양의 정수면 그 값, 아니면 4 로 한다 — 시험 프로세스 수 상한을 정하는 카드가 합쳐지면 그쪽 규칙을 따른다. | 레인 규칙(`RUST_TEST_THREADS=4`), 상한 카드의 기본값 4 | 카드 열림 |
| D-6 | 프로세스를 띄우는 시험의 묶음은 바이너리 이름 글롭(`home_*`, `update_*_cli`, `statusline_golden`)으로 정한다 — 정규식 필터셋은 쓰지 않는다. | 글롭이 `binary()` 의 기본 매처고, 정규식 필터셋의 최소 버전은 문서에서 못 찾았다 | 카드 열림 |

## Result

구현 에이전트가 2026-11-21에 채웠어요. 감사가 다시 확인해요. 기준 `dcd2a2651b`(작업 시작 때의 로컬 tip), 코드 커밋 `8d0ffd25c6`, 문서·SPEC 커밋 `e7cf139078`, 이 worktree 의 `target/notes/`, macOS arm64.

빨강(코드를 쓰기 전, 같은 시험): 설정 시험은 설정에 `[test-groups]` 가 없어서 `tests/nextest_local_checks_cli.rs:32` 에서 실패(`red1.out`: 0 passed, 1 failed). 러너 시험은 새 4개가 실패하고 기존 13개가 통과(`red2.out`: 13 passed, 4 failed). 그 4개 중 `without_a_usable_nextest…` 는 처음에 가짜 `cargo` 스크립트의 문법 오류(버전 답이 빈 문자열일 때 `; exit 0`)로 모든 단계가 2 로 끝나서 실패했고, 도우미를 고친 뒤 `a_*` 와 함께 초록이 됐어요.

- AC-1: met — `cargo test --locked --offline --test check_affected_cli the_test_steps_run_through_nextest_when_it_answers_a_version` 이 포함된 `green-it.out`: check_affected_cli 17 passed, nextest_local_checks_cli 1 passed, exit 0. `test_lib`·`test_bins`·`test_targets` 가 `cargo nextest run --locked --offline --no-fail-fast --no-tests=pass --test-threads 4 …` 로 불리고 `RUST_TEST_THREADS` 가 3·12 면 그 값, `abc`·`0`·빈 값·`-2`·`2.5` 면 4, clippy·build 단계는 `cargo clippy`·`cargo build` 그대로, `cargo test --no-fail-fast` 줄이 없어요.
- AC-2: met — 같은 실행의 `without_a_usable_nextest_the_steps_and_the_summary_are_as_before`: 답 없음·`0.9.84`·`garbage`·종료 1(stderr 에만 버전)의 네 경우 모두 `cargo test --locked --offline --no-fail-fast …` 이고 요약이 `… python=0 result=PASS` 로 끝나며 `runner=`·`flaky=` 가 없고 `runner` 필드는 `cargo`.
- AC-3: met — `flaky_tests_are_counted_in_the_summary_and_never_change_the_result`: 세 시험 단계가 한 줄씩 FLAKY 를 찍으면 `… python=0 runner=nextest flaky=3 result=PASS`, 이름에 flaky 가 든 PASS 줄은 세지 않음, `fmt` 를 실패시키면 같은 `flaky=3` 에 `result=FAIL`, FLAKY 줄이 없으면 `flaky=0`.
- AC-4: met — `list_names_the_runner_and_changes_nothing`: `runner` 가 `nextest` 일 때 `test_lib` 명령이 `cargo nextest run … --test-threads 4 --lib`, 답이 없으면 `cargo` 와 `cargo test`, 저장소 상태(status·ref)와 `target/` 불변.
- AC-5: met — `cargo test --locked --offline --test nextest_local_checks_cli the_profile_groups_the_spawning_binaries_ends_slow_tests_and_retries_twice` 1 passed(`green-it.out`): 시험 그룹 하나(`max-threads = 2`)에 세 이름 글롭, 기본 `slow-timeout` 60s×5, `statusline_golden` 은 12, `retries = 2`, `flaky-result` 없음, 기존 `ac_command_lint` 의 `threads-required` 유지.
- AC-6: met — `cargo test --locked --offline --lib check_affected::` (`green-lib.out`) 15 passed, exit 0: `…the_nextest_version_floor_is_read_from_the_version_line`(0.9.85 되고 0.9.84·`0.8.200`·`0.9`·낱말만·빈 줄·첫 줄이 빈 답은 안 됨), `…the_thread_cap_reads_a_positive_whole_number_else_four`, `…flaky_tests_are_counted_once_each_whatever_else_the_log_says`, `…the_test_steps_differ_only_in_the_runner_words`.
- AC-7: met — `grep -n "runner=nextest" docs/design-review/appendix/operations/check.md` 가 24번 줄을 찍어요(exit 0).

### 검사

`moai check affected dcd2a2651b..HEAD`(자기 가지 빌드의 binary, 헤드 `e7cf139078`, `--max 2`)의 요약 줄 그대로: `check-affected base=dcd2a2651bbd head=e7cf139078a9 targets=44 py=12 scans=listed:41 diff_check=0 fmt=0 clippy=0 test_lib=0 test_bins=0 test_targets=0 build_cli=0 python=0 result=PASS` (출력 파일은 `target/notes/affected.json`·`target/notes/affected/`). `trailer-check dcd2a2651b..HEAD` → 커밋 3개 failing 0. `spec check` 오류 0·경고 0, `--closure` 는 `closure_status` 하나만 남아요(승인·완료는 레인이 쓰지 않아요).

변이 시험 5개(`target/notes/mutate.py`, 커밋된 `8d0ffd25c6` 위, 변이마다 파일을 복원하고 `restored: True` 확인) 모두 죽었어요: M1 버전 바닥을 0.9.84 로 내림 → 버전 바닥 단위 시험, M2 `--no-tests=pass` 를 뺌 → 러너 낱말 단위 시험, M3 0 스레드를 받아들임 → 스레드 상한 단위 시험, M4 같은 FLAKY 시험을 두 번 셈 → FLAKY 한 번 세기 단위 시험, M5 요약 낱말을 뺌 → FLAKY 요약 통합 시험.

### 돌리지 않은 것

- 실제 `cargo-nextest` 로 돌려 본 것: 이 Mac 에 설치돼 있지 않아요(`cargo nextest --version` 이 없다고 답해요). 가짜 `cargo` 로 우리가 부르는 인자와 요약을, 설정 키는 공식 문서(retries·slow-tests·test-groups·running 쪽, 2026-11-21 읽음)와 대조만 했어요. 설치한 뒤 한 번 실제로 돌려 `FLAKY`·`TIMEOUT` 표시와 `binary(home_*)` 글롭 묶음이 문서대로인지 보는 일이 관리자 단계로 남아요.
- 실행 시간의 이득: 재지 않았고 이 SPEC 에서 주장하지 않아요.
- 설정의 시간 한도·재시도가 실제 시험 묶음에서 멀쩡한 시험을 끝내는지: 실제 nextest 가 없어서 보지 못했어요.
- 시험 스레드 상한을 정하는 다른 변경과의 병합: 아직 합쳐지지 않았어요. 같은 파일을 고치면 병합 때 `test_thread_cap` 주변을 사람이 합쳐야 해요.
