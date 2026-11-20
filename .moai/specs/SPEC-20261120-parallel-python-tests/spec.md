---
id: SPEC-20261120-parallel-python-tests
title: "The remaining Python tests run in parallel, one process per module"
status: draft
revision: 1
created: 2026-11-20
---

# The remaining Python tests run in parallel, one process per module

## Intent

사용자 결정(2026-11-20)이에요. 제품 동작의 테스트는 Rust 로만 쓰고, 개발 도구는 `moai` 하위 명령으로 옮겨요. 그 옮김이 끝날 때까지 남은 파이썬 테스트는 CI 와 로컬 검사에서 병렬로 돌려요. 조사 결과: 파이썬 테스트 모듈이 68개이고, macOS CI 의 `rest` 다리가 그 테스트를 `unittest discover` 한 줄로 직렬 실행해서 914초를 써요. 이 SPEC 은 새 의존성 없이 표준 라이브러리만 쓰는 작은 실행기 `tools/unittest_parallel.py` 를 더해요. 실행기는 테스트 모듈마다 `python3 -m unittest discover` 프로세스를 하나씩 띄우고(작업자 풀), 모듈마다 임시 HOME 과 임시 폴더를 따로 주고, 같은 `Ran N tests` / `OK` / `FAILED` 요약을 찍고, 하나라도 실패하면 실패해요. `tools/local_check.sh` 의 `unittest` 단계가 이 실행기를 부르고, CI 는 그 스크립트를 그대로 불러서 따로 바뀌지 않아요. 새 테스트는 Rust 로 써요(실행기를 픽스처 모듈로 돌려 봐요).

## Non-goals

- 파이썬 테스트를 Rust 로 옮기는 일: 별도 작업이고, 이 SPEC 은 옮기는 동안의 실행 시간만 줄여요.
- 테스트 내용이나 기대값을 바꾸는 일: 모듈은 그대로이고 실행 방법만 바뀌어요.
- 모듈 안의 테스트를 쪼개 병렬로 돌리는 일: 단위는 모듈이에요.
- `.github/workflows/ci.yml` 바꾸기: CI 는 `tools/local_check.sh` 를 불러서 그 한 줄이 바뀌면 충분해요.

## Requirements

- REQ-1: When `tools/unittest_parallel.py` is given a start folder and a file pattern, it shall run each matching module in its own `python3 -m unittest discover` process and keep no more processes running at once than the job count. · failure: none — a test with fixture modules records how many ran at the same time.
- REQ-2: When every module passes, the tool shall print `Ran N tests in S.SSSs` with N the sum over the modules, a blank line and `OK`, and exit 0. · failure: none — a test reads the last lines and the exit code.
- REQ-3: When any module fails or errors, the tool shall print that module's own output, then `FAILED (failures=F, errors=E)` with the sums, run the remaining modules to the end and exit 1. · failure: none — a test with one failing module reads the output and the exit code.
- REQ-4: When the tool is given `-v`, it shall print each module's own verbose lines unchanged and in one block, so every test line keeps the shape `name (module.Class.name) ... ok`. · failure: none — a test reads the lines for a passing, a skipped and a failing case.
- REQ-5: The tool shall run every module with its own temporary folder as `HOME` and as `TMPDIR` and keep the caller's `CARGO_HOME` and `RUSTUP_HOME`. · failure: none — a test reads the values each fixture module saw.
- REQ-6: The tool shall take its job count from `-j`, else from `MOAI_UNITTEST_JOBS`, else the smaller of four and the machine's processor count. · failure: none — a test runs with `-j 1` and with `MOAI_UNITTEST_JOBS`.
- REQ-7: When no module matches the pattern, the tool shall print that no module matched and exit 1. · failure: none — a test runs on an empty folder.
- REQ-9: When a module file is bigger than 32 KiB and unittest lists at least two test classes for it, the tool shall run each class in its own process, so that every test runs once, and shall run the module whole when unittest cannot list its classes (an import error, or a `load_tests` hook that adds a test of another module). · failure: none — a test with a 33 KB fixture module reads the process count and the test total, and one with a hook.
- REQ-8: The `tools/local_check.sh` step `unittest` shall run the tool with `-v -s tests -p '*_test.py'`, keep the log `ci-unittest.log` and the python slot, and exit with the tool's code. · failure: none — a test reads the step line and `tests/local_check_test.py` pins the recorded command.

## Acceptance

- AC-1 (REQ-1, REQ-6): `-j 1` 에서 동시에 도는 모듈은 하나이고 `-j 3` 에서는 둘 이상이에요. 환경 변수와 `-j` 의 우선순위가 맞아요.
  verify: `cargo test --locked --offline --test unittest_parallel_tool the_job_count_bounds_the_overlap`
- AC-2 (REQ-2, REQ-3, REQ-7): 모두 통과하면 합계와 `OK` 에 exit 0 이고, 하나가 실패하면 실패한 모듈의 출력과 `FAILED (…)` 에 exit 1 이며 나머지 모듈은 끝까지 돌아요. 맞는 모듈이 없으면 exit 1 이에요.
  verify: `cargo test --locked --offline --test unittest_parallel_tool`
- AC-3 (REQ-4): `-v` 에서 테스트 한 줄의 모양이 직렬 `unittest discover -v` 와 같아서 CI 의 macOS 전용 표가 그 줄을 세요.
  verify: `cargo test --locked --offline --test unittest_parallel_tool verbose_keeps_the_per_test_line_shape`
