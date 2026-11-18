---
id: SPEC-20261118-model-effort-questions
title: "model and effort questions: tool pick, selections, plan and digest-bound apply (C-47 slice 2a)"
status: draft
revision: 1
created: 2026-11-18
parent: SPEC-20261118-install-model-per-tool
---

# model and effort questions: tool pick, selections, plan and digest-bound apply (C-47 slice 2a)

## Intent

Slice 2a of card C-47. The parent SPEC wants `moai update` and the install to ask which tools to set up and then, for each picked tool, a leader and an agent model with its effort, and to save the choice through a preview and a digest-bound apply. The update command, the help table, the skill workflow and the source manifest belong to slice 2b, which waits for the card that holds the manifest (leader note 2191). This slice builds everything that does not touch them: a module that builds the questions from the support table of slice 1, reads an answer line, runs the question loop on any reader and writer, plans the change of the `launcher.models` key and applies it to the launcher settings file by path.
It also takes the audit follow-up N1 of slice 1: a saved entry outside the support table is listed by key path and left unchanged unless the person is asked about that tool and role, so a successful save never hides a file the launcher would then refuse.

## Non-goals

- The command line of `moai update`, the commands that expose the plan and the apply, the skill step, the help table and the source manifest (slice 2b).
- Resolving the real HOME or the launcher file path: the caller passes a path.
- A revert command: the plan keeps the old value of each change, and slice 2b builds the revert on it.
- Colors, translations and the terminal detection: the loop writes plain text to the writer it is given.

## Requirements

- REQ-1: When the tool question is built, the module shall list Claude Code, Codex and OMP as a multiple choice whose default is the installed tools the caller passes. · failure: none — no installed tool makes the default an empty pick.
- REQ-2: When a model or effort question is built for a tool and role, the module shall list only values of the support table, the default model first and the efforts in table order with the default marked. · failure: none — an unknown tool or role has no question.
- REQ-3: When an answer line is read, the module shall take an empty line as the default, a number or a list of numbers as those options, and end of input as cancelled. · failure: none — a line that is not a valid choice is REQ-4.
- REQ-4 (REQ-3): IF an answer line is not a valid choice or has more than 4096 content bytes (the newline is not counted), THEN the loop shall ask the same question again up to three times without echoing the line, and then end as cancelled.
- REQ-5: When the tool pick is empty, the loop shall ask no model question and return no pick. · failure: none — an empty pick is a valid answer, not an error.
- REQ-6: When picks are planned, the plan shall change only the entries of the picked tools and roles (an OMP pick sets omp and glm), keep every other key, list every other saved entry outside the support table by key path and leave it unchanged, and carry a digest of the launcher object before, the picks and the launcher object after. · failure: none — structural damage is REQ-7.
- REQ-7 (REQ-6): IF launcher.models is structurally broken (not an object, a tool or role name this version does not know, a level that is not an object), THEN the plan shall refuse, name the key path above the damage and quote no saved text.
- REQ-8: When a plan is applied with its digest, the apply shall plan again from the file and the picks, write nothing when the plan changes nothing, otherwise write the file atomically through a temporary name no other writer uses, keep the other root keys and the file mode, remove only a temporary file it created, and read the whole file back. · failure: none — refusals are REQ-9.
- REQ-9 (REQ-8): IF the digest differs from the new plan, or the target or its folder is a link, or the target is a directory, a special file or larger than one MiB, or the file to write would be larger than one MiB, THEN the apply shall write nothing and name the reason.

## Acceptance

- AC-1 (REQ-1): the tool question lists the three tools, and its default is the installed ones.
  verify: `cargo test --locked --offline --test model_effort_questions the_tool_question_lists_three_tools_and_defaults_to_the_installed_ones`
- AC-2 (REQ-2): model questions list only table values with the default first, effort questions list the efforts of the chosen model in table order with a default the model offers, and a name the table does not know has no question.
  verify: `cargo test --locked --offline --test model_effort_questions model_and_effort_questions_offer_only_table_values_with_the_default_first`
- AC-3 (REQ-3, REQ-4): an empty line, a number and a list are read; a bad line or a number out of range is invalid; end of input is cancelled; no result or message contains the line.
  verify: `cargo test --locked --offline --test model_effort_questions an_answer_line_is_read_and_a_bad_one_is_never_echoed`
- AC-4 (REQ-4, REQ-5): the loop asks the tool pick first, nothing after an empty pick, repeats a bad answer three times and then cancels.
  verify: `cargo test --locked --offline --test model_effort_questions the_loop_asks_the_tool_pick_first_and_stops_on_an_empty_or_cancelled_pick`
