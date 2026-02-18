---
paths:
  - "**/.claude/agents/**"
  - "**/.claude/hooks/**"
---

# Agent Hooks

Agent-specific hooks defined in agent frontmatter for lifecycle event handling. These hooks use the `handle-agent-hook.sh` wrapper script.

For general hook system reference, see @hooks-system.md.

## Configuration

Hooks are defined in agent YAML frontmatter using three event types:

- **PreToolUse**: Matcher `Write|Edit|MultiEdit` for pre-change validation
- **PostToolUse**: Matcher `Write|Edit|MultiEdit` for post-change verification
- **SubagentStop**: No matcher, triggers on agent completion

Configuration pattern per agent:

```yaml
hooks:
  PreToolUse:
    - matcher: "Write|Edit|MultiEdit"
      hooks:
        - type: command
          command: "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-agent-hook.sh\" {action}"
          timeout: 5
  PostToolUse:
    - matcher: "Write|Edit|MultiEdit"
      hooks:
        - type: command
          command: "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-agent-hook.sh\" {action}"
          timeout: 10
  SubagentStop:
    hooks:
      - type: command
        command: "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-agent-hook.sh\" {action}"
        timeout: 10
```

## Agent Hook Actions

Actions follow the naming pattern `{agent}-{phase}`:

| Agent | PreToolUse | PostToolUse | SubagentStop |
|-------|-----------|------------|-------------|
| manager-ddd | ddd-pre-transformation | ddd-post-transformation | ddd-completion |
| manager-tdd | tdd-pre-implementation | tdd-post-implementation | tdd-completion |
| expert-backend | backend-validation | backend-verification | - |
| expert-frontend | frontend-validation | frontend-verification | - |
| expert-testing | - | testing-verification | testing-completion |
| expert-debug | - | debug-verification | debug-completion |
| expert-devops | - | devops-verification | devops-completion |
| manager-quality | - | - | quality-completion |
| manager-spec | - | - | spec-completion |
| manager-docs | - | docs-verification | docs-completion |
| team-backend-dev | - | team-backend-verification | - |
| team-frontend-dev | - | team-frontend-verification | - |
| team-tester | - | team-testing-verification | team-testing-completion |
| team-quality | - | - | team-quality-completion |

## Hook Command Interface

Agent hooks are executed via `moai hook agent <action>`:

```bash
moai hook agent ddd-pre-transformation
moai hook agent backend-validation
```

stdin JSON structure:

```json
{
  "eventType": "SubagentStop",
  "toolName": "",
  "toolInput": null,
  "toolOutput": null,
  "session": { "id": "sess-123", "cwd": "/path/to/project", "projectDir": "/path/to/project" },
  "data": { "agent": "manager-ddd", "action": "ddd-completion" }
}
```

## Handler Architecture

The `internal/hook/agents/factory.go` implements handler creation per agent. Each agent type has a dedicated handler file: `{agent}_handler.go` (e.g., `ddd_handler.go`, `backend_handler.go`). Unknown actions fall through to `default_handler.go`.
