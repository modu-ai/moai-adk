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

- AC-1 (REQ-1, REQ-4): a child built with `clean_env` sees `MOAI_CLI_TEST_PROCESS_TABLE=empty`; after the test calls `use_real_process_table` its children see `real`, and another test thread still sees `empty` · verify: `cargo test --locked --offline --test process_table_default`
- AC-2 (REQ-2, REQ-3): the named slow test files pass with the switch on, and their in-use tests still see the program that holds the folder · verify: `cargo test --locked --offline --test home_install_apply_cli --test home_retire_cli --test home_state_apply --test home_state_plan --test home_install_go_retire --test home_install_plan_cli --test update_home_layer_cli --test hygiene_apply_cli --test hygiene_cli --test optimize_plan_cli`
- AC-3 (REQ-3): every named file sets the switch, and every named file that starts a holder program calls `use_real_process_table` · verify: `cargo test --locked --offline --test process_table_default the_named_files_choose_their_table`
- AC-4 (REQ-5): the Result lists the before and after time of each named target from one nextest run each, on the same machine · verify: read the Result

## Scope

- write: this SPEC folder, `tests/common/mod.rs`, the new `tests/process_table_default.rs`
- write: `tests/home_install_apply_cli.rs`, `tests/home_retire_cli.rs`, `tests/home_state_apply.rs`, `tests/home_state_plan.rs`, `tests/home_install_go_retire.rs`, `tests/home_install_plan_cli.rs`, `tests/update_home_layer_cli.rs`, `tests/hygiene_apply_cli.rs`, `tests/hygiene_cli.rs`, `tests/optimize_plan_cli.rs`
- write (found by the first check affected run): `tests/cargo_env_developer_dir.rs` and `tests/home_isolation_scan.rs` pin the variables `clean_env` passes on and learn the switch; `tests/process_scan_cost_cli.rs` counts lsof calls and reads the real table
- read-only: everything else, `src/` included

## Plan

1. Before: one nextest run of the ten named targets on the tip, durations kept.
2. RED: `tests/process_table_default.rs` asserts that a `clean_env` child sees the switch; it fails by assertion on the tip.
3. GREEN: `tests/common/mod.rs` names the switch (`PROCESS_TABLE_SWITCH`, `NO_PROCESS_TABLE`, `REAL_PROCESS_TABLE`) and keeps the choice per test thread: `process_table()` answers `empty` until the test calls `use_real_process_table()`; `clean_env` and the named files' command helpers pass `process_table()`; a test calls `use_real_process_table()` right before it starts a program that holds a folder or a file.
4. The file-list case (AC-3) reads the named files only (no directory walk).
5. After: the same nextest run; the Result records both.

## Risks

- A test that judges "in use" left under the switch would pass for the wrong reason only if it did not assert the in-use answer; the in-use tests assert `skip-in-use` or a refusal, so under the switch they fail and are found.
- `MOAI_CLI_TEST_PROCESS_TABLE` in a developer's shell does not reach these children: `clean_env` clears the environment first, and the named helpers set the variable explicitly.
- The choice lives in a thread-local: with `--test-threads=1` libtest runs every test on one thread, so a test after an in-use test reads the real table too — slower, never wrong.

## Decisions

- D-1: The switch goes into `clean_env` and into each named file's own helper, not into `.cargo/config.toml` `[env]`, because `[env]` would also reach `cargo run` of a developer's debug build (planning default).
- D-2: The real table is chosen per test thread by `use_real_process_table()`, called where a test starts its holder program, instead of an extra argument on every helper; the in-use tests then need one line each (planning default).

## Result

RED e7f3785f18 on the tip 2fb33d76bf; GREEN is the next commit. Both runs are one `cargo nextest run` of the eleven targets below through `moai heavy-slot-acquire --max 3`, on this machine (load about 17–24 at the start of each).

- AC-1 met: `a_clean_env_child_skips_the_process_table` failed first by assertion (`"unset"` instead of `"empty"`, `target/before.log`) and passes now; `a_test_that_judges_in_use_opts_back_in` passes.
- AC-2 met: the ten named targets pass with the switch on, their in-use tests included (173 passed, 0 failed, `target/after.log`).
- AC-3 met: `the_named_files_choose_their_table` passes.
- AC-4 met: summed test time per target, before → after: `home_install_apply_cli` 313.2 → 119.9 s, `home_retire_cli` 179.7 → 60.5 s, `home_install_go_retire` 121.4 → 13.1 s, `update_home_layer_cli` 111.2 → 50.2 s, `optimize_plan_cli` 95.2 → 112.0 s (no gain: its time is not the process scan), `home_state_apply` 81.7 → 1.6 s, `home_state_plan` 52.4 → 34.0 s, `hygiene_apply_cli` 44.4 → 21.2 s, `home_install_plan_cli` 30.1 → 11.0 s, `hygiene_cli` 5.2 → 3.3 s; all ten 1,034.5 → 426.8 s, and the run's wall clock 212 s → 96 s. The before run is the tip's files (the RED commit changes none of them).
- Not run: the lib unit tests that call the scans (a later slice), the whole suite, the independent audit.
