# SPEC-HOOK-008: Complete Hook Event System

## Metadata

| Field    | Value                                                  |
| -------- | ------------------------------------------------------ |
| SPEC ID  | SPEC-HOOK-008                                          |
| Title    | Complete Hook Event System - Add 7 Missing Hook Events |
| Status   | Active                                                 |
| Priority | P0 (Critical)                                          |
| Created  | 2026-02-14                                             |
| Domain   | internal/hook, internal/cli, internal/template          |

---

## 1. Environment

### Current System State

MoAI-ADK implements a hook system that integrates with Claude Code's event-driven architecture. The hook system reads JSON from stdin, dispatches to registered handlers, and writes JSON to stdout per the Claude Code hooks protocol.

**Implemented Events (7 of 14):**

| Event          | EventType Constant | Handler File      | CLI Subcommand  | Shell Wrapper                  |
| -------------- | ------------------ | ----------------- | --------------- | ------------------------------ |
| SessionStart   | EventSessionStart  | session_start.go  | session-start   | handle-session-start.sh.tmpl   |
| PreToolUse     | EventPreToolUse    | pre_tool.go       | pre-tool        | handle-pre-tool.sh.tmpl        |
| PostToolUse    | EventPostToolUse   | post_tool.go      | post-tool       | handle-post-tool.sh.tmpl       |
| SessionEnd     | EventSessionEnd    | session_end.go    | session-end     | handle-session-end.sh.tmpl     |
| Stop           | EventStop          | stop.go           | stop            | handle-stop.sh.tmpl            |
| SubagentStop   | EventSubagentStop  | (none, framework) | (none)          | (none)                         |
| PreCompact     | EventPreCompact    | compact.go        | compact         | handle-compact.sh.tmpl         |

**Missing Events (7 to implement):**

| Event              | Claude Code Name     | Protocol Output Type               |
| ------------------ | -------------------- | ----------------------------------- |
| PostToolUseFailure | PostToolUseFailure   | Top-level decision: "block"         |
| Notification       | Notification         | Standard (systemMessage)            |
| SubagentStart      | SubagentStart        | Standard (systemMessage)            |
| UserPromptSubmit   | UserPromptSubmit     | Top-level decision / updatedInput   |
| PermissionRequest  | PermissionRequest    | hookSpecificOutput.permissionDecision |
| TeammateIdle       | TeammateIdle         | Exit code 2 = keep working          |
| TaskCompleted      | TaskCompleted        | Exit code 2 = reject completion     |

### Codebase Architecture

- **Event types**: `internal/hook/types.go` (lines 18-39) - EventType constants
- **Handler interface**: `internal/hook/types.go` (line 264) - `Handle(ctx, input) (*HookOutput, error)`
- **Registry**: `internal/hook/registry.go` - Sequential dispatch with block short-circuit
- **Handler registration**: `internal/cli/deps.go` (lines 74-93) - Composition Root
- **CLI subcommands**: `internal/cli/hook.go` (lines 24-35) - Cobra subcommand table
- **Settings template**: `internal/template/templates/.claude/settings.json.tmpl`
- **Shell wrappers**: `internal/template/templates/.claude/hooks/moai/handle-*.sh.tmpl`
- **HookInput**: Already has JSON fields for all missing events (types.go lines 86-122)
- **HookOutput**: Already has decision/reason fields (types.go lines 148-157)

---

## 2. Assumptions

- **A1**: Claude Code's hook protocol for the 7 missing events is stable and follows the documented patterns.
- **A2**: The existing Handler interface (`Handle(ctx, input) (*HookOutput, error)`) is sufficient for all new events without modification.
- **A3**: HookInput and HookOutput structs already contain all necessary JSON fields for the missing events (verified in types.go).
- **A4**: The HookSpecificOutput struct already has `permissionDecision` and `permissionDecisionReason` fields needed for PermissionRequest.
- **A5**: Shell wrapper scripts follow the established `.sh.tmpl` Go template pattern with `posixPath` helpers.
- **A6**: The `settings.json.tmpl` template structure supports all Claude Code hook event names as top-level keys under `hooks`.
- **A7**: Exit code 2 semantics for TeammateIdle and TaskCompleted are handled by the `runHookEvent` function in `hook.go`, which already checks for deny decisions (line 90-92). This may need extension for exit code 2 with non-deny decisions.

