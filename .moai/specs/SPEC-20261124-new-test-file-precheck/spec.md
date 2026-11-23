---
id: SPEC-20261124-new-test-file-precheck
title: "check affected stops before the heavy steps when an added test file is missing from the area table or the state-folder allow list"
status: approved
revision: 1
created: 2026-11-24
approval:
  by: leader
  at: 2026-11-23
  basis: "리더 결정 노트 7835 (2026-11-23T19:44:54Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단, 사용자 직접 선택 아님, 노트 7834 에 답함): (A) 판 1 승인, 조건 넷(새 시험 파일이 새 단계를 통과, 멈춤과 통과 둘 다 임시 프로젝트 시험과 guard-mutant exit 100, 멈춤이 cargo 호출 없이 무거운 단계 앞이라는 시험, lane-rules.md 는 짧은 문장 하나); 레인은 그 노트를 인용만 해요"
  digest: "sha256:c10c10d8eb247c7606ac826ade4e80aa85ce64da0cca0999b4dbbd43feda8eb9"
---

# check affected stops before the heavy steps when an added test file is missing from the area table or the state-folder allow list

## Intent

재발 방지예요. 레인이 새 `tests/*.rs` 파일을 더하고 `moai check affected` 를 끝까지 돌린 뒤에야 `test_area_layout` 의 `area_table_maps_every_real_test_file` 과 `state_folder_literal_scan` 의 `the_repository_tests_stay_within_the_state_folder_baseline` 이 새 파일 때문에 빨개지는 일이 두 번 있었어요(`SPEC-20261121-moai-name` 의 조각 하나와 기준 조각 하나). 두 시험은 표 하나를 읽고 파일 이름 몇 줄을 세는 가벼운 검사인데, 무거운 단계(clippy, 빌드, 시험) 뒤에야 만나서 한 번에 수십 분을 버려요.

`check affected` 는 이미 "가벼운 단계가 실패하면 무거운 단계 앞에서 멈추는" 규칙을 가져요(`SPEC-20261122-check-light-stop`). 이 SPEC 은 범위가 새로 더한 최상위 `tests/*.rs` 파일을 두 표에 맞춰 보는 가벼운 단계 `test_tables` 를 하나 더해요. 어긋나면 그 단계가 실패하고, 줄 하나가 "어느 표에, 어느 줄을 더하라" 를 그대로 말해요. 레인 지침은 새 시험 파일을 어디에 등록하는지 한 곳에 적어요.

읽는 사람: 레인, 리더, 이 변경의 감사자.

## Interview

- Q1: 새 단계는 외부 프로그램을 부르는가? -> A: 아니요 (D-1). 범위의 사실(더한 경로, 시험 파일 글, 두 표의 글)은 `read_facts` 가 이미 읽고, 판정은 순수 함수예요. 단계는 판정 줄을 로그에 쓰고 종료 코드 1 을 내는 단계예요.
- Q2: 모든 실행에 이 단계가 들어가는가? -> A: 아니요 (D-2). 범위가 최상위 `tests/*.rs` 를 더했고 판정에 쓸 표 파일이 있을 때만 계획해요. 그 밖의 실행은 요약 줄의 낱말도 순서도 그대로예요.
- Q3: 표를 풀어 주는가? -> A: 아니요 (D-3). 두 표는 그대로이고, 이 단계는 읽기만 해요.
- Q4: 막는 검사는 먼저 observe 로 도는가? -> A: 이 단계는 저장소의 두 시험이 이미 막고 있는 규칙을 앞으로 당길 뿐이에요(D-4). 새 거절 규칙이 없고, 오탐 여부는 저장소의 실제 `tests/` 전체를 "더한 파일" 로 놓고 위반이 0 인지로 재요(AC-6).

## Non-goals

- `tests/areas.toml` 이나 `tests/fixtures/state_folder_literals/allow.tsv` 의 규칙을 바꾸거나 풀지 않아요. 두 시험은 그대로예요.
- `tests/<영역>/*.rs` 모듈 파일, 고쳐서 문자열 수가 늘어난 기존 파일, `.py` 시험 파일은 이 단계가 보지 않아요: 카드가 말한 "더한 `tests/*.rs`" 만이에요(그 밖은 두 시험이 그대로 잡아요).
- 표 파일을 자동으로 고치지 않아요: 어느 줄을 더할지만 알려 줘요.
- `--lane` 선택 규칙과 무거운 단계의 순서는 바꾸지 않아요.

## Requirements

