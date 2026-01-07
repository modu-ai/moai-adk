# OpenCode.ai Extended Reference

Comprehensive supplementary documentation for advanced OpenCode features, configurations, and integrations.

---

## Installation Methods

### Package Managers

```bash
# npm
npm install -g opencode

# Bun
bun install -g opencode

# pnpm
pnpm add -g opencode

# Yarn
yarn global add opencode

# Homebrew (macOS)
brew install opencode

# Paru (Arch Linux)
paru -S opencode

# Chocolatey (Windows)
choco install opencode

# Scoop (Windows)
scoop install opencode
```

### Docker

```bash
docker run -it --rm ghcr.io/anomalyco/opencode
```

### Version Management

```bash
# Upgrade to latest
opencode upgrade

# Upgrade to specific version
opencode upgrade 1.0.0

# Downgrade
opencode upgrade 0.15.31

# Uninstall
opencode uninstall
```

---

## TUI (Terminal User Interface)

### Keyboard Shortcuts

Leader Key: `ctrl+x` (prefix for most actions)

Application Control:
- `ctrl+c`, `ctrl+d`, `<leader>q` - Exit application
- `<leader>e` - Open external editor
- `<leader>b` - Toggle sidebar
- `<leader>t` - List themes

Session Management:
- `<leader>n` - New session
- `<leader>l` - List sessions
- `<leader>x` - Export session
- `<leader>c` - Compact session
- `<leader>g` - Session timeline

Navigation:
- `pageup` / `pagedown` - Scroll messages
- `ctrl+g`, `home` - First message
- `ctrl+alt+g`, `end` - Last message

Input:
- `return` - Submit
- `shift+return`, `ctrl+return` - New line
- `ctrl+a` / `ctrl+e` - Line start/end
- `ctrl+shift+d` - Delete line

Model/Agent:
- `<leader>m` - Model list
- `f2` - Cycle recent models
- `<leader>a` - Agent list
- `ctrl+p` - Command palette
- `Tab` - Switch agent

### Slash Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `/connect` | - | Add provider credentials |
| `/compact` | `/summarize` | Compress session context |
| `/details` | - | Toggle tool execution visibility |
| `/editor` | - | Open external editor |
| `/exit` | `/quit`, `/q` | Exit application |
| `/export` | - | Export to Markdown |
| `/help` | - | Display help |
| `/init` | - | Create/update AGENTS.md |
| `/models` | - | List available models |
| `/new` | `/clear` | Start fresh session |
| `/redo` | - | Restore undone message |
| `/sessions` | `/resume`, `/continue` | Switch sessions |
| `/share` | - | Share session |
| `/themes` | - | List themes |
| `/undo` | - | Revert last message |
| `/unshare` | - | Revoke sharing |

### File References

Use `@` for fuzzy file searching:
```
@packages/functions/src/api/index.ts
```

### Bash Commands

Prefix with `!` to execute shell commands:
```
!npm test
```

### Editor Configuration

Set `EDITOR` environment variable:
```bash
export EDITOR="code --wait"    # VS Code
export EDITOR="cursor --wait"  # Cursor
export EDITOR="nvim"           # Neovim
```

---

## CLI Commands

### Primary Commands

```bash
# TUI
opencode [--model <model>] [--agent <agent>] [--port <port>]

# Non-interactive
opencode run "Your prompt here"

# Headless server
opencode serve [--port 4096] [--hostname 0.0.0.0]

# Web interface
opencode web

# Attach to running server
opencode attach

# ACP mode (for editors)
opencode acp
```

### Management Commands

```bash
# Agent management
opencode agent create
opencode agent list

# MCP server management
opencode mcp add <name>
opencode mcp list
opencode mcp auth <name>
opencode mcp logout <name>
opencode mcp debug <name>

# Authentication
opencode auth login
opencode auth list
opencode auth logout

# Model listing
opencode models [--provider <provider>] [--refresh]

# Session management
opencode session list [--format json] [--limit 10]
opencode export <session-id>
opencode import <file-or-url>

# Statistics
opencode stats [--since <date>] [--tool <tool>]

# GitHub integration
opencode github install
opencode github run
```

