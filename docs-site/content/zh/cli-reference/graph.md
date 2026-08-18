---
title: moai graph
weight: 16
draft: false
new: true
added_in: "v3.1.1"
---

{{< new-badge v3.1.1 >}}

把代码库里的关系收进一份产物、用来回答**反向提问**的工具。“改这个包会波及到哪里”、“这个 SPEC 和实际代码连着吗” —— 这类问题 grep 答不了，必须把关系聚在一起才能回答。

{{< callout type="info" >}}
**一句话总结**: `moai graph build` 把散落在 codemaps、@MX 标签、SPEC、报告里的关系收进 `.moai/project/graph/edges.jsonl` 一个文件，`moai graph query` 对这份文件做反向查询。
{{< /callout >}}

## 为什么需要它

MoAI-ADK 已经在多个层面持有关系信息 —— codemaps 的导入图、代码里的 `@MX:SPEC` 标签、SPEC 文档的依赖声明、报告里的里程碑记录。问题在于这些层面各自散落在不同的文件里。想问“改这段代码会影响哪个 SPEC”，就必须把导入方向（导入带 `@MX:SPEC` 标签文件的那些文件）和 SPEC 依赖方向**放在同一张图里**反向追。edges.jsonl 就是这同一张图。

## moai graph build

```bash
$ moai graph build
```

汇总导入边、`@MX:SPEC` 连接、SPEC 间依赖，写入 `.moai/project/graph/edges.jsonl`。它在同一个 git HEAD 上跑两次会产出相同内容 —— 是确定性的。查询永远读这份产物，所以**查询之前要先跑一次 build**。

## moai graph query

一次调用只给**恰好一个**选择器。

| 选择器 | 提问 | 回答 |
|--------|------|-----|
| `--callers <节点>` | 谁直接依赖这个包/SPEC？ | 反向邻居 —— 导入它的包、依赖它的 SPEC、带 `@MX:SPEC` 标签的代码文件 |
| `--blast <节点>` | 从这里改起会波及多远？ | 沿反向边广度遍历 (BFS) 得到的影响半径。`@MX:SPEC` 边双向传播，能触及代码文件所实现的 SPEC |
| `--fanin [--limit N]` | 被用得最多的包是哪些？ | 按导入扇入排名 —— @MX:DEBT 扇入查询的代用品（还没有按标签种类分的边） |
| `--specs-no-code` | 哪些 SPEC 没有和代码相连？ | edges.jsonl 中 `@MX:SPEC` 边为 0 条的 SPEC 清单 |
| `--milestones-no-card` | 哪些里程碑没有卡片就过去了？ | 卡片交叉核对行未主张卡片、或主张的卡片不在存活积压队列里的里程碑 |

```bash
$ moai graph query --callers SPEC-FOO-001
$ moai graph query --blast internal/config
$ moai graph query --fanin --limit 20
$ moai graph query --specs-no-code
$ moai graph query --milestones-no-card
```

可以用 `--edges <路径>` 指向别的 edges.jsonl，或用根参数指定别的项目根。

## 两个选择器的注意事项

{{< callout type="warning" >}}
{{< icon warning warn >}} **`--specs-no-code`**: “未连接”不等于“未实现”。大多数 SPEC 只产出文档、规则、挽具，没有代码也算完成。请把这份清单当作覆盖度地图来读，而不是缺陷清单。
{{< /callout >}}

{{< callout type="warning" >}}
{{< icon warning warn >}} **`--milestones-no-card`**: 积压队列在卡片结束（`done`）时会删掉对应行。所以“不在队列里”**同时**混着已结束的卡片和从未签发的卡片。请用 `git log --oneline --grep 'merge: tNN'` 逐项判别 —— 有提交即通过，没有则是新卡片候选。grep 到 0 条也不代表“没干过活”：卡片可能换了新 id 重新签发，切断新卡片之前先核对谱系。
{{< /callout >}}

## 相关文档

- [看板模式](/zh/advanced/kanban-mode) — 里程碑-卡片交叉核对所守护的卡片流
- [`/moai mx`](/zh/utility-commands/moai-mx) — @MX 标签与 `@MX:SPEC` 连接的源头
- [Navigator](/zh/core-concepts/navigator) — 把设计决策、SPEC、符号绑成一张图的另一处图 (nav-graph.json)
