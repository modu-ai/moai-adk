---
id: SPEC-20261118-model-effort-settings
title: "model and effort settings: support table, defaults and the launcher.models key (C-47 slice 2)"
status: draft
revision: 1
created: 2026-11-18
parent: SPEC-20261118-install-model-per-tool
---

# model and effort settings: support table, defaults and the launcher.models key (C-47 slice 2)

## Intent

Slice 2 of card C-47 (leader decision, card note 2160: one SPEC and one pull request per code slice, settings file first). The parent SPEC fixes what the install and the launchers must do; this slice builds the part both of them need and nothing else: a module that holds the table of models and efforts each tool supports, the defaults, and two pure functions on the user launcher object, one that resolves the model and effort of a tool and role and one that sets a choice.
The choice lives in the existing user-only launcher settings file under the key `launcher.models`, shaped `launcher.models.<tool>.<role>` with the two fields `model` and `effort`. No new file is created; the file reader stays where it is and the update slice and the launcher slice call this module later.

## Non-goals

- Reading or writing the settings file, the preview and the digest-bound apply (the update slice).
- Passing the values to a tool at start (the launcher slice).
- Any change to the existing launcher keys `role_access` and `advisor`.
- Declaring the module in the library: that line is outside this card scope and the update slice adds it.

## Requirements

- REQ-1: When the saved launcher object holds no value for a tool and role, the resolver shall answer the default model and effort of that tool and role and mark the source default. · failure: none — the defaults cover all four tools and both roles.
- REQ-2: When the saved launcher object holds a model and an effort for a tool and role that the support table lists, the resolver shall answer them and mark the source saved. · failure: none — a value outside the table is REQ-3.
- REQ-3 (REQ-2): IF a saved value is outside the support table, has the wrong shape, or names a tool or role the module does not know, THEN the resolver shall refuse with a message that names the key path and quotes no saved text.
- REQ-4: When a choice is set, the setter shall return a copy of the launcher object in which only that tool and role entry is new and every other key is kept. · failure: none — an unsupported choice is REQ-5.
- REQ-5 (REQ-4): IF the model or the effort is not in the support table for the tool, THEN the setter shall refuse and leave its input unchanged.
- REQ-6: When the support table is read, it shall list for each tool only the model ids and efforts named in the research file of the parent SPEC. · failure: none — a manual comparison pins it.

## Acceptance

- AC-1 (REQ-1): with no saved value, every tool and role answers its default and the source default.
  verify: `cargo test --locked --offline --test model_effort_settings no_saved_value_answers_the_default_for_every_tool_and_role`
- AC-2 (REQ-2): a saved supported value wins over the default and the source is saved.
  verify: `cargo test --locked --offline --test model_effort_settings a_saved_supported_value_wins_and_is_marked_saved`
- AC-3 (REQ-3): an unsupported model, an unsupported effort, a wrong type, a missing field, an extra field, an unknown tool and an unknown role are all refused, and the refusal names the key path and never contains the saved text.
  verify: `cargo test --locked --offline --test model_effort_settings an_unsupported_or_malformed_saved_value_is_refused_without_quoting_it`
- AC-4 (REQ-4): setting one choice changes that entry only; role_access, advisor and the other entries are equal before and after, and the input object is not changed.
  verify: `cargo test --locked --offline --test model_effort_settings setting_a_choice_changes_one_entry_and_keeps_every_other_key`
- AC-5 (REQ-5): an unsupported model or effort is refused and the input is equal to a copy taken before.
  verify: `cargo test --locked --offline --test model_effort_settings an_unsupported_choice_is_refused_and_the_input_is_unchanged`
- AC-6 (REQ-1, REQ-5): every default model and effort is inside the support table of its tool.
  verify: `cargo test --locked --offline --test model_effort_settings every_default_is_inside_the_support_table`
- AC-7 (REQ-6): the table rows equal the rows of the research file.
  manual: read the research file of SPEC-20261118-install-model-per-tool beside the table in the module; the model ids and efforts per tool are the same, and neither offers ultra or ultracode.
- AC-8 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5): the code style checks.
  verify: `cargo fmt --check`

## Scope

- write: `.moai/specs/SPEC-20261118-model-effort-settings/spec.md`, `src/model_effort_settings.rs`, `tests/model_effort_settings.rs`
- read-only: `src/session.rs`, `.moai/specs/SPEC-20261118-install-model-per-tool/research.md`
- forbidden: `src/lib.rs`, `src/main.rs`, `src/session.rs`, `src/update.rs`, `tests/launcher*`, `tests/update*`, the real HOME, the network

## Plan

1. Tests first (RED): the six tests in `tests/model_effort_settings.rs`, which include the module by path, so they fail to compile until the module exists.
2. The module: tool and role names, the support table, the defaults, `resolve` and `with_choice`; no file access and no dependency beyond serde_json.
3. Later slices: the update slice declares the module in the library and calls it from the prompts; the launcher slice calls `resolve` at start.

## Risks

- The test includes the module by path because the library file is outside this card scope → when the update slice declares the module in the library, the test switches to the library path in that slice; until then a second copy of the code is never built into the product.
- Model ids change quickly → the table is one constant block with the read time in the research file, and a saved id that disappears is refused by REQ-3 and never replaced.
- A saved file written by hand may hold any text → REQ-3 quotes none of it, so a secret pasted into the file cannot reach an error message.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The key is `launcher.models.<tool>.<role>` with `model` and `effort`, inside the existing user launcher file | one settings file, as the card asks; the reader already validates the file type | rename the key in this module only |
| D-2 | Tool names are `claude`, `codex`, `omp` and `glm`; role names are `leader` and `agent` | the launchers are cc, glm and codex and the install offers Claude Code, Codex and OMP; the launcher host name is mapped to the tool name by the launcher slice | add or rename a name in the module |
| D-3 | The module does no file access | the update slice owns the preview and the digest-bound apply; the launcher slice owns the read | none |
| D-4 | Defaults follow decision D-2 of the parent SPEC: claude leader claude-opus-5-5 medium, agent claude-sonnet-5-5 high; codex leader gpt-6.1-sol high, agent gpt-6.1-terra xhigh; omp and glm leader zai/glm-5.3 high, agent zai/glm-5.3-flash max | the card table with the Codex ids corrected | edit the defaults block |
| D-5 | Efforts offered are low, medium, high, xhigh and max for Claude Code and Codex models, and low, high and max for the two GLM models; ultra and ultracode are not offered | decisions D-5 and D-6 of the parent SPEC | extend the lists |
| D-6 | A saved model is compared as an exact id; an alias such as opus is not accepted | one id per model keeps the saved file unambiguous | add aliases to the table |
