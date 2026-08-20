---
title: moai tokens
weight: 17
draft: false
new: true
added_in: "v3.1.1"
---

{{< new-badge v3.1.1 >}}

把 Claude Code 会话的 token 用量**按池 (pool) 与来源 (origin) 拆开**记录的记账工具。“这张卡片花了多少 token”如果只看会话总量一个数字，哪个后端花了多少、主对话和子智能体侧链各花了多少都会被埋掉。这条命令把这两层拆开留下。

{{< callout type="info" >}}
**一句话总结**: `moai tokens record` 读取会话转录，按池（glm/claude/other）、按来源（主对话/子智能体侧链）汇总量，附上上下文使用快照，以一行 append 到 `.moai/state/token-accounting.jsonl`。
{{< /callout >}}

## 用法

```bash
# 指定打开中/上一次的会话转录进行记录
$ moai tokens record --transcript <路径> --card t12 --role run

# 用会话 id 指定
$ moai tokens record --session <会话-ID> --card t12

# 以 JSON 输出记录 (同时写入文件)
$ moai tokens record --transcript <路径> --json
```

| 标志 | 说明 |
|--------|------|
| `--transcript <路径>` | 要汇总的 Claude Code 转录文件 |
| `--session <id>` | 用会话标识符指定转录 |
| `--card <卡片>` | 把这笔用量归入的看板卡片（如 `t12`） |
| `--role <角色>` | 会话的角色（如 `run`, `sync`, `worker-3`） |
| `--json` | 同时把记录以 JSON 输出到标准输出 |

## 记录的样子

记录以 **append-only** 方式堆在 `.moai/state/token-accounting.jsonl` —— 会话、卡片结束时各留一行的账本。每行包含:

- **按池用量** — 分成 `glm` / `claude` / `other` 的合计。哪个后端制造了账单，从池上一眼看出来。
- **按来源用量** — 主对话与子智能体侧链。在开着多个 worker 的 run 里，它能分清“实现的那份是不是全被 worker 花掉了”。
- **上下文快照** — 记录时点若存在上下文使用状态（`.moai/state/context-usage.json`），该值一并进入。

## 什么时候记录

按设计，这是卡片或会话**收尾时点**的记录。看板 run 里每结束一张卡片记一次、单会话里结束一件大工作时记一次，卡片之间的成本比较才能成立。这条命令本身不消耗 token —— 它是对已发生用量从转录里重新清点的记账。

## 相关文档

- [token 经济学概览](/zh/advanced/tokenomics-overview) — 为什么指派比单价更重要
- [状态栏](/zh/advanced/statusline) — 会话进行中查看用量的位置
- [看板模式](/zh/advanced/kanban-mode) — 以卡片、泳道为单位归集成本的 run 形态
