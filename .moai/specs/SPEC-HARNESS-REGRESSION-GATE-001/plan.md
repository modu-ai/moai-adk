---
id: SPEC-HARNESS-REGRESSION-GATE-001
title: "Implementation Plan — Harness M2-lite 비회귀 게이트"
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
tags: "harness, regression-gate, plan"
---

## A. Context

Run-phase implementation plan for SPEC-HARNESS-REGRESSION-GATE-001. cycle_type: `tdd`. This SPEC adds an in-Apply non-regression gate (project-health triple, `Δ ≥ 0`) plus a shared measurement leaf package, on top of the predecessor's M6 lineage + M7 human-gate skeleton.

The Design section (§F) is folded into this plan.md per the Tier M classification — the non-trivial decisions (import-cycle resolution, ordering, new error type, baseline format, gate-as-step-vs-layer, honest framing) are documented here so plan-auditor can verify they were considered.

## B. Known Issues / Risks

- **R1 — Measurement cost in Apply path**: running `go test ./...` twice (baseline + candidate) inside an Apply is expensive. Mitigation: scope the measurement to a configurable package set (default: whole module is acceptable for the current low Apply frequency — Apply runs are human-gated and rare). Document the cost honestly; do NOT optimize prematurely.
- **R2 — Measurement non-determinism**: flaky tests or coverage jitter could cause a false regression. Mitigation: the gate treats `tests_passed` decrease, `coverage` decrease, and `lint_count` increase as regressions; a tolerance band is OUT OF SCOPE (Phase5). Because the current markdown-only surface yields `Δ=0`, flakiness risk is dormant — document it as a future concern for when the allowlist widens.
- **R3 — Import-cycle/visibility (predecessor `L_import_cycle_ac_contradiction`)**: the SEC-HARDEN-002 incident had an import cycle missed by BOTH plan-auditor and the run-gate. DD-1 (§F) resolves this explicitly via a leaf package; the run-phase MUST verify `go list -deps` shows no cycle.
- **R4 — Refactor regression in `internal/loop`**: extracting the 3 parsers must be byte-identical. Mitigation: REQ-RG-003 requires delegation with no behavior change; existing loop tests MUST stay GREEN.

## C. Pre-flight (verify before M1)

```bash
# Confirm the Apply seam + error type collision are as the SPEC states
grep -n 'func (a \*Applier) Apply' internal/harness/applier.go        # expect :190
grep -n 'type ApplyPendingError struct' internal/harness/applier.go   # expect :147 (name TAKEN)
grep -n 'func parseGoTestJSON\|func parseCoverageFile\|func countNonEmptyLines' internal/loop/go_feedback.go
# Confirm FROZEN invariants present (will be re-asserted GREEN in acceptance)
grep -n 'auto_apply: false' .moai/config/sections/harness.yaml
go test ./internal/harness/... ./internal/loop/... 2>&1 | tail -5   # green baseline
```

## D. Constraints (from spec.md §D)

C1-C11 (see spec.md §D). The DO-NOT-MODIFY files are: `frozen_guard.go` (both), `tier/tier.go`, `scorer.go`, `harness.yaml`, `subagent_boundary_test.go`.

---

## E. Self-Verification (orchestrator read-only batch — run at run-phase completion)

```bash
# 1. Full harness + loop + measure suites
go test ./internal/harness/... ./internal/loop/... ./internal/measure/... 2>&1 | tail -10
# 2. Import-cycle proof for the leaf package (DD-1)
go list -deps github.com/modu-ai/moai-adk/internal/measure | grep -E 'internal/(lsp|gopls|harness|loop)' || echo "CLEAN: no forbidden deps"
# 3. C-HRA-008 boundary
grep -rn 'AskUserQuestion(\|mcp__askuser(' internal/harness/ internal/hook/ | grep -v '_test.go' | grep -v '^[^:]*:[0-9]*:[ \t]*//' || echo "CLEAN"
# 4. New error type present, old one untouched
grep -n 'type ApplyRegressionError struct\|type ApplyPendingError struct' internal/harness/applier.go
# 5. FROZEN unchanged (git diff against DO-NOT-MODIFY files)
git diff --stat internal/harness/frozen_guard.go internal/harness/safety/frozen_guard.go internal/harness/tier/tier.go internal/harness/scorer.go .moai/config/sections/harness.yaml
# 6. Coverage delta
go test -coverprofile=/tmp/rg.out ./internal/harness/... ./internal/measure/... && go tool cover -func=/tmp/rg.out | tail -1
# 7. Lint
golangci-lint run --timeout=2m ./internal/harness/... ./internal/measure/... ./internal/loop/...
```

