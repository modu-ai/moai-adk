---
id: SPEC-20261121-codex-launcher-and-models-removed
title: "The launcher and the model settings stop supporting Codex; a Codex value saved earlier is ignored on read and kept in the file"
status: draft
revision: 1
created: 2026-11-21
---

# The launcher and the model settings stop supporting Codex; a Codex value saved earlier is ignored on read and kept in the file

## Intent

The user decided on 2026-11-21 that moai-adk supports Claude Code and GLM only (GLM runs through the Claude Code launcher): Codex support is withdrawn. This change is the first part of the removal, for the launcher and the model settings. The second part, the home update, has its own SPEC and starts after this one is audited.

Three things stand in the way of a plain deletion, and the design follows them:

- The model settings refuse a whole `launcher.yaml` that names a tool they do not know. An earlier `moai update` offered Codex as a default candidate, so a saved `launcher.models.codex.*` is likely in real files. Removing `codex` from the tool list alone would make every Claude and GLM role launch fail on such a file. The saved Codex value must be ignored on read and left in the file (moai never deletes a user file for this change).
- `moai codex` becomes an unknown command. The word list of the installed switch script (`install/moai-switch.sh`) names `codex`; that script is pinned by a hash and belongs to another change, so it stays as it is.
- Codex stays a host elsewhere (hooks, project files, doctor): those belong to other changes. This change cuts only the launcher entry, the Codex command line, the Codex rows of the model table and question, and the Codex entry of the update's models step.

## Non-goals

- The home update (`~/.codex`, `~/.agents/skills`): the next SPEC.
- Hooks, project `.codex` files, the Codex doctor and projection commands, `src/hosts/*`: other changes.
- OMP: it stays in the table and the launcher (decision recorded on the card; a later change removes it).
- Deleting a user's `launcher.yaml` value or any file under the user's Codex folder.

## Requirements

- REQ-1: The launcher shall offer the tools `cc`, `glm` and `omp` and no `codex` subcommand, so that `moai codex` is an unknown command.
- REQ-2: When the host word after `-a` is `codex`, the launcher shall refuse it with the usage error that lists `cc`, `glm` and `omp`. · failure: none — the refusal is the required behavior
- REQ-3: The launcher shall build no Codex command line, so that a session request names no Codex program and the Codex argument notes are gone.
- REQ-4: The model settings shall hold a table and defaults for `claude`, `omp` and `glm` only.
- REQ-5: When the saved launcher settings hold a `launcher.models.codex` value of any shape, the model settings shall ignore it on read, so that a Claude or GLM choice still resolves and the entry is not listed as unsupported. · failure: none — an ignored value is not an error
- REQ-6: When a choice is written for another tool, the model settings shall keep a saved `launcher.models.codex` value as it was. · failure: none — keeping a value cannot fail
- REQ-7: The model question shall offer `claude` and `glm` as its tools and no Codex row.
- REQ-8: The models step of `moai update` shall look for the folders of `claude` only.
- REQ-9: The first-run profile and the missing-skill check of the launcher shall know no `codex` host.
- REQ-10: The launcher help and its goldens shall list `cc`, `glm` and `omp` and no `codex` word.
- REQ-11: The model settings shall refuse a choice written for the tool `codex`.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): `moai codex` exits 2 with the usage envelope `unknown command; see --help`; `moai -a codex` and `moai session --host codex` are refused exactly like an unknown word (same exit code, same message shape) and the message lists `cc|glm|omp`; the bare-run usage line and the `moai init` guidance name no `codex`; a session request holds no Codex program name · verify: `cargo test --test codex_launcher_removed`
- AC-2 (REQ-4, REQ-5, REQ-6, REQ-11): `supported("codex")` and `default_choice("codex", …)` are `None`, `TOOLS` has no `codex`, a `with_choice` for `codex` is refused; `resolve` for `claude` and `glm` works with a `launcher.models.codex` value of every shape (a valid old entry, a malformed one, a secret-looking text) and `unsupported_entries` lists none of it; writing a Claude choice leaves the Codex value byte for byte · verify: `cargo test --test model_effort_settings`
- AC-3 (REQ-7, REQ-8): the question tools are `claude` and `glm`; the installed tools of a home with `.claude` and `.codex` folders are `claude` only; a saved Codex value does not make the model step report an error · verify: `cargo test --test model_effort_questions --test update_model_pick`
- AC-4 (REQ-9): the first-run profile of `moai cc` and `moai glm` names its own host only and the missing-skill check refuses a `codex` word like any unknown host · verify: `cargo test --test launcher_auto_init --test launcher_skill_check`
- AC-5 (REQ-10): the root and the full help print the launcher words `cc, glm, omp` and the goldens match · verify: `cargo test --test help_layout`
- AC-6 (REQ-3): every other launcher test keeps its Claude, GLM and OMP cases green with the Codex cases removed · verify: `cargo test --test launcher_agent_entry --test launcher_agent_numbering --test launcher_reads_models --test launcher_roleless --test launcher_advisor_setting --test launcher_permissions_setting --test launcher_terminal_output --test omp_launcher --test output_names_moai --test cli_unknown_command`