### Global Flags

```bash
--help, -h           # Display help
--version, -v        # Print version
--print-logs         # Output logs to stderr
--log-level <level>  # DEBUG, INFO, WARN, ERROR
```

### Environment Variables

```bash
OPENCODE_CONFIG           # Custom config path
OPENCODE_CONFIG_DIR       # Custom config directory
OPENCODE_DISABLE_AUTOUPDATE=1
OPENCODE_ENABLE_EXPERIMENTAL_MODELS=1
OPENCODE_EXPERIMENTAL=icon_discovery,watcher,lsp_tools
```

---

## Models

### Recommended Models

- GPT 5.2
- GPT 5.1 Codex
- Claude Opus 4.5
- Claude Sonnet 4.5
- Minimax M2.1
- Gemini 3 Pro

### Model Configuration

```json
{
  "model": "anthropic/claude-sonnet-4-5",
  "small_model": "anthropic/claude-haiku-3-5",
  "provider": {
    "anthropic": {
      "options": {
        "reasoning_effort": "high",
        "max_tokens": 8192
      }
    }
  }
}
```

### Model Variants

Built-in variants for different reasoning levels:
- Anthropic: high/max thinking
- OpenAI: reasoning effort levels
- Google: low/high effort

Custom variants:
```json
{
  "model_variant": {
    "anthropic/claude-sonnet-4-5": {
      "high": { "reasoning_effort": "high" },
      "max": { "reasoning_effort": "max" }
    }
  }
}
```

Cycle variants with `variant_cycle` keybind.

### Model Loading Priority

1. Command-line flag (`--model`)
2. Config file setting
3. Last used model
4. First model by internal priority

---

## OpenCode Zen

OpenCode Zen is a curated AI gateway with pre-tested models.

### Setup

1. Sign in at opencode.ai/auth
2. Add billing details
3. Copy API key
4. Run `/connect` in TUI
5. Select OpenCode Zen
6. Paste key

### Pricing

Pay-as-you-go per-million-token pricing:
- Free models: Grok Code Fast 1, GLM 4.7, MiniMax M2.1
- Premium: Claude Opus 4.5 ($5/$25 per 1M tokens)
- Auto-reload: $20 when balance < $5

### Team Features

- Roles: Admin (full access) / Member (personal keys only)
- Model curation: Control team-accessible models
- Bring your own key: Use existing provider credentials
- Monthly spending limits per workspace/member

---

## Themes

### Built-in Themes

- `system` - Adapts to terminal background
- `opencode` - Default
- `tokyonight` - Tokyo Night
- `everforest` - Everforest
- `ayu` - Ayu Dark
- `catppuccin` - Catppuccin
- `gruvbox` - Gruvbox
- `kanagawa` - Kanagawa
- `nord` - Nord
- `matrix` - Green-on-black
- `one-dark` - Atom One Dark

### Custom Themes

Location hierarchy (later overrides earlier):
1. Built-in themes
2. `~/.config/opencode/themes/*.json`
3. `<project-root>/.opencode/themes/*.json`

Theme JSON structure:
```json
{
  "theme": {
    "primary": "#7aa2f7",
    "secondary": "#9ece6a",
    "accent": "#bb9af7",
    "error": "#f7768e",
    "warning": "#e0af68",
    "success": "#9ece6a",
    "info": "#7dcfff",
    "text": "#c0caf5",
    "textMuted": "#565f89",
    "background": "#1a1b26",
    "borders": "#3b4261"
  }
}
```

Color formats:
- Hex: `"#ffffff"`
- ANSI: `3` (0-255)
- Reference: `"primary"`
- Variants: `{"dark": "#000", "light": "#fff"}`
- Terminal default: `"none"`

---

## Formatters

### Built-in Formatters

