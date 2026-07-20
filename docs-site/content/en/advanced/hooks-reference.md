---
title: Hooks Event Reference
weight: 60
draft: false
---

Claude Code's hook system supports **29 event types**, **5 hook types**, **per-event matchers**, and **smart behaviors**. Hooks are the only deterministic control points in the agentic harness where execution is guaranteed — a prompt can be ignored, but a hook cannot.

> For hook fundamentals and configuration, see the [Hooks Guide](/en/advanced/hooks-guide). This page is the full event reference.

## Hook Types

Five hook types are available.

| Type | Description | Example |
|------|------|------|
| **command** | Executes a shell script | `".claude/hooks/moai/handle-session-start.sh"` |
| **prompt** | LLM evaluation | An LLM runs the prompt text and returns a result |
| **agent** | Sub-agent verification | An agent verifies the work and returns a result |
| **http** | Webhook endpoint | Delivers the event via an HTTP POST request |
| **mcp_tool** | MCP tool execution | Remotely invokes an MCP server tool |

## Full Event Reference (29)

### Lifecycle Events

| Event | Description | Matcher |
|--------|------|------|
| `SessionStart` | Session start | — |
| `SessionEnd` | Session end | — |
| `PostSession` | Runs after session end (a self-hosted runner lifecycle event, CC 2.1.169+). Fires after the session is fully released, later than `SessionEnd`. MoAI-ADK does not currently wire this hook. Documented as an available option for self-hosted deployments needing post-session cleanup/telemetry. | — |
| `Stop` | Agent stop | — |
| `SubagentStop` | Sub-agent stop | — |
| `SubagentStart` | Sub-agent start | — |
| `StopFailure` | Stop failure | `errorType` |
| `Setup` | Initial setup | — |

### Tool Events

| Event | Description | Matcher |
|--------|------|------|
| `PreToolUse` | Before tool execution | `toolName` |
| `PostToolUse` | After tool execution | `toolName` |
| `PostToolUseFailure` | Tool execution failure | `toolName`, `errorType` |
| `PostToolBatch` | After a parallel tool batch executes (v2.1.89+) | — |

### Context Events

| Event | Description | Matcher |
|--------|------|------|
| `PreCompact` | Before context compaction | — |
| `PostCompact` | After context compaction | — |
| `InstructionsLoaded` | Instructions finished loading | — |

### Input Events

| Event | Description | Matcher |
|--------|------|------|
| `UserPromptSubmit` | User prompt submitted | — |
| `UserPromptExpansion` | Slash-command prompt expansion (v2.1.90+) | — |
| `Elicitation` | Elicitation started | — |
| `ElicitationResult` | Elicitation completed | — |

### Security Events

| Event | Description | Matcher |
|--------|------|------|
| `PermissionRequest` | Permission request | `toolName` |
| `PermissionDenied` | Permission denied | `toolName` |

### Team Events

| Event | Description | Matcher |
|--------|------|------|
| `TeammateIdle` | Teammate transitions to idle | — |
| `TaskCompleted` | Task marked complete | — |
| `TaskCreated` | Task created | — |

### Worktree Events

| Event | Description | Matcher |
|--------|------|------|
| `WorktreeCreate` | Worktree creation | — |
| `WorktreeRemove` | Worktree removal | — |

### Environment Events

| Event | Description | Matcher |
|--------|------|------|
| `ConfigChange` | Configuration change | `configSource` |
| `CwdChanged` | Working directory change | — |
| `FileChanged` | File change | — |

### UI Events

| Event | Description | Matcher |
|--------|------|------|
| `Notification` | User notification | — |

## Smart Behaviors

MoAI-ADK hooks go beyond simple event handling to perform intelligent behaviors.

### PermissionDenied Auto-Retry

When permission for a read-only tool (Read, Grep, Glob) is denied, the hook automatically triggers a retry. This mitigates the problem of permission prompts not being displayed for background agents.

### StopFailure Error-Type Response

On agent stop failure, differentiated responses are provided by error type. This ensures stability in long-running sessions.

### PostCompact Session-Note Restoration

After context compaction, important session notes (progress state, SPEC references) are automatically restored. Context compaction is a trade that saves tokens at the cost of information — this hook protects the essential information from that loss.

### SubagentStart Context Injection

When a sub-agent starts, the necessary context (project rules, MX tags, progress state) is automatically injected.

## Matchers

Matchers let you filter so a hook runs only under specific conditions. Attaching hooks to every event increases execution cost accordingly, so narrowing scope with matchers is the default.

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

| Matcher field | Applicable events | Description |
|----------|-----------|------|
| `toolName` | PreToolUse, PostToolUse, PostToolUseFailure, PermissionRequest, PermissionDenied | Filter by tool name |
| `errorType` | StopFailure, PostToolUseFailure | Filter by error type |
| `configSource` | ConfigChange | Filter by configuration source |

## CLAUDE_ENV_FILE

Environment variables can be managed persistently via the `CwdChanged` and `FileChanged` hooks.

```bash
# .claude/hooks/moai/handle-cwd-changed.sh
# Persist environment variables via CLAUDE_ENV_FILE
echo "MOAI_PROJECT_DIR=$(pwd)" >> "$CLAUDE_ENV_FILE"
```

This lets you keep environment variables across sessions and automatically reconfigure the environment on directory changes.

## Key Hooks Used by MoAI-ADK

| Event | MoAI handler | Role |
|--------|-----------|------|
| `SessionStart` | `handle-session-start.sh` | Statusline initialization, metrics session start |
| `PostToolUse` | `handle-post-tool.sh` | Task metrics logging |
| `TeammateIdle` | `handle-teammate-idle.sh` | LSP quality-gate verification |
| `TaskCompleted` | `handle-task-completed.sh` | SPEC document existence check |
| `WorktreeCreate` | (none — not registered by MoAI by default) | Uses Claude Code's default worktree behavior (for `isolation: worktree` agents). If registered, the active-creator contract (directory creation + path stdout echo) is mandatory. |
| `WorktreeRemove` | (none — not registered by MoAI by default) | Uses Claude Code's default worktree cleanup behavior. If registered, the observer-only contract (no output required) applies. |
| `UserPromptSubmit` | `handle-user-prompt-submit.sh` | Prompt preprocessing (forwards user input) |
| `Stop` | `handle-stop-goal.sh` | goal engine — evaluates the `/goal`/`/moai goal` autonomous-continuation condition |
| `Stop` | `sync-phase-quality-gate.sh` | sync-phase quality gate (lint + test + coverage delta) |
| `PostToolUse` | `status-transition-ownership.sh` | SPEC frontmatter status-transition audit logging (advisory) |
| `TaskCompleted` | `team-ac-verify.sh` | team-mode per-AC PASS evidence-file verification (dormant by default) |
| `Stop` / `SubagentStop` / `UserPromptSubmit` | `handle-harness-observe-*.sh` | self-evolving harness observation (Loop 0) |

## Next Steps

- [Hooks Guide](/en/advanced/hooks-guide) — hook fundamentals and configuration
- [settings.json Guide](/en/advanced/settings-json) — the full settings.json reference
- [CLI Reference](/en/getting-started/cli) — details on the `moai hook` command
