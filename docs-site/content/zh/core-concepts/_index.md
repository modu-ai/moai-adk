---
title: 核心概念
weight: 20
draft: false
---

介绍理解 MoAI-ADK 的 4 个核心概念。

{{< callout type="info" >}}
新手吗？从上到下阅读，自然理解 MoAI-ADK 的全貌。
{{< /callout >}}

```mermaid
flowchart TD
    A["什么是 MoAI-ADK？"] --> B["基于 SPEC 的开发"]
    B --> C["领域驱动开发"]
    C --> D["TRUST 5 质量"]

    A -.- A1["理解 AI 框架的<br>必要性和结构"]
    B -.- B1["在 Plan 阶段<br>定义需求文档"]
    C -.- C1["在 Run 阶段<br>安全改进现有代码"]
    D -.- D1["用 5 个质量原则<br>验证所有阶段"]
```

## 学习顺序

| 顺序 | 文档 | 核心问题 |
|-------|----------|--------------|
| 1 | [什么是 MoAI-ADK？](/core-concepts/what-is-moai-adk) | 为什么需要 AI 开发工具，它是如何构建的？ |
| 2 | [基于 SPEC 的开发](/core-concepts/spec-based-dev) | 如何明确定义和管理需求？ |
| 3 | [领域驱动开发](/core-concepts/ddd) | 如何在不破坏现有功能的情况下改进代码？ |
| 4 | [TRUST 5 质量](/core-concepts/trust-5) | 什么标准确保代码质量？ |

{{< callout type="info" >}}
每个文档都可以独立阅读，但按顺序阅读可以自然连接 MoAI-ADK 的开发理念。用 **SPEC** 定义**内容**，用 **DDD** 安全实现，用 **TRUST 5** 验证质量。
{{< /callout >}}