---

## 3. Requirements

### 3.1 Event Type Registration

**REQ-HOOK-008-001** (Ubiquitous)
The system shall define EventType constants for all 7 missing Claude Code hook events: `EventPostToolUseFailure`, `EventNotification`, `EventSubagentStart`, `EventUserPromptSubmit`, `EventPermissionRequest`, `EventTeammateIdle`, `EventTaskCompleted`.

**REQ-HOOK-008-002** (Ubiquitous)
The system shall include all 7 new EventType constants in the `ValidEventTypes()` function return value.

**REQ-HOOK-008-003** (Ubiquitous)
The `IsValidEventType()` function shall return true for all 7 new event types.

### 3.2 Handler Implementations

**REQ-HOOK-008-010** (Event-Driven)
When a PostToolUseFailure event is received, the system shall log the tool failure including the `error` and `is_interrupt` fields, and return an empty HookOutput to allow default processing.

**REQ-HOOK-008-011** (Event-Driven)
When a PostToolUseFailure handler returns `decision: "block"`, the system shall prevent further processing of the failed tool result.

**REQ-HOOK-008-012** (Event-Driven)
When a Notification event is received, the system shall log the notification including `message`, `title`, and `notification_type` fields, and return an empty HookOutput.

**REQ-HOOK-008-013** (Event-Driven)
When a SubagentStart event is received, the system shall log the subagent startup including `agent_id` and `agent_transcript_path` fields, and return an empty HookOutput.

**REQ-HOOK-008-014** (Event-Driven)
When a UserPromptSubmit event is received, the system shall log the submitted prompt and return an empty HookOutput to allow default processing.

**REQ-HOOK-008-015** (Event-Driven)
When a UserPromptSubmit handler returns `decision: "block"`, the system shall prevent the prompt from being submitted.

**REQ-HOOK-008-016** (Event-Driven)
When a PermissionRequest event is received, the system shall log the permission request including `tool_name` and `tool_input` fields, and return a HookOutput with `hookSpecificOutput.permissionDecision` set to "ask" (deferring to default permission behavior).

**REQ-HOOK-008-017** (Event-Driven)
When a TeammateIdle event is received, the system shall log the idle notification and return an empty HookOutput to accept the idle state.

**REQ-HOOK-008-018** (Event-Driven)
When a TaskCompleted event is received, the system shall log the completion and return an empty HookOutput to accept the completion.

### 3.3 CLI Integration

**REQ-HOOK-008-020** (Ubiquitous)
The system shall register CLI subcommands for all 7 new events under the `moai hook` command group.

The required subcommand mappings are:

| Subcommand          | Event Constant            |
| ------------------- | ------------------------- |
| post-tool-failure   | EventPostToolUseFailure   |
| notification        | EventNotification         |
| subagent-start      | EventSubagentStart        |
| user-prompt-submit  | EventUserPromptSubmit     |
| permission-request  | EventPermissionRequest    |
| teammate-idle       | EventTeammateIdle         |
| task-completed      | EventTaskCompleted        |

**REQ-HOOK-008-021** (Event-Driven)
When a hook subcommand is executed, the system shall read JSON from stdin, dispatch to registered handlers, and write JSON to stdout, following the same pattern as existing subcommands.

**REQ-HOOK-008-022** (Event-Driven)
When a TeammateIdle or TaskCompleted handler signals "keep working" or "reject completion", the CLI shall exit with code 2 per Claude Code Agent Teams protocol.

### 3.4 Handler Registration

**REQ-HOOK-008-030** (Ubiquitous)
The system shall register all 7 new handlers in the Composition Root (`internal/cli/deps.go`).

### 3.5 Settings Template

**REQ-HOOK-008-040** (Optional)
Where the PostToolUseFailure event is implemented, the `settings.json.tmpl` shall include a hook configuration entry for PostToolUseFailure.

