---
id: SPEC-20261122-bin-test-false
title: "The moai binary target is not a test target, so the main.rs tests are built and run once"
status: draft
revision: 1
created: 2026-11-22
---

# The moai binary target is not a test target, so the main.rs tests are built and run once

## Intent

The user-approved redesign (`~/.moai/reports/moai-adk/agentic-redesign-2026-11-22.md` §3B) and its two reviews (build-review §5.6, test-review F7) measured this: `src/bin/moai.rs` loads `src/main.rs` as a module (`#[path = "../main.rs"] mod cli;`), so `cargo test --bins` (the `test_bins` step of `moai check affected`) compiles, links and runs the same 27–28 `main.rs` tests twice — once in the `moai-cli` harness and once in the `moai` harness. Each harness is a 63 MiB test binary, and the duplicate costs 5–45 s per check plus 63 MiB per worktree. `test = false` on the `[[bin]] name = "moai"` target stops cargo from building that second harness. The `moai-cli` harness still runs every `main.rs` test once. The `moai` binary itself is still built (the CLI tests run it), because `test = false` drops only the test harness. Measured on the tip with this change (2026-11-22): `cargo test --no-run` builds one harness (`unittests src/main.rs`), but `cargo test --bins --no-run` still builds both, because an explicit `--bins` selects every binary whatever its `test` key says. `check affected` passes `--bins`, so its step must name `--bin moai-cli` for the saving to reach the check (REQ-4). Readers are the people who run `moai check affected` and the auditors who read its receipt. No user-facing behaviour or document changes; the `test_bins` log shows one `running N tests` block instead of two.

## Non-goals

- Making `moai-cli` a link to `moai`, or moving `main.rs` into the library (build-review §5.6 "next"): those are crate-structure cards.
- Changing `check affected` step names or targets (SPEC-20261122-check-affected-one-slot owns that).
- Any change to the release or dist build, or to what `cargo build` produces.

## Requirements

- REQ-1: The file Cargo.toml shall set test to false on the binary target named moai.
- REQ-2: The binary target named moai-cli shall stay a test target that runs the tests of src/main.rs.
- REQ-3: The binary named moai shall still be built by cargo build and by the integration test builds.
- REQ-4: The test step of moai check affected shall select the binary named moai-cli by name instead of selecting every binary.

## Acceptance

- AC-1 (REQ-1, REQ-2): `Cargo.toml` has `test = false` in the `[[bin]]` table whose name is `moai`, and the `moai-cli` table has no `test = false`.
  verify: `cargo nextest run --locked --offline --test bin_test_once`
- AC-2 (REQ-4): the `test_bins` step of one `moai check affected` run on the change's head builds and runs one bin test harness: its log has one `running N tests` block for `src/main.rs`, not two.
  verify: the check-affected log `06-test_bins.log` of the READY receipt, quoted in the Result
- AC-3 (REQ-3): the integration tests that run the `moai` binary still find it (`CARGO_BIN_EXE_moai`), and the check-affected `build_cli` and `test_targets` steps pass.
  verify: the same check-affected receipt (`build_cli=0`, `test_targets=0`)

## Scope

- write: `.moai/specs/SPEC-20261122-bin-test-false/spec.md`, `Cargo.toml` (one line in the `moai` `[[bin]]` table), `tests/bin_test_once.rs` (new)
- read-only: `src/bin/moai.rs`, `src/main.rs`, `src/check_affected.rs` (REQ-4 is written there only by the leader's decision D-0)
- forbidden: any other `Cargo.toml` table, the release and dist profiles, shared build folders

## Decisions

- D-0: the REQ-4 argument lives in `src/check_affected.rs`, which the check-affected one-slot change (SPEC-20261122-check-affected-one-slot) reserves and rewrites into one nextest call. Who writes it is a leader decision (LANE-BLOCKED on the card); the default hands it to that rewrite.
- D-1: the guard test parses `Cargo.toml` with `toml_edit` (already a dependency) and reads the `[[bin]]` array by name, so a table moved or reordered still matches.

## Plan

1. RED: `tests/bin_test_once.rs` asserts AC-1. It fails on the tip, because the `moai` table has no `test` key.
2. GREEN: add `test = false` under `path = "src/bin/moai.rs"`.
3. Run nextest on `bin_test_once`, then `moai check affected` once on the head, and quote the `test_bins` log for AC-2.

## Risks

- `Cargo.toml` is edited by many cards: the change is one line inside the `moai` table, so a merge with another card's edit there is a text conflict at worst.
- If a future change adds tests to `src/bin/moai.rs` itself, they would not run. The file is a 3-line shell, and the guard test's message says so.
