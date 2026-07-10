---
paths: "**/.claude/hooks/**,**/.claude/settings.json,**/.claude/settings.local.json"
---

# Hooks System

Claude Code hooks for extending functionality with custom scripts.

## Hook Events

30 events documented below: 20 registered in settings.json + 4 RETIRE-OBS-ONLY (Go-only observability taps, opt-in via system.yaml) + 6 upstream events MoAI does not register by default (PostSession, PostToolBatch, UserPromptExpansion, WorktreeCreate, WorktreeRemove, MessageDisplay); plus 1 retired event (Setup).
**Note**: The moai-adk Go `EventSetup` constant is retired (orphan, no handler implementation); the upstream Claude Code `Setup` event remains a current, usable event.
Active settings.json keys: 20. RETIRE-OBS-ONLY (Go-only): 4.

**Event reference (20 registered in settings.json + 4 RETIRE-OBS-ONLY in Go; the table also documents upstream events MoAI does not register by default — PostSession, PostToolBatch, UserPromptExpansion, WorktreeCreate, WorktreeRemove):**

| Event | Matcher | Can Block | Description |
|-------|---------|-----------|-------------|
| SessionStart | Source | No | Runs when a new session begins. Matchers: startup, resume, clear, compact |
| SessionEnd | Reason | No | Runs when session terminates. Matchers: clear, resume, logout, prompt_input_exit, bypass_permissions_disabled, other |
| PostSession | No | No | Runs after a session ends (self-hosted runner lifecycle event, CC 2.1.169+). Fires once the session is fully torn down, later than SessionEnd. MoAI-ADK does not wire this hook today; documented as an available option for self-hosted deployments that need post-session cleanup/telemetry. |
| PreToolUse | Tool name | Yes | Runs before a tool executes |
| PostToolUse | Tool name | No | Runs after a tool completes successfully |
| PostToolUseFailure | Tool name | No | Runs after a tool execution fails |
| PostToolBatch | No | No | Runs after a batch of parallel tool calls resolves (v2.1.89+) |
| UserPromptExpansion | Slash command name | Yes | Runs when slash command expands into prompt (v2.1.90+) |
| PreCompact | Trigger | Yes | Runs before context compaction. Matchers: manual, auto |
| PostCompact | Trigger | No | Runs after context compaction completes (v2.1.76+). Matchers: manual, auto |
| Stop | No | Yes | Runs when Claude finishes responding |
| StopFailure | Error type | No | Runs when a turn ends due to API error (v2.1.78+). Matchers: rate_limit, overloaded, authentication_failed, oauth_org_not_allowed, billing_error, invalid_request, model_not_found, server_error, max_output_tokens, unknown |
| SubagentStart | Agent type | No | Runs when a subagent spawns |
| SubagentStop | Agent type | Yes | Runs when a subagent terminates |
| Notification | Type | No | Runs when notifications sent. Matchers: permission_prompt, idle_prompt, auth_success, elicitation_dialog, elicitation_complete, elicitation_response, agent_needs_input, agent_completed. **Go-only observability tap (see sub-table below).** |
| UserPromptSubmit | No | Yes | Runs when user submits a prompt, before processing |
| PermissionRequest | Tool name | Yes | Runs when permission dialog appears |
| PermissionDenied | Tool name | No | Runs after auto mode denies a tool call. Return {retry: true} to retry (v2.1.89+) |
| TeammateIdle | No | Yes | Runs when agent team teammate is about to go idle |
| TaskCompleted | No | Yes | Runs when a task is being marked complete. **Go-only observability tap (see sub-table below).** |
| TaskCreated | No | Yes | Runs when a task is created via TaskCreate (v2.1.84+). **Go-only observability tap (see sub-table below).** |
| WorktreeCreate | No | Yes | Runs when a worktree is created for agent isolation (v2.1.49+). **Active creator contract**: hook MUST create the directory and echo its absolute path to stdout (plain text); empty stdout or non-zero exit aborts creation. **Not registered by MoAI default** — see `.claude/rules/moai/workflow/worktree-integration.md` §WorktreeCreate and WorktreeRemove Hooks. |
| WorktreeRemove | No | No | Runs when a worktree is removed after agent terminates (v2.1.49+). Observer role; no output required. **Not registered by MoAI default** — see worktree-integration.md. |
| ConfigChange | Config source | Yes | Runs when config files change (v2.1.49+). Matchers: user_settings, project_settings, local_settings, policy_settings, skills |
| CwdChanged | No | No | Runs when working directory changes (v2.1.83+). Receives CLAUDE_ENV_FILE |
| FileChanged | Filename | No | Runs when a file is changed externally (v2.1.83+). The matcher takes **literal filenames** (NOT regex/glob) — the value is split on `|` and each segment is registered as a literal filename in the working directory. Receives CLAUDE_ENV_FILE |
| InstructionsLoaded | Load reason | No | Runs when CLAUDE.md or rules loaded (v2.1.69+). Matchers: session_start, nested_traversal, path_glob_match, include, compact |
| Elicitation | MCP server | Yes | Runs when MCP server requests user input (v2.1.76+). Handler types: command+http+mcp_tool only (prompt/agent not supported per handler-type matrix). **Go-only observability tap (see sub-table below).** |
| ElicitationResult | MCP server | Yes | Runs after user responds to MCP elicitation (v2.1.76+). Handler types: command+http+mcp_tool only (prompt/agent not supported per handler-type matrix). **Go-only observability tap (see sub-table below).** |

