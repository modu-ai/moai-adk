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

# SPEC-V3R6-AGENTIC-LOOP-CONFIG-001 — Progress

> **Lifecycle phase**: plan-phase (artifact authored 2026-07-08 by manager-spec).
> **Next phase**: run-phase (manager-develop, cycle_type=tdd per quality.yaml).
> **Run-phase entry gate**: Implementation Kickoff Approval (CLAUDE.local.md §19.1) — user MUST explicitly approve run-phase start; plan-phase PASS does NOT authorize autonomous run-phase entry.

## §A. Plan-phase Artifact Status

| Artifact | File | Status |
|---|---|---|
| Specification | `spec.md` | draft authored |
| Implementation Plan | `plan.md` | draft authored |
| Acceptance Criteria | `acceptance.md` | draft authored |
| Progress Ledger | `progress.md` (this file) | skeleton authored |

## §B. Run-phase Milestone Tracking

> Placeholder — manager-develop populates each row as milestones complete.

| Milestone | Priority | Files | Status | Evidence |
|---|---|---|---|---|
| M1 — RED test suite (anti-aliasing + default-fallback) | P0 | `internal/config/agentic_loop_distinctness_test.go` (6 test functions) | PASS | compile-fail RED confirmed; 6 tests pass after GREEN (see §E.2) |
| M2 — GREEN production code (types + defaults + schema) | P1 | `internal/config/types.go` (+AgenticLoopConfig struct + AgenticLoop field), `internal/config/defaults.go` (+DefaultAgenticLoopMaxIterations const + factory wiring), `internal/settings/schema_sections.go` (+registration) | PASS | `go build ./...` exit 0; `go test` all ok; loader.go NO-OP (whole-section unmarshal auto-wires struct tag) |
| M3 — REFACTOR + coverage verification | P2 | `@MX:ANCHOR` tag added on AgenticLoopConfig (distinctness invariant) | PASS | coverage 80.3%; golangci-lint 0 issues; go vet clean |

## §C. Decision Log

| Decision | Rationale | Date |
|---|---|---|
| SPEC ID `SPEC-V3R6-AGENTIC-LOOP-CONFIG-001` | V3R6-era prefix matches modern convention; AGENTIC-LOOP-CONFIG domain associates with parent `SPEC-MOAI-AGENTIC-LOOP-001` while `-CONFIG` suffix scopes this to the Go loader surface (distinct from orchestration/skill body). Decomposition PASS. | 2026-07-08 |
| Host struct = existing `WorkflowConfig` (types.go:318) | The struct already exists with nested sub-struct pattern (`LoopPrevention`, `AutoClear`, `Team`); new `AgenticLoop AgenticLoopConfig` follows the same pattern — no new top-level struct. | 2026-07-08 |
| Tier S | Single config struct + parser wiring + tests; 4-5 file touches; no runtime enforcement; no template change. Fits Tier S scope. | 2026-07-08 |
| `cycle_type=tdd` recommendation | Loader/parse changes are pure-function testable; REQ-ALC-009 mandates table-driven tests; RED-first on anti-aliasing test (AC-DISTINCT-002) is the natural first step. | 2026-07-08 |

## §D. Open Questions / Blockers

> None at plan-phase close.

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase audit-ready: YES. The 3 plan-phase artifacts (spec.md, plan.md, acceptance.md) are internally consistent — every REQ-ALC-* traces to an AC; the distinctness invariant (§A.4) has dedicated ACs (AC-DISTINCT-001/002); the file-by-file change plan (plan.md §F M2) enumerates every file to be touched; the era is V3R6 (matches parent); the SPEC ID passed the Pre-Write Self-Check Protocol decomposition.

Plan-phase frontmatter is `status: draft` across all 4 plan-phase files (spec.md + plan.md + acceptance.md + progress.md). Status transition `draft → in-progress` is owned by manager-develop at run-phase entry, gated by the Implementation Kickoff Approval (CLAUDE.local.md §19.1).

## §E.2 Run-phase Evidence

**Run commit**: M1-M3 single run-phase commit (cycle_type=tdd, RED→GREEN→REFACTOR).
**Worktree branch**: `worktree-agent-acc98729c6e73f16f`.

### AC PASS/FAIL Matrix

