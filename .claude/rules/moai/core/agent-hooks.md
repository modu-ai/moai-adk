---
paths: "**/.claude/agents/**,**/.claude/hooks/**"
---

# Agent Hooks

Agent-specific hooks defined in agent frontmatter for lifecycle event handling. These hooks use the `handle-agent-hook.sh` wrapper script.

For general hook system reference, see @hooks-system.md.

## Configuration

Hooks are defined in agent YAML frontmatter using these event types:

- **SubagentStart**: Matched by the agent-type name; triggers when the subagent begins
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
| manager-develop | cycle-pre-transformation | cycle-post-transformation | cycle-completion |
| manager-develop | develop-pre-implementation | develop-post-implementation | develop-completion |
| manager-spec | - | - | spec-completion |
| manager-docs | - | docs-verification | docs-completion |

Note: The archived `expert-backend` / `expert-frontend` / `expert-devops` / `manager-quality` rows that previously appeared here were removed during the catalog consolidation. Domain expertise formerly routed through those agents is now delivered via per-spawn `Agent(general-purpose)` per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C; quality-gate enforcement moved to the global Stop hook `sync-phase-quality-gate.sh` (see `.claude/rules/moai/core/agent-common-protocol.md` § Hook Invocation Surface).


Note: Dynamic team teammates (spawned via `Agent(subagent_type: "general-purpose")`) do not use agent-scoped hooks. Quality enforcement for teammates uses global TeammateIdle and TaskCompleted hooks in settings.json.

## Hook Command Interface

Agent hooks are executed via `moai hook agent <action>`:

```bash
moai hook agent cycle-pre-transformation
moai hook agent develop-pre-implementation
```

(The `ddd-*` action family was retired; the current action is `cycle-pre-transformation`. The help text in `internal/cli/hook.go` still carries the stale example and is tracked for a follow-up code update.)

stdin JSON structure:

```json
{
  "eventType": "SubagentStop",
  "toolName": "",
  "toolInput": null,
  "toolOutput": null,
  "session": { "id": "sess-123", "cwd": "/path/to/project", "projectDir": "/path/to/project" },
  "data": { "agent": "manager-develop", "action": "cycle-completion" }
}
```

## Handler Architecture

The `moai hook agent <action>` subcommand infers the hook `EventType` from the action suffix (`-validation` / `-pre-transformation` / `-pre-implementation` → `PreToolUse`; `-verification` / `-post-transformation` / `-post-implementation` → `PostToolUse`; `-completion` → `SubagentStop`; unknown suffix → `PreToolUse`) and dispatches through the generic `EventType`-keyed hook registry (`internal/hook/registry.go`). Handlers are registered by event type, not by per-agent identity — there is no per-agent handler factory.
