---
id: SPEC-20261123-worktree-slots
title: "Worktrees are fixed slots warmed from a seed by an APFS clone, under caps, and cleaned when their work ends"
status: draft
revision: 4
created: 2026-11-23
---

# Worktrees are fixed slots warmed from a seed by an APFS clone, under caps, and cleaned when their work ends

## Intent

User decision 2026-11-22 (ledger C-319 notes 6628 and 6646; redesign `~/.moai/reports/moai-adk/agentic-redesign-2026-11-22.md` §3A; research `worktree-practice-research-2026-11-22.md` §6; review `audit-landing-process-review-2026-11-22.md` D3). One machine had 465 to 520 linked worktrees, 155 GB under `.claude/worktrees` and 318 GB under the retired home build folder. Every new worktree started with an empty `target/` and rebuilt every dependency. The decision:

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
- REQ-6: If an ensured slot already exists, then the slot command shall report it as present with its checkout and its `target/` as they were.
- REQ-7: The cleanup plan shall never offer a fixed slot for removal; a slot is offered only for a detach and a trim of its cargo output.
- REQ-8: The cleanup plan shall skip a worktree that holds another linked worktree inside it or a live local claim, and name the reason.
- REQ-9: The cleanup plan shall offer a merged one-off worktree unchanged for 3 days for removal under the existing safe criteria.
- REQ-10: If a cap is exceeded, then the cleanup plan shall order its removal candidates oldest first with the number that brings the count back under the cap.
- REQ-11: If a branch ref is deleted, then the reference-transaction hook shall only append one line naming the ref to a queue file.
- REQ-12: If a card lands or its verdict is posted, then the landing and verdict steps shall apply the cleanup plan for that card's worktrees on a matching digest.

## Acceptance

- AC-1 (REQ-1, REQ-2): in a temporary repository, the plan lists the 14 slot names in order, with absent slots marked absent; an unknown name such as `lane-11`, `../x` or `seed2` is refused; the plan leaves the repository's worktree list unchanged.
  verify: `cargo nextest run --locked --offline --test worktree_slot_cli`
- AC-2 (REQ-3, REQ-4, REQ-6): ensuring `lane-01` at a commit adds a detached worktree there. With a seed that has `target/`, the slot's `target/` holds the seed's files with the same modification times, and the answer names the copy method. Ensuring it again changes nothing.
  verify: `cargo nextest run --locked --offline --test worktree_slot_cli`
- AC-3 (REQ-4): on macOS, a temporary crate with a dependency outside the repository, built in the seed and then built in a slot ensured from it, compiles only the crate and reports the dependency `Fresh`.
  verify: `cargo nextest run --locked --offline --test worktree_slot_cli`
- AC-4 (REQ-5): with the worktree cap or the free-share floor set below the measured value through the test seams, ensuring a new slot is refused, the message names the cap and both numbers, and no worktree is added.
  verify: `cargo nextest run --locked --offline --test worktree_slot_cli`
- AC-5 (REQ-7, REQ-8, REQ-9, REQ-10): a preview over a temporary repository lists exactly the worktrees that meet the safe criteria, never a slot, not one holding a nested worktree or a claim; apply removes them and their `target/` (slice 2).
  verify: `cargo nextest run --locked --offline --test worktree_gc_slots`
- AC-6 (REQ-11, REQ-12): a deleted branch adds one queue line and removes nothing; a landing or a verdict cleans that card's worktrees on a matching digest (slice 3).
  verify: `cargo nextest run --locked --offline --test worktree_cleanup_triggers`

## Scope

- slice 1 write: `.moai/specs/SPEC-20261123-worktree-slots/spec.md`, `src/help.rs` (one table row and its examples), `tests/fixtures/help/all*.txt` (regenerated, leader decision 6837), `src/worktree_slot.rs` (new; the git calls must reach `identity::scrub_git_env`, which only the library sees), `src/lib.rs` (one `pub mod` line), `src/ops/worktree_slot.rs` (new), `tests/worktree_slot_cli.rs` (new), `docs/design-review/appendix/operations/worktree_slot.md` (new)
- slice 2 write: `src/worktree_gc.rs`, `tests/worktree_gc*.rs`
- slice 3 write: the landing and verdict steps, a `reference-transaction` hook installer, the skill text
- read-only: everything else

