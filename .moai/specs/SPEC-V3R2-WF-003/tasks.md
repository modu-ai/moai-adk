# SPEC-V3R2-WF-003 Task Breakdown

> Granular task decomposition of M1-M5 milestones from `plan.md` §2.
> Companion to `spec.md` v0.2.0, `research.md` v0.1.0, `plan.md` v0.1.0, `acceptance.md` v0.1.0.

## HISTORY

| Version | Date       | Author                            | Description                                                            |
|---------|------------|-----------------------------------|------------------------------------------------------------------------|
| 0.1.0   | 2026-05-03 | MoAI Plan Workflow (Phase 1B)     | Initial task breakdown — 19 tasks (T-WF003-01..19) across M1-M5         |
| 0.2.0   | 2026-05-04 | MoAI Plan Workflow (iter 2)       | T-WF003-18 scope expanded from 2 (plan/sync) → 4 (plan, sync, project, db) mode-NA skill bodies per plan-auditor iteration 1 D4 fix. AC count references updated 15→17 (D1/D2 added AC-16/17). |

---

## Task ID Convention

- ID format: `T-WF003-NN`
- Priority: P0 (blocker), P1 (required), P2 (recommended), P3 (optional)
- Owner role: `manager-tdd`, `manager-docs`, `expert-backend` (Go test), `manager-git` (PR base transition)
- Dependencies: explicit task ID list; tasks with no deps may run in parallel within their milestone
- DDD/TDD alignment: per `.moai/config/sections/quality.yaml` `development_mode: tdd`, M1 (RED) precedes M2-M4 (GREEN) precedes M5 (REFACTOR + Trackable)

[HARD] No time estimates per `.claude/rules/moai/core/agent-common-protocol.md` §Time Estimation. Priority + dependencies only.

---

## M1: Test Scaffolding Extension (RED phase)

Goal: Extend WF-004's `agentless_audit_test.go` (created in WF-004 M1) with three new test functions that fail until M2-M4 add the required content.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-WF003-01 | Verify WF-004 M1's `internal/template/agentless_audit_test.go` exists in stacked base (HEAD `5ab409292`). If absent, escalate to user (M1 cannot proceed without WF-004 M1 baseline). | manager-tdd | n/a (verification only) | none | 0 files | RED prerequisite |
| T-WF003-02 | Implement `TestRunDesignSkillsContainModeUnknownSentinel` walking `.claude/skills/moai/workflows/{run,design}.md` and asserting each contains the literal `MODE_UNKNOWN`. | expert-backend | `internal/template/agentless_audit_test.go` (extend, ~30 LOC) | T-WF003-01 | 1 file (extend) | RED — must compile and fail (sentinel not yet in either skill) |
| T-WF003-03 | Implement `TestRunSkillContainsModeTeamUnavailableSentinel` walking `.claude/skills/moai/workflows/run.md` and asserting it contains the literal `MODE_TEAM_UNAVAILABLE`. | expert-backend | `internal/template/agentless_audit_test.go` (extend, ~20 LOC) | T-WF003-01 | 1 file (extend) | RED — must compile and fail |
| T-WF003-04 | Implement `TestLoopAliasCrossReference` walking `.claude/skills/moai/workflows/loop.md` and asserting it contains the literal phrase `/moai run --mode loop`. | expert-backend | `internal/template/agentless_audit_test.go` (extend, ~20 LOC) | T-WF003-01 | 1 file (extend) | RED — must compile and fail |
| T-WF003-05 | Run `go test ./internal/template/ -run "TestRunDesign\|TestRunSkill\|TestLoopAlias"` and confirm RED state (3 functions × ~4 subtests fail). Verify pre-existing WF-004 audit tests remain GREEN. | manager-tdd | n/a (verification only) | T-WF003-02, T-WF003-03, T-WF003-04 | 0 files | RED gate verification |

**M1 priority: P0** — blocks all subsequent milestones. M1 must complete before M2/M3/M4 begin (TDD discipline).

T-WF003-02 / 03 / 04 may technically execute in parallel — they all extend the same Go file, but each test func is independently writable. Per `CLAUDE.md` §14 Multi-File Decomposition HARD rule, however, since they touch the same file, sequential execution within a single edit pass is recommended.

---

## M2: Mode Dispatch section in `/moai run` skill (GREEN, part 1)

