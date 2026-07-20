---
title: Cost Optimization
weight: 70
draft: false
---

{{< callout type="info" >}}{{< icon flash primary >}} <strong>Belongs to</strong>: 🪙 Tokenomics
{{< /callout >}}
<!-- @value: tokenomics -->

MoAI-ADK tokenomics reduces inference cost along two axes. If **context dieting**
is the axis that shrinks the always-loaded context itself, then **prompt caching**
is the axis that reuses the remaining context at a 90%-discounted cost. This
section covers the break-even rule for deciding when to enable caching and how to
configure it.


## Documents in this section

- [Prompt Caching — Break-even Analysis and Implementation Guide](/en/cost-optimization/prompt-caching) — the 2-request break-even rule, `cacheStrategy` configuration, and statusline monitoring
