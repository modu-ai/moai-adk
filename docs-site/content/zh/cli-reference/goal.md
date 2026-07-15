---
title: moai goal 目标循环
weight: 72
draft: false
---

`moai goal` 针对当前会话 arm/查询/解除一个声明了条件的 agentic 目标循环。MoAI goal 引擎让会话跨越多个回合持续工作,直到所声明的条件满足或达到回合上限。

这是原生 `/goal`(仅用户可用的 TUI 命令)的编程式 MoAI 对应物,使编排器无需人工输入 `/goal` 行即可注册并 arm 目标。

## 子命令

| 命令 | 说明 |
|--------|------|
| `moai goal arm "<condition>"` | 向活动会话注册 + arm 目标(`moai goal "<condition>"` 也是 arm 的别名) |
| `moai goal status` | 输出活动会话的目标状态 |
| `moai goal clear` | 解除活动会话的目标 |

## 公共标志

| 标志 | 说明 |
|--------|------|
| `--session <id>` | 覆盖会话 id(默认:由 `moai session current` 解析) |
| `--json` | 机器可读 JSON 输出 |
| `--all` | (仅 `status`)不仅列出活动会话,还列出所有会话的目标 |

## 状态与评估

目标状态保存在 `.moai/state/goal/<session-id>.json`(每个会话 1 个文件)。Stop 钩子 `moai hook stop-goal` 在每个回合结束时评估目标。

**条件解析**:

- 可执行的 shell 命令(可选带 `exits <N>` 后缀)会成为**机械式(mechanical)条件**。
- 引用对话 transcript 的断言会成为由编排器评估的**模型(model)条件**。

## 示例

```bash
# 持续工作直到测试套件通过
moai goal arm "go test ./... exits 0"

# 检查当前目标状态
moai goal status

# 解除目标
moai goal clear
```

## 相关文档

- [自主连续循环](/zh/advanced/autonomous-loops) —— `/goal` 与 `/moai loop` 对比
- [moai loop](/zh/cli-reference/loop)
- [CLI 概览](/zh/getting-started/cli)
