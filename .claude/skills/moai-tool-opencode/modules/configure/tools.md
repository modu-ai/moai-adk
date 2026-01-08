---
source: https://opencode.ai/docs/tools/
fetched: 2026-01-08
title: Tools Configuration
---

# OpenCode Tools Documentation

## Overview

OpenCode provides a comprehensive tools system enabling LLMs to interact with your codebase. The platform includes built-in tools and supports extensibility through custom tools and MCP servers.

## Configuration Options

### Global Setup

Tools are enabled by default. Disable specific tools via the config:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "tools": {
    "write": false,
    "bash": false,
    "webfetch": true
  }
}
```

Wildcard patterns control multiple tools:

```json
{
  "tools": {
    "mymcp_*": false
  }
}
```

This disables all tools from that MCP server.

### Per-Agent Override

Agent-specific configurations supersede global settings:

```json
{
  "agent": {
    "plan": {
      "tools": {
        "write": false,
        "bash": false
      }
    }
  }
}
```

## Built-in Tools

| Tool | Purpose |
|------|---------|
| `bash` | Execute shell commands in your project environment |
| `edit` | Precise file modifications through exact string replacement |
| `write` | Create new or overwrite existing files |
| `read` | Retrieve file contents with line-range support |
| `grep` | Search file contents using regular expressions |
| `glob` | Pattern-based file discovery, sorted by modification time |
| `list` | Directory enumeration with glob pattern filtering |
| `lsp` | Code intelligence (experimental; requires flag) |
| `patch` | Apply patch files to codebases |
| `skill` | Load skill files (SKILL.md) into conversations |
| `todowrite` | Track multi-step task progress |
| `todoread` | Retrieve current task list state |
| `webfetch` | Retrieve web content for documentation lookup |

## Tool Descriptions

### bash

Execute shell commands in your project environment. Useful for running tests, build commands, and system operations.

### edit

Precise file modifications through exact string replacement. Matches and replaces specific text patterns.

### write

Create new files or overwrite existing files with specified content.

### read

Retrieve file contents with optional line-range support for large files.

### grep

Search file contents using regular expressions. Supports pattern matching across multiple files.

### glob

Pattern-based file discovery. Returns files matching glob patterns, sorted by modification time.

### list

Directory enumeration with optional glob pattern filtering.

### lsp

Code intelligence features including:
- Go to definition
- Find references
- Symbol search

**Note:** Experimental feature requiring `--experimental-lsp` flag.

### patch

Apply unified diff patch files to codebases.

### skill

Load skill files (SKILL.md) into conversations for specialized knowledge.

### todowrite / todoread

Track and retrieve multi-step task progress.

### webfetch

Retrieve web content for documentation lookup and research.

## Advanced Configuration

### Ignore Patterns

The system uses ripgrep, respecting `.gitignore` by default. Override with `.ignore`:

```
!node_modules/
!dist/
!build/
```

### Extensibility

**Custom Tools:**
User-defined functions in config files. See Custom Tools documentation.

**MCP Servers:**
External services and database integrations. See MCP Servers documentation.

## Permissions

Permissions can be configured separately to require approval before tool execution:

```json
{
  "permission": {
    "edit": "ask",
    "bash": "ask",
    "write": "allow"
  }
}
```

Options:
- `"ask"`: Require user approval
- `"allow"`: Auto-approve
- `"deny"`: Block tool usage
