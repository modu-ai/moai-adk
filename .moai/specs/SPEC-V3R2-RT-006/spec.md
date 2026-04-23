---
id: SPEC-V3R2-RT-006
title: "Hook Handler Completeness and 27-Event Coverage"
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: GOOS
priority: P0 Critical
phase: "v3.0.0 — Phase 2 — Runtime Hardening"
module: "internal/hook/"
dependencies:
  - SPEC-V3R2-CON-001
  - SPEC-V3R2-RT-001
  - SPEC-V3R2-RT-002
  - SPEC-V3R2-RT-004
  - SPEC-V3R2-RT-005
bc_id: [BC-V3R2-018]
related_principle: [P8 Hook JSON, P2 ACI, P6 Permission Bubble]
related_pattern: [T-5, T-1, S-1]
related_problem: [P-H01, P-H02, P-H03, P-H15, P-H16, P-H17, P-H19, P-R01]
related_theme: "Layer 3: Runtime"
breaking: true
lifecycle: spec-anchored
tags: "hook, handlers, coverage, v3r2, breaking, runtime, tmux, p-h02"
---

# SPEC-V3R2-RT-006: Hook Handler Completeness and 27-Event Coverage

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-23 | GOOS | Initial v3 Round-2 draft. Supersedes SPEC-V3-HOOKS-005 (missing events) and SPEC-V3-HOOKS-002 partial (handler upgrades). Addresses 10 logging-only handlers from P-H01, the CRITICAL subagent-stop tmux-pane-leak bug P-H02, orphan setupHandler P-H03, and 59% partial coverage P-H19. |

---

## 1. Goal (목적)

Decide each of the 27 Claude Code hook events with an explicit "full logic" or "retire from settings.json" verdict, eliminating the stub-handler debt that r6 §2.2 + master §7.3 enumerate. This SPEC composes on top of SPEC-V3R2-RT-001 (JSON-OR-ExitCode protocol) — the protocol unblocks rich business logic, this SPEC ships that logic. It also fixes the **P-H02 tmux pane leak bug** (subagentStop handler currently only logs; must read tmuxPaneId from team config and `kill-pane` before returning), removes orphan `setupHandler`, and reconciles the settings.json registration count with the Go handler count (r6 §2.3 + P-R01).

Every one of the 27 events receives one of three resolutions: **UPGRADE** (stub handler → full business logic), **KEEP** (existing full handler retained), or **RETIRE** (remove from settings.json; Go handler retained as observability tap with explicit disable option per master §8 BC-V3R2-018). The mapping is enumerated below in §5. No event is left in partial-coverage limbo.

Master §7.3 provides the inventory table; this SPEC converts the table into executable requirements. The critical bug-fix dimension (P-H02) is prioritized — the subagentStop tmux-pane-cleanup logic must ship before any other handler upgrade to unblock team-mode workflows that are currently leaking panes per MEMORY.md feedback entry "Team tmux pane cleanup".

## 2. Scope (범위)

In-scope:

- Per-event resolution for all 27 events enumerated in §5 below (columns: current state, target state, input JSON shape, handler action, output JSON shape).
- Upgrade of 5 critical handlers first per r6 §A top-5 gap list:
  1. `subagentStop` — read tmuxPaneId from team config, kill-pane, update team registry (FIXES P-H02 KNOWN BUG).
  2. `configChange` — re-render generated files when `.moai/config/sections/*.yaml` changes, re-parse typed loaders, emit ConfigReloaded SystemMessage (FIXES P-H15).
  3. `instructionsLoaded` — validate CLAUDE.md 40,000-char budget per coding-standards.md, emit warning SystemMessage if exceeded (FIXES P-H16).
  4. `fileChanged` — trigger MX re-scan for externally-edited files in 16 supported language extensions (FIXES P-H17).
  5. `postToolUseFailure` — error classification by signature (ExitError, TimeoutError, ContextCancelled, PermissionDenied), emit structured diagnostic via HookResponse.SystemMessage.
