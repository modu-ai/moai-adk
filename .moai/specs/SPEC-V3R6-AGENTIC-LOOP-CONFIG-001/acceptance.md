---
id: SPEC-V3R6-AGENTIC-LOOP-CONFIG-001
title: "Go-side loader registration for workflow.agentic_loop.max_iterations (prose-read → 기계적 상한)"
version: "0.1.0"
status: in-progress
created: 2026-07-08
updated: 2026-07-08
author: manager-spec
priority: P2
phase: "v3.x"
module: "internal/config"
lifecycle: spec-anchored
era: V3R6
tier: S
tags: "config, workflow, agentic-loop, loader, distinctness-invariant, follow-up"
depends_on:
  - SPEC-MOAI-AGENTIC-LOOP-001
---

# SPEC-V3R6-AGENTIC-LOOP-CONFIG-001 — Acceptance Criteria

## §A. Traceability Matrix

Every REQ-ALC-* in `spec.md §B` maps to one or more acceptance criteria below. Every AC is observable by a mechanical command (test output, grep match, build result) — not subjective judgment.

| REQ | AC(s) | Observation |
|---|---|---|
| REQ-ALC-001 (typed Go field) | AC-001, AC-006 | `grep` finds the field; build succeeds |
| REQ-ALC-002 (loader populates field) | AC-002, AC-003 | `go test` assertion output |
| REQ-ALC-003 (single const default = 10) | AC-004, AC-005 | `grep` finds exactly one const declaration; no literal `10` elsewhere |
| REQ-ALC-004 (absent block → default, no error) | AC-003 | `go test` assertion |
| REQ-ALC-005 (two distinct Go fields) | AC-DISTINCT-001, AC-DISTINCT-002 | dedicated anti-aliasing test |
| REQ-ALC-006 (no merge into LoopPrevention) | AC-DISTINCT-001, AC-DISTINCT-002 | dedicated anti-aliasing test |
| REQ-ALC-007 (schema_sections.go registration) | AC-007 | `grep` in schema_sections.go |
| REQ-ALC-008 (distinct path tokens) | AC-007, AC-008 | `grep` for the 4-token path |
| REQ-ALC-009 (table-driven tests) | AC-009 | test file exists, cases enumerated |
| REQ-ALC-010 (dedicated anti-aliasing test) | AC-DISTINCT-002 | test file exists |
| REQ-ALC-011 (template parity acknowledged) | AC-010 | template YAML not modified; parity check optional |

## §B. Severity Convention

- **MUST-PASS (blocker)**: AC-001..AC-005, AC-DISTINCT-001, AC-DISTINCT-002, AC-007..AC-009. Failure blocks SPEC close.
- **SHOULD-PASS**: AC-006, AC-010. Failure surfaces as debt, does not block close.
- **OPTIONAL**: AC-011 (template-parity CI — explicitly out of scope per spec.md §E).

## §C. Build & Lint Baselines

### AC-001 — Build succeeds with new field

**Given** the implementer has added `AgenticLoopConfig` struct to `types.go` and the `AgenticLoop` field to `WorkflowConfig`,
**When** `go build ./internal/config/... ./internal/settings/... ./internal/cli/...` is run,
**Then** the build exits 0 with zero new warnings (relative to the M0 baseline recorded in plan.md §C).

### AC-006 — Field is visible at the required Go path (grep-visibility)

**Given** the implementer has added the `AgenticLoop` field to `WorkflowConfig` and the `AgenticLoopConfig` struct to `types.go`,
**When** `grep -rn 'AgenticLoop' internal/config/types.go` is run,
**Then** at least one line is returned naming the field at the struct path required by REQ-ALC-001 (`WorkflowConfig.AgenticLoop.MaxIterations` or a structurally equivalent path with distinct semantics from `LoopPrevention.MaxIterations`).

**Note**: AC-001 (build succeeds) and AC-006 (grep-visible at the required path) are distinct observations — a build can succeed with a differently-named or differently-placed field that compiles but does not satisfy REQ-ALC-001's path requirement. AC-006 asserts the naming/path convention; AC-001 asserts compilation.

### AC-002 — Explicit custom value parses

**Given** a YAML fixture `.moai/config/sections/workflow.yaml` containing:
```yaml
workflow:
    agentic_loop:
        max_iterations: 42
```
**When** `Loader.Load()` parses this fixture,
**Then** `cfg.Workflow.AgenticLoop.MaxIterations == 42`.

## §D. AC Matrix — Default Fallback

### AC-003 — Absent block or sub-key falls back to default (NEVER zero, NEVER 100)

