---
id: SPEC-20261120-cli-text-citations
title: "user-visible command text carries no card number, Issue number or decision citation"
status: draft
revision: 1
created: 2026-11-20
---

# user-visible command text carries no card number, Issue number or decision citation

## Intent

The text the commands print (help text, previews, hints, error and note strings) quotes moai-adk's own history: Issue numbers, card numbers and "user decision" citations. A user of MoAI cannot follow those. This second change (after SPEC-20261120-skill-text-citations for shipped skill text) adds a scan over the non-test, non-comment code of `src/` with a list of files that still carry such text that can only shrink, cleans the strings no other card holds or a fixture pins, and keeps the removed citations in a developer document.

## Requirements

- REQ-1: The non-comment code of every source file except those on the pending list shall carry no card number, no Issue or pull-request number, no citation of a user or leader decision, and no name of a moai-adk development script.
- REQ-2: The pending list shall name only source files that still carry such text, so a file that was cleaned leaves the list in the same change.
- REQ-3: A cleaned string shall keep everything it tells the user, losing only a citation or a sentence that tells where a rule came from.
- REQ-4: The developer document `docs/cli-text-rationale.md` shall keep each removed citation under the path of the file it came from.
- REQ-5: A change to command text shall carry the new wording of every test that pinned the old text, in the same commit.

## Acceptance

- AC-1 (REQ-1, REQ-2): the scan passes, flags a card number, an Issue number, a pull-request number, a decision citation and a development script name in sample code, ignores a comment line, a trailing comment, the inline test module, a colour code such as `#8839EF` and the sample ids `C-1` to `C-9`, and fails for a source file outside the pending list and for a pending file that has gone clean · verify: `cargo test --locked --offline --test cli_text_citations scan`
- AC-2 (REQ-3, REQ-4): `docs/cli-text-rationale.md` has one section per cleaned file, titled with its path, and each cleaned file has no citation left · verify: `cargo test --locked --offline --test cli_text_citations rationale`
- AC-3 (REQ-5): the commands whose text changed still print what they printed, apart from the citations · verify: `cargo test --locked --offline --lib`; `cargo test --locked --offline --test help_layout --test product_cli`; `python3 -m unittest tests.repo_scans_test tests.ci_scope_test`

## Scope

- write: `.moai/specs/SPEC-20261120-cli-text-citations/spec.md`
- write: `tests/cli_text_citations.rs`
- write: `docs/cli-text-rationale.md`
- write: `tools/ci_scope.py`
- write: `tools/repo_scans.json`
- write: `src/help.rs`
- write: `tests/claim_cli.rs`
- write: `tests/install_candidates.rs`
- write: `tests/codex_adapter.rs`
- write: `src/hosts/codex.rs`
- write: `src/claim.rs`
- write: `src/claim_verify.rs`
- write: `src/gate_schedule.rs`
- write: `src/handoff.rs`
- write: `src/heavy_slot.rs`
- write: `src/home_state.rs`
- write: `src/launcher_init.rs`
- write: `src/ledger/open.rs`
- write: `src/project_hosts.rs`
- write: `src/recovery.rs`
- write: `src/shim_check.rs`
- write: `src/worktree.rs`
- write: `src/worktree_gc.rs`
- write: `src/worktree_include.rs`
- write: `src/worktree_report.rs`

## Plan

1. Tests first (RED): `tests/cli_text_citations.rs` reads the `src/` files of this checkout (no HOME, nothing is run). It holds the scan, a pending list of the files that still carry citations after this change, and the check of the rationale document; run it and see the files to clean fail.
2. Clean the strings file by file, keeping what each tells the user and writing the removed citations into `docs/cli-text-rationale.md`; fix the tests that pinned the old wording in the same commit; register the walking test in `tools/repo_scans.json` and deny the document in `tools/ci_scope.py`.
3. The pending list also holds what another card reserves (`bundle.rs`, `gate.rs`, `go_state.rs`, `home_layout.rs`, `optimize.rs`, `project_update.rs`, `session.rs`), what a fixture pins (`hooks.rs`), and the generated texts a later change handles together with the parsers that read them (`claude.rs`, `lane_env.rs`, `gate_merge.rs`, `claim_timeout.rs`, `flake_record.rs`, `code_refs.rs`).
4. Run the new file, the library tests, the tests that read command output, the scan checks of `tools/repo_scans.json` and the same-prefix rule, then audit and write the Result.

## Result

