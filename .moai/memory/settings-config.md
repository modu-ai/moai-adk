---
filename: settings-config.md
version: 1.0.0
updated_date: 2025-11-18
language: English
scope: Both
---

# Claude Code Settings Configuration

**Complete guide for configuring Claude Code v4.0 settings for MoAI-ADK.**

> **See also**: CLAUDE.md → "Claude Code Settings" for quick reference

---

## .claude/settings.json Structure

### Basic Template

```json
{
  "permissions": {
    "allowedTools": [
      "Read(**/*.{js,ts,json,md})",
      "Edit(**/*.{js,ts})",
      "Bash(git:*)",
      "Bash(npm:*)",
      "Bash(node:*)",
      "Bash(uv:*)",
      "Bash(pytest:*)",
      "Bash(mypy:*)"
    ],
    "deniedTools": [
      "Edit(/config/secrets.json)",
      "Edit(.env*)",
      "Bash(rm -rf:*)",
      "Bash(sudo:*)",
      "Bash(chmod:*)"
    ]
  },
  "permissionMode": "acceptEdits",
  "spinnerTipsEnabled": true,
  "sandbox": {
    "allowUnsandboxedCommands": false
  },
  "hooks": {
    "PreToolUse": [],
    "PostToolUse": [],
    "SessionStart": []
  },
  "mcpServers": {},
  "statusLine": {
    "enabled": true,
    "format": "{{model}} | {{tokens}} | {{thinking}}"
  }
}
```

---

## Permission Configuration

### Allowed Tools (Whitelist)

**Read Operations**:
```json
"allowedTools": [
  "Read(**/*.{js,ts,json,md})",
  "Read(**/*.{py,yaml,yml})"
]
```

**Edit Operations**:
```json
"allowedTools": [
  "Edit(**/*.{js,ts})",
  "Edit(**/*.py)"
]
```

**Bash Commands**:
```json
"allowedTools": [
  "Bash(git:*)",
  "Bash(npm:*)",
  "Bash(uv:*)",
  "Bash(pytest:*)",
  "Bash(mypy:*)",
  "Bash(ruff:*)"
]
```

### Denied Tools (Blacklist)

**High-Risk Operations**:
```json
"deniedTools": [
  "Edit(/config/secrets.json)",
  "Edit(.env*)",
  "Edit(.aws/**)",
  "Edit(.vercel/**)",
  "Bash(rm -rf:*)",
  "Bash(sudo:*)",
  "Bash(chmod 777:*)"
]
```

---

## Sandbox Configuration

### Sandbox Mode (Recommended)

```json
{
  "sandbox": {
    "allowUnsandboxedCommands": false,
    "validatedCommands": ["git:*", "npm:*", "uv:*"]
  }
}
```

**Benefits**:
- Prevents accidental destructive commands
- Validates shell expressions
- Logs all command execution
- Sandboxes execution environment

### Unsandboxed Mode (Not Recommended)

```json
{
  "sandbox": {
    "allowUnsandboxedCommands": true
  }
}
```

**Risks**:
- Direct system access
- No validation
- Potential security issues

---

## Hooks Configuration

### Pre-Tool-Use Hook (Validation)

**Purpose**: Validate commands before execution

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "python3 .claude/hooks/validate-command.py"
          }
        ]
      }
    ]
  }
}
```

**Hook Script** (.claude/hooks/validate-command.py):
```python
#!/usr/bin/env python3
import re
import sys
import json

DANGEROUS_PATTERNS = [
    r"rm -rf",
    r"sudo ",
    r":/.*\.\.",
    r"&&.*rm",
    r"\|.*sh"
]

def validate_command(command):
    for pattern in DANGEROUS_PATTERNS:
        if re.search(pattern, command):
            return False, f"Dangerous pattern: {pattern}"
    return True, "Safe"

if __name__ == "__main__":
    input_data = json.load(sys.stdin)
    command = input_data.get("command", "")
    is_safe, message = validate_command(command)

    if not is_safe:
        print(f"SECURITY BLOCK: {message}", file=sys.stderr)
        sys.exit(2)
    sys.exit(0)
```

### Session-Start Hook (Context Seeding)

**Purpose**: Initialize context at session start

```json
{
  "hooks": {
    "SessionStart": [
      {
        "type": "command",
        "command": "uv run .moai/scripts/statusline.py"
      }
    ]
  }
}
```

---

## MCP Server Configuration

### Context7 Integration

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

### GitHub MCP Server

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-server-github"],
      "oauth": {
        "clientId": "your-github-client-id",
        "clientSecret": "your-github-client-secret",
        "scopes": ["repo", "issues", "pull_requests"]
      }
    }
  }
}
```

### Filesystem MCP Server

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

## Status Line Configuration

### Format Options

```json
{
  "statusLine": {
    "enabled": true,
    "format": "{{model}} | {{tokens}} | {{thinking}}"
  }
}
```

**Available Variables**:
- `{{model}}` - Current Claude model (Haiku, Sonnet, Opus)
- `{{tokens}}` - Token usage
- `{{thinking}}` - Thinking mode status
- `{{temperature}}` - Temperature setting
- `{{maxTokens}}` - Max tokens per response

---

## Best Practices

### Security

- ✅ Keep sandbox mode enabled
- ✅ Use whitelist approach (explicit allow)
- ✅ Validate dangerous patterns
- ✅ Never allow `rm -rf` or `sudo` commands
- ✅ Protect `.env*` and credential files

### Development

- ✅ Allow `uv`, `pytest`, `mypy`, `ruff` commands
- ✅ Allow git operations
- ✅ Allow Node/Python package management
- ✅ Configure MCP for Context7 integration
- ✅ Use hooks for automatic validation

### Maintenance

- ✅ Review settings quarterly
- ✅ Monitor denied tool usage
- ✅ Update MCP server versions
- ✅ Test new tool permissions
- ✅ Log security events

---

## Common Issues

### "Permission Denied" Error

**Cause**: Tool not in `allowedTools` list

**Solution**:
```json
{
  "permissions": {
    "allowedTools": [
      "Your-Tool-Pattern:*"
    ]
  }
}
```

### Hook Execution Fails

**Cause**: Hook script missing or permission error

**Solution**:
```bash
chmod +x .claude/hooks/*.py
# Verify hook exists:
ls -la .claude/hooks/
```

### MCP Server Not Connecting

**Cause**: Missing environment variables or invalid config

**Solution**:
```bash
# Test MCP server:
npx @upstash/context7-mcp@latest

# Check config:
cat .claude/mcp.json | jq '.mcpServers'
```

---

## Security Checklist

- [ ] Sandbox mode: Enabled
- [ ] Dangerous patterns: Blocked
- [ ] `.env*` files: Protected
- [ ] `.aws/` directory: Protected
- [ ] `.vercel/` directory: Protected
- [ ] `sudo` commands: Disabled
- [ ] `rm -rf` commands: Disabled
- [ ] MCP servers: Configured
- [ ] Hooks: Validated
- [ ] Permissions: Tested

---

**Last Updated**: 2025-11-18
**Version**: v0.26.0
**Format**: Markdown | **Language**: Korean
