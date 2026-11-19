---
id: SPEC-20261119-models-revert-command
title: "models-revert: the command that puts the launcher settings back (C-47 slice 5)"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261119-launcher-settings-receipt
---

# models-revert: the command that puts the launcher settings back (C-47 slice 5)

## Intent

PR 3215 gave the model stage of `moai update` a receipt and a library revert, and its audit left two follow-ups (leader decision note 2400): a command a person can run, and a unit test for a receipt that holds only `{"before_sha256": null}`. This slice adds the command `models-revert HOME --receipt ID --permission-reference REF`, one file under `src/ops/` like `keys-revert`, with its help entry, its document fragment and a test that runs the real binary in a temporary HOME. It also pins the half-null receipts in a unit test.
The settings write and the revert themselves do not change. The skill text `skills/moai/workflows/update.md` still says there is no undo command; the manifest that carries its hash is held back for five other cards, so that sentence changes in a later piece.

## Non-goals

- Changing `skills/moai/workflows/update.md` or `skills/moai/references/source-manifest.json` (leader decision note 2400: the manifest waits for other cards).
- A hint line about the command in the terminal screens of `moai update`.
- The receipts folder check (audit N2 of PR 3215).
- The launcher reading the values (criteria 5 and 7 of the card, launcher slice).

## Requirements

- REQ-1: When `models-revert` runs with a home folder, a receipt id and a permission reference, the command shall run the receipt revert for that home and print its result in the standard envelope under the operation name `models.revert`. · failure: none — REQ-2 covers a refusal.
- REQ-2: IF the revert refuses because the reference, the id, the receipt or its settings path is not acceptable, THEN the command shall print the error envelope with the reason and change nothing.
- REQ-3: When the settings file changed after the save, the command shall answer ok with the state `skipped` and the reason `changed_since_apply`, write no marker and leave the file as it is. · failure: none — REQ-2 covers a refusal.
- REQ-4: When the help is printed, the command list and the per-command examples shall name `models-revert`. · failure: none — a missing entry fails the help fixtures.
- REQ-5: IF a receipt holds only one of its two before fields, or a null with a value, THEN the receipt reader shall treat it as damaged.
- REQ-6: When the command prints, the output shall hold ids, paths, states and reasons only and never the text of the settings file. · failure: none — a pinned output test finds a leak.

## Acceptance

- AC-1 (REQ-1, REQ-6): after a save, `models-revert` puts the earlier file back byte for byte, prints `ok`, `operation` `models.revert` and the state `reverted`, and the output holds no settings text.
  verify: `cargo test --locked --offline --test models_revert_cli models_revert_puts_the_earlier_file_back`
- AC-2 (REQ-2): an empty or placeholder reference, an id of another shape and an unknown receipt each exit with an error envelope and leave the file as it was.
  verify: `cargo test --locked --offline --test models_revert_cli models_revert_refuses_what_it_cannot_prove_is_its_own`
- AC-3 (REQ-3): a file changed after the save answers `skipped` with `changed_since_apply`, writes no marker and stays as it is.
  verify: `cargo test --locked --offline --test models_revert_cli models_revert_skips_a_file_changed_after_the_save`
- AC-4 (REQ-4): the full command list and the examples name `models-revert`, and the help fixtures and one fragment under `docs/design-review/appendix/operations/` carry it.
  verify: `cargo test --locked --offline --test help_layout`; `cargo test --locked --offline --test models_revert_cli models_revert_is_in_the_help`
- AC-5 (REQ-5): a receipt reader test holds `{"before_sha256": null}` alone, `{"before_text": null}` alone, and the other half-null pairs, and each is damaged.
  verify: `cargo test --locked --offline --lib update_models::`
- AC-6 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6): the neighbour checks and the style checks pass.
  verify: `cargo test --locked --offline --test launcher_settings_receipt`; `cargo test --locked --offline --test hot_file_ops_merge`; `cargo fmt --check`

## Scope

- write: `.moai/specs/SPEC-20261119-models-revert-command/spec.md`, `src/ops/models_revert.rs`, `src/update_models.rs`, `src/help.rs`, `tests/fixtures/help/all.txt`, `tests/fixtures/help/all-60.txt`, `tests/fixtures/help/all-80.txt`, `tests/fixtures/help/all-dumb.txt`, `docs/design-review/appendix/operations/models_revert.md`, `tests/models_revert_cli.rs`, `tests/launcher_settings_receipt.rs`
- read-only: `src/ops/keys_revert.rs`, `src/update_keys.rs`, `src/key_vault.rs`, `src/home_root.rs`
- forbidden: `src/main.rs`, `skills/moai/references/source-manifest.json`, `skills/moai/workflows/update.md`, the real HOME, a live tool session, the network

## Plan

1. Tests first (RED): `tests/models_revert_cli.rs` runs the binary, which has no such command yet; the half-null unit test runs against the reader.
2. `src/ops/models_revert.rs` (the registry collects it), the help group and example, the four help fixtures, the document fragment.
3. Neighbour checks, mutants, Result.

## Risks

- The help fixtures are generated text; a wrong group or wording fails `help_layout` → regenerate them with the fixture switch and read the diff.
- The skill text still says there is no undo command → named in the Intent and the pull request; the sentence changes when the manifest is free.
- The command reads a receipt a person could edit by hand → the revert already refuses damaged receipts, a path that is not the settings path of the given home, and a file that no longer hashes to what the save wrote.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The command is one file under `src/ops/` and does not touch `src/main.rs` | `keys-revert` and the registry rule (#2617) | add a variant to the main file |
| D-2 | The command takes the home folder as its first argument, like `keys-revert` | one shape for the revert commands | default to the current home |
| D-3 | Name `models-revert`, the receipt prefix is `models-` | the receipt id and the command say the same word | rename both |
| D-4 | The skill text and manifest wait | leader decision note 2400 | edit them in a later piece |

## Result

- AC-1: met — `models_revert_puts_the_earlier_file_back` passes (operation `models.revert`, state `reverted`, byte for byte, marker written, no settings text printed)
- AC-2: met — `models_revert_refuses_what_it_cannot_prove_is_its_own` passes (empty and placeholder reference, id of another shape, unknown receipt; each with its pinned reason, file unchanged)
- AC-3: met — `models_revert_skips_a_file_changed_after_the_save` passes (`skipped`, `changed_since_apply`, no marker)
- AC-4: met — `help_layout` 10 passed after the four golden files gained only the `models-revert` line (`MOAI_HELP_GOLDEN_WRITE`, diff read: 5 added lines); `models_revert_is_in_the_help` passes; `--bin moai` 28 passed (example commands exist in the command tree)
- AC-5: met — `--lib update_models::` passes with `{"before_sha256": null}` added to the damaged receipts
- AC-6: met — `launcher_settings_receipt` 11, `hot_file_ops_merge` 3, `update_keys_cli` 9, `update_model_pick` 10, `update_tux_golden` 14 passed, `cargo fmt --check` 0, `cargo clippy --lib --tests -- -D warnings` 0

RED before the command existed: `models_revert_cli` failed 4 of 4 (`unrecognized subcommand 'models-revert'`; the refusal test failed on the pinned reasons).
Mutation 6 mutants, 5 killed, 1 survived: N1 operation name, N2 reference dropped, N3 receipt swapped, N5 listing removed, N6 hash-only half-null accepted all KILLED; N4 (the `EXAMPLES` row removed) survives because no per-command help or golden prints that row, so nothing can observe it.
Not run: the full local `cargo test`, Python tty tests (terminal screens unchanged), Linux, Windows.
