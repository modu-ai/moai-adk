---
id: SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001-plan
title: "Implementation plan — Worktree --team contextual launch"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/cli/worktree, internal/tmux"
lifecycle: spec-anchored
tags: "plan, worktree, team-launch, milestones, tier-m, sprint-2"
tier: M
---

# Plan — SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001

## D. Milestones

Six milestones M1-M6 in priority order. No time estimates (per `agent-common-protocol.md` § Time Estimation).

### M1: CG Mode Detector + Pattern Dispatcher Scaffolding (Priority: High)

**Scope**: Add CG mode detection helper to `internal/tmux/` and pattern dispatcher skeleton to `internal/cli/worktree/team_launch.go`. No actual launch yet — only the decision matrix logic + struct definitions.

**Deliverables**:
- `internal/tmux/cg_detect.go` (NEW, ~50 LOC) — `IsCGMode(settingsPath string) (bool, error)` exported helper. Reads `.claude/settings.local.json` `teammateMode` field, checks `InTmuxSession()`, scans tmux session env for `ANTHROPIC_AUTH_TOKEN` / `ANTHROPIC_BASE_URL`. Returns `(false, nil)` on missing file (graceful, not error). Uses `glm.SettingsLocal` JSON unmarshalling pattern (read-only, no Write).
- `internal/tmux/cg_detect_test.go` (NEW, ~120 LOC) — 5 subtests: in tmux + teammateMode=tmux + GLM env → true / in tmux + teammateMode=tmux + no GLM env → false (P2 fallback) / no tmux → false / no settings.local.json → false / settings parse error → (false, error).
- `internal/cli/worktree/team_launch.go` (NEW, ~80 LOC scaffolding) — `Pattern` enum constants (P1/P2/P3/P4), `decidePattern(teamFlag bool, inTmux bool, cgMode bool) Pattern` pure function, `TeamLaunchConfig` struct with `Pattern`, `WorktreePath`, `Branch`, `SpecID`, `LLM` ("glm"|"cc"), `LaunchTime`. No execution yet.
- `internal/cli/worktree/new.go` (Modified, ~5 LOC added) — Add `--team` bool flag to `newNewCmd` `Flags()` definition (no dispatch logic yet — dispatch wiring lands in M2). Reason: the flag must exist as a cobra surface before any pattern logic so M3's tmux tests can invoke the command structure consistently. This unblocks M2's mutex enforcement check (depends on flag existing).

**Risks**: glm.SettingsLocal struct import cycle if `cg_detect.go` lives in `internal/tmux/` and imports `internal/cli/`. Mitigation: define a local minimal struct `type settingsLocalMin struct { TeammateMode string }` in `cg_detect.go` to avoid cross-package import.

### M2: Pattern P4 (No-Flag Handoff Guidance) + P3 (No-Tmux syscall.Exec) (Priority: High)

**Scope**: Implement the simplest two patterns first — P4 (stdout-only) and P3 (syscall.Exec). These are the foundation tests for the dispatcher.

**Deliverables**:
- `internal/cli/worktree/handoff_guidance.go` (NEW, ~60 LOC) — `printHandoff(out io.Writer, cfg *TeamLaunchConfig)` formats and writes paste-ready guidance:

  ```
  Worktree created at: <worktree-path>
  To start a session, run:
      cd <worktree-path> && moai cc
  ```

  Or `moai glm` when `cfg.LLM == "glm"`.

