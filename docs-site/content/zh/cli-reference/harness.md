---
title: moai harness 工具架
weight: 80
draft: false
---

`moai harness` 是管理 SPEC 复杂度路由与 harness 学习子系统的统合命令树。它提供路由、验证、生命周期、提案管理、v4 harness 生命周期、观测账本(ledger)子命令。

公共标志接受 `--project-root <path>`(默认:当前目录)。

## 路由 verb

| 命令 | 说明 |
|--------|------|
| `moai harness route --spec <id>` | 将 SPEC 路由至 minimal/standard/thorough harness 级别 |
| `moai harness validate` | 按 schema·不变量验证 harness.yaml |

`route` 接受 `--json`(JSON 输出)、`--path <harness.yaml>`、`--base-dir <dir>` 标志。

## 生命周期 verb

| 命令 | 说明 |
|--------|------|
| `moai harness status` | 显示观测/层级/进化摘要 |
| `moai harness apply` | 将待处理提案返回给编排器(或用 `--execute` 执行 Go apply 路径) |
| `moai harness rollback <date>` | 恢复指定日期的快照 |
| `moai harness disable` | 停用学习子系统(`learning.enabled: false`) |

## 提案管理 verb

| 命令 | 说明 |
|--------|------|
| `moai harness mute` | 静音提案类别(workflow.yaml) |
| `moai harness mute-list` | 输出当前被静音的类别 |
| `moai harness unmute` | 将类别从静音列表中移除 |
| `moai harness verify` | 验证 harness 确定性 |

## v4 harness 生命周期 verb

| 命令 | 说明 |
|--------|------|
| `moai harness list` | 列出所有 v4 harness(名称 + 领域 + 入口命令) |
| `moai harness edit <name>` | 显示 v4 harness manifest + specialist 编辑路径 |
| `moai harness remove <name>` | 原子化移除 v4 harness(命令 + workflow + specialist + skill + manifest) |
| `moai harness doctor` | 诊断 harness 安装状态 |

`list`、`edit` 接受 `--json` 标志。

## 观测账本(ledger)

`moai harness ledger` 管理路由观测账本。

| 命令 | 说明 |
|--------|------|
| `moai harness ledger record` | 记录 dispatch 时点的路由决策(pending row) |
| `moai harness ledger evidence` | 向 pending row 添加机器证据 ref(或委派项) |
| `moai harness ledger list` | 按过滤器列出最终账本 row |

> `ledger record` 与 `ledger evidence` 不暴露 `--outcome` 标志,以防伪造结果(outcome)。结果由机器证据导出。

## 相关文档

- [harness 自我进化](/zh/advanced/self-evolving)
- [Harness v4 Builder 进阶指南](/zh/advanced/harness-v4-builder)
- [harness 配置与评估](/zh/advanced/harness-profiles)
- [CLI 概览](/zh/getting-started/cli)
