---
id: SPEC-20261123-audit-slot
title: "An audit takes a fixed audit slot, switches it to the head and inherits the lane's build, and leaves it clean"
status: draft
revision: 4
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
- REQ-10: If the slot path or a folder above it inside the repository is a symbolic link, then the audit-slot command shall refuse every prep and done before its first write.
- REQ-11 (REQ-2): If the folder at a slot path is not that slot's own worktree of the repository, then the audit-slot command shall refuse every prep and done before its first write.
- REQ-12 (REQ-2): If another live holder has a slot, then the audit-slot command shall refuse to prepare or finish that slot for a different holder.
- REQ-13: When a build is placed in a slot, the audit-slot command shall give every tracked file of the slot the current modification time. · failure: none — a file that cannot be touched fails the prep and puts the earlier holder record back.
- REQ-14 (REQ-1): If the head is not a full 40-character lowercase commit id, then the audit-slot command shall refuse before it reads the repository.
- REQ-15 (REQ-7): If a slot still holds changes after a discard, then the audit-slot command shall fail and keep the slot held.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-6, REQ-8, REQ-9): in a temporary repository, `prep audit-1 --head H` leaves `audit-1` detached at H with an empty tracked status. A slot name outside `audit-1..3`, a head that is not a commit, a dirty slot and an unusable lane are refused, each naming its reason and changing nothing.
  verify: `cargo nextest run --locked --offline --test audit_slot_cli`
- AC-2 (REQ-4, REQ-5): with a lane worktree at H that has `target/`, the slot's `target/` holds the lane's files with the same modification times and the answer says `lane`. Without a lane, a slot lacking `target/` is warmed from the seed and the answer says `seed`.
  verify: `cargo nextest run --locked --offline --test audit_slot_cli`
- AC-3 (REQ-4): on macOS, a temporary crate with a dependency outside the repository, built in the lane worktree, then prepared into an audit slot from it, builds in the slot without compiling the dependency; a one-line change then recompiles only the crate.
  verify: `cargo nextest run --locked --offline --test audit_slot_cli`
- AC-4 (REQ-7, REQ-8): after an audit changed a tracked file and left an untracked file, `done audit-1` refuses and names both; `done audit-1 --discard` reverts and removes them, keeps `target/`, reports both paths, and leaves an empty `status --porcelain`.
  verify: `cargo nextest run --locked --offline --test audit_slot_cli`

- AC-5 (REQ-10): a slot path that is a link to another registered worktree is refused by prep without a lane, by prep with a lane, by done, and by done --discard. Each refusal names the link. The linked worktree keeps every byte, its tracked change, its untracked file and its HEAD. A `.claude/worktrees` that is a link is refused and nothing is written through it.
  verify: `cargo nextest run --locked --offline --test audit_slot_cli`
- AC-6 (REQ-11): at a registered slot path, a plain folder, another repository and a `.git` that borrows the lane's git dir are each refused by prep, done and done --discard. The primary checkout keeps its branch, its HEAD and its tracked change; the other repository keeps every byte; the lane keeps its staged content.
  verify: `cargo nextest run --locked --offline --test audit_slot_cli`
- AC-7 (REQ-12, REQ-9): a second holder is refused by prep and done of a held slot; the same holder may prep again; `prep auto` takes the first slot without a live holder and refuses when all three are held; `done` releases; a holder whose pid ended or whose lease ended is taken over with the record in the answer. With the slot lock held elsewhere and a one-second wait, prep and done refuse and change nothing.
  verify: `cargo nextest run --locked --offline --test audit_slot_cli`
- AC-8 (REQ-13): a lane builds a mutant and reverts its source, so its HEAD is the head with no tracked change. After prep from that lane, the slot's binary prints the head's output, not the mutant's.
  verify: `cargo nextest run --locked --offline --test audit_slot_cli`
- AC-9 (REQ-14, REQ-15, REQ-3): a short id, a branch name, `HEAD` and upper-case hex are refused as `--head`. A copy that a killed prep left (`.target-inherit-<pid>`) is named once as a folder in the refusal and removed whole by a discard. A cap refusal or a disk-floor refusal is `invalid_input` and adds nothing.
  verify: `cargo nextest run --locked --offline --test audit_slot_cli`

## Scope

- write:
  - `.moai/specs/SPEC-20261123-audit-slot/spec.md`
  - `src/audit_slot.rs` (new)
  - `src/lib.rs` (one `pub mod` line)
  - `src/ops/audit_slot.rs` (new)
  - `tests/audit_slot_cli.rs` (new)
  - `docs/design-review/appendix/operations/audit_slot.md` (new)
  - `src/worktree_slot.rs`: five helpers become `pub(crate)` (`common_dir`, `primary_of`, `same`, `linked`, `lock`), and the lock message names both commands; the repair adds `own_worktree`, the lock wait setting, the temporary-folder prefixes, and the link and own-worktree refusals in `ensure`
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