**REQ-HOOK-008-041** (Optional)
Where the Notification event is implemented, the `settings.json.tmpl` shall include a hook configuration entry for Notification.

**REQ-HOOK-008-042** (Optional)
Where the SubagentStart event is implemented, the `settings.json.tmpl` shall include a hook configuration entry for SubagentStart.

**REQ-HOOK-008-043** (Optional)
Where the UserPromptSubmit event is implemented, the `settings.json.tmpl` shall include a hook configuration entry for UserPromptSubmit.

**REQ-HOOK-008-044** (Optional)
Where the PermissionRequest event is implemented, the `settings.json.tmpl` shall include a hook configuration entry for PermissionRequest.

**REQ-HOOK-008-045** (Optional)
Where the TeammateIdle event is implemented, the `settings.json.tmpl` shall include a hook configuration entry for TeammateIdle.

**REQ-HOOK-008-046** (Optional)
Where the TaskCompleted event is implemented, the `settings.json.tmpl` shall include a hook configuration entry for TaskCompleted.

### 3.6 Shell Wrapper Scripts

**REQ-HOOK-008-050** (Event-Driven)
When a new hook event is registered in `settings.json.tmpl`, the system shall provide a corresponding shell wrapper script at `.claude/hooks/moai/handle-{event-name}.sh.tmpl` following the established Go template pattern.

**REQ-HOOK-008-051** (Ubiquitous)
Each shell wrapper script shall follow the established pattern: create a temp file from stdin, attempt moai binary resolution via PATH/GoBinPath/HomeDir fallback, and forward to `moai hook {subcommand}`.

### 3.7 Registry Default Output

**REQ-HOOK-008-060** (Event-Driven)
When no handlers are registered for a new event type, the registry's `defaultOutputForEvent()` shall return the protocol-appropriate default output for that event type.

The required defaults are:

| Event               | Default Output                                          |
| ------------------- | ------------------------------------------------------- |
| PostToolUseFailure  | Empty HookOutput `{}`                                   |
| Notification        | Empty HookOutput `{}`                                   |
| SubagentStart       | Empty HookOutput `{}`                                   |
| UserPromptSubmit    | Empty HookOutput `{}`                                   |
| PermissionRequest   | HookOutput with hookSpecificOutput.permissionDecision="ask" |
| TeammateIdle        | Empty HookOutput `{}`                                   |
| TaskCompleted       | Empty HookOutput `{}`                                   |

### 3.8 Exit Code Handling

**REQ-HOOK-008-070** (Event-Driven)
When the TeammateIdle handler output indicates the teammate should continue working, the `runHookEvent` function shall exit with code 2.

**REQ-HOOK-008-071** (Event-Driven)
When the TaskCompleted handler output indicates completion rejection, the `runHookEvent` function shall exit with code 2.

### 3.9 UserPromptSubmit Output Extension

**REQ-HOOK-008-080** (Optional)
Where prompt modification is needed, the HookOutput struct shall support an `updatedInput` field that allows UserPromptSubmit handlers to modify the user's prompt before processing.

---

## 4. Non-Functional Requirements

**REQ-HOOK-008-NF-001** (Ubiquitous)
The system shall maintain backward compatibility with all existing hook handlers and their behavior.

**REQ-HOOK-008-NF-002** (Ubiquitous)
All new handler implementations shall achieve at least 85% test coverage with table-driven tests.

**REQ-HOOK-008-NF-003** (Unwanted)
The system shall not introduce any regressions in existing hook handler tests.

**REQ-HOOK-008-NF-004** (Ubiquitous)
All new handler files shall follow the established naming convention: `{event_name}.go` with corresponding `{event_name}_test.go`.

**REQ-HOOK-008-NF-005** (Ubiquitous)
All new code shall pass `go vet`, `go test -race`, and `golangci-lint run` without errors.

**REQ-HOOK-008-NF-006** (Ubiquitous)
Each handler shall complete within the default 30-second timeout configured in the registry.

**REQ-HOOK-008-NF-007** (Ubiquitous)
The `make build` command shall succeed after all changes, regenerating embedded template files.

