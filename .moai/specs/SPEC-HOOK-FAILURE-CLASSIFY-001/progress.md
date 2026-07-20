---
id: SPEC-HOOK-FAILURE-CLASSIFY-001
title: "Progress — PostToolUseFailure nested-error classification"
version: "0.1.0"
status: completed
created: 2026-07-17
updated: 2026-07-21
author: manager-spec
tier: S
---

# Progress — SPEC-HOOK-FAILURE-CLASSIFY-001

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifact set authored (Tier S LEAN): `spec.md` (AC inline §D) + `plan.md` + `progress.md`.
- SPEC ID pre-write self-check: `decomposition: SPEC ✓ | HOOK ✓ | FAILURE ✓ | CLASSIFY ✓ | 001 ✓ → PASS` (canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`, executed → PASS). No collision in `.moai/specs/`.
- Root causes verified against source at HEAD `efd423b88`: (1) `classifyError` reads only top-level `input.Error`; (2) `validateInput` sets `session_id = "unknown"` → `trace-unknown.jsonl`.
- 6 REQ (GEARS) / 8 AC. Reuse candidate identified: `decodeToolResponse` (evidence_writer.go).
- Open run-phase investigation (NOT a product-decision clarification): session_id resolution source (M2 — likely `transcript_path`, evidence pending).
- No `[NEEDS CLARIFICATION]` markers — all choices resolved from code evidence.

## §E.2 Run-phase Evidence

Run-phase baseline: HEAD `35119252c` (green `go test ./...` before change). TDD RED confirmed (4 new tests failing pre-fix: nested classification, unknown-excerpt, precedence, resolver), then GREEN.

**session_id resolution investigation outcome (REQ-HFC-004)**: chosen source = `transcript_path` base name. Claude Code names the session transcript `~/.claude/projects/<hash>/<session-uuid>.jsonl`, so when `session_id` is absent (bug #541 → `validateInput` substitutes `"unknown"`), the session UUID is derived from `filepath.Base(transcript_path)` after stripping `.jsonl` and validating against the RFC 4122 UUID regex. Implemented in `internal/hook/trace_session.go` (`resolveTraceSessionID`), wired at the trace-filename binding site `internal/hook/registry.go` `Dispatch` → `ensureTraceWriter(resolveTraceSessionID(input))`. Last-resort fallback: when neither `session_id` nor a UUID-shaped transcript base is available, the original `"unknown"` value is kept (documented, unchanged) — the win is non-collision for RESOLVABLE sessions (plan.md R2). The global `validateInput` substitution for other events is untouched (§C exclusion).

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-HFC-001a | PASS | `go test -run TestPostToolUseFailureHandler_NestedToolResponse ./internal/hook/` | `ok` — nested `{"error":"permission denied: open /f"}` → `PermissionDenied:` prefix |
| AC-HFC-001b | PASS | same | nested `{"stderr":"...context deadline exceeded..."}` → `TimeoutError` |
| AC-HFC-002 | PASS | `go test -run TestPostToolUseFailureHandler_Handle ./internal/hook/` | pre-existing top-level-Error table passes UNMODIFIED (assertions untouched) |
| AC-HFC-003 | PASS | `grep -c 'ToolResponse\|tool_response' internal/hook/post_tool_failure_test.go` → `13` (≥7); 7 per-category nested cases in `TestPostToolUseFailureHandler_NestedToolResponse` | package `ok` |
| AC-HFC-004 | PASS | `go test -run TestResolveTraceSessionID ./internal/hook/` | UUID derived from transcript_path when session_id absent/"unknown"; explicit session_id wins (no regression); unresolvable → "unknown" |
| AC-HFC-005 | PASS | `go test -run TestPostToolUseFailureHandler_UnknownFailureExcerpt ./internal/hook/` | message ≠ content-free string; contains `something went wrong`; 500-char error truncated at 200 runes + `...` |
| AC-HFC-006 | PASS | `go test -run TestPostToolUseFailureHandler_ToolResponseResilience ./internal/hook/` | malformed `[}` / bare string / array / absent / `{}` → no error, no panic, non-empty systemMessage |
| AC-HFC-GATE | PASS | `go test ./...` exit=0 (106 pkgs ok); `go vet ./internal/hook/...` exit=0; `golangci-lint run internal/hook/...` → `0 issues`; boundary grep → 0 matches | evidence: `.moai/state/verify/hfc001/1-go-test.log` |

Invariants: no new `ErrorCategory`; ordered matcher untouched (aggregation widens the haystack only); `recordToolFailureEvent` / event-key format unchanged; no `TraceWriter` redesign (only the sessionID value bound at `ensureTraceWriter`).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-21
run_commit_sha: 99309448a
run_status: audit-ready
ac_pass_count: 8
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: "0 0 (synced with origin/main at spawn)"
l44_post_push_fetch: "n/a — push not performed per delegation (Do NOT push)"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: "go build ./... exit 0"
  windows: "GOOS=windows GOARCH=amd64 go build ./... exit 0"
total_run_phase_files: 5
m1_to_mN_commit_strategy: "single commit (Tier S, M1-M5 consolidated)"
```

Coverage (classification path, ≥85% target): `Handle` 100%, `classificationText` 100%, `rawErrorExcerpt` 100%, `classifyError` 100%, `formatMessage` 100%, `resolveTraceSessionID` 100%, `sessionUUIDFromTranscriptPath` 100% (`go tool cover -func`). Package-wide `internal/hook` 83.6% is the pre-existing baseline (unrelated handlers); every touched function meets the target.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-21
sync_commit_sha: 2325000a7
sync_status: audit-ready
```

CHANGELOG.md `[Unreleased]` § Fixed section updated with a #1089 entry (dedup pre-check: `grep -c 'SPEC-HOOK-FAILURE-CLASSIFY-001' CHANGELOG.md` was 0 before this commit). Frontmatter `status: in-progress → completed` transition rides this single sync commit per the 3-phase close (plan → run → sync); `updated:` refreshed to 2026-07-21. No spec.md/plan.md/acceptance.md body content modified. README/docs-site: out of scope (internal hook bugfix, no user-facing docs surface change beyond CHANGELOG).
