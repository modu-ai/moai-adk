# Claude Code Plugin Discovery and Installation - Official Documentation Reference

Source: https://code.claude.com/docs/en/discover-plugins
Related: https://code.claude.com/docs/en/plugins
Updated: 2026-01-06

## Overview

Plugin marketplaces are catalogs of pre-built plugins that extend Claude Code with custom commands, agents, hooks, and MCP servers. Discovery and installation is a two-step process:

Step 1: Add the marketplace to register the catalog with Claude Code for browsing available plugins

Step 2: Install individual plugins by selecting specific plugins you want to use

Think of it like adding an app store: adding the store gives you access to browse its collection, but you still choose which apps to download individually.

## Official Anthropic Marketplace

The official marketplace (claude-plugins-official) is automatically available when you start Claude Code. No manual registration required.

### Installation Command

```
/plugin install plugin-name@claude-plugins-official
```

### Code Intelligence Plugins (LSP-based)

These plugins provide deep codebase understanding with jump-to-definition, find references, and type error detection. Each requires the corresponding language server binary to be installed.

C/C++ Plugin:
- Plugin name: clangd-lsp
- Required binary: clangd

C# Plugin:
- Plugin name: csharp-lsp
- Required binary: csharp-ls

Go Plugin:
- Plugin name: gopls-lsp
- Required binary: gopls

Java Plugin:
- Plugin name: jdtls-lsp
- Required binary: jdtls

Lua Plugin:
- Plugin name: lua-lsp
- Required binary: lua-language-server

PHP Plugin:
- Plugin name: php-lsp
- Required binary: intelephense

Python Plugin:
- Plugin name: pyright-lsp
- Required binary: pyright-langserver

Rust Plugin:
- Plugin name: rust-analyzer-lsp
- Required binary: rust-analyzer

Swift Plugin:
- Plugin name: swift-lsp
- Required binary: sourcekit-lsp

TypeScript Plugin:
- Plugin name: typescript-lsp
- Required binary: typescript-language-server

### External Integrations (MCP Servers)

Pre-configured MCP servers for connecting to external services:

Source Control:
- github: GitHub integration
- gitlab: GitLab integration

Project Management:
- atlassian: Jira and Confluence integration
- asana: Asana project management
- linear: Linear issue tracking
- notion: Notion workspace integration

Design Tools:
- figma: Figma design platform integration

Infrastructure:
- vercel: Vercel deployment platform
- firebase: Google Firebase services
- supabase: Supabase backend services

Communication:
- slack: Slack messaging integration

Monitoring:
- sentry: Sentry error monitoring

### Development Workflow Plugins

- commit-commands: Git commit workflows including commit, push, and PR creation
- pr-review-toolkit: Specialized pull request review agents
- agent-sdk-dev: Tools for Claude Agent SDK development
- plugin-dev: Toolkit for creating plugins

### Output Style Plugins

- explanatory-output-style: Educational insights about implementation choices
- learning-output-style: Interactive learning mode for skill building

## Adding Marketplaces

### From GitHub

Basic format using owner/repo notation:

```
/plugin marketplace add owner/repo
```

Example:
```
/plugin marketplace add anthropics/claude-code
```

### From Other Git Hosts

HTTPS URL format:
```
/plugin marketplace add https://gitlab.com/company/plugins.git
```

SSH URL format:
```
/plugin marketplace add [email protected]:company/plugins.git
```

With specific branch or tag:
```
/plugin marketplace add https://gitlab.com/company/plugins.git#v1.0.0
```

### From Local Paths

From local directory:
```
/plugin marketplace add ./my-marketplace
```

From direct marketplace.json file path:
```
/plugin marketplace add ./path/to/marketplace.json
```

### From Remote URL

Direct URL to marketplace.json:
```
/plugin marketplace add https://example.com/marketplace.json
```

## Installing Plugins

### Command Line Installation

Default installation to user scope:
```
/plugin install plugin-name@marketplace-name
```

Installation with specific scope:
```
claude plugin install formatter@your-org --scope project
```

### Installation Scopes

User Scope (default):
- Install for yourself across all projects
- Files stored in user configuration

Project Scope:
- Install for all collaborators on the repository
- Configuration added to .claude/settings.json
- Shared via version control

Local Scope:
- Install for yourself in this repository only
- Not shared with collaborators

