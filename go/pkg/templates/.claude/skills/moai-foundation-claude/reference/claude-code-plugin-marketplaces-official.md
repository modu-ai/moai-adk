# Claude Code Plugin Marketplaces - Official Documentation Reference

Source: https://code.claude.com/docs/en/plugin-marketplaces
Related: https://code.claude.com/docs/en/plugins-reference
Updated: 2026-01-06

## What are Plugin Marketplaces?

A plugin marketplace is a catalog that distributes Claude Code extensions across teams and communities. It provides centralized discovery, version tracking, automatic updates, and supports multiple source types including git repositories and local paths.

Key Benefits:
- Centralized plugin discovery and management
- Version tracking and automatic updates
- Support for multiple source types (GitHub, GitLab, local paths)
- Team-wide distribution and enforcement

## Creating a Marketplace

### Directory Structure

```
my-marketplace/
- .claude-plugin/
  - marketplace.json
- plugins/
  - review-plugin/
    - .claude-plugin/plugin.json
    - commands/review.md
```

### Quick Start

Step 1: Create directory structure
```bash
mkdir -p my-marketplace/.claude-plugin
mkdir -p my-marketplace/plugins/review-plugin/.claude-plugin
mkdir -p my-marketplace/plugins/review-plugin/commands
```

Step 2: Create plugin manifest (plugins/review-plugin/.claude-plugin/plugin.json)
```json
{"name": "review-plugin", "description": "Quick code reviews", "version": "1.0.0"}
```

Step 3: Create marketplace manifest (.claude-plugin/marketplace.json)
```json
{
  "name": "my-plugins",
  "owner": {"name": "Your Name"},
  "plugins": [{"name": "review-plugin", "source": "./plugins/review-plugin"}]
}
```

## marketplace.json Schema

### Required Fields

- name: Marketplace identifier (kebab-case)
- owner: Object with name (required) and email (optional)
- plugins: Array of plugin entries

### Optional Metadata

- metadata.description: Brief marketplace description
- metadata.version: Marketplace version
- metadata.pluginRoot: Base directory for relative paths

### Complete Example

```json
{
  "name": "acme-dev-tools",
  "owner": {"name": "ACME DevTools Team", "email": "[email protected]"},
  "metadata": {"description": "ACME engineering tools", "version": "2.0.0", "pluginRoot": "./plugins"},
  "plugins": [
    {"name": "code-formatter", "source": "./formatter"},
    {"name": "security-scanner", "source": {"source": "github", "repo": "acme/security-plugin"}}
  ]
}
```

## Plugin Entry Configuration

### Required Fields

- name: Plugin identifier (kebab-case)
- source: Plugin location (string path or source object)

### Optional Fields

Metadata: description, version, author (object with name/email), homepage, repository, license, keywords, category, tags

Behavior: strict (boolean, default true) - whether plugin needs its own plugin.json

Components: commands, agents, hooks, mcpServers, lspServers - custom path overrides

## Plugin Source Types

### Relative Paths
```json
{"name": "my-plugin", "source": "./plugins/my-plugin"}
```

### GitHub Repositories
```json
{"name": "github-plugin", "source": {"source": "github", "repo": "owner/repo"}}
```

With specific ref:
```json
{"name": "github-plugin", "source": {"source": "github", "repo": "owner/repo", "ref": "v2.0"}}
```

### Git URL Repositories
```json
{"name": "git-plugin", "source": {"source": "url", "url": "https://gitlab.com/team/plugin.git"}}
```

## Hosting and Distribution

### GitHub (Recommended)
1. Create GitHub repository
2. Add .claude-plugin/marketplace.json at root
3. Users add with: /plugin marketplace add owner/repo

### Other Git Services
```bash
/plugin marketplace add https://gitlab.com/company/plugins.git
```

### Local Testing
```bash
/plugin marketplace add ./my-marketplace
/plugin install test-plugin@my-marketplace
/plugin validate .
```

## Team Configuration

### Add Marketplace to Settings (.claude/settings.json)
```json
{
  "extraKnownMarketplaces": {
    "company-tools": {"source": {"source": "github", "repo": "your-org/claude-plugins"}}
  }
}
```

### Auto-Enable Plugins
```json
{
  "enabledPlugins": {
    "code-formatter@company-tools": true,
    "security-scanner@company-tools": true
  }
}
```

