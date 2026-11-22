---
id: SPEC-20261122-remove-omp-launcher
title: "The launcher and the model settings stop supporting OMP; moai omp prints a removal notice and an OMP value saved earlier is ignored"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261121-remove-codex-support
---

# The launcher and the model settings stop supporting OMP; moai omp prints a removal notice and an OMP value saved earlier is ignored

## Intent

The user decided on 2026-11-21 that moai-adk supports Claude Code and GLM only. SPEC-20261121-codex-launcher-and-models-removed took the Codex launcher out and kept OMP for a later change; this change takes the OMP launcher, the `omp` host word, the OMP model settings and the Codex and OMP words of the install texts out, in the same shape as the Codex removal. Files a person keeps under the OMP home folder are never read, changed or deleted.

## Non-goals

- The OMP project layer, the hook adapters and generation: SPEC-20261121-remove-codex-project-layer (another branch).
- The OMP rendering of the base agent files, the lane duty and lane loop lines, the shipped skill and workflow texts and `docs/host-support-matrix.md`: other slices of the parent SPEC.
- New Python tests. The existing install script test is edited where it names the removed hosts.

## Requirements

- REQ-1: When a person runs `moai omp` with any arguments, the launcher shall print the one-line notice `moai omp was removed: MoAI supports Claude Code and GLM only (moai cc | moai glm)` on standard error and in the usage envelope and exit 2 without starting a host. · failure: none — the notice is the required behavior
- REQ-2: When a person runs `moai help omp` or `moai help codex`, the command line shall print the output of `moai omp` or `moai codex` with its exit code. · failure: none — the notice is the required behavior
- REQ-3: When the host word after `-a` or `session --host` is `omp`, the launcher shall refuse it like an unknown word with a message that lists `cc` and `glm` only. · failure: none — the refusal is the required behavior
- REQ-4: The launcher shall build no OMP command line, so that a `cc` or `glm` session names no OMP program, no OMP key variable and no `/skill:moai` spelling.
- REQ-5: The model settings shall hold a table and defaults and accepted choices for `claude` and `glm` only.
- REQ-6: When the saved launcher settings hold a `launcher.models.omp` value of any shape, the launcher shall start a `cc` or `glm` session with the argv and environment it builds without that value. · failure: none — an ignored value is the required behavior
- REQ-7: When `moai update` reports or saves model choices, the models step shall leave a saved `launcher.models.omp` value out of the report and out of the written file. · failure: none — the dropped value is the required behavior
- REQ-8: The help goldens and the recorded usage error and the install scripts and the install README shall name no `omp` launcher and no `moai codex`.
- REQ-9: The launcher shall leave every file under the OMP home folder unread and unchanged.

## Acceptance

- AC-1 (REQ-1, REQ-9): `moai omp`, `moai omp -a` and `moai omp -l --model x` each print the notice once on standard error and in the usage envelope, exit 2, start nothing and leave a temporary OMP home folder byte-identical · verify: `cargo test --locked --offline --test omp_launcher_removed moai_omp_says_it_was_removed`
- AC-2 (REQ-2): `moai help omp` and `moai help codex` print the same standard error and exit code as `moai omp` and `moai codex` · verify: `cargo test --locked --offline --test omp_launcher_removed help_omp_and_help_codex_print_the_removal_notice`
- AC-3 (REQ-3): `moai -a omp` and `moai session --host omp` are refused with a message that lists `cc|glm` · verify: `cargo test --locked --offline --test omp_launcher_removed the_omp_host_word_is_refused_like_any_unknown_host`
- AC-4 (REQ-4): the `cc -l`, `cc -a`, `glm -l` and `glm -a` dry runs name no `omp`, no `ZAI_API_KEY` and no `/skill:moai` in their command, environment or notes · verify: `cargo test --locked --offline --test omp_launcher_removed cc_and_glm_launches_are_unchanged_and_name_no_omp`
- AC-5 (REQ-5): the model settings list `claude` and `glm` and refuse an `omp` pick · verify: `cargo test --locked --offline --test model_effort_settings omp_is_not_a_tool_any_more`
- AC-6 (REQ-6): a saved `launcher.models.omp` value as a string, a table and a number leaves the `cc` and `glm` dry runs unchanged · verify: `cargo test --locked --offline --test omp_launcher_removed a_saved_omp_value_of_any_shape_never_changes_a_cc_or_glm_launch`
- AC-7 (REQ-7): the models step report lists no `omp` row, and a save for `claude` writes a file without `launcher.models.omp` · verify: `cargo test --locked --offline --test update_model_pick a_saved_omp_value_is_not_reported_and_leaves_the_file_on_the_next_save`
- AC-8 (REQ-8): the eight help goldens, the recorded usage error fixture, `install/install.sh`, `install/install.ps1` and `install/README.md` name no `moai omp` and no `moai codex` (named files only, no directory walk) · verify: `cargo test --locked --offline --test omp_launcher_removed no_help_usage_or_install_text_names_omp_or_moai_codex`

## Scope

