---
id: SPEC-HARNESS-OUTCOME-CAPTURE-001
title: "Implementation Plan — Harness Apply outcome capture"
version: "0.1.0"
status: draft
created: 2026-06-14
updated: 2026-06-14
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/harness"
lifecycle: spec-anchored
tier: M
tags: "harness, outcome-capture, plan"
---

## A. Context

Run-phase implementation plan for SPEC-HARNESS-OUTCOME-CAPTURE-001. cycle_type: `tdd`. This SPEC is a **brownfield extension** of the existing observer/Event: it adds a structured Apply-OUTCOME record (regression-gate verdict + project-health delta + `proposal_id` correlation key) that is recorded into the observer's `usage-log.jsonl` via the existing `RecordExtendedEvent` append path. It does NOT create a greenfield observer; it extends the one in `internal/harness/observer.go` + `types.go` and adds one composition seam in `applyWithRegressionGate` (`applier.go`).

The Design section (§F) is folded into this plan.md per the Tier M classification — the non-trivial decisions (representation, Apply-seam wiring point, backward compatibility, gate-inactive behavior, honest framing) are documented here so plan-auditor can verify they were considered.

## B. Known Issues / Risks

- **R1 — Apply-contract non-interference (most important)**: P1's `applyWithRegressionGate` is a recently-landed, tested pipeline with a precise keep/rollback contract (returns `nil` on keep, `*ApplyRegressionError` on rollback, plus baseline-store update + lineage write). The outcome capture MUST be purely additive — it composes from values the gate already computed and emits one extra observer write, never altering the decision or the returned error (REQ-OC-006 / C10). Mitigation: the outcome emit is placed AFTER the gate has decided in each branch; an emit error follows the existing additive-write error-wrapping pattern (mirror `writeLineage`'s "primary effect already happened" semantics).
- **R2 — Backward-compat of the JSONL schema**: existing readers (`learner.go` `AggregatePatterns`, `retention.go`, `cli/harness.go`) consume `Event` lines. New `omitempty` fields + a new `EventTypeApplyOutcome` must not break them. Mitigation: additive omitempty fields are ignored by readers that do not read them; an unknown event type is aggregated by key or skipped, never fatal (verify with a reader-tolerance test, AC-OC-006). `LogSchemaVersion` bump `"v1"`→`"v2"` marks the additive set (the established convention).
- **R3 — `lineage_id` does not exist (ground-truth correction)**: the roadmap referred to "lineage_id"; the actual lineage record (`LineageEntry`) has no such field — it is keyed by `ProposalID` + `Timestamp`. DD-3 reuses `proposal_id` as the correlation key and explicitly does NOT add a `lineage_id` field to the completed P1/P0 lineage type (C11). This avoids churning a completed SPEC's schema.
- **R4 — Δ=0 honest framing (MUST-PASS)**: the captured delta is typically `Δ=0` for the current markdown-only write surface (FROZEN allowlist). The SPEC must frame the value as capture+persist for downstream analysis, never as "improves"/"prevents" (REQ-OC-013 / C12 / DD-5). plan-auditor will flag overclaim.
- **R5 — Gate-inactive path ambiguity**: the `NewApplier()` default (no measurer/baseline store) runs the straight-line markdown modify path and does not reach `applyWithRegressionGate`. DD-4 resolves whether that path also emits an outcome (chosen: emit a `kept`/Δ=0 outcome only when an observer is wired; otherwise skip — no observer dependency is forced on the gate-inactive constructor). Bound by AC-OC-007.

## C. Pre-flight (verify before M1)

