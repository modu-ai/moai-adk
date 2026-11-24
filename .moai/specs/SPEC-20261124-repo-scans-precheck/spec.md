---
id: SPEC-20261124-repo-scans-precheck
title: "check affected stops before the heavy steps when an added test file is missing from the repository scan list"
status: approved
revision: 1
created: 2026-11-24
approval:
  by: leader
  at: 2026-11-24
  basis: "리더 결정 노트 8150 (2026-11-24T03:27:05Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단, 사용자 직접 선택 아님, 노트 8148 에 답함): 판 1 승인 선택지 (1) 기본(커밋 8fb9f23ea6d3b72250eb6d5c155e9fa91242bd9d); 판정을 따로 구현하되 AC-6 의 저장소 전체 대조가 두 판정의 어긋남을 잡는지를 뮤턴트로 보이고, tests/repo_scans_table.rs 의 새 테스트는 표를 느슨하게 하지 않는 추가만; 레인은 그 노트를 인용만 해요"
  digest: "sha256:2be56d7326c7b8fb2dbd00fb4ff107d1de9ae7966c357bea125d349971336b4e"
---

# check affected stops before the heavy steps when an added test file is missing from the repository scan list

## Intent

재발 방지예요. `SPEC-20261124-new-test-file-precheck` 가 만든 가벼운 단계 `test_tables` 는 새 최상위 `tests/*.rs` 파일을 두 표(`tests/areas.toml`, 상태 폴더 allow 목록)에 맞춰 보고 무거운 단계 앞에서 멈춰요. 그런데 같은 종류의 표가 하나 더 있고 그 단계는 보지 않아요: `tools/repo_scans.json`(저장소 스캔 목록)이에요. 새 테스트 파일이 저장소 트리를 걷거나(`CARGO_MANIFEST_DIR` 와 `read_dir(` 같은 호출), `include_str!`/`include_bytes!` 로 `tests/` 밖의 파일을 읽거나, 체크인된 문서의 경로를 문자열로 적고 열면, `tests/repo_scans_table.rs` 의 세 테스트가 그 파일을 목록에 등록하거나 예외(`EXEMPT`, `INCLUDE_STR_EXEMPT`, `FILE_READ_EXEMPT`)에 적으라고 요구해요. 이 요구는 무거운 테스트 단계에서야 만나요. 앞 SPEC 의 첫 구현 head 가 clippy 와 `repo_scans_table` 두 테스트에서 실패해 판 2 와 판 3 을 낳았고, 그 SPEC 의 Result 가 "세 번째 표는 단계가 보지 않는다" 를 빈틈으로 적었어요.

이 SPEC 은 같은 단계 `test_tables` 에 세 번째 표의 판정을 더해요. 새 단계도 새 표도 만들지 않고, 어긋나면 줄 하나가 "어디에 무엇을 더하라" 를 그대로 말해요. 표의 규칙은 풀지 않아요.

카드 C-405 감사의 비차단 후속 셋이 이 카드에 함께 넘어왔어요(카드 노트 7885): `FILE_READ_EXEMPT` 의 쓸모없는 두 행, `include_bytes!` 를 빼먹은 문구 둘, `tools` 리터럴 조건의 단위 테스트. 셋 다 표를 조이는 일이고 푸는 일이 아니에요.

읽는 사람: 레인, 리더, 이 변경의 감사자.

## Interview

- Q1: 세 번째 표의 판정을 `tests/repo_scans_table.rs` 와 하나로 합치는가? -> A: 아니요 (D-1). 따로 구현하고 AC-6 의 저장소 전체 대조가 어긋남을 잡아요.
- Q2: 막는 검사는 먼저 observe 로 도는가? -> A: 아니요 (D-2). 새 규칙이 아니라 이미 막는 규칙을 앞으로 당겨요.
- Q3: 예외 표는 어디서 읽는가? -> A: `tests/repo_scans_table.rs` 글에서 읽어요 (D-3).

## Non-goals