- Orphan resolution: `setupHandler` (P-H03) — REMOVE from `internal/hook/setup.go` since no shell wrapper and no settings.json entry exist; `AutoUpdate` composite SessionStart handler remains documented but not exposed as a separate event.
- Retire 4 events from settings.json: `notification`, `elicitation`, `elicitationResult`, `taskCreated` (per master §7.3). Their Go handlers remain available as observability taps behind `.moai/config/sections/system.yaml` `hook.observability_events` opt-in list.
- Handler documentation rule: each handler file MUST declare its event name, resolution category (UPGRADE/KEEP/RETIRE/FIX), and the `HookResponse` fields it may populate. Enforced by `internal/hook/audit_test.go` (new test file).
- Registration parity check: `internal/hook/audit_test.go` verifies `deps.go` handler count == settings.json event count + autoUpdate composite (1) + permitted observability-only handlers (RETIRE list).
- 27-event coverage table authoritative in this SPEC — future event additions require updating §5 before merging the new handler.

Out-of-scope (addressed by other SPECs):

- Hook JSON-OR-ExitCode protocol — SPEC-V3R2-RT-001 (consumed here).
- Permission decision wiring in PreToolUse — SPEC-V3R2-RT-002.
- Sandbox enforcement — SPEC-V3R2-RT-003.
- Session state checkpointing triggered by hooks — SPEC-V3R2-RT-004.
- Settings reload semantics beyond the diff-aware reload trigger — SPEC-V3R2-RT-005.
- Hardcoded path in shell wrappers — SPEC-V3R2-RT-007.
- @MX tag autonomous add/update/remove — SPEC-V3R2-SPC-002 (integration point only here).
- Sprint Contract injection via PostToolUse — SPEC-V3R2-HRN-002.

## 3. Environment (환경)

Current moai-adk state per r6-commands-hooks-style-rules.md §2-§A:

- 27 Go handlers registered in `internal/cli/deps.go:152-186`.
- 25 events in `.claude/settings.json` native registrations + 1 composite (autoUpdate on SessionStart) + 1 orphan (setupHandler has Go code but no shell wrapper or settings entry).
- 16 events with full business logic (59%); 10 logging-only (37%); 1 missing (Setup, 4%).
- Known bugs:
  - P-H02 subagentStop tmux pane leak (MEMORY.md feedback entry).
  - P-H03 setupHandler orphan.
  - P-H04 hardcoded `/Users/goos/go/bin/moai` in 26 shell wrappers (handled by SPEC-V3R2-RT-007).
  - P-H15 configChangeHandler no-op (should re-render).
  - P-H16 instructionsLoadedHandler no-op (should validate 40k char limit).
  - P-H17 fileChangedHandler no-op (should MX re-scan).

Team protocol reference:

- `.claude/rules/moai/workflow/team-protocol.md` describes team-mode coordination.
- MEMORY.md feedback entries: `feedback_team_tmux_cleanup.md` = "TeamDelete 전 tmuxPaneId로 kill-pane 필수 (SubagentStop 핸들러 미구현 근본원인)"; this SPEC codifies the fix.
- Team config file: `~/.claude/teams/{team-name}/config.json` per team-protocol §Team Discovery.

Affected modules:

- `internal/hook/subagent_stop.go` — UPGRADE logic (tmux pane kill).
- `internal/hook/config_change.go` — UPGRADE logic (reload).
- `internal/hook/instructions_loaded.go` — UPGRADE logic (length check).
- `internal/hook/file_changed.go` — UPGRADE logic (MX rescan).
- `internal/hook/post_tool_failure.go` — UPGRADE logic (error classification).
- `internal/hook/setup.go` — DELETE.
- `internal/hook/notification.go`, `elicitation.go`, `task_created.go` — KEEP code but remove from settings.json registration.
- `.claude/settings.json` + `internal/template/templates/.claude/settings.json.tmpl` — remove the 4 RETIRE events.
- `internal/hook/audit_test.go` — new file, registration parity CI rule.

