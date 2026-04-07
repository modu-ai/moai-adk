---
id: SPEC-PERSIST-001
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

# SPEC-PERSIST-001: Stop Persistent Mode (종료 방지 모드)

## Overview

Stop hook을 확장하여, `/moai loop` 또는 `/moai run` 진행 중 Claude의 조기 종료를 방지. 완료 마커 감지 또는 최대 지속 시간 초과 시 자동 해제.

## Motivation

현재 `stop.go`는 완료 마커를 관찰만 하고 실제로 차단(block)하지 않음. `/moai loop`에서 Claude가 중간에 멈추면 사용자가 수동으로 재시작해야 하며, 진행 상태가 유실될 수 있음.

## Requirements (EARS Format)

### REQ-PERSIST-001 (State-Driven)
When persistent-mode is active and `stop_hook_active` is false, the Stop handler shall return `{"decision": "block", "reason": "..."}` to prevent Claude from stopping.

### REQ-PERSIST-002 (State-Driven)
When a completion marker (`<moai>DONE</moai>` or `<moai>COMPLETE</moai>`) is detected in `last_assistant_message`, the handler shall deactivate persistent-mode and allow the stop.

### REQ-PERSIST-003 (State-Driven)
When `stop_hook_active` is true, the handler shall always allow the stop to prevent infinite loops.

### REQ-PERSIST-004 (State-Driven)
When `max_duration_minutes` is exceeded since activation, the handler shall deactivate persistent-mode and allow the stop.

### REQ-PERSIST-005 (Event-Driven)
When `/moai loop` or `/moai run` workflow starts, the system shall activate persistent-mode by writing `.moai/state/persistent-mode.json`.

### REQ-PERSIST-006 (Event-Driven)
When the workflow completes or the user manually cancels, the system shall deactivate persistent-mode.

## Affected Files

### New Files
- `internal/hook/lifecycle/persistent_mode.go` - Activate/deactivate/check API
- `internal/hook/lifecycle/persistent_mode_test.go`

### Modified Files
- `internal/hook/stop.go` - Add persistent-mode check before allowing stop
- `internal/hook/stop_test.go` - Add block/allow scenario tests
- `internal/hook/stop_completion_test.go` - Add persistent-mode + marker tests
- `internal/template/templates/.claude/skills/moai/workflows/loop.md` - Activate on start
- `internal/template/templates/.claude/skills/moai/workflows/run.md` - Activate on start

## Technical Design

### State File (`.moai/state/persistent-mode.json`)

```json
{
  "active": true,
  "workflow": "loop",
  "spec_id": "SPEC-AUTH-001",
  "started_at": "2026-04-07T10:00:00Z",
  "max_duration_minutes": 60
}
```

### Stop Handler Decision Tree

```
Stop event received
  |
  +-- stop_hook_active == true? --> ALLOW (prevent infinite loop)
  |
  +-- Read .moai/state/persistent-mode.json
  |     |
  |     +-- File missing or active == false? --> ALLOW (normal stop)
  |     |
  |     +-- active == true:
  |           |
  |           +-- Completion marker in last_assistant_message? --> DEACTIVATE + ALLOW
  |           |
  |           +-- max_duration exceeded? --> DEACTIVATE + ALLOW
  |           |
  |           +-- Otherwise --> BLOCK ("Workflow in progress. Continuing...")
```

### Block Output

Uses existing `NewStopBlockOutput()` from `types.go`:
```go
return &HookOutput{
    Decision: "block",
    Reason:   fmt.Sprintf("Persistent mode active: %s workflow on %s. Continuing...", mode.Workflow, mode.SpecID),
}
```

### Lifecycle API

```go
// internal/hook/lifecycle/persistent_mode.go
type PersistentMode struct {
    Active             bool      `json:"active"`
    Workflow           string    `json:"workflow"`
    SpecID             string    `json:"spec_id"`
    StartedAt          time.Time `json:"started_at"`
    MaxDurationMinutes int       `json:"max_duration_minutes"`
}

func Activate(projectDir, workflow, specID string, maxMinutes int) error
func Deactivate(projectDir string) error
func Check(projectDir string) (*PersistentMode, error)
```

### Workflow Integration

In `loop.md` and `run.md`, add Bash command at workflow start:
```bash
# Activate persistent mode (called by MoAI orchestrator via Bash tool)
echo '{"active":true,"workflow":"loop","spec_id":"SPEC-XXX","started_at":"...","max_duration_minutes":60}' > .moai/state/persistent-mode.json
```

Deactivation happens automatically via:
1. Completion marker detection in stop.go
2. Max duration timeout in stop.go
3. Explicit deactivation at workflow end

## Dependencies

- SPEC-MEMO-001 (optional): Session-memo includes persistent-mode status for context preservation

## Non-Goals

- User-facing persistent mode toggle (implicit via workflow lifecycle)
- Cross-session persistence (file is session-scoped, cleaned up by session_end)
