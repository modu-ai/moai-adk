---
name: "moai-plugin-builder"
description: "Claude Code plugin development patterns, templates, and best practices. Use when creating plugins, defining plugin components, or troubleshooting plugin issues."
version: 1.2.0
category: "foundation"
modularized: false
user-invocable: false
tags: ['plugin', 'claude-code', 'development', 'templates', 'hooks', 'commands', 'agents', 'skills', 'mcp', 'lsp', 'marketplace']
updated: 2026-01-08
allowed-tools:
  - Read
  - Write
  - Edit
  - Grep
  - Glob
  - Bash
  - TodoWrite
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
status: "active"
author: "MoAI-ADK Team"
---

# Claude Code Plugin Builder

## Quick Reference (30 seconds)

Plugin Development Essentials - Build Claude Code plugins with correct structure, components, and best practices.

Directory Structure:
- `.claude-plugin/plugin.json` - Plugin manifest (REQUIRED)
- `commands/` - Slash commands at plugin root
- `agents/` - Custom agents at plugin root
- `skills/` - Agent skills at plugin root
- `hooks/` - Event handlers at plugin root
- `.mcp.json` - MCP server configuration
- `.lsp.json` - LSP server configuration

Critical Constraint: Component directories (commands, agents, skills, hooks) MUST be at plugin root level, NOT inside `.claude-plugin/`.

When to Use:
- Creating new Claude Code plugins
- Defining plugin components (commands, agents, skills, hooks)
- Configuring MCP or LSP servers for plugins
- Troubleshooting plugin loading issues
- Migrating standalone configurations to plugin format

---

## Implementation Guide

### Plugin Directory Structure

Correct Plugin Layout:
```
my-plugin/
  .claude-plugin/
    plugin.json          # Required: Plugin metadata only
  commands/              # At root level
    my-command.md
  agents/                # At root level
    my-agent.md
  skills/                # At root level
    my-skill/
      SKILL.md
  hooks/                 # At root level
    hooks.json
  .mcp.json              # At root level
  .lsp.json              # At root level
  LICENSE
  CHANGELOG.md
  README.md
```

Common Mistake to Avoid: Never place component directories inside .claude-plugin folder.

### plugin.json Schema

Minimal Configuration:
```json
{
  "name": "my-plugin"
}
```

Complete Configuration:
```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "Plugin description explaining purpose",
  "author": {
    "name": "Author Name",
    "email": "author@example.com",
    "url": "https://example.com"
  },
  "homepage": "https://github.com/user/my-plugin",
  "repository": "https://github.com/user/my-plugin",
  "license": "MIT",
  "keywords": ["claude-code", "automation"],
  "commands": ["./commands"],
  "agents": ["./agents"],
  "skills": ["./skills"],
  "hooks": ["./hooks/hooks.json"],
  "mcpServers": ["./.mcp.json"],
  "lspServers": ["./.lsp.json"],
  "outputStyles": ["./output-styles"]
}
```

Required Fields:
- name: Kebab-case unique identifier (letters, numbers, hyphens only)

Optional Fields:
- version: Semantic versioning (MAJOR.MINOR.PATCH)
- description: Plugin purpose explanation
- author: Object containing name, email, url
- homepage, repository, license, keywords
- Component path references (commands, agents, skills, hooks)
- Server configurations (mcpServers, lspServers)
- Output style references (outputStyles)

### Path Configuration Rules

Path Format Requirements:
- All paths must be relative to plugin root
- Paths must start with `./`
- Arrays supported for multiple paths
- Default directories are additive (not replaced)

Environment Variables:
- `${CLAUDE_PLUGIN_ROOT}` - Plugin installation directory
- `${CLAUDE_PROJECT_DIR}` - Current project directory

Example Path Usage:
```json
{
  "commands": ["./commands", "./extra-commands"],
  "hooks": ["./hooks/main.json", "./hooks/validation.json"]
}
```

### Slash Commands

