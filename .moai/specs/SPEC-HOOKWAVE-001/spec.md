---
id: SPEC-HOOKWAVE-001
version: "1.0.0"
status: draft
created: "2026-04-07"
updated: "2026-04-07"
author: GOOS
priority: P2
issue_number: 0
---

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-04-07 | 1.0.0 | Initial draft |

---

# SPEC-HOOKWAVE-001: Hook Stub Improvement Wave 2

## Overview

현재 로깅만 수행하는 stub 핸들러 9개와 Go 핸들러 없이 shell wrapper만 존재하는 이벤트 6개를 실질적 기능으로 업그레이드. 특히 SubagentStop는 유일하게 EventType + CLI 라우팅 + shell wrapper가 모두 존재하는데 Go 핸들러만 없는 상태.

## Motivation

19개 등록된 Go 핸들러 중 9개가 stub(로깅만), 6개 shell wrapper는 Go 핸들러 미존재. hook 인프라는 잘 설계되어 있으나 핸들러의 39%가 아무 일도 하지 않음. Notification 핸들러는 현재 상태 유지 (개선 불필요).

## Requirements (EARS Format)

### Module A: SubagentStop Handler (유일한 미구현 이벤트)

#### REQ-HOOKWAVE-A01 (Ubiquitous)
The system shall implement a SubagentStop handler that logs agent completion with agent_id, agent_name, and transcript path.

#### REQ-HOOKWAVE-A02 (Event-Driven)
When SubagentStop fires, the handler shall update session-memo (SPEC-MEMO-001) with the agent's execution result summary if memo is active.

#### REQ-HOOKWAVE-A03 (Ubiquitous)
The handler shall register in `deps.go` alongside existing handlers.

### Module B: UserPromptSubmit Enhancement

#### REQ-HOOKWAVE-B01 (Event-Driven)
When the user submits a prompt containing workflow keywords (loop, run, plan), the handler shall inject workflow context via `additionalContext`.

### Module C: Worktree Lifecycle Enhancement

#### REQ-HOOKWAVE-C01 (Event-Driven)
When a worktree is created, the handler shall register it in `.moai/state/worktrees.json`.

#### REQ-HOOKWAVE-C02 (Event-Driven)
When a worktree is removed, the handler shall verify pending changes and update the registry.

### Module D: Missing Go Handler Creation

#### REQ-HOOKWAVE-D01 (Ubiquitous)
The system shall create Go handlers for: SubagentStop, TaskCreated, PermissionDenied, ConfigChange, CwdChanged.

#### REQ-HOOKWAVE-D02 (Ubiquitous)
Each new handler shall follow the existing Handler interface contract.

### Module E: New EventType Constants

#### REQ-HOOKWAVE-E01 (Ubiquitous)
The system shall add missing EventType constants to types.go: TaskCreated, PermissionDenied, ConfigChange, CwdChanged, FileChanged, Elicitation, ElicitationResult.

## Affected Files

### New Files
- `internal/hook/subagent_stop.go` + test
- `internal/hook/task_created.go` + test
- `internal/hook/permission_denied.go` + test
- `internal/hook/config_change.go` + test
- `internal/hook/cwd_changed.go` + test
- `internal/hook/worktree_registry.go` + test

### Modified Files
- `internal/hook/user_prompt_submit.go` - Keyword context injection
- `internal/hook/worktree_create.go` - Registry integration
- `internal/hook/worktree_remove.go` - Registry integration
- `internal/cli/deps.go` - Register 5 new handlers
- `internal/hook/types.go` - Add missing EventType constants

## Technical Design

### Module Priority

| Module | Priority | Complexity |
|--------|----------|-----------|
| A: SubagentStop | P1 | Low |
| D: Missing Handlers | P1 | Low |
| E: EventType Constants | P1 | Low |
| C: Worktree Lifecycle | P2 | Medium |
| B: UserPromptSubmit | P3 | Medium |

## Dependencies

- SPEC-MEMO-001 (optional): Module A updates session-memo on agent completion

## Non-Goals

- Agent-specific handlers (agents/ subdirectory)
- Notification handler enhancement (per user feedback: not needed)
- PermissionRequest auto-allow (deferred to separate SPEC)
