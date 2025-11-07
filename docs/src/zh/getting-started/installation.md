# 安装指南

在几分钟内即可在系统上安装并运行 MoAI-ADK。本指南涵盖系统要求、安装方法和验证步骤。

## 系统要求

### 最低要求

- **Python**: 3.13 或更高版本
- **操作系统**:
  - macOS (10.15+)
  - Linux (Ubuntu 20.04+, CentOS 8+, Debian 11+)
  - Windows 10+ (推荐使用 PowerShell)
- **Git**: 2.25 或更高版本
- **内存**: 最低 4GB RAM，推荐 8GB
- **存储空间**: 500MB 可用空间

### 推荐配置

- **Python**: 3.13+ (最新稳定版本)
- **包管理器**: UV 0.5.0+ (推荐) 或 pip 24.0+
- **IDE**: 安装了 Claude Code 扩展的 VS Code 或您喜欢的编辑器
- **终端**: 支持 UTF-8 的现代终端

## 安装方法

### 方法 1: UV 包管理器 (推荐)

UV 是安装 MoAI-ADK 最快速、最可靠的方法。它提供自动依赖管理和虚拟环境处理。

#### 步骤 1: 安装 UV

**macOS/Linux:**

```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

**Windows (PowerShell):**

```powershell
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"
```

#### 步骤 2: 验证 UV 安装

```bash
uv --version
# 预期输出: uv 0.5.1 或更高版本
```

#### 步骤 3: 安装 MoAI-ADK

```bash
uv tool install moai-adk
```

#### 步骤 4: 验证安装

```bash
moai-adk --version
# 预期输出: MoAI-ADK v1.0.0 或更高版本
```

### 方法 2: PyPI 安装 (替代方案)

如果您使用 pip 或无法使用 UV。

#### 步骤 1: 升级 pip (如需要)

```bash
python -m pip install --upgrade pip
```

#### 步骤 2: 安装 MoAI-ADK

```bash
pip install moai-adk
```

#### 步骤 3: 验证安装

```bash
moai-adk --version
```

### 方法 3: 开发者安装

适用于希望为 MoAI-ADK 做贡献的开发者。

#### 步骤 1: 克隆仓库

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk
```

#### 步骤 2: 以开发模式安装

```bash
# 使用 UV (推荐)
uv pip install -e .

# 或使用 pip
pip install -e .
```

#### 步骤 3: 验证安装

```bash
moai-adk --version
```

## 安装后配置

### 环境变量

可选但推荐的环境变量:

```bash
# 添加到 shell 配置文件 (~/.bashrc, ~/.zshrc 等)
export MOAI_LOG_LEVEL=INFO
export MOAI_CACHE_DIR="$HOME/.moai/cache"
export CLAUDE_PROJECT_DIR=$(pwd)
```

### Claude Code 集成

MoAI-ADK 需要 Claude Code 才能获得完整体验。

#### 安装 Claude Code

```bash
# macOS
brew install claude-ai/claude/claude

# Linux
curl -fsSL https://claude.ai/install.sh | sh

# Windows
winget install Anthropic.Claude
```

#### 验证 Claude Code

```bash
claude --version
# 预期: Claude Code v1.5.0 或更高版本
```

### 可选 MCP 服务器

MoAI-ADK 支持 Model Context Protocol (MCP) 服务器以增强功能。

#### 安装推荐的 MCP 服务器

```bash
# Context7 - 最新库文档
npx -y @upstash/context7-mcp

# Playwright - Web E2E 测试
npx -y @playwright/mcp

# Sequential Thinking - 复杂推理
npx -y @modelcontextprotocol/server-sequential-thinking
```

## 验证

### 检查系统状态

运行内置的 doctor 命令以验证您的安装:

```bash
moai-adk doctor
```

**预期输出:**

```
正在运行系统诊断...

┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━┓
┃ 检查项                                    ┃ 状态   ┃
┡━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━┩
│ Python >= 3.13                            │   ✓    │
│ uv 已安装                                 │   ✓    │
│ Git 已安装                                │   ✓    │
│ Claude Code 可用                          │   ✓    │
│ 可访问包注册表                            │   ✓    │
└───────────────────────────────────────────┴────────┘

✅ 所有检查通过!
```

### 创建测试项目

创建一个简单的测试项目以确认一切正常:

```bash
# 创建测试项目
moai-adk init test-project
cd test-project

# 启动 Claude Code
claude

# 在 Claude Code 中运行:
/alfred:0-project
```

## 故障排除

### 常见问题

#### 问题: "uv: command not found"

**解决方案:**

1. 确认 UV 已正确安装
2. 将 UV 添加到 PATH:
   ```bash
   export PATH="$HOME/.cargo/bin:$PATH"
   ```
3. 重启终端

#### 问题: "Python 3.8 found, but 3.13+ required"

**解决方案:**

```bash
# 使用 pyenv
curl https://pyenv.run | bash
pyenv install 3.13
pyenv global 3.13

# 或使用 UV
uv python install 3.13
uv python pin 3.13
```

#### 问题: 安装时 "Permission denied"

**解决方案:**

```bash
# 使用用户安装
pip install --user moai-adk

# 或使用 sudo (Linux/macOS)
sudo pip install moai-adk
```

#### 问题: 无法识别 Claude Code

**解决方案:**

1. 验证 Claude Code 安装: `claude --version`
2. 确认在 PATH 中
3. 如需要则重新安装

#### 问题: 依赖项的 ModuleNotFoundError

**解决方案:**

```bash
# 在项目目录中
uv sync

# 或安装特定依赖
uv add fastapi pytest
```

### 获取帮助

如果遇到此处未涵盖的问题:

1. **检查 GitHub Issues**: 在 https://github.com/modu-ai/moai-adk/issues 搜索现有问题
2. **运行详细诊断**: `moai-adk doctor --verbose`
3. **创建 Issue**: 在 Claude Code 中使用 `/alfred:9-feedback` 自动创建 GitHub issue

## 下一步

成功安装后:

1. **[快速入门指南](quick-start.md)** - 10 分钟内运行您的第一个项目
2. **[核心概念](concepts.md)** - 理解 SPEC-First、TDD、@TAG、TRUST 5 原则
3. **[项目初始化](../../guides/project/init.md)** - 学习项目设置和配置

## 安装总结

```bash
# 一行安装 (推荐)
curl -LsSf https://astral.sh/uv/install.sh | sh && uv tool install moai-adk

# 验证安装
moai-adk doctor

# 创建您的第一个项目
moai-adk init my-project && cd my-project && claude
```

现在您已准备好体验 Alfred 超级代理带来的 SPEC-First TDD 开发强大功能! 🚀