Command File Structure (commands/my-command.md):
```markdown
---
description: Command description for discovery
---

Command instructions and prompt content.

Arguments: $ARGUMENTS (all), $1, $2 (positional)
File references: @path/to/file.md
```

Frontmatter Fields:
- description (required): Command purpose for help display

Argument Handling:
- `$ARGUMENTS` - All arguments as single string
- `$1`, `$2`, `$3` - Individual positional arguments
- `@file.md` - File content injection

Command Namespacing: Commands accessed as `/plugin-name:command-name`

### Custom Agents

Agent File Structure (agents/my-agent.md):
```markdown
---
name: my-agent
description: Agent purpose and capabilities
tools: Read, Write, Edit, Grep, Glob, Bash
model: sonnet
permissionMode: default
skills:
  - skill-name-one
  - skill-name-two
---

Agent system prompt and instructions.
```

Frontmatter Fields:
- name (required): Agent identifier
- description: Agent purpose
- tools: Comma-separated tool list
- model: sonnet, opus, haiku, inherit
- permissionMode: default, bypassPermissions, plan, passthrough
- skills: Array of skill names to load

Available Tools:
- Read, Write, Edit - File operations
- Grep, Glob - Search operations
- Bash - Command execution
- WebFetch, WebSearch - Web access
- Task - Sub-agent delegation
- TodoWrite - Task management

### Agent Skills

Skill Structure (skills/my-skill/SKILL.md):
```markdown
---
name: my-skill
description: Skill purpose and when to use
allowed-tools: Read, Grep, Glob
---

# Skill Name

## Quick Reference (30 seconds)

Brief overview and key concepts.

## Implementation Guide

Detailed implementation patterns.

## Advanced Patterns

Expert-level knowledge.
```

Skill Discovery: Model-invoked based on context relevance. Skills load automatically when task context matches skill description.

### Hooks Configuration

hooks.json Structure:
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Write",
        "hooks": [
          {
            "type": "command",
            "command": "bash ./hooks/validate-write.sh"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "prompt",
            "prompt": "Verify operation completed successfully"
          }
        ]
      }
    ]
  }
}
```

Available Hook Events:
- PreToolUse, PostToolUse, PostToolUseFailure - Tool execution lifecycle
- PermissionRequest, UserPromptSubmit, Notification, Stop - User interaction
- SubagentStart, SubagentStop - Sub-agent lifecycle
- SessionStart, SessionEnd, PreCompact - Session lifecycle

Hook Types: command (bash), prompt (LLM), agent (invoke agent)

Matcher Patterns: Exact name ("Write"), wildcard ("*"), tool-specific filtering

### MCP Server Configuration

.mcp.json Structure:
```json
{
  "mcpServers": {
    "my-server": {
      "command": "npx",
      "args": ["-y", "@my-org/mcp-server"],
      "env": {
        "API_KEY": "${API_KEY}"
      }
    }
  }
}
```

Transport Types: stdio (default), http, sse

Fields: command (executable), args (array), env (variables), type (transport), url (for http/sse)

### LSP Server Configuration

.lsp.json Structure:
```json
{
  "lspServers": {
    "python": {
      "command": "pylsp",
      "args": [],
      "extensionToLanguage": {
        ".py": "python",
        ".pyi": "python"
      },
      "env": {
        "PYTHONPATH": "${CLAUDE_PROJECT_DIR}"
      }
    }
  }
}
```

Required Fields:
- command: LSP server executable
- extensionToLanguage: File extension to language mapping

Optional Fields: args, env, transport, initializationOptions, settings, workspaceFolder, startupTimeout, shutdownTimeout, restartOnCrash, maxRestarts, loggingConfig

---

## Advanced Patterns

### Development Workflow

Local Development:
```bash
# Test single plugin
claude --plugin-dir ./my-plugin

# Test multiple plugins
claude --plugin-dir ./plugin-one --plugin-dir ./plugin-two
```

Testing Components:
- Commands: `/plugin-name:command-name` invocation
- Agents: `/agents` to list, then invoke by name
- Skills: Ask questions relevant to skill domain
- Hooks: Trigger events and check debug logs

Debugging:
```bash
# Enable debug mode
claude --debug

