---
id: SPEC-20261123-worktree-slots
title: "Worktrees are fixed slots warmed from a seed by an APFS clone, under caps, and cleaned when their work ends"
status: draft
revision: 8
created: 2026-11-23
---

# Worktrees are fixed slots warmed from a seed by an APFS clone, under caps, and cleaned when their work ends

## Intent

User decision 2026-11-22 (recorded in the ledger; redesign `~/.moai/reports/moai-adk/agentic-redesign-2026-11-22.md` §3A; research `worktree-practice-research-2026-11-22.md` §6; review `audit-landing-process-review-2026-11-22.md` D3). One machine had 465 to 520 linked worktrees, 155 GB under `.claude/worktrees` and 318 GB under the retired home build folder. Every new worktree started with an empty `target/` and rebuilt every dependency. The decision:

- Fixed slots `.claude/worktrees/lane-01` to `lane-10`, `audit-1` to `audit-3` and `seed`; only subagents use one-off worktrees.
- A new or reset slot is warmed with `cp -c -R -p seed/target`. The seed is rebuilt incrementally after each landing.
- Caps: 25 worktrees, 300 GB under `.claude/worktrees`, and 15% of the disk free, judged with `df`.
- Cleanup on landing, on branch deletion (a `reference-transaction` hook only queues it), on an audit verdict, after 3 days unused, and when a cap is exceeded.
- Safe criteria: unlocked; no process or claim inside; no nested worktree; `status --porcelain --untracked-files=all` empty; HEAD in the integration branch or on the remote.
- `git worktree remove` without `--force`, then `branch -d`, then `prune`. Preview first.

Measured facts this rests on (research §4, [strong]):
- An APFS clone keeps modification times and costs almost no disk.
- Cargo then reports the dependencies `Fresh` and recompiles only the workspace crate.
- `git worktree remove` without `--force` deletes the ignored `target/` and refuses untracked or modified files.

Existing code covers part of the cleanup: `worktree-gc-plan` / `worktree-gc-apply` (merged, dirty, untracked, ignored-outside-build, in-use by cwd, recent change, locked, a digest-checked apply). This SPEC adds the slots, the warming and the caps, then the rules and triggers that code lacks.

Readers: lanes and the leader, who create, reuse and clean worktrees, and the auditor.

## Non-goals

