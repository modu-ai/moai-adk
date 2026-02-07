# Hooks System

Claude Code hooks for extending functionality with custom scripts.

## Hook Events

All 14 available hook event types:

| Event | Matcher | Can Block | Description |
|-------|---------|-----------|-------------|
| UserPromptSubmit | No | Yes | Runs when user submits a prompt, before processing |
| SessionStart | No | No | Runs when a new session begins |
| PreCompact | No | No | Runs before context compaction |
| PreToolUse | Tool name | Yes | Runs before a tool executes |
| PostToolUse | Tool name | No | Runs after a tool completes successfully |
| PostToolUseFailure | Tool name | No | Runs after a tool execution fails |
| PermissionRequest | Tool name | Yes | Runs when permission dialog appears |
| Notification | Type | No | Runs when Claude Code sends notifications |
| SubagentStart | Agent type | No | Runs when a subagent spawns |
| SubagentStop | No | No | Runs when a subagent terminates |
| Stop | No | No | Runs when conversation stops |
| TeammateIdle | No | Yes | Runs when agent team teammate is about to go idle |
| TaskCompleted | No | Yes | Runs when a task is being marked complete |
| SessionEnd | Reason | No | Runs when session terminates |

### Event Categories

**Lifecycle Events**: SessionStart, SessionEnd, Stop, PreCompact

**Prompt Events**: UserPromptSubmit, PermissionRequest, Notification

**Tool Events**: PreToolUse, PostToolUse, PostToolUseFailure

**Agent Events**: SubagentStart, SubagentStop, TeammateIdle, TaskCompleted

## Hook Event Reference

### UserPromptSubmit

Runs when the user submits a prompt, before Claude processes it. Can add context to the prompt or block it entirely.

