---
title: Profile Matrix
weight: 4
draft: false
---

MoAI-ADK maps each retained agent to a `{model, effort}` pair through a single **profile matrix**. The active **profile** (`max` / `medium` / `low`) selects one column of the matrix, and that column's values apply to every subagent spawn. This single 3-column profile axis replaces the former `plan_type × tier` 60-cell matrix (SPEC-MODEL-PROFILE-MATRIX-001).

## Profile axis

The profile has three values:

- `max` — highest-quality column. Places Fable at reasoning points, and Opus for design/harness/E2E.
- `medium` (default) — balanced column. Places Opus/high for reasoning and execution. An absent or empty value is interpreted as `medium`.
- `low` — economy column. Places Opus at low effort and routes mechanical work to Sonnet.

The profile is not a separate field from `performance_tier` but the same axis — `llm.profile` takes precedence, and when absent the legacy `performance_tier` is read as an alias (`high` → `max` normalization; `max`/`medium`/`low` pass through). The resolver reads this effective profile to determine each agent's cell.

## Setting the profile

```bash
moai init . --profile max              # set at init
moai update --profile low              # switch afterward
```

The current value is visible in the `llm.profile` field of `.moai/config/sections/llm.yaml`. In the `moai init` interactive wizard, a `high` answer is normalized to `max`.

## Profile matrix

The 10 grouped agents receive their `{model, effort}` from the matrix below. `Explore` and user-defined agents have no group, so they resolve to `inherit` (inherit the parent session model) and are not model-injection targets. Haiku appears nowhere in the matrix.

| Agent (group) | max | medium (default) | low |
|---|---|---|---|
| manager-spec (spec_auditors) | fable / medium | opus / high | opus / low |
| plan-auditor (spec_auditors) | fable / medium | opus / high | opus / low |
| sync-auditor (spec_auditors) | fable / medium | opus / high | opus / low |
| manager-develop (develop) | fable / low | opus / high | opus / medium |
| super-advisor (advisor) | fable / medium | fable / low | opus / high |
| manager-design (design_harness_e2e) | opus / high | opus / medium | opus / low |
| builder-harness (design_harness_e2e) | opus / high | opus / medium | opus / low |
| e2e-tester (design_harness_e2e) | opus / high | opus / medium | opus / low |
| manager-docs (docs) | sonnet / medium | sonnet / medium | sonnet / medium |
| manager-git (git) | sonnet / low | sonnet / low | sonnet / low |
| Explore (—) | inherit | inherit | inherit |

The `docs` and `git` rows are fixed regardless of the profile (sonnet/medium and sonnet/low respectively) — mechanical work does not raise its model class even when the profile changes.

## Agent groups

The matrix is defined over 6 **groups**, not individual agent names. The group → agent membership is as follows:

| Group | Agents |
|---|---|
| `spec_auditors` | manager-spec, plan-auditor, sync-auditor |
| `develop` | manager-develop |
| `advisor` | super-advisor |
| `design_harness_e2e` | manager-design, builder-harness, e2e-tester |
| `docs` | manager-docs |
| `git` | manager-git |

`Explore` and user-added agents have no membership and resolve to `inherit`.

## Resolver precedence

Each agent's effective `{model, effort}` is determined in this order:

1. If `llm.agent_overrides[agent]` exists, it wins.
2. Otherwise the active profile's group cell (config `llm.profiles`) is used.
3. If the config has no cell, the group cell of the Go default matrix (`template.DefaultProfileMatrix`) is used.
4. If there is no group membership, it is `inherit` (no injection).

`agent_overrides` is keyed by canonical agent name and validated against the catalog + enum:

```yaml
llm:
  agent_overrides:
    manager-develop: { model: opus, effort: xhigh }
```

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

- Model slot mapping: `fable` → `glm-5.2` (Fable slot, `ANTHROPIC_DEFAULT_FABLE_MODEL`)
- Claude's 5-step effort collapses into the 3-state z.ai can reach:
  - `low` → **thinking-off**
  - `medium` / `high` → **reasoning-high**
  - `xhigh` / `max` → **reasoning-max**
  - (unrecognized value → reasoning-max, to prevent under-reasoning)
- coding-max override: `manager-develop` is forced to **reasoning-max** regardless of the collapse result
- `manager-git` at low effort → **thinking-off**

Whether z.ai actually consumes the `ANTHROPIC_REASONING_EFFORT` value via the Anthropic-compat shim is an empirical task requiring live GLM session outbound observation. The runtime SSOT is `internal/template/glm_effort_overlay.go`.

## Next steps

- [3-Tier Agent Architecture](/en/advanced/no-haiku-3tier/) — the DeepSWE leaderboard rationale and the 3-tier definition
- [Tokenomics Overview](/en/advanced/tokenomics-overview/) — the B-layer routing of the 4-layer tokenomics structure
- [Model Policy](/en/multi-llm/model-policy/) — performance_tier alias and GLM backend details
