---
paths:
  - "**/.claude/hooks/**"
  - "**/.claude/settings.json"
  - "**/.claude/settings.local.json"
---

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

## Hook Event stdin/stdout Reference

| Event | stdin | stdout | Notes |
|-------|-------|--------|-------|
| UserPromptSubmit | `prompt` | `additionalContext`, `reason` | Exit 2 blocks prompt |
| PermissionRequest | `toolName`, `toolInput` | `reason` | Exit 0 = allow, exit 2 = deny |
| PostToolUseFailure | `toolName`, `toolInput`, `error`, `is_interrupt` | `systemMessage` | Non-blocking |
| Notification | `type`, `message` | - | Types: permission_prompt, idle_prompt, auth_success, elicitation_dialog |
| SubagentStart | `agentType`, `agentName` | `additionalContext` | Inject context into subagent |
| TeammateIdle | `agentType`, `agentName`, `tasksSummary` | `systemMessage` | Exit 2 = keep working. Critical for team quality |
| TaskCompleted | `taskId`, `taskSummary`, `agentName` | `reason` | Exit 2 = reject completion. Critical for team quality |
| SessionEnd | `reason`, `sessionId` | - | Reasons: clear, logout, prompt_input_exit, bypass_permissions_disabled, other |

Standard events (SessionStart, PreCompact, PreToolUse, PostToolUse, Stop) use common stdin/stdout patterns: stdin receives event-specific fields, stdout accepts optional `systemMessage`.

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

## Agent-Specific Hooks

Agent hooks are defined in agent frontmatter and executed for agent lifecycle events. For detailed configuration, actions table, and handler architecture, see @agent-hooks.md.

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
    }],
    "TeammateIdle": [{
      "hooks": [{
        "type": "command",
        "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-agent-hook.sh\"",
        "timeout": 10
      }]
    }],
    "TaskCompleted": [{
      "hooks": [{
        "type": "command",
        "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-agent-hook.sh\"",
        "timeout": 10
      }]
    }]
  }
}
```

## Path Syntax Rules

Hooks support `$CLAUDE_PROJECT_DIR` and `$HOME` environment variables:

```json
{
  "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/hook.sh\""
}
```

**Important**: Quote the entire path to handle project folders with spaces:
- Correct: `"\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/hook.sh\""`
- Wrong: `"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/hook.sh"`

For StatusLine path configuration, see @settings-management.md (StatusLine does NOT support environment variables).

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
- `.claude/hooks/moai/handle-agent-hook.sh`: TeammateIdle, TaskCompleted events (team mode)

## Rules

- Hook feedback is treated as user input
- When blocked, suggest alternatives
- Avoid infinite loops (no recursive tool calls)
- Keep hooks lightweight for performance
- Use proper path quoting to handle spaces in project paths
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
