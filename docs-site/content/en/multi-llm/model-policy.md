---
title: Model Policy
weight: 30
draft: false
---

## What is a model policy?

The model policy is the backbone of MoAI-ADK Tokenomics. Rather than "the best
model for everything," it declaratively assigns the right model to each agent
— reasoning-heavy work like planning and auditing, and lightweight work like
documentation and Git. It maximizes quality within your Claude Code
subscription plan while preventing rate-limit errors.

The MoAI-ADK v3.0 agent catalog contains **11** agents (10 MoAI-custom + the
Anthropic built-in `Explore`); the assignment table below covers the 7 core
agents the model policy assigns directly.

## The 3-level policy overview

| Policy (performance_tier) | CLI flag | Plan | Opus | Sonnet | Best for |
|------------------------|-----------|------|------|--------|-----------|
| **max** | `--model-policy max` | Max $200/month | 5 | 2 | Highest quality, maximum throughput |
| **medium** (default) | `--model-policy medium` | Max $100/month | 2 | 5 | Balance of quality and cost |
| **low** | `--model-policy low` | Plus $20/month | 0 | 7 | Low budget, no Opus |

> **Name axis**: The `performance_tier` field in `llm.yaml` and the CLI flag
> `--model-policy` both use the same three values `max`/`medium`/`low` and map
> 1:1 (no separate translation). The default is `medium`. The `--high` flag is
> a deprecated alias for `--model-policy max` (one-cycle backward compatibility;
> so is `--low`). `performance_tier` controls only subagent model assignment,
> and is a separate axis from the `plan_type` field that decides the pricing
> plan kind (api / subscription). User name and the like are kept separately in
> `user.yaml`.

> **Why does this matter?** The Plus $20 plan has no Opus access. Setting the `low` policy makes every agent use only Sonnet, preventing rate-limit errors. Higher plans assign Opus to the core agents (planning, auditing) and Sonnet to day-to-day work.

## Per-agent model assignment table

### Manager Agents (4)

| Agent | max | medium | low |
|---------|-----|--------|-----|
| manager-spec | opus | opus | sonnet |
| manager-develop | opus | sonnet | sonnet |
| manager-docs | sonnet | sonnet | sonnet |
| manager-git | sonnet | sonnet | sonnet |

### Evaluator & Builder Agents (3)

| Agent | max | medium | low |
|---------|-----|--------|-----|
| plan-auditor | opus | opus | sonnet |
| sync-auditor | opus | sonnet | sonnet |
| builder-harness | opus | sonnet | sonnet |

> The Anthropic built-in `Explore` is a read-only exploration agent that
> operates without a dedicated assignment. The Agent Teams static layer
> (static role profiles) was retired in v3.0; parallel work is now covered by
> sub-agent parallel execution and dynamic workflows. The `moai cg` teammate
> runtime (tmux panes) is preserved.

> **Haiku removal (v3.0)**: The former Haiku slots were replaced with
> `sonnet`/`effort:low`. This applies to the lightweight work of `manager-git`
> and `manager-docs` — the model is Sonnet, but reasoning depth is lowered to
> cut cost.

## Assignment principles

- **Always Opus**: plan auditing (plan-auditor), SPEC authoring (manager-spec) — needs high reasoning ability
- **Always Sonnet/effort:low**: Git (manager-git) — light and fast work
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
subscription plans even for the same workflow, the pricing-plan-kind
(`plan_type` — `api` or `subscription`) profile applies a separate matrix per
plan. `plan_type` is an axis independent from `performance_tier`; when absent,
it is interpreted as `subscription`.

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

### Setting it directly with a CLI flag

```bash
moai init my-project --model-policy max     # Highest quality (Opus-centric)
moai init my-project --model-policy medium  # Balanced (default)
moai init my-project --model-policy low     # Sonnet only, no Opus
```

`--model-policy` takes the three values `max`/`medium`/`low` and is stored
as-is in the `performance_tier` field of `llm.yaml`. The deprecated `--high`
flag is an alias for `--model-policy max`.

> The default policy is `medium` (llm.yaml `performance_tier: "medium"`, corresponding to CLI `--model-policy medium` — when absent, interpreted as `medium`). GLM settings are isolated in `settings.local.json` and never committed to Git.

## Next steps

- [CG Mode](/en/multi-llm/cg-mode) — cut costs with the Claude + GLM hybrid
- [Agent Guide](/en/advanced/agent-guide) — customizing agents
- [CLI Reference](/en/getting-started/cli) — moai init, moai update details