Goal: Make `TestRunDesignSkillsContainModeUnknownSentinel/run.md` and `TestRunSkillContainsModeTeamUnavailableSentinel` pass.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-WF003-06 | Locate the existing `## Mode Flag Compatibility` section in `.claude/skills/moai/workflows/run.md` (added by WF-004 M3). EXTEND that section into a `## Mode Dispatch (Multi-Mode Router)` section per plan.md §2 M2 template. Include sentinels `MODE_UNKNOWN`, `MODE_TEAM_UNAVAILABLE`, `MODE_PIPELINE_ONLY_UTILITY` (preserve last). Document 4 mode values, mode resolver pseudocode, harness-based default selection (REQ-WF003-002, 003), silent downgrade behavior (REQ-WF003-012), CLI > config > harness precedence (REQ-WF003-018). | manager-docs | `.claude/skills/moai/workflows/run.md` (extend ~60 lines) | T-WF003-05 | 1 file (edit) | GREEN |
| T-WF003-07 | Mirror T-WF003-06 edit into `internal/template/templates/.claude/skills/moai/workflows/run.md`. | manager-docs | `internal/template/templates/.claude/skills/moai/workflows/run.md` | T-WF003-06 | 1 file (edit, parity) | Embedded-template parity |
| T-WF003-08 | Run `make build` in worktree to regenerate `internal/template/embedded.go`. Verify diff is exactly the run.md content addition. | manager-docs | `internal/template/embedded.go` (regenerated) | T-WF003-07 | 1 file (regenerated) | Build verification |
| T-WF003-09 | Run `go test ./internal/template/ -run "TestRunSkillContainsModeTeamUnavailableSentinel"` and confirm PASS. Run `TestRunDesignSkillsContainModeUnknownSentinel/run.md` subtest only and confirm PASS (design.md still RED). | manager-tdd | n/a (verification only) | T-WF003-08 | 0 files | GREEN gate part 1 |

**M2 priority: P0** — blocks M3, M4d (which reference run.md Mode Dispatch section).

---

## M3: Mode Dispatch section in `/moai design` skill (GREEN, part 2)

Goal: Make `TestRunDesignSkillsContainModeUnknownSentinel/design.md` pass.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-WF003-10 | Locate the existing `## Mode Flag Compatibility` section in `.claude/skills/moai/workflows/design.md` (added by WF-004 M3). EXTEND that section into a `## Mode Dispatch (Multi-Mode Router)` section per plan.md §2 M3 template. Include 4 mode values for design (`autopilot|import|team|pipeline`), MODE_UNKNOWN sentinel, MODE_PIPELINE_ONLY_UTILITY (preserve), Path A/B-Common dispatch logic (REQ-WF003-009, 013), Path B1/B2 explicit-not-in-axis note. Cross-reference run.md for precedence + team prereqs. | manager-docs | `.claude/skills/moai/workflows/design.md` (extend ~50 lines) | T-WF003-09 | 1 file (edit) | GREEN |
| T-WF003-11 | Mirror T-WF003-10 edit into `internal/template/templates/.claude/skills/moai/workflows/design.md`. | manager-docs | `internal/template/templates/.claude/skills/moai/workflows/design.md` | T-WF003-10 | 1 file (edit, parity) | Embedded-template parity |
| T-WF003-12 | Run `make build` to regenerate `internal/template/embedded.go`. Verify diff. Run `go test ./internal/template/ -run "TestRunDesignSkillsContainModeUnknownSentinel"` and confirm PASS (both run.md and design.md subtests GREEN). | manager-tdd | n/a (verification only) | T-WF003-11 | 1 file (regenerated) | GREEN gate part 2 |

**M3 priority: P0** — required for spec-WF003-001 (design supports `--mode` axis) compliance.

[HARD] T-WF003-10 must NOT modify the existing Phase 0/1/A/B1/B2/B-Common/C structure of design.md. Insert-only into the Mode Flag Compatibility section.

---

## M4: Loop alias note + matrix extension + default_mode schema + plan/sync mode-NA (GREEN, part 3)

Goal: Make `TestLoopAliasCrossReference` pass and publish the unified Subcommand × Mode matrix + config schema extension.

This milestone has 4 sub-tasks (M4a-M4d) per plan.md §2 M4. Each sub-task has its own task IDs.