- `internal/cli/worktree/handoff_guidance_test.go` (NEW, ~80 LOC) — golden test with 3 cases: CC mode handoff / GLM mode handoff (CG detected but no `--team`) / handoff after tmux spawn failure (with stderr error notice prepended).
- Extend `internal/cli/worktree/team_launch.go` (~100 LOC added):
  - `launchP3(cfg *TeamLaunchConfig) error` — POSIX-only via `//go:build !windows`. Uses `os.Chdir(cfg.WorktreePath)` then `syscall.Exec(moaiBin, []string{"moai", cfg.LLM}, os.Environ())`. Returns error only on Chdir or LookPath failure (syscall.Exec doesn't return on success).
  - `team_launch_windows.go` (NEW, ~30 LOC, `//go:build windows`) — `launchP3Windows(cfg) error` falls back to `exec.Command` + `os.Exit`, prints unsupported notice via REQ-WTL-012.
  - Test scaffolding using injectable `syscallExecFn` var.
- Wire P3/P4 dispatch into `runNew` (`new.go`) — call `decidePattern` after worktree creation success, then `printHandoff` or `launchP3` per result.
- Add `MarkFlagsMutuallyExclusive("team", "tmux")` to `newNewCmd` Flags() setup to enforce R9 + OQ-2 (these two flags express conflicting launch intents). Verification: AC-WTL-007 extended with negative test that `--team --tmux` exits non-zero with cobra mutual exclusion error.

**Risks**: syscall.Exec replaces process — testing requires a fake `syscallExecFn` var. Pattern is established in `launcher.go` L541. Tests verify argv + cwd captured before exec; no real exec executed.

### M3: Pattern P1/P2 (Tmux Window Spawn) Implementation (Priority: High)

**Scope**: Implement P1 (CG mode → moai glm) and P2 (CC mode → moai cc) tmux window spawn. Reuse `CreateTmuxSession` from `tmux_integration.go` — adapt for "new window" instead of "new session" semantics.

**Deliverables**:
- Extend `internal/cli/worktree/team_launch.go` (~150 LOC added):
  - `launchP1P2(cfg *TeamLaunchConfig, mgr tmux.SessionManager) (paneID string, err error)` — Decision: spawn a new tmux window in the CURRENT tmux session (where `moai worktree new --team` was invoked). Uses `tmux new-window -d -P -F '#{pane_id}' -c <worktree-path> 'moai <cfg.LLM>'`. Returns the pane ID for swarm registry.
  - Sub-OQ resolved: `--team` does NOT detach a brand-new tmux session (P1/P2 stay in the user's current tmux session as a new window). This matches the CG-mode mental model where the leader pane stays where it is, and teammates spawn as new windows in the same session.
- `internal/cli/worktree/team_launch_test.go` updated (~150 LOC added) — 6 subtests for P1/P2: CG mode + tmux → P1 `moai glm` window / CC mode + tmux → P2 `moai cc` window / pane spawn failure → fallback to P4 with stderr notice / settings.local.json byte-identical pre/post `--team` / tmux env GLM token missing → warning + P2 fallback (REQ-WTL-009) / `tmux new-window -P` pane ID captured.

**Risks**: `tmux new-window` vs `tmux new-session` semantics — P1/P2 are inside-tmux operations, so `new-window` is correct. The existing `CreateTmuxSession` in `tmux_integration.go` creates a detached session — NOT what we want for `--team` which assumes the user is already in a tmux session.

### M4: Swarm Registry + Failure Mode Wiring (Priority: Medium)

**Scope**: Implement `.moai/state/swarm/<SPEC-ID>.json` write after successful P1/P2/P3 launch. Wire REQ-WTL-007 (pane spawn failure → fallback) and REQ-WTL-010 (worktree creation failure → no registry write).

**Deliverables**:
- `internal/cli/worktree/swarm_registry.go` (NEW, ~80 LOC) — `WriteSwarmEntry(repoRoot string, entry SwarmEntry) error` writes JSON to `.moai/state/swarm/<spec-id>.json` with mode `0o600`. `SwarmEntry` struct matches REQ-WTL-008 schema (spec_id / worktree_path / branch / pane_id / mode / created_at / created_by_pid).
- `internal/cli/worktree/swarm_registry_test.go` (NEW, ~100 LOC) — 4 subtests: P1 entry has mode=tmux-glm + pane_id populated / P2 entry has mode=tmux-cc + pane_id populated / P3 entry has mode=in-progress-* + pane_id empty / parent dir auto-created (mkdir 0o755).
- Wire failure modes in `runNew` (new.go ~40 LOC added):
  - Worktree creation fails (existing `WorktreeProvider.Add` error path) → return error, no registry, no launch.
  - Pane spawn fails (P1/P2) → call `printHandoff` with error notice + warning on stderr, exit 0, NO registry write.

**Risks**: Registry write between worktree creation and team launch — if the user Ctrl-C's during launch, registry may be orphaned. Mitigation: registry is written AFTER successful launch dispatch, NOT before. For P1/P2, the pane_id is the success signal. For P3, registry is written immediately before syscall.Exec.

### M5: Skill Body + Rule Updates + Template Mirrors (Priority: Medium)

**Scope**: Document `--team` flag in skill body and worktree-integration rule. Mirror to template (REQ-WTL-011).

**Deliverables**:
- `.claude/skills/moai-workflow-worktree/SKILL.md` updated (~50 LOC added) — new §`--team` Flag section with P1/P2/P3/P4 decision matrix as a Markdown table. Include the example command flow per pattern.
- `.claude/rules/moai/workflow/worktree-integration.md` updated (~30 LOC added) — new §Team Launch Patterns section cross-referencing this SPEC and the decision matrix.
- `internal/template/templates/.claude/skills/moai-workflow-worktree/SKILL.md` — byte-identical mirror (`diff` exit 0).
- `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` — byte-identical mirror.

**Risks**: Skill body length — `SKILL.md` currently 276 lines (verified via `wc -l`). Adding 50 LOC keeps it well under the 500-line skill body cap. Template-First Rule discipline per `CLAUDE.local.md §2` enforced via diff verification.

### M6: Quality Verification + Cross-Platform Build + Coverage (Priority: Medium)

**Scope**: Verification batch — run full quality gate, ensure cross-platform builds, coverage ≥85% for new files, lint zero NEW.

**Deliverables**:
- Verification batch executed as single-turn multi-Bash call per `agent-common-protocol.md` § Parallel Execution:
  1. `go test ./internal/cli/worktree/... ./internal/tmux/...` → all pass
  2. `go test -coverprofile=cover.out ./internal/cli/worktree/... ./internal/tmux/...` → coverage ≥85% for new files (`team_launch.go`, `handoff_guidance.go`, `swarm_registry.go`, `cg_detect.go`)
  3. `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/worktree/ | grep -v "_test.go" | grep -v "// "` → zero matches (BODP HARD)
  4. `go vet ./...` → zero issues
  5. `golangci-lint run --timeout=2m` → zero NEW violations (compared to `git merge-base HEAD origin/main`)
  6. `GOOS=darwin GOARCH=amd64 go build ./...` → exit 0
  7. `GOOS=linux GOARCH=amd64 go build ./...` → exit 0
  8. `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (Windows P3 simulation via `team_launch_windows.go`)
  9. `diff -r .claude/skills/moai-workflow-worktree/ internal/template/templates/.claude/skills/moai-workflow-worktree/` → zero output (Template-First mirror)
  10. `diff .claude/rules/moai/workflow/worktree-integration.md internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` → zero output
- Commit + push to feat branch with Conventional Commit `feat(SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001): contextual --team flag for worktree new` + `🗿 MoAI` trailer.

**Risks**: Coverage threshold — `team_launch.go` has POSIX syscall.Exec which cannot run in tests, leaving ~10% uncovered. Mitigation: `syscallExecFn` injectable var pattern (see `launcher.go` `launchClaudeFunc`). On Windows builds, the test gates the launch path via `//go:build` tag.

## File Inventory (LOC Estimates)

| File | Status | LOC Est. | Purpose |
|------|--------|----------|---------|
| `internal/cli/worktree/new.go` | Modified | +50 LOC | `--team` flag wiring in `newNewCmd` (M1) + dispatch in `runNew` + `MarkFlagsMutuallyExclusive("team", "tmux")` (M2) |
| `internal/cli/worktree/new_test.go` | Modified | +20 LOC | Extend `TestNew_NoAskUserQuestion` to scan `team_launch.go`, `handoff_guidance.go`, `swarm_registry.go`, `team_launch_windows.go`; add `TestNewTeamTmuxMutex` negative test for AC-WTL-007 extension |
| `internal/cli/worktree/team_launch.go` | NEW | ~330 LOC | Pattern dispatcher + P1/P2/P3 implementations |
| `internal/cli/worktree/team_launch_windows.go` | NEW | ~30 LOC | Windows fallback for P3 (`//go:build windows`) |
| `internal/cli/worktree/team_launch_test.go` | NEW | ~280 LOC | Full coverage 4 patterns + edges |
| `internal/cli/worktree/handoff_guidance.go` | NEW | ~60 LOC | P4 stdout formatter |
| `internal/cli/worktree/handoff_guidance_test.go` | NEW | ~80 LOC | Golden test |
| `internal/cli/worktree/swarm_registry.go` | NEW | ~80 LOC | `.moai/state/swarm/<SPEC>.json` writer |
| `internal/cli/worktree/swarm_registry_test.go` | NEW | ~100 LOC | Registry schema + permission test |
| `internal/tmux/cg_detect.go` | NEW | ~50 LOC | `IsCGMode` helper |
| `internal/tmux/cg_detect_test.go` | NEW | ~120 LOC | 5 subtests |
| `.claude/skills/moai-workflow-worktree/SKILL.md` | Modified | +50 LOC | §`--team` Flag + P1-P4 matrix |
| `.claude/rules/moai/workflow/worktree-integration.md` | Modified | +30 LOC | §Team Launch Patterns |
| `internal/template/templates/.claude/skills/moai-workflow-worktree/SKILL.md` | Modified | +50 LOC | byte-identical mirror |
| `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` | Modified | +30 LOC | byte-identical mirror |

**Totals**: 11 distinct logical files (5 modified — including `new_test.go` extension required by AC-WTL-006 + AC-WTL-007 mutex test; 6 NEW source + 4 NEW test = 10 NEW files; total 15 file-touches counting both source and template mirrors). Tier M (5-15 files) — sits at the upper boundary. If any further file is added during implementation, Tier reclassification to L is required. LOC ~640-740 actual source + ~600 test = ~1320 lines total. Tier reconciliation: stays Tier M (Tier L LOC threshold is >1000 LOC for source code OR constitutional changes — this is feature CLI work, no constitutional touch).

## Dependency Notes

- **No dependency on Sprint 2 in-progress SPECs**: this SPEC is independent of HOOK-OBSERVE-OPT-IN, HOOK-ASYNC-EXPAND, AGENT-MODEL-ROUTING, PROMPT-CACHE. Can ship out-of-order.
- **References existing infra**: SPEC-WORKTREE-001 (worktree creation), SPEC-WORKTREE-002 (tmux integration), SPEC-V3R3-CI-AUTONOMY-001 W7 (BODP audit trail). All three remain untouched.
- **Future SPEC enabler**: `moai swarm status/done/kill-all` reads from `.moai/state/swarm/<SPEC>.json` registry baseline written by this SPEC.

## Risk Register

| Risk | Severity | Mitigation |
|------|----------|------------|
| R1: `syscall.Exec` POSIX-only — tests cannot run real exec | High | Injectable `syscallExecFn` var pattern (proven in `launcher.go`); tests verify argv + cwd captured |
| R2: Windows build fail due to `syscall.Exec` import | High | `//go:build !windows` tag split into `team_launch.go` + `team_launch_windows.go` |
| R3: Coverage <85% due to syscall.Exec uncoverable lines | Medium | Per-package threshold acceptance — uncoverable lines documented with `//nolint:unused` and `// uncovered: syscall.Exec replaces process` comment |
| R4: glm.SettingsLocal import cycle if `cg_detect.go` in `internal/tmux/` | Medium | Define local minimal struct `settingsLocalMin{TeammateMode string}` in `cg_detect.go` to avoid `internal/cli/` import |
| R5: tmux new-window vs new-session semantics confusion | Low | `--team` uses `tmux new-window` (inside existing session); SPEC-WORKTREE-002 `--tmux` uses `tmux new-session` (detached). Both flags can coexist; `--tmux --team` combination errors |
| R6: Pane spawn failure between worktree-create and team-launch leaves orphaned worktree | Low | Worktree is valid state; user can manually `cd <wt> && moai cc`. P4 fallback handoff prints the exact command. Exit code 0 (worktree created OK). |
| R7: Settings.local.json mutation by accidental write — violates HARD constraint | High | All reads via `os.ReadFile`; no `os.WriteFile` on settings.local.json in team_launch.go. Test: AC-WTL-005 byte-identical pre/post check |
| R8: BODP HARD violation if `--team` adds AskUserQuestion | Critical | Static guard `TestNew_NoAskUserQuestion` extended to scan team_launch.go + handoff_guidance.go for "AskUserQuestion" substring |
| R9: `--team` + `--tmux` combination semantics undefined | Low | Mutual exclusion error on flag combination (cobra `MarkFlagsMutuallyExclusive`) |
| R10: Swarm registry orphans on Ctrl-C between worktree creation and registry write | Low | Registry write happens only AFTER successful launch dispatch. P3 syscall.Exec: write immediately before exec. P1/P2: write only on pane_id capture success. P4 (no --team): no registry write. |

## Technical Approach Summary

The implementation strategy follows the proven pattern from `internal/cli/launcher.go`:

1. **Pure decision function**: `decidePattern(teamFlag, inTmux, cgMode) Pattern` — single source of truth for the 4-way dispatch, table-testable.
2. **Injectable execution**: `syscallExecFn`, `tmuxNewWindowFn` vars override the real syscall/exec calls in tests.
3. **Static BODP guard**: `TestNew_NoAskUserQuestion` reads the source file as bytes and asserts "AskUserQuestion" substring is absent. Extended to cover new files.
4. **Byte-identical template mirroring**: each rule/skill edit is followed by a `cp` to `internal/template/templates/`, verified by `diff` returning empty.

The integration point with `runNew` (new.go) is a single new code block AFTER the worktree creation succeeds (post-L172 `wtSuccessCard` print). The block is conditionally entered only when `--team` is set, and falls back to P4 handoff guidance silently when `--team` is absent or pane spawn fails.

## OQ (Open Questions) — Resolved in Spec

| OQ | Resolution |
|----|------------|
| OQ-1: `--team` opens new tmux session vs new window in current session? | **New window in current session** (P1/P2). Matches CG-mode mental model. |
| OQ-2: `--team --tmux` combination behavior? | **Error — mutually exclusive flags**. The two flags target different use cases. |
| OQ-3: `--team` on Windows? | **Routes to P3 simulation** (REQ-WTL-012). Stderr notice that tmux unsupported. |
| OQ-4: Swarm registry pre-existing entry (re-creating worktree for same SPEC)? | **Overwrite silently**. The new launch replaces the old session reference. Future `moai swarm` may add overwrite warning. |
| OQ-5: P3 LLM choice when CG env partially set (e.g., teammateMode=tmux but no GLM token)? | **REQ-WTL-009 — warn + fallback to P2 `moai cc`**. Documented drift case. |

## Self-Check Result

Performed internal self-check per orchestrator instructions. Every REQ maps to ≥1 AC; every AC verifies a specific REQ. Traceability matrix in acceptance.md §3. Out of Scope section explicit in spec.md §4.1. Frontmatter 12-field canonical schema applied. No file modifications outside `.moai/specs/SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001/`.
