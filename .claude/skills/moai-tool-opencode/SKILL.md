---
name: moai-tool-opencode
description: OpenCode.ai open-source AI coding agent comprehensive reference. Use when working with OpenCode TUI, CLI, IDE integration, configuring agents, tools, MCP servers, creating plugins, or developing with the SDK.
version: 2.0.0
user-invocable: false
status: active
updated: 2026-01-08
allowed-tools:
  - Read
  - Grep
  - Glob
  - WebFetch
  - Bash
---

# OpenCode.ai Tool Skill

Comprehensive reference for OpenCode.ai - an open-source AI coding agent available as terminal interface, desktop application, and IDE extension.

## Quick Reference

**Installation:**
```bash
curl -fsSL https://opencode.ai/install | bash
# or
npm install -g opencode
```

**Start OpenCode:**
```bash
opencode              # Launch TUI in current directory
opencode /path/to/dir # Launch in specific directory
opencode run "prompt" # Non-interactive mode
opencode serve        # Start headless server
```

**Essential Commands:**
- `/connect` - Configure API keys
- `/init` - Generate AGENTS.md
- `/models` - List available models
- `/share` - Share session
- `/undo` - Undo changes
- `Tab` - Switch agents
- `Ctrl+P` - Command palette

**Configuration Files:**
- `~/.config/opencode/opencode.json` - Global config
- `./opencode.json` - Project config
- `AGENTS.md` - Project instructions
- `.opencode/agent/` - Custom agents
- `.opencode/skill/` - Agent skills
- `.opencode/tool/` - Custom tools
- `.opencode/command/` - Custom commands

---

## Documentation Index

This skill contains 32 comprehensive documentation modules organized by category.

### Core (7 modules)
- [Introduction](modules/core/intro.md) - Installation, setup, primary use cases
- [Configuration](modules/core/config.md) - JSON configuration options and merge order
- [Providers](modules/core/providers.md) - 75+ LLM providers setup
- [Network](modules/core/network.md) - Proxy and certificate configuration
- [Enterprise](modules/core/enterprise.md) - SSO, central config, AI gateway
- [Troubleshooting](modules/core/troubleshooting.md) - Common issues and solutions
- [Migration 1.0](modules/core/migration-1.0.md) - Upgrade guide

### Usage (7 modules)
- [TUI](modules/usage/tui.md) - Terminal interface reference
- [CLI](modules/usage/cli.md) - Command-line interface
- [IDE](modules/usage/ide.md) - VS Code/Cursor integration
- [Zen](modules/usage/zen.md) - Curated model gateway
- [Share](modules/usage/share.md) - Session sharing
- [GitHub](modules/usage/github.md) - GitHub Actions integration
- [GitLab](modules/usage/gitlab.md) - GitLab CI/CD integration

### Configuration (14 modules)
- [Tools](modules/configure/tools.md) - Built-in and custom tools
- [Rules](modules/configure/rules.md) - AGENTS.md instructions
- [Agents](modules/configure/agents.md) - Custom agent configuration
- [Models](modules/configure/models.md) - Model selection and variants
- [Themes](modules/configure/themes.md) - Custom theme creation
- [Keybinds](modules/configure/keybinds.md) - Keyboard shortcuts
- [Commands](modules/configure/commands.md) - Custom slash commands
- [Formatters](modules/configure/formatters.md) - Code formatting
- [Permissions](modules/configure/permissions.md) - Tool approval system
- [LSP Servers](modules/configure/lsp-servers.md) - Language server setup
- [MCP Servers](modules/configure/mcp-servers.md) - Model Context Protocol
- [ACP Support](modules/configure/acp.md) - Agent Client Protocol
- [Skills](modules/configure/skills.md) - Reusable instructions
- [Custom Tools](modules/configure/custom-tools.md) - TypeScript tools

### Development (4 modules)
- [SDK](modules/develop/sdk.md) - JavaScript/TypeScript client
- [Server](modules/develop/server.md) - HTTP API reference
- [Plugins](modules/develop/plugins.md) - Plugin development
- [Ecosystem](modules/develop/ecosystem.md) - Community projects

See [index.md](index.md) for complete navigation.

---

## Key Concepts

### Agent Architecture

**Primary Agents** - Main assistants you interact with directly:
- Build: Full tools enabled (default)
- Plan: Restricted, analysis-focused

**Subagents** - Specialized assistants invoked by primary agents:
- General: Research and multi-step tasks
- Explore: Fast codebase exploration

### Configuration Hierarchy

1. Command-line flags (highest priority)
2. Environment variables
3. Project config (`./opencode.json`)
4. Global config (`~/.config/opencode/opencode.json`)
5. Built-in defaults

### Tool Permission Levels

| Level | Description |
|-------|-------------|
| `allow` | Auto-approve execution |
| `ask` | Request user approval |
| `deny` | Block execution |

---

## Common Patterns

### Basic Configuration

```json
{
  "$schema": "https://opencode.ai/config.json",
  "model": "anthropic/claude-sonnet-4",
  "agent": {
    "security": {
      "description": "Security auditor",
      "permission": { "edit": "deny" }
    }
  }
}
```

### Custom Command

```markdown
---
description: Review current changes
agent: plan
---
Review the following changes:
!`git diff --staged`
```

### MCP Server Setup

```json
{
  "mcp": {
    "github": {
      "type": "local",
      "command": ["npx", "-y", "@modelcontextprotocol/server-github"],
      "environment": { "GITHUB_TOKEN": "{env:GITHUB_TOKEN}" }
    }
  }
}
```

### Custom Tool

```typescript
import { tool } from "@opencode-ai/plugin"

export default tool({
  description: "Query project database",
  args: {
    query: tool.schema.string().describe("SQL query")
  },
  async execute(args) {
    return `Executed: ${args.query}`
  }
})
```

---

## Resources

- Documentation: https://opencode.ai/docs
- GitHub: https://github.com/anomalyco/opencode
- Discord: https://opencode.ai/discord

Storage Locations:
- Config: `~/.config/opencode/`
- Data: `~/.local/share/opencode/`
- Cache: `~/.cache/opencode/`
- Logs: `~/.local/share/opencode/log/`

---

## Works Well With

- moai-lang-typescript - TypeScript development patterns
- moai-lang-python - Python development patterns
- moai-domain-backend - Backend API development
- moai-domain-frontend - Frontend development
- moai-workflow-testing - Testing workflows
