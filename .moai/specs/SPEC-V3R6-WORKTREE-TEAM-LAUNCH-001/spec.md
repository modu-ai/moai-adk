---
id: SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001
title: "Worktree --team flag: contextual Claude/GLM session launch in new worktree"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/cli/worktree, internal/tmux"
lifecycle: spec-anchored
tags: "worktree, team-launch, tmux, glm, cg-mode, syscall-exec, cli, bodp, v3.0, sprint-2"
tier: M
issue_number: null
depends_on: []
related_specs: [SPEC-WORKTREE-001, SPEC-WORKTREE-002, SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001]
---

# SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001: Worktree `--team` Contextual Session Launch

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-23 | manager-spec | Initial draft — Sprint 2 SPEC. Extend `moai worktree new <SPEC-ID>` with an optional `--team` flag that contextually launches a Claude or GLM session in the freshly created worktree depending on (1) tmux availability and (2) CG mode active state. Tier M (~600-900 LOC, 9-10 files in-scope: cobra flag wiring + pattern dispatcher + tmux helper + handoff formatter + skill/rule body + 2 template mirrors + tests). 4 canonical patterns (P1 tmux+cg → moai glm window, P2 tmux+cc → moai cc window, P3 no-tmux → syscall.Exec, P4 no-flag → handoff guidance). Swarm registry baseline (`.moai/state/swarm/<SPEC>.json`) for future `moai swarm status/done/kill-all`. CLI MUST NOT invoke AskUserQuestion (BODP HARD per `branch-origin-protocol.md` § HARD Rules). |

## 1. Goal

Extend `moai worktree new <SPEC-ID>` with an optional `--team` flag that launches a new Claude (or GLM) session in the freshly created worktree, **contextually choosing between four canonical patterns** based on the current tmux availability and CG mode state. The launch decision is made entirely from observable environment + settings.local.json state — the CLI never prompts the user (orchestrator-only HARD per BODP).

This is a usability improvement on top of SPEC-WORKTREE-002 (tmux integration). Where SPEC-WORKTREE-002 added the `--tmux` flag that creates a detached tmux session for the worktree, `--team` goes one step further by **immediately launching the appropriate LLM session in the new worktree** with cwd already anchored to the worktree path.

## 2. Why

Sprint 2 V3R6 multi-SPEC parallel workflow (memory entries: `V3R6 Sprint 2 4 SPECs plan complete`, `V3R6 Sprint 2 HOI blocked`) creates a new SPEC worktree several times per session via `moai worktree new <SPEC-ID>`. After worktree creation, the user manually:

1. `cd ~/.moai/worktrees/<project>/<SPEC-ID>`
2. `moai cc` (when CC mode) OR `moai glm` (when CG/GLM mode)

This 2-step pattern is mechanical and ergonomically painful, particularly when the user is in tmux + CG mode and wants the new session as a **separate tmux window** (not replacing the current pane). The `--team` flag collapses both steps into a single CLI invocation with cwd anchoring + LLM selection automated.

Beyond UX, the swarm registry baseline (`.moai/state/swarm/<SPEC>.json`) opens the door to a future `moai swarm` command family for orchestrating multiple SPEC sessions: `moai swarm status` to enumerate active workspaces, `moai swarm done <SPEC>` to mark completion, `moai swarm kill-all` for cleanup. This SPEC writes the baseline registry without implementing the swarm CLI (out of scope).

**Why CLI-only (not orchestrator)**: the launch decision is fully deterministic from observable state — there is no ambiguity that requires user interaction. Per `branch-origin-protocol.md` § HARD Rules, the CLI path MUST NOT invoke AskUserQuestion. BODP-style audit trail is already written by `writeWorktreeAuditTrail` and remains untouched.

## 3. Requirements (EARS)

### A. Pre-existing State Survey (REQ-WTL-000)

REQ-WTL-000 [Ubiquitous]: The CLI MUST grep-verify the following pre-existing infrastructure during pre-flight (Section C) before any new code is added:

- `internal/cli/worktree/new.go` (current: 346 LOC, BODP gate at L107-167, tmux gate at L175-195) — extension surface, the `--team` flag wires here
- `internal/tmux/detector.go` (current: 119 LOC, `SystemDetector` + `InTmuxSession()` at L66-68) — CG mode detection extension surface
- `internal/cli/launcher.go` `applyCGMode` (L143-204) — proves the CG mode pattern (tmux session + GLM env + settings.local.json TeammateMode=tmux). Reference reads only; `--team` does NOT mutate settings.local.json
- `internal/cli/worktree/tmux_integration.go` `CreateTmuxSession` (L48-94) + `buildTmuxInitialCommand` (L100-112) — proves pattern P1/P2 (detached tmux session with `cd <wt> && moai cc|glm`). The `--team` flag reuses `CreateTmuxSession` for P1/P2 with one difference: `--team` may attach the new session instead of leaving it detached. Sub-decision OQ in plan.md
- `internal/cli/glm.go` `SettingsLocal.TeammateMode` (L103-108) — read-only access for CG mode detection
- `internal/cli/worktree/new_test.go` `TestNew_NoAskUserQuestion` (L732-742) — static guard ensures BODP HARD compliance; the `--team` extension MUST keep this guard green

