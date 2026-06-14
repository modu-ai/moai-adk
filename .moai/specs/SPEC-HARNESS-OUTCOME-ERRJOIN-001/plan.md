---
id: SPEC-HARNESS-OUTCOME-ERRJOIN-001
title: "Apply rolled-back branch errors.Join — implementation plan"
version: "0.1.0"
status: draft
created: 2026-06-14
updated: 2026-06-14
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/harness"
lifecycle: spec-anchored
tags: "harness, error-propagation, errors-join, rolled-back, dormant-scaffold, tdd"
---

## §A Context

Tier **S (minimal)**, development_mode **tdd**. Single 2-file scope
(`internal/harness/applier.go` + `internal/harness/applier_test.go`). This is the F2
follow-up of `SPEC-HARNESS-OUTCOME-CAPTURE-001` (its sync-auditor flagged the
rolled-back-branch typed-signal loss SHOULD-FIX and deferred it). The fix is a
3-line change plus one import line, guarded by one new RED→GREEN test.

The defect, the fix, and the deliberate kept-branch asymmetry are fully specified in
`spec.md` §A. This plan does not restate them; it records the edit target, the TDD
milestone, the constraints self-check, and the anti-patterns to avoid.

## §B Known Issues / Pre-existing State

- The existing helper `asRegressionError` (applier_test.go:1125) uses a DIRECT type
  assertion, not `errors.As`. The new test MUST use `errors.As` directly (the standard
  caller idiom and the contract being restored), NOT the existing direct-assertion
  helper, because the post-fix joined error is not directly type-`*ApplyRegressionError`.
- `internal/harness` currently builds clean; `go test ./...` is GREEN at the predecessor
  close (`a11c268f1`). The pre-existing flaky `dbsync` timing test is unrelated to this
  package and out of scope.

## §C Pre-flight (run-phase entry checklist)

- [ ] Confirm `applier.go` rolled-back branch is still at lines 433-451 (the
  `if hasPriorBaseline { if regressed := … }` block ending in the bare
  `return &ApplyRegressionError{…}`); if the line range drifted, locate by the literal
  anchor `non-regression gate blocked (rolled back); outcome record failed`.
- [ ] Confirm the import block in `applier.go` does NOT already contain `"errors"`.
- [ ] `go test ./internal/harness/... ` GREEN before starting (baseline).

## §D Constraints (self-check, mirrors spec §C)

- ONLY `applier.go` + `applier_test.go` modified.
- `regression_gate.go` / `outcome.go` / `observer.go` / `internal/measure/measure.go`
  byte-unchanged (verified by `git diff --name-only`).
- kept branch (lines 462-470) byte-unchanged.
- only new import: stdlib `errors`.
- `go test ./...` exit 0, `go vet ./...` 0, `internal/harness` coverage no-regression.

## §E Self-Verification (post-implementation, read-only batch)

Run as a single parallel Bash batch at run-phase completion:

```bash
# 1. full suite
go test ./...
# 2. package coverage (no-regression vs baseline)
go test -coverprofile=cover.out ./internal/harness/...
# 3. vet
go vet ./...
# 4. errors.Join present at the fix site
grep -n 'errors.Join' internal/harness/applier.go
# 5. errors import added
grep -n '"errors"' internal/harness/applier.go
# 6. scope discipline — exactly two files changed
git diff --name-only
# 7. kept-branch wording preserved (must still be present, unchanged)
grep -n 'file modified but outcome record failed' internal/harness/applier.go
# 8. byte-frozen siblings (expect: empty diff)
git diff --stat -- internal/harness/regression_gate.go internal/harness/outcome.go internal/harness/observer.go internal/measure/measure.go
```

## §F Milestones

### M1 — RED→GREEN: errors.Join preserves the typed regression signal (single milestone)

Given the 2-file scope, a single milestone is appropriate.