- `tools/repo_scans.json` 이나 `tests/repo_scans_table.rs` 의 규칙을 풀지 않아요. 이 단계는 읽기만 하고, 파일을 고치지 않아요(후속 셋의 조이기는 REQ-10 이하이고 규칙을 늘려요, 줄이지 않아요).
- 파이썬 테스트(`*_test.py`)와 `src/**` 의 걷는 파일은 이 단계가 보지 않아요: 앞 SPEC 처럼 "더한 최상위 `tests/*.rs`" 만이에요(그 밖은 `repo_scans_table` 이 그대로 잡아요).
- 스캔 항목이 실제 테스트 함수를 가리키는지(`every_check_names_a_real_target_and_test`)와 이미 있는 항목끼리의 어긋남은 이 단계가 보지 않아요: 새 파일이 아니라 항목을 고치는 일이에요.
- `--lane` 선택 규칙과 무거운 단계의 순서는 바꾸지 않아요. 단계는 앞 SPEC 의 자리(`fmt` 뒤, 첫 무거운 단계 앞)에 그대로 있어요.
- `test_tables` 가 부르는 판정 함수를 `tests/repo_scans_table.rs` 와 하나로 합치지 않아요(D-1).

## Requirements

<!-- This SPEC adds no new rule: it pulls three rules that tests/repo_scans_table.rs already enforces forward to the cheap step, so it has no observe stage (D-2). -->
- REQ-1: When a range adds a top-level `tests/<name>.rs` that names `CARGO_MANIFEST_DIR` and calls one of the walk calls of `tests/repo_scans_table.rs`, and neither `tools/repo_scans.json` names the file as a rule or test file nor `EXEMPT` lists it, `check affected` shall fail the cheap step `test_tables` with a line that names the file and both places to register it · failure: none — the line is the only output of the step.
- REQ-2: When a range adds a top-level `tests/<name>.rs` that reads a file outside `tests/` with `include_str!` or `include_bytes!`, and no `always` scan names the file as a rule file with that path in its scope and `INCLUDE_STR_EXEMPT` lacks the file, `check affected` shall fail the cheap step `test_tables` with a line that names the file, the path and the entry to add · failure: none — the line is the only output of the step.
- REQ-3: When a range adds a top-level `tests/<name>.rs` that names `CARGO_MANIFEST_DIR` and spells a string literal that is a tracked path outside `tests/`, and no `always` scan names the file as a rule file with that path in its scope and `FILE_READ_EXEMPT` does not keep the path out, `check affected` shall fail the cheap step `test_tables` with a line that names the file, the path and the entry to add · failure: none — the line is the only output of the step.
- REQ-4: When a range adds a top-level `tests/<name>.rs` that an `always` scan names as a rule file and that scan runs no check with `--test <name>`, `check affected` shall fail the cheap step `test_tables` with a line that names the scan and the check to add · failure: none — the line is the only output of the step.
- REQ-5: When an added top-level `tests/<name>.rs` is registered or exempt under REQ-1 to REQ-4, `check affected` shall add no line for it from those rules · failure: none — a registered file is the passing case.
- REQ-6: When a range adds a top-level `tests/*.rs` file and the repository holds `tools/repo_scans.json` but neither of the two earlier tables, `check affected` shall plan the step `test_tables` · failure: none — the rule of the earlier SPEC (REQ-6: no added test file or none of the tables plans no step) stays as it was, with the scan list counted as a table.
- REQ-7: When a range adds a top-level `tests/<name>.rs` and `tools/repo_scans.json` exists but cannot be read, or its text is not JSON with a `scans` list, `check affected` shall fail the cheap step `test_tables` with a line that names the file and the error · failure: none — the line is the only output of the step.
- REQ-8: When the repository's own top-level `tests/*.rs` files are all taken as added, the three-table verdict shall find no violation · failure: none — a difference is a drift between this step and `tests/repo_scans_table.rs`.
- REQ-9: The operation page of `check` and the Tests section of `docs/lane-rules.md` shall say that a new test file that walks the tree or reads a checked-in file registers in `tools/repo_scans.json` or in the exemption lists of `tests/repo_scans_table.rs`, and that `test_tables` stops the run on it · failure: none — a test reads both texts.
- REQ-10: `FILE_READ_EXEMPT` shall hold no path whose removal leaves the reader check passing · failure: none — a test names the unneeded `file path` pair.
- REQ-11: The failure texts of the two `include_str!` reader tests of `tests/repo_scans_table.rs` shall name `include_bytes!` as well · failure: none — a test reads the texts.
- REQ-12: `file_read_paths` shall list the folder `tools` only when a test file spells an interpreter literal, the literal `tools` and a tracked `tools` folder together · failure: none — a unit test flips each of the three conditions.

