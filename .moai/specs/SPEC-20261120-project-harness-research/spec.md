---
id: SPEC-20261120-project-harness-research
title: "project asks for a researched harness: agents and skills proposed from the detected stack and current sources, with no fixed count"
status: draft
revision: 1
created: 2026-11-20
---

# project asks for a researched harness: agents and skills proposed from the detected stack and current sources, with no fixed count

## Intent

Step 6 of the project route offers at most three project agents and then project skills "the same way", built only from the interview answers. A project's useful harness is not three: it depends on the stack. This change makes step 6 study the detected project (languages, frameworks, test and build tools, CI, folders), search the web for current practice and well-regarded sources for that stack, and propose the agents and skills worth creating, each with its reason and the sources it rests on (URL and read time). The user picks what to create from one selection with the recommended ones first, and nothing is created without the pick. The proposals stay generic and portable and reuse the existing `generation-plan` and `generation-apply` contract. The files that carry the text (`project.md`, the manifest) are held by the card that rewrites the project and update texts, so the text change follows that card; the SPEC, the research notes of this folder and the request limit of `generation-plan` do not wait.

## Requirements

- REQ-1: The project route shall tell the agent to study the detected project, its languages, frameworks, test and build tools, CI and folders, before it proposes any agent or skill.
- REQ-2: WHEN web search is available, the project route shall tell the agent to search for current practice and well-regarded sources for the detected stack and to name each source, with its URL and read time, next to the proposal it supports. · failure: none
- REQ-3: The project route shall propose only what the stack and the sources justify, naming no count limit of project agents or project skills.
- REQ-4: Each proposal shall name its purpose, the tools or trigger it needs, its scope, the reason it is proposed and its sources.
- REQ-5: The project route shall offer the proposals as one several-of selection with the recommended ones first, creating nothing that the user did not pick.
- REQ-6: The proposals shall stay generic and portable, naming no internal tool of this project, no secret and no absolute home path.
- REQ-7 (REQ-2): IF web search is unavailable, THEN the project route shall tell the agent to propose from the detected facts alone and to say that no source was read.
- REQ-8: The `generation-plan` operation shall accept up to sixteen agents and sixteen skills in one request.
- REQ-9 (REQ-8): IF a request has more agents or skills than the limit or a name that is not safe, THEN `generation-plan` shall refuse it with the name of the field and the limit, and never with a value of the request.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7): step 6 of `skills/moai/workflows/project.md` carries the study, the web search with a source per proposal, the missing-web fallback, the fields of a proposal, the several-of selection and the portability rule, and no longer says "at most three"; the scan tests and `skill_references` pass · verify: `cargo test --locked --offline --test project_harness_step`; `cargo test --locked --offline --test skill_text_citations --test product_cli`; `python3 -m unittest tests.skill_references_test tests.source_manifest_test`
- AC-2 (REQ-8, REQ-9): `generation-plan` accepts sixteen agents and sixteen skills, and refuses seventeen of either, and an unsafe name, with the field named and no value shown · verify: `cargo test --locked --offline --test generation_request_limits`
- AC-3 (REQ-1…REQ-7): a dry run of the step on two fixture projects, a Rust command-line tool and a TypeScript web application, gives different proposals, each with at least one source line (URL and read time); the run and its sources are kept in `research.md` of this folder · verify: manual: read `research.md` of this SPEC and compare the two proposal lists

## Scope

- write: `.moai/specs/SPEC-20261120-project-harness-research/**`
- write (after the card that rewrites the project texts has returned them): `skills/moai/workflows/project.md`, `skills/moai/references/source-manifest.json`, `tests/project_harness_step.rs`
- write (widened when the file is free): `src/generation.rs`, `tests/generation_request_limits.rs`

## Decisions

- D-1: The limit of sixteen agents and sixteen skills stays as a safety bound, not a design count: it exists so one request cannot write an unbounded number of files, it is five times the old advice, and the refusal names it. "No fixed count" is read as no count in the advice the user sees; the leader may lift the bound by one constant.
- D-2: The rebuild record that lets `/moai update` recreate project agents and skills is another card's work (the generation record); this SPEC reuses the existing `generation-plan` and `generation-apply` contract and adds no record.
- D-3: The dry run of AC-3 is a manual criterion: the proposals are the agent's own output from a web search, which a test cannot repeat; the text and the request limit are what the tests pin. The run is recorded in `research.md` with its sources so a reviewer can read it.
- D-4: The official documents of the two hosts that have an agent file (Claude Code subagents and skills, Codex CLI) are the sources the proposal format must stay inside; this SPEC read the Claude Code ones (`research.md`) and did not read the Codex CLI ones, so the Codex file shape stays as `generation-plan` writes it today.

## Plan

1. This card, now: this SPEC, `research.md` (the official formats read, the facts a proposal may rely on) and the fixtures' proposals.
2. Tests first: `tests/generation_request_limits.rs` (red: the refusal names no field today) and `tests/project_harness_step.rs` (red against "at most three"); then `src/generation.rs` (the field named in the refusal), then the text of step 6 after the project-text card has returned `project.md` and the manifest.
3. Whole-area run, clippy, rustfmt, spec check, audit, merge.

## Result

Not yet written.
