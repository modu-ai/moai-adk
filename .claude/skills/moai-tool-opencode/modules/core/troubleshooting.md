---
source: https://opencode.ai/docs/troubleshooting/
fetched: 2026-01-08
title: Troubleshooting
---

# OpenCode Troubleshooting Guide

## Overview

OpenCode provides debugging tools through logs and local storage to help resolve issues.

## Logs Location

Log files are stored at:

| Platform | Path |
|----------|------|
| macOS/Linux | `~/.local/share/opencode/log/` |
| Windows | `%USERPROFILE%\.local\share\opencode\log\` |

Files use timestamp naming (e.g., `2025-01-09T123456.log`) with the 10 most recent retained.

### Enable Detailed Logging

```bash
opencode --log-level DEBUG
```

## Storage Location

Application data resides at:

| Platform | Path |
|----------|------|
| macOS/Linux | `~/.local/share/opencode/` |
| Windows | `%USERPROFILE%\.local\share\opencode` |

### Contents

- `auth.json` - API keys and OAuth tokens
- `log/` - Application logs
- `project/` - Session and message data
  - `./<project-slug>/storage/` for Git repos
  - `./global/storage/` otherwise

## Getting Help

### GitHub Issues

Report bugs at github.com/anomalyco/opencode/issues

### Discord Community

Real-time assistance available at opencode.ai/discord

## Common Issues & Solutions

### Won't Start

- Check logs
- Run with `--print-logs`
- Verify latest version via `opencode upgrade`

```bash
opencode --print-logs
opencode upgrade
```

### Authentication Problems

- Use `/connect` command in TUI
- Validate API keys
- Confirm network access to provider APIs

### Model Unavailable

- Confirm provider authentication
- Verify correct model naming format (`<providerId>/<modelId>`)
- Use `opencode models` command to check access

```bash
opencode models
```

### ProviderInitError

- Verify provider setup
- Clear configuration via `rm -rf ~/.local/share/opencode`
- Re-authenticate

```bash
rm -rf ~/.local/share/opencode
opencode
/connect
```

### API Call Errors

- Clear provider cache with `rm -rf ~/.cache/opencode`
- Restart to reinstall latest packages

```bash
rm -rf ~/.cache/opencode
opencode
```

### Linux Copy/Paste Issues

Install clipboard utilities based on your environment:

| Environment | Required Package |
|-------------|------------------|
| X11 | `xclip` or `xsel` |
| Wayland | `wl-clipboard` |
| Headless | `xvfb` |

```bash
# For X11
sudo apt install xclip

# For Wayland
sudo apt install wl-clipboard
```
