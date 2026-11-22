---
id: SPEC-20261123-audit-slot
title: "An audit takes a fixed audit slot, switches it to the head and inherits the lane's build, and leaves it clean"
status: draft
revision: 2
created: 2026-11-23
---

# An audit takes a fixed audit slot, switches it to the head and inherits the lane's build, and leaves it clean

## Intent

User decision 2026-11-22, recorded in the ledger. The sources are the redesign `~/.moai/reports/moai-adk/agentic-redesign-2026-11-22.md` §3D and the review `audit-landing-process-review-2026-11-22.md` A1, A2, A4 and "감사가 레인 빌드를 이어받는 방법".

Problem:
- An audit rebuilt the head the lane had already built: 5 to 11 builds a round, 50 to 70 minutes.
- Each audit made a new isolated worktree (`.claude/worktrees/agent-*`), or used a shared build folder that broke compiles.
- The review measured that an APFS clone (`cp -c`) keeps modification times, so Cargo reports the copied dependencies `Fresh`.
- The fixed slots `audit-1` to `audit-3` exist since SPEC-20261123-worktree-slots.

This SPEC adds `moai audit-slot`:
- `prep` takes an audit slot, switches it detached to the audited head and inherits a build. The build is the lane's `target/` when the lane's worktree is at the same head and clean; otherwise it is the seed's.
- `done` leaves the slot clean for the next audit. It reverts what the audit changed and keeps the warm `target/`.

Readers: the leader and the auditor, who prepare and finish an audit; the lane, whose READY worktree is the build source.

## Non-goals

- Running the audit's build or tests. The auditor runs the targeted targets; `prep` only places the build.
- Sharing one `target/` between worktrees, or a symlinked `target/`.
- Nextest archives shared from the lane. They need the tests to read root and binary paths at run time first (135 files use compile-time `env!`); that is a follow-up card.
- None for the help: the row is in. The four goldens were regenerated under a recorded leader decision, because another card reserves them.
- Removing a slot. Slots are fixed; cleanup rules belong to SPEC-20261123-worktree-slots.

## Requirements

- REQ-1: The audit-slot command shall accept only the slot names `audit-1`, `audit-2` and `audit-3`.
- REQ-2: When an audit slot is prepared for a head, the audit-slot command shall leave the slot detached at that head with no tracked change.
- REQ-3 (REQ-2): If the prepared slot holds a tracked change or an untracked file outside `target/`, then the audit-slot command shall refuse and name the files.
- REQ-4: If the lane worktree given is a worktree of this repository at the head with no tracked change and a real `target/` folder, then the audit-slot command shall replace the slot's `target/` with an APFS clone of the lane's `target/`.
- REQ-5: If no usable lane build is given and the slot has no `target/` while the seed has one, then the audit-slot command shall warm the slot from the seed.
- REQ-6 (REQ-4): If a lane worktree given is not usable, then the audit-slot command shall refuse and name the reason.
- REQ-7: When an audit slot is finished with discard, the audit-slot command shall revert its tracked changes, remove its untracked files outside `target/`, keep its `target/` and report every reverted and removed path.
- REQ-8 (REQ-7): If an audit slot holds changes and discard is not given, then the audit-slot command shall refuse and name the files.
- REQ-9: The audit-slot command shall hold the repository's slot lock while it changes a slot.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-6, REQ-8, REQ-9): in a temporary repository, `prep audit-1 --head H` leaves `audit-1` detached at H with an empty tracked status. A slot name outside `audit-1..3`, a head that is not a commit, a dirty slot and an unusable lane are refused, each naming its reason and changing nothing.
  verify: `cargo nextest run --locked --offline --test audit_slot_cli`
- AC-2 (REQ-4, REQ-5): with a lane worktree at H that has `target/`, the slot's `target/` holds the lane's files with the same modification times and the answer says `lane`. Without a lane, a slot lacking `target/` is warmed from the seed and the answer says `seed`.
  verify: `cargo nextest run --locked --offline --test audit_slot_cli`
- AC-3 (REQ-4): on macOS, a temporary crate with a dependency outside the repository, built in the lane worktree, then prepared into an audit slot from it, builds in the slot without compiling the dependency; a one-line change then recompiles only the crate.
  verify: `cargo nextest run --locked --offline --test audit_slot_cli`
