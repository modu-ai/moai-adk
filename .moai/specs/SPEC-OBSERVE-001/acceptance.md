---
id: SPEC-OBSERVE-001
title: "Structured Observability - Acceptance Criteria"
---

# Acceptance Criteria

## AC-OBS-001: Trace Writing
**Given** observability is enabled
**When** Registry.Dispatch() completes for any event
**Then** a JSON line SHALL be appended to `.moai/logs/trace-{session_id}.jsonl`

## AC-OBS-002: Trace Schema
**Given** a PreToolUse event dispatched for Bash tool taking 45ms returning "allow"
**When** trace entry is written
**Then** JSON SHALL contain: ts (ISO-8601), event "PreToolUse", tool "Bash", duration_ms 45, decision "allow"

## AC-OBS-003: Non-Blocking
**Given** trace writing takes 100ms due to disk I/O
**When** Registry.Dispatch() returns
**Then** the hook response SHALL NOT be delayed by trace writing (async)

## AC-OBS-004: Session Summary
**Given** a session with 50 hook invocations
**When** SessionEnd event fires
**Then** `.moai/reports/session-{session_id}.md` SHALL be created with invocation count, event breakdown, decision breakdown

## AC-OBS-005: Config Disable
**Given** `observability.enabled` is false
**When** Registry.Dispatch() completes
**Then** NO trace entry SHALL be written

## AC-OBS-006: Deny Reason Capture
**Given** PreToolUse returns Decision "deny" with reason "blocked by security"
**When** trace entry is written
**Then** trace SHALL include reason "blocked by security"

## AC-OBS-007: Summary Top 5 Slowest
**Given** a session with 20 hooks, 5 of which took > 100ms
**When** session summary is generated
**Then** summary SHALL list the 5 slowest hook executions with event type and duration

## AC-OBS-008: File Rotation
**Given** trace file exceeds 10MB
**When** next trace entry is written
**Then** current file SHALL be rotated to `.1.jsonl` and a new file created

## AC-OBS-009: Graceful Degradation
**Given** `.moai/logs/` directory is not writable
**When** trace writing is attempted
**Then** the error SHALL be logged via slog.Warn and hook execution SHALL continue normally

## AC-OBS-010: Test Coverage
Unit tests SHALL achieve >= 85% coverage for `internal/hook/trace/` package
