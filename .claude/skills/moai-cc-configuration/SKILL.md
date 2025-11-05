---
name: moai-cc-configuration
version: 1.0.0
created: 2025-11-06
updated: 2025-11-06
status: active
description: Complete Claude Code configuration system including settings.json, permissions, hooks, MCP servers, and plugin management. End-to-end setup for security, tool access, automation, and external integrations. Use when configuring Claude Code, setting up security, managing permissions, implementing hooks, or integrating external tools.
keywords: ['configuration', 'settings', 'permissions', 'hooks', 'mcp', 'plugins', 'security', 'automation']
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - Glob
---

# Claude Code Complete Configuration System

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-cc-configuration |
| **Version** | 1.0.0 (2025-11-06) |
| **Status** | Active |
| **Tier** | Operations |
| **Purpose** | Complete Claude Code configuration management |

---

## What It Does

Comprehensive configuration system for Claude Code covering security settings, permissions, hooks automation, MCP server integration, and plugin management.

**Core capabilities**:
- ✅ Complete settings.json configuration
- ✅ Permission system setup (allow/ask/deny modes)
- ✅ Hook system automation (PreToolUse/PostToolUse/SessionStart)
- ✅ MCP server integration (GitHub, Filesystem, Search)
- ✅ Plugin marketplace management
- ✅ Security best practices enforcement
- ✅ Environment variable management

---

## When to Use

**Initial Setup**:
- New project Claude Code configuration
- Team environment configuration
- Security hardening requirements
- External tool integration needs

**Ongoing Management**:
- Permission adjustments and security updates
- Hook system modifications and automation
- MCP server additions and updates
- Plugin installation and management

**Troubleshooting**:
- Configuration validation issues
- Permission problems
- Hook system failures
- MCP server connectivity issues

---

## Complete Configuration Template

### settings.json Master Template

```json
{
  "permissions": {
    "allowedTools": [
      "Read(**/*.{js,ts,json,md,py,go,rs})",
      "Edit(**/*.{js,ts,py,go,rs})",
      "Write(**/*.{js,ts,py,go,rs,json,md})",
      "Glob(**/*)",
      "Bash(git:*)",
      "Bash(npm:*)",
      "Bash(npm run:*)",
      "Bash(pytest:*)",
      "Bash(python:*)",
      "Bash(go:*)",
      "Bash(rustc:*)"
    ],
    "deniedTools": [
      "Read(./.env)",
      "Read(./.env.*)",
      "Read(./secrets/**)",
      "Read(./.ssh/**)",
      "Read(/etc/**)",
      "Bash(rm -rf:*)",
      "Bash(sudo:*)",
      "Bash(curl.*|.*bash)",
      "Edit(/etc/**)",
      "Write(/etc/**)"
    ]
  },
  "permissionMode": "ask",
  "spinnerTipsEnabled": true,
  "disableAllHooks": false,
  "env": {
    "ANTHROPIC_API_KEY": "${ANTHROPIC_API_KEY}",
    "GITHUB_TOKEN": "${GITHUB_TOKEN}",
    "BRAVE_SEARCH_API_KEY": "${BRAVE_SEARCH_API_KEY}",
    "NODE_ENV": "development",
    "CLAUDE_CODE_ENABLE_TELEMETRY": "1"
  },
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "bash ~/.claude/hooks/pre-bash-validator.sh"
          }
        ]
      },
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "bash ~/.claude/hooks/pre-edit-guard.sh"
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
      },
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "bash ~/.claude/hooks/post-bash-cleanup.sh"
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
            "command": "bash ~/.claude/hooks/session-status-card.sh"
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
        "scopes": ["repo", "issues", "pull_requests"]
      }
    },
    "filesystem": {
      "command": "npx",
      "args": [
        "-y", 
        "@modelcontextprotocol/server-filesystem",
        "${CLAUDE_PROJECT_DIR}/.moai",
        "${CLAUDE_PROJECT_DIR}/src",
        "${CLAUDE_PROJECT_DIR}/tests"
      ]
    },
    "sqlite": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-sqlite", "${CLAUDE_PROJECT_DIR}/data/app.db"]
    },
    "brave-search": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-brave-search"],
      "env": {
        "BRAVE_SEARCH_API_KEY": "${BRAVE_SEARCH_API_KEY}"
      }
    }
  },
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

---

## Permission System Configuration

### Permission Modes

| Mode | Behavior | Use Case | Security Level |
|------|----------|----------|----------------|
| **allow** | Execute all allowed tools without asking | Trusted environments, local development | Low |
| **ask** | Ask before executing each tool | Team development (recommended) | Medium |
| **deny** | Deny all tools except whitelisted | High-security environments | High |

### Permission Patterns by Environment

#### Development Environment (Permissive)
```json
{
  "permissionMode": "ask",
  "permissions": {
    "allowedTools": [
      "Read",
      "Write", 
      "Edit",
      "Bash(git:*)",
      "Bash(npm:*)",
      "Bash(python:*)",
      "Bash(go:*)",
      "Glob",
      "Grep"
    ],
    "deniedTools": [
      "Bash(sudo:*)",
      "Bash(rm -rf:*)",
      "Read(.env)",
      "Read(.ssh/**)"
    ]
  }
}
```

#### Team Environment (Balanced)
```json
{
  "permissionMode": "ask",
  "permissions": {
    "allowedTools": [
      "Read(src/**/*.{js,ts,py,go})",
      "Edit(src/**/*.{js,ts,py,go})",
      "Write(src/**/*.{js,ts,py,go})",
      "Glob(src/**)",
      "Bash(git status)",
      "Bash(git diff)",
      "Bash(npm run test:*)",
      "Bash(pytest:*)"
    ],
    "deniedTools": [
      "Bash(*)",
      "Read(.env*)",
      "Write(.env*)",
      "Edit(.env*)"
    ]
  }
}
```

#### Production Environment (Restrictive)
```json
{
  "permissionMode": "deny",
  "permissions": {
    "allowedTools": [
      "Read(./logs/**)",
      "Bash(git log)",
      "Bash(git status)"
    ]
  }
}
```

### Security Rule Validation

```bash
# Validate JSON syntax
jq . .claude/settings.json

