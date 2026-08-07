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
| `moai goal arm "<condition>"` | 向活动会话注册 + arm 目标(`moai goal "<condition>"` 也是 arm 的别名)。arm-only —— 其自身不会启动任何工作 |
| `moai goal status` | 输出活动会话的目标状态(用 `--all` 列出所有会话) |
| `moai goal clear` | 解除活动会话的目标 |
| `moai goal render` | 将活动会话的目标仪表盘渲染为 self-contained HTML 文件(保存到 `.moai/state/goal/` 旁)。如果没有已 arm 的 goal,以非零退出码结束。从 v3.1(PR #1388)起,仪表盘会显示判定段(天花板 exit 时从 sidecar 加载)和重新武装条件视图 — 详情参见 [/moai goal 仪表盘章节](/zh/utility-commands/moai-goal#目标看板) |

## 公共标志

| 标志 | 说明 |
|--------|------|
| `--session <id>` | 覆盖会话 id(默认:由 `moai session current` 解析) |
| `--json` | 机器可读 JSON 输出 |
| `--all` | (仅 `status`)不仅列出活动会话,还列出所有会话的目标 |

## arm 旗标

| 旗标 | 说明 |
|--------|------|
| `--max-turns <N>` | 回合上限。`0` = 无限(SPEC-INFINITE-GOAL-001);省略时默认 `30`(完全向后兼容)。**`0`(无限)必须搭配 `--max-duration <sec>`**(arm 时刻 fail-closed)。 |
| `--max-duration <sec>` | 实时上限(arm 时刻起的秒数)。**无限 goal(`--max-turns 0`)的实际墙上时间上限** —— 没有 this 标志无法 arm 无限 goal。 |
| `--cost-cap <value>` | 调用次数上限,**仅记录(recorded-only)** —— 当前没有强制执行逻辑,因此并不是实际 bound。它无法满足 `--max-turns 0` 的实际 bound 要求,因此 cost-cap 单独使用时会被拒绝。 |

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
