# SPEC-HARNESS-EVOLVE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-12
plan_version: 0.1.2   # iter-1 FAIL 0.75 → D1-D4+S1-S6; iter-2 PASS 0.89 → D-1 + N-1/N-2 amendment
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
req_count: 21
ac_count: 27          # 34 verification line items (022a-d, 023a-e)
open_clarifications: 0   # all resolved — 3 pinned user decisions (plan.md §H)
```

Plan-phase notes:
- Design SSOT: `.moai/reports/harness-self-evolving-redesign-final-20260712.html`
  (§4 3-Zone, §5 Loop 0 schema + A2/A4 deltas, §7 M1).
- Plan-audit iter-1: FAIL 0.75 (`.moai/reports/plan-audit/SPEC-HARNESS-EVOLVE-001-review-1.md`).
  v0.1.1 applies D1 (HOI dual-gate reconciliation — REQ-HEV-016 rewrite +
  spec.md §D.3 activation precondition + AC-HEV-021/Scenario 1/5/6 fixtures),
  D2 (CLI registration re-pinned to `newHarnessRouterCmd()`/harness_route.go +
  `v3r5RequiredHarnessVerbs` step + AC-HEV-027), D3 (AC-HEV-011 write-surface
  re-scope), D4 (§H markers struck), S1-S6.
- Resolved clarifications (AskUserQuestion round 2026-07-12, pinned in
  plan.md §H): (1) KEEP HOI-gated transport, activation precondition codified,
  no new gate / no default flip; (2) `request_class` INCLUDED in schema v1
  (coarse keyword enum, non-verbatim); (3) v1 no-rotation, retention deferred
  to EVOLVE-003 with `retention.go` reuse preserved.
- Gate note (spec.md §D.3): Stop-path outcome finalization requires
  `hook.opt_in.enabled: true` (fail-closed, default OFF per
  SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001) THEN `learning.enabled` (fail-open,
  default true). Default-config Stop-path dormancy is EXPECTED shipped
  behavior; this dev repo enables the opt-in locally during M4.
- AC baselines measured against main working tree 2026-07-12 (all
  discriminating tokens at 0 on registration surfaces; `.moai/state/`
  gitignore rule pre-existing; `"ledger"` in harness_retirement_test.go = 0;
  template workflow mirrors routing-ledger = 0 ×3).
- Pinned decisions D1-D6 in plan.md §A.2 (HOI-gated Stop-hook transport
  reuse, hook-side finalization authority, multi-turn pending, dual-gate
  inheritance, per-session pending isolation, best-effort identity
  resolution).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
