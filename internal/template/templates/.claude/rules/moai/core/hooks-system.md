# Hooks System

Claude Code hooks for extending functionality with custom scripts.

## Hook Types

Available hook event types:

- PreToolCall: Execute before a tool runs
- PostToolCall: Execute after a tool completes
- UserPromptSubmit: Process user input before handling
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
    "PreToolCall": [".claude/hooks/pre-write-check.py"],
    "UserPromptSubmit": [".claude/hooks/input-filter.sh"]
  }
}
```

## Rules

- Hook feedback is treated as user input
- When blocked, suggest alternatives
- Avoid infinite loops (no recursive tool calls)
- Keep hooks lightweight for performance

## Error Handling

- Failed hooks should exit with non-zero code
- Error messages are displayed to user
- Hooks can block operations by returning error

## Security

- Hooks run in sandbox by default
- Validate all hook inputs
- Do not store secrets in hook scripts

## MoAI Integration

- Skill("moai-foundation-claude") for detailed patterns
- Hook scripts must follow coding-standards.md
