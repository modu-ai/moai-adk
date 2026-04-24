---
id: SPEC-V3-HOOKS-001
title: "Hook Protocol v2 — Rich JSON IO"
version: "0.1.0"
status: draft
created: 2026-04-22
updated: 2026-04-22
author: GOOS
priority: P0 Critical
phase: "v3.0.0 — Phase 2 Hook Protocol v2 Core"
module: "internal/hook/"
dependencies:
  - SPEC-V3-SCH-001
related_gap:
  - gm#4
  - gm#5
  - gm#6
  - gm#7
related_theme: "Theme 1: Hook Protocol v2"
breaking: true
bc_id: [BC-001]
lifecycle: spec-anchored
tags: "hook, v3, breaking, protocol"
---

# SPEC-V3-HOOKS-001: Hook Protocol v2 — Rich JSON IO

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-22 | GOOS | Initial v3 draft from Wave 4 bundle (Hooks/Commands) |

---

## 1. Goal (목적)

Elevate moai-adk-go hook handlers from exit-code-only semantics to Claude Code's rich JSON output contract (`HookJSONOutput`). This unlocks the ability to inject model-turn context (`additionalContext`), rewrite tool input before execution (`updatedInput`), rewrite MCP tool output (`updatedMCPToolOutput`), control flow (`continue`/`stopReason`/`systemMessage`), and emit event-specific payloads via a discriminated union (`hookSpecificOutput`). Without this foundation, every higher-value Hook theme (permission decisions, watchPaths, condition filters) is blocked.

gap-matrix rows gm#4 (Critical), gm#5 (Critical), gm#6 (Critical), gm#7 (Medium) all collapse onto this SPEC.

## 2. Scope (범위)

In-scope:
- Extend `internal/hook/types.go` `HookOutput` struct to match CC's `HookJSONOutput` shape with every top-level field (continue, stopReason, systemMessage, suppressOutput, decision, reason, hookSpecificOutput).
- Introduce `HookSpecificOutput` as a Go interface implemented by per-event concrete types (discriminated by `hookEventName`).
- Ship concrete variants required in v3.0: PreToolUseOutput, PostToolUseOutput, PostToolUseFailureOutput, UserPromptSubmitOutput, SessionStartOutput (with `watchPaths` + `initialUserMessage`), SessionEndOutput, StopOutput, SubagentStopOutput, PreCompactOutput (with `newCustomInstructions` + `userDisplayMessage`), PostCompactOutput, NotificationOutput, TaskCompletedOutput, TeammateIdleOutput, WorktreeRemoveOutput, ConfigChangeOutput, InstructionsLoadedOutput.
- Add wire-protocol helpers: JSON marshal/unmarshal with `hookEventName` tag, strict schema validation via `go-playground/validator/v10` (per SPEC-V3-SCH-001).
- Introduce dual-parse shim: try JSON parse first, fall back to exit-code semantics when stdout is empty or not valid JSON. `ExitCode` field is synthesized from `Decision`.
- Emit deprecation warning (via `systemMessage`) when a hook returns exit-code-only output, referencing migration doc URL.
- Provide `MOAI_HOOK_LEGACY=1` opt-out env for CI / air-gapped installs.

Out-of-scope (deferred to other SPECs):
- Hook `if` condition filter and matcher upgrade → SPEC-V3-HOOKS-004.
- Hook source precedence 3-tier merge → SPEC-V3-HOOKS-006.
- Async / once / CLAUDE_ENV_FILE mechanics → SPEC-V3-HOOKS-003.
- Hook type: prompt/agent/http → SPEC-V3-HOOKS-002.
- PermissionRequest / PermissionDenied decision semantics → SPEC-V3-HOOKS-004/005 companions.

## 3. Environment (환경)