## 4. Assumptions (가정)

- SPEC-V3R2-RT-001 (hook JSON-OR-ExitCode protocol) ships before any handler upgrade in this SPEC — handlers use `HookResponse` throughout.
- Team config file format at `~/.claude/teams/{team-name}/config.json` exposes `tmuxPaneId` per teammate as currently implemented.
- `tmux kill-pane -t <id>` is the canonical pane-teardown command; error cases (pane already gone, session ended) do not raise errors from kill-pane.
- CLAUDE.md 40,000 char budget is documented in coding-standards.md as a HARD rule; validation emits a warning but does not block the session (developer choice preserved).
- MX tag re-scan uses the TagScanner from SPEC-V3R2-SPC-002 (integration interface called; implementation lives there).
- Retired events (`notification`, `elicitation`, `elicitationResult`, `taskCreated`) have no meaningful business-logic home in moai v3; retention is purely observability tap.
- `SessionEnd` 1500ms timeout noted in r3 §3 still applies; subagentStop tmux cleanup must complete within that budget.
- Windows does not use tmux; `subagentStop` tmux-cleanup path is no-op on Windows.

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

- REQ-V3R2-RT-006-001: The system SHALL implement a Go handler for every event listed in the 27-event table below in §5.7; handlers SHALL emit `HookResponse` (from SPEC-V3R2-RT-001) on stdout.
- REQ-V3R2-RT-006-002: Every handler file SHALL declare at the top: event name, resolution category (UPGRADE / KEEP / RETIRE-OBS-ONLY / FIX), and the `HookResponse` fields it may populate.
- REQ-V3R2-RT-006-003: `internal/hook/audit_test.go` SHALL verify that `deps.go` handler count matches the sum of `(settings.json native events) + 1 (autoUpdate composite) + |observability-only handlers declared in system.yaml|`.
- REQ-V3R2-RT-006-004: The 4 retired events (`notification`, `elicitation`, `elicitationResult`, `taskCreated`) SHALL be removed from `settings.json` + template; their Go handlers SHALL remain available as observability taps via `.moai/config/sections/system.yaml` key `hook.observability_events: []`.
- REQ-V3R2-RT-006-005: The orphan `setupHandler` SHALL be removed from `internal/hook/` entirely; no shell wrapper, no settings.json entry, no Go file.
- REQ-V3R2-RT-006-006: The upgraded `subagentStopHandler` SHALL read `tmuxPaneId` from `~/.claude/teams/{team-name}/config.json`, invoke `tmux kill-pane -t <id>`, and update the team registry to reflect pane teardown before returning.
- REQ-V3R2-RT-006-007: On Windows where `tmux` is absent, the `subagentStopHandler` tmux-cleanup path SHALL be a no-op; the handler SHALL return `HookResponse{}` without error.

### 5.2 Event-Driven Requirements — Upgraded Handlers

- REQ-V3R2-RT-006-010: WHEN `SubagentStop` fires with input JSON `{sessionId, teammateName, reason}`, the handler SHALL (a) look up the team config, (b) extract `tmuxPaneId` for that teammate, (c) invoke `tmux kill-pane -t <id>` with best-effort error swallowing, (d) update team config removing the teammate entry, (e) return `HookResponse{SystemMessage: "Teammate <name> shut down, pane <id> released"}`.
- REQ-V3R2-RT-006-011: WHEN `ConfigChange` fires with input `{path: "...yaml"}`, the handler SHALL (a) re-read the changed yaml file, (b) re-run typed loader + validator/v10, (c) if valid, invoke the diff-aware reload API from SPEC-V3R2-RT-005 naming the path, (d) emit `HookResponse{AdditionalContext: "<path> reloaded successfully"}`, (e) if invalid, emit `HookResponse{SystemMessage: "Config reload failed: <error>", Continue: false}`.
- REQ-V3R2-RT-006-012: WHEN `InstructionsLoaded` fires with input `{paths: [claude.md paths]}`, the handler SHALL (a) read each file, (b) count UTF-8 characters, (c) if any file exceeds 40,000 chars, emit `HookResponse{SystemMessage: "<path> exceeds 40,000 char budget at <N>; split content per coding-standards.md"}`, (d) otherwise return `HookResponse{}`.
- REQ-V3R2-RT-006-013: WHEN `FileChanged` fires with input `{path, changeType}` and the path ends in one of the 16 supported language extensions, the handler SHALL invoke the MX TagScanner from SPEC-V3R2-SPC-002 and, if new or removed @MX markers are detected, emit `HookResponse{AdditionalContext: "MX tag delta on <path>: <summary>"}`.
- REQ-V3R2-RT-006-014: WHEN `PostToolUseFailure` fires with input `{toolName, error, stderr}`, the handler SHALL classify the error (ExitError, TimeoutError, ContextCancelled, PermissionDenied, SandboxViolation, OOMKilled, UnknownFailure) and emit `HookResponse{SystemMessage: "<category>: <actionable message>", AdditionalContext: "<diagnostic hints>"}`.

