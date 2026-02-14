# SPEC-HOOK-008: Implementation Plan

## Metadata

| Field       | Value                                                  |
| ----------- | ------------------------------------------------------ |
| SPEC ID     | SPEC-HOOK-008                                          |
| Title       | Complete Hook Event System - Add 7 Missing Hook Events |
| Methodology | Hybrid (TDD for new files, DDD for modified files)     |

---

## Overview

This plan adds the 7 missing Claude Code hook events to MoAI-ADK's hook system. The implementation follows a bottom-up approach: define types first, implement handlers, wire CLI integration, and update templates.

---

## Phase 1: EventType Constants and Type Extensions

**Priority: Primary Goal**

### Changes

**File: `internal/hook/types.go`**

1. Add 7 new EventType constants after line 38 (after `EventPreCompact`):
   ```go
   EventPostToolUseFailure EventType = "PostToolUseFailure"
   EventNotification       EventType = "Notification"
   EventSubagentStart      EventType = "SubagentStart"
   EventUserPromptSubmit   EventType = "UserPromptSubmit"
   EventPermissionRequest  EventType = "PermissionRequest"
   EventTeammateIdle       EventType = "TeammateIdle"
   EventTaskCompleted      EventType = "TaskCompleted"
   ```

2. Update `ValidEventTypes()` (lines 42-52) to include all 7 new constants.

3. Add `UpdatedInput` field to HookOutput struct (after line 157):
   ```go
   UpdatedInput string `json:"updatedInput,omitempty"`
   ```

4. Add `ExitCode` field to HookOutput struct for exit code 2 protocol:
   ```go
   ExitCode int `json:"-"` // Internal: non-zero exit code (e.g., 2 for TeammateIdle/TaskCompleted)
   ```

5. Add helper constructors:
   - `NewPermissionRequestOutput(decision, reason string) *HookOutput` - For PermissionRequest responses
   - `NewUserPromptBlockOutput(reason string) *HookOutput` - For blocking user prompts
   - `NewTeammateKeepWorkingOutput() *HookOutput` - Exit code 2 for TeammateIdle
   - `NewTaskRejectedOutput() *HookOutput` - Exit code 2 for TaskCompleted

### Methodology: DDD (modifying existing file)

- ANALYZE: Read existing type constants, ValidEventTypes, HookOutput struct
- PRESERVE: Write characterization tests for IsValidEventType, existing constructors
- IMPROVE: Add new constants and fields

---

## Phase 2: Handler Implementations (7 new files)

**Priority: Primary Goal**

### Methodology: TDD (all new files)

For each handler, follow RED-GREEN-REFACTOR:
1. RED: Write test file first with table-driven tests
2. GREEN: Implement minimal handler to pass tests
3. REFACTOR: Clean up, add logging, edge case handling

### 2.1 PostToolUseFailure Handler

**New File: `internal/hook/post_tool_failure.go`**
**New File: `internal/hook/post_tool_failure_test.go`**

```go
type postToolUseFailureHandler struct{}

func NewPostToolUseFailureHandler() Handler
func (h *postToolUseFailureHandler) EventType() EventType  // EventPostToolUseFailure
func (h *postToolUseFailureHandler) Handle(ctx, input) (*HookOutput, error)
```

Behavior:
- Log tool failure with `error` and `is_interrupt` fields
- Log `tool_name` and `tool_use_id` for traceability
- Return empty `HookOutput{}` (allow default processing)
- Support `decision: "block"` via handler chain (no custom logic needed in base handler)

### 2.2 Notification Handler

**New File: `internal/hook/notification.go`**
**New File: `internal/hook/notification_test.go`**

```go
type notificationHandler struct{}

func NewNotificationHandler() Handler
func (h *notificationHandler) EventType() EventType  // EventNotification
func (h *notificationHandler) Handle(ctx, input) (*HookOutput, error)
```

Behavior:
- Log notification with `message`, `title`, and `notification_type` fields
- Return empty `HookOutput{}`
- No blocking capability (notifications are informational)

### 2.3 SubagentStart Handler

**New File: `internal/hook/subagent_start.go`**
**New File: `internal/hook/subagent_start_test.go`**

```go
type subagentStartHandler struct{}

func NewSubagentStartHandler() Handler
func (h *subagentStartHandler) EventType() EventType  // EventSubagentStart
func (h *subagentStartHandler) Handle(ctx, input) (*HookOutput, error)
```