## Enterprise Restrictions

### strictKnownMarketplaces Setting

Undefined (default): No restrictions

Empty array (lockdown): No external marketplaces allowed
```json
{"strictKnownMarketplaces": []}
```

Allowlist: Only specified marketplaces permitted
```json
{
  "strictKnownMarketplaces": [
    {"source": "github", "repo": "acme-corp/approved-plugins"},
    {"source": "url", "url": "https://plugins.example.com/marketplace.json"}
  ]
}
```

Note: Set in managed settings only (cannot be overridden by user/project settings)

## Validation and Testing

### Commands
```bash
claude plugin validate .    # CLI
/plugin validate .          # Slash command
```

### Common Errors

- File not found: marketplace.json - Create .claude-plugin/marketplace.json
- Invalid JSON syntax - Check commas, quotes
- Duplicate plugin name - Use unique names
- Path traversal not allowed - Remove ".." from paths

### Non-Blocking Warnings
- No plugins defined
- No marketplace description
- npm sources not fully implemented

## Advanced Plugin Entry

```json
{
  "name": "enterprise-tools",
  "source": {"source": "github", "repo": "company/enterprise-plugin"},
  "description": "Enterprise workflow automation",
  "version": "2.1.0",
  "author": {"name": "Enterprise Team", "email": "[email protected]"},
  "homepage": "https://docs.example.com/plugins/enterprise-tools",
  "license": "MIT",
  "keywords": ["enterprise", "workflow"],
  "category": "productivity",
  "commands": ["./commands/core/", "./commands/enterprise/"],
  "agents": ["./agents/security-reviewer.md"],
  "hooks": {
    "PostToolUse": [{
      "matcher": "Write|Edit",
      "hooks": [{"type": "command", "command": "${CLAUDE_PLUGIN_ROOT}/scripts/validate.sh"}]
    }]
  },
  "mcpServers": {
    "enterprise-db": {
      "command": "${CLAUDE_PLUGIN_ROOT}/servers/db-server",
      "args": ["--config", "${CLAUDE_PLUGIN_ROOT}/config.json"]
    }
  },
  "strict": false
}
```

Notes:
- ${CLAUDE_PLUGIN_ROOT} references plugin installation directory
- strict: false means plugin does not require its own plugin.json

## Reserved Marketplace Names

Cannot be used by third-party marketplaces:
- claude-code-marketplace, claude-code-plugins, claude-plugins-official
- anthropic-marketplace, anthropic-plugins
- agent-skills, life-sciences

Also blocked: Names impersonating official marketplaces

## Troubleshooting

### Marketplace Not Loading
- Verify URL is accessible
- Check .claude-plugin/marketplace.json exists at root
- Validate JSON with /plugin validate
- Confirm access permissions for private repos

### Plugin Installation Failures
- Verify plugin source URLs are accessible
- Check plugin directories contain required files
- For GitHub sources, ensure repos are public or accessible
- Test by cloning repository manually

### Files Not Found After Installation
Cause: Plugins are copied to cache. External paths (../shared) do not work.

Solutions:
- Use symlinks (followed during copying)
- Restructure shared directories inside plugin source
- Include all required files within plugin directory

## Commands Reference

### Marketplace Management
```bash
/plugin marketplace add owner/repo              # Add from GitHub
/plugin marketplace add https://url/repo.git   # Add from URL
/plugin marketplace add ./local-path           # Add local
/plugin marketplace list                        # List marketplaces
/plugin marketplace remove name                 # Remove marketplace
```

### Plugin Installation
```bash
/plugin install plugin-name@marketplace-name   # Install from marketplace
```

## Best Practices

Marketplace Organization:
- Group related plugins together
- Use clear, descriptive names
- Maintain consistent versioning
- Document all plugins

Security:
- Review plugin scripts before distribution
- Avoid hardcoded credentials
- Use environment variables for sensitive data
- Document required permissions

Distribution:
- Test locally before publishing
- Validate structure before sharing
- Provide clear installation instructions

## Additional Resources

- Plugin Creation: https://code.claude.com/docs/en/plugins
- Plugin Reference: https://code.claude.com/docs/en/plugins-reference
- Plugin Settings: https://code.claude.com/docs/en/settings#plugin-settings
- Discover Plugins: https://code.claude.com/docs/en/discover-plugins