**RETIRE-OBS-ONLY events (Go-only, not in settings.json — enable via system.yaml hook.observability_events):**

| Event | Notes |
|-------|-------|
| Notification | Observability tap; silent unless opted in |
| Elicitation | Observability tap; silent unless opted in |
| ElicitationResult | Observability tap; silent unless opted in |
| TaskCreated | Observability tap; silent unless opted in |

**Retired event:**

| Event | Status |
|-------|--------|
| Setup | Upstream CC event is CURRENT (triggered via --init, --init-only, or --maintenance flags, v2.1.10+). Only moai-adk internals retired: the Go `EventSetup` constant (orphan, no handler implementation) and the `moai hook setup` subcommand. The upstream Claude Code `Setup` event still exists and remains usable. |

### Event Categories

**Lifecycle Events**: SessionStart, Setup, SessionEnd, ConfigChange, InstructionsLoaded

**Context Events**: PreCompact, PostCompact, FileChanged, CwdChanged, WorktreeCreate, WorktreeRemove

**Prompt and Notification Events**: UserPromptSubmit, UserPromptExpansion, PermissionRequest, PermissionDenied, Notification, Elicitation, ElicitationResult

**Tool Events**: PreToolUse, PostToolUse, PostToolUseFailure, PostToolBatch

**Agent and Task Events**: SubagentStart, SubagentStop, TeammateIdle, TaskCompleted, TaskCreated

**Conversation State Events**: Stop, StopFailure

### Upstream Events Not Yet Adopted by MoAI

The following Claude Code hook event exists upstream but MoAI does not register a handler for it by default. Documented here for reference:

| Event | Matcher | Can Block | Description |
|-------|---------|-----------|-------------|
| MessageDisplay | No | No | Runs while assistant message text is displayed (v2.1.152+). Returns `hookSpecificOutput.displayContent` to replace the on-screen text; display-only — the transcript and what Claude sees keep the original. No MoAI handler registered by default. |

## Hook Event stdin/stdout Reference

