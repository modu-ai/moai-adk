---
id: SPEC-V3R6-HOOK-INPUT-SCHEMA-001
title: "Hook input schema robustness: globs array deserialization + empty-stdin graceful no-op"
version: "0.2.0"
status: completed
created: 2026-06-03
updated: 2026-06-03
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/hook"
lifecycle: spec-anchored
tags: "hook, schema, robustness, claude-code, stdin"
era: V3R6
tier: S
---

## HISTORY

- v0.1.0 (2026-06-03): Initial draft. Captures two confirmed `internal/hook` stdin-parse defects diagnosed from `~/.moai/logs/hook-stderr.log` — `globs` type mismatch (schema drift) and empty-stdin robustness gap.

## §A. Background and Diagnostic Evidence

The `internal/hook` package handles Claude Code hook events by reading a JSON payload from stdin and writing a structured response to stdout. Two distinct stdin-parse failures were captured in `~/.moai/logs/hook-stderr.log` this session, both verified against the current source.

### Defect 1 — `globs` type mismatch (schema drift, confirmed)

```
Error: read hook input: hook: invalid JSON input:
  json: cannot unmarshal array into Go struct field HookInput.globs of type string
  (moai hook instructions-loaded)
```

Root cause confirmed at `internal/hook/types.go:255`, in the InstructionsLoaded fields block (v2.1.69+):

```go
Globs string `json:"globs,omitempty"` // Glob patterns that triggered load
```

Claude Code's **InstructionsLoaded** hook event sends `globs` as a JSON **array** of glob-pattern strings (e.g. `["**/*.go", "**/*.md"]`), but the Go struct field is typed `string`. `json.Unmarshal` therefore fails, and `moai hook instructions-loaded` exits non-zero on every rule-file load. InstructionsLoaded fires on path-glob rule loading, so this failure is frequent.

A precedent for flexible typing already exists in the **same struct**: `ElicitationRequest json.RawMessage` (types.go:249) and `PermissionSuggestions json.RawMessage` (types.go:260) absorb shape-variable Claude Code fields without type-pinning the decoder.

Verified scoping fact: the `Globs` field currently has **no read consumer** anywhere in the package (`grep -rn '\.Globs' internal/hook/ --include='*.go'` returns zero matches outside the struct declaration). There is no string-form consumer to regress, so the compatibility requirement is satisfied by any chosen type.

### Defect 2 — empty/truncated stdin robustness gap (confirmed)

```
Error: read hook input: hook: invalid JSON input: unexpected end of JSON input
  (moai hook pre-tool)
```

The shared hook-input reader `ReadInput` (`internal/hook/protocol.go:26`) raises `ErrHookInvalidInput` (`internal/hook/errors.go:16`) when stdin is empty or truncated: `io.ReadAll` returns an empty byte slice, `normalizeHookInput`/`json.Unmarshal` then fail, and the wrapped error propagates. The central dispatcher `runHookEvent` (`internal/cli/hook.go:173-176`) returns this error to cobra, producing a **non-zero exit code**.

The PreToolUse hook matches `Write|Edit|Bash`. When `moai hook pre-tool` receives empty stdin and exits non-zero, the harness surfaces it as a **Bash tool execution failure** (`PostToolUseFailure: Bash says: UnknownFailure: Tool execution failed`) — the user-visible symptom.

Design alignment: `validateInput` (protocol.go:75-101) already establishes a graceful-fallback philosophy — missing `session_id`, `cwd`, and `hook_event_name` are defaulted rather than treated as fatal, because "a graceful fallback is preferable to hook execution failure." Treating empty stdin as a no-op success extends that existing philosophy to its logical conclusion: a non-blocking observer hook must never fail the tool it observes.

## §B. Out of Scope (What NOT to Build)

### Out of Scope — other hook events not exhibiting this failure

- Only the two captured failure paths (`globs` field decode + empty-stdin read) are in scope. No audit or hardening of other `HookInput` fields or other event handlers is performed.

### Out of Scope — settings.json hook registration

- No change to `.claude/settings.json` hook configuration. The bug is in the Go handler decode path, not in how hooks are registered or matched.

### Out of Scope — handle-*.sh wrapper scripts

- The `.claude/hooks/moai/handle-*.sh` wrappers already redirect stderr correctly. They are not the source of either defect and are left unchanged.

### Out of Scope — broad flexible-typing refactor

- This SPEC does not convert every shape-variable field to a custom decoder. It applies the minimum mechanism to the one field (`Globs`) and the one reader path (`ReadInput`) confirmed in the diagnostic evidence.

## §C. Requirements (GEARS)

- **REQ-HIS-001 (Defect 1 — globs array)**: When the InstructionsLoaded event sends `globs` as a JSON array of glob-pattern strings, the hook input decoder shall deserialize `HookInput.globs` without error, while preserving compatibility with any string-form consumer of `Globs`.

- **REQ-HIS-002 (Defect 2 — empty-stdin no-op)**: When the hook input reader receives empty, blank, or whitespace-only stdin, the reader shall return a no-op success (the caller exits 0) and shall not propagate a parse error that fails the wrapped tool (PreToolUse on Bash).

- **REQ-HIS-003 (regression — well-formed input unchanged)**: Where a hook handler already parses well-formed JSON, the decoder and reader shall remain behaviorally unchanged — no field value, error, or exit-code difference for valid input.

## §3. Acceptance Criteria (Tier S — inline)

- **AC-1 (REQ-HIS-001)**: A unit test feeds `instructions-loaded` (or directly decodes a `HookInput`) with `globs` as a JSON array `["**/*.go", "**/*.md"]` → decode succeeds with no `cannot unmarshal array` error, and `Globs` is populated with both patterns.

- **AC-2 (REQ-HIS-002)**: A unit test feeds the hook reader (`ReadInput`) empty stdin (`""`) → returns the no-op success path (a usable zero-value/default `*HookInput`, nil error), NOT `ErrHookInvalidInput`. A blank/whitespace-only variant (`"   \n"`) behaves identically.

- **AC-3 (REQ-HIS-003, regression)**: Existing `internal/hook` tests stay green (`go test ./internal/hook/...`), **EXCEPT the two empty/whitespace-stdin cases at `protocol_test.go:183-192`, which are intentionally inverted by AC-2** (`wantErr: true → false`, assert the returned `*HookInput` is a non-nil zero-value instead of `ErrHookInvalidInput`). A well-formed-input decode test confirms field values are unchanged from current behavior.

- **AC-4 (REQ-HIS-002, CLI smoke — mandatory)**: `moai hook pre-tool </dev/null` exits 0. The empty-stdin `ReadInput` change directly governs the `pre-tool` exit path — the exact user-visible symptom — so this smoke check is required, not optional.

## §D. Traceability

| REQ | AC | Source evidence | Fix site |
|-----|-----|-----------------|----------|
| REQ-HIS-001 | AC-1 | types.go:255 (`Globs string`); precedent types.go:249/260 | `internal/hook/types.go` |
| REQ-HIS-002 | AC-2 (mandatory) + AC-4 (mandatory) | protocol.go:26 (`ReadInput`); errors.go:16 (`ErrHookInvalidInput`); hook.go:173-176 (propagation) | `internal/hook/protocol.go` |
| REQ-HIS-003 | AC-3 (excludes the 2 inverted cases at protocol_test.go:183-192) | validateInput graceful-fallback precedent (protocol.go:75-101) | both fix sites |
