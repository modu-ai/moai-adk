---
source: https://opencode.ai/docs/themes/
fetched: 2026-01-08
title: Themes Configuration
---

# OpenCode Themes Documentation

## Overview

OpenCode allows users to select and customize themes. Users can choose from built-in themes, use a system-adaptive theme, or create custom ones.

## Terminal Requirements

Your terminal must support **truecolor** (24-bit color) for proper theme display.

### Verify Support

```bash
echo $COLORTERM
```

Should output `truecolor` or `24bit`.

### Enable Truecolor

Add to your shell profile:

```bash
export COLORTERM=truecolor
```

### Supported Terminals

Modern terminals with truecolor support:
- iTerm2
- Alacritty
- Kitty
- Windows Terminal
- WezTerm
- Ghostty

## Built-in Themes

| Theme | Description |
|-------|-------------|
| `system` | Adapts to terminal background |
| `tokyonight` | Based on popular editor theme |
| `everforest` | Nature-inspired color palette |
| `ayu` | Clean and modern design |
| `catppuccin` | Soothing pastel theme |
| `catppuccin-macchiato` | Catppuccin variant |
| `gruvbox` | Retro groove color scheme |
| `kanagawa` | Japanese-inspired palette |
| `nord` | Arctic, north-bluish palette |
| `matrix` | Green-on-black hacker style |
| `one-dark` | Atom-inspired theme |

## System Theme

The system theme automatically adapts to your terminal's color scheme:
- Generates custom gray scales
- Uses ANSI colors (0-15)
- Preserves terminal defaults using `none` values

## Configuration

### Using /theme Command

```bash
/themes
```

Select from available themes interactively.

### Config File

```json
{
  "$schema": "https://opencode.ai/config.json",
  "theme": "tokyonight"
}
```

## Custom Theme Creation

### Directory Hierarchy (Priority Order)

1. Built-in themes (highest priority)
2. `~/.config/opencode/themes/*.json` (global)
3. `./.opencode/themes/*.json` (project root)
4. `./.opencode/themes/*.json` (current directory)

### JSON Format

Custom themes support:
- Hex colors: `#ffffff`
- ANSI values: `0-255`
- Color references: Reference defined colors
- Dark/light variants
- `"none"`: Terminal defaults

### Example Theme Structure

```json
{
  "$schema": "https://opencode.ai/theme.json",
  "defs": {
    "nord0": "#2E3440",
    "nord1": "#3B4252",
    "nord2": "#434C5E",
    "nord3": "#4C566A",
    "nord4": "#D8DEE9",
    "nord5": "#E5E9F0",
    "nord6": "#ECEFF4",
    "nord7": "#8FBCBB",
    "nord8": "#88C0D0",
    "nord9": "#81A1C1",
    "nord10": "#5E81AC",
    "nord11": "#BF616A",
    "nord12": "#D08770",
    "nord13": "#EBCB8B",
    "nord14": "#A3BE8C",
    "nord15": "#B48EAD"
  },
  "theme": {
    "primary": {
      "dark": "nord8",
      "light": "nord10"
    },
    "secondary": {
      "dark": "nord9",
      "light": "nord9"
    },
    "accent": {
      "dark": "nord7",
      "light": "nord7"
    },
    "error": {
      "dark": "nord11",
      "light": "nord11"
    },
    "warning": {
      "dark": "nord13",
      "light": "nord13"
    },
    "success": {
      "dark": "nord14",
      "light": "nord14"
    },
    "text": {
      "dark": "nord4",
      "light": "nord0"
    },
    "background": {
      "dark": "nord0",
      "light": "nord6"
    },
    "border": {
      "dark": "nord3",
      "light": "nord4"
    },
    "diff": {
      "add": {
        "dark": "nord14",
        "light": "nord14"
      },
      "remove": {
        "dark": "nord11",
        "light": "nord11"
      }
    },
    "syntax": {
      "keyword": "nord9",
      "string": "nord14",
      "number": "nord15",
      "comment": "nord3",
      "function": "nord8",
      "variable": "nord4"
    }
  }
}
```

### Theme Properties

| Property | Description |
|----------|-------------|
| `primary` | Main accent color |
| `secondary` | Secondary accent |
| `accent` | Highlight color |
| `error` | Error messages |
| `warning` | Warning messages |
| `success` | Success messages |
| `text` | Default text |
| `background` | Background color |
| `border` | Border color |
| `diff.add` | Added lines in diffs |
| `diff.remove` | Removed lines in diffs |
| `syntax.*` | Syntax highlighting colors |
