---
id: SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001-progress
title: "Progress tracker — Worktree --team contextual launch"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/cli/worktree, internal/tmux"
lifecycle: spec-anchored
tags: "progress, tracker, milestones, tier-m"
tier: M
---

# Progress — SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001

## Status

**Current phase**: plan-phase (plan-auditor pending)
**Tier**: M
**Plan complete at**: 2026-05-23 (artifacts drafted)
**Plan status**: audit-ready

## Milestone Progress

| ID | Title | Status | Started | Completed | Notes |
|----|-------|--------|---------|-----------|-------|
| M1 | CG Mode Detector + Pattern Dispatcher Scaffolding | not-started | — | — | `internal/tmux/cg_detect.go` + `team_launch.go` skeleton + add `--team` cobra flag to `newNewCmd` (no dispatch yet) |
| M2 | Pattern P4 (Handoff) + P3 (syscall.Exec) + Mutex | not-started | — | — | Foundation patterns; injectable `syscallExecFn`; add `MarkFlagsMutuallyExclusive("team", "tmux")` |
| M3 | Pattern P1/P2 (Tmux Window Spawn) | not-started | — | — | `tmux new-window` inside existing session |
| M4 | Swarm Registry + Failure Mode Wiring | not-started | — | — | `.moai/state/swarm/<SPEC>.json` writer |
| M5 | Skill Body + Rule Updates + Template Mirrors | not-started | — | — | Template-First Rule discipline |
| M6 | Quality Verification + Cross-Platform Build + Coverage | not-started | — | — | Verification batch (10-command parallel) |

## Evidence Tracker

| ID | AC | Verification Command | Evidence (post-implementation) |
|----|----|---------------------|--------------------------------|
| AC-WTL-001 | P1 dispatch: tmux + CG → moai glm window | `go test -run TestTeamLaunch_P1_TmuxCG ./internal/cli/worktree/` | _pending_ |
| AC-WTL-002 | P2 dispatch: tmux + CC → moai cc window | `go test -run TestTeamLaunch_P2_TmuxCC ./internal/cli/worktree/` | _pending_ |
| AC-WTL-003 | P3 dispatch: no tmux → syscall.Exec | `go test -run TestTeamLaunch_P3_NoTmux_SyscallExec ./internal/cli/worktree/` | _pending_ |
| AC-WTL-004 | P4 dispatch: --team absent → handoff guidance | `go test -run TestTeamLaunch_P4_NoFlag_Handoff ./internal/cli/worktree/` | _pending_ |
| AC-WTL-005 | CG mode detection 4-scenario boolean | `go test -run TestIsCGMode ./internal/tmux/` | _pending_ |
| AC-WTL-006 | BODP HARD: TestNew_NoAskUserQuestion green | `go test -run TestNew_NoAskUserQuestion ./internal/cli/worktree/` | _pending_ |
| AC-WTL-007 | Pane spawn failure → P4 fallback | `go test -run TestTeamLaunch_PaneSpawnFailure_FallbackToP4 ./internal/cli/worktree/` | _pending_ |
| AC-WTL-008 | Swarm registry schema (7 fields + 0o600) | `go test -run TestSwarmRegistry_P1_Schema ./internal/cli/worktree/` | _pending_ |
| AC-WTL-009 | Cross-platform builds darwin/linux/windows | `GOOS={darwin,linux,windows} GOARCH=amd64 go build ./...` | _pending_ |
| AC-WTL-010 | Invalid SPEC → no launch, no registry | `go test -run TestTeamLaunch_WorktreeCreateFailure_NoLaunch ./internal/cli/worktree/` | _pending_ |
| AC-WTL-011 | Lint baseline: 0 NEW violations | `golangci-lint run --timeout=2m` (diff vs baseline) | _pending_ |
| AC-WTL-012 | Test coverage ≥85% for new files | `go tool cover -func=cover.out` per-file threshold | _pending_ |
| AC-WTL-013 | Skill body + template mirror byte-identical | `diff -r .claude/skills/moai-workflow-worktree/ internal/template/templates/.claude/skills/moai-workflow-worktree/` | _pending_ |
| AC-WTL-014 | BODP audit trail preserved | `go test -run TestTeamLaunch_BODPAuditTrailPreserved ./internal/cli/worktree/` | _pending_ |

## Test Plan

### Manual Verification (post-implementation, not in run-phase scope)