```bash
# Confirm the composition seam + reused types are as the SPEC states
grep -n 'func (a \*Applier) applyWithRegressionGate' internal/harness/applier.go      # gate flow (kept + rolled-back branches)
grep -n 'type MetricTriple struct\|func (base MetricTriple) Regressions\|type ApplyRegressionError struct' internal/harness/regression_gate.go
grep -n 'func (o \*Observer) RecordExtendedEvent\|func (o \*Observer) RecordEvent' internal/harness/observer.go
grep -n 'EventTypeApplyOutcome\|LogSchemaVersion' internal/harness/types.go             # EventTypeApplyOutcome expected ABSENT (new); LogSchemaVersion = "v1"
# Confirm NO existing Outcome capture (gap)
grep -rn 'RecordOutcome\|EventTypeApplyOutcome' internal/harness/*.go | grep -v '_test.go' || echo "GAP CONFIRMED: no outcome capture path"
# Confirm FROZEN invariants present (re-asserted GREEN in acceptance)
grep -n 'auto_apply: false' .moai/config/sections/harness.yaml
go test ./internal/harness/... 2>&1 | tail -5   # green baseline (incl. P1 regression-gate tests)
```

## D. Constraints (from spec.md §D)

C1-C12 (see spec.md §D). The DO-NOT-MODIFY files are: `frozen_guard.go` (both), `tier/tier.go`, `scorer.go`, `harness.yaml`, `subagent_boundary_test.go`, and `lineage.go` (the lineage schema is NOT extended — C11). The Apply keep/rollback contract is unchanged (C10).

---

## E. Self-Verification (orchestrator read-only batch — run at run-phase completion)

```bash
# 1. Full harness suite (incl. P1 regression-gate preservation)
go test ./internal/harness/... 2>&1 | tail -10
# 2. New outcome capture tests
go test -run 'TestRecordOutcome|TestApply_Outcome' -v ./internal/harness/ 2>&1 | tail -20
# 3. C-HRA-008 boundary
grep -rn 'AskUserQuestion(\|mcp__askuser(' internal/harness/ internal/hook/ | grep -v '_test.go' | grep -v '^[^:]*:[0-9]*:[ \t]*//' || echo "CLEAN"
# 4. New event type present, existing ones untouched; lineage schema untouched
grep -n 'EventTypeApplyOutcome\|LogSchemaVersion' internal/harness/types.go
git diff --stat internal/harness/lineage.go    # expect EMPTY (lineage schema not extended)
# 5. FROZEN unchanged (git diff against DO-NOT-MODIFY files)
git diff --stat internal/harness/frozen_guard.go internal/harness/safety/frozen_guard.go internal/harness/tier/tier.go internal/harness/scorer.go .moai/config/sections/harness.yaml
# 6. Backward-compat: existing readers still GREEN
go test -run 'TestAggregate|TestLearner|TestRetention' -v ./internal/harness/ 2>&1 | tail -10
# 7. Coverage + lint
go test -coverprofile=/tmp/oc.out ./internal/harness/... && go tool cover -func=/tmp/oc.out | tail -1
golangci-lint run --timeout=2m ./internal/harness/...
```

---

## F. Design (folded into plan.md — Tier M)

### DD-1 — Representation: additive `omitempty` fields on `Event` + new `EventType` (option a)

**Decision**: encode the outcome as an `Event` of a new event type `EventTypeApplyOutcome = "apply_outcome"`, with the outcome data carried in additive `omitempty` fields on the existing `Event` struct. The outcome record type (`OutcomeRecord`, a small struct holding verdict/triples/regressed/proposal_id/decision) is the in-memory composition value; `RecordOutcome` maps it onto an `Event` and delegates to `RecordExtendedEvent`.

**Why this is lowest-drift**: the `Event` struct ALREADY uses exactly this pattern for Stop / SubagentStop / UserPromptSubmit (per-event-type `omitempty` fields, REQ-HRN-OBS-009: "new fields are additive omitempty only"). Following the established pattern means: (a) `RecordExtendedEvent` already serializes the full `Event` with omitempty — no new low-level writer; (b) existing readers ignore omitempty fields they don't read — backward-compatible; (c) one time-ordered stream holds both WHAT-events and the OUTCOME, which is exactly what a Phase5 correlator needs.

Proposed additive fields (all `omitempty`, only set on `apply_outcome` events):

