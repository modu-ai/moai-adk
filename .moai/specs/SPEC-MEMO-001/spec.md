---
id: SPEC-MEMO-001
version: "1.0.0"
status: draft
created: "2026-04-07"
updated: "2026-04-07"
author: GOOS
priority: P1
issue_number: 0
---

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-04-07 | 1.0.0 | Initial draft |

---

# SPEC-MEMO-001: Session Memo (컨텍스트 압축 보존)

## Overview

PreCompact/PostCompact hook을 확장하여, 컨텍스트 압축 시 핵심 세션 상태를 `.moai/state/session-memo.md`에 보존하고, 압축 후 `systemMessage`로 재주입하는 시스템.

## Motivation

현재 `compact.go`와 `post_compact.go`는 stub 상태(로깅만 수행). 컨텍스트 압축 후 진행 중인 SPEC-ID, 워크플로우 단계, 에이전트 상태 등이 유실되어 Claude가 방향을 잃는 문제 발생.

## Requirements (EARS Format)

### REQ-MEMO-001 (Ubiquitous)
The system shall capture session state to `.moai/state/session-memo.md` during PreCompact events.

### REQ-MEMO-002 (Ubiquitous)
The system shall restore session state from `.moai/state/session-memo.md` via `systemMessage` during PostCompact events.

### REQ-MEMO-003 (Ubiquitous)
The system shall enforce a maximum token budget of 2,200 tokens for the session-memo content, trimming from lowest priority first.

### REQ-MEMO-004 (State-Driven)
When persistent-mode is active (SPEC-PERSIST-001), the session-memo shall include persistent-mode status and workflow context.

### REQ-MEMO-005 (Event-Driven)
When the session-memo file does not exist or is empty, the PostCompact handler shall return an empty response without error.

## Affected Files

### New Files
- `internal/hook/memo/writer.go` - Priority-based memo writer
- `internal/hook/memo/reader.go` - Memo reader with token trimming
- `internal/hook/memo/priority.go` - P1-P4 priority definitions
- `internal/hook/memo/writer_test.go`
- `internal/hook/memo/reader_test.go`
- `internal/hook/memo/priority_test.go`
- `internal/template/templates/.moai/config/sections/memo.yaml` - Configuration

### Modified Files
- `internal/hook/compact.go` - Integrate memo writer
- `internal/hook/compact_test.go` - Update tests
- `internal/hook/post_compact.go` - Integrate memo reader, return systemMessage
- `internal/hook/post_compact_test.go` - Update tests

## Technical Design

### Priority Levels

| Priority | Content | Token Budget |
|----------|---------|-------------|
| P1 (Required) | Active SPEC-ID, workflow phase, execution mode (solo/team/cg) | ~200 |
| P2 (High) | TaskList summary (in_progress/completed/pending counts), active agents | ~500 |
| P3 (Medium) | Last 3 agent execution result summaries | ~1000 |
| P4 (Low) | User decision history, major errors/warnings | ~500 |

### Data Flow

```
PreCompact event
  -> compact.go: collect state from HookInput (SessionID, CWD, Trigger)
  -> memo/writer.go: read existing state files (.moai/state/*, TaskList)
  -> memo/writer.go: write prioritized content to .moai/state/session-memo.md
  -> return HookOutput{Data: snapshot_metadata}

PostCompact event
  -> post_compact.go: read .moai/state/session-memo.md
  -> memo/reader.go: format + trim to token budget
  -> return HookOutput{SystemMessage: formatted_memo}
```

### HookOutput Integration

PostCompact handler returns:
```go
&HookOutput{
    SystemMessage: formattedMemo, // Injected into Claude context
}
```

Per `registry.go` line 128-134, `SystemMessage` fields are accumulated across handlers.

### Configuration (memo.yaml)

```yaml
memo:
  enabled: true
  max_tokens: 2200
  storage_path: ".moai/state/session-memo.md"
  priorities:
    p1_budget: 200
    p2_budget: 500
    p3_budget: 1000
    p4_budget: 500
```

## Dependencies

- None (foundational SPEC)

## Non-Goals

- Preserving conversation history (handled by Claude Code auto-compact)
- Cross-session memory (handled by auto-memory system)
- Real-time state streaming (single snapshot per compact event)
