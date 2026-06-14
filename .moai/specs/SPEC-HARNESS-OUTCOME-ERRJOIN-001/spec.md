---
id: SPEC-HARNESS-OUTCOME-ERRJOIN-001
title: "Apply rolled-back branch errors.Join — preserve typed ApplyRegressionError signal"
version: "0.1.0"
status: in-progress
created: 2026-06-14
updated: 2026-06-15
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/harness"
lifecycle: spec-anchored
tags: "harness, error-propagation, errors-join, rolled-back, dormant-scaffold, tdd"
era: V3R6
tier: S
depends_on: [SPEC-HARNESS-OUTCOME-CAPTURE-001]
related_specs: [SPEC-HARNESS-REGRESSION-GATE-001]
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-14 | manager-spec | Initial draft — F2 follow-up of SPEC-HARNESS-OUTCOME-CAPTURE-001 (sync-auditor SHOULD-FIX, deferred). |

---

## §A Background / Problem (verified ground-truth)

`SPEC-HARNESS-OUTCOME-CAPTURE-001` added an additive Apply-OUTCOME observer. In
`internal/harness/applier.go`, `applyWithRegressionGate()` calls `a.recordOutcome(...)`
at BOTH terminal branches (rolled-back and kept) AFTER the gate decision is fixed —
the capture is additive (C10) and never flips the verdict.

**F2 defect (the rolled-back branch only).** At the rolled-back branch
(`applier.go` lines 433-451, inside the `if hasPriorBaseline { if regressed := … }`
block) when the gate detects a regression it: builds `summary`, calls
`RestoreSnapshot`, writes the `regression-blocked` lineage, then calls
`recordOutcome("rolled-back", …)`. If `recordOutcome` returns an error, the current
code (applier.go:448-451) does:

```go
if oerr := a.recordOutcome("rolled-back", "regression-blocked", baseline, candidate, regressed, proposal.ID); oerr != nil {
    return fmt.Errorf("applier: non-regression gate blocked (rolled back); outcome record failed: %w", oerr)
}
return &ApplyRegressionError{Baseline: baseline, Candidate: candidate, Regressed: regressed}
```

The error path returns ONLY the wrapped `oerr` and **discards the typed
`*ApplyRegressionError`**. A caller using `errors.As(err, &target)` (target
`*ApplyRegressionError`) to detect the regression-block therefore FAILS to detect it —
the typed regression signal is silently lost whenever the outcome-record write also
fails. The regression was rolled back (correct), the lineage `regression-blocked`
entry was written (correct), but the caller can no longer learn *via the typed signal*
that a regression-block occurred — it only sees a generic wrapped outcome-record error.

**Fix.** In the rolled-back branch, when `recordOutcome` errors, return
`errors.Join(regErr, oerr)` where `regErr := &ApplyRegressionError{Baseline: baseline,
Candidate: candidate, Regressed: regressed}` — preserving BOTH the typed regression
signal (`errors.As` keeps working, because `errors.Join` produces an aggregate whose
`Unwrap() []error` is walked by `errors.As`) AND the outcome-record error
(`errors.Is`/unwrap retrievable). When `recordOutcome` succeeds, the branch still
returns the bare `&ApplyRegressionError{…}` (unchanged). Add `"errors"` to the import
block (currently `encoding/json`, `fmt`, `os`, `path/filepath`, `strings`, `time`).

**kept branch MUST NOT change** (applier.go lines 462-470). The asymmetry is
deliberate and MUST be preserved: on the success (kept) path there is no typed success
error to preserve (success == `nil`), so the existing
`return fmt.Errorf("applier: file modified but outcome record failed: %w", oerr)` on a
recordOutcome failure is already correct. There is nothing to `errors.Join` against on
the kept path — joining `nil` with `oerr` would just yield `oerr`, which is exactly
what `fmt.Errorf(... %w, oerr)` already conveys. A reviewer MUST NOT "symmetrize" the
two branches: only the rolled-back branch carries a typed sentinel (`*ApplyRegressionError`)
that must survive the join.

**Existing-helper note (informs §acceptance).** The current test helper
`asRegressionError(err, &regErr)` (applier_test.go:1125) uses a DIRECT type assertion
`err.(*ApplyRegressionError)`, NOT `errors.As`. After the fix returns
`errors.Join(regErr, oerr)`, the joined value is NOT directly type-`*ApplyRegressionError`,
so that direct-assertion helper would return `false` — but `errors.As(err, &regErr)`
WILL succeed because `errors.Join` exposes `Unwrap() []error` which `errors.As` walks.
The new test therefore asserts on `errors.As` (the standard caller idiom), which is the
contract this SPEC restores. Existing rolled-back tests that use the bare-error path
(recordOutcome SUCCESS) are unaffected.

## §A.2 HONEST framing (no overclaiming)

This fix improves the **error-propagation correctness of a dormant scaffold**. The full
Apply pipeline (`Applier.Apply()` → safety → snapshot → regression-gate →
outcome-capture → lineage) has **ZERO production callers** — it is invoked only in
tests (see §D for the dual-apply-path rationale). Therefore this change:

- does NOT activate anything,
- does NOT prevent regressions in production,
- does NOT change the runtime behavior of the current harness,
- is **forward-looking**: it matters once/if the Go Apply pipeline becomes a production
  apply path.

