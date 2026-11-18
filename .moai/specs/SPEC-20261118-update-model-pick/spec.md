---
id: SPEC-20261118-update-model-pick
title: "update: ask the tool pick and the model and effort per tool on a terminal (C-47 slice 3)"
status: draft
revision: 1
created: 2026-11-18
parent: SPEC-20261118-install-model-per-tool
---

# update: ask the tool pick and the model and effort per tool on a terminal (C-47 slice 3)

## Intent

Slice 3 of card C-47 (leader note 2211: the update questions start after the C-54 merge). The settings module and the question loop are merged (PRs #3207 and #3209) and nothing calls them yet. This slice wires them into `moai update`: on a person's terminal the update asks which tools to set up, then the leader and agent model and effort of each picked tool, shows what would change, and writes the choice to the user launcher settings file through the digest-bound apply. In every other run it asks nothing and reports which model and effort each tool and role would use.
The wiring wraps the existing `update_command` function: its signature, `src/main.rs` and `src/help.rs` stay as they are (the main file is held by card C-62), so the model stage runs after the update part and never changes how the update part is planned, confirmed or applied.

## Non-goals

- A command or flag for a host session to apply picks without a terminal; the host workflow only learns to read and show the reported values (a later slice adds the command when the main file is free).
- The launchers reading the values (launcher slice, after C-62 slice B) and the install script (install slice, after C-67).
- A localized wording of the stage: the stage prints English lines; `src/ui_text.rs` is outside this card scope.
- Any change to the update part, its plan digest or its receipts.

## Requirements

- REQ-1: When `moai update` runs with the person's screen (a terminal on stdin and stderr, screens on), without `--confirm`, `--confirm-cleanup` or `--confirm-home`, and the update part wrote files, the model stage shall ask the tool pick first and then the model and effort questions of the picked tools. · failure: none — REQ-2 covers the empty and cancelled pick.
- REQ-2 (REQ-1): IF no tool is picked or the pick is cancelled, THEN the model stage shall write nothing, ask no model question and report the state `nothing_picked` or `cancelled`.
- REQ-3: When the picks are made, the model stage shall show each change as tool, role, the saved value and the new value, and name the saved entries outside the support table that stay unchanged, before it asks once whether to save. · failure: none — a refusal is REQ-4.
- REQ-4 (REQ-3): IF the answer to the save question is not a yes, THEN the model stage shall write nothing and report the state `declined`.
- REQ-5 (REQ-3): When the answer is a yes, the model stage shall write through the digest-bound apply of the question module, which plans again, compares the digest and reads the file back, and report the state `applied` with the changes. · failure: none — REQ-6 covers a refusal.
- REQ-6 (REQ-5): IF the settings cannot be read, planned or written, or the digest no longer matches, THEN the model stage shall report the state `error` with the message of the apply and leave the update result and the settings file as they were.
- REQ-7 (REQ-3): IF the picks would change nothing, THEN the model stage shall write nothing and report the state `unchanged`.
- REQ-8: When `moai update` ends applied in any other way (a pipe, `--json`, a confirm flag, no terminal or an update part that wrote nothing), the update shall ask no model question and add a `models` object to its result that lists for each tool and role the model, the effort and the source (`saved` or `default`) and names the saved entries outside the support table by key path. · failure: none — an unreadable settings file is reported as the state `error` with its message.
- REQ-10 (REQ-8): IF the result is a preview, THEN the update shall leave its keys exactly as they were, as the preview test of the cleanup stage requires.
- REQ-9: When the tool question is asked, the model stage shall pass `claude` as installed when `~/.claude` is a folder and `codex` as installed when `~/.codex` is a folder, so that Enter picks them. · failure: none — no folder means the empty default and Enter picks nothing.

## Acceptance

- AC-1 (REQ-1, REQ-2): the first thing the stage prints is the tool question, and an empty pick or a cancelled pick ends without a model question and without a write.
  verify: `cargo test --locked --offline --test update_model_pick the_tool_pick_comes_first_and_an_empty_pick_asks_nothing`
- AC-2 (REQ-3, REQ-6): each picked tool offers only values from the support table, and a saved entry outside the table is named in the preview and stays unchanged in the file.
  verify: `cargo test --locked --offline --test update_model_pick only_supported_values_are_offered_and_an_unknown_saved_value_is_kept`
- AC-3 (REQ-3, REQ-9): Enter at every question selects the defaults of the installed tools, and the preview with them is shown before the file exists.
  verify: `cargo test --locked --offline --test update_model_pick enter_selects_the_defaults_and_the_preview_shows_them`
- AC-4 (REQ-5): the choice is written through the digest-bound apply and the file read back holds it, other launcher keys kept.
  verify: `cargo test --locked --offline --test update_model_pick the_choice_is_written_through_the_digest_bound_apply_and_read_back`
- AC-5 (REQ-4, REQ-7): a no writes nothing, and picks equal to the saved values report `unchanged` without a write.
  verify: `cargo test --locked --offline --test update_model_pick a_no_and_an_unchanged_pick_write_nothing`
- AC-6 (REQ-6): a settings file that cannot be used, or that changes between the preview and the save, is reported as `error` and the file is as it was.
  verify: `cargo test --locked --offline --test update_model_pick a_broken_or_changed_settings_file_is_an_error_and_stays_as_it_was`
- AC-7 (REQ-8, REQ-10): without a terminal, with a pipe or with `--json`, the real command asks nothing, a preview keeps its keys, and the applied result carries the reported values, saved or default, and the unsupported entries.
  verify: `cargo test --locked --offline --test update_model_pick a_non_terminal_run_asks_nothing_and_reports_the_values`
- AC-8 (REQ-9): the installed tools follow the two folders of the home.
  verify: `cargo test --locked --offline --test update_model_pick the_installed_tools_follow_the_home_folders`
- AC-9 (REQ-1, REQ-8): the existing update command tests still pass with the new `models` key, and a real terminal run (a pseudo terminal on macOS) that wrote files asks the model stage and saves the choice after a yes.
  verify: `cargo test --locked --offline --test update_command_cli`
- AC-10 (REQ-8): the update workflow text names the `models` key and the terminal stage, and the source manifest carries its new hash.
  verify: `cargo test --locked --offline --test product_cli source_manifest_covers_every_package_file`
- AC-11 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8, REQ-9): the code style checks.
  verify: `cargo fmt --check`

