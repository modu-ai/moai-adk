# SPEC-HARNESS-OUTCOME-ERRJOIN-001 — Progress

> Tier S minimal · era V3R6 · predecessor SPEC-HARNESS-OUTCOME-CAPTURE-001 (F2 SHOULD-FIX follow-up)

## §A Plan-phase (manager-spec)

- tier: S
- plan_complete_at: 2026-06-14T14:48:23Z
- plan_status: audit-ready
- plan_commit_sha: 5d25a5dcd
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

## §E.0 Phase 0.95 Mode Selection

- Decision: sub-agent
- Rationale: Tier S, 2 files, single-domain Go coding-heavy fix (one applier.go branch + one regression test) — per orchestration-mode-selection.md Mode 5 default; Anthropic coding-task parallelism caveat favors sequential sub-agent over parallel/team modes.
- GATE-2: user-approved (run-phase entry). Phase 0.5 SKIPPED (plan-auditor PASS 0.91 ≥ 0.90, skip-eligible; only post-verdict deltas were D1/D2 MINOR remediations + plan_commit_sha backfill, no new scope).

## §E.1 Run-phase Evidence (manager-develop, cycle_type=tdd)

- cycle_type: tdd · RED→GREEN single cycle · run_commit_sha: 5674734be

### RED→GREEN evidence

- **RED** (new test pre-fix): `go test ./internal/harness/ -run TestApply_Outcome_RolledBack_RecordError`
  → `--- FAIL`: `errors.As must reach *ApplyRegressionError on the joined error; got *fmt.wrapError: applier: non-regression gate blocked (rolled back); outcome record failed: observer: 디렉토리 생성 실패 …: not a directory`. Assertion (b) `errors.As(err, &*ApplyRegressionError)` was FALSE pre-fix — the F2 defect reproduced (the bare `fmt.Errorf` wrapper's only unwrap target is the observer error).
- **GREEN** (post-fix): rolled-back branch now returns `errors.Join(regErr, oerr)` (applier.go); the same test PASSES + all 5 existing `TestApply_Regression_*` remain GREEN. `errors.As` reaches the typed signal; the observer error (`디렉토리 생성 실패`) stays reachable; file rolled back to original bytes.

### AC Binary PASS/FAIL Matrix (acceptance.md §D.2 SSOT)

| AC ID | Severity | Status | Verification | Actual |
|-------|----------|--------|--------------|--------|
| AC-ERRJOIN-001 | MUST-PASS | PASS | new test `errors.As(err,&regErr)` TRUE | rolled-back + failing observer → `errors.As` TRUE, `regErr.Regressed` non-empty |
| AC-ERRJOIN-002 | MUST-PASS | PASS | outcome-record error reachable | `err.Error()` contains `디렉토리 생성 실패` |
| AC-ERRJOIN-003 | MUST-PASS | PASS | `TestApply_Regression_Blocks_RollsBack` GREEN | recordOutcome-SUCCESS path still bare `*ApplyRegressionError` |
| AC-ERRJOIN-004 | MUST-PASS | PASS | `grep 'file modified but outcome record failed'` | applier.go:474 present (kept branch byte-frozen) |
| AC-ERRJOIN-005 | MUST-PASS | PASS | `TestApply_Regression_*` 5/5 GREEN | rollback + `regression-blocked` lineage unaffected |
| AC-ERRJOIN-006 | MUST-PASS | PASS | `git diff --name-only` + frozen-sibling `git diff --stat` | only applier.go + applier_test.go; regression_gate/outcome/observer/measure.go empty diff |
| AC-ERRJOIN-007 | MUST-PASS | PASS | `go test ./...` exit 0 (96 pkg ok); `go vet ./...` 0; coverage | harness 87.3%→87.5% (no regression) |
| AC-ERRJOIN-008 | SHOULD | PASS | `grep 'errors.Join'` + `grep '"errors"'` | applier.go:455 `errors.Join`; applier.go:9 `"errors"` |

### Verification command outputs

- `go test ./internal/harness/... -run 'TestApply_Outcome_RolledBack_RecordError|TestApply_Regression'` → all PASS (6 tests)
- `go test ./...` → exit 0, 96 packages ok (one transient `internal/hook/wrapper_test.go` flake on first run — `moai`-binary PATH race, out of scope, passes on re-run + on clean baseline; NOT a regression)
- `go vet ./...` → exit 0 · `go test -cover ./internal/harness/` → 87.5% (baseline 87.3%)
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- `git diff --name-only` → `internal/harness/applier.go`, `internal/harness/applier_test.go` (exactly 2)
- frozen-sibling `git diff --stat -- regression_gate.go outcome.go observer.go ../measure/measure.go` → empty
- run_commit_sha: 5674734be

## §E.2 Sync-phase Audit-Ready Signal (manager-docs)

sync_commit_sha: 150d2745b

## §E.5 Mx-phase Audit-Ready Signal

mx_commit_sha: (backfill)

## §F Sync Audit (sync-auditor)

- verdict: PASS · overall 0.97 (harmonic mean, weighted)
- dimensions: Functionality 1.00 (MUST-PASS) / Security 1.00 / Craft 0.90 / Consistency 1.00 (MUST-PASS)
- RED genuineness: sync-auditor reconstructed the pre-fix branch in a scratch probe → `errors.As`=false (FAIL) pre-fix, true post-fix — defect real, test non-tautological
- adversarial bypass hunt (SEC-HARDEN-001 D1 lesson applied): extra `%w` re-wrap / Join-order / typed-nil — all defended, no demonstrated bypass
- D2 LOW (non-gating, NOT actioned): AC-002 partly relies on a Korean message substring (`디렉토리 생성 실패`); recommendation = export an observer sentinel for `errors.Is`-based assertion. Deferred — `observer.go` is a FROZEN sibling of this SPEC; actioning it would expand scope. Future-improvement candidate.
- report: `.moai/reports/sync-audit/SPEC-HARNESS-OUTCOME-ERRJOIN-001-2026-06-15.md` (gitignored, local)

## §G 4-Phase Close

- plan: `5d25a5dcd` (+ backfill `cc40fd876`) · run: `5674734be` · sync: `150d2745b` · Mx: (this commit)
- lifecycle: draft → in-progress → implemented → completed
- 8/8 AC PASS · plan-auditor 0.91 · sync-auditor 0.97 · coverage 87.5% · frozen siblings (regression_gate/outcome/observer/measure.go) byte-unchanged
- activation deferred to a future SPEC (dual-apply-path architecture decision; §D rationale)
