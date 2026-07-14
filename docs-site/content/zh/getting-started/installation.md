---
title: 安装
weight: 30
draft: false
---

介绍在系统中安装 MoAI-ADK 的方法。安装物是用 Go 构建的单一二进制 —— 无需 Python、虚拟环境或包管理器。

## 许可证

MoAI-ADK {{< version >}} 及以上在 **Apache-2.0 许可证** 下分发。

可自由用于商业、修改、分发,且无公开源代码的义务。详情请参阅 [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0)。

{{< callout type="info" >}}
**参考**:MoAI-ADK 1.x(Python 版本)采用 GPL-3.0 许可证。从 v2.0.0 起用 Go 语言重写并改为 Apache-2.0。
{{< /callout >}}

## 前置要求

安装前请确认以下项目:

### 1. Claude Code

MoAI-ADK 是运行在 Claude Code 之上的扩展框架。必须先安装 Claude Code。

```bash
claude --version
```

若尚未安装,请参阅 [Claude Code 官方文档](https://docs.anthropic.com/en/docs/claude-code)。

### 2. Git(必需)

MoAI-ADK 使用基于 Git 的工作流。系统中必须安装 Git。

```bash
git --version
```

{{< callout type="warning" >}}
**Windows 用户**:请务必在 **Git Bash** 或 **WSL** 环境中使用。不支持 Command Prompt(cmd.exe)。

若未安装 Git:
- **Windows**:在 [git-scm.com](https://git-scm.com) 安装 Git for Windows。会一并安装 Git Bash。
- **macOS**:`xcode-select --install` 或 [git-scm.com](https://git-scm.com)
- **Linux**:`sudo apt install git`(Ubuntu/Debian)或 `sudo dnf install git`(Fedora)
{{< /callout >}}

### 系统要求

| 项目 | 要求 |
|------|---------|
| **操作系统** | macOS, Linux, Windows(Git Bash / WSL) |
| **架构** | amd64, arm64 |
| **内存** | 至少 4GB RAM |
| **磁盘** | 至少 100MB 可用空间 |

## 安装方法

### 方法 1:快速安装(推荐)

用一条命令自动安装最新版本。

**macOS / Linux / WSL / Git Bash:**

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

**Windows(PowerShell):**

```powershell
irm https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.ps1 | iex
```

{{< callout type="info" >}}
安装脚本会自动检测平台、从 GitHub 下载预构建的二进制、验证 SHA256 校验和并配置 PATH。不需要 Python 或额外的运行时。
{{< /callout >}}

安装完成后请确认:

```bash
moai version
```

#### 安装选项

```bash
# 安装特定版本(指定所需的 release 标签)
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash -s -- --version <release-tag>

# 安装到自定义目录
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash -s -- --install-dir /usr/local/bin
```

### 方法 2:从源码构建

有 Go 开发环境时可直接从源码构建。

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk
make build
```

构建的二进制生成在 `./bin/moai`。请复制到 PATH 指定的位置:

```bash
cp ./bin/moai ~/.local/bin/
```

### 安装位置

安装脚本按以下顺序决定安装目录:

| 平台 | 优先级 |
|--------|---------|
| **macOS / Linux** | `$GOBIN` → `$GOPATH/bin` → `~/.local/bin` |
| **Windows** | `%LOCALAPPDATA%\Programs\moai` |

## 1.x 用户迁移

{{< callout type="error" >}}
**MoAI-ADK 1.x(Python 版本)用户必须先卸载既有版本。**

1.x 与 2.x 使用相同的 `moai` 命令,若残留旧版本会发生冲突。
{{< /callout >}}

### 第 1 步:卸载既有 1.x

```bash
# 用 uv 安装的情况
uv tool uninstall moai-adk

# 用 pip 安装的情况
pip uninstall moai-adk
```

### 第 2 步:备份既有配置(可选)

```bash
# 若想备份既有配置
cp -r ~/.moai ~/.moai-v1-backup
```

### 第 3 步:安装 2.x

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

### 第 4 步:确认安装

```bash
moai version
# 输出示例: moai <版本> (commit: <哈希>, built: <构建日期>)
```

{{< callout type="info" >}}
Go 版(v2.0+)是单一二进制,不需要 Python 运行时或虚拟环境。启动时间从约 800ms 大幅提升到 5ms。
{{< /callout >}}

## WSL 支持

为 Windows 用户介绍在 WSL(Windows Subsystem for Linux)环境下的安装与使用方法。

### WSL 安装

若未安装 WSL,请在 PowerShell(管理员权限)中运行以下命令:

```powershell
wsl --install
```

安装后重启 Windows,会自动安装 Ubuntu。

### 在 WSL 中安装 MoAI-ADK

在 WSL 终端中使用与 Linux 相同的命令:

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

### 路径处理

在 WSL 中需要区分 Windows 路径与 WSL 路径:

| Windows 路径 | WSL 路径 |
|-------------|----------|
| `C:\Users\name\project` | `/mnt/c/Users/name/project` |
| `D:\Projects\myapp` | `/mnt/d/Projects/myapp` |

{{< callout type="info" >}}
**推荐**:把项目创建在 WSL 的 Linux 文件系统(`~/projects/`)中,I/O 性能会提升 2-5 倍。访问 Windows 文件系统(`/mnt/c/`)可能导致性能下降。
{{< /callout >}}

### WSL 最佳实践

1. **使用 Linux 文件系统**:项目创建在 `~/projects/` 目录
2. **配置 Git 凭据**:在 WSL 中单独于 Windows 配置 Git 凭据
3. **推荐终端**:使用 Windows Terminal 管理多个 WSL 发行版

### WSL 问题排查

#### PATH 未加载

```bash
# 添加到 ~/.bashrc 或 ~/.zshrc
source ~/.cargo/env
export PATH="$HOME/.local/bin:$PATH"
```

#### 钩子/MCP 服务器执行权限问题

```bash
# 赋予执行权限
chmod +x ~/.claude/hooks/moai/*.sh
```

#### Windows 路径访问速度下降

请将项目移动到 Linux 文件系统:

```bash
# 从 Windows 移动到 WSL
cp -r /mnt/c/Users/name/project ~/projects/
cd ~/projects/project
```

## pip 与 uv 工具冲突

这是 MoAI-ADK 1.x(Python 版本)用户可能遇到的常见问题。

### 问题说明

pip 与 uv 把包安装到不同的位置。混用两个工具时,`moai` 命令可能运行意料之外的版本。

### 症状

- 运行 `moai version` 时显示 1.x 版本
- 发生 `command not found: moai` 错误
- 在与 `which moai` 不同的路径运行

### 原因

1. pip 安装到系统 Python 路径
2. uv tool 安装到 `~/.local/bin` 或 `~/.cargo/bin`
3. 依 PATH 顺序运行不同版本

### 解决方法

#### 完全卸载后重新安装

```bash
# 1. 卸载所有既有版本
uv tool uninstall moai-adk 2>/dev/null || true
pip uninstall moai-adk -y 2>/dev/null || true

# 2. 确认并删除残留的二进制
which moai && rm $(which moai) 2>/dev/null || true
ls ~/.local/bin/moai && rm ~/.local/bin/moai 2>/dev/null || true

# 3. 安装 2.x
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash

# 4. 确认
moai version
```

#### 更新 shell 配置

```bash
# 添加到 ~/.bashrc 或 ~/.zshrc
export PATH="$HOME/.local/bin:$PATH"

# 应用配置
source ~/.bashrc  # 或 source ~/.zshrc
```

### 预防方法

1. MoAI-ADK 2.x 是与 Python 无关的 Go 二进制
2. 卸载 1.x(Python 版本)后再安装 2.x
3. 不要同时使用 pip 与 uv tool

## 安装问题排查

### 问题:找不到命令

```bash
command not found: moai
```

**解决方法:**

1. 重启终端
2. 确认 PATH 设置:

```bash
echo $PATH
```

3. 确认二进制安装位置:

```bash
which moai || ls ~/.local/bin/moai
```

4. 手动添加到 PATH:

```bash
# Bash/Zsh
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### 问题:权限拒绝

```bash
Permission denied
```

**解决方法:**

```bash
chmod +x ~/.local/bin/moai
```

### 问题:1.x 与 2.x 冲突

当运行的是旧版本 `moai` 命令时:

```bash
# 确认运行的是哪个 moai
which moai

# 若残留 1.x 则卸载
uv tool uninstall moai-adk
# 或
pip uninstall moai-adk

# 重启终端后确认 2.x
moai version
```

## 安装后的下一步

安装完成后请初始化项目:

### 创建新项目

```bash
moai init my-project
```

### 应用到既有项目

```bash
cd my-existing-project
moai init
```

## 升级

要升级到最新版本:

```bash
moai update
```

### 更新选项

```bash
# 仅确认版本(不更新)
moai update --check

# 仅同步模板(跳过二进制更新)
moai update --templates-only

# 配置编辑模式(重新运行初始化向导)
moai update -c

# 强制更新(用户变更备份后覆盖)
moai update --force

# 自动批准模式(CI/CD)
moai update --yes
```

{{< callout type="info" >}}
**自动保留项目**:用户设置、自定义智能体、自定义命令、自定义技能、自定义钩子、SPEC 文档、报告在更新时会自动保留。用户修改过的模板文件会备份后 3-way 合并。
{{< /callout >}}

详情请参阅[更新指南](https://adk.mo.ai.kr/getting-started/update)。

## 卸载

要完全卸载 MoAI-ADK,请删除二进制与配置目录:

```bash
# 删除二进制(用 which moai 的结果删除)
rm "$(which moai)"

# 删除配置目录(可选)
rm -rf "$HOME/.moai"
```

---

## 下一步

在[初始设置向导](./init-wizard)中了解 MoAI-ADK 的配置方法。
