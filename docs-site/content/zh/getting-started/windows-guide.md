---
title: Windows使用指南
weight: 40
draft: false
---

## 支持的环境

| 环境 | 是否支持 | 备注 |
|------|----------|------|
| **WSL（推荐）** | ✅ 完全支持 | 最佳体验 |
| **PowerShell 7.x+** | ✅ 支持 | 备选环境 |
| PowerShell 5.x（旧版） | ❌ 不支持 | Windows PowerShell |
| cmd.exe | ❌ 不支持 | 命令提示符 |

**必要条件：**
- 必须安装 [Git for Windows](https://gitforwindows.org/)
- WSL 或 PowerShell 7.x 以上版本

## 安装方法

### WSL（推荐）

WSL在Windows上提供Linux环境，完整支持MoAI-ADK的所有功能。

```bash
# 安装WSL（在管理员PowerShell中执行）
wsl --install

# 在WSL内安装MoAI-ADK
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh \
  | bash
```

### PowerShell 7.x+

> **注**：为获得最佳体验，建议使用WSL。

```powershell
irm https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.ps1 | iex
```

## 非ASCII用户名路径错误

### 问题现象

如果Windows用户名包含中文、韩文等非ASCII字符，可能会发生 `EINVAL` 错误。这是Windows的8.3短文件名转换过程中产生的问题。

```
Error: EINVAL: invalid argument, open 'C:\Users\王伟\AppData\Local\Temp\...'
```

### 解决方法1：设置备用临时目录（推荐）

在仅包含ASCII字符的路径下创建临时目录：

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

要永久设置该环境变量，请将 `MOAI_TEMP_DIR` 添加到系统环境变量中。

### 解决方法2：禁用8.3文件名生成

以管理员权限执行：

```bash
fsutil 8dot3name set 1
```

> **注意**：此设置会影响整个系统，部分旧版程序可能会受到影响。

### 解决方法3：创建ASCII用户账户

使用英文名称创建新的Windows用户账户，可从根本上解决路径问题。

## WSL设置指南

### 安装WSL

```powershell
# 在管理员PowerShell中执行
wsl --install

# 默认发行版：Ubuntu（推荐）
# 重启后设置用户名和密码
```

### 项目文件访问

在WSL中访问Windows文件：

```bash
# 访问Windows文件系统
cd /mnt/c/Users/用户名/projects/

# 使用WSL原生文件系统（更快）
cd ~/projects/
```

> **性能提示**：在WSL原生文件系统（`~/` 下）中工作，可以在没有跨文件系统开销的情况下获得最佳性能。

### VS Code集成

1. 在VS Code中安装 [WSL扩展](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-wsl)
2. 在WSL终端中执行 `code .`
3. VS Code会自动以WSL模式打开

## 在CG模式下使用tmux

使用 [CG模式](/zh/multi-llm/cg-mode) 需要tmux。在WSL中安装：

```bash
# Ubuntu/Debian
sudo apt install tmux

# 启动tmux会话
tmux new -s moai

# 运行CG模式
moai cg
```

## 故障排除

| 问题 | 原因 | 解决方法 |
|------|------|------|
| `moai: command not found` | PATH中未包含Go bin目录 | 在 `.bashrc` 中添加 `export PATH="$HOME/go/bin:$PATH"` |
| `EINVAL` 错误 | 非ASCII用户名 | 参见上文 [非ASCII用户名路径错误](#非ascii用户名路径错误) |
| 权限被拒绝 | 安装脚本权限问题 | 执行 `chmod +x install.sh` 后重新运行 |
| Git命令失败 | 未安装Git for Windows | 安装 [Git for Windows](https://gitforwindows.org/) |
| 找不到tmux | 无法运行CG模式 | 执行 `sudo apt install tmux`（在WSL中） |

## 下一步

- [安装](/zh/getting-started/installation) — 详细安装指南
- [初始设置](/zh/getting-started/init-wizard) — 项目初始化
- [CG模式](/zh/multi-llm/cg-mode) — Claude + GLM混合模式
