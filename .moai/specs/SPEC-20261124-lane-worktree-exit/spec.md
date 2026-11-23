---
id: SPEC-20261124-lane-worktree-exit
title: "the card loop leaves the worktree after the hand-back, and a lane asks by note, never by a selection prompt"
status: approved
revision: 1
created: 2026-11-24
approval:
  by: leader
  at: 2026-11-23
  basis: "리더 결정 노트 7776 (2026-11-23T18:32:22Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단, 사용자 직접 선택 아님, 노트 7775 에 답함): 판 1 과 선택지 (A) 승인, 쓰기 범위 승인, 새 문단 핀은 문장 전체를 비교; 레인은 그 노트를 인용만 해요"
  digest: "sha256:705fafb24b54007d91cd0850298eaff662fbf82f91b73ad38bf89d474eb7c74e"
---

# the card loop leaves the worktree after the hand-back, and a lane asks by note, never by a selection prompt

## Intent

Card C-422 (recurrence prevention). The same class of block happened twice: lane `agent-2` stayed inside the isolated worktree session of its previous card after the READY note, so the next dispatch (C-359 after C-415, C-420 after C-359) could not run `git worktree add` — the guard refused it (C-359 note 7750, C-420 leader note). `CLAUDE.md` flow step 9 and `docs/lane-rules.md` already say "leave with `ExitWorktree(keep)`", but `skills/moai/workflows/work.md`, the text a lane actually reads in its card loop, has no step or sentence for it. The rule must sit where the lane reads, and a test must pin the sentence so it cannot silently drop out.

The card also asks that the same text confirm the other rule a lane already follows: it never opens a selection prompt for the person and asks by a `LANE-BLOCKED kind=decision` note. `work.md` says "never a question to the user" but not the prompt itself, and an existing pin forbids naming a host question tool in `work.md`.

Readers: lanes that run the card loop, the leader that dispatches the next card, and the auditor of this change.

## Interview

- Q1: Is the exit a new procedure step? -> A: No (D-1). The hand-back already lives in one paragraph of `work.md`, and `work_steps_run_in_order` pins the seven step keys.
- Q2: May `work.md` name the host's selection tool? -> A: No (D-2). `work_decision_becomes_a_blocked_note_on_the_card` forbids `AskUserQuestion`, `request_user_input` and `question tool`; the sentence says "selection prompt" instead.
- Q3: Does a new test file need a `tools/repo_scans.json` entry? -> A: No (D-3). The new tests live in `tests/next_flow_skill_scan.rs`, which already reads `work.md`.

## Non-goals

- Changing `CLAUDE.md`, `docs/lane-rules.md` or `docs/leader-rules.md`: they already carry the rule; this card puts it in the skill text.
- Adding a step to the WORK procedure: the seven steps `take … repeat` stay as pinned by `work_steps_run_in_order`.
- A tool or hook that forces the exit: the card asks for the workflow sentence and its pin only.
- Naming a host question tool in `work.md`: `work_decision_becomes_a_blocked_note_on_the_card` forbids it and stays.
- Removing a worktree: the leader removes landed worktrees (`CLAUDE.md`, "Worktrees and builds").

## Requirements

- REQ-1: When a session that works a card has written its READY, REPAIRED or blocked note and its `LANE-DONE` line, the WORK workflow shall tell the session to call `ExitWorktree(keep)` before it takes or enters any other worktree.
- REQ-2 (REQ-1): IF the card's worktree still holds an uncommitted change, THEN the WORK workflow shall tell the session to commit it or name it in the note before it leaves, and never to remove the worktree.
- REQ-3: The WORK workflow shall tell a session that works a card never to open a selection prompt and to put what it needs decided in a `LANE-BLOCKED kind=decision` note.
- REQ-4 (REQ-3): The WORK workflow shall make that statement without naming a host question tool.

## Acceptance

- AC-1 (REQ-1): `work.md` has one hand-back paragraph that names `ExitWorktree(keep)` after the `LANE-DONE` line and `EnterWorktree` with a `path` for the next card, and it stands after the paragraph that writes the READY note · verify: `cargo nextest run --locked --offline --test next_flow_skill_scan work_leaves_the_card_worktree_after_the_hand_back`
- AC-2 (REQ-2): the same paragraph names the uncommitted-change check and forbids removing the worktree · verify: `cargo nextest run --locked --offline --test next_flow_skill_scan work_leaves_the_card_worktree_after_the_hand_back`
- AC-3 (REQ-3): the "Exceptions and Permissions" section says that a session that works a card never opens a selection prompt and names the `LANE-BLOCKED kind=decision` note · verify: `cargo nextest run --locked --offline --test next_flow_skill_scan work_asks_by_note_and_never_by_a_selection_prompt`
- AC-4 (REQ-4): `work.md` still names none of `AskUserQuestion`, `request_user_input` and `question tool` · verify: `cargo nextest run --locked --offline --test next_flow_skill_scan work_decision_becomes_a_blocked_note_on_the_card`
- AC-5 (REQ-1): the `sha256` of `workflows/work.md` in `source-manifest.json` matches the file · verify: `cargo nextest run --locked --offline --test source_manifest_facts`

## Scope

