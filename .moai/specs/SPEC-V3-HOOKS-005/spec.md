---
id: SPEC-V3-HOOKS-005
title: "Missing Hook Event Handlers — 14 Events"
version: "0.1.0"
status: draft
created: 2026-04-22
updated: 2026-04-22
author: GOOS
priority: P0 Critical
phase: "v3.0.0 — Phase 2 Hook Protocol v2 Core"
module: "internal/hook/handlers/, internal/cli/deps.go"
dependencies:
  - SPEC-V3-HOOKS-001
  - SPEC-V3-HOOKS-003
related_gap:
  - gm#24
  - gm#25
  - gm#26
  - gm#27
  - gm#28
  - gm#29
  - gm#30
  - gm#31
  - gm#32
related_theme: "Theme 1: Hook Protocol v2 — Missing Event Handlers"
breaking: true
bc_id: BC-005
lifecycle: spec-anchored
tags: "hook, v3, handler, events, breaking"
---

# SPEC-V3-HOOKS-005: Missing Hook Event Handlers

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-22 | GOOS | Initial v3 draft from Wave 4 bundle (Hooks/Commands) |

---

## 1. Goal (목적)

moai-adk-go declares 27 `EventType` constants at `internal/hook/types.go:19-114` (findings-wave1-moai-current.md §5.1) but registers handlers only for a subset via `InitDependencies()` at `internal/cli/deps.go:151-186`. Fourteen events have type constants but no concrete handler implementations, meaning the rich behaviors they enable (auto-rollback on tool error, mid-session config-change blocking, programmatic permission decisions, watchPaths file watchers, MCP elicitation auto-response, etc.) are unreachable.

This SPEC registers handlers for:
- **PostToolUseFailure** — react to failed tool calls (auto-rollback, notify).
- **StopFailure** — handle API errors (rate_limit, auth_failed, billing_error).
- **SubagentStart** — pre-subagent setup (inject context, configure tools).
- **PostCompact** — observe compacted summary and persist.
- **PermissionRequest** — programmatic allow/deny/ask decisions BEFORE a dialog.
- **PermissionDenied** — emit retry hint `{retry: true}`.
- **Setup** — once-per-repo init/maintenance.
- **TaskCreated** — observe or block task creation.
- **Elicitation / ElicitationResult** — auto-respond or override MCP elicitation.
- **ConfigChange** — react to settings edits mid-session; may BLOCK changes.
- **InstructionsLoaded** — audit which CLAUDE.md / rule / memory file loaded.
- **CwdChanged** — react to cwd changes; inject env via CLAUDE_ENV_FILE.
- **FileChanged** — watch specified files; fires on change/add/unlink.

Additionally this SPEC redefines `WorktreeCreate` from observational to PROVIDER semantics (stdout must contain the absolute path to the created worktree). This is the only breaking change in this SPEC and corresponds to BC-005.

## 2. Scope (범위)

In-scope:
- Handler registration for 14 previously-unwired events listed in §1.
- Upgrade WorktreeCreate to PROVIDER contract (stdout = absolute path, non-zero exit = failed).
- Upgrade PreCompact, ConfigChange, InstructionsLoaded, Elicitation, ElicitationResult, WorktreeRemove from observational to structured-output emission per SPEC-V3-HOOKS-001 variants.
- File-watcher integration for SessionStart's `watchPaths` reply and CwdChanged / FileChanged events; uses `fsnotify` or equivalent.
- Task-notification bus for SubagentStart (pre-subagent context injection).
- Per-event structured output payloads (PermissionRequestDecision, PermissionDeniedRetry, PostCompactSummary, etc.) aligned with SPEC-V3-HOOKS-001 variants.

Out-of-scope:
- Hook output protocol itself → SPEC-V3-HOOKS-001.
- Async / once / CLAUDE_ENV_FILE wiring → SPEC-V3-HOOKS-003 (consumed here).
- Source precedence hierarchy → SPEC-V3-HOOKS-006.
- Workspace trust gating → deferred.
- Teams-specific event handlers (TaskCompleted, TeammateIdle already wired per findings-wave1-moai-current.md §5.1).

## 3. Environment (환경)

