---
id: SPEC-20261118-update-project-surveys
title: "Surveys of the update, project and setup text: run-it-yourself and GitHub-Issue assumptions"
status: draft
revision: 1
created: 2026-11-18
---

# Surveys of the update, project and setup text

## Intent

카드 C-51(사용자 지시 2026-11-16): `/moai update`·`/moai project`·`/moai setup`에서 세션이 명령을 직접 돌리고, 사용자는 권장안이 먼저 오는 선택만 해요. 또 GitHub Issue는 더 이상 작업 장부가 아니라 로컬 카드 장부(`moai card`)가 장부예요. 리더 note 2069가 tip에서 읽은 대로 완료 조건 3·4·5·6은 이미 tip에 있고, 남은 것은 조사표 둘(완료 조건 1·2)이에요. 이 SPEC은 그 조사표를 `docs/surveys/`에 쓰고, 모든 행의 `path:line`이 실제 파일에 있다는 증거를 남겨요.

## Non-goals

- 조사표가 가리키는 코드·문구·템플릿을 고치기(후속 묶음으로 리더가 카드를 만들어요): `src/agents.rs:269`·`:290`의 "Issue scope" 문구(행 B13)는 `tests/product_cli.rs`, `tests/language_setting.rs`, 저장소 `AGENTS.md`가 같은 문구를 고정하고 있어 이 카드의 쓰기 범위 밖이에요.
- 조사표에 시험을 붙이기: 표의 행을 고정하는 시험은 `docs/`를 읽는 시험이라 `ci_scope` 거부 목록과 새 시험 파일 범위가 필요해요.

## Requirements

- REQ-1: The survey of text that hands work to the user or asks an open question shall list every such place of the update, project and setup routes with `path:line`, the line as read and the planned change.
- REQ-2: The survey of GitHub-Issue assumptions shall list every such place of the same scope with `path:line`, the line as read, a class (defect or keep) and the planned change.
- REQ-3: Every `path:line` of both surveys shall name a line of the file at the surveyed head that contains the row's key word.
- REQ-4: Each survey shall say what its rows do not prove and what it did not cover.
- REQ-5: The Acceptance section shall cite each criterion that the tip already meets by the line or test that shows it.

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-4): `docs/surveys/update-project-user-requests.md` lists the workflow side (12 rows) and the CLI side (61 rows) with `path:line`, the line as read, a class and the planned change.
  verify: `rg -c '^\| [AW][0-9]+ ' docs/surveys/update-project-user-requests.md` prints 73
- AC-2 (REQ-2, REQ-3, REQ-4): `docs/surveys/update-project-github-issue-assumptions.md` lists 22 rows, each a defect or a keep.
  verify: `rg -c '^\| B[0-9]+ ' docs/surveys/update-project-github-issue-assumptions.md` prints 22
- AC-3 (REQ-5): the UPDATE workflow has the session run the user-layer update and its confirm steps itself (`skills/moai/workflows/update.md:5`).
  verify: `cargo test --locked --offline --test update_route the_update_route_runs_the_commands_itself_after_the_pick`
- AC-4 (REQ-5): every question in UPDATE and PROJECT is a selection with the recommended option first (`skills/moai/references/internal/questions.md:7-8`).
  verify: `cargo test --locked --offline --test update_route every_question_is_a_selection_through_the_hosts_question_tool`
- AC-5 (REQ-5): `/moai project` does not offer GitHub Issues as the ledger or the development method (`skills/moai/workflows/project.md:41`, `skills/moai/references/internal/interview.md:17`).
  verify: `cargo test --locked --offline --test update_route the_ledger_is_local_cards_and_previews_run_without_asking`
- AC-6 (REQ-5): the skill source manifest matches the files.
  verify: `python3 -I tools/source_manifest.py check`

## Scope

- write: `docs/surveys/update-project-user-requests.md`, `docs/surveys/update-project-github-issue-assumptions.md`, `.moai/specs/SPEC-20261118-update-project-surveys/spec.md`
- read-only: everything the surveys cite
- forbidden: any change to the cited code, templates, tests or the repository `AGENTS.md`

## Plan

1. Read the workflow and reference side by hand; have a read-only research subagent list the CLI side and the Issue assumptions with their pins.
2. Put every row through a mechanical check: the cited line contains the row's key word; fix the rows that fail (25 of 95 failed the first pass) and copy the "line as read" from the file.
3. Write the two documents and this SPEC; re-run the existing tests that show criteria 3 to 6.

## Risks

- A key word on a line does not prove the class or the planned change; the documents say so (REQ-4). A reader who changes a row must re-read the code.
- Line numbers drift as soon as another pull request edits a cited file; the documents name the head they were read at (`eea59f4ff1`).

## Decisions

| id | decision | why | alternative |
| --- | --- | --- | --- |
| D-1 | Docs only; no test pins the rows | a test reading `docs/` needs a `ci_scope` entry and a test file outside the card's scope | add `tests/update_project_surveys.rs` after the leader widens the scope |
| D-2 | The "line as read" column is copied from the file by the generator | a subagent's quotes cannot reach the document unverified | quotes from the subagent |
| D-3 | The wording fix of B13 is a follow-up row, not part of this PR | three pins outside the scope and a managed-section regeneration; one SPEC is one pull request | widen the scope and include it |

## Result

- AC-1: met — `docs/surveys/update-project-user-requests.md` has 12 workflow rows and 61 CLI rows (`rg -c '^\| [AW][0-9]+ '` printed 73); each row's cited line contains its key word in the file at head `eea59f4ff1` (95 of 95 after 25 corrections, with the survey B rows); classes and planned changes are a reading, stated in the document. Not covered: `src/hook_engine.rs`, claim/lane/gate/trailer modules, the `docs/design-review/appendix/operations` files.
- AC-2: met — `docs/surveys/update-project-github-issue-assumptions.md` has 22 rows, 16 defects and 6 kept (`rg -c '^\| B[0-9]+ '` printed 22); the wording fix of B13 (`src/agents.rs:269`, `:290`) is a follow-up row because three pins sit outside this card's scope.
- AC-3: met — `cargo test --locked --offline --test update_route` 8 passed, including `the_update_route_runs_the_commands_itself_after_the_pick` (`skills/moai/workflows/update.md:5`).
- AC-4: met — same run, `every_question_is_a_selection_through_the_hosts_question_tool` (`skills/moai/references/internal/questions.md:7-8`).
- AC-5: met — same run, `the_ledger_is_local_cards_and_previews_run_without_asking` (`skills/moai/workflows/project.md:41`, `skills/moai/references/internal/interview.md:17`).
- AC-6: met — `python3 -I tools/source_manifest.py check` exit 0, `{"ok": true, "changed": []}`; `python3 -m unittest tests.ci_scope_test` OK.
