---
id: SPEC-20261119-bundle-build
title: "moai bundle build: one new branch and worktree, one merge commit per head, a conflicting head dropped and named"
status: draft
revision: 2
created: 2026-11-19
parent: SPEC-20261119-bundle-tool
---

# moai bundle build: one new branch and worktree, one merge commit per head, a conflicting head dropped and named

## Intent

Card C-119, slice 3 of `SPEC-20261119-bundle-tool` (slice 2, `bundle ready|check|body`, is READY at head `b49fd5f070` after a tier 1 audit PASS). The leader accepted the slice table and directed this slice on 2026-11-19: push only, no pull request.

The hand procedure the leader runs today for one M-Flow bundle: pick the audited READY heads, create `m-flow/bundle-<date>-<n>` from the tip in a new worktree, merge each head with one merge commit, stop at a conflict, rebuild without the head that conflicts, and write down the pinned SHA. Bundle 5 shows what the manual rebuild costs: every dropped head meant working out the new branch name and the head list again. This slice makes the rebuild one command:

- `moai bundle build C-<a> C-<b> …` creates the branch and worktree and merges the heads in the order given.
- A head whose merge conflicts is dropped, named with its files, and the answer carries the command line that builds again without it (a new branch name: the old branch is never moved, D-3 of the parent).
- The answer says which heads merged, which were dropped, which were skipped and the pinned SHA, so the next step (`check`, `body`, push, pull request) stays the leader's.

The tool writes a branch, a worktree and merge commits and nothing else: no push, no fetch, no pull request, no card note, no ledger row, no setting (REQ-12 of the parent, REQ-10 here).

## Non-goals

- Choosing the heads or their order: the caller names them (parent D-2).
- Running the checks on the built branch: that is slice 4 (`bundle verify`).
- Deleting or moving a branch or a worktree, including the ones this command made: cleanup stays a preview plus the user's confirmation.
- Pushing the branch or opening the pull request.
- Judging a head (audit evidence, trailers, SPEC): `bundle check` does it; `build` only needs a head it can merge.

## Requirements

- REQ-1: When `moai bundle build` is given heads in an order, `moai bundle build` shall create a new branch `m-flow/bundle-<yyyymmdd>-<n>` and a new worktree `<HOME>/.moai/worktrees/<project>-bundle-<n>` at the tip's commit, merge each head into the branch with one merge commit in the given order, and print the pinned SHA that `git rev-parse` gives for the branch.
- REQ-2 (REQ-1): IF a head cannot be used (a card with no readable `LANE-READY-INTEGRATION` line, a head that is not a commit here, a card named twice, a `--date` that is not a date, or no head left to merge after the heads already in the tip), THEN `moai bundle build` shall refuse before it creates anything and name every such head in one answer.
- REQ-3 (REQ-1): IF a head's merge conflicts under the text merge GitHub runs, THEN `moai bundle build` shall abort that merge, leave the branch without that head, name the head and its conflicting files, and go on with the next head.
- REQ-4 (REQ-1): IF a head is already in the tip or already in the branch when its turn comes, THEN `moai bundle build` shall merge nothing for it and name it as skipped with the reason.
- REQ-5: `moai bundle build` shall take `<n>` as the lowest integer from 1 to 999 for which no local branch and no `origin` branch `m-flow/bundle-<yyyymmdd>-<n>`, no registered worktree (also one whose folder is gone) and no folder `<project>-bundle-<n>` under `<HOME>/.moai/worktrees/` exists.
- REQ-6: `moai bundle build` shall never move, reset, force or delete an existing branch, ref, worktree or folder.
- REQ-7 (REQ-6): IF the branch or the folder is taken when `moai bundle build` creates it, THEN `moai bundle build` shall stop and change nothing.
- REQ-8 (REQ-1): IF git fails after the branch exists (no author identity, a hook refuses a merge commit, an abort does not restore the branch), THEN `moai bundle build` shall stop, name the branch and the worktree it leaves behind, and delete nothing.
- REQ-9: `moai bundle build` shall merge with the text merge GitHub runs, so that no merge driver named in the git configuration of the clone or of the user decides.
- REQ-10: `moai bundle build` shall write only the new branch, its worktree and the merge commits, and no push, fetch, pull request, card note, ledger row or setting.
- REQ-11: `moai bundle build` shall print `action`, `branch`, `worktree`, `tip`, `pinned`, `merged`, `dropped`, `skipped`, `rebuild` and `ok`, with `ok` false exactly when a head was dropped and `rebuild` the command line of the same build without the dropped heads, its words quoted for a shell, or an empty string when nothing was dropped or no head would be left.
- REQ-12 (REQ-5): IF all the numbers from 1 to 999 are used for the date, THEN `moai bundle build` shall refuse before it creates anything.

