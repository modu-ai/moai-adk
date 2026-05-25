---
id: SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001
title: "Progress — v2-to-v3 Clean Reinstall (Tier M, cycle_type=tdd)"
version: "0.1.0"
status: in-progress
created: 2026-05-25
updated: 2026-05-25
author: manager-develop
priority: P1
phase: "v3.0.0-rc2"
module: "internal/cli, internal/defs, pkg/version, internal/template/templates/.claude/agents, .claude/agents"
lifecycle: spec-anchored
tags: "moai-update, v2-v3-migration, run-phase, progress, milestone-tracking"
tier: M
---

# Progress — SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 (Run-phase)

## §E — Phase 0.95 Mode Selection

### Input Parameters

- **tier**: M (per spec.md frontmatter `tier: M`)
- **scope (file count)**: ~38-42 files estimated per plan.md §F total
- **domain count**: 4 distinct domains touched (Go source `pkg/version/`, `internal/defs/`, `internal/cli/`; template mirrors `internal/template/templates/.claude/agents/`; on-disk agent layout `.claude/agents/`; CHANGELOG.md + SPEC artifact frontmatter)
- **file language mix**: Go source (~70% of LOC) + YAML config + Markdown frontmatter + git mv operations
- **concurrency benefit**: LOW (coding-heavy work per Finding A4 caveat — milestones M1-M6 have sequential dependencies M1→M2→M2a→M3→M4→M5→M6 per plan.md §F)
- **Agent Teams prereqs**: harness level NOT verified `thorough`; `workflow.team.enabled` NOT explicitly verified; `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` NOT verified — assumed FALSE for safety

### Mode Evaluation Table

| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 (trivial) | NO | Multi-milestone scope, ~1000 LOC, ~38 files — not a single-line typo |
| 2 (background) | NO | Run-phase requires Write/Edit operations — `run_in_background: true` would auto-deny per CONST-V3R2-020 |
| 3 (agent-team) | NO | Agent Teams capability gate prerequisites not all verified true; coding-heavy work per Finding A4 caveat |
| 4 (parallel) | NO | Milestones M1→M2→M2a→M3→M4→M5→M6 have sequential dependencies per plan.md §F; parallelism would violate dependency ordering |
| 5 (sub-agent) | **YES** | Sequential single-agent execution matches the dependency chain; coding-heavy work prefers Mode 5 per Finding A4 caveat |

### Decision: sub-agent (Mode 5)

### Justification

The SPEC has 6+1 milestones (M1, M2, M2a, M3, M4, M5, M6) with strict sequential dependencies declared in plan.md §F (e.g., M2 blocks M2a/M3/M4/M5; M3 blocks M4; M4 blocks M5; M5 blocks M6). Anthropic Finding A4 explicitly cautions against parallelism for coding tasks: *"most coding tasks involve fewer truly parallelizable tasks than research, and LLM agents are not yet great at coordinating and delegating to other agents in real time."* The single-agent (this `manager-develop` invocation) executes milestones in order, with each commit serving as a checkpoint. Modes 3 (agent-team) and 4 (parallel) are not selected because their prerequisites are unmet and the work profile is coding-heavy with cross-milestone state propagation (e.g., M2 extends `DeprecatedPaths` which M2a Category-C entries reference; M4 uses M3 detection logic).

## §A — Milestone Lifecycle

| Milestone | Status | Commit SHA | LOC delta | AC progress |
|-----------|--------|------------|-----------|-------------|
| M1 (Version bump + CHANGELOG) | IN-PROGRESS | (pending commit) | ~30 | (M1 sets foundation, no AC binary verifiable yet) |
| M2 (Extend DeprecatedPaths) | PENDING | — | — | — |
| M2a (FLAT Layout Restoration) | PENDING | — | — | — |
| M3 (v2 detection logic) | PENDING | — | — | — |
| M4 (Clean reinstall impl) | PENDING | — | — | — |
| M5 (runUpdate integration + catalog regen) | PENDING | — | — | — |
| M6 (Test coverage + cross-platform) | PENDING | — | — | — |

## §B — Status Transitions

| Date | Transition | Owning Agent | Trigger |
|------|------------|--------------|---------|
| 2026-05-25 | draft → in-progress | manager-develop | M1 commit start (per .claude/rules/moai/development/spec-frontmatter-schema.md Status Transition Ownership Matrix) |

All 5 SPEC artifacts (spec.md / plan.md / acceptance.md / design.md / research.md) frontmatter `status:` field updated `draft` → `in-progress`. The `updated:` field remains `2026-05-25` (same calendar day).

## §C — Run-phase Notes

### M1 — Version bump + CHANGELOG entry

Deliverables completed:
- `pkg/version/version.go` line 8: `Version = "v3.0.0-rc1"` → `Version = "v3.0.0-rc2"`
- `.moai/config/sections/system.yaml` lines 46-48: `template_version: v3.0.0-rc1` → `v3.0.0-rc2`; `version: v3.0.0-rc1` → `v3.0.0-rc2`
- `CHANGELOG.md`: Added new `### Added` subsection under `## [Unreleased]` for SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 describing the paradigm shift + FLAT layout restoration

Note on `internal/template/templates/.moai/config/config.yaml`: this file does NOT exist in the template directory (only `.moai/config/sections/*.yaml` files exist in templates). The plan.md M1 deliverable list mentioned this file; pre-flight verification confirmed its absence. The actionable version-bump scope therefore reduces to `pkg/version/version.go` + `.moai/config/sections/system.yaml` (project-local mirror).

`pkg/version/version_test.go` was inspected: existing tests (TestGetVersion, TestGetFullVersion, etc.) verify format and contract rather than the literal version string — they pass unchanged after the constant bump.

