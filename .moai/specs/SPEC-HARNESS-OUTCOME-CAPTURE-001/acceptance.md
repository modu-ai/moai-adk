---
id: SPEC-HARNESS-OUTCOME-CAPTURE-001
title: "Acceptance Criteria — Harness Apply outcome capture"
version: "0.1.0"
status: completed
created: 2026-06-14
updated: 2026-06-14
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/harness"
lifecycle: spec-anchored
tier: M
tags: "harness, outcome-capture, acceptance"
---

## A. Overview

Acceptance criteria for SPEC-HARNESS-OUTCOME-CAPTURE-001 in GEARS Given-When-Then form. Each AC carries a concrete `go test -run '<pattern>' -v` command AND an anti-vacuous guard (`grep '--- PASS'` and `! grep 'no tests to run'`).

[HARD] Run-phase MUST create test functions whose names exactly match the `-run` regexes below. Patterns are `$`-anchored to avoid infix/substring mismatch (`L_ac_run_pattern_vacuous_guard`): e.g. `-run 'TestRecordOutcome_Kept$'` matches `func TestRecordOutcome_Kept` but the `$` anchor prevents accidental over-match (it will not also match `TestRecordOutcome_KeptAndAudited`).

The anti-vacuous guard idiom (applied to every AC):

```bash
OUT=$(go test -run '<pattern>' -v ./<pkg>/ 2>&1)
echo "$OUT" | grep -q -- '--- PASS' || { echo "FAIL: no PASS line"; exit 1; }
echo "$OUT" | grep -q 'no tests to run' && { echo "FAIL: vacuous — pattern matched nothing"; exit 1; }
echo "AC PASS"
```

---

## B. Acceptance Criteria

### AC-OC-001 — Outcome event type + additive omitempty fields round-trip (REQ-OC-001, REQ-OC-004, C8, C9)

- **Given** the new `EventTypeApplyOutcome` constant and the additive `omitempty` outcome fields on `Event`,
- **When** an `apply_outcome` `Event` is JSON-marshaled and unmarshaled,
- **Then** it round-trips the verdict / decision / proposal_id / baseline+candidate triples / regressed list, AND non-outcome `Event` values (e.g. a `moai_subcommand` event) marshal WITHOUT any `outcome_*` keys (omitempty proven).

```bash
OUT=$(go test -run 'TestApplyOutcomeEvent_RoundTrip$|TestApplyOutcomeEvent_OmitemptyOnOtherEvents$' -v ./internal/harness/ 2>&1)
echo "$OUT" | grep -q -- '--- PASS' || { echo "FAIL: no PASS"; exit 1; }
echo "$OUT" | grep -q 'no tests to run' && { echo "FAIL: vacuous"; exit 1; }
# New const present; existing consts untouched (additive only)
grep -q 'EventTypeApplyOutcome\s*EventType\s*=\s*"apply_outcome"' internal/harness/types.go || { echo "FAIL: EventTypeApplyOutcome missing/wrong value"; exit 1; }
grep -q 'EventTypeMoaiSubcommand\s*EventType\s*=\s*"moai_subcommand"' internal/harness/types.go || { echo "FAIL: existing const clobbered"; exit 1; }
echo "AC-OC-001 PASS"
```

### AC-OC-002 — LogSchemaVersion bumped to v2 (REQ-OC-010)

- **Given** the additive outcome field set,
- **When** `LogSchemaVersion` is inspected,
- **Then** it is `"v2"` (marking the additive set), and newly recorded events carry `schema_version: "v2"`.

```bash
grep -q 'LogSchemaVersion\s*=\s*"v2"' internal/harness/types.go || { echo "FAIL: LogSchemaVersion not bumped to v2"; exit 1; }
OUT=$(go test -run 'TestApplyOutcomeEvent_SchemaVersionV2$' -v ./internal/harness/ 2>&1)
echo "$OUT" | grep -q -- '--- PASS' || { echo "FAIL: no PASS"; exit 1; }
echo "$OUT" | grep -q 'no tests to run' && { echo "FAIL: vacuous"; exit 1; }
echo "AC-OC-002 PASS"
```

