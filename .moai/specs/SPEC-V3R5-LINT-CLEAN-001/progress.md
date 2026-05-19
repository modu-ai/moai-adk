# SPEC-V3R5-LINT-CLEAN-001 Progress

- Started: 2026-05-19 (run-phase entry, LCLN-Phase 1)
- Base: main HEAD `6604ca234` (PR #1008 plan-PR admin-merged 2026-05-19T12:06:15Z)
- Branch: `chore/SPEC-V3R5-LINT-CLEAN-001-phase-1`

## Phase 0.5 — Plan Audit Gate

- Verdict: **PASS** (re-using plan-phase iter-2 verdict, harmonic mean 0.9251)
- Source: `spec.md` HISTORY v0.2.0 (2026-05-19)
- Re-audit: NOT triggered (Grace Window 7-day active, plan artifacts unchanged since PR #1008 merge)
- Report: `.moai/reports/plan-audit/SPEC-V3R5-LINT-CLEAN-001-2026-05-19.md`

## LCLN-Phase 1 Progress (current)

- Goal: reduce `moai agent lint --strict` baseline by **60 findings** (176 → 116)
- Scope: T1.1 LR-03 (24) + T1.2 LR-12 (6) + T1.3 LR-06 (28) + T1.4 LR-05 (2) = 60
- Approach: template-first edits to `internal/template/templates/.claude/agents/moai/*.md`, then mirror to live `.claude/agents/moai/*.md`, then `make build`.

### T1.0 — Baseline Capture

- Status: **completed** (2026-05-19)
- Output: `.moai/state/lint-baseline-pre-LCLN-P1.json` (gitignored)
- Total: 176 (93 errors + 83 warnings)
- Breakdown: LR-08=83, LR-06=29, LR-01=28, LR-03=25, LR-12=6, LR-02=3, LR-05=2

### T1.1–T1.4 — D1+D2+D4 Cleanup (in progress)

Delegated to manager-develop. See PR description for fix details.

### T1.5 — make build

Pending.

### T1.6 — Post-Phase-1 Delta Verification

Pending.

### T1.7 — Orthogonal Lints + Frozen Guard

Pending.

## LCLN-Phase 2-4 (deferred to subsequent run-phase invocations)

- LCLN-Phase 2 (D7): live `expert-mobile.md` deletion (−4 findings)
- LCLN-Phase 3 (D3): tool-boundary cleanup (−30 findings)
- LCLN-Phase 4 (D5): preload drift W4-resolvable subset (−70 findings)
- Mega-Sprint W2 (CORE-SLIM-001): dissolves residual 12 W2-deferred LR-08