# Check for secrets
rg "sk-ant-|ghp_|gho_|ghu_" .claude/settings.json

# Validate permission patterns
jq '.permissions.allowedTools[]' .claude/settings.json
jq '.permissions.deniedTools[]' .claude/settings.json

# Check environment variables are referenced correctly
jq '.env' .claude/settings.json | grep -E '\$\{[A-Z_]+\}'
```

---

## Hook System Automation

### Hook Types & Use Cases

| Hook Type | Trigger | Execution Limit | Primary Use Cases |
|-----------|---------|------------------|------------------|
| **PreToolUse** | Before any tool execution | <100ms | Input validation, safety checks, permission verification |
| **PostToolUse** | After successful tool execution | <100ms | Auto-formatting, cleanup, linting, permissions restoration |
| **SessionStart** | Claude Code session initialization | <500ms | Context seeding, status display, environment setup |
| **Notification** | User notification events | N/A | macOS notifications, alerts, status updates |
| **Stop** | Session termination | N/A | Cleanup, final reports, state persistence |

### Essential Hook Scripts

#### Pre-Bash Safety Validator
```bash
#!/bin/bash
# ~/.claude/hooks/pre-bash-validator.sh

FORBIDDEN_PATTERNS=(
  "rm -rf /"
  "sudo rm"
  "chmod 777 /"
  "eval \$(curl"
  "curl.*\|.*bash"
  "dd if="
)

COMMAND="$*"
for pattern in "${FORBIDDEN_PATTERNS[@]}"; do
  if [[ "$COMMAND" =~ $pattern ]]; then
    echo "🔴 BLOCKED: Dangerous pattern detected: $pattern" >&2
    exit 2  # Block execution
  fi
done

# Check for suspicious file operations
if [[ "$COMMAND" =~ (>|>>)\s*/etc/ ]]; then
  echo "🔴 BLOCKED: Writing to system files not allowed" >&2
  exit 2
fi

