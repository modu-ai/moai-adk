---
id: SPEC-20261121-codex-launcher-and-models-removed
title: "The launcher and the model settings stop supporting Codex; a Codex value saved earlier is ignored on read and dropped on the next save"
status: draft
revision: 1
created: 2026-11-21
---

# The launcher and the model settings stop supporting Codex; a Codex value saved earlier is ignored on read and dropped on the next save

## Intent

The user decided on 2026-11-21 that moai-adk supports Claude Code and GLM only (GLM runs through the Claude Code launcher): Codex support is withdrawn. This change is the first part of the removal, for the launcher and the model settings. The second part, the home update, has its own SPEC and starts after this one is audited.

Three things stand in the way of a plain deletion, and the design follows them:

- The model settings refuse a whole `launcher.yaml` that names a tool they do not know. An earlier `moai update` offered Codex as a default candidate, so a saved `launcher.models.codex.*` is likely in real files. Removing `codex` from the tool list alone would make every Claude and GLM role launch fail on such a file. The saved Codex value must be ignored on read and left out of the file the next save writes (leader decision D2; moai deletes no user file for this change, it only stops writing the value).
- `moai codex` stops launching and says in one line that it was removed (the same shape as the removed `moai init`), not clap's generic unknown-subcommand text with a suggestion of the internal `codex-*` commands. The word list of the installed switch script (`install/moai-switch.sh`) names `codex`; that script is pinned by a hash and belongs to another change, so it stays as it is.
- Codex stays a host elsewhere (hooks, project files, doctor): those belong to other changes. This change cuts only the launcher entry, the Codex command line, the Codex rows of the model table and question, and the Codex entry of the update's models step.

## Non-goals

- The home update (`~/.codex`, `~/.agents/skills`): the next SPEC.
- Hooks, project `.codex` files, the Codex doctor and projection commands, `src/hosts/*`: other changes.
- OMP: it stays in the table and the launcher (decision recorded on the card; a later change removes it).
- Deleting a user's `launcher.yaml` value or any file under the user's Codex folder.

## Requirements

- REQ-1: The launcher shall offer the tools `cc`, `glm` and `omp` and no `codex` subcommand, so that `moai codex` prints the one-line removal notice `moai codex was removed: MoAI supports Claude Code and GLM only (moai cc | moai glm)` in English and Korean and exits 2.
- REQ-2: When the host word after `-a` is `codex`, the launcher shall refuse it with the usage error that lists `cc`, `glm` and `omp`. · failure: none — the refusal is the required behavior
- REQ-3: The launcher shall build no Codex command line, so that a session request names no Codex program and the Codex argument notes are gone.
- REQ-4: The model settings shall hold a table and defaults for `claude`, `omp` and `glm` only.
- REQ-5: When the saved launcher settings hold a `launcher.models.codex` value of any shape, the model settings shall ignore it on read, so that a Claude or GLM choice still resolves and the entry is not listed as unsupported. · failure: none — an ignored value is not an error
- REQ-6: When a choice is written for another tool, the model settings shall leave a saved `launcher.models.codex` value out of the written file, as the retired-tool rule of SPEC-20261121-retired-hosts-read-safely does (leader decision D2 of 2026-11-21). · failure: none — leaving a key out cannot fail
- REQ-7: The model question shall offer `claude` and `glm` as its tools and no Codex row.
- REQ-8: The models step of `moai update` shall look for the folders of `claude` only.
- REQ-9: The first-run profile and the missing-skill check of the launcher shall know no `codex` host.
- REQ-10: The launcher help and its goldens shall list `cc`, `glm` and `omp` and no `codex` word.
- REQ-11: The model settings shall refuse a choice written for the tool `codex`.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): `moai codex` (alone or with flags) exits 2, prints the removal sentence once on stderr with no clap text and no suggestion of a `codex-*` command, and the usage envelope on stdout carries the same sentence; `moai -a codex` and `moai session --host codex` are refused exactly like an unknown word (same exit code, same message shape) and the message lists `cc|glm|omp`; the bare-run usage line and the `moai init` guidance name no `codex`; a session request holds no Codex program name · verify: `cargo test --test codex_launcher_removed`
- AC-2 (REQ-4, REQ-5, REQ-6, REQ-11): `supported("codex")` and `default_choice("codex", …)` are `None`, `TOOLS` has no `codex`, a `with_choice` for `codex` is refused; `resolve` for `claude` and `glm` works with a `launcher.models.codex` value of every shape (a valid old entry, a malformed one, a secret-looking text) and `unsupported_entries` lists none of it; writing a Claude choice leaves the Codex value out of the written file and keeps every other key · verify: `cargo test --test model_effort_settings`
- AC-3 (REQ-7, REQ-8): the question tools are `claude` and `glm`; the installed tools of a home with `.claude` and `.codex` folders are `claude` only; a saved Codex value does not make the model step report an error · verify: `cargo test --test model_effort_questions --test update_model_pick`
- AC-4 (REQ-9): the first-run profile of `moai cc` and `moai glm` names its own host only and the missing-skill check refuses a `codex` word like any unknown host · verify: `cargo test --test launcher_auto_init --test launcher_skill_check`
- AC-5 (REQ-10): the root and the full help print the launcher words `cc, glm, omp` and the goldens match · verify: `cargo test --test help_layout`
- AC-6 (REQ-3): every other launcher test keeps its Claude, GLM and OMP cases green with the Codex cases removed · verify: `cargo test --test launcher_agent_entry --test launcher_agent_numbering --test launcher_reads_models --test launcher_roleless --test launcher_advisor_setting --test launcher_permissions_setting --test launcher_terminal_output --test omp_launcher --test output_names_moai --test cli_unknown_command`