## Scope

- write: `src/session.rs`, `src/main.rs`, `src/launcher_init.rs`, `src/launcher_terminal.rs`, `src/model_effort_settings.rs`, `src/model_effort_questions.rs`, `src/update_models.rs`, `src/help.rs` (the launcher lines only), the tests named in the acceptance lines, `tests/product_cli.rs` (its Codex launcher cases), `tests/launcher_codex_effort.rs` (deleted), the eight help goldens under `tests/fixtures/help/`, a new `tests/codex_launcher_removed.rs`, and this SPEC folder
- files held by other changes at the time of writing (asked for back by pair lines): `src/help.rs`, `src/update_models.rs`, `tests/product_cli.rs`, `tests/update_model_pick.rs`, `tests/launcher_codex_effort.rs`
- read-only: `src/hooks.rs`, `src/hosts/*`, `src/ops/codex_*`, `src/go_state.rs`, `src/lane_loop.rs`, `src/host_home.rs`, `install/moai-switch.sh`
- forbidden: the Go repository, the real HOME, any file under the user's Codex or agents folders in the home, any card number in a file name, SPEC id, branch or commit

## Plan

Tests first (RED): the new `tests/codex_launcher_removed.rs` runs the built binary for the launcher words and calls the model functions for the table; the existing launcher and model tests lose their Codex cases and gain the new expectations (a saved Codex value is ignored and kept). On the tip these fail for the right reasons (`moai codex` still launches, `codex` is still in the table).

Order of the change, as the research of the area gave it: first the read safety in `src/model_effort_settings.rs` (ignored tool names are skipped in the structure check, the entry check and the unsupported list; `with_choice` keeps every other key as it does now), then the table, the defaults, the question and the update's tool list; then the launcher (`src/session.rs` Codex argv, notes and request field, `src/main.rs` command, host word, usage lines and request literals, `src/launcher_terminal.rs` warning list length, `src/launcher_init.rs`); last the help lines and goldens when `src/help.rs` is free.

`host_home::Host::Codex` stays (hooks and doctor use it). `install/moai-switch.sh` stays (hash pinned).

## Risks

- A saved Codex value reaches the structure check, the entry check and the unsupported list: each needs its own test, or a leftover path refuses a Claude launch (the trap the research file lists first).
- `src/main.rs` and `src/help.rs` are shared with the hooks, lane and doctor changes: only the launcher lines are cut.
- Removing the Codex notes shortens `WARNINGS` in `src/launcher_terminal.rs`: the array length changes with them.
- The goldens of the help are shared with other changes: a regenerated golden must show only the launcher lines changed.

## Result

(not yet written)
