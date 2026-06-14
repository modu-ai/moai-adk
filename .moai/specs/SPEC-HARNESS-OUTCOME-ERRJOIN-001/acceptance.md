---
id: SPEC-HARNESS-OUTCOME-ERRJOIN-001
title: "Apply rolled-back branch errors.Join — acceptance criteria"
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

## §D Acceptance Criteria Matrix

All AC are testable, anchored, and anti-vacuous. Grep-based AC use explicit literals
(no `^\s*//` non-greedy idiom, per the C-HRA-008 plan-audit pitfall).

| AC ID | REQ | Severity | Assertion |
|-------|-----|----------|-----------|
| AC-ERRJOIN-001 | REQ-ERRJOIN-001 | MUST-PASS | New test: rolled-back + failing observer → `errors.As(err, &regErr)` is TRUE. |
| AC-ERRJOIN-002 | REQ-ERRJOIN-002 | MUST-PASS | Same test: outcome-record error reachable via `errors.Is`/unwrap or message substring. |
| AC-ERRJOIN-003 | REQ-ERRJOIN-003 | MUST-PASS | recordOutcome-SUCCESS rolled-back path still yields a bare `*ApplyRegressionError`. |
| AC-ERRJOIN-004 | REQ-ERRJOIN-004 | MUST-PASS | kept branch wording byte-unchanged (grep + git diff). |
| AC-ERRJOIN-005 | REQ-ERRJOIN-005 | MUST-PASS | Snapshot rollback + `regression-blocked` lineage unaffected (existing tests GREEN). |
| AC-ERRJOIN-006 | C-ERRJOIN-02 | MUST-PASS | Scope discipline: only `applier.go` + `applier_test.go` changed. |
| AC-ERRJOIN-007 | C-ERRJOIN-04 | MUST-PASS | `go test ./...` exit 0; `go vet ./...` 0 findings; `internal/harness` coverage no-regression. |
| AC-ERRJOIN-008 | C-ERRJOIN-05 | SHOULD | Only new import is stdlib `errors`; `errors.Join` present at fix site. |

## §D.1 Detailed Given-When-Then Scenarios

### Scenario 1 — Typed signal preserved when outcome-record write fails (AC-ERRJOIN-001 / -002)

New test `TestApply_Outcome_RolledBack_RecordError` in `internal/harness/applier_test.go`:

- **Given** a gated Applier with a seeded prior baseline (`hasPriorBaseline` true), a
  `seqMeasurer` returning a baseline triple then a regressed candidate triple (coverage
  drop + lint up, as in `TestApply_Regression_Blocks_RollsBack`), AND a
  deliberately-failing `*Observer` wired via `WithOutcomeObserver` — the observer's
  `logPath` parent is an EXISTING REGULAR FILE so `os.MkdirAll(filepath.Dir(logPath))`
  inside `RecordExtendedEvent` fails and `recordOutcome` returns a non-nil error.
- **When** `Apply(proposal, approvedEvaluator(), snapshotBase, nil)` runs and reaches
  the rolled-back branch.
- **Then**:
  - (a) the target file is rolled back to its original bytes
    (`bytes.Equal(originalContent, after)` is true);
  - (b) **`errors.As(err, &regErr)` is TRUE** for `var regErr *ApplyRegressionError`,
    and `regErr.Regressed` is non-empty (the F2 regression-prevention assertion);
  - (c) the outcome-record error is also reachable — the returned error's message
    contains the observer-failure substring (`observer:` and/or `디렉토리 생성 실패`,
    the `os.MkdirAll`-origin text that fails first when the logPath parent is a regular
    file), OR `errors.Is`/manual unwrap reaches the observer error (preferred).
- **RED first**: assertion (b) MUST FAIL before the fix (pre-fix the branch returns only
  the `fmt.Errorf("… outcome record failed: %w", oerr)` wrapper, which is NOT
  `*ApplyRegressionError` and whose single unwrap target is `oerr`, so `errors.As` for
  `*ApplyRegressionError` returns false). After the fix it PASSES.