## Acceptance

- AC-1 (REQ-1, REQ-5, REQ-11): two heads give one new branch and one new worktree at `<HOME>/.moai/worktrees/<project>-bundle-1`, two merge commits in the given order, `pinned` equal to `git rev-parse` of the branch, `ok` true and `rebuild` empty.
  verify: `cargo test --locked --offline --test bundle_build_cli two_heads_make_one_branch_and_worktree_with_a_merge_commit_each`
- AC-2 (REQ-2): a card with no READY line, a head SHA that is not here, a card named twice, and heads that are all already in the tip are each refused with every such head named, and no branch, worktree or folder exists afterwards; a `--date` that is not a date (`20261340`, `20260132`, `2026111a`) is refused the same way; a head that names a tag is merged as the commit it points to.
  verify: `cargo test --locked --offline --test bundle_build_cli heads_that_cannot_be_used_are_refused_before_anything_is_created`; `cargo test --locked --offline --test bundle_build_cli a_date_that_is_not_a_date_is_refused_and_a_tag_head_is_the_commit`
- AC-3 (REQ-3, REQ-9, REQ-11): of three heads where the second conflicts with the first in a file a merge driver registered in the clone would hide, the branch holds the first and third merge commits, the answer names the second head and the file, `ok` is false and `rebuild` is the build command of the first and third head, and the working tree of the worktree is clean.
  verify: `cargo test --locked --offline --test bundle_build_cli a_conflicting_head_is_dropped_and_named_and_the_rest_merge`; `cargo test --locked --offline --test bundle_build_cli a_second_registered_merge_driver_does_not_decide`
- AC-4 (REQ-4): a head already in the tip and a head already contained in an earlier head's merge are skipped with their reasons and make no merge commit.
  verify: `cargo test --locked --offline --test bundle_build_cli a_head_in_the_tip_or_in_the_branch_is_skipped`
- AC-5 (REQ-5, REQ-6, REQ-12): with a local branch for `-1`, an `origin` branch for `-2`, a worktree folder for `-3` and a registered worktree for `-4`, the build uses `-5`; every pre-existing ref, folder and worktree is byte-equal afterwards; a second build is `-6`; a registered worktree whose folder is gone still takes its number; with the numbers 1 to 999 all used for the date the build is refused and nothing is created.
  verify: `cargo test --locked --offline --test bundle_build_cli the_number_is_the_lowest_unused_and_existing_names_are_never_touched`; `cargo test --locked --offline --test bundle_build_cli a_registered_worktree_whose_folder_is_gone_still_takes_its_number`; `cargo test --locked --offline --test bundle_build_cli all_999_numbers_used_is_refused_before_anything_is_created`
- AC-6 (REQ-7): the creation step on a branch name or folder that already exists returns an error and changes no ref and no folder, also for an existing branch that is checked out nowhere (so `-b`, not `-B`, is what refuses).
  verify: `cargo test --locked --offline --test bundle_build_cli creation_refuses_a_taken_branch_or_folder`
- AC-7 (REQ-8): in a clone whose git has no author identity (`user.useConfigOnly`), the build stops at the first merge, the error names the branch and the worktree, both still exist and the branch is at the tip's commit.
  verify: `cargo test --locked --offline --test bundle_build_cli a_git_failure_after_creation_names_what_is_left_behind`
- AC-8 (REQ-10): after a build the primary clone's HEAD, status and every ref except the one new branch, the bare `origin`'s refs and the ledger bytes are as before.
  verify: `cargo test --locked --offline --test bundle_build_cli build_changes_only_the_new_branch_and_worktree`
- AC-9 (REQ-10, REQ-11): `bundle build --help` starts with `Writes a new branch and worktree.`, the operations page has the `build` action with its answer keys, and the neighbours stay green.
  verify: `cargo test --locked --offline --test bundle_build_cli the_help_and_the_ops_doc_say_what_build_writes`; `cargo test --locked --offline --test git_env_scrub`; `cargo test --locked --offline --test home_isolation_scan`; `cargo test --locked --offline --test help_layout`; `cargo test --locked --offline --test product_cli`

