---
id: SPEC-FACTORY-WORKER-NAMING-001
title: "Factory worker naming — GTD todo naming closure, agent→worker join token, lane-N→worker-N notation"
version: "0.1.0"
status: completed
created: 2026-09-22
updated: 2026-09-23
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/kanban, internal/hook, internal/template/templates/.claude/rules"
lifecycle: spec-anchored
tags: "factory, worker, naming, lane, agent-token, i18n, gtd, card-t1085"
tier: M
card: t1085
depends_on: [SPEC-FACTORY-MIXED-HOOK-001]
related_specs: [SPEC-FACTORY-MIXED-HOOK-001, SPEC-GTD-AUTONOMY-001, SPEC-GTD-CANON-BODY-001]
---

# SPEC-FACTORY-WORKER-NAMING-001 — Factory worker naming axis

## §A Background and Motivation

Three related naming work items are consolidated into one SPEC (card t1085, Class C):

1. **GTD naming-axis closure (independent, immediate).** Operator decision of 2026-09-22: the `moai todo` command keeps its name — it is NOT renamed to a GTD-branded name. The GTD card family (t855 naming, t867 canon body, t899 landed store, t939–t941, autonomy ×2 — 8 trees total) is closed as investigation-only scope. t855 has zero work commits, so a closure record suffices. The physical DISPOSAL of the old GTD trees belongs to card t1084 and is explicitly out of this SPEC's scope.
2. **Factory roll vocabulary agent→worker (trails SPEC-FACTORY-MIXED-HOOK-001 / card t1074).** The `-f agent` join token and the agent-lane vocabulary rename to worker.
3. **lane-N → worker-N notation (trails SPEC-FACTORY-MIXED-HOOK-001).** All `lane-N` surface notation renames to `worker-N` across code, tests, 4-locale i18n status strings, and the kanban-dispatch rule-doc twins (local + template mirror).

Items 2 and 3 MUST be serialized after SPEC-FACTORY-MIXED-HOOK-001 lands on develop: that in-flight SPEC is actively repairing `internal/factorymsg/` and has already modified the factory CLI sources on its own branch. Renaming vocabulary in parallel would collide with its uncommitted work.

## §B Requirements (GEARS)

- **REQ-001** (Ubiquitous): The GTD naming-axis closure record shall state the operator decision of 2026-09-22 that `moai todo` keeps its name and is NOT renamed to a GTD-branded name, and shall record the GTD card family as investigation-only scope with the t1084 disposal boundary named.
- **REQ-002** (Event-driven): **When** the run phase reaches any agent→worker or lane→worker rename work, it shall first verify the serialization precondition by reading the develop tree: `internal/factorymsg/` exists there AND the `SPEC-FACTORY-MIXED-HOOK-001` frontmatter status read from develop reports landed (`implemented` or `completed`).
- **REQ-003** (Event-detected): **When** the serialization precondition is not satisfied, the run phase shall halt all rename milestones (M3, M4) with a blocker report and shall not perform any rename edit on a tree that lacks the landed SPEC-FACTORY-MIXED-HOOK-001 state.
- **REQ-004** (Ubiquitous): The factory CLI shall present worker-axis vocabulary on every user-facing surface: the `-f worker` join token in place of `-f agent`, and `worker-N` notation in place of `lane-N` / `agent-N` in help text, error messages, and flag documentation.
- **REQ-005** (Ubiquitous): The factory status-surface i18n table shall carry the worker-axis wording in all four locales (en / ko / ja / zh) in lockstep — no locale left on lane or agent wording after the rename.
- **REQ-006** (Ubiquitous): The kanban-dispatch rule-doc twins shall be judged and renamed per file individually — the local copy under `.claude/rules/moai/workflow/` AND the template mirror under `internal/template/templates/.claude/rules/moai/workflow/` are intentionally NOT assumed byte-identical — and every template-mirror edit shall be followed by `make build` so the embedded copy refreshes.
- **REQ-007** (Event-driven): **When** the old-token inventory measurement (milestone M2) completes, the run phase shall execute the compatibility decision — keep-alias vs remove — for `-f agent`, `-f lane-<n>`, and the `--name lane-<n>` join-label form FROM that measurement, and shall record the decision plus its measured basis in progress.md.
- **REQ-008** (Unwanted): The run phase shall not silently destroy the old tokens: no removal of `-f agent`, `-f lane-<n>`, or `lane-N` notation before the M4 measured compatibility decision. An a-priori removal without the M2 inventory is prohibited.
- **REQ-009** (Ubiquitous): The GTD compatibility surface shall be preserved by this SPEC: the `gtd`↔`todo` alias coverage guarded by `gtd_compat_test.go` continues to pass, and the closure record (item 1) removes nothing from that alias surface.
- **REQ-010** (Capability gate): **Where** the serialization precondition of REQ-002 is satisfied, the test files that reference the lane/agent vocabulary (factory, codex-factory, goal-mission, goal-blocked-question, doctor-jev, and gtd-compat test files under `internal/cli/`) shall be updated to the worker vocabulary in the same rename, except any legacy-form coverage explicitly retained by the REQ-007 decision.
- **REQ-011** (Event-driven): **When** the rename milestone (M3) completes, the literal `lane-` token shall be absent from the factory CLI user-facing surfaces and the 4-locale i18n strings, except entries explicitly retained by the REQ-007 compatibility decision and listed in progress.md.

## §C Success Criteria

Acceptance criteria, Given-When-Then scenarios, severity, and closure gates: see `acceptance.md` (§D AC Matrix). Traceability: AC-001→REQ-001, AC-002/003→REQ-002/003, AC-004→REQ-004, AC-005→REQ-005, AC-006→REQ-006, AC-007→REQ-007/008, AC-008→REQ-009, AC-009→REQ-010, AC-010→REQ-011.

## §D Exclusions

### Out of Scope — GTD tree disposal

- The physical disposal of the 8 old GTD work trees is card t1084's concern; this SPEC records the naming closure decision only.

### Out of Scope — kanban companion session vocabulary

- The prose word "lane" as used for Kanban Mode column-companion sessions (non-Factory context) is not renamed by a blanket rule; each documentation line is judged individually during the run phase and the per-line disposition is recorded.

### Out of Scope — mixed-hook repair work

- Any repair, extension, or behavior change to `internal/factorymsg/` or the mixed-hook logic itself: that is SPEC-FACTORY-MIXED-HOOK-001's scope; this SPEC only renames vocabulary after it lands.

### Out of Scope — release activities

- No release cut, push, or PR creation is part of this SPEC; integration follows the git-flow lane protocol.

## §E History

- 2026-09-22 — v0.1.0 — manager-spec — initial plan-phase draft (card t1085, Tier M + card-mandated research.md).
