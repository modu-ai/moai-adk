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
3. The pending list also holds what another card reserves (`bundle.rs`, `gate.rs`, `go_state.rs`, `home_layout.rs`, `optimize.rs`, `project_update.rs`, `session.rs`), what a fixture pins (`hosts/codex.rs`, `hooks.rs`), and the generated texts a later change handles together with the parsers that read them (`claude.rs`, `lane_env.rs`, `gate_merge.rs`, `claim_timeout.rs`, `flake_record.rs`, `code_refs.rs`).
4. Run the new file, the library tests, the tests that read command output, the scan checks of `tools/repo_scans.json` and the same-prefix rule, then audit and write the Result.
