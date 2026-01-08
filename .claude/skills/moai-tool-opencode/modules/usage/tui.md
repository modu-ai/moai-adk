---
source: https://opencode.ai/docs/tui/
fetched: 2026-01-08
title: Terminal User Interface (TUI)
---

# OpenCode TUI Documentation

## Overview

OpenCode provides an interactive terminal interface (TUI) for working on projects with an LLM. Launch it by running `opencode` in your project directory or specify a path:

```bash
opencode
# or
opencode /path/to/project
```

## File References

Users can reference files in messages using `@` to perform fuzzy file searching within the current working directory.

**Example:**
```
How is auth handled in @packages/functions/src/api/index.ts?
```

The file content is automatically included in conversations.

## Bash Commands

Messages beginning with `!` execute shell commands. The command output is then added to the conversation as a tool result.

**Example:**
```
!npm test
```

## Commands (Slash Commands)

All commands use `/` prefix and most include `ctrl+x` keybind shortcuts (ctrl+x is the default leader key):

| Command | Aliases | Description | Keybind |
|---------|---------|-------------|---------|
| `/connect` | - | Add a provider and configure API keys | - |
| `/compact` | `/summarize` | Compact current session | `ctrl+x c` |
| `/details` | - | Toggle tool execution details | `ctrl+x d` |
| `/editor` | - | Open external editor for message composition | `ctrl+x e` |
| `/exit` | `/quit`, `/q` | Exit OpenCode | `ctrl+x q` |
| `/export` | - | Export conversation to Markdown | `ctrl+x x` |
| `/help` | - | Show help dialog | `ctrl+x h` |
| `/init` | - | Create/update AGENTS.md file | `ctrl+x i` |
| `/models` | - | List available models | `ctrl+x m` |
| `/new` | `/clear` | Start new session | `ctrl+x n` |
| `/redo` | - | Redo previously undone message | `ctrl+x r` |
| `/sessions` | `/resume`, `/continue` | List/switch sessions | `ctrl+x l` |
| `/share` | - | Share current session | `ctrl+x s` |
| `/themes` | - | List available themes | `ctrl+x t` |
| `/undo` | - | Undo last message and revert file changes | `ctrl+x u` |
| `/unshare` | - | Unshare current session | - |

## Editor Setup

The `EDITOR` environment variable controls which editor opens for `/editor` and `/export` commands.

### Linux/macOS

```bash
export EDITOR=nano
export EDITOR=vim
export EDITOR="code --wait"
```

### Windows CMD

```cmd
set EDITOR=notepad
set EDITOR="code --wait"
```

### Windows PowerShell

```powershell
$env:EDITOR = "notepad"
$env:EDITOR = "code --wait"
```

### Popular Editors

| Editor | Command |
|--------|---------|
| VS Code | `code` |
| Cursor | `cursor` |
| Windsurf | `windsurf` |
| Neovim | `nvim` |
| Vim | `vim` |
| Nano | `nano` |
| Sublime | `subl` |

**Note:** Some editors like VS Code need to be started with the `--wait` flag to block until closed.

## Configuration

```json
{
  "$schema": "https://opencode.ai/config.json",
  "tui": {
    "scroll_speed": 3,
    "scroll_acceleration": {
      "enabled": true
    }
  }
}
```

### Options

| Option | Description |
|--------|-------------|
| `scroll_acceleration` | Enable macOS-style scroll acceleration; takes precedence over scroll_speed |
| `scroll_speed` | Controls scroll speed (minimum: 1); defaults to 1 on Unix, 3 on Windows; ignored if scroll_acceleration is enabled |

## Customization

Users can customize TUI view settings through the command palette (`ctrl+x h` or `/help`). Settings persist across restarts.

### Username Display

Toggle whether your username appears in chat messages through the command palette by searching "username" or "hide username."
