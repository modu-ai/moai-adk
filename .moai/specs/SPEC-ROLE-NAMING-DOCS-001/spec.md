---
id: SPEC-ROLE-NAMING-DOCS-001
title: "Role naming — unify Kanban/Factory role vocabulary to leader · lane in the document layer"
version: "0.1.0"
status: draft
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/template/templates/.claude, .claude/rules, .claude/agents/moai, .claude/skills, .claude/output-styles, docs-site/content, README"
lifecycle: spec-anchored
tags: "naming, leader, lane, kanban, factory, docs, i18n, template-mirror, card-t1257"
tier: L
card: t1257
related_specs: [SPEC-FACTORY-WORKER-NAMING-001, SPEC-KANBAN-RENAME-001, SPEC-LANE-PROVIDER-AXIS-001, SPEC-HIERARCHICAL-TEAM-001]
---

# SPEC-ROLE-NAMING-DOCS-001 — Role naming, document layer

## HISTORY

- **v0.1.0** (2026-09-26) — manager-spec — initial plan-phase draft for card t1257 (Tier L, class C). Measurements taken in worktree `.claude/worktrees/t1257` at HEAD `e62c3e183`; the inventory is `.moai/reports/t1257/inventory.md`. Substitution is gated behind the code-layer card t1256 (§C, REQ-RND-002).

## §A Background

The operator directive of 2026-09-26 unifies the Kanban Mode (`moai cc -k`) and Factory Mode (`moai cc -f`) role vocabulary to two words in the document layer:

- **leader** — the single session that manages and operates multiple lanes.
- **lane** — a session that either self-dispatches a card or receives the leader's dispatch, and carries that card through plan > run > sync.

The document layer today uses at least eight role words for these two roles and their sub-roles: `lead`, `leader`, `lane`, `worker`, `companion`, `foreman`, `deputy`, and `coordinator` (inventory §2.1). Three of the target words already carry other meanings in the same documents: `leader` is the `moai cg` Claude leader pane and the Agent Teams leader (57 of 87 occurrences), `lane` is the "Lane A / Lane B" command batch and the "Epic N Lane A" schedule track, and `worker` is the `manager-lead` leaf worker (a subagent) (inventory §4).

Three days before this SPEC, SPEC-FACTORY-WORKER-NAMING-001 (card t1085, completed) and its public-docs follow-up (card t1102) renamed the Factory session notation in the opposite direction — `lane-N` → `worker-N`, `-f agent` → `-f worker` — in code, tests, 4-locale i18n strings, docs-site, and READMEs, and kept `-f lane-<n>` as a deprecated alias. The code therefore emits `worker-N` today. A document-only move to `lane-N` would make every such line false until the code layer moves with it.

The in-scope surfaces, measured in the inventory: 376 files carrying at least one role token — template `.claude/**` (100 files), C3 Codex TOML (10), template `.moai` and root (5), local `.claude/**` (115), root `CLAUDE.md` / `AGENTS.md` / `CLAUDE.local.md` (3), local `.moai/config` (2), docs-site in four locales (137 page-locale files), and the four READMEs.

## §B Requirements (GEARS)

