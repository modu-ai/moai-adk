# Hooks

Execute custom shell commands at various points in Claude Code's lifecycle to customize and extend behavior.

## Overview

Claude Code hooks are user-defined shell commands that execute at specific points during Claude Code's lifecycle. Hooks provide deterministic control over Claude Code's behavior, ensuring specific actions always occur without relying on what the LLM chooses to execute.

Hooks enable powerful automation and customization scenarios:

- **Notifications**: Customize how you receive alerts when Claude Code waits for input or permissions
- **Auto-formatting**: Run `prettier` on .ts files, `gofmt` on .go files after every file edit
- **Logging**: Track and audit all executed commands for compliance or debugging
- **Feedback**: Provide automated feedback when Claude Code generates code that doesn't follow codebase rules
- **Custom permissions**: Block modifications to production files or sensitive directories
- **Context injection**: Add dynamic information like current time, git status, or environment details

By encoding these rules as hooks rather than prompt instructions, you transform suggestions into app-level code that executes every time as expected.

## Hook Events

Claude Code provides several hook events that execute at different points in the workflow:

### Tool-Related Events

| Event | When It Executes | Can Block | Common Use Cases |
|-------|------------------|-----------|------------------|
| **PreToolUse** | Before tool execution | Yes | Permission validation, input sanitization, logging |
| **PostToolUse** | After tool completion | No (provides feedback) | Code formatting, quality checks, notifications |

### Session Events

| Event | When It Executes | Can Block | Common Use Cases |
|-------|------------------|-----------|------------------|
| **SessionStart** | Session initialization or resume | No | Context loading, environment setup |
| **SessionEnd** | Session termination | No | Cleanup, final logging |
| **UserPromptSubmit** | Before processing user input | Yes | Prompt validation, context injection |

### Workflow Events

| Event | When It Executes | Can Block | Common Use Cases |
|-------|------------------|-----------|------------------|
| **Stop** | When main agent completes response | Yes | Continuation logic, follow-up actions |
| **SubagentStop** | When subagent completes | Yes | Subagent coordination, result validation |
| **PreCompact** | Before context compaction | No | Context preservation, custom instructions |
| **Notification** | When sending notifications | No | Custom notification handling |

## Configuration

### Settings File Structure

Hooks are configured in [settings files](settings):

- `~/.claude/settings.json` - User settings (apply across all projects)
- `.claude/settings.json` - Project settings (shared with team)
- `.claude/settings.local.json` - Local project settings (not committed)
- Enterprise managed policy settings

### Basic Configuration Structure

```json
{
  "hooks": {
    "EventName": [
      {
        "matcher": "ToolPattern",
        "hooks": [
          {
            "type": "command",
            "command": "your-command-here",
            "timeout": 60
          }
        ]
      }
    ]
  }
}
```

### Configuration Parameters

| Parameter | Description | Required | Example |
|-----------|-------------|----------|---------|
| `matcher` | Tool name pattern to match (case-sensitive) | No* | `"Bash"`, `"Edit\|Write"`, `"*"` |
| `type` | Command type (currently only `"command"` supported) | Yes | `"command"` |
| `command` | Shell command to execute | Yes | `"jq -r '.tool_input.command'"` |
| `timeout` | Execution timeout in seconds (default: 60) | No | `30` |

*Required for PreToolUse and PostToolUse; optional for other events

### Matcher Patterns

- **Exact match**: `"Write"` matches only the Write tool
- **Regex patterns**: `"Edit|MultiEdit|Write"` or `"Notebook.*"`
- **Wildcard**: `"*"` matches all tools
- **Empty**: `""` or omit matcher field for non-tool events

### Environment Variables

Hooks have access to Claude Code's environment plus:

- `CLAUDE_PROJECT_DIR`: Absolute path to project root directory
- Standard environment variables from your shell

## Hook Input

Hooks receive JSON data via stdin containing session information and event-specific data:

### Common Fields

All hooks receive these base fields:

```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/conversation.jsonl",
  "cwd": "/current/working/directory",
  "hook_event_name": "EventName"
}
```

### PreToolUse Input

```json
{
  "session_id": "abc123",
  "transcript_path": "/Users/.../conversation.jsonl",
  "cwd": "/Users/...",
  "hook_event_name": "PreToolUse",
  "tool_name": "Write",
  "tool_input": {
    "file_path": "/path/to/file.txt",
    "content": "file content"
  }
}
```

### PostToolUse Input

```json
{
  "session_id": "abc123",
  "transcript_path": "/Users/.../conversation.jsonl",
  "cwd": "/Users/...",
  "hook_event_name": "PostToolUse",
  "tool_name": "Write",
  "tool_input": {
    "file_path": "/path/to/file.txt",
    "content": "file content"
  },
  "tool_response": {
    "filePath": "/path/to/file.txt",
    "success": true
  }
}
```

