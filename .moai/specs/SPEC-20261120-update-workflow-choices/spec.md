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