Current moai-adk state:
- `internal/hook/types.go:19-114` defines 27 `EventType` constants (findings-wave1-moai-current.md §5.1).
- `internal/hook/types.go:167-311` (per findings-wave1-moai-current.md §5.7) defines a partial `HookOutput` with `Continue`, `StopReason`, `SystemMessage`, `SuppressOutput`, `Decision`, `Reason`, `HookSpecificOutput` (map), `UpdatedInput`, `Retry`, `ExitCode`. Missing: `AdditionalContext`, `UpdatedMCPToolOutput`, `WatchPaths`, `InitialUserMessage`, `NewCustomInstructions`, `UserDisplayMessage` as typed fields; `HookSpecificOutput` is an untyped map instead of a discriminated union.
- 26 shell wrappers in `.claude/hooks/moai/` all emit exit codes only; no wrapper writes structured JSON to stdout today.
- `internal/hook/registry.go:18-325` dispatches handlers by `EventType` (findings-wave1-moai-current.md §5.1). No current code path consumes the richer JSON fields.

Claude Code reference:
- `entrypoints/sdk/coreSchemas.ts:806-935` defines `HookJSONOutput` as discriminated union keyed on `hookEventName` (findings-wave1-hooks-commands.md §4.1).
- `types/hooks.ts:29-166` defines the full wire format (findings-wave1-hooks-commands.md §4.1).
- `utils/hooks.ts:622-628, 644-652` injects `additionalContext` into model turn; `utils/hooks.ts:618-620, 668-672` applies `updatedInput` (findings-wave1-hooks-commands.md §14.3).
- `utils/hooks.ts:645-649` handles `updatedMCPToolOutput` rewrite (findings-wave1-hooks-commands.md §14.3).

Affected modules:
- `internal/hook/types.go` — struct expansion.
- `internal/hook/protocol.go` — new file: dual-parse reader.
- `internal/hook/registry.go` — downstream consumption of rich fields.
- `.claude/hooks/moai/*.sh` — optional wrapper rewrite to emit JSON (not strictly required in v3.0 due to dual-parse).

## 4. Assumptions (가정)

- CC-native HookJSONOutput schema at `entrypoints/sdk/coreSchemas.ts:806-935` is stable for v2.1.111+ and will remain the reference for v3.0.
- go-playground/validator/v10 is available (provided by SPEC-V3-SCH-001); discriminated-union validation uses `oneof` tag on `hookEventName` per variant.
- All existing 26 shell hook wrappers emit valid legacy exit-code-only output today; dual-parse fallback is the migration mechanism.
- External hook authors (plugin hooks, custom scripts) exist in small numbers; grace period via legacy shim is sufficient.
- Go's `encoding/json` with `json.RawMessage` handles the `hookSpecificOutput` discrimination in the dual-pass decode pattern.

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

- REQ-HOOKS-001-001: The `HookOutput` struct SHALL expose every field defined in CC's `HookJSONOutput` schema as typed Go fields, including `Continue *bool`, `StopReason string`, `SystemMessage string`, `SuppressOutput bool`, `Decision string`, `Reason string`, `HookSpecificOutput HookSpecificOutput`.
- REQ-HOOKS-001-002: The `HookSpecificOutput` Go interface SHALL expose a `HookEventName() string` method implemented by every concrete per-event variant struct.
- REQ-HOOKS-001-003: The system SHALL provide concrete variant structs for at least the following events in v3.0: PreToolUse, PostToolUse, PostToolUseFailure, UserPromptSubmit, SessionStart, SessionEnd, Stop, SubagentStop, PreCompact, PostCompact, Notification, TaskCompleted, TeammateIdle, WorktreeRemove, ConfigChange, InstructionsLoaded.
- REQ-HOOKS-001-004: The `SessionStartOutput` variant SHALL expose `WatchPaths []string`, `AdditionalContext string`, and `InitialUserMessage string` fields with JSON tags matching CC's schema.
- REQ-HOOKS-001-005: The `PreCompactOutput` variant SHALL expose `NewCustomInstructions string` and `UserDisplayMessage string` fields.
- REQ-HOOKS-001-006: The `PostToolUseOutput` variant SHALL expose `UpdatedMCPToolOutput json.RawMessage` field preserving arbitrary shape.
- REQ-HOOKS-001-007: The system SHALL validate every received `HookOutput` using go-playground/validator/v10 schema tags before enqueueing downstream effects.

### 5.2 Event-driven Requirements

