---
id: SPEC-ROLE-NAMING-CODE-001
title: "Role naming unification, code and CLI layer — leader and lane"
version: "0.2.0"
status: draft
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/kanban, internal/hook, internal/factorymsg, internal/homestate, internal/config"
lifecycle: spec-anchored
tags: "factory, kanban, naming, leader, lane, worker, agent, no-alias, rejection, i18n, card-t1256"
tier: L
card: t1256
related_specs: [SPEC-FACTORY-WORKER-NAMING-001, SPEC-FACTORY-WORKER-FANOUT-001, SPEC-CODEX-FACTORY-RETIRE-001, SPEC-AUTONOMY-PRECONDITION-001]
---

# SPEC-ROLE-NAMING-CODE-001 — Role naming unification (code + CLI layer)

## HISTORY

| Version | Date | Author | Change |
|---|---|---|---|
| 0.1.0 | 2026-09-26 | manager-spec | Initial plan-phase draft for card t1256 (operator directive 2026-09-26, Tier L, class C). Census `.moai/reports/t1256/census.md`, conflict map `.moai/reports/t1256/conflicts.md`. |
| 0.2.0 | 2026-09-26 | manager-spec | Operator decisions relayed by the leader applied (plan.md §B). `lane` canonical; legacy `worker`/`agent` spellings — and, by the same no-alias rule, `lead` — are **rejected**, not aliased (REQ-RNC-003/-005/-007/-013 rewritten, REQ-RNC-020 now tests rejection). Persisted values follow the new vocabulary with a run-boundary rule instead of read-both (REQ-RNC-009/-010 rewritten, REQ-RNC-022 new). Role-marker guard accepts `lane` only (REQ-RNC-012). Lane self-dispatch help text (REQ-RNC-021 new), homonym qualifiers (REQ-RNC-023 new). t1193 demoted from hard precondition to a recorded dependency (REQ-RNC-014). `manager-lead` kept by operator decision (REQ-RNC-017). |

**Authority for the vocabulary change.** This SPEC reverses the worker-canonical contract that SPEC-FACTORY-WORKER-NAMING-001 (card t1085) established on 2026-09-22 and that commit `2a3af0c1e` (card t1193) restored on 2026-09-26 on the ground that a human decision had been reversed without approval. The authority here is the operator directive recorded in card t1256 (added 2026-09-26T06:41Z) and the operator's direct answers of 2026-09-26 relayed by the leader (plan.md §B O0: `lane` canonical, `worker`/`agent` aliases removed immediately, no deprecation path, reversal of t1085 approved). No run-phase edit precedes Implementation Kickoff Approval.

## §A Background

Factory Mode (`-f`) and Kanban Mode (`-k`) currently name their roles with four nouns that overlap: the managing session is `lead` in labels, CLI text, and English notices, but `leader` in socket wording and in the Korean, Japanese, and Chinese notices; a card-processing session is `worker` in the CLI token and labels, `agent` in a still-accepted legacy token, and `lane` in the persisted role key, the web dashboard, the handoff protocol, and every CJK notice. The census (`.moai/reports/t1256/census.md` §2) counts 1,261 rename-candidate occurrences in 88 production files and 2,584 in 202 test files.

The operator model for this SPEC: the **leader** manages and operates several lanes; each **lane** takes a card — by self-dispatch, up to and including promoting a queued card, or by the leader's dispatch — and carries it through plan, run, and sync. Two nouns, one meaning each, and no second spelling for either.

This card is layer A (code and CLI). Layer B (documentation, templates, rules, agent descriptions, docs-site, README) is sibling card t1257, which consumes the canonical term table in `design.md` §3 and owns the amendment of the documented promotion clause.

## §B Requirements (GEARS)

### B.1 Vocabulary and input surfaces

