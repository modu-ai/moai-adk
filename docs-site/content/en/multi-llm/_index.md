---
title: "Multi-LLM"
weight: 80
draft: false
---

MoAI-ADK supports **z.ai GLM** as an alternative AI backend alongside the Claude API, enabling multi-LLM development workflows.

## What is z.ai GLM?

GLM (Generative Language Model) is an AI model service provided by z.ai that is compatible with Claude Code. You can switch backends with environment variables alone — no code changes required.

| Item | Details |
|------|---------|
| **GLM Coding Plan** | Starting at **$10**/month ([sign up](https://z.ai/subscribe?ic=1NDV03BGWU)) |
| **Compatibility** | Compatible with Claude Code — no code changes |
| **Models** | glm-5.2[1m], GLM-4.7, GLM-4.5-Air, free models |

## Default Model Mapping

| Claude Tier | GLM Model | Input (per 1M tokens) | Output (per 1M tokens) |
|-------------|-----------|-----------------------|------------------------|
| Opus / Sonnet / Haiku / Fable | glm-5.2[1m] | $2.00 | $8.00 |

> All four Claude tiers (Opus, Sonnet, Haiku, Fable) are unified to `glm-5.2[1m]` — a single 1M-context model. GLM models are intentionally NOT mapped disparately across these slots (e.g. opus→glm-5.2, sonnet→glm-4.7, haiku→glm-4.5-air) because mixing a 1M-context model with 200K-context models across those slots breaks agent-spawn session sharing: a 1M-context model and a 200K-context model cannot share a session.

> The mapping is implemented via four Claude Code `ANTHROPIC_DEFAULT_*_MODEL` environment variables — `ANTHROPIC_DEFAULT_OPUS_MODEL`, `ANTHROPIC_DEFAULT_SONNET_MODEL`, `ANTHROPIC_DEFAULT_HAIKU_MODEL`, and `ANTHROPIC_DEFAULT_FABLE_MODEL` — all set to `glm-5.2`. The Fable variable is officially supported as of Claude Code v2.1.202.

> Free models are also available: GLM-4.7-Flash, GLM-4.5-Flash. See [z.ai Pricing](https://docs.z.ai/guides/overview/pricing) for the full price list.

## Three Execution Modes

MoAI-ADK provides three LLM execution modes:

| Command | Leader | Worker | tmux required | Cost saving | Use case |
|---------|--------|--------|---------------|-------------|----------|
| `moai cc` | Claude | Claude | No | - | Highest quality, complex tasks |
| `moai glm` | GLM | GLM | Recommended | ~70% | Cost optimization |
| `moai cg` | Claude | GLM | **Required** | **~60%** | Quality + cost balance |

```mermaid
graph TD
    A["🗿 MoAI Orchestrator"] --> B{"Execution mode selection"}
    B -->|"moai cc"| C["Claude Only<br/>Highest quality"]
    B -->|"moai glm"| D["GLM Only<br/>Cost saving"]
    B -->|"moai cg"| E["CG Hybrid<br/>Balanced"]

    C --> F["Leader: Claude<br/>Worker: Claude"]
    D --> G["Leader: GLM<br/>Worker: GLM"]
    E --> H["Leader: Claude<br/>Worker: GLM"]

    style C fill:#7C3AED,color:#fff
    style D fill:#059669,color:#fff
    style E fill:#D97706,color:#fff
```

### Quick Start

```bash
# 1. Store your GLM API key (first time only)
moai glm sk-your-glm-api-key

# 2. Select a mode
moai cc            # Claude only
moai glm           # GLM only
moai cg            # CG hybrid (tmux required)
```

> **As of v2.7.1**, CG mode is the **default team mode** for the `--team` flag. It runs in CG mode unless you explicitly switch to `moai cc` or `moai glm`.

## Next Steps

- [CG Mode (Claude + GLM)](/en/multi-llm/cg-mode) — tmux isolation architecture details
- [Model Policy](/en/multi-llm/model-policy) — per-agent model assignment table
