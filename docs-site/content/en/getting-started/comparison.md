---
title: What's Different — Three Usage Patterns Compared
weight: 35
draft: false
---
# What's Different — Three Usage Patterns Compared

There are three broad ways to use Claude Code for agent-tic development. (1) Use Claude Code as-is, (2) Wrap it in a generic wrapper, (3) Wrap it in an agent-tic harness like MoAI-ADK. This page summarizes where the three patterns converge and diverge. The evaluation criteria are the three core pillars the README emphasizes — cost (tokenomics), self-improvement (loop engineering), and quality control (agent-tic harness).

## Three Usage Patterns

| | Claude Code alone | Generic wrapper | MoAI-ADK (agent-tic harness) |
|---|------------------|-----------------|------------------------------|
| **Identity** | The official model·tools themselves | A thin layer wrapping model calls | An execution environment (harness) wrapping models |
| **Distribution unit** | Official installation | Varies per tool | Single Go binary (no Python·runtime needed) |
| **Cost control** | Model uses whatever it wants | Usually none | Per-task model·inference-depth assignment + budget guard |
| **Quality gates** | User checks manually each time | Varies per tool | SPEC 3-phase + TRUST 5, auto-verification |
| **Learning loop** | Fresh start each session | Usually none | Self-evolution accumulating observations into rules |
| **Session continuity** | Breaks at every `/clear` | Varies per tool | Paste-ready history + auto-injection |
| **Agent roster** | Single session | Single session | 12-agent catalog + 3-phase workflow |

Claude Code alone is not bad. Rather, MoAI-ADK does NOT replace Claude Code — it **wraps** it and adds structure on top. Model routing, quality gates, cost control, learning loops, session continuity — the parts Claude Code left to users, the harness takes as system responsibility.

## Why These Three Core Pillars

Push just one and you fall into a trap. That's why the README sets up all three together.

{{< callout type="warning" >}}
- **Optimize only cost** and quality quietly breaks. The ensuing rework and debug loops become the most expensive token spend.
- **Set up quality gates only with no learning loop** and you repeat the same mistakes every session.
- **Run autonomous loops only with no cost ceiling** and one over-execution swallows the entire quota.
{{< /callout >}}

The three pillars support each other. Cost is economic when quality prevents rework; quality becomes enforceable when loops capture what worked; loops become affordable when cost gates stop them before overflow. Every design decision in MoAI-ADK serves one of these three.

## Cost — Tokenomics

Unit prices dropped 98% in three years (Linux Foundation), yet enterprise AI spend rose 320% over the same period. Usage increase covered the unit-price drop. An agent running a task burns through tens to hundreds of steps, spending tokens proportionally.

{{< icon target >}} **Claude Code alone / Generic wrapper** — The model plans its own steps, and the user watches cost. Even with low unit prices, many steps still make large bills.

{{< icon target primary >}} **MoAI-ADK** — What divides cost is not unit price but **assignment**. In the DeepSWE benchmark, Opus 5's lowest inference scored higher than Sonnet 5's highest while costing one-sixteenth per task. Retry loops write the bill, not token unit prices. So assign the right model and inference depth per task, diet the context, and stop before budget overflow. `moai cg`'s Claude+GLM hybrid mode brings 60-70% cost reduction on implementation-heavy work.

Covered in detail in [Tokenomics overview](/en/advanced/tokenomics-overview/) and [Cost optimization](/en/cost-optimization/).

## Self-Improvement — Agent-tic Loop Engineering

The cheapest session is the one that doesn't repeat the last session's mistakes.

{{< icon rotate >}} **Claude Code alone / Generic wrapper** — When the session ends, observations disappear with it. The next session starts from scratch every time.

{{< icon rotate primary >}} **MoAI-ADK** — Each execution becomes material for the next. Routing decisions and gate evidence are recorded, recurring patterns become rules, declared goals (`/moai goal`) drive the session until conditions are satisfied. Observed failure patterns surface as rule-change proposals, not silently applied — they require approval. Realistic short-term self-improvement is recursively improving the **harness around the model**, not the model weights themselves.

Covered in detail in [Self-evolving system](/en/advanced/self-evolving/), [Autonomous loops](/en/advanced/autonomous-loops/), [Decision memory](/en/advanced/decision-memory/).

## Quality Control — Agent-tic Harness

Rework is the largest token waste. A bug that returns once costs more than all routing optimization combined.

{{< icon package >}} **Claude Code alone** — "All done" means the user must manually verify every time.

{{< icon package >}} **Generic wrapper** — Quality standards vary per tool or don't exist at all.

{{< icon package primary >}} **MoAI-ADK** — Changes "done" to *verified done*. SPEC 3-phase (plan → run → sync) and TRUST 5 gates (tested·readable·unified·secured·trackable) apply to every change. Gates judge verification, not agents. The 12-agent catalog separates planning and auditing from the start, so the planning side can't mark its own homework. The [Verification-Claim Integrity](/en/core-concepts/verification-claim-integrity/) rule prevents unobserved "passes" from slipping through as gaps.

Covered in detail in [Harness engineering](/en/core-concepts/harness-engineering/), [TRUST 5 Quality](/en/core-concepts/trust-5/), [SPEC-based development](/en/core-concepts/spec-based-dev/).

## Which Pattern to Choose

- **Claude Code alone** is sufficient — For exploratory coding, simple fixes to one or two files, short sessions where you directly monitor cost and quality.
- **Generic wrapper** fits — When you want to automate just one specific workflow and need no further structure.
- **MoAI-ADK** is needed — When the system must take responsibility for per-session cost and quality, when agents working in parallel must not step on each other, when last session's learning must carry into the next.

The three patterns are not mutually exclusive. MoAI-ADK wraps Claude Code, it doesn't replace it. A single Go binary runs on macOS·Linux·Windows with no additional dependencies.

## Next Steps

- [Installation](/en/getting-started/installation/) — Single binary installation
- [What is MoAI-ADK?](/en/core-concepts/what-is-moai-adk/) — Identity and philosophy
- [Harness engineering](/en/core-concepts/harness-engineering/) — Where the three core pillars meet
