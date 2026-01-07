---
source: https://opencode.ai/docs/config/
fetched: 2026-01-08
title: Configuration
---

# OpenCode Configuration Guide

## Overview

OpenCode uses JSON/JSONC config files to customize the tool's behavior across TUI, CLI, IDE, and other interfaces.

## Configuration Locations (Merge Order)

Configuration files are merged together, not replaced. Settings from the following config locations are combined:

1. **Global**: `~/.config/opencode/opencode.json`
2. **Per-project**: `opencode.json` in project root (merged with global)
3. **Custom path**: Via `OPENCODE_CONFIG` environment variable
4. **Custom directory**: Via `OPENCODE_CONFIG_DIR` environment variable

## Core Configuration Options

### Models & Providers

```json
{
  "$schema": "https://opencode.ai/config.json",
  "model": "anthropic/claude-sonnet-4-5",
  "small_model": "anthropic/claude-haiku-3-5",
  "provider": {
    "anthropic": {
      "timeout": 30000,
      "setCacheKey": true,
      "region": "us-east-1",
      "profile": "default",
      "endpoint": "https://api.anthropic.com"
    }
  }
}
```

- `model`: Primary model (e.g., "anthropic/claude-sonnet-4-5")
- `small_model`: Lightweight tasks like title generation
- `provider`: Provider-specific settings (timeout, setCacheKey, region, profile, endpoint)

### Interface Settings

```json
{
  "theme": "dark",
  "tui": {
    "scroll_speed": 3,
    "scroll_acceleration": 1.5,
    "diff_style": "unified"
  },
  "server": {
    "port": 3000,
    "hostname": "localhost",
    "mdns": true,
    "cors": true
  }
}
```

- `theme`: Visual appearance configuration
- `tui`: Terminal UI options (scroll_speed, scroll_acceleration, diff_style)
- `server`: Server settings (port, hostname, mdns, cors)

### Functionality

```json
{
  "tools": {
    "write": true,
    "bash": true
  },
  "agent": {
    "code-review": {
      "description": "Code review specialist",
      "model": "anthropic/claude-sonnet-4",
      "prompt": "You are a code review expert..."
    }
  },
  "default_agent": "code-review",
  "command": {
    "test": {
      "description": "Run tests",
      "command": "npm test"
    }
  },
  "keybinds": {
    "submit": "ctrl+enter"
  },
  "formatter": {
    "typescript": "prettier"
  },
  "permission": {
    "write": "ask",
    "bash": "allow"
  }
}
```

- `tools`: Enable/disable tools like write, bash
- `agent`: Define specialized agents for specific tasks
- `default_agent`: Agent used when none specified
- `command`: Custom commands for repetitive workflows
- `keybinds`: Keyboard shortcuts customization
- `formatter`: Code formatter configuration
- `permission`: Tool approval requirements ("ask", "allow", etc.)

### Features & Behavior

```json
{
  "share": "manual",
  "autoupdate": true,
  "compaction": "auto",
  "watcher": {
    "ignore": ["node_modules", ".git"]
  },
  "mcp": {
    "servers": {}
  },
  "plugin": [],
  "instructions": ["AGENTS.md", ".opencode/*.md"],
  "disabled_providers": [],
  "enabled_providers": [],
  "experimental": {}
}
```

- `share`: Sharing mode ("manual", "auto", "disabled")
- `autoupdate`: Auto-update behavior (true/false/"notify")
- `compaction`: Context management (auto, prune)
- `watcher`: File watching ignore patterns
- `mcp`: Model Context Protocol server configuration
- `plugin`: Load custom plugins from npm or local files
- `instructions`: Array of instruction files/glob patterns
- `disabled_providers`: Blocklist of providers
- `enabled_providers`: Allowlist of providers
- `experimental`: Unstable features under development

## Variable Substitution

### Environment Variables

Use `{env:VARIABLE_NAME}` syntax:

```json
{
  "model": "{env:OPENCODE_MODEL}"
}
```

### File Contents

Use `{file:path/to/file}` syntax:

```json
{
  "apiKey": "{file:~/.secrets/openai-key}"
}
```

File paths support relative (to config) and absolute paths (`/` or `~`).

## Provider-Specific Options

### Amazon Bedrock

Amazon Bedrock supports the following options:

- `region`: AWS region (defaults to AWS_REGION env or us-east-1)
- `profile`: AWS named profile
- `endpoint`: Custom VPC endpoint URL

**Note:** Bearer tokens take precedence over profile-based authentication.

```json
{
  "provider": {
    "amazon-bedrock": {
      "region": "us-west-2",
      "profile": "production",
      "endpoint": "https://bedrock.us-west-2.amazonaws.com"
    }
  }
}
```

## Agent Configuration

Agents can be defined inline in the config or via markdown files:

- Global: `~/.config/opencode/agent/`
- Project: `.opencode/agent/`

Configuration includes:
- `description`: What the agent does
- `model`: Which model to use
- `prompt`: System prompt for the agent
- Tool restrictions

## Important Notes

- The schema is defined at `opencode.ai/config.json`
- Experimental options are not stable. They may change or be removed without notice.
- `disabled_providers` takes priority over `enabled_providers`
- Sharing defaults to manual mode requiring explicit `/share` command
