---
id: SPEC-HARNESS-REGRESSION-GATE-001
title: "Progress — Harness M2-lite 비회귀 게이트"
version: "0.1.1"
status: in-progress
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

### AC PASS/FAIL Matrix (populated at run completion)

| AC | Status | Actual Output |
|----|--------|---------------|
| AC-RG-001 | (pending) | |
| AC-RG-002 | (pending) | |
| AC-RG-003 | (pending) | |
| AC-RG-004 | (pending) | |
| AC-RG-005 | (pending) | |
| AC-RG-006 | (pending) | |
| AC-RG-007 | (pending) | |
| AC-RG-008 | (pending) | |
| AC-RG-009 | (pending) | |
| AC-RG-010 | (pending) | |
| AC-RG-011 | (pending) | |
| AC-RG-012 | (pending) | |
| AC-RG-013 | (pending) | |

### Milestone commit log

| Milestone | Status | Commit |
|-----------|--------|--------|
| M1 (extract parsers → internal/measure) | done | (this commit) |
| M2 (loop delegates to measure) | done | (this commit) |
| M3 (MetricTriple + ApplyRegressionError + baseline store) | pending | |
| M4 (in-Apply gate wiring) | pending | |
| M5 (FROZEN preservation + quality gate) | pending | |

---

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: (pending)
run_commit_sha: (pending — backfill)
run_status: in-progress
ac_pass_count: 0
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: (pending)
l44_post_push_fetch: (orchestrator-owned)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  host: pass
  windows_amd64: pass
total_run_phase_files: 8
m1_to_mN_commit_strategy: per-milestone scoped commits, Authored-By-Agent trailer
```
