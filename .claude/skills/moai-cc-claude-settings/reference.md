## Claude Code Settings Reference

### Core Settings Structure

```json
{
  "permissions": {
    "allowedTools": [],    // Allowed tool patterns
    "deniedTools": [],     // Denied tool patterns
    "allow": [],          // Shorthand allowed list
    "deny": []            // Shorthand denied list
  },
  "permissionMode": "ask", // ask | allow | deny
  "spinnerTipsEnabled": true,
  "disableAllHooks": false,
  "env": {}               // Environment variables
}
```

### Permission Patterns

```javascript
// Tool permission examples
"Read(**/*.js)"         // Read all JS files
"Edit(src/**/*)"        // Edit files in src
"Bash(git:*)"          // All git commands
"Bash(npm run:*)"      // All npm run scripts
```

### Settings Priority

1. `.claude/settings.local.json` (highest)
2. `.claude/settings.json`
3. Default settings (lowest)

### Common Operations

```bash
# Merge settings
jq -s '.[0] * .[1]' settings.json settings.local.json

# Validate JSON
python -m json.tool < settings.json

# Pretty print
jq '.' settings.json
```

### Migration Helper

From `moai-cc-settings` to this bridge:
1. Copy old settings files
2. Update permission patterns if needed
3. Test with limited permissions first
4. Gradually expand permissions

### Troubleshooting

| Issue | Solution |
|-------|----------|
| Settings not loading | Check JSON syntax |
| Permissions denied | Review permission patterns |
| Hooks not working | Check `disableAllHooks` |
| Environment vars missing | Add to `env` section |