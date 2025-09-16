# Claude Code Settings Configuration

## Overview

Claude Code offers flexible configuration options through various settings files and environment variables. The configuration system supports hierarchical settings with different levels of precedence.

## Settings Files Locations

1. **User Settings**: `~/.claude/settings.json`
   - Applies to all projects
   - Personal global settings

2. **Project Settings**:
   - `.claude/settings.json`: Shared team settings
   - `.claude/settings.local.json`: Personal project-specific settings

3. **Enterprise Managed Policies**:
   - macOS: `/Library/Application Support/ClaudeCode/managed-settings.json`
   - Linux/WSL: `/etc/claude-code/managed-settings.json`
   - Windows: `C:\ProgramData\ClaudeCode\managed-settings.json`

## Example Settings File

```json
{
  "permissions": {
    "allow": ["Bash(npm run lint)", "Bash(npm run test:*)", "Read(~/.zshrc)"],
    "deny": ["Bash(curl:*)", "Read(./.env)", "Read(./secrets/**)"]
  },
  "env": {
    "CLAUDE_CODE_ENABLE_TELEMETRY": "1"
  }
}
```

## Key Configuration Options

### Permissions Settings

- `allow`: Permitted tool usage rules
- `deny`: Denied tool usage rules
- `ask`: Tools requiring confirmation
- `additionalDirectories`: Extra working directories
- `defaultMode`: Default permission mode

### Global Configuration Keys

- `autoUpdates`: Enable/disable automatic updates
- `preferredNotifChannel`: Notification delivery method
- `theme`: Color theme selection
- `verbose`: Show detailed command outputs

## Environment Variables

Claude Code supports numerous environment variables for fine-tuned control, including:

- `ANTHROPIC_API_KEY`: API authentication
- `CLAUDE_CODE_USE_BEDROCK`: Use Amazon Bedrock
- `CLAUDE_CODE_USE_VERTEX`: Use Google Vertex AI
- `DISABLE_AUTOUPDATER`: Disable automatic updates
- `DISABLE_TELEMETRY`: Opt out of telemetry
- `HTTP_PROXY` / `HTTPS_PROXY`: Proxy configuration
- `MAX_THINKING_TOKENS`: Control thinking token usage

## Tools Available to Claude

Claude Code has access to powerful tools:

| Tool             | Description                  | Needs Permission |
| ---------------- | ---------------------------- | ---------------- |
| **Bash**         | Execute shell commands       | Yes              |
| **Edit**         | Make targeted edits to files | Yes              |
| **Glob**         | Find files based on patterns | No               |
| **Grep**         | Search for patterns in files | No               |
| **LS**           | List files and directories   | No               |
| **MultiEdit**    | Multiple atomic edits        | Yes              |
| **NotebookEdit** | Modify Jupyter notebooks     | Yes              |
| **Read**         | Read file contents           | No               |
| **Task**         | Launch subagents             | No               |
| **TodoWrite**    | Manage task lists            | No               |
| **WebFetch**     | Fetch web content            | Yes              |
| **WebSearch**    | Perform web searches         | Yes              |
| **Write**        | Create or overwrite files    | Yes              |

## Configuration Priority

Settings are applied in priority order (highest to lowest):

1. **Enterprise managed policies**
2. **Command-line arguments**
3. **Local project settings** (`.claude/settings.local.json`)
4. **Shared project settings** (`.claude/settings.json`)
5. **User settings** (`~/.claude/settings.json`)

## Configuration Commands

- List settings: `claude config list`
- View setting: `claude config get <key>`
- Change setting: `claude config set <key> <value>`
- Add to setting: `claude config add <key> <value>`
- Remove from setting: `claude config remove <key> <value>`

Use `--global` flag for global configuration management.