exit 0  # Allow execution
```

#### Post-Edit Auto-Formatter
```bash
#!/bin/bash
# ~/.claude/hooks/post-edit-format.sh

FILE="$1"
EXT="${FILE##*.}"

# Skip if no file argument or non-existent file
[ -z "$FILE" ] || [ ! -f "$FILE" ] && exit 0

case "$EXT" in
  js|ts|jsx|tsx)
    command -v prettier >/dev/null 2>&1 && npx prettier --write "$FILE" 2>/dev/null &
    ;;
  py)
    command -v black >/dev/null 2>&1 && python3 -m black "$FILE" 2>/dev/null &
    ;;
  go)
    command -v gofmt >/dev/null 2>&1 && gofmt -w "$FILE" 2>/dev/null &
    ;;
  rs)
    command -v rustfmt >/dev/null 2>&1 && rustfmt "$FILE" 2>/dev/null &
    ;;
esac

exit 0
```

#### SessionStart Status Display
```bash
#!/bin/bash
# ~/.claude/hooks/session-status-card.sh

echo "🚀 Claude Code Session Started"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Project information
if [ -f ".moai/config.json" ]; then
  PROJECT=$(jq -r '.name // "Unknown"' .moai/config.json 2>/dev/null)
  VERSION=$(jq -r '.moai.version // "unknown"' .moai/config.json 2>/dev/null)
  echo "📦 Project: $PROJECT (v$VERSION)"
  
  TECH_STACK=$(jq -r '.tech_stack // "Auto-detecting..."' .moai/config.json 2>/dev/null)
  echo "🏗️  Stack: $TECH_STACK"
fi

# Recent SPEC activity
if [ -d ".moai/specs" ]; then
  echo ""
  echo "📋 Recent SPECs:"
  ls -t .moai/specs/SPEC-* 2>/dev/null | head -3 | while read spec; do
    SPEC_NAME=$(basename "$spec")
    STATUS=$(jq -r '.status // "unknown"' "$spec/spec.md" 2>/dev/null)
    echo "  ✓ $SPEC_NAME ($STATUS)"
  done
fi

# Git status
if [ -d ".git" ]; then
  echo ""
  echo "🌿 Git: $(git branch --show-current 2>/dev/null || echo 'detached')"
  if [ -n "$(git status --porcelain 2>/dev/null)" ]; then
    echo "  ⚠️  Uncommitted changes"
  else
    echo "  ✅ Working tree clean"
  fi
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "💡 Type /help for available commands"
```

#### Permission Preserver
```bash
#!/bin/bash
# ~/.claude/hooks/pre-edit-permissions.sh

FILE="$1"
[ -z "$FILE" ] && exit 0

# Save file permissions before edit
if [ -f "$FILE" ]; then
  PERMS_FILE="/tmp/claude_perms_${FILE//\//_}"
  stat -c "%a %U:%G" "$FILE" > "$PERMS_FILE" 2>/dev/null || true
fi

exit 0
```

```bash
#!/bin/bash
# ~/.claude/hooks/post-edit-permissions.sh

FILE="$1"
[ -z "$FILE" ] && exit 0

