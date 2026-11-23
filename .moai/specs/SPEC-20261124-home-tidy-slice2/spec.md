---
id: SPEC-20261124-home-tidy-slice2
title: "The home moai folder inventory and the read-only tidy plan, with the memory index size line"
status: approved
revision: 2
created: 2026-11-24
approval:
  by: leader
  at: 2026-11-23
  basis: "리더 결정 노트 8007 (2026-11-23T23:36:33Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단, 사용자 직접 선택 아님, 노트 8006 에 답함): 판 2 승인(head cff13cb09c); src/help.rs 네 줄은 line-lease; 레인은 그 노트를 인용만 해요"
  digest: "sha256:73cd70adf1aea87b6c61706231ad3061c5074ccb93f7ed29c3370b54106f670c"
---

# The home moai folder inventory and the read-only tidy plan, with the memory index size line

## Intent

Slice 2 of the home tidy work: the two read-only operations `tidy-inventory` and `tidy-plan`, and the size line of the automatic memory index. Related: SPEC-20261122-home-tidy (slice 1, landed as the design; its Requirements, Decisions D-1 to D-10 and class names are read as written there and are not repeated here). Slice 3 (apply, archive, restore) and slice 4 (worktree rules, the `tidy` block of `moai update`, the session start notice) stay as that SPEC's Plan names them.

What this slice answers, and for whom: a user whose home moai folder grows (reports, logs, old build output, leftovers of retired tools) sees every entry with a class, a size, its newest change and the reason, and a plan that names what could be archived or deleted and why, with no write at all. The plan is the preview the later apply and the `moai update` step run on.

What this slice does not yet decide by itself: an item whose safety needs git proof (a worktree, a build folder of a branch, evidence "landed more than 30 days ago", a receipt tree the integration branch does not reach) is measured and shown with the folder kind's threshold state, and each of its items is kept with the reason `proof_pending`. The proof rules are slice 4 (parent REQ-9, D-3).

Also in this slice, from the leader's coverage note of 2026-11-22 (card note 6245): A27 the tidy text names `claude purge` (the host command `claude purge [options] [path]`, read from `claude purge --help` on 2026-11-24; it replaces the older name `claude project purge`), and A50 the memory index targets. A26, the project folder tidy proposal (untracked leftovers, empty folders, old tool traces, overlapping docs, a project `.moai/archive`), is a different tree (the project, not the home folder) and needs its own SPEC after slice 3; it is recorded in Decisions D-4 and is not a requirement here.

## Interview

- none: no question was put to the user for this slice; its choices D-1 to D-5 follow the parent SPEC and the leader's card notes 6060 and 6245, and each is listed in Decisions.

## Non-goals

- Any write: both operations write nothing, not even the summary file for the session start notice (slice 4, parent REQ-12) and not an archive (slice 3).
- Worktree and build-folder proof of merge, broken registrations, `claude rm` and `git worktree remove` (slice 4).
- Wiring the thresholds to profile keys: `src/profile.rs` is a shared table; this slice takes the thresholds as arguments with the parent D-5 defaults, and the profile wiring comes by pair request with the update step.
- Editing, moving or deleting any `MEMORY.md`, and running `claude purge`: the operations only measure the file and print text for the user to run in a Claude session.
- The project folder tidy proposal (D-4).
- Opening a key, a credential or a secret-shaped file; touching the real HOME in a test (every test runs in a temporary home).

## Requirements

