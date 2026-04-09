---
id: SPEC-OBSERVE-001
title: "Structured Observability"
status: draft
priority: P0
created: "2026-04-07"
harness_pillar: "P5: Observability"
---

# SPEC-OBSERVE-001: Structured Observability

## Overview

매 훅 실행마다 구조화된 JSON 트레이스를 `.moai/logs/trace-{session}.jsonl`에 기록하고, 세션 종료 시 요약 리포트를 생성하는 관측 가능성 시스템.

하네스 엔지니어링: "기업 73%가 AI 에이전트를 모니터링 없이 통제 불능 상태로 방치한다"

## Requirements (EARS Format)

### REQ-OBS-001 (Ubiquitous)
Every hook dispatch SHALL write a JSON line to `.moai/logs/trace-{session_id}.jsonl` with: timestamp, event_type, handler_name, duration_ms, decision, and error (if any).

### REQ-OBS-002 (Ubiquitous)
The trace entry JSON schema SHALL be:
```json
{
  "ts": "2026-04-07T10:30:00Z",
  "event": "PreToolUse",
  "handler": "preToolHandler",
  "tool": "Bash",
  "duration_ms": 45,
  "decision": "allow",
  "error": null,
  "session_id": "abc123"
}
```

### REQ-OBS-003 (Ubiquitous)
Trace writing SHALL be async (non-blocking) and SHALL NOT affect hook response latency.

### REQ-OBS-004 (Event-Driven)
When SessionEnd event fires, the system SHALL generate a session summary at `.moai/reports/session-{session_id}.md` with: total hooks fired, duration, decisions breakdown, errors count.

### REQ-OBS-005 (State-Driven)
When `observability.enabled` is false in `.moai/config/sections/observability.yaml`, all tracing SHALL be disabled.

### REQ-OBS-006 (Ubiquitous)
The trace writer SHALL rotate files when they exceed 10MB, creating `trace-{session_id}.1.jsonl`.

### REQ-OBS-007 (Ubiquitous)
Trace files older than 30 days SHALL be eligible for cleanup via `moai clean --traces`.

### REQ-OBS-008 (Event-Driven)
When a hook handler returns Decision "deny" or "block", the trace entry SHALL include the reason field.

### REQ-OBS-009 (Ubiquitous)
The session summary SHALL include:
- Total hook invocations count
- Breakdown by event type
- Breakdown by decision (allow/deny/block)
- Top 5 slowest hook executions
- Error count and list

## Architecture

```
Registry.Dispatch()
  → handler.Handle() [existing]
  → traceWriter.Write(traceEntry) [NEW, async]
  → return HookOutput [unchanged latency]

SessionEnd handler
  → traceReader.Summarize(sessionID) [NEW]
  → write .moai/reports/session-{id}.md
```

## Implementation Scope

### New Files
- `internal/hook/trace/writer.go` — TraceWriter with async buffered writes
- `internal/hook/trace/writer_test.go` — Unit tests
- `internal/hook/trace/entry.go` — TraceEntry struct
- `internal/hook/trace/summary.go` — Session summary generator
- `internal/hook/trace/summary_test.go` — Unit tests
- `internal/template/templates/.moai/config/sections/observability.yaml` — Default config

### Modified Files
- `internal/hook/registry.go` — Inject TraceWriter, write trace after each dispatch
- `internal/hook/session_end.go` — Generate session summary on SessionEnd
- `internal/cli/deps.go` — Wire TraceWriter into Registry

## Non-Goals
- Real-time dashboard UI (CLI query only for now)
- OpenTelemetry export (future consideration)
- Token usage estimation (separate SPEC)
- `moai observe` CLI command (P1, separate SPEC)