- AC-6 (REQ-9): 32 KiB 보다 큰 모듈이 클래스마다 한 프로세스로 돌고(상속한 테스트도 한 번씩), 합계가 같아요. `load_tests` 가 다른 모듈의 테스트를 더하는 모듈은 통째로 돌아요.
  verify: `cargo test --locked --offline --test unittest_parallel_tool a_big_module_is_split_by_class`; `cargo test --locked --offline --test unittest_parallel_tool a_big_module_with_a_load_tests_hook_runs_whole`
- AC-4 (REQ-5): 모듈마다 HOME 과 TMPDIR 이 서로 다르고 호출자의 것과도 달라요.
  verify: `cargo test --locked --offline --test unittest_parallel_tool each_module_gets_its_own_home_and_tmpdir`
- AC-5 (REQ-8): `local_check.sh` 의 `unittest` 단계 줄과 `tests/local_check_test.py` 의 고정이 새 명령이에요.
  verify: `cargo test --locked --offline --test unittest_parallel_tool local_check_runs_the_tool`; `python3 -m unittest tests.local_check_test`

## Scope

- write: `.moai/specs/SPEC-20261120-parallel-python-tests/spec.md`, `tools/unittest_parallel.py`, `tools/local_check.sh`, `tests/unittest_parallel_tool.rs`, `tests/local_check_test.py`, `tests/repo_scans_test.py` (the list of tests that walk the tree names the new Rust test), `tests/github_assets_test.py` (its CI test pins the command text; one line)
- read-only: `.github/workflows/ci.yml`, `tools/repo_scans.json`
- forbidden: 실제 HOME, 사용자의 전역 git 설정, push 이외의 원격 쓰기

## Plan

1. 테스트 먼저: `tests/unittest_parallel_tool.rs` 가 임시 폴더에 픽스처 모듈(통과, 실패, 건너뜀, 겹침을 재는 느린 모듈, HOME 을 적는 모듈)을 만들고 실행기를 돌려요. 실행기가 없어서 실패(RED)하는 것을 읽어요.
2. `tools/unittest_parallel.py` 를 표준 라이브러리만으로 써요: 모듈 목록(파일 이름 순), 작업자 풀, 모듈별 임시 HOME·TMPDIR, 결과를 모듈 끝날 때 한 덩어리로 출력, 합계 요약.
3. `tools/local_check.sh` 의 `unittest` 줄과 `tests/local_check_test.py` 의 고정을 바꿔요.
4. 직렬 `unittest discover` 와 실행기를 같은 트리에서 돌려 모듈별 결과(통과·건너뜀·실패 수)가 같은지 비교하고, 걸린 시간을 SPEC Result 에 적어요. 한 모듈이 병렬에서만 실패하면(고정 경로·포트 충돌) 그 모듈의 충돌을 먼저 고쳐요.
5. READY 전 6종, `tools/repo_scans.json` 전부, 파이썬 영역 테스트, `--lib`, `--bins` 를 돌려요.

## Risks

- 병렬에서만 깨지는 모듈(고정 `/tmp` 경로, 같은 포트, 같은 캐시 폴더)이 있을 수 있어요 → 직렬과 병렬 결과를 모듈 단위로 비교하고, 충돌하는 모듈은 임시 폴더를 쓰게 고쳐요.
- 임시 HOME 이 `cargo`, `rustup`, `git` 설정을 가려 도구가 안 돌 수 있어요 → `CARGO_HOME`·`RUSTUP_HOME` 을 호출자의 값으로 못 박고, 직렬 결과와 비교해서 확인해요.
- 기계가 바쁠 때 시간이 흔들려요 → 측정 때 부하(`uptime`)를 같이 적어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 새 의존성 없이 표준 라이브러리의 스레드 풀이 모듈마다 프로세스를 띄워요 | 사용자 결정 2026-11-20: 파이썬은 옮기는 동안만; 도구를 더 들이지 않아요 | `local_check.sh` 줄을 이전 `unittest discover` 로 되돌려요 |
| D-2 | 단위는 모듈 하나이고 모듈마다 `unittest discover -p <파일>` 로 돌려요. 32 KiB 보다 큰 모듈만 클래스 단위로 쪼개요 | 직렬 실행과 같은 시작 폴더·임포트 규칙이라 결과가 같아요. 쪼개기는 측정에서 나왔어요: 모듈 단위 병렬은 가장 긴 모듈 하나(install_script_test.py, 945.8초)가 전체 시간(945.8초)을 정했어요. 목록은 unittest 가 직접 만들어서(AST 를 읽지 않아요) 상속한 테스트도 빠지거나 겹치지 않아요 | 임계값을 올리거나 쪼개기를 지워요 |
| D-3 | CI 워크플로는 바꾸지 않아요 | CI 는 `tools/local_check.sh` 를 불러요; `ci.yml` 은 다른 작업도 만져서 줄을 아껴요 | 해당 없음 |
| D-4 | 새 테스트는 Rust 로 써요 | 사용자 결정: 제품과 도구의 테스트는 Rust, 파이썬은 옮기는 동안만 | 테스트 파일을 지워요 |

## Result

(채워요: RED, GREEN, 직렬·병렬 비교와 시간, 이웃, Not run)