Behavior:
- Log subagent startup with `agent_id` and `agent_transcript_path`
- Return empty `HookOutput{}`

### 2.4 UserPromptSubmit Handler

**New File: `internal/hook/user_prompt_submit.go`**
**New File: `internal/hook/user_prompt_submit_test.go`**

```go
type userPromptSubmitHandler struct{}

func NewUserPromptSubmitHandler() Handler
func (h *userPromptSubmitHandler) EventType() EventType  // EventUserPromptSubmit
func (h *userPromptSubmitHandler) Handle(ctx, input) (*HookOutput, error)
```

Behavior:
- Log prompt submission (truncated for privacy, first 100 chars)
- Return empty `HookOutput{}` (allow prompt)
- Support `decision: "block"` and `updatedInput` via output fields

### 2.5 PermissionRequest Handler

**New File: `internal/hook/permission_request.go`**
**New File: `internal/hook/permission_request_test.go`**

```go
type permissionRequestHandler struct{}

func NewPermissionRequestHandler() Handler
func (h *permissionRequestHandler) EventType() EventType  // EventPermissionRequest
func (h *permissionRequestHandler) Handle(ctx, input) (*HookOutput, error)
```

Behavior:
- Log permission request with `tool_name`
- Return HookOutput with `hookSpecificOutput.permissionDecision = "ask"` (defer to default)
- Uses existing `HookSpecificOutput` struct with `PermissionDecision` field

### 2.6 TeammateIdle Handler

**New File: `internal/hook/teammate_idle.go`**
**New File: `internal/hook/teammate_idle_test.go`**

```go
type teammateIdleHandler struct{}

func NewTeammateIdleHandler() Handler
func (h *teammateIdleHandler) EventType() EventType  // EventTeammateIdle
func (h *teammateIdleHandler) Handle(ctx, input) (*HookOutput, error)
```

Behavior:
- Log teammate idle event with `agent_id`
- Return empty `HookOutput{}` (accept idle state)
- When keep-working is desired, return `HookOutput{ExitCode: 2}`

### 2.7 TaskCompleted Handler

**New File: `internal/hook/task_completed.go`**
**New File: `internal/hook/task_completed_test.go`**

```go
type taskCompletedHandler struct{}

func NewTaskCompletedHandler() Handler
func (h *taskCompletedHandler) EventType() EventType  // EventTaskCompleted
func (h *taskCompletedHandler) Handle(ctx, input) (*HookOutput, error)
```

Behavior:
- Log task completion with session context
- Return empty `HookOutput{}` (accept completion)
- When rejection is desired, return `HookOutput{ExitCode: 2}`

---

## Phase 3: Handler Registration

**Priority: Primary Goal**

### Changes

**File: `internal/cli/deps.go`** (DDD - modifying existing file)

Add handler registration after line 93 (after `NewCompactHandler()`):

```go
deps.HookRegistry.Register(hook.NewPostToolUseFailureHandler())
deps.HookRegistry.Register(hook.NewNotificationHandler())
deps.HookRegistry.Register(hook.NewSubagentStartHandler())
deps.HookRegistry.Register(hook.NewUserPromptSubmitHandler())
deps.HookRegistry.Register(hook.NewPermissionRequestHandler())
deps.HookRegistry.Register(hook.NewTeammateIdleHandler())
deps.HookRegistry.Register(hook.NewTaskCompletedHandler())
```

---

## Phase 4: CLI Subcommand Registration

**Priority: Primary Goal**

### Changes

**File: `internal/cli/hook.go`** (DDD - modifying existing file)

1. Add 7 entries to the `hookSubcommands` table (after line 34):
   ```go
   {"post-tool-failure",  "Handle post-tool-use failure event",  hook.EventPostToolUseFailure},
   {"notification",       "Handle notification event",           hook.EventNotification},
   {"subagent-start",     "Handle subagent start event",         hook.EventSubagentStart},
   {"user-prompt-submit", "Handle user prompt submit event",     hook.EventUserPromptSubmit},
   {"permission-request", "Handle permission request event",     hook.EventPermissionRequest},
   {"teammate-idle",      "Handle teammate idle event",          hook.EventTeammateIdle},
   {"task-completed",     "Handle task completed event",         hook.EventTaskCompleted},
   ```