# Validate plugin structure
claude plugin validate

# View plugin errors
/plugin errors
```

### Security Best Practices

Path Security:
- Always use `${CLAUDE_PLUGIN_ROOT}` for plugin-relative paths
- Never hardcode absolute paths
- Validate all inputs in hook scripts
- Prevent path traversal attacks

Permission Guidelines:
- Apply least privilege for tool access
- Limit agent permissions to required operations
- Validate hook command inputs
- Sanitize environment variables

### Distribution and Installation

Plugin Installation Scopes:
- user: `~/.claude/settings.json` (personal, default)
- project: `.claude/settings.json` (team, version controlled)
- local: `.claude/settings.local.json` (developer, gitignored)
- managed: `managed-settings.json` (enterprise, read-only)

CLI Commands:
```bash
# Plugin management
claude plugin install <plugin-name>
claude plugin uninstall <plugin-name>
claude plugin list
claude plugin enable <plugin-name>
claude plugin disable <plugin-name>
claude plugin update <plugin-name>

# Marketplace
claude plugin marketplace add <url>
claude plugin marketplace list
```

### Marketplace Creation

marketplace.json Structure:
```json
{
  "name": "my-marketplace",
  "owner": {
    "name": "Organization Name",
    "email": "contact@example.com"
  },
  "metadata": {
    "description": "Custom plugins for our team",
    "version": "1.0.0",
    "pluginRoot": "./plugins"
  },
  "plugins": [
    {
      "name": "my-plugin",
      "source": "./plugins/my-plugin",
      "description": "Plugin description",
      "version": "1.0.0",
      "category": "development",
      "keywords": ["automation", "workflow"]
    }
  ]
}
```

Marketplace Required Fields:
- name: Marketplace identifier in kebab-case
- owner: Object with name (required) and email (optional)
- plugins: Array of plugin entries

Plugin Source Types:
- Relative paths: `"source": "./plugins/my-plugin"`
- GitHub: `{"source": "github", "repo": "owner/repo"}`
- Git URL: `{"source": "url", "url": "https://gitlab.com/org/plugin.git"}`

Reserved Marketplace Names (cannot be used):
- claude-code-marketplace, claude-code-plugins, claude-plugins-official
- anthropic-marketplace, anthropic-plugins
- agent-skills, life-sciences

Marketplace Hosting Options:
- GitHub repository (recommended): Users add via `/plugin marketplace add owner/repo`
- Other Git services: Full URL with `/plugin marketplace add https://...`
- Local testing: `/plugin marketplace add ./path/to/marketplace`

---

## Troubleshooting

Plugin Not Loading: Verify plugin.json exists and has valid syntax; confirm kebab-case name; ensure component directories at root level

Commands Not Found: Check .md extension; verify YAML frontmatter with description; test with `/plugin-name:command-name`

Hooks Not Triggering: Verify hooks.json syntax; check matcher patterns; confirm command executable; enable debug mode

MCP Server Failures: Verify command in PATH; check env variables; confirm transport type; test server independently

---

## Works Well With

- moai-foundation-claude - Claude Code configuration and patterns
- moai-foundation-core - Core development workflows
- moai-workflow-project - Project initialization
- moai-domain-backend - Backend plugin development
- moai-domain-frontend - Frontend plugin development

---

## Reference Files

Extended Documentation:
- [Templates Reference](reference.md) - Complete plugin templates
- [Migration Guide](migration.md) - Converting standalone configs
- [Examples](examples.md) - Working plugin examples
- [Validation](validation.md) - Plugin validation rules

---

Status: Production Ready
Last Updated: 2026-01-06
Maintained by: MoAI-ADK Team
Version Changes: v1.2.0 - Added marketplace creation and hosting section; Added reserved marketplace names; Added plugin source types documentation
v1.1.0 - Added PostToolUseFailure, SubagentStart hook events; Added agent hook type; Added LSP advanced options; Added managed installation scope