### AC-OC-003 — `RecordOutcome` writes one JSONL line via the existing append path (REQ-OC-003, DD-6)

- **Given** an `Observer` bound to a temp `usage-log.jsonl` and an `OutcomeRecord`,
- **When** `RecordOutcome` is called,
- **Then** exactly one `apply_outcome` JSONL line is appended (parent dir auto-created), carrying the verdict + proposal_id + triples — proving delegation to the shared `RecordExtendedEvent` append machinery.

```bash
OUT=$(go test -run 'TestRecordOutcome_AppendsOneLine$|TestRecordOutcome_AutoCreatesDir$' -v ./internal/harness/ 2>&1)
echo "$OUT" | grep -q -- '--- PASS' || { echo "FAIL: no PASS"; exit 1; }
echo "$OUT" | grep -q 'no tests to run' && { echo "FAIL: vacuous"; exit 1; }
echo "AC-OC-003 PASS"
```

### AC-OC-004 — In-Apply seam emits a `kept` outcome on a non-regressing Apply (REQ-OC-002, REQ-OC-005)

- **Given** an active regression gate (stub measurer + baseline store + injected observer) where the candidate Δ ≥ 0 (kept),
- **When** `Apply` runs through `applyWithRegressionGate` and keeps the change,
- **Then** the gate's return value is unchanged (`nil`) AND exactly one `apply_outcome` event is recorded with `outcome_verdict: "kept"`, `outcome_decision: "approved"`, and the matching `proposal_id`.

```bash
OUT=$(go test -run 'TestApply_Outcome_Kept$' -v ./internal/harness/ 2>&1)
echo "$OUT" | grep -q -- '--- PASS' || { echo "FAIL: no PASS"; exit 1; }
echo "$OUT" | grep -q 'no tests to run' && { echo "FAIL: vacuous"; exit 1; }
echo "AC-OC-004 PASS"
```

### AC-OC-005 — In-Apply seam emits a `rolled-back` outcome on a regressing Apply (REQ-OC-002, REQ-OC-005)

- **Given** an active regression gate where the candidate regresses (stub measurer injects a decreased coverage), so the gate rolls back and returns `*ApplyRegressionError`,
- **When** `Apply` runs through `applyWithRegressionGate`,
- **Then** the gate's return value is unchanged (still `*ApplyRegressionError`, snapshot still restored, `"regression-blocked"` lineage still written) AND exactly one `apply_outcome` event is recorded with `outcome_verdict: "rolled-back"`, `outcome_decision: "regression-blocked"`, the regressed-dimension list, and the matching `proposal_id`.

```bash
OUT=$(go test -run 'TestApply_Outcome_RolledBack$' -v ./internal/harness/ 2>&1)
echo "$OUT" | grep -q -- '--- PASS' || { echo "FAIL: no PASS"; exit 1; }
echo "$OUT" | grep -q 'no tests to run' && { echo "FAIL: vacuous"; exit 1; }
echo "AC-OC-005 PASS"
```

### AC-OC-006 — Apply contract unchanged + readers tolerate the new event (REQ-OC-006, REQ-OC-009, C10)

- **Given** the outcome capture wired into `applyWithRegressionGate` with an outcome observer **INJECTED** (the observer-active seam — non-interference must be proven with capture ACTIVE, not only against the bare P1 Applier),
- **When** the observer-active keep/rollback tests (`TestApply_Outcome_Kept` / `TestApply_Outcome_RolledBack`, which inject an observer) run AND the P1 regression-gate keep/rollback contract tests run (regression backstop) AND a usage-log containing `apply_outcome` events is fed to the existing reader (`AggregatePatterns` / retention),
- **Then** the observer-active tests assert the returned decision/error MATCH the P1 keep/rollback contract (non-interference proven with capture active) AND the P1 contract tests stay GREEN (decision/error/baseline-store/lineage unchanged) AND the existing reader processes the log without error (unknown event type + new omitempty fields tolerated).

