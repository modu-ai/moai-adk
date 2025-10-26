---
name: moai-cc-settings
description: "Configuring Claude Code settings.json & Security. Set up permissions (allow/deny), permission modes, environment variables, tool restrictions. Use when securing Claude Code, restricting tool access, or optimizing session settings."
allowed-tools: "Read, Write, Edit, Bash"
---

## Skill Metadata

| Field | Value |
| ----- | ----- |
| Version | 1.0.0 |
| Tier | Ops |
| Auto-load | When configuring security & permissions |

## What It Does

settings.json 설정 및 보안 구성을 위한 전체 가이드를 제공합니다. Permissions (allow/deny), permission modes, environment variables, tool restrictions 설정 방법을 다룹니다.

## When to Use

- 새 프로젝트의 settings.json을 설정할 때
- Tool access를 제한하거나 보안을 강화할 때
- Environment variables를 구성할 때
- Permission mode (ask/allow/deny)를 변경할 때


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

**Reference**: Claude Code settings.json documentation
**Version**: 1.0.0