---

## F. Design (folded into plan.md — Tier M)

### DD-1 — Import-cycle / visibility resolution: **extract to leaf package `internal/measure`** (option b)

**Decision**: Extract the 3 pure parsers (`parseGoTestJSON`, `parseCoverageFile`, `countNonEmptyLines`) into a NEW leaf package `internal/measure` with exported names (`ParseGoTestJSON`, `ParseCoverageFile`, `CountNonEmptyLines`). Both `internal/loop` and `internal/harness` import `internal/measure`. The transitive unexported helpers `parseCoverageFile` depends on (`parseIntFromString`, `mustParseInt`, `mustParseIntErr` at `go_feedback.go:239-250`) MUST move into `internal/measure` together with the parsers (kept unexported), or the leaf package will not compile — the "~60 LOC extracted" estimate includes these helpers.

**Alternatives rejected**:
- **(a) import `internal/loop` + use `GoFeedbackGenerator.Collect()`** — REJECTED. `internal/loop/go_feedback.go` imports `internal/lsp` and `internal/lsp/gopls` (verified: line 14-15). Importing `internal/loop` into `internal/harness` would pull the entire LSP/gopls dependency tree into the harness Apply path — unjustified heavyweight coupling for 3 pure string parsers.
- **(c) reimplement ~50 LOC inside harness** — REJECTED. Duplicating the parsers risks drift between the loop's parsing and the gate's parsing (a future bug-fix in one is missed in the other). The leaf package is the single source of truth.

**Why this is safe (import-cycle proof)**:
- The 3 parsers are **pure**: they use only `bufio`, `bytes`, `encoding/json`, `os`, `strings` (verified by reading `go_feedback.go:166-240`). Zero LSP coupling.
- `internal/measure` will import NONE of `internal/lsp`, `internal/lsp/gopls`, `internal/harness`, `internal/loop` (C9).
- `internal/loop` already does NOT import `internal/harness`, and `internal/harness` does NOT import `internal/loop` (verified — no relationship either direction today). Adding `loop → measure` and `harness → measure` creates a DAG (both point down to the leaf), no cycle.
- Run-phase MUST prove this with `go list -deps internal/measure` (E.2 above). This directly addresses the predecessor `L_import_cycle_ac_contradiction` incident where a cycle was missed by both plan-auditor and the run-gate.

**Tier impact**: touching `internal/loop` (refactor-delegate) is the reason Tier M sits at its upper boundary. It does NOT cross into Tier L because the move is mechanical and behavior-preserving (REQ-RG-003).

### DD-2 — measure → apply → rollback ordering

The gate runs as a step **inside `Apply`, after `DecisionApproved` (applier.go:218), replacing the current straight-line snapshot→modify→lineage block** with:

```
DecisionApproved:
  baseline   = measure()                      # (1) BEFORE snapshot/apply
  createSnapshot(proposal, snapshotBase)      # (2) existing primitive — unchanged
  applyFileModification(proposal)             # (3) existing EnrichDescription/InjectTrigger
  candidate  = measure()                      # (4) AFTER apply
  deltas     = compare(baseline, candidate)   # (5)
  if any regression in deltas:
      RestoreSnapshot(snapshotDir)            # (6a) rollback — existing primitive
      writeLineage(proposal, "regression-blocked", "", regressionSummary)  # audit (REQ-RG-010)
      return &ApplyRegressionError{Baseline, Candidate, Regressed}
  else:
      updateBaselineStore(candidate)          # (6b) persist new baseline
      writeLineage(proposal, "approved", ...) # existing M6 approved-transition (unchanged)
      return nil
```