```bash
# Non-interference proven with the observer ACTIVE (load-bearing — not just the green-by-default P1 tests)
OUT0=$(go test -run 'TestApply_Outcome_Kept$|TestApply_Outcome_RolledBack$' -v ./internal/harness/ 2>&1)
echo "$OUT0" | grep -q -- '--- PASS' || { echo "FAIL: observer-active non-interference no PASS"; exit 1; }
echo "$OUT0" | grep -q 'no tests to run' && { echo "FAIL: observer-active non-interference vacuous"; exit 1; }
# P1 contract preserved (regression backstop — the outcome emit must not break keep/rollback)
OUT1=$(go test -run 'TestApply_Regression_NonRegressing_Keeps$|TestApply_Regression_Blocks_RollsBack$' -v ./internal/harness/ 2>&1)
echo "$OUT1" | grep -q -- '--- PASS' || { echo "FAIL: P1 contract regressed"; exit 1; }
echo "$OUT1" | grep -q 'no tests to run' && { echo "FAIL: P1 contract vacuous"; exit 1; }
# Reader tolerance for the new event type + fields
OUT2=$(go test -run 'TestApplyOutcome_ReaderTolerance$' -v ./internal/harness/ 2>&1)
echo "$OUT2" | grep -q -- '--- PASS' || { echo "FAIL: reader tolerance no PASS"; exit 1; }
echo "$OUT2" | grep -q 'no tests to run' && { echo "FAIL: reader tolerance vacuous"; exit 1; }
echo "AC-OC-006 PASS"
```

### AC-OC-007 — Gate-inactive path emits no outcome; nil-observer is a no-op (REQ-OC-008, DD-4, DD-2)

- **Given** the gate-INACTIVE Apply path (`NewApplier()` default, no measurer/baseline store) OR an active gate with a nil observer,
- **When** `Apply` runs,
- **Then** no `apply_outcome` event is recorded (gate-inactive path has no measured triple — DD-4) AND a nil observer makes `recordOutcome` a safe no-op (no panic, Apply outcome unchanged).

```bash
OUT=$(go test -run 'TestApply_Outcome_GateInactive_NoEmit$|TestApply_Outcome_NilObserver_NoOp$' -v ./internal/harness/ 2>&1)
echo "$OUT" | grep -q -- '--- PASS' || { echo "FAIL: no PASS"; exit 1; }
echo "$OUT" | grep -q 'no tests to run' && { echo "FAIL: vacuous"; exit 1; }
echo "AC-OC-007 PASS"
```

### AC-OC-008 — Outcome emit-error does not flip the verdict (REQ-OC-007)

- **Given** an active gate that kept (or rolled back) a change, but the injected observer's write fails (stub observer returns an I/O error),
- **When** `applyWithRegressionGate` attempts to record the outcome,
- **Then** the decided verdict is NOT flipped — a kept Apply stays kept (change not undone), a rolled-back Apply stays rolled back — and the observer-write failure is surfaced as a wrapped error consistent with the existing `writeLineage`-error semantics (the primary effect already happened).

```bash
OUT=$(go test -run 'TestApply_Outcome_RecordError_DoesNotFlipVerdict$' -v ./internal/harness/ 2>&1)
echo "$OUT" | grep -q -- '--- PASS' || { echo "FAIL: no PASS"; exit 1; }
echo "$OUT" | grep -q 'no tests to run' && { echo "FAIL: vacuous"; exit 1; }
echo "AC-OC-008 PASS"
```

### AC-OC-009 — Lineage schema + correlation key unchanged (REQ-OC-006, C11, DD-3)

- **Given** the outcome capture reuses `proposal_id` as the correlation key,
- **When** `lineage.go` / `LineageEntry` are inspected,
- **Then** no `lineage_id` field was added, `lineage.go` shows zero git diff, and the `apply_outcome` event carries `outcome_proposal_id` matching the lineage entry's `proposal_id` for the same Apply.

