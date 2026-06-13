---
id: SPEC-HARNESS-REGRESSION-GATE-001
title: "Acceptance Criteria — Harness M2-lite 비회귀 게이트"
version: "0.1.1"
status: draft
created: 2026-06-14
updated: 2026-06-14
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/harness, internal/measure"
lifecycle: spec-anchored
tier: M
tags: "harness, regression-gate, acceptance"
---

## A. Overview

Acceptance criteria for SPEC-HARNESS-REGRESSION-GATE-001 in GEARS Given-When-Then form. Each AC carries a concrete `go test -run '<pattern>' -v` command AND an anti-vacuous guard (`grep '--- PASS'` and `! grep 'no tests to run'`).

[HARD] Run-phase MUST create test functions whose names exactly match the `-run` regexes below. The patterns are anchored to avoid infix/substring mismatch (`L_ac_run_pattern_vacuous_guard`): e.g. `-run 'TestApplyRegression_Blocks$'` matches `func TestApplyRegression_Blocks` but the `$` anchor prevents accidental over-match.

The anti-vacuous guard idiom (applied to every AC):

```bash
OUT=$(go test -run '<pattern>' -v ./<pkg>/ 2>&1)
echo "$OUT" | grep -q -- '--- PASS' || { echo "FAIL: no PASS line"; exit 1; }
echo "$OUT" | grep -q 'no tests to run' && { echo "FAIL: vacuous — pattern matched nothing"; exit 1; }
echo "AC PASS"
```

---

## B. Acceptance Criteria

### AC-RG-001 — Leaf package parsers are pure and exported (REQ-RG-002, C9)

- **Given** the `internal/measure` leaf package exists with `ParseGoTestJSON`, `ParseCoverageFile`, `CountNonEmptyLines`,
- **When** the unit tests run and `go list -deps` is inspected,
- **Then** all three exported parsers produce correct results AND the package imports none of `internal/lsp`, `internal/lsp/gopls`, `internal/harness`, `internal/loop`.

```bash
OUT=$(go test -run 'TestParseGoTestJSON|TestParseCoverageFile|TestCountNonEmptyLines' -v ./internal/measure/ 2>&1)
echo "$OUT" | grep -q -- '--- PASS' || { echo "FAIL: no PASS"; exit 1; }
echo "$OUT" | grep -q 'no tests to run' && { echo "FAIL: vacuous"; exit 1; }
# Import-cycle / forbidden-dep proof (DD-1)
go list -deps github.com/modu-ai/moai-adk/internal/measure | grep -E 'internal/(lsp|gopls|harness|loop)' \
  && { echo "FAIL: forbidden dependency present"; exit 1; } || echo "CLEAN: no forbidden deps"
echo "AC-RG-001 PASS"
```

### AC-RG-002 — `internal/loop` delegates without behavior change (REQ-RG-003, R4)

- **Given** `GoFeedbackGenerator` now delegates parsing to `internal/measure`,
- **When** the existing loop feedback tests run,
- **Then** they stay GREEN (byte-identical parsing behavior, no regression).

```bash
OUT=$(go test -run 'TestGoFeedback|TestParse|TestCollect' -v ./internal/loop/ 2>&1)
echo "$OUT" | grep -q -- '--- PASS' || { echo "FAIL: no PASS"; exit 1; }
echo "$OUT" | grep -q 'no tests to run' && { echo "FAIL: vacuous"; exit 1; }
echo "AC-RG-002 PASS"
```

### AC-RG-003 — Baseline store atomic write + absent-file first-run (REQ-RG-004, REQ-RG-005)

- **Given** the baseline store at `.moai/harness/measurements-baseline.yaml`,
- **When** the store is read on first run (absent file) and then written atomically,
- **Then** an absent file yields a no-block "treat candidate as baseline" outcome AND the written file round-trips the metric triple via temp-file+rename.