- AC-5 (REQ-3, REQ-5): pressing Enter at every question picks the defaults of the installed tools.
  verify: `cargo test --locked --offline --test model_effort_questions enter_everywhere_picks_the_defaults_of_the_installed_tools`
- AC-6 (REQ-6): the plan changes the picked entries only, keeps every other key, and lists a saved unsupported entry by key path unchanged unless it is picked.
  verify: `cargo test --locked --offline --test model_effort_questions the_plan_changes_only_the_picked_entries_and_lists_unsupported_ones`
- AC-7 (REQ-6): an OMP pick sets the omp and glm entries to the same choice.
  verify: `cargo test --locked --offline --test model_effort_questions an_omp_pick_sets_omp_and_glm`
- AC-8 (REQ-7): a broken launcher.models is refused by key path without quoting it.
  verify: `cargo test --locked --offline --test model_effort_questions a_structurally_broken_models_key_is_refused_without_quoting_it`
- AC-9 (REQ-8): the apply writes through the digest, keeps the other root keys and the mode, creates an absent file with only the launcher key, and reads the result back.
  verify: `cargo test --locked --offline --test model_effort_questions apply_writes_through_the_digest_and_reads_back`
- AC-10 (REQ-9): a wrong digest, a changed file, a link, a directory and an oversize file are refused and the bytes stay the same.
  verify: `cargo test --locked --offline --test model_effort_questions apply_refuses_a_stale_digest_and_an_unsafe_target_and_writes_nothing`
- AC-11 (REQ-6): the slice 1 module lists unsupported saved entries by key path, and the cases the slice 1 audit named (a non-object tool entry, extra fields of a replaced entry, the omp tool with a saved value, the refusal text of a refused set) are pinned.
  verify: `cargo test --locked --offline --test model_effort_settings`
- AC-12 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8, REQ-9): the code style checks.
  verify: `cargo fmt --check`
- AC-13 (REQ-8, REQ-9): a folder that is a link is refused with nothing behind it, a mode the umask would change is kept, a temporary-looking file of another writer is never touched, a folder that cannot be written fails clean, exactly one MiB is read and one byte more is refused, and a plan that changes nothing writes nothing (also for an absent file).
  verify: `cargo test --locked --offline --test model_effort_questions the_apply_guards_hold_at_their_edges`
- AC-14 (REQ-4): an answer line with 4096 content bytes is an ordinary line, one with 4097 is one invalid answer whose rest is dropped, a long tail never becomes a second answer, and at the end of input it ends as cancelled.
  verify: `cargo test --locked --offline --test model_effort_questions an_overlong_answer_line_is_one_invalid_answer`; `cargo test --locked --offline --test model_effort_questions the_answer_line_cap_is_exact_and_a_long_tail_never_leaks`
- AC-15 (REQ-8, REQ-9): a rename that fails removes the temporary file this call created, and a rewrite that would grow a file past one MiB is refused before anything is created.
  verify: `cargo test --locked --offline --test model_effort_questions a_failed_rename_removes_its_temporary_file_and_a_growing_file_is_refused`

## Scope

- write: `.moai/specs/SPEC-20261118-model-effort-questions/spec.md`, `src/model_effort_questions.rs`, `tests/model_effort_questions.rs`, `src/model_effort_settings.rs`, `tests/model_effort_settings.rs`
- read-only: `src/session.rs`, `src/progress.rs`, `src/update.rs`
- forbidden: `src/lib.rs`, `src/main.rs`, `src/help.rs`, `skills/moai/references/source-manifest.json`, `skills/moai/workflows/update.md`, the real HOME, the network

## Plan

1. Tests first (RED): `tests/model_effort_questions.rs` includes both modules by path, so it fails to compile until the new module exists; the extra cases of slice 1 are added to `tests/model_effort_settings.rs`.
2. `src/model_effort_settings.rs` gets `unsupported_entries`; the new module `src/model_effort_questions.rs` gets the questions, the answer reader, the loop, the plan and the apply.
3. Slice 2b (after the card that holds the manifest merges): the update command line, the plan and apply commands, the skill step, the help table, the manifest entry, the library declaration of both modules, and the revert.

## Risks

