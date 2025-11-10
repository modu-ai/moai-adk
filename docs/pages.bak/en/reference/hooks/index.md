# Hooks System Reference

Automatic guardrails and context management through Claude Code's Hook system.

## Overview

**Hooks** are scripts that automatically execute when specific events occur. MoAI-ADK provides 4 main Hooks.

### Hook Types

| Hook                 | Timing           | Purpose               | Timeout |
| -------------------- | ---------------- | --------------------- | ------- |
| **SessionStart**     | Session start    | Project status check  | 5s      |
| **PreToolUse**       | Before tool execution | Block dangerous commands | 5s      |
| **UserPromptSubmit** | After user input | Input validation      | 5s      |
| **PostToolUse**      | After tool execution | Result analysis      | 5s      |

## Hook Location

```
.claude/
├── hooks/
│   ├── session_start.sh       # SessionStart Hook
│   ├── pre_tool_use.sh        # PreToolUse Hook
│   ├── post_tool_use.sh       # PostToolUse Hook
│   └── user_prompt_submit.sh  # UserPromptSubmit Hook
├── settings.json              # Hook settings
└── permissions.json           # Permission settings
```

## Hook Execution Flow

```
┌─────────────────────────────────────┐
│  Claude Code Session Start           │
└────────────────┬────────────────────┘
                 │
            ┌────▼────────────┐
            │ SessionStart    │ (Project status check)
            └────┬────────────┘
                 │
      ┌──────────▼──────────┐
      │ User command input   │
      └──────────┬──────────┘
                 │
            ┌────▼────────────┐
            │PreToolUse       │ (Pre-execution validation)
            └────┬────────────┘
                 │
          ┌──────▼──────────┐
          │ Tool execution  │
          └──────┬──────────┘
                 │
            ┌────▼────────────┐
            │PostToolUse      │ (Result analysis)
            └────┬────────────┘
                 │
      ┌──────────▼──────────┐
      │  Deliver result to user│
      └─────────────────────┘
```

## Hook Configuration

### .claude/settings.json

```json
{
  "hooks": {
    "enabled": true,
    "timeout": 5000,
    "session_start": ".claude/hooks/session_start.sh",
    "pre_tool_use": ".claude/hooks/pre_tool_use.sh",
    "post_tool_use": ".claude/hooks/post_tool_use.sh",
    "user_prompt_submit": ".claude/hooks/user_prompt_submit.sh"
  }
}
```

## 🆘 Hook Error Handling

### Hook Execution Failure

```
:x: Hook failure
│
├─ Timeout (exceeds 5s)
│  └─→ Tool execution blocked
│
├─ Script error
│  └─→ Error log saved
│
└─ Permission error
   └─→ Permission adjustment needed
```

### Debugging

```bash
# Check Hook logs
cat ~/.claude/projects/*/hook-logs/*.log

# Manually execute Hook
bash .claude/hooks/session_start.sh

# Disable Hooks
# Set "hooks.enabled" → false in settings.json
```

## <span class="material-icons">library_books</span> Detailed Guides

- **[SessionStart Hook](session.md)** - Auto-execute on session start
- **[Tool Hooks](tool.md)** - Pre/post tool execution processing

______________________________________________________________________

**Next**: [SessionStart Hook](session.md) or [Tool Hooks](tool.md)
