---
name: "Configuring Claude Code settings.json & Security"
description: "Set up permissions (allow/deny), permission modes, environment variables, tool restrictions. Use when securing Claude Code, restricting tool access, or optimizing session settings."
allowed-tools: "Read, Write, Edit, Bash"
---

# Configuring Claude Code settings.json

`settings.json` centralizes all Claude Code configuration: permissions, tool access, environment variables, and session behavior.

**Location**: `.claude/settings.json`

## Complete Configuration Template

```json
{
  "permissions": {
    "allowedTools": [
      "Read(**/*.{js,ts,json,md})",
      "Edit(**/*.{js,ts})",
      "Glob(**/*)",
      "Grep(**/*)",
      "Bash(git:*)",
      "Bash(npm:*)",
      "Bash(npm run:*)",
      "Bash(pytest:*)",
      "Bash(python:*)"
    ],
    "deniedTools": [
      "Read(./.env)",
      "Read(./.env.*)",
      "Read(./secrets/**)",
      "Bash(rm -rf:*)",
      "Bash(sudo:*)",
      "Bash(curl:*)"
    ]
  },
  "permissionMode": "ask",
  "spinnerTipsEnabled": true,
  "disableAllHooks": false,
  "env": {
    "ANTHROPIC_API_KEY": "${ANTHROPIC_API_KEY}",
    "GITHUB_TOKEN": "${GITHUB_TOKEN}",
    "CLAUDE_CODE_ENABLE_TELEMETRY": "1"
  },
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "bash ~/.claude/hooks/pre-bash-check.sh"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Edit",
        "hooks": [
          {
            "type": "command",
            "command": "bash ~/.claude/hooks/post-edit-format.sh"
          }
        ]
      }
    ],
    "SessionStart": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "bash ~/.claude/hooks/session-init.sh"
          }
        ]
      }
    ]
  },
  "statusLine": {
    "enabled": true,
    "type": "command",
    "command": "~/.claude/statusline.sh"
  },
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-server-github"],
      "oauth": {
        "clientId": "${GITHUB_CLIENT_ID}",
        "clientSecret": "${GITHUB_CLIENT_SECRET}",
        "scopes": ["repo", "issues"]
      }
    },
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "${CLAUDE_PROJECT_DIR}/.moai", "${CLAUDE_PROJECT_DIR}/src"]
    }
  },
  "extraKnownMarketplaces": [
    {
      "name": "company-plugins",
      "url": "https://github.com/your-org/claude-plugins"
    }
  ]
}
```

## Permission Modes

| Mode | Behavior | Use Case |
|------|----------|----------|
| **allow** | Execute all allowed tools without asking | Trusted environments |
| **ask** | Ask before executing each tool | Development (safer) |
| **deny** | Deny all tools except whitelisted | Restrictive (default) |

```json
{
  "permissionMode": "ask"
}
```

## Tool Permission Patterns

### Restrictive (Recommended for teams)
```json
{
  "allowedTools": [
    "Read(src/**)",
    "Edit(src/**/*.ts)",
    "Bash(npm run test:*)",
    "Glob(src/**)"
  ],
  "deniedTools": [
    "Bash(rm:*)",
    "Bash(sudo:*)",
    "Read(.env)"
  ]
}
```

### Permissive (Local development only)
```json
{
  "allowedTools": [
    "Read",
    "Write",
    "Edit",
    "Bash(git:*)",
    "Bash(npm:*)",
    "Bash(python:*)",
    "Glob",
    "Grep"
  ]
}
```

## Environment Variables Pattern

```json
{
  "env": {
    "ANTHROPIC_API_KEY": "${ANTHROPIC_API_KEY}",
    "GITHUB_TOKEN": "${GITHUB_TOKEN}",
    "BRAVE_SEARCH_API_KEY": "${BRAVE_SEARCH_API_KEY}",
    "NODE_ENV": "development"
  }
}
```

**Security rule**: Never hardcode secrets; always use `${VAR_NAME}` syntax.

