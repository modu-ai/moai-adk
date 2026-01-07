---
source: https://opencode.ai/docs/acp/
fetched: 2026-01-08
title: ACP Support
---

# ACP Support in OpenCode

## Overview

OpenCode integrates with the Agent Client Protocol (ACP), an open standard that enables communication between code editors and AI agents. This allows you to use OpenCode directly in compatible IDEs.

## Supported Editors

- Zed Editor
- JetBrains IDEs
- Neovim (via plugins)

## Configuration Examples

### Zed Editor

Add to Zed settings:

```json
{
  "agent_servers": {
    "OpenCode": {
      "command": "opencode",
      "args": ["acp"]
    }
  }
}
```

### JetBrains IDEs

Requires absolute path to the OpenCode binary in `acp.json`:

```json
{
  "agent_servers": {
    "OpenCode": {
      "command": "/absolute/path/bin/opencode",
      "args": ["acp"]
    }
  }
}
```

**Note:** JetBrains requires the full absolute path to the binary.

### Avante.nvim

Add to your Neovim configuration:

```lua
{
  acp_providers = {
    ["opencode"] = {
      command = "opencode",
      args = { "acp" }
    }
  }
}
```

### CodeCompanion.nvim

```lua
require("codecompanion").setup({
  strategies = {
    chat = {
      adapter = {
        name = "opencode",
        model = "claude-sonnet-4",
      },
    },
  },
})
```

## Supported Features

The ACP interface maintains full feature parity with terminal usage:

| Feature | Supported |
|---------|-----------|
| File operations | Yes |
| Custom tools | Yes |
| MCP servers | Yes |
| Project rules | Yes |
| Formatters | Yes |
| Permissions system | Yes |
| `/undo` command | No |
| `/redo` command | No |

**Note:** Built-in slash commands like `/undo` and `/redo` remain unsupported via ACP.

## Running ACP Mode

Start OpenCode in ACP mode:

```bash
opencode acp
```

This starts OpenCode in Agent Client Protocol mode via stdin/stdout.

## Best Practices

1. **Use absolute paths**: JetBrains requires full paths
2. **Verify binary location**: Use `which opencode` to find path
3. **Test connection**: Start with simple commands to verify setup
4. **Configure per-project**: Use project-level config for specific needs
