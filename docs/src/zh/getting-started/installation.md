______________________________________________________________________

## title: 安装指南 description: 完整的 MoAI-ADK 安装和配置指南，支持多种操作系统和 Python 版本

# 安装指南

本指南将帮助您在各种操作系统上安装 MoAI-ADK。

## 系统要求

### 最低要求

- **Python**: 3.13 或更高版本
- **操作系统**: Windows 10+、macOS 10.15+、Ubuntu 20.04+ 或 equivalent
- **Git**: 2.25 或更高版本
- **内存**: 最少 4GB RAM（推荐 8GB+）
- **存储**: 至少 1GB 可用空间

### 推荐配置

- **Python**: 3.13（最新稳定版）
- **uv**: 0.5.0 或更高版本
- **Claude Code**: 1.5.0 或更高版本
- **IDE**: VS Code、PyCharm 或其他支持 Python 的编辑器

______________________________________________________________________

## 方法一：使用 UV（推荐）

UV 是一个现代 Python 包管理器，安装速度快，依赖管理优秀。

### 步骤 1：安装 UV

#### macOS 和 Linux

```bash
# 官方安装脚本
curl -LsSf https://astral.sh/uv/install.sh | sh

# 或使用 wget
wget -qO- https://astral.sh/uv/install.sh | sh
```

#### Windows (PowerShell)

```powershell
# 官方安装脚本
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"
```

#### 验证 UV 安装

```bash
# 检查版本
uv --version

# 输出示例: uv 0.5.1
```

### 步骤 2：安装 MoAI-ADK

```bash
# 安装 MoAI-ADK
uv tool install moai-adk

# 验证安装
moai-adk --version
```

### 步骤 3：配置 PATH（如需要）

如果命令找不到，请添加 uv 工具目录到 PATH：

```bash
# macOS/Linux
export PATH="$HOME/.cargo/bin:$PATH"

# Windows (PowerShell)
$env:PATH = "$env:USERPROFILE\.cargo\bin;$env:PATH"
```

______________________________________________________________________

## 方法二：使用 Pip

如果您更喜欢使用传统的 pip 包管理器。

### 步骤 1：确保 Python 版本

```bash
# 检查 Python 版本
python --version

# 需要 3.13 或更高版本
# 如果不是，请先升级 Python
```

### 步骤 2：创建虚拟环境

```bash
# 创建虚拟环境
python -m venv moai-adk-env

# 激活虚拟环境
# macOS/Linux:
source moai-adk-env/bin/activate
# Windows:
moai-adk-env\Scripts\activate
```

### 步骤 3：安装 MoAI-ADK

```bash
# 升级 pip
pip install --upgrade pip

# 安装 MoAI-ADK
pip install moai-adk

# 验证安装
moai-adk --version
```

______________________________________________________________________

## 方法三：从源码安装

如果您想安装开发版本或贡献代码。

### 步骤 1：克隆仓库

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk
```

### 步骤 2：安装开发依赖

```bash
# 使用 uv（推荐）
uv sync --dev

# 或使用 pip
pip install -e ".[dev]"
```

### 步骤 3：验证安装

```bash
# 运行测试
pytest

# 检查安装
python -m moai_adk --version
```

______________________________________________________________________

## 安装后验证

### 运行系统诊断

```bash
# 完整系统检查
moai-adk doctor

# 详细输出
moai-adk doctor --verbose
```

**预期输出**：

```
Running system diagnostics...

┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━┓
┃ Check                                    ┃ Status ┃
┡━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━┩
│ Python >= 3.13                           │   ✓    │
│ uv installed                            │   ✓    │
│ Git installed                            │   ✓    │
│ Claude Code installed                   │   ✓    │
│ MoAI-ADK package                        │   ✓    │
└──────────────────────────────────────────┴────────┘

✓ All checks passed
```

### 验证命令可用性

```bash
# 测试核心命令
moai-adk --help
moai-adk init --help
moai-adk doctor --help
```

______________________________________________________________________

## Claude Code 设置

MoAI-ADK 需要 Claude Code 来运行 AI 代理。

### 安装 Claude Code

#### 官方安装方法

```bash
# macOS (Homebrew)
brew install claude-code

# 其他系统请访问官方文档
# https://docs.claude.com/installation
```

#### 验证安装

```bash
# 检查版本
claude --version

# 需要 1.5.0 或更高版本
```

### 配置 Claude Code

```bash
# 启动 Claude Code
claude

# 检查设置
claude-code settings
```

______________________________________________________________________

## MCP 服务器配置

MoAI-ADK 自动配置 4 个核心 MCP 服务器以增强 AI 功能。

### 自动配置

创建新项目时，MCP 服务器会自动配置：

```bash
# 包含 MCP 服务器的项目初始化
moai-adk init my-project --with-mcp
```

### 手动配置

如果需要手动配置：

```bash
# 进入现有项目
cd your-project