> Note: the test MUST use `errors.As(err, &regErr)` directly. It MUST NOT use the
> existing `asRegressionError(err, &regErr)` helper (applier_test.go:1125), which uses a
> direct type assertion `err.(*ApplyRegressionError)` that the joined error correctly
> fails — that distinction is the whole point of the contract being restored.

### Scenario 2 — No regression on the recordOutcome-SUCCESS rolled-back path (AC-ERRJOIN-003 / -005)

- **Given** the existing `TestApply_Regression_Blocks_RollsBack` (no observer wired, or a
  succeeding observer) reaching the rolled-back branch.
- **When** `Apply(...)` runs.
- **Then** the returned error is a bare `*ApplyRegressionError`
  (`asRegressionError` / `errors.As` both TRUE, `regErr.Regressed` non-empty), the file
  is rolled back, the baseline store is unchanged, and the `regression-blocked` lineage
  entry is appended — all exactly as before this SPEC.
- This is satisfied by the existing rolled-back tests continuing to PASS unchanged
  (REQ-ERRJOIN-003 guard); optionally extend one to assert the no-observer path yields a
  bare typed error with no joined second error.

## §D.2 Verification Commands (read-only, run as one parallel batch)

```bash
# AC-ERRJOIN-001/-002/-003/-005/-007 — new + existing tests
go test ./internal/harness/... -run 'TestApply_Outcome_RolledBack_RecordError|TestApply_Regression'
go test ./...

# AC-ERRJOIN-007 — vet + coverage
go vet ./...
go test -coverprofile=cover.out ./internal/harness/...

# AC-ERRJOIN-008 — errors.Join at fix site + errors import
grep -n 'errors.Join' internal/harness/applier.go
grep -n '"errors"' internal/harness/applier.go

# AC-ERRJOIN-004 — kept-branch wording preserved
grep -n 'file modified but outcome record failed' internal/harness/applier.go

# AC-ERRJOIN-006 — scope discipline (exactly two files), trailing-$ anchored
git diff --name-only | grep -E '^internal/harness/applier(_test)?\.go$'
git diff --name-only | grep -vE '^internal/harness/applier(_test)?\.go$'   # expect: empty

# AC-ERRJOIN-006 (sibling freeze) — expect empty diff
git diff --stat -- internal/harness/regression_gate.go internal/harness/outcome.go internal/harness/observer.go internal/measure/measure.go
```

## §D.3 Definition of Done

- [ ] `TestApply_Outcome_RolledBack_RecordError` added, demonstrated RED→GREEN.
- [ ] `errors.As(err, &*ApplyRegressionError)` TRUE on rolled-back + failing-observer (AC-001).
- [ ] Outcome-record error also reachable (AC-002).
- [ ] recordOutcome-SUCCESS rolled-back path yields bare `*ApplyRegressionError` (AC-003).
- [ ] kept branch byte-unchanged; wording grep present (AC-004).
- [ ] Existing `TestApply_Regression_*` GREEN — rollback/lineage unaffected (AC-005).
- [ ] `git diff --name-only` lists ONLY `applier.go` + `applier_test.go` (AC-006).
- [ ] Frozen siblings show empty diff (AC-006).
- [ ] `go test ./...` exit 0; `go vet ./...` 0; `internal/harness` coverage ≥ baseline (AC-007).
- [ ] Only new import is stdlib `errors`; `errors.Join` at fix site (AC-008).

## §D.4 Quality Gate

| Gate | Threshold | Source |
|------|-----------|--------|
| Test suite | `go test ./...` exit 0 | AC-ERRJOIN-007 |
| Static analysis | `go vet ./...` 0 findings | AC-ERRJOIN-007 |
| Coverage | `internal/harness` no-regression vs baseline | AC-ERRJOIN-007 |
| Scope | exactly 2 files changed | AC-ERRJOIN-006 |
| Dependency | no new third-party import | C-ERRJOIN-05 |
