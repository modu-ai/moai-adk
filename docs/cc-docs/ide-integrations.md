# Add Claude Code to your IDE

Claude Code works with any IDE that has a terminal. Simply run `claude`, and you're ready to go.

## Supported IDEs

Claude Code provides dedicated integrations for:

- **Visual Studio Code** (including forks like Cursor, Windsurf, and VSCodium)
- **JetBrains IDEs** (including IntelliJ, PyCharm, Android Studio, WebStorm, PhpStorm, and GoLand)

## Features

- **Quick launch**: Use `Cmd+Esc` (Mac) or `Ctrl+Esc` (Windows/Linux) to open Claude Code
- **Diff viewing**: Code changes displayed in IDE diff viewer
- **Selection context**: Current selection/tab automatically shared with Claude Code
- **File reference shortcuts**:
  - Mac: `Cmd+Option+K`
  - Windows/Linux: `Alt+Ctrl+K`
- **Diagnostic sharing**: IDE errors automatically shared with Claude

## Installation

### VS Code and Forks

1. Open VS Code
2. Open integrated terminal
3. Run `claude` - extension will auto-install

### JetBrains IDEs

1. Find and install Claude Code plugin from marketplace
2. Restart IDE

**Note for Remote Development**:

- In JetBrains Remote Development, install plugin in remote host via `Settings > Plugin (Host)`

## Usage

### From IDE

Run `claude` in the integrated terminal to activate all features.

### From External Terminals

Use `/ide` command to connect Claude Code to your IDE.

## Configuration

1. Run `claude`
2. Enter `/config` command
3. Adjust preferences (set diff tool to `auto` for automatic IDE detection)

## Troubleshooting

### VS Code Extension Issues

- Ensure running from VS Code's integrated terminal
- Verify CLI is available (`code`, `cursor`, `windsurf`, `codium`)
- Check VS Code extension installation permissions

### JetBrains Plugin Issues

- Run from project root directory
- Ensure plugin is enabled in IDE settings
- Completely restart IDE
- For remote development, install plugin on remote host