This AC has TWO sub-cases that MUST both pass:

**AC-003a (absent block)**:
**Given** a YAML fixture with NO `agentic_loop:` block,
**When** `Loader.Load()` parses this fixture,
**Then** `cfg.Workflow.AgenticLoop.MaxIterations == DefaultAgenticLoopMaxIterations` (i.e. `10`),
**And** `cfg.Workflow.AgenticLoop.MaxIterations != 0`,
**And** `cfg.Workflow.AgenticLoop.MaxIterations != cfg.Workflow.LoopPrevention.MaxIterations` (which is `100` by default).

**AC-003b (present parent, absent sub-key)**:
**Given** a YAML fixture with:
```yaml
workflow:
    agentic_loop:
        # max_iterations omitted
```
**When** `Loader.Load()` parses this fixture,
**Then** `cfg.Workflow.AgenticLoop.MaxIterations == DefaultAgenticLoopMaxIterations`.

## §E. Single-Source-of-Truth Const

### AC-004 — Default const declared exactly once

**Given** the implementation is complete,
**When** `grep -n 'DefaultAgenticLoopMaxIterations' internal/config/defaults.go` is run,
**Then** exactly ONE declaration line is returned (the const definition).

### AC-005 — No hardcoded literal `10` outside defaults.go

**Given** the implementation is complete,
**When** the following grep is run across non-test Go source:
```bash
grep -rnE '\b10\b' internal/config/*.go internal/settings/*.go | \
  grep -v '_test.go' | \
  grep -v '^\S*defaults\.go:'
```
**Then** the output contains NO lines that set `MaxIterations` or default the agentic-loop field to the literal `10`. (Comments and unrelated occurrences of `10` are permitted; the AC fails only if a code-path default uses the literal where the const should be used.)

## §F. AC Matrix — Distinctness Invariant (the critical ACs)

### AC-DISTINCT-001 — Two distinct Go fields, separate defaults (SPEC-LEVEL INVARIANT)

**Given** the default configuration (no YAML overrides) is loaded via `NewDefaultConfig()` or `Loader.Load()` on a project with no `workflow.yaml`,
**When** the resulting `cfg.Workflow` struct is inspected,
**Then** ALL THREE of the following hold simultaneously:
1. `cfg.Workflow.AgenticLoop.MaxIterations == 10` (i.e. `DefaultAgenticLoopMaxIterations`),
2. `cfg.Workflow.LoopPrevention.MaxIterations == 100`,
3. The two fields occupy distinct memory addresses (i.e. modifying one at runtime does NOT modify the other — they are not backed by the same storage via pointer aliasing or a shared interface).

**Evidence format (5-section per verification-claim-integrity §3)**:
- **Claim**: the two fields are distinct.
- **Evidence**: test assertion output (AC-DISTINCT-002 output transcript) + a runtime mutation test `cfg.Workflow.AgenticLoop.MaxIterations = 999; assert cfg.Workflow.LoopPrevention.MaxIterations != 999`.
- **Baseline-attribution**: measured against this SPEC's commit, this tree, this test run.
- **Gaps**: none anticipated (the test is deterministic).
- **Residual-risk**: a future refactor that introduces a pointer-aliased default factory could regress this; the anti-aliasing test catches it.

### AC-DISTINCT-002 — Dedicated anti-aliasing test (parse-time distinctness)

**Given** a YAML fixture `.moai/config/sections/workflow.yaml` containing BOTH:
```yaml
workflow:
    agentic_loop:
        max_iterations: 7
    loop_prevention:
        max_iterations: 99
```
**When** `Loader.Load()` parses this fixture,
**Then** ALL FOUR of the following hold simultaneously:
1. `cfg.Workflow.AgenticLoop.MaxIterations == 7`,
2. `cfg.Workflow.LoopPrevention.MaxIterations == 99`,
3. Setting `agentic_loop.max_iterations` did NOT propagate to `LoopPrevention.MaxIterations` (i.e. `LoopPrevention.MaxIterations != 7`),
4. Setting `loop_prevention.max_iterations` did NOT propagate to `AgenticLoop.MaxIterations` (i.e. `AgenticLoop.MaxIterations != 99`).

**Test file location**: `internal/config/agentic_loop_distinctness_test.go` (dedicated; name makes the invariant grep-discoverable).

## §G. Schema Registration

### AC-007 — `schema_sections.go` registers the new key

**Given** the implementation is complete,
**When** `grep -n 'agentic_loop' internal/settings/schema_sections.go` is run,
**Then** at least one line is returned, in the form:
```go
s(SectionWorkflow, "workflow", TypeInt, "workflow", "agentic_loop", "max_iterations"),
```
**And** the path tokens are `"agentic_loop", "max_iterations"` (distinct from the `"loop_prevention", "max_iterations"` path at line 194).