2. Extend `runHookEvent` exit code handling (around line 89-93) to support exit code 2:
   ```go
   // Exit code 2 for special protocol decisions
   if output != nil && output.ExitCode == 2 {
       os.Exit(2)
   }
   // Exit code 2 for deny decisions per Claude Code protocol
   if output != nil && output.Decision == hook.DecisionDeny {
       os.Exit(2)
   }
   ```

---

## Phase 5: Registry Default Output Update

**Priority: Primary Goal**

### Changes

**File: `internal/hook/registry.go`** (DDD - modifying existing file)

Update `defaultOutputForEvent()` switch statement (lines 144-157) to handle new events:

```go
case EventPostToolUseFailure, EventNotification, EventSubagentStart,
     EventUserPromptSubmit, EventTeammateIdle, EventTaskCompleted:
    return &HookOutput{}
case EventPermissionRequest:
    return &HookOutput{
        HookSpecificOutput: &HookSpecificOutput{
            HookEventName:      "PermissionRequest",
            PermissionDecision: DecisionAsk,
        },
    }
```

---

## Phase 6: Settings Template and Shell Wrappers

**Priority: Secondary Goal**

### 6.1 Settings Template

**File: `internal/template/templates/.claude/settings.json.tmpl`** (DDD)

Add hook configuration entries for all 7 new events. Each follows the established pattern:

```json
"PostToolUseFailure": [{
  "hooks": [{
    "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-post-tool-failure.sh\"",
    "timeout": 5,
    "type": "command"
  }]
}],
```

Event-specific considerations:
- **PostToolUseFailure**: No matcher needed (global)
- **Notification**: No matcher needed (global)
- **SubagentStart**: No matcher needed (global)
- **UserPromptSubmit**: No matcher needed (global)
- **PermissionRequest**: No matcher needed (global)
- **TeammateIdle**: No matcher needed (global)
- **TaskCompleted**: No matcher needed (global)

### 6.2 Shell Wrapper Scripts (7 new files)

Each follows the established template pattern from `handle-stop.sh.tmpl`:

| Template File                             | CLI Subcommand     |
| ----------------------------------------- | ------------------ |
| handle-post-tool-failure.sh.tmpl          | post-tool-failure  |
| handle-notification.sh.tmpl              | notification       |
| handle-subagent-start.sh.tmpl            | subagent-start     |
| handle-user-prompt-submit.sh.tmpl        | user-prompt-submit |
| handle-permission-request.sh.tmpl        | permission-request |
| handle-teammate-idle.sh.tmpl             | teammate-idle      |
| handle-task-completed.sh.tmpl            | task-completed     |

---

## Phase 7: Testing

**Priority: Primary Goal**

### 7.1 Unit Tests (per handler)

Each handler test file follows the table-driven test pattern:

```go
func TestXxxHandler_Handle(t *testing.T) {
    tests := []struct {
        name     string
        input    *HookInput
        wantErr  bool
        validate func(t *testing.T, output *HookOutput)
    }{
        // test cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test body
        })
    }
}
```

### 7.2 Types Tests

Update `internal/hook/types_test.go`:
- Test all 7 new EventType constants are in `ValidEventTypes()`
- Test `IsValidEventType()` returns true for all new types
- Test new constructor functions
- Test HookOutput `UpdatedInput` JSON serialization
- Test HookOutput `ExitCode` is not serialized to JSON

### 7.3 Registry Tests

Update `internal/hook/registry_test.go`:
- Test `defaultOutputForEvent()` for all new event types
- Test PermissionRequest default returns `permissionDecision: "ask"`
- Test dispatch with new event types

### 7.4 Integration Test

- Verify `make build` succeeds (template embedding)
- Verify `go test -race ./internal/hook/...` passes
- Verify `go test -race ./internal/cli/...` passes
- Verify `go vet ./...` passes

---

## File Change Summary

### New Files (14)

