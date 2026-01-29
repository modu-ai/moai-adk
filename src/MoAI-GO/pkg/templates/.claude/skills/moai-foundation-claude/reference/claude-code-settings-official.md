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

Base Settings Framework:
```json
{
 "model": "claude-3-5-sonnet-20241022",
 "permissionMode": "default",
 "maxFileSize": 10000000,
 "maxTokens": 200000,
 "temperature": 1.0,
 "environment": {},
 "hooks": {},
 "plugins": {},
 "subagents": {},
 "mcpServers": {},
 "allowedTools": [],
 "toolRestrictions": {},
 "memory": {},
 "logging": {},
 "security": {}
}
```

### Essential Configuration Fields

Model Configuration:
```json
{
 "model": "claude-3-5-sonnet-20241022",
 "maxTokens": 200000,
 "temperature": 1.0,
 "topP": 1.0,
 "topK": 0
}
```

Permission Management:
```json
{
 "permissionMode": "default",
 "allowedTools": [
 "Read",
 "Write",
 "Edit",
 "Bash",
 "Grep",
 "Glob",
 "WebFetch",
 "AskUserQuestion"
 ],
 "toolRestrictions": {
 "Bash": "prompt",
 "Write": "prompt",
 "Edit": "prompt"
 }
}
```

## Detailed Configuration Sections

### Model Settings

Available Models:
- `claude-3-5-sonnet-20241022`: Balanced performance (default)
- `claude-3-5-haiku-20241022`: Fast and cost-effective
- `claude-3-opus-20240229`: Highest quality, higher cost

Model Configuration Examples:
```json
{
 "model": "claude-3-5-sonnet-20241022",
 "maxTokens": 200000,
 "temperature": 0.7,
 "topP": 0.9,
 "topK": 40,
 "stopSequences": ["---", "##"],
 "timeout": 300000
}
```

Model Selection Guidelines:
```json
{
 "modelProfiles": {
 "development": {
 "model": "claude-3-5-haiku-20241022",
 "temperature": 0.3,
 "maxTokens": 50000
 },
 "testing": {
 "model": "claude-3-5-sonnet-20241022",
 "temperature": 0.1,
 "maxTokens": 100000
 },
 "production": {
 "model": "claude-3-5-sonnet-20241022",
 "temperature": 0.0,
 "maxTokens": 200000
 }
 }
}
```

### Permission System

Permission Modes:
- `default`: Standard permission prompts for sensitive operations
- `acceptEdits`: Automatically accept file edits without prompts
- `dontAsk`: Suppress all permission dialogs

Tool-Specific Permissions:
```json
{
 "allowedTools": [
 "Read",
 "Write",
 "Edit",
 "Bash",
 "Grep",
 "Glob",
 "WebFetch",
 "WebSearch",
 "AskUserQuestion",
 "TodoWrite",
 "Task",
 "Skill",
 "SlashCommand"
 ],
 "toolRestrictions": {
 "Read": {
 "allowedPaths": ["./", "~/.claude/"],
 "blockedPaths": [".env*", "*.key", "*.pem"]
 },
 "Bash": {
 "allowedCommands": ["git", "npm", "python", "make", "docker"],
 "blockedCommands": ["rm -rf", "sudo", "chmod 777", "dd", "mkfs"],
 "requireConfirmation": true
 },
 "Write": {
 "allowedExtensions": [".md", ".py", ".js", ".ts", ".json", ".yaml"],
 "blockedExtensions": [".exe", ".bat", ".sh", ".key"],
 "maxFileSize": 10000000
 },
 "WebFetch": {
 "allowedDomains": ["*.github.com", "*.npmjs.com", "docs.python.org"],
 "blockedDomains": ["*.malicious-site.com"],
 "requireConfirmation": false
 }
 }
}
```

### Environment Variables

Environment Configuration:
```json
{
 "environment": {
 "NODE_ENV": "development",
 "PYTHONPATH": "./src",
 "API_KEY": "$ENV_VAR", // Environment variable reference
 "PROJECT_ROOT": ".", // Static value
 "DEBUG": "true",
 "LOG_LEVEL": "$DEFAULT_LOG_LEVEL"
 }
}
```

Variable Resolution:
```json
{
 "environmentResolution": {
 "precedence": [
 "runtime_environment",
 "settings_json",
 "default_values"
 ],
 "validation": {
 "required": ["PROJECT_ROOT"],
 "optional": ["DEBUG", "LOG_LEVEL"],
 "typeChecking": true
 }
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
 "command": "echo 'Bash command: $COMMAND' >> ~/.claude/hooks.log"
 }
 ]
 }
 ],
 "PostToolUse": [
 {
 "matcher": "*",
 "hooks": [
 {
 "type": "command",
 "command": "echo 'Tool executed: $TOOL_NAME' >> ~/.claude/activity.log"
 }
 ]
 }
 ],
 "UserPromptSubmit": [
 {
 "hooks": [
 {
 "type": "validation",
 "pattern": "^[\\w\\s\\.\\?!]+$",
 "message": "Invalid characters in prompt"
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
