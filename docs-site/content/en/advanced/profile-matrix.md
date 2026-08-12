---
title: Profile Matrix
weight: 4
draft: false
---

# Profile Matrix

MoAI-ADK maps each of the 11 retained agents to its own `{model, effort}` pair through a single **profile matrix**. The active **profile** (`high` / `medium` / `low`) selects one column of the matrix, and that column's values apply to every subagent spawn. The matrix is **33 cells** keyed by agent name (11 agents × 3 profiles), replacing both the former group abstraction and the `plan_type × tier` axis.

## Profile axis

The profile has three values:

- `high` — quality-first column. Opus 5 carries every multi-turn agentic row, and `max` is reserved for the two rarest-invocation rows (`manager-develop`, `super-advisor`). `xhigh` appears in no cell: on Opus 5 it scores the same as `high` while costing materially more.
- `medium` (default) — the balanced column, and the anchor the rest of the matrix derives from. `manager-develop` sits at Opus 5 `medium`, the knee of the cost/score curve. An absent or empty value is interpreted as `medium`.
- `low` — economical column. Opus 5 at `low` still scores higher **and** costs less per task than Sonnet 5 at any effort, so Opus is retained on every agentic row; Sonnet appears only on single-shot, input-dominated rows.

`max` is a **read-time alias** of `high`. An existing `profile: max` still resolves to `high`, and saves always write the canonical name `high`. No migration step is required.

The profile is not a separate field from `performance_tier` but the same axis — `llm.profile` takes precedence, and when absent the legacy `performance_tier` is read as an alias. Both fields share the `high`/`medium`/`low` vocabulary. The resolver reads this effective profile to determine each agent's cell.

## Setting the profile

```bash
moai init . --profile high             # set at init
moai update --profile low              # switch afterward
```

The accepted values are `high` / `medium` / `low`; the legacy `max` is also accepted as input and normalized to `high`. The current value is visible in the `llm.profile` field of `.moai/config/sections/llm.yaml`.

## Profile matrix

The 11 retained agents receive their `{model, effort}` directly from the matrix below. Only user-added agents resolve to `inherit` (inherit the parent session model) and are excluded from model injection. Haiku appears nowhere in the matrix.

| Agent | high | medium (default) | low |
|---|---|---|---|
| manager-spec | opus / high | opus / medium | opus / low |
| plan-auditor | opus / high | opus / medium | opus / low |
| sync-auditor | opus / high | opus / medium | opus / low |
| manager-develop | opus / max | opus / medium | opus / low |
| super-advisor | opus / max | opus / high | opus / medium |
| manager-design | opus / high | opus / medium | opus / low |
| builder-harness | opus / high | opus / medium | opus / low |
| e2e-tester | opus / medium | opus / low | sonnet / low |
| manager-docs | opus / medium | opus / low | sonnet / low |
| manager-git | sonnet / low | sonnet / low | sonnet / low |
| Explore | sonnet / low | sonnet / low | sonnet / low |

Model distribution across the 33 cells is Opus 25 / Sonnet 8. Fable appears in no cell, and no cell uses `xhigh`.

The `manager-git` and `Explore` rows are fixed at `sonnet / low` regardless of the profile — mechanical work and read-only exploration do not raise their model class even when the profile rises.

Every row is monotone: `high` ≥ `medium` ≥ `low`. Lowering the profile never gives any agent a stronger combination than before.

### Why these cells

The cells are derived from a long-horizon coding-agent benchmark that reports score, cost per task, output tokens, and agent steps **at every effort level** — not from unit token price. Three measurements drive the layout:

- **Opus dominates Sonnet at every effort.** Opus 5 at `low` (58%, $1.66/task, 36 steps) scores higher and costs less per task than Sonnet 5 at any level, including Sonnet 5 at `max` (54%, $26.40/task, 268 steps). What drives per-task cost is completion efficiency — the steps and output tokens spent finishing the task — not the per-token price. Sonnet is therefore retained only where multi-step completion does not apply: single-shot, input-dominated rows (`Explore` search, `manager-git` mechanics) where its lower input price is the operative factor.
- **`xhigh` is strictly dominated on Opus.** `high` scores 73% at $6.08 while `xhigh` scores the same 73% at $9.07 — no gain, +49% cost, +22% steps. It is retired from the matrix (6 cells → 0). `max` survives only in the two rarest-invocation cells.
- **`medium` is the knee of the curve.** Marginal cost per point rises several-fold above it: `low`→`medium` costs $0.15 per point, `medium`→`high` costs $0.70 per point (4.7×). This is why `manager-develop` at `medium` anchors the default column.