- write (own reservation): this SPEC folder, `src/session.rs`, `src/launcher_terminal.rs`, `src/model_effort_settings.rs`, `src/help.rs`, `install/install.sh`, `install/install.ps1`, `install/README.md`, `tests/omp_launcher.rs` (deleted)
- write (to add to the reservation): `src/ui_text.rs`, `src/model_effort_questions.rs`, the new `tests/omp_launcher_removed.rs`, and the tests that name removed things: `tests/codex_launcher_removed.rs`, `tests/launcher_agent_entry.rs`, `tests/launcher_terminal_output.rs`, `tests/launcher_reads_models.rs`, `tests/output_names_moai.rs`, `tests/help_layout.rs`, `tests/model_effort_settings.rs`, `tests/model_effort_questions.rs`, `tests/update_model_pick.rs`, `tests/install_script_test.py`, `tests/fixtures/help/root*.txt`
- write (held by other cards; only on the leader's word): `src/main.rs` (C-321), `src/launcher_init.rs`, `tests/fixtures/help/all*.txt`, `tests/launcher_auto_init.rs`, `tests/retired_hosts_layer.rs`, `tests/product_cli.rs` with `tests/fixtures/protocol/v1/current/01-usage-unknown-subcommand.json` (C-328)
- read-only: everything else

## Plan

1. RED: the new `tests/omp_launcher_removed.rs` (AC-1, AC-2, AC-3, AC-4, AC-6, AC-8), `omp_is_not_a_tool_any_more` (AC-5) and the AC-7 case; each fails by assertion on the tip, except AC-4 and AC-6 where the tip may already hold (they then guard the change).
2. Launcher: the `Omp` command and the `omp` host word go, `OMP_REMOVED` and `Msg::OmpRemoved` join the Codex notice, `help omp` and `help codex` route to the notices, the session drops its OMP command line, key and spellings, and the help tables, goldens and recorded usage error follow; the tests that name removed things change in the same commit.
3. Model settings: `TOOLS` becomes `claude` and `glm`; the retired-host read rule then ignores a saved `omp` value; the tests that used `omp` as the other tool move to `glm`.
4. Install texts: the two scripts and the README name `moai cc` and `moai glm` only; the install script test follows.
5. Checks: fmt, clippy `--all-targets`, the touched targets, `moai spec check`, one `moai check affected` over the branch.

## Risks

- Clippy `--all-targets` flags a one-element `for` left by the removal; such loops are rewritten without a loop.
- The shorter launcher row may move the help about column; the goldens show it.
- The new notice texts carry no citation, because `src/main.rs` and `src/help.rs` are scanned as citation-free.

## Decisions

- D-1: A separate `Msg::OmpRemoved` and `OMP_REMOVED`, mirroring the Codex notice, rather than one notice with a parameter (planning default).
- D-2: The removal test is a new file `tests/omp_launcher_removed.rs`; `tests/omp_launcher.rs` is deleted (planning default).
- D-3: The existing Python install test is edited in the install commit; no Python test is added.

## Result

All REQs are coded on this branch after the RED commit 3d9a8b0830: the launcher (5afdc18d93), the model settings (1c33e18fef) and the install texts (b6ef0c3968), written by a moai-impl agent in its own worktree on top of the RED commit and brought here by fast-forward. The RED lines come from this lane's run on 3d9a8b0830 (`target/red2.log`); the passes come from `moai check affected 59cefecb76..b6ef0c3968` (PASS, 70 targets, `target/check-g.log`, step logs in `target/check-g/`).

- AC-1 met: `moai_omp_says_it_was_removed` failed first (`moai omp` composed an `omp` launch and exited 0) and passes now; `omp_launcher_removed` 6 passed.
- AC-2 met: `help_omp_and_help_codex_print_the_removal_notice` failed first (`help omp` exited 0 with an `omp` launch) and passes in the same run.
- AC-3 met: `the_omp_host_word_is_refused_like_any_unknown_host` failed first and passes in the same run.
- AC-4 met: `cc_and_glm_launches_are_unchanged_and_name_no_omp` failed first (the `cc -l` dry run named `omp` in a note) and passes in the same run.
- AC-5 met: `omp_is_not_a_tool_any_more` failed first (`TOOLS` was `["claude", "omp", "glm"]`) and passes now; `model_effort_settings` 12 passed.
- AC-6 met: `a_saved_omp_value_of_any_shape_never_changes_a_cc_or_glm_launch` failed first (`launcher.models.omp must be an object` refused the `cc -l` dry run) and passes in the same run.
- AC-7 met: `a_saved_omp_value_is_not_reported_and_leaves_the_file_on_the_next_save` failed first and passes now; `update_model_pick` 12 passed.
- AC-8 met: `no_help_usage_or_install_text_names_omp_or_moai_codex` failed first (`tests/fixtures/help/root.txt:4` named `cc|glm|omp`) and passes in the same run; `help_layout` 10 passed and `tests/install_script_test.py` ran in the python step (302 tests, OK, 7 skipped).
- Check results: check affected 59cefecb76..b6ef0c3968 PASS: diff_check 0, fmt 0, bash_n `install/install.sh` 0, clippy 0, test_lib 0, test_bins 0, test_targets 0, build_cli 0, python 0.
- Changes beyond the listed lines, all inside granted files: `launch_settings` in `src/session.rs` lost its `host` argument (its only use was the removed OMP advisor guard) with its two callers; `default_profile` in `src/launcher_init.rs` keeps its signature with `_host` and always answers `claude`; `tests/install_script_test.py` renames `test_omp_gets_its_skill_entry` to `test_an_omp_or_codex_folder_alone_is_no_host`; `tests/model_effort_questions.rs` gains a supported `claude.agent` entry to keep a "from Some" case; the `src/launcher_terminal.rs` startup-lines test uses a GLM vault note instead of the removed OMP note.
- Left as is (outside this SPEC): the `hooks-plan` doc text at `src/main.rs:614` and the comment at `src/retired_hosts.rs:5`, both project-layer wording.
- Not run: the whole suite, a test run after merging the tip, and the independent audit.