## Acceptance

- AC-1 (REQ-1): 임시 저장소 범위가 `CARGO_MANIFEST_DIR` 와 `read_dir(` 를 적은 새 `tests/zzwalk_cli.rs` 를 더하고 `tools/repo_scans.json` 에도 `EXEMPT` 에도 없으면 요약 줄이 `test_tables=1 stopped_at=test_tables result=FAIL` 이고 가짜 clippy·테스트 호출 흔적이 없고, 단계 로그가 그 파일과 `tools/repo_scans.json` 과 `EXEMPT` 를 말해요 · verify: `cargo nextest run --locked --offline --test check_affected_scans_cli a_new_walking_test_file_missing_from_the_scan_list_stops_before_the_heavy_steps`
- AC-2 (REQ-2, REQ-3): `include_str!("../src/x.rs")` 를 적은 새 파일과 체크인된 문서 경로 리터럴을 적고 여는 새 파일이 각각 한 줄을 받고, 줄이 파일과 경로를 말해요; `FILE_READ_EXEMPT` 가 그 경로를 빼 주는 파일은 줄이 없어요 · verify: `cargo nextest run --locked --offline --test check_affected_scans_cli a_new_reader_without_an_always_scan_stops_before_the_heavy_steps`
- AC-3 (REQ-4, REQ-5, REQ-6): 항목의 규칙 파일에 이름이 있고 `--test <이름>` 검사가 있는 새 파일, `EXEMPT` 에 있는 새 파일은 `test_tables=0` 으로 지나가 무거운 단계가 돌고; 규칙 파일에 이름이 있는데 `--test <이름>` 검사가 없으면 그 검사를 말하며 멈추고; 두 표 파일이 없고 스캔 목록만 있는 저장소도 새 테스트 파일에 대해 단계를 계획하고; 세 표가 모두 없는 저장소와 테스트 파일을 더하지 않은 범위의 요약 줄에는 `test_tables` 낱말이 없어요 · verify: `cargo nextest run --locked --offline --test check_affected_scans_cli a_registered_or_exempt_new_test_file_passes_and_the_step_is_planned_only_with_a_table`
- AC-4 (REQ-7): 스캔 목록이 잘못된 UTF-8 이거나 JSON 이 아니거나 `scans` 목록이 없으면 단계가 그 표 이름과 오류를 한 줄로 말하고 멈춰요; 읽을 수 있는 같은 범위는 그 줄이 없어요 · verify: `cargo nextest run --locked --offline --test check_affected_scans_cli a_scan_list_that_cannot_be_read_stops_before_the_heavy_steps`
- AC-5 (REQ-1…REQ-5): 순수 판정 함수 단위 테스트가 위반 종류별 줄(걷는 파일, `include_str!`/`include_bytes!` 읽기, 파일 읽기, 검사 없음)과 위반 없음, 그리고 세 `EXEMPT` 표의 읽기를 가려요 · verify: `cargo nextest run --locked --offline --lib check_affected::tests::scan_violations`
- AC-6 (REQ-8): 저장소의 실제 `tests/*.rs` 전체를 더한 파일로 놓고 세 표를 판정하면 위반이 0 개예요 · verify: `cargo nextest run --locked --offline --lib check_affected::tests::the_repository_tests_pass_the_scan_step_as_added_files`
- AC-7 (REQ-9): `check.md` 와 `docs/lane-rules.md` 가 `tools/repo_scans.json` 과 `EXEMPT` 를 `test_tables` 와 함께 말해요 · verify: `cargo nextest run --locked --offline --lib check_affected::tests::the_guidance_says_where_a_new_test_file_registers`
- AC-8 (REQ-10): `FILE_READ_EXEMPT` 의 각 `file path` 쌍을 빼면 `every_file_reader_has_an_always_scan_or_an_exemption` 이 실패해요; 쓸모없는 쌍은 지워지고, 새 테스트가 지켜 줘요 · verify: `cargo nextest run --locked --offline --test repo_scans_table every_file_read_exemption_is_needed`
- AC-9 (REQ-11): 두 `include_str!` 읽기 테스트의 실패 문구가 `include_bytes!` 를 말해요 · verify: `cargo nextest run --locked --offline --test repo_scans_table the_include_str_reader_texts_name_include_bytes`
- AC-10 (REQ-12): 단위 테스트가 인터프리터 리터럴, 리터럴 `tools`, 추적되는 `tools` 폴더 중 하나씩만 없앨 때 `tools` 가 나오지 않고 셋이 있을 때만 나오는 것을 가려요 · verify: `cargo nextest run --locked --offline --test repo_scans_table file_read_paths_adds_the_tools_folder_only_with_all_three_conditions`

