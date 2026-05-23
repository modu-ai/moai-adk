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

**Current phase**: run-phase (M4 complete)
**Tier**: M
**Plan complete at**: 2026-05-23 (artifacts drafted)
**Plan status**: audit-ready

## Milestone Progress

| ID | Title | Status | Started | Completed | Notes |
|----|-------|--------|---------|-----------|-------|
| M1 | CG Mode Detector + Pattern Dispatcher Scaffolding | completed | 2026-05-23 | 2026-05-23 | `internal/tmux/cg_detect.go` + `team_launch.go` skeleton + `--team` cobra flag added to `newNewCmd` (no dispatch yet). TDD RED→GREEN: 9 IsCGMode subtests + TestDecidePattern (5 cases) + TestPatternString (5 cases) + TestTeamLaunchConfig_ZeroValue all PASS. Coverage cg_detect.go 94.1%, team_launch.go 100% (String + decidePattern). Cross-platform builds (darwin/linux/windows amd64) exit 0. `go vet` clean. C-HRA-008 grep: 0 matches. golangci-lint M1 files: 0 issues. |
| M2 | Pattern P4 (Handoff) + P3 (syscall.Exec) + Mutex | completed | 2026-05-23 | 2026-05-23 | `handoff_guidance.go` (printHandoff + printHandoffWithError) + `team_launch_posix.go` (launchP3 with injectable `syscallExecFn` + `lookPathFn`) + `team_launch_windows.go` (Windows fallback to handoff with notice). `MarkFlagsMutuallyExclusive("team", "tmux")` added to `newNewCmd`. `dispatchTeamLaunch` wired into `runNew` after worktree creation: P4 → printHandoff, P3 → launchP3 (POSIX exec / Windows fallback), P1/P2 → handoff with M3-pending info notice. TDD RED→GREEN: 4 handoff_guidance tests + 4 launchP3 POSIX tests + 1 Windows P3 fallback test + TestNewTeamTmuxMutex + extended TestNew_NoAskUserQuestion (scans 5 files). Coverage: printHandoff 87.5%, printHandoffWithError 100%, launchP3 90.9% (syscall.Exec line uncoverable by design per plan §R3). Cross-platform builds (darwin/linux/windows amd64) exit 0. `go vet` clean. C-HRA-008 grep: 0 matches. golangci-lint baseline 4 issues = 0 NEW issues. |
| M3 | Pattern P1/P2 (Tmux Window Spawn) | completed | 2026-05-23 | 2026-05-23 | `team_launch_posix.go` extended with `launchP1P2` (tmux new-window spawn returning pane_id) + injectable `tmuxNewWindowFn` package var + `defaultTmuxNewWindow` production binding (`tmux new-window -d -P -F '#{pane_id}' -c <cwd> '<cmd>'`). `team_launch_windows.go` extended with defensive `launchP1P2` stub returning error (tmux POSIX-only). `dispatchTeamLaunch` in `new.go` P1/P2 branch wired to `launchP1P2`: success → stdout `tmux window spawned in pane %s`; failure → `printHandoffWithError` with stderr substring `tmux pane spawn failed` (REQ-WTL-007 fallback, exit 0 since worktree was created). TDD RED→GREEN: 7 P1/P2 subtests (TestLaunchP1_CG_TmuxGLMWindow + TestLaunchP2_NoCG_TmuxCCWindow + TestLaunchP1P2_PaneSpawnFailure_FallbackHandoff + TestLaunchP1P2_PaneIDCaptured + TestLaunchP1P2_SettingsLocalJSON_ByteIdentical + TestLaunchP1P2_GLMDriftFallback_P2 + TestLaunchP1P2_EmptyLLMDefaultsToCC). All M1/M2 tests still PASS. Coverage `launchP1P2` 100%, `defaultTmuxNewWindow` 0% (real-tmux-only, matches M2 syscall.Exec uncoverable pattern). Cross-platform builds (darwin/linux/windows amd64) exit 0. `go vet` clean. C-HRA-008 grep: 0 matches. golangci-lint baseline 27 issues (pre-existing) = 0 NEW issues. |
| M4 | Swarm Registry + Failure Mode Wiring | completed | 2026-05-23 | 2026-05-23 | `swarm_registry.go` (`WriteSwarmEntry` + `patternToMode` + `SwarmEntry` 7-field schema) + dispatch wiring in `new.go` `dispatchTeamLaunch`: P1/P2 success → registry write with captured pane_id + mode `tmux-glm`/`tmux-cc`; P3 → registry write BEFORE `launchP3` (syscall.Exec replaces process, so write order matters) with pane_id="" + mode `in-progress-glm`/`in-progress-cc`; P4 → no registry (no spawn); P1/P2 pane spawn failure → no registry (REQ-WTL-007 fallback). TDD RED→GREEN: 8 swarm_registry subtests (P1_Schema_PaneIDPopulated + P2_Schema_PaneIDPopulated + P3_Schema_PaneIDEmpty {CC/GLM} + ParentDirAutoCreated + FilePerm_0o600 + JSONFields + MkdirFailure + WriteFailure) + 5 PatternToMode subtests + 4 dispatch-wiring subtests (DispatchTeamLaunch_P1_WritesRegistry + DispatchTeamLaunch_P1_PaneSpawnFails_NoRegistry + DispatchTeamLaunch_P2_WritesRegistry + DispatchTeamLaunch_P4_NoRegistry) + 1 worktree-create-failure test (TestDispatchTeamLaunch_WorktreeCreateFailure_NoRegistry) + extended TestNew_NoAskUserQuestion scan to swarm_registry.go (6 files). All M1/M2/M3 tests still PASS. Coverage swarm_registry.go: WriteSwarmEntry 90%, patternToMode 100%, swarmRegistryDir 100% (uncovered: json.MarshalIndent error path is structurally unreachable since SwarmEntry has no un-marshalable fields). Cross-platform builds (darwin/linux/windows amd64) exit 0. `go vet` clean. C-HRA-008 grep: 0 matches. golangci-lint internal/cli/worktree/: 0 issues (baseline 4 pre-M4 issues in internal/tmux/cg_detect.go + session_sensitive_test.go unchanged). |
| M5 | Skill Body + Rule Updates + Template Mirrors | completed | 2026-05-23 | 2026-05-23 | `.claude/skills/moai-workflow-worktree/SKILL.md` extended +68 LOC (276→344) with new `### 5. --team Flag - Contextual Session Launch` section after Section 4 (Integration Patterns). New section includes P1-P4 decision matrix as 6-column Markdown table (tmux session? / CG mode active? / --team flag? / Behavior / LLM), detection logic (3-condition CG mode definition + drift fallback to P2 per REQ-WTL-009), 4 example invocations (tmux+CG/tmux-only/no-tmux/no-flag), mutual exclusion with --tmux note, swarm registry baseline 7-field schema reference, failure modes (pane spawn fail → P4 fallback + Windows route → handoff). `.claude/rules/moai/workflow/worktree-integration.md` extended +25 LOC (368→393) with new `## Team Launch Patterns` section after `## Prompt Path Rules for Worktree-Isolated Agents`, before `## Minimum Version Requirements`. New rule section includes 2 ZONE:Frozen HARD rules (no AskUserQuestion per CONST-V3R5-030 + --team/--tmux mutex), static guard reference to TestNew_NoAskUserQuestion, swarm registry baseline contract, and 4-item Cross-references list. Template mirrors byte-identical via `cp` — both `diff` invocations exit 0. M1-M4 Go tests still PASS (no source code touched). Literal markers verified: `| P[1-4] |` 4 rows in SKILL.md table; SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001 referenced in both files; CONST-V3R5-030 referenced in both files. |
| M6 | Quality Verification + Cross-Platform Build + Coverage | not-started | — | — | Verification batch (10-command parallel) |