### 5.3 Event-Driven Requirements — Kept Handlers (semantic reaffirmation)

- REQ-V3R2-RT-006-020: WHEN `SessionStart` fires, the handler SHALL continue to perform GLM tmux setup, skill discovery, and memory evaluation per current behavior + emit `HookResponse.AdditionalContext` with the welcome/status block.
- REQ-V3R2-RT-006-021: WHEN `SessionEnd` fires, the handler SHALL continue memo save and MX scan within the 1500 ms timeout per r3 §3 item 5.
- REQ-V3R2-RT-006-022: WHEN `PreToolUse` fires, the handler SHALL emit `HookResponse.PermissionDecision` (from SPEC-V3R2-RT-002) and optional `UpdatedInput` for rewrite cases.
- REQ-V3R2-RT-006-023: WHEN `PostToolUse` fires, the handler SHALL emit `HookResponse.AdditionalContext` with @MX tag injections and LSP/metrics diagnostics.
- REQ-V3R2-RT-006-024: WHEN `PreCompact` / `PostCompact` fire, the handlers SHALL perform memo save and restore respectively with no change from v2 behavior.
- REQ-V3R2-RT-006-025: WHEN `Stop` / `StopFailure` fire, the handlers SHALL emit completion markers and error-class-specific systemMessage respectively.
- REQ-V3R2-RT-006-026: WHEN `SubagentStart` fires, the handler SHALL inject project context into the subagent turn via `HookResponse.AdditionalContext`.
- REQ-V3R2-RT-006-027: WHEN `TeammateIdle` fires, the handler SHALL evaluate LSP error thresholds and emit `HookResponse{Continue: false}` when quality gate fails.
- REQ-V3R2-RT-006-028: WHEN `TaskCompleted` fires, the handler SHALL validate SPEC acceptance criteria and emit `HookResponse{Continue: false}` when validation fails.
- REQ-V3R2-RT-006-029: WHEN `UserPromptSubmit` fires, the handler SHALL perform SPEC detection, session title update, and workflow-keyword routing via `HookResponse.AdditionalContext`.
- REQ-V3R2-RT-006-030: WHEN `PermissionRequest` fires, the handler SHALL re-verify `UpdatedInput` per SPEC-V3R2-RT-002 resolver contract.
- REQ-V3R2-RT-006-031: WHEN `PermissionDenied` fires, the handler SHALL suggest read-only retry via `HookResponse.SystemMessage`.
- REQ-V3R2-RT-006-032: WHEN `WorktreeCreate` / `WorktreeRemove` fire, handlers SHALL update the worktree registry in `.moai/state/worktrees.json`.
- REQ-V3R2-RT-006-033: WHEN `CwdChanged` fires, the handler SHALL write `CLAUDE_ENV_FILE` per current behavior.

### 5.4 State-Driven Requirements

