---
id: SPEC-V3R6-I18N-VALIDATOR-BUDGET-001
title: "i18n-validator TestBudget Threshold 30s → 35s — acceptance"
version: "1.0.0"
status: implemented
created: 2026-05-24
updated: 2026-05-24
author: manager-spec
priority: P3
phase: "v3.0.0"
module: "scripts/i18n-validator"
lifecycle: spec-anchored
tags: "i18n, test, budget, tier-s, lcl-001-followup"
tier: S
---

# Acceptance Criteria — SPEC-V3R6-I18N-VALIDATOR-BUDGET-001

This file is the SSOT for AC enumeration. spec.md §B-REQ references AC IDs; AC bodies live here only.

## AC-IVB-001 — Budget threshold raised to 35 seconds at line 376

**Given** `scripts/i18n-validator/main_test.go` is at HEAD post-run.
**When** the orchestrator runs:
```bash
grep -n "if elapsed > 35\*time.Second" scripts/i18n-validator/main_test.go
```
**Then** output **shall** contain exactly one line with line number `376` (or the equivalent line if comment/blank line shifts moved it by ≤ 2 lines).

**And when** the orchestrator runs:
```bash
grep -cn "if elapsed > 30\*time.Second" scripts/i18n-validator/main_test.go
```
**Then** output **shall** be `0` (zero stale 30-second threshold references in the budget test).

**Anchors**: REQ-IVB-001.
**Severity**: Critical (must pass for SPEC completion).

## AC-IVB-002 — Function renamed and JP godoc comment updated

**Given** `scripts/i18n-validator/main_test.go` is at HEAD post-run.
**When** the orchestrator runs:
```bash
grep -n "TestBudget_FullRepoScanWithin35Sec" scripts/i18n-validator/main_test.go
```
**Then** output **shall** contain exactly 2 lines:
- one godoc comment line containing `35秒以内に完了することを検証します` (Japanese, line ≈ 359)
- one function declaration line `func TestBudget_FullRepoScanWithin35Sec(t *testing.T) {` (line ≈ 360)

**And when** the orchestrator runs:
```bash
grep -cn "TestBudget_FullRepoScanWithin30Sec\|30秒以内" scripts/i18n-validator/main_test.go
```
**Then** output **shall** be `0` (no stale `Within30Sec` function references or `30秒以内` Japanese comment fragments remain in the source file).

**Anchors**: REQ-IVB-002.
**Severity**: Major (naming coherence; degraded developer experience if fails but no functional impact).

## AC-IVB-003 — Renamed test passes locally with new threshold

**Given** the edits per AC-IVB-001 + AC-IVB-002 are applied and committed.
**When** the orchestrator runs:
```bash
go test -timeout 60s -v -run TestBudget_FullRepoScanWithin35Sec \
  ./scripts/i18n-validator/...
```
**Then** stdout **shall** contain:
- `=== RUN   TestBudget_FullRepoScanWithin35Sec`
- `--- PASS: TestBudget_FullRepoScanWithin35Sec (<duration>s)` where `<duration>` parses as a positive float strictly less than 35.0
- `PASS` on a standalone line
- `ok  scripts/i18n-validator` package summary line

**And** exit code **shall** be `0`.

**Anchors**: REQ-IVB-003.
**Severity**: Critical (must pass for SPEC completion).

## AC-IVB-004 — Elapsed time recorded in progress evidence

**Given** AC-IVB-003 has been verified.
**When** the orchestrator reads `.moai/specs/SPEC-V3R6-I18N-VALIDATOR-BUDGET-001/progress.md` §Run-phase Evidence section.
**Then** the section **shall** contain a row or paragraph stating the recorded elapsed time of `TestBudget_FullRepoScanWithin35Sec` from the AC-IVB-003 invocation, formatted as `elapsed = <X.YYs>` or equivalent (the exact figure from `--- PASS: TestBudget_FullRepoScanWithin35Sec (X.YYs)` output).

**And** the recorded value **shall** be strictly less than `35.00s`.

**And if** the recorded value is greater than or equal to `33.00s` (94 % of new budget per Risk-E1 mitigation), **then** the progress.md row **shall** include a `WARNING` annotation noting the proximity to the new ceiling and recommending a future optimization SPEC.

**Anchors**: REQ-IVB-003, Risk-E1 mitigation.
**Severity**: Major (evidence preservation; PASS-WITH-DEBT clearance traceability).

## AC-IVB-005 — No regression in other i18n-validator tests

**Given** AC-IVB-001 + AC-IVB-002 edits are applied and committed.
**When** the orchestrator runs:
```bash
go test -timeout 90s -count=1 ./scripts/i18n-validator/... 2>&1 | tee /tmp/ivb-full.log
```
**Then** the package summary line **shall** read `ok  scripts/i18n-validator <duration>s` (exit code 0).

**And** the PASS count for non-budget tests (any test name NOT matching `^TestBudget`) **shall** equal the baseline PRE_PASS count captured before the edits (manager-develop M1 step 1 pre-flight should capture this baseline via `go test -timeout 90s -count=1 ./scripts/i18n-validator/... 2>&1 | grep -cE '^--- PASS'` minus 2 budget-test PASS lines).

**Anchors**: REQ-IVB-003, scope discipline (no collateral regression).
**Severity**: Critical (regression detection; LCL-001 PASS-WITH-DEBT pattern must not recur in same file under this SPEC).

## Definition of Done

All 5 ACs marked PASS in `.moai/specs/SPEC-V3R6-I18N-VALIDATOR-BUDGET-001/progress.md` §Run-phase Evidence section with independently verifiable command outputs captured (either inline or referenced /tmp log paths).

## Out of scope (no AC)

- `defaultBudget = 30 * time.Second` constant in `scripts/i18n-validator/main.go` (production validator default; unchanged per §A.3).
- `TestBudget_TimeoutExitOnExcess` at line 382 (1ns synthetic budget; unchanged per §A.3).
- Archival SPEC body references to the old function name (immutable per D-3).
