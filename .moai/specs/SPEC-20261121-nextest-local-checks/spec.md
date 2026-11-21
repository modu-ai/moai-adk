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
- REQ-8: When a nextest test step's log holds lines that begin with `FLAKY`, the summary line shall show `runner=nextest` and `flaky=<count>` before `result=` · failure: none — a run without a nextest step shows neither word.
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

(구현 뒤 채워요.)
