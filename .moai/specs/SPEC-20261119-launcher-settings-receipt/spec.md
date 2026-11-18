---
id: SPEC-20261119-launcher-settings-receipt
title: "update: a receipt and a revert for the launcher settings write (C-47 slice 4)"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261118-install-model-per-tool
---

# update: a receipt and a revert for the launcher settings write (C-47 slice 4)

## Intent

Slice 4 of card C-47 (the audit of PR 3212 named the missing way back as note N1; leader decision note 2226 moved it here). The model stage of `moai update` writes the user launcher settings file after a yes, and nothing can undo that write. This slice makes the write leave a receipt in the receipts folder of the state folder, like the other home writes, and adds a revert function that puts the file back byte for byte while it is still the file the write produced.
The receipt keeps the text of the file before the write, so the folder is owner-only like every receipt folder. The revert command line (`src/main.rs`, `src/help.rs`) waits for the card that holds those files; this slice ships the library function with its tests and the receipt id in the stage result.

## Non-goals

- The command that calls the revert from a shell or a session: the main file and the help file are held by card C-62; the next piece of this card adds the command and the help text when they are free.
- A way back for a write that changed nothing: a run that saves the values already saved writes no file and leaves no receipt.
- The launcher reading the values (launcher slice) and the install script (install slice).
- Reverting a file that changed after the write: the revert names it and leaves it alone.

## Requirements

- REQ-1: When the model stage saves after a yes, the stage shall write a receipt before the file is written, named by the word models, the unix milliseconds and the process id joined by dashes, in the receipts folder of the state folder, holding the operation `update.models`, the settings path, the text of the file before the write or null when the file was absent, the sha256 of the file before and of the file the write will produce, and the permission reference, so that the result of the stage can name the receipt id. · failure: none — REQ-2 and REQ-3 cover a receipt that cannot be kept.
- REQ-2 (REQ-1): IF the receipt would be larger than 60000 bytes because the file before the write is large, THEN the stage shall write nothing, keep no receipt and report the state `error` naming the limit.
- REQ-3 (REQ-1): IF the receipt cannot be written, THEN the stage shall write no settings and report the state `error`.
- REQ-4 (REQ-1): IF the file read back after the write does not equal the text the receipt promised, THEN the stage shall report the state `error` with the receipt id and say that the file must be checked.
- REQ-5: When the revert is called with a permission reference and a receipt id and the settings file equals the file the receipt names as written, the revert shall put the text before the write back byte for byte, with the file mode kept, or remove the file when it was absent, read the file back, write the marker named by the receipt id and the suffix .reverted and report the state `reverted`. · failure: none — REQ-6 and REQ-7.
- REQ-6 (REQ-5): IF the settings file is not the file the receipt names as written, THEN the revert shall change nothing, write no marker and report the state `skipped` with the reason `changed_since_apply`.
- REQ-7 (REQ-5): IF the permission reference is empty or a placeholder, the id is not a models receipt id, the receipt is missing or is not an `update.models` receipt, it names another settings file than the current one, or it was already reverted, THEN the revert shall change nothing and refuse with an error.
- REQ-8: When the stage or the revert reports, the result shall hold ids, paths, states and reasons only and never the text of the file or of the receipt. · failure: none — a pinned output test finds a leak.

## Acceptance

- AC-1 (REQ-1, REQ-8): after a yes the result carries the receipt id, the receipt file holds the fields of REQ-1 with the real before text and hashes, its mode is 0600 in a folder of mode 0700, and neither the result nor the printed lines contain the file text.
  verify: `cargo test --locked --offline --test launcher_settings_receipt a_save_leaves_a_receipt_with_the_text_before_and_no_text_in_the_result`
- AC-2 (REQ-1, REQ-5): a save into an absent file records the file as absent, and the revert removes the file the save created.
  verify: `cargo test --locked --offline --test launcher_settings_receipt reverting_a_save_into_an_absent_file_removes_it`
- AC-3 (REQ-5): the revert puts back the earlier text byte for byte with its mode (0640) and an unusual layout, reads it back and writes the marker.
  verify: `cargo test --locked --offline --test launcher_settings_receipt the_revert_restores_the_earlier_bytes_and_mode_and_marks_the_receipt`
- AC-4 (REQ-6): a file changed after the save is left as it is, no marker is written, and a later revert after the change is undone works.
  verify: `cargo test --locked --offline --test launcher_settings_receipt a_file_changed_after_the_save_is_not_reverted`
