---
id: SPEC-V3R6-TEST-REFACTOR-001
title: "Go test suite refactor — phase progress tracker"
version: "0.1.1"
status: in-progress
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
| M1 | pending | in-progress | 2026-05-25 | Ground truth re-measured at HEAD 40dc43f5b: exactly 15 FAIL lines matching §A.4 (zero drift). Frontmatter draft→in-progress applied to 4 SPEC artifacts. |
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

## Section F — Run-phase verification + close marker

### F.0 Pre-run-phase ground truth (M1)

Measured at HEAD `40dc43f5b` on 2026-05-25:

```
$ go test ./... 2>&1 | grep -c "^--- FAIL"
15
```

Per-package breakdown:
- `internal/template`: 11 failures (TestContractSchemaVerification, TestBackwardCompatibility, TestContractAssertionsNaturalLanguage, TestAgentFrontmatterAudit, TestTemplateAgentsStructure, TestEmbeddedTemplates_AgentDefinitions, TestLoadEmbeddedCatalog_Success, TestLoadCatalog, TestAllAgentsInCatalog, TestRuleTemplateMirrorDrift, TestRetirementCompletenessAssertion)
- `internal/skills`: 2 failures (TestTemplateMirrorParity, TestSubSkillLOCCeiling)
- `internal/harness`: 1 failure (TestSubagentBoundary_NoAskUserQuestion)
- `internal/statusline`: 1 failure (TestRenderPRSegment_Absence)

**Zero drift from §A.4 baseline (measured at HEAD `e7b119924`).** Ground truth verification PASS — proceed to M2.

### F.x Run-phase milestone records

*(populated by manager-develop M1–M6)*

### 4-phase close marker

*(emitted at Mx-phase close terminator commit per L60 atomic chicken-and-egg pattern)*

## HISTORY

### v0.1.1 (2026-05-25) — run-phase M1 frontmatter status:in-progress + ground truth verification

- Run-phase M1 entry: frontmatter status transition `draft → in-progress` per Status Transition Ownership Matrix exception (manager-develop allowed on draft → in-progress only).
- §F.0 Pre-run-phase ground truth populated: 15 failures measured at HEAD `40dc43f5b`, zero drift from §A.4 baseline.
- §B Run-phase milestone log §B M1 row populated (commit SHA pending L60 atomic backfill).

### v0.1.0 (2026-05-25) — initial draft

- Plan-phase progress tracker scaffold authored.
- §A Lifecycle Sync row populated with `pending` placeholders for L60 atomic backfill protocol.
- §B Run-phase milestone log scaffolded with 6 M-rows ready for manager-develop population.
- §E phase-evidence & audit-ready signal slots scaffolded for all 4 phases.
