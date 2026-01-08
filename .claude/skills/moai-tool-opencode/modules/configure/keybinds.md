---
source: https://opencode.ai/docs/keybinds/
fetched: 2026-01-08
title: Keybinds Configuration
---

# OpenCode Keybinds Documentation

## Overview

OpenCode allows customization of keyboard shortcuts through the configuration file.

## Leader Key Concept

OpenCode implements a leader key system, defaulting to `ctrl+x`, which users press before the actual command shortcut.

This design "avoids conflicts in your terminal."

## Configuration Method

Customize keybinds through the `opencode.json` file using the `"keybinds"` object:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "keybinds": {
    "leader": "ctrl+x",
    "submit": "return",
    "cancel": "escape"
  }
}
```

## Notable Default Bindings

| Action | Default Keybind |
|--------|-----------------|
| Leader key | `ctrl+x` |
| New session | `<leader>n` |
| Model list | `<leader>m` |
| Agent list | `<leader>a` |
| Command list | `ctrl+p` |
| Input submission | `return` |
| Add newline | `shift+return`, `ctrl+return`, `alt+return`, `ctrl+j` |

## Common Keybinds

| Action | Keybind |
|--------|---------|
| `session_new` | `<leader>n` |
| `session_list` | `<leader>l` |
| `model_list` | `<leader>m` |
| `agent_list` | `<leader>a` |
| `command_list` | `ctrl+p` |
| `help` | `<leader>h` |
| `compact` | `<leader>c` |
| `details` | `<leader>d` |
| `editor` | `<leader>e` |
| `export` | `<leader>x` |
| `share` | `<leader>s` |
| `themes` | `<leader>t` |
| `undo` | `<leader>u` |
| `redo` | `<leader>r` |
| `quit` | `<leader>q` |

## Disabling Bindings

Set any keybind value to `"none"` to disable it:

```json
{
  "keybinds": {
    "share": "none",
    "export": "none"
  }
}
```

## Built-in Readline/Emacs Shortcuts

The desktop prompt supports standard text editing shortcuts. These are **not configurable** through `opencode.json`:

| Shortcut | Action |
|----------|--------|
| `ctrl+a` | Move to line start |
| `ctrl+e` | Move to line end |
| `ctrl+w` | Delete word backward |
| `ctrl+u` | Delete to line start |
| `ctrl+k` | Delete to line end |
| `ctrl+b` | Move backward one character |
| `ctrl+f` | Move forward one character |
| `alt+b` | Move backward one word |
| `alt+f` | Move forward one word |

## Platform-Specific Configuration

### Windows Terminal

Users may need to manually configure escape sequence handling for `Shift+Enter` functionality.

Add to Windows Terminal `settings.json`:

```json
{
  "actions": [
    {
      "command": {
        "action": "sendInput",
        "input": "\u001b[13;2u"
      },
      "keys": "shift+enter"
    }
  ]
}
```

## Custom Keybind Examples

### Minimal Configuration

```json
{
  "keybinds": {
    "submit": "ctrl+enter",
    "cancel": "ctrl+c"
  }
}
```

### Vim-Style Configuration

```json
{
  "keybinds": {
    "leader": "space",
    "session_new": "<leader>n",
    "session_list": "<leader>b"
  }
}
```

## Full Keybind Reference

OpenCode provides 80+ configurable actions. Use `/help` or check the documentation for the complete list.

## Best Practices

1. **Keep leader key default**: `ctrl+x` avoids most conflicts
2. **Test in your terminal**: Some terminals intercept certain key combinations
3. **Document custom bindings**: Keep a reference for team consistency
