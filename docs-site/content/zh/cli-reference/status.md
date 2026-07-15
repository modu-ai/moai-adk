---
title: 项目状态
weight: 20
draft: false
---

用 `moai status` 命令一目了然地查看当前项目的初始化状态、SPEC 数量、配置文件情况。这是一个无标志的只读命令。

## 用法

```bash
moai status
```

不带标志运行时,会以方框形式输出项目状态。

## 输出内容

### 已初始化的项目

在存在 `.moai/` 目录的项目中运行时,会显示以下信息。

| 项目 | 说明 |
|------|------|
| **Project** | 项目名称(当前目录名) |
| **ADK** | 已安装的 MoAI-ADK 版本 |
| **Config** | 配置文件路径(`.moai/config/sections`) |
| **SPECs** | `.moai/specs/` 下的 SPEC 目录数量 |
| **Configs** | `.moai/config/sections/` 中的 YAML 文件数量 |

底部会一并输出表示初始化状态和 SPEC 数量的状态标识。

### 未初始化的项目

若不存在 `.moai/` 目录,会显示 "Not initialized" 状态标识并引导执行 `moai init`。

## BODP 分支提醒

若项目是 Git 仓库,当当前分支偏离 BODP (Branch-Oriented Development Practice) 规约时,会在 stderr 输出提醒。该提醒是在分散式单人 OSS 工作流中提示分支命名规则的装置。

提醒会自动输出;若未安装 Git 或当前目录不是 Git 仓库,则静默省略。

## 相关命令

| 命令 | 说明 |
|--------|------|
| `moai doctor` | 系统诊断与环境验证(详细检查) |
| `moai inventory` | 活跃会话、工作树、harness 综合查询 |
| `moai init` | 项目初始化(未初始化时执行) |

## 参考

- [CLI 参考](./cli) — 全部 CLI 命令
- [moai inventory](./inventory) — 活跃资源综合查询
- [初始设置](./init-wizard) — 项目初始化向导
