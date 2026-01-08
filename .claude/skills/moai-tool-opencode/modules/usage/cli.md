---
source: https://opencode.ai/docs/cli/
fetched: 2026-01-08
title: Command Line Interface (CLI)
---

# OpenCode CLI Documentation

## Overview

OpenCode is a CLI tool that defaults to launching a terminal user interface (TUI) when run without arguments. It supports both interactive and programmatic modes, allowing users to interact with the system through various commands and flags.

## Basic Usage

```bash
# Launch TUI in current directory
opencode

# Launch TUI in specific directory
opencode /path/to/project

# Run non-interactive with a prompt
opencode run "Your prompt here"
```

## Global Flags

| Flag | Description |
|------|-------------|
| `--help` | Display help information |
| `--version` | Display version information |
| `--print-logs` | Print logs to stdout |
| `--log-level` | Set logging level (DEBUG, INFO, WARN, ERROR) |

## Core Commands

### Agent Management

Create, list, and manage custom agents with configurable system prompts and tool configurations.

```bash
opencode agent list
opencode agent create my-agent
```

### Authentication

Handle credentials across multiple providers.

```bash
opencode auth login
opencode auth list
opencode auth logout
```

Credentials are stored in `~/.local/share/opencode/auth.json`.

### Session Control

List, export, and import session data in JSON format.

```bash
opencode session list
opencode session export <session-id>
opencode session import <file.json>
```

Supports OpenCode share URLs for importing sessions.

### Model Operations

Display available models from configured providers in `provider/model` format.

```bash
opencode models
opencode models --refresh  # Refresh cache
```

### Server Modes

**Headless HTTP API Server:**
```bash
opencode serve
```

**HTTP Server with Web Browser Interface:**
```bash
opencode web
```

**Agent Client Protocol Server (stdin/stdout):**
```bash
opencode acp
```

### GitHub Integration

Install workflows and run agents within GitHub Actions environments.

```bash
opencode github install
opencode github run
```

### MCP Server Management

Add, list, and authenticate Model Context Protocol servers with OAuth support.

```bash
opencode mcp add <server-name>
opencode mcp list
opencode mcp auth <server-name>
```

## Non-Interactive Mode

The `opencode run` command enables non-interactive execution:

```bash
opencode run "Your prompt here"
```

"Run opencode in non-interactive mode by passing a prompt directly."

## Environment Variables

| Variable | Description |
|----------|-------------|
| `OPENCODE_CONFIG_DIR` | Custom configuration directory |
| `OPENCODE_DISABLE_AUTOUPDATE` | Disable automatic updates |

## Advanced Features

### Attachment Capability

Connect to running servers to avoid cold boot times on subsequent invocations.

### Experimental Flags

Advanced features may be available through experimental flags. Check `--help` for current options.

## Examples

```bash
# Start TUI with debug logging
opencode --log-level DEBUG

# Run a quick prompt
opencode run "Explain the authentication flow in this codebase"

# Start headless server on custom port
opencode serve --port 8080

# List all available models
opencode models

# Export current session
opencode session export my-session > session.json
```
