# SPEC-PREPUSH-MODE-WIRING-001 — Progress

## Phase status

| Phase | Status | Commit SHA | Notes |
|-------|--------|-----------|-------|
| Plan  | complete | (pending) | spec.md + plan.md + acceptance.md + progress.md authored. status: draft. |
| Run   | pending  | — | cycle_type=tdd, M1-M5 (M4 conditional on REQ-PMW-012). |
| Sync  | pending  | — | — |
| Mx    | pending  | — | — |

## Plan-phase summary

- **Tier**: S (minimal — ~50-80 LOC production change: two pure helpers `resolvePrePushAction` +
  `decideExit` + one branch in `runPrePush` + optional env const).
- **REQ count**: 13 (REQ-PMW-001 .. REQ-PMW-012 + REQ-PMW-002a testability seam; REQ-PMW-012 SHOULD/optional).
- **AC count**: 13 (AC-PMW-001 .. AC-PMW-013; AC-PMW-012 conditional on REQ-PMW-012; AC-PMW-013 = gate-OFF predecessor-preservation regression pin).
- **Module**: `internal/cli` (+ `internal/config` only if REQ-PMW-012 env const).
- **Predecessor**: SPEC-PREPUSH-WIRING-001 (completed) — 1st dead-config follow-up (`enforce_on_push`).
  This is the 2nd dead-config follow-up (`git_strategy.<mode>.hooks.pre_push` severity dial).

## Precedence model (as encoded)

```
env(MOAI_ENFORCE_ON_PUSH)  >  enforce_on_push (MASTER GATE)  >  pre_push (SEVERITY dial)
                                       |                              |
                              gate OFF (default) ⇒ no-op,      gate ON ⇒ skip / warn / enforce
                              pre_push NEVER consulted          via ActiveModeProfile().Hooks.PrePush
```

- Fail-safe defaults: nil ModeProfile → `enforce`; unknown pre_push value → `enforce`.
- Optional `MOAI_PRE_PUSH` severity override sits BELOW the gate (never opens the gate).

## §E.1 Plan-phase audit-ready signal

- plan_complete_at: 2026-06-08
- plan_status: audit-ready
- SPEC ID self-check: `decomposition: SPEC ✓ | PREPUSH ✓ | MODE ✓ | WIRING ✓ | 001 ✓ → PASS`
- plan-auditor verdict: PASS-WITH-DEBT 0.84 (Tier S threshold 0.80); 4 defects, all orchestrator-re-verified against live source, all 4 patched:
  - D1 (SHOULD-FIX): drifted template citation — `pre_push` default at git-strategy.yaml.tmpl:34/66/104; `enforce|warn|skip` vocabulary on sibling `pre_commit` line :33/65/103 (NOT on pre_push). Fixed §A.1 + Cross-References.
  - D4 (MINOR): off-by-two — `HooksConfig.PrePush` field at types.go:92 (line 90 is the struct decl). Fixed §A.1 + Cross-References.
  - D2 (SHOULD-FIX, borderline BLOCKING): exit-2 path not in-process testable; `TestRunPrePush_WithViolations` false-named (fails at /dev/stdin, never reaches os.Exit); no subprocess harness in internal/cli/*_test.go. Added REQ-PMW-002a testability seam (pure `decideExit` + `resolvePrePushAction`); rewrote AC-PMW-002/003/005/006/007 to assert pure helpers; flagged the barrier in plan.md §A.1 + §E.
  - D3 (MINOR): added AC-PMW-013 gate-OFF predecessor-preservation regression pin (existing `TestRunPrePush_EnforcementDisabled_ReturnsNilImmediately` + new `TestRunPrePush_GateOff_PrePushNotConsulted`); noted gate-OFF is the only legacy-harness-reachable row in plan.md §E.

## §E.2 Run-phase Evidence

(populated by manager-develop)

## §E.3 Run-phase Audit-Ready Signal

(populated by manager-develop)

## §E.4 Sync-phase Audit-Ready Signal

(populated by manager-docs)

## §E.5 Mx-phase Audit-Ready Signal

(populated at 4-phase close)