| Formatter | Extensions | Requirement |
|-----------|-----------|-------------|
| gofmt | .go | `gofmt` command |
| prettier | .js, .jsx, .ts, .tsx, .html, .css, .md, .json, .yaml | package.json |
| biome | .js, .jsx, .ts, .tsx, .html, .css, .md, .json | biome.json(c) |
| rustfmt | .rs | `rustfmt` |
| ruff | .py, .pyi | `ruff` + config |
| rubocop | .rb, .rake, .gemspec | `rubocop` |
| terraform | .tf, .tfvars | `terraform` |
| mix | .ex, .exs | mix format |
| zig | .zig | `zig fmt` |
| clang-format | .c, .cpp, .h, .hpp | clang-format |

### Custom Formatter

```json
{
  "formatter": {
    "my-prettier": {
      "command": ["npx", "prettier", "--write", "$FILE"],
      "environment": { "NODE_ENV": "development" },
      "extensions": [".js", ".ts", ".jsx", ".tsx"]
    }
  }
}
```

### Disable Formatters

```json
{
  "formatter": false
}
```

Or specific:
```json
{
  "formatter": {
    "prettier": { "disabled": true }
  }
}
```

---

## LSP Servers

### Built-in LSP Servers

- JavaScript/TypeScript: TypeScript, ESLint, Deno
- Python: Pyright
- Rust: rust-analyzer
- Go: gopls
- Java: JDTLS
- PHP: Intelephense
- C/C++: clangd
- C#: OmniSharp
- Ruby: Solargraph
- Swift: SourceKit-LSP
- And 15+ more languages

### Configuration

```json
{
  "lsp": {
    "typescript": { "disabled": true },
    "custom-lsp": {
      "command": ["custom-lsp-server", "--stdio"],
      "extensions": [".custom"],
      "env": { "DEBUG": "1" }
    }
  }
}
```

Disable all:
```json
{
  "lsp": false
}
```

---

## GitHub Integration

### Installation

```bash
opencode github install
```

Or manual:
1. Install GitHub app: `github.com/apps/opencode-agent`
2. Add workflow file: `.github/workflows/opencode.yml`
3. Store API keys in repository secrets

### Workflow Configuration

```yaml
- uses: opencode-agent/opencode-action@v1
  with:
    model: anthropic/claude-sonnet-4-20250514
    agent: build
    share: true
```

### Trigger Events

| Event | Trigger | Context |
|-------|---------|---------|
| `issue_comment` | `/oc` in comments | Issue/PR context |
| `pull_request_review_comment` | Code line comments | File path, diff |
| `schedule` | Cron jobs | Requires prompt |
| `issues` | Issue creation | Requires prompt |
| `pull_request` | PR changes | Full PR context |

### Usage

```
/opencode explain this issue
/opencode fix this
/oc review this merge request
```

---

## GitLab Integration

### GitLab CI

Using community component `nagyv/gitlab-opencode`:

```yaml
include:
  - component: $CI_SERVER_HOST/nagyv/gitlab-opencode@main
    inputs:
      config_dir: .opencode
      auth_var: OPENCODE_AUTH

variables:
  OPENCODE_AUTH: $OPENCODE_AUTH_JSON
```

### GitLab Duo

Configure `.gitlab/flows/opencode.yaml` with:
- Docker image
- Service account
- API key as CI/CD variable

Usage: `@opencode explain this issue` or `@opencode fix this`

---

## ACP (Agent Client Protocol)

### Zed Configuration

`~/.config/zed/settings.json`:
```json
{
  "agent_servers": {
    "OpenCode": {
      "command": "opencode",
      "args": ["acp"]
    }
  }
}
```

### JetBrains Configuration

`acp.json`:
```json
{
  "command": "/path/to/opencode",
  "args": ["acp"]
}
```

### Neovim (Avante.nvim / CodeCompanion)

Configure adapter to use OpenCode as ACP agent.

### Supported Features via ACP

- All built-in tools
- Custom tools and commands
- MCP servers
- Project rules (AGENTS.md)
- Formatters and linters
- Agents and permissions

Unsupported: `/undo`, `/redo` commands

---

## Network Configuration

### Proxy Settings

```bash
export HTTPS_PROXY=https://proxy.example.com:8080
export HTTP_PROXY=http://proxy.example.com:8080
export NO_PROXY=localhost,127.0.0.1
```