- REQ-HOOKS-001-010: WHEN a hook wrapper writes a JSON object to stdout, the protocol parser SHALL attempt `encoding/json` unmarshal into `HookOutput` first; on success the exit-code path is bypassed.
- REQ-HOOKS-001-011: WHEN JSON parse fails or stdout is empty, the protocol parser SHALL fall back to exit-code semantics (0 = allow/continue, 2 = block with stderr as reason, other = user-visible stderr) and synthesize `Decision` accordingly.
- REQ-HOOKS-001-012: WHEN a hook returns `HookSpecificOutput` whose `hookEventName` does not match the dispatched event, the system SHALL reject the output with a `HookSpecificOutputMismatch` error and log the mismatch to trace.
- REQ-HOOKS-001-013: WHEN a hook returns `AdditionalContext` on an event that supports context injection, the system SHALL forward the string to the model turn context assembly path.
- REQ-HOOKS-001-014: WHEN a hook returns `UpdatedInput` for PreToolUse, the system SHALL replace the tool-call input with the rewritten value before executing the tool.

### 5.3 State-driven Requirements

- REQ-HOOKS-001-020: WHILE the environment variable `MOAI_HOOK_LEGACY=1` is set, the system SHALL suppress the dual-parse deprecation warning banner while continuing to accept both legacy and v2 outputs.
- REQ-HOOKS-001-021: WHILE a hook wrapper emits only exit-code output (no JSON on stdout) and `MOAI_HOOK_LEGACY` is unset, the system SHALL emit a one-time-per-session `systemMessage` pointing to the migration document URL.

### 5.4 Optional Features

- REQ-HOOKS-001-030: WHERE the configuration flag `hook.strict_mode: true` is set in `.moai/config/sections/system.yaml`, the system SHALL refuse exit-code-only output entirely and return `HookProtocolLegacyRejected` error.
- REQ-HOOKS-001-031: WHERE a hook wrapper opts into `api_version: 2` frontmatter declaration, the system SHALL skip the dual-parse fallback for that wrapper even in non-strict mode.

### 5.5 Complex Requirements

- REQ-HOOKS-001-040: IF a hook output contains both the legacy `decision` field and a `hookSpecificOutput` with a conflicting `permissionDecision`, THEN the system SHALL prefer `hookSpecificOutput.permissionDecision` and log the conflict; ELSE the legacy `decision` path is preserved for backward compatibility.
- REQ-HOOKS-001-041: IF a hook output contains `Continue: false` and `StopReason` is empty, THEN the system SHALL use a default message "Hook requested stop (no reason supplied)" before propagating halt; ELSE the provided `StopReason` is used verbatim.

## 6. Acceptance Criteria (수용 기준 요약)

- AC-HOOKS-001-01: Given a PreToolUse hook wrapper writing `{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow","updatedInput":{"file_path":"/tmp/x"}}}` to stdout, When the hook runs, Then the tool input is rewritten to `{"file_path":"/tmp/x"}` before execution. (maps REQ-HOOKS-001-014)
- AC-HOOKS-001-02: Given a SessionStart hook wrapper returning `{"hookSpecificOutput":{"hookEventName":"SessionStart","watchPaths":["/abs/.env"],"additionalContext":"ctx"}}`, When parsed, Then `HookOutput.HookSpecificOutput` is typed as `SessionStartOutput` and `WatchPaths` contains the absolute path. (maps REQ-HOOKS-001-004, REQ-HOOKS-001-010)
- AC-HOOKS-001-03: Given a legacy hook wrapper exiting with code 2 and stderr "blocked", When dual-parse falls back, Then `HookOutput.Decision == "deny"` and `HookOutput.Reason == "blocked"`. (maps REQ-HOOKS-001-011)
- AC-HOOKS-001-04: Given `MOAI_HOOK_LEGACY=1`, When a legacy wrapper runs, Then no `systemMessage` is emitted. (maps REQ-HOOKS-001-020)
- AC-HOOKS-001-05: Given a PreToolUse hook returns `hookSpecificOutput.hookEventName == "PostToolUse"` (mismatch), When parsed, Then the protocol layer returns `HookSpecificOutputMismatch` error and the hook is treated as failed. (maps REQ-HOOKS-001-012)
- AC-HOOKS-001-06: Given `hook.strict_mode: true`, When a legacy hook emits only an exit code, Then the system returns `HookProtocolLegacyRejected` and halts. (maps REQ-HOOKS-001-030)
- AC-HOOKS-001-07: Given a hook output with `decision: "block"` and `hookSpecificOutput.permissionDecision: "allow"`, When resolved, Then `permissionDecision` wins and the conflict is written to trace log. (maps REQ-HOOKS-001-040)
- AC-HOOKS-001-08: Given `HookOutput` validated via validator/v10, When any concrete variant has an invalid `hookEventName` discriminator, Then `Validate()` returns a non-nil error. (maps REQ-HOOKS-001-007)

