---
title: MoAI-ADK Documentation
weight: 99
draft: false
---

MoAI-ADK (Agentic Development Kit) is a strategic orchestration framework for Claude Code.

> **Current version:** {{< version >}} — version information follows the single source of truth `params.version` in `hugo.toml`.

> {{< icon book primary >}} **[Practical Agentic Coding with Claude Code](/book)** — the end of vibe coding, the beginning of harness engineering. A 488-page practical guide written by MoAI-ADK maintainers (with endorsements from 9 experts).

![MoAI-ADK](/og.jpg)

![Documentation map](/images/sections/doc-map-en.png)

## New in v3.2 — Factory Mode {{< new-badge v3.2 >}}

A session holds one context window, and a long SPEC fills it — everything that follows carries everything that came before. Factory Mode spreads the work across **several terminals**: one leader session routes picked cards, and each lane is a session of its own that carries one card whole through `plan`, `run`, and `sync`. The review verdict is not a separate stage: the sync gate absorbs it. Limits do not disappear, but no session carries more than one card's history, so the same budget goes considerably further.

Each lane can run a different backend. The leader and every lane choose between `cc`, `glm`, and `codex` on their own.

{{< terminal title="factory mode" raw="true" >}}
moai cc -f                    # leader — opens the factory run
moai cc -l                    # lane, in its own terminal — joins as the next free lane
moai glm -l                   # a lane on the GLM backend
moai codex -l                 # a Codex lane
{{< /terminal >}}

The queue is the channel: `backlog` has no owning session by design — work enters it only when you put it there with [`/moai todo`](/en/utility-commands/moai-todo). The leader advances a card only on evidence it read from the card's `progress.md`, never on a lane's reply.

Run `moai web` to watch the factory lanes and the SPEC pipeline side by side on the Factory screen.

![moai web console — Overview screen with SPEC counts, in-progress SPECs, and session registry](/images/profile/web-console-v31-overview.png)

More: [Factory Mode](/en/advanced/factory-mode) · [manager-lead Leader Coordinator](/en/advanced/manager-lead) · [`/moai todo`](/en/utility-commands/moai-todo) · [moai web console](/en/advanced/moai-web-console)

## Three Core Values of MoAI 3.1

- {{< icon database primary >}} **Tokenomics** — Reduces inference costs by 60-70% through context dieting and prompt caching. See [Multi-LLM](/en/multi-llm), [Cost Optimization](/en/cost-optimization), and [Advanced/Tokenomics Overview](/en/advanced/tokenomics-overview).

- {{< icon rotate primary >}} **Agentic Loop Engineering** — An autonomous improvement cycle where loops work on their own and observations accumulate so harness guidance evolves (recursive self-learning). See [Self-Evolving Systems](/en/advanced/self-evolving), [Autonomous Loops](/en/advanced/autonomous-loops), and [Decision Memory](/en/advanced/decision-memory).

- {{< icon package primary >}} **Agentic Harness** — Composable execution environment with skills, hooks, and MCP for extensible agent orchestration. See [Core Concepts](/en/core-concepts), [Workflow Commands](/en/workflow-commands), and [Agent Guide](/en/advanced/agent-guide).

## Key Features

- **MoAI Orchestrator**: Strategic task delegation through specialized agents
- **SPEC-Based TDD/DDD**: Adaptive methodology — TDD for new projects, DDD for legacy code
- **TRUST 5 Framework**: 5 quality principles: Tested, Readable, Unified, Secured, Trackable
- **Progressive Disclosure**: 3-level skill loading for 67% token reduction

## Getting Started

To start with MoAI-ADK, see the [Getting Started](/en/getting-started) section.

## Documentation Structure

- [Getting Started](/en/getting-started) - Installation, basic setup, quick start
- [Core Concepts](/en/core-concepts) - SPEC format, agents, workflows
- [Advanced](/en/advanced) - Advanced patterns, skill usage, performance optimization
- [Git Worktree](/en/worktree) - Complete Git Worktree CLI guide
- {{< icon book primary >}} [Book: Practical Agentic Coding](/book) - Master agentic coding with MoAI-ADK (488 pages · endorsed by 9 experts)
