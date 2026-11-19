---
id: SPEC-20261120-hooks-conflict-choices
title: "codex-hooks and omp-hooks conflicts carry keep, diff and replace-with-backup choices"
status: draft
revision: 1
created: 2026-11-20
---

# codex-hooks and omp-hooks conflicts carry keep, diff and replace-with-backup choices

## Intent

Card C-57 criterion 4: a conflict step of `moai project-update-plan` carries the choices `keep`, `diff` and `replace-with-backup`, each with its command. The `codex-config` half is on the tip (SPEC-20261119-codex-conflict-choices: `codex-config-plan` and `codex-config-apply` with `--replace-with-backup`, a backup copy named by its digest, a plan digest that binds the mode). The `codex-hooks` half is not: when `.codex/hooks.json` exists and no receipt matches it, `hooks::codex_state` answers `existing .codex/hooks.json requires preservation review`, `hooks_step` in `src/project_hosts.rs` turns that into a bare `conflict` step with no preview and no choices, and `hooks-plan` and `hooks-apply` have no way through. The person who hand-edited that file is stuck, and the skill has nothing to offer them. `omp-hooks` takes the same shared path (`hooks::plan` with another host name and file), so the same change serves both.

This SPEC adds a `--replace-with-backup` mode to `hooks-plan` and `hooks-apply` for the hosts `codex` and `omp` that mirrors the `codex-config` one, and gives the conflict step its three choices. Everything else about hooks stays as it is: the plain plan still refuses a file it does not own, with the same words and the same digest.

## Non-goals

- Changing the plain `hooks-plan` so that a conflict is a plan JSON instead of an error: the words and the digest of the plain path are pinned by existing tests and by receipts already previewed.
- `hooks-revert` for a replacement: the way back is to copy the backup over the file; a revert that understands replacements is a later SPEC.
- The `claude` host: its settings conflict already has `confirm_replace` and `settings-backups`.
- The AGENTS.md, bundle and CLAUDE.md conflicts, and the workflow sentences in `skills/moai/**` (card C-57 criterion 5 is its own slice after this lands).
- Recording a `keep` pick: `keep` writes nothing, so the next plan shows the same conflict (as `codex-config` does).

## Requirements

- REQ-1: WHEN `hooks-plan --replace-with-backup` is given for the host `codex` or `omp` and the hook file exists with no matching receipt, the plan shall report the action `replace`, the replaced file, the backup path, a `diff` and its own digest without writing anything.
- REQ-2 (REQ-1): The `diff` shall list the lines that differ between the file and the planned file, at most 40 per side, with a line shaped like a secret withheld.
- REQ-3: WHEN `hooks-apply --replace-with-backup` is given the digest of the replace plan, the apply shall write the backup, then the pending receipt, then the hook file, then the final receipt.
- REQ-4 (REQ-3): IF the digest is not the one of the replace plan made now, THEN the apply shall refuse as a stale plan before it writes the backup.
- REQ-5 (REQ-3): IF a backup of the planned name holds other bytes, THEN the apply shall refuse before it writes the hook file.
- REQ-6 (REQ-1): IF the `backups` folder is a link, THEN the replace plan shall refuse and write nothing.
- REQ-7: The backup shall hold the exact bytes the hook file had before the apply, with mode 0600, under the state folder `backups`.
- REQ-8 (REQ-3): IF `hooks-apply` meets a hook file that no receipt matches and is not given `--replace-with-backup`, THEN it shall refuse with a conflict and write nothing.
- REQ-9 (REQ-1): IF `--replace-with-backup` is given and the hook file is absent or owned by its receipt, THEN the plan and the apply shall behave as without the flag and write no backup.
- REQ-10 (REQ-1): IF `--replace-with-backup` is given with `--host claude`, THEN `hooks-plan` and `hooks-apply` shall refuse it and write nothing.
- REQ-11: WHEN `project-update-plan` finds a `codex-hooks` or `omp-hooks` conflict whose replace plan can be made, its step shall carry the choices `keep`, `diff` and `replace` in that order, with `keep` recommended and without a command, `diff` with the read-only replace plan command and `replace` with the apply command of that plan's digest.
- REQ-12 (REQ-11): IF the replace plan cannot be made, THEN the step shall stay a conflict without choices.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-9, REQ-10): the codex and omp replace plans report `replace`, the replaced file, a backup path and a `diff` and leave the tree as it was; an owned file and an absent file plan as without the flag (same JSON and digest, no backup); `claude` is refused; the diff stops at 40 lines and withholds a secret-shaped line · verify: `cargo test --locked --offline --test hooks_conflict_choices_cli plan`
- AC-2 (REQ-3, REQ-4, REQ-5, REQ-7, REQ-8): the apply writes the backup, then the file, then the receipts; the backup holds the original bytes with mode 0600; the plain plan afterwards reads `unchanged`; a digest of a file changed since is refused before any backup exists; an apply without the flag is refused; a backup of the same name with other bytes is refused and one with the same bytes is reused; two replaces in a row keep two backups · verify: `cargo test --locked --offline --test hooks_conflict_choices_cli apply`
- AC-3 (REQ-6): a linked `backups` folder refuses the replace plan and does not block the plain plan of a file that needs no copy · verify: `cargo test --locked --offline --test hooks_conflict_choices_cli linked`
- AC-4 (REQ-11, REQ-12): `codex-hooks` and `omp-hooks` in conflict at once each carry their own `keep`, `diff`, `replace`; the printed commands run as printed and the next plan reads `unchanged`; without a replace plan the step is a bare conflict · verify: `cargo test --locked --offline --test hooks_conflict_choices_cli step`; `cargo test --locked --offline --lib project_hosts::tests::hooks_choices`
- AC-5 (REQ-3): a replace is refused where the plain apply is refused because the `moai` first on `PATH` is the earlier-generation build · verify: `cargo test --locked --offline --test hooks_conflict_choices_cli gate`
- AC-6 (REQ-8, REQ-9): the plain path is unchanged: its words, its digest and the step JSON of a plain plan · verify: `cargo test --locked --offline --lib hooks::`; `cargo test --locked --offline --test hooks_cli`; `cargo test --locked --offline --test hook_legacy_cli`; `cargo test --locked --offline --test update_hooks_cli`; `cargo test --locked --offline --test project_update_hooks_cli`; `cargo test --locked --offline --test project_update_plan_cli`; `cargo test --locked --offline --test codex_conflict_choices_cli`; `cargo test --locked --offline --test ac_command_lint`; `cargo test --locked --offline --test help_layout`; `cargo test --locked --offline --test home_isolation_scan`; `cargo test --locked --offline --test spec_cli`