The captured project-health delta is typically Δ=0 for the current markdown-only
harness write surface, so the rolled-back branch this SPEC corrects is itself rarely
reached in practice. The value of the fix is purely correctness-of-contract: when a
future production caller relies on `errors.As(err, &*ApplyRegressionError)` to detect a
regression-block, that detection must not be silently defeated by a coincident
outcome-record write failure. CHANGELOG/spec wording MUST state the dormant-scaffold +
forward-looking nature plainly.

---

## §B Requirements (GEARS notation)

- **REQ-ERRJOIN-001** (Event-driven): **When** the in-Apply regression gate reaches the
  rolled-back branch (`hasPriorBaseline` true AND `baseline.Regressions(candidate)`
  non-empty) **and** `recordOutcome` returns a non-nil error, the applier **shall**
  return an error for which `errors.As(err, &target)` with `target *ApplyRegressionError`
  evaluates to `true` (the typed regression signal is preserved).

- **REQ-ERRJOIN-002** (Event-driven): **When** the rolled-back branch's `recordOutcome`
  returns a non-nil error, the applier **shall** ALSO surface that outcome-record error
  in the returned value such that it is retrievable via `errors.Is`/unwrap (neither the
  typed regression signal nor the outcome-record error is suppressed).

- **REQ-ERRJOIN-003** (Event-driven): **When** the rolled-back branch's `recordOutcome`
  returns nil (success), the applier **shall** return a bare `*ApplyRegressionError`
  (i.e. `errors.As` succeeds and there is no joined second error) — no behavioral
  regression versus the pre-fix success path.

- **REQ-ERRJOIN-004** (Ubiquitous): The kept branch (`applier.go` lines 462-470)
  error/return semantics **shall** remain byte-unchanged — on a kept-path
  `recordOutcome` failure the applier shall continue to return
  `fmt.Errorf("applier: file modified but outcome record failed: %w", oerr)`.

- **REQ-ERRJOIN-005** (Ubiquitous): The capture **shall** remain additive (C10) — the
  keep/rollback decision, the snapshot rollback, the baseline-store state, the
  `regression-blocked` lineage write, and the returned verdict **shall** be unaffected
  by this change (the change touches only how the rolled-back branch composes its
  returned error when an outcome-record error coincides).

- **REQ-ERRJOIN-006** (Capability gate): **Where** no outcome observer is wired (the
  gate-active-but-no-observer case, or the gate-inactive default), `recordOutcome`
  remains a no-op (`nil`), so the rolled-back branch returns the bare
  `*ApplyRegressionError` exactly as before — this change introduces no new behavior on
  the no-observer path.

---

## §C Constraints (HARD)

- **C-ERRJOIN-01 (FROZEN guard)**: `internal/harness/frozen_guard.go` is unchanged;
  `learning.auto_apply: false` is unchanged; the M7 human-gate flow (`ApplyPendingError`
  → orchestrator AskUserQuestion) is unchanged.
- **C-ERRJOIN-02 (scope discipline)**: ONLY `internal/harness/applier.go` and
  `internal/harness/applier_test.go` may be modified. `regression_gate.go`,
  `outcome.go`, `observer.go`, and `internal/measure/measure.go` MUST remain
  byte-unchanged (git diff verification AC).
- **C-ERRJOIN-03 (asymmetry preserved)**: The kept branch MUST NOT be "symmetrized"
  with the rolled-back branch. Only the rolled-back branch carries a typed sentinel to
  preserve via `errors.Join`.
- **C-ERRJOIN-04 (quality gates)**: `go test ./...` exits 0; `go vet ./...` reports 0
  findings; `internal/harness` coverage does not regress below its current baseline.
- **C-ERRJOIN-05 (import discipline)**: The only new import is the standard-library
  `errors` package in `applier.go`. No third-party dependency is added.

---

## §D Exclusions (What NOT to Build)

### §D.1 Out of Scope — with ground-truth rationale

- **Observer / gate ACTIVATION is entirely out of scope.** Rationale (recorded verbatim
  so the future activation SPEC inherits it): `Applier.Apply()` — the full pipeline
  safety → snapshot → regression-gate → outcome-capture → lineage — has **ZERO
  production callers** (only invoked in tests). The production `moai harness apply` verb
  (`runHarnessApply` in `internal/cli/harness.go`) only surfaces a pending-proposal JSON
  payload to the orchestrator; it never calls `Applier.Apply()`. The actual production
  apply is performed by the orchestrator's skill-workflow Edit path
  (`.claude/skills/moai/workflows/harness.md`), bypassing the Go pipeline.
  `NewApplierWithRegressionGate()` / `WithOutcomeObserver()` likewise have 0 production
  callers. Therefore "wire the observer/gate into the production Apply flow" is blocked
  on a dual-apply-path architecture decision (whether the Go pipeline should become the
  canonical apply path) and is deferred to a dedicated future SPEC. The user explicitly
  chose this F2-only scope (decision A) over activation options.
- **Documentation-drift reconciliation is out of scope**: reconciling `harness.md`
  (V3R4, claims "Go CLI never registered") vs `harness_route.go` (V3R5, re-registers
  verbs) is a separate concern.
- **Phase5 learning rigor is out of scope**: D1 failure-signature clustering, D3
  held-out evaluation, D4 scorer-loop are downstream and out of scope.
- **Kept-branch error reshaping is out of scope**: the kept branch is byte-frozen
  (REQ-ERRJOIN-004); no `errors.Join` or wording change there.
