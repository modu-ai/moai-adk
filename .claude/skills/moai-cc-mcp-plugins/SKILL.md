---
name: "Configuring MCP Servers & Plugins for Claude Code"
description: "Set up Model Context Protocol servers (GitHub, Filesystem, Brave Search, SQLite). Configure OAuth, manage permissions, validate MCP structure. Use when integrating external tools, APIs, or expanding Claude Code capabilities."
allowed-tools: "Read, Write, Edit, Bash, Glob"
---

# Configuring MCP Servers & Plugins

MCP servers extend Claude Code with external tool integrations. Each server provides tools that Claude can invoke directly.

## MCP Server Setup in settings.json

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-server-github"],
      "oauth": {
        "clientId": "your-client-id",
        "clientSecret": "your-client-secret",
        "scopes": ["repo", "issues", "pull_requests"]
      }
    },
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/allowed/files"]
    },
    "sqlite": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-sqlite", "/path/to/database.db"]
    },
    "brave-search": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-brave-search"],
      "env": {
        "BRAVE_SEARCH_API_KEY": "${BRAVE_SEARCH_API_KEY}"
      }
    }
  }
}
```

## Common MCP Servers

| Server | Purpose | Installation | Config |
|--------|---------|--------------|--------|
| **GitHub** | PR/issue management, code search | `@anthropic-ai/mcp-server-github` | OAuth required |
| **Filesystem** | Safe file access with path restrictions | `@modelcontextprotocol/server-filesystem` | Path whitelist required |
| **SQLite** | Database queries & migrations | `@modelcontextprotocol/server-sqlite` | DB file path |
| **Brave Search** | Web search integration | `@modelcontextprotocol/server-brave-search` | API key required |

## OAuth Configuration Pattern

```json
{
  "oauth": {
    "clientId": "your-client-id",
    "clientSecret": "your-client-secret",
    "scopes": ["repo", "issues"]
  }
}
```

**Scope Minimization** (principle of least privilege):
- GitHub: `repo` (code access), `issues` (PR/issue access)
- NOT `admin`, NOT `delete_repo`

## Filesystem MCP: Path Whitelisting

```json
{
  "filesystem": {
    "command": "npx",
    "args": [
      "-y",
      "@modelcontextprotocol/server-filesystem",
      "${CLAUDE_PROJECT_DIR}/.moai",
      "${CLAUDE_PROJECT_DIR}/src",
      "${CLAUDE_PROJECT_DIR}/tests"
    ]
  }
}
```

**Security Principle**: Explicitly list allowed directories, no wildcards.

## Plugin Marketplace Integration

```json
{
  "extraKnownMarketplaces": [
    {
      "name": "company-plugins",
      "url": "https://github.com/your-org/claude-plugins"
    },
    {
      "name": "community-plugins",
      "url": "https://glama.ai/mcp/servers"
    }
  ]
}
```

## MCP Health Check

```bash
# Inside Claude Code terminal
/mcp                    # List active MCP servers
/plugin validate        # Validate plugin structure
/plugin install         # Install from marketplace
/plugin enable github   # Enable specific server
/plugin disable github  # Disable specific server
```

## Environment Variables for MCP

```bash
# Set in ~/.bash_profile or .claude/config.json
export GITHUB_TOKEN="gh_xxxx..."
export BRAVE_SEARCH_API_KEY="xxxx..."
export ANTHROPIC_API_KEY="sk-ant-..."

# Launch Claude Code with env
GITHUB_TOKEN=gh_xxxx claude
```

## MCP Troubleshooting

| Issue | Cause | Fix |
|-------|-------|-----|
| Server not connecting | Invalid JSON in mcpServers | Validate with `jq .mcpServers settings.json` |
| OAuth error | Token expired or invalid scopes | Check `claude /usage`, regenerate token |
| Permission denied | Path not whitelisted | Add to Filesystem MCP args |
| Slow response | Network latency or server overload | Check server logs, reduce scope |

## Best Practices

✅ **DO**:
- Use environment variables for secrets
- Whitelist Filesystem paths explicitly
- Start with minimal scopes, expand only if needed
- Test MCP connection: `/mcp` command

❌ **DON'T**:
- Hardcode credentials in settings.json
- Use wildcard paths (`/` in Filesystem MCP)
- Install untrusted plugins
- Give admin scopes unnecessarily

## Plugin Custom Directory Structure

```
my-plugin/
├── .claude-plugin/
│   └── plugin.json
├── commands/
│   ├── deploy.md
│   └── rollback.md
├── agents/
│   └── reviewer.md
└── hooks/
    └── pre-deploy-check.sh
```

## Validation Checklist

- [ ] All server paths are absolute
- [ ] OAuth secrets stored in env vars
- [ ] Filesystem paths are whitelisted
- [ ] No hardcoded tokens or credentials
- [ ] MCP server installed: `which npx`
- [ ] Health check passes: `/mcp`
- [ ] Scopes follow least-privilege principle

---

## 🤝 Works Well With

**Complementary Skills:**
- **moai-cc-settings** - Register MCP servers in settings.json
- **moai-foundation-git** - GitHub MCP integrates with GitFlow automation
- **moai-cc-hooks** - PostToolUse Hook can trigger external APIs via MCP
- **cc-manager agent** - Automate MCP configuration and testing

**MoAI-ADK Workflows:**
- **`/alfred:0-project`** - Optional: Enable GitHub MCP for PR automation
- **`/alfred:3-sync`** - GitHub MCP creates/updates PR with TAG metadata
- **Document generation** - Filesystem MCP safe access to .moai/ directory
- **Research** - Brave Search MCP finds technical documentation

**Example Integration (MoAI-ADK):**
```bash
# 1. Configure GitHub MCP in settings.json
@agent-cc-manager "Set up GitHub MCP for /alfred:3-sync PR creation"

# 2. Filesystem MCP for .moai/ access
@agent-cc-manager "Configure Filesystem MCP restricted to .moai/, src/, tests/"

# 3. Use in workflow
/alfred:3-sync  # Creates PR via GitHub MCP with all changes
```

**Common MoAI Patterns:**
- ✅ **GitHub MCP** = Automate PR/issue creation in /alfred:3-sync
- ✅ **Filesystem MCP** = Safe structured access to .moai/ directory
- ✅ **Brave Search MCP** = Research technical docs when writing SPEC
- ✅ **SQLite MCP** = Store SPEC progress and project metadata

**General Claude Code Patterns:**
- ✅ External tool integration (APIs, services)
- ✅ OAuth authentication for secure access
- ✅ Marketplace plugins for team sharing
- ✅ Least-privilege path restrictions

**See Also:**
- 📖 **Orchestrator Guide:** `Skill("moai-cc-guide")` → SKILL.md
- 📖 **Sync Phase:** `Skill("moai-cc-guide")` → workflows/alfred-3-sync-flow.md
- 📖 **Settings Integration:** `Skill("moai-cc-settings")` → MCP Configuration

---

**Reference**: Claude Code MCP documentation
**Version**: 1.0.0
