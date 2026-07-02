# Progress — SPEC-DEADPKG-INVESTIGATE-001

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **M** | era: **V3R6** | status: **draft**
- Artifacts: `spec.md` + `plan.md` + `acceptance.md` + `progress.md` created (plan-phase).
- SPEC ID Pre-Write Self-Check: `decomposition: SPEC ✓ | DEADPKG ✓ | INVESTIGATE ✓ | 001 ✓ → PASS`
- Frontmatter: 12 canonical fields + `tier: M` + `era: V3R6` + `related_specs` — validated.
- Plan-phase reachability baseline **observed** (not assumed) against the working tree
  (module `github.com/modu-ai/moai-adk`, go 1.26.4):
  - `internal/design` 19 files, 0 external importers.
  - `internal/research` 17 files, 0 external importers (stale `@MX:REASON` in `safety/limiter.go`).
  - `internal/runtime` (top-level) 8 files, 0 production importers; owning `SPEC-WF-AUDIT-GATE-001`
    is `implemented`; `gobin` subpackage LIVE (protected).
  - `internal/migrate` 1 file (`CleanupUserSettings`), 0 callers; `@MX:REASON` in
    `retired_events.go`; overlaps LIVE `internal/migration`.
  - `internal/i18n` 2 files, imported only by `internal/github/*_test.go`.
- Milestones defined: M1 runtime, M2 research, M3 design, M4 i18n, M5 migrate, M6 verification.
- Governing invariant recorded: a "dead package" is a defect claim = hypothesis until the
  run-phase tool checks confirm it (verification-claim-integrity §1.1 surface 3 + §5).
- Plan-phase gate: **ready for plan-auditor review**.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