## Scope

- write: `src/hooks.rs`, `src/hook_legacy.rs`, `src/update_hooks.rs`, `src/project_hosts.rs`, `src/main.rs`, `src/hosts/codex_config.rs`, `tests/hooks_conflict_choices_cli.rs`, `.moai/specs/SPEC-20261120-hooks-conflict-choices/spec.md`
- read-only: `src/claude.rs`, `src/agents.rs`, `src/apply.rs`, `src/project_update.rs`, `src/lib.rs`, `src/contract.rs`, `src/help.rs`, `tests/hooks_cli.rs`, `tests/common/**`
- forbidden: `schemas/**`, `tests/fixtures/help/**`, `docs/hooks-generation.md`, `docs/design-review/appendix/operations/project_update_plan.md`

## Plan

Scope notes: the backup writer in `src/hosts/codex_config.rs` becomes crate-visible; the real HOME, its ledger and the network are never touched; no schema names hooks; the new flags carry doc comments and do not change a help list's first line, so the help fixtures stay as they are; the two documents are written in a later slice after the card takes them.

Two slices in one branch, tests first (RED), then code, then mutants, neighbours and the Result with one line per AC:

1. Slice 1, the core: `hooks::plan_with_replace` and `hooks::apply_with_replace` (the old `plan_with` and `apply_with` stay as wrappers with the flag off, so their callers and inline tests are untouched); `codex_state` learns the replace mode and the backup; `hook_diff`; the flags on `HooksPlan` and `HooksApply`; `update_hooks::apply_direct` carries the flag; `hook_legacy::check` and `gate_plan` treat `replace` like `create` and `update`. REQ-1 to REQ-10, AC-1 to AC-3, AC-5, AC-6.
2. Slice 2, the step: `hooks_step` in `src/project_hosts.rs` asks for the replace plan and builds `keep`, `diff`, `replace` through a new `hooks_conflict_choices`. REQ-11, REQ-12, AC-4.
3. Then the docs (`docs/hooks-generation.md`, `project_update_plan.md`) and the workflow sentences (criterion 5) as their own steps.

## Risks

