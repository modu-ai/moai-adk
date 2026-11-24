---
id: SPEC-20261124-cleanup-leftovers
title: "Six small leftovers of the 2026-11-24 card cleanup land as separate slices of one SPEC"
status: approved
revision: 1
created: 2026-11-24
approval:
  by: leader
  at: 2026-11-24
  basis: "리더 결정 노트 8093 (2026-11-24T01:39:02Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단, 사용자 직접 선택 아님, 노트 8092 에 답함): 판 1 승인 선택지 A, 커밋 95cd8e1d91; 조각 순서 A, B, C, D; 조각 C 의 D-6, D-7, D-8 은 패치가 정한 대로 승인; 조각 D 는 위험 경로 실행과 tier 2 감사 필수; 기준 1 과 4 는 Deferred 그대로; 레인은 그 노트를 인용만 해요"
  digest: "sha256:6cb40812e2a5652fa202c07d2fb38a313872184a89f7fff0469532c3bc841aef"
---

# Six small leftovers of the 2026-11-24 card cleanup land as separate slices of one SPEC

## Intent

사용자가 2026-11-24 에 열린 카드를 정리하며 개수를 줄이라고 정했어요. 닫힌 카드 여섯 장에 작은 조각이 하나씩 남았고, 새 카드 여섯 대신 한 카드의 기준 여섯으로 묶였어요. 이 SPEC 은 그 여섯 조각 가운데 지금 할 수 있는 네 조각(기준 6, 3, 5, 2)을 정하고, 지금 할 수 없는 두 조각(기준 1, 4)이 왜 기다리는지 적어요. 조각마다 따로 커밋하고 따로 감사받을 수 있어요.

원본 카드의 마지막 노트를 기준마다 읽고 현재 통합 브랜치(`NeXT`, `546cd451a2`)와 맞춰 봤어요(Evidence).

## Interview

- none: 사용자와 리더의 결정이 이미 기준 여섯과 순서로 적혀 있어 새로 물을 것이 없어요(배차 글 2026-11-24, 카드 C-446 본문; 사용자 결정은 카드 정리 2026-11-24).

## Non-goals

- 기준 1(하위 `.gitignore`)과 기준 4(설치·업데이트 보고)는 이 SPEC 에서 만들지 않아요. 둘 다 아래 Deferred 에 이유와 다시 시작할 조건이 있어요.
- 출력 스타일의 렌더 본문(`skills/moai/references/home/output-style.md`)과 전역 사본은 건드리지 않아요. 템플릿 세 개만이에요.
- 노트 7169 의 나머지 비차단 항목(`render.rs` 테마 주석, `docs/rust-product-path.md` 예시, `SPEC-20261123-progress-board-flow-icons` 의 superseded 표시)은 기준 문장에 없어서 하지 않아요.
- `docs/rust-product-path.md` 는 다른 카드가 쥐고 있어서 기준 3 에서도 건드리지 않아요.

## Requirements

- REQ-1: Each of the three shipped output-style templates (`moai.md`; `moai-easy.md`; `moai-learn.md`) shall name only the progress icons of the one flow set (🗂️ 🟡 📨 🔍 🔁 🛬 ✅ beside 🚧 🙋 🔴 ❌) and none of ⬜ 🟢 ⏸️ 🔵 👀 ⛔ · failure: none — the retired icons leave the three files and a scan test fails when one returns.
- REQ-2: The update-recovery section of `docs/support-matrix.md` shall describe the Claude projection and the user skill folder that the `release_suite` fixture uses now · failure: none — the second-host sentence is false on the integration branch since the fixture moved to the Claude host alone.
- REQ-3: The pointers in `docs/host-support-matrix.md` and `docs/product-naming-inventory.md` to the deleted Codex adapter page shall say that the page was deleted and name the deleting commit · failure: none — a dangling pointer becomes a dated record.
- REQ-4: The test `the_repository_tracks_no_codex_folder` shall read the committed tree with `git ls-tree` and fail when git itself fails · failure: none — a git failure no longer passes the check as "no such entry".
- REQ-5: The refusal of an unknown `permissions` key shall print the key only through `shown_value` · failure: none — a token-shaped key split by an invisible character is no longer echoed whole.
- REQ-6: Every acceptance screen of a settings value (a `claudeMdExcludes` entry; the `model` value; the other screened values) shall judge the text as written and the text with invisible characters removed · failure: none — a token split by U+200B is refused and nothing is planned.
- REQ-7: The invisible-character set shall name each near-invisible class as put in or left out, with the reason, in the SPEC decisions of the project-scale settings SPEC · failure: none — a class left out (spaces, combining marks, look-alike letters) is written down as a remaining risk, not forgotten.
- REQ-8: The test of the invisible-character screen shall pin one split-token point for each pattern `credential_like` matches · failure: none — a pattern removed from the screen fails its own point.
- REQ-9: A refusal that removed invisible characters shall say `(invisible characters removed)` · failure: none — a stripped name no longer reads as an allowed one (display only).
- REQ-10: The process-scan snapshot slice of `SPEC-20261121-process-scan-snapshot` shall be present on the integration branch with its three deletion-safety tests: a working-directory question reads the cwd list only, a re-check right before a delete or move reads fresh, and a time-limited `lsof` is unknown and refuses · failure: none — a conflict in `src/worktree_gc.rs` is resolved by the lander of this card and audited at tier 2.