## Decisions

- D-1: The slot names are fixed in code, not configured, so a typo can never create a 15th slot; the count matches the user decision (10 lanes, 3 audits, 1 seed).
- D-2: A slot is created detached at a commit; the lane switches its own branch inside it. A slot never owns a branch, so deleting a card branch never strands a slot.
- D-3: Warming copies only when the slot has no `target/`: a present `target/` is the slot's own warm build and is never overwritten (REQ-6).
- D-4: The 300 GB cap is not enforced in slice 1. `du` overcounts APFS clones (research §4 experiment 1), and an honest per-folder figure needs block-sharing data that `df` does not give. Slice 1 enforces the worktree count and the free share from `statvfs`. The plan reports the 300 GB cap as `not_measured`; slice 2 measures it from the trims it makes.
- D-5: Caps are checked by the command that adds a slot. Entry gates in the launcher and a `PreToolUse` hook belong to slice 3.
- D-6: Test seams (`--max-worktrees`, `--min-free-percent`) set lower limits for the refusal tests; the defaults are 25 and 15.
- D-7: The help table row and its examples are in slice 1: the binary's `help_rules` tests require every command in the table with an `Examples:` section. The four `help --all` goldens (`tests/fixtures/help/all*.txt`) are reserved by another card; the leader decision 6837 on this card lets slice 1 regenerate them from the tree merged with the integration tip, and the diff is the one `worktree-slot` row.

## Plan

1. This SPEC.
2. Slice 1, tests first (RED): `tests/worktree_slot_cli.rs` AC-1 to AC-4 on temporary repositories; then `src/ops/worktree_slot.rs` (`plan`, `ensure`), its operation document, the help row and the regenerated help goldens.
3. Slice 2, tests first: the cleanup rules of REQ-7 to REQ-10 in `src/worktree_gc.rs`.
4. Slice 3, tests first: the queue hook, the landing and verdict cleanup, the seed rebuild after a landing, the entry gates.
5. Per slice: light checks, one affected run, then the READY note with head and base.

## Risks

- A clone shares blocks until written; a full disk can make a later write fail with ENOSPC (clonefile(2)). The free-share floor guards it.
- A slot reused for a new card keeps the old card's build: Cargo rebuilds what changed by modification time, which a fresh checkout of changed files always updates.
- Two sessions ensuring the same slot at once: `git worktree add` refuses the second, which reports the slot present.

## Result

Slice 1, measured 2026-11-23 on macOS (APFS).
- AC-1 met: `plan_lists_the_fourteen_slots_in_order_and_writes_nothing` and `ensure_refuses_a_name_that_is_not_a_fixed_slot`.
- AC-2 met: `ensure_adds_a_detached_slot_warms_it_from_the_seed_and_leaves_it_alone_after`. The copied file keeps the seed file's 3-day-old modification time, and the method is `clone` on macOS. `ensure_without_a_seed_build_adds_the_slot_cold` covers the cold case.
- AC-3 met: `a_slot_warmed_from_the_seed_compiles_only_the_workspace_crate`. `cargo build -v` in the warmed slot prints `Fresh slotdep` and `Compiling slotapp`, and no `Compiling slotdep`.
- AC-4 met: `ensure_refuses_past_the_worktree_cap_or_under_the_free_floor_and_adds_nothing` and `ensure_adds_right_at_the_caps` (N against N+1 for both caps). The two boundary mutants (`>=` to `>` on the count, `<` to `<=` on the free share) each fail one test.
- Runs: red 6 run, 0 passed; green 7 run, 7 passed. `help_layout` and `git_env_scrub` pass with the new module (20 with the slot tests before the boundary test).
- AC-5 and AC-6 belong to slices 2 and 3.
- The first check affected (on cb078b1009) failed on the help table, the examples, the home isolation scan and the test identity list. The fixes are the help row with examples, TempHome in the tests and a listed test address. After one merge of the integration tip 6f4117db80, the four goldens gain only the `worktree-slot` row (4 files, 5 insertions).
