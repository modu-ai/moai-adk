# Claude Code Settings - Reference

> **Illustrative reference, NOT the official Claude Code specification.**
> This file is an MoAI-authored summary intended for orientation only. The
> authoritative, up-to-date settings schema lives in the official Claude Code
> documentation at https://code.claude.com/docs/en/settings — consult that
> source before relying on any field for a production `settings.json`. Some
> blocks below are illustrative patterns rather than real `settings.json`
> keys; the headers flag those.

Source: https://code.claude.com/docs/en/settings

## Key Concepts

### What are Claude Code Settings?

Claude Code Settings provide a hierarchical configuration system that controls Claude Code's behavior, tool permissions, model selection, and integration preferences. Settings are managed through JSON configuration files with clear inheritance and override patterns.

### Settings Hierarchy

Configuration Priority (highest to lowest):
1. Enterprise Settings: Organization-wide policies and restrictions
2. User Settings: `~/.claude/settings.json` (personal preferences)
3. Project Settings: `.claude/settings.json` (team-shared)
4. Local Settings: `.claude/settings.local.json` (local overrides)

Inheritance Flow:
```
Enterprise Policy → User Settings → Project Settings → Local Settings
 (Applied) (Personal) (Team) (Local)
 ↓ ↓ ↓ ↓
 Overrides Overrides Overrides Overrides
```

## Core Settings Structure

### Complete Configuration Schema

Base Settings Framework (valid top-level fields):
```json
{
 "model": "claude-sonnet-5",
 "permissions": {},
 "hooks": {},
 "disableAllHooks": false,
 "env": {},
 "statusLine": {},
 "outputStyle": "",
 "cleanupPeriodDays": 30,
 "sandbox": {},
 "enabledPlugins": {},
 "enabledMcpjsonServers": [],
 "disabledMcpjsonServers": []
}
```

### Essential Configuration Fields

Key fields frequently used in settings.json:
- `model`: Default model identifier
- `permissions`: Tool allow/ask/deny lists
- `hooks`: Lifecycle event hooks
- `env`: Environment variables
- `statusLine`: Status bar configuration
- `outputStyle`: Output formatting style
- `cleanupPeriodDays`: Session cleanup period
- `sandbox`: Sandboxing configuration

## Detailed Configuration Sections

### Model Settings

The `model` field sets the default model. Only this single field is valid in settings.json for model selection.

```json
{
 "model": "claude-sonnet-5"
}
```

### Permission System

Permission Modes: `default`, `plan`, `acceptEdits`, `bypassPermissions`.

Permissions use allow/ask/deny lists with tool-path patterns:
```json
{
 "permissions": {
 "defaultMode": "default",
 "allow": [
 "Read",
 "Glob",
 "Grep",
 "Bash(git status:*)",
 "Bash(git log:*)"
 ],
 "ask": [
 "Bash(rm:*)",
 "Bash(sudo:*)"
 ],
 "deny": [
 "Read(~/.ssh/**)",
 "Bash(rm -rf /:*)"
 ],
 "additionalDirectories": []
 }
}
```

### Environment Variables

The `env` field sets environment variables for the Claude Code session:
```json
{
 "env": {
 "NODE_ENV": "development",
 "PYTHONPATH": "./src",
 "DEBUG": "true"
 }
}
```

### MCP Server Configuration

MCP Server Setup:
```json
{
 "mcpServers": {
 "example-server": {
 "command": "npx",
 "args": ["@example/mcp-server"],
 "env": {
 "EXAMPLE_API_KEY": "$EXAMPLE_KEY"
 },
 "timeout": 30000
 },
 "figma": {
 "command": "npx",
 "args": ["@figma/mcp-server"],
 "env": {
 "FIGMA_API_KEY": "$FIGMA_KEY"
 }
 }
 }
}
```

MCP Permission Management:
```json
{
 "mcpPermissions": {
 "example-server": {
 "allowed": ["tool-a", "tool-b"],
 "rateLimit": {
 "requestsPerMinute": 60,
 "burstSize": 10
 }
 }
 }
}
```

### Hooks Configuration

Hook events: SessionStart, UserPromptSubmit, PreToolUse, PermissionRequest, PostToolUse, PostToolUseFailure, Notification, SubagentStart, SubagentStop, Stop, PreCompact, SessionEnd.

Hook handler types: "command" (shell command), "prompt" (LLM evaluation), "agent" (subagent with tool access).

Timeout unit: seconds. Defaults: 600 for command, 30 for prompt, 60 for agent.

Hooks Setup:
```json
{
 "hooks": {
 "PreToolUse": [
 {
 "matcher": "Bash",
 "hooks": [
 {
 "type": "command",
 "command": ".claude/hooks/block-rm.sh",
 "timeout": 10
 }
 ]
 }
 ],
 "PostToolUse": [
 {
 "matcher": "Write|Edit",
 "hooks": [
 {
 "type": "command",
 "command": "\"$CLAUDE_PROJECT_DIR\"/.claude/hooks/lint-check.sh",
 "timeout": 30
 }
 ]
 }
 ],
 "Stop": [
 {
 "hooks": [
 {
 "type": "prompt",
 "prompt": "Check if all tasks are complete: $ARGUMENTS",
 "timeout": 30
 }
 ]
 }
 ]
 }
}
```