### B. Decision Matrix (REQ-WTL-001 to REQ-WTL-005)

REQ-WTL-001 [Event-Driven]: When `--team` is set AND the CLI process is inside a tmux session AND CG mode is active, the CLI SHALL spawn a new tmux window with cwd=worktree-path and command `moai glm`. (Pattern P1)

REQ-WTL-002 [Event-Driven]: When `--team` is set AND the CLI process is inside a tmux session AND CG mode is NOT active, the CLI SHALL spawn a new tmux window with cwd=worktree-path and command `moai cc`. (Pattern P2)

REQ-WTL-003 [Event-Driven]: When `--team` is set AND no tmux session is detected, the CLI SHALL invoke `syscall.Exec("moai", "glm" or "cc")` with the process's cwd switched to the worktree path. LLM choice mirrors CG mode detection (`moai glm` if CG, otherwise `moai cc`). The current shell process is replaced and the orchestrator session terminates. (Pattern P3)

REQ-WTL-004 [Ubiquitous]: When `--team` is absent, the CLI SHALL print handoff guidance to stdout containing a paste-ready `cd <worktree-path> && moai cc` (or `moai glm` when CG detected) line. No process spawn, no tmux window, no syscall.Exec. (Pattern P4)

REQ-WTL-005 [State-Driven]: While CG mode is active (detection per §4 below), team launch defaults to `moai glm` regardless of which pane invoked the command. The leader/teammate role is determined by tmux session env, not by the launching pane.

### C. Detection Logic (REQ-WTL-006, REQ-WTL-009)

REQ-WTL-006 [State-Driven]: CG mode detection SHALL be a read-only operation against:

- `tmux.NewDetector().InTmuxSession()` — true when `$TMUX` env var is non-empty
- `.claude/settings.local.json` `teammateMode == "tmux"` (per `glm.go` L228)
- Tmux session env contains GLM credentials: at least one of `ANTHROPIC_AUTH_TOKEN` (Z.AI token) OR `ANTHROPIC_BASE_URL` pointing to `*.z.ai`

CG mode is true iff: `InTmuxSession() == true` AND `teammateMode == "tmux"` AND (GLM token OR GLM base URL is set).

REQ-WTL-009 [Unwanted]: If `teammateMode == "tmux"` is set but GLM env vars are absent, then the CLI SHALL warn on stderr and treat the situation as P2 (`moai cc` fallback). This is the documented `settings.local.json` drift case — the user previously ran `moai cg` but the tmux session was restarted or env was cleared.

### D. Safety and Invariants (REQ-WTL-007, REQ-WTL-008, REQ-WTL-010 to REQ-WTL-013)

REQ-WTL-013 [Ubiquitous]: The CLI MUST NOT invoke AskUserQuestion. Static check: `TestNew_NoAskUserQuestion` (existing) is extended to cover `team_launch.go` and `handoff_guidance.go`. Traces to CONST-V3R5-030 (BODP HARD — CLI orchestrator-only AskUserQuestion).

REQ-WTL-007 [Event-Driven]: When tmux pane spawn fails (P1/P2), the CLI SHALL fall back to P4 (handoff guidance) with an error notice on stderr. The exit code MUST be 0 (worktree was created successfully; only the team launch convenience step failed). Worktree state remains valid.

REQ-WTL-008 [Ubiquitous]: After successful team launch (P1, P2, or P3), the CLI SHALL write `.moai/state/swarm/<SPEC-ID>.json` with the following exact fields:

```json
{
  "spec_id": "SPEC-V3R6-EXAMPLE-001",
  "worktree_path": "/Users/.../SPEC-V3R6-EXAMPLE-001",
  "branch": "feature/SPEC-V3R6-EXAMPLE-001",
  "pane_id": "%5",
  "mode": "tmux-glm",
  "created_at": "2026-05-23T14:30:00Z",
  "created_by_pid": 12345
}
```

