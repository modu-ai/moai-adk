---
id: SPEC-20261118-install-model-per-tool
title: "install and update: pick tools, then leader and agent model and effort per tool"
status: draft
revision: 1
created: 2026-11-18
parent: SPEC-20261117-sdd-foundation
---

# install and update: pick tools, then leader and agent model and effort per tool

## Intent

Card C-47 (user instruction 2026-11-16): during `moai update` and the moai-adk install the user first picks which coding tools to set up (Codex, Claude Code, OMP), then for each picked tool chooses a leader model with its effort and an agent model with its effort from the values that tool really supports. One settings file keeps the choice, and the launchers `moai cc`, `moai glm` and `moai codex` read it when they start the tool (`-l` leader values, `-a` agent values) instead of hard-coded values.
This SPEC fixes the requirements and records the research the card asked for first: the supported ids and efforts per tool, with source and read time (`research.md` beside this file). The code lands in later slices under this SPEC; the launcher entry text (PR #3201) is already merged and is not repeated.

## Non-goals

- The question tool each host uses for these prompts (card C-52, with C-51).
- Model ids beyond the six in the defaults table; any provider other than the three tools.
- Code, tests or the settings file in this slice: it adds the SPEC and the research only.

## Requirements

- REQ-1: When a person runs `moai update` on a terminal, the update shall ask which of Codex, Claude Code and OMP to set up before it asks any model question. · failure: none — REQ-2 covers the empty pick.
- REQ-2 (REQ-1): IF no tool is picked, THEN the update shall ask no model question and save nothing.
- REQ-3: When a tool is picked, the update shall offer a leader and an agent selection of model and effort limited to the values in the support table of `research.md`. · failure: none — a saved value outside the table is REQ-4.
- REQ-4 (REQ-3): IF a saved model or effort is not in the support table, THEN the update shall report it by name and keep it unchanged.
- REQ-5: When the person confirms without choosing, the update shall select the defaults of the defaults table and show them in the preview before any write. · failure: none — the preview is read only.
- REQ-6: When the person confirms the preview, the update shall write the choice to one settings file through the digest-bound apply and read it back. · failure: none — a stale digest is refused by the existing apply.
- REQ-7: When a launcher starts a tool with the leader entry or the agent entry, the launcher shall pass the model and effort of that role from the settings file to the tool. · failure: none — a missing file falls back to REQ-8.
- REQ-8 (REQ-7): IF the settings file has no value for the tool and role, THEN the launcher shall pass the default of the defaults table and say so.
- REQ-9: When the output is not a terminal or is JSON, the update shall ask no question and report which values it used. · failure: none — it reads the saved values or the defaults.
- REQ-10: When the settings file changes, the launcher shall use the new values at its next start without a reinstall. · failure: none — the file is read at every start.

## Acceptance

- AC-1 (REQ-1, REQ-2): on a terminal capture the update asks the tool pick first, and a pick of no tool ends without a model question and without a write.
  verify: `cargo test --locked --offline --test update_model_pick the_tool_pick_comes_first_and_an_empty_pick_asks_nothing`
- AC-2 (REQ-3, REQ-4): each picked tool offers only values listed in `research.md`, and a saved value outside the list is reported and unchanged.
  verify: `cargo test --locked --offline --test update_model_pick only_supported_values_are_offered_and_an_unknown_saved_value_is_kept`
- AC-3 (REQ-5): pressing Enter selects the defaults and the preview shows them before any write.
  verify: `cargo test --locked --offline --test update_model_pick enter_selects_the_defaults_and_the_preview_shows_them`
- AC-4 (REQ-6): the choice is written through the preview and the digest-bound apply and read back from the file.
  verify: `cargo test --locked --offline --test update_model_pick the_choice_is_written_through_the_digest_bound_apply_and_read_back`
- AC-5 (REQ-7, REQ-8): a test shows the exact argv and environment passed for each launcher and role, and the default with a notice when the file has no value.
  verify: `cargo test --locked --offline --test launcher_model_settings each_launcher_and_role_passes_the_saved_model_and_effort`
- AC-6 (REQ-9): without a terminal, with a pipe or with `--json`, nothing is asked and the values used are reported.
  verify: `cargo test --locked --offline --test update_model_pick a_non_terminal_run_asks_nothing_and_reports_the_values`
- AC-7 (REQ-10): changing the file and relaunching uses the new values.
  verify: `cargo test --locked --offline --test launcher_model_settings a_changed_file_is_used_at_the_next_start`
- AC-8 (REQ-3): the research table names for every id and effort its source and read time, and marks what was not verified.
  manual: read `research.md` beside this file; every table row has a source line with a read time and an Not verified line.

## Scope

- write: `.moai/specs/SPEC-20261118-install-model-per-tool/spec.md`, `.moai/specs/SPEC-20261118-install-model-per-tool/research.md`
- read-only: `src/session.rs`, `src/update.rs`, `skills/moai/workflows/update.md`
- forbidden: the real HOME, any live tool session, keys and secrets, the network at test time

## Plan

1. This slice: SPEC and research (documents only, audit-tier 0).
2. Slice settings: the settings file, its schema and the read and write functions, tests first.
3. Slice update: the tool pick and the model and effort selections in `moai update` and the install, with the preview.
4. Slice launchers: `moai cc|glm|codex` read the file for `-l` and `-a`; the argv and environment tests per launcher and role.
5. Each later slice states its own write scope on the card before it starts.

## Risks

- Model names change quickly → the table in `research.md` carries read times, and a value that disappears is reported by REQ-4, never replaced.
- The environment variable `CLAUDE_CODE_EFFORT_LEVEL` beats `--effort` in Claude Code → the launcher slice removes it from the child environment (D-3), checked by the argv and environment test of AC-5.
- Effort values listed in a catalog are not proof that a live session accepts them → the defaults use values the docs name for that model; the launcher slice runs one live start per tool as evidence, or reports `not_run`.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Ids written exactly as the research found them: `claude-opus-5-5`, `claude-sonnet-5-5`, `gpt-6.1-sol`, `gpt-6.1-terra`, `zai/glm-5.3`, `zai/glm-5.3-flash`; the card names `gpt-6.1 sol` and `GPT-6.1 Terra` map to the two Codex ids | `codex debug models` lists the hyphenated ids; the spaced form is not an id | change the table row |
| D-2 | Defaults: Claude Code leader Opus 5.5 medium, agent Sonnet 5.5 high; Codex leader `gpt-6.1-sol` high, agent `gpt-6.1-terra` xhigh; OMP and `moai glm` leader `zai/glm-5.3` high, agent `zai/glm-5.3-flash` max. Every default is inside the supported list of its model | the card table, checked against `research.md` | edit the card table and the defaults |
| D-3 | The launcher passes the effort as a flag and removes `CLAUDE_CODE_EFFORT_LEVEL` from the child environment | the variable wins over `--effort` in Claude Code | stop removing it |
| D-4 | Sonnet 5.5 high is passed explicitly | Claude Code defaults it to medium (the API default high is another surface) | none |
| D-5 | The selection lists offer low to max where the model lists them; `ultra` (Codex) and `ultracode` (Claude Code) are not offered | `ultra` adds automatic task delegation and `ultracode` needs a newer Claude Code; neither is in the card | add them to the list |
| D-6 | OMP efforts offered are `low`, `high` and `max` only | OMP offers exactly these for GLM-5.3; `--thinking off` is clamped to `low` | none |
| D-7 | The support table is a repository table with read times, not a live query | a live `codex debug models` or `omp models` call would make install depend on the tool being present | add a refresh command later |
| D-8 | The settings file path, schema and the `moai glm` transport are decided in their own slices | not researched here | none |
