---
title: Model Policy
weight: 30
draft: false
---

## What is a model policy?

The model policy is the backbone of MoAI-ADK Tokenomics. Rather than "the best
model for everything", it declaratively assigns the right model to each agent
— reasoning-heavy work like planning and auditing versus lightweight work like
documentation and Git. It maximizes quality within your Claude Code
subscription plan while preventing rate-limit errors.

The MoAI-ADK v3.0 agent catalog contains **11** agents (10 MoAI-custom + the
Anthropic built-in `Explore`); the assignment table below covers the 7 core
agents the model policy assigns directly.

## The 3-level policy overview

| Policy | Plan | Opus | Sonnet | Haiku | Best for |
|------|------|---------|-----------|----------|-----------|
| **High** | Max $200/month | 5 | 1 | 1 | Highest quality, maximum throughput |
| **Medium** | Max $100/month | 2 | 3 | 2 | Balance of quality and cost |
| **Low** | Plus $20/month | 0 | 4 | 3 | Low budget, no Opus |

> **Why does this matter?** The Plus $20 plan has no Opus access. Setting the `Low` policy makes every agent use only Sonnet and Haiku, preventing rate-limit errors. Higher plans assign Opus to the core agents (planning, auditing) and Sonnet/Haiku to day-to-day work.

## Per-agent model assignment table

### Manager Agents (4)

| Agent | High | Medium | Low |
|---------|------|--------|-----|
| manager-spec | opus | opus | sonnet |
| manager-develop | opus | sonnet | sonnet |
| manager-docs | sonnet | haiku | haiku |
| manager-git | haiku | haiku | haiku |

### Evaluator & Builder Agents (3)

| Agent | High | Medium | Low |
|---------|------|--------|-----|
| plan-auditor | opus | opus | sonnet |
| sync-auditor | opus | sonnet | sonnet |
| builder-harness | opus | sonnet | haiku |

> The Anthropic built-in `Explore` is a read-only exploration agent that
> operates without a dedicated assignment. The Agent Teams static layer
> (static role profiles) was retired in v3.0; parallel work is now covered by
> sub-agent parallel execution and dynamic workflows. The `moai cg` teammate
> runtime (tmux panes) is preserved.

## Assignment principles

- **Always Opus**: plan auditing (plan-auditor), SPEC authoring (manager-spec) — needs high reasoning ability
- **Always Haiku**: Git (manager-git) — light and fast work
- **Varies by plan**: implementation (manager-develop, cycle_type=tdd/ddd) — the higher the plan, the more Opus

To ensure the agent that authored a plan never audits it, plan-auditor and
sync-auditor keep independent assignments — the table is designed along both
the cost axis and the quality axis (bias prevention).

## v3.0 extension: the Tier×Phase declaration axis

v3.0 adds a **work phase and SPEC size (Tier)** axis on top of per-agent
assignment. `internal/config/model_routing.go` declaratively manages the
Tier×Phase → {model, effort} matrix:

- **model**: inherit / sonnet / opus / glm / fable
- **effort** (reasoning depth): low / medium / high / xhigh / max
- **tier** (SPEC size): S / M / L
- **phase** (work phase): plan / run / sync / mx

Because the optimal allocation differs between pay-as-you-go API usage and
subscription plans even for the same workflow, plan-aware (plan_type)
profiles apply separate matrices per pricing plan.

## Configuration

### At project initialization

```bash
moai init my-project
# The interactive wizard includes model policy selection
```

### Reconfiguring an existing project

```bash
moai update
# Interactive prompts:
# - Reset model policy? (y/n) — reset the model policy
# - Update GLM settings? (y/n) — configure GLM environment variables
```

> The default policy is `High`. GLM settings are isolated in `settings.local.json` and never committed to Git.

## Next steps

- [CG Mode](/en/multi-llm/cg-mode) — cut costs with the Claude + GLM hybrid
- [Agent Guide](/en/advanced/agent-guide) — customizing agents
- [CLI Reference](/en/getting-started/cli) — moai init, moai update details