## Dangerous Tools to Deny

```json
{
  "deniedTools": [
    "Bash(rm -rf:*)",           // Recursive delete
    "Bash(sudo:*)",             // Privilege escalation
    "Bash(curl.*|.*bash)",      // Code injection
    "Read(.env)",               // Secrets
    "Read(.ssh/**)",            // SSH keys
    "Read(/etc/shadow)",        // System secrets
    "Edit(/etc/**)",            // System files
  ]
}
```

## Permission Validation

```bash
# Check current permissions
cat .claude/settings.json | jq '.permissions'

# Validate JSON syntax
jq . .claude/settings.json

# List allowed tools
jq '.permissions.allowedTools[]' .claude/settings.json
```

## Spinner Tips Configuration

```json
{
  "spinnerTipsEnabled": true
}
```

Custom tips can be added for better UX during long operations.

## Best Practices

✅ **DO**:
- Use `ask` mode for teams
- Explicitly whitelist paths
- Environment variables for all secrets
- Review permissions regularly
- Document why each denial exists

❌ **DON'T**:
- Hardcode credentials in settings.json
- Use `allow` mode for untrusted contexts
- Grant `Bash(*)` without restrictions
- Include secrets in version control
- Mix personal and project settings

## Permission Checklist

- [ ] All secrets use `${VAR_NAME}` syntax
- [ ] Dangerous patterns are denied
- [ ] File paths are explicit (not wildcards)
- [ ] Permission mode matches use case (ask/allow/deny)
- [ ] Hooks are not left in commented state
- [ ] MCP servers have proper OAuth configuration
- [ ] No `.env` file is readable
- [ ] Sudo commands are denied

---

## 🤝 Works Well With

**Complementary Skills:**
- **moai-cc-hooks** - Register Hooks in settings.json
- **moai-cc-mcp-plugins** - Configure MCP servers in settings.json
- **moai-cc-agents** - Apply tool restrictions per agent YAML
- **cc-manager agent** - Automate settings.json generation and validation

**MoAI-ADK Workflows:**
- **`/alfred:0-project`** - Creates initial settings.json with project permissions
- **`/alfred:2-run`** - Uses settings to restrict edits to src/, tests/
- **`/alfred:3-sync`** - Uses GitHub MCP configured in settings.json
- **All phases** - Permission mode ("ask") protects against accidents

**Example Integration (MoAI-ADK):**
```bash
# 1. Setup project security
@agent-cc-manager "Configure settings.json for /alfred:0-project:
  - allowedTools: Read(src/**, tests/**), Edit(src/**), Bash(pytest:*)
  - deniedTools: Read(.env), Bash(rm -rf:*), Bash(sudo:*)
  - permissionMode: 'ask'"

# 2. Register Hooks for quality gates
@agent-cc-manager "Add PreToolUse Hook to validate @TAG"

# 3. Configure GitHub MCP for /alfred:3-sync
@agent-cc-manager "Register GitHub MCP with GITHUB_TOKEN"
```

**Common MoAI Patterns:**
- ✅ Restrict edits to specific paths (src/, tests/, docs/)
- ✅ Block dangerous patterns (rm -rf, sudo, curl | bash)
- ✅ Deny access to .env, secrets/, SSH keys
- ✅ Allow language-specific commands (pytest, npm, go, cargo)
- ✅ Integrate Hooks and MCP via settings.json

**General Claude Code Patterns:**
- ✅ Fine-grained permissions (patterns, paths)
- ✅ Permission modes (ask, allow, deny)
- ✅ Environment variable management
- ✅ MCP server orchestration

**See Also:**
- 📖 **Orchestrator Guide:** `Skill("moai-cc-guide")` → SKILL.md
- 📖 **Project Setup:** `Skill("moai-cc-guide")` → workflows/alfred-0-project-setup.md
- 📖 **Hooks Integration:** `Skill("moai-cc-hooks")` → Hook Configuration

---

**Reference**: Claude Code settings.json documentation
**Version**: 1.0.0