Authenticated proxy:
```bash
export HTTPS_PROXY=http://username:password@proxy.example.com:8080
```

### Custom Certificates

```bash
export NODE_EXTRA_CA_CERTS=/path/to/ca-cert.pem
```

**Important**: Bypass proxy for local server to prevent routing loops.

---

## Enterprise Features

### Data Security

- Code never leaves infrastructure
- No data storage by OpenCode
- Direct API calls to trusted providers
- Share feature can be disabled

### Deployment Options

- Central configuration with SSO integration
- Internal AI gateway routing
- Self-hosted share infrastructure (roadmap)
- Private npm registry support

### Configuration

```json
{
  "share": "disabled"
}
```

### Pricing

- Per-seat licensing
- No token charges for proprietary LLM gateways
- Custom enterprise quotes available

Contact: contact@anoma.ly

---

## Sharing

### Configuration

```json
{
  "share": "manual"    // Default: use /share command
  "share": "auto"      // Auto-share all conversations
  "share": "disabled"  // Disable sharing entirely
}
```

### Commands

- `/share` - Create public URL
- `/unshare` - Revoke access and delete data

### Privacy Considerations

- Shared conversations persist until unshared
- Review content before sharing
- Avoid sharing sensitive/proprietary code
- Disable for sensitive projects

---

## Context Management

### Compaction

```json
{
  "compaction": {
    "auto": true,
    "prune": true
  }
}
```

Commands:
- `/compact` or `/summarize` - Compress session context

### File Watching

```json
{
  "watcher": {
    "ignore": ["node_modules/**", "dist/**", ".git/**"]
  }
}
```

---

## Troubleshooting

### Log Locations

- macOS/Linux: `~/.local/share/opencode/log/`
- Windows: `%USERPROFILE%\.local\share\opencode\log\`

Log files named with timestamps; 10 most recent kept.

### Debug Mode

```bash
opencode --log-level DEBUG --print-logs
```

### Common Issues

**OpenCode Won't Start**
- Check logs for errors
- Run with `--print-logs`
- Update: `opencode upgrade`

**Authentication Issues**
- Re-authenticate via `/connect`
- Verify API keys
- Check network connectivity

**Model Not Available**
- Confirm provider authentication
- Verify format: `<providerId>/<modelId>`
- List models: `opencode models`

**ProviderInitError**
- Review provider configuration
- Clear data: `rm -rf ~/.local/share/opencode`
- Re-authenticate

**API Call Errors**
- Clear cache: `rm -rf ~/.cache/opencode`
- Restart application

**Copy/Paste Issues (Linux)**
- Install: xclip, xsel, or wl-clipboard

---

## Migration to 1.0

### Upgrade

```bash
opencode upgrade 1.0.0
```

### Breaking Changes

Renamed Keybinds:
- `messages_revert` -> `messages_undo`
- `switch_agent` -> `agent_cycle`
- `switch_agent_reverse` -> `agent_cycle_reverse`
- `switch_mode` -> `agent_cycle`
- `switch_mode_reverse` -> `agent_cycle_reverse`

Removed Keybinds:
- messages_layout_toggle, messages_next, messages_previous
- file_diff_toggle, file_search, file_close, file_list
- app_help, project_init, tool_details, thinking_blocks

### New Features

- Compressed session history
- New command bar (Ctrl+P)
- Toggleable session sidebar
- OpenTUI framework (Zig + SolidJS)

---

## Ecosystem

### Community Plugins

Authentication:
- Gemini billing
- Antigravity
- ChatGPT Plus integration

Development:
- TypeScript type injection
- Devcontainer support
- PTY processes
- Dynamic context pruning

Integrations:
- Helicone session tracking
- Wakatime analytics
- Desktop notifications

### Projects

- Discord bot
- Neovim plugins
- Web interfaces
- Vercel AI SDK provider
- VS Code extension
- Obsidian plugin

### Resources

- awesome-opencode repository
- opencode.cafe aggregator
- GitHub: github.com/anomalyco/opencode
- Discord: opencode.ai/discord