### M4a: `/moai loop` alias header note

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-WF003-13 | Insert `## Invocation Routes (SPEC-V3R2-WF-003)` section in `.claude/skills/moai/workflows/loop.md` after line 33 (flow declaration). Body per plan.md §2 M4a template. Include literal phrase `/moai run --mode loop` to satisfy `TestLoopAliasCrossReference`. | manager-docs | `.claude/skills/moai/workflows/loop.md` (insert ~10 lines) | T-WF003-12 | 1 file (edit) | GREEN |
| T-WF003-14 | Mirror T-WF003-13 into `internal/template/templates/.claude/skills/moai/workflows/loop.md`. Run `make build`. | manager-docs | template file + `internal/template/embedded.go` | T-WF003-13 | 2 files (edit + regenerated) | Embedded-template parity |

### M4b: Subcommand × Mode matrix extension in spec-workflow.md

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-WF003-15 | Locate the existing `## Subcommand Classification (Pipeline vs Multi-Agent)` section in `.claude/rules/moai/workflow/spec-workflow.md` (added by WF-004 M4). EXTEND the matrix table by adding 3 new columns (Default mode, Valid `--mode` values, Sentinel on invalid mode) and 1 new row for `/moai loop` alias. INSERT a new `### Mode Dispatch Cross-Reference` sub-section immediately after the matrix per plan.md §2 M4b template (~40 lines). | manager-docs | `.claude/rules/moai/workflow/spec-workflow.md` (extend matrix + insert sub-section) | T-WF003-14 | 1 file (edit) | GREEN |
| T-WF003-16 | Mirror T-WF003-15 into `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md`. Run `make build`. | manager-docs | template file + `internal/template/embedded.go` | T-WF003-15 | 2 files (edit + regenerated) | Embedded-template parity |

### M4c: default_mode schema extension in workflow.yaml

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-WF003-17 | Add new optional `default_mode: ""` field at top of `workflow:` block in `.moai/config/sections/workflow.yaml` (before `auto_clear:`) with comment per plan.md §2 M4c template. Mirror to `internal/template/templates/.moai/config/sections/workflow.yaml`. Run `make build`. | manager-docs | `.moai/config/sections/workflow.yaml` (insert ~5 lines) + template mirror + embedded.go | T-WF003-12 (no dep on M4a/b) | 3 files (edit + parity + regen) | GREEN |

[HARD] T-WF003-17 MUST NOT modify any other key in `workflow.yaml`. Only the new `default_mode: ""` field is added.

### M4d: Mode-NA notation in plan.md and sync.md

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-WF003-18 | Locate the existing `## Mode Flag Compatibility` section in **all 4** `.claude/skills/moai/workflows/{plan,sync,project,db}.md` skill bodies (all added by WF-004 M3 — covers all 4 mode-NA implementation subcommands per REQ-WF003-005). Pre-edit: verify each skill currently does NOT parse `--mode` (grep check); if any does, escalate. EXTEND each `## Mode Flag Compatibility` section with REQ-WF003-005 clarification per plan.md §2 M4d template (~10 lines each, with subcommand name substitution). Mirror all 4 into `internal/template/templates/.claude/skills/moai/workflows/{plan,sync,project,db}.md`. Run `make build`. Run full `go test ./...` and confirm 0 failures (per CLAUDE.local.md §6 HARD rule). | manager-docs | 8 skill files (4 source + 4 template) + embedded.go | T-WF003-12 (no dep on M4a/b/c) | 9 files (edit + parity + regen) | GREEN final gate |

**M4 priority: P1** — required deliverable; M4a/M4b/M4c/M4d may execute in parallel (they touch independent files), but per `CLAUDE.md` §14 Multi-File Decomposition HARD rule with 3+ files, sequential execution per sub-task is recommended.

After M4 completion: all 3 new audit tests pass; pre-existing WF-004 tests still pass; full repository test suite is GREEN.

---

## M5: Documentation Sync + MX Tags + Cross-Links (REFACTOR + Trackable)

Goal: TRUST 5 Trackable + MX tag insertion + cross-link footers.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-WF003-19 | Add CHANGELOG entry under `## [Unreleased]` per plan.md §2 M5. Add cross-link footer to 5 affected skill files (`run.md`, `design.md`, `loop.md`, `plan.md`, `sync.md`) pointing to `spec-workflow.md#subcommand-classification`. Insert 8 MX tags across 6 distinct file locations per plan.md §6. Mirror all skill edits to template tree. Update `progress.md` with `run_complete_at` and `run_status: implementation-complete`. Run `make build` + final `go test ./...` verification. | manager-docs | `CHANGELOG.md` (~10 lines), 5 skill files (cross-link footers, ~3 lines each), 6 files (MX tag inserts), 5 template mirrors, `progress.md` update, `internal/template/embedded.go` (regenerated) | T-WF003-18 | ~17 files | REFACTOR / Trackable |

