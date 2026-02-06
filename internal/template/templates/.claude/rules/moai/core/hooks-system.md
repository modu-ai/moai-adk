# Hooks System

Claude Code hooks for extending functionality with custom scripts.

## Hook Types

Available hook event types:

- SessionStart: Execute when a new session begins
- PreCompact: Execute before context compaction
- PreToolUse: Execute before a tool runs (supports matcher for tool-specific hooks)
- PostToolUse: Execute after a tool completes (supports matcher for tool-specific hooks)
- Stop: Execute when conversation stops
- SubagentStop: Execute when a subagent terminates (agent-specific hooks)

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
- ✅ Correct: `"\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/hook.sh\""`
- ❌ Wrong: `"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/hook.sh"`

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

## Error Handling

- Failed hooks should exit with non-zero code
- Error messages are displayed to user
- Hooks can block operations by returning error
- Missing hooks exit silently (Claude Code handles gracefully)

## Security

- Hooks run in sandbox by default
- Validate all hook inputs
- Do not store secrets in hook scripts

## MoAI Integration

- Skill("moai-foundation-claude") for detailed patterns
- Hook scripts must follow coding-standards.md
- Hook wrappers are managed by `internal/hook/` package