- REQ-V3R2-RT-006-040: WHILE `hook.observability_events: [notification, elicitation, ...]` is set in system.yaml, the retired-event Go handlers SHALL remain live but emit only structured logs (never side-effect user-facing state).
- REQ-V3R2-RT-006-041: WHILE `hook.strict_mode: true` (from SPEC-V3R2-RT-001) AND a retired event fires, the handler SHALL still succeed silently (retired events are informational-only, not strict-mode relevant).
- REQ-V3R2-RT-006-042: WHILE `internal/hook/audit_test.go` is enabled, any merge that adds a new handler without updating §5.7 table SHALL fail CI.

### 5.5 Optional Features

- REQ-V3R2-RT-006-050: WHERE `moai doctor hook` is invoked, the system SHALL print the 27-event table from §5.7 with per-event resolution status and observability opt-in state.
- REQ-V3R2-RT-006-051: WHERE `moai doctor hook --trace <event>` is invoked, the system SHALL stream the handler's decision path for the most recent invocation.

### 5.6 Unwanted Behavior

- REQ-V3R2-RT-006-060: IF the subagentStop handler cannot find `tmuxPaneId` in the team config (entry already removed by concurrent call), THEN the handler SHALL log at DEBUG level and return `HookResponse{}` without error.
- REQ-V3R2-RT-006-061: IF `tmux kill-pane` fails with "pane not found", THEN the handler SHALL treat the cleanup as successful (pane already gone) and remove the teammate from registry.
- REQ-V3R2-RT-006-062: IF `configChange` reload produces a validation error AND the old settings are still valid in memory, THEN the handler SHALL retain the old settings and emit `HookResponse{SystemMessage: "Config reload rejected: <field>: <error>; old settings retained"}`.
- REQ-V3R2-RT-006-063: IF a retired event is somehow registered in settings.json despite removal, THEN `audit_test.go` SHALL fail the build naming the event.

### 5.7 27-Event Coverage Table (authoritative)

| # | Event | Current (v2) | v3 Resolution | Handler Action | HookResponse Fields |
|---|-------|--------------|---------------|----------------|---------------------|
| 1 | SessionStart | Full | KEEP | GLM setup, skill discovery, memory eval | `AdditionalContext`, `WatchPaths` |
| 2 | SessionEnd | Full | KEEP | Memo save, MX scan within 1500ms | `SystemMessage` |
| 3 | PreToolUse | Full | KEEP+JSON | Security scan, secrets, reflective write; emit PermissionDecision | `PermissionDecision`, `UpdatedInput`, `AdditionalContext` |
| 4 | PostToolUse | Full | KEEP+JSON | MX validate, LSP convert, metrics, MX tag injection | `AdditionalContext` (MX markers) |
| 5 | PostToolUseFailure | Stub | **UPGRADE** | Error classification (ExitError/TimeoutError/ContextCancelled/PermissionDenied/SandboxViolation/OOMKilled/UnknownFailure) | `SystemMessage`, `AdditionalContext` |
| 6 | PreCompact | Full | KEEP | Memo save | `SystemMessage` |
| 7 | PostCompact | Full | KEEP | Memo restore | `AdditionalContext` |
| 8 | Stop | Full | KEEP | Completion markers, Ralph state | `SystemMessage` |
| 9 | StopFailure | Full | KEEP | Error-class systemMessage | `SystemMessage` |
| 10 | SubagentStart | Full | KEEP | Project context injection | `AdditionalContext` |
| 11 | SubagentStop | **Stub (BUG P-H02)** | **FIX** | Read tmuxPaneId, `tmux kill-pane`, update team registry | `SystemMessage` |
| 12 | Notification | Stub | **RETIRE-OBS-ONLY** | Remove from settings.json; observability tap via system.yaml | (none unless observability opt-in) |
| 13 | UserPromptSubmit | Full | KEEP+JSON | SPEC detect, session title, workflow kw | `AdditionalContext` |
| 14 | PermissionRequest | Full | KEEP+JSON | UpdatedInput re-verify | `PermissionDecision`, `UpdatedInput` |
| 15 | PermissionDenied | Full | KEEP | Read-only retry suggestion | `SystemMessage` |
| 16 | TeammateIdle | Full | KEEP+JSON | Quality gate enforcement; emit `continue: false` on fail | `Continue`, `SystemMessage` |
| 17 | TaskCompleted | Full | KEEP+JSON | SPEC acceptance criteria validation; `continue: false` on fail | `Continue`, `SystemMessage` |
| 18 | TaskCreated | Stub | **RETIRE-OBS-ONLY** | Remove from settings.json; observability tap | (none unless opt-in) |
| 19 | WorktreeCreate | Full | KEEP | Registry update at `.moai/state/worktrees.json` | `SystemMessage` |
| 20 | WorktreeRemove | Full | KEEP | Registry cleanup | `SystemMessage` |
| 21 | ConfigChange | Stub | **UPGRADE** | Re-render + diff-aware reload via SPEC-V3R2-RT-005 | `AdditionalContext`, `SystemMessage`, `Continue` (on validation fail) |
| 22 | CwdChanged | Full | KEEP | CLAUDE_ENV_FILE write | `SystemMessage` |
| 23 | FileChanged | Stub | **UPGRADE** | MX re-scan for 16 supported languages | `AdditionalContext` |
| 24 | InstructionsLoaded | Stub | **UPGRADE** | CLAUDE.md 40,000-char budget check | `SystemMessage` |
| 25 | Elicitation | Stub | **RETIRE-OBS-ONLY** | Remove from settings.json; observability tap | (none unless opt-in) |
| 26 | ElicitationResult | Stub | **RETIRE-OBS-ONLY** | Remove from settings.json; observability tap | (none unless opt-in) |
| 27 | Setup | Orphan Go handler (no wrapper, no settings) | **REMOVE** | Delete `internal/hook/setup.go`; no wrapper; no registration | (n/a) |