```go
// ── ApplyOutcome event optional fields (SPEC-HARNESS-OUTCOME-CAPTURE-001) ──
OutcomeVerdict      string  `json:"outcome_verdict,omitempty"`        // "kept" | "rolled-back"
OutcomeDecision     string  `json:"outcome_decision,omitempty"`       // "approved" | "regression-blocked"
OutcomeProposalID   string  `json:"outcome_proposal_id,omitempty"`    // lineage correlation key
OutcomeBaseTests    int     `json:"outcome_baseline_tests,omitempty"`
OutcomeBaseCoverage float64 `json:"outcome_baseline_coverage,omitempty"`
OutcomeBaseLint     int     `json:"outcome_baseline_lint,omitempty"`
OutcomeCandTests    int     `json:"outcome_candidate_tests,omitempty"`
OutcomeCandCoverage float64 `json:"outcome_candidate_coverage,omitempty"`
OutcomeCandLint     int     `json:"outcome_candidate_lint,omitempty"`
OutcomeRegressed    []string `json:"outcome_regressed,omitempty"`     // e.g. ["coverage"]
```

(Exact field names/grouping are a run-phase detail; the run-phase MUST keep them additive `omitempty` and MUST NOT reorder/rename existing fields — C9. `int`/`float64` omitempty drops genuine zeros; that is acceptable here because a `kept` Δ=0 outcome is still identifiable by `outcome_verdict: "kept"` + the absent triple fields meaning "no change measured" — the verdict, not the numeric fields, is the load-bearing signal.)

**Alternatives rejected**:
- **(b) dedicated `Outcome` struct + separate outcome log file** (`.moai/harness/apply-outcomes.jsonl`) — REJECTED. A separate file fragments the analysis substrate: a Phase5 correlator would have to join two logs by timestamp to relate a usage pattern to its outcome. The single `usage-log.jsonl` stream is the substrate `learner.go` already reads; keeping the outcome there is lower drift for the consumer. A separate file also duplicates the append/mkdir/pruning machinery.
- **(c) extend `LineageEntry` with the delta + a `lineage_id`** — REJECTED. `LineageEntry` belongs to the completed P0 SPEC; adding fields churns a completed schema (C11) and, more importantly, the lineage manifest is the human-audit record, NOT the machine-analysis substrate the Phase5 consumers read. The two records serve different purposes; conflating them couples human-audit to machine-analysis.

### DD-2 — Wiring point: compose + emit inside `applyWithRegressionGate`, after the gate decides

**Decision**: the outcome is composed and recorded inside `applyWithRegressionGate` (`applier.go`), in BOTH terminal branches, AFTER the gate has fully decided:

```
applyWithRegressionGate:
  ... measure baseline → apply → measure candidate → compare ...
  if regression:
      RestoreSnapshot(snapshotDir)            # (existing — unchanged)
      writeLineage(..., "regression-blocked") # (existing — unchanged)
      recordOutcome(verdict="rolled-back",     # (NEW — additive, after decision)
                    decision="regression-blocked",
                    baseline, candidate, regressed, proposal_id)
      return &ApplyRegressionError{...}        # (existing — unchanged)
  else:  # non-regressing or first-run
      baselineStore.Save(candidate)            # (existing — unchanged)
      writeLineage(..., "approved")            # (existing — unchanged)
      recordOutcome(verdict="kept",            # (NEW — additive, after decision)
                    decision="approved",
                    baseline, candidate, regressed=nil, proposal_id)
      return nil                               # (existing — unchanged)
```

**Why inside the gate flow (not in `Apply`)**: the verdict + baseline triple + candidate triple + regressed dimensions are local to `applyWithRegressionGate`. Composing the outcome there reuses those values directly. The straight-line gate-INACTIVE path (`Apply` → `applyFileModification`) does not have a measured triple — DD-4 governs that path separately.

**Why AFTER the decision in each branch**: the capture must never influence the decision (REQ-OC-006 / C10). Placing the `recordOutcome` call after `RestoreSnapshot`/`writeLineage` in the rollback branch and after `baselineStore.Save`/`writeLineage` in the keep branch guarantees the decision + returned error are already fixed before the observer write happens.

