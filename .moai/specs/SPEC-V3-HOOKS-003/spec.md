---
id: SPEC-V3-HOOKS-003
title: "Async Hook Execution — async, asyncRewake, once"
version: "0.1.0"
status: draft
created: 2026-04-22
updated: 2026-04-22
author: GOOS
priority: P1 High
phase: "v3.0.0 — Phase 2 Hook Protocol v2 Core"
module: "internal/hook/async_registry.go, internal/hook/once.go"
dependencies:
  - SPEC-V3-HOOKS-001
related_gap:
  - gm#9
  - gm#10
  - gm#11
  - gm#20
related_theme: "Theme 1: Hook Protocol v2 — Async Execution"
breaking: false
bc_id: null
lifecycle: spec-anchored
tags: "hook, v3, async, once, env-file"
---

# SPEC-V3-HOOKS-003: Async Hook Execution

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-22 | GOOS | Initial v3 draft from Wave 4 bundle (Hooks/Commands) |

---

## 1. Goal (목적)

Enable long-running hooks without blocking the model turn. Claude Code supports `async: true` (backgrounded execution), `asyncRewake: true` (backgrounded + wake model on exit 2 via task-notification queue), and `once: true` (self-remove after first successful execution). Combined with the `CLAUDE_ENV_FILE` mechanism for SessionStart / Setup / CwdChanged / FileChanged events (hook writes bash exports picked up by subsequent BashTool commands), this unlocks Ralph-style long-running quality loops, one-shot initializers, and dynamic env management that moai cannot perform today.

## 2. Scope (범위)

In-scope:
- `async: true` flag: background the subprocess, register in `AsyncHookRegistry`, default timeout 15,000 ms, signal completion via first-line JSON `{"async": true, "asyncTimeout": N}` on stdout.
- `asyncRewake: true` flag: implies async; when the backgrounded subprocess exits with code 2, enqueue a task-notification that wakes the model for the next turn.
- `once: true` flag: self-remove from the session's active hook list after first successful (exit 0 or valid JSON decision) execution.
- `CLAUDE_ENV_FILE` mechanism: for SessionStart, Setup, CwdChanged, and FileChanged events, expose an env var `CLAUDE_ENV_FILE` pointing to a unique temp file; hook writes bash `export K=V` lines; subsequent BashTool calls source the file before running.
- Bookkeeping for `once`: `.moai/state/hook-once.json` with entries keyed by `{event, source_tier, dedup_key}`.
- `AsyncHookRegistry` in `internal/hook/async_registry.go` tracking in-flight goroutines with cancellation via `context.Context`.
- Task-notification queue abstraction in `internal/hook/notify.go` for `asyncRewake` wake events.

Out-of-scope:
- Hook type system (prompt/agent/http) → SPEC-V3-HOOKS-002.
- Matcher / if condition → SPEC-V3-HOOKS-004.
- Source precedence hierarchy → SPEC-V3-HOOKS-006.
- Full CC-compat prompt elicitation protocol (bidirectional stdin/stdout JSON dialogs) → deferred to post-v3.0 release.

## 3. Environment (환경)

Current moai-adk state:
- Hooks run synchronously with per-event timeout `DefaultHookTimeout = 30s` (findings-wave1-moai-current.md §5.1). No async primitive exists.
- No self-removing hooks; every declaration is active for the full session.
- No `CLAUDE_ENV_FILE` mechanism; env stays static across the session.
- No persistent `.moai/state/hook-once.json` bookkeeping file.

Claude Code reference:
- `utils/hooks.ts:995-1030` — `async: true` subprocess backgrounding (findings-wave1-hooks-commands.md §14.3).
- `utils/hooks.ts:205-245` — `asyncRewake` task-notification queue integration (findings-wave1-hooks-commands.md §14.3).
- `utils/hooks.ts:1112-1165` — first-line async JSON protocol (findings-wave1-hooks-commands.md §4.1).
- `utils/hooks/registerSkillHooks.ts:36-43` — `once` self-remove semantics (findings-wave1-hooks-commands.md §3.7).
- `utils/hooks.ts:917-926` — `CLAUDE_ENV_FILE` injection for 4 events (findings-wave1-hooks-commands.md §4.3).
- `utils/hooks/AsyncHookRegistry.ts` — registry implementation (findings-wave1-hooks-commands.md §0).

Affected modules:
- `internal/hook/async_registry.go` — new file.
- `internal/hook/once.go` — new file handling bookkeeping.
- `internal/hook/env_file.go` — new file for CLAUDE_ENV_FILE lifecycle.
- `internal/hook/registry.go` — dispatch path updates.
- `internal/hook/types.go` — add `Async`, `AsyncRewake`, `Once` boolean fields to HookDeclaration.
- `.moai/state/` — new persistent bookkeeping directory for once-hook state.