## 7. Constraints (제약)

- Technical: Go 1.22+. No new top-level dependencies (validator/v10 enters via SPEC-V3-SCH-001; `encoding/json` from stdlib suffices).
- Backward compat: v2 exit-code hooks MUST continue to function during the v3.0 → v3.2 deprecation window (BC-001 in master-v3 Section 4); exit-code fallback is removed entirely in v4.0.
- Platform: macOS / Linux / Windows. On Windows, shell paths pass through existing `windowsPathToPosixPath()` translator before invocation (no new behavior required here).
- Performance: JSON parse of hook stdout MUST not add more than 5ms p99 overhead on top of subprocess wall time for payloads up to 64KB.
- Binary size: struct additions MUST NOT grow `bin/moai` by more than 200KB.

## 8. Risks & Mitigations (리스크 및 완화)

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| External hook authors emit invalid JSON and rely on exit-code fallback forever | M | M | Dual-parse spans v3.0 through v3.2 with warn phase; removal at v4.0 gives two minor versions of notice. `MOAI_HOOK_LEGACY=1` provides indefinite escape for CI/air-gapped installs. |
| Discriminated union mismatches produce cryptic errors | M | M | REQ-HOOKS-001-012 mandates specific `HookSpecificOutputMismatch` error and trace logging. `moai doctor hook --validate` surfaces these. |
| Large `updatedInput` or `additionalContext` payloads balloon token usage | L | M | Document recommended payload caps in docs-site migration guide; add budget check in SPEC-V3-HOOKS-004 (handler richness upgrade) for events that accept large payloads. |
| legacy/JSON both present with conflicting fields confuses users | L | L | REQ-HOOKS-001-040 defines a deterministic precedence; trace log surfaces conflicts for debugging. |

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3-SCH-001 (provides go-playground/validator/v10 for struct validation).

### 9.2 Blocks

- SPEC-V3-HOOKS-002 (Hook type system extends HookOutput consumers to non-command types).
- SPEC-V3-HOOKS-003 (Async/once mechanics rely on `continue`/`async:true` first-line protocol which is defined here).
- SPEC-V3-HOOKS-004 (Matcher/if condition filter emits decisions via HookSpecificOutput).
- SPEC-V3-HOOKS-005 (Missing event handlers use concrete output variants defined here).
- SPEC-V3-HOOKS-006 (Scoping hierarchy surfaces dedup via HookSpecificOutput for source-tagged outputs).

### 9.3 Related

- SPEC-V3-CMDS-001 (command frontmatter extensions may emit hook-adjacent metadata that consumes typed output).
- SPEC-V3-PLG-001 (plugin manifests declare hooks that must emit protocol v2 output).

## 10. Traceability (추적성)

- Theme: master-v3 Section 3.1 (Theme 1 — Hook Protocol v2), specifically the `HookJSONOutput` alignment in the API/schema sketch.
- Gap rows: gm#4 (Critical — JSON output protocol), gm#5 (Critical — additionalContext), gm#6 (Critical — updatedInput), gm#7 (Medium — updatedMCPToolOutput).
- BC-ID: BC-001 (Hook output protocol — exit-code-only deprecated).
- Wave 1 sources: findings-wave1-hooks-commands.md §4.1 (wire format), §14.3 (feature gap table), §12 (source references); findings-wave1-moai-current.md §5.1 (registry), §5.7 (current HookOutput partial shape).
- Priority: P0 Critical (blocks all other Hook SPECs in v3.0 Phase 2).
