---
id: SPEC-HOOK-FAILURE-CLASSIFY-001
title: "PostToolUseFailure nested-error classification and session_id trace resolution"
version: "0.1.0"
status: completed
created: 2026-07-17
updated: 2026-07-21
author: manager-spec
priority: P1
phase: "v3.0.0 target"
module: "internal/hook"
lifecycle: spec-anchored
tags: "hook, post-tool-failure, classification, observability, bugfix"
issue_number: 1089
tier: S
---

# SPEC-HOOK-FAILURE-CLASSIFY-001 — PostToolUseFailure nested-error classification

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-17 | manager-spec | Initial draft — fixes GitHub issue #1089 (UnknownFailure always + trace-unknown.jsonl) |

## §A Context / Why

GitHub issue #1089 (moai-adk-go v3.0.0-rc12): `moai hook post-tool-failure` **always** classifies real tool failures as `UnknownFailure` and writes the trace to `trace-unknown.jsonl`. Both symptoms were mechanically reproduced by the orchestrator at HEAD `efd423b88`.

Ground-truth root causes (verified against current source):

1. **Misclassification.** `internal/hook/post_tool_failure.go` `classifyError` (lines 87-122) lowercases and substring-matches **only** the top-level `input.Error` field. The actual Claude Code `PostToolUseFailure` stdin payload nests failure text under `tool_response.error` / `tool_response.stderr` (`input.ToolResponse`, `internal/hook/types.go:220`, a `json.RawMessage`). `input.ToolResponse` is never parsed by the classifier, so every real payload arrives with an empty `input.Error`, falls through every matcher, and returns `UnknownFailure`. Reproduction: a nested-error payload yields `{"systemMessage":"UnknownFailure: ..."}`, while a synthetic top-level `"error":"exit status 1"` correctly yields `ExitError` — proving the matcher works only at the top level.

