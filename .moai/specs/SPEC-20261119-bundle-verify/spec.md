---
id: SPEC-20261119-bundle-verify
title: "moai bundle verify: run the check list on the pinned bundle branch under one heavy slot and write the run record"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261119-bundle-tool
---

# moai bundle verify: run the check list on the pinned bundle branch under one heavy slot and write the run record

## Intent

Card C-119, slice 4 (the last of the plan in SPEC-20261119-bundle-tool: REQ-8 and REQ-9 there). Slices 1 to 3 shipped in the release pull request 3231 (merge 4a45b73eea): `bundle ready`, `check`, `body` and `build`. What the leader still does by hand after `build` is the third step of the hand procedure: run the verification list on the pinned SHA and write down what ran. Measured at the local tip e341e361cc: `moai bundle body --run <file>` needs a run record (`commands`, `not_run`, `self_audit`) and `bundle verify` does not exist (`moai bundle --help` lists `ready`, `check`, `body`, `build`), so the leader writes the record file by hand, once per bundle.

`bundle verify` takes the heads and the bundle branch that `build` made, runs every command of the check list in that branch's worktree under one heavy slot (one slot hold for the whole list, not one wait per command), goes on after a failed command, and writes the run record that `body` reads. When a command fails it names the heads whose diff against the tip touches a file the failing output names, so the leader can drop that head and run `build` again without it (the `rebuild` line of `build`).