## Scope

- write: `.moai/specs/SPEC-20261124-repo-scans-precheck/**`, `src/check_affected.rs`, `tests/check_affected_scans_cli.rs`(새), `tests/repo_scans_table.rs`, `docs/design-review/appendix/operations/check.md`, `docs/lane-rules.md`
- write, by line lease (shared files): `tools/repo_scans.json`(항목 하나 `table-step-matches-the-two-repository-tests` 에 테스트·검사·규칙 파일·범위 줄만, D-7)
- read-only: `tests/areas.toml`(새 파일의 첫 낱말 `check` 는 이미 `spec` 영역에 있어서 줄을 더하지 않아요), `tests/fixtures/state_folder_literals/allow.tsv`, `tests/check_affected_tables_cli.rs`, `SPEC-20261124-new-test-file-precheck`, `SPEC-20261123-land-repo-scans`
- forbidden: `tests/repo_scans_table.rs` 규칙(걷기·읽기 판정, `RUST_WALK`, 스캔 항목 모양)을 풀거나 줄이는 것, `src/lib.rs`, 새 Python 테스트, 카드 번호를 SPEC id·가지·커밋·PR 에 쓰는 것, 실제 HOME, 다른 레인의 파일

## Plan

1. RED: `tests/check_affected_scans_cli.rs` 의 테스트 넷, `check_affected::tests::scan_violations*` 단위 테스트, `tests/repo_scans_table.rs` 의 새 테스트 셋(`every_file_read_exemption_is_needed`, `the_include_str_reader_texts_name_include_bytes`, `file_read_paths_adds_the_tools_folder_only_with_all_three_conditions`)을 먼저 쓰고 코드 없이 빨간 것을 확인해요(저장소 전체 대조 테스트는 빈 판정에서 통과하는 것이 정상이에요).
2. GREEN, `src/check_affected.rs`: `Facts` 에 추적 경로(`git ls-files -z`)와 스캔 목록 읽기 오류를 더하고, 순수 함수 `scan_violations`(걷기·`include_str!`·파일 읽기 판정, 세 `EXEMPT` 표 읽기, `--test` 검사 대조)를 더해 `table_step` 이 `table_violations` 와 함께 불러요. 계획 조건에 스캔 목록을 더해요(REQ-6).
3. `tests/repo_scans_table.rs`: 후속 셋의 조이기.
4. 글: `check.md` 와 `docs/lane-rules.md` 의 `test_tables` 문단에 세 번째 표.
5. `tools/repo_scans.json`: `moai card line-lease` 로 항목 하나의 줄만 고쳐요.
6. 뮤턴트: 각 새 가드 테스트마다 코드를 지우고 실패를 봐요. 영향 테스트 한 번, READY.

## Risks