## Evidence Tracker

| ID | AC | Verification Command | Evidence (post-implementation) |
|----|----|---------------------|--------------------------------|
| AC-WTL-001 | P1 dispatch: tmux + CG → moai glm window | `go test -run TestLaunchP1_CG_TmuxGLMWindow ./internal/cli/worktree/` | **PASS (M3 2026-05-23)**: TestLaunchP1_CG_TmuxGLMWindow captures cwd = worktree-path AND command contains `moai glm` AND NOT `moai cc`. Pane_id `%5` correctly propagated. Sub-assertion (settings.local.json byte-identical pre/post) verified via TestLaunchP1P2_SettingsLocalJSON_ByteIdentical (SHA-256 unchanged after launch). |
| AC-WTL-002 | P2 dispatch: tmux + CC → moai cc window | `go test -run TestLaunchP2_NoCG_TmuxCCWindow ./internal/cli/worktree/` | **PASS (M3 2026-05-23)**: TestLaunchP2_NoCG_TmuxCCWindow captures cwd = worktree-path AND command contains `moai cc` AND NOT `moai glm`. Pane_id `%7` propagated. REQ-WTL-009 drift case verified by TestLaunchP1P2_GLMDriftFallback_P2 (given P2 + LLM=cc, launchP1P2 dispatches `moai cc` despite original CG-mode user intent). |
| AC-WTL-003 | P3 dispatch: no tmux → syscall.Exec | `go test -run TestLaunchP3 ./internal/cli/worktree/` | **PASS (M2 2026-05-23)**: 4 launchP3 POSIX tests PASS (CapturesArgvAndCwd_CC + CapturesArgvAndCwd_GLM + ChdirFails + LookPathFails). Injectable `syscallExecFn` captures argv [moai, cc] / [moai, glm] correctly; `lookPathFn` injection makes tests independent of $PATH state. cwd after exec asserted via symlink-resolved comparison. |
| AC-WTL-004 | P4 dispatch: --team absent → handoff guidance | `go test -run TestPrintHandoff ./internal/cli/worktree/` | **PASS (M2 2026-05-23)**: 4 printHandoff tests PASS (CC_Mode + GLM_Mode + PathWithSpaces + WithError fallback). stdout contains both `cd ` and `&& moai` literals per AC; CC vs GLM routes to correct LLM invocation; path with spaces preserved verbatim. |
| AC-WTL-005 | CG mode detection 4-scenario boolean | `go test -run TestIsCGMode ./internal/tmux/` | **PASS (M1 2026-05-23)**: 9 subtests PASS (4 AC scenarios + drift warning + nil-sink + base-URL + corrupt-JSON + no-file). Drift test asserts stderr substring `GLM env vars are absent`. |
| AC-WTL-006 | BODP HARD: TestNew_NoAskUserQuestion green | `go test -run TestNew_NoAskUserQuestion ./internal/cli/worktree/` | **PASS (M4 2026-05-23, extended)**: scans 6 files (new.go + team_launch.go + team_launch_posix.go + team_launch_windows.go + handoff_guidance.go + swarm_registry.go). Raw grep `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/worktree/ internal/tmux/` returns 0 non-comment matches. M4 extended scan to swarm_registry.go. |
| AC-WTL-007 | Pane spawn failure → P4 fallback | `go test -run TestLaunchP1P2_PaneSpawnFailure_FallbackHandoff ./internal/cli/worktree/` | **PASS (M3 2026-05-23 complete)**: TestLaunchP1P2_PaneSpawnFailure_FallbackHandoff injects `tmuxNewWindowFn` returning `tmux: no server running` error; launchP1P2 returns wrapped error containing `tmux` substring with empty pane_id. `dispatchTeamLaunch` in new.go detects this error and calls `printHandoffWithError(out, stderr, cfg, "tmux pane spawn failed: ...")` then returns nil (exit 0 — worktree was created OK). The "tmux pane spawn failed" substring is the verification anchor for AC-WTL-007 stderr assertion. Edge-case mutex path (TestNewTeamTmuxMutex) also still PASS. |
| AC-WTL-008 | Swarm registry schema (7 fields + 0o600) | `go test -run TestSwarmRegistry_P1_Schema ./internal/cli/worktree/` | **PASS (M4 2026-05-23)**: TestSwarmRegistry_P1_Schema_PaneIDPopulated verifies all 7 fields (spec_id / worktree_path / branch / pane_id / mode / created_at / created_by_pid) round-trip identical via JSON unmarshal. TestSwarmRegistry_JSONFields asserts exact JSON keys with map-based unmarshal (no extra keys). TestSwarmRegistry_FilePerm_0o600 asserts `info.Mode().Perm() == 0o600` via os.Stat (POSIX-only; Windows skipped per file-mode model). TestSwarmRegistry_ParentDirAutoCreated verifies `.moai/state/swarm/` auto-creation via os.MkdirAll(0o755). TestSwarmRegistry_P2_Schema_PaneIDPopulated covers tmux-cc mode. TestSwarmRegistry_P3_Schema_PaneIDEmpty covers in-progress-cc and in-progress-glm modes with empty pane_id. Dispatch integration verified via TestDispatchTeamLaunch_P1_WritesRegistry + TestDispatchTeamLaunch_P2_WritesRegistry end-to-end with injected tmuxNewWindowFn. |
| AC-WTL-009 | Cross-platform builds darwin/linux/windows | `GOOS={darwin,linux,windows} GOARCH=amd64 go build ./...` | **PASS (M4 2026-05-23)**: All 3 OS targets exit 0. swarm_registry.go has no build tag (portable Go using encoding/json + os + time stdlib). Verified in single-turn parallel batch alongside test suite. |
| AC-WTL-010 | Invalid SPEC → no launch, no registry | `go test -run TestDispatchTeamLaunch_WorktreeCreateFailure_NoRegistry ./internal/cli/worktree/` | **PASS (M4 2026-05-23)**: TestDispatchTeamLaunch_WorktreeCreateFailure_NoRegistry injects a failing WorktreeProvider whose `Add` returns os.ErrPermission. runNew returns the wrapped error ("create worktree: permission denied") BEFORE dispatchTeamLaunch is invoked, so no registry write is reachable. Negative assertion `os.Stat(registry-path)` confirms file does NOT exist. Structural guarantee: line 159 of new.go returns early on WorktreeProvider.Add failure, structurally preventing both launch and registry write. Additional negative assertions in TestDispatchTeamLaunch_P1_PaneSpawnFails_NoRegistry + TestDispatchTeamLaunch_P4_NoRegistry. |
| AC-WTL-011 | Lint baseline: 0 NEW violations | `golangci-lint run --timeout=2m` (diff vs baseline) | _pending_ |
| AC-WTL-012 | Test coverage ≥85% for new files | `go tool cover -func=cover.out` per-file threshold | _pending_ |
| AC-WTL-013 | Skill body + template mirror byte-identical | `diff -r .claude/skills/moai-workflow-worktree/ internal/template/templates/.claude/skills/moai-workflow-worktree/` | **PASS (M5 2026-05-23)**: `diff .claude/skills/moai-workflow-worktree/SKILL.md internal/template/templates/.claude/skills/moai-workflow-worktree/SKILL.md` → exit 0 (zero output). `diff .claude/rules/moai/workflow/worktree-integration.md internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` → exit 0 (zero output). Both pairs byte-identical via `cp` mirror operation. Pre-M5 baseline already had zero drift; M5 preserved invariant. REQ-WTL-011 (Template-First mirror discipline) satisfied. |
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
| R4: glm.SettingsLocal import cycle | resolved (M1 2026-05-23) | Local `settingsLocalMin` struct in `internal/tmux/cg_detect.go`; verified no `internal/cli/*` import (build clean cross-platform) |
| R5: tmux new-window vs new-session confusion | resolved | OQ-1 resolved: new-window |
| R6: Orphaned worktree on pane spawn failure | accepted | Exit 0 + P4 handoff guidance gives user manual recovery |
| R7: settings.local.json accidental mutation | open | No os.WriteFile on settings.local.json in team_launch.go |
| R8: BODP HARD violation via AskUserQuestion | open | Static guard extended in AC-WTL-006 |
| R9: --team + --tmux combination undefined | resolved | OQ-2 resolved: mutually exclusive |
| R10: Swarm registry orphans on Ctrl-C | resolved (M4 2026-05-23) | P1/P2: registry write AFTER `launchP1P2` success (Ctrl-C during tmux spawn → no write). P3: registry write BEFORE `launchP3` syscall.Exec (Ctrl-C during write → no write; Ctrl-C after write but before exec replacement → orphan accepted, no rollback since worktree already exists). P4 + worktree-create-failure: structurally no registry write. Negative tests cover P1/P2 pane-fail + P4 + worktree-create-failure paths. |

