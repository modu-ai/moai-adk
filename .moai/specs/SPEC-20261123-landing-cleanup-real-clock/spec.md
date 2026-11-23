---
id: SPEC-20261123-landing-cleanup-real-clock
title: "A landing cleanup removes the landed card's clean merged worktree on the real clock and keeps what it skipped"
status: draft
revision: 1
created: 2026-11-23
---

# A landing cleanup removes the landed card's clean merged worktree on the real clock and keeps what it skipped

## Intent

The independent audit of slice 4 of `SPEC-20261123-worktree-slots` (the landing cleanup) found four follow-ups that did not block it. All four are one defect seen from different sides: the cleanup passed its tests only because they moved the clock three days ahead.

- F-1: `worktree_cleanup::clean` plans with the default 120-minute recent-change window and, for a worktree with no build output, the 24-hour empty-worktree guard. A lane lands within hours of its last change, so on the real clock the landed card's worktree is skipped as `recent_change` or `empty_worktree`. A closed card never reaches the landing trigger again, so nothing is ever removed.
- F-2: `Taken::done` deletes the whole queue even when every worktree of a queue line was skipped for a temporary reason (recent change, a process inside, a claim), so those lines are lost.
- F-3: the queue is taken by a rename, then read. A line that a hook writes into the renamed file after the read is lost when the file is deleted; the read-failure path renames the file back over a queue a hook has created since; and a `worktree-cleanup-queue.draining-<pid>` file left by a run that died is never read again.
- F-5: the operation document row of `card land`, the `apply` doc comment and the `writes` key of the `card land` answer still describe a landing that removes nothing.

Readers: the leader and lanes who land cards, and the auditor.

## Non-goals

- Widening what the cleanup may remove: the safe criteria, the preserved `.claude/worktrees/` folder and the fixed slots keep their effect. A user pick for removing under `.claude/worktrees` is still open.
- Installing the branch-delete hook in the real repository.
- Wiring the verdict trigger.
- A new trigger for a landed worktree that was skipped for a temporary reason: its answer already names the reasons, and the 3-day removal candidate of the plan offers it later.

## Requirements

- REQ-1: When a landing closes a card, the cleanup shall plan that card's worktrees without the recent-change window and the empty-worktree guard, because every reservation of the card has ended.
- REQ-2 (REQ-1): If another active reservation, a nested worktree, a process inside, an unmerged head, a tracked change, an untracked or non-regenerable ignored file, a linked `target/` or a path under `.claude/worktrees/` applies to a worktree, then the cleanup shall keep it as before.
- REQ-3 (REQ-4): If a worktree of a queue line is skipped for a temporary reason, then the cleanup shall write that line back to the queue.
- REQ-4: When the cleanup reads the queue, it shall also read every `worktree-cleanup-queue.draining-*` file left in the common git folder and remove each one only after its lines are used or written back.
- REQ-5 (REQ-4): If a hook appended a line to the taken queue file after the cleanup read it, then the cleanup shall write that line back to the queue.
- REQ-6 (REQ-4): If the taken queue file cannot be read, then the cleanup shall restore it without replacing a queue file that exists.
- REQ-7: When a landing removes or trims a worktree, the `card land` answer shall report `writes: true`.
- REQ-8: The operation document row of `card land` and the `apply` doc comment shall say that a landing removes the closed card's merged clean worktrees without `--force` and name the `cleanup` key.
- REQ-9 (REQ-7): If the cleanup answers an error or removes nothing, then the `card land` answer shall report `writes` from the ledger and the branch only.

## Acceptance

- AC-1 (REQ-1, REQ-2): in a temporary repository on the real clock (no clock seam), landing a card removes its clean merged worktree, including one with no `target/`; a worktree with a tracked change, an unmerged head, an untracked file, a linked `target/` pointing outside, another card's claim or a `.claude/worktrees/` path is kept and the file outside the worktree is unchanged; `git worktree remove` is never called with `--force`.
  verify: `cargo nextest run --locked --offline --test card_land_cli --test worktree_cleanup_triggers`
- AC-2 (REQ-3): a queue line whose worktree is skipped as `process_in_use`, `live_claim` or `recent_change` is in the queue after the call; a line whose worktree was removed or is kept for a lasting reason (`not_merged`, `tracked_changes`) is gone.
  verify: `cargo nextest run --locked --offline --test worktree_cleanup_triggers --test card_land_cli`
- AC-3 (REQ-4, REQ-5, REQ-6): a leftover draining file is read by the next run and removed after; a line added to the taken file after the read is in the queue after the call; an unreadable queue is restored under its name when no queue exists, and left in place under its draining name when a new queue exists, with no line lost.
  verify: `cargo nextest run --locked --offline --test worktree_cleanup_triggers`
- AC-4 (REQ-7, REQ-8, REQ-9): a landing that removes a worktree answers `writes: true`; `worktree_cleanup::wrote` is true only for an answer with a `removed` or `trimmed` worktree, so a cleanup that removed nothing or failed adds nothing to `writes` (the close itself already writes events); the operation document row and the `apply` comment name the removal and the `cleanup` key.
  verify: `cargo nextest run --locked --offline --test card_land_cli`

## Scope