- Sharing one `target/`, `CARGO_TARGET_DIR` or build-dir between worktrees (cargo#12516), or symlinking `target` (research §7).
- Rewinding modification times so Cargo skips our own crate (research §6(b)3).
- A `WorktreeCreate` hook (it disables Claude Code's own subagent cleanup).
- Moving or deleting existing worktrees under `~/.moai/worktrees` or `.moai-cli/worktrees` outside the safe criteria and the user's pick.
- Spotlight or XProtect settings, which belong to the user's OS settings.

## Requirements

- REQ-1: The slot command shall accept only the fixed slot names `lane-01` to `lane-10`, `audit-1` to `audit-3` and `seed`, each at `.claude/worktrees/<name>` of the primary checkout.
- REQ-2: The slot plan shall read, without writing, every fixed slot with whether it exists, its HEAD, its branch or `detached` and whether its `target/` exists, and the caps with their measured values.
- REQ-3: If an ensured slot does not exist, then the slot command shall add it detached at the given commit with `git worktree add --detach`.
- REQ-4: If an ensured slot has no `target/` while the seed has one, then the slot command shall copy the seed's `target/` into it with `cp -c -R -p` on macOS or a recursive copy keeping times elsewhere, naming the method in its answer.
- REQ-5: If the linked worktree count is already at the worktree cap or the free share of the disk is below the floor, then the slot command shall refuse to add a slot and name the cap, its limit and the measured value.
- REQ-6: If an ensured slot already exists, then the slot command shall report it as present with its checkout as it was, warming it only when it has no `target/`.
- REQ-7: The cleanup plan shall never offer a fixed slot for removal or for a trim of its cargo output.
- REQ-8: The cleanup plan shall skip a worktree that holds another linked worktree inside it or a live local claim, and name the reason.
- REQ-9: The cleanup plan shall offer a merged one-off worktree unchanged for 3 days for removal under the existing safe criteria.
- REQ-10: If a cap is exceeded, then the cleanup plan shall order its removal candidates oldest first with the number that brings the count back under the cap.
- REQ-11: If a branch ref is deleted, then the reference-transaction hook shall only append one line naming the ref to a queue file.
- REQ-12: If a card lands or its verdict is posted, then the landing and verdict steps shall apply the cleanup plan for that card's worktrees on a matching digest.
- REQ-13: If a slot path is a symbolic link, then the slot command shall refuse it before any write.
- REQ-14: If the cleanup apply is about to act on a worktree that a claim now names, or that now holds another worktree, or whose claims cannot be read, then the cleanup apply shall refuse that worktree.

## Acceptance

- AC-1 (REQ-1, REQ-2): in a temporary repository, the plan lists the 14 slot names in order, with absent slots marked absent; an unknown name such as `lane-11`, `../x` or `seed2` is refused; the plan leaves the repository's worktree list unchanged.
  verify: `cargo nextest run --locked --offline --test worktree_slot_cli`
- AC-2 (REQ-3, REQ-4, REQ-6): ensuring `lane-01` at a commit adds a detached worktree there. With a seed that has `target/`, the slot's `target/` holds the seed's files with the same modification times, and the answer names the copy method. Ensuring it again changes nothing.
  verify: `cargo nextest run --locked --offline --test worktree_slot_cli`
- AC-3 (REQ-4): on macOS, a temporary crate with a dependency outside the repository, built in the seed and then built in a slot ensured from it, compiles only the crate and reports the dependency `Fresh`.
  verify: `cargo nextest run --locked --offline --test worktree_slot_cli`
- AC-4 (REQ-5): with the worktree cap or the free-share floor set below the measured value through the test seams, ensuring a new slot is refused, the message names the cap and both numbers, and no worktree is added.
  verify: `cargo nextest run --locked --offline --test worktree_slot_cli`
- AC-5 (REQ-7, REQ-8, REQ-9, REQ-10): a preview over a temporary repository keeps a fixed slot (`fixed_slot`), a worktree holding a nested worktree (`nested_worktree`) and a worktree an active card reservation names (`live_claim`); a merged clean worktree unchanged for 3 days stays a removal candidate; the `cap` block orders the candidates oldest first with the number needed (slice 2).
  verify: `cargo nextest run --locked --offline --test worktree_gc_slots`
- AC-7 (REQ-13, REQ-14, REQ-5, REQ-6): the follow-ups of the slice 1 and 2 verdict, in temporary repositories.
  - A linked slot path is refused, and nothing is written through it.
  - Re-warming a present slot meets the free-disk floor.
  - Leftover temporary copies of a lane slot or the seed are removed and named in `cleaned`; in an audit slot (`audit-1` to `audit-3`, the names of `audit_slot::AUDIT_SLOTS`) `ensure` leaves them, and `audit-slot` names them from the slot folder, ignored ones included, and `done --discard` removes them, a link as a link.
  - A held lock makes ensure wait the configured time, then refuse.
  - apply leaves a claimed worktree and a nested one, whether the claim or the nesting came before or after the plan.
  - The act-time re-check refuses a nested worktree and unknown claims.
  - An unreadable profile keeps every linked worktree.
  verify: `cargo nextest run --locked --offline --test worktree_slot_cli --test worktree_gc_slots`
- AC-6 (REQ-11, REQ-12): a deleted branch adds one queue line and removes nothing; a landing cleans that card's worktrees on the plan's own digest; a verdict does the same through the same function (slice 4).
  - A deleted branch adds one line `refs/heads/<name>` to `moai/worktree-cleanup-queue` in the common git folder, only at the committed state; a created or moved branch, a symbolic write, a tag delete and an aborted delete add none; the worktrees, their `target/` and the other branches stay; deleting the integration branch is still refused.
  - `card land` cleans the closed card's worktrees through the cleanup plan and its apply on the digest of that plan; a worktree under `.claude/worktrees`, a card that stays open and another card's worktree keep theirs; the queue is read in the same call: a ref that is gone brings the worktrees of the cards that held that branch, a ref that still exists only loses its line, and the queue file is used up; a cleanup that cannot run leaves the landing as it is and the queue with its lines.
  - Verdict half: the function `worktree_cleanup::run` with `Trigger::Verdict` is the entry the verdict step calls; no verdict command exists yet, so this half is `not_run` here.
  verify: `cargo nextest run --locked --offline --test worktree_cleanup_triggers --test card_land_cli`

## Scope

- slice 1 write: `.moai/specs/SPEC-20261123-worktree-slots/spec.md`, `src/help.rs` (one table row and its examples), `tests/fixtures/help/all*.txt` (regenerated under a recorded leader decision), `src/worktree_slot.rs` (new; the git calls must reach `identity::scrub_git_env`, which only the library sees), `src/lib.rs` (one `pub mod` line), `src/ops/worktree_slot.rs` (new), `tests/worktree_slot_cli.rs` (new), `docs/design-review/appendix/operations/worktree_slot.md` (new)
- slice 2 write: `src/worktree_gc.rs`, `src/ledger/reservation.rs` (one read-only query), `src/ops/worktree_gc_plan.rs` (`--max-worktrees`), `src/ops/worktree_gc_apply.rs` (the new `Pace` field), `tests/worktree_gc_slots.rs` (new), `docs/design-review/appendix/operations/worktree_gc_plan.md`
- slice 3 write: `src/worktree_slot.rs`, `src/worktree_gc.rs`, `tests/worktree_slot_cli.rs`, `tests/worktree_gc_slots.rs`, `tests/audit_slot_cli.rs`, `docs/design-review/appendix/operations/worktree_slot.md` (the follow-ups of the slice 1 and 2 verdict)
- slice 4 write: `src/worktree_cleanup.rs` (new), `src/lib.rs` (one `pub mod` line), `src/card/guard.rs` (the queue part of the hook text), `src/card/land.rs` (the call after the close and the `cleanup` key), `src/ledger/reservation.rs` (two read-only queries), `src/audit_slot.rs` and `src/worktree_slot.rs` (the follow-ups N1 and N2), `tests/worktree_cleanup_triggers.rs` (new), `tests/card_land_cli.rs`, `tests/audit_slot_cli.rs`, `tests/worktree_slot_cli.rs` and the two operation documents
- not in this slice: the landing-cleanup paragraph of `skills/moai/references/flows/next-flow.md`. That file has 847 bytes of room under its size cap, and a change to it also changes `skills/moai/references/source-manifest.json`, which this slice does not reserve. It goes to a slice whose scope holds both.
- not in this SPEC: a hook installer (the managed-hook installer of another SPEC adds the hook text as a managed name; installing it in the real repository is a change of repository protection and waits for the user's pick), a verdict command (another SPEC), the seed rebuild after a landing and the entry gates
- read-only: everything else

## Decisions

- D-1: The slot names are fixed in code, not configured, so a typo can never create a 15th slot; the count matches the user decision (10 lanes, 3 audits, 1 seed).
- D-2: A slot is created detached at a commit; the lane switches its own branch inside it. A slot never owns a branch, so deleting a card branch never strands a slot.
- D-3: Warming copies only when the slot has no `target/`: a present `target/` is the slot's own warm build and is never overwritten (REQ-6).
- D-4: The 300 GB cap is not enforced in slice 1. `du` overcounts APFS clones (research §4 experiment 1), and an honest per-folder figure needs block-sharing data that `df` does not give. Slice 1 enforces the worktree count and the free share from `df -P -k`. The plan reports the 300 GB cap as `not_measured`; slice 2 measures it from the trims it makes.
- D-5: Caps are checked by the command that adds a slot. Entry gates in the launcher and a `PreToolUse` hook belong to a later slice.
- D-6: Test seams (`--max-worktrees`, `--min-free-percent`) set lower limits for the refusal tests; the defaults are 25 and 15.
- D-7: The help table row and its examples are in slice 1: the binary's `help_rules` tests require every command in the table with an `Examples:` section. The four `help --all` goldens (`tests/fixtures/help/all*.txt`) are reserved by another card; a recorded leader decision on this card lets slice 1 regenerate them from the tree merged with the integration tip, and the diff is the one `worktree-slot` row.

- D-8: Slice 2 follows option (b) of the leader's recorded decision on the lane's deletion-scope question (category guarded): its rules only skip or report, and no path becomes removable that was not before. The Q8 preservation of every `.claude/worktrees/` path and its test `a_worktree_under_claude_worktrees_is_preserved` stay. Option (a) (this repository's own `.claude/worktrees/` under the normal safe criteria, slots trim-only) waits for the user's pick and becomes a slice of its own.
- D-9: A claim is an active card reservation (`held`, `expired_candidate`, `held_for_review`) whose `worktree` names the path. A `held` reservation whose lease has ended still counts, because a lane's lease can end during a long run. A ledger that exists but cannot be read keeps every linked worktree (`claim_check_unavailable`); a project without a profile or a ledger has no claims.
- D-11: Only a missing profile means no claims. A profile that exists but cannot be read is `claim_check_unavailable`, as an unreadable ledger is. This replaces part of D-9.
- D-12: `apply` reads the nesting (from the worktree list it read) and the claims again right before it acts on each entry, beside the existing busy re-check. A claim or a nested worktree that appears after the plan already changes the recomputed digest. So the act-time re-check covers the window between the digest match and the action.
- D-13: The lock wait is shortened by `MOAI_WORKTREE_SLOT_LOCK_WAIT_SECS`, for tests. The default stays 300 seconds.
- D-14: Under the slot lock, a present slot's `.target-warming-*`, `.target-inherit-*` and `.target-old-*` folders belong to a killed run, because no other run copies while the lock is held. ensure removes them and names them in `cleaned`, except in an audit slot, whose leftover copy audit-slot names and discards with `done --discard`.
- D-15: The hook queues at the `committed` state only, because a delete that is prepared can still be aborted. The line is the full ref and nothing else; a delete of a ref that still exists (a pack or a prune sends `value → zero` while a packed copy stays) and a ref seen twice are possible, so a line is a hint and the reader checks the ref again (D-16). The loop uses only shell built-ins, and its `read` is written `IFS=' ' read` so the existing check that no line loop forks per line still finds the one loop of the guard.
- D-16: `worktree_cleanup::run` takes the queue by renaming it to `<queue>.draining-<pid>`, so a hook that appends meanwhile writes a new queue and loses no line. A ref that still exists loses its line. For a ref that is gone, the worktrees of every card whose reservation named that branch are added to the card's own. If the cleanup does not finish, the lines are appended back; a queue that cannot be read is renamed back.
- D-17: The cleanup hands the plan only registered linked worktrees of this clone (`only`), runs the plan, and runs the apply on the digest the plan gave, which the apply recomputes before it writes anything. It adds no way to remove a path: the safe criteria, the preserved `.claude/worktrees/` folder (D-8) and the fixed slots are the plan's, so in this repository a lane worktree under `.claude/worktrees/` is still kept until the user picks option (a) of D-8.
- D-18: The landing calls the cleanup after the card closed, never before and never for a card that stays open, and a cleanup that fails never undoes the landing: its answer carries `error`, and the queue keeps its lines. The verdict half is the same function with `Trigger::Verdict`; it is not wired to a command (the leader's recorded decision on this card), and the verdict command of another SPEC calls it. `MOAI_WORKTREE_CLEANUP_NOW_UNIX` sets the cleanup's clock, for tests.
- D-10: REQ-9 needs no new code: a merged, clean worktree whose newest change is older than the existing 120-minute window is already a `remove` candidate, and 3 days is past it. The control test measures it with the clock three days ahead.

## Plan

1. This SPEC.
2. Slice 1, tests first (RED): `tests/worktree_slot_cli.rs` AC-1 to AC-4 on temporary repositories; then `src/ops/worktree_slot.rs` (`plan`, `ensure`), its operation document, the help row and the regenerated help goldens.
3. Slice 2, tests first: the cleanup rules of REQ-7 to REQ-10 in `src/worktree_gc.rs`.
4. Slice 3 and slice 4, tests first: the follow-ups of the verdicts, the queue hook, the landing cleanup and the verdict entry. The seed rebuild after a landing and the entry gates are later slices.
5. Per slice: light checks, one affected run, then the READY note with head and base.

## Risks

- A clone shares blocks until written; a full disk can make a later write fail with ENOSPC (clonefile(2)). The free-share floor guards it.
- A slot reused for a new card keeps the old card's build: Cargo rebuilds what changed by modification time, which a fresh checkout of changed files always updates.
- Two sessions ensuring the same slot at once: an exclusive lock on a file in the git common directory serializes `ensure`, so the second run finds the slot present. Without it, both runs added and warmed the slot (the audit measured a doubled registration and a nested `target/target`).

- A queue line can be false (a pack or a prune of a ref that still exists) or doubled; D-15 and D-16 make the reader check the ref and the plan decide, so a false line costs one read and removes nothing.
- The hook text is also the integration-branch guard: its tests (`card_guard_cli`) must stay green with the queue part, and it must still start the moai program only for a transaction on the recorded branch.

## Result

Slice 1, measured 2026-11-23 on macOS (APFS).
- AC-1 met: `plan_lists_the_fourteen_slots_in_order_and_writes_nothing` and `ensure_refuses_a_name_that_is_not_a_fixed_slot`.
- AC-2 met: `ensure_adds_a_detached_slot_warms_it_from_the_seed_and_leaves_it_alone_after`. The copied file keeps the seed file's 3-day-old modification time, and the method is `clone` on macOS. `ensure_without_a_seed_build_adds_the_slot_cold` covers the cold case.
- AC-3 met: `a_slot_warmed_from_the_seed_compiles_only_the_workspace_crate`. `cargo build -v` in the warmed slot prints `Fresh slotdep` and `Compiling slotapp`, and no `Compiling slotdep`.
- AC-4 met: `ensure_refuses_past_the_worktree_cap_or_under_the_free_floor_and_adds_nothing` and `ensure_adds_right_at_the_caps` (N against N+1 for both caps). The two boundary mutants (`>=` to `>` on the count, `<` to `<=` on the free share) each fail one test.
- Runs: red 6 run, 0 passed; green 7 run, 7 passed. `help_layout` and `git_env_scrub` pass with the new module (20 with the slot tests before the boundary test).
- AC-5 belonged to slice 2 and is met there (below). AC-6 belongs to slice 4 and is met there, its verdict half excepted.
- The first check affected (on cb078b1009) failed on the help table, the examples, the home isolation scan and the test identity list. The fixes are the help row with examples, TempHome in the tests and a listed test address. After one merge of the integration tip 6f4117db80, the four goldens gain only the `worktree-slot` row (4 files, 5 insertions).

Slice 2, measured 2026-11-23 on macOS, under option (b) (D-8).
- AC-5 met: `tests/worktree_gc_slots.rs`, 8 passed. The tests cover the control remove three days on, `nested_worktree`, `live_claim` and its end on release, `fixed_slot` only in the slot folder, an unreadable ledger giving `claim_check_unavailable`, and the `cap` block (oldest first, `needed`, `enough` at over equal to the candidates, outside the digest). `tests/worktree_gc.rs` still passes 14 of 14, the Q8 test included.
- Red: 7 run, 2 passed (the two controls), 5 failed. Before this slice, a worktree that a held reservation named was a `remove` candidate.
- Mutants, all killed: the nested check matching itself; the claim match inverted; claim_check_unavailable switched off; the age sort reversed; `needed` without the minimum; `enough` with `>`; a slot name matched outside the slot folder.

Slice 1 audit repair, measured 2026-11-23 on macOS.
- The audit (FAIL, one blocking) found that concurrent ensure runs of one slot both added and warmed it. Repair: one lock per repository around check, add and warm, and a warm copied to a temporary name and renamed into place only after `cp` succeeded.
- Also repaired:
  - A slot that has `target/` is never copied into.
  - A present slot without `target/` is warmed (REQ-6 revised).
  - A failed warm fails the command and leaves no `target/`.
  - A registered slot whose folder is gone is reported `missing` and refused.
  - A linked seed `target` or a seed that is not a worktree is refused before anything is added.
  - A failing `df` refuses.
  - git failures keep the `fatal:` line.
- `tests/worktree_slot_cli.rs`: 14 passed, 7 of them new. One of them runs three concurrent ensures of one slot for three slots.
- Mutants, all killed: the lock removed; the guard on an existing `target/` removed; the temporary copy kept after a failure; the `df` exit status ignored.

Follow-ups of the slice 1 and 2 verdict, measured 2026-11-23 on macOS (before AC-6).
- AC-7 met. Integration tests: 43 of 43 passed across `worktree_slot_cli`, `worktree_gc_slots` and `worktree_gc`. Unit tests of `worktree_gc` and `worktree_slot`: 18 of 18 passed.
- The first runs failed in two fixtures, not in the code: two cards reserved one scope path, and a trim control had no `target/`. Both fixtures were fixed.
- Six mutants were killed, each by its targeted test: the link refusal removed; the re-warm floor removed; the leftover cleanup removed; the act-time nesting check removed; unknown claims read as none at act time; a profile read error read as no profile.
- REQ-7 was reworded to match the code: a fixed slot is offered neither for removal nor for a trim. The stale slice 1 line and the doc's `failed` warm value were fixed.
- The commit messages that cite a decision number were left as they are; new ones carry none.
- Branch: from the audit-slot head, because both change the slot module.

Slice 4, measured 2026-11-23 on macOS (APFS), from the integration tip.
- AC-6 met for the hook and the landing: `tests/worktree_cleanup_triggers.rs` (4 tests: a deleted branch queues one line and removes nothing, only a committed delete queues, the integration branch stays protected, a cleanup that fails after taking the queue puts every line back) and `tests/card_land_cli.rs` (ac35 to ac38 and the second half of ac36: the closed card's merged worktree is removed on the plan's digest, a preserved folder, an open card and another card's worktree stay, the queue is read and used up, a failed cleanup leaves the landing and the queue).
- AC-6 verdict half: `not_run`. `worktree_cleanup::run` with `Trigger::Verdict` exists and its failure path is tested through it; no verdict command exists to call it, so nothing runs it on a verdict.
- AC-7 follow-ups N1 and N2 met: an ignored `.target-*` copy and a link named like a copy are named once, refused without `--discard` and removed by it in each of `audit-1`, `audit-2` and `audit-3` (a link is removed as a link and the folder it pointed at keeps its file); `ensure` leaves the copy of each audit slot (`cleaned` is `[]`), tested over all three slot names.
- N3 met: the text of AC-7, this Result and the two operation documents name `cleaned` and the audit-slot exception.
- Runs (nextest, `--lib` and the seven changed integration targets, one run): 1823 run, 1823 passed. Integration targets: `worktree_slot_cli` 19, `audit_slot_cli` 19, `worktree_gc_slots` 11, `worktree_gc` 14, `worktree_cleanup_triggers` 4, `card_land_cli` 37, `card_guard_cli` 18; the other 1701 are library tests.
- Mutants, all killed by a test of this slice: the slot-folder read of `audit-slot` removed; the audit-slot rule narrowed to `audit-1`; the landing cleaning a card that stays open; the apply never run; a failed cleanup dropping the queue (the first run of this mutant survived, which is why the failure test was added); a ref that still exists read as gone; every worktree planned instead of the card's; the hook queueing at every state.
