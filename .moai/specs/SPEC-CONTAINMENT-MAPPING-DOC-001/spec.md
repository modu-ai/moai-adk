---
id: SPEC-CONTAINMENT-MAPPING-DOC-001
title: "Multi-agent containment and source-control mapping reference document"
version: "0.1.0"
status: completed
created: 2026-10-16
updated: 2026-10-17
author: manager-spec
priority: P2
phase: "v3.2.0"
module: ".moai/docs"
lifecycle: spec-anchored
tags: "documentation, reference-only, multi-agent, containment, source-control, factory"
tier: M
---

# SPEC-CONTAINMENT-MAPPING-DOC-001: Multi-agent containment and source-control mapping reference document

## HISTORY

- 0.1.0: Initial draft. Plan-phase artifacts (Tier M: spec, plan, acceptance, plus the progress skeleton).
- 0.1.0 (plan-audit fix round 1): REQ-CMM-007 scoped to eight distinct paths so that it matches the revised acceptance command; the acceptance commands for the source register, the mapping table, the path check, the forbidden-token check, and the scope guard were rebuilt (see `plan.md` § I).
- 0.1.0 (plan-audit fix round 2): REQ-CMM-001 now names the card-evidence directory `.moai/reports/` as a third allowed location, so that it carries the same scope as the revised scope-guard criterion (AC-CMM-010 excludes that directory because card evidence is tracked there); the forbidden-token check gained full and abbreviated English month names in both orders and any case (see `plan.md` § J).
- 0.1.0 (plan-audit fix round 3): REQ-CMM-001 narrowed from the whole `.moai/reports/` tree to this card's own evidence directory `.moai/reports/t1760/`, so that it carries exactly the scope of the narrowed scope-guard pathspec (AC-CMM-010); the date-check criterion AC-CMM-008 was reworded to what its command matches (see `plan.md` § J.1).
- 0.1.0 (plan-audit fix round 4): REQ-CMM-001 now binds tracked or non-ignored files, because `.gitignore` ignores `.moai/reports/*` and the scope guard (AC-CMM-010) cannot see a new untracked file there; the criterion's claim about other cards' report directories was narrowed to the same scope and the ignored-file limit stated (see `plan.md` § J.2).

## 1. Purpose

### 1.1 Why

Published engineering writing on multi-agent systems makes three points the factory already has to live with: components of an agent system should be able to fail independently of one another, boundaries belong at the environment layer rather than in prompts, and a merge queue stops making sense once agents merge at very high volume. The repository holds no document that sets these points beside the defenses the factory actually carries. Measured in this tree: none of the three anthropic.com engineering URLs in § 2 appears in `.moai/docs`, `.claude/rules`, `.claude/skills`, `AGENTS.md`, or `docs-site/content`, and no `.moai/docs` filename names containment, mapping, or multi-agent.

### 1.2 Deliverable

One reference-only markdown document, `.moai/docs/multi-agent-containment-mapping.md`, holding a source register and a mapping table. It is a comparison. It recommends no change to any rule, configuration, code, or template, and it makes no claim that the factory should adopt any external approach.

### 1.3 Where it lives

`.moai/docs` is tracked in the repository, but only part of it is deployed to user projects through `internal/template/templates/.moai/docs/`. The three nearest siblings (the git workflow doctrine, the integration-chain procedure, and the managed-session guide) have no copy in that template directory, and the integration-chain document declares itself local-only. This document is local-development material in the same way, so it gets no Template-First mirror. Command evidence: `plan.md` § E.

## 2. Verified source basis

Four external sources are in scope. The identifiers are stable labels used by the document; gaps in the numbering are intentional because only these four sources belong to this deliverable.

Verification method, applied to every row below in the planning run: the URL was fetched with WebFetch, the page was downloaded raw, and each quoted fragment was matched as a fixed string against the raw bytes (`grep -c -F`, count 1 or more). A fragment is quoted below only when that match succeeded. Per-fragment counts are in `plan.md` § C.

| ID | Title (as shown on the page) | URL | Fetched | Flag |
|---|---|---|---|---|
| BP-5 | Redesigning source control at Anthropic | https://www.anthropic.com/engineering/redesigning-source-control | yes | VERIFIED |
| BP-6 | How we contain Claude across products | https://www.anthropic.com/engineering/how-we-contain-claude | yes | VERIFIED |
| BP-1 | Scaling Managed Agents: Decoupling the brain from the hands | https://www.anthropic.com/engineering/managed-agents | yes | VERIFIED |
| BP-2 | Prompting Claude Opus 5.5 | https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/prompting-claude-opus-5-5 | yes | VERIFIED |

Quoted fragments observed on the fetched pages (verbatim):

- BP-5: "source control became our biggest development bottleneck"; "around 2,500 code changes merged per day"; "we merged over 30,000 per day"; "A proper merge queue, for example"; "much sense at agent scale"; "global repository lock on the merge queue"; "post-merge and pre-deploy automated testing"; "concurrent changes can conflict"; "Claude can typically fix these conflicts faster than humans".
- BP-6: "users approved roughly 93% of permission prompts"; "Design for containment at the environment layer first, then steer behavior at the model layer"; "The goal is to set a hard boundary on what an agent can reach".
- BP-1: "Decoupling the brain from the hands meant the harness no longer lived inside the container"; "each could fail or be replaced independently"; "the tokens are never reachable from the sandbox where Claude"; "nothing in the harness needs to survive a crash".
- BP-2: "Treat a text-only end of turn as a report rather than as proof the task is done"; "stop after two or three automatic continuations on the same task".