**M5 priority: P2** — quality polish; could be split if size becomes unwieldy in the run phase.

[HARD] T-WF003-19 ends with a final `go test ./...` after `make build`. No tests may regress (per CLAUDE.local.md §6 HARD).

---

## Pre-Merge Hook (Orthogonal to Milestones)

This is NOT a milestone — it's a CI / human-managed coordination step that must occur before PR #765 (WF-004) merges to main.

| ID | Subject | Owner role | Trigger | Touch points | Notes |
|----|---------|-----------|---------|--------------|-------|
| T-WF003-PRE | Monitor PR #765 (WF-004) status. When CI all-green AND reviews approved AND merge imminent: run `gh pr edit <WF-003-PR-#> --base main` to retarget the WF-003 PR base from `feature/SPEC-V3R2-WF-004` to `main`. Verify diff still contains only WF-003-specific changes. Resolve any merge conflicts that appear. | manager-git (or user manual) | PR #765 approaching merge | 0 source files (PR metadata change) | CLAUDE.local.md §18.11 v2.14.0 case study reference. Failure to do so = WF-003 PR auto-closes when PR #765 merges. |

---

## Aggregate Statistics

- **Total tasks**: 19 (numbered T-WF003-01..19) + 1 pre-merge hook (T-WF003-PRE)
- **Total milestones**: 5 (M1: 5 tasks, M2: 4 tasks, M3: 3 tasks, M4: 6 tasks, M5: 1 task)
- **Files created**: 0 (all edits are extensions of existing files)
- **Files modified (source-of-truth)**: 10
  - 7 skills (`run.md`, `design.md`, `loop.md`, `plan.md`, `sync.md`, `project.md`, `db.md`)
  - 1 rule (`spec-workflow.md`)
  - 1 yaml (`workflow.yaml`)
  - 1 Go test (`agentless_audit_test.go`, extend)
- **Files modified (CHANGELOG)**: 1
- **Files modified (embedded-template parity)**: 9 (mirrors of 7 skills + 1 rule + 1 yaml; agentless_audit_test.go is not in templates dir)
- **Files modified (regenerated)**: 1 (`internal/template/embedded.go`)
- **MX tag insertions**: 8 across 6 distinct file locations (per plan.md §6)
- **Owner-role distribution**:
  - `expert-backend`: 3 tasks (T-WF003-02..04, all Go test work)
  - `manager-tdd`: 5 tasks (T-WF003-01, 05, 09, 12, 18 — TDD gate verification)
  - `manager-docs`: 11 tasks (all content authoring + template parity + final sync)
  - `manager-git` (or user): 1 hook (T-WF003-PRE — pre-merge base transition)

---

## Owner-Role Rationale

- **Go test work** (`expert-backend`): the audit test extensions are Go code; needs Go expertise (regex, fs.WalkDir, embedded FS bytes.Contains).
- **TDD gate verification** (`manager-tdd`): the project's `quality.yaml` declares `development_mode: tdd`, so manager-tdd is the methodology owner. Each gate (RED, GREEN parts 1/2/3, final GREEN) is a manager-tdd checkpoint. T-WF003-01 verification of WF-004 baseline is also manager-tdd's responsibility.
- **Content authoring** (`manager-docs`): all skill/rule/yaml/CHANGELOG edits are documentation per `.claude/rules/moai/development/coding-standards.md` (skills are config-as-code documents). manager-docs is the owner per `CLAUDE.md` §4 Manager Agents catalog.
- **PR base transition** (`manager-git` or user): git operation requiring `gh` CLI; manager-git is the agent for git operations per CLAUDE.md. May be performed manually by user if agent is not invoked at the right moment.

---

## Parallel Execution Opportunities

These task groups have no inter-dependencies and may run in parallel within their milestone:

- **M1 parallel**: T-WF003-02, T-WF003-03, T-WF003-04 (all extend the same file; cannot truly parallelize without merge conflict — execute sequentially in same edit pass)
- **M4 parallel**: T-WF003-13 (M4a), T-WF003-15 (M4b), T-WF003-17 (M4c), T-WF003-18 (M4d) touch independent files → true parallelism possible. Per `CLAUDE.md` §14 Multi-File Decomposition HARD rule (3+ files), decomposition + sequencing recommended; the per-sub-task task IDs already encode this.

