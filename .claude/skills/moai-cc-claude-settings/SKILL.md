---
name: moai-cc-claude-settings
description: Claude Code user settings management bridge - handles Claude Code specific configurations while integrating with enterprise configuration system
version: 1.0.0
modularized: false
last_updated: 2025-11-22
compliance_score: 71
auto_trigger_keywords:
  - cc
  - claude
  - settings
category_tier: 1
---

# Claude Code Settings Management Bridge

## Purpose

Bridges the gap between the deleted `moai-cc-settings` and the new `moai-cc-configuration` by providing Claude Code specific settings management.

## When to Use

**Automatic triggers**:
- Managing Claude Code user preferences
- Customizing Claude Code behavior
- Configuring permissions and tools
- Setting up user profiles

**Manual invocation**:
- Troubleshooting Claude Code settings issues
- Optimizing development experience
- Managing local vs global settings

## Key Features

### 1. Claude Code Settings Management
- `.claude/settings.json` management
- `.claude/settings.local.json` management
- Permission configuration
- Tool allowlist/denylist

### 2. User Experience Optimization
- Spinner tips configuration
- Hook management
- Permission mode settings (ask/allow/deny)
- Custom keyboard shortcuts

### 3. Profile Management
- User profile creation
- Profile switching
- Settings backup and restore

## Integration with moai-cc-configuration

This skill acts as a bridge:
- Uses `moai-cc-configuration` for enterprise features
- Provides Claude Code specific functionality
- Maintains backward compatibility with old workflows

## Settings Template

Located at: `templates/settings-template.json`

```json
{
  "permissions": {
    "allowedTools": ["Read", "Edit", "Bash", "Grep"],
    "deniedTools": ["rm -rf", "sudo"]
  },
  "permissionMode": "ask",
  "spinnerTipsEnabled": true,
  "disableAllHooks": false
}
```

## Quick Commands

```bash
# View current settings
cat .claude/settings.json

# Update local settings
cat > .claude/settings.local.json << 'EOF'
{
  "permissions": {
    "allow": ["Bash(git:*)", "Bash(npm:*)"]
  }
}
EOF

# Validate settings
node -e "console.log(JSON.parse(require('fs').readFileSync('.claude/settings.json')))"
```

## Best Practices

1. **Always backup before changes**: `cp .claude/settings.json .claude/settings.json.backup`
2. **Use local settings for personal preferences**: Keep in `.claude/settings.local.json`
3. **Never commit sensitive data**: Add to `.gitignore` if needed
4. **Test permission changes carefully**: Can affect tool functionality

## Migration from moai-cc-settings

If you were using `moai-cc-settings`, this bridge skill provides:
- ✅ All user-facing features
- ✅ Settings file management
- ✅ Template support
- ✅ Profile management
- ✅ UX optimization

For enterprise configuration needs, use `moai-cc-configuration`.

## Related Skills
- `moai-cc-configuration` - Enterprise configuration management
- `moai-core-config-schema` - Configuration schema validation
- `moai-cc-permission-mode` - Permission system management

---

*Bridge skill created to restore Claude Code settings functionality after moai-cc-settings deletion*