- **REQ-RNC-001** (Ubiquitous) — The factory and kanban command surfaces shall present exactly two role nouns on every surface they produce — help and usage text, error messages, SessionStart notices, doctor output, and dashboard labels: `leader` for the session that manages a run and `lane` for a session that processes cards.
- **REQ-RNC-002** (Event-driven) — When a factory launch receives the role token `lane`, the launcher shall join the running factory as the next free numbered lane, labelled `lane-<n>`.
- **REQ-RNC-003** (Event-driven) — When a factory launch receives the role token `worker` or `agent`, the launcher shall launch no session, write no registry, broker, or role-declaration record, print one error line that names `-f lane` as the canonical form, and exit with the same non-zero status it returns for any other invalid `-f` value.
- **REQ-RNC-004** (Ubiquitous) — Every lane label the launcher produces shall have the shape `lane-<n>`, and no launcher input path shall accept `worker-<n>` or `agent-<n>` as a lane label.
- **REQ-RNC-005** (Event-driven) — When an operator supplies a lane label as `worker-<n>` or `agent-<n>` through `-f` or `--name`, the launcher shall launch no session, write no record, print one error line naming `lane-<n>` with the same `<n>`, and exit non-zero.
- **REQ-RNC-006** (Ubiquitous) — The leader session of a kanban or factory run shall be launched under the bare label `leader`, and its collision-bumped and run-id forms shall be `leader-<n>` and `leader-<run-id>`.
- **REQ-RNC-007** (Event-driven) — When an operator supplies the leader label as `lead` or `lead-<suffix>`, the launcher shall launch no session, write no record, print one error line naming the corresponding `leader` or `leader-<suffix>` form, and exit non-zero.

### B.2 Persisted state, environment, broker

- **REQ-RNC-008** (Ubiquitous) — The persisted state schema shall not change: every table name, column name, `CREATE TABLE` body, `ALTER TABLE` statement, and index statement in the factory database and the factory message broker database shall be byte-identical before and after this SPEC.
- **REQ-RNC-009** (Ubiquitous) — Every reader that interprets a persisted role, slot, or label value as a role shall recognize only `leader` and `leader-<suffix>` as the leader and only `lane` and `lane-<n>` as a lane; a legacy value (`lead`, `lead-<suffix>`, `worker`, `agent`, `worker-<n>`, `agent-<n>`) shall not be mapped to either role.
- **REQ-RNC-010** (Ubiquitous) — Writers shall write the new vocabulary into every persisted role key, slot key, session label, registry key, and factory card owner value they create — `leader` for the leader role and slot, `lane` for the lane role, `lane-<n>` for a lane slot and label — and shall not rewrite any record that already exists.
- **REQ-RNC-011** (Ubiquitous) — The environment variable names that carry the leader address, the leader name, the factory signal, and the lane label shall stay unchanged, with no second name added alongside them, and the Codex MCP `env_vars` allowlist generated for existing projects shall stay unchanged; the values these variables carry shall follow REQ-RNC-006 and REQ-RNC-004, and any user-facing text that names one of these variables shall describe its meaning in the leader/lane vocabulary.
- **REQ-RNC-012** (Capability gate) — **Where** the factory role marker environment variable exists in the tree, the lane role value shall be `lane`; the role guard shall treat only `lane` as the lane marker; the site that stamps the marker, the guard that reads it, the `-f` role token, and the lane-label prefix shall reference one shared value definition; and no production code shall stamp `worker` or `agent` into the marker.
- **REQ-RNC-013** (Event-driven) — When a factory message addresses `leader` or `lane-<n>`, the broker shall deliver it to that endpoint; **when** a factory message addresses a legacy slot (`lead`, `worker-<n>`, `agent-<n>`, or the role inputs `worker`/`agent`), the broker shall deliver nothing and return an error naming the canonical slot.
- **REQ-RNC-022** (Event-driven) — When a leader launch, a lane join, or a factory or kanban hook of the post-change binary encounters a live record of its own run that carries a legacy role, slot, or label value, it shall not adopt, rewrite, or number around the record, shall surface one message naming the legacy value, the run, and the retire-and-relaunch step, and — for a launch or join — shall exit non-zero; a dead legacy record shall be treated as stale exactly as a dead record of the new vocabulary is, and historical provenance rows (factory card owner and event history) shall be kept and displayed exactly as recorded.

### B.3 Locales, boundaries, dependencies

