---
id: SPEC-ROLE-NAMING-CODE-001
title: "Role naming unification, code and CLI layer — leader and lane"
version: "0.1.0"
status: draft
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/kanban, internal/hook, internal/factorymsg, internal/homestate, internal/config"
lifecycle: spec-anchored
tags: "factory, kanban, naming, leader, lane, worker, agent, compat, alias, i18n, card-t1256"
tier: L
card: t1256
related_specs: [SPEC-FACTORY-WORKER-NAMING-001, SPEC-FACTORY-WORKER-FANOUT-001, SPEC-CODEX-FACTORY-RETIRE-001, SPEC-AUTONOMY-PRECONDITION-001]
---

# SPEC-ROLE-NAMING-CODE-001 — Role naming unification (code + CLI layer)

## HISTORY

| Version | Date | Author | Change |
|---|---|---|---|
| 0.1.0 | 2026-09-26 | manager-spec | Initial plan-phase draft for card t1256 (operator directive 2026-09-26, Tier L, class C). Census `.moai/reports/t1256/census.md`, conflict map `.moai/reports/t1256/conflicts.md`. |

**Authority for the vocabulary change.** This SPEC reverses the worker-canonical contract that SPEC-FACTORY-WORKER-NAMING-001 (card t1085) established on 2026-09-22 and that commit `2a3af0c1e` (card t1193) restored on 2026-09-26 on the ground that a human decision had been reversed without approval. The authority here is the operator directive recorded in card t1256 (added 2026-09-26T06:41Z), to be confirmed at the Implementation Kickoff Approval gate. No run-phase edit precedes that confirmation.

## §A Background

Factory Mode (`-f`) and Kanban Mode (`-k`) currently name their roles with four nouns that overlap: the managing session is `lead` in labels, CLI text, and English notices, but `leader` in socket wording and in the Korean, Japanese, and Chinese notices; a card-processing session is `worker` in the CLI token and labels, `agent` in a still-accepted legacy token, and `lane` in the persisted role key, the web dashboard, the handoff protocol, and every CJK notice. The census (`.moai/reports/t1256/census.md` §2) counts 1,261 rename-candidate occurrences in 88 production files and 2,584 in 202 test files.

The operator model for this SPEC: the **leader** manages and operates several lanes; each **lane** takes a card — by self-dispatch or by the leader's dispatch — and carries it through plan, run, and sync. Two nouns, one meaning each.

This card is layer A (code and CLI). Layer B (documentation, templates, rules, agent descriptions, docs-site, README) is sibling card t1257, which consumes the canonical term table in `design.md` §3.

## §B Requirements (GEARS)

### B.1 Vocabulary on user-facing surfaces

- **REQ-RNC-001** (Ubiquitous) — The factory and kanban command surfaces shall present exactly two role nouns on every surface they produce — help and usage text, error messages, deprecation hints, SessionStart notices, doctor output, and dashboard labels: `leader` for the session that manages a run and `lane` for a session that processes cards.
- **REQ-RNC-002** (Event-driven) — When a factory launch receives the role token `lane`, the launcher shall join the running factory as the next free numbered lane, labelled `lane-<n>`.
- **REQ-RNC-003** (Event-driven) — When a factory launch receives the role token `worker` or `agent`, the launcher shall behave exactly as for `lane` and shall print one deprecation line that names `-f lane` as the replacement.
- **REQ-RNC-004** (Ubiquitous) — Every lane label the launcher produces shall have the shape `lane-<n>`; the labels `worker-<n>` and `agent-<n>` shall remain readable and shall share the one lane number space, so a live claim under any of the three spellings holds its number.
- **REQ-RNC-005** (Event-driven) — When an operator supplies a lane label in a legacy spelling (`worker-<n>` or `agent-<n>`), the launcher shall launch under the canonical `lane-<n>` form and shall print one deprecation line naming that form.
- **REQ-RNC-006** (Ubiquitous) — The leader session of a kanban or factory run shall be launched under the bare label `leader`, and its collision-bumped and run-id forms shall be `leader-<n>` and `leader-<run-id>`.
- **REQ-RNC-007** (Event-driven) — When an operator supplies the leader label as `lead` or `lead-<suffix>`, the launcher shall accept it as the leader label, launch under the corresponding `leader` form, and print one deprecation line; `lead` and `leader` shall share one name namespace, so a live session holding `lead` holds the bare leader name.

### B.2 Compatibility and persisted state

