---
id: SPEC-HOOK-FAILURE-CLASSIFY-001
title: "Implementation plan — PostToolUseFailure nested-error classification"
version: "0.1.0"
status: in-progress
created: 2026-07-17
updated: 2026-07-21
author: manager-spec
tier: S
---

# Plan — SPEC-HOOK-FAILURE-CLASSIFY-001

## §A Context

Single-package fix in `internal/hook`. Tier S (LEAN): ≤5 files, no cross-package surface. Development mode: TDD (RED-GREEN-REFACTOR) is the natural fit — the bug is reproducible as a failing test, and every REQ has a mechanically verifiable assertion (Reproduction-First, CLAUDE.md §7 Rule 4).

Anchors (verify at run-phase entry — line numbers drift, prefer content tokens):
- `internal/hook/post_tool_failure.go` — `classifyError` (reads only `input.Error`), `formatMessage` (content-free `UnknownFailure` default).
- `internal/hook/types.go` — `HookInput.ToolResponse json.RawMessage` (`tool_response`).
- `internal/hook/evidence_writer.go` — `decodeToolResponse(result []byte) (text string, isObject bool)` + `textKeys` family (`stdout`/`stderr`/`output`/`content`/`result`). **Reuse candidate** (Simplicity ladder step 2): it already normalizes wrapped-object vs plain-text `tool_response` into text.
- `internal/hook/protocol.go` — `validateInput` sets `input.SessionID = "unknown"` on empty `session_id`.
- `internal/hook/trace/writer.go` — `filePath()` = `trace-<sessionID>.jsonl`; `registry.go` `ensureTraceWriter` binds `sessionID` from `input.SessionID`.
- `internal/hook/post_tool_failure_test.go` — existing top-level-`Error` table (must keep passing = backward-compat gate AC-HFC-002).

## §B Known Issues / Risks

- **R1 — `tool_response` shape variance.** Live Bash `tool_response` is a wrapped object; some tools may send a bare JSON string or nest `error` under a sub-object. Mitigation: reuse `decodeToolResponse` (already handles object/plain-text) and aggregate; REQ-HFC-006 resilience table guards malformed/absent shapes. Cross-ref: boundary-verification.md Case 6/7 (async/generic shape).
- **R2 — session_id genuinely absent.** If neither `session_id` nor `transcript_path` resolves a UUID, the fix cannot invent one. Mitigation: REQ-HFC-004 permits a documented last-resort fallback; the win is not-colliding for RESOLVABLE sessions, not eliminating "unknown" entirely.
- **R3 — precedence ambiguity.** Aggregating all error sources for classification could let a stray token in one field mis-steer the category. Mitigation: keep the existing ordered matcher (timeout→permission→context→sandbox→oom→exit→unknown); aggregation only widens the haystack, and the ordered first-match semantics are unchanged.

## §C Pre-flight

- `git rev-parse --short HEAD` (baseline anchor; audit ran at `efd423b88`).
- `go test ./internal/hook/...` green baseline before any change.
- Re-verify the five anchors above by content token (not line number).

## §D Constraints

- No new `ErrorCategory`; no learning-loop event-format change; no `TraceWriter` redesign (§C exclusions).
- Fail-open: handler never returns an error for a malformed payload (matches `recordToolFailureEvent` contract).
- Subagent boundary: no `AskUserQuestion`/`mcp__askuser` in `internal/hook` (C-HRA-008).
- ≤5s hook budget — parsing is O(payload size), well within budget.

## §F Milestones (ordered by decision-reversibility — highest-change-likelihood first)

### M1 — Classification-input sourcing + precedence (highest change likelihood: semantics)
The core design decision. Broaden `classifyError` to build its lowercased haystack from top-level `input.Error` **plus** the normalized `tool_response` text (via reused `decodeToolResponse`). Define precedence: top-level `Error` wins for the displayed excerpt; classification sees the union. RED: add AC-HFC-001a/001b nested cases (currently → `UnknownFailure`). GREEN: implement aggregation. Keep the ordered matcher untouched. This milestone is where human review should focus — it changes observable classification semantics.

### M2 — session_id resolution for the trace filename (behavioral / observability-facing; investigation)
Investigate the root cause end-to-end (audit left this evidence-pending). Determine whether `transcript_path` (present in the payload, `types.go:211`) encodes the session UUID and can seed `sessionID` when `session_id` is empty, BEFORE `validateInput` substitutes `"unknown"`. Choose the resolution source, implement it at the trace-filename binding path, and record the chosen source + last-resort fallback in `progress.md` §E.2. AC-HFC-004. Reversible-risky because it changes observable trace filenames and depends on an investigated external payload shape.

### M3 — Unclassifiable message policy (user-facing message content)
Change `formatMessage` so the `UnknownFailure` branch appends a bounded excerpt (e.g. first ~200 chars) of the observed raw error (top-level precedence, else nested) instead of the fixed content-free string. Decision recorded: emit-with-excerpt (NOT emit-nothing) — the `systemMessage` is the sole model-facing failure surface; emitting nothing regresses the hook's purpose. AC-HFC-005.

### M4 — Regression + resilience test suite (mechanical)
Add per-category nested-payload cases (AC-HFC-003) and the malformed/bare-string/absent resilience table (AC-HFC-006). Preserve all existing top-level-`Error` cases verbatim (AC-HFC-002).

### M5 — Verification gate (mechanical)
`go test ./internal/hook/...`, coverage ≥85% on the classification path, `go vet`, `golangci-lint run internal/hook/...`, subagent-boundary grep = 0. Record verbatim output in `progress.md` §E.2/§E.3.

## §G Anti-Patterns to avoid

- Writing a bespoke `tool_response` parser when `decodeToolResponse` exists (Simplicity ladder violation).
- Adding an eighth `ErrorCategory` (out of scope).
- Claiming the session_id fix works without a test that exercises a session-id-less payload (verify-don't-assume; the audit explicitly did NOT verify this end-to-end).
- Editing existing top-level-`Error` test assertions to make them pass (that would mask a backward-compat regression).

## §H Cross-References

- Issue #1089; `internal/hook/CLAUDE.md` (JSON I/O contract, subagent boundary, fail-open).
- `.claude/rules/moai/quality/boundary-verification.md` Case 6/7 (payload shape variance).
- `internal/hook/evidence_writer.go` `decodeToolResponse` (reuse).