| File                                                                    | Purpose                     |
| ----------------------------------------------------------------------- | --------------------------- |
| internal/hook/post_tool_failure.go                                      | PostToolUseFailure handler  |
| internal/hook/post_tool_failure_test.go                                 | PostToolUseFailure tests    |
| internal/hook/notification.go                                           | Notification handler        |
| internal/hook/notification_test.go                                      | Notification tests          |
| internal/hook/subagent_start.go                                         | SubagentStart handler       |
| internal/hook/subagent_start_test.go                                    | SubagentStart tests         |
| internal/hook/user_prompt_submit.go                                     | UserPromptSubmit handler    |
| internal/hook/user_prompt_submit_test.go                                | UserPromptSubmit tests      |
| internal/hook/permission_request.go                                     | PermissionRequest handler   |
| internal/hook/permission_request_test.go                                | PermissionRequest tests     |
| internal/hook/teammate_idle.go                                          | TeammateIdle handler        |
| internal/hook/teammate_idle_test.go                                     | TeammateIdle tests          |
| internal/hook/task_completed.go                                         | TaskCompleted handler       |
| internal/hook/task_completed_test.go                                    | TaskCompleted tests         |

### New Template Files (7)

| File                                                                              | Purpose                    |
| --------------------------------------------------------------------------------- | -------------------------- |
| internal/template/templates/.claude/hooks/moai/handle-post-tool-failure.sh.tmpl   | Shell wrapper              |
| internal/template/templates/.claude/hooks/moai/handle-notification.sh.tmpl        | Shell wrapper              |
| internal/template/templates/.claude/hooks/moai/handle-subagent-start.sh.tmpl      | Shell wrapper              |
| internal/template/templates/.claude/hooks/moai/handle-user-prompt-submit.sh.tmpl  | Shell wrapper              |
| internal/template/templates/.claude/hooks/moai/handle-permission-request.sh.tmpl  | Shell wrapper              |
| internal/template/templates/.claude/hooks/moai/handle-teammate-idle.sh.tmpl       | Shell wrapper              |
| internal/template/templates/.claude/hooks/moai/handle-task-completed.sh.tmpl      | Shell wrapper              |

### Modified Files (5)

| File                                                                | Changes                                |
| ------------------------------------------------------------------- | -------------------------------------- |
| internal/hook/types.go                                              | 7 constants, UpdatedInput, ExitCode    |
| internal/hook/registry.go                                           | defaultOutputForEvent switch cases     |
| internal/cli/hook.go                                                | 7 subcommands, exit code 2 handling    |
| internal/cli/deps.go                                                | 7 handler registrations                |
| internal/template/templates/.claude/settings.json.tmpl              | 7 hook config entries                  |

### Test Files Updated (2)

| File                                    | Changes                              |
| --------------------------------------- | ------------------------------------ |
| internal/hook/types_test.go             | New constant and constructor tests   |
| internal/hook/registry_test.go          | Default output tests for new events  |

---

## Risks and Mitigation

### Risk 1: Exit Code 2 Semantics

**Risk**: The existing `runHookEvent` function uses exit code 2 only for `DecisionDeny`. TeammateIdle and TaskCompleted need exit code 2 for a different semantic (keep working / reject completion).

**Mitigation**: Add an `ExitCode` field to HookOutput (not serialized to JSON) that handlers can set explicitly. The CLI checks this field before the Decision field.

### Risk 2: Template Embedding Size

**Risk**: Adding 7 new shell wrapper templates increases embedded binary size.

**Mitigation**: Each shell wrapper is ~30 lines (~800 bytes). Total increase: ~5.6KB. Negligible impact on binary size.

### Risk 3: Backward Compatibility

**Risk**: Changes to types.go and registry.go could break existing handler behavior.

**Mitigation**:
- DDD methodology for all modified files (characterization tests first)
- New constants are additive (no removal or rename)
- HookOutput extensions use `omitempty` tags
- ExitCode field uses `json:"-"` to prevent serialization

### Risk 4: Agent Teams Feature Maturity

**Risk**: TeammateIdle and TaskCompleted are experimental Agent Teams features.

**Mitigation**: Handlers are minimal (log and return empty output). The exit code 2 mechanism is implemented but the default behavior is to accept idle/completion (no exit code 2 by default).

---

## Execution Order

1. Phase 1 (types.go) - Foundation for all other phases
2. Phase 5 (registry.go) - Default outputs needed before handlers
3. Phase 2 (7 handlers) - Can be done in parallel per handler
4. Phase 3 (deps.go) - Depends on Phase 2
5. Phase 4 (hook.go) - Depends on Phase 1
6. Phase 6 (templates) - Independent, can parallel with Phase 2-4
7. Phase 7 (testing) - Continuous throughout, final verification at end