**Observer injection**: the `Applier` gains an optional `outcomeObserver *Observer` (or a minimal `outcomeRecorder` interface) field, injected by the production constructor (`NewApplierWithRegressionGate`) and by tests. When nil, `recordOutcome` is a no-op (preserves callers that did not opt in — mirrors the existing `manifestPath == "" → skip` pattern for lineage). Run-phase decides whether to thread it through `NewApplierWithRegressionGate`'s signature or add a separate `WithOutcomeObserver` option; the SPEC requires only that the gate-active production seam can emit, and that a nil observer is a safe no-op.

**Emit-error semantics (REQ-OC-007)**: a `recordOutcome` failure is wrapped consistently with `writeLineage`'s pattern — the decision (keep/rollback) and the returned error stand; the observer-write failure is surfaced as a wrapped error on the same return, never flipping the verdict. (Run-phase mirrors the exact `writeLineage`-error wrapping idiom already in `applyWithRegressionGate`.)

### DD-3 — Correlation key: reuse `proposal_id`, do NOT add `lineage_id`

**Decision**: the outcome's lineage correlation key is `proposal_id` (`Proposal.ID`) + the transition `decision` string. No `lineage_id` field is introduced anywhere.

**Justification (ground-truth)**: `LineageEntry` (types.go:399) has fields `ProposalID`, `TargetPath`, `AppliedSurface`, `Decision`, `Timestamp`, `Reason` — there is **no `lineage_id`**. Lineage entries are correlated by `ProposalID` + write-order `Timestamp`. The outcome record reuses the same `ProposalID`, so a Phase5 consumer can join an `apply_outcome` observer event to its `manifest.jsonl` lineage entry by `proposal_id` + `decision`. Introducing a `lineage_id` would (a) require changing the completed P0 lineage schema (C11 forbids) and (b) add a synthetic key where `proposal_id` already serves. If a richer lineage key is ever genuinely needed, it is a separate lineage-schema SPEC (recorded in Exclusions).

### DD-4 — Gate-inactive path behavior (REQ-OC-008)

**Decision**: the outcome capture is emitted ONLY when the gate is active (`applyWithRegressionGate`) AND an observer is injected. The straight-line gate-INACTIVE markdown modify path (`Apply` → `applyFileModification`, used by `NewApplier()`) does NOT emit an outcome by default.

**Justification**: the gate-inactive path has no measured triple (no baseline, no candidate) — fabricating a Δ=0 outcome there would invent numbers the path never measured, which contradicts the honest-framing bar. The gate-inactive path is the pre-P1 behavior; it is preserved untouched. If a future SPEC wants outcomes on the gate-inactive path, it can add an optional "verdict=kept, delta=unmeasured" emit — explicitly out of scope here. AC-OC-007 binds this: the gate-inactive `Apply` produces no `apply_outcome` event.

### DD-5 — Honest framing rationale (REQ-OC-013, MUST-PASS)

This SPEC delivers **capture + persist ONLY**. The captured delta is typically `Δ=0` for the current markdown-only write surface (FROZEN allowlist → harness edits markdown frontmatter, which cannot move Go test/coverage/lint counts). The enabler's two honest values:

1. **Observability** — harness Apply outcomes (verdict + delta + proposal_id) become persisted in `usage-log.jsonl`, where they were previously transient (consumed by the gate, audited only in the human-review lineage manifest).
2. **Phase5 substrate** — the recorded stream is the precondition that D1 clustering + canary-effectiveness will read. Those consumers are downstream and OUT OF SCOPE.

The SPEC MUST NOT frame the capture as "improving" the harness or "preventing" anything — it is a passive observer write of an already-decided outcome. This is the MUST-PASS quality bar; plan-auditor will flag any overclaim.

### DD-6 — `RecordOutcome` reuses `RecordExtendedEvent` (no new low-level writer)

