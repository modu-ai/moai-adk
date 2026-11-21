---
id: SPEC-20261121-go-residue-detect
title: "A read-only go-residue step finds the files the Go generation left behind, proven against the manifests it can still find"
status: draft
revision: 1
created: 2026-11-21
---

# A read-only go-residue step finds the files the Go generation left behind, proven against the manifests it can still find

## Intent

The Go-generation update survey (2026-11-21, mo.ai.kr measured) found the root gap G1/G2: the update recognises the Go generation only by three marker files and `.moai/manifest.json`, so once go-state has moved the state (or the markers were removed by hand) the Go files left behind — on mo.ai.kr 222 tracked files byte-identical to the backup manifest — are invisible, and the plan answers "nothing to move". The survey's slice S1 adds the detection half: a read-only `go-residue` step of `moai project-update-plan` that finds those files, proves them against a Go manifest it can still find, and classifies them. Applying and reverting are the next slice (S2), on its own card.

## Non-goals

- Applying, moving or reverting anything: the step writes nothing; the apply and revert commands are S2.
- The three Go markers and `state_root` classification: unchanged; a `mixed` verdict still blocks go-state as today (S3 owns the resurrected-marker case).
- The `claude-md` step, the leftovers bookkeeping, `all_unchanged` semantics and the managed ignore block: untouched (S4, S5 own them).
- Codex: withdrawn by decision, so no Codex step or text.
- The `settings_referenced`, `tracked_reports` and `unmanaged_host` classes: the slices that own them (the claude-md step, the leftovers and reports bookkeeping, the unmanaged-host step) extend this survey; this slice adds none of them.

## Requirements

- REQ-1: The project-update plan shall include a read-only `go-residue` step that surveys the project for files the Go generation left behind and writes nothing.
- REQ-2: The go-residue step shall prove a Go-made file against the first manifest it finds in the order `.moai/manifest.json`, then a `manifest.json` under any `.moai-backup*` folder, then a `manifest.json` under `.moai/backups/go-markers-*`, then the backup manifest that the last committed go-state receipt names.
- REQ-3: The go-residue step shall sort every file it surveys into `go_same` (byte-identical to a `deployed_hash` of the found manifest), `go_changed` (named by the manifest with different bytes), `go_marker` (a marker file back after a committed move), `user_data` (in no manifest) and `leftover_folder` (`.moai-cli/`, a `.moai.rust-<UTC>` folder, `.moai-backups/`).
- REQ-4: Each residue item shall carry its path, its sha256, whether git tracks it, whether a committed go-state receipt says it moved and it is back, and the suggested action.
- REQ-5: The go-residue step shall count secret-shaped names without opening their files and never follow a link.
- REQ-6: When no manifest proves any file, the go-residue step shall report no proof with zero candidates and the plan still succeeds. · failure: none — the no-manifest case is this requirement's own subject and leaves nothing to retry
- REQ-7: A second run of the plan shall read the go-residue step unchanged with the same items and the same digests.

## Acceptance

- AC-1 (REQ-1, REQ-7): on the survey's temp frame the plan's `go-residue` step reports `go_same` for the byte-identical files, `go_changed` for the changed one, `user_data` for the unmanifested one, counts the one secret-shaped name without opening it, marks the tracked files `tracked: true`, and a second plan run reads the same · verify: `cargo test --test project_update_go_residue_cli`
- AC-2 (REQ-2): with `.moai/manifest.json` absent the step still proves the same files from the backup manifest and the go-markers manifest · verify: `cargo test --test project_update_go_residue_cli`
- AC-3 (REQ-6): with no manifest anywhere the step reports no proof and zero candidates while the plan exits 0 · verify: `cargo test --test project_update_go_residue_cli`
- AC-4 (REQ-3, REQ-4, REQ-5): every item carries path, sha256, tracked, resurrected and a suggested action, a symlink is never followed, and the snapshots around the plan are identical · verify: `cargo test --test project_update_go_residue_cli`

## Scope

- write: `src/go_residue.rs` (new), `src/project_update.rs` (the step), `tests/project_update_go_residue_cli.rs` (new), `.moai/specs/SPEC-20261121-go-residue-detect/**`
- read-only: `src/go_state.rs`, `src/state_root.rs`, the survey report sections 4-8
- forbidden: any card number in shipped text, real HOME or real project writes, apply or revert behaviour

## Plan

The temp frame follows the survey's section 8: a git repository with a valid Rust marker and a profile, a backup manifest with `deployed_hash` values, a resurrected policies file and a workflow file byte-identical (`go_same`), a changed config section (`go_changed`), an unmanifested design file (`user_data`), one secret-shaped name, an output style a settings file references, and a `.moai-backup` manifest chain with the live `.moai/manifest.json` removed for AC-2. RED: `tests/project_update_go_residue_cli.rs` fails on a tip without the step. GREEN: the detection module plus the plan step.

## Risks

- A `deployed_hash` collision across manifests: the first-found manifest wins and the item names which manifest proved it.
- Large projects: the survey walks the state folders and the manifest-named paths only, never the whole tree, and stops at the same depth limits go-state uses.
- The step must stay read-only even on error: the test asserts identical snapshots after a failing run.