```bash
# No lineage_id introduced; lineage.go untouched
grep -q 'lineage_id\|LineageID' internal/harness/lineage.go internal/harness/types.go && { echo "FAIL: lineage_id introduced (C11 violation)"; exit 1; } || echo "OK: no lineage_id"
DIFF=$(git diff --stat internal/harness/lineage.go)
[ -z "$DIFF" ] || { echo "FAIL: lineage.go modified: $DIFF"; exit 1; }
# Correlation key reuse proven in a test
OUT=$(go test -run 'TestApply_Outcome_ProposalIDMatchesLineage$' -v ./internal/harness/ 2>&1)
echo "$OUT" | grep -q -- '--- PASS' || { echo "FAIL: no PASS"; exit 1; }
echo "$OUT" | grep -q 'no tests to run' && { echo "FAIL: vacuous"; exit 1; }
echo "AC-OC-009 PASS"
```

### AC-OC-010 — Subagent boundary preserved (REQ-OC-011, C7)

- **Given** the outcome capture code in `internal/harness/`,
- **When** the C-HRA-008 binary guard runs,
- **Then** no `AskUserQuestion(` / `mcp__askuser(` call-site exists in harness/hook Go source.

```bash
OUT=$(go test -run 'TestSubagentBoundary_NoAskUserQuestion$' -v ./internal/harness/ 2>&1)
echo "$OUT" | grep -q -- '--- PASS' || { echo "FAIL: no PASS"; exit 1; }
echo "$OUT" | grep -q 'no tests to run' && { echo "FAIL: vacuous"; exit 1; }
echo "AC-OC-010 PASS"
```

### AC-OC-011 — FROZEN invariants + `auto_apply` unchanged (REQ-OC-012, C1-C6)

- **Given** the DO-NOT-MODIFY files,
- **When** the FROZEN preservation tests run and `git diff` is checked,
- **Then** all preservation tests stay GREEN AND the DO-NOT-MODIFY files show zero diff AND `auto_apply` remains `false`.

```bash
# NOTE (L_grep_ac_substring_collision): run each package NON-RECURSIVELY and
# scope each -run to its owning package so sibling sub-packages do not emit a
# misread "[no tests to run]".
OUT1=$(go test -run 'TestSafetyArchitecture_LayerCount$|TestSafetyArchitecture_FrozenZoneUnchanged$|TestSentinelCatalog' -v ./internal/harness/ 2>&1)
echo "$OUT1" | grep -q -- '--- PASS' || { echo "FAIL: harness preservation no PASS"; exit 1; }
echo "$OUT1" | grep -q 'no tests to run' && { echo "FAIL: harness preservation vacuous"; exit 1; }
OUT2=$(go test -run 'TestIsFrozen' -v ./internal/harness/safety/ 2>&1)
echo "$OUT2" | grep -q -- '--- PASS' || { echo "FAIL: safety frozen no PASS"; exit 1; }
echo "$OUT2" | grep -q 'no tests to run' && { echo "FAIL: safety frozen vacuous"; exit 1; }
go test ./internal/harness/tier/ 2>&1 | grep -q '^ok\|^PASS' || { echo "FAIL: tier tests"; exit 1; }
DIFF=$(git diff --stat internal/harness/frozen_guard.go internal/harness/safety/frozen_guard.go internal/harness/tier/tier.go internal/harness/scorer.go .moai/config/sections/harness.yaml)
[ -z "$DIFF" ] || { echo "FAIL: FROZEN file modified: $DIFF"; exit 1; }
grep -q 'auto_apply: false' .moai/config/sections/harness.yaml || { echo "FAIL: auto_apply changed"; exit 1; }
echo "AC-OC-011 PASS"
```

### AC-OC-012 — Honest framing documented (REQ-OC-013, C12, MUST-PASS)

