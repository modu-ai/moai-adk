---
name: moai-cc-permission-mode
description: Claude Code permission system, IAM patterns, and security policy automation
allowed-tools: [Read, Bash]
---

# Claude Code Permission Mode

## Quick Reference

Claude Code's permission system implements a **zero-trust, ask-before-act** security model with three policy levels: **Allow** (automatic approval), **Ask** (manual review), and **Deny** (hard block). This skill covers permission modes, IAM patterns, wildcard matching, file/command protection, and enterprise security policies for production Claude Code deployments.

**Permission Architecture** (November 2025):
- **Default Mode**: Read-only with explicit approvals for mutations
- **Policy Scopes**: User-level (`~/.claude/settings.json`) + Project-level (`.claude/settings.json`)
- **Rule Types**: `allowedTools` (whitelist), `deniedTools` (blacklist), `permissionMode` (session behavior)
- **Tool Categories**: Bash commands, Read/Write/Edit operations, WebFetch URLs, MCP servers
- **Pattern Matching**: Exact matches, wildcards (`*`), glob patterns (`**/*.ts`)

**Security Principle**: Prefer allowlist over blocklist. Only permit routine, low-risk actions automatically; keep everything else on **Ask** mode for manual review.

---

## Implementation Guide

### Phase 1: Permission Policy Configuration

**Basic Permission Structure** (`~/.claude/settings.json` or `.claude/settings.json`):
```json
{
  "permissions": {
    "allowedTools": [
      "Read(**/*.{js,ts,json,md})",
      "Edit(**/*.{js,ts})",
      "Bash(git:*)",
      "Bash(npm:test)",
      "Bash(uv:*)",
      "WebFetch(https://docs.example.com/**)"
    ],
    "deniedTools": [
      "Edit(/config/secrets.json)",
      "Edit(/.env*)",
      "Bash(rm -rf:*)",
      "Bash(sudo:*)",
      "Bash(chmod 777:*)",
      "Write(/build/**)",
      "Read(/secrets/**)"
    ]
  },
  "permissionMode": "acceptEdits"
}
```

**Permission Modes**:
- `"permissionMode": "ask"` - Require manual approval for all edits (safest)
- `"permissionMode": "acceptEdits"` - Auto-approve allowed edits, ask for others
- `"permissionMode": "acceptAll"` - Auto-approve everything (⚠️ use only in trusted environments)

### Phase 2: Wildcard Pattern Matching

**Glob Pattern Examples**:
```json
{
  "allowedTools": [
    // Exact match
    "Read(/src/index.ts)",
    
    // Single-level wildcard
    "Read(/src/*.ts)",
    
    // Multi-level wildcard
    "Read(/src/**/*.ts)",
    
    // File extension groups
    "Edit(**/*.{js,ts,jsx,tsx})",
    
    // Command prefix matching
    "Bash(npm:*)",          // npm install, npm test, npm run build
    "Bash(git:status)",     // Only git status (exact)
    "Bash(git:add *)",      // git add with any arguments
    
    // URL pattern matching
    "WebFetch(https://api.example.com/**)"
  ]
}
```

**Pattern Priority** (first match wins):
```json
{
  "deniedTools": [
    "Edit(/.env)"           // Block .env (highest priority)
  ],
  "allowedTools": [
    "Edit(/.env.example)"   // Allow .env.example (ignored if .env denied)
  ]
}
```

**Recommendation**: Place specific denials before broad allows to prevent accidental overrides.

### Phase 3: Security Policy Templates

**Development Environment** (Permissive for fast iteration):
```json
{
  "permissions": {
    "allowedTools": [
      "Read(**/*)",
      "Edit(**/*.{js,ts,jsx,tsx,css,html})",
      "Bash(npm:*)",
      "Bash(git:*)",
      "Bash(node:*)",
      "Bash(uv:*)"
    ],
    "deniedTools": [
      "Edit(/.env*)",
      "Bash(rm -rf:*)",
      "Bash(sudo:*)"
    ]
  },
  "permissionMode": "acceptEdits"
}
```

