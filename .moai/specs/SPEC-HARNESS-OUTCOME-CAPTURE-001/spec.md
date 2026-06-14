---
id: SPEC-HARNESS-OUTCOME-CAPTURE-001
title: "Harness Apply outcome capture (observer OUTCOME enabler)"
version: "0.1.0"
status: implemented
created: 2026-06-14
updated: 2026-06-14
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/harness"
lifecycle: spec-anchored
tags: "harness, observer, outcome-capture, self-harness, phase5-enabler"
tier: M
era: V3R6
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-14 | manager-spec | Initial draft — Self-Harness 로드맵 P2/Phase5 첫 sub-item. 기존 observer/Event(WHAT 캡처)를 brownfield 확장해 매 harness Apply의 OUTCOME(regression-gate verdict + project-health delta + lineage 상관키)을 구조화 기록. Predecessor P0 SPEC-HARNESS-LOOP-CLOSURE-001(M6 lineage + M7 human-gate, completed) + P1 SPEC-HARNESS-REGRESSION-GATE-001(M2-lite 비회귀 게이트, completed) 위에 얹는 enabler. |

---

## A. Overview

### A.1 Purpose

This SPEC closes the **observer OUTCOME-capture gap**. The predecessor SPECs delivered:

- **P0** (SPEC-HARNESS-LOOP-CLOSURE-001, completed) — M6 auditable lineage (`WriteLineageEntry` → `manifest.jsonl`) + M7 human-gated loop.
- **P1** (SPEC-HARNESS-REGRESSION-GATE-001, completed) — the M2-lite non-regression gate (`MetricTriple`, `Measurer`, `BaselineStore`, `ApplyRegressionError`, `applyWithRegressionGate`). The gate produces an Apply **verdict** (kept | rolled-back) and a **project-health delta** (baseline triple vs candidate triple).

The existing observer (`internal/harness/observer.go`, `RecordEvent` / `RecordExtendedEvent` → `.moai/harness/usage-log.jsonl`) captures **WHAT happened** — `/moai` subcommand invocations, agent invocations, SPEC references, Stop / SubagentStop / UserPromptSubmit events. The `Event` struct (`types.go`) records the event subject and metadata, but it carries **no field for the OUTCOME of a harness Apply**. The regression gate's verdict + delta exist transiently (in the `applyWithRegressionGate` return value and the lineage `manifest.jsonl` audit append), but they are **not surfaced into the observer's `usage-log.jsonl`** — the substrate that the learner (`learner.go`) and future Phase5 consumers read.

This SPEC adds a structured **Apply OUTCOME record** to the observer so each completed harness Apply emits one observable, persisted outcome composed of:

1. the regression-gate **verdict** (`kept` | `rolled-back`) — reused from the `applyWithRegressionGate` flow,
2. the project-health **delta** (baseline triple + candidate triple + regressed dimensions) — reused from `MetricTriple` / `Regressions`,
3. a **lineage correlation key** (`proposal_id` + the transition decision) — reused from the existing lineage record (NB: no `lineage_id` field exists today; §B.3 documents the correlation-key decision).

This outcome record is the **shared precondition** for two downstream Phase5 consumers — D1 failure-signature clustering (paper mechanism C5) and canary-effectiveness analysis — but this SPEC delivers ONLY the capture + persist mechanism. No consumer is built.

### A.2 Honest Framing (MUST-PASS quality bar)

[HARD] This SPEC MUST honestly state the enabler's **present-day value** and MUST NOT oversell.

**What this SPEC genuinely delivers:** harness Apply outcomes become **observable and persisted** in `usage-log.jsonl`. Before this SPEC, an Apply's verdict + delta were transient (consumed by the gate, audited in `manifest.jsonl`) and absent from the observer substrate that Phase5 analysis reads.

**What this SPEC does NOT deliver (and MUST NOT claim):**

- It does NOT cluster failures, score canary effectiveness, validate held-out splits, or wire a scorer loop. Those are downstream consumers, each a separate future SPEC (see Exclusions).
- It does NOT "improve" the harness, "prevent" regressions, or change any Apply decision. The capture is a passive observer write — it records the outcome the gate already decided; it never alters keep/rollback.
- Because the current FROZEN allowlist restricts harness writes to markdown skill/agent descriptions (which cannot move Go test/coverage/lint counts), the captured delta is, in the typical current case, **Δ=0** (an Apply that was kept with no project-health change). The record is still genuine and still worth persisting — its value is making the outcome **observable for downstream analysis**, not demonstrating an active effect today.

