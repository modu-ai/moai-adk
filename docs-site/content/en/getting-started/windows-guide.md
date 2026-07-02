---
title: Windows Guide
weight: 40
draft: false
---

## Supported Environments

| Environment | Supported | Notes |
|------|----------|------|
| **WSL (recommended)** | ✅ Fully supported | Best experience |
| **PowerShell 7.x+** | ✅ Supported | Alternative environment |
| PowerShell 5.x (legacy) | ❌ Not supported | Windows PowerShell |
| cmd.exe | ❌ Not supported | Command Prompt |

**Requirements:**
- [Git for Windows](https://gitforwindows.org/) must be installed
- WSL or PowerShell 7.x or later

## Installation

### WSL (recommended)

WSL provides a Linux environment on Windows and fully supports every MoAI-ADK feature.

```bash
# Install WSL (run from an administrator PowerShell)
wsl --install

# Install MoAI-ADK inside WSL
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh \
  | bash
```

### PowerShell 7.x+

> **Note**: WSL is recommended for the best experience.

```powershell
irm https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.ps1 | iex
```

## Non-ASCII Username Path Error

### Symptom

If the Windows username contains non-ASCII characters such as Korean or Chinese, an `EINVAL` error can occur. This is caused by Windows' 8.3 short-filename conversion process.

```
Error: EINVAL: invalid argument, open 'C:\Users\李明\AppData\Local\Temp\...'
```

### Fix 1: Set an alternate temp directory (recommended)

Create a temp directory at a path that contains only ASCII characters:

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

To set the environment variable permanently, add `MOAI_TEMP_DIR` to your system environment variables.

### Fix 2: Disable 8.3 filename generation

Run as administrator:

```bash
fsutil 8dot3name set 1
```

> **Caution**: This setting affects the entire system. Some legacy programs may be affected.

### Fix 3: Create an ASCII user account

Creating a new Windows user account with an English name resolves the path issue at the root.

## WSL Setup Guide

### Installing WSL

```powershell
# Run from an administrator PowerShell
wsl --install

# Default distribution: Ubuntu (recommended)
# Set your username and password after the restart
```

### Accessing Project Files

Accessing Windows files from WSL:

```bash
# Access the Windows filesystem
cd /mnt/c/Users/username/projects/

# Use the WSL-native filesystem (faster)
cd ~/projects/
```

> **Performance tip**: Working from the WSL-native filesystem (under `~/`) gives you the best performance, with no cross-filesystem overhead.

### VS Code Integration

1. Install the [WSL extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-wsl) in VS Code
2. Run `code .` from a WSL terminal
3. VS Code opens automatically in WSL mode

## Using tmux in CG Mode

[CG mode](/multi-llm/cg-mode) requires tmux. Install it inside WSL:

```bash
# Ubuntu/Debian
sudo apt install tmux

# Start a tmux session
tmux new -s moai

# Run CG mode
moai cg
```

## Troubleshooting

| Issue | Cause | Fix |
|------|------|------|
| `moai: command not found` | Go bin directory not in PATH | Add `export PATH="$HOME/go/bin:$PATH"` to `.bashrc` |
| `EINVAL` error | Non-ASCII username | See [Non-ASCII Username Path Error](#non-ascii-username-path-error) above |
| Permission denied | Install script permissions | Run `chmod +x install.sh`, then retry |
| Git command fails | Git for Windows not installed | Install [Git for Windows](https://gitforwindows.org/) |
| tmux missing | Cannot run CG mode | `sudo apt install tmux` (inside WSL) |

## Next Steps

- [Installation](/getting-started/installation) — Detailed installation guide
- [Initial Setup](/getting-started/init-wizard) — Project initialization
- [CG Mode](/multi-llm/cg-mode) — Claude + GLM hybrid mode
