---
title: Quality Commands
weight: 50
draft: false
---

Introduction to MoAI-ADK's code quality management commands.

{{< callout type="info" >}}
Quality commands are specialized for **code review, test coverage, E2E testing, and architecture analysis**. They help you systematically manage and improve code quality.
{{< /callout >}}

## Command Comparison

| Command | Purpose | Execution Method | When to Use |
|---------|---------|------------------|-------------|
| `/moai review` | Code review | 4-perspective analysis (security/performance/quality/UX) | Need code review before PR |
| `/moai codemaps` | Architecture docs | Codebase structure analysis and documentation | Want to understand project architecture |

## Command Relationship Diagram

```mermaid
flowchart TD
    A[Quality Commands] --> B[Analysis Commands]

    B --> D["/moai review<br/>Code Review"]
    B --> E["/moai codemaps<br/>Architecture Docs"]

    D -->|Issues found| H["/moai fix"]
```

{{< callout type="info" >}}
**Not sure which command to use?**

- Want to check overall code quality → `/moai review`
- Want to understand and document project structure → `/moai codemaps`
{{< /callout >}}
