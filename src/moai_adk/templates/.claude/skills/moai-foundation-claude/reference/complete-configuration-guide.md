# Claude Code Complete Configuration Guide

## IAM & Permission Rules

### Tool-Specific Permission Rules

Tiered Permission System:
1. Read-only: No approval required
2. Bash Commands: User approval required
3. File Modification: User approval required

Permission Rule Format:
```json
{
 "allowedTools": [
 "Read", // Read-only access
 "Bash", // Commands with approval
 "Write", // File modification with approval
 "WebFetch(domain:*.example.com)" // Domain-specific web access
 ]
}
```

### Enterprise Policy Overrides

Enterprise IAM Structure:
```json
{
 "enterprise": {
 "policies": {
 "tools": {
 "Bash": "never", // Enterprise-wide restriction
 "WebFetch": ["domain:*.company.com"] // Approved domains only
 },
 "mcpServers": {
 "allowed": ["context7", "figma"], // Approved MCP servers
 "blocked": ["custom-mcp"] // Blocked servers
 }
 }
 }
}
```

### Permission Configuration Examples

Development Environment:
```json
{
 "allowedTools": [
 "Read",
 "Write",
 "Edit",
 "Bash",
 "WebFetch",
 "Grep",
 "Glob"
 ],
 "toolRestrictions": {
 "Bash": {
 "allowedCommands": ["npm", "python", "git", "make"],
 "blockedCommands": ["rm -rf", "sudo", "chmod 777"]
 },
 "WebFetch": {
 "allowedDomains": ["*.github.com", "*.npmjs.com", "docs.python.org"],
 "blockedDomains": ["*.malicious-site.com"]
 }
 }
}
```

Production Environment:
```json
{
 "allowedTools": [
 "Read",
 "Grep",
 "Glob"
 ],
 "toolRestrictions": {
 "Write": "never",
 "Edit": "never",
 "Bash": "never"
 }
}
```

### MCP Permission Management

MCP servers do not support wildcards - specific server names required:

```json
{
 "allowedMcpServers": [
 "context7",
 "figma-dev-mode-mcp-server",
 "playwright"
 ],
 "blockedMcpServers": [
 "custom-unverified-mcp"
 ]
}
```

## Claude Code Settings

### Settings Hierarchy

Configuration Priority (highest to lowest):
1. Enterprise Settings: Organization-wide policies
2. User Settings: `~/.claude/settings.json` (personal)
3. Project Settings: `.claude/settings.json` (shared)
4. Local Settings: `.claude/settings.local.json` (local overrides)

### Core Settings Structure

```json
{
 "model": "claude-3-5-sonnet-20241022",
 "permissionMode": "default",
 "maxFileSize": 10000000,
 "maxTokens": 200000,
 "environment": {},
 "hooks": {},
 "plugins": {},
 "subagents": {},
 "mcpServers": {}
}
```

### Key Configuration Options

Model Settings:
```json
{
 "model": "claude-3-5-sonnet-20241022", // or haiku, opus
 "maxTokens": 200000, // Context window limit
 "temperature": 1.0 // Creativity level (0.0-1.0)
}
```

Permission Management:
```json
{
 "permissionMode": "default", // default, acceptEdits, dontAsk
 "tools": {
 "Bash": "prompt", // always, prompt, never
 "Write": "prompt",
 "Edit": "prompt"
 }
}
```

Environment Variables:
```json
{
 "environment": {
 "NODE_ENV": "development",
 "API_KEY": "$ENV_VAR", // Environment variable reference
 "PROJECT_ROOT": "." // Static value
 }
}
```

MCP Server Configuration:
```json
{
 "mcpServers": {
 "context7": {
 "command": "npx",
 "args": ["@upstash/context7-mcp"],
 "env": {"CONTEXT7_API_KEY": "$CONTEXT7_KEY"}
 }
 }
}
```
