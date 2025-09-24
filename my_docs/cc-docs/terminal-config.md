# Optimize your terminal setup

Claude Code works best when your terminal is properly configured. Follow these guidelines to optimize your experience.

## Themes and appearance

Claude cannot control the theme of your terminal. That's handled by your terminal application.

You can match Claude Code's theme to your terminal via the `/config` command.

For additional customization, you can configure a custom status line to display contextual information like:

- Current model
- Working directory
- Git branch

## Line breaks

You have several options for entering linebreaks into Claude Code:

- **Quick escape**: Type `\` followed by Enter to create a newline
- **Keyboard shortcut**: Set up a keybinding to insert a newline

### Set up Shift+Enter (VS Code or iTerm2)

Run `/terminal-setup` within Claude Code to automatically configure Shift+Enter.

### Set up Option+Enter (VS Code, iTerm2 or macOS Terminal.app)

**For Mac Terminal.app:**

1. Open Settings → Profiles → Keyboard
2. Check "Use Option as Meta Key"

**For iTerm2 and VS Code terminal:**

1. Open Settings → Profiles → Keys
2. Under General, set Left/Right Option key to "Esc+"

## Notification setup

### Terminal bell notifications

Enable sound alerts when tasks complete:

```sh
claude config set --global preferredNotifChannel terminal_bell
```

**For macOS users**: Enable notification permissions in System Settings.

### iTerm 2 system notifications

For iTerm 2 alerts when tasks complete:

1. Open iTerm 2 Preferences
2. Navigate to Profiles → Terminal
3. Enable "Silence bell" and Filter Alerts
4. Set your preferred notification delay

### Custom notification hooks

Create notification hooks for advanced notification handling.

## Handling large inputs

- **Avoid direct pasting**: Claude Code may struggle with very long pasted content
- **Use file-based workflows**: Write content to a file and ask Claude to read it
- **Be aware of VS Code limitations**: The VS Code terminal can truncate large pastes