Current moai-adk state:
- `EventType` constants exist for all 27 events in `internal/hook/types.go:19-114` (findings-wave1-moai-current.md §5.1).
- `InitDependencies()` at `internal/cli/deps.go:151-186` (findings-wave1-moai-current.md §15.5) registers 28 handler calls against 27 events — extra call is the `AutoUpdateHandler` composed with SessionStart; documented mismatch (gm#193).
- Handlers MISSING for: PostToolUseFailure, StopFailure, SubagentStart, PostCompact, PermissionRequest, PermissionDenied, Setup, TaskCreated, Elicitation, ElicitationResult, ConfigChange, InstructionsLoaded, CwdChanged, FileChanged (per gap-matrix §14.1).
- WorktreeCreate handler exists but is observational only (findings-wave1-hooks-commands.md §14.4). Does not return absolute path on stdout.

Claude Code reference:
- `entrypoints/sdk/coreTypes.ts:25-53` — 27 event list (findings-wave1-hooks-commands.md §1).
- `utils/hooks.ts:3394, 3450, 3492, 3570, 3594, 3639, 3709, 3745, 3789, 3826, 3867, 3902, 3932, 3961, 4034, 4097, 4157, 4214, 4335, 4470, 4525, 4928, 4967` — per-event trigger file line references (findings-wave1-hooks-commands.md §1.2).
- `utils/hooks.ts:4928-4958` — WorktreeCreate as PROVIDER hook, stdout = absolute path (findings-wave1-hooks-commands.md §14.4).
- `utils/hooks/fileChangedWatcher.ts` (5,309 bytes) — watchPaths dynamic update (findings-wave1-hooks-commands.md §1.2, CwdChanged).
- `utils/hooks/hooksConfigManager.ts:216-217` — ConfigChange exit-code 2 blocks settings change (findings-wave1-hooks-commands.md §1.2, ConfigChange).
- `utils/hooks/hooksConfigManager.ts:231-232` — InstructionsLoaded is observational (cannot block) (findings-wave1-hooks-commands.md §1.2, InstructionsLoaded).

Affected modules:
- `internal/hook/handlers/` — 14 new handler files (one per event).
- `internal/cli/deps.go` — register new handlers.
- `internal/hook/watcher.go` — new file: file watcher for watchPaths / FileChanged.
- `internal/hook/registry.go` — route events to new handlers.

## 4. Assumptions (가정)

- `fsnotify` (stdlib-adjacent, Apache-2.0) is an acceptable dependency for file-system watching; it is already indirectly depended on via tooling though not yet in `go.mod` direct deps.
- MCP elicitation semantics match CC's `mcp_server_name` + `elicitation_id` + `action: accept|decline|cancel` shape (findings-wave1-hooks-commands.md §1.2).
- InstructionsLoaded is observability-only: return value cannot block.
- ConfigChange exit-code 2 semantics: the value is the "block" decision; JSON-protocol equivalent is `HookOutput.Decision = "block"`.
- WorktreeCreate provider contract: handler MUST write absolute path as last line of stdout when it successfully creates a worktree.

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

- REQ-HOOKS-005-001: The system SHALL register concrete handlers for every event listed in `internal/hook/types.go` EventType constants such that `moai doctor hook --validate` reports 27/27 events wired.
- REQ-HOOKS-005-002: Each new handler SHALL consume and emit rich JSON output per SPEC-V3-HOOKS-001 variants.
- REQ-HOOKS-005-003: The PermissionRequest handler SHALL be capable of returning a `PermissionRequestOutput` with `decision: {behavior: "allow"|"deny", updatedInput?, updatedPermissions?, message?, interrupt?}`.
- REQ-HOOKS-005-004: The PermissionDenied handler SHALL be capable of returning `{retry: boolean}` to signal the model may retry.
- REQ-HOOKS-005-005: The Setup handler SHALL fire exactly once per `trigger` value per session (e.g., once on `init`, once on `maintenance`).
- REQ-HOOKS-005-006: The InstructionsLoaded handler SHALL NOT block the instruction load (observational-only per CC contract).
- REQ-HOOKS-005-007: The WorktreeCreate handler SHALL implement PROVIDER semantics: stdout MUST contain the absolute path to the created worktree; non-zero exit indicates creation failure.
- REQ-HOOKS-005-008: The file watcher subsystem SHALL register absolute paths from SessionStart's `watchPaths` reply and from CwdChanged's reply, and SHALL fire FileChanged events for `change`, `add`, `unlink` file-system operations on those paths.

### 5.2 Event-driven Requirements

- REQ-HOOKS-005-010: WHEN a PostToolUseFailure event fires, the handler SHALL receive `tool_name`, `tool_input`, `error`, and `is_interrupt` payload fields.
- REQ-HOOKS-005-011: WHEN a StopFailure event fires, the handler SHALL receive an `error` field whose value is one of: `rate_limit`, `authentication_failed`, `billing_error`, `invalid_request`, `server_error`, `max_output_tokens`, `unknown`.
- REQ-HOOKS-005-012: WHEN a ConfigChange handler returns `HookOutput.Decision = "block"`, the system SHALL reject the settings change and revert the in-memory config to the pre-change state.
- REQ-HOOKS-005-013: WHEN an Elicitation handler returns `{action: "accept", content: {...}}`, the system SHALL forward the response to the MCP server as the elicitation result.
- REQ-HOOKS-005-014: WHEN a SubagentStart event fires with `agent_id` and `agent_type`, the handler MAY return `additionalContext` which SHALL be injected into the subagent's first turn.
- REQ-HOOKS-005-015: WHEN a FileChanged event fires and the hook returns `watchPaths`, the file watcher SHALL dynamically update its watch list to union the new paths with the existing set.
- REQ-HOOKS-005-016: WHEN a CwdChanged event fires and the hook writes `export K=V` lines to `$CLAUDE_ENV_FILE`, the subsequent BashTool invocation SHALL source the file (per SPEC-V3-HOOKS-003).
- REQ-HOOKS-005-017: WHEN a WorktreeCreate handler exits with code 0 but stdout does not contain a valid absolute path, the system SHALL treat the hook as failed and emit `HookWorktreeCreateMissingPath` error.

### 5.3 State-driven Requirements

- REQ-HOOKS-005-020: WHILE the file watcher is active, it SHALL batch events within a 100 ms window to coalesce multiple rapid-fire changes to the same path into a single FileChanged dispatch.
- REQ-HOOKS-005-021: WHILE a PermissionRequest handler is executing, the system SHALL suspend the corresponding permission dialog until the handler returns or times out (per `DefaultHookTimeout`).

### 5.4 Optional Features

- REQ-HOOKS-005-030: WHERE the configuration key `hook.watcher.max_paths` is set, the file watcher SHALL refuse to register more than the configured number of paths and SHALL emit `HookWatcherCapacityExceeded`.
- REQ-HOOKS-005-031: WHERE a Setup handler's `trigger` is `maintenance`, the handler MAY perform destructive cleanup; WHERE the trigger is `init`, the handler SHALL NOT perform destructive cleanup.

### 5.5 Complex Requirements

- REQ-HOOKS-005-040: IF a ConfigChange handler returns `decision: "block"` AND the changed setting belongs to the `policy_settings` tier, THEN the block is ignored (policy-tier overrides), a warn log is emitted, and the change proceeds; ELSE the block is honored.
- REQ-HOOKS-005-041: IF an InstructionsLoaded handler emits `Continue: false`, THEN the value is ignored (observational-only per contract) and a warn log is emitted explaining why the directive had no effect; ELSE normal flow.

## 6. Acceptance Criteria (수용 기준 요약)

- AC-HOOKS-005-01: Given `moai doctor hook --validate` on a fresh v3 install, When executed, Then the output reports 27/27 events wired and zero unhandled event types. (maps REQ-HOOKS-005-001)
- AC-HOOKS-005-02: Given a PermissionRequest handler returning `{"hookSpecificOutput":{"hookEventName":"PermissionRequest","decision":{"behavior":"allow","updatedInput":{"file_path":"/safe/path"}}}}`, When resolved, Then the tool call proceeds with `/safe/path` and no dialog is shown to the user. (maps REQ-HOOKS-005-003)
- AC-HOOKS-005-03: Given a PermissionDenied handler returning `{"hookSpecificOutput":{"hookEventName":"PermissionDenied","retry":true}}`, When the model receives the response, Then the model is told it may retry the tool call. (maps REQ-HOOKS-005-004)
- AC-HOOKS-005-04: Given a WorktreeCreate handler that writes `/abs/worktrees/my-branch\n` to stdout and exits 0, When resolved, Then the system records the worktree at that path. (maps REQ-HOOKS-005-007)
- AC-HOOKS-005-05: Given a WorktreeCreate handler that exits 0 with empty stdout, When resolved, Then `HookWorktreeCreateMissingPath` error is returned. (maps REQ-HOOKS-005-017)
- AC-HOOKS-005-06: Given a SessionStart handler returning `watchPaths: ["/abs/.env"]`, When the file `/abs/.env` is modified, Then a FileChanged event fires with `file_path: "/abs/.env"`, `event: "change"`. (maps REQ-HOOKS-005-008, REQ-HOOKS-005-015)
- AC-HOOKS-005-07: Given a ConfigChange handler returning `Decision: "block"` for a project-tier change, When resolved, Then the in-memory config is reverted and the file-level change is rejected at next load. (maps REQ-HOOKS-005-012)
- AC-HOOKS-005-08: Given a ConfigChange handler returning `Decision: "block"` for a policy-tier change, When resolved, Then the block is ignored and a warn log is emitted. (maps REQ-HOOKS-005-040)
- AC-HOOKS-005-09: Given an InstructionsLoaded handler returning `Continue: false`, When resolved, Then the instruction load still completes and a warn log is emitted. (maps REQ-HOOKS-005-041)
- AC-HOOKS-005-10: Given `hook.watcher.max_paths: 10` and 11 watchPaths requested, When resolved, Then the 11th registration fails with `HookWatcherCapacityExceeded`. (maps REQ-HOOKS-005-030)
- AC-HOOKS-005-11: Given a Setup handler with `trigger: "init"`, When fired, Then the handler detects `init` and does not execute cleanup code paths. (maps REQ-HOOKS-005-031)

## 7. Constraints (제약)

- Technical: Go 1.22+. `fsnotify` added to `go.mod` direct deps (brings moai to 10 direct deps; justified as critical per master-v3 §1.2 design principles).
- Backward compat: Existing 13 wired handlers remain unchanged; 14 new handlers are additive. Exception: WorktreeCreate upgrade to PROVIDER semantics is BC-005 (breaking).
- Platform: File watcher semantics differ slightly across OSes (macOS FSEvents vs Linux inotify vs Windows ReadDirectoryChangesW); `fsnotify` abstracts these but batching (REQ-HOOKS-005-020) standardizes delivery.
- Performance: File-watcher max active paths default cap 1,000; configurable via `hook.watcher.max_paths`.

## 8. Risks & Mitigations (리스크 및 완화)

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| WorktreeCreate provider upgrade breaks users with custom observational handlers | M | H | BC-005 documented in migration guide; `moai doctor hook --validate` detects legacy observational shape and suggests fix; dual-parse phase for one minor version. |
| File watcher leaks FDs when watching many paths | L | M | `hook.watcher.max_paths` cap with `HookWatcherCapacityExceeded` error; `moai doctor` surfaces active path count. |
| ConfigChange block races with parallel config writes | L | M | ConfigChange handler runs synchronously before the in-memory config is swapped; no race possible on single-writer path. |
| MCP elicitation handler hangs if server disconnects mid-dialog | L | M | Handler bound by `DefaultHookTimeout`; timeout triggers `action: "cancel"`. |
| PermissionRequest handler abuse auto-approves dangerous tools | L | H | `moai doctor hook --validate` highlights PermissionRequest handlers that return `allow` for Write/Bash without additional gating; docs warn users about the security implications. |

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3-HOOKS-001 (rich JSON output variants).
- SPEC-V3-HOOKS-003 (CLAUDE_ENV_FILE mechanism for CwdChanged / FileChanged).

### 9.2 Blocks

- SPEC-V3-PLG-001 (plugin manifests may ship handlers for any of these events).

### 9.3 Related

- SPEC-V3-HOOKS-004 (matcher/if filter applies before handler dispatch).
- SPEC-V3-HOOKS-006 (source precedence determines handler source tier when multiple are declared).

## 10. Traceability (추적성)

- Theme: master-v3 Section 3.1 (Theme 1 — Hook Protocol v2) "handler richness upgrade" sub-theme.
- Gap rows: gm#24 (Medium — WorktreeCreate provider), gm#25 (Medium — SessionStart watchPaths), gm#26 (Low — CwdChanged watchPaths), gm#27 (High — PermissionRequest decision), gm#28 (Medium — PermissionDenied retry), gm#29 (Medium — PreCompact), gm#30 (Low — Elicitation), gm#31 (Medium — ConfigChange block), gm#32 (Medium — InstructionsLoaded observational).
- BC-ID: BC-005 (WorktreeCreate provider contract upgrade).
- Wave 1 sources: findings-wave1-hooks-commands.md §1 (27-event catalog), §1.2 (per-event schema), §14.1 (missing event table), §14.4 (WorktreeCreate PROVIDER semantics).
- Priority: P0 Critical (achieves HOOK-PARITY success metric in master-v3 §1.3).
