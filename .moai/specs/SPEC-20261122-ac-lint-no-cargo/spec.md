---
id: SPEC-20261122-ac-lint-no-cargo
title: "The recorded-AC-command lint reads test names from the test sources and starts no cargo"
status: draft
revision: 1
created: 2026-11-22
---

# The recorded-AC-command lint reads test names from the test sources and starts no cargo

## Intent

The user-approved redesign (`~/.moai/reports/moai-adk/agentic-redesign-2026-11-22.md` §3C, §5 card ④) and test-review F6 measured this: `tests/ac_command_lint.rs` runs `tools/check_ac_commands.py --tree`, and for every `cargo test --test X` line in the recorded documents the script runs `cargo test --offline --locked --test X -- --list` (`tools/check_ac_commands.py` `list_tests`). A target that is not built yet is compiled and linked in the middle of the test: 68–112 s locally against 1.9 s in CI. Because that cargo call also re-copies `target/debug/moai-cli`, `.config/nextest.toml` makes the binary run alone (`threads-required = 'num-test-threads'`), so every other test of the machine stops for those minutes. The lint only needs the test names. It reads them from `tests/<X>.rs` and the module files it declares, with cargo's `module::name` paths, and so starts no cargo. The nextest special case then has no reason and is removed. Readers are the people who run `check affected` or the lint, and the auditors who read the lint's output; the output lines keep their form.

## Non-goals

- Changing which documents the lint scans or how it finds a command in them.
- Running the recorded commands, or checking flags other than the target and its filters.
- Reading test names that a `macro_rules!` body makes (no target the documents name uses one today).

## Requirements

- REQ-1: The lint shall read the test names of a target from `tests/<name>.rs` and from the module files that file declares, without starting a cargo process.
- REQ-2: The lint shall give each test name the module path cargo prints, joined with `::`, for inline modules and for module files.
- REQ-3: The lint shall count a function as a test when an attribute whose last path segment is `test` stands before it, with other attributes and comments allowed in between.
- REQ-4: If a target makes tests in a `macro_rules!` body and the filters select no read test, then the lint shall report a violation that says those names cannot be read.
- REQ-5: The file .config/nextest.toml shall have no override for the binary ac_command_lint.

## Acceptance

- AC-1 (REQ-1): the lint run over the repository with a `PATH` whose `cargo` only records that it was started exits 0, the record stays empty, and the run ends in under 5 s.
  verify: `cargo nextest run --locked --offline --test ac_command_lint`
- AC-2 (REQ-2, REQ-3): in a temporary tree, filters that name a test in an inline module, in a module file, under `#[tokio::test]` and after `#[ignore]` with a comment each select it; braces inside strings and comments do not move the module path; a filter that names nothing is a violation that names the file.
  verify: `cargo nextest run --locked --offline --test ac_command_lint`
- AC-3 (REQ-4): in a temporary tree, a filter on a target whose tests come only from a `macro_rules!` body is a violation whose text says the names are made by a macro.
  verify: `cargo nextest run --locked --offline --test ac_command_lint`
- AC-4 (REQ-5): `.config/nextest.toml` has no `binary(ac_command_lint)` override, and the targets the documents name today get the same answer as from `cargo test -- --list`.
  verify: `cargo nextest run --locked --offline --test ac_command_lint`, and the name comparison quoted in the Result

## Scope

- write: `.moai/specs/SPEC-20261122-ac-lint-no-cargo/spec.md`, `tools/check_ac_commands.py`, `tests/ac_command_lint.rs`, `.config/nextest.toml`
- read-only: `tests/*.rs` (as the lint's input), `tools/repo_scans.json`
- forbidden: any product code under `src/`, the documents the lint scans

## Decisions

- D-1: a `#[cfg(...)]` test for another platform is read as a name, while `cargo --list` on this platform leaves it out. A filter on it then passes here although cargo would run 0 tests on this platform. That is accepted: the documents record one command for every platform.
- D-2: module files resolve as rustc does for a crate root in `tests/`: `mod x;` in `tests/a.rs` is `tests/x.rs` or `tests/x/mod.rs`, and inside `tests/x/mod.rs` a `mod y;` is `tests/x/y.rs` or `tests/x/y/mod.rs`. A `#[path = "…"]` attribute before `mod` wins.

## Plan

1. RED: in `tests/ac_command_lint.rs`, replace `nextest_runs_this_binary_alone` with the no-override test, and add the no-cargo run (AC-1), the temporary-tree cases (AC-2) and the macro case (AC-3). They fail on the tip.
2. GREEN: a small Rust-source reader in `tools/check_ac_commands.py` (skips strings, chars, raw strings and comments; tracks `mod name {` braces; follows `mod name;` files) replaces `list_tests`; remove the override and its comment from `.config/nextest.toml`.
3. Compare the read names with `cargo nextest list` for the targets the documents name, quote it in the Result, then one `moai check affected` and HANDOFF READY.

## Risks

- A test name the reader misses would turn a good command into a false violation: the comparison in Plan step 3 checks every target the documents name, and the temporary-tree cases pin the shapes the repository uses.
- Removing the nextest override lets this binary run beside the others again: the reason for it (the cargo call re-copying `target/debug/moai-cli`) is gone with REQ-1.

## Result

Measured 2026-11-22 on macOS in this branch's worktree (in-tree `target/`, `CARGO_BUILD_JOBS=5`).

- AC-1 met: `the_repository_lint_starts_no_cargo_and_ends_within_five_seconds` passes. With a fake `cargo` first on `PATH`, the lint over the repository (334 recorded files, 6 named targets) exits 0 with 0 cargo starts and ends in 1.31 s inside nextest (1.78 s when run directly). The tip's script, run the same way, started cargo 6 times and reported every target as listing no tests.
- AC-2 met: `test_names_are_read_from_the_sources_with_their_module_paths` passes. The 6 good filters select a read test (inline module, nested inline module under `#[tokio::test]`, module file, `#[path]` module, top level, after the modules), and exactly the 4 bad filters are violations (names inside a string and a raw string, a wrong module path twice).
- AC-3 met: `a_filter_on_macro_made_tests_says_the_names_cannot_be_read` passes.
- AC-4 met: `nextest_has_no_special_case_for_this_binary` passes. Name comparison, `cargo nextest list --message-format oneline` against the names read by `list_tests`, for every target the documents name today: archive_privacy_suite 13/13, boundary_output_scope 2/2, e2e_host_fixture 6/6, index_cli 23/23, install_cli 7/7, product_cli 73/73, with no name found on only one side.
- `cargo nextest run --test ac_command_lint`: red on the tip code 8 run, 4 passed, 4 failed (the four new tests); green 8 run, 8 passed, 1.48 s.