### UserPromptSubmit Input

```json
{
  "session_id": "abc123",
  "transcript_path": "/Users/.../conversation.jsonl",
  "cwd": "/Users/...",
  "hook_event_name": "UserPromptSubmit",
  "prompt": "Write a function to calculate the factorial of a number"
}
```

### Notification Input

```json
{
  "session_id": "abc123",
  "transcript_path": "/Users/.../conversation.jsonl",
  "cwd": "/Users/...",
  "hook_event_name": "Notification",
  "message": "Task completed successfully"
}
```

### SessionStart Input

```json
{
  "session_id": "abc123",
  "transcript_path": "/Users/.../conversation.jsonl",
  "hook_event_name": "SessionStart",
  "source": "startup"
}
```

**Source values:**
- `"startup"` - Called on startup
- `"resume"` - Called from `--resume`, `--continue`, or `/resume`
- `"clear"` - Called from `/clear`

### Stop/SubagentStop Input

```json
{
  "session_id": "abc123",
  "transcript_path": "/Users/.../conversation.jsonl",
  "hook_event_name": "Stop",
  "stop_hook_active": true
}
```

## Hook Output

Hooks can return output to Claude Code using two methods: exit codes or structured JSON.

### Exit Code Method

Simple method using exit codes and standard streams:

| Exit Code | Meaning | Behavior |
|-----------|---------|----------|
| `0` | Success | `stdout` shown to user in transcript mode; for `UserPromptSubmit` and `SessionStart`, stdout added to context |
| `2` | Blocking error | `stderr` automatically fed back to Claude; specific behavior by event type |
| Other | Non-blocking error | `stderr` shown to user, execution continues |

#### Exit Code 2 Behavior by Event

| Hook Event | Behavior |
|------------|----------|
| `PreToolUse` | Blocks tool call, shows stderr to Claude |
| `PostToolUse` | Shows stderr to Claude (tool already executed) |
| `UserPromptSubmit` | Blocks prompt processing, clears prompt, shows stderr to user only |
| `Stop`/`SubagentStop` | Blocks stopping, shows stderr to Claude |
| `Notification` | Shows stderr to user only |
| `PreCompact`/`SessionStart` | Shows stderr to user only |

### JSON Output Method

Advanced method for sophisticated control using structured JSON output:

#### Common JSON Fields

```json
{
  "continue": true,
  "stopReason": "Optional message when continue is false",
  "suppressOutput": false,
  "systemMessage": "Optional warning for user"
}
```

#### PreToolUse Decision Control

```json
{
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "allow" | "deny" | "ask",
    "permissionDecisionReason": "Explanation shown to user or Claude"
  }
}
```

**Permission decisions:**
- `"allow"`: Bypasses permission system, reason shown to user
- `"deny"`: Prevents tool execution, reason shown to Claude
- `"ask"`: Prompts user for confirmation, reason shown to user

#### PostToolUse Decision Control

```json
{
  "decision": "block" | undefined,
  "reason": "Explanation for decision",
  "hookSpecificOutput": {
    "hookEventName": "PostToolUse",
    "additionalContext": "Additional information for Claude"
  }
}
```

#### UserPromptSubmit Decision Control

```json
{
  "decision": "block" | undefined,
  "reason": "Explanation for blocking",
  "hookSpecificOutput": {
    "hookEventName": "UserPromptSubmit",
    "additionalContext": "Context to inject if not blocked"
  }
}
```

#### SessionStart Context Injection

```json
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": "Context to load at session start"
  }
}
```

## Common Hook Examples

### Bash Command Logging

Log all bash commands before execution:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "jq -r '\"\\(.tool_input.command) - \\(.tool_input.description // \"No description\")\"' >> ~/.claude/bash-command-log.txt"
          }
        ]
      }
    ]
  }
}
```

### Automatic Code Formatting

Format TypeScript files after editing:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|MultiEdit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "jq -r '.tool_input.file_path' | { read file_path; if echo \"$file_path\" | grep -q '\\.ts$'; then npx prettier --write \"$file_path\"; fi; }"
          }
        ]
      }
    ]
  }
}
```

### Desktop Notifications

Receive desktop notifications when Claude needs input:

```json
{
  "hooks": {
    "Notification": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "notify-send 'Claude Code' 'Awaiting your input'"
          }
        ]
      }
    ]
  }
}
```

### File Protection

