---
id: SPEC-20261119-bundle-ready-check
title: "moai bundle ready, check, body: read the READY heads, judge them, write the bundle PR body"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261119-bundle-tool
---

# moai bundle ready, check, body: read the READY heads, judge them, write the bundle PR body

## Intent

Card C-119, slice 2 (the first code slice of `SPEC-20261119-bundle-tool`, whose Plan fixes the boundary: the CLI reads and judges and writes a text, the leader chooses, pushes and opens the pull request). The leader directed this slice: audit notes N-a to N-j of slice 1 are applied, the slice is read-only verification plus the pull request body, and it goes on the same branch, pushed only (no pull request).

Three actions under one `moai bundle`, none of which writes the ledger, a ref, a branch or a remote:

- `bundle ready`: the READY heads of the ledger, one line per card, newest line wins.
- `bundle check`: every per-head check for the heads the caller names, all findings in one result (bundle 5 lost two hours because each failure stopped the whole bundle).
- `bundle body`: the bundle pull request body from the heads' facts and a run record file, refused with every missing fact named when a fact has no record.

`build` (slice 3) and `verify` (slice 4) are not here. Until `verify` exists the leader writes the run record file by hand; the body is still generated from facts, never from memory of what ran.

Measured facts this slice rests on are those of `SPEC-20261119-bundle-tool` (Intent; counts pinned at note 2931). Two more, read in this slice from the code at tip `d25ecb1ba1`: `git merge-tree --write-tree` exits 0 clean and 1 with conflicts and prints the tree id first, then, with `--name-only`, the conflicted file names up to an empty line; the Self-Audit checker (`self_audit::check`, `src/self_audit.rs`) requires the `## Self-Audit` section to hold `mutation:` (the line must say FAILED), `clone:`, `host:`, `receipt:`, `plan:` (a number of seven digits or more) and `red:`.

## Non-goals

- Fetching, pushing, opening, editing or merging a pull request, writing a card note, a ref or any ledger row (REQ-12). The caller fetches first.
- Choosing heads, accepting an audit, deciding a repair delta is enough: `check` reports `audit_missing` with the nearest ancestor `PASS` (`ancestor_pass`) and the leader decides.
- Building the integration branch (slice 3) and running the check list (slice 4). A conflict found by `check` is a finding, never resolved.
- A conflict check over three or more heads together: `check` merges each head with the tip and each head with every earlier head named, two at a time; the exact three-way result is `build`'s.
- Reading the verdict's tier or author: any `PASS` naming the head counts, and the result says so.

## Requirements