- AC-5 (REQ-7): an empty reference, a bad id, a missing receipt, a receipt of another operation, another settings path and a second revert are each refused and change nothing.
  verify: `cargo test --locked --offline --test launcher_settings_receipt the_revert_refuses_what_it_cannot_prove_is_its_own`
- AC-6 (REQ-2, REQ-3): a file whose receipt would pass 60000 bytes and a receipts folder that cannot be made each stop the stage before any write, with no receipt.
  verify: `cargo test --locked --offline --test launcher_settings_receipt no_receipt_means_no_write`
- AC-7 (REQ-1): a run that changes nothing writes no file and no receipt, and a no writes no receipt.
  verify: `cargo test --locked --offline --test launcher_settings_receipt a_no_and_an_unchanged_run_leave_no_receipt`
- AC-8 (REQ-5): a revert whose write-back fails (the settings folder cannot be written) reports an error and leaves the file and the receipt unmarked.
  verify: `cargo test --locked --offline --test launcher_settings_receipt a_failed_restore_changes_nothing_and_stays_revertable`
- AC-9 (REQ-1, REQ-5): the existing stage and question tests still pass with the receipt, and the pseudo terminal flow that saves reports the receipt.
  verify: `cargo test --locked --offline --test update_model_pick`; `cargo test --locked --offline --test model_effort_questions`; `python3 -m unittest tests.update_tux_tty_test`
- AC-10 (REQ-1): the update workflow text names the receipt, and the source manifest carries its new hash.
  verify: `cargo test --locked --offline --test product_cli source_manifest_covers_every_package_file`
- AC-11 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8): the code style checks.
  verify: `cargo fmt --check`

## Scope

- write: `.moai/specs/SPEC-20261119-launcher-settings-receipt/spec.md`, `src/update_models.rs`, `src/model_effort_questions.rs`, `tests/launcher_settings_receipt.rs`, `tests/update_model_pick.rs`, `tests/model_effort_questions.rs`, `skills/moai/workflows/update.md`, `skills/moai/references/source-manifest.json`
- read-only: `src/key_vault.rs`, `src/update_keys.rs`, `src/contract.rs`, `src/home_root.rs`, `src/update.rs`
- forbidden: `src/main.rs`, `src/help.rs`, `src/session.rs`, the real HOME, a live tool session, the network

## Plan

1. Tests first (RED): `tests/launcher_settings_receipt.rs` with the eight cases, which fail to compile until the revert function exists.
2. The question module gets the shared reader, the text writer, the rendered text of the file after a plan and a remover; the stage writes the receipt before the apply; the revert function is added to the stage module.
3. The update workflow text and the manifest hash; the checks of the earlier test files.
4. Later: the command line for the revert when the main file is free.

## Risks

- The receipt holds a copy of the user settings file, which a person may have edited by hand → the receipts folder is owner-only (0700, files 0600, never replaced), the receipt is limited to 60000 bytes, and nothing prints the text (REQ-8).
- The receipt is written before the settings, so a failed write leaves a receipt that was never applied → the revert finds the file equal to the file before the write and reports `skipped` with `changed_since_apply`, never a false revert; the unused receipt is harmless.
- Without a command line the revert cannot be run from a shell yet → the receipt id is in the result and the function is tested; the command follows when the main file is free, and this slice does not claim a working undo for a person.
- A write by another program between the receipt and the settings → the digest-bound apply refuses a changed launcher object, and REQ-4 catches any other difference after the write.
- The limits of the settings write recorded by the audits of PRs 3207, 3209 and 3212 stay as listed in their SPEC files.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The receipt is written before the settings file | a failed receipt must never leave a write without a way back | write it after and roll back on failure |
| D-2 | The receipt keeps the before text instead of per entry values | a replaced entry outside the support table may hold any text, and only the whole text restores byte for byte | keep per entry values for supported entries only |
| D-3 | The receipt id is the word models, the unix milliseconds and the process id joined by dashes, and the marker is the id with the suffix .reverted, like the keys receipt | one receipt shape across the home writes | rename the prefix |
| D-4 | A revert that finds a changed file does not write the marker | the person may undo their change and revert later | write the marker on every call |
| D-5 | The command line waits for the main file | card C-62 holds `src/main.rs` and `src/help.rs` | add the command in the next piece |

## Result

- AC-1: pending