1. **RED** — add `TestApply_Outcome_RolledBack_RecordError` to `applier_test.go`:
   - Construct a regressing scenario reaching the rolled-back branch: seed a prior
     baseline (so `hasPriorBaseline` is true), inject a `seqMeasurer` whose candidate
     measurement is worse than baseline (coverage drop + lint up), matching the existing
     `TestApply_Regression_Blocks_RollsBack` setup (applier_test.go:921).
   - Wire a **deliberately-failing** `*Observer` via `WithOutcomeObserver`: construct an
     `Observer` whose `logPath` parent directory cannot be created — e.g. create a
     regular file at path `P` then set `logPath = filepath.Join(P, "usage-log.jsonl")`,
     so `os.MkdirAll(filepath.Dir(o.logPath), 0o755)` inside `RecordExtendedEvent`
     fails (the parent `P` is a file, not a dir). This makes `recordOutcome` return a
     non-nil error on the rolled-back branch.
   - Assert (RED, pre-fix these FAIL):
     - (a) the snapshot was rolled back — target file restored to original bytes
       (`bytes.Equal(originalContent, after)`).
     - (b) `errors.As(err, &regErr)` is `true` for `var regErr *ApplyRegressionError`
       (the F2 regression-prevention assertion).
     - (c) the outcome-record error is also reachable — assert the returned error's
       message contains the observer failure substring (e.g. `"파일 열기 실패"` /
       `"observer:"`), or use `errors.Is`/unwrap to reach it.
   - Verify the test FAILS for the right reason before the fix (assertion (b) fails
     because pre-fix the branch returns only the wrapped `oerr`).

2. **GREEN** — apply the fix in `applier.go` rolled-back branch (lines 448-451):
   - Add `"errors"` to the import block.
   - Replace the rolled-back `recordOutcome`-error handling so that on `oerr != nil` it
     returns `errors.Join(&ApplyRegressionError{Baseline: baseline, Candidate: candidate,
     Regressed: regressed}, oerr)`; on `oerr == nil` it continues to the existing bare
     `return &ApplyRegressionError{…}`.
   - Concretely: hoist `regErr := &ApplyRegressionError{Baseline: baseline, Candidate:
     candidate, Regressed: regressed}`; on outcome error return
     `errors.Join(regErr, oerr)`; otherwise `return regErr`.
   - kept branch (lines 462-470) untouched.

3. **VERIFY** — run §E batch; new test GREEN, all prior rolled-back tests
   (`TestApply_Regression_Blocks_RollsBack`, `_AppendsBlockedLineage`,
   `_MeasurementError_FailsClosed`) still GREEN (REQ-ERRJOIN-003/005 guard — the
   bare-error success path is unchanged).

## §G Anti-Patterns to Avoid

- **AP-1 — Symmetrizing the kept branch.** Do NOT add `errors.Join` to the kept branch.
  Success == nil; there is no typed sentinel to preserve there (REQ-ERRJOIN-004 / C-ERRJOIN-03).
- **AP-2 — Using the direct-assertion helper.** Do NOT assert via
  `asRegressionError(...)` for the new test — it uses a direct type assertion that the
  joined error fails. Use `errors.As(err, &regErr)` directly.
- **AP-3 — Editing a frozen sibling.** Do NOT touch `outcome.go` / `observer.go` to
  "make the observer fail more cleanly" — fail it from the TEST by constructing an
  un-creatable `logPath` parent. The siblings are byte-frozen (C-ERRJOIN-02).
- **AP-4 — Overclaiming in CHANGELOG.** Per §A.2, the sync-phase CHANGELOG MUST state
  this is a dormant-scaffold error-propagation correctness fix (0 production callers),
  not an activation or a production regression-prevention change.

## §H Cross-References

- Predecessor: `SPEC-HARNESS-OUTCOME-CAPTURE-001` (this is its F2 follow-up).
- Sibling: `SPEC-HARNESS-REGRESSION-GATE-001` (introduced `ApplyRegressionError` +
  the rolled-back branch).
- Fix site: `internal/harness/applier.go:448-451` (rolled-back branch); import block
  `internal/harness/applier.go:7-14`.
- Test home: `internal/harness/applier_test.go` (alongside `TestApply_Regression_*`).
