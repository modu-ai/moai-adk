---
title: 成本优化
weight: 70
draft: false
---

{{< callout type="info" >}}{{< icon flash primary >}} <strong>所属价值</strong>: 代币经济学
{{< /callout >}}
<!-- @value: tokenomics -->

MoAI-ADK 代币经济学通过两种方式削减推理成本。**上下文瘦身**缩减常驻加载的
上下文本身,而 **提示缓存**则以 90% 折扣的成本复用剩余上下文。
本节讲解判断何时启用缓存的盈亏平衡规则以及配置方法。


## 本站中提示缓存的两篇文档

站内有两篇文档讲解提示缓存。它们从不同视角阐释同一机制，你可以根据自己的疑问选择起点。

| 文档 | 视角 | 何时阅读 |
|------|------|-----------|
| [本节：提示缓存](/zh/cost-optimization/prompt-caching) | **成本** — 盈亏平衡、价格倍数、缓存何时保持与何时失效 | 想了解节省效果与缓存失效行为时 |
| [Claude Code: 提示缓存](/zh/claude-code/context-memory/prompt-caching) | **上下文** — 每轮如何复用前文以降低延迟 | 想了解 Claude Code 逐轮如何处理上下文时 |

若主要关心成本，请从本节文档开始。若更想知道 Claude Code 运行时每轮如何构建并复用上下文，也可以从上下文视角的文档入手 — 两篇文档互相引用，从哪边都能跳转。

## 本节文档

- [提示缓存 — 盈亏平衡分析与实现指南](/zh/cost-optimization/prompt-caching) — 2 次请求盈亏平衡规则、断点放置、statusline 监控
