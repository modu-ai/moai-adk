---
title: 质量命令
weight: 50
draft: false
---

介绍 MoAI-ADK 的代码质量管理命令。

{{< callout type="info" >}}
质量命令专注于**代码审查、测试覆盖率、E2E 测试和架构分析**。帮助您系统地管理和改善代码质量。
{{< /callout >}}

## 命令对比

| 命令 | 目的 | 执行方式 | 使用时机 |
|------|------|----------|----------|
| `/moai review` | 代码审查 | 安全/性能/质量/UX 四维分析 | PR 前需要代码审查时 |
| `/moai codemaps` | 架构文档 | 代码库结构分析和文档化 | 想了解项目架构时 |

## 命令关系图

```mermaid
flowchart TD
    A[质量命令] --> B[分析命令]

    B --> D["/moai review<br/>代码审查"]
    B --> E["/moai codemaps<br/>架构文档"]

    D -->|发现问题时| H["/moai fix"]
```

{{< callout type="info" >}}
**不确定使用哪个命令？**

- 想全面检查代码质量 → `/moai review`
- 想了解并记录项目结构 → `/moai codemaps`
{{< /callout >}}
