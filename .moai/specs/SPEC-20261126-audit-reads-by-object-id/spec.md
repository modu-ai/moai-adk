---
id: SPEC-20261126-audit-reads-by-object-id
title: "감사는 다른 레인의 head 를 자기 워크트리에서 객체 id 로 읽어요"
status: draft
revision: 1
created: 2026-11-26
---

# 감사는 다른 레인의 head 를 자기 워크트리에서 객체 id 로 읽어요

## Interview

- none: 감사 중에 세션 위치가 옮겨진 관찰에서 나온 변경이라 사용자에게 물은 것이 없어요.

## Intent

감사 세션이 감사 대상 워크트리로 `cd` 해서 호스트 세션의 작업 폴더가 옮겨진 일이 하룻밤에 다섯 번 있었어요. 워크트리에 격리된 세션에서 `git -C` 로 다른 워크트리를 가리키면 호스트 관문이 거절해요. 모든 워크트리는 객체 저장소를 함께 쓰므로 감사는 자기 워크트리에서 객체 id 로 다른 head 를 읽을 수 있어요. 이 길을 감사 계약과 레인 규칙에 한 줄씩 적고, 감사 글에 `cd` 나 `git -C` 가 새로 들어오면 실패하는 스캔 테스트를 둬요.

## Non-goals

- 감사 에이전트 글(`agents/audit.md`)과 AUDIT 워크플로(`skills/moai/workflows/audit.md`) 수정: 감사자가 만든 감사 워크트리를 쓰는 기존 줄은 baseline 에 남겨요.
- 훅이나 관문으로 `cd` 를 거절하는 일.
- 글자만 보고 대상 워크트리를 판별하는 일, `pushd`, `--git-dir`, `--work-tree`, 줄이 나뉜 명령 탐지.
- `skills/`, `agents/`, `docs/` 밖의 파일(`templates/` 아래 감사 이름 파일 포함).

## Requirements

- REQ-1: The audit contract shall tell its reader on one line to read another head from its own worktree by object id with `git show <sha>:<path>`, `git diff <base> <head>` and `git merge-tree` and to state the ban in the words "never `cd` or `git -C` into a worktree that is not your own".
- REQ-2: The lane rules shall state the same read by object id and the same ban on one line.
- REQ-3: When an audit text holds a `cd` or a `git -C` that no baseline entry covers, the scan test shall fail and name the file and the line · failure: none — a covered occurrence passes
- REQ-4: When a baseline substring occurs in its file a number of times other than its recorded count, the scan test shall fail and name the entry · failure: none — an exact count passes
- REQ-5: The scan test shall treat as audit texts every file under `skills/`, `agents/` and `docs/` whose file name holds `audit` and the five rule files it names.

## Acceptance

- AC-1 (REQ-1, REQ-2): The audit contract and the lane rules each hold one line with the own worktree, the object id, the three read commands and the ban · verify: `cargo test --locked --test audit_reads_by_object_id -- ac1_the_contract_and_the_lane_rules_say_to_read_another_head_by_object_id`
- AC-2 (REQ-3, REQ-4): Every `cd` and `git -C` of the checked-in audit texts lies inside a baseline substring whose count matches · verify: `cargo test --locked --test audit_reads_by_object_id -- ac2_the_checked_in_audit_texts_hold_only_the_baseline_cd_and_git_c`
- AC-3 (REQ-3): A planted `cd` or `git -C` line is reported with its file and line, and `AUDIT-CLAIM` and `git diff` are not reported · verify: `cargo test --locked --test audit_reads_by_object_id -- ac3_a_new_cd_or_git_c_instruction_is_named_with_its_file_and_line`
- AC-4 (REQ-4): A repeated or removed baseline substring is reported with its entry · verify: `cargo test --locked --test audit_reads_by_object_id -- ac4_a_baseline_count_that_differs_names_the_entry`
- AC-5 (REQ-5): A file is chosen by its name or by the named list and another file is not · verify: `cargo test --locked --test audit_reads_by_object_id -- ac5_audit_texts_are_chosen_by_file_name_or_by_the_named_list`

## Scope

- write: `.moai/specs/SPEC-20261126-audit-reads-by-object-id/**`, `skills/moai/references/audit.md`, `docs/lane-rules.md`, `skills/moai/references/source-manifest.json`, `tests/audit_reads_by_object_id.rs`, `tools/repo_scans.json`
- shared files (`card line-lease`, 커밋 뒤 바로 반납): `skills/moai/references/source-manifest.json`, `tools/repo_scans.json`
- read-only: `src/gate.rs`, `tests/repo_scans_table.rs`, `tests/next_flow_skill_scan.rs`
- forbidden: `agents/audit.md`, `skills/moai/workflows/audit.md`

## Plan

1. 실패하는 테스트와 스캔 표 항목을 함께 커밋해요. `tools/repo_scans.json` 은 공유 파일이라 `card line-lease` 로 써요.
2. 감사 계약의 "Rules for every audit" 끝과 레인 규칙의 "Worktree entry and exit" 에 한 줄씩 넣어요. 넣는 글은 아래 그대로예요.
   - 감사 계약: "6. Read another head from your own worktree by object id (`git show <sha>:<path>`, `git diff <base> <head>`, `git merge-tree`); never `cd` or `git -C` into a worktree that is not your own."
   - 레인 규칙: "- An audit reads another head from your own worktree by object id (`git show <sha>:<path>`, `git diff <base> <head>`, `git merge-tree`); never `cd` or `git -C` into a worktree that is not your own."
3. `moai check fix` 로 소스 매니페스트의 sha256 을 써요(`card line-lease`).
4. 대상 테스트와 가벼운 검사를 돌리고 Result 를 쓴 뒤 READY 에서 `check affected --card` 를 한 번 돌려요.

## Risks

- 감사 계약은 8192 바이트 한도이고 지금 7966 바이트예요. 새 줄을 넣고도 한도 아래여야 하고 한도는 올리지 않아요.
- 스캔 표를 `test_targets` 형식으로 바꾸는 SPEC-20261125-repo-scans-convert-test-targets 가 먼저 착지하면 새 항목은 `checks` 대신 `test_targets` 를 써야 해요.
- 탐지는 글자 기준이라 자기 워크트리를 가리키는 새 줄도 실패해요. 그런 줄은 baseline 을 함께 고쳐요.
- 감사 에이전트 글은 감사 워크트리를 `git -C` 로 가리키라고 하는데 호스트 관문 안내는 격리 세션이 다른 워크트리를 가리키는 그 형태를 거절한다고 적어요. 이 SPEC 은 그 차이를 고치지 않아요.

## Decisions

Reversible, category `none`.

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | the ban names a worktree that is not the auditor's own | the criterion says another worktree, and the two audit texts outside the scope let the auditor use the audit worktree it made | widen the line and its baseline substring |
| D-2 | audit texts are the files named audit under three folders and five named rule files | small and stable, and a new audit file joins by its name | edit the named list |
| D-3 | each occurrence is covered by a baseline substring with a count | an added occurrence fails and a removed one asks to shrink the baseline | edit the baseline |
| D-4 | the scan has its own always entry over the three folders | the existing read groups list named files and would miss a new audit file | remove the entry and join those groups |