- AC-10 (REQ-11): the `rebuild` line quotes a tip name that has shell characters (`odd;tip` is written `'odd;tip'`) and is an empty string when the only head conflicts and none would be left.
  verify: `cargo test --locked --offline --test bundle_build_cli the_rebuild_line_is_quoted_and_empty_when_no_head_is_left`

## Scope

- write: `src/bundle.rs` (the line `pub mod build;`, the merge-driver list, and what `build` takes from the module), `src/bundle/build.rs`, `src/ops/bundle.rs`, `tests/bundle_build_cli.rs`, `docs/design-review/appendix/operations/bundle.md`, `.moai/specs/SPEC-20261119-bundle-build/**`, `.moai/specs/SPEC-20261119-bundle-tool/**`.
- read-only: `src/home_workdir.rs`, `src/worktree.rs`, `src/identity.rs`, `src/audit_delta.rs`, `src/hygiene.rs`, `src/gate.rs`, `src/card/command.rs`, `src/ledger/command.rs`.
- forbidden: `src/lib.rs` and `src/main.rs` (no new module line is needed, D-1), any other change under `src/`, the real HOME and ledger as a write target, `skills/`, `.github/`, any push of a ref other than this branch, any pull request.

## Plan

**Where it lives (D-1).** `src/bundle/build.rs` is a child module of `src/bundle.rs` (`pub mod build;`), so `src/lib.rs` is untouched. It reuses the head arguments and the tip of slice 2 (`parse_head_args`, `tip_of`, the newest readable READY line) through `pub(super)` items; the check of a head (audit, trailers, SPEC) stays `bundle check`.

**Steps of `build`.**
1. Resolve the tip to a commit (`tip_of`, never fetched) and every head: `C-<n>` through the newest readable READY line, `C-<n>=<sha>` as given. Collect every problem (REQ-2); refuse if any, before a branch or folder exists. Heads already contained in the tip are set aside as skipped; if none is left, refuse.
2. Number (REQ-5): `n` is 1, 2, … until none of these exists: the local branch `refs/heads/m-flow/bundle-<date>-<n>`, the remote branch `refs/remotes/origin/m-flow/bundle-<date>-<n>`, a path of `git worktree list --porcelain` that equals the folder, the folder `<HOME>/.moai/worktrees/<project>-bundle-<n>` (`home_workdir::worktree_dir`, which makes the name safe and refuses a reserved one). `<date>` is today (UTC, `hygiene::utc_date`) unless `--date YYYYMMDD` is given (tests; D-13). `n` runs from 1 to 999; a registered worktree whose folder is gone still takes its number (`git worktree add` would refuse it).
3. Create: `git worktree add -b <branch> <folder> <tip sha>`. `-b`, never `-B`, so an existing branch is an error (REQ-7); the parent folder `<HOME>/.moai/worktrees/` is created if missing; a folder that already exists is refused by the tool before git runs.
4. For each head in order: skip it if it is an ancestor of the worktree's HEAD (REQ-4, reason `in_tip` or `in_branch`); otherwise `git -c merge.<name>.driver=<the text merge of slice 2> … merge --no-ff -m "Merge <card> head <short sha> into <branch>" <head sha>` in the worktree, with one `-c` for the manifest driver and one for every merge driver the configuration names (D-12). Exit 0: record the merge commit. Exit 1 with unmerged paths: record the files (`git diff --name-only --diff-filter=U`), `git merge --abort`, check that HEAD is the commit before the merge (REQ-3, REQ-8). Any other exit: stop (REQ-8).
5. Answer: `{action:"build", branch, worktree, tip:{name,sha,committed,note}, pinned, merged:[{card,head,merge}], dropped:[{card,head,files}], skipped:[{card,head,reason}], rebuild, ok, writes:true}`. `pinned` is `git rev-parse <branch>` read after the last merge. `rebuild` is `moai bundle build <card>=<head> … --tip <tip>` of the heads that were not dropped, in the given order, every word quoted when it holds a shell character (D-11), or an empty string when nothing was dropped or no head is left.

