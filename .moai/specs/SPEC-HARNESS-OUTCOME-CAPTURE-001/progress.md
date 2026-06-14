---
id: SPEC-HARNESS-OUTCOME-CAPTURE-001
title: "Progress — Harness Apply outcome capture"
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
tags: "harness, outcome-capture, progress"
era: V3R6
---

## §F.1 Plan-phase

### Tier + artifact set

- **Tier**: M (standard) — 3-artifact set (spec.md + plan.md + acceptance.md) + this progress.md.
- **Rationale**: brownfield extension of the observer/Event (~250-400 LOC, ~4 files: types.go additive fields + new outcome.go + applier.go composition seam + tests). The Apply-seam composition (DD-2) and the representation choice (DD-1) are genuine non-mechanical decisions in a recently-landed tested pipeline, warranting plan.md design documentation. NOT Tier S (under-documents the seam decision). NOT Tier L (no new package, no FROZEN change, no safety layer, one new exported const).

### Roadmap position

- Self-Harness P2/Phase5 FIRST sub-item: the observer OUTCOME-capture enabler.
- Predecessors completed: P0 (LOOP-CLOSURE-001, M6 lineage + M7 human-gate), P1 (REGRESSION-GATE-001, M2-lite gate — supplies the verdict + delta reused here).
- Downstream consumers (D1 clustering, canary-effectiveness, held-out, scorer-loop) are OUT OF SCOPE — each a separate future SPEC.

### Ground-truth corrections applied (authored from verified file reads, not premise)

- `internal/harness/observer.go` EXISTS — `RecordEvent` + `RecordExtendedEvent` append `Event` structs to `usage-log.jsonl`. The new `RecordOutcome` reuses `RecordExtendedEvent` (DD-6).
- `internal/harness/regression_gate.go` (P1) ALREADY landed: `MetricTriple`, `Regressions`, `ApplyRegressionError`, `BaselineStore`, `goMeasurer`, and `applyWithRegressionGate` are in place. The outcome record REUSES these (no re-creation).
- The Apply pipeline ALREADY emits lineage with `Decision ∈ {"approved","rejected","regression-blocked"}`. The outcome capture is a SEPARATE observer write (the lineage manifest is the human-audit record; the observer log is the machine-analysis substrate — B.2).
- **No `lineage_id` field exists** — `LineageEntry` is keyed by `ProposalID` + `Timestamp`. The outcome reuses `proposal_id` as the correlation key; no `lineage_id` is introduced (DD-3 / C11).
- The GAP this SPEC closes: the `Event` struct has NO OUTCOME fields, and there is NO `RecordOutcome` path surfacing an Apply's verdict+delta+proposal_id into `usage-log.jsonl`. Confirmed by grep (no `RecordOutcome`/`Outcome` capture path in `internal/harness/*.go`).

### Design decisions (folded in plan.md §F)

- **DD-1** Representation: additive `omitempty` fields on `Event` + new `EventTypeApplyOutcome` const (follows the established Stop/SubagentStop/UserPromptSubmit omitempty pattern). Alternatives (b) separate log file / (c) extend `LineageEntry` rejected (substrate fragmentation / completed-schema churn).
- **DD-2** Wiring point: compose + emit inside `applyWithRegressionGate`, AFTER the decision, in BOTH kept + rolled-back branches; nil-observer no-op; emit-error wrapping mirrors `writeLineage`.
- **DD-3** Correlation key: reuse `proposal_id`; do NOT add `lineage_id` (ground-truth: none exists; C11 forbids lineage-schema churn).
- **DD-4** Gate-inactive path emits no outcome (no measured triple) — pre-P1 behavior untouched.
- **DD-5** Honest framing (MUST-PASS): capture+persist ONLY; typically Δ=0; no "improves"/"prevents".
- **DD-6** `RecordOutcome` reuses `RecordExtendedEvent` (no new low-level writer).

### Self-assessment

| Metric | Value |
|--------|-------|
| Tier | M |
| REQ count | 13 (REQ-OC-001 … REQ-OC-013) |
| HARD constraints | 12 (C1 … C12) |
| AC count | 12 (AC-OC-001 … AC-OC-012) |
| Files affected (run-phase est.) | ~4: `types.go` (additive), `outcome.go` (new), `applier.go` (seam), `outcome_test.go` (new) + `applier_test.go` (additive tests) |
| LOC est. (run-phase) | ~250-400 |
| Representation choice | DD-1 option (a): additive `omitempty` Event fields + `EventTypeApplyOutcome` |
| Wiring choice | DD-2: in-`applyWithRegressionGate` post-decision emit, both branches |

### Open questions (for plan-auditor / GATE-2)

- **OQ1 (observer injection signature)**: thread the observer through `NewApplierWithRegressionGate`'s existing signature vs add a separate `WithOutcomeObserver(*Observer)` option. Run-phase detail; SPEC requires only that the gate-active production seam can emit and that a nil observer is a safe no-op. No SPEC-level blocker.
- **OQ2 (omitempty zero-drop)**: `int`/`float64` omitempty drops genuine zero triples on a `kept` Δ=0 outcome. plan.md DD-1 argues the `outcome_verdict` field (always non-empty) is the load-bearing signal, so the dropped zeros are acceptable. If a future consumer needs the explicit zeros, it can drop omitempty on those fields then (additive, backward-compatible). Surfaced for plan-auditor judgment — not a blocker.

plan_complete_at: 2026-06-14T00:00:00Z
plan_status: audit-ready