Final counts: 15 KEEP + 5 UPGRADE + 1 FIX (subagentStop) + 4 RETIRE-OBS-ONLY + 1 REMOVE + 1 composite (autoUpdate on SessionStart) = 27 events resolved.

Go handler count after this SPEC: 27 current − 1 removed (setupHandler) = 26 Go handlers. Settings.json event count: 25 − 4 retired = 21 native events + 1 composite = 22 registrations. `audit_test.go` verifies: `26 Go handlers == 22 (registrations + composite) + 4 (observability-only retained) = 26`. ✓

## 6. Acceptance Criteria (수용 기준)

- AC-V3R2-RT-006-01: Given team mode with 3 teammates and `tmuxPaneId` entries in team config, When any teammate's SubagentStop fires, Then `tmux kill-pane` is invoked with the right id and the teammate entry is removed from `~/.claude/teams/{team-name}/config.json`. (maps REQ-V3R2-RT-006-006, -010)
- AC-V3R2-RT-006-02: Given a pane already gone and tmux returns "pane not found", When SubagentStop invokes kill-pane, Then the handler treats cleanup as successful and registry entry is still removed. (maps REQ-V3R2-RT-006-061)
- AC-V3R2-RT-006-03: Given Windows host, When SubagentStop fires, Then the handler returns `HookResponse{}` without error (tmux path is no-op). (maps REQ-V3R2-RT-006-007)
- AC-V3R2-RT-006-04: Given `.moai/config/sections/quality.yaml` is edited, When ConfigChange fires, Then the handler re-reads the file, runs typed loader, calls diff-aware reload, and emits `AdditionalContext: "<path> reloaded successfully"`. (maps REQ-V3R2-RT-006-011)
- AC-V3R2-RT-006-05: Given a quality.yaml edit introduces an invalid field, When ConfigChange reload fails validation, Then `SystemMessage` reports the error and `Continue: false` is set, AND old settings remain in memory. (maps REQ-V3R2-RT-006-011, -062)
- AC-V3R2-RT-006-06: Given CLAUDE.md is 42,000 characters, When InstructionsLoaded fires, Then HookResponse.SystemMessage names the file and the budget overage. (maps REQ-V3R2-RT-006-012)
- AC-V3R2-RT-006-07: Given `internal/auth/handler.go` is externally edited adding `@MX:WARN at line 42`, When FileChanged fires, Then HookResponse.AdditionalContext contains an MX tag delta summary. (maps REQ-V3R2-RT-006-013)
- AC-V3R2-RT-006-08: Given a Bash tool fails with exit code 124 (timeout), When PostToolUseFailure fires, Then SystemMessage classifies the error as "TimeoutError" with actionable hint. (maps REQ-V3R2-RT-006-014)
- AC-V3R2-RT-006-09: Given `internal/hook/setup.go` exists, When `go build ./...` compiles, Then the file is absent (removed by this SPEC). (maps REQ-V3R2-RT-006-005)
- AC-V3R2-RT-006-10: Given `settings.json` contains `"notification"` registration, When `audit_test.go` runs, Then the test fails naming `notification` as illegally registered. (maps REQ-V3R2-RT-006-063)
- AC-V3R2-RT-006-11: Given `system.yaml` key `hook.observability_events: [notification]`, When a Notification event fires, Then the Go handler emits structured log only, no user-facing side effect. (maps REQ-V3R2-RT-006-040)
- AC-V3R2-RT-006-12: Given `moai doctor hook`, When invoked, Then the 27-event table is printed with per-event resolution state. (maps REQ-V3R2-RT-006-050)
- AC-V3R2-RT-006-13: Given a developer adds a new Go handler without updating §5.7 table, When `audit_test.go` runs in CI, Then the test fails naming the new handler as undocumented. (maps REQ-V3R2-RT-006-003, -042)
- AC-V3R2-RT-006-14: Given every handler file declares its event + resolution category at top, When grep runs for the declaration, Then all 26 remaining handlers match the pattern. (maps REQ-V3R2-RT-006-002)
- AC-V3R2-RT-006-15: Given teammate shutdown-request → response flow (per team-protocol), When SubagentStop fires, Then pane teardown happens AFTER shutdown_response is sent, preserving the protocol order. (maps REQ-V3R2-RT-006-010)
- AC-V3R2-RT-006-16: Given `hook.observability_events: []` (empty), When a retired event fires, Then the Go handler returns `HookResponse{}` without logging. (maps REQ-V3R2-RT-006-040)
- AC-V3R2-RT-006-17: Given a PreToolUse handler needs to consult the permission stack, When it returns HookResponse, Then `PermissionDecision` field is populated (handler integration with SPEC-V3R2-RT-002). (maps REQ-V3R2-RT-006-022)