- **REQ-RND-001** (Ubiquitous): The document layer shall name the single coordinating session of Kanban and Factory Mode `leader` and each card-carrying session `lane`, on every in-scope surface, once the substitution milestones complete.
- **REQ-RND-002** (Event-driven): **When** the run phase reaches any substitution milestone, it shall first verify the gate — the t1256 code-layer conclusion is landed on develop (its SPEC frontmatter reads `implemented` or `completed`, or the operator records the code-layer decision in progress.md) AND the operator answers to research.md §F Q1–Q5 are recorded in progress.md.
- **REQ-RND-003** (Event-detected): **When** the REQ-RND-002 gate is not satisfied, the run phase shall halt every substitution milestone with a blocker report and shall perform no body-text substitution on any in-scope surface.
- **REQ-RND-004** (Ubiquitous): Every identifier-class occurrence in documentation — CLI tokens, session notation, environment variables, sentinels, agent names, and file paths — shall match the form the code accepts at the develop HEAD being edited against.
- **REQ-RND-005** (Unwanted): The run phase shall not substitute an occurrence whose meaning is not the Kanban/Factory role — the `moai cg` leader pane, the Agent Teams lead or leader, "Lane A / Lane B" command batches, "Epic N Lane X" schedule tracks, "detail companion" document pairs, `manager-lead` leaf workers, or plain English.
- **REQ-RND-006** (Ubiquitous): Every substitution shall be recorded in a per-line disposition ledger (path, line, old text, new text or "kept", reason class) committed under `.moai/reports/t1257/`.
- **REQ-RND-007** (Ubiquitous): Every `[HARD]` clause touched by a substitution shall keep its conditions, prohibitions, and authority scope unchanged, and the count of `[HARD]` markers in each touched file shall be equal before and after the change.
- **REQ-RND-008** (Event-driven): **When** a heading or bold-paragraph anchor containing a renamed word changes, every site that references that anchor shall change in the same commit, and a re-run anchor scan shall report zero references to an anchor text that no longer exists.
- **REQ-RND-009** (Ubiquitous): Every change to a file that has a template mirror shall be made first under `internal/template/templates/`, then in the local copy, and `make build` shall run after the template edits.
- **REQ-RND-010** (Where): **Where** a change touches an agent definition under `internal/template/templates/.claude/agents/moai/`, the run phase shall regenerate the Codex copies with `make agents-emit` and `make agents-emit-check` shall exit 0; the files under `internal/template/templates/.codex/agents/moai/` shall not be edited by hand.
- **REQ-RND-011** (Where): **Where** the operator selects option A of research.md §F Q4, the agent identifier `manager-lead` shall stay unchanged and its definition sites shall state that it is the leader's coordination agent; **where** the operator selects option B, C, or D, the agent rename shall be executed by the code-layer work (t1256 or a successor card), and this SPEC shall update only prose after that rename lands.
- **REQ-RND-012** (Ubiquitous): Template edits shall pass the template-neutrality guard — no SPEC IDs, card IDs, or internal dates introduced into `internal/template/templates/**`.
- **REQ-RND-013** (Ubiquitous): The docs-site pages in all four locales (en, ko, ja, zh) and the four README files shall change in the same change set, using one fixed per-locale lexicon for `leader` and `lane` recorded in progress.md before the first docs edit.
- **REQ-RND-014** (Ubiquitous): The docs-site glossary page `core-concepts/kanban-board-terms.md` shall define `leader` and `lane` with the operator's model statement and shall carry one disambiguation line separating the Kanban/Factory leader from the `moai cg` leader pane and the Agent Teams leader, in all four locales.
- **REQ-RND-015** (Ubiquitous): The `CLAUDE.local.md` lane-obligation text (§4.1) shall be edited only in the develop copy carried by the card worktree and merged to develop; the primary checkout's working copy shall not be edited or restored.
- **REQ-RND-016** (Event-detected): **When** a document edit would change a string that a Go test reads and asserts (inventory §7), the run phase shall halt that edit with a blocker report routing the test change to the code-layer card, and this SPEC shall not edit any `.go` file.
- **REQ-RND-017** (Unwanted): The run phase shall not rewrite frozen records — CHANGELOG entries already released, `.moai/specs/**` records of other SPECs, HISTORY sections, and deprecated-alias disclosures that describe the legacy form as legacy.
- **REQ-RND-018** (Unwanted): The run phase shall not change the semantics of the queue-production, promotion, dispatch, or column-ownership clauses — "the lead is the queue's sole producer", "promotion is the operator's act, always", and the per-column companion model — unless the operator's answers to research.md §F Q2/Q3 authorize a model change, in which case that change shall be carried by a separate SPEC.
- **REQ-RND-019** (Where): **Where** a role-bearing file has no template mirror (the local-only files listed in inventory §5), the run phase shall edit the local file only and shall not create a template mirror for it.
- **REQ-RND-020** (Event-driven): **When** the substitution milestones complete, a re-run of the inventory scripts on the edited tree shall report zero `role`-class occurrences of the retired role words (`lead`, `companion`, and `worker` in its Factory-session sense) on the in-scope surfaces, except the entries the disposition ledger lists as kept.

## §C Dependency

- **Code layer first.** This SPEC depends on the code-layer naming decision of card t1256. At authoring time no t1256 SPEC exists: the t1256 worktree holds only `.moai/reports/t1256/raw/{agent,lane,lead,worker}.txt` (checked 2026-09-26), so the dependency is recorded by card id and `depends_on:` is left empty until that SPEC ID exists. The run phase fills `depends_on:` through a mid-run amendment when it does.
- **Relation to SPEC-FACTORY-WORKER-NAMING-001** (completed, card t1085): that SPEC set the current code vocabulary (`worker-N`, `-f worker`, deprecated `-f agent` / `-f lane-<n>` / `--name lane-<n>` aliases). This SPEC may reverse its prose direction only as far as the code-layer decision reverses its identifiers (REQ-RND-004).
- **Relation to SPEC-KANBAN-RENAME-001** (completed): that SPEC renamed Factory Mode to Kanban Mode and set the precedent this SPEC follows — a rename-only SPEC with measured inventories, zero-residue completion greps, and no behavior change.
- **Relation to SPEC-LANE-PROVIDER-AXIS-001** (draft): uses "lane" for the Factory session pool crossed with the provider axis; its vocabulary is consistent with this SPEC's target and needs no change.
- **Relation to SPEC-HIERARCHICAL-TEAM-001** (completed): introduced `manager-lead` as a "leader" of a hierarchical team with leaf workers; its leaf-worker sense of `worker` is one of the meanings REQ-RND-005 protects.

## §D Success Criteria

Acceptance criteria, Given-When-Then scenarios, and closure gates live in `acceptance.md`. Traceability: see acceptance.md §B.

## §E Exclusions

### Out of Scope — code layer

- Any `.go` source or test change, CLI flag, session-notation emission, i18n Go string table, environment variable, or hook behavior. These belong to card t1256 (REQ-RND-016).
- Regenerating or editing test fixtures under `internal/cli/testdata/**`.

### Out of Scope — model change

- Any change to who may produce, promote, or dispatch a card, or to the Kanban per-column companion structure. The operator's "lanes self-dispatch" wording is recorded as research.md §F Q3; a behavior change goes to a separate SPEC (REQ-RND-018).

### Out of Scope — frozen records

- Released CHANGELOG entries, other SPECs' records under `.moai/specs/**`, HISTORY sections, and `.moai/reports/**` of other cards.

### Out of Scope — other meanings

- `moai cg` leader-pane wording, Agent Teams leader wording, command-batch "Lane A/B", Epic "Lane X", "detail companion", and leaf-worker wording (REQ-RND-005).

### Out of Scope — release activities

- No push, PR, tag, or release. Integration follows the git-flow lane protocol; the leader pushes develop in batch.