## 4. Assumptions (가정)

- Go's `context.Context` + `sync.WaitGroup` suffices to track async hook lifecycles without adding concurrency frameworks.
- Claude Code's "task-notification queue" for `asyncRewake` is an in-process channel that surfaces a wake signal on the next user turn; the moai-side abstraction mirrors this via a small in-memory event bus.
- `.moai/state/hook-once.json` does not need cross-machine coordination; it is session / project-local.
- `CLAUDE_ENV_FILE` is a platform-agnostic path under `os.TempDir()`; on Windows we still pass a POSIX-style path when the hook shell is bash, matching CC's existing behavior (findings-wave1-hooks-commands.md §4.3).
- The default async timeout 15,000 ms is correct per CC source; users may override via hook frontmatter `timeout` field.

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

- REQ-HOOKS-003-001: The `HookDeclaration` schema SHALL expose boolean fields `async`, `asyncRewake`, and `once`, all defaulting to false.
- REQ-HOOKS-003-002: The `AsyncHookRegistry` SHALL track every in-flight async hook by a unique `asyncID` with fields `{event, declarationKey, startedAt, deadline, cancel context.CancelFunc}`.
- REQ-HOOKS-003-003: The system SHALL export `CLAUDE_ENV_FILE` pointing to a unique temp file ONLY for events SessionStart, Setup, CwdChanged, and FileChanged.
- REQ-HOOKS-003-004: The once-hook bookkeeping SHALL persist across process restarts via `.moai/state/hook-once.json`.
- REQ-HOOKS-003-005: The once-hook dedup key SHALL be `{event}\0{source_tier}\0{dedup_key}` where `dedup_key` matches SPEC-V3-HOOKS-006 source-precedence convention.

### 5.2 Event-driven Requirements

- REQ-HOOKS-003-010: WHEN a hook declaration has `async: true` and the subprocess starts, the executor SHALL return control to the caller within 200 ms without waiting for subprocess exit.
- REQ-HOOKS-003-011: WHEN an async hook's stdout's first line is `{"async": true, "asyncTimeout": N}`, the registry SHALL apply timeout N ms (or 15,000 ms default) to the background goroutine.
- REQ-HOOKS-003-012: WHEN an `asyncRewake: true` hook exits with code 2, the system SHALL enqueue a task-notification that wakes the model on the next user turn with the hook's stderr as `additionalContext`.
- REQ-HOOKS-003-013: WHEN a hook with `once: true` completes successfully (exit 0 or valid JSON `continue: true`), the system SHALL append the dedup key to `.moai/state/hook-once.json` and SHALL NOT execute that hook again in the current or future sessions while the entry persists.
- REQ-HOOKS-003-014: WHEN a SessionStart / Setup / CwdChanged / FileChanged hook writes valid `export K=V` lines to its `CLAUDE_ENV_FILE`, the next BashTool invocation SHALL source the file before executing the user's command.

### 5.3 State-driven Requirements

- REQ-HOOKS-003-020: WHILE an async hook is running and the user's process receives SIGINT, the AsyncHookRegistry SHALL cancel the goroutine's `context.Context` and wait up to 2 seconds for graceful shutdown before SIGKILL.
- REQ-HOOKS-003-021: WHILE `.moai/state/hook-once.json` contains an entry for a given dedup key, the registry SHALL treat the associated hook declaration as inert regardless of source tier.

### 5.4 Optional Features

- REQ-HOOKS-003-030: WHERE the configuration key `hook.async.max_concurrent` is set in `.moai/config/sections/system.yaml`, the AsyncHookRegistry SHALL refuse to start a new async hook when the live count would exceed the limit and SHALL enqueue the request.
- REQ-HOOKS-003-031: WHERE the environment variable `MOAI_HOOK_CLEAR_ONCE=1` is set on process start, the registry SHALL truncate `.moai/state/hook-once.json` before loading declarations.

### 5.5 Complex Requirements

- REQ-HOOKS-003-040: IF an async hook writes invalid JSON to its first line AND `asyncRewake: true` is set, THEN the registry SHALL still treat the subprocess as async (backgrounded) with the 15,000 ms default timeout, and log a parse warning; ELSE (no asyncRewake) the hook degrades to synchronous behavior.
- REQ-HOOKS-003-041: IF a `CLAUDE_ENV_FILE` write contains lines that are not of the form `export KEY=VALUE`, THEN the system SHALL skip those lines with a trace-level warning and SHALL NOT fail the hook; ELSE valid exports are applied verbatim.

## 6. Acceptance Criteria (수용 기준 요약)