The check list (parent SPEC, D-4) is data in the code and pinned by a test: the ten check keys of `bundle::CHECKS`, with the commands of the two merged bundle bodies (#3225, #3227) and the lesson of bundle 3 (a Python test is never a cargo target):

| key | command |
| --- | --- |
| `build` | `cargo build --locked --offline --bin moai-cli` |
| `lib_tests` | `cargo test --locked --offline --lib` |
| `home_isolation_scan` | `cargo test --locked --offline --test home_isolation_scan` |
| `product_cli` | `cargo test --locked --offline --test product_cli` |
| `fmt` | `cargo fmt --check` |
| `clippy` | `cargo clippy --locked --offline --all-targets -- -D warnings` |
| `trailer_check` | `<moai-cli> trailer-check <tip>..HEAD` |
| `spec_check` | `<moai-cli> spec check --all` |
| `result_coverage` | `<moai-cli> spec check <SPEC id> --closure`, one per SPEC the heads name |
| `unittest` | `python3 -m unittest discover -s tests -p '*_test.py'` |

and, as commands without a key, `cargo test --locked --offline --test <name>` for every `tests/<name>.rs` that `tip...branch` adds or changes and every `--test <name>` named in a `verify:` line of a SPEC that `tip...branch` adds or changes. The commands are run as the leader runs them; whether a lane may run the whole `--lib` or the Python discovery on its own machine is the lane rules' question, not this tool's: this tool is the leader's, and its tests run it against stub `cargo` and `python3`.

## Non-goals

- Pushing the branch, opening or editing a pull request, fetching, writing a card note or a ledger row, changing a repository setting, moving or deleting a ref or worktree (REQ-12 of the parent SPEC).
- Deciding that a failure is acceptable: the record names every failed command and `body` refuses a record with one; the leader decides.
- Writing the Self-Audit lines (`plan`, `red`, `mutation`, `clone`, `host`, `receipt`): they are what a person did. `--self-audit <file>` copies a JSON object the leader wrote; without it the record's `self_audit` is empty and `body` names the six gaps.
- The standing not-run items (Linux, the CI run): they are not commands of this list; the leader adds them as `not_run` entries to the record file before `body` reads it (`body` asks for each check of the list to have run or to be named not run, and these are the leader's words).
- A configuration file for the list (parent D-4).

## Requirements

- REQ-1: When `bundle verify` is given heads and a bundle branch, `bundle verify` shall run every command of the check list in the worktree of that branch under one heavy slot, record each command, and print the SHA the branch had when it began.
- REQ-2 (REQ-1): IF the branch has no worktree, the HEAD of its worktree is not the branch's SHA, the worktree has uncommitted changes to tracked files, or a head is in neither the tip nor the branch, THEN `bundle verify` shall refuse, name every such fact, run nothing and write nothing.
- REQ-3: While the check list runs, `bundle verify` shall go on after a command that fails, so that every command of the list has an exit status or a `not_run` reason in the record.
- REQ-4 (REQ-3): IF a command needs something an earlier command should have made (the built `moai-cli` after a failed build) or has nothing to run on (no SPEC named by the heads), THEN `bundle verify` shall list its check in `not_run` with that reason, not counting it as a failed command.
- REQ-5: When the list has ended, `bundle verify` shall write one new folder holding `run.json`, a JSON run record with `commands`, `not_run` and `self_audit` in the shape `bundle body --run` reads, and one log file per command.
- REQ-6 (REQ-5): IF the output folder exists, cannot be created, or the file given with `--self-audit` is not a JSON object, THEN `bundle verify` shall refuse before it runs anything and write nothing.
- REQ-7 (REQ-3): IF a command fails, THEN `bundle verify` shall name for it the heads whose `tip...head` diff touches a path the failing output names (the file of a failing `--test <name>` target included), or `unattributed`.
- REQ-8: The check list shall be data in the code, pinned by a test: the ten keys of `bundle::CHECKS`, in one order, each with one command, plus the commands for changed tests and SPEC `verify:` targets.
- REQ-9: `bundle verify` shall not fetch, push, open or edit a pull request, write a card note or a ledger row, change a repository setting or move a ref.
- REQ-10: The help of `bundle verify` shall say first that it runs commands and writes files under the worktree's `target/`.

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-5): in a temporary repository with a bundle branch built from two heads, stub `cargo` and `python3` on `PATH`, `bundle verify` runs every command of the list with the worktree as its folder, prints the pinned SHA, writes the folder with `run.json` and a log per command, and `bundle body --run run.json` accepts the record when every command exited 0.
  verify: `cargo test --locked --offline --test bundle_verify_cli every_command_of_the_list_runs_in_the_branch_worktree`
- AC-2 (REQ-2, REQ-6): a branch without a worktree, a worktree at another commit, a worktree with a modified tracked file, a head that is in neither the tip nor the branch, an existing output folder, a branch that is not there and a `--self-audit` file that is not an object are each refused with every fact named, and no stub ran and no file was written; a slot that is refused leaves no folder; an untracked file is allowed; a relative `--out` or `--moai-cli` is the caller's own.
  verify: `cargo test --locked --offline --test bundle_verify_cli refusals_run_nothing_and_write_nothing`; `cargo test --locked --offline --test bundle_verify_cli a_missing_branch_and_a_busy_slot_are_refused_with_nothing_left`; `cargo test --locked --offline --test bundle_verify_cli a_relative_path_is_the_callers_own_not_the_worktrees`; `cargo test --locked --offline --test bundle_verify_cli the_default_folder_is_under_the_worktree_target_and_untracked_files_are_allowed`
- AC-3 (REQ-3, REQ-4): a failing `fmt` does not stop `clippy`, the tests or the rest; a failing `build` puts `trailer_check`, `spec_check` and `result_coverage` in `not_run` with the reason and not among the failed commands; a bundle that names no SPEC puts `result_coverage` in `not_run`.
  verify: `cargo test --locked --offline --test bundle_verify_cli a_failed_command_does_not_stop_the_list_and_a_missing_binary_is_not_run`; `cargo test --locked --offline --test bundle_verify_cli a_bundle_that_names_no_spec_has_no_result_coverage_to_run`
- AC-4 (REQ-7): a failing `--test <name>` target is attributed to the head whose diff adds that test file, a failure naming a file two heads touch names both, a failure that names a changed file by its absolute path in the worktree (`cargo fmt --check` prints `Diff in <path>:<line>:`) is attributed too, and a failure naming no changed file is `unattributed`.
  verify: `cargo test --locked --offline --test bundle_verify_cli a_failure_names_the_head_that_touches_its_file`
- AC-5 (REQ-8): the keys and commands are the table of the Intent, in order; a changed or added `tests/<name>.rs` and a `--test <name>` of a changed SPEC's `verify:` line each add one command, with no repeat; a name with no `tests/<name>.rs` is in `not_run`.
  verify: `cargo test --locked --offline --test bundle_verify_cli the_list_is_the_documented_data`
- AC-6 (REQ-9): the refs of the repository and the dump of the ledger are the same before and after `bundle verify`.
  verify: `cargo test --locked --offline --test bundle_verify_cli verify_writes_no_ref_and_no_ledger_row`
- AC-7 (REQ-10): the help of `bundle verify` starts with `Runs commands.` and the operations document names the table, the record, the refusals and the attribution.
  verify: `cargo test --locked --offline --test bundle_verify_cli the_help_and_the_ops_doc_name_what_verify_does`
- AC-8 (REQ-1…REQ-10): neighbours.
  verify: `cargo test --locked --offline --test bundle_build_cli`; `cargo test --locked --offline --test bundle_check_cli`; `cargo test --locked --offline --test bundle_body_cli`; `cargo test --locked --offline --test bundle_ready_cli`; `cargo test --locked --offline --test help_layout`; `cargo test --locked --offline --bins`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/bundle.rs` (the line `pub mod verify;` and what `verify` takes from the module), `src/bundle/verify.rs`, `src/ops/bundle.rs`, `tests/bundle_verify_cli.rs`, `docs/design-review/appendix/operations/bundle.md`, `.moai/specs/SPEC-20261119-bundle-verify/**`, `.moai/specs/SPEC-20261119-bundle-tool/**` (its Result only)
- read-only: `src/bundle/build.rs`, `src/heavy_slot.rs`, `src/lane_env.rs`, `src/identity.rs`, `src/audit_delta.rs`, `src/gate.rs`, `tests/bundle_build_cli.rs`, `tests/common/**`
- forbidden: `src/help.rs` (reserved by C-120; no help example is added), `src/lib.rs` and `src/main.rs` (no new module line is needed), any other change under `src/`, the real HOME and ledger as a write target, `skills/`, `.github/`, any push of a ref other than this branch, any pull request

## Plan

1. Tests first (`tests/bundle_verify_cli.rs`, AC-1 to AC-7): they fail before the code because `bundle verify` does not exist.
2. `src/bundle/verify.rs`: the list as data (`LIST`, the extras from `git diff --name-status tip...branch` and the changed SPECs' `verify:` lines); the refusals of REQ-2 and REQ-6 before anything runs; an output folder (`--out`, default `<worktree>/target/bundle-verify/<sha12>/`) with a script that has one line per command (its own log, `.exit`, `.start`, `.end`), started as `<this executable> heavy-slot-acquire --max <n> --min-free-gb <g> --name bundle-verify -- sh <script>` so the whole list holds one slot; after it ends the logs and exit files are read into the record.
3. Attribution (REQ-7): a pure function over the failing output, the paths of `tip...branch` and the paths of each head; unit tests.
4. `src/ops/bundle.rs` (`Verify` action and its help), the operations document, mutants, the SPEC Result.

## Risks

- A verify run is long (a build and the whole `--lib`); the slot may wait or refuse (exit 75). Nothing is recorded then and the command says so; a second run is a new folder.
- The commands run with the caller's environment. A `CARGO_TARGET_DIR` the caller set is used for the built `moai-cli`; without it the worktree's own `target/` is.
- Attribution by paths named in an output can name a head that only shares a file name with the cause, or none: `unattributed` is the honest answer and the leader decides (parent D-10).
- `python3` is whatever `PATH` says (`--python` overrides); on a machine whose `python3` is old the `unittest` entry fails and says so in its log.
- The tool trusts the worktree it is given only as far as REQ-2: HEAD equals the branch's SHA and nothing is uncommitted; a command that edits tracked files while it runs is not detected here (the leader sees it in `git status`).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | One script under one `heavy-slot-acquire` call, not one slot per command | parent REQ-8 and D-5: one hold for the whole list, a failure does not release it | A flag for one slot per command |
| D-2 | The slot is taken by running this same executable's `heavy-slot-acquire` as a child | the slot code lives in the binary; no second implementation of the lock; the slot options are passed through (`--slot-dir`, `--max`, `--min-free-gb`) so tests use their own | Call `heavy_slot` in process |
| D-3 | Heads are given as for `build` (`C-<n>` or `C-<n>=<sha>`) and used only for the "head is in the branch" refusal and the attribution | REQ-7 needs each head's diff; the branch alone does not say which head touched a file | Attribute by commit message trailer |
| D-4 | The record's `self_audit` comes from `--self-audit <file>` or stays empty | the six lines say what a person did; inventing them would be a placeholder | Generate from the card notes |
| D-5 | A check that cannot run is `not_run` with a reason, never a failed command and never absent | parent "What a check run records": a command that was not run is `not_run` | Count it as failed |
| D-6 | The built binary for `trailer_check`, `spec_check` and `result_coverage` is `<CARGO_TARGET_DIR or worktree target>/debug/moai-cli` (`--moai-cli` overrides) | the check must judge the branch with the branch's own code | Always use the installed `moai` |
| D-7 | `python3 -m unittest discover -s tests -p '*_test.py'` is the one Python entry | parent D-4: the discovery CI runs; no module list to get wrong | A module list from the heads |

## Result

- AC-1: met — `cargo test --locked --offline --test bundle_verify_cli every_command_of_the_list_runs_in_the_branch_worktree`: ok (ten commands in the table's order, each logged in the worktree, `run.json` accepted by `bundle body`).
- AC-2: met — `refusals_run_nothing_and_write_nothing`, `a_missing_branch_and_a_busy_slot_are_refused_with_nothing_left`, `a_relative_path_is_the_callers_own_not_the_worktrees` and `the_default_folder_is_under_the_worktree_target_and_untracked_files_are_allowed`: ok (no stub ran, no folder, state unchanged).
- AC-3: met — `a_failed_command_does_not_stop_the_list_and_a_missing_binary_is_not_run` and `a_bundle_that_names_no_spec_has_no_result_coverage_to_run`: ok.
- AC-4: met — `a_failure_names_the_head_that_touches_its_file` and the fmt case of `a_failed_command_does_not_stop_the_list_and_a_missing_binary_is_not_run` (absolute path): ok; unit tests in `src/bundle/verify.rs` (`heads_are_named_by_their_own_diff_in_the_order_given`) ok.
- AC-5: met — `the_list_is_the_documented_data`: ok.
- AC-6: met — `verify_writes_no_ref_and_no_ledger_row`: ok.
- AC-7: met — `the_help_and_the_ops_doc_name_what_verify_does`: ok.
- AC-8: met — `cargo test --locked --offline --test <name>` for bundle_build_cli, bundle_check_cli, bundle_body_cli, bundle_ready_cli: ok (every `tests/bundle*.rs`, 5 targets, 0 non-zero); the pre-READY checks are in the card's READY note.