---

## 5. Specifications

### 5.1 New EventType Constants (types.go)

```go
const (
    EventPostToolUseFailure EventType = "PostToolUseFailure"
    EventNotification       EventType = "Notification"
    EventSubagentStart      EventType = "SubagentStart"
    EventUserPromptSubmit   EventType = "UserPromptSubmit"
    EventPermissionRequest  EventType = "PermissionRequest"
    EventTeammateIdle       EventType = "TeammateIdle"
    EventTaskCompleted      EventType = "TaskCompleted"
)
```

### 5.2 HookOutput Extension (types.go)

```go
type HookOutput struct {
    // ... existing fields ...

    // For UserPromptSubmit: modified prompt text
    UpdatedInput string `json:"updatedInput,omitempty"`
}
```

### 5.3 New Handler Files

| File                      | Handler Type              | Event                     |
| ------------------------- | ------------------------- | ------------------------- |
| post_tool_failure.go      | postToolUseFailureHandler | EventPostToolUseFailure   |
| notification.go           | notificationHandler       | EventNotification         |
| subagent_start.go         | subagentStartHandler      | EventSubagentStart        |
| user_prompt_submit.go     | userPromptSubmitHandler   | EventUserPromptSubmit     |
| permission_request.go     | permissionRequestHandler  | EventPermissionRequest    |
| teammate_idle.go          | teammateIdleHandler       | EventTeammateIdle         |
| task_completed.go         | taskCompletedHandler      | EventTaskCompleted        |

### 5.4 Exit Code 2 Protocol

For TeammateIdle and TaskCompleted, exit code 2 signals a special state to Claude Code:
- **TeammateIdle**: Exit code 2 means "keep working" (reject idle state)
- **TaskCompleted**: Exit code 2 means "reject completion" (task not actually done)

The `runHookEvent` function in `hook.go` must be extended to support this protocol. A new field or convention on HookOutput is needed to signal exit code 2 without using `Decision: "deny"`.

### 5.5 CLI Subcommand Registration

All 7 new subcommands are added to the `hookSubcommands` table in `internal/cli/hook.go`:

```go
{"post-tool-failure",  "Handle post-tool-use failure event",  hook.EventPostToolUseFailure},
{"notification",       "Handle notification event",           hook.EventNotification},
{"subagent-start",     "Handle subagent start event",         hook.EventSubagentStart},
{"user-prompt-submit", "Handle user prompt submit event",     hook.EventUserPromptSubmit},
{"permission-request", "Handle permission request event",     hook.EventPermissionRequest},
{"teammate-idle",      "Handle teammate idle event",          hook.EventTeammateIdle},
{"task-completed",     "Handle task completed event",         hook.EventTaskCompleted},
```

---

## 6. Traceability

| Requirement        | Implementation File(s)                                      | Test File(s)                    |
| ------------------ | ----------------------------------------------------------- | ------------------------------- |
| REQ-HOOK-008-001   | internal/hook/types.go                                      | internal/hook/types_test.go     |
| REQ-HOOK-008-002   | internal/hook/types.go                                      | internal/hook/types_test.go     |
| REQ-HOOK-008-010-018 | internal/hook/{handler}.go                                | internal/hook/{handler}_test.go |
| REQ-HOOK-008-020   | internal/cli/hook.go                                        | (CLI integration tests)         |
| REQ-HOOK-008-030   | internal/cli/deps.go                                        | (composition root tests)        |
| REQ-HOOK-008-040-046 | internal/template/templates/.claude/settings.json.tmpl    | (template render tests)         |
| REQ-HOOK-008-050-051 | internal/template/templates/.claude/hooks/moai/*.sh.tmpl  | (template tests)                |
| REQ-HOOK-008-060   | internal/hook/registry.go                                   | internal/hook/registry_test.go  |
| REQ-HOOK-008-070-071 | internal/cli/hook.go                                      | (CLI integration tests)         |
| REQ-HOOK-008-080   | internal/hook/types.go                                      | internal/hook/types_test.go     |
