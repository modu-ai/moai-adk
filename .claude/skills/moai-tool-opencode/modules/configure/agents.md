---
source: https://opencode.ai/docs/agents/
fetched: 2026-01-08
title: Agents Configuration
---

# OpenCode Agents Documentation

## Overview

OpenCode provides specialized AI assistants called agents that can be configured for specific tasks and workflows. They enable focused tools with custom prompts, models, and tool access.

## Agent Types

### Primary Agents

Main assistants you interact with directly. Switch between them using Tab or configured keybinds.

**Built-in Primary Agents:**
- **Build**: Default agent with all tools enabled for full development work
- **Plan**: Restricted agent for analysis (file edits and bash set to "ask")

### Subagents

Specialized assistants that primary agents can invoke. Manually invoked via @ mentions.

**Built-in Subagents:**
- **General**: Research and multi-step tasks
- **Explore**: Fast codebase exploration

## Built-in Agents Summary

| Agent | Type | Description |
|-------|------|-------------|
| Build | Primary | All tools enabled for full development work |
| Plan | Primary | Restricted permissions for analysis |
| General | Subagent | Researching questions and executing multi-step tasks |
| Explore | Subagent | Quick codebase exploration |

## Usage

### Switching Primary Agents

- Press `Tab` to cycle through primary agents
- Use `switch_agent` keybind

### Invoking Subagents

- Automatic: Primary agents invoke subagents as needed
- Manual: Use `@general` or `@explore` syntax

### Navigation

- `<Leader>+Right`: Navigate to child session
- `<Leader>+Left`: Navigate to parent session
- Configure with `session_child_cycle` keybinds

## Configuration Methods

### JSON Configuration

In `opencode.json`:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "agent": {
    "security-auditor": {
      "description": "Security vulnerability scanner",
      "model": "anthropic/claude-sonnet-4",
      "temperature": 0.3,
      "maxSteps": 20,
      "tools": {
        "write": false,
        "bash": false
      },
      "permission": {
        "edit": "deny"
      }
    }
  }
}
```

### Markdown Files

Create markdown files in:
- Global: `~/.config/opencode/agent/`
- Per-project: `.opencode/agent/`

**Example: `.opencode/agent/docs-writer.md`**

```markdown
---
description: Technical documentation writer
model: anthropic/claude-sonnet-4
temperature: 0.7
---

You are a technical documentation specialist.

Focus on:
- Clear, concise writing
- Code examples
- API documentation
- User guides
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `description` | Brief description (required) | - |
| `temperature` | Randomness (0.0-1.0) | Model default |
| `maxSteps` | Max iterations before text-only response | Unlimited |
| `disable` | Set true to disable | false |
| `prompt` | Custom system prompt or file path | - |
| `model` | Override global model | Global model |
| `tools` | Enable/disable specific tools | All enabled |
| `permission` | Tool permissions (ask/allow/deny) | - |
| `mode` | "primary", "subagent", or "all" | "all" |
| `hidden` | Hide from @ autocomplete | false |
| `taskPermissions` | Control subagent invocation | - |

### Tools Configuration

```json
{
  "agent": {
    "readonly": {
      "tools": {
        "write": false,
        "edit": false,
        "bash": false,
        "mcp_*": false
      }
    }
  }
}
```

Supports wildcards for pattern matching.

### Permissions Configuration

```json
{
  "agent": {
    "careful": {
      "permission": {
        "edit": "ask",
        "bash": "ask",
        "webfetch": "allow"
      }
    }
  }
}
```

## Creating Agents

### Interactive Wizard

```bash
opencode agent create
```

The wizard handles:
1. Location selection (global/project)
2. Description input
3. System prompt generation
4. Tool selection
5. Markdown file creation

## Example Use Cases

### Documentation Agent

```json
{
  "agent": {
    "docs": {
      "description": "Technical writing specialist",
      "model": "anthropic/claude-sonnet-4",
      "temperature": 0.7,
      "tools": {
        "bash": false
      }
    }
  }
}
```

### Security Auditor

```json
{
  "agent": {
    "security": {
      "description": "Vulnerability identification",
      "model": "anthropic/claude-opus-4",
      "temperature": 0.1,
      "permission": {
        "edit": "deny",
        "write": "deny"
      }
    }
  }
}
```

### Code Reviewer

```json
{
  "agent": {
    "reviewer": {
      "description": "Code review with read-only access",
      "tools": {
        "write": false,
        "edit": false,
        "bash": false
      }
    }
  }
}
```

### Debug Agent

```json
{
  "agent": {
    "debug": {
      "description": "Issue investigation specialist",
      "model": "anthropic/claude-sonnet-4",
      "prompt": "Focus on debugging and root cause analysis."
    }
  }
}
```
