---
source: https://opencode.ai/docs/permissions/
fetched: 2026-01-08
title: Permissions Configuration
---

# OpenCode Permissions Documentation

## Overview

OpenCode manages action approval through a `permission` config system that determines whether actions run automatically, require approval, or are blocked entirely.

## Core Permission States

Actions resolve to three outcomes:

| State | Description |
|-------|-------------|
| `"allow"` | Automatic execution |
| `"ask"` | User approval required |
| `"deny"` | Blocked |

## Configuration Structure

### Global Configuration

Set permissions globally using `"*"` as a catch-all, with specific tool overrides:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "permission": {
    "*": "ask",
    "bash": "allow",
    "edit": "deny"
  }
}
```

### Granular Pattern Control

Object syntax enables granular control based on input patterns:

```json
{
  "permission": {
    "bash": {
      "*": "ask",
      "git *": "allow",
      "rm *": "deny"
    }
  }
}
```

### Pattern Matching

- `*`: Matches any characters
- `?`: Matches single character
- Last matching rule takes precedence

## Available Permissions

OpenCode protects the following actions:

| Permission | Description |
|------------|-------------|
| `read` | File reading |
| `edit` | File editing |
| `glob` | Pattern matching |
| `grep` | Content searching |
| `list` | Directory listing |
| `bash` | Shell command execution |
| `task` | Subagent invocation |
| `skill` | Skill loading |
| `lsp` | Language server operations |
| `todoread` | Task list reading |
| `todowrite` | Task list writing |
| `webfetch` | Web content retrieval |
| `websearch` | Web searching |
| `codesearch` | Code searching |
| `external_directory` | External directory access |
| `doom_loop` | Runaway loop prevention |

## Default Behavior

| Permission | Default |
|------------|---------|
| Most permissions | `"allow"` |
| `doom_loop` | `"ask"` |
| `external_directory` | `"ask"` |
| `*.env` files | `"deny"` |

**Note:** Files matching `*.env` patterns are denied by default to protect sensitive data.

## Agent-Level Overrides

Agents can override global permissions:

```json
{
  "agent": {
    "careful-agent": {
      "permission": {
        "*": "ask",
        "read": "allow"
      }
    }
  }
}
```

Agent rules take precedence in merged configurations.

## Examples

### Restrictive Configuration

```json
{
  "permission": {
    "*": "ask",
    "read": "allow",
    "grep": "allow",
    "glob": "allow"
  }
}
```

### Allow Git, Deny Destructive

```json
{
  "permission": {
    "bash": {
      "*": "ask",
      "git *": "allow",
      "rm *": "deny",
      "sudo *": "deny"
    }
  }
}
```

### Safe Read-Only Mode

```json
{
  "permission": {
    "edit": "deny",
    "write": "deny",
    "bash": "deny"
  }
}
```

## Best Practices

1. **Start restrictive**: Use `"*": "ask"` and allow specific actions
2. **Protect sensitive files**: Ensure `.env` and credentials are denied
3. **Use patterns**: Create specific rules for common commands
4. **Test permissions**: Verify behavior before production use
