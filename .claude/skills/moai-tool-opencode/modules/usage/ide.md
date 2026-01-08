---
source: https://opencode.ai/docs/ide/
fetched: 2026-01-08
title: IDE Integration
---

# OpenCode IDE Documentation

## Overview

This documentation covers the OpenCode extension for VS Code, Cursor, and other IDEs that support terminal integration.

## Keyboard Shortcuts

| Action | macOS | Windows/Linux |
|--------|-------|---------------|
| Quick launch | `Cmd+Esc` | `Ctrl+Esc` |
| New session | `Cmd+Shift+Esc` | `Ctrl+Shift+Esc` |
| Insert file reference | `Cmd+Option+K` | `Alt+Ctrl+K` |

### File References

The file reference shortcut enables quick insertion like `@File#L37-42`. The tool automatically shares current selection or active tab with OpenCode.

## Installation

### Automatic Installation

The extension installs automatically when running the `opencode` command in VS Code's integrated terminal.

### Manual Installation

1. Open the Extension Marketplace in your IDE
2. Search for "OpenCode"
3. Click Install

## Supported IDEs

- VS Code
- Cursor
- Windsurf
- Codium

## Troubleshooting

If automatic installation fails:

1. **Run in integrated terminal:** Run `opencode` specifically in the integrated terminal (not an external terminal)

2. **Verify CLI tools:** Ensure IDE CLI tools are installed:
   - `code` (VS Code)
   - `cursor` (Cursor)
   - `windsurf` (Windsurf)
   - `codium` (Codium)

3. **Install shell commands:** Use command palette (`Cmd+Shift+P` or `Ctrl+Shift+P`) to install shell commands in PATH

4. **Check permissions:** Confirm VS Code extension installation permissions are enabled

## Editor Configuration

For using custom editors with `/editor` or `/export` commands, set the `EDITOR` environment variable:

```bash
export EDITOR="code --wait"
```

The `--wait` flag ensures the editor blocks until the file is closed.

## Usage Tips

- Use file references to quickly share code context
- The extension automatically detects active selections
- Works with multiple IDE instances simultaneously