```bash
OUT=$(go test -run 'TestBaselineStore_AbsentFile$|TestBaselineStore_AtomicRoundTrip$' -v ./internal/harness/ 2>&1)
echo "$OUT" | grep -q -- '--- PASS' || { echo "FAIL: no PASS"; exit 1; }
echo "$OUT" | grep -q 'no tests to run' && { echo "FAIL: vacuous"; exit 1; }
echo "AC-RG-003 PASS"
```

### AC-RG-004 — `ApplyRegressionError` is distinct and carries metrics (REQ-RG-008, C8, DD-3)

- **Given** the new `ApplyRegressionError` type,
- **When** it is constructed with baseline, candidate, and regressed dimensions,
- **Then** its `Error()` string names the regressed dimensions AND it is a distinct type from `ApplyPendingError`.

```bash
OUT=$(go test -run 'TestApplyRegressionError_Error$|TestApplyRegressionError_DistinctFromPending$' -v ./internal/harness/ 2>&1)
echo "$OUT" | grep -q -- '--- PASS' || { echo "FAIL: no PASS"; exit 1; }
echo "$OUT" | grep -q 'no tests to run' && { echo "FAIL: vacuous"; exit 1; }
# Confirm both error types coexist
grep -q 'type ApplyRegressionError struct' internal/harness/applier.go || grep -q 'type ApplyRegressionError struct' internal/harness/regression_gate.go || { echo "FAIL: new type missing"; exit 1; }
grep -q 'type ApplyPendingError struct' internal/harness/applier.go || { echo "FAIL: existing type clobbered"; exit 1; }
echo "AC-RG-004 PASS"
```

### AC-RG-005 — In-Apply gate keeps non-regressing change + updates baseline (REQ-RG-007, REQ-RG-009)

- **Given** an Apply that reached `DecisionApproved` and measures `Δ ≥ 0` for all dimensions,
- **When** the gate runs (measure baseline → snapshot → apply → measure candidate → compare),
- **Then** the change is kept, the baseline store is updated to the candidate triple, and the existing M6 `"approved"` lineage entry is written.

```bash
OUT=$(go test -run 'TestApply_Regression_NonRegressing_Keeps$' -v ./internal/harness/ 2>&1)
echo "$OUT" | grep -q -- '--- PASS' || { echo "FAIL: no PASS"; exit 1; }
echo "$OUT" | grep -q 'no tests to run' && { echo "FAIL: vacuous"; exit 1; }
echo "AC-RG-005 PASS"
```

### AC-RG-006 — In-Apply gate rolls back regressing change + returns error + audit (REQ-RG-008, REQ-RG-010)

- **Given** an Apply that reached `DecisionApproved` but a regression is injected (tests_passed decreased OR coverage decreased OR lint_count increased) via a stub measurer,
- **When** the gate compares the candidate to the baseline,
- **Then** `RestoreSnapshot` is called (file rolled back to original), `ApplyRegressionError` is returned, AND a `"regression-blocked"` lineage entry is appended.

```bash
OUT=$(go test -run 'TestApply_Regression_Blocks_RollsBack$|TestApply_Regression_AppendsBlockedLineage$' -v ./internal/harness/ 2>&1)
echo "$OUT" | grep -q -- '--- PASS' || { echo "FAIL: no PASS"; exit 1; }
echo "$OUT" | grep -q 'no tests to run' && { echo "FAIL: vacuous"; exit 1; }
echo "AC-RG-006 PASS"
```

### AC-RG-007 — Subagent boundary preserved (REQ-RG-011, C7)

- **Given** the regression gate code in `internal/harness/`,
- **When** the C-HRA-008 binary guard runs,
- **Then** no `AskUserQuestion(` / `mcp__askuser(` call-site exists in harness/hook Go source.

```bash
OUT=$(go test -run 'TestSubagentBoundary_NoAskUserQuestion$' -v ./internal/harness/ 2>&1)
echo "$OUT" | grep -q -- '--- PASS' || { echo "FAIL: no PASS"; exit 1; }
echo "$OUT" | grep -q 'no tests to run' && { echo "FAIL: vacuous"; exit 1; }
echo "AC-RG-007 PASS"
```

### AC-RG-008 — FROZEN invariants unchanged (REQ-RG-012, C1-C6)

