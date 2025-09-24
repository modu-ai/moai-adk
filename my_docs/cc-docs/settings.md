# Settings

Configure Claude Code with global and project-level settings and environment variables.

Claude Code provides various settings to configure its behavior to your needs. When using the interactive REPL, you can run `/config` to configure Claude Code.

## Settings files

The `settings.json` file is the official mechanism to configure Claude Code through hierarchical settings:

- **User settings** are defined at `~/.claude/settings.json` and apply to all projects
- **Project settings** are stored in your project directory:
  - `.claude/settings.json` for settings checked into source control and shared with teams
  - `.claude/settings.local.json` for settings that are not checked in, useful for personal preferences and experimentation. Claude Code will configure git to ignore `.claude/settings.local.json` when it's created.
- For enterprise deployments of Claude Code, **enterprise managed policy settings** are also supported. These take precedence over user and project settings. System administrators can deploy policies to the following locations:
  - macOS: `/Library/Application Support/ClaudeCode/managed-settings.json`
  - Linux and WSL: `/etc/claude-code/managed-settings.json`
  - Windows: `C:\ProgramData\ClaudeCode\managed-settings.json`

### Example settings.json

```json
{
  "permissions": {
    "allow": [
      "Bash(npm run lint)",
      "Bash(npm run test:*)",
      "Read(~/.zshrc)"
    ],
    "deny": [
      "Bash(curl:*)",
      ".env",
      "secrets/**"
    ],
    "ask": [
      "Bash(rm *)",
      "Write(**/.env)"
    ]
  },
  "model": "opus",
  "forceLoginMethod": "claudeai",
  "enableAllProjectMcpServers": true,
  "enabledMcpjsonServers": ["github", "postgres"],
  "disabledMcpjsonServers": ["experimental-server"]
}
```

## Available settings

The `settings.json` supports several options:

