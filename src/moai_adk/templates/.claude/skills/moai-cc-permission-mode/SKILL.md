---
name: moai-cc-permission-mode
version: 4.0.0
updated: 2025-11-19
status: stable
stability: stable
description: Claude Code permission configuration and policy management strategies for enterprise security. Covers permission modes, tool access control, whitelist/blacklist patterns, and enterprise deployment best practices.
focus_areas:
  - Permission configuration patterns
  - Tool access control
  - Security policy enforcement
  - Enterprise deployment
keywords:
  - permissions
  - tool-control
  - security
  - sandbox
  - enterprise
  - access-control
allowed-tools: Read, Glob, Bash
---

# Claude Code Permission Modes

## Overview

Claude Code permission management provides fine-grained control over tool access, command execution, and system operations through comprehensive permission policies. Enables secure, controlled automation in enterprise environments.

## Permission Architecture

### Two-Layer Permission System

```
Layer 1: allowedTools (Whitelist)
  ├─ Read patterns: Read(**/*.{js,ts})
  ├─ Edit patterns: Edit(src/**)
  └─ Bash patterns: Bash(git:*), Bash(npm:*)

Layer 2: deniedTools (Blacklist Override)
  ├─ Secrets files: .env*, .aws/**, .vercel/**
  ├─ Destructive: rm -rf:*, sudo:*
  └─ Dangerous: chmod 777:*
```

## Permission Modes

### Whitelist Approach (Recommended)

Define explicitly allowed tool patterns:

```json
{
  "permissions": {
    "allowedTools": [
      "Read(**/*.{js,ts,json,md})",
      "Edit(**/*.{js,ts})",
      "Bash(git:*)",
      "Bash(npm:*)",
      "Bash(uv:*)",
      "Bash(pytest:*)",
      "Bash(mypy:*)"
    ]
  }
}
```

**Benefits**:
- Secure by default (deny unless explicitly allowed)
- Clear audit trail
- Easy to review
- Enterprise-ready

### Blacklist Approach (Not Recommended)

Explicitly block dangerous operations:

```json
{
  "permissions": {
    "deniedTools": [
      "Edit(/config/secrets.json)",
      "Edit(.env*)",
      "Edit(.aws/**)",
      "Bash(rm -rf:*)",
      "Bash(sudo:*)",
      "Bash(chmod 777:*)"
    ]
  }
}
```

**Issues**:
- Requires knowing all dangerous patterns
- New vulnerabilities not covered
- Hard to maintain
- Not enterprise-recommended

## Tool Pattern Syntax

### Read Operations

```
Read(glob_pattern)

Examples:
- Read(**/*.{js,ts})          # All JS/TS files recursively
- Read(.claude/**)            # All Claude files
- Read(docs/api/**/*.md)      # API documentation
- Read(config/production.json) # Specific file
```

### Edit Operations

```
Edit(glob_pattern)

Examples:
- Edit(src/**/*.py)           # Source code only
- Edit(src/services/*.ts)     # Specific directory
```

**Never Edit**:
- `.env*` files
- `.aws/` credentials
- `.vercel/` project config
- `secrets.json`, `credentials.json`

### Bash Commands

```
Bash(command:*)

Examples:
- Bash(git:*)                 # All git operations
- Bash(npm:*)                 # NPM package management
- Bash(uv:*)                  # UV (Python) operations
- Bash(pytest:*)              # Testing
- Bash(mypy:*)                # Type checking
- Bash(ruff:*)                # Python linting

Dangerous (Block):
- Bash(rm -rf:*)              # Recursive delete
- Bash(sudo:*)                # Superuser access
- Bash(chmod 777:*)           # Permission changes
- Bash(find.*-delete:*)       # File deletion
```

## Security Patterns

### Production-Grade Configuration