**Decision**: `RecordOutcome(rec OutcomeRecord) error` maps the outcome record onto an `Event{EventType: EventTypeApplyOutcome, ...outcome omitempty fields...}` and delegates to the existing `RecordExtendedEvent`. No new file-handling code is written.

**Justification**: `RecordExtendedEvent` (observer.go:103) already does the full job — fill default Timestamp/SchemaVersion, auto-mkdir, JSONL marshal, `O_APPEND|O_CREATE|O_WRONLY` append, lazy retention pruning. Reusing it means the outcome write inherits all the observer's correctness (atomic append, pruning) for free and stays consistent with every other event write. Duplicating the append machinery would risk drift (a future fix to one path missing the other).

---

## G. Milestones

| M | Goal | Files | Verify |
|---|------|-------|--------|
| M1 | Add `EventTypeApplyOutcome` const + additive `omitempty` outcome fields on `Event` + bump `LogSchemaVersion` "v1"→"v2". RED: write a serialization round-trip test first. | `internal/harness/types.go`, `internal/harness/outcome_test.go` (new) | `go test -run 'TestApplyOutcomeEvent_RoundTrip' -v ./internal/harness/` ; `git diff --stat internal/harness/lineage.go` empty |
| M2 | Add `OutcomeRecord` type + `RecordOutcome` observer method delegating to `RecordExtendedEvent` (DD-1/DD-6). RED first. | `internal/harness/outcome.go` (new), `internal/harness/outcome_test.go` | `go test -run 'TestRecordOutcome' -v ./internal/harness/` |
| M3 | Wire the composition seam into `applyWithRegressionGate` — emit outcome in BOTH kept + rolled-back branches AFTER the decision; nil-observer no-op; emit-error wrapping mirrors `writeLineage` (DD-2/DD-4). RED first. | `internal/harness/applier.go`, `internal/harness/applier_test.go` | `go test -run 'TestApply_Outcome_Kept|TestApply_Outcome_RolledBack|TestApply_Outcome_GateInactive_NoEmit' -v ./internal/harness/` |
| M4 | Backward-compat: assert existing readers tolerate the new event type + fields; assert P1 regression-gate keep/rollback contract unchanged. | `internal/harness/outcome_test.go` (reader-tolerance) | `go test -run 'TestApplyOutcome_ReaderTolerance|TestApply_Regression_' -v ./internal/harness/` |
| M5 | FROZEN preservation re-assert + C-HRA-008 + lineage-schema-untouched + full suite + coverage + lint. | (tests only) | `go test ./internal/harness/...` ; FROZEN + lineage git-diff stat clean ; `golangci-lint run --timeout=2m` |

Milestones are priority-ordered (no time estimates). M1→M2 build the record + recorder; M3 wires the seam; M4 proves non-interference + backward-compat; M5 is the preservation/quality gate.

## H. Anti-Patterns to avoid (predecessor lessons)

- `L_ac_run_pattern_vacuous_guard` — every acceptance `go test -run '<pattern>'` MUST match a real test name (anchored, no infix/substring mismatch) and MUST guard against vacuous "[no tests to run]".
- `L_grep_ac_substring_collision` — outcome field names (`outcome_*`) must not collide with existing grep patterns; run package-scoped, anchored greps.
- Over-claim / honest-framing breach — the enabler captures+persists ONLY; no "improves"/"prevents" language (DD-5, MUST-PASS).
- Apply-contract drift — the outcome emit must be purely additive; do NOT alter keep/rollback, the returned error, the baseline-store update, or the lineage write (C10).
- Lineage-schema churn — do NOT add `lineage_id` to `LineageEntry` or modify `lineage.go` (C11).
- Over-engineering — no separate outcome log file, no query CLI, no Phase5 consumer (all out of scope). The enabler is the minimal capture path.

## I. Cross-References

See spec.md §E for the full file:line + DO-NOT-MODIFY table. Predecessors: SPEC-HARNESS-LOOP-CLOSURE-001 (P0, completed — M6 lineage), SPEC-HARNESS-REGRESSION-GATE-001 (P1, completed — verdict + delta source).
