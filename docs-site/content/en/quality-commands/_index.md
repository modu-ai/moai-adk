---
title: Quality Commands
weight: 40
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
| `/moai coverage` | Coverage analysis | Test gap analysis and test generation | Want to improve test coverage |
| `/moai e2e` | E2E testing | Browser automation test creation/execution | Want to verify user flows |
| `/moai codemaps` | Architecture docs | Codebase structure analysis and documentation | Want to understand project architecture |

## Command Relationship Diagram

```mermaid
flowchart TD
    A[Quality Commands] --> B[Analysis Commands]
    A --> C[Testing Commands]

    B --> D["/moai review<br/>Code Review"]
    B --> E["/moai codemaps<br/>Architecture Docs"]
    C --> F["/moai coverage<br/>Coverage Analysis"]
    C --> G["/moai e2e<br/>E2E Testing"]

    D -->|Issues found| H["/moai fix"]
    F -->|Tests lacking| G
```

{{< callout type="info" >}}
**Not sure which command to use?**

- Want to check overall code quality → `/moai review`
- Want to find and fill test gaps → `/moai coverage`
- Want to verify the app works from user perspective → `/moai e2e`
- Want to understand and document project structure → `/moai codemaps`
{{< /callout >}}