- REQ-1: When a range adds a top-level `tests/<name>.rs` whose first word no array of `tests/areas.toml` lists, `check affected` shall fail the cheap step `test_tables` with a line that names the file, the table and the word to add · failure: none — the line is the only output of the step.
- REQ-2: When a range adds a top-level `tests/<name>.rs` that `tests/areas.toml` maps to a moved area and does not list as solo, or that is named like an area crate, `check affected` shall fail the cheap step `test_tables` with a line that names the file and the folder it belongs in · failure: none — the line is the only output of the step.
- REQ-3: When a range adds a top-level `tests/<name>.rs` that spells a state folder name and has no row in `tests/fixtures/state_folder_literals/allow.tsv`, `check affected` shall fail the cheap step `test_tables` with a line that names the file and the exact `path<TAB>earlier<TAB>new` row to add · failure: none — the line is the only output of the step.
- REQ-4: When the added file has a row but spells a name more often than the row allows, `check affected` shall fail the cheap step `test_tables` with a line that names the file, the counts and the row · failure: none — the line is the only output of the step.
- REQ-5: When `test_tables` fails, `check affected` shall stop before the heavy steps and name `test_tables` in the summary as `stopped_at=test_tables` · failure: none — the rule of the cheap steps already does this.
- REQ-6: When a range adds no top-level `tests/*.rs` file, or the repository has neither table file, `check affected` shall plan no `test_tables` step · failure: none — the summary line stays as it was.
- REQ-7: `check affected --list` shall show `test_tables` as a step with `heavy: false` before the first heavy step when it is planned · failure: none — a step that is not planned has no entry.
- REQ-8: When the repository's own `tests/*.rs` files are all taken as added, the `test_tables` verdict shall find no violation · failure: none — a difference is a drift between this step and the two tests.
- REQ-9: The operation page of `check` and the Tests section of `docs/lane-rules.md` shall say where a new test file registers and that `test_tables` stops the run · failure: none — a test reads both texts.

## Acceptance

- AC-1 (REQ-1, REQ-2): 임시 저장소 범위가 첫 낱말이 표에 없는 `tests/zzfresh_cli.rs` 를 더하면 요약 줄이 `test_tables=1 stopped_at=test_tables result=FAIL` 이고 가짜 clippy·시험 호출 흔적이 없고, 단계 로그가 파일·`tests/areas.toml`·`"zzfresh"` 를 말해요; 옮겨진 영역의 낱말 파일과 영역 크레이트 이름 파일도 각각 한 줄을 받아요 · verify: `cargo nextest run --locked --offline --test check_affected_tables_cli a_new_test_file_missing_from_the_area_table_stops_before_the_heavy_steps`
- AC-2 (REQ-3, REQ-4): 상태 폴더 이름을 적은 새 시험 파일이 allow 목록에 없으면 같은 멈춤이 오고 로그에 정확한 `path<TAB>0<TAB>1` 행이 있어요; 행이 있어도 세는 수가 더 크면 그 수들이 있어요; 행이 같거나 더 크면 단계가 0 이에요 · verify: `cargo nextest run --locked --offline --test check_affected_tables_cli a_new_test_file_missing_from_the_state_folder_allow_list_stops_before_the_heavy_steps`
- AC-3 (REQ-5, REQ-6): 두 표에 맞는 새 시험 파일은 `test_tables=0` 으로 지나가 무거운 단계가 돌고, 시험 파일을 더하지 않았거나 표 파일이 없는 저장소의 요약 줄에는 `test_tables` 낱말이 없어요 · verify: `cargo nextest run --locked --offline --test check_affected_tables_cli a_registered_new_test_file_passes_and_other_ranges_plan_no_table_step`
- AC-4 (REQ-7): `--list` 의 단계 목록에서 `test_tables` 는 `heavy: false` 이고 첫 `clippy` 앞에 있어요 · verify: `cargo nextest run --locked --offline --test check_affected_tables_cli list_shows_the_table_step_before_the_heavy_steps`
- AC-5 (REQ-1…REQ-4): 순수 판정 함수 단위 시험이 위반 종류별 줄(미등록 낱말, 옮겨진 영역, 영역 크레이트 이름, 목록에 없는 행, 초과 행)과 위반 없음을 가려요 · verify: `cargo nextest run --locked --offline --lib check_affected::tests::table_violations`
- AC-6 (REQ-8): 저장소의 실제 `tests/*.rs` 전체를 더한 파일로 놓고 판정하면 위반이 0 개예요 · verify: `cargo nextest run --locked --offline --test check_affected_tables_cli the_repository_tests_pass_the_table_step_as_added_files`
- AC-7 (REQ-9): `check.md` 와 `docs/lane-rules.md` 가 `test_tables`, `tests/areas.toml`, `allow.tsv` 를 말해요 · verify: `cargo nextest run --locked --offline --test check_affected_tables_cli the_guidance_says_where_a_new_test_file_registers`