- AC-HOOKS-003-01: Given a PreToolUse hook with `async: true`, When the hook runs a 5-second sleep script, Then the tool call proceeds without waiting and the background goroutine is registered in `AsyncHookRegistry`. (maps REQ-HOOKS-003-002, REQ-HOOKS-003-010)
- AC-HOOKS-003-02: Given an async hook whose stdout first line is `{"async": true, "asyncTimeout": 5000}`, When registered, Then the background goroutine's deadline is now + 5s. (maps REQ-HOOKS-003-011)
- AC-HOOKS-003-03: Given a hook with `asyncRewake: true` that exits code 2, When it completes, Then a task-notification is enqueued and on the next user turn the model receives the stderr as `additionalContext`. (maps REQ-HOOKS-003-012)
- AC-HOOKS-003-04: Given a hook with `once: true` that completes successfully, When the same event fires again in the session, Then the hook is not re-executed and `.moai/state/hook-once.json` contains its dedup key. (maps REQ-HOOKS-003-013)
- AC-HOOKS-003-05: Given a SessionStart hook that writes `export API_KEY=xyz` to `$CLAUDE_ENV_FILE`, When a subsequent BashTool invocation runs, Then `API_KEY` is visible in the subprocess environment. (maps REQ-HOOKS-003-014)
- AC-HOOKS-003-06: Given a PostToolUse hook (not in the SessionStart/Setup/CwdChanged/FileChanged set), When it runs, Then `CLAUDE_ENV_FILE` is NOT set in its environment. (maps REQ-HOOKS-003-003)
- AC-HOOKS-003-07: Given `hook.async.max_concurrent: 2` and 2 async hooks already running, When a 3rd async hook is dispatched, Then the 3rd is enqueued and started only after one of the prior two completes. (maps REQ-HOOKS-003-030)
- AC-HOOKS-003-08: Given `MOAI_HOOK_CLEAR_ONCE=1`, When the process starts, Then `.moai/state/hook-once.json` is truncated before declarations load. (maps REQ-HOOKS-003-031)
- AC-HOOKS-003-09: Given an async hook running when SIGINT is received, When the signal fires, Then the registry cancels the context and waits up to 2 seconds before force-killing the subprocess. (maps REQ-HOOKS-003-020)

## 7. Constraints (제약)

- Technical: Go 1.22+. `os/exec` with `CommandContext` for cancellation. `sync.Map` for the live async-hook table.
- Backward compat: All new fields default to false; existing v2 hooks are unaffected.
- Platform: `CLAUDE_ENV_FILE` uses `os.TempDir()` + `os.CreateTemp()`. On Windows, file is created with forward-slash path when hook shell is bash (matching CC behavior).
- State file: `.moai/state/hook-once.json` is newline-delimited JSON entries to support append-only writes; readers must tolerate trailing whitespace.
- Performance: async-hook dispatch overhead ≤ 10 ms p99 beyond the existing synchronous overhead.

## 8. Risks & Mitigations (리스크 및 완화)

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Leaked goroutines accumulate over a long session | M | M | Every async hook tied to a `context.Context` with explicit deadline; `AsyncHookRegistry` exposes `ActiveCount()` for `moai doctor hook --async-status`. |
| `.moai/state/hook-once.json` grows unbounded | L | L | Cap to 10,000 entries with oldest-first eviction; size is monitored in `moai doctor`. |
| `CLAUDE_ENV_FILE` injects malicious env (e.g., `LD_PRELOAD`) | L | H | Env file only applied to subsequent BashTool subprocesses (already sandboxed by CC permission rules); we refuse writes containing newline-embedded shell metacharacters. |
| `asyncRewake` task-notification arrives after the session ends | L | M | Registry drops wake events for terminated sessions; logs one-line notice to trace. |
| Windows tempfile path has backslashes confusing bash hooks | L | M | Path translation through existing `windowsPathToPosixPath()` helper. |

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3-HOOKS-001 (async-protocol first-line JSON consumes HookOutput.Continue field conventions).

### 9.2 Blocks

- SPEC-V3-HOOKS-005 (FileChanged / CwdChanged handlers require CLAUDE_ENV_FILE mechanism).

### 9.3 Related

- SPEC-V3-HOOKS-004 (matcher resolution applies BEFORE async dispatch).
- SPEC-V3-HOOKS-006 (source precedence and dedup key influence once-hook bookkeeping).

## 10. Traceability (추적성)

- Theme: master-v3 Section 3.1 (Theme 1 — Hook Protocol v2) async/once sub-feature.
- Gap rows: gm#9 (High — async), gm#10 (High — asyncRewake), gm#11 (Medium — once), gm#20 (High — CLAUDE_ENV_FILE).
- BC-ID: None (purely additive).
- Wave 1 sources: findings-wave1-hooks-commands.md §2.1 (once flag), §4.1 (async wire format), §4.3 (CLAUDE_ENV_FILE), §14.3 (feature table), §0 (AsyncHookRegistry source anchor).
- Priority: P1 High (unlocks long-running quality loops; ships Phase 2).