Block edits to sensitive files:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|MultiEdit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "python3 -c \"import json, sys; data=json.load(sys.stdin); path=data.get('tool_input',{}).get('file_path',''); sys.exit(2 if any(p in path for p in ['.env', 'package-lock.json', '.git/']) else 0)\""
          }
        ]
      }
    ]
  }
}
```

### Context Injection

Add current timestamp to user prompts:

```python
#!/usr/bin/env python3
import json
import sys
import datetime

input_data = json.load(sys.stdin)
context = f"Current time: {datetime.datetime.now()}"
print(context)
sys.exit(0)
```

## Advanced Hook Patterns

### Project-Specific Scripts

Use project-relative paths with `CLAUDE_PROJECT_DIR`:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/check-style.sh"
          }
        ]
      }
    ]
  }
}
```

### Complex Validation Logic

Python script for bash command validation:

```python
#!/usr/bin/env python3
import json
import re
import sys

# Validation rules as (regex pattern, message) tuples
VALIDATION_RULES = [
    (
        r"\bgrep\b(?!.*\|)",
        "Use 'rg' (ripgrep) instead of 'grep' for better performance"
    ),
    (
        r"\bfind\s+\S+\s+-name\b",
        "Use 'rg --files -g pattern' instead of 'find -name'"
    ),
]

def validate_command(command: str) -> list[str]:
    issues = []
    for pattern, message in VALIDATION_RULES:
        if re.search(pattern, command):
            issues.append(message)
    return issues

try:
    input_data = json.load(sys.stdin)
except json.JSONDecodeError as e:
    print(f"Error: Invalid JSON input: {e}", file=sys.stderr)
    sys.exit(1)

tool_name = input_data.get("tool_name", "")
tool_input = input_data.get("tool_input", {})
command = tool_input.get("command", "")

if tool_name != "Bash" or not command:
    sys.exit(0)

# Validate command
issues = validate_command(command)

if issues:
    for message in issues:
        print(f"• {message}", file=sys.stderr)
    # Exit code 2 blocks tool call and shows stderr to Claude
    sys.exit(2)
```

### Auto-Approval with JSON Output

Auto-approve documentation file reads:

```python
#!/usr/bin/env python3
import json
import sys

try:
    input_data = json.load(sys.stdin)
except json.JSONDecodeError as e:
    print(f"Error: Invalid JSON input: {e}", file=sys.stderr)
    sys.exit(1)

tool_name = input_data.get("tool_name", "")
tool_input = input_data.get("tool_input", {})

# Auto-approve file reads for documentation files
if tool_name == "Read":
    file_path = tool_input.get("file_path", "")
    if file_path.endswith((".md", ".mdx", ".txt", ".json")):
        output = {
            "hookSpecificOutput": {
                "hookEventName": "PreToolUse",
                "permissionDecision": "allow",
                "permissionDecisionReason": "Documentation file auto-approved"
            },
            "suppressOutput": True
        }
        print(json.dumps(output))
        sys.exit(0)

# Let normal permission flow proceed
sys.exit(0)
```

## MCP Tool Integration

Claude Code hooks work seamlessly with [Model Context Protocol (MCP) tools](mcp). MCP tools appear in hooks with a special naming pattern.

### MCP Tool Naming

MCP tools follow the pattern: `mcp__<server>__<tool>`

Examples:
- `mcp__memory__create_entities` - Memory server's create entities tool
- `mcp__filesystem__read_file` - Filesystem server's read file tool
- `mcp__github__search_repositories` - GitHub server's search tool

### Hook Configuration for MCP Tools

Target specific MCP tools or entire servers:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "mcp__memory__.*",
        "hooks": [
          {
            "type": "command",
            "command": "echo 'Memory operation initiated' >> ~/mcp-operations.log"
          }
        ]
      },
      {
        "matcher": "mcp__.*__write.*",
        "hooks": [
          {
            "type": "command",
            "command": "/home/user/scripts/validate-mcp-write.py"
          }
        ]
      }
    ]
  }
}
```

## Hook Management

### Configuration via `/hooks` Command

Use the `/hooks` slash command for interactive hook management:

1. Run `/hooks` to open the configuration interface
2. Select the hook event type (e.g., `PreToolUse`)
3. Add matchers for specific tools
4. Configure hook commands
5. Choose save location (user vs project settings)

### Configuration File Locations

| Location | Scope | Use Case |
|----------|-------|----------|
| `~/.claude/settings.json` | All projects for user | Personal automation, notification preferences |
| `.claude/settings.json` | Current project, shared | Team workflows, project-specific rules |
| `.claude/settings.local.json` | Current project, local only | Personal project customizations |

### Configuration Security

Claude Code protects against malicious hook modification:

1. Captures hook snapshot at startup
2. Uses this snapshot throughout the session
3. Warns if hooks are modified externally
4. Requires review via `/hooks` menu for changes to take effect

## Security Considerations

### Important Disclaimers

**Use at your own risk**: Claude Code hooks automatically execute arbitrary shell commands on your system. By using hooks, you acknowledge that:

- You are fully responsible for the commands you configure
- Hooks can modify, delete, or access any files your user account can access
- Malicious or poorly written hooks can cause data loss or system compromise
- Anthropic provides no warranties and is not liable for damages from hook usage
- You must thoroughly test hooks in a safe environment before production use

Always review and understand hook commands before adding them to your configuration.

### Security Best Practices

1. **Input Validation**: Never blindly trust input data
2. **Quote Shell Variables**: Use `"$VAR"` not `$VAR`
3. **Block Path Traversal**: Check for `..` in file paths
4. **Use Absolute Paths**: Specify full paths for scripts (use `$CLAUDE_PROJECT_DIR` for project paths)
5. **Skip Sensitive Files**: Avoid `.env`, `.git/`, keys, etc.
6. **Limit Permissions**: Only grant necessary file access
7. **Test Thoroughly**: Validate hooks in isolated environments first

### Example Secure Hook

```python
#!/usr/bin/env python3
import json
import sys
import os
import re

