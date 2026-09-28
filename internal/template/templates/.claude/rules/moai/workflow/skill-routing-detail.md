---
description: "Detail companion for skill-routing.md — why the static preload and on-demand invocation have different cost profiles"
paths: "**/skill-routing.md,**/skill-authoring.md,**/agent-authoring.md"
---

# Skill Routing — Detail Companion

> Detail companion of `skill-routing.md` (the always-loaded stub). The stub owns the orchestrator
> injection obligation, the orchestrator-direct routing table, the intent-based matching rule, and
> the agent-side obligation. This file owns the cost reasoning behind them.

## Why the two loading mechanisms differ

The two loading mechanisms have different cost profiles:

- `skills:` frontmatter injects each listed skill's FULL body into the agent context at spawn — a fixed cost paid on every invocation, whether or not the skill is used.
- `Skill()` invocation loads on demand: only the ~100-token metadata line is always visible; the ~5K-token body is paid only when the skill is actually invoked.

Keeping the static preload minimal and routing the rest through explicit `Skill()` instructions converts a fixed per-spawn cost into a pay-per-use cost, while the orchestrator-side injection (section 1) preserves discoverability for domain skills the agent would not know to load.


---

Classification: Lazy companion — cost rationale only. Every routing obligation stays in
`skill-routing.md`.

## 2. Agent Obligation

Agents whose `tools:` include `Skill` load conditional skills per the "Conditional Skill Loading" section in their own body. The static `skills:` frontmatter preload stays at most 2 entries per agent (token diet — progressive disclosure does the rest). An agent loads a conditional skill when its body's stated trigger situation actually arises, not preemptively.

## 3. Rationale

The two loading mechanisms have different cost profiles: `skills:` frontmatter injects each listed
skill's full body at spawn, paid whether or not it is used, while `Skill()` pays ~100 tokens of
metadata always and the ~5K body only on invocation. Keeping the static preload minimal and routing
the rest through explicit `Skill()` instructions converts a fixed per-spawn cost into a
pay-per-use one. Worked detail: `skill-routing-detail.md`.

