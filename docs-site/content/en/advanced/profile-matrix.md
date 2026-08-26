---
title: Profile Matrix
weight: 4
draft: false
---

MoAI-ADK maps each of the 12 retained agents to its own `{model, effort}` pair through a single **profile matrix**. The active **profile** (`high` / `medium` / `low`) selects one column of the matrix, and that column's values apply to every subagent spawn. The matrix is **36 cells** keyed by agent name (12 agents × 3 profiles), replacing both the former group abstraction and the `plan_type × tier` axis.

## Profile axis

The profile has three values:

- `high` — quality-first column. The spend goes to the rows that judge rather than the rows that produce: the auditing/advising rows (`plan-auditor`, `sync-auditor`, `super-advisor`) and the coordinating rows (`manager-design`, `manager-lead`) hold `high`, while the authoring and implementing rows (`manager-spec`, `manager-develop`) sit at `medium` in all three columns. No row takes `max`. `xhigh` appears in no cell: on Opus 5 it scores the same as `high` while costing materially more.
- `medium` (default) — the balanced column. It differs from `high` in exactly two rows: `builder-harness` steps down to `medium` and `e2e-tester` to `low`. An absent or empty value is interpreted as `medium`.
- `low` — economical column. Opus 5 at `low` still scores higher **and** costs less per task than Sonnet 5 at any effort, so Opus is retained on every agentic row; most Opus rows land on `medium`, with `super-advisor` alone keeping `high` — the escalation path is what a cheap column most needs to keep sound. Sonnet appears only on single-shot, input-dominated rows.

`max` is a **read-time alias** of `high`. An existing `profile: max` still resolves to `high`, and saves always write the canonical name `high`. No migration step is required.

The profile is not a separate field from `performance_tier` but the same axis — `llm.profile` takes precedence, and when absent the legacy `performance_tier` is read as an alias. Both fields share the `high`/`medium`/`low` vocabulary. The resolver reads this effective profile to determine each agent's cell.

## Setting the profile

```bash
moai init . --profile high             # set at init
moai update --profile low              # switch afterward
```

The accepted values are `high` / `medium` / `low`; the legacy `max` is also accepted as input and normalized to `high`. The current value is visible in the `llm.profile` field of `.moai/config/sections/llm.yaml`.

## Profile matrix

The 12 retained agents receive their `{model, effort}` directly from the matrix below. Only user-added agents resolve to `inherit` (inherit the parent session model) and are excluded from model injection. Haiku appears nowhere in the matrix.

| Agent | high | medium (default) | low |
|---|---|---|---|
| manager-spec | opus / medium | opus / medium | opus / medium |
| plan-auditor | opus / high | opus / high | opus / medium |
| sync-auditor | opus / high | opus / high | opus / medium |
| manager-develop | opus / medium | opus / medium | opus / medium |
| super-advisor | opus / high | opus / high | opus / high |
| manager-design | opus / high | opus / high | opus / medium |
| manager-lead | opus / high | opus / high | opus / medium |
| builder-harness | opus / high | opus / medium | opus / low |
| e2e-tester | opus / medium | opus / low | sonnet / low |
| manager-docs | sonnet / low | sonnet / low | sonnet / low |
| manager-git | sonnet / low | sonnet / low | sonnet / low |
| Explore | sonnet / low | sonnet / low | sonnet / low |

Model distribution across the 36 cells is Opus 26 / Sonnet 10. Fable appears in no cell, and no cell uses `xhigh` or `max`.

The `manager-docs`, `manager-git`, and `Explore` rows are fixed at `sonnet / low` regardless of the profile — documentation synthesis, mechanical work, and read-only exploration do not raise their model class even when the profile rises.

Every row is monotone: `high` ≥ `medium` ≥ `low`. Lowering the profile never gives any agent a stronger combination than before.

### Why these cells

The cells are not derived from a cost/score curve — they are **settled operator input**. The single principle: the spend goes to the rows that judge, not the rows that produce. The auditing/advising rows (`plan-auditor`, `sync-auditor`, `super-advisor`) and the coordinating rows (`manager-design`, `manager-lead`) hold `high`, while the authoring and implementing rows (`manager-spec`, `manager-develop`) sit at `medium` in all three columns, `manager-docs` drops to `sonnet / low`, and no row takes `max`. Re-deriving these cells from a curve would silently walk the producing rows back up — treat a value change as an operator-judgment update, not a recalculation.

