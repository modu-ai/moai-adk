---
id: SPEC-20261122-home-tidy
title: "The home moai folder is inventoried, tidied on the user's pick and kept in order by thresholds, also as a step of update"
status: draft
revision: 1
created: 2026-11-22
---

# The home moai folder is inventoried, tidied on the user's pick and kept in order by thresholds, also as a step of update

## Intent

The user decided on 2026-11-22: the home moai folder (`~/.moai`) is analysed, tidied and kept in order; what is not needed and what is old goes into an archive first; `moai update` runs the same optimising step; and the folders the user works in (the project folders moai works in and the home moai folder) are looked after regularly: measured, and past a threshold a tidy plan is offered for the user's pick. No resident loop runs: the triggers are `moai update`, one short notice in the session start line only when a threshold is passed, and an on-demand command. Only what the user picks runs, every destructive step needs that pick, and entries of other products in the same folder are listed and never changed.

Measured on 2026-11-22 by the leader's read-only inventory (card note 5902, a partial read: the projects group of the worktree folder and 313 worktrees of another repository were not sized; `du` sizes overcount APFS clones, so a reclaim is measured with `df` before and after): about 416 GiB are build output that must be kept until its branch is done, about 136 GiB are candidates (build folders of merged or missing worktrees 78.5, merged clean worktrees older than a day 55.3, old Go releases 1.84, temporary files 0.5), 86.2 GiB are merged and clean but touched by git within 24 hours (a fresh branch looks merged), at least 117 GiB belong to another product, about 24.6 GiB are secrets (listed by name only) and about 6.75 GiB are evidence, reports, handoffs and a backup that an archive can take. 101 worktree registrations are broken (91 point at a repository that no longer exists), 107 of 143 build folders carry no ownership record and can only be matched by name, and a few small entries (`state`, `logs`, `dev-login.sh`, a thread-media folder) have a writer nobody knows.

The code already holds the parts this SPEC joins: the read-only preview of leftovers (`optimize`), the plan and apply of worktree cleanup (`worktree_gc`), the archive of the tool's own unchanged files with a confirm token and a revert (`update_cleanup`, shown in `moai update` as the `cleanup` block), the move of old report files (`reports_move`), the verified copy and manifest of the SPEC archive (`archive`) and the entry classes of the Go folder (`home_state`). None of them lists the whole folder, none has a threshold, and none restores from one list file; this SPEC adds that and reuses the rest.

This SPEC is slice 1 of the card: the SPEC only. The code and the tests follow in the slices of the Plan.

## Non-goals

- Moving or deleting an existing worktree because of where it lives. The worktree place SPEC (SPEC-20261122-worktree-place) keeps old worktrees where they are until their branch is finished; this SPEC offers one only as a finished item under the rules below.
- Anything outside the home moai folder and the projects moai works in: `~/.claude*`, the Orca and Hide folders, `claude-profiles/`, `gpt-appserver/`, `local-plugins/`, `mcp/`, `studio/`, keys and vaults, tracked files and the automatic memory are listed or skipped and never changed.
- A resident loop, a scheduler or a background process: the triggers are the three named in the Intent.
- Opening a key, a credential or a secret-shaped file for any reason, including its size by reading: only the directory entry's own size and times are used.
- Rewriting the Codex and OMP writers of `moai update` (cards C-290 and C-291): this SPEC adds one step to the update flow and leaves the rest of that file to them.
- Choosing the thresholds for the user: the proposed values of the hooks report are the defaults, and every one is a profile key the user may change.
- Running a real tidy on the real HOME inside a test: every destructive step is proven on a copy in a temporary home.

## Requirements