- write: `.moai/specs/SPEC-20261123-landing-cleanup-real-clock/spec.md`, `src/worktree_cleanup.rs`, `src/worktree_gc.rs`, `src/card/land.rs`, `docs/design-review/appendix/operations/card.md`, `tests/worktree_cleanup_triggers.rs`, `tests/card_land_cli.rs`
- read-only: everything else

## Decisions

- D-1: The landing sets `Pace.settled` to the worktrees of the landed card only. A worktree reached through a queue line keeps the normal window, because its card ended earlier and nothing says how long ago. `settled` removes `recent_change` and `empty_worktree` and nothing else; `live_claim`, the process check and the file census still decide, and the apply re-checks them right before acting.
- D-2: Temporary reasons are `recent_change`, `empty_worktree`, `process_in_use`, `process_check_unavailable`, `file_open_in_use`, `open_file_check_unavailable`, `live_claim`, `claim_check_unavailable` and `budget_spent`. A dirty or unmerged worktree is not temporary: its line is used up as before.
- D-3: The taken file gets a name with the process id and a time, so a second run in one process never replaces a leftover file.
- D-4: A line is written back by appending to the queue path, never by a rename over it.

## Plan

1. Tests first (red): real-clock landing, kept worktrees, requeue, draining files, tail, restore, `writes`.
2. `Pace.settled` and `Facts.settled` in `src/worktree_gc.rs`; `worktree_cleanup` sets it for a landing.
3. Queue handling in `src/worktree_cleanup.rs` (requeue, orphans, tail, restore).
4. `writes` in `src/card/land.rs`; the document row and the `apply` comment.
5. Light checks, one affected run, a run of the dangerous path in a temporary repository, the READY note.

## Risks

- Removing a worktree is not undoable. `settled` therefore never relaxes a guard other than time, and the run of the dangerous path in a temporary repository goes into the READY note.
- A leftover draining file of a run that is still alive is read by a second run; both then plan the same worktrees, and the apply refuses what the other already removed.

## Result

Measured 2026-11-23 on macOS, from the integration tip `0e7ed6190d`, first commit `9cbe1f48b0`.

- AC-1 met. `tests/worktree_cleanup_triggers.rs` `real_clock` (4 new tests, no clock seam): a clean merged worktree goes with or without build output; a tracked change, an unmerged commit, an untracked file, a linked `target`, another card's claim and a `.claude/worktrees` path stay, and the files outside and the bystander file are as they were. A run of the dangerous path with the built binary in a throwaway repository, on the real clock (`dangerous-path.sh`, 7 real `card land` calls): `clean` and `built` removed (`action: remove`, `status: removed`); `dirty` (`tracked_changes`), `unmerged` (`not_merged`), `untracked` (`untracked_files`), `linked` (`ignored_files_outside_build_output`, the `target` link still a link) and `preserved` (`preserved_location`) kept; `outside/keep.txt`, `bystander.txt` and the dirty file unchanged. `git worktree remove` is called without `--force` (the dirty ones were kept, not forced).
- AC-2 met. `a_queue_line_skipped_for_a_temporary_reason_stays_in_the_queue`: the lines of a claimed and a fresh worktree stay in the queue; the line of a removed worktree and of a dirty one is used up. A dirty worktree with a fresh change is kept for a lasting reason, not requeued.
- AC-3 met. Unit tests in `src/worktree_cleanup.rs` (9) and `a_leftover_draining_file_is_read_by_the_next_run`: a leftover taken file is read and removed, a line written after the read goes back, a run that did not finish puts back every line, an unread file does not replace a queue that exists, and an unreadable queue is restored under its name.
- AC-4 partly met. The `card land` row of `docs/design-review/appendix/operations/card.md` and the `apply` doc comment now name the removal and the `cleanup` key. The `writes` value in `src/card/land.rs` and the two landing tests in `tests/card_land_cli.rs` are `not_run`: another card reserves both files (`card take` and `card line-lease` answered `conflict`, nothing was written). The close of a landing already writes events, so `writes` is `true` whenever the cleanup runs; the change only counts the removal itself (`worktree_cleanup::wrote`, already in place and unit-tested).
- Runs: unit `worktree_cleanup` 9 and `worktree_gc` 18 passed; `worktree_cleanup_triggers` 8 of 8; six targets together (`card_land_cli`, `worktree_gc`, `worktree_gc_slots`, `worktree_slot_cli`, `card_guard_cli`, `worktree_cleanup_triggers`) 107 of 107 passed; `cargo clippy --all-targets -- -D warnings` clean; `cargo fmt --check` clean.
- Mutants, all killed by a test of this SPEC: settled never applied (4 integration tests fail with the old `recent_change` skip); `any` instead of `all` temporary reasons (2 tests); the tail after the read ignored (2 unit tests); the restore by rename over an existing queue (1 unit test); skipped lines never written back (1 test); leftover files never read (2 unit tests).
- First red: the first run of the new integration tests failed in three fixtures, not in the code: the in-process plan reads claims under the process `HOME`, and a fake `target/` without `.cargo-lock` is not build output. Both fixtures were fixed; the lasting-reason rule (`all` temporary reasons) came from the dirty-and-fresh case.