- **Given** the DO-NOT-MODIFY files,
- **When** the FROZEN preservation tests run and `git diff` is checked,
- **Then** all preservation tests stay GREEN AND the DO-NOT-MODIFY files show zero diff.

```bash
# NOTE (D1 fix, L_grep_ac_substring_collision): run each package NON-RECURSIVELY.
# A recursive ./internal/harness/... emits "[no tests to run]" for sibling sub-packages
# (capture/proposalgen/router/seeds/throttle/tier) which a blanket negative guard would
# misread as vacuous. Scope each -run to its owning package and assert per-package.
# harness-package preservation tests (TestSafetyArchitecture_*, TestSentinelCatalog_*)
OUT1=$(go test -run 'TestSafetyArchitecture_LayerCount$|TestSafetyArchitecture_FrozenZoneUnchanged$|TestSentinelCatalog' -v ./internal/harness/ 2>&1)
echo "$OUT1" | grep -q -- '--- PASS' || { echo "FAIL: harness preservation no PASS"; exit 1; }
echo "$OUT1" | grep -q 'no tests to run' && { echo "FAIL: harness preservation vacuous"; exit 1; }
# safety-package frozen tests (TestIsFrozen_* live in internal/harness/safety)
OUT2=$(go test -run 'TestIsFrozen' -v ./internal/harness/safety/ 2>&1)
echo "$OUT2" | grep -q -- '--- PASS' || { echo "FAIL: safety frozen no PASS"; exit 1; }
echo "$OUT2" | grep -q 'no tests to run' && { echo "FAIL: safety frozen vacuous"; exit 1; }
# tier tests
go test ./internal/harness/tier/ 2>&1 | grep -q '^ok\|^PASS' || { echo "FAIL: tier tests"; exit 1; }
# DO-NOT-MODIFY git-diff must be empty
DIFF=$(git diff --stat internal/harness/frozen_guard.go internal/harness/safety/frozen_guard.go internal/harness/tier/tier.go internal/harness/scorer.go .moai/config/sections/harness.yaml)
[ -z "$DIFF" ] || { echo "FAIL: FROZEN file modified: $DIFF"; exit 1; }
echo "AC-RG-008 PASS"
```

### AC-RG-009 — Honest framing documented (REQ-RG-013, C10, MUST-PASS)

- **Given** the SPEC artifacts,
- **When** spec.md / plan.md are inspected,
- **Then** they explicitly document that the measured delta is typically `Δ=0` for the current markdown-only write surface AND frame the gate as (1) measurement scaffold + (2) defense-in-depth, NOT as an active current-operation regression preventer.

```bash
SPEC=.moai/specs/SPEC-HARNESS-REGRESSION-GATE-001/spec.md
grep -qi 'typically.*Δ=0\|typically.*delta.*0\|always-pass' "$SPEC" || { echo "FAIL: missing Δ=0 honest disclosure"; exit 1; }
grep -qi 'measurement-infrastructure scaffold\|measurement scaffold' "$SPEC" || { echo "FAIL: missing scaffold framing"; exit 1; }
grep -qi 'defense-in-depth' "$SPEC" || { echo "FAIL: missing defense-in-depth framing"; exit 1; }
echo "AC-RG-009 PASS"
```

### AC-RG-010 — `auto_apply` and human gate unchanged (REQ-RG-012, C5)

- **Given** the harness config,
- **When** `auto_apply` is inspected,
- **Then** it remains `false` (L5 human gate preserved; no autonomous apply enabled).

```bash
grep -q 'auto_apply: false' .moai/config/sections/harness.yaml || { echo "FAIL: auto_apply changed"; exit 1; }
echo "AC-RG-010 PASS"
```

### AC-RG-011 — Measurement collector assembles the metric triple (REQ-RG-001)

- **Given** the measurement collector built on the `internal/measure` parsers,
- **When** `Collect` runs against a project state,
- **Then** it returns a `{tests_passed, coverage, lint_count}` triple assembled from `ParseGoTestJSON` / `ParseCoverageFile` / `CountNonEmptyLines`.