| Event | stdin | stdout | Notes |
|-------|-------|--------|-------|
| UserPromptSubmit | `prompt` | `additionalContext`, `reason`, `decision:{block,reason}`, `sessionTitle`, `suppressOriginalPrompt` | Exit 2 blocks prompt; JSON `decision:"block"` also available |
| PermissionRequest | `toolName`, `toolInput` | `reason`, `decision.behavior`, `updatedInput`, `updatedPermissions` | Exit 0 = allow, exit 2 = deny; JSON `decision.behavior` (allow/deny/ask) also available |
| PermissionDenied | `toolName`, `toolInput` | `{retry: true}` | Return retry to allow model to retry (v2.1.89+) |
| PostToolUseFailure | `toolName`, `toolInput`, `error`, `is_interrupt` | `systemMessage` | Non-blocking |
| Notification | `type`, `message` | - | Types: permission_prompt, idle_prompt, auth_success, elicitation_dialog, elicitation_complete, elicitation_response, agent_needs_input, agent_completed |
| Setup | `trigger` | `systemMessage` | trigger: init, init-only, or maintenance (v2.1.10+) |
| InstructionsLoaded | `files`, `source` | - | Lists loaded instruction files (v2.1.69+) |
| SubagentStart | `agentType`, `agentName`, `agent_id` | `additionalContext` | Inject context into subagent. `agent_id` added in v2.1.69 |
| TeammateIdle | `agentType`, `agentName`, `tasksSummary`, `agent_id` | `systemMessage` or JSON | Exit 2 = keep working. Also accepts JSON: `{"continue": false, "stopReason": "..."}` to stop teammate (v2.1.69+) |
| TaskCompleted | `taskId`, `taskSummary`, `agentName`, `agent_id` | `reason` or JSON | Exit 2 = reject completion. Also accepts JSON: `{"continue": false, "stopReason": "..."}` to reject (v2.1.69+) |
| SessionStart | `source` | `hookSpecificOutput`: `additionalContext`, `reloadSkills`, `sessionTitle` | `reloadSkills` (bool): when `true`, re-scans skill/command directories after SessionStart hooks complete, so skills the hook installed are available in the same session. `sessionTitle`: sets the session title (same effect as `/rename`); applies on `startup`/`resume` only, ignored on `clear`/`compact` (v2.1.152+) |
| SessionEnd | `reason`, `sessionId` | - | Reasons: clear, resume, logout, prompt_input_exit, bypass_permissions_disabled, other |
| Stop | `last_assistant_message` | `systemMessage` | Includes last assistant message (v2.1.49+) |
| SubagentStop | `agentType`, `agentName`, `last_assistant_message`, `agent_id`, `agent_transcript_path` | `decision:{block,reason}`, `additionalContext`, `systemMessage` | `agent_id` and `agent_transcript_path` added in v2.1.42/v2.1.69. Also accepts `hookSpecificOutput.additionalContext` for non-error feedback that continues the conversation |
| ConfigChange | `configPath`, `changes` | - | Triggered on settings.json modification (v2.1.49+). The MoAI runtime handler is continue-only — reload failures surface via slog observability logs, NOT via stdout JSON or exit 2 (the handler unconditionally returns empty output) |
| StopFailure | `error_type`, `error_message` | `systemMessage` | Error types: rate_limit, overloaded, authentication_failed, oauth_org_not_allowed, billing_error, invalid_request, model_not_found, server_error, max_output_tokens, unknown (v2.1.78+) |
| CwdChanged | `old_cwd`, `new_cwd` | - | Receives CLAUDE_ENV_FILE env var for environment persistence |
| FileChanged | `file_path`, `change_type` | - | change_type: modified, created, deleted. Receives CLAUDE_ENV_FILE |
| Elicitation | `mcp_server_name`, `mcp_tool_name`, `elicitation_request` | `action`, `content` | action: accept, decline, cancel |
| ElicitationResult | `mcp_server_name`, `mcp_tool_name` | `action`, `content` | Overrides user response |
| PostToolBatch | `batch_id`, `tool_results` array | `decision`, `reason` | Exit 2 blocks batch execution. `decision: "block"` prevents execution |
| UserPromptExpansion | `expansion_type`, `command_name`, `command_args` | `decision`, `additionalContext` | Exit 2 blocks expansion. `decision: "block"` prevents command expansion |

All hook events include `agent_id` and `agent_type` fields when triggered from a subagent context (v2.1.69+).

Standard events (SessionStart, PreCompact, PreToolUse, PostToolUse) use common stdin/stdout patterns: stdin receives event-specific fields, stdout accepts optional `systemMessage`.

## Hook Execution Types

Five hook types are available:

### Command Hooks (type: "command")

Default hook type. Executes a shell command, communicates via stdin/stdout JSON.