| Setting | Description | Example |
|---------|-------------|---------|
| `apiKeyHelper` | Custom script to be executed in `/bin/sh` to generate authentication values. This value is sent as `X-Api-Key` and `Authorization: Bearer` headers for model requests | `/bin/generate_temp_api_key.sh` |
| `cleanupPeriodDays` | How long to keep chat history locally (default: 30 days) | `20` |
| `env` | Environment variables to apply to all sessions | `{"FOO": "bar"}` |
| `includeCoAuthoredBy` | Whether to include `co-authored-by Claude` signatures in git commits and pull requests (default: `true`) | `false` |
| `permissions` | Permission structure, see permissions table below | |
| `hooks` | Configure custom commands to run before and after tool execution. See [hooks documentation](hooks) | `{"PreToolUse": {"Bash": "echo 'Running command...'"}}` |
| `model` | Override the default model to use for Claude Code | `"claude-3-5-sonnet-20241022"` |
| `ask` | Array of permission rules requiring confirmation (new 2025 feature) | `["Bash(rm *)", "Write(**/.env)"]` |
| `forceLoginMethod` | Use `claudeai` to restrict login to Claude.ai accounts, `console` to restrict to Anthropic Console (API billing) accounts | `claudeai` |
| `enableAllProjectMcpServers` | Automatically approve all MCP servers defined in project `.mcp.json` files | `true` |
| `enabledMcpjsonServers` | List of specific MCP servers to approve from `.mcp.json` files | `["memory", "github"]` |
| `disabledMcpjsonServers` | List of specific MCP servers to reject from `.mcp.json` files | `["filesystem"]` |
| `awsAuthRefresh` | Custom script that modifies the `.aws` directory (see [Advanced credential configuration](amazon-bedrock#advanced-credential-configuration)) | `aws sso login --profile myprofile` |
| `awsCredentialExport` | Custom script that outputs JSON containing AWS credentials (see [Advanced credential configuration](amazon-bedrock#advanced-credential-configuration)) | `/bin/generate_aws_grant.sh` |

### Permissions settings

| Setting | Description | Example |
|---------|-------------|---------|
| `allow` | Array of [permission rules](iam#configuring-permissions) to allow tool usage | `["Bash(git diff:*)"]` |
| `deny` | Array of [permission rules](iam#configuring-permissions) to deny tool usage | `["WebFetch", "Bash(curl:*)"]` |
| `ask` | Array of permission rules requiring user confirmation before execution (2025 feature) | `["Bash(rm *)", "Write(**/.env)"]` |
| `additionalDirectories` | Additional [working directories](iam#working-directories) Claude can access | `["../docs/"]` |
| `defaultMode` | Default [permission mode](iam#permission-modes) when opening Claude Code | `"acceptEdits"` |
| `disableBypassPermissionsMode` | Set to `"disable"` to prevent `bypassPermissions` mode from being enabled. See [enterprise managed policy settings](iam#enterprise-managed-policy-settings) | `"disable"` |

## Settings priority

Settings are applied in priority order (high to low):

1. **Enterprise managed policy** (`managed-settings.json`)
   - Deployed by IT/DevOps
   - Cannot be overridden

2. **Command line arguments**
   - Temporary overrides for specific sessions

3. **Local project settings** (`.claude/settings.local.json`)
   - Personal per-project settings

4. **Shared project settings** (`.claude/settings.json`)
   - Team-shared project settings in source control

5. **User settings** (`~/.claude/settings.json`)
   - Personal global settings

This hierarchy ensures enterprise security policies are always enforced while allowing teams and individuals to customize their experience.

### Configuration system overview

- **Memory files (CLAUDE.md)**: Contains instructions and context that Claude loads at startup
- **Settings files (JSON)**: Configure permissions, environment variables, and tool behavior
- **Slash commands**: Custom commands that can be invoked during sessions with `/command-name`
- **MCP servers**: Extend Claude Code with additional tools and integrations
- **Priority**: Higher-level configurations (enterprise) override lower-level configurations (user/project)
- **Inheritance**: Settings are merged, with more specific settings adding to or overriding broader settings

## Subagent configuration

Claude Code supports custom AI subagents that can be configured at the user and project level. These subagents are stored as Markdown files with YAML frontmatter:

- **User subagents**: `~/.claude/agents/` - Available across all projects
- **Project subagents**: `.claude/agents/` - Project-specific and shareable with teams

Subagent files define specialized AI assistants with custom prompts and tool permissions. For more details on creating and using subagents, see the [subagents documentation](sub-agents).

## Environment variables

Claude Code supports the following environment variables to control behavior:

> **Note**: All environment variables can also be configured in [`settings.json`](#available-settings). This is a useful way to automatically set environment variables for each session or deploy a set of environment variables to an entire team or organization.

| Variable | Purpose |
|----------|---------|
| `ANTHROPIC_API_KEY` | API key sent as `X-Api-Key` header, typically for Claude SDK (run `/login` for interactive use) |
| `ANTHROPIC_AUTH_TOKEN` | Custom value for `Authorization` header (value you set here is prefixed with `Bearer `) |
| `ANTHROPIC_CUSTOM_HEADERS` | Custom headers you want to add to requests (format: `Name: Value`) |
| `ANTHROPIC_MODEL` | Custom model name to use (see [model configuration](model-config)) |
| `ANTHROPIC_SMALL_FAST_MODEL` | Name of Haiku-class model for background tasks |
| `ANTHROPIC_SMALL_FAST_MODEL_AWS_REGION` | Override AWS region for small/fast model when using Bedrock |
| `AWS_BEARER_TOKEN_BEDROCK` | Bedrock API key for authentication (see [Bedrock API keys](https://aws.amazon.com/blogs/machine-learning/accelerate-ai-development-with-amazon-bedrock-api-keys/)) |
| `BASH_DEFAULT_TIMEOUT_MS` | Default timeout for long-running bash commands |
| `BASH_MAX_TIMEOUT_MS` | Maximum timeout the model can set for long-running bash commands |
| `BASH_MAX_OUTPUT_LENGTH` | Maximum number of characters before bash output is mid-truncated |
| `CLAUDE_BASH_MAINTAIN_PROJECT_WORKING_DIR` | Return to original working directory after each Bash command |
| `CLAUDE_CODE_API_KEY_HELPER_TTL_MS` | Interval in milliseconds that credentials should be refreshed (when using `apiKeyHelper`) |
| `CLAUDE_CODE_IDE_SKIP_AUTO_INSTALL` | Skip auto-installation of IDE extensions |
| `CLAUDE_CODE_MAX_OUTPUT_TOKENS` | Set maximum output tokens for most requests |
| `CLAUDE_CODE_USE_BEDROCK` | Use Bedrock |
| `CLAUDE_CODE_USE_VERTEX` | Use Vertex |
| `CLAUDE_CODE_SKIP_BEDROCK_AUTH` | Skip AWS authentication to Bedrock (e.g., when using an LLM gateway) |
| `CLAUDE_CODE_SKIP_VERTEX_AUTH` | Skip Google authentication to Vertex (e.g., when using an LLM gateway) |
| `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC` | Same as setting `DISABLE_AUTOUPDATER`, `DISABLE_BUG_COMMAND`, `DISABLE_ERROR_REPORTING`, `DISABLE_TELEMETRY` |
| `CLAUDE_CODE_DISABLE_TERMINAL_TITLE` | Set to `1` to disable automatic terminal title updates based on conversation context |
| `DISABLE_AUTOUPDATER` | Set to `1` to disable auto-updates. This takes precedence over the `autoUpdates` config setting. |
| `DISABLE_BUG_COMMAND` | Set to `1` to disable the `/bug` command |
| `DISABLE_COST_WARNINGS` | Set to `1` to disable cost warning messages |
| `DISABLE_ERROR_REPORTING` | Set to `1` to opt out of Sentry error reporting |
| `DISABLE_NON_ESSENTIAL_MODEL_CALLS` | Set to `1` to disable model calls for non-essential paths like flavor text |
| `DISABLE_TELEMETRY` | Set to `1` to opt out of Statsig telemetry (Statsig events do not include user data like code, file paths, or bash commands) |
| `HTTP_PROXY` | Specify HTTP proxy server for network connections |
| `HTTPS_PROXY` | Specify HTTPS proxy server for network connections |
| `MAX_THINKING_TOKENS` | Force thinking for model budget |
| `MCP_TIMEOUT` | Timeout in milliseconds for MCP server startup |
| `MCP_TOOL_TIMEOUT` | Timeout in milliseconds for MCP tool execution |
| `MAX_MCP_OUTPUT_TOKENS` | Maximum number of tokens allowed in MCP tool responses (default: 25000) |
| `VERTEX_REGION_CLAUDE_3_5_HAIKU` | Override region for Claude 3.5 Haiku when using Vertex AI |
| `VERTEX_REGION_CLAUDE_3_5_SONNET` | Override region for Claude 3.5 Sonnet when using Vertex AI |
| `VERTEX_REGION_CLAUDE_3_7_SONNET` | Override region for Claude 3.7 Sonnet when using Vertex AI (2025 model) |
| `VERTEX_REGION_CLAUDE_4_0_OPUS` | Override region for Claude 4.0 Opus when using Vertex AI (2025 model) |
| `VERTEX_REGION_CLAUDE_4_0_SONNET` | Override region for Claude 4.0 Sonnet when using Vertex AI (2025 model) |
| `VERTEX_REGION_CLAUDE_4_1_OPUS` | Override region for Claude 4.1 Opus when using Vertex AI (2025 model) |

## Configuration commands

To manage configuration, use the following commands:

- List settings: `claude config list`
- View setting: `claude config get <key>`
- Change setting: `claude config set <key> <value>`
- Add to setting (for lists): `claude config add <key> <value>`
- Remove from setting (for lists): `claude config remove <key> <value>`

By default, `config` changes project configuration. To manage global configuration, use the `--global` (or `-g`) flag.

### Global configuration

To set global configuration, use `claude config set -g <key> <value>`:

| Setting | Description | Example |
|---------|-------------|---------|
| `autoUpdates` | Whether to enable auto-updates (default: `true`). When enabled, Claude Code will automatically download and install updates in the background. Updates are applied when you restart Claude Code. | `false` |
| `preferredNotifChannel` | Where you want to receive notifications (default: `iterm2`) | `iterm2`, `iterm2_with_bell`, `terminal_bell`, or `notifications_disabled` |
| `theme` | Color theme | `dark`, `light`, `light-daltonized`, or `dark-daltonized` |
| `verbose` | Whether to show full bash and command output (default: `false`) | `true` |

## Available tools for Claude

Claude Code has access to a powerful set of tools to help understand and modify your codebase:

| Tool | Description | Permissions Required |
|------|-------------|---------------------|
| **Bash** | Execute shell commands in your environment | Yes |
| **Edit** | Perform targeted edits on specific files | Yes |
| **Glob** | Find files based on pattern matching | No |
| **Grep** | Search for patterns in file contents | No |
| **LS** | List files and directories | No |
| **MultiEdit** | Perform multiple edits atomically on a single file | Yes |
| **NotebookEdit** | Modify Jupyter notebook cells | Yes |
| **NotebookRead** | Read and display Jupyter notebook contents | No |
| **Read** | Read the contents of files | No |
| **Task** | Launch subagents to handle complex multi-step tasks | No |
| **TodoWrite** | Create and manage structured task lists | No |
| **WebFetch** | Fetch content from specified URLs | Yes |
| **WebSearch** | Perform web searches with domain filtering | Yes |
| **Write** | Create or overwrite files | Yes |

Permission rules can be configured using `/allowed-tools` or in [permission settings](#permissions-settings).

### Extending tools with hooks

You can use Claude Code hooks to execute custom commands before and after tool execution.

For example, you could automatically run a Python formatter after Claude modifies a Python file, or block Write operations to certain paths to prevent modification of production configuration files.
