# Claude Code Settings - Official Documentation Reference

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
 "model": "claude-sonnet-4-5-20250929",
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
 "model": "claude-sonnet-4-5-20250929"
}
```

### Permission System

Permission Modes: `default`, `plan`, `acceptEdits`, `dontAsk`, `bypassPermissions`.

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
 "context7": {
 "command": "npx",
 "args": ["@upstash/context7-mcp"],
 "env": {
 "CONTEXT7_API_KEY": "$CONTEXT7_KEY"
 },
 "timeout": 30000
 },
 "sequential-thinking": {
 "command": "npx",
 "args": ["@modelcontextprotocol/server-sequential-thinking"],
 "env": {},
 "timeout": 60000
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
 "context7": {
 "allowed": ["resolve-library-id", "get-library-docs"],
 "rateLimit": {
 "requestsPerMinute": 60,
 "burstSize": 10
 }
 },
 "sequential-thinking": {
 "allowed": ["*"], // All permissions
 "maxContextSize": 100000
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

Sub-agent Settings:
```json
{
 "subagents": {
 "defaultModel": "claude-3-5-sonnet-20241022",
 "defaultPermissionMode": "default",
 "maxConcurrentTasks": 5,
 "taskTimeout": 300000,
 "allowedSubagents": [
 "spec-builder",
 "ddd-implementer",
 "security-expert",
 "backend-expert",
 "frontend-expert"
 ],
 "customSubagents": {
 "custom-analyzer": {
 "description": "Custom code analysis agent",
 "tools": ["Read", "Grep", "Bash"],
 "model": "claude-3-5-sonnet-20241022"
 }
 }
 }
}
```

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

Configuration Commands:
```bash
# View current settings
claude settings show
claude settings show --model
claude settings show --permissions

# Set individual settings
claude config set model "claude-3-5-sonnet-20241022"
claude config set maxTokens 200000
claude config set permissionMode "default"

# Edit settings file
claude config edit
claude config edit --local
claude config edit --user

# Reset settings
claude config reset
claude config reset --local
claude config reset --user

# Validate settings
claude config validate
claude config validate --strict
```

Environment-Specific Settings:
```bash
# Set environment-specific settings
claude config set --environment development model "claude-3-5-haiku-20241022"
claude config set --environment production maxTokens 200000

# Switch between environments
claude config use-environment development
claude config use-environment production

# List available environments
claude config list-environments
```

## Advanced Configuration

### Context Management

Context Window Settings:
```json
{
 "context": {
 "maxTokens": 200000,
 "compressionThreshold": 150000,
 "compressionStrategy": "importance-based",
 "memoryIntegration": true,
 "cacheStrategy": {
 "enabled": true,
 "maxSize": "100MB",
 "ttl": 3600
 }
 }
}
```

### Logging and Debugging

Logging Configuration:
```json
{
 "logging": {
 "level": "info",
 "file": "~/.claude/logs/claude.log",
 "maxFileSize": "10MB",
 "maxFiles": 5,
 "format": "json",
 "include": [
 "tool_usage",
 "agent_delegation",
 "errors",
 "performance"
 ],
 "exclude": [
 "sensitive_data"
 ]
 }
}
```

Debug Settings:
```json
{
 "debug": {
 "enabled": false,
 "verboseOutput": false,
 "timingInfo": false,
 "tokenUsage": true,
 "stackTraces": false,
 "apiCalls": false
 }
}
```

### Performance Optimization

Performance Settings:
```json
{
 "performance": {
 "parallelExecution": true,
 "maxConcurrency": 5,
 "caching": {
 "enabled": true,
 "strategy": "lru",
 "maxSize": "500MB"
 },
 "optimization": {
 "contextCompression": true,
 "responseStreaming": false,
 "batchProcessing": true
 }
 }
}
```

## Integration Settings

### Git Integration

Git Configuration:
```json
{
 "git": {
 "autoCommit": false,
 "autoPush": false,
 "branchStrategy": "feature-branch",
 "commitTemplate": {
 "prefix": "feat:",
 "includeScope": true,
 "includeBody": true
 },
 "hooks": {
 "preCommit": "lint && test",
 "prePush": "security-scan"
 }
 }
}
```

### CI/CD Integration

CI/CD Settings:
```json
{
 "cicd": {
 "platform": "github-actions",
 "configPath": ".github/workflows/",
 "autoGenerate": false,
 "pipelines": {
 "test": {
 "trigger": ["push", "pull_request"],
 "steps": ["lint", "test", "security-scan"]
 },
 "deploy": {
 "trigger": ["release"],
 "steps": ["build", "deploy"]
 }
 }
 }
}
```

## Security Configuration

### Security Settings

Security Configuration:
```json
{
 "security": {
 "level": "standard",
 "encryption": {
 "enabled": true,
 "algorithm": "AES-256-GCM"
 },
 "accessControl": {
 "authentication": "required",
 "authorization": "role-based"
 },
 "audit": {
 "enabled": true,
 "logLevel": "detailed",
 "retention": "90d"
 }
 }
}
```

### Privacy Settings

Privacy Configuration:
```json
{
 "privacy": {
 "dataCollection": "minimal",
 "analytics": false,
 "crashReporting": true,
 "usageStatistics": false,
 "dataRetention": {
 "logs": "30d",
 "cache": "7d",
 "temp": "1d"
 }
 }
}
```

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

Team Configuration:
```json
{
 "team": {
 "standards": {
 "model": "claude-3-5-sonnet-20241022",
 "testCoverage": 90,
 "codeStyle": "prettier",
 "documentation": "required"
 },
 "workflow": {
 "branching": "gitflow",
 "reviews": "required",
 "ciCd": "automated"
 }
 }
}
```

Enterprise Policies:
```json
{
 "enterprise": {
 "policies": {
 "allowedModels": ["claude-3-5-sonnet-20241022"],
 "maxTokens": 100000,
 "restrictedTools": ["Bash", "WebFetch"],
 "auditRequired": true
 },
 "compliance": {
 "standards": ["SOC2", "ISO27001"],
 "dataResidency": "us-east-1",
 "retentionPolicy": "7y"
 }
 }
}
```

This comprehensive reference provides all the information needed to configure Claude Code effectively for any use case, from personal development to enterprise deployment.