{{< icon warning warn >}} **Scope of the evidence**: the benchmark measures *coding* agents. Documentation authoring, audit judgment, and SPEC authoring quality are **not** directly measured — those row placements rest on a similarity inference to multi-turn agentic work. Any row is reversible per-agent via `llm.agent_overrides`.

The Anthropic built-in `Explore` no longer resolves to `inherit` but to its own cell (`sonnet / low`). The `inherit` sentinel now survives only for user-added agents.

## Harness specialist model + effort

Specialists generated by `/moai:harness` are **model-uniform on `opus`** and **differentiated by effort alone**. Harness agents are persistent, user-owned specialists whose distinguishing axis is reasoning depth, not model tier. Pinning the model costs no context, since every non-Haiku model now carries a 1M context window.

Each purpose class borrows its effort from a corresponding retained-agent row:

| Purpose class | Effort source row | high | medium | low |
|---|---|---|---|---|
| `read-only-extract` | Explore | opus / low | opus / low | opus / low |
| `mechanical-transform` | manager-git | opus / low | opus / low | opus / low |
| `synthesize` | manager-docs | opus / medium | opus / low | opus / low |
| `research` | plan-auditor | opus / high | opus / medium | opus / low |
| `verify-judge` | sync-auditor | opus / high | opus / medium | opus / low |
| `implement` | manager-develop | opus / max | opus / medium | opus / low |
| `design-architecture` | manager-design | opus / high | opus / medium | opus / low |

`llm.harness_agents[profile][class].effort` overrides a class's effort. The model never changes through any path. An unrecognized class falls back to `implement`.

## Resolver precedence

Each agent's effective `{model, effort}` is determined in this order:

1. If `llm.agent_overrides[agent]` exists, it wins.
2. Otherwise the active profile's agent cell (config `llm.profiles`) is used.
3. If the config has no cell, the agent cell of the Go default matrix (`template.DefaultProfileMatrix`) is used.
4. An agent absent from the matrix (user-added) is `inherit` (no injection).

`agent_overrides` is keyed by canonical agent name and validated against the catalog + enum:

```yaml
llm:
  agent_overrides:
    manager-develop: { model: opus, effort: high }
```

The enum still accepts `fable` as a model and `xhigh` as an effort — they are absent from the default matrix, not removed from the vocabulary, so an override may still select either.

The consumption paths of **model** and **effort** differ. The resolved **model** is the value the orchestrator injects as the `Agent(model: <alias>)` runtime argument at spawn time (`[1m]`-safe, separate from the frontmatter `model:` field). The agent `.md` frontmatter stays `model: inherit`, and init/update/web saves do not change it. The resolved **effort** is the *documented intent* for a NAMED subagent — the Agent/Task tool takes no per-spawn effort argument for named subagents, so effort is consumed only through (a) the agent frontmatter effort default, (b) the GLM effort overlay, and (c) Workflow / `Agent(general-purpose)` prompt-level steering.

## moai model profile

The per-agent model+effort resolved for the active profile is inspected via a read-only accessor:

```bash
moai model profile          # human-readable table
moai model profile --json   # machine-readable
```

This command changes nothing — it exposes exactly the values the orchestrator will inject at spawn time.

## GLM backend effort overlay

{{< icon warning warn >}} **Honesty notice**: the GLM backend effort overlay is **implemented + wired**, but wire effectiveness (live effectiveness) is pending empirical verification — it is not described as "behavior guaranteed".

On the GLM backend (`moai glm` / `moai cg` GLM panes), an overlay is applied on top of the profile matrix:

- Model slot mapping: `fable` → `glm-5.2` (Fable slot, `ANTHROPIC_DEFAULT_FABLE_MODEL`). This slot is a GLM environment binding, independent of the profile matrix — it stays wired even though no matrix cell selects Fable.
- Claude's 5-step effort collapses into the 3-state z.ai can reach:
  - `low` → **thinking-off**
  - `medium` / `high` → **reasoning-high**
  - `xhigh` / `max` (legacy effort value) → **reasoning-max**
  - (unrecognized value → reasoning-max, to prevent under-reasoning)
- coding-max override: `manager-develop` is forced to **reasoning-max** regardless of the collapse result
- `manager-git` at low effort → **thinking-off**

Whether z.ai actually consumes the `ANTHROPIC_REASONING_EFFORT` value via the Anthropic-compat shim is an empirical task requiring live GLM session outbound observation. The runtime SSOT is `internal/template/glm_effort_overlay.go`.

## Next steps

- [3-Tier Agent Architecture](/en/advanced/no-haiku-3tier/) — the DeepSWE leaderboard rationale and the 3-tier definition
- [Tokenomics Overview](/en/advanced/tokenomics-overview/) — the B-layer routing of the 4-layer tokenomics structure
- [Model Policy](/en/multi-llm/model-policy/) — performance_tier alias and GLM backend details