- **REQ-RNC-008** (Ubiquitous) — The persisted state schema shall not change: every table name, column name, `CREATE TABLE` body, `ALTER TABLE` statement, and index statement in the factory database and the factory message broker database shall be byte-identical before and after this SPEC.
- **REQ-RNC-009** (Event-driven) — When any reader encounters a persisted role or slot value written in either vocabulary, it shall resolve `lead` and `leader` to the leader role, and `worker`, `agent`, and `lane` to the lane role, with identical downstream behavior.
- **REQ-RNC-010** (Ubiquitous) — Writers shall keep writing the persisted role and slot key values that the current release writes (the leader key `lead`; the broker lane role `worker`; the role-declaration lane key `lane`), so that a binary from the current release can still read state written after this SPEC. Session labels are identity values, not role keys, and follow REQ-RNC-004 and REQ-RNC-006.
- **REQ-RNC-011** (Ubiquitous) — The environment variable names that carry the leader address, the leader name, the factory signal, and the lane label shall stay unchanged, and the Codex MCP `env_vars` allowlist generated for existing projects shall stay unchanged; any user-facing text that names one of these variables shall describe its meaning in the leader/lane vocabulary.
- **REQ-RNC-012** (Capability gate) — **Where** the factory role marker environment variable exists in the tree, the lane role value shall be `lane`; the role guard shall treat `lane`, `worker`, and `agent` all as the lane marker; and the site that stamps the marker and the guard that reads it shall reference one shared value definition, so the two cannot drift apart silently.
- **REQ-RNC-013** (Event-driven) — When a factory message addresses the leader as `leader` or a lane in any of the three lane spellings, the broker shall deliver it to the same endpoint it would have reached under the current spelling.

### B.3 Locales, boundaries, and preconditions

- **REQ-RNC-014** (Event-driven) — When the run phase starts, it shall read the develop tree and record in `progress.md` whether card t1242's deletion of the codex kanban and codex factory entry files has landed and whether card t1193 is merged or closed; **when** either is unresolved, it shall halt every milestone that edits a file those cards change, with a blocker report to the leader.
- **REQ-RNC-015** (Ubiquitous) — The SessionStart notices for Factory Mode and Kanban Mode shall use the canonical term table (`design.md` §3) in all four locales — en, ko, ja, zh — in one change, with no locale left on `worker`, `agent`, or a mixed `lead`/`leader` wording.
- **REQ-RNC-016** (Unwanted) — The change shall not rename any token used in a sense other than the factory or kanban role: Claude Code agent and subagent vocabulary, the `Agent` tool name, hook input fields for agent type and name, the `moai agent` command group, the `--agent-name` flag, the session-message broker's agent kind, the Codex `agent_role` field, the Claude Agent Teams `leadSessionId` field, the CG-mode teammate leader, goroutine worker pools, and SPEC identifiers.
- **REQ-RNC-017** (Unwanted) — The change shall not rename the `manager-lead` agent or any code that keys on its name; that decision belongs to card t1257, and this SPEC follows it only after t1257 records it.
- **REQ-RNC-018** (Ubiquitous) — A guard test shall fail when a production user-facing string presents `worker` or `agent` as a factory role, or `lead` as the leader noun, outside an explicit allowlist that contains only the legacy-spelling parse values and the deprecation-hint text.
- **REQ-RNC-019** (Ubiquitous) — Go identifiers and comments that name the factory or kanban role shall use `leader` and `lane`, except the persisted-key value definitions retained by REQ-RNC-010, which shall say in a comment that the value is an on-disk key and not notation.
- **REQ-RNC-020** (Ubiquitous) — Test coverage shall retain at least one test per legacy spelling (`-f worker`, `-f agent`, `worker-<n>`, `agent-<n>`, `lead`, `lead-<suffix>`) proving it still parses and prints its hint.

## §C Success Criteria

Acceptance criteria, Given-When-Then scenarios, edge cases, and closure gates: `acceptance.md`. Traceability: every REQ-RNC-0NN maps to at least one AC-RNC-0NN (matrix in `acceptance.md` §C).

## §D Exclusions

### Out of Scope — documentation and template text
- Rules, skills, agent descriptions, output styles, `CLAUDE.md`, `AGENTS.md`, `CLAUDE.local.md`, docs-site (4 locales), and README (4 files) — sibling card t1257.
- Everything under `internal/template/templates/**`, including the `.codex` agent TOML emitted from it.

### Out of Scope — renames that change on-disk format
- Renaming the `workers` table, the `runs.lead_*` columns, the `legacy_workers_imported` meta key, the `leads.json` registry file, or any broker table.
- Rewriting existing persisted rows to new values (no data migration).

### Out of Scope — removing the legacy spellings
- Removing `-f worker`, `-f agent`, `worker-<n>`, `agent-<n>`, `lane`-era hints, or `lead` acceptance. Removal is a separate, measured card after at least one minor release.
- Changing environment variable names (REQ-RNC-011 keeps them; plan.md open decision O1 records the alternative).

### Out of Scope — adjacent cards' own work
- Creating the factory role marker variable or its guard (card t1245 owns it; this SPEC only fixes its value per REQ-RNC-012).
- The F2 self-dispatch mode (card t1240) and the codex factory retirement (card t1242).
- Broker safety repairs on PR #1722 (card t1193).

### Out of Scope — release activities
- Version bump, CHANGELOG, tagging, and the release PR.
