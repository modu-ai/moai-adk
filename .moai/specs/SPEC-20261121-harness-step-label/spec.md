---
id: SPEC-20261121-harness-step-label
title: "The generation step shows users the name Project harness while its id and commands stay"
status: draft
revision: 1
created: 2026-11-21
---

# The generation step shows users the name Project harness while its id and commands stay

## Intent

The step of `moai project-update-plan` that writes the managed `AGENTS.md` section and the project agents, skills and workflows has the id `generation`. The user decided (2026-11-21, "맞다, 표시 이름만 변경") that wherever a person reads about this step it is called `프로젝트 맞춤 하네스` in Korean and `Project harness` in English, and that the internal names stay for compatibility: the step id `generation`, the commands `generation-plan` and `generation-apply`, the file `manifests/generation.json` and the JSON fields.

The step has no label field in the plan; people meet its name in two places. The `reason` of the install phase lists the steps by name (`profile, generation (AGENTS.md managed section and generated files), …`), and the `/moai update` and `/moai project` workflows tell the model to read the plan out, which is where Korean and English screens come from. This change names the display name in both places and pins it, and changes no identifier.

The plan stays English bytes (the JSON envelope never speaks Korean); the Korean name reaches a Korean user through the workflow text, which tells the model what to call the step.

## Non-goals

- Renaming the step id, the commands, the manifest path, any JSON field or any error text that names the manifest file (`generation manifest unreadable …` names the real file `manifests/generation.json`).
- A new `label` field in the plan JSON, or a per-language label table in `src/ui_text.rs`: the plan never reaches a human screen of its own.
- Japanese or Chinese text templates: none exist for the plan or the workflows in this repository; the workflow sentence tells the model to translate the display name for any other conversation language.
- The word `generation` where it means the era of the tool (`Go-generation`, `earlier generation`) or a card claim generation: those are different things.

## Requirements

- REQ-1: The install phase reason of the plan shall name the generation step by its display name `project harness`.
- REQ-2: The install phase reason of the plan shall not contain the word `generation`.
- REQ-3: The step id `generation` and the commands `generation-plan` and `generation-apply` and the path `manifests/generation.json` shall stay unchanged.
- REQ-4: The `/moai update` workflow text shall tell the model to call the step `Project harness` in English and `프로젝트 맞춤 하네스` in Korean.
- REQ-5: The `/moai project` workflow text shall tell the model to call the step `Project harness` in English and `프로젝트 맞춤 하네스` in Korean.
- REQ-6: Each of the two workflow texts shall tell the model to translate the display name for any other conversation language and to keep the id `generation` for tools.

## Acceptance

- AC-1 (REQ-1, REQ-2): the constant `INSTALL_REASON` of `src/project_update.rs` contains `project harness` and does not contain `generation`, and it still lists the same other steps in the same order · verify: `cargo test --test harness_step_label install_reason`
- AC-2 (REQ-3): `src/project_update.rs` still names the step `generation` in its plan code, still prints the apply command `generation-apply`, and still reads `manifests/generation.json`; `src/help.rs` still lists `generation-plan` and `generation-apply` · verify: `cargo test --test harness_step_label identifiers`
- AC-3 (REQ-4, REQ-5, REQ-6): `skills/moai/workflows/update.md` and `skills/moai/workflows/project.md` each hold one sentence that names `Project harness` and `프로젝트 맞춤 하네스`, tells to translate it for another language and keeps the id `generation` for tools · verify: `cargo test --test harness_step_label workflow_text`

## Scope

- write: `tests/harness_step_label.rs` (new), `.moai/specs/SPEC-20261121-harness-step-label/**`
- paired (written by the cards that hold these files, from a pair line, and merged in one chain with this change): `src/project_update.rs` (one string, `INSTALL_REASON`), `skills/moai/workflows/update.md` and `skills/moai/workflows/project.md` (one sentence each), `skills/moai/references/source-manifest.json` (the two sha256 entries)
- read-only: `src/help.rs`, `skills/moai/workflows/setup.md`
- forbidden: any identifier, command name, manifest path, JSON field or error text that names a file; the Go repository; the real HOME

## Plan

Tests first (RED). `tests/harness_step_label.rs` reads the source and the two workflow texts, runs no command and needs no HOME, like `tests/project_preview_order.rs`. On the tip alone the three groups fail for the right reasons (the install reason says `generation`; the workflows have no sentence). The three paired edits arrive from their holders as pair lines; the chain is merged together and the tests read green on the merge.

The workflow sentence (one per file, placed where the file first reads the plan out): `Call the generation step "Project harness" in English and "프로젝트 맞춤 하네스" in Korean, and translate that name for any other conversation language; the id stays `generation` for the commands and the JSON, never as the name you say to the user.`

The install reason becomes `profile, project harness (AGENTS.md managed section and generated files), Claude Code settings, …` and the rest of the string is unchanged.

## Risks

- The tests read red on this branch alone: the three edits live in files other cards hold. The pair lines name this change as the pair, and the READY waits for the merged chain.
- The shipped text must not gain a maintainer token (a card number, an issue number, a repository link or the local-rules file name): the sentence has none, and the ratchet scan of the shipped paths still passes with the sentence added.
- `skills/moai` edits need their sha256 entries in the source manifest in the same change; the holder of the manifest updates them with the sentence.
- A reader of the plan JSON who looks for the display name in a field will not find it (non-goal); the install reason and the workflow text are where it is.

## Result

(not yet written)
