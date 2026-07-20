---
title: moai loop 反馈循环
weight: 76
draft: false
---

`moai loop` 管理 SPEC 生命周期 Ralph 反馈循环控制器。它控制针对单个 SPEC 反复处理工具诊断所检测出的工作的状态机。

> 此 CLI 命令与 Claude Code 对话窗口的 `/moai loop` 技能是分开的 —— CLI 操作循环控制器的状态,而 `/moai loop` 技能由编排器执行实际的反复修复。

## 子命令

| 命令 | 说明 |
|--------|------|
| `moai loop start <SPEC-ID>` | 为 SPEC 启动反馈循环 |
| `moai loop status` | 显示当前循环状态 |
| `moai loop pause` | 暂停正在运行的循环 |
| `moai loop resume <SPEC-ID>` | 恢复被暂停的循环 |
| `moai loop cancel` | 取消正在运行的循环 |

## 示例

```bash
# 为 SPEC 启动循环
moai loop start SPEC-AUTH-001

# 检查当前状态
moai loop status

# 暂停后稍后恢复
moai loop pause
moai loop resume SPEC-AUTH-001

# 取消循环
moai loop cancel
```

## 相关文档

- [自主连续循环](/zh/advanced/autonomous-loops) —— Ralph 引擎与基于目标的循环
- [moai goal](/zh/cli-reference/goal)
- [CLI 概览](/zh/getting-started/cli)