- REQ-1: When `tidy-inventory` runs on a home folder, the inventory operation shall list every top-level entry of its `.moai` folder with class (`owned_live`, `owned_stale`, `retired_tool`, `other_product`, `unknown` or `protected`), size, newest change and the reason for the class · failure: none — an entry that cannot be read is listed as `unknown` with the reason, a size walk that passes its entry limit or its time limit keeps its group with the reason `over_limit` and the size seen so far, and the run goes on.
- REQ-2: When the inventory meets an entry that is secret-shaped, a link or a special file, the inventory operation shall list its name and its own directory data (type, size, times) only, as class `protected`, without opening, following or hashing it · failure: none — the entry is never a candidate and the run goes on.
- REQ-3: When the inventory lists an entry of a folder kind that grows (`reports`, `logs`, `evidence`, `worktrees`, `build`, `receipts`), the inventory operation shall add one row per kind with its count, total size, oldest and newest change and whether each threshold of the kind is passed · failure: none — a kind whose folder is absent has no row.
- REQ-4: When the caller gives `--threshold <kind>.<field>=<value>`, the inventory and plan operations shall use that value instead of the default of the parent D-5 table for that field · failure: none — an argument that is not valid is the case of REQ-12.
- REQ-5: When `tidy-plan` runs, the plan operation shall name each candidate with an action (`archive`, `delete` or `keep`, where `delete` goes only to an empty folder of the retired-tool class), the rule that made it a candidate and, for `keep`, the reason it is kept · failure: none — a fact that cannot be established keeps the item with the reason `unknown`, and build output and broken registrations wait for slice 4.
- REQ-6: When a folder kind is over a threshold, the plan operation shall offer the items that the age rule of that kind reaches, oldest first · failure: none — an item the rule does not reach is kept with the reason `not_old_enough`, and a kind under every threshold offers nothing and keeps its items with the reason `under_threshold`.
- REQ-7: When an item of a kind whose safety needs git proof is planned, the plan operation shall keep it with the reason `proof_pending` · failure: none — nothing of such a kind is ever offered by this slice.
- REQ-8: When the plan is built, the plan operation shall give the digest of its candidate list, so that slice 3 can use the same digest as the confirm token of the recomputed plan · failure: none — two runs on an unchanged folder give the same digest.
- REQ-9: When `--memory-index <path>` is given, the inventory and plan operations shall report the lines, the bytes and the share of the host load cap (200 lines or 25 KB, whichever share is higher) of that one file read with a size limit, whether it is over 80 percent, and the targets of under 100 lines and 12 KB and 120 bytes per line · failure: none — a missing, unreadable or over-limit file is reported with the reason and the run goes on.
- REQ-10: When a memory index is over 80 percent of the cap, the plan operation shall print the three groups of the memory tidy (rules to move into instructions, retired items, merges) as text for the user to run in a Claude session, and `claude purge --all` as a reference line marked never run by moai · failure: none — under 80 percent the plan prints no such text.
- REQ-11: When either operation runs, the operation shall change no file and start no command that writes (no `git worktree` change, no `claude` command) · failure: none — a probe that records every write and every started command of a run sees none that writes.
- REQ-12 (REQ-4): IF a `--threshold` argument names an unknown kind or an unknown field or carries a value that is not a number, THEN the operation shall refuse before any read, naming the argument.

## Acceptance

- AC-1 (REQ-1, REQ-2): in a crafted home folder with a live entry, a stale entry, a retired-tool entry, another product's entry, an unknown entry, a secret-named file, a link and a special file, the inventory gives each entry its class, size, newest change and reason; the secret-named file, the link and the special file are listed as `protected` without being opened (a file the test makes unreadable stays unread and the run succeeds, and a probe that records every open call sees none on those three); a walk over a lowered limit keeps its group with `over_limit` · verify: `cargo test --locked --offline --test home_tidy_cli inventory_`
- AC-2 (REQ-3, REQ-4, REQ-12): the kind rows carry count, size, oldest and newest and the passed thresholds; each default of the parent D-5 table is overridden by `--threshold`, and a bad argument (unknown kind, unknown field, a value that is not a number) refuses before any read · verify: `cargo test --locked --offline --test home_tidy_cli kinds_` and `--test home_tidy_cli threshold_`
- AC-3 (REQ-5, REQ-6, REQ-7, REQ-8): the plan names action, rule and reason for each candidate; a kind over its threshold offers oldest first, a kind under every threshold offers nothing, the only `delete` is an empty folder of the retired-tool class, a worktree, a build folder and an evidence folder are `keep` with `proof_pending`, and two runs on one folder give one digest while a changed folder gives another · verify: `cargo test --locked --offline --test home_tidy_cli plan_`
- AC-4 (REQ-9, REQ-10): a crafted `MEMORY.md` of 150 lines, one of 170 lines and one of 30 KB give lines, bytes, percent and the over-80 flag as measured, a file under 80 percent gives no text, one over gives the three groups and the `claude purge --all` reference line, a missing file is reported with a reason, and the file is byte-identical after the run · verify: `cargo test --locked --offline --test home_tidy_cli memory_`
- AC-5 (REQ-11): a run in a temporary home leaves every file of that home with the same bytes and times, and a `claude` and a `git` stand-in on the path record no call that writes · verify: `cargo test --locked --offline --test home_tidy_cli readonly_`
- AC-6 (REQ-1, REQ-5): the help text of the two operations says they write nothing, the help table lists both once with an example each, the operations document fragments match the code, and the SPEC checks · verify: `cargo test --locked --offline --test home_tidy_cli help_`, `cargo test --locked --offline --bin moai-cli help_rules` and `operations_doc`, and `moai spec check SPEC-20261124-home-tidy-slice2 --project .`