Managed Scope:
- Enterprise admin-installed plugins
- Read-only for users

### Interactive Installation

Open plugin manager:
```
/plugin
```

Navigate to Discover tab, press Enter on a plugin to see scope options.

## Managing Installed Plugins

### Disable Without Uninstalling

```
/plugin disable plugin-name@marketplace-name
```

### Re-enable a Disabled Plugin

```
/plugin enable plugin-name@marketplace-name
```

### Uninstall a Plugin

```
/plugin uninstall plugin-name@marketplace-name
```

### Target Specific Scope

```
claude plugin uninstall formatter@your-org --scope project
```

## Managing Marketplaces

### List All Marketplaces

```
/plugin marketplace list
```

### Refresh Plugin Listings

```
/plugin marketplace update marketplace-name
```

### Remove a Marketplace

```
/plugin marketplace remove marketplace-name
```

### Command Shortcuts

- Use /plugin market instead of /plugin marketplace
- Use rm instead of remove

### Auto-Update Configuration

Enable or disable auto-updates via interactive manager:

1. Run /plugin
2. Select Marketplaces tab
3. Choose a marketplace
4. Select Enable auto-update or Disable auto-update

Default behavior:
- Official Anthropic marketplaces: auto-update enabled
- Third-party and local marketplaces: auto-update disabled

Disable all auto-updates globally:
```
export DISABLE_AUTOUPDATER=true
```

## Interactive Plugin Manager

The /plugin command opens a tabbed interface. Use Tab to cycle forward and Shift+Tab to cycle backward.

Discover Tab:
- Browse available plugins from all added marketplaces
- View plugin descriptions and details
- Install plugins with scope selection

Installed Tab:
- View installed plugins grouped by scope
- Enable, disable, or uninstall plugins
- Check plugin status

Marketplaces Tab:
- View all registered marketplaces
- Add new marketplaces
- Update or remove existing marketplaces
- Configure auto-update settings

Errors Tab:
- View plugin loading errors
- Diagnose installation issues
- Check for missing dependencies

## Team Configuration

Team admins can configure automatic marketplace and plugin installation via .claude/settings.json:

```json
{
  "extraKnownMarketplaces": [
    {
      "source": "https://github.com/company/plugins",
      "name": "company-plugins"
    }
  ],
  "enabledPlugins": [
    {
      "name": "plugin-name",
      "marketplaceId": "marketplace-name",
      "scope": "project"
    }
  ]
}
```

When team members trust the repository, Claude Code prompts them to install configured marketplaces and plugins automatically.

## Troubleshooting

/plugin Command Not Recognized:
- Check version with: claude --version (requires 1.0.33+)
- Update via: brew upgrade claude-code or npm update -g @anthropic-ai/claude-code
- Restart terminal after updating

Marketplace Not Loading:
- Verify URL is accessible and .claude-plugin/marketplace.json exists at repository root

Plugin Installation Failures:
- Check plugin source URLs are accessible
- Verify repositories are public or you have access credentials

Files Not Found After Installation:
- Plugins are copied to cache; paths outside plugin directory will not work

Executable Not Found in PATH:
- Install required language server binary from Code Intelligence section

Skills Not Appearing:
- Clear cache: rm -rf ~/.claude/plugins/cache
- Restart Claude Code and reinstall affected plugins

## Quick Start Example

```
/plugin marketplace add anthropics/claude-code
/plugin
/plugin install commit-commands@anthropics-claude-code
/commit-commands:commit
```

## Best Practices

For Individual Users:
- Start with official marketplace plugins matching your development stack
- Install LSP plugins for primary languages to enable code intelligence
- Use project scope for plugins shared with team members
- Keep plugins updated by periodically running marketplace updates

For Teams:
- Configure extraKnownMarketplaces in project settings for team-wide access
- Use enabledPlugins to auto-install required plugins for new members
- Document plugin requirements in project README
- Consider custom marketplace for organization-specific plugins

For Plugin Developers:
- Test plugins locally before publishing to marketplace
- Use plugin-dev from official marketplace for development tooling
- Follow plugin structure conventions from plugins reference
- Version plugins semantically and maintain changelogs

## Related Documentation

- Plugin Development: https://code.claude.com/docs/en/plugins
- Plugin Marketplaces: https://code.claude.com/docs/en/plugin-marketplaces
- Plugins Reference: https://code.claude.com/docs/en/plugins-reference