- Configuration: `type`, `command`, `timeout`
- stdin: JSON with event data
- stdout: JSON with response (optional `systemMessage`, `additionalContext`, `reason`)
- Exit codes: 0 = success, 1 = error (shown to user), 2 = block/reject (honored only by events marked "Can Block: Yes" in the event reference above)
- PreToolUse permission decisions: `allow`, `deny`, `ask`, `defer` (defer pauses headless sessions for --resume, v2.1.89+)
- Hook stdout over 50K characters is saved to disk; only a file path + preview is injected into context (v2.1.89+)
- Exec form (shell-bypass): supply `"args": []` alongside `"command"` to run the program directly without a shell, avoiding shell-quoting and word-splitting issues. When a hook script DOES depend on a shell and must not run under a non-interactive invocation, guard the shell-only branch with an interactive-shell check — `if [[ $- == *i* ]]; then ... fi` — so the body is skipped when the script is sourced non-interactively by the hook runner.

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

### HTTP Hooks (type: "http")

Send hook input as JSON POST to a URL and receive JSON response. Useful for remote CI/CD integration and webhook-based workflows.

- Configuration: `type`, `url`, `timeout`
- The `url` field specifies the endpoint to POST event data to
- Request: JSON body containing hook event data (same as stdin for command hooks)
- Response: JSON with optional `systemMessage`, `additionalContext`, `reason`
- Same blocking behavior as command hooks (HTTP status codes map to exit codes)
- Available since v2.1.63

### MCP Tool Hooks (type: "mcp_tool")

Call a tool on a connected MCP server to make validation decisions.

- Configuration: `type`, `server`, `tool`, `timeout`
- The `server` field specifies the MCP server name (e.g., "security_server")
- The `tool` field specifies the tool name on that server
- Request: Hook event data passed as MCP tool arguments
- Response: JSON with optional `decision`, `reason`, `additionalContext`
- Same blocking behavior as command hooks
- Available since v2.1.85+

### Async Command Hooks (async: true)

Run command hooks in the background without blocking the conversation.

- Only available for `type: "command"` hooks
- Configuration: Add `async: true` to any command hook definition
- Results are delivered on the next conversation turn via `additionalContext` only (async hooks cannot control `decision` or `updatedToolOutput` — the sole async-deliverable stdout field is `additionalContext`)
- Useful for long-running validations (linting, test execution, deployments)
- Async PostToolUse can only deliver `additionalContext` — it cannot control `decision` or `updatedToolOutput` (those require synchronous PostToolUse). The shipped PostToolUse harness-observe tap is async, so it observes only; it never blocks.

### Single-Fire Hooks (once: true)

Execute a hook only once per session, then automatically skip subsequent triggers.

- Configuration: Add `once: true` to any hook definition
- Useful for one-time session initialization, first-write validation, or setup tasks
- Available since v2.1.0

### Conditional Hook Execution (if field)

Filter when hooks run using permission rule syntax (v2.1.85+).

The `if` field accepts permission rule patterns to prevent unnecessary hook execution and reduce process spawning overhead. Use tool patterns like `Bash(git *)` for git commands, `Write|Edit` for write operations, or `Bash(npm *)` for npm commands.

Example configurations:
- `"if": "Bash(git *)"` - Only run for git bash commands
- `"if": "Write|Edit"` - Only run for write/edit operations
- `"if": "Bash(npm *)"` - Only run for npm commands
- `"if": "Bash(pytest *)"` - Only run for pytest commands

This field significantly reduces performance overhead by skipping hook evaluation for non-matching operations.

### Stop Hook Block Cap

A Stop hook that keeps blocking (exit 2) would otherwise loop indefinitely. The runtime applies a block cap: after 8 consecutive Stop-hook blocks the cap is reached and the block is overridden so the turn can end. The cap is tunable via the `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` environment variable.

## Agent-Specific Hooks

Agent hooks are defined in agent frontmatter and executed for agent lifecycle events. For detailed configuration, actions table, and handler architecture, see @agent-hooks.md.

## Hook Location

Hooks are defined in `.claude/hooks/` directory:

