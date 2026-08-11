# Hook Development Guidelines

> Relocated from `CLAUDE.local.md` §7 to reduce launch-time context weight. CLAUDE.local.md retains a 1-line pointer.

### [HARD] Shell Script Hooks Only

moai-adk-go uses shell scripts for hooks, NOT Python:

**Hook Wrapper Pattern:**
```bash
#!/bin/bash
# .claude/hooks/moai/handle-session-start.sh

# Read stdin JSON from Claude Code
INPUT=$(cat)

# Call moai binary with hook subcommand
moai hook session-start <<< "$INPUT"
```

**Why Shell Scripts:**
- Faster execution (no Python startup overhead)
- Always available (no dependency on uv/python)
- Cross-platform (bash, /bin/sh)

### Hook Command Format

**settings.json hook configuration:**
```json
{
  "hooks": {
    "SessionStart": [{
      "hooks": [{
        "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-session-start.sh\"",
        "timeout": 5
      }]
    }]
  }
}
```

**Key Rules:**
- [HARD] Always quote `$CLAUDE_PROJECT_DIR`: `"$CLAUDE_PROJECT_DIR"`
- [HARD] Use full path to hook wrapper script
- [HARD] Set appropriate timeout. MoAI policy default is 5 seconds (the Claude Code platform default is 10 minutes; MoAI tightens this to 5 seconds to avoid stalling the session).

### Platform Differences

**macOS/Linux:**
```json
"command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/hook.sh\""
```

**Windows:**
```json
"command": "\"%CLAUDE_PROJECT_DIR%\\.claude\\hooks\\moai\\hook.sh\""
```