| AC | Status | Verification Command | Result |
|---|---|---|---|
| AC-001 (build succeeds) | PASS | `go build ./...` | exit 0 (no output) |
| AC-002 (explicit value parses) | PASS | `go test -run TestAgenticLoopMaxIterations_ExplicitValue` | PASS |
| AC-003a (absent block → default) | PASS | `go test -run TestAgenticLoopMaxIterations_AbsentBlock_Default` | PASS |
| AC-003b (absent sub-key → default) | PASS | `go test -run TestAgenticLoopMaxIterations_AbsentSubKey_Default` | PASS |
| AC-004 (const declared once) | PASS | `grep -n DefaultAgenticLoopMaxIterations defaults.go` | exactly 1 declaration (line 41) |
| AC-005 (no hardcoded 10) | PASS | grep for agentic_loop literal 10 outside defaults.go | none |
| AC-006 (field grep-visible) | PASS | `grep AgenticLoop types.go` | field + struct confirmed |
| AC-007 (schema registered) | PASS | `grep agentic_loop schema_sections.go` | line 193 confirmed |
| AC-008 (distinct path tokens) | PASS | grep both path registrations | 2 distinct lines (193 agentic_loop, 195 loop_prevention) |
| AC-009 (table-driven tests) | PASS | `ls internal/config/*agentic*` | file exists, 6 cases |
| AC-010 (coverage non-decreasing) | PASS | `go test -cover ./internal/config/...` | 80.3% |
| **AC-DISTINCT-001** (DoD blocker) | **PASS** | `TestAgenticLoopDefault_NeverZero_NeverLoopPrevention` + `TestAgenticLoopDistinctness_RuntimeMutation` | both PASS |
| **AC-DISTINCT-002** (DoD blocker) | **PASS** | `TestAgenticLoopDistinctness_AntiAliasing` (7 vs 99) | PASS |

### Verbatim test output (AC-DISTINCT-001/002 + AC-002/003a/b)

```
$ go test ./internal/config/ -run 'TestAgenticLoop' -v
=== RUN   TestAgenticLoopMaxIterations_ExplicitValue
=== RUN   TestAgenticLoopMaxIterations_AbsentBlock_Default
=== RUN   TestAgenticLoopMaxIterations_AbsentSubKey_Default
=== RUN   TestAgenticLoopDistinctness_AntiAliasing
=== RUN   TestAgenticLoopDistinctness_RuntimeMutation
=== RUN   TestAgenticLoopDefault_NeverZero_NeverLoopPrevention
--- PASS: TestAgenticLoopDefault_NeverZero_NeverLoopPrevention (0.00s)
--- PASS: TestAgenticLoopDistinctness_RuntimeMutation (0.00s)
--- PASS: TestAgenticLoopMaxIterations_AbsentSubKey_Default (0.00s)
--- PASS: TestAgenticLoopMaxIterations_ExplicitValue (0.00s)
--- PASS: TestAgenticLoopDistinctness_AntiAliasing (0.00s)
--- PASS: TestAgenticLoopMaxIterations_AbsentBlock_Default (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/config	0.420s
```

### Verbatim full-suite + lint output

```
$ go test ./internal/config/... ./internal/settings/...
ok  	github.com/modu-ai/moai-adk/internal/config	1.410s
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	(cached)
ok  	github.com/modu-ai/moai-adk/internal/settings	0.766s
ok  	github.com/modu-ai/moai-adk/internal/settings/agentfm	(cached)
ok  	github.com/modu-ai/moai-adk/internal/settings/yamlpatch	(cached)

$ go test -cover ./internal/config/...
ok  	github.com/modu-ai/moai-adk/internal/config	1.193s	coverage: 80.3% of statements
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.485s	coverage: 91.1% of statements

$ golangci-lint run --timeout=2m
0 issues.

$ go vet ./internal/config/... ./internal/settings/...
(exit 0, no output)
```

### Files modified (run-phase)

| File | Change |
|---|---|
| `internal/config/agentic_loop_distinctness_test.go` | NEW — 6 test functions (anti-aliasing, default-fallback, runtime-mutation) |
| `internal/config/types.go` | +`AgenticLoopConfig` struct (with `@MX:ANCHOR` distinctness tag); +`AgenticLoop` field on `WorkflowConfig` |
| `internal/config/defaults.go` | +`DefaultAgenticLoopMaxIterations = 10` const; +factory wiring in `NewDefaultWorkflowConfig()` |
| `internal/settings/schema_sections.go` | +`s(SectionWorkflow, "workflow", TypeInt, "workflow", "agentic_loop", "max_iterations")` registration |

loader.go: NO-OP (plan-auditor confirmed whole-section unmarshal auto-wires the struct tag).
audit_struct_yaml_symmetry_test.go: NO-OP (WorkflowConfig not in the symmetryCases list — uses reflection, not enumeration).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-08
run_commit_sha: "<pending commit — orchestrator rebases + pushes>"
run_status: audit-ready
ac_pass_count: 13
ac_fail_count: 0
preserve_list_post_run_count: 0  # parent SPEC + template untouched; no out-of-scope modification
l44_pre_commit_fetch: n/a (worktree-isolated; orchestrator handles origin sync)
l44_post_push_fetch: n/a (deferred to orchestrator — no push per race-defense protocol)
new_warnings_or_lints_introduced: 0  # golangci-lint: 0 issues; go vet: clean
cross_platform_build:
  go_build_all: PASS  # exit 0
  goos_windows: not_run  # no syscall/build-tag changes; cross-platform N/A for config-struct addition
total_run_phase_files: 4  # 1 new test + 3 production files (types.go, defaults.go, schema_sections.go)
m1_to_mN_commit_strategy: single_commit  # M1+M2+M3 in one run-phase commit (Tier S, tight TDD cycle)
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

> manager-docs populates this section at sync-phase (3-phase close: lint + test + coverage delta + CHANGELOG/README sync).
