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
Anthropic built-in `Explore`). Under the **No-Haiku policy**, every worker agent
is **pinned to Sonnet 5** regardless of tier, and the policy tier controls only
two axes — (a) where to place Opus, and (b) how far to lower Sonnet's reasoning
depth (effort).

## The 3-level policy overview

| Policy (performance_tier) | CLI flag | Plan | Opus placement | Workers | Best for |
|------------------------|-----------|------|-----------|------|-----------|
| **max** | `--model-policy max` | Max $200/month | 5 sites | Sonnet-pinned | Highest quality, maximum throughput |
| **medium** (default) | `--model-policy medium` | Max $100/month | 2 sites (on-demand) | Sonnet-pinned | Balance of quality and cost |
| **low** | `--model-policy low` | Plus $20/month | None (0) | Sonnet-pinned | Low budget, no Opus |

> **Name axis**: The `performance_tier` field in `llm.yaml` and the CLI flag
> `--model-policy` both use the same three values `max`/`medium`/`low` and map
> 1:1 (no separate translation). The default is `medium`. The `--high` flag is
> a deprecated alias for `--model-policy max` (one-cycle backward compatibility;
> so is `--low`). `performance_tier` is a legacy alias field for `profile` (the
> profile matrix column), read only when `profile` is absent and normalized
> `high`→`max`. The two fields are the same `max`/`medium`/`low` axis. User name
> and the like are kept separately in `user.yaml`.

> **Why does this matter?** The Plus $20 plan has no Opus access. Setting the `low` policy makes every agent use only Sonnet, preventing rate-limit errors. Higher plans assign Opus only to the core sites (plan authoring, auditing, advisory) and use Sonnet for the rest.

## Per-agent model assignment table

Every worker agent is pinned to Sonnet 5, and Opus is placed only at the specific
sites below. (The orchestrator main session also runs on Opus at `max`, but since
it is not a spawned agent, it is not included in the table.)

### Manager Agents (5)

| Agent | max | medium | low |
|---------|-----|--------|-----|
| manager-spec (plan) | opus | opus (Tier L only) | sonnet |
| manager-develop | sonnet | sonnet | sonnet |
| manager-docs | sonnet | sonnet | sonnet |
| manager-git | sonnet | sonnet | sonnet |
| manager-design | sonnet | sonnet | sonnet |

### Evaluator · Advisor · Builder Agents (4)

| Agent | max | medium | low |
|---------|-----|--------|-----|
| plan-auditor | opus | sonnet | sonnet |
| sync-auditor | opus | sonnet | sonnet |
| super-advisor | opus | opus | sonnet |
| builder-harness | sonnet | sonnet | sonnet |

> The Anthropic built-in `Explore` is a read-only exploration agent that
> operates without a dedicated assignment. The Agent Teams static layer
> (static role profiles) was retired in v3.0; parallel work is now covered by
> sub-agent parallel execution and dynamic workflows. The `moai cg` teammate
> runtime (tmux panes) is preserved.

> **Haiku removal (v3.0)**: The former Haiku slots (documentation, MX tagging, Git
> procedures) were replaced with `sonnet`/`effort:low`. The model is Sonnet, but
> reasoning depth is lowered to cut cost — the model class was not lowered.

## Assignment principles

- **All workers pinned to Sonnet**: manager-develop, manager-docs, manager-git, manager-design, builder-harness — the tier controls only where Opus is placed and how far Sonnet's effort is adjusted
- **Opus placement at max (5 sites)**: orchestrator, super-advisor, manager-spec (plan), plan-auditor, sync-auditor — where high reasoning ability is needed
- **medium minimizes Opus (2 sites, on-demand)**: Opus only for super-advisor and Tier L planning (manager-spec); Sonnet for the rest
- **low has 0 Opus**: everything, including advisory (super-advisor), is Sonnet, adjusted only by effort tiering

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

Per-agent model+effort assignment is handled by a single profile matrix. The
active profile (`profile` — `max`/`medium`/`low`) selects one column of the
matrix; when `profile` is absent the legacy `performance_tier` is read as an
alias, and failing that it is interpreted as `medium`. For the detailed
per-agent mapping, see the [Profile Matrix](/en/advanced/profile-matrix/) page.

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