def is_safe_path(path):
    """Check if path is safe from traversal attacks"""
    if not path:
        return False

    # Block path traversal
    if '..' in path:
        return False

    # Block sensitive files
    sensitive_patterns = ['.env', '.git/', 'id_rsa', 'password']
    if any(pattern in path.lower() for pattern in sensitive_patterns):
        return False

    return True

try:
    input_data = json.load(sys.stdin)
    file_path = input_data.get('tool_input', {}).get('file_path', '')

    if not is_safe_path(file_path):
        print("Blocked access to sensitive file", file=sys.stderr)
        sys.exit(2)

except Exception as e:
    print(f"Hook error: {e}", file=sys.stderr)
    sys.exit(1)

# Safe to proceed
sys.exit(0)
```

## Debugging and Troubleshooting

### Basic Troubleshooting

When hooks aren't working:

1. **Check Configuration**: Run `/hooks` to verify hooks are registered
2. **Validate JSON**: Ensure settings file has valid JSON syntax
3. **Test Commands**: Run hook commands manually first
4. **Check Permissions**: Verify scripts are executable (`chmod +x`)
5. **Review Logs**: Use `claude --debug` for detailed execution information

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| Hook not executing | Invalid matcher pattern | Check tool name matches exactly (case-sensitive) |
| Command not found | Relative path in command | Use absolute paths or `$CLAUDE_PROJECT_DIR` |
| JSON syntax error | Unescaped quotes | Use `\"` for quotes in JSON strings |
| Permission denied | Script not executable | Run `chmod +x script-path` |
| Timeout | Hook takes too long | Optimize script or increase `timeout` value |

### Debug Mode

Use `claude --debug` to see detailed hook execution:

```
[DEBUG] Executing hooks for PostToolUse:Write
[DEBUG] Getting matching hook commands for PostToolUse with query: Write
[DEBUG] Found 1 hook matchers in settings
[DEBUG] Matched 1 hooks for query "Write"
[DEBUG] Found 1 hook commands to execute
[DEBUG] Executing hook command: <Your command> with timeout 60000ms
[DEBUG] Hook command completed with status 0: <Your stdout>
```

### Hook Testing

Test hooks before deployment:

```bash
# Test hook command with mock input
echo '{"tool_name": "Write", "tool_input": {"file_path": "test.ts"}}' | your-hook-script.py

# Verify hook configuration
claude --debug
> /hooks

# Test with actual Claude Code usage
claude "create a simple test file"
```

## Hook Execution Details

### Execution Environment

- **Working Directory**: Current directory when hook triggers
- **Environment**: Inherits Claude Code's environment
- **Special Variables**: `CLAUDE_PROJECT_DIR` points to project root
- **Timeout**: 60 seconds default, configurable per command
- **Parallelization**: All matching hooks run in parallel

### Performance Considerations

- **Keep hooks fast**: Hooks run synchronously and can slow down Claude Code
- **Use timeouts**: Set appropriate timeouts to prevent hanging
- **Optimize scripts**: Cache expensive operations where possible
- **Minimize I/O**: Reduce file system operations in critical paths

### Hook Lifecycle

1. Event triggers in Claude Code
2. Find all matching hooks for the event
3. Execute hooks in parallel
4. Collect exit codes and output
5. Process results according to hook type
6. Continue or block based on hook decisions

Hooks provide a powerful extension mechanism for Claude Code, enabling custom workflows, security policies, and automation that integrates seamlessly with your development process.