- A replace plan that did not bind the mode could be applied without the flag: the replace digest adds the mode and the replaced file, the plain digest stays as it was (`BASE_PLAN_DIGEST` in `src/hooks.rs` pins it), and AC-2 applies the replace digest without the flag and expects a refusal.
- `hook_legacy::check` and `gate_plan` pass only `create` and `update`: without `replace` in both, a replace would skip the wrong-generation refusal. AC-5 covers it.
- The pending receipt names the user's bytes by `sha` for a moment, as the create and update path does; an interruption between the backup and the pending receipt leaves the backup and a plan that reads `update`. Nothing is lost.
- The diff prints hook command lines: a line shaped like a secret is withheld (`secret_shaped`), the backup is mode 0600, and the plan prints no more than 40 lines a side. Whether `secret_shaped` fits every hook command is read in the tests, not assumed.
- `src/main.rs` and `src/hooks.rs` are hot files: the card holds them only while this slice is written, and the new tests are in a new file so no other lane's test file changes.
- A replacement has no `hooks-revert`: the person copies the backup back (Non-goals); the step's `reason` and the choices' text say where the backup is.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The flag is `--replace-with-backup` on `hooks-plan` and `hooks-apply`, for `codex` and `omp`; `claude` is refused with a message that says it must be used with `--host codex` or `omp`. Option not taken: allow `claude` (it has `confirm_replace` and `settings-backups`). | The planner read both paths; the `codex-config` flag has the same name, so one word means one thing. | Remove the flag; the plain path never changed. |
| D-2 | The plain `hooks-plan` keeps its error for a file it does not own. Option not taken: turn the conflict into a plan JSON (the words and digest are pinned by `tests/hooks_cli.rs` and by receipts already shown). | Smallest change that adds the way through. | A later SPEC may change it with its own tests. |
| D-3 | A replace is planned only when the flag is given, the file exists and no receipt matches it (no receipt, or a different `sha`); a broken receipt gives the same error as today. An owned file and an absent file ignore the flag: same JSON, same digest, no backup. Option not taken: reuse the `update` action (its digest cannot bind the mode). | Mirrors `codex-config` REQ-8: the flag changes nothing unless there is something to replace. | n/a |
| D-4 | The replace digest is `(1, host, before, receipt, after, "replace-with-backup", [file])` plus the lane when there is one; the plain digest is unchanged. Option not taken: bind the mode always (it would change `BASE_PLAN_DIGEST` and void every digest already previewed). | The plain digest is pinned by an inline test. | n/a |
| D-5 | The `diff` is `{path, current{sha256, bytes}, desired{sha256, bytes}, removed[], added[], truncated}`: lines only in the file and lines only in the plan, at most 40 and 200 characters each, secret-shaped lines withheld. Option not taken: the whole body (a secret could leak) or a per-event JSON diff (an omp TypeScript module is not JSON). | The `codex-config` diff is key-level; a hook file has no keys on omp, so a line set is the common form. | Change the form with a SPEC. |
| D-6 | The backup is `.moai-cli/backups/<host>-hooks-<first 16 hex of the sha>.<json|ts>`, the exact bytes, 0600, exclusive create, reused when the same bytes are already there, refused when other bytes hold the name, and the `backups` folder must not be a link. The writer is the `codex-config` one made `pub(crate)`. Option not taken: a copy under `receipts/` (it would tie the backup to the `hooks-revert` receipt model); fifteen copied lines instead of a shared writer. | One backup rule for both. | Keep the copy and move the writer back. |
| D-7 | The apply order is backup, pending receipt, hook file, final receipt (the existing `codex_receipt_bytes`). The backup comes first so an interruption still leaves the copy. | Same ordering idea as `codex-config`. | n/a |
| D-8 | A file whose bytes already equal the plan but has no receipt is a replacement: a backup of the same bytes is written and the receipt adopts the file. Option not taken: `unchanged` (the person could never leave the conflict). | The receipt is the only owner record. | n/a |
| D-9 | `keep` writes nothing and records nothing, so the next plan still reads `conflict`. `hooks-revert` is not extended. `omp-hooks` is included because it shares the path. Option not taken: record `keep` as a skip (a ledger or profile change). | The leader gave no answer on `omp-hooks`; the stated default (include) is used and written here. | Drop `omp` from D-1's host list. |
| D-10 | When the replace plan cannot be made (a linked `backups` folder, a broken receipt) the step stays a conflict without choices (REQ-12), where the `codex-config` step would still print `keep` and `diff`: for hooks the plain plan is an error, so the diff content comes only from the replace plan and a `diff` command would fail for the same reason. | Printing a command that is known to fail is worse than printing none. | Add `keep` alone later if a person asks for it. |
