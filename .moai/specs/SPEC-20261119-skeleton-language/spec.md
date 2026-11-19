---
id: SPEC-20261119-skeleton-language
title: "first-creation skeleton text follows the instruction language (C-60)"
status: draft
revision: 1
created: 2026-11-19
---

# first-creation skeleton text follows the instruction language (C-60)

## Intent

Card C-60. A `/moai project` run in a Korean project (profile `language.instruction_files` = `ko`) wrote a REVIEW.md whose managed block was Korean but whose first heading, written only when the file is created, was the English text `# Review instructions`; the Codex session then edited the generated file by hand. The managed block already follows the instruction language; the text that `agents-apply` writes outside the markers when it creates a file did not.
Measured at tip 061177768c: the only first-creation strings outside markers are the REVIEW.md heading (`src/agents.rs`, hard-coded English), the AGENTS.md heading (the project id, no words), the whole `AGENTS.local.md` text (chosen by the request language) and nothing for `.gitignore`. Only the REVIEW.md heading ignores the language. This SPEC fixes that heading and pins the full list so a new hard-coded string is seen.

## Non-goals

- Rewriting a heading of a file that already exists (the card says never).
- Changing the managed blocks, `AGENTS.local.md` wording or the `.gitignore` block.
- Languages other than `en` and `ko`: another instruction language keeps the request language, as the managed block does today.
- The `project-update-plan` command itself (it reads the same plan; this SPEC checks the plan rows).

## Requirements

- REQ-1: When `agents-apply` creates REVIEW.md, the apply shall write the heading in the instruction language, `# 검토 지침` for `ko` and `# Review instructions` for `en`, where the instruction language is the profile's `language.instruction_files` when it is `en` or `ko`, else the request language. · failure: none — REQ-2 covers an existing file.
- REQ-2: IF REVIEW.md already exists, THEN the apply shall leave its text outside the managed block as it is, whatever the instruction language. · failure: none — the file is never rewritten outside its block.
- REQ-3: When `agents-apply` creates AGENTS.md, REVIEW.md, AGENTS.local.md and `.gitignore` for a project, the text outside their managed markers shall hold no English skeleton words for a `ko` project and no Korean for an `en` project. · failure: none — the pinned list test finds a new hard-coded string.
- REQ-4: When the same apply is planned again without a hand edit, the plan shall mark every row `unchanged`. · failure: none — a changed row fails the test.

## Acceptance

- AC-1 (REQ-1): a new REVIEW.md starts with the Korean heading for `ko` and the English one for `en`, from the instruction setting and, without it, from the request language; an instruction code other than `en` and `ko` uses the request language.
  verify: `cargo test --locked --offline --test skeleton_language a_new_review_file_has_a_heading_in_the_instruction_language`
- AC-2 (REQ-2): an existing REVIEW.md keeps its heading and every byte outside the block, in both directions (English file with a `ko` profile, Korean file with an `en` profile).
  verify: `cargo test --locked --offline --test skeleton_language an_existing_review_file_keeps_its_own_heading`
- AC-3 (REQ-3): the four files created for a `ko` project and for an `en` project carry skeleton text in their own language only.
  verify: `cargo test --locked --offline --test skeleton_language every_first_creation_text_follows_its_language`
- AC-4 (REQ-4): after the apply the plan reads `unchanged` for every row, in `ko` and `en`.
  verify: `cargo test --locked --offline --test skeleton_language the_plan_reads_unchanged_right_after_the_apply`
- AC-5 (REQ-1, REQ-2, REQ-3, REQ-4): the neighbour tests pass.
  verify: `cargo test --locked --offline --test review_section_cli`; `cargo test --locked --offline --test language_setting`; `cargo test --locked --offline --test agents_project_cli`; `cargo fmt --check`

## Scope

- write: `.moai/specs/SPEC-20261119-skeleton-language/spec.md`, `src/agents.rs`, `tests/skeleton_language.rs`
- read-only: `tests/review_section_cli.rs`, `tests/language_setting.rs`, `src/language.rs`, `src/profile.rs`
- forbidden: `skills/moai/references/source-manifest.json`, the real HOME, a live tool session, the network

## Plan

1. Tests first (RED): the Korean heading is missing today.
2. One function for the REVIEW.md heading by language in `src/agents.rs`, used at the create site.
3. Neighbour tests, mutants, Result.

## Risks

- `AGENTS.local.md` follows the request language, not `instruction_files`: it states the answer language to the model, which is the conversation language the interview passes as the request language → kept and listed (D-2); `is_generated_local` pins its eight texts.
- The AGENTS.md heading is the project id and has no words → listed, nothing to translate.
- The Korean heading wording is a choice (the Codex session wrote `# 검토 지침` by hand) → the managed block heading says `프로젝트 리뷰 규칙`; both are Korean for the same file.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Korean REVIEW.md heading is `# 검토 지침` | what the person's Codex session wrote by hand | another wording |
| D-2 | `AGENTS.local.md` keeps the request language | it names the response language of the conversation | follow `instruction_files` and extend `is_generated_local` |
| D-3 | The heading follows the same rule as the managed block (`instruction_language`) | the file must read in one language | split the two |

## Result

- AC-1: met — `a_new_review_file_has_a_heading_in_the_instruction_language` passes (seven combinations: setting and request `ko`/`en`, mixed, no setting, and code `ja` falling back to the request language)
- AC-2: met — `an_existing_review_file_keeps_its_own_heading` passes (English file with a `ko` profile, Korean file with an `en` profile; every byte outside the block unchanged, one block)
- AC-3: met — `every_first_creation_text_follows_its_language` passes (AGENTS.md, REVIEW.md, AGENTS.local.md and `.gitignore` for `ko` and `en`: no English skeleton words in `ko`, no Hangul outside markers in `en`; AGENTS.md heading is the project id; `.gitignore` has nothing outside its block)
- AC-4: met — `the_plan_reads_unchanged_right_after_the_apply` passes (every row `unchanged`, `ko` and `en`)
- AC-5: met — `review_section_cli` 8, `language_setting` 7, `agents_project_cli` 5 passed, `cargo fmt --check` 0

RED before the fix: `a_new_review_file_has_a_heading_in_the_instruction_language` failed with `# Review instructions` for `instruction_files` `ko` and `every_first_creation_text_follows_its_language` failed with `REVIEW.md in ko holds Review instructions outside its markers` (2 failed, 2 passed).
Mutation 6 mutants, 6 killed: K1 heading always English, K2 always Korean, K3 request language only, K4 argument ignored at the call, K5 Korean wording changed, K6 code `ja` not falling back.
Repair 2026-11-19 (bundle 3 verification): `tests/skeleton_language.rs` ran the binary with no HOME, so `home_isolation_scan` `no_new_test_runs_the_binary_without_a_home` failed; `cli()` now runs through `TempHome::command` (clean environment, temporary HOME). After it: `skeleton_language` 4 passed, `home_isolation_scan` 18 passed, `cargo fmt --check` 0, `cargo clippy --locked --offline --all-targets -- -D warnings` 0.
Not run: the full local `cargo test`, `project-update-plan` itself (the plan rows of `agents-plan` are checked instead), Linux, Windows.