**Git.** Reads go through `crate::audit_delta::command` (a scrubbed environment, global and system configuration off). The writes of steps 3 and 4 need the user's own author identity and hook settings, so they run through `std::process::Command` with `identity::scrub_git_env` (the pattern of `worktree.rs`), in the lib crate (D-4). No `--no-verify`: a hook that refuses a merge commit is REQ-8.

**Tests** (`tests/bundle_build_cli.rs`, a temp HOME, a bare `origin` and a work clone as in slice 2; the clone's `merge.moai-manifest.driver` is `true`): the nine AC lines above. The creation step is `pub fn create` of the module, so the test of AC-6 calls it directly in a temporary repository.

## Risks

- A merge commit written under the user's own git configuration may be signed, or refused by a local hook: the build then stops with the branch and worktree named (REQ-8) and the leader decides; the tool never bypasses a hook (D-8).
- The conflict check of `bundle check` looks at pairs of heads; `build` merges in sequence, so an interaction of three heads shows up here as a conflict on the third and the third is dropped: the leader sees it in `dropped` and `rebuild` is one command.
- Two builds in the same second race for the same `<n>`: the second `git worktree add -b` fails on the existing branch and changes nothing (REQ-7).
- The worktree lives under `<HOME>/.moai/worktrees/`: the free-space floor of `moai worktree` is not applied here; cleanup of a built worktree is a separate, confirmed step.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `build` is a child module `src/bundle/build.rs` of `src/bundle.rs` | no new line in `src/lib.rs`, which another card held until PR 3229 | a sibling module and one `lib.rs` line |
| D-2 | Slice 3 has its own branch `spec/20261119-bundle-build`, cut from the READY head `b49fd5f070` of slice 2 | the parent SPEC says a later slice gets its own branch; the code needs slice 2's, which is not in the tip yet | the same branch as slice 2 |
| D-3 | The merge commits the tool writes carry no `Moai-*` trailers | the user's 2026-11-09 decision: a merge commit a tool makes is a record, trailers are not required, and `trailer-check` skips merges | add trailers |
| D-4 | Writes run with `identity::scrub_git_env` (the user's global identity and settings), reads with `audit_delta::command` | `audit_delta::command` turns the global configuration off, so no author identity | one wrapper |
| D-5 | `ok` is false exactly when a head was dropped; a skipped head does not fail the build | a skipped head is a fact, a dropped one is work the leader must look at | `ok` false on any skip |
| D-6 | `--date YYYYMMDD` sets the date part of the branch name, default today in UTC | tests need a fixed name; bundles 2 to 5 used `m-flow/bundle-20261119-<n>` | no flag |
| D-7 | A build with no head left after the heads already in the tip is refused before creating anything | an empty branch is a name used for nothing | create the branch anyway |
| D-8 | No `--no-verify`; a hook that refuses is a REQ-8 stop | hooks are the repository's own rules; the ruleset and CI are not bypassed (card C-119) | a flag |
| D-9 | `pinned` is read back with `git rev-parse` after the last merge, never computed | the report states what the read shows | n/a |
| D-10 | `rebuild` names heads as `C-<n>=<sha>` | a pinned head cannot change between the build and the rebuild | `C-<n>` |
| D-11 | `rebuild` quotes a word with a shell character in single quotes (`'` becomes `'\''`) and is empty when no head would be left | audit N2 and N5 of slice 3: a ref name may hold `;` or `$(…)`, and a line with no heads cannot run | no quoting |
| D-12 | The merge drivers named in the git configuration (`git config --get-regexp '^merge\..*\.driver$'`, the clone's and the user's) are all overridden with the text merge of slice 2, next to the manifest driver | audit N1 of slice 3: a second registered driver (`merge=ours`) silently decided a conflict; `bundle check` keeps the clone's own configuration only, because it runs with the user's global configuration off | override the manifest driver only |
| D-13 | A `--date` is eight digits with a month 01 to 12 and a day 01 to 31; a head that names a tag is used as the commit the tag points to (the peeled commit is what is merged and reported) | audit N6 of slice 3 | digits only |

## Result

Revision 2. Run on the code commit `2565f863b3` plus the commit that carries this Result (documents only), branch `spec/20261119-bundle-build`, which holds slice 2 (READY head `b49fd5f070`), slice 3 (audited head `b2c0d412aa`, PASS card note 3139) and the repair of audit notes N1 to N7 below; tip `747738af63` is an ancestor of the head (`git merge-base --is-ancestor`), so no merge is needed; `git diff --check origin/v2026.11.0...HEAD` prints nothing. The binary was built from that tree (`moai lane-env --lane agent-4`).

- AC-1: met - `bundle_build_cli two_heads_make_one_branch_and_worktree_with_a_merge_commit_each` (unchanged, passes).
- AC-2: met - `heads_that_cannot_be_used_are_refused_before_anything_is_created` and, new, `a_date_that_is_not_a_date_is_refused_and_a_tag_head_is_the_commit`: `20261340`, `20260132`, `00000000`, `20260000`, `2026111a`, `2026111`, `202611190` and an empty `--date` are each refused with a message naming `--date`; the repository state and the `worktrees` folder are as before; a tag object's SHA as a head merges the commit and reports it, and `pinned^2` is that commit.
- AC-3: met - `a_conflicting_head_is_dropped_and_named_and_the_rest_merge` and, new, `a_second_registered_merge_driver_does_not_decide`: with `merge.ours.driver=true` chosen by a `.gitattributes` line for `q.txt`, a plain `git merge-tree` exits 0 (asserted) while `build` drops the second head naming `q.txt` and `bundle check` reports one `conflict` with `q.txt` (killed M3, only the manifest driver overridden).
- AC-4: met - `a_head_in_the_tip_or_in_the_branch_is_skipped` (unchanged, passes).
- AC-5: met - `the_number_is_the_lowest_unused_and_existing_names_are_never_touched`, new `a_registered_worktree_whose_folder_is_gone_still_takes_its_number` (`git worktree list --porcelain` shows `prunable`, the build takes `-2`) and new `all_999_numbers_used_is_refused_before_anything_is_created` (999 loose refs; the message "the numbers 1 to 999 of this date are all used"; refs and `worktrees` unchanged). The numbers now come from one `git for-each-ref` of `refs/heads/m-flow/` and `refs/remotes/origin/m-flow/`.
- AC-6: met - `creation_refuses_a_taken_branch_or_folder`, now also with an existing branch checked out nowhere that points to another commit: `create` refuses, the branch did not move, no folder (it pins `-b`, which `-B` would have broken).
- AC-7, AC-8, AC-9: met - `a_git_failure_after_creation_names_what_is_left_behind`, `build_changes_only_the_new_branch_and_worktree`, `the_help_and_the_ops_doc_say_what_build_writes` (unchanged, pass).
- AC-10: met - `the_rebuild_line_is_quoted_and_empty_when_no_head_is_left`: tip `odd;tip` gives `moai bundle build <card>=<sha> --tip 'odd;tip'`; when the only head conflicts the line is empty (killed M4 quoting off, M5 empty rule off).

Audit notes of slice 3: N1 (second driver) AC-3, N2 (no head left) AC-10, N3 (`-b` pinned) AC-6, N4 (registered folder gone) AC-5, N5 (quoting) AC-10, N6 (tag, `--date`, 999 limit, docs heading with `--project`) AC-2 and AC-5, N7 (the earlier Result said trailer-check 10 commits, the range had 11) is fixed below by the number measured now. Open and left for later: none from slice 3.

Also run, all in one heavy slot (`target/agent-4/verify119k.sh`): `cargo fmt --check` exit 0; `cargo clippy --locked --offline --all-targets -- -D warnings` exit 0; whole `cargo test --lib` 1464 passed; targets `bundle_ready_cli` 2, `bundle_check_cli` 11, `bundle_body_cli` 7, `bundle_build_cli` 14, `git_env_scrub` 3, `help_layout` 10, `home_isolation_scan` 18, `hot_file_ops_merge` 3, `product_cli` 79, all passed; `python3 -m unittest tests.ci_scope_test tests.repo_scans_test` 65 OK; five mutants (`target/agent-4/mutate119d.py`: M3 to M5 above and M1, M2 of `SPEC-20261119-bundle-audit-tier`) all KILLED, the files restored (`restored: True`); `moai trailer-check origin/v2026.11.0..HEAD --branch spec/20261119-bundle-build` ok true, failing 0 over 14 commits (merges skipped; the range includes slices 2 and 3).

Not run: the whole Python discover, Linux, the unfiltered full `cargo test` (the head's CI run is cited once it exists), mutants beyond the tier-1 cap of five. Not exercised: a signing or hook refusal of a merge commit (REQ-8 is shown with a missing author identity only), `build` against the real ledger and repository (the leader runs it).
