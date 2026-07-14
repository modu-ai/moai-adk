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

**Rationale**: Intent clarity is 100%. The orchestrator's investigation (this session) fully established: (a) the 38-call baseline and per-file distribution, (b) the existing Printer interface method catalogue (printer.go:64-112), (c) the call-site classification (24 migratable, 13 gap, 1 dead code), (d) the dependency chain (SPEC-001 completed, SPEC-003 completed). The two remaining [NEEDS CLARIFICATION] items (banner.go scope, branch_protection.go dead-code disposition) are M1-gate decisions that require user input at the Implementation Kickoff Approval, not Socratic discovery rounds. The domain (CLI output layer) is familiar; no Unknown-Unknowns are suspected.

---

## §C. Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-07-14 | SPEC ID: SPEC-CLI-TUX-V3-005 | Series continuity with TUX-V3-001/002/003/004. Alternative SPEC-CLI-PRINTER-001 considered but rejected — the debt originates from the TUX v3 series and the Printer was introduced by SPEC-001. |
| 2026-07-14 | Tier: M (standard) | Single domain (CLI output), interface already exists, 5 non-test files. Interface-design risk (Tier L signal) is retired. |
| 2026-07-14 | M1 architecture gate leads milestone ordering | banner.go + branch_protection.go scope decision determines whether Printer interface needs extension — highest reversibility. |

---

## §D. [NEEDS CLARIFICATION] Items (gate Implementation Kickoff Approval)

These items MUST be resolved via orchestrator AskUserQuestion before run-phase entry:

1. **[NEEDS CLARIFICATION: uikit/banner.go scope]** — Should the 12 `fmt.Print*` calls in banner.go (TUI render via lipgloss) be migrated to Printer (requiring a new styled-stdout interface method), or excluded as TUI render out-of-scope? Default recommendation: exclude. (See research.md §C.6, plan.md §F M1.)

2. **[NEEDS CLARIFICATION: branch_protection.go ttyConfirmer]** — Should the 1 `fmt.Printf` prompt in `ttyConfirmer.Confirm()` (dead code, `nolint:unused`) be migrated for ratchet credit, or excluded as deferred code? Default recommendation: exclude. (See research.md §C.5, plan.md §F M1.)

3. **[NEEDS CLARIFICATION: state.go/migration.go human-format channel]** — Should the multi-line human-readable state/migration display route to stdout via `Data()` (as a composed string) or be restructured? This affects 6+ call sites across state.go and migration.go. (See research.md §C.1/C.2, plan.md §F M2.)

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