## 7. Constraints (제약)

- Technical: Go 1.22+; no new external dependencies. `tmux kill-pane` invoked via `os/exec` with `.Run()`; errors classified but not propagated as hook failure.
- Backward compat: Per master §8 BC-V3R2-018, retired events still fire at CC level but no moai native handler runs (observability tap is opt-in). v2.x users who depend on the 4 retired events can opt in via system.yaml.
- Platform: macOS / Linux / Windows. Windows tmux path is no-op per REQ-V3R2-RT-006-007.
- Performance: Each handler MUST complete in under 200 ms p99 except SessionStart (up to 2 s acceptable for GLM tmux setup) and SessionEnd (1500 ms ceiling per r3 §3).
- Handler size: Per-handler file MUST NOT exceed 300 LOC of business logic after upgrades; larger handlers split by concern.
- Test coverage: Each upgraded handler SHALL have unit tests with at least 85% line coverage; critical paths (subagentStop tmux, configChange reload-fail, instructionsLoaded 40k edge, fileChanged MX delta, postToolUseFailure error classification) SHALL have integration tests.
- Memory regression: MEMORY.md `teammate_mode_regression.md` notes that `make build && make install` + Claude Code restart is required; CI builds the binary but manual QA must verify the subagentStop fix in a live team-mode session before release.