- The loop and the apply are tested on temporary folders only; a live terminal and the real launcher file are slice 2b evidence, reported not_run here.
- A launcher file written by hand may hold anything → every refusal names a key path and quotes no text, so a secret in the file cannot reach a message.
- The apply replaces a file under HOME in slice 2b → this slice only takes a path, and the apply re-plans from the file, so a file that changed after the preview is refused.
- Limits the code does not close and the tests do not pin (slice 1 and 2a audits): a link in an ancestor above the immediate folder is not checked; the lstat before the read and the open are not one atomic step, so a swap between them is not caught (a swap to a link only changes what is read, the write replaces the entry and never writes through it); the failure of the read-back after the rename cannot be induced by a test (it needs a fault between the rename and the read), and a failed rename can leave behind the empty folder the apply created; a hand-edited file with duplicate root keys keeps the last value, and a number beyond 64 bits loses precision when the file is rewritten → slice 2b shows the plan, which lists the launcher changes, before any write.
- The two modules are included by path in the tests because the library file is outside this card scope; slice 2b declares them in the library and switches the includes.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Slice 2 of the leader order is split: 2a is this SPEC (no manifest file), 2b is the wiring | leader note 2191: tests outside the manifest first, wiring after the manifest holder merges | merge 2b into one SPEC later |
| D-2 | An OMP pick sets the omp and glm entries to the same choice | the card says `moai glm` is configured like OMP | ask for glm separately |
| D-3 | Enter at the tool question picks the installed tools; Enter at a model or effort question picks its default | the card: with no input the defaults are chosen; an empty pick must still be possible | change the default in the question builder |
| D-4 | A saved entry outside the table is listed and left unchanged unless picked; structural damage is refused | slice 1 audit N1; the parent REQ-4 keeps an unsupported value unchanged | refuse every unsupported entry |
| D-5 | The apply plans again from the file and the picks and compares the digest, like the other apply commands | a stale preview must never write | none |
| D-6 | The digest covers the sorted launcher object before, the sorted picks and the sorted launcher object after; other root keys are kept but not part of the digest | key order of a file is not meaningful and the library keeps insertion order; the apply never changes another root key | add the other root keys to the digest |
| D-7 | Three invalid answers to one question end the loop as cancelled | a bounded loop never hangs a script | change the limit |

## Result

- AC-1: met — `the_tool_question_lists_three_tools_and_defaults_to_the_installed_ones` passes
- AC-2: met — `model_and_effort_questions_offer_only_table_values_with_the_default_first` passes (default model first, efforts in table order with a default the model offers, no question for an unknown name)
- AC-3: met — `an_answer_line_is_read_and_a_bad_one_is_never_echoed` passes (empty, numbers, lists, 8 bad lines, end of input)
- AC-4: met — `the_loop_asks_the_tool_pick_first_and_stops_on_an_empty_or_cancelled_pick` passes (a fourth try would complete, so three tries are pinned)
- AC-5: met — `enter_everywhere_picks_the_defaults_of_the_installed_tools` passes
- AC-6: met — `the_plan_changes_only_the_picked_entries_and_lists_unsupported_ones` passes (digest independent of key order)
- AC-7: met — `an_omp_pick_sets_omp_and_glm` passes
- AC-8: met — `a_structurally_broken_models_key_is_refused_without_quoting_it` passes (5 shapes)
- AC-9: met — `apply_writes_through_the_digest_and_reads_back` passes (other root keys, mode 0644 kept, absent file created with mode 0600, no temp file left)
- AC-10: met — `apply_refuses_a_stale_digest_and_an_unsafe_target_and_writes_nothing` passes (wrong digest, changed file, link, folder, FIFO, oversize file, broken file)
- AC-11: met — `cargo test --locked --offline --test model_effort_settings` 8 passed, including `unsupported_saved_entries_are_listed_by_key_path_without_quoting_them` and `the_slice_one_audit_cases_are_pinned`
- AC-12: met — `cargo fmt --check` exits 0 and `cargo clippy --test model_effort_settings --test model_effort_questions -- -D warnings` exits 0
- AC-13: met — `the_apply_guards_hold_at_their_edges` passes (link folder, mode 0664 kept, foreign temporary file untouched, unwritable folder clean, 1 MiB and 1 MiB + 1, no-op writes nothing)
- AC-14: met — `an_overlong_answer_line_is_one_invalid_answer` and `the_answer_line_cap_is_exact_and_a_long_tail_never_leaks` pass (4096 accepted, 4097 refused, a tail ending in 1 does not leak)
- AC-15: met — `a_failed_rename_removes_its_temporary_file_and_a_growing_file_is_refused` passes (rename onto a parent entry fails and leaves no temporary file; a 960 KB file that would become 3 MB is refused and left as it was)