- D-9 (revision 4, AUDIT FAIL B1): a slot is its own worktree when git run in it names the slot as its top level, the repository's git common dir as its common dir, and a private git dir whose `gitdir` file points back at the slot's `.git`. `prep` checks it before `worktree-slot ensure` (for a slot that exists) and again under the lock; `done` checks it under the lock. `ensure` makes the same check for a present slot before it warms one, so the seed is never copied into a foreign folder.
- D-10 (revision 4, AUDIT FAIL B2, leader direction): a lane at the head with no tracked change does not prove that its `target/` was built from the head; a mutant run leaves a build that Cargo calls `Fresh`. After any build is placed (lane, seed or kept), every file `git ls-files` lists in the slot gets the current modification time. Cargo then compiles the workspace crate once from the audited sources. Dependencies outside the slot keep their times and stay `Fresh`, as AC-3 measures. The answer gives the count as `build.sources_touched`.
- D-11 (revision 4, AUDIT FAIL B3): a slot has one holder. `prep --holder <name>` writes a record (name, the host session when the host names one, an optional `--pid`, the head, the time and `lease_until`) to `<git common dir>/moai-audit-slots/<slot>.json` under the slot lock, through a temporary file and a rename. A holder is live while its lease runs (`--lease-secs`, default three hours, at most a day) and its pid, when given, has not ended. Another holder is refused while the record is live; the same name (and the same session when both name one) may prep again; a dead holder is taken over and the old record is returned as `took_over`. `done --holder` refuses a live different holder and removes the record when the slot ends clean (`released`). `prep auto` tries `audit-1` to `audit-3` in order and refuses with every holder named when all are held. A failed switch or build after the record was written puts the earlier record back. Reversible: the lease length and the record place are defaults of this command.
- D-12 (revision 4, findings N1, N2, N4, N6, N7, N8): the link check covers `.claude` and `.claude/worktrees`, not only the slot. `ensure` refuses a linked slot under its own lock with the same words as the slot cards' later slice. After `ensure` ran, a refusal says what it did instead of "nothing was changed". `--head` must be 40 lowercase hex. A copy a killed prep left is one entry, its folder. A discard that leaves changes fails (non-zero) and keeps the holder. A cap or disk-floor refusal from `ensure` is reworded as an audit-slot refusal, so it is `invalid_input` as D-7 says. The lock wait can be shortened with `MOAI_WORKTREE_SLOT_LOCK_WAIT_SECS`, so a test proves the lock is taken.
- D-13 (revision 4, finding N5, by design D-3): ignored files outside `target/` survive a discard, including what the user's global ignore file matches, so the next audit sees them. The operation document says so. Finding N11 (no free-disk check before the clone of the lane's `target/`) stays as accepted in Risks.

- D-8: The link check runs first in prep (before the lane check and before `worktree-slot ensure`, which on this branch does not refuse a linked slot yet) and first in done. Both commands check again after taking the slot lock. Leader decision on the card, made before the audit was dispatched.

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
- Repair before the audit: a linked slot path is refused by every writing action (REQ-10, AC-5). Test: `every_writing_action_refuses_a_slot_path_that_is_a_link`.
- Repair of the audit FAIL (revision 4): `cargo nextest run --locked --offline --no-fail-fast --test audit_slot_cli` ran 18, 18 passed.
- AC-1 met: `prep_refuses_a_name_that_is_not_an_audit_slot`, `prep_refuses_a_slot_with_changes_and_names_them`, `prep_refuses_a_lane_it_cannot_inherit_from_and_changes_nothing`.
- AC-2 met: `prep_adds_the_slot_at_the_head_and_inherits_the_lane_build`, `prep_again_replaces_the_slot_target_with_the_lane_build`, `prep_without_a_lane_keeps_the_slot_build_or_warms_from_the_seed`.
- AC-3 met on macOS: `a_slot_prepared_from_the_lane_does_not_compile_the_dependency`.
- AC-4 met: `done_refuses_changes_without_discard_then_discards_them_and_keeps_target`, `done_refuses_a_slot_that_is_not_there`.
- AC-5 met: `every_writing_action_refuses_a_slot_path_that_is_a_link`, `a_link_above_the_slot_is_refused`.
- AC-6 met: `a_registered_slot_that_is_not_its_own_worktree_is_refused_and_nothing_changes`.
- AC-7 met: `a_slot_has_one_live_holder_and_done_releases_it`, `a_dead_holder_is_taken_over_with_a_line`, `prep_and_done_wait_for_the_slot_lock_then_refuse`.
- AC-8 met: `an_inherited_mutant_build_is_compiled_again_from_the_head`.
- AC-9 met: `prep_refuses_a_name_that_is_not_an_audit_slot`, `a_copy_a_killed_prep_left_is_named_once_and_removed_whole`, `a_cap_refusal_is_invalid_input_and_adds_nothing`.
- Not tested: a discard that leaves changes behind (REQ-15). No fixture makes `reset --hard` and the removals leave a change.
