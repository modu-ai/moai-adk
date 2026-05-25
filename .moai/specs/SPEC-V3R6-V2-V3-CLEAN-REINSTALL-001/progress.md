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
| M1 (Version bump + CHANGELOG) | COMPLETE | 5a18dd98f | ~89 lines (incl. 8 SPEC frontmatter + progress.md NEW) | Foundation (no binary AC) |
| M2 (Extend DeprecatedPaths) | COMPLETE | 68e3af7b1 | +474 net (dirs.go +234 + dirs_test.go +218 + golden +22) | AC-VVCR-005 PASS (43 entries; 9/31/3 split) |
| M2a (FLAT Layout Restoration) | PENDING — orchestrator handoff | — | est. ~50 lines path-string subst | AC-VVCR-LR-001/002/003/004/005 (5 ACs) |
| M3 (v2 detection logic) | PENDING — orchestrator handoff | — | est. ~250 lines (NEW Go file + test) | AC-VVCR-001 |
| M4 (Clean reinstall impl) | PENDING — orchestrator handoff | — | est. ~550 lines (2 NEW Go files + 2 NEW tests) | AC-VVCR-002/003/007/008/009/010/011/012/013 (9 ACs) |
| M5 (runUpdate integration + catalog regen) | PENDING — orchestrator handoff | — | est. ~80 lines | (wires M4 into CLI) |
| M6 (Test coverage + cross-platform) | PENDING — orchestrator handoff | — | est. ~400 lines | AC-VVCR-004/014/015/016/017 (5 ACs) |

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

### M1-cascade — Golden file updates (folded into M2 commit)

6 testdata golden files contained literal `v3.0.0-rc1` string snapshots:
- `internal/cli/testdata/doctor-{light,dark,nocolor}.golden`
- `internal/cli/testdata/status-{light,dark,nocolor}.golden`

The version bump from M1 broke `TestDoctor_Current_{Light,Dark}`, `TestDoctor_NoColor`, `TestStatus_Current_{Light,Dark}`, `TestStatus_NoColor`. Updated `v3.0.0-rc1` → `v3.0.0-rc2` in all 6 files (binary-safe Python script). Tests restored to PASS. Folded into the M2 commit because the cascade is one atomic consequence of the M1 version constant change and would create unnecessary commit fragmentation.

### M2 — Extend DeprecatedPaths table

Deliverables completed (commit `68e3af7b1`):
- `internal/defs/dirs.go` extended: 34 NEW `DeprecatedPathEntry` entries appended (Category B 31 v.2.x-era + Category C 3 rc1-stage staging artifacts) per spec.md §A.4. Total: 9 (Category A pre-existing, untouched) + 34 NEW = 43 entries.
- `internal/defs/dirs_test.go` NEW: 8 table-driven tests covering total count (43), per-category subtotals (9/31/3), required-fields invariant, Category B/C exact path enumeration, DeprecatedBy consistency, uniqueness, slash-separated normalization.
- @MX:ANCHOR added on DeprecatedPaths slice with @MX:REASON citing the SPEC and the 43-entry contract.

RED-GREEN cycle: dirs_test.go authored first → run failed (12 Category B missing + 3 Category C missing entries from the count mismatch + uniqueness checks) → dirs.go extended atomically → all 8 tests PASS.

Verification:
- `go test ./internal/defs/... -v` → 8/8 PASS
- `go test ./internal/cli/...` → all PASS (TestDoctor + TestStatus golden updates folded in)
- `go build ./...` → clean
- `go vet ./...` → clean

AC progress: **AC-VVCR-005 PASS** (Extended DeprecatedPaths enumeration verified by 43-entry count + per-category split assertion).

## §D — Partial-Completion Checkpoint (Run-phase Handoff)

Per `.claude/rules/moai/workflow/context-window-management.md` (1M context handoff threshold 50%) and `feedback_large_spec_wave_split.md` (>30-task SPEC wave-split mitigation), this run-phase is checkpointing after M1 + M2 (2 of 7 milestones) and returning a partial-completion report to the orchestrator.

### Rationale for handoff

The remaining milestones span substantial scope:
- **M2a**: ~14 git mv operations + 6 empty-dir removals + ~30 cross-reference grep+replace across `.claude/skills/`, `.claude/rules/`, `.claude/agents/moai/`, `CLAUDE.md`, `CLAUDE.local.md`, `.moai/specs/*/{spec,plan,acceptance,design,research}.md`
- **M3**: NEW `internal/cli/v2_detection.go` (~250 LOC) + `v2_detection_test.go` table-driven tests
- **M4**: NEW `internal/cli/update_clean_install.go` + `update_preserve_inventory.go` (~550 LOC combined) + 2 NEW test files
- **M5**: `internal/cli/update.go` integration (`runUpdate` modification + `runMigrateAgency` auto-invoke pattern) + `internal/template/catalog.yaml` regeneration via `gen-catalog-hashes.go --all`
- **M6**: 3-5 integration tests + cross-platform CI verification

Combined LOC: ~1300+ lines remaining, across multiple new packages and mechanical refactoring. The current context already carries the SPEC artifacts (140KB across 5 files), several large project rules (~50KB), and the M1+M2 implementation cycle. Continuing into M2a-M6 in this same session risks SSE stream stall and degraded reasoning quality on the architecturally novel M3+M4 work.

### What is safe to defer

M2a (FLAT layout restoration) is mechanical revert work — predictable scope, predictable risk. M3+M4+M5+M6 are novel implementation work that benefits from fresh context.

### Orchestrator next action

Resume the run-phase in a fresh session by re-spawning `manager-develop` with `cycle_type=tdd` and the following continuation prompt:

> Continue SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 run-phase from M2a. M1 (commit `5a18dd98f`) and M2 (commit `68e3af7b1`) are complete; 43-entry DeprecatedPaths table verified by 8 PASSing tests. progress.md §A reflects current state. Status frontmatter is already `in-progress` on all 5 SPEC artifacts. Begin with M2a FLAT layout restoration (mechanical revert of commit `1bd083725` — see plan.md §F.M2a for the 6-step deliverable list).

### Status transition note

The `draft → in-progress` transition is owned by manager-develop on M1 commit start per the Status Transition Ownership Matrix and was applied in commit `5a18dd98f`. The `in-progress → implemented` transition is reserved for manager-docs at sync-phase per the matrix and MUST NOT be applied at this checkpoint.