- write: `.moai/specs/SPEC-20261124-lane-worktree-exit/spec.md`, `skills/moai/workflows/work.md`, `skills/moai/references/source-manifest.json` (the `sha256` of `work.md` only), `tests/next_flow_skill_scan.rs`
- read-only: `CLAUDE.md`, `docs/lane-rules.md`, `tests/lane_completion_handshake.rs`, `tests/lane_next_rules.rs`, `tests/source_manifest_facts.rs`
- forbidden: `tools/repo_scans.json` (no new file is read; the new tests live in a file that already reads `work.md`), every other file of `skills/moai`

## Plan

1. RED: two tests in `tests/next_flow_skill_scan.rs` — `work_leaves_the_card_worktree_after_the_hand_back` and `work_asks_by_note_and_never_by_a_selection_prompt`; both fail on the current `work.md`.
2. GREEN: one paragraph after the hand-back paragraph of `work.md` (exit after `LANE-DONE`, the uncommitted check, the next card entered by `EnterWorktree` with the path the leader made or made after the exit) and one sentence in "Exceptions and Permissions".
3. Update the `sha256` of `workflows/work.md` in `source-manifest.json`.
4. Guard mutants: remove each new sentence and see its test fail.
5. SYNC: Result with each AC met or not met, then the READY note with the `check affected` receipt.

## Risks

- `LANE-DONE` must stay the last line the lane writes (`tests/lane_completion_handshake.rs`): `ExitWorktree(keep)` is a tool call and prints no line, and the paragraph says so.
- The `sha256` in the manifest goes stale with every edit of `work.md`; `source_manifest_facts` notices, and step 3 repairs it.
- A lane that is not in a worktree entered by `EnterWorktree` (entered by a shell `cd`) gets a no-op from `ExitWorktree`; `docs/lane-rules.md` already says to start the next card with `EnterWorktree`, and the new paragraph repeats that the next card starts outside every card worktree.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The exit is a paragraph after the hand-back paragraph, not an eighth procedure step | the hand-back already lives in that paragraph, and `work_steps_run_in_order` pins the seven keys | move the sentence into a step and change the pin |
| D-2 | The selection-prompt sentence avoids every host tool name | the existing pin forbids `AskUserQuestion`, `request_user_input` and `question tool` in `work.md`, because the skill is host neutral | none needed; it adds one sentence |
| D-3 | No `tools/repo_scans.json` entry | the new tests live in `tests/next_flow_skill_scan.rs`, which already reads `work.md` | none |

## Result

Measured 2026-11-23 (UTC) on macOS in the worktree `.claude/worktrees/lane-worktree-exit`, code head `af62ff3ed0932a2ea8773fdd1a8dc0481eeb7f55`, base `ec6233e7b4f9ec3c35b1f8f50918c21b25181890` (the `NeXT` tip when the worktree was made).
- AC-1 met: `skills/moai/workflows/work.md` has one paragraph after the hand-back paragraph and before "Exceptions and Permissions" that calls `ExitWorktree(keep)` after the `LANE-DONE` line and starts the next card with `EnterWorktree` and its `path`; `work_leaves_the_card_worktree_after_the_hand_back` compares the three sentences whole and their order. RED at commit `f33e336886`: both new tests failed (exit 100) with `work.md lacks …`; GREEN at `af62ff3ed0`.
- AC-2 met: the same paragraph reads `git status --short`, says commit or name what is uncommitted, and never remove the worktree or leave it with a shell `cd`; the same test pins it, and no `ExitWorktree(remove)` appears.
- AC-3 met: "Exceptions and Permissions" holds the whole sentence "A session that works a card never opens a selection prompt for the person; whatever it needs decided goes in that note." beside the `LANE-BLOCKED kind=decision` sentence; `work_asks_by_note_and_never_by_a_selection_prompt` pins it.
- AC-4 met: `work_decision_becomes_a_blocked_note_on_the_card` still passes, so `work.md` names none of `AskUserQuestion`, `request_user_input` and `question tool`.
- AC-5 met: `source-manifest.json` holds the new `sha256` of `workflows/work.md` (`304a8d1d18673301cdd6955df39db01cb09fe4c7876b19ef87acc00da7a5d3f8`, read with `shasum -a 256`); `source_manifest_facts` passes.
- `cargo nextest run --locked --offline` over `next_flow_skill_scan`, `source_manifest_facts`, `lane_completion_handshake`, `lane_next_rules`, `mode_cards_53`, `mode_workflow_steps` and `skill_structure` ran 105 tests, 105 passed (exit 0).
- guard mutants, each run on the committed head `af62ff3ed0` and restored (`git diff` of `work.md` empty after each and after the last): exit sentence 1 removed whole, `keep` changed to `remove`, `no` removed from "prints no line"; exit sentence 2 without `never`, without `uncommitted`; exit sentence 3 without `first`, without `path`; the no-prompt sentence removed whole, without `never`, without `selection`. All ten ended exit 100 (the new tests failed); none was left unapplied.
- Gap: the first draft of this SPEC had no `## Interview` section and `moai spec check` refused it once approved (`interview_missing`); the section was added with the approval commit `c7031c55a5` and the check then answered 0 errors for this SPEC. The digest `moai spec show` prints stayed the same.
- Gap: `ExitWorktree` is a no-op for a worktree first entered by a shell `cd` (`docs/lane-rules.md`); the new text only names `EnterWorktree` for the next card and a test cannot prove a lane obeys it. A third recurrence would need a hook, not text.