**Why measure-baseline BEFORE snapshot**: the baseline must reflect the pre-apply project state. Measuring after the modification would compare the candidate to itself. The snapshot is a file backup (for rollback), independent of the measurement.

**Why measure-candidate AFTER apply (not predicted)**: you cannot predict the post-apply test/coverage/lint result without applying. Hence the realistic flow is apply-then-measure-then-maybe-rollback, using the existing `createSnapshot`/`RestoreSnapshot` primitives for safe rollback.

**Measurement-error path (REQ-RG-014, fail-closed)**: if `measure()` cannot execute for either baseline or candidate (build error, `go test` exec failure, or timeout), the gate fails closed — `RestoreSnapshot` (if already snapshotted/applied) then return a wrapped measurement error; the change is NOT kept. This is distinct from a still-running suite that reports failing tests (which yields a valid `tests_passed` count and is compared normally). Rationale: the §A.2 / DD-7 defense-in-depth value is only real if an unverifiable Apply is never kept.

**Note on the `RestoreSnapshot` argument**: `createSnapshot` writes to `snapshotBase/<ISO-DATE>/`; `RestoreSnapshot(snapshotDir)` takes the dated directory. The gate must capture the exact `snapshotDir` returned/derivable from `createSnapshot` to roll back. Run-phase: `createSnapshot` currently does not return the dir — a minimal change to surface the created `snapshotDir` (return value or capture) is required. This is an internal seam edit within `applier.go`, not a public API change.

### DD-3 — New error type: `ApplyRegressionError`

`ApplyPendingError` (applier.go:147) is TAKEN for the L5 human-gate pending case. The regression gate returns a distinct type:

```go
// ApplyRegressionError is returned when the in-Apply non-regression gate detects
// a project-health regression and rolls back the applied change.
type ApplyRegressionError struct {
    Baseline  MetricTriple   // tests_passed, coverage, lint_count BEFORE apply
    Candidate MetricTriple   // ... AFTER apply
    Regressed []string       // e.g. ["coverage", "lint_count"]
}
func (e *ApplyRegressionError) Error() string { /* "apply: non-regression gate blocked (regressed: ...)" */ }
```

Mirrors the existing `ApplyPendingError` pattern (struct + `Error() string`). Carries baseline, candidate, and regressed-dimension list per REQ-RG-008.

### DD-4 — Baseline store format

New file `.moai/harness/measurements-baseline.yaml`:

```yaml
tests_passed: 1234
coverage: 87.7
lint_count: 0
updated_at: "2026-06-14T03:39:00Z"
```

- Atomic write (temp-file + `os.Rename`), 0o644.
- Absent file (first run) → REQ-RG-005: candidate becomes the baseline, no block.
- MUST NOT touch `usage-log.jsonl`, lineage `manifest.jsonl`, `observations.yaml`, `tier-promotions.jsonl` (C11).

### DD-5 — Audit trail via existing lineage

A regression-blocked Apply appends a `LineageEntry{Decision: "regression-blocked"}` via the existing `WriteLineageEntry` (lineage.go:28). `LineageEntry.Decision` is a free `string` (types.go) — no schema change needed; it currently carries `"approved"` | `"rejected"`, and now additionally `"regression-blocked"`. The `Reason` field carries the regressed-dimension summary (REQ-RG-010).

### DD-6 — Gate-as-step, NOT a 6th safety layer

**Decision**: the regression gate is an **in-Apply step after `DecisionApproved`**, NOT a 6th safety pipeline layer.

**Justification**: the 5-layer architecture (L1 frozen → L2 canary → L3 contradiction → L4 rate-limit → L5 human gate) is a FROZEN invariant asserted by `TestSafetyArchitecture_LayerCount` (C6). The safety pipeline evaluates a proposal *before* any file is touched and returns a `Decision`; it cannot measure post-apply project health because nothing has been applied yet. The regression gate inherently needs the post-apply state, so it MUST live in the Apply execution path, after the pipeline has approved. Modeling it as a 6th layer would (a) break the FROZEN 5-layer count test and (b) be architecturally wrong (the pipeline is a pre-apply evaluator). Therefore: in-Apply step.

