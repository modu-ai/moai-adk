# SPEC-HOOK-008: Acceptance Criteria

## Metadata

| Field   | Value                                                  |
| ------- | ------------------------------------------------------ |
| SPEC ID | SPEC-HOOK-008                                          |
| Title   | Complete Hook Event System - Add 7 Missing Hook Events |

---

## AC-001: EventType Constants Registered

**Requirement**: REQ-HOOK-008-001, REQ-HOOK-008-002, REQ-HOOK-008-003

```gherkin
Given the hook package is compiled
When ValidEventTypes() is called
Then the returned slice shall contain exactly 14 event types
And the slice shall include EventPostToolUseFailure with value "PostToolUseFailure"
And the slice shall include EventNotification with value "Notification"
And the slice shall include EventSubagentStart with value "SubagentStart"
And the slice shall include EventUserPromptSubmit with value "UserPromptSubmit"
And the slice shall include EventPermissionRequest with value "PermissionRequest"
And the slice shall include EventTeammateIdle with value "TeammateIdle"
And the slice shall include EventTaskCompleted with value "TaskCompleted"

Given the hook package is compiled
When IsValidEventType() is called with each of the 7 new EventType values
Then it shall return true for every new event type
```

---

## AC-002: PostToolUseFailure Handler

**Requirement**: REQ-HOOK-008-010, REQ-HOOK-008-011

```gherkin
Given a PostToolUseFailure handler is created via NewPostToolUseFailureHandler()
When Handle() is called with a HookInput containing error="tool crashed" and is_interrupt=false
Then the handler shall return a non-nil HookOutput
And the HookOutput shall have empty Decision field
And the handler shall not return an error

Given a PostToolUseFailure handler is created
When Handle() is called with a HookInput containing error="" (empty error)
Then the handler shall return a non-nil HookOutput without error

Given a PostToolUseFailure handler is created
When EventType() is called
Then it shall return EventPostToolUseFailure
```

---

## AC-003: Notification Handler

**Requirement**: REQ-HOOK-008-012

```gherkin
Given a Notification handler is created via NewNotificationHandler()
When Handle() is called with a HookInput containing message="Build complete", title="CI", notification_type="info"
Then the handler shall return a non-nil HookOutput
And the HookOutput shall have empty Decision field
And the handler shall not return an error

Given a Notification handler is created
When Handle() is called with a HookInput containing all empty notification fields
Then the handler shall return a non-nil HookOutput without error

Given a Notification handler is created
When EventType() is called
Then it shall return EventNotification
```

---

## AC-004: SubagentStart Handler

**Requirement**: REQ-HOOK-008-013

```gherkin
Given a SubagentStart handler is created via NewSubagentStartHandler()
When Handle() is called with a HookInput containing agent_id="agent-123" and agent_transcript_path="/tmp/transcript.json"
Then the handler shall return a non-nil HookOutput
And the HookOutput shall have empty Decision field
And the handler shall not return an error

Given a SubagentStart handler is created
When Handle() is called with a HookInput containing empty agent fields
Then the handler shall return a non-nil HookOutput without error

Given a SubagentStart handler is created
When EventType() is called
Then it shall return EventSubagentStart
```

---

## AC-005: UserPromptSubmit Handler

**Requirement**: REQ-HOOK-008-014, REQ-HOOK-008-015

```gherkin
Given a UserPromptSubmit handler is created via NewUserPromptSubmitHandler()
When Handle() is called with a HookInput containing prompt="Fix the authentication bug"
Then the handler shall return a non-nil HookOutput
And the HookOutput shall have empty Decision field
And the handler shall not return an error

Given a UserPromptSubmit handler is created
When Handle() is called with a HookInput containing an empty prompt
Then the handler shall return a non-nil HookOutput without error

Given a UserPromptSubmit handler is created
When EventType() is called
Then it shall return EventUserPromptSubmit
```

---

## AC-006: PermissionRequest Handler

**Requirement**: REQ-HOOK-008-016

```gherkin
Given a PermissionRequest handler is created via NewPermissionRequestHandler()
When Handle() is called with a HookInput containing tool_name="Bash" and tool_input with command data
Then the handler shall return a non-nil HookOutput
And the HookOutput shall have a non-nil HookSpecificOutput
And the HookSpecificOutput.PermissionDecision shall equal "ask"
And the handler shall not return an error

Given a PermissionRequest handler is created
When Handle() is called with a HookInput containing empty tool fields
Then the handler shall return a HookOutput with PermissionDecision "ask"

Given a PermissionRequest handler is created
When EventType() is called
Then it shall return EventPermissionRequest
```