## Scope

- write: `.moai/specs/SPEC-20261124-home-tidy-slice2/spec.md`
- write (after the approval): `src/tidy/` (`mod.rs`, `inventory.rs`, `rules.rs`, `plan.rs`, `memory.rs`), `src/ops/tidy_inventory.rs`, `src/ops/tidy_plan.rs`, `tests/home_tidy_cli.rs`, `docs/design-review/appendix/operations/tidy_inventory.md`, `docs/design-review/appendix/operations/tidy_plan.md`
- shared lines, by line lease: `src/lib.rs` (one line, `pub mod tidy;`) and `src/help.rs` (the `GROUPS` row and the `EXAMPLES` row of each of the two operations, four lines); the operations register themselves from their files (`src/ops/mod.rs`), so `src/main.rs` stays untouched
- read-only: `src/home_state.rs` (`is_secret`, `table_class`), `src/hygiene.rs`, `src/optimize.rs`, `src/worktree_report.rs` (`bytes_under`), the parent SPEC
- forbidden: `src/update.rs`, `src/profile.rs`, `src/hook_session_start.rs`, any file outside this list

## Plan

1. This SPEC, read by the leader before any code (the approval is the leader's note, never recorded here by the author).
2. RED: `tests/home_tidy_cli.rs` with `inventory_`, `kinds_`, `threshold_`, `plan_`, `memory_` and `readonly_` on a crafted temporary home; each fails before the code exists.
3. `src/tidy/rules.rs` (class table built on `home_state::table_class` and `is_secret`, the kind table, the thresholds), `inventory.rs` (one bounded walk), `memory.rs`, `plan.rs`, then the two operation files and their document fragments.
4. Guard mutants (one per new guard test), one `moai check affected` run at READY, and the READY note with head and base.

## Risks

- A class decided from names alone would be wrong for a new entry: an entry in no table row is `unknown` and never offered, and a case found later is added to the crafted folder first (parent Risks).
- Sizes from one walk overcount APFS clones; this slice reports the walk's size and leaves the `df` figure to slice 3 (parent D-4).
- `claude purge` may change its name or options; the plan prints the reference line from a constant and says it was read on 2026-11-24, and nothing runs it.
- The 80 percent notice and the 100 line target come from the host's documented load cap (first 200 lines or 25 KB) and the leader's targets; if the host changes the cap, the constants change in one place.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Slice 2 is read-only: no summary file, no archive, no delete, and the operations are `tidy-inventory` and `tidy-plan` (hyphen form, one file each), not the `moai tidy <action>` of parent D-1, so that `src/main.rs` is not touched; the help table `src/help.rs` takes one `GROUPS` row and one `EXAMPLES` row per operation by line lease (revision 2: revision 1 said the table stays untouched, which was wrong) | Parent Scope lists `src/ops/tidy_*.rs`; the operations registry (`src/ops/mod.rs`) takes one file per operation; the binary's help tests `help_table_names_every_command_once`, `every_command_has_examples` and `every_one_line_about_follows_the_rule` refuse an operation the table lacks (found by the READY check run, `home-tidy-slice2` log `05-test.log`); the one-command front is added with slice 4 by pair request | Add a `tidy` front command that calls the same functions |
| D-2 | Kinds that need git proof are measured and kept with `proof_pending` in this slice | Parent D-3 and REQ-9: merged state is recomputed per item; no proof, no offer | Slice 4 replaces the reason with the recomputed proof |
| D-3 | Thresholds are arguments with the parent D-5 defaults, not profile keys yet | `src/profile.rs` is a shared key table; the key wiring needs the update step's pair request | Slice 4 reads the profile keys and passes the same values |
| D-4 | A26 (the project folder tidy proposal and a project `.moai/archive`) gets its own SPEC after slice 3 | Different tree and different proof rules (tracked files, untracked leftovers); keeping it out holds this slice to the home folder | Fold it into a later slice of this SPEC if the leader prefers |
| D-5 | The memory index size line reads one file named by `--memory-index`; nothing under `~/.claude*` is walked | Parent Non-goals: the automatic memory is never changed; an explicit path keeps the read to one file | Add a discovery option in a later slice by decision |
