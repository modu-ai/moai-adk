---
source: https://opencode.ai/docs/lsp/
fetched: 2026-01-08
title: LSP Servers Configuration
---

# LSP Servers Documentation

## Overview

OpenCode integrates Language Server Protocol (LSP) functionality to help LLMs interact with codebases through diagnostic feedback. The platform supports numerous built-in LSP servers across popular programming languages.

## Key Features

### Automatic Detection

LSP servers activate when:
- File extensions matching supported languages are detected
- Server requirements are met (runtime, dependencies)

### Built-in Support

OpenCode includes 25+ LSP servers:

| Language | LSP Server |
|----------|------------|
| TypeScript/JavaScript | TypeScript Language Server |
| Python | Pyright |
| Rust | rust-analyzer |
| Go | gopls |
| PHP | Intelephense |
| C/C++ | clangd |
| Java | Eclipse JDT LS |
| Ruby | Solargraph |
| Swift | SourceKit-LSP |
| Kotlin | Kotlin Language Server |
| Scala | Metals |
| Elixir | ElixirLS |
| Lua | lua-language-server |
| Zig | zls |
| And more... | |

## Configuration Options

Configure LSP behavior in `opencode.json`:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "lsp": {
    "typescript": {
      "disabled": false,
      "command": ["typescript-language-server", "--stdio"],
      "extensions": [".ts", ".tsx", ".js", ".jsx"],
      "env": {
        "NODE_ENV": "development"
      },
      "initialization": {
        "preferences": {
          "includeCompletionsForModuleExports": true
        }
      }
    }
  }
}
```

### Configuration Properties

| Property | Type | Description |
|----------|------|-------------|
| `disabled` | boolean | Turn off specific server |
| `command` | string[] | Custom startup command |
| `extensions` | string[] | Associated file types |
| `env` | object | Environment variables |
| `initialization` | object | LSP initialization parameters |

## Disabling LSP

### Disable All LSP Servers

```json
{
  "$schema": "https://opencode.ai/config.json",
  "lsp": false
}
```

### Disable Specific Server

```json
{
  "lsp": {
    "typescript": {
      "disabled": true
    }
  }
}
```

## Custom LSP Server

Add custom LSP servers:

```json
{
  "lsp": {
    "custom-lsp": {
      "command": ["custom-lsp-server", "--stdio"],
      "extensions": [".custom", ".cst"]
    }
  }
}
```

## Special Configurations

### PHP Intelephense Premium

Supports premium licensing through key files:

| Platform | License Path |
|----------|--------------|
| Unix/Linux/macOS | `~/intelephense/licence.txt` |
| Windows | `%USERPROFILE%\intelephense\licence.txt` |

### Pyright Configuration

```json
{
  "lsp": {
    "pyright": {
      "initialization": {
        "python": {
          "analysis": {
            "typeCheckingMode": "strict"
          }
        }
      }
    }
  }
}
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `OPENCODE_DISABLE_LSP_DOWNLOAD` | Prevents automatic server installations |

## Best Practices

1. **Let auto-detection work**: Most servers activate automatically
2. **Disable unused servers**: Reduce resource usage
3. **Configure per-project**: Use project-level config for specific needs
4. **Check requirements**: Ensure language runtimes are installed
