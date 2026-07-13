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
# WSL 설치 (관리자 PowerShell에서 실행)
wsl --install

# WSL 내에서 MoAI-ADK 설치
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

当 Windows 用户名包含韩文、中文等非 ASCII 字符时，可能出现 `EINVAL` 错误。这是 Windows 8.3 短文件名转换过程引发的问题。

```
Error: EINVAL: invalid argument, open 'C:\Users\홍길동\AppData\Local\Temp\...'
```

### 解决方法 1：设置替代临时目录 (推荐)

在仅包含 ASCII 字符的路径下创建临时目录：

```bash
# Command Prompt
set MOAI_TEMP_DIR=C:\temp
mkdir C:\temp 2>/dev/null
```

```powershell
# PowerShell
$env:MOAI_TEMP_DIR="C:\temp"
New-Item -ItemType Directory -Path "C:\temp" -Force
```

若要永久设置环境变量，请在系统环境变量中添加 `MOAI_TEMP_DIR`。

### 解决方法 2：禁用 8.3 文件名生成

以管理员权限运行：

```bash
fsutil 8dot3name set 1
```

> **注意**：此设置影响整个系统，部分旧程序可能受到影响。

### 解决方法 3：创建 ASCII 用户账户

用英文名创建新的 Windows 用户账户，可从根本上解决路径问题。

## WSL 设置指南

### 安装 WSL

```powershell
# 관리자 PowerShell에서 실행
wsl --install

# 기본 배포판: Ubuntu (권장)
# 재시작 후 사용자명 및 비밀번호 설정
```

### 访问项目文件

在 WSL 中访问 Windows 文件：

```bash
# Windows 파일시스템 접근
cd /mnt/c/Users/사용자명/projects/

# WSL 네이티브 파일시스템 사용 (더 빠름)
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

# tmux 세션 시작
tmux new -s moai

# CG 모드 실행
moai cg
```

## 问题排查

| 问题 | 原因 | 解决 |
|------|------|------|
| `moai: command not found` | PATH 未包含 Go bin 目录 | 在 `.bashrc` 中添加 `export PATH="$HOME/go/bin:$PATH"` |
| `EINVAL` 错误 | 非 ASCII 用户名 | 参考上文[非 ASCII 用户名路径错误](#非-ascii-用户名路径错误) |
| 权限被拒绝 | 安装脚本权限 | 执行 `chmod +x install.sh` 后重试 |
| Git 命令失败 | 未安装 Git for Windows | 安装 [Git for Windows](https://gitforwindows.org/) |
| 没有 tmux | 无法运行 CG 模式 | `sudo apt install tmux`（在 WSL 中） |

## 下一步

- [安装](/zh/getting-started/installation) — 安装详细指南
- [初始设置](/zh/getting-started/init-wizard) — 项目初始化
- [CG 模式](/zh/multi-llm/cg-mode) — Claude + GLM 混合模式
