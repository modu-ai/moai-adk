---
title: Hooks Event Reference
weight: 60
draft: false
---

As of MoAI-ADK v2.10.1, Claude Code's hook system supports **29 event types**, **5 hook types**, **per-event matchers**, and **smart behaviors**.

> For foundational concepts and setup instructions, see [Hooks Guide](/en/advanced/hooks-guide). This page is the complete event reference.

## Hook Types

**Five hook types are available:**

| Type | Description | Example |
|------|-------------|---------|
| **command** | Shell script execution | `".claude/hooks/moai/handle-session-start.sh"` |
| **prompt** | LLM evaluation | LLM executes prompt text and returns result |
| **agent** | Sub-agent validation | Agent validates task and returns result |
| **http** | Webhook endpoint | HTTP POST request to remote endpoint |
| **mcp_tool** | MCP tool invocation | Remote call to MCP server tool |

## Complete Event Reference (29 events)

### Lifecycle Events

| Event | Description | Matcher |
|-------|-------------|---------|
| `SessionStart` | Session started | — |
| `SessionEnd` | Session ended | — |
| `Stop` | Agent stopped | — |
| `SubagentStop` | Sub-agent stopped | — |
| `SubagentStart` | Sub-agent started | — |
| `StopFailure` | Stop failed | `errorType` |
| `Setup` | Initial setup | — |

### Tool Events

| Event | Description | Matcher |
|-------|-------------|---------|
| `PreToolUse` | Before tool execution | `toolName` |
| `PostToolUse` | After tool execution | `toolName` |
| `PostToolUseFailure` | Tool execution failed | `toolName`, `errorType` |
| `PostToolBatch` | After batch of parallel tool calls (v2.1.89+) | — |

### Context Events

| Event | Description | Matcher |
|-------|-------------|---------|
| `PreCompact` | Before context compaction | — |
| `PostCompact` | After context compaction | — |
| `InstructionsLoaded` | Instructions loaded | — |

### Input Events

| Event | Description | Matcher |
|-------|-------------|---------|
| `UserPromptSubmit` | User prompt submitted | — |
| `UserPromptExpansion` | Slash command expanded to prompt (v2.1.90+) | — |
| `Elicitation` | Elicitation started | — |
| `ElicitationResult` | Elicitation completed | — |

### Security Events

| Event | Description | Matcher |
|-------|-------------|---------|
| `PermissionRequest` | Permission requested | `toolName` |
| `PermissionDenied` | Permission denied | `toolName` |

### Team Events

| Event | Description | Matcher |
|-------|-------------|---------|
| `TeammateIdle` | Teammate idle state | — |
| `TaskCompleted` | Task marked complete | — |
| `TaskCreated` | Task created | — |

### Worktree Events

| Event | Description | Matcher |
|-------|-------------|---------|
| `WorktreeCreate` | Worktree created | — |
| `WorktreeRemove` | Worktree removed | — |

### Environment Events

| Event | Description | Matcher |
|-------|-------------|---------|
| `ConfigChange` | Config changed | `configSource` |
| `CwdChanged` | Working directory changed | — |
| `FileChanged` | File changed | — |

### UI Events

| Event | Description | Matcher |
|-------|-------------|---------|
| `Notification` | User notification | — |

## Smart Behaviors

MoAI-ADK hooks perform intelligent actions beyond simple event handling:

### PermissionDenied Auto-Retry

When read-only tools (Read, Grep, Glob) are denied, hooks automatically trigger retry. This mitigates issues with permission prompts not displaying in background agents.

### StopFailure Error-Type Response

When agent stop fails, responses vary by error type. Ensures stability in long-running sessions.

### PostCompact Session Notes Recovery

Critical session notes (progress, SPEC references) are automatically restored after context compaction. This prevents loss of essential information during compression.

### SubagentStart Context Injection

Required context (project rules, MX tags, progress) is automatically injected when sub-agents start.

## Matchers

Use matchers to filter hooks to run only under specific conditions:

```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": { "toolName": "Bash" },
      "hooks": [{
        "type": "command",
        "command": "echo 'Bash tool detected'",
        "timeout": 5
      }]
    }]
  }
}
```

### Available Matcher Fields

| Matcher Field | Applied Events | Description |
|---------------|----------------|-------------|
| `toolName` | PreToolUse, PostToolUse, PostToolUseFailure, PermissionRequest, PermissionDenied | Filter by tool name |
| `errorType` | StopFailure, PostToolUseFailure | Filter by error type |
| `configSource` | ConfigChange | Filter by config source |

## CLAUDE_ENV_FILE

Use `CwdChanged` and `FileChanged` hooks to manage environment variables persistently:

```bash
# .claude/hooks/moai/handle-cwd-changed.sh
# Persist environment variables via CLAUDE_ENV_FILE
echo "MOAI_PROJECT_DIR=$(pwd)" >> "$CLAUDE_ENV_FILE"
```

This maintains environment variables across sessions and automatically reconfigures the environment when directories change.

## Key Hooks Used by MoAI-ADK

| Event | MoAI Handler | Role |
|-------|-------------|------|
| `SessionStart` | `handle-session-start.sh` | Initialize statusline, start metrics session |
| `PostToolUse` | `handle-post-tool.sh` | Log task metrics |
| `TeammateIdle` | `handle-teammate-idle.sh` | Validate LSP quality gates |
| `TaskCompleted` | `handle-task-completed.sh` | Verify SPEC document exists |
| `WorktreeCreate` | (none — MoAI default unregistered) | Uses Claude Code default worktree behavior (for `isolation: worktree` agents). Registration requires active creator contract (create directory + echo absolute path to stdout). |
| `WorktreeRemove` | (none — MoAI default unregistered) | Uses Claude Code default worktree cleanup. Registration is observer-only (no stdout required). |
| `UserPromptSubmit` | `handle-user-prompt.sh` | Auto-execute quality gates |

## Next Steps

- [Hooks Guide](/en/advanced/hooks-guide) — Foundational concepts and setup instructions
- [settings.json Guide](/en/advanced/settings-json) — Complete settings.json reference
- [CLI Reference](/en/getting-started/cli) — Detailed `moai hook` command documentation
