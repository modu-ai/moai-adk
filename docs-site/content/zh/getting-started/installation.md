---
title: 安装
weight: 30
draft: false
---

本文介绍在系统上安装 MoAI-ADK 的方法。安装物只是一个用 Go 构建的单体二进制 — 不需要 Python，不需要虚拟环境，也不需要包管理器。

## 许可证

MoAI-ADK {{< version >}} 及以上版本在 **Apache-2.0 许可证**下分发。

可以自由地商业使用、修改、分发，且没有公开源代码的义务。详情请参考 [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0)。

{{< callout type="info" >}}
**提示**：MoAI-ADK 1.x（Python 版本）曾使用 GPL-3.0 许可证。自 v2.0.0 起用 Go 语言重写，并变更为 Apache-2.0。
{{< /callout >}}

## 前置要求

安装前请确认以下条目：

### 1. Claude Code

MoAI-ADK 是运行在 Claude Code 之上的扩展框架，须先安装 Claude Code。

```bash
claude --version
```

如果尚未安装，请参考 [Claude Code 官方文档](https://docs.anthropic.com/en/docs/claude-code)。

### 2. Git（必需）

MoAI-ADK 使用基于 Git 的工作流，系统上必须安装 Git。

```bash
git --version
```

{{< callout type="warning" >}}
**Windows 用户**：请务必在 **Git Bash** 或 **WSL** 环境中使用。不支持 Command Prompt (cmd.exe)。

如果未安装 Git：
- **Windows**：在 [git-scm.com](https://git-scm.com) 安装 Git for Windows，会同时安装 Git Bash。
- **macOS**：`xcode-select --install` 或 [git-scm.com](https://git-scm.com)
- **Linux**：`sudo apt install git`（Ubuntu/Debian）或 `sudo dnf install git`（Fedora）
{{< /callout >}}

### 系统要求

| 项目 | 要求 |
|------|---------|
| **操作系统** | macOS、Linux、Windows（Git Bash / WSL） |
| **架构** | amd64、arm64 |
| **内存** | 至少 4GB RAM |
| **磁盘** | 至少 100MB 可用空间 |

## 安装方法

### 方法 1：快速安装（推荐）

用一条命令自动安装最新版本。

**macOS / Linux / WSL / Git Bash：**

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

**Windows (PowerShell)：**

```powershell
irm https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.ps1 | iex
```

{{< callout type="info" >}}
安装脚本会自动检测平台，从 GitHub 下载预构建的二进制文件，验证 SHA256 校验和，并配置 PATH。不需要 Python 或其他运行时。
{{< /callout >}}

安装完成后进行确认：

```bash
moai version
```

#### 安装选项

```bash
# 특정 버전 설치 (원하는 릴리스 태그 지정)
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash -s -- --version <릴리스-태그>

# 커스텀 디렉터리에 설치
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash -s -- --install-dir /usr/local/bin
```

### 方法 2：从源码构建

如果有 Go 开发环境，也可以直接从源码构建。

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk
make build
```

构建出的二进制位于 `./bin/moai`。请复制到 PATH 中的位置：

```bash
cp ./bin/moai ~/.local/bin/
```

### 安装位置

安装脚本按以下顺序决定安装目录：

| 平台 | 优先级 |
|--------|---------|
| **macOS / Linux** | `$GOBIN` → `$GOPATH/bin` → `~/.local/bin` |
| **Windows** | `%LOCALAPPDATA%\Programs\moai` |

## 1.x 用户迁移

{{< callout type="error" >}}
**MoAI-ADK 1.x（Python 版本）用户务必先卸载旧版本。**

1.x 与 2.x 使用同样的 `moai` 命令，旧版本残留会导致冲突。
{{< /callout >}}

### 第 1 步：卸载旧的 1.x

```bash
# uv로 설치한 경우
uv tool uninstall moai-adk

# pip로 설치한 경우
pip uninstall moai-adk
```

### 第 2 步：备份旧设置（可选）

```bash
# 기존 설정을 백업하고 싶다면
cp -r ~/.moai ~/.moai-v1-backup
```

### 第 3 步：安装 2.x

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

### 第 4 步：确认安装

```bash
moai version
# 출력 예시: moai <버전> (commit: <해시>, built: <빌드 날짜>)
```

{{< callout type="info" >}}
Go 版本 (v2.0+) 是单体二进制，不需要 Python 运行时或虚拟环境。启动时间从约 800ms 大幅提升至 5ms。
{{< /callout >}}

## WSL 支持

为 Windows 用户介绍在 WSL (Windows Subsystem for Linux) 环境中的安装与使用方法。

### 安装 WSL

若未安装 WSL，请在 PowerShell（管理员权限）中执行以下命令：

```powershell
wsl --install
```

安装后重启 Windows，Ubuntu 会自动完成安装。

### 在 WSL 中安装 MoAI-ADK

在 WSL 终端中使用与 Linux 相同的命令：

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

### 路径处理

在 WSL 中需要区分 Windows 路径与 WSL 路径：

| Windows 路径 | WSL 路径 |
|-------------|----------|
| `C:\Users\name\project` | `/mnt/c/Users/name/project` |
| `D:\Projects\myapp` | `/mnt/d/Projects/myapp` |

{{< callout type="info" >}}
**推荐**：在 WSL 的 Linux 文件系统（`~/projects/`）中创建项目，I/O 性能提升 2-5 倍。访问 Windows 文件系统（`/mnt/c/`）可能导致性能下降。
{{< /callout >}}

### WSL 最佳实践

1. **使用 Linux 文件系统**：项目创建在 `~/projects/` 目录下
2. **配置 Git 凭据**：在 WSL 中单独配置 Git 凭据，与 Windows 分开
3. **推荐终端**：使用 Windows Terminal 管理多个 WSL 发行版

### WSL 问题排查

#### PATH 未加载

```bash
# ~/.bashrc 또는 ~/.zshrc에 추가
source ~/.cargo/env
export PATH="$HOME/.local/bin:$PATH"
```

#### 钩子/MCP 服务器执行权限问题

```bash
# 실행 권한 부여
chmod +x ~/.claude/hooks/moai/*.sh
```

#### 访问 Windows 路径速度慢

请把项目移动到 Linux 文件系统：

```bash
# Windows에서 WSL로 이동
cp -r /mnt/c/Users/name/project ~/projects/
cd ~/projects/project
```

## pip 与 uv 工具冲突

这是 MoAI-ADK 1.x（Python 版本）用户可能遇到的常见问题。

### 问题说明

pip 与 uv 会把包安装到不同位置。混用两个工具时，`moai` 命令可能执行到意料之外的版本。

### 症状

- 执行 `moai version` 时显示 1.x 版本
- 出现 `command not found: moai` 错误
- 从与 `which moai` 不同的路径执行

### 原因

1. pip 安装到系统 Python 路径
2. uv tool 安装到 `~/.local/bin` 或 `~/.cargo/bin`
3. 根据 PATH 顺序执行到不同版本

### 解决方法

#### 完全卸载后重装

```bash
# 1. 모든 기존 버전 제거
uv tool uninstall moai-adk 2>/dev/null || true
pip uninstall moai-adk -y 2>/dev/null || true

# 2. 남은 바이너리 확인 및 삭제
which moai && rm $(which moai) 2>/dev/null || true
ls ~/.local/bin/moai && rm ~/.local/bin/moai 2>/dev/null || true

# 3. 2.x 설치
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash

# 4. 확인
moai version
```

#### 更新 shell 配置

```bash
# ~/.bashrc 또는 ~/.zshrc에 추가
export PATH="$HOME/.local/bin:$PATH"

# 설정 적용
source ~/.bashrc  # 또는 source ~/.zshrc
```

### 预防方法

1. MoAI-ADK 2.x 是与 Python 无关的 Go 二进制
2. 卸载 1.x（Python 版本）之后再安装 2.x
3. 不要同时使用 pip 与 uv tool

## 安装问题排查

### 问题：找不到命令

```bash
command not found: moai
```

**解决方法：**

1. 重启终端
2. 确认 PATH 设置：

```bash
echo $PATH
```

3. 确认二进制的安装位置：

```bash
which moai || ls ~/.local/bin/moai
```

4. 手动添加到 PATH：

```bash
# Bash/Zsh
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### 问题：权限被拒绝

```bash
Permission denied
```

**解决方法：**

```bash
chmod +x ~/.local/bin/moai
```

### 问题：1.x 与 2.x 冲突

若执行到旧版本的 `moai` 命令：

```bash
# 어떤 moai가 실행되는지 확인
which moai

# 1.x가 남아있다면 제거
uv tool uninstall moai-adk
# 또는
pip uninstall moai-adk

# 터미널 재시작 후 2.x 확인
moai version
```

## 安装后的下一步

安装完成后请初始化项目：

### 创建新项目

```bash
moai init my-project
```

### 应用到现有项目

```bash
cd my-existing-project
moai init
```

## 升级

要升级到最新版本：

```bash
moai update
```

### 更新选项

```bash
# 버전 확인만 (업데이트 안 함)
moai update --check

# 템플릿 동기화만 (패키지 업그레이드 건너뜀)
moai update --templates-only

# 설정 편집 모드 (초기화 마법사 다시 실행)
moai update --config
moai update -c

# 백업 없이 강제 업데이트
moai update --force

# 자동 승인 모드 (모든 확인 자동 승인)
moai update --yes
```

### 合并策略

```bash
# 자동 병합 강제 (기본값)
moai update --merge

# 수동 병합 강제
moai update --manual
```

{{< callout type="info" >}}
**自动保留项目**：用户设置、自定义智能体、自定义命令、自定义技能、自定义钩子、SPEC 文档、报告在更新时会自动保留。
{{< /callout >}}

详细内容请参考[更新指南](https://adk.mo.ai.kr/getting-started/update)。

## 卸载

要完全卸载 MoAI-ADK，请删除二进制与设置目录：

```bash
# 바이너리 삭제 (which moai 결과로 삭제)
rm "$(which moai)"

# 설정 디렉토리 삭제 (선택사항)
rm -rf "$HOME/.moai"
```

---

## 下一步

在[初始设置向导](./init-wizard)中了解 MoAI-ADK 的配置方法。