# 添加 MCP 支持
moai-adk init . --with-mcp
```

### 验证 MCP 配置

```bash
# 检查 MCP 配置文件
cat .claude/mcp.json

# 重启 Claude Code
exit
claude
```

______________________________________________________________________

## 常见安装问题

### 问题 1：uv 命令未找到

**症状**：

```bash
uv: command not found
```

**解决方案**：

```bash
# 1. 确认安装路径
ls -la ~/.cargo/bin/uv

# 2. 添加到 PATH（临时）
export PATH="$HOME/.cargo/bin:$PATH"

# 3. 永久添加到 shell 配置
echo 'export PATH="$HOME/.cargo/bin:$PATH"' >> ~/.bashrc
# 或
echo 'export PATH="$HOME/.cargo/bin:$PATH"' >> ~/.zshrc

# 4. 重新加载 shell
source ~/.bashrc  # 或 source ~/.zshrc
```

### 问题 2：Python 版本不兼容

**症状**：

```bash
Python 3.12 found, but 3.13+ required
```

**解决方案**：

#### 使用 pyenv 管理 Python 版本

```bash
# 安装 pyenv
curl https://pyenv.run | bash

# 配置 shell
echo 'export PYENV_ROOT="$HOME/.pyenv"' >> ~/.bashrc
echo 'command -v pyenv >/dev/null || export PATH="$PYENV_ROOT/bin:$PATH"' >> ~/.bashrc
echo 'eval "$(pyenv init -)"' >> ~/.bashrc

# 重新加载 shell
source ~/.bashrc

# 安装 Python 3.13
pyenv install 3.13.0
pyenv global 3.13.0

# 验证
python --version
```

#### 使用 uv 管理 Python

```bash
# 安装 Python 3.13
uv python install 3.13

# 为项目设置 Python 版本
uv python pin 3.13

# 验证
python --version
```

### 问题 3：权限错误

**症状**：

```bash
Permission denied: '/usr/local/bin/moai-adk'
```

**解决方案**：

```bash
# 使用用户安装
pip install --user moai-adk

# 或使用 uv（自动处理权限）
uv tool install moai-adk
```

### 问题 4：网络连接问题

**症状**：

```bash
Could not fetch URL https://pypi.org/simple/moai-adk/
```

**解决方案**：

```bash
# 使用国内镜像源
pip install -i https://pypi.tuna.tsinghua.edu.cn/simple moai-adk

# 或配置 pip 镜像源
pip config set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple
```

### 问题 5：Windows 权限问题

**症状**：

```bash
'uv' is not recognized as an internal or external command
```

**解决方案**：

```powershell
# 1. 以管理员身份运行 PowerShell
# 2. 设置执行策略
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# 3. 添加到 PATH
$env:PATH = "$env:USERPROFILE\.cargo\bin;$env:PATH"

# 4. 永久添加到系统 PATH
# 系统属性 → 环境变量 → PATH → 添加 %USERPROFILE%\.cargo\bin
```

______________________________________________________________________

## 升级 MoAI-ADK

### 检查当前版本

```bash
moai-adk --version
```

### 升级到最新版本

#### 使用 UV

```bash
# 升级 MoAI-ADK
uv tool upgrade moai-adk

# 验证升级
moai-adk --version
```

#### 使用 Pip

```bash
# 升级包
pip install --upgrade moai-adk

# 验证升级
moai-adk --version
```

### 同步项目模板

升级后，同步现有项目的模板：

```bash
# 进入项目目录
cd your-project

# 同步模板
moai-adk update

# 验证同步
moai-adk doctor
```

______________________________________________________________________

## 卸载

### 卸载 MoAI-ADK

#### 使用 UV

```bash
# 卸载包
uv tool uninstall moai-adk

# 清理缓存
uv cache clean
```

#### 使用 Pip

```bash
# 卸载包
pip uninstall moai-adk

# 清理缓存
pip cache purge
```

### 删除配置文件

```bash
# 删除全局配置（可选）
rm -rf ~/.config/moai-adk/

# 删除项目特定配置（谨慎操作）
# rm -rf .moai/ .claude/
```

______________________________________________________________________

## 下一步

安装完成后，您可以：

1. [创建第一个项目](quick-start.md)
2. [学习核心概念](../guides/concepts.md)
3. [查看命令参考](../reference/cli/)

______________________________________________________________________

## 获取帮助

如果在安装过程中遇到问题：

- **GitHub Issues**: [moai-adk/issues](https://github.com/modu-ai/moai-adk/issues)
- **GitHub Discussions**: [moai-adk/discussions](https://github.com/modu-ai/moai-adk/discussions)
- **文档**: [完整文档](../../)

### 报告问题时请包含

1. 操作系统和版本
2. Python 版本
3. MoAI-ADK 版本
4. 完整的错误消息
5. `moai-adk doctor --verbose` 输出
