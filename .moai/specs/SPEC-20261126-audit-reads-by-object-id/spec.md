---
id: SPEC-20261126-audit-reads-by-object-id
title: "감사는 다른 레인의 head 를 자기 워크트리에서 객체 id 로 읽어요"
status: draft
revision: 2
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
- 스캔 표를 `test_targets` 형식으로 바꾸는 SPEC-20261125-repo-scans-convert-test-targets 가 착지한 뒤에는 새 항목도 `checks` 대신 `test_targets` 를 써요. 그 전에는 변환 전 가드 테스트가 그 형식을 거절해요.
- 스캔 항목의 docs 범위는 세 경로뿐이에요. `docs/` 의 다른 자리에 audit 이름 파일이 새로 생기면 바뀐 범위로 고르는 검사는 이 스캔을 고르지 않고, 이 테스트가 도는 실행이 잡아요.
- 탐지는 글자 기준이라 자기 워크트리를 가리키는 새 줄도 실패해요. 그런 줄은 baseline 을 함께 고쳐요.
- 감사 에이전트 글은 감사 워크트리를 `git -C` 로 가리키라고 하는데 호스트 관문 안내는 격리 세션이 다른 워크트리를 가리키는 그 형태를 거절한다고 적어요. 이 SPEC 은 그 차이를 고치지 않아요.

## Decisions

Reversible, category `none`.

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | the ban names a worktree that is not the auditor's own | the criterion says another worktree, and the two audit texts outside the scope let the auditor use the audit worktree it made | widen the line and its baseline substring |
| D-2 | audit texts are the files named audit under three folders and five named rule files | small and stable, and a new audit file joins by its name | edit the named list |
| D-3 | each occurrence is covered by a baseline substring with a count | an added occurrence fails and a removed one asks to shrink the baseline | edit the baseline |
| D-4 | the scan has its own always entry over `skills/`, `agents/`, `docs/lane-rules.md`, `docs/leader-rules.md` and `docs/design-review/appendix/operations/` | the existing read groups list named files and would miss a new audit file, and a scope over all of `docs/` makes a docs-only change an own hit, which an existing repository test refuses | remove the entry and join those groups |

## Result

Measured 2026-11-26 (UTC, `date -u` 03:21:58Z when written) on macOS in the worktree `.claude/worktrees/audit-object-read-abel`, base `91dd561a5aa2740aad0fae713a2e51a74525e7fc` (the `NeXT` tip the leader named; the branch was cut from `e3450e0ede` and took that tip once, merge `79dffcece3`, while only the tip side had changed the skill package manifest). `NeXT` has moved since and the checks did not follow it. Approval: the leader's delegated approvals, notes 9728 (revision 1) and 9889 (revision 2), approval digest `sha256:111085483ca7dc5eebde9d499db4d8eb2b0c4afd45d7ebac48ab2b552fb6057d` for both: the digest did not change with the revised decision row, so note 9889 quotes that row. Commits: SPEC draft `ba5d1094ba`, test and scan entry `553b149f51`, the two rule lines `f24bd136ca`, the `test_targets` form `6b59ce0006` and its revert `f8bf505bca`, manifest `d7e5a82870`, revision 2 `c784f3ad74`, narrowed scan scope `844f46c004`. RED on `553b149f51`: `audit_reads_by_object_id` 3 passed and 2 failed (AC-1 and AC-2, for the missing lines), `repo_scans_table` 33 passed. All numbers below are from one slot run on `844f46c004` (exit 0): `audit_reads_by_object_id` 5, `audit_text_facts` 7, `next_flow_skill_scan` 76, `repo_scans_cli` 25, `repo_scans_table` 33, `source_manifest_facts` 8 passed, 154 passed and 0 failed.
- AC-1 met: `cargo test --locked --test audit_reads_by_object_id -- ac1_the_contract_and_the_lane_rules_say_to_read_another_head_by_object_id`, 1 passed, exit 0. The audit contract holds the line as rule 6 of "Rules for every audit" (8155 of 8192 bytes) and the lane rules hold it under "Worktree entry and exit".
- AC-2 met: `... -- ac2_the_checked_in_audit_texts_hold_only_the_baseline_cd_and_git_c`, 1 passed, exit 0. The baseline holds eight rows over `agents/audit.md`, one operations text, the lane rules and the contract.
- AC-3 met: `... -- ac3_a_new_cd_or_git_c_instruction_is_named_with_its_file_and_line`, 1 passed, exit 0.
- AC-4 met: `... -- ac4_a_baseline_count_that_differs_names_the_entry`, 1 passed, exit 0.
- AC-5 met: `... -- ac5_audit_texts_are_chosen_by_file_name_or_by_the_named_list`, 1 passed, exit 0.
- guard mutants, run by `target/abel/checks-and-mutants.sh` on `844f46c004`, each restored with `git checkout --`, five of five caught with exit 101: the new contract line removed (AC-1), a `cd` line added to the contract (AC-2), the `git -C` match made never to match (AC-3), the baseline count comparison removed (AC-4), the file-name rule of the selection removed (AC-5).
- The first scan entry had all of `docs/` in its scope. `repo_scans_cli` then failed (`always_the_checked_in_table_flags_the_two_whole_tree_scans`: a docs-only change became an own hit; 153 passed, 1 failed on `d7e5a82870`). The run recorded as green for `f24bd136ca` covered two targets only and did not show it. Decision D-4 was revised and the scope narrowed; the existing test is unchanged.
- The `test_targets` form was tried and refused at this base (`the_checked_in_table_names_no_test_targets_before_the_conversion` and three tests of `repo_scans_cli`, one of them the failure above), so the entry keeps the `checks` form.
- Landing: `NeXT` has since changed the skill package manifest and the scan table, which this branch also changes, so the lander merges the tip. If the conversion SPEC `SPEC-20261125-repo-scans-convert-test-targets` is landed first, the new entry has to be converted to `test_targets` in that merge, and the scan test has to run once on the merged tree.
- Gap: the `check affected` receipt of the READY head is in the READY note, not here. The whole `cargo test --locked` was not run. No dangerous path: nothing is deleted and no process is started beyond the cargo test runs; the new test only reads files.
- Gap: a new audit-named file at another place under `docs/` does not select the scan in a changed-path run; the test itself walks all of `docs/` whenever it runs. `pushd`, `--git-dir`, `--work-tree` and wrapped commands are not detected.
- Follow-up candidates, not changed here: `agents/audit.md` and the AUDIT workflow tell the auditor to name the audit worktree with `git -C`, while the host guard refuses that form from an isolated session; the `rejects` text of the scan entry says "under skills/, agents/ or docs/", wider than its scope.