- 단계가 세는 규칙이 `tests/repo_scans_table.rs` 와 어긋날 수 있어요: 걷기 판정, `include_str!` 경로 해석, 문자열 리터럴 읽기, `Tracked::repo_path` 를 두 곳이 가져요. AC-6 이 한쪽만 바뀐 드리프트를 잡아요(테스트 파일 전체를 더한 파일로 놓고 두 테스트가 이미 통과하는 트리에서 위반이 나오면 어긋난 거예요).
- 세 `EXEMPT` 표는 테스트 파일 안의 Rust 상수라서 단계가 글로 읽어요. 읽기가 어긋나 예외를 놓치면 오탐(막힘)이고, 예외를 지어내면 미탐이에요: AC-5 가 세 표의 읽기를, AC-6 이 실제 표의 모든 행을 가려요.
- `git ls-files` 를 읽을 수 없으면 추적 경로를 구할 수 없어요: 그때는 파일 읽기 판정(REQ-3)만 건너뛰고 로그에 한 줄로 그 사실을 말해요(건너뛴 판정을 통과로 세지 않아요). 범위를 읽은 `git diff` 가 이미 같은 저장소에서 돌았으니 드문 경우예요.
- 요약 줄의 `test_tables` 낱말이 스캔 목록만 있는 저장소의 새 테스트 파일에도 생겨요(REQ-6): 낱말 순서를 읽는 곳은 연산 페이지와 `check_affected*` 테스트뿐이고, 영향 테스트 실행이 그 테스트들을 돌려요.
- 후속 REQ-10 이 쓸모없다고 본 두 행을 지우면 그 두 파일이 뒤의 테스트에서 빨개질 수 있어요: RED 에서 행마다 지우고 읽기 테스트를 돌려 측정하고, 빨개지는 쌍은 남겨요(카드 노트 7885 는 뮤턴트 C1·C2 exit 0 을 말하지만 이 레인이 다시 재요).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 세 번째 표의 판정을 `src/check_affected.rs` 에 따로 구현하고 `tests/repo_scans_table.rs` 와 합치지 않아요; 어긋남은 AC-6 이 잡아요 | 앞 SPEC 의 선례(`table_violations` 와 두 표 테스트). 합치려면 판정을 `src/` 모듈로 옮기고 `src/lib.rs` 에 줄을 더해야 하고, 게이트 테스트의 코드를 바꾸는 일이라 이 카드의 "표를 풀지 않는다" 와 맞지 않아요 | 판정을 공유 모듈로 옮기는 후속 카드 |
| D-2 | observe 단계 없이 처음부터 막아요 | 새 규칙이 아니라 `repo_scans_table` 의 세 테스트가 이미 막는 규칙을 앞으로 당긴 것이고 오탐은 AC-6 으로 재요 | 단계를 계획에서 빼기 |
| D-3 | 세 `EXEMPT` 표(`EXEMPT`, `INCLUDE_STR_EXEMPT`, `FILE_READ_EXEMPT`)는 `tests/repo_scans_table.rs` 글에서 읽어요; 그 파일이 없으면 예외가 없는 것으로 봐요 | 표가 그 파일에만 있어요. 이 단계는 읽기만 해요 | 표를 `tools/` 의 데이터 파일로 옮기는 후속 카드 |
| D-4 | 파일 읽기 판정(REQ-3)의 추적 경로는 `git ls-files -z` 이고 단계를 계획할 때만(더한 테스트 파일이 있을 때만) 읽어요 | `repo_scans_table` 의 `Tracked::of_repo` 와 같은 길이고 다른 실행의 비용·요약 줄이 그대로예요 | 늘 읽기 |
| D-5 | REQ-4(`--test <이름>` 검사)를 넣어요 | 항목에 등록하고 검사를 빼먹는 것이 같은 종류의 무거운 단계 실패예요(`every_always_scan_runs_the_readers_it_names`) | 규칙 하나 빼기 |
| D-6 | 파이썬 테스트와 `src/**` 의 걷는 파일은 이 단계가 보지 않아요 | 앞 SPEC 의 범위 결정과 같아요(테스트는 Rust 만이에요) | 없음 |
| D-7 | `tools/repo_scans.json` 은 새 항목이 아니라 앞 SPEC 의 항목 `table-step-matches-the-two-repository-tests` 에 줄을 더해요: 새 단위 테스트 둘과 그 검사, 규칙 파일 `tests/repo_scans_table.rs`·`tools/repo_scans.json`, 범위 `tools/repo_scans.json` | 같은 단계의 같은 규칙 파일을 읽는 테스트들이라 한 항목이에요. 공유 파일이라 `moai card line-lease` 로만 써요 | 더한 줄을 지우기 |
| D-8 | 카드 노트 7885 의 후속 셋을 이 SPEC 의 REQ-10~12 로 함께 해요 | 리더 결정 노트 7885(2026-11-23T20:54:25Z): 이 카드를 잡는 레인이 함께 해요. 셋 다 규칙을 늘리고 줄이지 않아요 | 없음 |