### AC-008 — Path tokens are distinct (anti-collision)

**Given** the implementation is complete,
**When** `grep -nE '"agentic_loop", "max_iterations"|"loop_prevention", "max_iterations"' internal/settings/schema_sections.go` is run,
**Then** exactly TWO lines are returned — one for each distinct key path. The two MUST NOT collapse to a single registration.

## §H. Test Coverage

### AC-009 — Table-driven test suite exists

**Given** the implementation is complete,
**When** `ls internal/config/*agentic*` (or `grep -l 'AgenticLoop' internal/config/*_test.go`) is run,
**Then** at least one test file exists containing table-driven cases for: explicit value (AC-002), absent block (AC-003a), absent sub-key (AC-003b), and distinctness (AC-DISTINCT-002).

### AC-010 (SHOULD-PASS) — Coverage does not decrease

**Given** the implementation is complete,
**When** `go test -cover ./internal/config/...` is run,
**Then** the coverage percentage for `internal/config` is equal to or higher than the M0 baseline (plan.md §C).

## §I. Forward-Looking & Indirect Verification

### AC-011 (OPTIONAL) — Template-parity CI

**Given** the template-side `internal/template/templates/.moai/config/sections/workflow.yaml` ships `agentic_loop.max_iterations: 10` (verified),
**When** a future CI guard asserts `template_value == DefaultAgenticLoopMaxIterations`,
**Then** the parity holds.

This AC is OPTIONAL and explicitly out of scope for this SPEC (see spec.md §E "Out of Scope — Parity CI"). It is recorded here as a forward-looking check for whatever SPEC (or the run-phase implementer, if deemed warranted) adds the parity test. This SPEC does NOT require it.

## §J. Edge Cases (run-phase MUST consider)

| Edge case | Expected behavior | Covered by |
|---|---|---|
| YAML fixture sets `agentic_loop.max_iterations: 0` | Field is `0` (user-explicit; loader does NOT override to default — only absent-block triggers default). Distinctness invariant still holds. | AC-DISTINCT-002 logic (zero is a legal user-set value, distinct from absent) |
| YAML fixture sets `agentic_loop.max_iterations: -1` | Loader parses as-is (no validation required by REQ-ALC-* — runtime enforcement is out-of-scope per spec.md §E). | N/A (negative-value validation deferred to runtime SPEC) |
| YAML fixture has `agentic_loop:` as a map with sibling unknown keys | Non-strict mode silently ignores siblings (per `internal/config/CLAUDE.md` non-strict policy). | REQ-ALC-004 |
| YAML fixture has `agentic_loop: "not a map"` (type mismatch) | Loader raises `ConfigTypeError` (per `internal/config/CLAUDE.md` — type mismatch IS reported). | N/A (existing loader behavior; not in scope to change) |
| `.moai/config/sections/workflow.yaml` is entirely absent | Loader returns defaults — `AgenticLoop.MaxIterations == DefaultAgenticLoopMaxIterations`, `LoopPrevention.MaxIterations == 100`. Both defaults active. | AC-003a + defaults.go wiring |

## §K. Definition of Done (DoD)

The SPEC is "done" (ready for `draft → in-progress` transition by manager-develop, eventually `in-progress → implemented → completed` by manager-docs) when ALL of the following hold:

1. All MUST-PASS ACs (§B) pass — evidence recorded in `progress.md §E.2` by manager-develop.
2. `go test ./internal/config/... ./internal/settings/...` is green.
3. `go build ./...` succeeds.
4. `golangci-lint run --timeout=2m` is clean (or remaining warnings are pre-existing and unrelated).
5. Coverage for `internal/config` is non-decreasing (AC-010 SHOULD-PASS).
6. No inline literal `10` outside `defaults.go` (AC-005).
7. Anti-aliasing test (AC-DISTINCT-002) passes — this is the single most important check.
8. Parent SPEC `SPEC-MOAI-AGENTIC-LOOP-001` is untouched (era integrity).
9. No template file is modified (template already ships the block).
10. The `progress.md §E.2 Run-phase Evidence` section is populated by manager-develop with verbatim command output (per verification-claim-integrity §3.2 — verbatim output is the load-bearing artifact, not a summary).

**DoD blocker**: ANY failure on AC-DISTINCT-001 or AC-DISTINCT-002 blocks SPEC close, regardless of other ACs passing. The distinctness invariant is non-negotiable.