- `pane_id` is the tmux pane identifier returned by `tmux split-window -P -F '#{pane_id}'` — populated only for P1/P2; empty string for P3
- `mode` enum: `"tmux-glm"` (P1) | `"tmux-cc"` (P2) | `"in-progress-glm"` (P3 with GLM) | `"in-progress-cc"` (P3 with Claude)
- `created_by_pid` is the launching `moai worktree new` PID (not the new session's PID, which is not knowable for P1/P2 from the parent)
- File MUST be written with mode `0o600` (per project security convention for state files)

REQ-WTL-010 [Event-Driven]: When worktree creation fails (any reason — invalid SPEC ID, git error, path exists), no team launch is attempted and no swarm registry file is written. The original error is returned unchanged.

REQ-WTL-011 [Ubiquitous]: Template-First Rule per `CLAUDE.local.md §2` MUST be honored: any rule/skill update gets a byte-identical mirror in `internal/template/templates/`. Verification: `diff -r .claude/skills/moai-workflow-worktree/ internal/template/templates/.claude/skills/moai-workflow-worktree/` and `diff .claude/rules/moai/workflow/worktree-integration.md internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` MUST return zero output.

REQ-WTL-012 [Optional]: Where the user runs on Windows (no tmux), `--team` automatically routes to P3 (syscall.Exec equivalent) with a stderr notice that tmux is unsupported on Windows. Implementation detail: `syscall.Exec` is POSIX-specific; on Windows the implementation uses `exec.Command(...).Run()` followed by `os.Exit(cmd.ProcessState.ExitCode())` to simulate process replacement.

## 4. Acceptance Criteria Summary

12 binary ACs across 5 categories (full Given/When/Then in `acceptance.md`):

1. AC-WTL-001 through AC-WTL-005: pattern dispatch (P1, P2, P3, P4, CG detection)
2. AC-WTL-006 through AC-WTL-008: safety (BODP HARD guard, pane spawn failure, swarm registry)
3. AC-WTL-009 through AC-WTL-010: cross-platform + error handling
4. AC-WTL-011 through AC-WTL-012: quality (lint baseline, coverage ≥85%)
5. AC-WTL-013 through AC-WTL-014: documentation + audit trail preservation

Every AC has a single deterministic verification command. See `acceptance.md` § Traceability Matrix.

### 4.1 Out of Scope

- `moai swarm status` / `moai swarm done` / `moai swarm kill-all` CLI commands — registry baseline ONLY in this SPEC; future SPEC
- Attaching to an existing tmux session vs creating a new one — `--team` always spawns a new window
- Auto-detecting the swarm registry to drive `moai cc`/`moai glm` workflow integration — future SPEC
- Customizing the launched command (e.g., `--team --command "moai cc -p work"`) — minimal flag surface for v0.1; future SPEC may add `--launch-cmd` or profile support
- Migrating SPEC-WORKTREE-002 `--tmux` flag — `--tmux` remains independent; combining `--tmux --team` is not supported in v0.1 (mutual exclusion error)
- Windows-specific `syscall.Exec` shim implementation depth — REQ-WTL-012 specifies the high-level behavior; detailed implementation is plan.md M4

## 5. HARD Constraints

- [HARD] CLI MUST NOT invoke AskUserQuestion (BODP per `branch-origin-protocol.md` § HARD Rules + CONST-V3R5-030). Static guard `TestNew_NoAskUserQuestion` extended.
- [HARD] settings.local.json MUST NOT be mutated by `--team` — read-only detection only. Mutation is the exclusive domain of `moai cg`/`moai glm`/`moai cc`. Test: post-`--team` settings.local.json MUST be byte-identical to pre-`--team`.
- [HARD] BODP audit trail (`bodp.WriteDecision`) MUST run before team launch — the existing `writeWorktreeAuditTrail` call at `new.go` L165 remains untouched and precedes any team-launch step.
- [HARD] Template-First Rule per `CLAUDE.local.md §2` — every rule/skill update has a byte-identical template mirror under `internal/template/templates/`.
- [HARD] `swarm/` registry directory is **per-project** under `.moai/state/swarm/`, NOT user-global. This isolates swarm state to the project and avoids cross-project leaks.
- [HARD] Cross-platform build: `GOOS=darwin GOARCH=amd64 go build ./...` AND `GOOS=linux GOARCH=amd64 go build ./...` AND `GOOS=windows GOARCH=amd64 go build ./...` MUST all exit 0. Use `//go:build` tags to separate POSIX `syscall.Exec` from Windows simulation.
- [HARD] C-HRA-008-style subagent boundary discipline does not apply (this is CLI code, not harness/hook), but the spiritual equivalent is in REQ-WTL-013: no AskUserQuestion in CLI path.
- [HARD] Frontmatter canonical schema per `.claude/rules/moai/development/spec-frontmatter-schema.md` — 12 required fields including `created`/`updated`/`tags` (NOT `created_at`/`updated_at`/`labels`).
