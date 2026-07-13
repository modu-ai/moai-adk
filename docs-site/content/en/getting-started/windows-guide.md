---
title: Windows Guide
weight: 40
draft: false
---

This page collects the environment requirements and common pitfalls you should know when using MoAI-ADK on Windows. Bottom line first: **WSL is the most comfortable option** — most of the path and permission issues you hit in a native Windows environment simply do not occur under WSL.

## Supported Environments

| Environment | Supported | Notes |
|------|----------|------|
| **WSL (recommended)** | {{< icon check ok >}} Fully supported | Best experience |
| **PowerShell 7.x+** | {{< icon check ok >}} Supported | Alternative environment |
| PowerShell 5.x (legacy) | {{< icon x danger >}} Not supported | Windows PowerShell |
| cmd.exe | {{< icon x danger >}} Not supported | Command Prompt |

**Required:**
- [Git for Windows](https://gitforwindows.org/) must be installed
- WSL or PowerShell 7.x or later

## Installation

### WSL (Recommended)

WSL provides a Linux environment on Windows and fully supports every MoAI-ADK feature.

```bash
# Install WSL (run in an administrator PowerShell)
wsl --install

# Install MoAI-ADK inside WSL
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh \
  | bash
```

### PowerShell 7.x+

> **Note**: For the best experience, WSL is recommended.

```powershell
irm https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.ps1 | iex
```

## Non-ASCII Username Path Errors

### Symptom

When a Windows username contains non-ASCII characters such as Korean or Chinese, an `EINVAL` error may occur. This is caused by Windows' 8.3 short-filename conversion.

```
Error: EINVAL: invalid argument, open 'C:\Users\홍길동\AppData\Local\Temp\...'
```

### Solution 1: Set an Alternative Temp Directory (Recommended)

Create a temp directory on a path containing only ASCII characters:

```bash
# Command Prompt
set MOAI_TEMP_DIR=C:\temp
mkdir C:\temp 2>/dev/null
```

```powershell
# PowerShell
$env:MOAI_TEMP_DIR="C:\temp"
New-Item -ItemType Directory -Path "C:\temp" -Force
```

To make the environment variable permanent, add `MOAI_TEMP_DIR` to your system environment variables.

### Solution 2: Disable 8.3 Filename Generation

Run with administrator privileges:

```bash
fsutil 8dot3name set 1
```

> **Caution**: This setting affects the entire system. Some legacy programs may be affected.

### Solution 3: Create an ASCII User Account

Creating a new Windows user account with an English name fixes the path problem at its root.

## WSL Setup Guide

### Installing WSL

```powershell
# Run in an administrator PowerShell
wsl --install

# Default distribution: Ubuntu (recommended)
# After restarting, set your username and password
```

### Accessing Project Files

Accessing Windows files from WSL:

```bash
# Access the Windows filesystem
cd /mnt/c/Users/<username>/projects/

# Use the WSL native filesystem (faster)
cd ~/projects/
```

> **Performance tip**: Working in the WSL native filesystem (under `~/`) gives you optimal performance with no cross-filesystem overhead.

### VS Code Integration

1. Install the [WSL extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-wsl) in VS Code
2. Run `code .` from a WSL terminal
3. VS Code opens in WSL mode automatically

## Using tmux in CG Mode

[CG Mode](/en/multi-llm/cg-mode) requires tmux. Install it in WSL:

```bash
# Ubuntu/Debian
sudo apt install tmux

# Start a tmux session
tmux new -s moai

# Run CG mode
moai cg
```

## Troubleshooting

| Problem | Cause | Solution |
|------|------|------|
| `moai: command not found` | Go bin directory not on PATH | Add `export PATH="$HOME/go/bin:$PATH"` to `.bashrc` |
| `EINVAL` error | Non-ASCII username | See [Non-ASCII Username Path Errors](#non-ascii-username-path-errors) above |
| Permission denied | Install script permissions | Run `chmod +x install.sh` and retry |
| Git commands fail | Git for Windows not installed | Install [Git for Windows](https://gitforwindows.org/) |
| tmux missing | CG mode cannot run | `sudo apt install tmux` (in WSL) |

## Next Steps

- [Installation](/en/getting-started/installation) — Detailed installation guide
- [Initial Setup](/en/getting-started/init-wizard) — Project initialization
- [CG Mode](/en/multi-llm/cg-mode) — Claude + GLM hybrid mode