- AC-4 (REQ-7, REQ-8): after an audit changed a tracked file and left an untracked file, `done audit-1` refuses and names both; `done audit-1 --discard` reverts and removes them, keeps `target/`, reports both paths, and leaves an empty `status --porcelain`.
  verify: `cargo nextest run --locked --offline --test audit_slot_cli`

## Scope

- write:
  - `.moai/specs/SPEC-20261123-audit-slot/spec.md`
  - `src/audit_slot.rs` (new)
  - `src/lib.rs` (one `pub mod` line)
  - `src/ops/audit_slot.rs` (new)
  - `tests/audit_slot_cli.rs` (new)
  - `docs/design-review/appendix/operations/audit_slot.md` (new)
  - `src/worktree_slot.rs`: five helpers become `pub(crate)` (`common_dir`, `primary_of`, `same`, `linked`, `lock`), and the lock message names both commands
  - `src/help.rs`: the `audit-slot` row and its examples, added under a one-command line lease
- read-only: everything else

## Decisions

- D-1: `prep` creates a missing audit slot through `worktree_slot::ensure`, which takes and releases the slot lock itself. `prep` then takes the same lock (the same file in the git common directory) for the switch and the warm.
- D-2: The lane build replaces the slot's own `target/`. The copy goes to a temporary name inside the slot first. The old `target/` is renamed aside only after the copy succeeded, and is removed after the new one is in place. A failed copy leaves the old `target/` as it was.
- D-3: `done` changes only the slot worktree.
  - Tracked changes, staged or not, are reverted by `git reset --hard HEAD` inside the slot. A staged new file is removed with them.
  - The untracked files `git status --porcelain=v2 -z --untracked-files=all` lists outside `target/` are removed. Each path is checked to stay inside the slot.
  - Ignored files other than `target/` are left as they are.
- D-5: `prep` checks the lane before it adds or changes the slot, so a refused lane changes nothing. When a lane is given and the slot is new, the seed warm of `ensure` runs first and the lane build then replaces it. A clone costs little on macOS.
- D-6: measured on macOS by the AC-3 test. After `prep` from the lane, the dependency outside the repository is `Fresh`. The workspace crate is `Dirty` once, because the switch writes its source files after the lane's build. So the head gets one incremental build, which meets the card's criterion "one incremental build of the head and no full rebuild".
- D-7: a refusal message contains `refused`, so the envelope reports it as `invalid_input`. A git failure reads `git failed (<step>)` and is reported as `io`.
- D-4: The branch starts from the head of SPEC-20261123-worktree-slots, which is in the leader queue, and merges the integration tip once after that lands.

## Plan

1. This SPEC.
2. Tests first (RED): `tests/audit_slot_cli.rs` AC-1 to AC-4 on temporary repositories.
3. `src/audit_slot.rs`, `src/ops/audit_slot.rs` and the operation document.
4. Light checks, one affected run, READY with head and base.

## Risks

- A clone shares blocks until written. The free-disk floor of the slot command guards it when a slot is added; `prep` adds no worktree beyond that.
- A lane that moves its worktree to another head between `prep`'s check and its copy would hand over a mismatched build. `prep` reads the lane's HEAD again after the copy and refuses when it changed.

## Result

- Slice 1: `audit-slot prep` and `done`.
- RED first: `cargo nextest run --locked --offline --test audit_slot_cli` ran before the code. 5 tests ran and failed; the 4 others were skipped by fail-fast.
- After the code, with `--no-fail-fast`: 9 run, 9 passed.
- AC-3 was measured on macOS: the dependency is `Fresh`; the crate is `Dirty` once (D-6).
- The integration tip 072e359488 was merged once (fbd40fd1da). The four help goldens were then regenerated, adding the one audit-slot row (a3ee3bfc6f), under the recorded leader decision.
- After the merge, `build.rs` writes no absolute path. The ops list holds names only, and the bundle packs use `env!("CARGO_MANIFEST_DIR")`, which the compile resolves to the slot. So a copied build never points at the lane's files.