- REQ-1: When the inventory command runs, the inventory command shall list every entry of the home moai folder with its class (`owned_live`, `owned_stale`, `retired_tool`, `other_product` or `unknown`), size, newest change and the reason for the class · failure: none — an entry that cannot be read is listed as `unknown` with the reason and the run goes on.
- REQ-2: When the inventory meets an entry that is secret-shaped, a link or a special file, the inventory command shall list its name and its own directory data without opening, following or hashing it · failure: none — such an entry is `protected` and never a candidate.
- REQ-3: When the tidy plan runs, the tidy plan command shall name each candidate with an action (`archive`, `delete` or `keep`), the rule that made it a candidate and the reason it is kept when it is kept · failure: none — a fact that cannot be established keeps the item with the reason `unknown`.
- REQ-4: The tidy plan command shall offer `archive` for every entry that cannot be rebuilt and `delete` only for rebuildable build output, broken registrations and empty folders of the retired-tool class · failure: none — the plan has no other delete.
- REQ-5: When the user applies picked items of a plan, the tidy apply command shall run only the picked items under the confirm token of the plan it has just recomputed, with merged state, live use and recent activity of each item recomputed right before it runs · failure: none — an unknown pick, a token that is not the recomputed one or an item that is no longer a candidate refuses before any write, and an item whose proof failed in the recompute is skipped and named.
- REQ-6: When the tidy apply command archives an item, the tidy apply command shall write one compressed archive with a `list.json` of the path, size, sha256, time and mode of every file and remove the original only after the archive was read back and every hash matched · failure: none — a mismatch removes nothing and reports the item.
- REQ-7: When the user restores an archive, the tidy restore command shall put back the picked items from the `list.json` with each hash verified and no existing file overwritten · failure: none — an existing path is named and skipped and the rest of the pick is restored.
- REQ-8: When a threshold of the standing rules is passed, the tidy plan command shall mark that folder kind as over and offer its candidates oldest first, with the thresholds of the table in Decisions as profile keys with those defaults · failure: none — a folder under every threshold offers nothing.
- REQ-9: When a worktree is offered for removal, the tidy plan command shall offer it only when its branch is landed or squash-merged by a recomputed proof, it has no untracked file and no ignored file outside cargo output, it was quiet for the recent window, git did not touch it within 24 hours and its branch has no unpushed commit · failure: none — any other worktree is kept with the named reason.
- REQ-10: When `claude rm` or `git worktree remove` refuses a worktree, the tidy apply command shall show the refusal text and the values it printed, leave the worktree as it is and never pass `--discard-unpushed` or `--force-remove-worktree` · failure: none — the item is reported as refused and the run goes on.
- REQ-11: When `moai update` previews, the update command shall show a `tidy` block only when the plan has candidates and apply it only with `--confirm-tidy <token>` of that preview · failure: none — a token that is not the preview's refuses before any write, and no preview key is shown for an empty plan.
- REQ-12: When a threshold is passed, the session start line shall show one short notice read from one small summary file written by the last plan, and no scan runs at session start · failure: none — a missing or old summary file shows nothing.

## Acceptance

- AC-1 (REQ-1, REQ-2): in a crafted home folder with a live entry, a stale entry, a retired-tool entry, another product's entry, an unknown entry, a secret-named file, a link and a special file, the inventory gives each entry its class, size, newest change and reason, the secret-named file, the link and the special file are listed as protected without being opened (a file the test makes unreadable stays unread and the run succeeds), and a probe that records every open call of the run sees none on those three · verify: `cargo test --locked --offline --test home_tidy_cli inventory_`
- AC-2 (REQ-3, REQ-4, REQ-8): the plan names action, rule and reason for each candidate; a folder over its threshold is marked and offers oldest first, a folder under every threshold offers nothing, no entry that cannot be rebuilt has the action `delete`, and each default threshold is overridden by its profile key · verify: `cargo test --locked --offline --test home_tidy_cli plan_`
- AC-3 (REQ-5): the apply runs only the picked items with the recomputed token; an unknown pick or a stale token refuses with nothing written; an item that became a branch with a new commit between the plan and the apply is skipped and named; the answer carries `df` before and after · verify: `cargo test --locked --offline --test home_tidy_cli apply_`
- AC-4 (REQ-6, REQ-7): an archived item is read back and hash-checked before its original goes; an archive whose byte is flipped before the check removes nothing; the restore puts the item back by its list file with the same bytes and modes, and an existing path is named and not overwritten · verify: `cargo test --locked --offline --test home_tidy_cli archive_` and `restore_`
- AC-5 (REQ-9, REQ-10): a worktree with an unpushed commit, one with a file git ignores outside cargo output, one touched by git within 24 hours and one that is not landed are each kept with their reason, a landed clean quiet pushed one is offered, and a `claude rm` stand-in that refuses leaves the worktree, shows its text and is never called with the two force flags · verify: `cargo test --locked --offline --test home_tidy_cli worktree_`
- AC-6 (REQ-11): `moai update` on a copy of a home folder with candidates shows the `tidy` block and applies nothing without `--confirm-tidy`; with an empty plan the preview has no `tidy` key and the thirteen preview keys the existing test pins are unchanged · verify: `cargo test --locked --offline --test update_user_layer_cli` and `--test home_tidy_cli update_`
- AC-7 (REQ-12): the session start line shows the notice from a summary file that says a threshold is passed, shows nothing for a missing or old file, and runs no directory walk (a probe that counts reads sees only the one file) · verify: `cargo test --locked --offline --test home_tidy_cli notice_`
- AC-8 (REQ-1, REQ-3, REQ-6, REQ-11): the card operations document and the help text of the new command are present, and the card closes on the three criteria of the card: the inventory classification, the pick-only archive-before-delete with a restore, and the update preview proven on a copy · verify: `cargo test --locked --offline --test card_cli` and the spec check of this SPEC

## Scope

- write: `.moai/specs/SPEC-20261122-home-tidy/spec.md`
- write (slices 2 to 4): `src/tidy/` (inventory, rules, plan, apply, archive, restore, notice), `src/ops/tidy_*.rs`, `tests/home_tidy_cli.rs`, `docs/design-review/appendix/operations/tidy.md`
- last step of slice 4, by pair request: `src/update.rs` (the `tidy` block next to the `cleanup` block), `src/main.rs` and `src/help.rs` (the command and its examples), the session start line module