- Matcher: None
- Blocking: Yes (exit code 2 blocks the prompt)
- stdin fields: `prompt` (user's submitted text)
- stdout: JSON with optional `additionalContext` (string injected into prompt) or `reason` (when blocking)

### PermissionRequest

Runs when a permission dialog would appear for a tool. Matches on tool name. Can auto-allow or auto-deny.

- Matcher: Tool name (e.g., `Write|Edit|Bash`)
- Blocking: Yes (exit code 0 = allow, exit code 2 = deny)
- stdin fields: `toolName`, `toolInput`
- stdout: JSON with optional `reason` (when denying)

### PostToolUseFailure

Runs after a tool execution fails. Matches on tool name. Receives error details.

- Matcher: Tool name (e.g., `Bash|Write`)
- Blocking: No
- stdin fields: `toolName`, `toolInput`, `error` (error message), `is_interrupt` (boolean, true if user interrupted)
- stdout: JSON with optional `systemMessage`

### Notification

Runs when Claude Code sends a notification. Matches on notification type.

- Matcher: Notification type
- Blocking: No
- Notification types: `permission_prompt`, `idle_prompt`, `auth_success`, `elicitation_dialog`
- stdin fields: `type` (notification type), `message`

### SubagentStart

Runs when a subagent spawns. Matches on agent type name. Can inject additional context into the subagent.

- Matcher: Agent type name
- Blocking: No
- stdin fields: `agentType`, `agentName`
- stdout: JSON with optional `additionalContext` (string injected into the subagent prompt)

### TeammateIdle

Runs when an agent team teammate is about to go idle. Exit code 2 keeps the teammate working instead of idling.

- Matcher: None
- Blocking: Yes (exit code 2 = keep working)
- stdin fields: `agentType`, `agentName`, `tasksSummary`
- stdout: JSON with optional `systemMessage` (guidance for continued work)
- **Critical for team quality enforcement**: Use to verify quality gates before allowing idle

### TaskCompleted

Runs when a task is being marked complete. Exit code 2 prevents completion and sends the task back for more work.

- Matcher: None
- Blocking: Yes (exit code 2 = reject completion)
- stdin fields: `taskId`, `taskSummary`, `agentName`
- stdout: JSON with optional `reason` (why completion was rejected)
- **Critical for team quality enforcement**: Use to verify deliverables meet standards

### SessionEnd

Runs when the session terminates. Matches on the termination reason. Cannot block termination.

- Matcher: Reason (`clear`, `logout`, `prompt_input_exit`, `bypass_permissions_disabled`, `other`)
- Blocking: No
- stdin fields: `reason` (termination reason), `sessionId`
- stdout: Ignored (session is ending)

## Hook Execution Types

### Command Hooks (type: "command")

Default hook type. Executes a shell command, communicates via stdin/stdout JSON.

- Configuration: `type`, `command`, `timeout`
- stdin: JSON with event data
- stdout: JSON with response (optional `systemMessage`, `additionalContext`, `reason`)
- Exit codes: 0 = success, 1 = error (shown to user), 2 = block/reject (for blocking events)

### Prompt Hooks (type: "prompt")

Send hook input to an LLM for single-turn evaluation. The LLM receives the event data and returns a judgment.

- Configuration: `type`, `prompt`, `model`, `timeout`
- The `prompt` field contains instructions for the LLM evaluator
- Returns JSON: `ok` (boolean), `reason` (string explanation)
- When `ok` is false on a blocking event, the operation is blocked with the provided reason

### Agent Hooks (type: "agent")

Spawn a subagent with tool access to verify conditions. The agent can read files, search code, and make informed decisions.

- Configuration: `type`, `prompt`, `model`, `timeout`
- Agent has access to: Read, Grep, Glob
- Returns JSON: `ok` (boolean), `reason` (string explanation)
- Same blocking behavior as prompt hooks

### Async Command Hooks (async: true)

Run command hooks in the background without blocking the conversation.

- Only available for `type: "command"` hooks
- Configuration: Add `async: true` to any command hook definition
- Results are delivered on the next conversation turn via `systemMessage`
- Useful for long-running validations (linting, test execution, deployments)

## Agent Hooks

Agent-specific hooks are defined in agent frontmatter (`.claude/agents/**/*.md`) and are executed for specific agent lifecycle events. These hooks use the `handle-agent-hook.sh` wrapper script.

### Agent Hook Configuration

Hooks are defined in agent YAML frontmatter:

```yaml
---
name: manager-ddd
description: DDD workflow specialist
hooks:
  PreToolUse:
    - matcher: "Write|Edit|MultiEdit"
      hooks:
        - type: command
          command: "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-agent-hook.sh\" ddd-pre-transformation"
          timeout: 5
  PostToolUse:
    - matcher: "Write|Edit|MultiEdit"
      hooks:
        - type: command
          command: "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-agent-hook.sh\" ddd-post-transformation"
          timeout: 10
  SubagentStop:
    hooks:
      - type: command
        command: "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-agent-hook.sh\" ddd-completion"
        timeout: 10
---
```

### Agent Hook Actions

Available agent hook actions:

| Agent | Action | Event | Purpose |
|-------|--------|-------|---------|
| manager-ddd | ddd-pre-transformation | PreToolUse | Check characterization tests before code changes |
| manager-ddd | ddd-post-transformation | PostToolUse | Verify behavior preservation after changes |
| manager-ddd | ddd-completion | SubagentStop | Report DDD workflow completion |
| manager-tdd | tdd-pre-implementation | PreToolUse | Ensure test exists (RED phase) |
| manager-tdd | tdd-post-implementation | PostToolUse | Verify tests pass (GREEN phase) |
| manager-tdd | tdd-completion | SubagentStop | Report TDD workflow completion |
| expert-backend | backend-validation | PreToolUse | Validate backend code before changes |
| expert-backend | backend-verification | PostToolUse | Verify backend code after changes |
| expert-frontend | frontend-validation | PreToolUse | Validate frontend code before changes |
| expert-frontend | frontend-verification | PostToolUse | Verify frontend code after changes |
| expert-testing | testing-verification | PostToolUse | Verify test quality |
| expert-testing | testing-completion | SubagentStop | Report testing workflow completion |
| expert-debug | debug-verification | PostToolUse | Verify debugging results |
| expert-debug | debug-completion | SubagentStop | Report debugging completion |
| expert-devops | devops-verification | PostToolUse | Verify DevOps configurations |
| expert-devops | devops-completion | SubagentStop | Report DevOps workflow completion |
| manager-quality | quality-completion | SubagentStop | Report quality validation completion |
| manager-spec | spec-completion | SubagentStop | Report SPEC document generation completion |
| manager-docs | docs-verification | PostToolUse | Verify documentation quality |
| manager-docs | docs-completion | SubagentStop | Report documentation generation completion |

### Hook Command Interface

Agent hooks are executed via the `moai hook agent <action>` command:

```bash
moai hook agent ddd-pre-transformation
moai hook agent backend-validation
moai hook agent tdd-completion
```

The hook receives JSON input via stdin with the following structure:

```json
{
  "eventType": "SubagentStop",
  "toolName": "",
  "toolInput": null,
  "toolOutput": null,
  "session": {
    "id": "sess-123",
    "cwd": "/path/to/project",
    "projectDir": "/path/to/project"
  },
  "data": {
    "agent": "manager-ddd",
    "action": "ddd-completion"
  }
}
```

### Agent Handler Factory

The `internal/hook/agents/factory.go` file implements the factory pattern for creating agent-specific handlers. Each agent type has its own handler file:

- `ddd_handler.go`: DDD workflow hooks
- `tdd_handler.go`: TDD workflow hooks
- `backend_handler.go`: Backend expert hooks
- `frontend_handler.go`: Frontend expert hooks
- `testing_handler.go`: Testing expert hooks
- `debug_handler.go`: Debug expert hooks
- `devops_handler.go`: DevOps expert hooks
- `quality_handler.go`: Quality manager hooks
- `spec_handler.go`: SPEC manager hooks
- `docs_handler.go`: Documentation manager hooks
- `default_handler.go`: Default handler for unknown actions

## Hook Location

Hooks are defined in `.claude/hooks/` directory:

- Shell scripts: `*.sh`
- Python scripts: `*.py`

## Configuration

Define hooks in `.claude/settings.json`:

```json
{
  "hooks": {
    "SessionStart": [{
      "type": "command",
      "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-session-start.sh\"",
      "timeout": 5
    }],
    "PreCompact": [{
      "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-compact.sh\"",
      "timeout": 5
    }],
    "PreToolUse": [{
      "matcher": "Write|Edit|Bash",
      "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-pre-tool.sh\"",
      "timeout": 5
    }],
    "PostToolUse": [{
      "matcher": "Write|Edit",
      "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-post-tool.sh\"",
      "timeout": 60
    }],
    "Stop": [{
      "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-stop.sh\"",
      "timeout": 5
    }]
  }
}
```

## Path Syntax Rules

### Hooks (Support Environment Variables)

Hooks support `$CLAUDE_PROJECT_DIR` and `$HOME` environment variables:

```json
{
  "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/hook.sh\""
}
```

**Important**: Quote the entire path to handle project folders with spaces:
- Correct: `"\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/hook.sh\""`
- Wrong: `"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/hook.sh"`

### StatusLine (No Environment Variable Support)

StatusLine does NOT support environment variable expansion (GitHub Issue #7925). Use relative paths from project root:

```json
{
  "statusLine": {
    "type": "command",
    "command": ".moai/status_line.sh"
  }
}
```

## Hook Wrappers

MoAI-ADK generates hook wrapper scripts during `moai init` that:

1. Read stdin JSON from Claude Code
2. Forward it to the moai binary via `moai hook <event>` command
3. Support multiple moai binary locations:
   - `moai` command in PATH
   - Detected Go bin path from initialization
   - Default `~/go/bin/moai`

Wrapper scripts are located at:
- `.claude/hooks/moai/handle-session-start.sh`
- `.claude/hooks/moai/handle-compact.sh`
- `.claude/hooks/moai/handle-pre-tool.sh`
- `.claude/hooks/moai/handle-post-tool.sh`
- `.claude/hooks/moai/handle-stop.sh`

## Rules

- Hook feedback is treated as user input
- When blocked, suggest alternatives
- Avoid infinite loops (no recursive tool calls)
- Keep hooks lightweight for performance
- Use proper path quoting to handle spaces in project paths
- StatusLine uses relative paths only (no env var expansion)
- Prompt and agent hooks return JSON with `ok` and `reason` fields
- Async hooks deliver results via `systemMessage` on the next turn
- Exit code 2 is the universal "block/reject" signal for blocking events

## Error Handling

- Failed hooks should exit with non-zero code
- Error messages are displayed to user
- Hooks can block operations by returning error
- Missing hooks exit silently (Claude Code handles gracefully)
- Prompt/agent hooks that fail return `ok: false` with a reason

## Security

- Hooks run in sandbox by default
- Validate all hook inputs
- Do not store secrets in hook scripts
- Agent hooks (type: "agent") have read-only tool access (Read, Grep, Glob)

## MoAI Integration

- Skill("moai-foundation-claude") for detailed patterns
- Hook scripts must follow coding-standards.md
- Hook wrappers are managed by `internal/hook/` package
- TeammateIdle and TaskCompleted hooks are critical for Agent Teams quality enforcement