```bash
OUT=$(go test -run 'TestCollector_AssemblesTriple$' -v ./internal/harness/ 2>&1)
echo "$OUT" | grep -q -- '--- PASS' || { echo "FAIL: no PASS"; exit 1; }
echo "$OUT" | grep -q 'no tests to run' && { echo "FAIL: vacuous"; exit 1; }
echo "AC-RG-011 PASS"
```

### AC-RG-012 — Gate does not touch forbidden runtime files (REQ-RG-006, C11)

- **Given** an Apply executed through the regression gate,
- **When** the gate runs (including the regression-block audit append),
- **Then** `usage-log.jsonl`, `observations.yaml`, and `tier-promotions.jsonl` are unmodified, and the lineage `manifest.jsonl` is only appended to (never truncated/rewritten).

```bash
OUT=$(go test -run 'TestApply_Regression_ForbiddenFilesUntouched$' -v ./internal/harness/ 2>&1)
echo "$OUT" | grep -q -- '--- PASS' || { echo "FAIL: no PASS"; exit 1; }
echo "$OUT" | grep -q 'no tests to run' && { echo "FAIL: vacuous"; exit 1; }
echo "AC-RG-012 PASS"
```

### AC-RG-013 — Measurement failure fails closed (REQ-RG-014)

- **Given** an Apply at `DecisionApproved` where the measurement step cannot execute (stubbed build/exec error or timeout),
- **When** the gate attempts to measure baseline or candidate,
- **Then** the gate returns a wrapped measurement error (fail-closed) and does NOT keep the change — distinct from the valid-count path where a still-running-but-red suite yields a number and is compared normally.

```bash
OUT=$(go test -run 'TestApply_Regression_MeasurementError_FailsClosed$' -v ./internal/harness/ 2>&1)
echo "$OUT" | grep -q -- '--- PASS' || { echo "FAIL: no PASS"; exit 1; }
echo "$OUT" | grep -q 'no tests to run' && { echo "FAIL: vacuous"; exit 1; }
echo "AC-RG-013 PASS"
```

---

## C. Edge Cases

- **EC-1 — First run, no baseline file**: candidate becomes baseline, Apply not blocked (AC-RG-003).
- **EC-2 — Markdown-only Apply (the real case today)**: `Δ=0` for all dimensions → non-regressing → kept (AC-RG-005). This is the typical path and honestly yields no protection (it is the always-pass case per DD-7).
- **EC-3 — Snapshot dir capture**: `createSnapshot` must surface its dated `snapshotDir` so `RestoreSnapshot` can roll back the exact backup (DD-2 note). Tested via the rollback AC (AC-RG-006).
- **EC-4 — Lineage write failure during regression-block**: the rollback + error return are the primary effects; a failed audit append is wrapped but does not suppress the regression block (mirrors existing `writeLineage` error semantics).

## D. Definition of Done

- [ ] All 13 ACs PASS (AC-RG-001 … AC-RG-013).
- [ ] `go test ./internal/harness/... ./internal/loop/... ./internal/measure/...` GREEN.
- [ ] `go list -deps internal/measure` shows no `internal/(lsp|gopls|harness|loop)` (import-cycle proof).
- [ ] DO-NOT-MODIFY files: zero git diff.
- [ ] `golangci-lint run --timeout=2m` clean for touched packages.
- [ ] Coverage ≥ existing baseline for `internal/harness` + new `internal/measure` package ≥ 85%.
- [ ] C-HRA-008 boundary GREEN.
- [ ] Honest framing (DD-7 / AC-RG-009) — plan-auditor confirms no vacuous-gate overstatement.

## E. Quality Gate Criteria

- TRUST 5: Tested (new ACs + preservation GREEN), Readable (mirrors existing error/lineage patterns), Unified (gofmt/golangci-lint), Secured (no new attack surface; baseline file is local 0o644), Trackable (lineage audit append).
- MUST-PASS dimension: **honest framing** (REQ-RG-013 / C10). A SPEC that overstates the gate's current-surface protective value fails this gate regardless of other scores.