Done for the printed text of 17 source files; the rest of the command text stays for later pieces (below). RED `4df7536a64` (3 of 5 new tests failed: 16 files still carried citations, the rationale document was missing), GREEN `c0f5249264` (16 files cleaned, the document, the scan registered in `tools/repo_scans.json`, the document denied in `tools/ci_scope.py`), `4ba61950f3` (three tests that pinned a removed citation: the inline help test, the inline `shim_check` test, `tests/claim_cli.rs`). The independent audit then found two blocking problems and fixed them in `ac98be18a6`: `tests/install_candidates.rs` still looked for the old shim note ending (my name-prefix list did not include `install_`), and six Korean help sentences had lost a space or a full stop because Rust drops the blanks after a line-ending backslash. The same commit hardened the scan: a `#[cfg(test)]` on a function or constant no longer ends the scanned code; only a test module does (four files were scanned only in part before: `agents.rs`, `index.rs`, `ui_summary.rs`, `update.rs`; the RED sample failed first, they carry no citation). `fcf7c82707` cleaned `src/hosts/codex.rs` too (the audit showed its pin was a test, not a fixture). I also compared every string literal of the changed files before and after with the Rust line-continuation rule applied: the only differences beyond a removed citation are the sample ids of the `bundle` examples, the words "파일 이동 단계", the launcher permission reference, "a later step" in two `home_state` previews, the worktree report reason, a full stop in the shim note and one test line.

- AC-1: met — `cli_text_citations` 5 passed on the final head `fcf7c82707`: the scan flags a card number, an Issue number, a pull-request number, a decision citation (with a date, an id or a record number) and a development script name in sample code; ignores a comment line, a trailing comment, a block comment, the test module, colour codes such as `#8839EF`, the sample ids `C-1` to `C-9` and `--issue 2904`; reads a `//` inside a string, a raw string and a character literal correctly; fails for a source file outside `PENDING` and for a pending file that went clean. The auditor's own count with a separate regular expression gave the same 15 files as the `PENDING` list at the time (now 14 after the Codex file left it); its extra hits were two false positives of its own.
- AC-2: met — `rationale_has_a_section_per_cleaned_file` and `rationale_cleaned_files_carry_no_citation` pass: 17 cleaned files have a section titled with their path in `docs/cli-text-rationale.md`; the auditor compared the removed citations of the diff with the document one to one (nothing missing, nothing extra).
- AC-3: met — each `test result` read on the final head: `--lib` 1604 passed 0 failed, `--bins` 56 passed, the five light tests 42 passed, 85 integration targets whose first word matches a changed source file (`claim`, `gate`, `handoff`, `heavy`, `help`, `home`, `launcher`, `ledger`, `project`, `recovery`, `shim`, `worktree`) plus `lane_rules_2879`, `agent_defs_package`, `mode_workflow_steps`, `bundle_body`, `go_state` and `update_route` 1440 passed 0 failed; python `repo_scans_test`, `source_manifest_test`, `local_check_test` and `ci_scope_test` 164 OK, the python tests of the changed areas (`home_entry_count_test`, `heavy_check_slot_test`, `help_english_tty_test`) 51 OK, `skill_references_test` 5 OK; all 40 checks of `tools/repo_scans.json` exit 0; `cargo fmt --check`, `rustfmt --check --edition 2021` over `src/ops`, `cargo clippy --all-targets -D warnings` and `git diff --check` exit 0. One timing flake in an earlier run: `home_retire_cli::retire_keeps_the_go_markers_until_the_entries_in_use_are_free` failed once while the machine was loaded and passed alone (71 s) and as a whole target (24 passed); it is not related to the changed text. The auditor's mutations (a citation put back into a Korean help string and into a string of `shim_check.rs`, a renamed rationale heading) were caught; one comment mutation stayed clean as intended.

Not done: the English `--help` text, which clap builds from `///` doc comments in `src/main.rs` and other command structs, still carries citations in about 65 commands (the scan removes comments before it looks, so it does not see them): the next piece of this work; `docs/rust-product-path.md` line 1053 still names the old launcher permission reference (held by another card); the 14 files of `PENDING` (generated texts that parsers read, and files another card reserves: `bundle.rs`, `gate.rs`, `go_state.rs`, `home_layout.rs`, `optimize.rs`, `project_update.rs`, `session.rs`); the internal id `(06-B)` after "operator approval" in `src/claim.rs`, which `tests/claim_cli.rs` pins. Not run: Windows, Linux, CI, a Korean terminal rendering of the help (the Korean table is reached only through `configure_in(..., Lang::Ko)`, since the language of the help is English today), and no full local test suite.