- Shell scripts: `*.sh`
- Python scripts: `*.py`

## Configuration

Define hooks in `.claude/settings.json`. Each event key maps to an array of matcher groups; each group carries an inner `"hooks"` array of hook definitions (the inner array is REQUIRED — a hook definition placed directly in the outer array is not valid schema):

```json
{
  "hooks": {
    "SessionStart": [{
      "matcher": "startup|resume|clear|compact",
      "hooks": [{
        "type": "command",
        "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-session-start.sh\"",
        "timeout": 30
      }]
    }],
    "PreCompact": [{
      "matcher": "manual|auto",
      "hooks": [{
        "type": "command",
        "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-compact.sh\"",
        "timeout": 5
      }]
    }],
    "PreToolUse": [{
      "matcher": "Write|Edit|Bash",
      "hooks": [{
        "type": "command",
        "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-pre-tool.sh\"",
        "timeout": 5
      }]
    }],
    "PostToolUse": [{
      "matcher": "Write|Edit",
      "hooks": [{
        "type": "command",
        "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-post-tool.sh\"",
        "timeout": 10,
        "async": true
      }]
    }],
    "Stop": [{
      "hooks": [{
        "type": "command",
        "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-stop.sh\"",
        "timeout": 5
      }, {
        "type": "command",
        "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/sync-phase-quality-gate.sh\"",
        "timeout": 60
      }]
    }],
    "TeammateIdle": [{
      "hooks": [{
        "type": "command",
        "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-teammate-idle.sh\"",
        "timeout": 5
      }]
    }],
    "TaskCompleted": [{
      "hooks": [{
        "type": "command",
        "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-task-completed.sh\"",
        "timeout": 5
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

For StatusLine path configuration, see @settings-management.md (StatusLine supports the built-in `$CLAUDE_PROJECT_DIR` token, same as hooks).

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
- `.claude/hooks/moai/handle-agent-hook.sh`: agent frontmatter lifecycle hooks (PreToolUse/PostToolUse/SubagentStop — see agent-hooks.md)
- `.claude/hooks/moai/handle-teammate-idle.sh`: TeammateIdle event (team mode)
- `.claude/hooks/moai/handle-task-completed.sh`: TaskCompleted event (team mode)

## Smart Hook Behaviors (v2.10.1)

MoAI-ADK implements intelligent handler logic beyond simple logging:

- **PermissionDenied auto-retry**: Read-only tools (Read, Grep, Glob, WebFetch, WebSearch, Skill) automatically return `{retry: true}` when denied by auto mode
- **StopFailure error-type responses**: Returns targeted `systemMessage` based on `error_type` (rate_limit, authentication_failed, billing_error, max_output_tokens)
- **PostCompact memo restoration**: Reads session-memo.md saved by PreCompact and injects it as `systemMessage` for context recovery
- **SubagentStart context injection**: Injects project metadata (name, type, language, active SPEC) via `additionalContext` into spawned subagents
- **CwdChanged environment persistence**: Writes project-specific env vars to `CLAUDE_ENV_FILE` when directory changes to a MoAI project
- **UserPromptSubmit session title**: Sets Claude Code session title via `sessionTitle` field with SPEC ID or project/branch info

## Timeout Configuration

All hook types support a `timeout` field (in seconds). The maximum timeout is **600 seconds (10 minutes)** across all hook types.

MoAI-ADK uses shorter independent timeout policies for operational efficiency. The values below are the timeouts actually registered in the shipped settings.json:

| Hook registration | Shipped timeout | Max Value | Notes |
|-------------------|-----------------|-----------|-------|
| SessionStart | 30s | 600s | Session bootstrap (context injection, version checks) needs more headroom than the 5s fast-hook default |
| SessionEnd | 10s | 600s | Session teardown |
| PostToolUse (handle-post-tool) | 10s + `async: true` | 600s | Exception: background LSP/AST/MX validation runs async (10s is a per-run background ceiling, not a blocking wait) |
| PostToolUse (status-transition-ownership.sh) | 5s | 600s | Advisory governance gate |
| Stop (handle-stop) | 5s | 600s | Turn-end wrapper |
| Stop (sync-phase-quality-gate.sh) | 60s | 600s | Runs fast structural checks (`go vet` + `go build` etc.) — needs headroom beyond the 5s wrapper default |
| UserPromptSubmit | 5s | 30s | Blocks user interaction; reduced max |
| All other registered events (PreCompact, PreToolUse, PostToolUseFailure, SubagentStart, SubagentStop, TeammateIdle, TaskCompleted, ConfigChange, StopFailure, PostCompact, InstructionsLoaded, CwdChanged, FileChanged, PermissionDenied, PermissionRequest) and the opt-in harness-observe entries | 5s | 600s | Synchronous fast lifecycle hooks (blocking default; PostToolUse harness-observe is `async: true`) |
| prompt, agent hooks | 30s-60s | 600s | Evaluation/verification hooks |

The **5s default applies to synchronous blocking hooks** (PreCompact, PreToolUse, etc.); **SessionStart is 30s** (session bootstrap) and **SessionEnd is 10s**. **PostToolUse (handle-post-tool) is the documented exception at 10s + `async: true`** because its LSP/AST/MX validations run in the background — this matches the JSON example below (`PostToolUse` block with `timeout: 10, async: true`). **The Stop-event sync-phase-quality-gate entry is 60s** so the gate's compile/vet checks are not killed mid-run. These MoAI values (5s, 10s, 30s, 60s) are valid independent policies and do NOT violate the 600s upper bound. Customize the `timeout` field in hook definitions to adjust per-hook timing as needed.

## Rules

- Hook feedback is treated as user input
- When blocked, suggest alternatives
- Avoid infinite loops (no recursive tool calls)
- Keep hooks lightweight for performance
- Use proper path quoting to handle spaces in project paths
- Prompt and agent hooks return JSON with `ok` and `reason` fields
- Async hooks deliver results via `additionalContext` on the next turn (the only async-deliverable field; `systemMessage`, `decision`, and `updatedToolOutput` are NOT delivered for async hooks)
- Exit code 2 blocks on events marked "Can Block: Yes" (PreToolUse, PermissionRequest, UserPromptSubmit, UserPromptExpansion, Stop, SubagentStop, TeammateIdle, TaskCreated, TaskCompleted, ConfigChange, PostToolBatch, PreCompact, Elicitation, ElicitationResult, WorktreeCreate); events marked "No" ignore it (StopFailure, PostToolUse, PostToolUseFailure, PermissionDenied, Notification, SubagentStart, SessionStart, Setup, SessionEnd, CwdChanged, FileChanged, PostCompact, WorktreeRemove, InstructionsLoaded, MessageDisplay). Some events also support JSON `decision:"block"` (PostToolUse, PostToolBatch, SubagentStop, ConfigChange, PreCompact, UserPromptSubmit, UserPromptExpansion) or `continue:false` (TeammateIdle, TaskCreated, TaskCompleted) as alternative block mechanisms — exit 2 is NOT universal
- Stop and SubagentStop hooks receive `last_assistant_message` field (v2.1.49+)

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

## MX Tag Integration with Hooks

PostToolUse hooks can trigger MX tag validation after code modifications:

**Trigger Conditions:**
- Write or Edit tool used on source files (`.go`, `.py`, `.ts`, etc.)
- New functions or classes added
- Function signatures changed

**PostToolUse MX Check Flow:**
1. Detect if modified file is a source code file
2. Check if file has `.moai/config/sections/mx.yaml` exclusion
3. If new exported function added without @MX tag, log warning
4. If function with @MX:ANCHOR modified, flag for review

**Hook Wrapper Enhancement:**
```bash
# handle-post-tool.sh MX check
if [[ "$TOOL_NAME" =~ ^(Write|Edit)$ ]] && is_source_file "$FILE_PATH"; then
  # Check for MX tag needs
  moai mx check --file "$FILE_PATH" --dry
fi
```

**Non-Blocking Behavior:**
- MX checks are informational only during hook execution
- Actual tag insertion happens during workflow phases (run, sync)
- Use `/moai mx --dry` to preview tag recommendations