## Scope

- write: `src/session.rs`, `src/main.rs`, `src/launcher_init.rs`, `src/launcher_terminal.rs`, `src/model_effort_settings.rs`, `src/model_effort_questions.rs`, `src/update_models.rs`, `src/help.rs` (the launcher lines only), the tests named in the acceptance lines, `tests/product_cli.rs` (its Codex launcher cases), `tests/launcher_codex_effort.rs` (deleted), the eight help goldens under `tests/fixtures/help/`, the recorded usage error `tests/fixtures/protocol/v1/current/01-usage-unknown-subcommand.json`, `src/ui_text.rs` (the two `InitRemoved` texts and their test only, by a line-range agreement with the holder), a new `tests/codex_launcher_removed.rs`, and this SPEC folder
- files held by other changes at the time of writing (asked for back by pair lines): `src/help.rs`, `src/update_models.rs`, `tests/product_cli.rs`, `tests/update_model_pick.rs`, `tests/launcher_codex_effort.rs`
- read-only: `src/hooks.rs`, `src/hosts/*`, `src/ops/codex_*`, `src/go_state.rs`, `src/lane_loop.rs`, `src/host_home.rs`, `install/moai-switch.sh`
- forbidden: the Go repository, the real HOME, any file under the user's Codex or agents folders in the home, any card number in a file name, SPEC id, branch or commit

## Plan

Tests first (RED): the new `tests/codex_launcher_removed.rs` runs the built binary for the launcher words and calls the model functions for the table; the existing launcher and model tests lose their Codex cases and gain the new expectations (a saved Codex value is ignored on read and dropped on the next save). On the tip these fail for the right reasons (`moai codex` still launches, `codex` is still in the table).

Order of the change, as the research of the area gave it: first the read safety in `src/model_effort_settings.rs` (ignored tool names are skipped in the structure check, the entry check and the unsupported list; `with_choice` keeps every other key as it does now and leaves retired tools out; the shared retired-tool rule landed first in SPEC-20261121-retired-hosts-read-safely), then the table, the defaults, the question and the update's tool list; then the launcher (`src/session.rs` Codex argv, notes and request field, `src/main.rs` command, host word, usage lines and request literals, `src/launcher_terminal.rs` warning list length, `src/launcher_init.rs`); last the help lines and goldens when `src/help.rs` is free.

`host_home::Host::Codex` stays (hooks and doctor use it). `install/moai-switch.sh` stays (hash pinned).

## Risks

- A saved Codex value reaches the structure check, the entry check and the unsupported list: each needs its own test, or a leftover path refuses a Claude launch (the trap the research file lists first).
- `src/main.rs` and `src/help.rs` are shared with the hooks, lane and doctor changes: only the launcher lines are cut.
- Removing the Codex notes shortens `WARNINGS` in `src/launcher_terminal.rs`: the array length changes with them.
- The goldens of the help are shared with other changes: a regenerated golden must show only the launcher lines changed.

## Result

Built tests first, one slice at a time: the red commit held the launcher, model-table and question cases (recorded with `moai tdd red`, `failure_class: test-failure`), the models-step slice has its own red and green commit, and the help words followed once `src/help.rs` was free. The branch then took the local tip, which already held the retired-host layer (SPEC-20261121-retired-hosts-read-safely): the three conflict spots of `src/model_effort_settings.rs` were resolved to its `retired(tool, active)` rule and this change's own `IGNORED_TOOLS` list is gone. That layer drops a retired key on the next save (leader decision D2), so REQ-6, AC-2 and the title of this SPEC were changed from "kept" to "left out of the written file", with the cases of `model_effort_settings`, `model_effort_questions` and `update_model_pick` following.

