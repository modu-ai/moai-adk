---
source: https://opencode.ai/docs/1-0/
fetched: 2026-01-08
title: Migrating to 1.0
---

# OpenCode 1.0 Migration Guide

## Key Overview

OpenCode 1.0 represents "a complete rewrite of the TUI," transitioning from a Go/Bubbletea framework to an in-house solution built with Zig and SolidJS. The new version maintains compatibility with the existing OpenCode server while improving performance and capabilities.

## Upgrade Instructions

### Manual Upgrade

```bash
opencode upgrade 1.0.0
```

### Revert to Previous Version

```bash
opencode upgrade 0.15.31
```

**Note:** Some older versions automatically fetch the latest release.

## Notable UX Improvements

### Command Bar

The interface now includes "a command bar which almost everything flows through," accessible via `Ctrl+P`.

### Session History

Session history display has been simplified to focus on edit and bash tool details.

### Session Sidebar

A toggleable session sidebar provides additional context.

## Breaking Changes

### Renamed Keybinds

Four keybinds were renamed:

| Old Name | New Name |
|----------|----------|
| `messages_revert` | `messages_undo` |
| `switch_agent` | `agent_cycle` |
| `switch_agent_reverse` | `agent_cycle_reverse` |
| `switch_mode` | `agent_cycle` |

### Removed Keybinds

Twelve keybinds were removed entirely, including functionality for:

- Layout toggling
- File operations
- Help display
- Thinking block management

## Feedback

The documentation encourages users to report missing features via GitHub issues.

## Technical Details

### Framework Changes

- **Previous:** Go/Bubbletea
- **Current:** Zig and SolidJS (in-house solution)

### Compatibility

The new version maintains compatibility with the existing OpenCode server.

### Performance

The rewrite focuses on improving performance and capabilities.