---

## AC-007: TeammateIdle Handler

**Requirement**: REQ-HOOK-008-017

```gherkin
Given a TeammateIdle handler is created via NewTeammateIdleHandler()
When Handle() is called with a HookInput containing agent_id="teammate-backend-dev"
Then the handler shall return a non-nil HookOutput
And the HookOutput shall have ExitCode equal to 0 (accept idle)
And the handler shall not return an error

Given a TeammateIdle handler is created
When EventType() is called
Then it shall return EventTeammateIdle
```

---

## AC-008: TaskCompleted Handler

**Requirement**: REQ-HOOK-008-018

```gherkin
Given a TaskCompleted handler is created via NewTaskCompletedHandler()
When Handle() is called with a HookInput containing session context
Then the handler shall return a non-nil HookOutput
And the HookOutput shall have ExitCode equal to 0 (accept completion)
And the handler shall not return an error

Given a TaskCompleted handler is created
When EventType() is called
Then it shall return EventTaskCompleted
```

---

## AC-009: CLI Subcommand Registration

**Requirement**: REQ-HOOK-008-020, REQ-HOOK-008-021

```gherkin
Given the moai CLI is initialized
When "moai hook --help" is executed
Then the output shall list subcommands for all 7 new events:
  | Subcommand          |
  | post-tool-failure   |
  | notification        |
  | subagent-start      |
  | user-prompt-submit  |
  | permission-request  |
  | teammate-idle       |
  | task-completed      |

Given the moai CLI is initialized with hook handlers registered
When "moai hook post-tool-failure" is executed with valid JSON on stdin
Then the command shall write valid JSON to stdout
And the command shall exit with code 0
```

---

## AC-010: Exit Code 2 Protocol

**Requirement**: REQ-HOOK-008-022, REQ-HOOK-008-070, REQ-HOOK-008-071

```gherkin
Given the runHookEvent function processes a TeammateIdle event
When the handler returns a HookOutput with ExitCode=2
Then the process shall exit with code 2

Given the runHookEvent function processes a TaskCompleted event
When the handler returns a HookOutput with ExitCode=2
Then the process shall exit with code 2

Given the runHookEvent function processes a TeammateIdle event
When the handler returns a HookOutput with ExitCode=0 (default)
Then the process shall exit with code 0
```

---

## AC-011: Handler Registration in Composition Root

**Requirement**: REQ-HOOK-008-030

```gherkin
Given InitDependencies() is called
When the hook registry is inspected
Then handlers shall be registered for all 7 new event types:
  | Event Type            |
  | PostToolUseFailure    |
  | Notification          |
  | SubagentStart         |
  | UserPromptSubmit      |
  | PermissionRequest     |
  | TeammateIdle          |
  | TaskCompleted         |
```

---

## AC-012: Registry Default Output

**Requirement**: REQ-HOOK-008-060

```gherkin
Given a registry with no handlers registered for PostToolUseFailure
When Dispatch() is called for EventPostToolUseFailure
Then the output shall be an empty HookOutput

Given a registry with no handlers registered for Notification
When Dispatch() is called for EventNotification
Then the output shall be an empty HookOutput

Given a registry with no handlers registered for SubagentStart
When Dispatch() is called for EventSubagentStart
Then the output shall be an empty HookOutput

Given a registry with no handlers registered for UserPromptSubmit
When Dispatch() is called for EventUserPromptSubmit
Then the output shall be an empty HookOutput

Given a registry with no handlers registered for PermissionRequest
When Dispatch() is called for EventPermissionRequest
Then the output shall have HookSpecificOutput with PermissionDecision="ask"

Given a registry with no handlers registered for TeammateIdle
When Dispatch() is called for EventTeammateIdle
Then the output shall be an empty HookOutput

Given a registry with no handlers registered for TaskCompleted
When Dispatch() is called for EventTaskCompleted
Then the output shall be an empty HookOutput
```

---

## AC-013: Settings Template Configuration

**Requirement**: REQ-HOOK-008-040 through REQ-HOOK-008-046

```gherkin
Given the settings.json.tmpl is rendered
When the output JSON is parsed
Then the "hooks" object shall contain keys for all 7 new events:
  | Hook Key             |
  | PostToolUseFailure   |
  | Notification         |
  | SubagentStart        |
  | UserPromptSubmit     |
  | PermissionRequest    |
  | TeammateIdle         |
  | TaskCompleted        |
And each key shall reference a shell wrapper script in .claude/hooks/moai/
And each hook entry shall have "type": "command" and a "timeout" value
```