[HARD] Any framing in this SPEC, plan.md, or acceptance.md that implies the captured outcome "improves" the harness or that the enabler does anything beyond capture+persist is a defect. plan-auditor will correctly flag overclaim that is not honestly disclosed.

### A.3 Tier Classification

**Tier M (standard)** — explicit justification:

- New outcome record type + a new `EventType` constant + additive `omitempty` fields on the existing `Event` struct + a focused `RecordOutcome` observer method + one Apply-seam composition point (where verdict + delta + correlation key are assembled) + tests.
- Estimated ~250-400 LOC across ~4 files: `internal/harness/outcome.go` (new — outcome record type + `RecordOutcome` method + `EventType` constant), `internal/harness/outcome_test.go` (new), `internal/harness/types.go` (additive `omitempty` Event fields + SchemaVersion bump), `internal/harness/applier.go` (one composition seam where the gate's verdict + delta + proposal_id are passed to the outcome recorder).
- The Apply-seam touch (composing the outcome inside `applyWithRegressionGate` without altering its keep/rollback contract) is the reason this is Tier M rather than Tier S — it is the one non-mechanical decision (where + how to emit, without breaking the gate).

**Why NOT Tier S**: the Apply-seam composition (DD-2) is a genuine wiring decision in a tested, recently-landed pipeline, and the representation choice (DD-1) has multiple alternatives with backward-compatibility implications. A 2-file (spec + plan) Tier S artifact set would under-document these.

**Why NOT Tier L**: no new package, no FROZEN constant change, no safety layer added, no public-API surface beyond one new exported `EventType` constant + outcome fields. The change is bounded to the observer + the Apply composition seam.

### A.4 cycle_type (run-phase)

`tdd` — new feature, test-first.

---

## B. Background

### B.1 Self-Harness Roadmap Context

The Self-Harness paper (arXiv 2606.09498v1) describes **bounded, evidence-based, NON-recursive self-improvement** — recursive self-evolution is explicitly excluded; the improvement machine (scorer, tier thresholds, FROZEN allowlist) is itself frozen. MoAI-ADK's harness is identically non-recursive.

Roadmap position:

- **P0** ✅ (SPEC-HARNESS-LOOP-CLOSURE-001) — M6 lineage + M7 human-gate.
- **P1** ✅ (SPEC-HARNESS-REGRESSION-GATE-001) — M2-lite non-regression gate (verdict + delta now exist transiently).
- **P2 / Phase5** — failure-signature clustering (C5), canary-effectiveness, held-out validation (C7), scorer-loop wiring (D4).
- **THIS SPEC** — the FIRST sub-item of P2/Phase5: the observer OUTCOME-capture enabler. It does NOT implement any Phase5 consumer; it makes the Apply outcome observable so the consumers have a substrate to read.
- **DROPPED (permanent)** — LLM proposer / K-candidate diversity (C6): FROZEN hallucination risk + K=1 unproven.

### B.2 Why an observer-side capture (not a lineage-only record)

The Apply transition is already audited in the lineage `manifest.jsonl` (`Decision: "approved" | "rejected" | "regression-blocked"`). That audit is sufficient for **human review of a single Apply**. It is NOT the substrate the Phase5 analysis consumers read — those consumers read the observer's `usage-log.jsonl` (e.g., `learner.go` aggregates Events from it). Adding the outcome to the observer log places the verdict + delta in the same time-ordered stream as the WHAT-events, so a future clustering/canary consumer can correlate "this Apply followed this usage pattern and produced this outcome" from one log. The lineage manifest remains the human-audit record; the observer outcome record is the machine-analysis substrate. The two are complementary — this SPEC does not replace or modify the lineage write.

### B.3 The `lineage_id` correlation-key reality (ground-truth)

The roadmap framing referred to "lineage_id" as a reuse target. Ground-truth (verified by reading `types.go` / `lineage.go`): **no `lineage_id` field exists**. Lineage entries (`LineageEntry`) are keyed by `ProposalID` + `Timestamp` (write order). The canonical correlation key already shared by both the lineage record and the Apply flow is **`ProposalID`** (`Proposal.ID`). The outcome record therefore reuses `proposal_id` (plus the transition `decision` string) as its lineage correlation key — NOT a non-existent `lineage_id`. This SPEC does NOT introduce a `lineage_id` field to `LineageEntry` (that would be schema churn in a completed SPEC's type); it reuses the existing `proposal_id` key. See plan.md DD-3.

---

## C. Requirements (GEARS)

### C.1 Outcome Record Type

**REQ-OC-001** (Ubiquitous): The harness package shall expose an outcome record type capturing, for one completed Apply: the verdict (`kept` | `rolled-back`), the baseline `MetricTriple`, the candidate `MetricTriple`, the list of regressed dimensions, the `proposal_id` correlation key, and the transition `decision` string (`"approved"` | `"regression-blocked"`).

**REQ-OC-002** (Ubiquitous): The outcome verdict shall be derived mechanically from the Apply result — `kept` when the change was kept (non-regressing, first-run adopt, or gate-inactive markdown-only apply), `rolled-back` when the regression gate restored the snapshot and returned `ApplyRegressionError`.

### C.2 Observer Recording Path

**REQ-OC-003** (Ubiquitous): The observer shall expose a `RecordOutcome` method that serializes the outcome record into a single `usage-log.jsonl` line via the existing `RecordExtendedEvent` low-level append path (reusing the established `O_APPEND|O_CREATE|O_WRONLY`, auto-mkdir, lazy-pruning machinery — no new low-level writer).

**REQ-OC-004** (Ubiquitous): The outcome record shall be encoded as an `Event` of a new event type `EventTypeApplyOutcome` (`"apply_outcome"`), carrying the verdict / triples / regressed-dimensions / proposal_id / decision in additive `omitempty` fields on the existing `Event` struct — following the established per-event-type `omitempty` field pattern (Stop / SubagentStop / UserPromptSubmit).

**REQ-OC-005** (Event-driven): When a harness Apply completes through the active regression gate (`applyWithRegressionGate`), the Apply path shall compose one outcome record from the gate's verdict + baseline triple + candidate triple + regressed dimensions + `proposal_id` + decision, and shall record it via `RecordOutcome` — for both the kept branch and the rolled-back branch.

### C.3 Non-Interference with the Apply Contract

**REQ-OC-006** (Ubiquitous): The outcome capture shall NOT alter the regression gate's keep/rollback decision, the returned error (`nil` | `*ApplyRegressionError`), the baseline-store update, or the M6 lineage write. The capture is an additive observer write composed from values the gate already computed.

**REQ-OC-007** (Event-driven): When the `RecordOutcome` observer write fails (file I/O error), the Apply path shall NOT change its keep/rollback outcome — the recording error shall be surfaced/wrapped consistently with the existing additive-write error semantics (the lineage-write error pattern), never converting a kept Apply into a failed one nor a rolled-back Apply into a kept one.

**REQ-OC-008** (Optional / capability gate): Where the regression gate is inactive (the `NewApplier()` default with no measurer/baseline store — the straight-line markdown modify path), the Apply path may still record an outcome with verdict `kept` and a zero/empty delta, OR omit the outcome record; the chosen behavior shall be documented in plan.md DD-4 and bound by an acceptance criterion. (The capture mechanism MUST be correct in the gate-active path regardless; the gate-inactive path behavior is a documented design choice, not left ambiguous.)

### C.4 Backward Compatibility

**REQ-OC-009** (Ubiquitous): Existing `usage-log.jsonl` readers (`learner.go` `AggregatePatterns`, `retention.go`, `cli/harness.go`) shall tolerate the new `EventTypeApplyOutcome` events and the new `omitempty` fields without error — additive `omitempty` fields are ignored by readers that do not expect them, and an unknown event type is skipped or counted but never fatal.

**REQ-OC-010** (Ubiquitous): The log schema version constant `LogSchemaVersion` shall be bumped (`"v1"` → `"v2"`) to mark the additive field set, preserving the established additive-omitempty backward-compatibility convention (per the `Event` struct's REQ-HRN-OBS-009 note: new fields are additive omitempty only).

### C.5 Subagent Boundary & Human Gate Preservation

**REQ-OC-011** (Unwanted behavior): The outcome capture shall not invoke `AskUserQuestion` or `mcp__askuser`; it is an observer write inside `internal/harness/` and the C-HRA-008 binary guard shall stay GREEN.

**REQ-OC-012** (Ubiquitous): The outcome capture shall preserve the L5 human gate unchanged — it runs only after the safety pipeline + regression gate have decided, records the already-decided outcome, and shall NOT enable autonomous apply nor alter `auto_apply: false`.

### C.6 Honest Framing

**REQ-OC-013** (Ubiquitous): The SPEC artifacts (spec.md §A.2, plan.md Design section, acceptance.md) shall explicitly document that this SPEC delivers capture+persist ONLY — that the captured delta is typically `Δ=0` for the current markdown-only write surface, that the CONSUMING analysis (clustering, canary-effectiveness) is downstream and out of scope, and that no "improves" / "prevents" claim is made.

---

## D. HARD Constraints

| ID | Constraint | Bound REQ |
|----|------------|-----------|
| C1 | `internal/harness/frozen_guard.go` `allowedPrefixes` (line 18-22) + `frozenPrefixes` (line 27-32) UNCHANGED. | REQ-OC-012 |
| C2 | `internal/harness/safety/frozen_guard.go` `frozenPrefixes` UNCHANGED. | REQ-OC-012 |
| C3 | `internal/harness/tier/tier.go` `tierThresholds = [4]int{1,3,5,10}` UNCHANGED. | REQ-OC-012 |
| C4 | `internal/harness/scorer.go` 4 Dimensions {Functionality, Security, Craft, Consistency} + `DefaultMustPassDimensions` UNCHANGED. | REQ-OC-012 |
| C5 | `.moai/config/sections/harness.yaml` `auto_apply: false` UNCHANGED. The L5 human gate is preserved; this SPEC does NOT enable autonomous apply. | REQ-OC-012 |
| C6 | The 5-layer safety pipeline (L1 frozen → L2 canary → L3 contradiction → L4 rate-limit → L5 human gate) architecture UNCHANGED. The regression gate (P1) and this outcome capture are in-Apply steps after `DecisionApproved`, NOT additional safety layers. | REQ-OC-006, REQ-OC-012 |
| C7 | C-HRA-008 subagent boundary: no `AskUserQuestion(` / `mcp__askuser(` call-site in `internal/harness/` or `internal/hook/` Go source. `TestSubagentBoundary_NoAskUserQuestion` MUST stay GREEN. | REQ-OC-011 |
| C8 | The new event type constant MUST be named distinctly and additive (`EventTypeApplyOutcome = "apply_outcome"`); it MUST NOT redefine, reorder, or remove any existing `EventType` constant. | REQ-OC-004 |
| C9 | New `Event` fields MUST be additive `omitempty` only — no existing field renamed, reordered (semantically), or removed; the existing 4-field core schema (timestamp/event_type/subject/context_hash) + the established omitempty fields stay byte-compatible for existing readers. | REQ-OC-009 |
| C10 | The regression gate decision (keep/rollback), returned error, baseline-store update, and M6 lineage write MUST be unchanged by the outcome capture — the capture is composed from values the gate already produced and emitted as an additional observer write. | REQ-OC-006 |
| C11 | The outcome capture MUST NOT introduce a `lineage_id` field to `LineageEntry`, nor modify `lineage.go` / the lineage manifest schema — it reuses the existing `proposal_id` + `decision` correlation key. | REQ-OC-001, REQ-OC-006 |
| C12 | Honest framing (§A.2) — no overstatement of the enabler's value; capture+persist ONLY, no "improves"/"prevents" claim. | REQ-OC-013 |

### Preservation tests required GREEN (run in acceptance)

- `TestSafetyArchitecture_LayerCount`
- `TestSafetyArchitecture_FrozenZoneUnchanged`
- `TestIsFrozen_*` (table, `internal/harness/safety`)
- `TestSentinelCatalog_*`
- `TestSubagentBoundary_NoAskUserQuestion`
- P1 regression-gate tests (`TestApply_Regression_*`) — MUST stay GREEN (the outcome capture must not break the gate's keep/rollback paths)
- tier tests (`internal/harness/tier`)

---

## Exclusions (What NOT to Build)

This section uses h3 (`###`) sub-headings with dash-bullet items per `moai spec lint --strict` `MissingExclusions` / `OutOfScopeRule`.

### Out of Scope (Phase5 deferrals — each a SEPARATE future SPEC)

- D1 failure-signature clustering (paper mechanism C5) — the CONSUMER that reads the outcome records and clusters failure signatures. This SPEC delivers only the capture substrate; the clustering itself is a future SPEC.
- Canary-effectiveness analysis — measuring whether an applied harness change improved agent outcomes by reading the captured outcome stream. Future SPEC.
- D3 held-out split validation (paper mechanism C7) — train/validate evaluation set. Future SPEC.
- D4 scorer-loop wiring — feeding the captured outcomes back into the scorer. Future SPEC.
- A `lineage_id` field on `LineageEntry` — not introduced. Correlation reuses the existing `proposal_id`. If a richer lineage key is ever required, it is a separate lineage-schema SPEC.
- Outcome-record analytics / query CLI (`moai harness outcomes`) — no read/query surface is built; only the write path. Future SPEC if a CLI consumer is wanted.

### Explicitly Dropped (not merely deferred)

- LLM proposer / K-candidate diversity (paper mechanism C6) — DROPPED from the roadmap, not deferred. FROZEN hallucination risk + K=1 unproven. Recorded as a permanent drop.

### Permanently Rejected

- Enabling autonomous apply / changing `auto_apply` from `false` — PERMANENTLY REJECTED. The L5 constitutional human gate is non-negotiable.
- Using the captured outcome to alter any Apply decision — PERMANENTLY REJECTED for this enabler. The capture is passive; the outcome is recorded AFTER the gate decided and never feeds back into the same Apply.
- Recursive threshold/weight self-modification (FROZEN constants + 16-language template neutrality) — PERMANENTLY REJECTED.

---

## E. Cross-References

| File | Line anchor | Role | Marker |
|------|-------------|------|--------|
| `internal/harness/observer.go` | 103 (`RecordExtendedEvent`) | Low-level append path the new `RecordOutcome` delegates to | REFERENCE (reuse, do not modify) |
| `internal/harness/observer.go` | 53 (`RecordEvent`) | Existing 3-arg recorder — pattern to mirror for `RecordOutcome` | REFERENCE |
| `internal/harness/types.go` | 17-43 (`EventType` consts) | Add `EventTypeApplyOutcome` additively after the existing constants | EDIT (additive const only) |
| `internal/harness/types.go` | 53-120 (`Event` struct) | Add outcome `omitempty` fields after the existing per-event-type omitempty fields | EDIT (additive omitempty fields) |
| `internal/harness/types.go` | 11 (`LogSchemaVersion = "v1"`) | Bump to `"v2"` (additive field set marker) | EDIT (version bump) |
| `internal/harness/applier.go` | 346-407 (`applyWithRegressionGate`) | Composition seam — emit outcome record in BOTH the kept branch (line ~399-406) and the rolled-back branch (line ~384-396) without altering keep/rollback | EDIT (additive outcome emit) |
| `internal/harness/regression_gate.go` | 35-39 (`MetricTriple`), 50-62 (`Regressions`), 71-92 (`ApplyRegressionError`) | Verdict + delta source values reused by the outcome record | REFERENCE (reuse, no change) |
| `internal/harness/lineage.go` | 28 (`WriteLineageEntry`) + types.go 399 (`LineageEntry`) | The lineage audit record — UNCHANGED; outcome record is a separate observer write | **DO-NOT-MODIFY** |
| `internal/harness/frozen_guard.go` | 18-22, 27-32 | allowedPrefixes / frozenPrefixes | **DO-NOT-MODIFY** |
| `internal/harness/safety/frozen_guard.go` | frozenPrefixes | frozenPrefixes | **DO-NOT-MODIFY** |
| `internal/harness/tier/tier.go` | tierThresholds [1,3,5,10] | **DO-NOT-MODIFY** | **DO-NOT-MODIFY** |
| `internal/harness/scorer.go` | 4 dimensions + DefaultMustPassDimensions | **DO-NOT-MODIFY** | **DO-NOT-MODIFY** |
| `.moai/config/sections/harness.yaml` | `auto_apply: false` | **DO-NOT-MODIFY** | **DO-NOT-MODIFY** |
| `internal/harness/subagent_boundary_test.go` | `TestSubagentBoundary_NoAskUserQuestion` | C-HRA-008 binary guard (must stay GREEN) | **DO-NOT-MODIFY** |

Predecessor SPECs: `.moai/specs/SPEC-HARNESS-LOOP-CLOSURE-001/` (P0, completed — M6 lineage + M7 human-gate), `.moai/specs/SPEC-HARNESS-REGRESSION-GATE-001/` (P1, completed — M2-lite non-regression gate; this SPEC reuses its `MetricTriple` / `Regressions` / `ApplyRegressionError` / `applyWithRegressionGate`).
