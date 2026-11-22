---
id: SPEC-20261122-test-process-table-default
title: "Tests skip the machine-wide process scan by default; the in-use tests opt back in"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261121-cheap-process-scan
---

# Tests skip the machine-wide process scan by default; the in-use tests opt back in

## Intent

The parent SPEC gave debug builds a test switch, `MOAI_CLI_TEST_PROCESS_TABLE=empty`, that makes
the two process scans of `src/worktree_gc.rs` answer "no process uses anything" without running
`lsof`. Only one test file uses it, so the home, hygiene and update tests still read the whole
machine's process table on every command: on 2026-11-22 the review of the test load measured
`home_install_apply_cli` at 170 s, `home_retire_cli` at 75 s, `home_state_apply` at 66 s and
`update_home_layer_cli` at 35–63 s locally, against a few seconds in CI (test-review F3 (a); the
user-approved redesign, section C). This slice turns the switch on by default in the test command
helpers, and only the tests that judge "a program uses this folder" read the real table.

## Non-goals

- The lib unit tests that call the scans directly (`src/home_state.rs`, `src/home_install_apply.rs`):
  an environment variable is process-wide there, so they need the table passed in; that is a later
  slice (test-review F3 (b)).
- `tests/worktree_gc.rs` (held by another card) keeps its own command helper.
- The release build: the switch is read only under `debug_assertions` (parent SPEC); nothing here
  changes `src/`.
- Slice 2 of the test-load work (`SPEC-20261121-process-scan-snapshot`) is the leader's to land.

## Requirements

- REQ-1: The shared `clean_env` test helper shall set the test switch to `empty` on the child command.
- REQ-2: The command helpers of the named slow test files shall set the test switch to `empty`.
- REQ-3: A test that judges whether a program uses a folder shall run its command with the real process table.
- REQ-4: The shared test module shall name the switch and its two values in one place.
- REQ-5: The slice Result shall record the time of the named test targets before and after the change.

## Acceptance

- AC-1 (REQ-1, REQ-4): a child built with `clean_env` sees `MOAI_CLI_TEST_PROCESS_TABLE=empty`, and one built with `clean_env` and then `real_process_table` sees no such variable · verify: `cargo test --locked --offline --test process_table_default`
- AC-2 (REQ-2, REQ-3): the named slow test files pass with the switch on, and their in-use tests still see the program that holds the folder · verify: `cargo test --locked --offline --test home_install_apply_cli --test home_retire_cli --test home_state_apply --test home_state_plan --test home_install_go_retire --test home_install_plan_cli --test update_home_layer_cli --test hygiene_apply_cli --test hygiene_cli --test optimize_plan_cli`
- AC-3 (REQ-3): every named file whose tests judge a folder in use calls `real_process_table`, and every other named file sets the switch · verify: `cargo test --locked --offline --test process_table_default the_named_files_choose_their_table`
- AC-4 (REQ-5): the Result lists the before and after time of each named target from one nextest run each, on the same machine · verify: read the Result

## Scope

- write: this SPEC folder, `tests/common/mod.rs`, the new `tests/process_table_default.rs`
- write: `tests/home_install_apply_cli.rs`, `tests/home_retire_cli.rs`, `tests/home_state_apply.rs`, `tests/home_state_plan.rs`, `tests/home_install_go_retire.rs`, `tests/home_install_plan_cli.rs`, `tests/update_home_layer_cli.rs`, `tests/hygiene_apply_cli.rs`, `tests/hygiene_cli.rs`, `tests/optimize_plan_cli.rs`
- read-only: everything else, `src/` included

## Plan

1. Before: one nextest run of the ten named targets on the tip, durations kept.
2. RED: `tests/process_table_default.rs` asserts that a `clean_env` child sees the switch; it fails by assertion on the tip.
3. GREEN: `tests/common/mod.rs` names the switch (`PROCESS_TABLE_SWITCH`, `NO_PROCESS_TABLE`, `REAL_PROCESS_TABLE`), `clean_env` sets it, and a `real_process_table` method removes it; the named files' command helpers set it; the tests that hold a folder with a program opt back in.
4. The file-list case (AC-3) reads the named files only (no directory walk).
5. After: the same nextest run; the Result records both.

## Risks

- A test that judges "in use" left under the switch would pass for the wrong reason only if it did not assert the in-use answer; the in-use tests assert `skip-in-use` or a refusal, so under the switch they fail and are found.
- `MOAI_CLI_TEST_PROCESS_TABLE` in a developer's shell does not reach these children: `clean_env` clears the environment first.

## Decisions

- D-1: The switch goes into `clean_env` and into each named file's own helper, not into `.cargo/config.toml` `[env]`, because `[env]` would also reach `cargo run` of a developer's debug build (planning default).