- AC-1: met — `codex_launcher_removed` 5 passed (`moai codex` and its flags print the one-line removal notice on stderr with no clap text and no `codex-*` suggestion, the usage envelope carries the same sentence, exit 2, no stub host started; `-a codex` and `session --host codex` differ from an unknown word only by the word; the bare usage line and a session request hold no codex); `ui_text::` unit tests 11 passed (the `moai init` guidance names `moai cc` and `moai glm` and no codex; the `moai codex` notice names `moai cc`, `moai glm`, Claude Code and GLM in English and Korean).
- AC-2: met — `model_effort_settings` 11 passed (table, defaults, refusal of a choice for `codex`, a saved value of five shapes read without error and not listed, the next write leaves it out and keeps every other key).
- AC-3: met — `model_effort_questions` 16 and `update_model_pick` 11 passed (a home with `.claude` and `.codex` offers `claude` only, the command report holds six rows and none for codex, a saved Codex value of three shapes gives no error, is not quoted and is dropped by the next save).
- AC-4: met — `launcher_auto_init` 18 and `launcher_skill_check` 4 passed.
- AC-5: met — `help_layout` 10 passed after the goldens were written with `MOAI_HELP_GOLDEN_WRITE` and checked by a second run; `cargo test --lib help::` 18 passed before the tip merge. The `codex-*` commands keep their lines until their own removal.
- AC-6: met — `launcher_agent_entry` 6, `launcher_agent_numbering` 25, `launcher_reads_models` 10, `launcher_roleless` 3, `launcher_advisor_setting` 7, `launcher_permissions_setting` 6, `launcher_terminal_output` 11, `omp_launcher` 17, `output_names_moai` 8, `cli_unknown_command` 4 passed; also `launcher_settings_receipt` 17, `models_revert_cli` 5, `update_always_asks` 11 and `product_cli` 79 (after its Codex argv case was replaced by the Claude prompt-note case and the recorded usage error was updated).

Fail-to-pass, measured: the three launcher targets `cli_unknown_command`, `launcher_agent_entry` and `omp_launcher` and `update_model_pick` failed before their green commits (tdd red records and the first runs), and passed after.

Check-affected chain (base `a840fa32fa`, head `e78275d388`, the tip that holds the retired-host layer): `diff_check`, `fmt`, `clippy` (`--all-targets -D warnings`), `test_lib` (1546 passed), `test_bins` (28 + 28 passed), `test_targets` (81 targets, none failed), `build_cli` and `python` (185 tests OK) all exit 0. The first two runs of the chain failed on `fmt` (three assert lines) and on `clippy` (four `for` loops over one element left by the removal of the codex host in `tests/omp_launcher.rs`, `tests/codex_launcher_removed.rs`, `tests/launcher_agent_numbering.rs`, `tests/launcher_advisor_setting.rs` and `tests/launcher_skill_check.rs`); both were fixed in test text only and the whole chain was run again on the final head.

Mutants (5, one rule each, run on the committed text, each restored with `git checkout` and the tree read clean): `installed_tools` lists `.codex` again → `update_model_pick` 1 failed; `supported()` has a codex row again → `model_effort_settings` 1 failed; `host_of` accepts the word `codex` → `codex_launcher_removed` 1 failed; the `moai init` guidance names `moai codex` again → `ui_text::` 1 failed; the retired-key skip dropped from `check_structure` → `model_effort_settings` 2 failed. All five are killed.

Not run: Windows and Linux, mutants of the help words and the launcher notes (the goldens and `help_layout` hold them), and the `codex-*` command help lines (left by design).

Repair after the audit (one blocking item, one wording item): `moai codex` printed clap's generic unknown-subcommand text and suggested the internal `codex-*` commands; it now prints the one-line removal notice (`Msg::CodexRemoved`, English and Korean, handled in `usage_error` like `moai init`), with the `codex_launcher_removed` case rewritten first (red commit, `test-failure`) and `ui_text::` pinning both texts. The wording of this SPEC (intent, REQ-1, AC-1) and of `tests/model_effort_questions.rs` now says a saved Codex value leaves the file on the next save (leader decision D2). Check-affected chain over the whole branch on the repaired head (base `a840fa32fa`, head `2523a8dcbe`): targets 81, python 8, scans listed 41, `diff_check` 0, `fmt` 0, `clippy` 0, `test_lib` 0, `test_bins` 0, `test_targets` 0, `build_cli` 0, `python` 0, result PASS. Mutants of the repair, both killed: the notice branch never taken → `codex_launcher_removed` 1 failed; the Korean text without GLM → `ui_text::` 1 failed. Not run: Windows and Linux.
