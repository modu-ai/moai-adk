---
title: 成本优化
weight: 70
draft: false
---

{{< callout type="info" >}}{{< icon flash primary >}} <strong>所属价值</strong>: 🪙 代币经济学
{{< /callout >}}
<!-- @value: tokenomics -->

MoAI-ADK 代币经济学通过两种方式削减推理成本。**上下文瘦身**缩减常驻加载的
上下文本身,而 **提示缓存**则以 90% 折扣的成本复用剩余上下文。
本节讲解判断何时启用缓存的盈亏平衡规则以及配置方法。


## 本节文档

- [提示缓存 — 盈亏平衡分析与实现指南](/zh/cost-optimization/prompt-caching) — 2 次请求盈亏平衡规则、`cacheStrategy` 配置、statusline 监控