## Acceptance

- AC-1 (REQ-1): none of the three templates holds a retired icon and each names the flow icons it uses · verify: `cargo test --locked --offline --test style_banner_scan the_three_style_templates_use_the_one_icon_set`
- AC-2 (REQ-2, REQ-3): the three documents carry the corrected sentences and rows of commit `9706de1802` and the cherry-pick applies with no conflict on the tip · verify: `git diff --check`
- AC-3 (REQ-4): the check fails on a failing git and passes on an absent `.codex` entry; the old `rev-parse` form fails the same test when put back (guard mutant) · verify: `cargo test --locked --offline --test retired_host_words the_repository_tracks_no_codex_folder`
- AC-4 (REQ-5, REQ-6, REQ-8, REQ-9): the four tests of the C4 patch pass and the lib test of the unknown key passes · verify: `cargo test --locked --offline --test claude_settings_scale_keys`
- AC-5 (REQ-7): the decisions D-6 to D-8 of `SPEC-20261122-project-scale-settings` list each near-invisible class as in or out · verify: `moai spec check SPEC-20261122-project-scale-settings`
- AC-6 (REQ-10): the snapshot slice passes its own tests on the tip: the cwd-only scan, the fresh re-check of a delete and a move, and the time-limit refusal, and the two earlier survivors MB and MH stay killed · verify: `cargo test --locked --offline --test process_scan_snapshot_cli`
- AC-7 (REQ-10): the hygiene re-check test and the cost test of the two neighbours pass · verify: `cargo test --locked --offline --test hygiene_apply_cli`

## Scope

- write: `.moai/specs/SPEC-20261124-cleanup-leftovers/**`, `templates/claude/output-styles/moai.md`, `moai-easy.md`, `moai-learn.md`, `tests/style_banner_scan.rs`
- write: `docs/support-matrix.md`, `docs/host-support-matrix.md`, `docs/product-naming-inventory.md`, `tests/retired_host_words.rs`, `.moai/specs/SPEC-20261122-remove-codex-shipped-text/spec.md`
- write: `src/claude.rs` (the value screens only, new lines), `tests/claude_settings_scale_keys.rs`, `.moai/specs/SPEC-20261122-project-scale-settings/spec.md`
- write: `src/worktree_gc.rs`, `src/hygiene.rs`, `tests/process_scan_snapshot_cli.rs`, `tests/process_scan_cost_cli.rs`, `tests/hygiene_apply_cli.rs`, `.moai/specs/SPEC-20261121-process-scan-snapshot/**`
- read-only: everything else; `docs/rust-product-path.md` stays with its holder
- forbidden: push, pull request, merge of a moved tip, the real HOME, the user's global files, card numbers in SPEC ids, branches and commits

## Plan

Order of slices, one commit group each, so each can be audited and landed alone:

