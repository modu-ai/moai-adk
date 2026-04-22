---
title: Design System
description: Hybrid design workflow combining Claude Design and code-based paths
weight: 10
draft: false
---

# Design System

MoAI-ADK's design system supports a **hybrid approach**. Choose between Claude Design or code-based design to build brand-aligned web experiences.

## Two Paths

```mermaid
flowchart TD
    A["User Request<br>/moai design"] --> B["Establish<br>Brand Context"]
    B --> C{Choose Path}
    C -->|Path A| D["Use Claude Design<br>Create design in<br>claude.ai/design"]
    C -->|Path B| E["Code-Based Design<br>brand-voice.md +<br>visual-identity.md"]
    D --> F["Export<br>handoff bundle"]
    E --> G["Generate<br>copywriting +<br>design tokens"]
    F --> H["Parse and<br>convert bundle"]
    G --> H
    H --> I["expert-frontend<br>Code implementation"]
    I --> J["GAN Loop<br>Evaluate and iterate"]
    J --> K["Sprint Contract<br>Based completion"]
    K --> L["Final artifacts"]
```

## Key Features

- **Brand Consistency** — Brand context applied at every stage
- **Sprint Contract Protocol** — Clear acceptance criteria per iteration
- **4-Dimensional Scoring** — Design quality, originality, completeness, functionality
- **Anti-AI-Slop** — Rules preventing shallow AI-generated content
- **Accessibility Compliance** — WCAG AA standard automated validation

## Next Steps

- **[Getting Started](./getting-started.md)** — Start your first project with `/moai design`
- **[Claude Design Handoff](./claude-design-handoff.md)** — Learn Claude Design features and bundle export
- **[Code-Based Path](./code-based-path.md)** — Design using brand-voice.md
- **[GAN Loop](./gan-loop.md)** — Builder-Evaluator iteration process
- **[Migration Guide](./migration-guide.md)** — Convert existing .agency/ projects

## Requirements

- Latest MoAI-ADK version
- Claude Code desktop client v2.1.50 or later
- Path A: Claude.ai Pro or higher subscription
- Path B: Complete brand context files