## Decisions

- D-1: The command is `moai tidy` with the actions `inventory`, `plan`, `apply` and `restore`, as the hooks report proposed; `inventory` and `plan` read only. The existing `cleanup-*` operations stay and keep their own confirm token: their candidates appear in the tidy plan as class `retired_tool` with `apply` delegated to them, so that one proof of a tool's own unchanged file exists and not two.
- D-2: Classes are decided from facts, not from names alone: `owned_live` has a live owner (an ownership record whose worktree or session exists, or a process inside), `owned_stale` has an ownership record whose owner is gone or whose branch is landed, `retired_tool` is a table entry of a retired generation (`home_state::SELECTIVE_TABLE` and the Go-era names: releases, db, run, cache), `other_product` is a table entry of another product, and everything else is `unknown` and never offered. Build folders with no ownership record are matched by name only and are `keep-recheck`, offered for delete only when the matched branch is proven landed at the moment of the apply.
- D-3: Merged means an ancestor of the integration branch at one moment, and it is recomputed right before every pick is applied; a squash-landed branch reads as not merged and stays kept unless the squash test of the worktree plan proves it; a branch with git activity within 24 hours stays `keep-recheck` because a fresh branch looks merged.
- D-4: Sizes come from one walk with an entry limit and are reported as `du` does; the reclaim is the difference of `df` before and after the apply, because APFS clones make the sum overcount. A walk over its limit keeps its group with the reason `over_limit`.
- D-5: Thresholds (the proposed values the user approved as the start): reports over 500 MiB or 2,000 files a project, archive what is older than 90 days; evidence over 2 GiB, archive what landed more than 30 days ago; a log file over 4 MiB is rotated and three compressed copies are kept; build output over 50 GiB or free disk under 40 GiB, only landed branches; worktrees over 8 a project or 30 GiB, landed, clean and quiet for 120 minutes; receipts over 200 MiB, trees that the integration branch does not reach. Each is a profile key with that default.
- D-6: The archive is one compressed tar with a `list.json` beside it in `<home>/.moai/archive/<UTC time>/`, mode 0700 for the folder and 0600 for the files; the compression is zstd when the crate is available offline and gzip when it is not, named in `list.json`. Writing, reading back and comparing hashes come before any removal (the verified copy of `archive.rs` is the model, and its manifest type is reused if it fits).
- D-7: A worktree is removed by the tool that owns it: a Claude Code background session's worktree by `claude rm <id>`, any other by `git worktree remove` without `--force`. `claude rm` refuses a worktree with unpushed commits and reports a commit and a worktree id for the force flags; the tidy never passes those flags (REQ-10), shows the printed values to the user, and a user who wants the discard runs the command himself. A branch with unpushed commits is kept from the plan on (REQ-9), so the refusal is a second line of defence.
- D-8: The session start notice reads one summary file, `<home>/.moai/tidy-summary.json`, written by the last inventory or plan with the folder kinds over their threshold and the time; the session start line never walks a folder. A summary older than 7 days shows nothing and says no more.
- D-9: The update step reuses the cleanup block's pattern: the preview carries `tidy` only with candidates, the apply takes `--confirm-tidy <token>`, the confirm token is the digest of the recomputed plan, and nothing outside the home moai folder and the project is touched.
- D-10: Entries whose writer is unknown (`state`, `logs`, `dev-login.sh`, the thread-media folder) are `unknown`, listed with size and newest change and never offered; a later card names their writers.

## Plan

1. This SPEC, read by the leader and by the audit before any code.
2. Slice 2, tests first (RED): `tests/home_tidy_cli.rs` `inventory_` and `plan_` on a crafted home folder; then `src/tidy/` inventory and rules; the document row.
3. Slice 3, tests first: `apply_`, `archive_` and `restore_`; the archive writer, the read-back check, the restore and the `df` figure.
4. Slice 4, tests first: `worktree_`, `update_` and `notice_`; the worktree rules and the `claude rm` stand-in, the `tidy` block of `moai update` by pair request, the session start notice.
5. Per slice: light checks, at most the mutant cap, one affected run, then the READY note with head and base only.

## Risks

- A hand-built crafted home folder is not the leader's real one: the numbers of the Intent are not test data, and a slice that finds a class the table misses adds the case to the crafted folder first.
- Proof of merge by ancestor only misses squash landings and hits fresh branches; the 24-hour rule and the squash test narrow it, and a doubt keeps the item.
- `claude rm` is the host's command and its output format may change; the stand-in of the test pins only the two facts the rule uses (it refuses an unpushed worktree and it prints values for the force flags), and a changed host shows as an unparsed refusal that keeps the worktree.
- A compression crate that is not vendored offline forces gzip; `list.json` names the algorithm so a later zstd archive and a gzip one restore the same way.
- `src/update.rs` is being edited by the Codex and OMP removal cards; the update step waits for their landing or a pair request, whichever the leader orders.
- 313 worktrees of another repository sit in the same folder and were not sized: they are `other_product` by their repository and are never offered.
