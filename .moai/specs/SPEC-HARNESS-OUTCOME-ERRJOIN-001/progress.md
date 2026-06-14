# SPEC-HARNESS-OUTCOME-ERRJOIN-001 — Progress

> Tier S minimal · era V3R6 · predecessor SPEC-HARNESS-OUTCOME-CAPTURE-001 (F2 SHOULD-FIX follow-up)

## §A Plan-phase (manager-spec)

- tier: S
- plan_complete_at: 2026-06-14T14:48:23Z
- plan_status: audit-ready
- plan_commit_sha: (backfill)
- artifacts: spec.md + plan.md + acceptance.md + progress.md
- spec-lint: strict mode, 0 findings, 0 warnings
- scope: `internal/harness/applier.go` (errors.Join at rolled-back branch ~L448-451 + `"errors"` import) + `internal/harness/applier_test.go` (new `TestApply_Outcome_RolledBack_RecordError`); kept branch (~L462-470) byte-frozen

## §B Plan Audit (plan-auditor iter1)

- verdict: PASS
- overall: 0.91 (Tier S threshold 0.75, delta +0.16; skip-eligible ≥ 0.90)
- dimensions: Clarity 0.95 / Completeness 0.92 / Testability 0.92 / Traceability 1.00
- must-pass: MP-1 PASS / MP-2 PASS / MP-3 PASS / MP-4 N/A (single-language internal pkg)
- report: `.moai/reports/plan-audit/SPEC-HARNESS-OUTCOME-ERRJOIN-001-iter1.md` (gitignored, local artifact)
- ground-truth: empirical RED/GREEN `errors.As` probe confirmed (PRE-FIX false → RED, POST-FIX true → GREEN); `Applier.Apply()` 0 production callers reconfirmed (§D activation-deferral rationale grounded)
- orchestrator-direct plan patches (L_orchestrator_direct_plan_patch): D1 MINOR (`tier: S` frontmatter added) · D2 MINOR (AC-ERRJOIN-002 Scenario 1(c) substring `파일 열기 실패` → `디렉토리 생성 실패`, the os.MkdirAll-first failure)

## §C Phase 0.5 SKIP Rationale (for /moai run)

plan-auditor PASS 0.91 ≥ 0.90 → skip-eligible. If no plan-PR commit lands after this verdict, the run-phase MAY skip Phase 0.5 re-execution (record in run delegation Section A: Context). GATE-2 (plan→run HUMAN GATE) remains mandatory and score-independent per CLAUDE.local.md §19.1 — skip-eligibility applies to Phase 0.5 verdict re-execution only, never to GATE-2.

## §D Scope ground-truth (carried for the future activation SPEC)

Out of scope: observer/gate ACTIVATION. Rationale — `Applier.Apply()` (safety→snapshot→regression-gate→outcome-capture→lineage) has ZERO production callers; `moai harness apply` (`runHarnessApply`) only surfaces a pending-proposal JSON payload to the orchestrator; production apply is performed by the skill-workflow Edit path (`.claude/skills/moai/workflows/harness.md`), bypassing the Go pipeline. `NewApplierWithRegressionGate()` / `WithOutcomeObserver()` also have 0 production callers. Activation is blocked on a dual-apply-path architecture decision (whether the Go pipeline becomes canonical) → deferred to a dedicated future SPEC. User decision A (F2-only) selected over activation options B/C/D.

## §E.1 Run-phase Evidence (manager-develop, cycle_type=tdd)

(pending — RED first: `TestApply_Outcome_RolledBack_RecordError` asserts `errors.As(err, &*ApplyRegressionError)` fails pre-fix; GREEN: `errors.Join(regErr, oerr)` at the rolled-back branch)

## §E.2 Sync-phase Audit-Ready Signal (manager-docs)

sync_commit_sha: (pending)

## §E.5 Mx-phase Audit-Ready Signal

mx_commit_sha: (pending)
