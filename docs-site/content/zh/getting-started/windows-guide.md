---
title: Windows 使用指南
weight: 40
draft: false
---

本文整理了在 Windows 上使用 MoAI-ADK 时需要了解的环境要求与常见陷阱。先说结论：**WSL 最省心** — 原生 Windows 环境中遇到的大部分路径·权限问题在 WSL 中都不会出现。

## 支持的环境

| 环境 | 是否支持 | 备注 |
|------|----------|------|
| **WSL (推荐)** | {{< icon check ok >}} 完全支持 | 最佳体验 |
| **PowerShell 7.x+** | {{< icon check ok >}} 支持 | 备选环境 |
| PowerShell 5.x (旧版) | {{< icon x danger >}} 不支持 | Windows PowerShell |
| cmd.exe | {{< icon x danger >}} 不支持 | 命令提示符 |

**必要条件：**
- 必须安装 [Git for Windows](https://gitforwindows.org/)
- WSL 或 PowerShell 7.x 以上

## 安装方法

### WSL (推荐)

WSL 在 Windows 上提供 Linux 环境，可完整支持 MoAI-ADK 的全部功能。

```bash
# 安装 WSL（在管理员 PowerShell 中运行）
wsl --install

# 在 WSL 内安装 MoAI-ADK
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh \
  | bash
```

### PowerShell 7.x+

> **提示**：为获得最佳体验，建议使用 WSL。

```powershell
irm https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.ps1 | iex
```

## 非 ASCII 用户名路径错误

### 问题现象

当 Windows 用户名包含韩文、中文等非 ASCII 字符时，某些旧工具或在 8.3 短文件名转换过程中可能出现路径处理问题。若主目录路径混有非 ASCII 字符，特定命令可能失败。

```
C:\Users\홍길동\...
```

此时用以下方法准备一个纯 ASCII 路径环境最为可靠。

### 解决方法 1：启用 8.3 文件名生成

以管理员权限设置，使 8.3 短文件名（ASCII 替代路径）得以生成。

```powershell
fsutil 8dot3name set 1
```

> **注意**：此设置影响整个系统，部分旧程序可能受到影响。

### 解决方法 2：创建 ASCII 用户账户

用英文名创建新的 Windows 用户账户，可从根本上解决主目录路径问题。

### 解决方法 3：使用 WSL

最推荐的方法是在 WSL（见下文 [WSL 设置指南](#wsl-设置指南)）环境中工作。WSL 原生文件系统不受非 ASCII 主目录路径问题的影响。

## WSL 设置指南

### 安装 WSL

```powershell
# 在管理员 PowerShell 中运行
wsl --install

# 默认发行版：Ubuntu（推荐）
# 重启后设置用户名和密码
```

### 访问项目文件

在 WSL 中访问 Windows 文件：

```bash
# 访问 Windows 文件系统
cd /mnt/c/Users/用户名/projects/

# 使用 WSL 原生文件系统（更快）
cd ~/projects/
```

> **性能提示**：在 WSL 原生文件系统（`~/` 下）中工作，可以避免跨文件系统开销，获得最佳性能。

### VS Code 联动

1. 在 VS Code 中安装 [WSL 扩展](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-wsl)
2. 在 WSL 终端中运行 `code .`
3. VS Code 会自动以 WSL 模式打开

## 在 CG 模式中使用 tmux

使用 [CG 模式](/zh/multi-llm/cg-mode)需要 tmux。在 WSL 中安装：

```bash
# Ubuntu/Debian
sudo apt install tmux

# 启动 tmux 会话
tmux new -s moai

# 运行 CG 模式
moai cg
```

## 问题排查

| 问题 | 原因 | 解决 |
|------|------|------|
| `moai: command not found` | PATH 未包含安装目录 | 安装脚本安装到 `~/.local/bin` —— 在 `.bashrc` 中添加 `export PATH="$HOME/.local/bin:$PATH"`（用 `go install` 安装时为 `$HOME/go/bin`） |
| 韩文路径处理失败 | 韩文用户名 | 参考上文[非 ASCII 用户名路径错误](#非-ascii-用户名路径错误) |
| 权限被拒绝 | 安装脚本权限 | 执行 `chmod +x install.sh` 后重试 |
| Git 命令失败 | 未安装 Git for Windows | 安装 [Git for Windows](https://gitforwindows.org/) |
| 没有 tmux | 无法运行 CG 模式 | `sudo apt install tmux`（在 WSL 中） |

## 下一步

- [安装](/zh/getting-started/installation) — 安装详细指南
- [初始设置](/zh/getting-started/init-wizard) — 项目初始化
- [CG 模式](/zh/multi-llm/cg-mode) — Claude + GLM 混合模式
