---
id: SPEC-20261125-state-folder-test-helpers
title: "테스트가 옛 상태 폴더 철자를 직접 쓰지 않게 도우미와 지침과 스캔 문구를 둬요"
status: draft
revision: 1
created: 2026-11-25
---

# 테스트가 옛 상태 폴더 철자를 직접 쓰지 않게 도우미와 지침과 스캔 문구를 둬요

## Intent

테스트 도구와 문서만 바꾸는 일이에요(tier 0·1). 테스트가 옛 상태 폴더 철자를 직접 쓰는 일이 세 번 반복됐어요(상태 폴더를 한 곳에서만 쓰게 한 조각의 첫 영수증, 그 앞 조각, 데스크탑 설치 SPEC 의 `tests/desktop_folder_rules_cli.rs` 가 여섯째 묶음 넓은 검사에서). 레인의 base 가 스캔보다 앞이면 착지 묶음에서야 드러나 묶음 하나를 다시 돌려요. 임시 HOME 의 상태 폴더를 주는 도우미 둘, 스캔이 거절할 때 그 도우미를 이름으로 알려 주는 문구, 레인 규칙 한 줄로 막아요. 이미 있는 테스트는 한꺼번에 고치지 않아요.

### 결정

- D-1: 지침 한 줄은 `docs/lane-rules.md` 에 둬요. `skills/moai/references/internal/test-design.md` 는 사용자 프로젝트에도 나가는 글이고 `source-manifest.json`(공유 파일) 과 여러 패키지 테스트가 묶고 있어서, moai-adk 안의 `tests/common` 이름과 `allow.tsv` 를 거기 적지 않아요.
- D-2: 새 도우미는 폴더 이름을 글자로 쓰지 않아요. 새 배치는 `moai_cli::home_root::path`, 옛 배치는 `moai_cli::HOME_STATE_DIR` 상수로 만들어요. 그래서 도우미를 쓰는 테스트에는 `allow.tsv` 행이 필요 없어요.

## Requirements

- REQ-1: When a test asks `TempHome::state_dir`, the helper shall create the HOME's state folder and return the path that `moai_cli::home_root::path` gives for that HOME. · failure: none — a folder that cannot be created fails the test with the io error.
- REQ-2: When a test asks `TempHome::make_earlier_state_dir`, the helper shall create the earlier state folder under that HOME and return its path, so that `moai_cli::home_root::dir` names the earlier folder for the HOME. · failure: none — a folder that cannot be created fails the test with the io error.
- REQ-3: When a test file spells a state folder name outside its row of `allow.tsv`, the scan test shall fail with a message that names `TempHome::state_dir` and `TempHome::make_earlier_state_dir`. · failure: none — the scan only reports.
- REQ-4: When a new test file spells a state folder name outside `allow.tsv`, the `test_tables` step of `check affected` shall stop with a message that names `TempHome::state_dir` and `TempHome::make_earlier_state_dir`. · failure: none — the step only reports.
- REQ-5: The lane rules shall say in one line that a test finds a state folder through the two helpers and spells the name only in a fixture with its `allow.tsv` row. · failure: none — a document.

## Acceptance

- AC-1 (REQ-1, REQ-2): a test of the two helpers shows each path equals what the resolver answers for a fresh HOME and for a HOME made by the earlier-layout helper · verify: `cargo test --offline --locked --test state_folder_literal_scan the_state_folder_helpers_agree_with_the_resolver`
- AC-2 (REQ-1, REQ-2): the text of a test that uses the two helpers counts no state folder name, and the scan flags the same text with a literal · verify: `cargo test --offline --locked --test state_folder_literal_scan a_test_that_uses_the_helpers_spells_no_state_folder_name`
- AC-3 (REQ-3): the scan test's rejection names both helpers · verify: `cargo test --offline --locked --test state_folder_literal_scan the_scan_rejection_names_the_helpers`
- AC-4 (REQ-4): the table step's rejection names both helpers · verify: `cargo test --offline --locked --lib a_new_test_file_with_a_state_folder_literal_is_told_the_helpers`
- AC-5 (REQ-5): `docs/lane-rules.md` carries the line · verify: `cargo test --offline --locked --lib the_guidance_says_where_a_new_test_file_registers`

## Scope

- write: `tests/common/home.rs`
- write: `tests/state_folder_literal_scan.rs`
- write: `src/check_affected.rs`
- write: `docs/lane-rules.md`
- write: `docs/later.md`
- write: `.moai/specs/SPEC-20261125-state-folder-test-helpers/**`

## Plan

1. Write the tests of the helpers and of both messages first, run them red.
2. Add the two helpers to `tests/common/home.rs`.
3. Name the helpers in the scan test's message and in the `test_tables` message.
4. Add the lane rules line; read whether `check affected --card` reads the tip's scan table and note the answer.
