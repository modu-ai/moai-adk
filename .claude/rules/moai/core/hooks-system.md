# Hooks System

Claude Code hooks for extending functionality with custom scripts.

## Hook Types

Available hook event types:

- SessionStart: Execute when a new session begins
- PreCompact: Execute before context compaction
- PreToolUse: Execute before a tool runs (supports matcher for tool-specific hooks)
- PostToolUse: Execute after a tool completes (supports matcher for tool-specific hooks)
- Stop: Execute when conversation stops

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
