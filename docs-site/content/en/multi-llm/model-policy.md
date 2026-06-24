---
title: "Model Policy"
weight: 30
draft: false
---

## What is the Model Policy?

MoAI-ADK assigns the optimal AI model to each of the 8 retained agents (7 MoAI-custom + Anthropic built-in Explore). It maximizes quality while preventing rate-limit errors, tailored to your Claude Code subscription plan.

## 3-Tier Policy Overview

| Policy | Plan | 🟣 Opus | 🔵 Sonnet | 🟡 Haiku | Suitable use case |
|--------|------|---------|-----------|----------|-------------------|
| **High** | Max $200/mo | 5 | 1 | 1 | Highest quality, maximum throughput |
| **Medium** | Max $100/mo | 2 | 3 | 2 | Balance of quality and cost |
| **Low** | Plus $20/mo | 0 | 4 | 3 | Low budget, no Opus |

> **Why does this matter?** The Plus $20 plan cannot access Opus. Setting the `Low` policy makes all agents use only Sonnet and Haiku, preventing rate-limit errors. Higher plans assign Opus to core agents (security, strategy, architecture) and use Sonnet/Haiku for routine tasks.

## Per-Agent Model Assignment Table

### Manager Agents (4)

| Agent | High | Medium | Low |
|-------|------|--------|-----|
| manager-spec | 🟣 opus | 🟣 opus | 🔵 sonnet |
| manager-develop | 🟣 opus | 🔵 sonnet | 🔵 sonnet |
| manager-docs | 🔵 sonnet | 🟡 haiku | 🟡 haiku |
| manager-git | 🟡 haiku | 🟡 haiku | 🟡 haiku |

### Evaluator & Builder Agents (3)

| Agent | High | Medium | Low |
|-------|------|--------|-----|
| plan-auditor | 🟣 opus | 🟣 opus | 🔵 sonnet |
| sync-auditor | 🟣 opus | 🔵 sonnet | 🔵 sonnet |
| builder-harness | 🟣 opus | 🔵 sonnet | 🟡 haiku |

> Team-mode roles (researcher, analyst, architect, implementer, tester, designer, reviewer) are not static agents — they are dynamically generated via `Agent(general-purpose)` from the role profiles in `workflow.yaml`.

## Assignment Principles

- **Always Opus**: plan audit (plan-auditor), SPEC authoring (manager-spec) — require high reasoning
- **Always Haiku**: documentation (manager-docs), Git (manager-git) — lightweight, fast tasks
- **Varies by plan**: implementation (manager-develop, cycle_type=tdd/ddd) — higher plans get Opus

## How to Configure

### At project initialization

```bash
moai init my-project
# The interactive wizard includes model policy selection
```

### Reconfiguring an existing project

```bash
moai update
# Interactive prompts:
# - Reset model policy? (y/n) — reset model policy
# - Update GLM settings? (y/n) — configure GLM environment variables
```

> The default policy is `High`. GLM settings are isolated in `settings.local.json` and are not committed to Git.

## Next Steps

- [CG Mode](/en/multi-llm/cg-mode) — save cost with the Claude + GLM hybrid
- [Agent Guide](/en/advanced/agent-guide) — customizing agents
- [CLI Reference](/en/getting-started/cli) — moai init, moai update details
