---
title: Installation
weight: 30
draft: false
---

Learn how to install MoAI-ADK 2.x on your system.

## Prerequisites

Verify the following before installation:

### 1. Claude Code

MoAI-ADK is an extension framework that runs on top of Claude Code. Claude Code must be installed first.

```bash
claude --version
```

If not yet installed, refer to the [Claude Code official documentation](https://docs.anthropic.com/en/docs/claude-code).

### 2. Git (Required)

MoAI-ADK uses Git-based workflows. Git must be installed on your system.

```bash
git --version
```

{{< callout type="warning" >}}
**Windows Users**: You must use **Git Bash** or **WSL**. Command Prompt (cmd.exe) is not supported.

If Git is not installed:
- **Windows**: Install Git for Windows from [git-scm.com](https://git-scm.com). Git Bash is included.
- **macOS**: `xcode-select --install` or [git-scm.com](https://git-scm.com)
- **Linux**: `sudo apt install git` (Ubuntu/Debian) or `sudo dnf install git` (Fedora)
{{< /callout >}}

### System Requirements

| Item | Requirement |
|------|------------|
| **OS** | macOS, Linux, Windows (Git Bash / WSL) |
| **Architecture** | amd64, arm64 |
| **Memory** | Minimum 4GB RAM |
| **Disk** | Minimum 100MB free space |

## Installation Methods

### Method 1: Quick Install (Recommended)

Install the latest version automatically with a single command.

**macOS / Linux / WSL / Git Bash:**

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

**Windows (PowerShell):**

```powershell
irm https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.ps1 | iex
```

{{< callout type="info" >}}
The install script automatically detects your platform, downloads the prebuilt binary from GitHub, verifies the SHA256 checksum, and configures PATH. No Python or separate runtime is required.
{{< /callout >}}

After installation, verify:

```bash
moai version
```

#### Install Options

```bash
# Install a specific version
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash -s -- --version 2.0.0

# Install to a custom directory
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash -s -- --install-dir /usr/local/bin
```

### Method 2: Build from Source

If you have a Go development environment, you can build from source.

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk
make build
```

The built binary will be at `./bin/moai`. Copy it to a directory in your PATH:

```bash
cp ./bin/moai ~/.local/bin/
```

### Install Locations

The install script determines the installation directory in this order:

| Platform | Priority |
|----------|---------|
| **macOS / Linux** | `$GOBIN` → `$GOPATH/bin` → `~/.local/bin` |
| **Windows** | `%LOCALAPPDATA%\Programs\moai` |

## Migrating from 1.x

{{< callout type="error" >}}
**MoAI-ADK 1.x (Python version) users must uninstall the old version first.**

Both 1.x and 2.x use the same `moai` command, so keeping the old version will cause conflicts.
{{< /callout >}}

### Step 1: Remove existing 1.x

```bash
# If installed via uv
uv tool uninstall moai-adk

# If installed via pip
pip uninstall moai-adk
```

### Step 2: Backup existing config (optional)

```bash
# If you want to back up existing settings
cp -r ~/.moai ~/.moai-v1-backup
```

### Step 3: Install 2.x

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

### Step 4: Verify installation

```bash
moai version
# Example output: moai v2.x.x (commit: abc1234, built: 2026-01-15)
```

{{< callout type="info" >}}
Version 2.x is a single Go binary with no Python runtime or virtual environment required. Startup time has improved dramatically from ~800ms to ~5ms.
{{< /callout >}}

## WSL Support

Guide for installing and using MoAI-ADK in WSL (Windows Subsystem for Linux) on Windows.

### Installing WSL

If WSL is not installed, run the following command in PowerShell (Administrator):

```powershell
wsl --install
```

After installation, restart Windows and Ubuntu will be automatically installed.

### Installing MoAI-ADK in WSL

Use the same command as Linux in the WSL terminal:

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

### Path Handling

Distinguish between Windows paths and WSL paths:

| Windows Path | WSL Path |
|-------------|----------|
| `C:\Users\name\project` | `/mnt/c/Users/name/project` |
| `D:\Projects\myapp` | `/mnt/d/Projects/myapp` |

{{< callout type="info" >}}
**Recommended**: Create projects in WSL's Linux filesystem (`~/projects/`) for 2-5x better I/O performance. Accessing the Windows filesystem (`/mnt/c/`) may result in slower performance.
{{< /callout >}}

### WSL Best Practices

1. **Use Linux filesystem**: Create projects in `~/projects/` directory
2. **Configure Git credentials**: Set up Git credentials separately in WSL from Windows
3. **Recommended terminal**: Use Windows Terminal to manage multiple WSL distributions

### WSL Troubleshooting

#### PATH Not Loaded

```bash
# Add to ~/.bashrc or ~/.zshrc
source ~/.cargo/env
export PATH="$HOME/.local/bin:$PATH"
```

#### Hook/MCP Server Permission Issues

```bash
# Grant execute permissions
chmod +x ~/.claude/hooks/moai/*.sh
```

#### Slow Windows Path Access

Move the project to the Linux filesystem:

```bash
# Move from Windows to WSL
cp -r /mnt/c/Users/name/project ~/projects/
cd ~/projects/project
```

## pip and uv Tool Conflict

A common issue for MoAI-ADK 1.x (Python version) users.

### Problem Description

pip and uv install packages in different locations. Using both tools interchangeably may cause the `moai` command to execute an unexpected version.

### Symptoms

- `moai version` shows 1.x version
- `command not found: moai` error
- `which moai` shows a different path than expected

### Root Cause

1. pip installs to system Python paths
2. uv tool installs to `~/.local/bin` or `~/.cargo/bin`
3. PATH order determines which version runs

### Solutions

#### Clean Reinstall

```bash
# 1. Remove all existing versions
uv tool uninstall moai-adk 2>/dev/null || true
pip uninstall moai-adk -y 2>/dev/null || true

# 2. Check and remove remaining binaries
which moai && rm $(which moai) 2>/dev/null || true
ls ~/.local/bin/moai && rm ~/.local/bin/moai 2>/dev/null || true

# 3. Install 2.x
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash

# 4. Verify
moai version
```

#### Update Shell Configuration

```bash
# Add to ~/.bashrc or ~/.zshrc
export PATH="$HOME/.local/bin:$PATH"

# Apply settings
source ~/.bashrc  # or source ~/.zshrc
```

### Prevention Tips

1. MoAI-ADK 2.x is a Python-independent Go binary
2. Uninstall 1.x (Python version) before installing 2.x
3. Do not use pip and uv tool simultaneously

## Troubleshooting

### Problem: Command Not Found

```bash
command not found: moai
```

**Solution:**

1. Restart your terminal
2. Check your PATH:

```bash
echo $PATH
```

3. Verify the binary location:

```bash
which moai || ls ~/.local/bin/moai
```

4. Manually add to PATH:

```bash
# Bash/Zsh
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Problem: Permission Denied

```bash
Permission denied
```

**Solution:**

```bash
chmod +x ~/.local/bin/moai
```

### Problem: 1.x and 2.x Conflict

If the old version of `moai` is being executed:

```bash
# Check which moai is running
which moai

# Remove 1.x if still present
uv tool uninstall moai-adk
# or
pip uninstall moai-adk

# Restart terminal and verify 2.x
moai version
```

## Next Steps After Installation

Once installed, initialize your project:

### Create a New Project

```bash
moai init my-project
```

### Apply to Existing Project

```bash
cd my-existing-project
moai init
```

## Upgrade

To upgrade to the latest version:

```bash
moai update
```

### Update Options

```bash
# Check version only (no update)
moai update --check

# Synchronize templates only (skip package upgrade)
moai update --templates-only

# Config edit mode (re-run init wizard)
moai update --config
moai update -c

# Force update without backup
moai update --force

# Auto-approve mode (auto-approve all confirmations)
moai update --yes
```

### Merge Strategy

```bash
# Force auto-merge (default)
moai update --merge

# Force manual merge
moai update --manual
```

{{< callout type="info" >}}
**Automatically Preserved Items**: User settings, custom agents, custom commands, custom skills, custom hooks, SPEC documents, and reports are automatically preserved during updates.
{{< /callout >}}

See the [Update Guide](https://adk.mo.ai.kr/getting-started/update) for details.

## Uninstall

To completely remove MoAI-ADK:

```bash
# Remove the binary
rm $(which moai)

# Remove config directory (optional)
rm -rf ~/.moai
```

---

## Next Steps

Learn how to configure MoAI-ADK in the [Initial Setup Wizard](./init-wizard).