- **Given** the SPEC artifacts,
- **When** spec.md / plan.md are inspected,
- **Then** they explicitly document that this SPEC is capture+persist ONLY, that the captured delta is typically `Δ=0` for the current markdown-only write surface, that the consuming analysis (clustering / canary-effectiveness) is downstream and out of scope, AND they make no "improves"/"prevents" claim about the enabler.

```bash
SPEC=.moai/specs/SPEC-HARNESS-OUTCOME-CAPTURE-001/spec.md
grep -qi 'capture *+ *persist\|capture+persist' "$SPEC" || { echo "FAIL: missing capture+persist framing"; exit 1; }
grep -qi 'typically.*Δ=0\|typically.*delta.*0' "$SPEC" || { echo "FAIL: missing Δ=0 honest disclosure"; exit 1; }
grep -qi 'downstream' "$SPEC" || { echo "FAIL: missing downstream-consumer disclosure"; exit 1; }
# MUST-NOT overclaim: the enabler section must NOT assert it "improves" or "prevents".
# (A literal scan; run-phase keeps spec.md prose free of "this SPEC improves"/"this SPEC prevents".)
grep -niE 'this (spec|enabler) (improves|prevents)' "$SPEC" && { echo "FAIL: overclaim detected"; exit 1; } || echo "OK: no overclaim"
echo "AC-OC-012 PASS"
```

---

## C. Edge Cases

- **EC-1 — Markdown-only Apply (the real case today)**: gate-active, `Δ=0`, kept → one `apply_outcome` event with `outcome_verdict: "kept"` and equal baseline/candidate triples (numeric fields may be omitempty-dropped when zero, but the verdict is the load-bearing signal). This is the typical path and honestly records "kept, no measured change" (AC-OC-004).
- **EC-2 — First run (no baseline)**: gate adopts the candidate as baseline without blocking (P1 REQ-RG-005) → outcome is `kept` / `approved` (regressed list empty). The capture must not double-count or treat first-run as a regression (covered by AC-OC-004 with a first-run fixture).
- **EC-3 — Gate-inactive markdown modify (`NewApplier()`)**: no measured triple → no `apply_outcome` event (DD-4 / AC-OC-007). The pre-P1 straight-line path is untouched.
- **EC-4 — Observer write failure**: the verdict is not flipped; the error is wrapped (AC-OC-008), mirroring the existing `writeLineage`-error semantics.
- **EC-5 — Reader on a mixed log**: an existing reader (`AggregatePatterns`) reading a log with both legacy events and `apply_outcome` events processes it without error (AC-OC-006).

## D. Definition of Done

- [ ] All 12 ACs PASS (AC-OC-001 … AC-OC-012).
- [ ] `go test ./internal/harness/...` GREEN (incl. P1 regression-gate preservation tests).
- [ ] `LogSchemaVersion = "v2"`; `EventTypeApplyOutcome = "apply_outcome"` present additively.
- [ ] `lineage.go` zero git diff; no `lineage_id` field introduced.
- [ ] DO-NOT-MODIFY files (frozen_guard ×2, tier, scorer, harness.yaml): zero git diff.
- [ ] `golangci-lint run --timeout=2m ./internal/harness/...` clean for touched packages.
- [ ] Coverage ≥ existing baseline for `internal/harness`.
- [ ] C-HRA-008 boundary GREEN.
- [ ] Honest framing (DD-5 / AC-OC-012) — plan-auditor confirms no overclaim (capture+persist ONLY).

## E. Quality Gate Criteria

- TRUST 5: Tested (12 ACs + P1/FROZEN preservation GREEN), Readable (mirrors existing observer/Event omitempty + writeLineage error patterns), Unified (gofmt/golangci-lint), Secured (no new attack surface; outcome write reuses the existing local-file observer path), Trackable (outcome is an observable persisted record; reuses `proposal_id` correlation key).
- MUST-PASS dimension: **honest framing** (REQ-OC-013 / C12). A SPEC that frames the enabler as "improving" or "preventing" anything, rather than capture+persist for downstream analysis, fails this gate regardless of other scores.
