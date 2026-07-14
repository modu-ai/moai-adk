---
id: SPEC-CLI-TUX-V3-005
title: "AC-TUX3-020 Printer Migration — fmt.Print* Ratchet Succession"
version: "0.1.0"
status: draft
created: 2026-07-14
updated: 2026-07-14
author: manager-spec
priority: P2
phase: "v3.0.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "cli, printer, ratchet, debt, tux-v3"
depends_on: [SPEC-CLI-TUX-V3-001]
tier: M
era: V3R6
---

# progress.md — SPEC-CLI-TUX-V3-005

> Plan-phase progress tracker. Run-phase evidence (§E.2-§E.4) is populated by manager-develop; sync-phase evidence (§E.4) by manager-docs. This file carries the §E section skeleton at plan-phase per the era-classification engine requirement.

---

## §A. Current State

| Field | Value |
|-------|-------|
| Phase | plan (artifacts authored, awaiting plan-audit) |
| Status | draft |
| plan_complete_at | _pending audit-ready verdict_ |
| Baseline (fmt.Print* count) | 38 (measured 2026-07-14) |

---

## §B. Phase Skip Rationale

### Phase 1 (Brain/analyze) — SKIP

**Rationale**: This SPEC is a debt split from SPEC-CLI-TUX-V3-003 (amendment re-close), not a new brain-proposal. The requirement (REQ-TUX3-020) was already analyzed and defined in SPEC-003 spec.md:98. The orchestrator's read-only investigation this session established all ground-truth context (38-call baseline distribution, existing Printer interface surface, gap-site classification). No new requirement analysis is needed — the migration target and scope are fully determined by the inherited REQ and the existing `printer.Printer` interface.

### Phase 3 (Context-First Discovery) — SKIP

**Rationale**: Intent clarity is 100%. The orchestrator's investigation (this session) fully established: (a) the 38-call baseline and per-file distribution, (b) the existing Printer interface method catalogue (printer.go:64-112), (c) the call-site classification (24 migratable, 13 gap, 1 dead code), (d) the dependency chain (SPEC-001 completed, SPEC-003 completed). The three clarification items (banner.go scope, branch_protection.go dead-code, state.go/migration.go human-format channel) were M1-gate decisions RESOLVED 2026-07-14 via user AskUserQuestion — see §D below. All three resolved with the default-recommended option (exclude / exclude / stdout Data() composed string). The domain (CLI output layer) is familiar; no Unknown-Unknowns are suspected.

---

## §C. Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-07-14 | SPEC ID: SPEC-CLI-TUX-V3-005 | Series continuity with TUX-V3-001/002/003/004. Alternative SPEC-CLI-PRINTER-001 considered but rejected — the debt originates from the TUX v3 series and the Printer was introduced by SPEC-001. |
| 2026-07-14 | Tier: M (standard) | Single domain (CLI output), interface already exists, 5 non-test files. Interface-design risk (Tier L signal) is retired. |
| 2026-07-14 | M1 architecture gate leads milestone ordering | banner.go + branch_protection.go scope decision determines whether Printer interface needs extension — highest reversibility. |

---

## §D. Resolved Decisions (formerly open clarification items; resolved 2026-07-14 via user AskUserQuestion)

These items MUST be resolved via orchestrator AskUserQuestion before run-phase entry:

1. **[DECISION 2026-07-14: uikit/banner.go scope — EXCLUDED as TUI render]** — The 12 `fmt.Print*` calls in banner.go (TUI render via lipgloss) are EXCLUDED; Printer interface not extended. (See research.md §C.6, plan.md §F M1.)

2. **[DECISION 2026-07-14: branch_protection.go ttyConfirmer — EXCLUDED as dead code]** — The 1 `fmt.Printf` prompt in `ttyConfirmer.Confirm()` (dead code, `nolint:unused`) is EXCLUDED; active path is yesConfirmer + --yes-branch-protection. (See research.md §C.5, plan.md §F M1.)

3. **[DECISION 2026-07-14: state.go/migration.go human-format — stdout Data() composed string]** — The multi-line human-readable state/migration display routes to stdout via `Data()` (composed string); scripted consumers unaffected. (See research.md §C.1/C.2, plan.md §F M2.)

---

## §E. Section Skeleton

> The following §E subsections are placeholder headings only at plan-phase. They are populated by manager-develop (§E.2, §E.3) and manager-docs (§E.4) during run/sync phases. The era-classification engine (`internal/spec/era.go` `hasAnyProgressMarker`) greps for the literal `§E.2`/`§E.3`/`§E.4` substrings to classify the SPEC era — emitting these placeholder headings at plan-phase prevents misclassification.

## §E.1 Plan-phase Audit-Ready Signal

_status: pending plan-auditor verdict_

Plan-phase artifacts (spec.md, plan.md, acceptance.md, research.md, progress.md) authored 2026-07-14. Awaiting plan-auditor independent audit. The §E.1 audit-ready signal will be populated upon receiving a PASS or PASS-WITH-DEBT verdict.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

**Input parameters**:
- tier: M
- scope: 5 non-test files (state.go, migration.go, tmux_integration.go, banner.go, branch_protection.go); 24 migratable call sites
- domain count: 1 (CLI output layer)
- file language mix: Go (100%)
- concurrency benefit: LOW (coding-heavy migration per Anthropic coding-task parallelism caveat)

**Mode evaluation**:
- Mode 1 (trivial): NO — multi-file migration with characterization tests
- Mode 2 (background): NO — write-capable implementation work
- Mode 3 (agent-team): RETIRED
- Mode 4 (parallel): NO — coding-heavy, single domain, few parallelizable tasks
- Mode 5 (sub-agent): SELECTED — sequential per-milestone DDD cycles
- Mode 6 (workflow): NO — semantic per-file migration (not a single uniform mechanical transform); <30 files

**Decision**: sub-agent (Mode 5)

**Justification**: Tier M single-domain coding-heavy migration. Each milestone (M2 state.go 11 calls, M3 migration.go 8 calls, M4 tmux_integration.go 5 calls) is an independent DDD ANALYZE-PRESERVE-IMPROVE cycle on one file — sequential sub-agent per milestone matches Anthropic's coding-task parallelism caveat ("most coding tasks involve fewer truly parallelizable tasks than research"). M1 architecture gate already RESOLVED (3 decisions recorded 2026-07-14: banner.go excluded, ttyConfirmer excluded, human-format stdout Data()). M5 is ratchet verification. Per-milestone dispatch also avoids the autocompact-thrash hazard of a single large-scope spawn.

**Implementation Kickoff Approval**: OBTAINED 2026-07-14 (user AskUserQuestion — run-phase entry approved, score-independent mandatory gate cleared).

**Plan Audit Gate (Phase 1)**: PASS 0.95 (plan-auditor iter-2, 2026-07-14). skip-eligible (score ≥0.90 + artifact hash stable post-acd7d8046 + within 24h) — Phase 1 re-execution skipped; the iter-2 audit IS the verdict of record.