- REQ-1: When `moai bundle ready` is run, `moai bundle ready` shall list, for each card whose newest note with the first-line marker `LANE-READY-INTEGRATION` reads under the grammar of D-6 of the parent SPEC, the card, the note number, the head, the SPEC ids named by the `Moai-Spec` trailers of `tip..head` and whether the head is on `origin`, with the tip's SHA and commit time.
- REQ-2 (REQ-1): IF the newest marker note of a card does not read under that grammar or names another card, THEN `moai bundle ready` shall list the note as `unreadable` with its card and note number and use no head from the card's older notes.
- REQ-3: When `moai bundle check` is given heads as `C-<n>` (the card's newest readable READY head) or `C-<n>=<40 hex>`, `moai bundle check` shall judge every head on every check and report all findings in one result, each with its code, card, head and detail.
- REQ-4 (REQ-3): IF a head fails a check, THEN `moai bundle check` shall report it under one stable code: `no_ready_line`, `head_missing`, `no_merge_base`, `not_on_origin`, `already_merged`, `audit_missing`, `trailers`, `spec`, `result_missing` or `conflict`, and for a conflict the files and the other head, if any.
- REQ-5 (REQ-4): IF the audit floor of a head (`gate::tier_of(Some(0), <paths of tip...head>)`) is 1 or more and no note of the head's card is a verdict (first line `AUDIT `, read by `gate::parse_verdict` unchanged) with the word `PASS` naming exactly that head, THEN `moai bundle check` shall report `audit_missing` with the floor and the nearest `PASS` on an ancestor of the head.
- REQ-6 (REQ-3): IF a head is not present in the local repository, THEN `moai bundle check` shall report `head_missing` ("fetch first"), skip that head's other checks and judge the remaining heads.
- REQ-7 (REQ-3): IF `moai bundle check` finds anything, THEN `moai bundle check` shall print the whole result with `ok` false.
- REQ-8: When `moai bundle body` is given heads, a bundle branch and a run record file, `moai bundle body` shall write a pull request body with the `audit-tier:` line from `gate::tier_of(Some(0), <paths of tip...branch>)`, one `Moai-Spec:` line per SPEC, a table of the heads, an `Integration:` line with the SHA that `git rev-parse` gives for the branch and its size counted from the merge base with the tip, `## Self-Audit` from the record and `## Result` from the record and a count line from each SPEC's own `## Result`.
- REQ-9 (REQ-8): IF a fact the body needs has no record, THEN `moai bundle body` shall refuse, name every such fact in one answer and write no body: a head not contained in the branch, a head with any finding of `bundle check`, a check of the list in the Plan that neither ran nor is named in `not_run` with a reason, a command whose exit is not 0, a Self-Audit line or a record value with a line break, a SPEC with no `## Result` line for one of its ACs.
- REQ-10 (REQ-8): IF the text fails `self_audit::check`, THEN `moai bundle body` shall refuse with the names it prints and write no body.
- REQ-11: IF the path named by `--out` already holds a file, THEN `moai bundle body` shall refuse and leave that file unchanged.
- REQ-12: `moai bundle` shall not fetch, push, open, edit or merge a pull request, create or move a ref or branch, write a card note or any ledger row, or change a setting; the git commands it runs write no ref, index or work tree (the list is in the Plan; `merge-tree --write-tree` leaves unreferenced objects that `git gc` removes), through the scrubbed git environment.

## Acceptance

- AC-1 (REQ-1, REQ-2): in a temporary ledger with a card whose notes are an old line, a newer repair line and a note that starts with the marker but does not read, `moai bundle ready` lists the newest readable head per card, the SPEC ids from the trailers, `on_origin`, and the unreadable note with its card and note number; a card whose newest marker note is unreadable lists no head.
  verify: `cargo test --locked --offline --test bundle_ready_cli ready_lists_the_newest_readable_line_per_card_and_names_the_unreadable_ones`
- AC-2 (REQ-3, REQ-4, REQ-7): `check` on a clean head answers `ok` true with no finding and exit 0; on two heads that each fail something it answers every finding of both heads, with the codes `not_on_origin`, `already_merged`, `trailers`, `spec`, `result_missing` and `audit_missing` each produced once, and `ok` false.
  verify: `cargo test --locked --offline --test bundle_check_cli a_clean_head_has_no_finding`; `cargo test --locked --offline --test bundle_check_cli findings_of_all_heads_come_together_with_stable_codes`
- AC-3 (REQ-5): a head whose tip has moved on since its merge base is judged by what the head adds (three dots), not by the tip's own changes; a head that changes only `docs/` and `tests/` needs no verdict; a head that changes `src/` needs a `PASS` naming exactly its SHA; a `PASS` for an ancestor is reported as `ancestor_pass` and still `audit_missing`; a note starting `AUDIT ` that does not read is listed as `unreadable_verdict` and does not fail the check.
  verify: `cargo test --locked --offline --test bundle_check_cli the_audit_floor_decides_which_heads_need_a_pass`; `cargo test --locked --offline --test bundle_check_cli a_head_behind_the_tip_does_not_count_the_tips_own_changes`
- AC-4 (REQ-4): two heads that change the same lines conflict with each other and with the tip when one of them is merged first; the files are named; a repository whose config registers a merge driver for `skills/moai/references/source-manifest.json` that always succeeds still reports the conflict, because the check runs the text merge GitHub runs.
  verify: `cargo test --locked --offline --test bundle_check_cli a_conflict_is_found_without_the_merge_driver`
- AC-5 (REQ-4, REQ-6): a head SHA that does not exist locally gives `head_missing` for that head and a head with no merge base with the tip gives `no_merge_base`; in both cases the other heads are judged and the answer is `ok` false (exit 0, D-3).
  verify: `cargo test --locked --offline --test bundle_check_cli a_head_that_is_not_here_is_reported_and_the_rest_are_judged`; `cargo test --locked --offline --test bundle_check_cli a_head_with_no_merge_base_is_one_finding_and_not_an_abort`
- AC-6 (REQ-8, REQ-10): on a bundle branch of two merged heads and a complete run record, `body` writes text whose first line is `audit-tier: N <paths>`, one `Moai-Spec:` line per SPEC, the heads table with each head's verdict note, an `Integration:` line with the branch's SHA and the size `git diff --numstat` gives from the merge base (also when the tip has moved on since), and which passes `self_audit::check`.
  verify: `cargo test --locked --offline --test bundle_body_cli the_body_carries_the_facts_and_passes_the_self_audit_checker`; `cargo test --locked --offline --test bundle_body_cli the_integration_line_counts_what_the_branch_adds_when_the_tip_moved`
- AC-7 (REQ-9, REQ-10): a record with a missing check, a failed command and a missing Self-Audit line, and a branch that lacks one head, is refused with all four facts named in one answer and no body or file; a check both run and named not run, a record value with a line break and a head with a `check` finding are refused; a Self-Audit text the checker rejects is refused by REQ-10.
  verify: `cargo test --locked --offline --test bundle_body_cli a_body_with_missing_facts_is_refused_naming_all_of_them`; `cargo test --locked --offline --test bundle_body_cli more_records_the_body_refuses`
- AC-8 (REQ-11): `--out` to an existing path is refused and the file is unchanged; to a new path the file holds the body.
  verify: `cargo test --locked --offline --test bundle_body_cli out_writes_a_new_file_only`
- AC-9 (REQ-12): after `ready`, `check` and `body`, every ledger table, `git for-each-ref`, the branch list and `git status` are as before.
  verify: `cargo test --locked --offline --test bundle_check_cli the_commands_change_no_ledger_row_and_no_ref`
- AC-10 (REQ-12): the first line of each action's help says what it does to the repository, and the operations document has a `bundle` page naming the three actions, the finding codes and the run record file.
  verify: `cargo test --locked --offline --test bundle_body_cli the_help_and_the_ops_doc_say_what_each_action_does`
- AC-11 (REQ-1 … REQ-12): neighbours.
  verify: `cargo test --locked --offline --test git_env_scrub`; `cargo test --locked --offline --test home_isolation_scan`; `cargo test --locked --offline --test help_layout`; `cargo test --locked --offline --test product_cli`

## Scope

- write: `src/bundle.rs`, `src/ops/bundle.rs`, `tests/bundle_ready_cli.rs`, `tests/bundle_check_cli.rs`, `tests/bundle_body_cli.rs`, `docs/design-review/appendix/operations/bundle.md`, `.moai/specs/SPEC-20261119-bundle-ready-check/**`, `.moai/specs/SPEC-20261119-bundle-tool/**`, and one line in `src/lib.rs` (`pub mod bundle;`; reserved by card C-75; leader decision B, card note 2970: taken after pull request #3229 is merged and C-75's holder has released, and left uncommitted until then).
- read-only: `src/gate.rs`, `src/trailers.rs`, `src/spec_doc.rs`, `src/self_audit.rs`, `src/identity.rs`, `src/card/command.rs`, `src/ledger/command.rs`, `src/ops/card.rs`, `src/ops/trailer_check.rs`.
- forbidden: any other change under `src/`, the real HOME and ledger as a write target, `skills/`, `.github/`, any push of a ref other than this branch, any pull request.

## Plan

**Code.** `src/bundle.rs` in the library crate (the git calls and the scrubbed environment live there; `src/ops` spawns no git): `ready`, `check` and `body` as functions returning `serde_json::Value`; the ledger is read through `card::command::list` and `card::command::show`; the SPEC text of a head is read with `git show <head>:.moai/specs/<id>/spec.md` and parsed by `spec_doc::parse` (no closure: the status is `draft` until the integration). `src/ops/bundle.rs` is the clap layer (`--project`, `--home`, `--dir`, `--tip`, default tip `origin/v2026.11.0`), the same one-word-per-action shape as `card`.

**Git commands (the whole list; REQ-12).** `rev-parse --verify <rev>^{commit}`, `merge-base --is-ancestor`, `branch -r --contains` and `for-each-ref refs/remotes/origin`, `log --no-merges --format=…`, `diff --name-only <tip>...<head>`, `show <head>:<path>`, `-c merge.moai-manifest.driver='git merge-file -L ours -L base -L theirs %A %O %B' merge-tree --write-tree --name-only <a> <b>` (exit status read from the process; 0 clean, 1 conflict; the driver override because GitHub never runs a registered driver), `rev-list --count`. No `fetch`, `push`, `merge`, `checkout`, `branch` write, `update-ref`, `worktree`. Every command runs through `identity::scrub_git_env`.

**Findings.** Each finding is `{code, card, head, detail}`; codes are the nine of REQ-4 plus `unreadable_verdict` as a note (not a failure). `trailers` is `trailers::check(dir, "<tip>..<head>", <branch>, None)`, where branch is the `origin` branch whose tip is the head (`detached` when none: the SPEC check below then catches a missing `Moai-Spec`). `spec` is any error of `spec_doc::parse` for each SPEC the head names, or a missing file, or no `Moai-Spec` trailer at all. `result_missing` lists the ACs of a named SPEC without a line in its `## Result`.

**Run record file (JSON, one object).**

```
{
  "commands": [ {"check": "<key>", "command": "<text>", "exit": 0, "counts": "<optional text>"} ],
  "not_run":  [ {"check": "<key>", "reason": "<text>"} ],
  "self_audit": { "mutation": "…FAILED…", "clone": "…", "host": "…", "receipt": "…", "plan": "<text with a 7-digit id>", "red": "…" }
}
```

The check keys are data in the code, pinned by a test (D-4 of the parent): `build`, `lib_tests`, `home_isolation_scan`, `product_cli`, `fmt`, `clippy`, `trailer_check`, `spec_check`, `result_coverage`, `unittest`. A command whose `check` is not a key is carried into `## Result` as an extra line. `body` takes the `Moai-Spec` line, the table row and the SPEC result line from the heads; the `Self-Audit` lines and the `## Result` command lines from the record.

**Order.** (1) `src/lib.rs` line once the leader decides; (2) `bundle ready`; (3) `bundle check`; (4) `bundle body`; (5) help, operations page, tests, mutants; each step a commit.

## Risks

- `check` trusts the card notes it reads (see the parent's Risks): a `PASS` note is a note, not proof of an independent audit. The result carries `audit_note` saying so.
- `merge-tree --write-tree` leaves unreferenced objects in the object store (`git gc` removes them) and needs git 2.38 or newer; no ref, index or work tree changes.
- The body carries fixed sentences that are not read from facts (the M-Flow one-line introduction, `built from`, `Local verification on`, `see the SPEC's Result`, the Claude Code attribution line).
- The pairwise conflict check can miss a three-way conflict; `build` (slice 3) is the exact check. The result names the limit.
- Every ledger read goes through `card show` JSON; a change of its keys breaks the reader, which is why the tests build a real temporary ledger through the product commands.
- `src/lib.rs` is reserved by C-75 (leader decision B): the module compiles and its tests run in this worktree with the one line present, but the line cannot be committed until C-75's change is merged and released, so the head cannot be pushed before then.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `body` is in this slice and reads a run record file | the leader's direction on card C-119; until `verify` exists a file written by hand carries what ran | move `body` back to a later slice |
| D-2 | Heads are `C-<n>` or `C-<n>=<sha>`, never a bare SHA | the audit check needs the card; a bare SHA could be on two cards | accept a bare SHA when exactly one READY line names it |
| D-3 | `check` answers `ok:false` with every finding and exits 0 | the exit rule of operations (`ok:false` → exit 1) is written by operation name in `src/main.rs`, which card C-75 reserves; one line there gives `bundle` the exit 1 of `trailers.check` | add that line to `src/main.rs` |
| D-4 | The conflict check is pairwise (each head with the tip, each with every earlier head) | one `merge-tree` per pair needs no worktree; the exact three-way merge is `build` | a temporary index per bundle |
| D-5 | `body` refuses on every missing fact in one answer | the card's rule: no placeholder; one round, not one refusal per fact | stop at the first |
| D-6 | The check keys are ten fixed words | the union of the two merged bodies' lists (parent D-4) | a configuration file |
| D-7 | Proposal, not built: a verdict line may carry the tier it was audited at in the text before ` head `, as `AUDIT C-<n> (<SPEC>, tier <N>, …) head <sha> base <sha> — verdict: PASS`; `parse_verdict` reads from the last ` head ` on, so the line still reads as today | the tool cannot tell at which tier a `PASS` was given (audit N-d of slice 1); a floor-2 head is satisfied by a tier-1 `PASS` today | the leader decides whether lanes and auditors write it; then `check` reads `tier <N>` and reports a `PASS` below the floor |

## Result

Run on the head `a98161107b` plus the edits of the commit that carries this Result and the one-line `pub mod bundle;` in `src/lib.rs` (uncommitted, leader decision B), with a binary built from that tree; tip `924da079d5`.

- AC-1: met — `cargo test --test bundle_ready_cli` → 1 passed. Read-only smoke on the real ledger (`moai bundle ready --project /Users/goos/MoAI/moai-adk`): 20 heads, `unreadable` = C-33 note 2854 and C-60 note 2646, C-99 `already_in_tip` true.
- AC-2: met — `cargo test --test bundle_check_cli` → 7 passed, among them `a_clean_head_has_no_finding` and `findings_of_all_heads_come_together_with_stable_codes` (six codes over three heads, exact set).
- AC-3: met — `the_audit_floor_decides_which_heads_need_a_pass` (floor 0 and 1, an ancestor `PASS` shown with `commits_behind` 1, a `FAIL` and an unreadable note not counted) and `a_head_behind_the_tip_does_not_count_the_tips_own_changes` (killed the three-dot mutant M1).
- AC-4: met — `a_conflict_is_found_without_the_merge_driver`: the clone's driver (`true`) hides the conflict from a plain `git merge-tree` (exit 0, asserted), and the check still reports it with the file (killed M2).
- AC-5: met — `a_head_that_is_not_here_is_reported_and_the_rest_are_judged` (also `no_ready_line`).
- AC-6: met — `cargo test --test bundle_body_cli` → 4 passed; `the_body_carries_the_facts_and_passes_the_self_audit_checker` also runs `moai pr-self-audit-check --file` on the body, exit 0.
- AC-7: met — `a_body_with_missing_facts_is_refused_naming_all_of_them`: "4 missing fact(s)" naming the missing `fmt` check, the failed command, the empty `host` line and the head not in the branch; no file (killed M5).
- AC-8: met — `out_writes_a_new_file_only` (an existing file and a folder refused, file unchanged).
- AC-9: met — `the_commands_change_no_ledger_row_and_no_ref`: the HOME snapshot (ledger bytes included) and the refs, status and HEAD of the work clone and the bare origin are equal before and after `ready`, `check` and a refused `body`; `body` with a result also leaves the refs as before.
- AC-10: met — `the_help_and_the_ops_doc_say_what_each_action_does`: each action's help starts `Reads only.`; the operations page names the actions, codes and record keys.
- AC-11: met — `git_env_scrub` 3 passed, `home_isolation_scan` 18, `help_layout` 10, `product_cli` 79, `hot_file_ops_merge` 3.

Also run: `cargo fmt --check` exit 0; `cargo clippy --locked --offline --all-targets -- -D warnings` exit 0; the whole `cargo test --lib` 1458 passed (5 are this slice's: the READY grammar, the check keys, the head arguments); `python3 -m unittest tests.ci_scope_test tests.repo_scans_test` 65 OK (it caught `docs/…` literals in the tests, renamed to `notes/…`); five mutants (`target/agent-4/mutate119.py`: M1 three dots, M2 driver override, M3 oldest note, M4 floor 2, M5 exit threshold) all KILLED; `moai trailer-check origin/v2026.11.0..HEAD` 4 commits, failing 0; `git merge-tree` against the tip exit 0 and no file shared with the tip's change.

Not run: the whole Python discover (about 40 minutes here), Linux, the unfiltered full `cargo test` (the head's CI run is cited once it exists), mutants beyond the tier-1 cap of five. First failures on the way, all fixed: the refusal text "exited N" was classified as an `io` error by `contract::classify` (reworded "ended with status N"); `clippy::single_match`; a help-test word.
