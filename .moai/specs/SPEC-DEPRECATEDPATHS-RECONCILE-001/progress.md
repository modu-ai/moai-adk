---
id: SPEC-DEPRECATEDPATHS-RECONCILE-001
title: "Reconcile DeprecatedPaths — progress tracking"
version: "0.1.0"
status: draft
tier: S
era: V3R6
---

# Progress — SPEC-DEPRECATEDPATHS-RECONCILE-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_complete_at: 2026-07-02
- plan_status: audit-ready
- tier: S (2 files edited: `internal/defs/dirs.go` + `internal/defs/dirs_test.go`; + doc-comment sites; no template)
- recommended_direction: **A (un-deprecate)** — remove `design.yaml` + `db.yaml` from `DeprecatedPaths`
- artifacts: spec.md (§1 context + §2 GEARS REQ-DPR-001..005 + §3 inline AC-DPR-001..009 + §4 Out of Scope + §5 cross-refs), plan.md (M1 RED→GREEN, M2 reconcile+verify), progress.md
- investigation summary: origin SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 REQ-VVCR-011/012 deliberately marked both files for removal on premises TRUE at authoring (design-system retirement; `_TBD_` placeholder), which became STALE — `db.yaml` re-established live by SPEC-DB-SYNC-001 (real `migration_patterns`, read by `loadMigrationPatterns`), `design.yaml` retained + reclassified live by SPEC-DEAD-CONFIG-001 §D. Ground-truth-shift reconciliation, not repudiation.
- count reconciliation: total 43→41, Category B 31→29 (Category A 9 + Category C 3 unchanged)
- no-overlap check: SPEC-DEAD-CONFIG-001 (completed) scoped `runtime.yaml` only and explicitly left `DeprecatedPaths` intact + parked `design.yaml` "Do NOT touch" — no duplication with this SPEC
- run-phase decision point: §A.4 count-derivation reconciliation (Option A.2 recommended — redirect `@MX:ANCHOR` to this SPEC, preserve completed origin body immutability)

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