## Plan-Phase Self-Check Result

Performed internal self-check per orchestrator instructions:

- [x] Every REQ has at least one mapped AC (14 ACs across 13 REQs; REQ-WTL-000 is pre-flight gate)
- [x] Every AC verifies a specific REQ or HARD constraint
- [x] Traceability matrix is complete (acceptance.md §3)
- [x] Out of Scope is explicit (spec.md §4.1 + plan.md OQ resolutions)
- [x] HARD constraints listed (spec.md §5 — 8 constraints)
- [x] Tier classification: M (5-15 files affected, 10 distinct files in inventory, ~640-740 source LOC + ~580 test LOC; stays below Tier L >1000 source LOC threshold for non-constitutional changes)
- [x] Frontmatter 12-field canonical schema applied (id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags, plus optional tier)
- [x] No file modifications outside `.moai/specs/SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001/`
- [x] No implementation code touched (plan-phase only; M1-M6 deferred to run-phase)
- [x] BODP HARD compliance: REQ-WTL-013 + AC-WTL-006 enforce CLI no-AskUserQuestion rule (CONST-V3R5-030)
- [x] Template-First Rule: REQ-WTL-011 + AC-WTL-013 enforce byte-identical mirror in `internal/template/templates/`

## Notes

- This SPEC ships independently of Sprint 2 in-progress SPECs (HOOK-OBSERVE-OPT-IN, HOOK-ASYNC-EXPAND, AGENT-MODEL-ROUTING, PROMPT-CACHE). Can be queued anytime in the Sprint 2 sequence.
- Baseline references verified during artifact drafting: `internal/cli/worktree/new.go` (346 LOC), `internal/tmux/detector.go` (119 LOC), `internal/cli/launcher.go` `applyCGMode` (L143-204), `internal/cli/worktree/tmux_integration.go` `CreateTmuxSession` (L48-94), `internal/cli/glm.go` `SettingsLocal.TeammateMode` (L103-108), `TestNew_NoAskUserQuestion` (L732-742).
- After plan-PR merge, run-phase invocation: `/moai run SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001` (Tier M → Section A-E delegation template REQUIRED per `manager-develop-prompt-template.md` § Applicability).
