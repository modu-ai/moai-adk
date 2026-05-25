---
id: SPEC-V3R6-TEST-REFACTOR-001
title: "Go test suite refactor — phase progress tracker"
version: "0.1.0"
status: draft
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P1
phase: "v3.0.0 follow-up"
module: "internal/template, internal/skills, internal/harness, internal/statusline"
lifecycle: spec-anchored
tags: "test-refactor, progress, atr-001-debt-discharge"
depends_on: [SPEC-V3R6-AGENT-TEAM-REBUILD-001]
tier: M
---

# SPEC-V3R6-TEST-REFACTOR-001 — Phase progress

## Section A — Lifecycle Sync

| Field | Value |
|-------|-------|
| plan_commit_sha | pending |
| sync_commit_sha | pending |
| mx_commit_sha | pending |
| supersedes | (none) |
| superseded_by | (none) |
| anchor SPEC | SPEC-V3R6-AGENT-TEAM-REBUILD-001 |

L60 atomic backfill protocol: `pending` placeholders replaced by actual SHAs via separate chore commits AFTER each phase's primary commit lands. plan_commit_sha is backfilled immediately following this plan-phase commit; sync_commit_sha is backfilled following manager-docs sync-phase; mx_commit_sha is backfilled following Mx-phase close marker emission.

## Section B — Run-phase milestone log

*(empty — populated by manager-develop during run-phase execution)*

| Milestone | Commit SHA | Status | Date | Notes |
|-----------|-----------|--------|------|-------|
| M1 | pending | not started | — | — |
| M2 | pending | not started | — | — |
| M3 | pending | not started | — | — |
| M4 | pending | not started | — | — |
| M5 | pending | not started | — | — |
| M6 | pending | not started | — | — |

## Section C — Sync-phase log

*(empty — populated by manager-docs during sync-phase execution)*

## Section D — Mx-phase log

*(empty — populated during Mx-phase execution)*

### D.1 Mx Step C (EVALUATE) decision

*(empty — manager-develop or orchestrator records EVALUATE-EXECUTE vs EVALUATE-SKIP decision per coverage delta heuristic)*

## Section E — Phase evidence & audit-ready signals

### E.1 Plan-phase audit-ready signal

| Field | Value |
|-------|-------|
| plan_auditor_iteration | pending (iter-1 expected on plan-phase entry) |
| plan_auditor_verdict | pending |
| plan_auditor_score | pending |
| phase_0_5_skip_eligible | pending (≥0.90 threshold per CONST-V3R5-026) |

### E.2 Run-phase evidence

*(populated by manager-develop)*

### E.3 Run-phase audit-ready signal

*(populated by manager-develop)*

### E.4 Sync-phase audit-ready signal

*(populated by manager-docs)*

### E.5 Mx-phase audit-ready signal

*(populated during Mx-phase)*

## Section F — 4-phase close marker

*(emitted at Mx-phase close terminator commit per L60 atomic chicken-and-egg pattern)*

## HISTORY

### v0.1.0 (2026-05-25) — initial draft

- Plan-phase progress tracker scaffold authored.
- §A Lifecycle Sync row populated with `pending` placeholders for L60 atomic backfill protocol.
- §B Run-phase milestone log scaffolded with 6 M-rows ready for manager-develop population.
- §E phase-evidence & audit-ready signal slots scaffolded for all 4 phases.
