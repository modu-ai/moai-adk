---
id: SPEC-20261120-update-workflow-choices
title: "the UPDATE workflow presents conflict choices as a selection with the recommended option first"
status: draft
revision: 1
created: 2026-11-20
---

# the UPDATE workflow presents conflict choices as a selection with the recommended option first

## Intent

Card C-57 criterion 5: the `codex-config` conflict step (on the tip) and the `codex-hooks` and `omp-hooks` conflict steps (SPEC-20261120-hooks-conflict-choices) carry `preview.choices` (`keep`, `diff`, `replace`), but `skills/moai/workflows/update.md` still treats every conflict as "keep it and report it", and step 6 says these steps have no revert. The session therefore never offers the choices. This change adds the sentence that presents them as one selection, `keep` first and recommended, and the note that a replacement leaves a copy.

## Requirements

- REQ-1: The text of step 2 shall tell the session to ask the `preview.choices` of a conflict step as one selection in the printed order with `keep` first and recommended.
- REQ-2: The text of step 2 shall say that `keep` and `diff` write nothing and that `replace` runs only after the user picked it, through its printed command with a permission reference composed from that pick.
- REQ-3: The text of step 6 shall say that a replacement made from a conflict choice has no revert command and that the copy named by the answer's `backup` is the way back.
- REQ-4: The sentence of step 2 shall stay within 1200 bytes and `update.md` within the package file bound.

## Acceptance

- AC-1 (REQ-1, REQ-2): step 2 of `update.md` holds the sentence naming `preview.choices`, the three ids in the printed order, `keep` as recommended with no command, `diff` as read-only, `replace` as a write after the pick and the permission reference · verify: `cargo test --locked --offline --test workflow_update_choices step_two`
- AC-2 (REQ-3): step 6 names the copy and says a replacement has no revert command · verify: `cargo test --locked --offline --test workflow_update_choices step_six`
- AC-3 (REQ-4): the sentence is at most 1200 bytes and the file at most the bound `profile::read` puts on a package file · verify: `cargo test --locked --offline --test workflow_update_choices budget`
- AC-4 (REQ-1, REQ-3): the package manifest holds the new hash of `update.md` and the neighbours still pass · verify: `python3 -m unittest tests.source_manifest_test tests.skill_references_test`; `cargo test --locked --offline --test workflow_size_budget`; `cargo test --locked --offline --test product_cli`

## Scope

- write: `.moai/specs/SPEC-20261120-update-workflow-choices/spec.md`
- write: `skills/moai/workflows/update.md`
- write: `skills/moai/references/source-manifest.json`
- write: `tests/workflow_update_choices.rs`

## Plan

1. Tests first (RED): `tests/workflow_update_choices.rs` reads `update.md` of this checkout (no HOME, nothing is run) and pins the sentence of step 2, the note of step 6 and the byte budget; run it and see the three groups fail.
2. Add the sentence after the conflict sentence of step 2 and the note to step 6; update the manifest hash and the `reason` of `workflows/update.md`.
3. Run the new file, the neighbours (the manifest and reference tests, `workflow_size_budget`, `product_cli`), the checks of `tools/repo_scans.json` and the same-prefix rule (`tests/workflow*.rs`, `tests/skill*.rs`), then the audit and the Result.
4. The text names behaviour of the hook steps that SPEC-20261120-hooks-conflict-choices adds; both go into the same release, and the READY note says so.

## Result

Done. RED `61997f87e0`: all 5 new tests failed, each with "`Conflict choices (card C-57): …` text missing" (the sentence and the note were not in `update.md` yet). GREEN `2f76a69bd5`: the sentence is in step 2 after the conflict rule, the note in step 6, and the `workflows/update.md` entry of `skills/moai/references/source-manifest.json` has the new sha256 and a `reason` that names this change. Mutations on `update.md` (each text change reverted with `git checkout`, tree clean after): M1 `replace` without "only after the user picked it" (`step_two_names_the_choices_and_how_each_one_runs`), M2 step 6 without the `backup` (`step_six_names_the_copy_of_a_replacement`), M3 the sentence padded over 1200 bytes (`budget_holds_for_the_sentence_and_the_file`; my first M3 put the padding after the end marker and so did not test the budget, I redid it inside the sentence and it was killed), M4 `keep` no longer first (`step_two_lists_keep_diff_replace_in_that_order`): all killed.

- AC-1: met — `workflow_update_choices`: `step_two_names_the_choices_and_how_each_one_runs`, `step_two_lists_keep_diff_replace_in_that_order`, `step_two_holds_the_sentence_after_the_conflict_rule` pass (5 passed, 0 failed in the file): the sentence names `preview.choices`, the three ids in the printed order with `keep` recommended, `diff` as read-only, `replace` as a write only after the pick with the permission reference and `command.stdin`; M1 and M4 killed.
- AC-2: met — `step_six_names_the_copy_of_a_replacement` passes: step 6 says a replacement has no revert command and names the copy of the answer's `backup`, and the older rule for steps that print no command is still there; M2 killed.
- AC-3: met — `budget_holds_for_the_sentence_and_the_file` passes: the sentence is under 1200 bytes and `update.md` is 23389 bytes, under the 65536 bound; M3 killed.
- AC-4: met — each file's `test result` line read from the run's log, all `ok` with 0 failed: `workflow_size_budget` 6, `product_cli` 79 (it checks the manifest against the package files), `doc_references` 12, `ac_command_lint` 5, `spec_cli` 18, `help_layout` 10, `--lib` 1577, `--bins` 28 on each of two targets; python `source_manifest_test` and `skill_references_test` 50 OK, `repo_scans_test`, `source_manifest_test` and `local_check_test` 116 OK, `skill_references_test` 5 OK; the same-prefix rule (`workflow*`, `skill*`, `update_*`, `source_*`, `project_skill*`: 28 targets, 254 tests, one cargo run) 0 failed; all 39 checks of `tools/repo_scans.json` exit 0.

Not done: a run of the installed skill in a host (the text is read by tests only; host use is `not_run` until the installed skill carries it), and Windows (nothing here is platform specific). The text names `codex-hooks` and `omp-hooks` choices that SPEC-20261120-hooks-conflict-choices adds; that change is already on the local integration branch.