2. **`trace-unknown.jsonl`.** The trace filename is `trace-<sessionID>.jsonl` (`internal/hook/trace/writer.go:99`), and `sessionID` flows from `input.SessionID` (`internal/hook/registry.go:341-350`). `validateInput` (`internal/hook/protocol.go:110-112`) substitutes `input.SessionID = "unknown"` whenever `session_id` is empty. Per the existing comment (`protocol.go:93-95`), `PostToolUse`/`PostToolUseFailure` payloads may omit `session_id` (Claude Code bug #541), so the substitution fires and the trace is written to `trace-unknown.jsonl`. The end-to-end path was NOT verified by the audit — the resolution mechanism is a run-phase investigation (evidence pending).

WHY it matters: the failure handler feeds the harness learning loop (`recordToolFailureEvent`, `tool_failure:<tool>:<category>`) and surfaces an actionable `systemMessage` to the model. A universally-`UnknownFailure` signal poisons the learning loop with a single degenerate category and gives the model no actionable guidance. A single shared `trace-unknown.jsonl` collides across all session-id-less events, corrupting per-session observability.

## §B Requirements (GEARS)

### REQ-HFC-001 — Nested-error classification input (Event-driven)

**When** a `PostToolUseFailure` payload carries failure text under `tool_response.error` and/or `tool_response.stderr`, the failure handler **shall** derive its classification input from those nested `tool_response` fields (in addition to the top-level `error` field), so that a realistic nested payload classifies to its correct `ErrorCategory` rather than `UnknownFailure`.

### REQ-HFC-002 — Top-level precedence and backward compatibility (Ubiquitous)

The failure handler **shall** continue to honor a non-empty top-level `error` field, classifying such a payload identically to the pre-fix behavior. **Where** both the top-level `error` field and nested `tool_response` error text are present, the failure handler **shall** apply a defined precedence: the top-level `error` field takes precedence for the human-facing error excerpt, while classification considers all available sources so no correct category is lost.

### REQ-HFC-003 — Per-category nested regression tests (Ubiquitous)

The regression test suite **shall** include realistic nested `tool_response` payload cases covering each error category (`TimeoutError`, `PermissionDenied`, `ContextCancelled`, `SandboxViolation`, `OOMKilled`, `ExitError`, and a genuinely unclassifiable case), asserting the classified category and the presence of a non-degenerate `systemMessage`.

### REQ-HFC-004 — session_id resolution for the trace filename (Event-driven)

**When** a `PostToolUseFailure` (or any hook) payload omits `session_id`, the failure/trace subsystem **shall** resolve a session identifier from an available payload source, or fall back to a documented deterministic identifier, such that a resolvable session is NOT written to the shared `trace-unknown.jsonl`. The concrete resolution mechanism (e.g. deriving the session UUID from `transcript_path`) **shall** be investigated during run-phase and the chosen source recorded; **when** no source is resolvable, the documented last-resort fallback applies.

### REQ-HFC-005 — No content-free UnknownFailure message (unwanted behavior)

**When** a failure is genuinely unclassifiable, the failure handler **shall not** emit a content-free `UnknownFailure` `systemMessage`; it **shall** include a bounded excerpt of the observed raw error text (drawn from the top-level `error` field with precedence, else the nested `tool_response` text) so the model receives an actionable signal.

### REQ-HFC-006 — tool_response shape resilience (capability gate / fail-open)

**Where** `tool_response` is a wrapped JSON object, a bare JSON string, malformed, or absent, the failure handler **shall** degrade gracefully — never panic and never return a handler error — falling back to whatever error text is available (fail-open, consistent with the existing `recordToolFailureEvent` fail-open contract).

## §C Exclusions

This SPEC fixes classification-input sourcing and the trace session_id for the `PostToolUseFailure` handler only. The following are explicitly out of scope.

### Out of Scope — new error categories

- No new `ErrorCategory` constants are introduced. The seven existing categories (`TimeoutError`, `PermissionDenied`, `ContextCancelled`, `SandboxViolation`, `OOMKilled`, `ExitError`, `UnknownFailure`) are the complete set.
- No change to the substring-matching heuristics themselves (which token maps to which category); only the **input** the matchers see is broadened.

### Out of Scope — harness learning-loop semantics

- No change to `recordToolFailureEvent`, the `tool_failure:<tool>:<category>` event key format, or `usage-log.jsonl` recording. The category flows into the learning loop unchanged; only its accuracy improves as a side effect.

### Out of Scope — trace subsystem redesign

- No change to the async `TraceWriter`, rotation, channel-drop degradation, or the `trace-<sessionID>.jsonl` naming scheme. Only the value bound to `sessionID` is corrected.
- The global `validateInput` `session_id = "unknown"` substitution for OTHER events is not redesigned here; the fix targets the trace-filename resolution path.

### Out of Scope — other hook events

- Only `PostToolUseFailure` classification and its trace filename are addressed. `PostToolUse`, `Stop`, `StopFailure`, and other events are untouched.

## §D Acceptance Criteria (inline — Tier S LEAN)

Each AC is observable and independently verifiable. Given/When/Then.

- **AC-HFC-001a** (REQ-HFC-001): **Given** a `PostToolUseFailure` `HookInput` with an empty top-level `Error` and `tool_response` = `{"error":"permission denied: open /f","stderr":""}`, **When** `Handle` runs, **Then** the classified category is `PermissionDenied` (not `UnknownFailure`) and the `systemMessage` has the `PermissionDenied:` prefix.
- **AC-HFC-001b** (REQ-HFC-001): **Given** a `PostToolUseFailure` payload with empty top-level `Error` and `tool_response` = `{"stderr":"...context deadline exceeded..."}`, **When** `Handle` runs, **Then** the classified category is `TimeoutError`.
- **AC-HFC-002** (REQ-HFC-002): **Given** the existing `post_tool_failure_test.go` table cases that set the top-level `Error` field, **When** the suite runs after the fix, **Then** every pre-existing case still classifies to its original expected category (zero regression) — the existing test file passes unmodified in its top-level-Error assertions.
- **AC-HFC-003** (REQ-HFC-003): **Given** the regression suite, **When** `go test ./internal/hook/...` runs, **Then** the suite contains at least one nested-`tool_response` case for each of the seven categories, each asserting the classified category. Verify: `grep -c 'tool_response\|ToolResponse' internal/hook/post_tool_failure_test.go` returns ≥ 7 nested-payload cases and the test package passes.
- **AC-HFC-004** (REQ-HFC-004): **Given** a payload that omits `session_id` but carries a resolvable source (per the run-phase-chosen mechanism), **When** the trace filename is computed, **Then** it is NOT `trace-unknown.jsonl`; **and** the chosen resolution source and the last-resort fallback are documented in code and in `progress.md` §E.2. A payload with a present `session_id` still produces `trace-<session_id>.jsonl` (no regression).
- **AC-HFC-005** (REQ-HFC-005): **Given** a genuinely unclassifiable payload (top-level `Error` empty, `tool_response` = `{"error":"something went wrong"}`), **When** `Handle` runs, **Then** the `systemMessage` is `UnknownFailure`-classified BUT is NOT the exact content-free string `"UnknownFailure: Tool execution failed. Review error logs for details."` — it contains a bounded excerpt of `something went wrong`.
- **AC-HFC-006** (REQ-HFC-006): **Given** malformed `tool_response` (e.g. `[}`), a bare JSON string, and an absent `tool_response`, **When** `Handle` runs for each, **Then** it returns no error, does not panic, and produces a `systemMessage` (fail-open). Verify with a table-driven resilience test.
- **AC-HFC-GATE** (quality gate): **Given** the full package, **When** `go test ./internal/hook/... && go vet ./internal/hook/... && golangci-lint run internal/hook/...` runs, **Then** all pass with zero errors and the subagent-boundary grep (`grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ | grep -v _test.go`) returns 0 matches.

## §E Definition of Done

- All six REQs implemented; all AC-HFC-* observable checks pass with cited command output.
- No new `ErrorCategory`; no change to learning-loop event format.
- Coverage for `internal/hook` classification path ≥ 85% (per-package target).
- `progress.md` §E.2 records the session_id resolution investigation outcome (chosen source + fallback).