---

## AC-014: Shell Wrapper Scripts

**Requirement**: REQ-HOOK-008-050, REQ-HOOK-008-051

```gherkin
Given the template directory at internal/template/templates/.claude/hooks/moai/
Then the following .sh.tmpl files shall exist:
  | File                               |
  | handle-post-tool-failure.sh.tmpl   |
  | handle-notification.sh.tmpl        |
  | handle-subagent-start.sh.tmpl      |
  | handle-user-prompt-submit.sh.tmpl  |
  | handle-permission-request.sh.tmpl  |
  | handle-teammate-idle.sh.tmpl       |
  | handle-task-completed.sh.tmpl      |

Given any new shell wrapper template file
When its content is inspected
Then it shall contain:
  - A shebang line "#!/bin/bash"
  - A temp file creation with mktemp
  - A trap for cleanup on EXIT
  - A stdin redirect via cat
  - A PATH-based moai binary lookup
  - A GoBinPath fallback using posixPath template function
  - A HomeDir fallback using posixPath template function
  - A silent exit 0 fallback
```

---

## AC-015: HookOutput Extension

**Requirement**: REQ-HOOK-008-080

```gherkin
Given a HookOutput with UpdatedInput="modified prompt text"
When the output is serialized to JSON
Then the JSON shall contain "updatedInput":"modified prompt text"

Given a HookOutput with UpdatedInput="" (empty)
When the output is serialized to JSON
Then the JSON shall not contain the "updatedInput" key (omitempty)

Given a HookOutput with ExitCode=2
When the output is serialized to JSON
Then the JSON shall not contain the ExitCode field (json:"-" tag)
```

---

## AC-016: Backward Compatibility

**Requirement**: REQ-HOOK-008-NF-001, REQ-HOOK-008-NF-003

```gherkin
Given all changes from SPEC-HOOK-008 are applied
When "go test -race ./internal/hook/..." is executed
Then all existing tests shall pass without modification
And no existing handler behavior shall change

Given all changes from SPEC-HOOK-008 are applied
When "go test -race ./internal/cli/..." is executed
Then all existing CLI tests shall pass without modification
```

---

## AC-017: Test Coverage

**Requirement**: REQ-HOOK-008-NF-002

```gherkin
Given all new handler files are created
When "go test -cover ./internal/hook/..." is executed
Then each new handler file shall have at least 85% coverage
And each test file shall use table-driven test patterns
And tests shall cover: normal input, empty input, context cancellation
```

---

## AC-018: Build Verification

**Requirement**: REQ-HOOK-008-NF-005, REQ-HOOK-008-NF-007

```gherkin
Given all changes from SPEC-HOOK-008 are applied
When "go vet ./..." is executed
Then no errors shall be reported

Given all changes from SPEC-HOOK-008 are applied
When "go test -race ./..." is executed
Then no race conditions shall be detected

Given all changes from SPEC-HOOK-008 are applied
When "make build" is executed
Then the build shall succeed
And embedded template files shall be regenerated
And the binary shall include all new shell wrapper templates
```

---

## AC-019: Hook List Command

**Requirement**: REQ-HOOK-008-020 (implied)

```gherkin
Given all new handlers are registered
When "moai hook list" is executed
Then the output shall show handler counts for all 14 event types
And each new event type shall show at least 1 handler
```

---

## Quality Gate Criteria

### Definition of Done

- [ ] All 7 EventType constants defined in types.go
- [ ] ValidEventTypes() returns 14 event types
- [ ] 7 new handler files created with corresponding test files
- [ ] All handlers registered in deps.go Composition Root
- [ ] 7 CLI subcommands added to hook.go
- [ ] Exit code 2 handling implemented for TeammateIdle and TaskCompleted
- [ ] Registry defaultOutputForEvent updated for all new events
- [ ] settings.json.tmpl updated with 7 new hook entries
- [ ] 7 shell wrapper .sh.tmpl files created
- [ ] HookOutput extended with UpdatedInput and ExitCode fields
- [ ] All existing tests pass (`go test -race ./...`)
- [ ] All new handlers have 85%+ test coverage
- [ ] `go vet ./...` passes
- [ ] `make build` succeeds
- [ ] No regressions in existing hook behavior

### Verification Commands

```bash
# Unit tests with race detection
go test -race ./internal/hook/...
go test -race ./internal/cli/...

# Coverage verification
go test -cover ./internal/hook/...

# Static analysis
go vet ./...

# Build verification (includes template embedding)
make build

# Full test suite
go test -race ./...
```