### DD-7 — Honest framing rationale (REQ-RG-013, MUST-PASS)

The gate's measured delta is **typically `Δ=0`** for the current markdown-only write surface (FROZEN allowlist → harness only edits `.claude/skills/my-harness-*` / `.claude/agents/my-harness/` / `.moai/harness/` markdown frontmatter, which cannot change Go test/coverage/lint counts). The gate's two honest values:

1. **Measurement scaffold** — baseline store + delta comparator + collector that Phase5 reuses. This is the present-day value.
2. **Defense-in-depth** — fires only if the allowlist is widened OR an applier defect writes outside the allowlist into tested code. Dormant under the current correct narrow allowlist.

The SPEC MUST NOT frame the gate as actively preventing current-operation regressions. This is the MUST-PASS quality bar; plan-auditor will flag any vacuous/tautological framing not honestly disclosed.

---

## G. Milestones

| M | Goal | Files | Verify |
|---|------|-------|--------|
| M1 | Extract pure parsers + transitive helpers (`parseIntFromString`/`mustParseInt`/`mustParseIntErr`, kept unexported) to `internal/measure` leaf package (3 exported parsers, zero LSP dep). RED: write `measure_test.go` first. | `internal/measure/measure.go` (new), `internal/measure/measure_test.go` (new) | `go test ./internal/measure/...` ; `go list -deps github.com/modu-ai/moai-adk/internal/measure \| grep -E 'internal/(lsp\|gopls\|harness\|loop)' \|\| echo CLEAN` |
| M2 | Delegate `internal/loop` parsers to `internal/measure` (refactor, byte-identical). | `internal/loop/go_feedback.go` | `go test ./internal/loop/...` (existing tests stay GREEN, no behavior change) |
| M3 | Add `MetricTriple` + `ApplyRegressionError` + baseline store (atomic YAML read/write). RED first. | `internal/harness/regression_gate.go` (new), `internal/harness/regression_gate_test.go` (new) | `go test -run 'TestApplyRegressionError\|TestBaselineStore' -v ./internal/harness/` |
| M4 | Wire the in-Apply gate (DD-2 ordering) into `Apply` after `DecisionApproved`; surface `snapshotDir` from `createSnapshot` for rollback; audit append `"regression-blocked"`; fail-closed on measurement-exec error (REQ-RG-014). RED first. | `internal/harness/applier.go`, `internal/harness/regression_gate.go`, `internal/harness/applier_test.go` | `go test -run 'TestApply_Regression' -v ./internal/harness/` |
| M5 | FROZEN preservation re-assert + C-HRA-008 + full suite + coverage + lint. | (tests only) | `go test ./internal/harness/... ./internal/loop/... ./internal/measure/...` ; FROZEN git-diff stat clean ; `golangci-lint run --timeout=2m` |

Milestones are priority-ordered (no time estimates). M1→M2 establish the shared parser; M3→M4 build and wire the gate; M5 is the preservation/quality gate.

## H. Anti-Patterns to avoid (predecessor lessons)

- `L_ac_run_pattern_vacuous_guard` — every acceptance `go test -run '<pattern>'` MUST match a real test name (no infix/substring mismatch like `TestFoo` not matching `TestFoo_Bar`) and MUST guard against vacuous "[no tests to run]".
- `L_import_cycle_ac_contradiction` — DD-1 resolves import visibility explicitly; M1 verify proves no cycle.
- `L_orchestrator_direct_plan_patch` — orchestrator may apply mechanical plan-auditor remediations directly.
- Over-engineering: no tolerance band, no composite weighting, no canary signal (all Phase5/dropped). The gate is the minimal triple comparator.

## I. Cross-References

See spec.md §E for the full file:line + DO-NOT-MODIFY table. Predecessor: SPEC-HARNESS-LOOP-CLOSURE-001 (P0, completed).