### Sub-agent Configuration

Claude Code does NOT expose a `subagents` block in `settings.json`. Sub-agents
are configured as individual files under `.claude/agents/*.md` (frontmatter:
`name`, `description`, `tools`, `model`, etc.) and spawned via the `Agent`
tool. The earlier `subagents` / `defaultModel` / `allowedSubagents` /
`customSubagents` schema that appeared here was illustrative-only and did not
correspond to a real `settings.json` key — it has been removed. See
`claude-code-sub-agents-official.md` for the authoritative sub-agent
authoring surface.

### Plugin System

Plugin Configuration:
```json
{
 "plugins": {
 "enabled": true,
 "pluginPaths": ["./plugins", "~/.claude/plugins"],
 "loadedPlugins": [
 "git-integration",
 "docker-helper",
 "database-tools"
 ],
 "pluginSettings": {
 "git-integration": {
 "autoCommit": false,
 "branchStrategy": "feature-branch"
 },
 "docker-helper": {
 "defaultRegistry": "docker.io",
 "buildTimeout": 300000
 }
 }
 }
}
```

## File Locations and Management

### Settings File Paths

Standard Locations:
```bash
# Enterprise settings (system-wide)
/etc/claude/settings.json

# User settings (personal preferences)
~/.claude/settings.json

# Project settings (team-shared)
./.claude/settings.json

# Local overrides (development)
./.claude/settings.local.json

# Environment-specific overrides
./.claude/settings.${ENVIRONMENT}.json
```

### Settings Management Commands

> The commands below are illustrative. Claude Code's real CLI surface is
> narrower than this list — it exposes `claude config set`/`get`/`list` for
> a small set of keys plus the interactive `/config` command inside the
> TUI. The `--environment` flag, `use-environment`, `list-environments`,
> and any `maxTokens` setter shown below do not exist in the real CLI
> (there is no `maxTokens` field). Treat this block as a pattern sketch,
> not a copy-paste source; verify against
> https://code.claude.com/docs/en/settings.

Configuration Commands (illustrative):
```bash
# View current settings
claude config list
claude config get model

# Set individual settings (real keys only — model, permissions, etc.)
claude config set model "claude-sonnet-5"

# Edit settings file directly
$EDITOR .claude/settings.json
```

For interactive configuration, use the `/config` command inside a Claude
Code session — it writes to the appropriate settings scope (local/project/
user) automatically.

## Advanced Configuration

> The blocks that appeared here earlier — `context.maxTokens` /
> `compressionThreshold` / `cacheStrategy`, `logging`, `debug`, and
> `performance` top-level settings — were illustrative-only and did not
> correspond to real `settings.json` keys. Claude Code does not expose a
> `context`, `logging`, `debug`, or `performance` top-level object. They have
> been removed. Context-window behavior is governed by the runtime and by
> `cleanupPeriodDays`; logging is governed by the runtime's debug flags
> (`claude --debug ...`) rather than a `settings.json` block.

## Integration Settings

> The `git` and `cicd` top-level settings blocks that appeared here earlier
> were illustrative-only and did not correspond to real `settings.json` keys.
> Claude Code does not commit, push, or generate CI pipelines from a
> `settings.json` block; git/CI behavior is driven by your shell, your
> repository hooks, and the MoAI orchestrator. They have been removed.

## Security Configuration

> The `security` and `privacy` top-level settings blocks that appeared here
> earlier were illustrative-only and did not correspond to real
> `settings.json` keys. Claude Code's security surface is the
> `permissions` block (allow/ask/deny lists, permission modes), sandboxing
> (`sandbox`), and `disableBypassPermissionsMode` — not an inline
> `security`/`privacy` object. They have been removed.

## Best Practices

### Configuration Management

Development Practices:
- Use version control for project settings
- Keep local overrides in `.gitignore`
- Document all custom settings
- Validate settings before deployment

Security Practices:
- Never commit sensitive credentials
- Use environment variables for secrets
- Implement principle of least privilege
- Regular security audits

Performance Practices:
- Optimize context window usage
- Enable caching where appropriate
- Monitor token usage
- Use appropriate models for tasks

### Organization Standards

> The `team` and `enterprise` top-level settings blocks that appeared here
> earlier were illustrative-only and did not correspond to real
> `settings.json` keys. Claude Code's enterprise surface uses separate
> managed-settings files (`/etc/claude/settings.json` /
> `/Library/Application Support/ClaudeCode/managed-settings.json`) plus
> `disableBypassPermissionsMode`, not an inline `team`/`enterprise` object
> with SOC2/ISO27001 compliance fields. They have been removed. See the
> official IAM documentation at
> https://code.claude.com/docs/en/iam for the real enterprise/managed
> settings surface.

This reference orients you to the categories of Claude Code configuration. For the authoritative and current schema — including fields added after this file was last reviewed — consult the official documentation at https://code.claude.com/docs/en/settings.
