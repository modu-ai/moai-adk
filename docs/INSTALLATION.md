# MoAI-ADK Installation Guide

**Available Languages:** [🇰🇷 한국어](./INSTALLATION.ko.md) | [🇺🇸 English](./INSTALLATION.md) | [🇯🇵 日本語](./INSTALLATION.ja.md) | [🇨🇳 中文](./INSTALLATION.zh.md)

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Installation Methods](#installation-methods)
3. [Platform-Specific Instructions](#platform-specific-instructions)
4. [Post-Installation Setup](#post-installation-setup)
5. [Verification](#verification)
6. [Troubleshooting](#troubleshooting)
7. [Uninstallation](#uninstallation)

---

## Prerequisites

### Required Software

| Software | Minimum Version | Recommended | Installation Guide |
|----------|----------------|-------------|---------------------|
| **Python** | 3.11 | 3.13+ | [python.org](https://www.python.org/downloads/) |
| **Git** | 2.30+ | Latest | [git-scm.com](https://git-scm.com/downloads) |
| **uv** (Recommended) | 0.1.0 | Latest | See below |
| **Claude Code** | Latest | - | [claude.ai/code](https://claude.ai/code) |

### Optional Software

- **pyenv**: For Python version management (macOS/Linux)
- **virtualenv**: For isolated Python environments
- **Node.js 18+**: Required for MCP servers

---

## Installation Methods

### Method 1: Quick Install (Recommended)

**Best for**: Most users, quick setup

```bash
curl -LsSf https://modu-ai.github.io/moai-adk/install.sh | sh
```

**What it does**:
- Downloads and installs `uv` if not present
- Installs MoAI-ADK using `uv tool install`
- Verifies installation
- Configures shell environment

**Output**:
```text
✓ MoAI-ADK installed successfully
  Version: 1.5.0
  Location: ~/.local/bin/moai

Next steps:
  1. Run 'moai init' in your project directory
  2. Start Claude Code
  3. Run /moai:0-project for full setup
```

---

### Method 2: Manual Install with uv

**Best for**: Users who want control over the installation process

#### Step 1: Install uv

**macOS/Linux**:
```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

**Windows**:
```powershell
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"
```

#### Step 2: Install MoAI-ADK

```bash
uv tool install moai-adk
```

#### Step 3: Verify Installation

```bash
which moai
# Should show: ~/.local/bin/moai (Unix) or %USERPROFILE%\.local\bin\moai.exe (Windows)

moai --version
# Should show: MoAI-ADK, version 1.5.0
```

---

### Method 3: Install with pip (Not Recommended)

**Best for**: Users without uv, or in restricted environments

```bash
pip install moai-adk
```

**⚠️ Warning**: pip installation may conflict with uv tool installation. See [Troubleshooting](#troubleshooting) for details.

---

## Platform-Specific Instructions

### macOS

#### Prerequisites Installation

```bash
# Install Homebrew (if not installed)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install Python 3.13
brew install python@3.13

# Install uv
curl -LsSf https://astral.sh/uv/install.sh | sh

# Install pyenv (optional)
brew install pyenv
```

#### Shell Configuration

**Zsh (default on macOS)**:

Add to `~/.zshrc`:
```zsh
# ===== UV Tool Priority =====
# uv tool 설치 명령어가 pyenv shims보다 우선하도록 설정
export PATH="$HOME/.local/bin:$PATH"
```

**Bash**:

Add to `~/.bash_profile` or `~/.bashrc`:
```bash
# ===== UV Tool Priority =====
export PATH="$HOME/.local/bin:$PATH"
```

#### Apply Changes

```bash
source ~/.zshrc  # or source ~/.bash_profile
```

---

### Linux

#### Ubuntu/Debian

```bash
# Update package list
sudo apt update

# Install Python and dependencies
sudo apt install -y python3 python3-pip python3-venv git

# Install uv
curl -LsSf https://astral.sh/uv/install.sh | sh

# Install MoAI-ADK
uv tool install moai-adk
```

#### Fedora/RHEL

```bash
# Install Python and dependencies
sudo dnf install -y python3 python3-pip git

# Install uv
curl -LsSf https://astral.sh/uv/install.sh | sh

# Install MoAI-ADK
uv tool install moai-adk
```

#### Shell Configuration

Add to `~/.bashrc` or `~/.zshrc`:
```bash
# ===== UV Tool Priority =====
export PATH="$HOME/.local/bin:$PATH"
```

---

### Windows

#### Prerequisites Installation

**Option 1: Using PowerShell**

```powershell
# Install uv
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"

# Install MoAI-ADK
uv tool install moai-adk
```

**Option 2: Using Chocolatey**

```powershell
# Install uv
choco install uv

# Install MoAI-ADK
uv tool install moai-adk
```

#### PATH Configuration

Add to your PowerShell profile (`$PROFILE`):
```powershell
# ===== UV Tool Priority =====
$env:PATH = "$env:USERPROFILE\.local\bin;$env:PATH"
```

To make it permanent, add to System Environment Variables:
1. Open **System Properties** → **Advanced** → **Environment Variables**
2. Edit **Path** under **User variables**
3. Add: `%USERPROFILE%\.local\bin`

#### Apply Changes

```powershell
# Refresh environment variables
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","User") + ";" + [System.Environment]]::GetEnvironmentVariable("Path","Machine")
```

---

## Post-Installation Setup

### Step 1: Initialize Project

```bash
# For new project
mkdir my-project && cd my-project
moai init

# For existing project
cd existing-project
moai init .
```

### Step 2: Configure Claude Code

```bash
# Start Claude Code in your project
claude

# Run project initialization
> /moai:0-project
```

### Step 3: Verify Setup

```bash
# Check MoAI-ADK status
moai status

# List installed tools
uv tool list | grep moai

# Verify command availability
which moai moai-adk moai-worktree moai-wt
```

---

## Verification

### Quick Health Check

```bash
moai doctor
```

**Expected Output**:
```text
✓ MoAI-ADK: 1.5.0
✓ Python: 3.13.1
✓ uv: 0.9.3
✓ Git: 2.39.0
✓ Claude Code: Installed

All systems operational!
```

### Manual Verification

```bash
# 1. Check version
moai --version

# 2. Check command location
which moai

# 3. Test basic functionality
moai init /tmp/test-moai-project

# 4. Check help
moai --help
```

---

## Troubleshooting

### Issue: "command not found: moai"

**Symptoms**:
```bash
moai --version
# zsh: command not found: moai
```

**Solutions**:

1. **Check installation**:
   ```bash
   uv tool list | grep moai
   ```

2. **Verify PATH**:
   ```bash
   echo $PATH | grep -o '[^:]*\.local/bin[^:]*'
   ```

3. **Add to PATH** (see Platform-Specific Instructions above)

4. **Reload shell**:
   ```bash
   source ~/.zshrc  # or source ~/.bashrc
   ```

---

### Issue: pip and uv tool conflict

**Symptoms**:
```bash
moai --version
# Shows different version than expected

which moai
# Shows pyenv shims instead of ~/.local/bin
```

**Solutions**:

1. **Uninstall pip version**:
   ```bash
   pip uninstall moai-adk -y
   ```

2. **Reinstall with uv**:
   ```bash
   uv tool uninstall moai-adk
   uv tool install moai-adk
   ```

3. **Prioritize uv in PATH**:
   ```bash
   export PATH="$HOME/.local/bin:$PATH"
   ```

4. **Verify**:
   ```bash
   which moai  # Should show ~/.local/bin/moai
   ```

**For Windows**:
```powershell
# Uninstall pip version
pip uninstall moai-adk

# Reinstall with uv
uv tool uninstall moai-adk
uv tool install moai-adk

# Update PATH in $PROFILE
$env:PATH = "$env:USERPROFILE\.local\bin;$env:PATH"
```

---

### Issue: "ModuleNotFoundError: No module named 'yaml'"

**Symptoms**:
```bash
# When running hooks or commands
SessionStart:resume hook error: ModuleNotFoundError: No module named 'yaml'
```

**Root Cause**:
- pip-installed version lacks dependencies
- Version mismatch between pip and uv installations

**Solutions**:

1. **Uninstall all versions**:
   ```bash
   pip uninstall moai-adk -y
   uv tool uninstall moai-adk
   ```

2. **Clean reinstall with uv**:
   ```bash
   uv tool install moai-adk
   ```

3. **Verify dependencies**:
   ```bash
   uv tool list
   # Should show moai-adk with all dependencies
   ```

---

### Issue: "Permission denied" when installing

**On Unix/macOS**:
```bash
# Don't use sudo with uv tool
# Instead, ensure ~/.local/bin is writable
mkdir -p ~/.local/bin
chmod u+w ~/.local/bin

# Reinstall
uv tool install moai-adk
```

**On Windows**:
```powershell
# Run PowerShell as Administrator
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Reinstall
uv tool install moai-adk
```

---

### Issue: Hook errors after update

**Symptoms**:
```bash
SessionStart hook error: ...
```

**Solutions**:

1. **Clear hooks cache**:
   ```bash
   rm -rf ~/.claude/hooks/__pycache__
   ```

2. **Reinstall**:
   ```bash
   uv tool install moai-adk --reinstall
   ```

3. **Check Python version**:
   ```bash
   python --version  # Should be 3.11+
   ```

---

## Uninstallation

### Remove MoAI-ADK

**Using uv**:
```bash
uv tool uninstall moai-adk
```

**Using pip**:
```bash
pip uninstall moai-adk
```

### Remove Configuration

```bash
# Remove global config
rm -rf ~/.moai

# Remove hooks (optional)
rm -rf ~/.claude/hooks/moai

# Remove project-specific .moai folders (per project)
rm -rf /path/to/project/.moai
```

### Remove Claude Code Integration (Optional)

```bash
# This removes CLAUDE.md and .claude folder from projects
# Only do this if you want to completely remove MoAI-ADK integration
```

---

## Getting Help

If you encounter issues not covered here:

1. **Check GitHub Issues**: [https://github.com/modu-ai/moai-adk/issues](https://github.com/modu-ai/moai-adk/issues)
2. **Join Discord**: [https://discord.gg/umywNygN](https://discord.gg/umywNygN)
3. **Email Support**: [support@mo.ai.kr](mailto:support@mo.ai.kr)
4. **Documentation**: [https://adk.mo.ai.kr](https://adk.mo.ai.kr)

---

**Last Updated**: 2026-01-20
**Version**: 1.5.0