- **REQ-RNC-014** (Event-driven) — When the run phase starts, it shall read the develop tree and record in `progress.md` the develop SHA, whether card t1242's deletion of the codex kanban and codex factory entry files has landed, whether card t1245's role-marker constants exist, and card t1193's state; **when** the t1242 deletion has not landed, it shall halt every milestone that edits a file t1242 deletes, with a blocker report to the leader.
- **REQ-RNC-015** (Ubiquitous) — The SessionStart notices for Factory Mode and Kanban Mode shall use the canonical term table (`design.md` §3) in all four locales — en, ko, ja, zh — in one change, with no locale left on `worker`, `agent`, or a mixed `lead`/`leader` wording.
- **REQ-RNC-016** (Unwanted) — The change shall not rename any token used in a sense other than the factory or kanban role: Claude Code agent and subagent vocabulary, the `Agent` tool name, hook input fields for agent type and name, the `moai agent` command group, the `--agent-name` flag, the session-message broker's agent kind, the Codex `agent_role` field, the Claude Agent Teams `leadSessionId` field, the CG-mode leader, goroutine worker pools, and SPEC identifiers.
- **REQ-RNC-017** (Unwanted) — The change shall not rename the `manager-lead` agent or any code that keys on its name (operator decision Q4: the name is kept).
- **REQ-RNC-018** (Ubiquitous) — A guard test shall fail when a production user-facing string presents `worker` or `agent` as a factory role, or `lead` as the leader noun, outside an explicit allowlist that contains only the legacy-value literals the rejection and stale-record paths compare against and the text of their error messages.
- **REQ-RNC-019** (Ubiquitous) — Go identifiers and comments that name the factory or kanban role shall use `leader` and `lane`, except the retained environment variable name constants of REQ-RNC-011 and the legacy-value literals REQ-RNC-018 allowlists, each of which shall say in a comment that it exists only to be refused or detected.
- **REQ-RNC-020** (Ubiquitous) — Test coverage shall retain at least one test per legacy spelling (`-f worker`, `-f agent`, `worker-<n>`, `agent-<n>`, `lead`, `lead-<suffix>`) proving it is refused: non-zero exit, an error naming the canonical form, and no record written.
- **REQ-RNC-021** (Ubiquitous) — The queue pick command's help text shall state, in leader/lane vocabulary, that a queued card is promoted either by the operator's pick through the leader or by a lane's self-dispatch; the change shall add no role guard to the pick path and remove none.
- **REQ-RNC-023** (Ubiquitous) — Every production user-facing string that names the CG-mode leader or the Claude Agent Teams lead shall carry its qualifier (`CG` or `team`) next to the noun, so it cannot be read as the factory or kanban leader, without renaming the underlying identifier or field.

## §C Success Criteria

Acceptance criteria, Given-When-Then scenarios, edge cases, and closure gates: `acceptance.md`. Traceability: every REQ-RNC-0NN maps to at least one AC-RNC-0NN (matrix in `acceptance.md` §C).

## §D Exclusions

### Out of Scope — documentation and template text
- Rules, skills, agent descriptions, output styles, `CLAUDE.md`, `AGENTS.md`, `CLAUDE.local.md`, docs-site (4 locales), and README (4 files) — sibling card t1257.
- Amending the documented HARD promotion clause for lane self-dispatch — t1257 (this SPEC changes only the pick command's help text, REQ-RNC-021).
- Everything under `internal/template/templates/**`, including the `.codex` agent TOML emitted from it.

### Out of Scope — on-disk format changes
- Renaming the `workers` table, the `runs.lead_*` columns, the `legacy_workers_imported` meta key, the `leads.json` registry file, or any broker table.
- Rewriting existing persisted rows to new values (no data migration; REQ-RNC-010, REQ-RNC-022).

### Out of Scope — environment variable renames
- Renaming, or adding a second name for, `MOAI_KANBAN_LEAD_ADDR`, `MOAI_KANBAN_LEAD_NAME`, `MOAI_FACTORY_WORKER`, `MOAI_FACTORY_WORKERS` (REQ-RNC-011; plan.md §B O1).

### Out of Scope — adjacent names and cards
- Renaming `manager-lead` (operator decision Q4) or the CG-mode / Agent Teams homonyms (operator decision Q5 — qualifiers only, REQ-RNC-023).
- Creating the factory role marker variable or its guard (card t1245 owns it; this SPEC only converts its value per REQ-RNC-012).
- The F2 self-dispatch mode (card t1240) and the codex factory retirement (card t1242).
- Broker safety repairs on PR #1722 (card t1193).

### Out of Scope — release activities
- Version bump, CHANGELOG, tagging, and the release PR.