The two rules governing model class are grounded in measurement:

- **Opus dominates Sonnet at every effort.** Opus 5 at `low` (58%, $1.66/task, 36 steps) scores higher and costs less per task than Sonnet 5 at any level, including Sonnet 5 at `max` (54%, $26.40/task, 268 steps). What drives per-task cost is completion efficiency — the steps and output tokens spent finishing the task — not the per-token price. Sonnet is therefore retained only where multi-step completion does not apply: single-shot, input-dominated rows (`Explore` search, `manager-git` mechanics) where its lower input price is the operative factor. This is why every multi-turn agentic row is Opus.
- **`xhigh` is strictly dominated on Opus.** `high` scores 73% at $6.08 while `xhigh` scores the same 73% at $9.07 — no gain, +49% cost, +22% steps. It is retired from the matrix (6 cells → 0). `max` remains the only level above `high` in the vocabulary, but no row currently takes it.

{{< icon warning warn >}} **Scope of the evidence**: the benchmark measures *coding* agents. Documentation authoring, audit judgment, and SPEC authoring quality are **not** directly measured — those row placements rest on a similarity inference to multi-turn agentic work. Any row is reversible per-agent via `llm.agent_overrides`.

The Anthropic built-in `Explore` no longer resolves to `inherit` but to its own cell (`sonnet / low`). The `inherit` sentinel now survives only for user-added agents.

## Harness specialist model + effort

Specialists generated by `/moai:harness` are **model-uniform on `opus`** and **differentiated by effort alone**. Harness agents are persistent, user-owned specialists whose distinguishing axis is reasoning depth, not model tier. Pinning the model costs no context, since every non-Haiku model now carries a 1M context window.

Each purpose class borrows its effort from a corresponding retained-agent row:

| Purpose class | Effort source row | high | medium | low |
|---|---|---|---|---|
| `read-only-extract` | Explore | opus / low | opus / low | opus / low |
| `mechanical-transform` | manager-git | opus / low | opus / low | opus / low |
| `synthesize` | manager-docs | opus / low | opus / low | opus / low |
| `research` | plan-auditor | opus / high | opus / high | opus / medium |
| `verify-judge` | sync-auditor | opus / high | opus / high | opus / medium |
| `implement` | manager-develop | opus / medium | opus / medium | opus / medium |
| `design-architecture` | manager-design | opus / high | opus / high | opus / medium |

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

- Model slot mapping: `fable` → `glm-5.3` (Fable slot, `ANTHROPIC_DEFAULT_FABLE_MODEL`). This slot is a GLM environment binding, independent of the profile matrix — it stays wired even though no matrix cell selects Fable.
- Claude's 5-step effort collapses onto z.ai's reasoning ceiling. GLM-5.3 reasons **always** — disabling reasoning is not supported, and a request asking for it fails — so the control is a single 3-level `reasoning_effort` (low / high / max):
  - `low` → **reasoning-low**
  - `medium` / `high` / `xhigh` / `max` → **reasoning-max**
  - (unrecognized value → reasoning-max, the totality clause: never under-reason)
  - reasoning-high remains a legal wire value, but no Claude effort collapses onto it
  - a GLM session without an explicit override defaults to **reasoning-max**
- coding-max override: `manager-develop` is forced to **reasoning-max** regardless of the collapse result (z.ai's "reasoning max for coding tasks" recommendation)
- `manager-git`, at `low` effort in all three profiles, occupies the reasoning-low tier

The runtime SSOT is `internal/template/glm_effort_overlay.go`.

## Next steps

- [3-Tier Agent Architecture](/en/advanced/no-haiku-3tier/) — the DeepSWE leaderboard rationale and the 3-tier definition
- [Tokenomics Overview](/en/advanced/tokenomics-overview/) — the B-layer routing of the 4-layer tokenomics structure
- [Model Policy](/en/multi-llm/model-policy/) — performance_tier alias and GLM backend details
