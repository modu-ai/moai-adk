---
source: https://opencode.ai/docs/commands/
fetched: 2026-01-08
title: Custom Commands
---

# OpenCode Commands Documentation

## Overview

OpenCode enables users to create custom commands for repetitive tasks using the `/command-name` syntax, supplementing built-in commands like `/init`, `/undo`, and `/share`.

## File Creation & Configuration

### Markdown Approach

Create `.md` files in:
- Project-level: `.opencode/command/`
- Global: `~/.config/opencode/command/`

The filename becomes the command identifier.

**Example: `.opencode/command/review.md`**

```markdown
---
description: Review current changes
agent: plan
---

Review the current git diff and provide feedback on:
- Code quality
- Potential bugs
- Performance concerns
```

### JSON Configuration

Use the `command` object in `opencode.jsonc`:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "command": {
    "review": {
      "description": "Review current changes",
      "agent": "plan",
      "template": "Review the current git diff and provide feedback."
    }
  }
}
```

## Template Features

### Arguments

Use placeholders for dynamic content:

- `$ARGUMENTS`: All arguments as string
- `$1`, `$2`, `$3`: Positional arguments

**Example:**
```markdown
Fix the issue described: $ARGUMENTS
```

Usage: `/fix The button is not clickable`

### Shell Integration

Execute bash and embed output with `` !`command` `` syntax:

```markdown
Review the following changes:

!`git diff --staged`
```

### File References

Include file content with `@filename` syntax:

```markdown
Analyze this component:

@src/components/Button.tsx
```

## Configuration Options

| Option | Purpose | Required |
|--------|---------|----------|
| `template` | The prompt sent to the LLM | Yes |
| `description` | UI label displayed in TUI | No |
| `agent` | Specifies which agent executes | No |
| `subtask` | Forces subagent invocation | No |
| `model` | Overrides default model | No |

## Override Built-in Commands

Custom commands can override built-in commands:

```json
{
  "command": {
    "init": {
      "description": "Custom project initialization",
      "template": "Initialize project with my custom setup..."
    }
  }
}
```

If you define a custom command with the same name as a built-in, it will override the built-in command.

## Examples

### Code Review Command

```markdown
---
description: Comprehensive code review
agent: plan
---

Review the following code changes:

!`git diff HEAD~1`

Focus on:
1. Security vulnerabilities
2. Performance issues
3. Code style consistency
4. Test coverage gaps
```

### Test Generator

```markdown
---
description: Generate tests for a file
---

Generate comprehensive tests for $1:

@$1

Include:
- Unit tests
- Edge cases
- Error handling tests
```

Usage: `/test src/utils/validator.ts`

### Documentation Generator

```markdown
---
description: Generate API documentation
agent: docs
model: anthropic/claude-sonnet-4
---

Generate API documentation for:

@$1

Format as:
- Function signature
- Parameters table
- Return value
- Usage examples
```

### Commit Message Generator

```markdown
---
description: Generate commit message
---

Generate a conventional commit message for these changes:

!`git diff --staged`

Follow the format: type(scope): description
```

## Best Practices

1. **Use descriptive names**: Command names should indicate their purpose
2. **Add descriptions**: Help users understand command purpose
3. **Specify agents**: Use appropriate agents for different tasks
4. **Combine features**: Mix file refs, shell output, and arguments
5. **Version control**: Store project commands in `.opencode/command/`
