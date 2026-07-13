---
title: Installation
weight: 30
draft: false
---

This guide covers installing MoAI-ADK on your system. The install artifact is a single binary built with Go — no Python, no virtual environment, no package manager needed.

## License

MoAI-ADK {{< version >}} and later is distributed under the **Apache-2.0 license**.

Commercial use, modification, and distribution are free, with no obligation to publish source code. See the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0) for details.

{{< callout type="info" >}}
**Note**: MoAI-ADK 1.x (the Python version) was GPL-3.0 licensed. Starting with v2.0.0, it was rewritten in Go and relicensed under Apache-2.0.
{{< /callout >}}

## Prerequisites

Check the following before installing:

### 1. Claude Code

MoAI-ADK is an extension framework that runs on top of Claude Code. Claude Code must be installed first.

```bash
claude --version
```

If you have not installed it yet, see the [official Claude Code documentation](https://docs.anthropic.com/en/docs/claude-code).

### 2. Git (Required)

MoAI-ADK uses a Git-based workflow. Git must be installed on your system.

```bash
git --version
```

{{< callout type="warning" >}}
**Windows users**: You must use a **Git Bash** or **WSL** environment. Command Prompt (cmd.exe) is not supported.

If Git is not installed:
- **Windows**: Install Git for Windows from [git-scm.com](https://git-scm.com). Git Bash is installed along with it.
- **macOS**: `xcode-select --install` or [git-scm.com](https://git-scm.com)
- **Linux**: `sudo apt install git` (Ubuntu/Debian) or `sudo dnf install git` (Fedora)
{{< /callout >}}

### System Requirements

| Item | Requirement |
|------|---------|
| **Operating system** | macOS, Linux, Windows (Git Bash / WSL) |
| **Architecture** | amd64, arm64 |
| **Memory** | Minimum 4GB RAM |
| **Disk** | Minimum 100MB free space |

## Installation Methods

### Method 1: Quick Install (Recommended)

Installs the latest version automatically with a single command.

**macOS / Linux / WSL / Git Bash:**

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

**Windows (PowerShell):**

```powershell
irm https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.ps1 | iex
```

{{< callout type="info" >}}
The install script automatically detects your platform, downloads the prebuilt binary from GitHub, verifies the SHA256 checksum, and configures your PATH. No Python or separate runtime is required.
{{< /callout >}}

Once installation is complete, verify it:

```bash
moai version
```

#### Install Options

```bash
# Install a specific version (specify the desired release tag)
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash -s -- --version <release-tag>

# Install to a custom directory
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash -s -- --install-dir /usr/local/bin
```

### Method 2: Build from Source

If you have a Go development environment, you can build directly from source.

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk
make build
```

The built binary is generated at `./bin/moai`. Copy it to a location on your PATH:

```bash
cp ./bin/moai ~/.local/bin/
```

### Install Location

The install script chooses the install directory in this order:

| Platform | Priority |
|--------|---------|
| **macOS / Linux** | `$GOBIN` → `$GOPATH/bin` → `~/.local/bin` |
| **Windows** | `%LOCALAPPDATA%\Programs\moai` |

## Migrating from 1.x

{{< callout type="error" >}}
**MoAI-ADK 1.x (Python version) users must uninstall the old version first.**

1.x and 2.x use the same `moai` command, so a leftover old version causes conflicts.
{{< /callout >}}

### Step 1: Remove the Old 1.x

```bash
# If installed with uv
uv tool uninstall moai-adk

# If installed with pip
pip uninstall moai-adk
```

### Step 2: Back Up Existing Settings (Optional)

```bash
# If you want to back up your existing settings
cp -r ~/.moai ~/.moai-v1-backup
```

### Step 3: Install 2.x

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

### Step 4: Verify the Installation

```bash
moai version
# Example output: moai <version> (commit: <hash>, built: <build date>)
```

{{< callout type="info" >}}
The Go edition (v2.0+) is a single binary that needs no Python runtime or virtual environment. Startup time improved dramatically, from about 800ms to 5ms.
{{< /callout >}}

## WSL Support

For Windows users, here is how to install and use MoAI-ADK in a WSL (Windows Subsystem for Linux) environment.

### Installing WSL

If WSL is not installed, run the following in PowerShell (as administrator):

```powershell
wsl --install
```

After installing, restart Windows and Ubuntu is installed automatically.

### Installing MoAI-ADK in WSL

Use the same command as on Linux from a WSL terminal:

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

### Path Handling

In WSL, you need to distinguish Windows paths from WSL paths:

| Windows path | WSL path |
|-------------|----------|
| `C:\Users\name\project` | `/mnt/c/Users/name/project` |
| `D:\Projects\myapp` | `/mnt/d/Projects/myapp` |

{{< callout type="info" >}}
**Recommended**: Creating projects on WSL's Linux filesystem (`~/projects/`) improves I/O performance 2-5x. Accessing the Windows filesystem (`/mnt/c/`) can degrade performance.
{{< /callout >}}

### WSL Best Practices

1. **Use the Linux filesystem**: Create projects under the `~/projects/` directory
2. **Configure Git credentials**: Set up Git credentials in WSL separately from Windows
3. **Recommended terminal**: Use Windows Terminal to manage multiple WSL distributions

### WSL Troubleshooting

#### PATH Not Loading

```bash
# Add to ~/.bashrc or ~/.zshrc
source ~/.cargo/env
export PATH="$HOME/.local/bin:$PATH"
```

#### Hook/MCP Server Execution Permission Issues

```bash
# Grant execute permission
chmod +x ~/.claude/hooks/moai/*.sh
```

#### Slow Access to Windows Paths

Move your project to the Linux filesystem:

```bash
# Move from Windows to WSL
cp -r /mnt/c/Users/name/project ~/projects/
cd ~/projects/project
```

## pip and uv Tool Conflicts

A common issue for MoAI-ADK 1.x (Python version) users.

### The Problem

pip and uv install packages to different locations. Mixing the two tools can cause the `moai` command to run an unexpected version.

### Symptoms

- Running `moai version` shows a 1.x version
- `command not found: moai` errors
- The binary runs from a different path than `which moai` shows

### Cause

1. pip installs to the system Python path
2. uv tool installs to `~/.local/bin` or `~/.cargo/bin`
3. Depending on PATH order, a different version runs

### Solution

#### Complete Removal, Then Reinstall

```bash
# 1. Remove all existing versions
uv tool uninstall moai-adk 2>/dev/null || true
pip uninstall moai-adk -y 2>/dev/null || true

# 2. Check for and delete leftover binaries
which moai && rm $(which moai) 2>/dev/null || true
ls ~/.local/bin/moai && rm ~/.local/bin/moai 2>/dev/null || true

# 3. Install 2.x
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash

# 4. Verify
moai version
```

#### Update Your Shell Configuration

```bash
# Add to ~/.bashrc or ~/.zshrc
export PATH="$HOME/.local/bin:$PATH"

# Apply the change
source ~/.bashrc  # or source ~/.zshrc
```

### Prevention

1. MoAI-ADK 2.x is a Go binary with no relation to Python
2. Uninstall 1.x (the Python version) before installing 2.x
3. Do not use pip and uv tool at the same time

## Installation Troubleshooting

### Problem: Command Not Found

```bash
command not found: moai
```

**Solution:**

1. Restart your terminal
2. Check your PATH configuration:

```bash
echo $PATH
```

3. Check where the binary is installed:

```bash
which moai || ls ~/.local/bin/moai
```

4. Add it to PATH manually:

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

If the old version of the `moai` command is running:

```bash
# Check which moai is running
which moai

# If 1.x remains, remove it
uv tool uninstall moai-adk
# or
pip uninstall moai-adk

# Restart the terminal and verify 2.x
moai version
```

## Next Steps After Installation

Once installation is complete, initialize a project:

### Create a New Project

```bash
moai init my-project
```

### Apply to an Existing Project

```bash
cd my-existing-project
moai init
```

## Upgrading

To upgrade to the latest version:

```bash
moai update
```

### Update Options

```bash
# Check the version only (no update)
moai update --check

# Sync templates only (skip package upgrade)
moai update --templates-only

# Configuration edit mode (re-run the init wizard)
moai update --config
moai update -c

# Force update without backup
moai update --force

# Auto-approve mode (approve all confirmations automatically)
moai update --yes
```

### Merge Strategies

```bash
# Force automatic merge (default)
moai update --merge

# Force manual merge
moai update --manual
```

{{< callout type="info" >}}
**Automatically preserved items**: User settings, custom agents, custom commands, custom skills, custom hooks, SPEC documents, and reports are preserved automatically during updates.
{{< /callout >}}

For details, see the [Update Guide](https://adk.mo.ai.kr/getting-started/update).

## Uninstalling

To remove MoAI-ADK completely, delete the binary and the configuration directory:

```bash
# Delete the binary (using the result of which moai)
rm "$(which moai)"

# Delete the configuration directory (optional)
rm -rf "$HOME/.moai"
```

---

## Next Steps

Learn how to configure MoAI-ADK in the [Initial Setup Wizard](./init-wizard).
