---
id: SPEC-HARNESS-REGRESSION-GATE-001
title: "Harness M2-lite 비회귀 게이트 (measurement scaffold + defense-in-depth)"
version: "0.1.1"
status: implemented
created: 2026-06-14
updated: 2026-06-14
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/harness, internal/measure"
lifecycle: spec-anchored
tags: "harness, regression-gate, measurement, self-harness, defense-in-depth"
tier: M
era: V3R6
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-14 | manager-spec | Initial draft — Self-Harness 로드맵 P1 (M2-lite 비회귀 게이트). Predecessor SPEC-HARNESS-LOOP-CLOSURE-001 (P0, completed) 의 M6 lineage + M7 human-gate 위에 측정-인프라 scaffold + 방어층 추가. |
| 0.1.1 | 2026-06-14 | orchestrator | plan-auditor iter-1 PASS-WITH-DEBT 0.86 (Tier M 임계 0.80) 후속 orchestrator-direct 결함 remediation: D2 GEARS 라벨(REQ-RG-008 Event-driven), D3 transitive helper 추출 명시(REQ-RG-002), D5 측정-실패 fail-closed 신규 REQ-RG-014 + §C.6, D6 line-cite drift(C1/C2). acceptance.md: D1 AC-RG-008 guard 분리, D4 신규 AC-RG-011/012, D5 신규 AC-RG-013. |

---

## A. Overview

### A.1 Purpose

This SPEC ports the **M2-lite non-regression gate** from the Self-Harness paper (arXiv 2606.09498v1) roadmap item **P1**. The predecessor SPEC-HARNESS-LOOP-CLOSURE-001 (P0, completed) ported M6 (auditable lineage) and M7 (human-gated loop), achieving one clean closure of the harness Apply loop.

This SPEC adds a **non-regression gate** to the Apply pipeline: before keeping an approved harness change, the gate measures three project-health dimensions (**test pass count, coverage%, lint count**) on a baseline and on the post-apply candidate, and requires **every dimension delta ≥ 0 (non-regressing)**. If any dimension regresses, the gate rolls back the snapshot and returns a new error to the orchestrator. The L5 human gate is preserved unchanged.

### A.2 Honest Framing (MUST-PASS quality bar)

[HARD] This SPEC MUST honestly and explicitly disclose the gate's **actual protective value for the current write surface**, and MUST NOT overstate it.

**The measured signal is genuine** — test pass count + coverage% + lint count, with `Δ ≥ 0` required, exactly per the roadmap intent.

**BUT** — because the FROZEN allowlist (`internal/harness/frozen_guard.go:18-22`) restricts every harness proposal to:

- `.claude/agents/my-harness/`
- `.claude/skills/my-harness-`
- `.moai/harness/`

…harness proposals modify **markdown skill/agent descriptions and triggers ONLY — never Go code, never tested templates**. Modifying a markdown frontmatter `description:` or `triggers:` field cannot change Go test pass count, coverage%, or `go vet` lint count. **Therefore, for the current markdown-only write surface, the measured delta is typically `Δ=0` (the gate is always-pass).**

The gate's genuine value is stated honestly as **exactly two things**, NOT as "prevents harness regressions":

1. **Measurement-infrastructure scaffold** — establishes the baseline store + delta comparator + measurement collector that Phase5's richer signals (canary effectiveness, held-out validation) will reuse. This is the gate's primary present-day value: an infrastructure seam, not an active guard.

2. **Defense-in-depth safety net** — catches a regression *if and only if* one of these latent conditions ever occurs:
   - the FROZEN allowlist is widened in a future SPEC to permit writes into tested code/templates, OR
   - an Apply operation has a bug that writes outside the allowlist (e.g., a path-traversal or symlink escape inadvertently touching a tested Go/template file).

   Under the current correct, narrow allowlist, this safety net never fires. It is insurance against a future allowlist widening or an applier defect — not a guard against the markdown writes that actually occur today.

[HARD] Any framing in this SPEC, plan.md, or acceptance.md that implies the gate actively prevents regressions in the *current* harness operation (rather than being a scaffold + dormant safety net) is a defect. plan-auditor will correctly flag vacuous/tautological-gate framing that is not honestly disclosed.

### A.3 Tier Classification

**Tier M (standard)** — explicit justification:

- New measurement collector + baseline store + regression gate + new error type + applier seam wiring + leaf-package extraction + tests.
- Estimated ~350-500 LOC across ~6 files: `internal/measure/measure.go` (new leaf package, ~60 LOC extracted parsers), `internal/measure/measure_test.go`, `internal/harness/regression_gate.go` (new), `internal/harness/regression_gate_test.go`, `internal/harness/applier.go` (seam edit + new error type), `internal/loop/go_feedback.go` (delegate to leaf package — refactor only).
- Touches `internal/loop` (refactors 3 unexported parsers to delegate to the new leaf package). This cross-package touch is the reason Tier M is at its upper boundary rather than Tier S.

