# MoAI-ADK Troubleshooting Guide

**Available Languages:** [🇰🇷 한국어](./TROUBLESHOOTING.ko.md) | [🇺🇸 English](./TROUBLESHOOTING.md) | [🇯🇵 日本語](./TROUBLESHOOTING.ja.md) | [🇨🇳 中文](./TROUBLESHOOTING.zh.md)

---

## Table of Contents

1. [Common Installation Issues](#common-installation-issues)
2. [Runtime Issues](#runtime-issues)
3. [Hook Issues](#hook-issues)
4. [Git Integration Issues](#git-integration-issues)
5. [Claude Code Integration Issues](#claude-code-integration-issues)
6. [Platform-Specific Issues](#platform-specific-issues)
7. [Performance Issues](#performance-issues)

---

## Common Installation Issues

### Issue: "command not found: moai"

**Severity**: 🔴 High

**Symptoms**:
```bash
$ moai --version
zsh: command not found: moai
# or
moai: command not found
```

**Diagnosis**:
```bash
# 1. Check if moai is installed
uv tool list | grep moai

# 2. Check PATH
echo $PATH

# 3. Find moai location
find ~ -name "moai" -type f 2>/dev/null
```

**Solutions**:

#### Solution 1: Add uv tool bin to PATH

**macOS/Linux (~/.zshrc or ~/.bashrc)**:
```bash
# Add this line
export PATH="$HOME/.local/bin:$PATH"

# Apply changes
source ~/.zshrc  # or source ~/.bashrc
```

**Windows (PowerShell $PROFILE)**:
```powershell
# Add this line
$env:PATH = "$env:USERPROFILE\.local\bin;$env:PATH"

# Make permanent
[System.Environment]::SetEnvironmentVariable("Path", $env:PATH, [System.EnvironmentVariableTarget]::User)
```

#### Solution 2: Reinstall MoAI-ADK

```bash
uv tool uninstall moai-adk
uv tool install moai-adk
```

#### Solution 3: Install uv first

```bash
# macOS/Linux
curl -LsSf https://astral.sh/uv/install.sh | sh

# Windows
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"

# Then install MoAI-ADK
uv tool install moai-adk
```

**Verification**:
```bash
which moai  # Should show ~/.local/bin/moai
moai --version  # Should show version number
```

---

### Issue: pip and uv tool version conflict

**Severity**: 🟡 Medium

**Symptoms**:
```bash
$ moai update
✓ Package already up to date (1.5.0)

$ moai --version
MoAI-ADK, version 1.1.0  # Different version!

$ which moai
~/.pyenv/shims/moai  # Not ~/.local/bin/moai

$ moai init
SessionStart hook error: ModuleNotFoundError: No module named 'yaml'
```

**Root Cause**:
- Both pip and uv tool have installed MoAI-ADK
- PATH priority determines which version is used
- pip-installed version may be outdated or missing dependencies

**Diagnosis**:
```bash
# Check all installed moai versions
pip show moai-adk 2>/dev/null
uv tool list | grep moai

# Check which moai is being used
which moai
moai --version

# Check PATH priority
echo $PATH | tr ':' '\n' | grep -E '(pyenv|local/bin)'
```

**Solutions**:

#### Solution 1: Uninstall pip version (Recommended)

```bash
# Remove pip version
pip uninstall moai-adk -y

# Verify uv tool version
uv tool list | grep moai

# Check which moai is active
which moai  # Should show ~/.local/bin/moai

# Verify version
moai --version  # Should match uv tool version
```

#### Solution 2: Prioritize uv tool in PATH

**macOS/Linux (~/.zshrc)**:
```bash
# Add AFTER pyenv initialization
# ===== UV Tool Priority =====
# uv tool 설치 명령어가 pyenv shims보다 우선하도록 설정
# MoAI-ADK (moai, moai-adk, moai-worktree 등)는 uv tool로 관리
export PATH="$HOME/.local/bin:$PATH"
```

**Windows ($PROFILE)**:
```powershell
# UV Tool Priority
$env:PATH = "$env:USERPROFILE\.local\bin;$env:PATH"
```

#### Solution 3: Force reinstall with uv

```bash
# Uninstall all versions
pip uninstall moai-adk -y
uv tool uninstall moai-adk

# Reinstall with uv
uv tool install moai-adk

# Verify
uv tool list
which moai
moai --version
```

**Prevention**:
- Always use `uv tool install moai-adk` for installation
- Avoid mixing pip and uv for the same package
- Regularly check `which moai` to verify active installation

**Windows-Specific Notes**:
Windows users may experience more severe issues due to:
1. Multiple Python installations (Python Launcher, Store, Anaconda, etc.)
2. Different PATH priorities across shells (cmd, PowerShell, Git Bash)
3. Virtual environment isolation differences

**Windows Prevention**:
```powershell
# 1. Uninstall all Python versions except one
# 2. Use uv exclusively for package management
# 3. Update PATH in System Environment Variables
# 4. Restart all terminals after PATH changes
```

---

### Issue: "Permission denied" when installing

**Severity**: 🟡 Medium

**Symptoms**:
```bash
$ uv tool install moai-adk
error: failed to create directory: Permission denied
```

**Solutions**:

#### Solution 1: Fix directory permissions

```bash
# Ensure ~/.local/bin exists and is writable
mkdir -p ~/.local/bin
chmod u+w ~/.local/bin

# Retry installation
uv tool install moai-adk
```

#### Solution 2: Don't use sudo with uv

```bash
# ❌ WRONG
sudo uv tool install moai-adk

# ✅ CORRECT
uv tool install moai-adk
```

**Windows**:
```powershell
# Run PowerShell as Administrator
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Install
uv tool install moai-adk
```

---

## Runtime Issues

### Issue: Hook execution fails

**Severity**: 🟡 Medium

**Symptoms**:
```bash
SessionStart:resume hook error: ...
SessionStart:show_project_info hook warning: ...
```

**Diagnosis**:
```bash
# Check hook files
ls -la ~/.claude/hooks/moai/

# Test hook manually
python ~/.claude/hooks/moai/session_start__show_project_info.py

# Check Python version
python --version  # Should be 3.11+
```

**Solutions**:

#### Solution 1: Clear hooks cache

```bash
# Clear Python cache
rm -rf ~/.claude/hooks/moai/__pycache__
rm -rf ~/.claude/hooks/moai/*/__pycache__

# Reinstall
uv tool install moai-adk --reinstall
```

#### Solution 2: Check Python dependencies

```bash
# Verify required modules
python -c "import yaml; import requests; import psutil"

# If any import fails, reinstall
uv tool install moai-adk --reinstall
```

#### Solution 3: Update hooks

```bash
# Force template sync
moai update --templates-only
```

---

### Issue: "ModuleNotFoundError: No module named 'yaml'"

**Severity**: 🔴 High

**Symptoms**:
```bash
SessionStart:resume hook error:
  File ".../version_reader.py", line 19
    import yaml
ModuleNotFoundError: No module named 'yaml'
```

**Root Cause**:
- pip-installed version missing PyYAML dependency
- Version mismatch between pip and uv installations

**Solutions**:

#### Solution 1: Complete reinstall with uv

```bash
# Uninstall all versions
pip uninstall moai-adk -y
uv tool uninstall moai-adk

# Clean install
uv tool install moai-adk

# Verify
python -c "import yaml; print('PyYAML:', yaml.__version__)"
moai --version
```

#### Solution 2: Install missing dependency

```bash
# For pip installation
pip install pyyaml

# For uv tool installation (should be automatic)
uv tool install moai-adk --reinstall
```

**Prevention**:
- Always use uv tool for installation
- Avoid mixing pip and uv
- Run `moai doctor` after installation

---

## Git Integration Issues

### Issue: "git: command not found"

**Severity**: 🟡 Medium

**Solutions**:

**macOS**:
```bash
brew install git
```

**Ubuntu/Debian**:
```bash
sudo apt install git
```

**Windows**:
- Download from [git-scm.com](https://git-scm.com/downloads)

---

### Issue: MoAI-ADK commands not committing changes

**Severity**: 🟡 Medium

**Diagnosis**:
```bash
# Check git status
git status

# Check recent commits
git log --oneline -5
```

**Solutions**:

1. **Check Git mode**:
   ```bash
   cat .moai/config/config.yaml | grep git_mode
   ```

2. **Verify git config**:
   ```bash
   git config user.name
   git config user.email
   ```

3. **Manual commit**:
   ```bash
   git add .
   git commit -m "Manual commit"
   ```

---

## Claude Code Integration Issues

### Issue: CLAUDE.md not recognized

**Severity**: 🟢 Low

**Symptoms**:
- Claude Code doesn't follow CLAUDE.md rules
- Commands not recognized

**Solutions**:

1. **Restart Claude Code**:
   - Close and reopen Claude Code
   - Or run `/clear` in chat

2. **Verify CLAUDE.md**:
   ```bash
   ls -la CLAUDE.md
   cat CLAUDE.md | head -20
   ```

3. **Check settings**:
   ```bash
   cat .claude/settings.json | jq '.markdown'
   ```

---

### Issue: MCP servers not connecting

**Severity**: 🟡 Medium

**Diagnosis**:
```bash
# Check MCP settings
cat .claude/settings.json | jq '.mcpServers'

# Check MCP servers
ls -la ~/.claude/mcp-servers/
```

**Solutions**:

1. **Reinstall MCP servers**:
   ```bash
   npm install -g @modelcontextprotocol/server-context7
   ```

2. **Update settings**:
   ```bash
   moai update
   ```

---

## Platform-Specific Issues

### macOS

#### Issue: "xcrun: error: invalid active developer path"

**Symptoms**:
```bash
xcrun: error: invalid active developer path (/Library/Developer/CommandLineTools)
```

**Solution**:
```bash
xcode-select --install
```

---

### Windows

#### Issue: PowerShell execution policy

**Symptoms**:
```powershell
uv : Cannot load because running scripts is disabled on this system
```

**Solution**:
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

#### Issue: Python not in PATH

**Symptoms**:
```powershell
python : The term 'python' is not recognized
```

**Solution**:
1. Add Python to PATH in System Environment Variables
2. Or use Python Launcher:
   ```powershell
   py -m uv tool install moai-adk
   ```

---

### Linux

#### Issue: Missing system dependencies

**Symptoms**:
```bash
error: failed to build wheel for psutil
```

**Solution**:
```bash
# Ubuntu/Debian
sudo apt install -y python3-dev build-essential

# Fedora/RHEL
sudo dnf install -y python3-devel gcc
```

---

## Performance Issues

### Issue: Slow hook execution

**Severity**: 🟢 Low

**Symptoms**:
- Hooks take >5 seconds to run
- Claude Code startup is slow

**Solutions**:

1. **Disable heavy hooks**:
   ```bash
   # In .claude/settings.json
   {
     "hooks": {
       "SessionStart:resume": false
     }
   }
   ```

2. **Use Python optimization**:
   ```bash
   export PYTHONOPTIMIZE=2
   ```

3. **Clear cache**:
   ```bash
   rm -rf ~/.claude/hooks/__pycache__
   ```

---

### Issue: High memory usage

**Severity**: 🟢 Low

**Solutions**:

1. **Disable statusline memory tracking**:
   ```yaml
   # In .moai/config/statusline-config.yaml
   display:
     memory_usage: false
   ```

2. **Limit session history**:
   ```bash
   # In .claude/settings.json
   {
     "sessionHistory": {
       "maxSessions": 10
     }
   }
   ```

---

## Getting Additional Help

If issues persist:

1. **Run diagnostics**:
   ```bash
   moai doctor
   ```

2. **Check logs**:
   ```bash
   cat ~/.claude/logs/moai-*.log
   ```

3. **Report issue**:
   - GitHub: [https://github.com/modu-ai/moai-adk/issues](https://github.com/modu-ai/moai-adk/issues)
   - Discord: [https://discord.gg/umywNygN](https://discord.gg/umywNygN)
   - Email: [support@mo.ai.kr](mailto:support@mo.ai.kr)

4. **Include diagnostic information**:
   ```bash
   moai doctor > diagnostics.txt
   moai --version >> diagnostics.txt
   python --version >> diagnostics.txt
   uv --version >> diagnostics.txt
   which moai >> diagnostics.txt
   ```

---

**Last Updated**: 2026-01-20
**Version**: 1.5.0
