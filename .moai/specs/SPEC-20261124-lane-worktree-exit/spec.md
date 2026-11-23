---
id: SPEC-20261124-lane-worktree-exit
title: "the card loop leaves the worktree after the hand-back, and a lane asks by note, never by a selection prompt"
status: draft
revision: 2
created: 2026-11-24
---

# the card loop leaves the worktree after the hand-back, and a lane asks by note, never by a selection prompt

## Intent

Recurrence prevention. The same class of block happened twice: lane `agent-2` stayed inside the isolated worktree session of its previous card after the READY note, so the next dispatch (once after the lane's first card, once after its second) could not run `git worktree add` — the guard refused it (the blocked-dispatch note of the first and the leader note of the second say so). `CLAUDE.md` flow step 9 and `docs/lane-rules.md` already say "leave with `ExitWorktree(keep)`", but `skills/moai/workflows/work.md`, the text a lane actually reads in its card loop, has no step or sentence for it. The rule must sit where the lane reads, and a test must pin the sentence so it cannot silently drop out.

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

- REQ-1: When a session that works a card has written its READY, REPAIRED or blocked note and given the card back, the WORK workflow shall tell the session to call `ExitWorktree(keep)` and then to write `LANE-DONE` as the last line of the turn, before it takes or enters any other worktree.
- REQ-2 (REQ-1): IF the card's worktree holds an uncommitted change when the session is about to write a READY, REPAIRED or blocked note, THEN the WORK workflow shall tell the session to commit it before it writes the note, never to commit after a READY or REPAIRED note, and never to remove the worktree.
- REQ-3: The WORK workflow shall tell a session that works a card never to open a selection prompt and to put what it needs decided in a `LANE-BLOCKED kind=decision` note.
- REQ-4 (REQ-3): The WORK workflow shall make that statement without naming a host question tool.

## Acceptance

- AC-1 (REQ-1): `work.md` has one hand-back paragraph that calls `ExitWorktree(keep)` after the note, the card's hand-back and the message to the leader, then names `LANE-DONE` as the last line of the turn, and `EnterWorktree` with a `path` for the next card, and it stands after the paragraph that writes the READY note · verify: `cargo nextest run --locked --offline --test next_flow_skill_scan work_leaves_the_card_worktree_after_the_hand_back`
- AC-2 (REQ-2): the same paragraph puts the `git status --short` check before the note, forbids a commit after a READY or REPAIRED note, and forbids removing the worktree, in the order of `docs/lane-rules.md` · verify: `cargo nextest run --locked --offline --test next_flow_skill_scan work_leaves_the_card_worktree_after_the_hand_back`
- AC-3 (REQ-3): the "Exceptions and Permissions" section says that a session that works a card never opens a selection prompt and names the `LANE-BLOCKED kind=decision` note · verify: `cargo nextest run --locked --offline --test next_flow_skill_scan work_asks_by_note_and_never_by_a_selection_prompt`
- AC-4 (REQ-4): `work.md` still names none of `AskUserQuestion`, `request_user_input` and `question tool` · verify: `cargo nextest run --locked --offline --test next_flow_skill_scan work_decision_becomes_a_blocked_note_on_the_card`
- AC-5 (REQ-1): the `sha256` of `workflows/work.md` in `source-manifest.json` matches the file · verify: `cargo nextest run --locked --offline --test source_manifest_facts`

## Scope

- write: `.moai/specs/SPEC-20261124-lane-worktree-exit/spec.md`, `skills/moai/workflows/work.md`, `skills/moai/references/source-manifest.json`, `tests/next_flow_skill_scan.rs`
- read-only: `CLAUDE.md`, `docs/lane-rules.md`, `tests/lane_completion_handshake.rs`, `tests/lane_next_rules.rs`, `tests/source_manifest_facts.rs`
- forbidden: `tools/repo_scans.json`

Only the sha256 line of the work workflow changes in the manifest, and no other file under the skill folder is touched. `tools/repo_scans.json` stays as it is because no new file is read (D-3).

## Plan

1. RED: two tests in `tests/next_flow_skill_scan.rs` — `work_leaves_the_card_worktree_after_the_hand_back` and `work_asks_by_note_and_never_by_a_selection_prompt`; both fail on the current `work.md`.
2. GREEN: one paragraph after the hand-back paragraph of `work.md` (revision 2 order of `docs/lane-rules.md`: the `git status --short` check before the note, no commit after a READY or REPAIRED note, `ExitWorktree(keep)` after the note and the hand-back, `LANE-DONE` as the last line of the turn, the next card entered by `EnterWorktree` with the path the leader made or made after the exit) and one sentence in "Exceptions and Permissions".
3. Update the `sha256` of `workflows/work.md` in `source-manifest.json`.
4. Guard mutants: remove each new sentence and see its test fail.
5. SYNC: Result with each AC met or not met, then the READY note with the `check affected` receipt.

## Risks

- `LANE-DONE` must stay the last line the lane writes (`tests/lane_completion_handshake.rs`): `ExitWorktree(keep)` is a tool call that comes before it, and the paragraph says `LANE-DONE` is the last line of the turn, after that tool call.
- A commit after the READY or REPAIRED note makes the branch pass the head the note names; the paragraph forbids that commit and puts the clean-tree check before the note (revision 2, after the first audit).
- The `sha256` in the manifest goes stale with every edit of `work.md`; `source_manifest_facts` notices, and step 3 repairs it.
- A lane that is not in a worktree entered by `EnterWorktree` (entered by a shell `cd`) gets a no-op from `ExitWorktree`; `docs/lane-rules.md` already says to start the next card with `EnterWorktree`, and the new paragraph repeats that the next card starts outside every card worktree.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The exit is a paragraph after the hand-back paragraph, not an eighth procedure step | the hand-back already lives in that paragraph, and `work_steps_run_in_order` pins the seven keys | move the sentence into a step and change the pin |
| D-2 | The selection-prompt sentence avoids every host tool name | the existing pin forbids `AskUserQuestion`, `request_user_input` and `question tool` in `work.md`, because the skill is host neutral | none needed; it adds one sentence |
| D-3 | No `tools/repo_scans.json` entry | the new tests live in `tests/next_flow_skill_scan.rs`, which already reads `work.md` | none |

## Result

Revision 2 measured 2026-11-23 (UTC) on macOS in the worktree `.claude/worktrees/lane-worktree-exit`, code head `759ed8a8f5fd6751493e5b14183897576995fd7f`, base `ec6233e7b4f9ec3c35b1f8f50918c21b25181890` (the `NeXT` tip when the worktree was made; the checks did not follow the tip). Revision 1 landed no commit on `NeXT`; the first audit failed it for two reasons, both repaired here: the paragraph told a lane to commit or name uncommitted work after the READY note, against the order of `docs/lane-rules.md`, and this SPEC's body held seven card numbers.
- AC-1 met: `skills/moai/workflows/work.md` has one paragraph after the hand-back paragraph and before "Exceptions and Permissions" that calls `ExitWorktree(keep)` after the note, the card's hand-back and the message to the leader, then names `LANE-DONE` as the last line of the turn, and starts the next card with `EnterWorktree` and its `path`; `work_leaves_the_card_worktree_after_the_hand_back` compares the five sentences whole and their order. RED at commit `492bf2efdb` against the revision 1 text (exit 100, `work.md lacks …`); GREEN at `759ed8a8f5`.
- AC-2 met: the same paragraph puts the `git status --short` check, which must print nothing, before the READY, REPAIRED or blocked note, forbids a commit after a READY or REPAIRED note, and forbids removing the worktree or leaving it with a shell `cd`; the same test pins it, and no `ExitWorktree(remove)` appears. `docs/lane-rules.md` is unchanged.
- AC-3 met: "Exceptions and Permissions" holds the whole sentence "A session that works a card never opens a selection prompt for the person; whatever it needs decided goes in that note." beside the `LANE-BLOCKED kind=decision` sentence; `work_asks_by_note_and_never_by_a_selection_prompt` pins it.
- AC-4 met: `work_decision_becomes_a_blocked_note_on_the_card` still passes, so `work.md` names none of `AskUserQuestion`, `request_user_input` and `question tool`.
- AC-5 met: `source-manifest.json` holds the new `sha256` of `workflows/work.md` (`61075b668bb5c3cace3ba3c9c2ef1585213dac41f31b64232517319deb122760`, read with `shasum -a 256`); `source_manifest_facts` passes.
- `cargo nextest run --locked --offline` over `next_flow_skill_scan`, `source_manifest_facts`, `lane_completion_handshake`, `lane_next_rules`, `mode_cards_53`, `mode_workflow_steps` and `skill_structure` ran 105 tests, 105 passed (exit 0).
- guard mutants of `work_leaves_the_card_worktree_after_the_hand_back`, each run on the committed head `759ed8a8f5` and restored (`git status --short` and `git diff --stat` empty after the last): sentence 1 removed whole, without `first`, "Before" changed to "After", without `nothing`; sentence 2 removed whole, "no commit" changed to "a commit", "READY or REPAIRED" shortened to "READY"; sentence 3 with `keep` changed to `remove`, without `never`, without `message`; sentence 4 removed whole, without `last`; sentence 5 without `first`, without `path`. All fourteen ended exit 100 and every one was applied (the first run of the `keep` to `remove` mutant did not match, and was rerun with a corrected pattern). The sentences of the no-prompt pin did not change in revision 2, and its three mutants of revision 1 (sentence removed whole, without `never`, without `selection`) stand.
- Gap: revision 2 is a draft; its approval comes from the leader as a decision note, and this lane records none. The 105-test run above is the targeted run, not the `moai check affected` receipt, which the REPAIRED note quotes for the final head.
- Gap: the first draft of this SPEC had no `## Interview` section and `moai spec check` refused it once approved (`interview_missing`); the section was added with the approval commit `c7031c55a5` and the check then answered 0 errors for this SPEC. The digest `moai spec show` prints stayed the same.
- Gap: `ExitWorktree` is a no-op for a worktree first entered by a shell `cd` (`docs/lane-rules.md`); the new text only names `EnterWorktree` for the next card and a test cannot prove a lane obeys it. A third recurrence would need a hook, not text.