**Production Environment** (Strict for safety):
```json
{
  "permissions": {
    "allowedTools": [
      "Read(/src/**/*.{js,ts})",
      "Read(/docs/**/*.md)",
      "Bash(git:status)",
      "Bash(git:log)"
    ],
    "deniedTools": [
      "Edit(**/*)",           // Block all edits
      "Bash(rm:*)",
      "Bash(sudo:*)",
      "Bash(chmod:*)",
      "Write(**/*)"
    ]
  },
  "permissionMode": "ask"
}
```

**Regulated/Compliance Environment** (Audit trail required):
```json
{
  "permissions": {
    "allowedTools": [
      "Read(/approved-docs/**/*.md)"
    ],
    "deniedTools": [
      "Edit(**/*)",
      "Bash(**)",
      "Write(**/*)",
      "WebFetch(**)"
    ]
  },
  "permissionMode": "ask",
  "hooks": {
    "PreToolUse": [
      {
        "hooks": [{
          "type": "command",
          "command": "uv run .claude/hooks/audit_log.py"
        }]
      }
    ]
  }
}
```

---

## Advanced Patterns

### Multi-Layer IAM Strategy

**Layered Permissions** (User-level + Project-level):
```
~/.claude/settings.json (User-level - global defaults)
  ├─ Broad organizational policies
  ├─ Universal denials (sudo, rm -rf)
  └─ Common tools (git, npm)

.claude/settings.json (Project-level - overrides)
  ├─ Project-specific allows
  ├─ Tech stack permissions (uv, pytest, docker)
  └─ Custom hooks and validations
```

**Precedence Rule**: Project-level settings override user-level settings for the same tool.

**Example User-Level** (`~/.claude/settings.json`):
```json
{
  "permissions": {
    "allowedTools": [
      "Bash(git:*)",
      "Read(**/*.{md,txt})"
    ],
    "deniedTools": [
      "Bash(sudo:*)",
      "Bash(rm -rf:*)"
    ]
  },
  "permissionMode": "ask"
}
```

**Example Project-Level** (`.claude/settings.json`):
```json
{
  "permissions": {
    "allowedTools": [
      "Read(/src/**/*.py)",
      "Edit(/src/**/*.py)",
      "Bash(uv:*)",
      "Bash(pytest:*)"
    ],
    "deniedTools": [
      "Edit(/src/core/security/*.py)"  // Block critical security modules
    ]
  },
  "permissionMode": "acceptEdits"
}
```

### Dynamic Permission Validation (Hook-Based)

**Advanced File Protection** (Context-aware validation):
```python
#!/usr/bin/env python3
import json
import sys
from pathlib import Path

def is_sensitive_file(file_path: str) -> bool:
    """Detect sensitive files beyond static patterns."""
    sensitive_patterns = [
        '.env', 'secrets.', 'credentials.', 
        'private_key', 'id_rsa', '.pem'
    ]
    
    path = Path(file_path)
    
    # Check filename
    if any(pattern in path.name.lower() for pattern in sensitive_patterns):
        return True
    
    # Check parent directories
    if any(part in ['.git', '.ssh', 'secrets'] for part in path.parts):
        return True
    
    # Check file content (for small files)
    if path.exists() and path.stat().st_size < 10000:
        try:
            content = path.read_text()
            if 'BEGIN PRIVATE KEY' in content or 'API_KEY' in content:
                return True
        except:
            pass
    
    return False

input_data = json.loads(sys.stdin.read())
tool_name = input_data.get('tool_name', '')
tool_input = input_data.get('tool_input', {})

if tool_name in ['Read', 'Edit', 'Write']:
    file_path = tool_input.get('file_path', '')
    if is_sensitive_file(file_path):
        print(f"BLOCKED: Sensitive file detected - {file_path}", file=sys.stderr)
        sys.exit(2)  # Block operation

sys.exit(0)  # Allow
```

