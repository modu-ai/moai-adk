---
title: plan_type Tier Profiles
weight: 4
draft: false
---

MoAI-ADK recognizes that even the same workflow has different optimal allocation under API pay-as-you-go vs subscription plans. The `plan_type` axis applies separate Tier × Phase model/effort matrices per billing model. This page officially documents the 60-cell profile matrix implemented by SPEC-MODEL-TIER-PLANTYPE-001 (CLOSED).

## The plan_type Axis

`plan_type` has two values:

- `api` — pay-as-you-go. Dollars are the sole constraint. Per-task cost optimization is the goal.
- `subscription` — subscription plan. Weekly token quota + Opus weighted deduction are the constraints. Maximizing solved-tasks-per-quota is the goal.

In subscription plans, Opus hours are separately weighted (Max 5x: Sonnet 140-280h vs Opus 15-35h, roughly 1/8). Therefore, in subscriptions, Opus is assigned only to reasoning and execution runs on abundant Sonnet hours — the opusplan structure is optimal.

## Setting plan_type

```bash
moai init . --plan-type api           # set at initialization
moai update --plan-type subscription  # post-hoc switch
```

The current value can be checked in the `llm.plan_type` field of `llm.yaml`.

## 60-Cell Profile Matrix

10 agents × 3 tiers × 2 plan_types = 60 cells. The tables below are the ApplyTierProfile implementation from SPEC-MODEL-TIER-PLANTYPE-001.

### Plan A — API Pay-As-You-Go (rev2)

In API, dollars are the sole constraint. rev2 revision: Sonnet's nominal price is half of Opus, but per-task cost inverts (Opus $13.22 < Sonnet $26.40). Therefore, API uses Opus for execution too. Reasoning = Fable (quality leader), execution = Opus (per-task cost leader), mechanical = Sonnet low.

| Agent (role) | A-max (quality) | A-medium (recommended) | A-low (cost) |
|---|---|---|---|
| manager-spec (reasoning) | fable / high | fable / high | opus / high |
| plan-auditor (reasoning) | fable / high | fable / high | opus / high |
| sync-auditor (reasoning) | fable / high | opus / high | opus / medium |
| manager-design (reasoning) | fable / high | fable / high | opus / high |
| super-advisor (top reasoning) | fable / xhigh | fable / high | opus / high |
| manager-develop (execution) | fable / high | opus / high | opus / medium |
| builder-harness (execution) | opus / high | opus / medium | opus / medium |
| manager-docs (mechanical) | sonnet / medium | sonnet / low | sonnet / low |
| manager-git (mechanical) | sonnet / low | sonnet / low | sonnet / low |
| Explore (exploration) | inherit / medium | inherit / low | inherit / low |

### Plan B — Subscription (Availability-First)

The subscription constraint is not dollars but weekly token quota + Opus weighted deduction. The goal is maximizing solved-tasks-per-quota = excluding retry loops + assigning Opus only to reasoning. This is a precise version of Anthropic's official opusplan pattern ("Opus for planning, Sonnet for execution").

| Agent (role) | B-max (recommended) | B-medium | B-low (Pro) |
|---|---|---|---|
| manager-spec (reasoning) | opus / high | opus / high | opus / medium |
| plan-auditor (reasoning) | opus / high | opus / medium | sonnet / high |
| sync-auditor (reasoning) | opus / high | opus / medium | sonnet / high |
| manager-design (reasoning) | opus / high | opus / medium | sonnet / high |
| super-advisor (top reasoning) | opus / xhigh | opus / high | opus / medium |
| manager-develop (execution) | sonnet / high | sonnet / high | sonnet / high |
| builder-harness (execution) | sonnet / high | sonnet / medium | sonnet / medium |
| manager-docs (mechanical) | sonnet / low | sonnet / low | sonnet / low |
| manager-git (mechanical) | sonnet / low | sonnet / low | sonnet / low |
| Explore (exploration) | inherit / medium | inherit / low | inherit / low |

## ApplyTierProfile Mechanism

`ApplyTierProfile` replaces both `model` and `effort` in agent frontmatter (replace-both). Since all agents have an `effort:` field, "preserve" mode is ineffective, so it always operates as replace-both.

This mechanism was implemented in SPEC-MODEL-TIER-PLANTYPE-001 (run-phase complete, CLOSED). All cells in the tables above are verified live behavior.

## GLM Backend Effort Overlay

{{< icon warning warn >}} **Honesty caveat (REQ-DA-060)**: The GLM backend effort overlay's wire effectiveness is a verification item requiring live GLM session outbound observation.

The GLM backend (`moai glm` / `moai cg` GLM panels) collapses Claude's 5-level effort (max / xhigh / high / medium / low) into GLM's 3-level reasoning_effort (high / max). Implementation:

- `IsGLMBackend` detection identifies GLM sessions
- 5-level → 3-level collapse mapping (max/xhigh → max, high → high, medium/low → GLM unsupported)
- coding-max override for coding tasks

**Implemented + wired, wire validity pending live verification** — whether z.ai's Anthropic-compat shim actually consumes the `ANTHROPIC_REASONING_EFFORT` environment variable value is a run-phase verification item requiring live GLM session outbound observation. This page does not state "works guaranteed"; it states "implemented + wired, wire validity pending live verification."

## Model Policy Board (moai web)

The `/model-policy` board in `moai web` visually displays and configures plan_type and tier profiles. This board is an approved exception under SPEC-WEB-CONSOLE-013 allowing plan_type writes.

## Roadmap

{{< icon clock >}} **Spawn-time 36-cell routing** (SPEC-MODEL-TIER-ROUTING-PROFILES-001) — currently ApplyTierProfile operates at the agent level. Precision 36-cell routing combining phase and SPEC Tier at spawn-time is a descoped follow-up SPEC. Currently, agent frontmatter model/effort is replace-both'd by ApplyTierProfile.

## Next Steps

- [3-Tier Agent Architecture](/en/advanced/no-haiku-3tier/) — DeepSWE leaderboard rationale and 3-tier definition
- [Tokenomics Overview](/en/advanced/tokenomics-overview/) — Layer B routing of the 4-layer tokenomics structure