## Scope

- write: `.moai/specs/SPEC-20261124-new-test-file-precheck/**`, `src/check_affected.rs`, `tests/check_affected_tables_cli.rs`(새), `docs/design-review/appendix/operations/check.md`, `docs/lane-rules.md`
- read-only: `tests/areas.toml`, `tests/test_area_layout.rs`, `tests/state_folder_literal_scan.rs`, `tests/fixtures/state_folder_literals/allow.tsv`, `tests/check_affected_cli.rs`, `SPEC-20261122-check-light-stop`
- forbidden: `tests/areas.toml` 와 `allow.tsv` 의 규칙·내용 변경, `src/lib.rs`, 새 Python 시험, 카드 번호를 SPEC id·가지·커밋·PR 에 쓰는 것, 실제 HOME, 다른 레인의 파일

## Plan

1. RED: `tests/check_affected_tables_cli.rs` 의 시험 다섯과 `check_affected::tests::table_violations*` 단위 시험을 먼저 쓰고, 코드 없이 실패하는 것을 확인해요.
2. GREEN, `src/check_affected.rs`: `Facts` 에 더한 경로, 두 표의 글, 옮겨진 영역 이름을 더하고, 순수 함수 `table_violations` 와 단계 `test_tables`(로그에 줄을 쓰고 1 을 내는 안쪽 단계)를 더해요. 단계는 첫 무거운 단계 앞에 끼워서 `run` 과 `list` 가 같은 계획을 써요.
3. 글: `check.md` 의 "시험 영역 표" 와 "쓰는 법과 답", `docs/lane-rules.md` 의 Tests.
4. 뮤턴트: 각 새 가드 시험마다 코드를 지우고 실패를 봐요. 영향 시험 한 번, READY.

## Risks

- 단계가 세는 규칙이 `state_folder_literal_scan` 과 어긋날 수 있어요: `.rs` 파일의 두 이름 세는 규칙을 두 곳이 가져요. AC-6 의 저장소 전체 대조 시험이 한쪽만 바뀐 드리프트를 잡아요(두 시험이 이미 통과하는 트리에서 이 단계가 위반을 내면 어긋난 거예요).
- 요약 줄에 새 낱말 `test_tables` 가 생겨요(단계가 계획될 때만): 낱말 순서를 읽는 외부 도구가 있다면 깨질 수 있어요. 저장소 안에서 줄을 읽는 곳은 연산 페이지와 `check_affected*` 시험뿐이에요.
- 표 파일이 없는 저장소(임시 저장소 시험)에서는 단계가 계획되지 않아요(REQ-6). 진짜 저장소에서 표가 사라지면 두 표 시험이 먼저 빨개져요.
- `tests/areas.toml` 은 한 줄에 배열 하나인 줄 단위 읽기를 전제해요(Python 도구 때문이에요). 이 단계도 같은 줄 읽기만 써서 `toml_edit` 에 기대지 않아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 새 단계는 바깥 프로그램이 아니라 계획할 때 판정을 끝내고 줄만 로그에 쓰는 안쪽 단계예요 | 범위의 사실은 이미 읽혀 있고, 가짜 PATH 시험에서도 같은 길을 타요 | 단계를 외부 명령으로 바꾸기 |
| D-2 | 더한 최상위 `tests/*.rs` 가 있고 표 파일이 있을 때만 계획해요 | 다른 실행의 요약 줄이 그대로여야 하고, 표 없는 임시 저장소 시험이 깨지지 않아요 | 항상 계획 |
| D-3 | 두 표는 읽기만 하고 규칙은 그대로예요 | 사용자 지시 "표를 풀지 않는다" | 없음 |
| D-4 | observe 단계 없이 처음부터 막아요 | 새 규칙이 아니라 저장소의 두 시험이 이미 막는 규칙을 앞으로 당긴 것이고, 오탐은 AC-6 으로 재요 | 단계를 계획에서 빼기 |
| D-5 | 첫 낱말이 미등록일 때 줄은 낱말 하나와 영역 이름들을 말하고, 어느 영역인지는 고르지 않아요 | 영역 선택은 사람(파일의 성격)의 일이에요 | 없음 |

## Result

(status draft — written when the slice lands)
