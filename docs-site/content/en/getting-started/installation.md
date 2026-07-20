---
title: Installation
weight: 30
draft: false
---

A guide to installing MoAI-ADK on your system. The installation is a single Go-built binary — no Python, no virtual environment, no package manager required.


## License

MoAI-ADK {{< version >}} and later is distributed under the **Apache-2.0 license**.

Commercial use, modification, and distribution are free, with no obligation to open-source your code. For details, see [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).

{{< callout type="info" >}}
**Note**: MoAI-ADK 1.x (the Python version) was under the GPL-3.0 license. From v2.0.0 it was rewritten in Go and changed to Apache-2.0.
{{< /callout >}}

## Prerequisites

Check the following before installing:

### 1. Claude Code

MoAI-ADK is an extension framework that runs on top of Claude Code. Claude Code must be installed first.

```bash
claude --version
```

If you have not installed it yet, see the [Claude Code official documentation](https://docs.anthropic.com/en/docs/claude-code).

### 2. Git (required)

MoAI-ADK uses a Git-based workflow. Git must be installed on your system.

```bash
git --version
```

{{< callout type="warning" >}}
**Windows users**: You must use **Git Bash** or **WSL**. Command Prompt (cmd.exe) is not supported.

If Git is not installed:
- **Windows**: Install Git for Windows from [git-scm.com](https://git-scm.com). Git Bash is installed alongside it.
- **macOS**: `xcode-select --install` or [git-scm.com](https://git-scm.com)
- **Linux**: `sudo apt install git` (Ubuntu/Debian) or `sudo dnf install git` (Fedora)
{{< /callout >}}

### System requirements

| Item | Requirement |
|------|---------|
| **OS** | macOS, Linux, Windows (Git Bash / WSL) |
| **Architecture** | amd64, arm64 |
| **Memory** | 4GB RAM minimum |
| **Disk** | 100MB free space minimum |

## Installation methods

### Method 1: Quick install (recommended)

Automatically installs the latest version with a single command.

**macOS / Linux / WSL / Git Bash:**

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

**Windows (PowerShell):**

```powershell
irm https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.ps1 | iex
```

{{< callout type="info" >}}
The install script automatically detects the platform, downloads a pre-built binary from GitHub, verifies the SHA256 checksum, and configures the PATH. No Python or separate runtime is required.
{{< /callout >}}

Once the install completes, verify it:

```bash
moai version
```

#### Install options

```bash
# Install a specific version (specify the desired release tag)
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash -s -- --version <release-tag>

# Install to a custom directory
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash -s -- --install-dir /usr/local/bin
```

### Method 2: Build from source

If you have a Go development environment, you can build directly from source.

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk
make build
```

The built binary is created at `./bin/moai`. Copy it to a location on your PATH:

```bash
cp ./bin/moai ~/.local/bin/
```

### Install location

The install script determines the install directory in the following order:

| Platform | Priority |
|--------|---------|
| **macOS / Linux** | `$GOBIN` → `$GOPATH/bin` → `~/.local/bin` |
| **Windows** | `%LOCALAPPDATA%\Programs\moai` |

## Migration for 1.x users

{{< callout type="error" >}}
**MoAI-ADK 1.x (Python version) users must remove the old version first.**

Since 1.x and 2.x use the same `moai` command, a leftover old version causes conflicts.
{{< /callout >}}

### Step 1: remove the old 1.x

```bash
# If installed with uv
uv tool uninstall moai-adk

# If installed with pip
pip uninstall moai-adk
```

### Step 2: back up your old settings (optional)

```bash
# If you want to back up your old settings
cp -r ~/.moai ~/.moai-v1-backup
```

### Step 3: install 2.x

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

### Step 4: verify the installation

```bash
moai version
```

```text
╭────────────────────────╮
│                        │
│    moai-adk v3.0.0     │
│                        │
│                        │
╰────────────────────────╯
 v3.0.0   none   built unknown
```

{{< callout type="info" >}}
The Go edition (v2.0+) is a single binary that needs no Python runtime or virtual environment. Startup time improved dramatically from about 800ms to 5ms.
{{< /callout >}}

## WSL support

For Windows users, here is how to install and use MoAI-ADK in the WSL (Windows Subsystem for Linux) environment.

### Installing WSL

If WSL is not installed, run the following in PowerShell (as administrator):

```powershell
wsl --install
```

After installing and restarting Windows, Ubuntu is installed automatically.

### Installing MoAI-ADK in WSL

In the WSL terminal, use the same command as Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

### Path handling

In WSL you must distinguish Windows paths from WSL paths:

| Windows path | WSL path |
|-------------|----------|
| `C:\Users\name\project` | `/mnt/c/Users/name/project` |
| `D:\Projects\myapp` | `/mnt/d/Projects/myapp` |

{{< callout type="info" >}}
**Recommended**: Creating projects on the WSL Linux filesystem (`~/projects/`) improves I/O performance 2-5x. Accessing the Windows filesystem (`/mnt/c/`) can degrade performance.
{{< /callout >}}

### WSL best practices

1. **Use the Linux filesystem**: create projects in the `~/projects/` directory
2. **Set up Git credentials**: configure Git credentials in WSL separately from Windows
3. **Recommended terminal**: use Windows Terminal to manage multiple WSL distributions

### WSL troubleshooting

#### PATH not loaded

```bash
# Add to ~/.bashrc or ~/.zshrc
source ~/.cargo/env
export PATH="$HOME/.local/bin:$PATH"
```

#### Hook / MCP server execute-permission issues

```bash
# Grant execute permission
chmod +x ~/.claude/hooks/moai/*.sh
```

#### Slow Windows path access

Move your project to the Linux filesystem:

```bash
# Move from Windows to WSL
cp -r /mnt/c/Users/name/project ~/projects/
cd ~/projects/project
```

## pip and uv tool conflicts

A common problem MoAI-ADK 1.x (Python version) users can encounter.

### Problem description

pip and uv install packages to different locations. Mixing the two tools can make the `moai` command run an unexpected version.

### Symptoms

- Running `moai version` shows a 1.x version
- A `command not found: moai` error occurs
- It runs from a path different from `which moai`

### Cause

1. pip installs to the system Python path
2. uv tool installs to `~/.local/bin` or `~/.cargo/bin`
3. A different version runs depending on the PATH order

### Solution

#### Fully remove, then reinstall

```bash
# 1. Remove all existing versions
uv tool uninstall moai-adk 2>/dev/null || true
pip uninstall moai-adk -y 2>/dev/null || true

# 2. Find and delete leftover binaries
which moai && rm $(which moai) 2>/dev/null || true
ls ~/.local/bin/moai && rm ~/.local/bin/moai 2>/dev/null || true

# 3. Install 2.x
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash

# 4. Verify
moai version
```

#### Update your shell configuration

```bash
# Add to ~/.bashrc or ~/.zshrc
export PATH="$HOME/.local/bin:$PATH"

# Apply the configuration
source ~/.bashrc  # or source ~/.zshrc
```

### Prevention

1. MoAI-ADK 2.x is a Go binary unrelated to Python
2. Remove 1.x (Python version) before installing 2.x
3. Do not use pip and uv tool at the same time

## Installation troubleshooting

### Problem: command not found

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

### Problem: permission denied

```bash
Permission denied
```

**Solution:**

```bash
chmod +x ~/.local/bin/moai
```

### Problem: 1.x and 2.x conflict

If an older `moai` command runs:

```bash
# Check which moai runs
which moai

# If 1.x remains, remove it
uv tool uninstall moai-adk
# or
pip uninstall moai-adk

# Restart the terminal, then verify 2.x
moai version
```

## Next steps after installation

Once installation is complete, initialize a project:

### Create a new project

```bash
moai init my-project
```

### Apply to an existing project

```bash
cd my-existing-project
moai init
```

## Upgrading

To upgrade to the latest version:

```bash
moai update
```

### Update options

```bash
# Check the version only (no update)
moai update --check

# Sync templates only (skip the binary update)
moai update --templates-only

# Config edit mode (re-run the setup wizard)
moai update -c

# Force update (user changes are backed up, then overwritten)
moai update --force

# Auto-approve mode (CI/CD)
moai update --yes
```

{{< callout type="info" >}}
**Auto-preserved items**: User settings, custom agents, custom commands, custom skills, custom hooks, SPEC documents, and reports are preserved automatically on update. Template files you modified are backed up and then 3-way merged.
{{< /callout >}}

For details, see the [Update guide](https://adk.mo.ai.kr/cli-reference/update).

## Uninstalling

To fully remove MoAI-ADK, delete the binary and the settings directory:

```bash
# Delete the binary (delete the result of which moai)
rm "$(which moai)"

# Delete the settings directory (optional)
rm -rf "$HOME/.moai"
```

---

## Next steps

Learn how to configure MoAI-ADK in the [Initial Setup wizard](./init-wizard).