The flags above describe the planning run. The document author re-verifies at authoring time (REQ-CMM-004) and records the flag from that run. A source that cannot be fetched or whose fragment cannot be matched is carried as UNVERIFIED, never as established.

## 3. Requirements

Notation: GEARS. The mapping table, source register, and flag tokens are fixed English tokens so that the acceptance commands can key on them.

- **REQ-CMM-001** (Ubiquitous): The deliverable shall be a single markdown document at `.moai/docs/multi-agent-containment-mapping.md`, and the run phase shall add, change, or remove no tracked or non-ignored file outside that document, this SPEC directory, and this card's own evidence directory `.moai/reports/t1760/`.
- **REQ-CMM-002** (Ubiquitous): The document shall state in its opening section that it is reference-only, using the literal token `reference-only`: a comparison of published external practice with defenses already present in the repository, carrying no recommendation to change any rule, configuration, code, or template.
- **REQ-CMM-003** (Ubiquitous): The document shall carry a source register with one row for each of BP-1, BP-2, BP-5, and BP-6, each row giving the source title, its URL, one double-quoted fragment of at least 20 characters copied verbatim from the page, and, as its final cell, the flag `VERIFIED` or `UNVERIFIED`.
- **REQ-CMM-004** (Event-driven): When the author assigns `VERIFIED` to a source, the author shall have fetched that URL during the authoring run and observed the quoted fragment on the fetched page; where either condition is not met, the author shall assign `UNVERIFIED`.
- **REQ-CMM-005** (State-driven): While a source's flag is `UNVERIFIED`, the document shall present every mapping row that cites that source as a lead awaiting confirmation and shall not present its claim as established.
- **REQ-CMM-006** (Ubiquitous): The document shall contain one mapping table with the columns Theme, Source, Source claim, MoAI defense, and Fit, covering the themes `correlated-failure`, `isolation-first`, and `merge-queue-alternative` with at least one row per theme; the Fit cell shall be the last cell of its row and shall be exactly one of `aligned`, `differs`, or `gap`.
- **REQ-CMM-007** (Ubiquitous): The MoAI-defense cell of every non-gap row shall name a repository path and a section of that file in the form `path § section`, for a file the author opened during the authoring run; every cited path shall resolve in the tree, and the document shall cite at least eight distinct paths in total.
- **REQ-CMM-008** (Event-driven): When no MoAI defense matching a source claim is found among the files the author opened, the document shall record a row whose Fit is `gap` and shall not synthesize a match; a gap row states that no defense was located, which is not a statement that none exists.
- **REQ-CMM-009** (Ubiquitous): The document shall attribute the source's volume figures and its post-merge-testing conclusion to the source, shall frame the `merge-queue-alternative` comparison as one across differing scales, and shall not assert that the factory should adopt the source's approach.
- **REQ-CMM-010** (Unwanted): The document shall not contain card identifiers, SPEC identifiers, calendar years or dates, the rule-marker tokens `[HARD]` or `[ZONE`, or the word `develop`; defense cells shall paraphrase and point to repository text and shall not restate it.
- **REQ-CMM-011** (Ubiquitous): The document shall remain local-development material, and no Template-First mirror of it shall be created under `internal/template/templates/`.

Traceability to acceptance criteria is in `acceptance.md` § Traceability.

## 4. Constraints

- Reference-only: the document introduces no new rule and no new `[HARD]` clause, and it edits no existing rule, configuration key, code path, or template.
- Scale: the merge-queue comparison is a comparison. The source reports its own volume; no measurement of the factory's merge volume was opened, so the document makes no scale claim about the factory.
- Evidence: every claim about a MoAI defense rests on a file the author opened in the authoring run. An unread file supports no row.
- Language: the document body follows the project's conversation language setting, with the fixed tokens, identifiers, paths, URLs, and verbatim source fragments kept in their original form (decision recorded in `plan.md` § B).

## 5. Exclusions

### Out of Scope — Rule, configuration, code, and template changes

- Editing any file under `.claude/rules/`, `.claude/skills/`, `.claude/agents/`, `.moai/config/`, `internal/`, or `AGENTS.md`.
- Adding a new rule, a new `[HARD]` clause, a new configuration key, or a new check script.

### Out of Scope — Recommendations to change the factory

- Proposing that the factory adopt post-merge testing, drop its integration window, or change any guard's default.
- Deciding whether any `gap` row should be closed. Closing a gap is a separate decision for a separate card.

### Out of Scope — Template-First mirror

- Creating `internal/template/templates/.moai/docs/multi-agent-containment-mapping.md` or any generated embed entry for it.

### Out of Scope — Further sources

- Any external source beyond the four in § 2, including secondary summaries and unofficial figures.
- Quantitative claims about the factory's own merge volume, lane count effects, or review latency.