**Why NOT Tier L**: no new safety layer is added (the gate is an in-Apply step after L5-approval, NOT a 6th layer — see Design Decision DD-6 in plan.md). No FROZEN constant changes. No public API surface expansion beyond one new error type. The leaf-package extraction is a mechanical move-and-delegate with no behavior change. The change is bounded to the harness Apply pipeline + one shared parser package.

### A.4 cycle_type (run-phase)

`tdd` — new feature, test-first.

---

## B. Background

### B.1 Self-Harness Roadmap Context

The Self-Harness paper (arXiv 2606.09498v1) describes **bounded, evidence-based self-improvement** — explicitly NOT recursive self-evolution. The improvement machine itself is frozen. MoAI-ADK's harness is identically non-recursive: a human-gated self-improvement loop where the improvement mechanism (scorer, tier thresholds, FROZEN allowlist) never changes itself.

The roadmap (from the predecessor's adversarial-review analysis):

- **P0** ✅ (SPEC-HARNESS-LOOP-CLOSURE-001) — M6 auditable lineage + M7 human-gated loop.
- **P1** (this SPEC) — M2-lite non-regression gate.
- **P2 (Phase5)** — failure-signature clustering (C5), held-out split validation (C7), scorer-loop wiring (D4).
- **DROPPED** — LLM proposer / K-candidate diversity (C6): FROZEN hallucination risk + K=1 unproven.
- **PERMANENTLY REJECTED** — autonomous apply / `auto_apply: true`: L5 constitutional human gate.

### B.2 Why a project-health signal (not canary effectiveness)

The user chose the **project-health signal** (test/coverage/lint delta) over the semantically-richer canary-effectiveness signal. Canary effectiveness — measuring whether an applied harness change actually improved agent outcomes — is the Phase5 alternative (it requires observer outcome capture first, which does not yet exist). The project-health signal is the M2-lite ("lite") choice: mechanically simple, reuses existing parsers, and establishes the measurement seam Phase5 will extend.

---

## C. Requirements (GEARS)

### C.1 Measurement Collection

**REQ-RG-001** (Ubiquitous): The measurement collector shall compute a metric triple `{tests_passed int, coverage float64, lint_count int}` for the project, reusing the pure parsers extracted into the `internal/measure` leaf package.

**REQ-RG-002** (Ubiquitous): The `internal/measure` package shall expose the three pure parsers (`ParseGoTestJSON`, `ParseCoverageFile`, `CountNonEmptyLines`) as exported functions with zero dependency on `internal/lsp`, `internal/lsp/gopls`, or `internal/harness`, so both `internal/loop` and `internal/harness` may import it without an import cycle. The unexported transitive helpers the parsers depend on (`parseIntFromString`, `mustParseInt`, `mustParseIntErr` in `go_feedback.go:243+`) shall move into `internal/measure` alongside the parsers (remaining unexported) so the leaf package compiles standalone.

**REQ-RG-003** (Ubiquitous): The `internal/loop` `GoFeedbackGenerator` shall delegate its test/coverage/lint parsing to `internal/measure` exported functions, producing byte-identical parsing behavior to the pre-extraction implementation (refactor-only, no behavior change).

### C.2 Baseline Store

**REQ-RG-004** (Ubiquitous): The regression gate shall persist a baseline metric triple to `.moai/harness/measurements-baseline.yaml` using an atomic write (temp-file + rename).

**REQ-RG-005** (Event-driven): When the baseline file is absent (first run), the regression gate shall treat the candidate measurement as the new baseline and shall NOT block the Apply (no prior baseline to regress against).

**REQ-RG-006** (Ubiquitous): The regression gate shall NOT read or write any of the following existing harness files: `usage-log.jsonl`, the lineage `manifest.jsonl`, `observations.yaml`, `tier-promotions.jsonl`. (The baseline store is a new, separate file.)

### C.3 Regression Gate (in-Apply step)

**REQ-RG-007** (State-driven): While an Apply has reached `DecisionApproved`, the regression gate shall execute in this order: (1) measure baseline, (2) `createSnapshot`, (3) apply the file modification, (4) measure candidate, (5) compare deltas, (6) keep + record on `Δ ≥ 0` for every dimension, or rollback + return error on any regression.

**REQ-RG-008** (Event-driven): When any dimension regresses (`tests_passed` decreased OR `coverage` decreased OR `lint_count` increased), the regression gate shall call `RestoreSnapshot` to roll back the applied change and shall return a new `ApplyRegressionError` carrying the baseline triple, the candidate triple, and the list of regressed dimensions.

**REQ-RG-009** (Ubiquitous): On a non-regressing Apply (`Δ ≥ 0` for all dimensions), the regression gate shall keep the change, update the baseline store to the candidate triple, and proceed to the existing M6 approved-transition lineage write.

**REQ-RG-010** (Event-driven): When the regression gate blocks an Apply, it shall append a lineage entry with decision value `"regression-blocked"` to the existing lineage `manifest.jsonl` (via `WriteLineageEntry`) for auditability, carrying the regressed-dimension summary in the `Reason` field.

### C.4 Subagent Boundary & Human Gate Preservation

**REQ-RG-011** (Unwanted behavior): The regression gate shall not invoke `AskUserQuestion` or `mcp__askuser`; it shall return `ApplyRegressionError` to the orchestrator, which owns user interaction.

**REQ-RG-012** (Ubiquitous): The regression gate shall preserve the L5 human gate unchanged — it runs *after* `DecisionApproved` (which already passed the L5 human gate via the safety pipeline) and shall NOT enable autonomous apply nor alter `auto_apply: false`.

### C.5 Honest Framing

**REQ-RG-013** (Ubiquitous): The SPEC artifacts (spec.md §A.2, plan.md Design section, acceptance.md) shall explicitly document that the measured delta is typically `Δ=0` for the current markdown-only write surface, and shall frame the gate's value as (1) measurement-infrastructure scaffold and (2) defense-in-depth safety net — NOT as an active regression preventer for current harness operation.

### C.6 Measurement Robustness

**REQ-RG-014** (Event-driven): When the baseline or candidate measurement cannot be executed (build error, `go test` exec failure, or timeout) — as distinct from tests merely failing while the suite still runs, which yields a valid `tests_passed` count and is compared normally — the regression gate shall **fail closed**: it shall NOT keep the change on an unverifiable basis, and shall return a wrapped measurement error to the orchestrator. Rationale: the §A.2 defense-in-depth value is only enforceable if an unverifiable Apply is never kept (a defect that both writes outside the allowlist AND breaks the build must not slip through). Richer measurement-error resilience (retries, partial-failure tolerance, timeout tuning) is deferred to Phase5.

---

## D. HARD Constraints

| ID | Constraint | Bound REQ |
|----|------------|-----------|
| C1 | `internal/harness/frozen_guard.go` `allowedPrefixes` (line 18-22) + `frozenPrefixes` (line 27-31) UNCHANGED. | REQ-RG-012 |
| C2 | `internal/harness/safety/frozen_guard.go` `frozenPrefixes` (line 18-23) UNCHANGED. | REQ-RG-012 |
| C3 | `internal/harness/tier/tier.go` `tierThresholds = [4]int{1,3,5,10}` (line 50) UNCHANGED. | REQ-RG-012 |
| C4 | `internal/harness/scorer.go` 4 Dimensions {Functionality, Security, Craft, Consistency} (line 13-24) + `DefaultMustPassDimensions` (line 50) UNCHANGED. | REQ-RG-012 |
| C5 | `.moai/config/sections/harness.yaml:116` `auto_apply: false` UNCHANGED. The L5 human gate is preserved; this SPEC does NOT enable autonomous apply. | REQ-RG-012 |
| C6 | The 5-layer safety pipeline (L1 frozen → L2 canary → L3 contradiction → L4 rate-limit → L5 human gate) architecture UNCHANGED. The regression gate is an in-Apply step after `DecisionApproved`, NOT a 6th safety layer. | REQ-RG-007, REQ-RG-012 |
| C7 | C-HRA-008 subagent boundary: no `AskUserQuestion(` / `mcp__askuser(` call-site in `internal/harness/` or `internal/hook/` Go source. `TestSubagentBoundary_NoAskUserQuestion` MUST stay GREEN. | REQ-RG-011 |
| C8 | New error type MUST be named distinctly from the existing `ApplyPendingError` (applier.go:147). Canonical name: `ApplyRegressionError`. Mirror the existing error pattern (struct + `Error() string`). | REQ-RG-008 |
| C9 | The 3 pure parsers in `internal/measure` MUST have zero import of `internal/lsp`, `internal/lsp/gopls`, `internal/harness`, or `internal/loop` — verified by `go list -deps`. | REQ-RG-002 |
| C10 | Honest framing (§A.2) — no overstatement of the gate's protective value for the current write surface. | REQ-RG-013 |
| C11 | The baseline store is a NEW file (`.moai/harness/measurements-baseline.yaml`); the gate MUST NOT touch `usage-log.jsonl`, lineage `manifest.jsonl` (except the audit append of REQ-RG-010), `observations.yaml`, `tier-promotions.jsonl`. | REQ-RG-006 |

### Preservation tests required GREEN (run in acceptance)

- `TestSafetyArchitecture_LayerCount`
- `TestSafetyArchitecture_FrozenZoneUnchanged`
- `TestIsFrozen_*` (table)
- `TestSentinelCatalog_*`
- `TestSubagentBoundary_NoAskUserQuestion`
- tier tests (`internal/harness/tier`)

---

## Exclusions (What NOT to Build)

This section uses h3 (`###`) sub-headings with dash-bullet items per `moai spec lint --strict` `MissingExclusions` / `OutOfScopeRule`.

### Out of Scope (Phase5 deferrals)

- Composite weighting — the scorer 4-dimension verdict wired into the regression decision is deferred to Phase5. This SPEC's gate uses only the project-health triple, never the scorer verdict.
- Held-out split validation — train/validate evaluation set is deferred to Phase5.
- Failure-signature clustering (C5 paper mechanism) — deferred to Phase5; requires observer outcome capture to land first.
- Canary-effectiveness signal as the gate's measure — out of scope. The user chose the project-health signal. Canary effectiveness is noted as the semantically-richer Phase5 alternative (requires observer outcome capture).
- Measurement-error resilience — retries, partial-failure tolerance, and timeout tuning for the measurement step are deferred to Phase5. M2-lite uses a fail-closed default (REQ-RG-014): an unexecutable measurement returns an error rather than silently keeping the change.

### Explicitly Dropped (not merely deferred)

- LLM proposer / K-candidate diversity (C6 paper mechanism) — DROPPED from the roadmap, not deferred. FROZEN hallucination risk + K=1 unproven. Recorded as a permanent drop.

### Permanently Rejected

- Enabling autonomous apply / changing `auto_apply` from `false` — PERMANENTLY REJECTED. The L5 constitutional human gate is non-negotiable.
- Recursive threshold/weight self-modification (FROZEN constants + 16-language template neutrality) — PERMANENTLY REJECTED.

---

## E. Cross-References

| File | Line anchor | Role | Marker |
|------|-------------|------|--------|
| `internal/harness/applier.go` | 190 (`Apply`), 216 (`DecisionApproved`), 222 (`createSnapshot`), 226-240 (file modify), 246 (lineage) | Apply seam — gate hooks after line 218, around snapshot/apply/measure | EDIT (seam wiring + new error type) |
| `internal/harness/applier.go` | 147-157 (`ApplyPendingError`) | Error-type pattern to mirror for the new `ApplyRegressionError` | REFERENCE (do not modify existing type) |
| `internal/harness/applier.go` | 333 (`RestoreSnapshot`), 279 (`createSnapshot`) | Snapshot/rollback primitives (already exist) | REFERENCE (reuse, do not modify) |
| `internal/harness/lineage.go` | 28 (`WriteLineageEntry`) | Audit append for `"regression-blocked"` decision | REFERENCE (reuse) |
| `internal/harness/types.go` | 399 (`LineageEntry`, `Decision` free string) | Carries `"regression-blocked"` decision value | REFERENCE (reuse, no schema change) |
| `internal/loop/go_feedback.go` | 166 (`parseGoTestJSON`), 188 (`parseCoverageFile`), 227 (`countNonEmptyLines`) | Parsers to extract → `internal/measure`; then delegate | EDIT (delegate to leaf package) |
| `internal/loop/state.go` | 112 (`Feedback` struct) | `TestsPassed`/`Coverage`/`LintErrors` fields | REFERENCE (no schema change) |
| `internal/harness/frozen_guard.go` | 18-22, 27-31 | allowedPrefixes / frozenPrefixes | **DO-NOT-MODIFY** |
| `internal/harness/safety/frozen_guard.go` | 18-23 | frozenPrefixes | **DO-NOT-MODIFY** |
| `internal/harness/tier/tier.go` | 50 | tierThresholds [1,3,5,10] | **DO-NOT-MODIFY** |
| `internal/harness/scorer.go` | 13-24, 50 | 4 dimensions + DefaultMustPassDimensions | **DO-NOT-MODIFY** |
| `.moai/config/sections/harness.yaml` | 116 | `auto_apply: false` | **DO-NOT-MODIFY** |
| `internal/harness/subagent_boundary_test.go` | — | C-HRA-008 binary guard (must stay GREEN) | **DO-NOT-MODIFY** |

Predecessor SPEC: `.moai/specs/SPEC-HARNESS-LOOP-CLOSURE-001/` (P0, completed — M6 lineage + M7 human-gate).