Per `CLAUDE.md` §1 HARD rule "Multi-File Decomposition: Split work when modifying 3+ files," M4 SHOULD be split into per-sub-task edits. The task IDs T-WF003-13..18 already encode this.

---

## Cross-Reference Map

Each task references which REQ(s) and which AC(s) it advances toward DoD:

| Task ID | REQ coverage | AC coverage |
|---------|--------------|-------------|
| T-WF003-01 | (prerequisite verification) | (gate) |
| T-WF003-02 | REQ-WF003-010 | AC-WF003-06 |
| T-WF003-03 | REQ-WF003-011 | AC-WF003-07 |
| T-WF003-04 | REQ-WF003-004 | AC-WF003-03 |
| T-WF003-05 | (gate) | (gate) |
| T-WF003-06 | REQ-WF003-001, 002, 003, 008, 010, 011, 012, 014, 016, 017, 018 | AC-WF003-01, 02, 04, 06, 07, 08, 11, 12, 13, 14 |
| T-WF003-07 | (parity) | (parity) |
| T-WF003-08 | (build) | (build) |
| T-WF003-09 | (gate) | AC-WF003-07 GREEN |
| T-WF003-10 | REQ-WF003-001, 009, 010, 013, 014, 016, 018 | AC-WF003-05, 06, 11, 12, 13, 15 |
| T-WF003-11 | (parity) | (parity) |
| T-WF003-12 | (gate) | AC-WF003-06 GREEN (full) |
| T-WF003-13 | REQ-WF003-004 | AC-WF003-03 |
| T-WF003-14 | (parity) | (parity) |
| T-WF003-15 | REQ-WF003-007, 016, 018 | AC-WF003-09, 11, 13, 16 |
| T-WF003-16 | (parity) | (parity) |
| T-WF003-17 | REQ-WF003-014 | AC-WF003-12 |
| T-WF003-18 | REQ-WF003-005, 016 | AC-WF003-09, 11 (plan/sync sentinel preservation) |
| T-WF003-19 | (Trackable / MX) | (DoD steps 11-13) |
| T-WF003-PRE | (PR coordination) | (DoD step 14) |

REQ coverage summary: 18 REQs, all referenced by at least one task (verified by transitive lookup against `spec.md` §5). REQ-WF003-007 → T-WF003-15 (matrix publication); REQ-WF003-015 → T-WF003-15 + T-WF003-19 (architectural-contract layer documented in spec-workflow.md cross-ref + CHANGELOG).

AC coverage summary: 17 ACs, all advanced toward DoD by tasks. AC-WF003-10 is implicit via WF-004's existing utility skill audit (not modified by WF-003 — preserved). AC-WF003-16 (matrix publication) advanced by T-WF003-15/16. AC-WF003-17 (future-extension schema contract) advanced by T-WF003-15/16 + T-WF003-19 (architectural-contract documentation).

---

## Final Verification Pass (subset of T-WF003-19)

Before marking SPEC implementation complete, execute these checks in order:

1. **New audit tests green**: `go test ./internal/template/ -run "TestRunDesign|TestRunSkill|TestLoopAlias" -v` — 3 funcs, ~7 subtests, all PASS.
2. **WF-004 audit tests still green**: `go test ./internal/template/ -run "TestAgentless|TestUtilitySkills|TestImplementationSkills" -v` — 3 funcs, ~14 subtests, all PASS.
3. **Full repo green**: `go test ./...` — 0 failures (per `CLAUDE.local.md` §6 HARD).
4. **Lint clean**: `golangci-lint run` — 0 errors, ≤10 warnings (per `quality.yaml` `lsp_quality_gates.sync`).
5. **Vet clean**: `go vet ./...` — 0 issues.
6. **Embedded parity**: diff `internal/template/embedded.go` after `make build` shows only the 7 expected file content changes (no spurious diffs).
7. **DoD checklist**: all 14 items in `acceptance.md` §"Definition of Done" checked.
8. **Pre-merge hook awareness**: T-WF003-PRE is documented; user/manager-git aware that PR base transition is required before PR #765 merges.

If any step fails, return to the appropriate milestone for remediation. Do NOT advance to `/moai sync` until all 8 checks pass.

---

End of tasks.md.