## Result

Measured 2026-11-24 (UTC, `date -u` 04:08:51Z) on macOS in the worktree `.claude/worktrees/repo-scans-precheck`, base `e9c6b93400f6a18b2e450eeaec6e06a9bd837a18` (the `NeXT` tip the branch was cut from; the checks did not follow the tip). Approval: leader note 8150, digest `sha256:2be56d7326c7b8fb2dbd00fb4ff107d1de9ae7966c357bea125d349971336b4e` (unchanged by the added Interview section). RED was the commit `f87f482884`: 14 tests ran, 11 failed and 3 passed (exit 100): the passing three are the repository-wide test and the `tools` condition test, which pass on an empty stub, and the skip-what-is-absent unit test. The code head is `ca8ad056e54e82e4024bd2c9d13cfbcdb25f576c`; GREEN is `68a5109142`. The first bare `moai check affected` run (head `64c8c46a45a4`, the earlier Result commit) ended `result=FAIL`: `diff_check=0 fmt=0 clippy=0 test=100`, 3786 tests, 3785 passed, 1 failed: `git_env_scrub every_git_subprocess_in_src_is_scrubbed`, because the repository-wide unit test spawned `git` without `scrub_git_env` (#2337). The commit `ca8ad056e5` uses the file's scrubbed `git()` helper; the re-run receipt is in the READY note.
- AC-1 met: `a_new_walking_test_file_missing_from_the_scan_list_stops_before_the_heavy_steps` passes: one line that names `tests/check_walk_cli.rs`, `tools/repo_scans.json` and `EXEMPT`; the registered, exempt and plain files get none; `test_tables=1 stopped_at=test_tables result=FAIL`, no fake `clippy` or test call.
- AC-2 met: `a_new_reader_without_an_always_scan_stops_before_the_heavy_steps` passes: `include_str!`, `include_bytes!` and a path read each get one line with the file and the path; the files kept out by `INCLUDE_STR_EXEMPT` and `FILE_READ_EXEMPT` get none (3 lines).
- AC-3 met: `a_registered_or_exempt_new_test_file_passes_and_the_step_is_planned_only_with_a_table` passes: registered and exempt files give `test_tables=0` and the heavy steps run; a reader that its `always` scan never runs with `--test` stops the run with the scan name; a repository with only the scan list plans the step; a repository with no table, and a range with no added test file, plan none.
- AC-4 met: `a_scan_list_that_cannot_be_read_stops_before_the_heavy_steps` passes for invalid UTF-8, text that is no JSON and JSON with no `scans` list (one line each, naming the file and `cannot be read`); the readable list gives `test_tables=0`.
- AC-5 met: the five `check_affected::tests::scan_violations_*` unit tests pass (each kind of miss, registered and exempt files and absent inputs, the three exemption lists with a commented row and escaped quotes, the rule file without its `--test` check, unknown tracked files).
- AC-6 met: `check_affected::tests::the_repository_tests_pass_the_scan_step_as_added_files` passes: the repository's own top-level `tests/*.rs` files, all taken as added, give no violation (0 false positives against `tests/repo_scans_table.rs`, which passes on the same tree).
- AC-7 met: `check_affected::tests::the_guidance_says_where_a_new_test_file_registers` passes with `tools/repo_scans.json` and `EXEMPT` added to the words both pages must name.
- AC-8 met: `every_file_read_exemption_is_needed` found 11 idle `FILE_READ_EXEMPT` pairs at RED (more than the two rows of note 7885): `README.md` of `boundary_repro`, `bundle_body_cli`, `bundle_build_cli`, `bundle_verify_cli`, `review_checklist_cli`, `run_verify`, `worktree_cli` and `project_flow_closing`, `tools` of `check_pipeline_config`, and `src/lib.rs` and `tools` of `repo_scans_table`. The rows left with no path (`review_checklist_cli`, `project_flow_closing`, `check_pipeline_config`) are removed whole and the other pairs one by one, the reasons of the rows that stay name only their remaining paths, and the test passes. Removing a pair never loosens the table: each of the 11 was measured as covered by an `always` scan with the pair taken out.
- AC-9 met: `the_include_str_reader_texts_name_include_bytes` passes; the failure text of the reader test is now the constant `INCLUDE_READER_HINT` and the two doc comments name `include_bytes!` too.
- AC-10 met: `file_read_paths_adds_the_tools_folder_only_with_all_three_conditions` passes.
- affected tests on GREEN `68a5109142`: `cargo nextest run --locked --offline --no-fail-fast` over `--lib` and the targets `check_affected_scans_cli`, `repo_scans_table`, `check_affected_tables_cli`, `test_area_layout`, `state_folder_literal_scan` with a filterset on those binaries and `check_affected::tests` ran 96 tests, 96 passed (exit 0), in one heavy slot (`--max 3`), `RUSTC_WRAPPER=sccache`.
- guard mutants, each run on the committed head, restored with `git checkout -- <file>` (`git status --short` empty after the last), each checked applied with `git diff --stat` before the run; a driver script in the scratchpad ran them. In `src/check_affected.rs`, tests `--lib check_affected::tests` and `--test check_affected_scans_cli`: (s1) the `EXEMPT` row ignored — 4 failed, exit 100, among them `the_repository_tests_pass_the_scan_step_as_added_files` (a stricter step is a drift: AC-6 catches it); (s2) the `FILE_READ_EXEMPT` row ignored — 5 failed, including the repository-wide test; (s3) the `INCLUDE_STR_EXEMPT` row ignored — 2 failed; (s4) `include_bytes!` dropped from the readers — 2 failed; (s5) a scan without `always` counts as covering — exit 0 on the first head `68a5109142` (a gap), then 1 failed on `11fa39ed66` after the case with a non-`always` scan was added to `scan_violations_name_each_kind_of_miss`; (s6) the scope suffix ignored — 1 failed; (s7) the `--test` check rule removed — 2 failed; (s8) unknown tracked files read as an empty set — 1 failed; (s9) the not-a-scan-list line removed — 1 failed; (s10) the step needs an area table to be planned — 1 failed; (s11) an unreadable scan list read as absent — 1 failed; (s12) a commented row read as a row — 1 failed; (s13) drift, the folder `tests` counted as a repository read — the repository-wide test failed (exit 100); (s14) drift, a walk needs no named root — the repository-wide test failed (exit 100). In `tests/repo_scans_table.rs`, target `repo_scans_table`: (t1) an idle row put back — `every_file_read_exemption_is_needed` failed; (t2) the reader hint without `include_bytes!` — `the_include_str_reader_texts_name_include_bytes` failed; (t3) the literal `tools` condition dropped — `file_read_paths_adds_the_tools_folder_only_with_all_three_conditions` failed, and so did the live reader test; (t4) the interpreter condition dropped — 3 failed; (t5) the tracked `tools` folder condition dropped — 2 failed. Each mutant ended with exit 100 except s5 on the first head.
- Clarifications of the approved text: (a) D-7 names two new unit tests for the entry in `tools/repo_scans.json`; one test is new (`the_repository_tests_pass_the_scan_step_as_added_files`), the guidance test was already listed, and the entry also names `tests/repo_scans_table.rs` and `tools/repo_scans.json` as rule files and gets the scope `tools/repo_scans.json`. (b) The risk line "skip the file-read rule when `git ls-files` cannot be read" is built as a line that makes the step fail (`names the repository root but git ls-files cannot be read`), so a skipped judgement is never a pass.
- Gap: the step reads the three exemption lists as text of `tests/repo_scans_table.rs` with a small reader of Rust strings and brackets; a list that changes its shape (a row that is not a tuple) is caught by AC-6 only when it hides a row the repository uses.
- Gap: the whole `cargo test --locked` and `check affected` of the READY head are not in this Result; the `check affected` receipt is in the READY note.
- Gap: no dangerous path (the step starts no process beyond `git ls-files` and deletes nothing).
