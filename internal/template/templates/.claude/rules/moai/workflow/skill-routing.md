# Skill Routing Protocol

Canonical rule for dynamic skill chaining: how the orchestrator routes domain skills into agent spawns, and how agents load conditional skills on demand.

## 1. Orchestrator Obligation

[ZONE:Evolvable] [HARD] Before spawning an implementation or review `Agent()`, the orchestrator MUST match the mission's domain against the available `moai-ref-*` / `moai-domain-*` skill descriptions and inject an explicit instruction into the spawn prompt for each matched skill (0-3 matches maximum):

```
At start, invoke Skill("<name>") for <reason>.
```

Examples:
- Backend API implementation → `At start, invoke Skill("moai-ref-api-patterns") for REST endpoint and error-handling conventions.`
- Security-sensitive review → `At start, invoke Skill("moai-ref-owasp-checklist") for the OWASP Top 10 review baseline.`
- React/Next.js work → `At start, invoke Skill("moai-ref-react-patterns") for component and state-management patterns.`

When no skill description matches the mission's domain, inject nothing — zero matches is a valid outcome, not a gap.

## 2. Agent Obligation

Agents whose `tools:` include `Skill` load conditional skills per the "Conditional Skill Loading" section in their own body. The static `skills:` frontmatter preload stays at most 2 entries per agent (token diet — progressive disclosure does the rest). An agent loads a conditional skill when its body's stated trigger situation actually arises, not preemptively.

## 3. Rationale

The two loading mechanisms have different cost profiles:

- `skills:` frontmatter injects each listed skill's FULL body into the agent context at spawn — a fixed cost paid on every invocation, whether or not the skill is used.
- `Skill()` invocation loads on demand: only the ~100-token metadata line is always visible; the ~5K-token body is paid only when the skill is actually invoked.

Keeping the static preload minimal and routing the rest through explicit `Skill()` instructions converts a fixed per-spawn cost into a pay-per-use cost, while the orchestrator-side injection (section 1) preserves discoverability for domain skills the agent would not know to load.

## 4. Cross-references

- `.claude/rules/moai/core/moai-constitution.md` § Agent Core Behaviors — cross-cutting agent obligations
- `.claude/rules/moai/development/agent-authoring.md` — agent frontmatter format (`skills:` YAML array, `tools:` CSV) and the Extension-Mechanism Context-Cost Ladder
- `CLAUDE.md` §4 — the retained agent catalog this protocol applies to
- `.claude/rules/moai/development/skill-authoring.md` § Progressive Disclosure — the 3-level token budget behind the on-demand cost profile

---

Classification: Evolvable operational rule — applies to all agent spawns and agent bodies with the Skill tool.