```json
{
  "permissions": {
    "allowedTools": [
      "Read(**)",
      "Edit(src/**)",
      "Edit(tests/**)",
      "Bash(git:*)",
      "Bash(uv:*)",
      "Bash(pytest:*)",
      "Bash(mypy:*)",
      "Bash(ruff:*)"
    ],
    "deniedTools": [
      "Edit(.*)",
      "Edit(.env*)",
      "Edit(.aws/**)",
      "Edit(.vercel/**)",
      "Bash(rm:*)",
      "Bash(sudo:*)",
      "Bash(chmod:*)"
    ]
  },
  "sandbox": {
    "allowUnsandboxedCommands": false
  }
}
```

### Sensitive Data Protection

Critical patterns to block:

```json
{
  "deniedTools": [
    "Edit(.env*)",
    "Edit(.env.local)",
    "Edit(.env.production)",
    "Edit(.aws/**)",
    "Edit(.aws/credentials)",
    "Edit(.aws/config)",
    "Edit(.vercel/**)",
    "Edit(config/**/secrets.json)",
    "Edit(**/*credentials*.json)",
    "Edit(**/*password*.json)",
    "Edit(**/*token*.json)"
  ]
}
```

## Permission Validation

Always validate permission configurations:

```bash
# Check current settings
cat .claude/settings.json | jq '.permissions'

# Verify allowedTools patterns
cat .claude/settings.json | jq '.permissions.allowedTools[]'

# Verify deniedTools patterns
cat .claude/settings.json | jq '.permissions.deniedTools[]'

# Test specific operation
# Try Read(test_file.md) → allowed?
# Try Edit(.env) → denied?
# Try Bash(git status) → allowed?
```

## Best Practices

### Security-First Approach

- ✅ Use whitelist (allowedTools) instead of blacklist
- ✅ Protect secrets files (.env*, .aws/, .vercel/)
- ✅ Block destructive commands (rm -rf, sudo, chmod)
- ✅ Enable sandbox mode
- ✅ Regularly review permissions
- ✅ Audit permission violations
- ✅ Document permission decisions

### Team Collaboration

- ✅ Document permission changes in commit
- ✅ Explain security rationale
- ✅ Test with different roles
- ✅ Keep audit log of changes

## Common Patterns by Use Case

### Development Environment

```json
{
  "allowedTools": [
    "Read(**)",
    "Edit(src/**)",
    "Edit(tests/**)",
    "Edit(.claude/**)",
    "Bash(git:*)",
    "Bash(uv:*)",
    "Bash(pytest:*)",
    "Bash(mypy:*)",
    "Bash(ruff:*)"
  ],
  "deniedTools": [
    "Edit(.env*)",
    "Edit(.aws/**)",
    "Bash(rm -rf:*)",
    "Bash(sudo:*)"
  ]
}
```

### CI/CD Pipeline

```json
{
  "allowedTools": [
    "Read(src/**)",
    "Read(tests/**)",
    "Read(.github/workflows/**)",
    "Bash(git:*)",
    "Bash(uv:*)",
    "Bash(pytest:*)",
    "Bash(mypy:*)"
  ],
  "deniedTools": [
    "Edit(**)",
    "Bash(sudo:*)",
    "Bash(rm:*)"
  ]
}
```

### Read-Only Analysis

```json
{
  "allowedTools": [
    "Read(**)",
    "Bash(git:log)",
    "Bash(git:show)"
  ],
  "deniedTools": [
    "Edit(**)",
    "Bash(git:push)",
    "Bash(git:pull)"
  ]
}
```

## TRUST 5 Compliance

- **Test-First**: Permission patterns tested with actual Claude Code usage scenarios
- **Readable**: Clear permission naming and organization, documented rationale
- **Unified**: Consistent permission approach across all environments
- **Secured**: Whitelist-based, blocking all dangerous operations
- **Trackable**: Audit trail of permission modifications and changes

## Related Skills

- `moai-cc-hooks` - Hook execution for pre/post-tool validation
- `moai-core-env-security` - Environment variable security
- `moai-cc-sandbox-isolation` - Sandbox mode configuration

---

**Last Updated**: 2025-11-19
**Version**: 4.0.0
**Enterprise Production Ready**: Yes ✅
**Maturity**: Stable
