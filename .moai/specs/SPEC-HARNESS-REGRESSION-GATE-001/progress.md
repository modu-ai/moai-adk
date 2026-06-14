---
id: SPEC-HARNESS-REGRESSION-GATE-001
title: "Progress — Harness M2-lite 비회귀 게이트"
version: "0.1.1"
status: completed
created: 2026-06-14
updated: 2026-06-14
author: manager-develop
priority: P1
phase: "v3.0.0"
module: "internal/harness, internal/measure"
lifecycle: spec-anchored
tier: M
tags: "harness, regression-gate, progress"
---

## §E — Phase 0.95 Mode Selection

**Input parameters**
- tier: M
- scope (file count): 8 files (2 new measure + 2 new regression_gate + applier.go edit + applier_test.go edit + go_feedback.go edit + coverage_boost_test.go edit)
- domain count: 2 (Go source — internal/measure leaf + internal/harness gate; one shared internal/loop refactor)
- file language mix: 100% Go source
- concurrency benefit: LOW (coding-heavy, sequential milestones with M1→M2→M3→M4 inter-milestone dependency)
- Agent Teams prereqs status: not evaluated (coding-heavy work; sequential default)

**Mode evaluation**

| Mode | Selected | Rationale |
|------|----------|-----------|
| trivial | no | Multi-file new-feature implementation, not a typo. |
| background | no | Write-heavy (Write/Edit) — background agents auto-deny writes. |
| agent-team | no | Coding-heavy, not multi-domain research; Agent Teams overhead unjustified. |
| parallel | no | Anthropic coding-task parallelism caveat; milestones have inter-dependency (M2 depends on M1, M4 on M3). |
| sub-agent | YES | Default fallback for coding-heavy sequential-milestone work (Mode 5). |
| workflow | no | Not ≥30-file mechanical-uniform transform; this is semantic new-code work. |

**Decision: sub-agent** (Mode 5 — sequential sub-agent per milestone)

**Justification**: This SPEC is coding-heavy with strict milestone ordering (M1 extract → M2 delegate → M3 build gate types → M4 wire gate → M5 preservation). Per Anthropic's coding-task parallelism caveat ("most coding tasks involve fewer truly parallelizable tasks than research"), the sequential sub-agent path is the safe default. The work fits in a single sub-agent operating through the milestones in order; no fan-out benefit applies.

---

## §E.2 Run-phase Evidence

### AC PASS/FAIL Matrix (run completion — all 13 PASS)

| AC | Status | Actual Output |
|----|--------|---------------|
| AC-RG-001 | PASS | `--- PASS` (measure parsers) + import-cycle CLEAN (no lsp/gopls/harness/loop dep) |
| AC-RG-002 | PASS | `--- PASS` (loop delegates; existing feedback/parse/collect tests GREEN) |
| AC-RG-003 | PASS | `--- PASS` TestBaselineStore_AbsentFile$ + TestBaselineStore_AtomicRoundTrip$ (no leftover .tmp) |
| AC-RG-004 | PASS | `--- PASS` Error$ + DistinctFromPending$; both error types coexist in source |
| AC-RG-005 | PASS | `--- PASS` TestApply_Regression_NonRegressing_Keeps$ (baseline updated, approved lineage) |
| AC-RG-006 | PASS | `--- PASS` Blocks_RollsBack$ + AppendsBlockedLineage$ (rollback byte-identical + regression-blocked entry) |
| AC-RG-007 | PASS | `--- PASS` TestSubagentBoundary_NoAskUserQuestion$ |
| AC-RG-008 | PASS | `--- PASS` SafetyArchitecture + SentinelCatalog + safety TestIsFrozen + tier ok; FROZEN git-diff empty |
| AC-RG-009 | PASS | grep: `typically Δ=0`/`always-pass` + `measurement-infrastructure scaffold` + `defense-in-depth` all present in spec.md |
| AC-RG-010 | PASS | grep `auto_apply: false` present in harness.yaml (unchanged) |
| AC-RG-011 | PASS | `--- PASS` TestCollector_AssemblesTriple$ |
| AC-RG-012 | PASS | `--- PASS` TestApply_Regression_ForbiddenFilesUntouched$ (usage-log/observations/tier-promotions unmodified) |
| AC-RG-013 | PASS | `--- PASS` TestApply_Regression_MeasurementError_FailsClosed$ (fail-closed roll back, no baseline update) |

### Milestone commit log

| Milestone | Status | Commit |
|-----------|--------|--------|
| M1+M2 (extract parsers → internal/measure + loop delegates) | done | `a384ce79c` |
| M3 (MetricTriple + ApplyRegressionError + baseline store) | done | `855a8b418` |
| M4 (in-Apply gate wiring + production seam) | done | `e202df00c` |
| M5 (FROZEN preservation + quality gate) | done | `4863466f7` |

### Quality gate (M5)

- `go test ./internal/harness/... ./internal/loop/... ./internal/measure/...` → all GREEN
- `go test ./...` → GREEN (note: `internal/hook/wrapper_test.go` is pre-existing flaky under full-parallel contention — 5s mock-binary subprocess timeout; passes in isolation and on re-run; OUTSIDE this SPEC's scope, never touched)
- `go list -deps internal/measure` → no `internal/(lsp|gopls|harness|loop)` (import-cycle proof)
- DO-NOT-MODIFY files → zero git diff
- `golangci-lint run --timeout=2m ./internal/harness/... ./internal/measure/... ./internal/loop/...` → 0 issues
- Coverage: `internal/measure` 98.0% (≥85%), `internal/harness` 86.1% (no-regression)
- C-HRA-008 boundary → GREEN (no AskUserQuestion in internal/harness non-test source)
- Cross-platform: `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` → exit 0

---

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-14
run_commit_sha: 4863466f7
run_status: implemented
ac_pass_count: 13
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: done (origin/main divergence 1 3 — parallel SEC-HARDEN-003 sync, disjoint scope; orchestrator owns rebase+push)
l44_post_push_fetch: (orchestrator-owned)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  host: pass
  windows_amd64: pass
total_run_phase_files: 8
m1_to_mN_commit_strategy: per-milestone scoped commits, Authored-By-Agent trailer (a384ce79c, 855a8b418, e202df00c, 4863466f7)
```

---

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_commit_sha: dcd5c2a3b
sync_phase_artifacts: 3 files (CHANGELOG.md + spec.md frontmatter + progress.md §E.2/§E.4/§E.5)
sync_status: completed
status_transition: in-progress → implemented
changelog_entry_position: [Unreleased] → Added section (second entry)
frontmatter_status_transitions: updated (status field only)
sync_phase_verification:
  changelog_entry_grep: SPEC-HARNESS-REGRESSION-GATE-001 (0 pre-existing duplicates confirmed)
  ac_count_match: 13 ACs from acceptance.md (matches CHANGELOG disclosure count)
  file_path_verification: CHANGELOG.md, spec.md, progress.md all exist
  git_status_clean: 3 files modified only (no collateral changes)
  build_pre_sync: success (run-phase pushed cleanly; no additional build needed)
```

---

## §E.5 Mx-phase Completion Signal

```yaml
mx_phase_status: completed (orchestrator-direct 4-phase close)
mx_commit_sha: 0ef92581d
mx_audit_ready: close-confirmed (spec.md status: completed, progress.md §E.2/§E.4/§E.5 complete, CHANGELOG entry signed, run+sync pushed to origin/main)
status_transition: implemented → completed
final_phase_marker: "chore(SPEC-HARNESS-REGRESSION-GATE-001): Mx-phase audit-ready signal + 4-phase close"
```
