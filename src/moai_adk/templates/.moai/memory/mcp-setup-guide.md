# MCP Integration Setup Guide

**Model Context Protocol (MCP) configuration and best practices for MoAI-ADK.**

> **See also**: CLAUDE.md → "MCP Integration & External Services" for quick overview

---

## Context7 MCP Setup

Context7 provides up-to-date API documentation and code references.

### Installation

```bash
# Context7 auto-installed with Claude Code
# Verify in .claude/mcp.json

npx @upstash/context7-mcp@latest
```

### Configuration (.claude/mcp.json)

```json
{
  "mcpServers": {
    "context7": {
      "command": "npx",
      "args": ["-y", "@upstash/context7-mcp@latest"],
      "env": {
        "CONTEXT7_SESSION_STORAGE": ".moai/sessions/",
        "CONTEXT7_CACHE_SIZE": "1GB",
        "CONTEXT7_SESSION_TTL": "30d"
      }
    }
  }
}
```

### Usage

```bash
# Resolve library ID
mcp__context7__resolve-library-id("React")

# Get documentation
mcp__context7__get-library-docs("/facebook/react")
mcp__context7__get-library-docs("/facebook/react", topic="hooks")
```

---

## GitHub MCP Setup

GitHub MCP enables repository operations through Claude Code.

### Configuration

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-server-github"],
      "oauth": {
        "clientId": "your-client-id",
        "clientSecret": "your-client-secret",
        "scopes": ["repo", "issues"]
      }
    }
  }
}
```

### Usage

```bash
@github list issues
@github create pull-request --title "Feature" --body "Description"
gh pr create --title "Title"
```

---

## Filesystem MCP Setup

Safe filesystem access through MCP.

### Configuration

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/files"]
    }
  }
}
```

---

## Troubleshooting MCP

### MCP Connection Issues

```bash
# Check MCP server status
claude mcp serve

# Validate configuration
claude /doctor

# Restart MCP servers
/mcp restart
```

### Context7 Caching

```bash
# Clear Context7 cache if outdated
rm -rf .moai/sessions/

# Check session storage
ls -la .moai/sessions/
```

---

**Last Updated**: 2025-11-18
**Format**: Markdown | **Language**: Korean
