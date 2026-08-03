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
Anthropic built-in `Explore`). Under the **No-Haiku policy**, Haiku appears
nowhere. Opus carries every multi-turn agentic row and Sonnet is confined to
single-shot, input-dominated rows; the policy tier controls where each agent
sits on the Opus effort ladder, not which model class it gets.

## The 3-level policy overview

| Policy (profile) | CLI flag | Opus cells | Sonnet cells | Best for |
|------------------|----------|-----------|--------------|----------|
| **high** | `--model-policy high` | 9 of 11 | 2 of 11 | Highest quality; `max` effort on the two rarest-invocation rows |
| **medium** (default) | `--model-policy medium` | 9 of 11 | 2 of 11 | Balance of quality and cost; the knee of the cost/score curve |
| **low** | `--model-policy low` | 7 of 11 | 4 of 11 | Lowest cost per task; agentic rows drop to Opus `low` |

> **Name mapping**: The `profile` field in `llm.yaml`, the legacy
> `performance_tier` alias, and the CLI flag `--model-policy` all use the same
> three values `high`/`medium`/`low` and map 1:1 (no separate translation). The
> default is `medium`. The former top-tier name `max` is still **read** as an
> alias for `high` so existing configs keep resolving, but saves always write
> `high` — no migration step is required. `performance_tier` is read only when
> `profile` is absent. User name and the like are kept separately in `user.yaml`.

> **Why does this matter?** Lowering the policy no longer means switching to a
> weaker model class. On a long-horizon agentic task, Opus at `low` effort
> scores higher **and** costs less per task than Sonnet at any effort, because
> the bill is set by how many steps a model spends finishing — not by the
> per-token rate. So the `low` policy economizes *within* Opus by lowering
> reasoning depth, and reaches for Sonnet only on the single-shot rows where
> multi-step completion failure does not apply.

## Per-agent model assignment table

The 33 cells below are the profile matrix (11 agents × 3 profiles). Each cell is
the `{model, effort}` pair the resolver injects at spawn time. (The orchestrator
main session is not a spawned agent, so it is not in the table.)

### Manager Agents (5)

| Agent | high | medium | low |
|---------|------|--------|-----|
| manager-spec | opus / high | opus / medium | opus / low |
| manager-develop | opus / max | opus / medium | opus / low |
| manager-docs | opus / medium | opus / low | sonnet / low |
| manager-git | sonnet / low | sonnet / low | sonnet / low |
| manager-design | opus / high | opus / medium | opus / low |

### Evaluator · Advisor · Builder · Specialist Agents (5)

| Agent | high | medium | low |
|---------|------|--------|-----|
| plan-auditor | opus / high | opus / medium | opus / low |
| sync-auditor | opus / high | opus / medium | opus / low |
| super-advisor | opus / max | opus / high | opus / medium |
| builder-harness | opus / high | opus / medium | opus / low |
| e2e-tester | opus / medium | opus / low | sonnet / low |

### Built-in Agent (1)

| Agent | high | medium | low |
|---------|------|--------|-----|
| Explore | sonnet / low | sonnet / low | sonnet / low |

> `Explore` has no agent file on disk, so its effort cannot be pinned in
> frontmatter — the matrix records `sonnet / low` as the call-time default,
> stated in the spawn prompt. The Agent Teams static layer (static role
> profiles) was retired in v3.0; parallel work is now covered by sub-agent
> parallel execution and dynamic workflows. The `moai cg` teammate runtime
> (tmux panes) is preserved.

> **Haiku removal (v3.0)**: The former Haiku slots (documentation, MX tagging, Git
> procedures) were replaced with lower reasoning depth rather than a lower model
> class — cost is cut by effort tiering, not by model substitution.

## Assignment principles

- **Opus on every agentic row**: `manager-spec`, `manager-develop`, `plan-auditor`, `sync-auditor`, `manager-design`, `builder-harness`, `manager-docs`, `e2e-tester` — all multi-turn work stays on Opus, because Opus at `low` outscores Sonnet at any effort while costing less per task
- **Sonnet only on single-shot rows**: `manager-git` mechanics and `Explore` search complete in one input-dominated pass, so multi-step completion failure does not apply and Sonnet's lower input price is the operative factor. These two rows are fixed across all three profiles
- **`max` is confined to two cells**: `manager-develop` and `super-advisor`, in the `high` profile only — the rarest-invocation rows, where one decision carries disproportionate downstream cost
- **`xhigh` is used nowhere**: on Opus it matches `high` on score at 49% higher cost
- **`low` steps down effort, not model class**: agentic rows move to Opus `low`; only `manager-docs` and `e2e-tester` additionally fall back to Sonnet

To ensure the agent that authored a plan never audits it, `plan-auditor` and
`sync-auditor` keep assignments independent of `manager-spec` — bias prevention
is a structural property of the catalog, not of the cell values.

## v3.0 extension: the Tier×Phase declaration matrix

v3.0 adds a **work phase and SPEC size (Tier)** axis on top of per-agent
assignment. `internal/config/model_routing.go` declaratively manages the
Tier×Phase → {model, effort} matrix:

- **model**: inherit / sonnet / opus / glm / fable
- **effort** (reasoning depth): low / medium / high / xhigh / max
- **tier** (SPEC size): S / M / L
- **phase** (work phase): plan / run / sync / mx

Per-agent model+effort assignment is handled by a single profile matrix. The
active profile (`profile` — `high`/`medium`/`low`) selects one column of the
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
moai init my-project --model-policy high    # Highest quality (max effort on 2 rows)
moai init my-project --model-policy medium  # Balanced (default)
moai init my-project --model-policy low     # Lowest cost per task
```

`--model-policy` takes the three values `high`/`medium`/`low` and persists to the
`performance_tier` field of `llm.yaml`. The former top-tier name `max` is still
accepted as input and treated as an alias of `high`.

> The default policy is `medium` (llm.yaml `performance_tier: "medium"`, corresponding to CLI `--model-policy medium` — when absent, interpreted as `medium`). GLM settings are isolated in `settings.local.json` and never committed to Git.

## Next steps

- [CG Mode](/en/multi-llm/cg-mode) — cut costs with the Claude + GLM hybrid
- [Agent Guide](/en/advanced/agent-guide) — customizing agents
- [CLI Reference](/en/getting-started/cli) — moai init, moai update details