## 8. Risks & Mitigations (리스크 및 완화)

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| tmux kill-pane latency exceeds SessionEnd 1500ms ceiling | L | M | Best-effort with error swallowing; async via goroutine with 500ms timeout per pane. |
| subagentStop handler upgrade doesn't reach users because of stale binary regression (MEMORY.md teammate_mode_regression) | H | H | Release notes explicitly instruct `make build && make install` + restart; `moai doctor --binary-hash` verifies freshness. |
| Retired events' Go handlers confuse future maintainers | M | L | Per-file resolution category declaration (REQ-V3R2-RT-006-002) makes status explicit; `moai doctor hook` surfaces it. |
| ConfigChange reload race (file mid-write when fsnotify fires) | M | M | Handler performs a brief 20 ms debounce; validation catches truncated writes. |
| Windows cohort lacks tmux coverage testing | M | L | Windows path is explicit no-op per REQ-V3R2-RT-006-007; skips test on non-tmux hosts. |
| InstructionsLoaded warnings fatigue users with large CLAUDE.md tolerances | L | L | Warning only — no block; users can acknowledge and move on; threshold adjustable via coding-standards.md amendment. |
| MX re-scan on every FileChanged event adds I/O load | L | M | Scanner uses memoized hash per path; re-scan only on actual content change. |
| Orphan setupHandler removal breaks an undocumented use case | L | L | r6 §2.2 confirms setupHandler is fully dead (no wrapper, no settings, logging-only); removal is safe. |

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3R2-RT-001 (HookResponse schema used by every handler here).
- SPEC-V3R2-RT-002 (PreToolUse PermissionDecision integration — REQ-022).
- SPEC-V3R2-RT-004 (TeammateIdle + TaskCompleted may record BlockerReport per HRN-002 flow).
- SPEC-V3R2-RT-005 (ConfigChange diff-aware reload API).
- SPEC-V3R2-SPC-002 (MX TagScanner for FileChanged handler).

### 9.2 Blocks

- SPEC-V3R2-HRN-002 (evaluator memory per-iteration composes with SubagentStart/SubagentStop semantics).
- SPEC-V3R2-WF-003 (Multi-mode router — Ralph loop consumes FileChanged MX re-scan output).
- SPEC-V3R2-MIG-002 (hook registration cleanup aligns to the 22-native + 1-composite + 4-observability count established here).

### 9.3 Related

- SPEC-V3R2-RT-007 (hardcoded-path fix in shell wrappers ships in parallel).
- SPEC-V3R2-ORC-002 (common protocol CI lint extends `audit_test.go` pattern established here).
- SPEC-V3R2-CON-003 (consolidation pass moves hook-system.md rule updates).

## 10. Traceability (추적성)

- Theme: master §4.3 Layer 3 Runtime; §7.3 Hook audit table (authoritative input); §8 BC-V3R2-018.
- Principle: P8 (Hook JSON Protocol); P2 (ACI — handlers expose structured observations).
- Pattern: T-5 (Hook Dual Protocol); T-1 (ACI); S-1 (Permission stack via PreToolUse).
- Problem: P-H01 (10 stub handlers), P-H02 (subagent-stop tmux leak CRITICAL), P-H03 (setupHandler orphan), P-H15/16/17 (ConfigChange/InstructionsLoaded/FileChanged no-ops), P-H19 (59% partial coverage), P-R01 (handler count drift).
- Master Appendix B: P-H02 → SPEC-V3R2-RT-006 primary resolution; P-H19 → same.
- Wave 1 sources: r6-commands-hooks-style-rules.md §2.2 (Go handler audit), §A Hook Coverage Matrix (27-event table input); MEMORY.md feedback entries `feedback_team_tmux_cleanup.md`, `teammate_mode_regression.md`.
- Wave 2 sources: problem-catalog.md Cluster 3 (Hook Completeness and Safety, CRITICAL).
- BC-ID: BC-V3R2-018 (retired events removed from settings.json; observability tap via opt-in).
- Priority: P0 Critical — contains P-H02 CRITICAL bug fix (known tmux pane leak); unblocks team-mode reliability.