**Register Hook**:
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Read|Edit|Write",
        "hooks": [{
          "type": "command",
          "command": "python3 .claude/hooks/file_protection.py"
        }]
      }
    ]
  }
}
```

### Sandbox Mode for Bash Commands

**Enable Sandbox** (Linux/Mac only):
```json
{
  "sandbox": {
    "allowUnsandboxedCommands": false
  },
  "permissions": {
    "allowedTools": [
      "Bash(ls:*)",
      "Bash(cat:*)",
      "Bash(grep:*)"
    ]
  }
}
```

**Security Impact**:
- ✅ Bash commands run in isolated environment
- ✅ Prevents system-level modifications
- ✅ Limits network access
- ⚠️ Some tools may not work (docker, kubernetes)

---

## Permission Policy Best Practices

### ✅ DO
- Use **allowlist philosophy**: Block by default, allow specific actions
- Prefer **per-file approvals** tied to specific tasks (not broad scopes)
- Configure **project-level permissions** for tech stack specificity
- Implement **hook-based validation** for complex security logic
- Enable **audit logging** for compliance requirements
- Use **sandbox mode** for untrusted environments
- Review **permission UI** regularly (`/permissions` command)
- Document **permission rationale** in `.claude/PERMISSIONS.md`

### ❌ DON'T
- Use `"permissionMode": "acceptAll"` in production
- Allow broad wildcards (`Bash(**)`) without validation
- Forget to deny dangerous patterns (`rm -rf`, `sudo rm`)
- Trust deprecated permission features (known bugs in Read/Write deny)
- Skip project-level overrides (one-size-fits-all policies fail)
- Ignore permission request logs (security indicators)
- Allow WebFetch to arbitrary URLs (SSRF risk)

---

## Permission UI & Debugging

**View Active Permissions**:
```bash
/permissions
```

**Output Example**:
```
Active Permissions:
┌─────────────────────────────────────────────┐
│ Source: Project (.claude/settings.json)    │
├─────────────────────────────────────────────┤
│ ✅ Read(/src/**/*.py)                       │
│ ✅ Edit(/src/**/*.py)                       │
│ ✅ Bash(uv:*)                               │
│ ❌ Edit(/src/core/security/*.py)            │
├─────────────────────────────────────────────┤
│ Source: User (~/.claude/settings.json)     │
├─────────────────────────────────────────────┤
│ ✅ Bash(git:*)                              │
│ ❌ Bash(sudo:*)                             │
│ ❌ Bash(rm -rf:*)                           │
└─────────────────────────────────────────────┘
```

**Debug Permission Denials**:
```bash
# Check stderr output for blocked operations
tail -f ~/.claude/logs/tool_execution.log
```

**Test Permission Rules** (Dry-run):
```python
#!/usr/bin/env python3
# Test if a tool call would be allowed
import json

test_tool = {
    "tool_name": "Edit",
    "tool_input": {"file_path": "/src/config.py"}
}

# Simulate permission check
# (Run through your hook logic)
```

---

## Enterprise Integration

**CI/CD Permission Template**:
```yaml
# .claude/ci-permissions.json
{
  "permissions": {
    "allowedTools": [
      "Read(**/*)",
      "Bash(pytest:*)",
      "Bash(git:status)",
      "Bash(git:log)"
    ],
    "deniedTools": [
      "Edit(**/*)",
      "Write(**/*)",
      "Bash(rm:*)",
      "WebFetch(**)"
    ]
  },
  "permissionMode": "ask",
  "hooks": {
    "PreToolUse": [{
      "hooks": [{
        "type": "command",
        "command": "python3 .claude/hooks/ci_audit.py"
      }]
    }]
  }
}
```

**Load in CI**:
```bash
export CLAUDE_SETTINGS_PATH=./.claude/ci-permissions.json
claude --settings $CLAUDE_SETTINGS_PATH
```

---

## Known Issues & Workarounds

**Issue #6631**: Permission deny for Read/Write tools not enforced (reported vulnerability)
- **Workaround**: Use `PreToolUse` hooks for reliable blocking
- **Status**: Fix planned for future release

**Issue #2720**: Local settings bypass in certain configurations
- **Workaround**: Validate permissions in hooks
- **Status**: Under investigation

---

## Works Well With

- `moai-cc-hook-model-strategy` (Hook-based permission validation)
- `moai-cc-subagent-lifecycle` (Subagent permission inheritance)
- `moai-security-api` (Security policy patterns)
- `moai-core-env-security` (Environment variable protection)

---

**Version**: 1.0.0  
**Last Updated**: 2025-11-21  
**Status**: Production Ready (Enterprise)  
**Official Reference**: https://code.claude.com/docs/en/settings  
**Security Guide**: https://www.backslash.security/blog/claude-code-security-best-practices
