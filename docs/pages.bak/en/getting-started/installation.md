# Installation Guide

Install and run MoAI-ADK on your system in just minutes. This guide covers system requirements, installation methods, and verification steps.

## System Requirements

### Minimum Requirements

- **Python**: 3.13 or higher
- **Operating System**:
  - macOS (10.15+)
  - Linux (Ubuntu 20.04+, CentOS 8+, Debian 11+)
  - Windows 10+ (PowerShell recommended)
- **Git**: 2.25 or higher
- **Memory**: 4GB RAM minimum, 8GB recommended
- **Storage**: 500MB free space

### Recommended Requirements

- **Python**: 3.13+ (latest stable version)
- **Package Manager**: UV 0.5.0+ (recommended) or pip 24.0+
- **IDE**: VS Code with Claude Code extension or preferred editor
- **Terminal**: Modern terminal with UTF-8 support

## Installation Methods

### Method 1: UV Package Manager (Recommended)

UV is the fastest and most reliable way to install MoAI-ADK. It provides automatic dependency management and virtual environment handling.

#### Step 1: Install UV

**macOS/Linux:**

```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

**Windows (PowerShell):**

```powershell
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"
```

#### Step 2: Verify UV Installation

```bash
uv --version
# Expected output: uv 0.5.1 or higher
```

#### Step 3: Install MoAI-ADK

```bash
uv tool install moai-adk
```

#### Step 4: Verify Installation

```bash
moai-adk --version
# Expected output: MoAI-ADK v1.0.0 or higher
```

### Method 2: PyPI Installation (Alternative)

For pip users or when UV is unavailable.

#### Step 1: Upgrade pip (if needed)

```bash
python -m pip install --upgrade pip
```

#### Step 2: Install MoAI-ADK

```bash
pip install moai-adk
```

#### Step 3: Verify Installation

```bash
moai-adk --version
```

### Method 3: Development Installation

For developers who want to contribute to MoAI-ADK.

#### Step 1: Clone Repository

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk
```

#### Step 2: Install in Development Mode

```bash
# Using UV (recommended)
uv pip install -e .

# Or using pip
pip install -e .
```

#### Step 3: Verify Installation

```bash
moai-adk --version
```

## Post-Installation Setup

### Environment Variables

Optional but recommended environment variables:

```bash
# Add to shell profile (~/.bashrc, ~/.zshrc, etc.)
export MOAI_LOG_LEVEL=INFO
export MOAI_CACHE_DIR="$HOME/.moai/cache"
export CLAUDE_PROJECT_DIR=$(pwd)
```

### Claude Code Integration

MoAI-ADK requires Claude Code for the complete experience.

#### Install Claude Code

```bash
# macOS
brew install claude-ai/claude/claude

# Linux
curl -fsSL https://claude.ai/install.sh | sh

# Windows
winget install Anthropic.Claude
```

#### Verify Claude Code

```bash
claude --version
# Expected: Claude Code v1.5.0 or higher
```

### Optional MCP Servers

MoAI-ADK supports Model Context Protocol (MCP) servers for enhanced functionality.

#### Install Recommended MCP Servers

```bash
# Context7 - Latest library documentation
npx -y @upstash/context7-mcp

# Playwright - Web E2E testing
npx -y @playwright/mcp

# Sequential Thinking - Complex reasoning
npx -y @modelcontextprotocol/server-sequential-thinking
```

## Verification

### Check System Status

Run the built-in doctor command to verify installation:

```bash
moai-adk doctor
```

**Expected Output:**

```
Running system diagnostics...

┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━┓
┃ Check                                    ┃ Status ┃
┡━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━┩
│ Python >= 3.13                           │   ✓    │
│ uv installed                             │   ✓    │
│ Git installed                            │   ✓    │
│ Claude Code available                    │   ✓    │
│ Package registry accessible              │   ✓    │
└──────────────────────────────────────────┴────────┘

✅ All checks passed!
```

### Create Test Project

Create a simple test project to ensure everything works:

```bash
# Create test project
moai-adk init test-project
cd test-project

# Start Claude Code
claude

# In Claude Code, run:
/alfred:0-project
```

## Troubleshooting

### Common Issues

#### Issue: "uv: command not found"

**Solution:**

1. Verify UV is properly installed
2. Add UV to PATH:
   ```bash
   export PATH="$HOME/.cargo/bin:$PATH"
   ```
3. Restart terminal

#### Issue: "Python 3.8 found, but 3.13+ required"

**Solution:**

```bash
# Using pyenv
curl https://pyenv.run | bash
pyenv install 3.13
pyenv global 3.13

# Or using UV
uv python install 3.13
uv python pin 3.13
```

#### Issue: "Permission denied" during installation

**Solution:**

```bash
# Use user installation
pip install --user moai-adk

# Or use sudo (Linux/macOS)
sudo pip install moai-adk
```

#### Issue: Claude Code not recognized

**Solution:**

1. Verify Claude Code installation: `claude --version`
2. Check if in PATH
3. Reinstall if necessary

#### Issue: ModuleNotFoundError for dependencies

**Solution:**

```bash
# In project directory
uv sync

# Or install specific dependency
uv add fastapi pytest
```

### Getting Help

If you encounter issues not covered here:

1. **Check GitHub Issues**: Search existing issues at https://github.com/modu-ai/moai-adk/issues
2. **Run Detailed Diagnostics**: `moai-adk doctor --verbose`
3. **Create Issue**: Use `/alfred:9-feedback` in Claude Code to automatically create GitHub issue

## Next Steps

After successful installation:

1. **[Quick Start Guide](quick-start.md)** - Run your first project in 10 minutes
2. **[Core Concepts](concepts.md)** - Understand SPEC-First, TDD, @TAG, and TRUST 5 principles
3. **[Project Initialization](../../guides/project/init.md)** - Learn project setup and configuration

## Installation Summary

```bash
# One-line installation (recommended)
curl -LsSf https://astral.sh/uv/install.sh | sh && uv tool install moai-adk

# Verify installation
moai-adk doctor

# Create first project
moai-adk init my-project && cd my-project && claude
```

You're now ready to experience the power of SPEC-First TDD development with Alfred SuperAgent! 🚀