# Restore file permissions after edit
PERMS_FILE="/tmp/claude_perms_${FILE//\//_}"
if [ -f "$PERMS_FILE" ]; then
  SAVED_PERMS=$(cat "$PERMS_FILE")
  chmod ${SAVED_PERMS%% *} "$FILE" 2>/dev/null || true
  chown ${SAVED_PERMS##* } "$FILE" 2>/dev/null || true
  rm "$PERMS_FILE"
fi

exit 0
```

### Hook Installation

```bash
# Create hooks directory
mkdir -p ~/.claude/hooks

# Make scripts executable
chmod +x ~/.claude/hooks/*.sh

# Test hook configuration
jq '.hooks' .claude/settings.json
```

---

## MCP Server Integration

### Essential MCP Servers

#### GitHub Integration
```json
{
  "github": {
    "command": "npx",
    "args": ["-y", "@anthropic-ai/mcp-server-github"],
    "oauth": {
      "clientId": "${GITHUB_CLIENT_ID}",
      "clientSecret": "${GITHUB_CLIENT_SECRET}",
      "scopes": ["repo", "issues", "pull_requests"]
    }
  }
}
```

**Required Environment Variables**:
```bash
export GITHUB_CLIENT_ID="your-github-oauth-app-client-id"
export GITHUB_CLIENT_SECRET="your-github-oauth-app-client-secret"
```

#### Secure Filesystem Access
```json
{
  "filesystem": {
    "command": "npx",
    "args": [
      "-y",
      "@modelcontextprotocol/server-filesystem",
      "${CLAUDE_PROJECT_DIR}/.moai",
      "${CLAUDE_PROJECT_DIR}/src",
      "${CLAUDE_PROJECT_DIR}/tests",
      "${CLAUDE_PROJECT_DIR}/docs"
    ]
  }
}
```

#### Database Integration
```json
{
  "sqlite": {
    "command": "npx",
    "args": ["-y", "@modelcontextprotocol/server-sqlite", "${CLAUDE_PROJECT_DIR}/data/app.db"]
  }
}
```

#### Web Search Integration
```json
{
  "brave-search": {
    "command": "npx",
    "args": ["-y", "@modelcontextprotocol/server-brave-search"],
    "env": {
      "BRAVE_SEARCH_API_KEY": "${BRAVE_SEARCH_API_KEY}"
    }
  }
}
```

### MCP Server Validation

```bash
# Validate MCP configuration
jq '.mcpServers' .claude/settings.json

# Test MCP server connectivity
/mcp                    # List active servers
/mcp validate          # Validate configuration
/mcp health           # Check server health

# Install missing MCP servers
npx -y @anthropic-ai/mcp-server-github --version
npx -y @modelcontextprotocol/server-filesystem --version
```

### MCP Security Best Practices

✅ **DO**:
- Use environment variables for all secrets
- Whitelist specific directories for filesystem access
- Use minimal OAuth scopes
- Regularly rotate API keys

❌ **DON'T**:
- Hardcode credentials in settings.json
- Use root directory `/` for filesystem access
- Request unnecessary OAuth scopes
- Install untrusted MCP servers

---

## Environment Variable Management

### Required Environment Variables

```bash
# Claude Code Core
export ANTHROPIC_API_KEY="sk-ant-..."
export CLAUDE_CODE_ENABLE_TELEMETRY="1"

# MCP Integration
export GITHUB_TOKEN="ghp_..."
export GITHUB_CLIENT_ID="your-oauth-client-id"
export GITHUB_CLIENT_SECRET="your-oauth-client-secret"
export BRAVE_SEARCH_API_KEY="your-brave-search-api-key"

# Project Specific
export CLAUDE_PROJECT_DIR="$(pwd)"
export NODE_ENV="development"
export PYTHON_ENV="development"
```

### Environment Setup Script

```bash
#!/bin/bash
# ~/.claude/setup-env.sh

echo "Setting up Claude Code environment..."

# Check for required environment variables
check_env() {
  local var_name="$1"
  local var_value="${!var_name}"
  
  if [ -z "$var_value" ]; then
    echo "⚠️  $var_name is not set"
    echo "   Add to your ~/.bash_profile or ~/.zshrc:"
    echo "   export $var_name=\"your-value\""
    return 1
  else
    echo "✅ $var_name is set"
    return 0
  fi
}

# Core variables
check_env "ANTHROPIC_API_KEY"
check_env "GITHUB_TOKEN"

# Optional variables
check_env "BRAVE_SEARCH_API_KEY"
check_env "GITHUB_CLIENT_ID"
check_env "GITHUB_CLIENT_SECRET"

# Set project directory
export CLAUDE_PROJECT_DIR="$(pwd)"
echo "✅ CLAUDE_PROJECT_DIR=$CLAUDE_PROJECT_DIR"

echo ""
echo "Environment setup complete!"
```

---

## Plugin Marketplace Management

### Adding Custom Marketplaces

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
    },
    {
      "name": "official-plugins",
      "url": "https://github.com/anthropics/claude-plugins"
    }
  ]
}
```

### Plugin Management Commands

```bash
# Available in Claude Code terminal
/plugin list                    # List installed plugins
/plugin install <plugin-name>   # Install from marketplace
/plugin enable <plugin-name>    # Enable specific plugin
/plugin disable <plugin-name>   # Disable specific plugin
/plugin validate               # Validate plugin structure
/plugin update                 # Update all plugins
```

---

## Configuration Validation

### Complete Validation Checklist

**Settings.json Validation**:
- [ ] JSON syntax is valid: `jq . .claude/settings.json`
- [ ] No hardcoded secrets: `rg "sk-ant-|ghp_|gho_|ghu_" .claude/settings.json`
- [ ] Environment variables properly referenced: `grep -E '\$\{[A-Z_]+\}'`
- [ ] Permission mode matches use case
- [ ] Dangerous operations are in deniedTools

**Hook System Validation**:
- [ ] All hook scripts exist and are executable
- [ ] Hook scripts complete within time limits
- [ ] Hook paths are absolute
- [ ] Error handling is robust
- [ ] No sensitive data in hook scripts

**MCP Server Validation**:
- [ ] All required packages installed: `npx -y @anthropic-ai/mcp-server-github`
- [ ] Environment variables set correctly
- [ ] OAuth scopes follow principle of least privilege
- [ ] Filesystem paths are whitelisted (no wildcards)
- [ ] Server connectivity test passes: `/mcp`

**Security Validation**:
- [ ] No secrets in version control
- [ ] File permissions are appropriate (600 for sensitive files)
- [ ] Environment variables are not logged
- [ ] Rate limiting configured where applicable
- [ ] Audit trail is enabled

### Validation Script

```bash
#!/bin/bash
# ~/.claude/validate-config.sh

echo "🔍 Validating Claude Code Configuration..."
echo "=========================================="

# Check settings.json syntax
if jq empty .claude/settings.json 2>/dev/null; then
  echo "✅ settings.json syntax is valid"
else
  echo "❌ settings.json has syntax errors"
  exit 1
fi

# Check for hardcoded secrets
if rg "sk-ant-|ghp_|gho_|ghu_|API_KEY" .claude/settings.json; then
  echo "❌ Hardcoded secrets found in settings.json"
  exit 1
else
  echo "✅ No hardcoded secrets found"
fi

# Check hook scripts
HOOK_SCRIPTS=$(jq -r '.hooks | to_entries[] | .value[] | .hooks[] | .command' .claude/settings.json 2>/dev/null)
for script in $HOOK_SCRIPTS; do
  if [ -f "$script" ] && [ -x "$script" ]; then
    echo "✅ Hook script exists and executable: $script"
  else
    echo "⚠️  Hook script missing or not executable: $script"
  fi
done

# Check MCP dependencies
if command -v npx >/dev/null 2>&1; then
  echo "✅ npx is available for MCP servers"
else
  echo "❌ npx is not available - MCP servers won't work"
fi

echo "=========================================="
echo "Validation complete!"
```

---

## Troubleshooting Common Issues

### Permission Problems

| Issue | Cause | Solution |
|-------|-------|----------|
| Tool blocked unexpectedly | Too restrictive deniedTools | Review and refine permission patterns |
| Sensitive files accessible | Missing access controls | Add to deniedTools with specific paths |
| Commands require confirmation | permissionMode set to "ask" | Change to "allow" for trusted environments |

### Hook System Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| Hooks not executing | Invalid JSON syntax | Validate with `jq .` |
| Slow performance | Hook scripts taking too long | Optimize scripts, add background execution |
| Permission errors | Hook scripts not executable | Run `chmod +x` on hook scripts |

### MCP Server Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| Server not connecting | Invalid JSON configuration | Validate mcpServers section with jq |
| OAuth authentication failed | Invalid credentials or scopes | Check environment variables and OAuth app settings |
| Filesystem access denied | Paths not whitelisted | Add specific paths to filesystem MCP args |

---

**End of Skill** | Consolidated from moai-cc-settings + moai-cc-hooks + moai-cc-mcp-plugins + moai-cc-commands