## Scope

- write: `.moai/specs/SPEC-20261118-update-model-pick/spec.md`, `src/update_models.rs`, `src/update.rs`, `src/lib.rs`, `src/model_effort_questions.rs`, `tests/update_model_pick.rs`, `tests/model_effort_questions.rs`, `tests/model_effort_settings.rs`, `skills/moai/workflows/update.md`, `skills/moai/references/source-manifest.json`
- read-only: `src/update_tux.rs`, `src/progress.rs`, `src/home_root.rs`, `src/session.rs`, `tests/update_command_cli.rs`
- forbidden: `src/main.rs`, `src/help.rs`, `src/ui_text.rs`, the real HOME, a live tool session, the network

## Plan

1. Tests first (RED): `tests/update_model_pick.rs` with the eight cases, which fail to compile until the module exists.
2. The module `src/update_models.rs`: the settings path, the installed tools, the report, the stage; the library declares the three model modules; the two earlier test files move from the path include to the library path.
3. `update_command` becomes a wrapper: the old body is kept as the core, the wrapper adds `models` and runs the stage on the person's screen.
4. The update workflow text and the manifest hash; the checks of the six update test files.

## Risks

- The stage runs only after an update that wrote files, so a person whose files are already up to date is not asked and cannot change the values by `moai update` until the next real update → a later slice adds a command or flag that asks on demand; `models` still reports the values.
- A new key in the update result could break a test that compares the whole result → only the applied result gets the key (REQ-10), and AC-9 runs the update test files.
- The stage prints English lines on the person's screen while the update screens are localized → the follow-up listed in Non-goals.
- The terminal path is covered by the stage tests with a captured input and output, not by a pseudo terminal: a real terminal run is `not_run` until a person or a later slice does it.
- The saved file limits found in the audits of PRs #3207 and #3209 stay (ancestor links, the lstat and read race, duplicate keys, a read-back fault) → they are listed in `SPEC-20261118-model-effort-questions`.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Wrap `update_command` instead of changing its signature or `src/main.rs` | the main file is held by card C-62 and the signature is called from the stale retry path | move the call into the main file when it is free |
| D-2 | The stage runs only for the person's screen and only after an update that wrote files | a preview result means the person said no or the run was scripted; asking at every up-to-date run would nag a person who updates often, and the first install and every real update still ask | also ask when nothing is to apply, or add a flag that asks on demand |
| D-3 | One save question after the full preview, separate from the update question | the update digest does not cover the settings file; the question module has its own digest | fold both into one preview later |
| D-4 | The installed tools are the folders `~/.claude` and `~/.codex` | the update manages these two hosts; OMP is not an update host and is picked by hand | list more hosts |
| D-5 | The report uses the saved value only when the support table lists it, else the default, and names the other entries | REQ-4 of the parent SPEC keeps and reports a value outside the table | none |

## Result

- AC-1: pending
- AC-2: pending
- AC-3: pending
- AC-4: pending
- AC-5: pending
- AC-6: pending
- AC-7: pending
- AC-8: pending
- AC-9: pending
- AC-10: pending
- AC-11: pending