1. Slice A, REQ-1 (card criterion 6): RED `the_three_style_templates_use_the_one_icon_set` fails on the retired icons of the three files; GREEN rewrites the three icon sentences with the flow set; one test only, no source change.
2. Slice B, REQ-2 to REQ-4 (card criterion 3): `git cherry-pick 9706de1802` onto this branch (a dry merge applied with no conflict); re-run `retired_host_words`; run the guard mutant (put the old `rev-parse` form back, the test must fail). The old commit's `Moai-Session` trailer is the earlier lane's; the commit gets this session's trailers when the check demands them.
3. Slice C, REQ-5 to REQ-9 (card criterion 5): the written but never compiled patch `c334c4.patch` of the earlier lane applies on the tip (files: `src/claude.rs`, `tests/claude_settings_scale_keys.rs`, the settings SPEC); RED on the base copy, GREEN, mutants N1 raw key, N2 raw only, U+2800 left out, N5 note removed. Risk named by that lane: the added "(invisible characters removed)" text may change wording that other tests read, so the neighbours of `claude_settings_cli` run too.
4. Slice D, REQ-10 (card criterion 2): merge the branch head `7059238965` of `spec/20261121-process-scan-snapshot` (audit PASS head `e803e25056` for the test delta; the `src` rounds passed earlier) onto the tip. The merge conflicts in `src/worktree_gc.rs` (dry run `git merge-tree`: one conflict, `hygiene_apply_cli` and `process_scan_cost_cli` merge cleanly). The conflict is resolved once and the result is a tier 2 change (it guards deletions), so it goes to the independent audit with the guard-mutant lines MB and MH re-run on the merged tree.

### Deferred (not part of the four slices)

- Criterion 1, C-239 AC2 (nested `.gitignore` in `newly_visible`): blocked. `newly_visible` and its caller are not in the integration branch (`git grep newly_visible NeXT` finds nothing); they arrive with the C-174 layout slice (audit PASS head `ce8e7e7513`, card C-174 still `ready`, not an ancestor of `NeXT`). The fix reads nested `.gitignore` files inside that function, so it can start only after C-174 lands. Resume condition: `git merge-base --is-ancestor ce8e7e7513 NeXT` exits 0. Result state until then: `blocked`.
- Criterion 4, C-331 AC2 (improvement reports at install and update): `not_run`, after RC1 by the leader's dispatch.

## Risks

- Slice D touches the deletion guards (`src/worktree_gc.rs`, `src/hygiene.rs`); a wrong conflict resolution can let a busy folder be deleted. The dangerous path runs in a throwaway: a held-open folder is refused by plan, by re-check and by a time-limited `lsof`.
- Slice C is a patch that was never compiled; an expected compile error or a changed refusal wording is repaired inside slice C, not outside its scope.
- The scope above is wide because one card holds six slices. `src/claude.rs`, `src/worktree_gc.rs` and `src/hygiene.rs` are shared files: a holder of another card that needs them gets a line-lease; this card writes only the lines named in the slice.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Criterion 1 waits for C-174; the lane does not copy C-174 code to work around it. | `git grep newly_visible NeXT` finds none; scope discipline. | card open |
| D-2 | Criterion 4 stays `not_run` until after RC1. | the dispatch of the leader, 2026-11-24. | card open |
| D-3 | Slice B cherry-picks the earlier commit as written; its docs text is not edited again. | audit 6935 B1 to B4 are still true on the tip (`docs/support-matrix.md:111`). | revert the commit |
| D-4 | Slice C takes the earlier patch's decisions D-6 to D-8 of the project-scale settings SPEC as the proposal for approval (D-6 judge both texts, D-7 per-class list, D-8 display note). | note 7032 of the source card. | card open |

### Evidence

Read on 2026-11-24 from the card ledger and the repository at `546cd451a2`:

| Criterion | Source read | Finding |
| --- | --- | --- |
| 6 | C-372 notes 7134, 7169 | three files still hold ⬜ 🟢 ⏸️ 🔵 (`moai.md:17`, `moai-easy.md:15`, `moai-learn.md:15`) |
| 3 | C-292 notes 6935, 7000, 7008 | commit `9706de1802` exists locally, is not in `HEAD`, and cherry-picks without conflict |
| 1 | C-239 notes 4638, 4833, 4932; C-174 notes 4597, 4638 | note 4833 is the slice 1 merge note; the open part is N5 of audit 4638, and its code is not on `NeXT` |
| 5 | C-334 notes 6840, 6842, 7032 | C3 (`609a497d1c`) is in `NeXT`; patch `c334c4.patch` (21031 bytes) exists in its worktree |
| 2 | C-275 notes 5717, 5733, 6743, 7062 | slice 2 head `7059238965` is not in `NeXT`; a dry merge has one conflict in `src/worktree_gc.rs` |
| 4 | C-331 | not read; `not_run` by the dispatch |

## Result

(not started; this SPEC is a draft and waits for approval)
