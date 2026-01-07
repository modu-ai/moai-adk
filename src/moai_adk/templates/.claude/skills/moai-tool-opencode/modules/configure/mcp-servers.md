---
source: https://opencode.ai/docs/mcp-servers/
fetched: 2026-01-08
title: MCP Servers Configuration
---

# MCP Servers in OpenCode

## Overview

OpenCode enables integration of external tools through the Model Context Protocol (MCP), supporting both local and remote servers that become automatically available to language models.

## Local MCP Servers

Use a command-based approach:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "filesystem": {
      "type": "local",
      "command": ["npx", "-y", "@modelcontextprotocol/server-filesystem"],
      "environment": {
        "MCP_ROOT": "/path/to/root"
      }
    }
  }
}
```

### Local Server Options

| Option | Description |
|--------|-------------|
| `type` | Set to `"local"` |
| `command` | Array of command and arguments |
| `environment` | Environment variables |

## Remote MCP Servers

Connect via HTTP:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "my-remote-server": {
      "type": "remote",
      "url": "https://example.com/mcp",
      "headers": {
        "Authorization": "Bearer TOKEN"
      }
    }
  }
}
```

### Remote Server Options

| Option | Description |
|--------|-------------|
| `type` | Set to `"remote"` |
| `url` | Server endpoint URL |
| `headers` | HTTP headers (auth, etc.) |

## Authentication

### Automatic OAuth

OpenCode handles OAuth automatically for remote servers, including dynamic client registration.

### Manual Authentication

```bash
# Trigger OAuth flow
opencode mcp auth <server-name>

# Check auth status
opencode mcp list

# Remove credentials
opencode mcp logout <server-name>
```

### Disable OAuth

For API-key-based servers, disable OAuth:

```json
{
  "mcp": {
    "api-server": {
      "type": "remote",
      "url": "https://api.example.com/mcp",
      "oauth": false,
      "headers": {
        "X-API-Key": "{env:API_KEY}"
      }
    }
  }
}
```

## Management

### Global Control

Enable/disable across all agents:

```json
{
  "tools": {
    "mcp_filesystem_*": false
  }
}
```

### Per-Agent Configuration

Activate selectively for specific agents:

```json
{
  "agent": {
    "file-worker": {
      "tools": {
        "mcp_filesystem_*": true
      }
    }
  }
}
```

### Glob Patterns

Use wildcards to manage multiple servers:

```json
{
  "tools": {
    "mcp_*": false,
    "mcp_approved_*": true
  }
}
```

## Performance Consideration

"MCP servers add to your context, so you want to be careful with which ones you enable."

Excessive tools can quickly exhaust context limits.

### Best Practices

1. **Enable only needed servers**: Disable unused MCP servers
2. **Use per-agent configuration**: Only enable for relevant agents
3. **Monitor context usage**: Watch for context limit warnings
4. **Use glob patterns**: Manage groups of tools efficiently

## Examples

### GitHub MCP Server

```json
{
  "mcp": {
    "github": {
      "type": "local",
      "command": ["npx", "-y", "@modelcontextprotocol/server-github"],
      "environment": {
        "GITHUB_TOKEN": "{env:GITHUB_TOKEN}"
      }
    }
  }
}
```

### Database MCP Server

```json
{
  "mcp": {
    "postgres": {
      "type": "local",
      "command": ["npx", "-y", "@modelcontextprotocol/server-postgres"],
      "environment": {
        "DATABASE_URL": "{env:DATABASE_URL}"
      }
    }
  }
}
```

### Multiple Servers

```json
{
  "mcp": {
    "filesystem": {
      "type": "local",
      "command": ["npx", "-y", "@modelcontextprotocol/server-filesystem"]
    },
    "github": {
      "type": "local",
      "command": ["npx", "-y", "@modelcontextprotocol/server-github"]
    },
    "custom-api": {
      "type": "remote",
      "url": "https://api.myservice.com/mcp"
    }
  }
}
```