1. **End-to-end P1**: tmux + `moai cg` (sets CG mode) → `moai worktree new SPEC-WTL-E2E-001 --team` → confirm new tmux window with `moai glm` running, cwd = worktree path.
2. **End-to-end P2**: tmux + `moai cc` (resets CG) → `moai worktree new SPEC-WTL-E2E-002 --team` → confirm new tmux window with `moai cc` running.
3. **End-to-end P3**: outside tmux (fresh shell) → `moai worktree new SPEC-WTL-E2E-003 --team` → confirm process replaces itself with `moai cc`, cwd = worktree.
4. **End-to-end P4**: `moai worktree new SPEC-WTL-E2E-004` (no --team) → confirm handoff guidance printed, no spawn.

### Automated Coverage (run-phase scope)

- M1 unit tests: `IsCGMode` 5 subtests
- M2 unit tests: P3 syscall.Exec capture + P4 golden output
- M3 unit tests: P1/P2 tmux new-window capture (6 subtests)
- M4 unit tests: swarm registry schema + permissions (4 subtests)
- M5 verification: byte-identical mirror diff
- M6 verification batch: 10-command parallel exec via single-turn multi-Bash call

## Risks Tracker

| Risk | Status | Mitigation Applied |
|------|--------|-------------------|
| R1: syscall.Exec uncoverable in tests | open | Injectable `syscallExecFn` var (proven pattern) |
| R2: Windows build fail due to syscall.Exec | open | `//go:build !windows` tag split |
| R3: Coverage <85% due to uncoverable lines | open | Per-package threshold + nolint annotations |
| R4: glm.SettingsLocal import cycle | open | Local minimal struct in `cg_detect.go` |
| R5: tmux new-window vs new-session confusion | resolved | OQ-1 resolved: new-window |
| R6: Orphaned worktree on pane spawn failure | accepted | Exit 0 + P4 handoff guidance gives user manual recovery |
| R7: settings.local.json accidental mutation | open | No os.WriteFile on settings.local.json in team_launch.go |
| R8: BODP HARD violation via AskUserQuestion | open | Static guard extended in AC-WTL-006 |
| R9: --team + --tmux combination undefined | resolved | OQ-2 resolved: mutually exclusive |
| R10: Swarm registry orphans on Ctrl-C | accepted | Registry write happens only AFTER successful launch dispatch |

## Plan-Phase Self-Check Result

Performed internal self-check per orchestrator instructions:

- [x] Every REQ has at least one mapped AC (14 ACs across 12 REQs; REQ-WTL-000 is pre-flight gate)
- [x] Every AC verifies a specific REQ or HARD constraint
- [x] Traceability matrix is complete (acceptance.md §3)
- [x] Out of Scope is explicit (spec.md §4.1 + plan.md OQ resolutions)
- [x] HARD constraints listed (spec.md §5 — 8 constraints)
- [x] Tier classification: M (5-15 files affected, 10 distinct files in inventory, ~640-740 source LOC + ~580 test LOC; stays below Tier L >1000 source LOC threshold for non-constitutional changes)
- [x] Frontmatter 12-field canonical schema applied (id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags, plus optional tier)
- [x] No file modifications outside `.moai/specs/SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001/`
- [x] No implementation code touched (plan-phase only; M1-M6 deferred to run-phase)
- [x] BODP HARD compliance: REQ-WTL-006 + AC-WTL-006 enforce CLI no-AskUserQuestion rule (CONST-V3R5-030)
- [x] Template-First Rule: REQ-WTL-011 + AC-WTL-013 enforce byte-identical mirror in `internal/template/templates/`

## Notes

- This SPEC ships independently of Sprint 2 in-progress SPECs (HOOK-OBSERVE-OPT-IN, HOOK-ASYNC-EXPAND, AGENT-MODEL-ROUTING, PROMPT-CACHE). Can be queued anytime in the Sprint 2 sequence.
- Baseline references verified during artifact drafting: `internal/cli/worktree/new.go` (346 LOC), `internal/tmux/detector.go` (119 LOC), `internal/cli/launcher.go` `applyCGMode` (L143-204), `internal/cli/worktree/tmux_integration.go` `CreateTmuxSession` (L48-94), `internal/cli/glm.go` `SettingsLocal.TeammateMode` (L103-108), `TestNew_NoAskUserQuestion` (L732-742).
- After plan-PR merge, run-phase invocation: `/moai run SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001` (Tier M → Section A-E delegation template REQUIRED per `manager-develop-prompt-template.md` § Applicability).
