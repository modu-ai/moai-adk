---
id: SPEC-20261118-home-install-go-retire
title: "home-install after the Go retire: movable preview with retire evidence"
status: draft
revision: 1
created: 2026-11-18
parent: SPEC-20261117-sdd-foundation
---

# home-install after the Go retire: movable preview with retire evidence

## Intent

Card C-50 (observed on a real HOME 2026-11-16, moai-cli 2026.11.0-rc.2+local.2): `home-install-plan` is blocked before the Go retire
(`go_items_not_retired`) and again after it (`generation_unknown`), so `~/.moai-cli` never moves to `~/.moai` (user decision D-M1, #2871).
The retire takes every Go marker away, the kept items (unknown owner, user data) stay, and `classify_home` calls such a `~/.moai`
`unknown`. `ledger-register-plan` on the same HOME reads `register`, so the two previews disagree on whether the HOME is ready.
This SPEC lets the install preview offer the move when a finished retire is on record, keeps today's refusal for an unknown
`~/.moai` with no record, and makes both previews give one answer.

## Non-goals

- Retiring the leftover `~/.moai-cli` folder itself (#2871 U6, planned `home-cli-retire-*`): a later SPEC.
- Changing `classify_home`'s verdicts or the marker schema.
- Removing or rewriting `~/.moai/logs/hook-stderr.log`: its writer is named here, the Go generation's hooks stay as they are.

## Requirements

- REQ-1: When the state folder is classed unknown and the retire folder holds a manifest of a finished go-retire for this HOME, the home-install preview shall list the moves as it does for an empty HOME and answer install with a confirmation token.
- REQ-2 (REQ-1): IF the state folder is classed unknown and no such manifest is there (none, one of another HOME, a link, a file that is not a manifest), THEN the home-install preview shall answer blocked with generation_unknown and a help line that names what the user can do.
- REQ-3 (REQ-1): When the install is applied, the apply shall move the vault, versions and path items and write the schema-2 marker last. · failure: none — the existing apply tests cover a stale plan, a clash and a stopped move.
- REQ-4 (REQ-3): When the install has been applied, the key status, the launcher path and the home root shall resolve under the new state folder. · failure: none — a read of the earlier folder is the old behavior and AC-3 pins it.
- REQ-5 (REQ-4): When a command runs after the install, the command shall write no new entry into the earlier state folder. · failure: none — AC-3 compares the listing before and after.
- REQ-6: When both previews run on the same HOME, the ledger-register preview and the home-install preview shall give one answer. · failure: none — a mismatch fails AC-4.
- REQ-7 (REQ-3): When the install is reverted, the revert shall restore the earlier tree. · failure: none — the existing revert tests cover a changed marker and a moved link.
- REQ-8: When the hook-stderr log is asked about, the SPEC shall name its writer with evidence. · failure: none — a document statement, pinned by AC-5.

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-7): in a temporary HOME with Go state and kept unknown items, after `home-retire-apply` the install preview is `install` with a token; apply moves the Rust items and writes the marker last; revert restores the tree.
  verify: `cargo test --locked --offline --test home_install_go_retire a_retired_go_home_gets_a_movable_install_preview_and_the_move_reverts`
- AC-2 (REQ-2): an unknown state folder with no retire manifest, a manifest of another HOME, a manifest that is a link and a manifest of another kind and a retire whose journal is pending, reverted, missing or has an item not moved are all still blocked with generation_unknown, a help line, and unchanged bytes under the temporary HOME.
  verify: `cargo test --locked --offline --test home_install_go_retire an_unknown_home_without_a_retire_record_is_still_refused`; `cargo test --locked --offline --test home_install_go_retire a_retire_that_did_not_finish_does_not_unlock_the_install`
- AC-3 (REQ-4, REQ-5): after the apply, key-status reads the vault from the new state folder, the home root is .moai, the path stub points into its versions folder, and the listing of the earlier state folder is the same before and after those commands.
  verify: `cargo test --locked --offline --test home_install_go_retire after_the_install_readers_resolve_under_moai_and_moai_cli_stays_untouched`
- AC-4 (REQ-6): for the same temporary HOME the two previews agree in the three states (retire record, no record, applied); a blocked registration digest never authorizes the registration after the HOME turns ready; an existing ledger keeps its idempotent registration.
  verify: `cargo test --locked --offline --test home_install_go_retire ledger_register_and_install_agree_on_home_readiness`; `cargo test --locked --offline --test home_install_go_retire a_blocked_registration_digest_cannot_authorize_after_the_home_turns_ready`; `cargo test --locked --offline --test home_install_go_retire an_existing_ledger_keeps_its_idempotent_registration`
- AC-5 (REQ-8): the writer is the Go generation's hook wrapper `.claude/hooks/moai/handle-*.sh` (`MOAI_HOOK_STDERR_LOG=…/.moai/logs/hook-stderr.log`), started by the Go project's `.claude/settings.json`; no Rust source or skill file names the log.
  manual: run `grep -rl hook-stderr src skills` in the repository (it prints nothing and exits 1) and read line 3 of any `handle-*.sh` of the Go repository (no test walks the tree: `tools/repo_scans.json` would need an entry)
- AC-6 (REQ-1…REQ-8): the neighbours.
  verify: `cargo test --locked --offline --test home_install_plan_cli`; `cargo test --locked --offline --test home_install_apply_cli`; `cargo test --locked --offline --test home_retire_cli`; `cargo test --locked --offline --test ledger_cli`; `cargo test --locked --offline --test ledger_open`; `cargo test --locked --offline --bin moai`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/home_install.rs`, `src/home_retire.rs`, `src/ops/ledger_register_plan.rs`, `src/ops/ledger_register_apply.rs`, `tests/home_install_go_retire.rs`, `.moai/specs/SPEC-20261118-home-install-go-retire/spec.md`
- read-only: `src/state_root.rs`, `src/home_layout.rs`, `src/home_root.rs`, `src/move_journal.rs`, `tests/home_install_plan_cli.rs`, `tests/home_install_apply_cli.rs`
- forbidden: the real HOME and its three state folders, the hooks of the Go repository, the network

## Plan

1. Tests first (RED): the four tests above in `tests/home_install_go_retire.rs`, each in a temporary HOME.
2. `home_retire::finished_retire(home)`: newest manifest of kind `go-retire` for this HOME, read by `lstat` (no link followed, size-bounded).
3. `home_install::plan`: `unknown` with a finished retire is movable; the record's stamp joins the content digest; `generation_unknown` gets a `help` line.
4. `home_install::registration_blocked`: the two ledger-register commands call it (the preview answers blocked, the apply refuses before writing); the ledger library and opening an existing ledger keep `check_home_ready`. The card write scope has no `src/ledger/**`, so the check lives in the command wrappers.

## Risks

- A fresh HOME whose `~/.moai` holds only a Go hook's `logs/` folder is `unknown` with no record, so ledger registration is now `blocked` there (it registered before) → the `help` line names the way out (retire or move the entries); D-2 records the choice.
- A forged manifest could unlock the preview → the apply still re-plans, moves only table-listed Rust items, refuses a clash (`destination_taken`) and writes the marker last; a manifest proves a retire happened, it grants nothing else.
- `unknown` plus a record is also what a stopped install looks like after its moves → the install journal in `~/.moai-cli/home-install/` decides a resume first; the existing stop-and-resume tests run unchanged (AC-6).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Evidence is the retire manifest (`retire-manifest-<stamp>.json`, kind `go-retire`, `home` equal to this HOME), not a changed verdict in `classify_home` | the card names the manifest as the evidence; eight verdicts stay | add a verdict later |
| D-2 | Ledger registration is blocked on `unknown` with no record, like the install | the card wants one answer; the install refusal for an unknown `~/.moai` is kept (card AC-2) | drop the `registration_blocked` calls |
| D-3 | Opening an existing ledger keeps `check_home_ready` | a registered ledger must keep working whatever `~/.moai` holds | none needed |
| D-5 | A retire counts as finished only when its journal is committed and every item is moved, not on the manifest alone | the manifest is written before the first rename (Codex review P2 on #3199) | accept the manifest alone |
| D-4 | `hook-stderr.log` is named, not changed | the writer is the Go repository's hooks, outside this card's scope | a card in the Go hook set |

## Result

- AC-1: met — `a_retired_go_home_gets_a_movable_install_preview_and_the_move_reverts` passes: after a real retire in a temporary HOME the preview is install with a token, the apply moves the Rust items and writes the marker last, the revert restores them
- AC-2: met — `an_unknown_home_without_a_retire_record_is_still_refused` (five variants) and `a_retire_that_did_not_finish_does_not_unlock_the_install` (pending, reverted, item not moved, journal missing) pass with unchanged bytes
- AC-3: met — `after_the_install_readers_resolve_under_moai_and_moai_cli_stays_untouched` passes: key-status, home root and path stub resolve under the new folder, the earlier folder listing is unchanged
- AC-4: met — `ledger_register_and_install_agree_on_home_readiness`, `a_blocked_registration_digest_cannot_authorize_after_the_home_turns_ready` and `an_existing_ledger_keeps_its_idempotent_registration` pass
- AC-5: met — writer is the Go hook wrapper handle-*.sh (`MOAI_HOOK_STDERR_LOG` line 3); `grep -rl hook-stderr src skills` prints nothing and exits 1
- AC-6: met — fmt, clippy -D warnings, the neighbour test files, repo_scans_test, ci_scope_test and the CI run of the PR head pass
