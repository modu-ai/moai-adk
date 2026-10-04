# Agents Reference - MoAI-ADK Agent Catalog

Purpose: Reference catalog of MoAI-ADK's retained agents. MoAI delegates specialist work to a small, flat set of agents — there is no tier hierarchy and no `{domain}-{role}` naming scheme.

Version: 3.0.0

---

## Quick Reference (30 seconds)

MoAI delegates specialist tasks to **11 retained agents**: 10 MoAI-custom agents plus the Anthropic built-in `Explore`. The catalog is intentionally flat — agents are invoked through natural-language delegation, never via a `subagent_type` string literal in code.

| Agent | Phase scope |
|-------|-------------|
| `manager-spec` | Plan-phase artifact authoring (spec / plan / acceptance / research / design) |
| `manager-develop` | Run-phase implementation (cycle_type ∈ {ddd, tdd, autofix}) |
| `manager-docs` | Sync-phase documentation + project-doc scaffolding |
| `manager-git` | PR creation per Tier-based routing + branch closure |
| `plan-auditor` | Independent plan-phase audit, bias prevention, GEARS compliance |
| `sync-auditor` | Independent post-implementation quality scoring (4 dimensions) |
| `builder-harness` | Dynamic project-specific harness specialist generation |
| `super-advisor` | On-demand high-reasoning consultation (non-binding, E1-E4 escalation) |
| `manager-design` | Design-phase collaboration (Claude Design bidirectional sync) |
| `e2e-tester` | E2E test execution across web/mobile/desktop (CLI-first, token-minimized) |
| `Explore` | Read-only codebase exploration (Anthropic built-in) |

Agent Selection:
- Simple (1 file): 1 agent, or direct execution
- Medium (3-5 files): 1-2 agents sequential
- Complex (10+ files): multiple agents (parallel where independent)

Delegation pattern (natural language, not a code call):
- "Use the manager-develop subagent to implement the API (cycle_type=tdd)."
- "Use the Explore subagent to analyze the codebase structure."

---

## Implementation Guide (5 minutes)

### Flat Catalog (no tiers)

MoAI-ADK does NOT use a tier hierarchy or a `{domain}-{role}` naming convention. The catalog is a flat set of 11 retained agents, aligned with Anthropic's published guidance: "Subagents cannot spawn other subagents", "Start with 3-5 teammates for most workflows", and "Define a custom subagent when you keep spawning the same kind of worker".

### Selection Decision Tree

1. Read-only codebase exploration? Use the `Explore` subagent (Anthropic built-in).
2. External documentation or API research? Use WebSearch / WebFetch.
3. SPEC plan-phase authoring (spec / plan / acceptance / research / design)? Use the `manager-spec` subagent.
4. Run-phase implementation (DDD / TDD / autofix)? Use the `manager-develop` subagent with the appropriate `cycle_type`.
5. Sync-phase documentation (CHANGELOG / README / docs)? Use the `manager-docs` subagent.
6. PR creation per Tier-based routing (Tier L OR explicit `--pr`)? Use the `manager-git` subagent.
7. Plan-phase independent audit (bias prevention)? Use the `plan-auditor` subagent.
8. Post-implementation quality scoring? Use the `sync-auditor` subagent.
9. Dynamic specialist generation (project-specific harness)? Use the `builder-harness` subagent.
10. On-demand high-reasoning consultation / second opinion (E1-E4 escalation)? Use the `super-advisor` subagent.
11. Design-phase collaboration (Claude Design bidirectional sync, UI-surfaced SPECs)? Use the `manager-design` subagent.
12. E2E test execution across web/mobile/desktop (journey scripting, CLI-first suite runs)? Use the `e2e-tester` subagent.

### Retained Agents (detail)

| Agent | Class | Phase scope |
|-------|-------|-------------|
| `manager-spec` | core / manager | Plan-phase artifact authoring (spec / plan / acceptance / research / design); emits initial `status: draft` |
| `manager-develop` | core / manager | Run-phase implementation across three `cycle_type` modes: `tdd` (RED-GREEN-REFACTOR), `ddd` (ANALYZE-PRESERVE-IMPROVE), `autofix` (localize → repair → validate) |
| `manager-docs` | core / manager | Sync-phase documentation (CHANGELOG, README, docs) + frontmatter status transitions + project-doc scaffolding (product / structure / tech) |
| `manager-git` | core / manager | PR creation per Tier-based routing + branch closure |
| `plan-auditor` | meta / evaluator | Independent plan-phase audit, bias prevention, GEARS / EARS compliance verification |
| `sync-auditor` | meta / evaluator | Independent skeptical quality assessment, 4-dimension scoring (Functionality / Security / Craft / Consistency) |
| `builder-harness` | builder | Dynamic project-specific harness specialist generation |
| `super-advisor` | meta / advisor | On-demand high-reasoning consultation; returns non-binding prescriptions (E1-E4 escalation entry) |
| `manager-design` | core / manager | Design-phase collaboration (Claude Design bidirectional sync, D1-D5 pipeline) |
| `e2e-tester` | core / specialist | E2E test execution (web/mobile/desktop journey scripting, CLI-first runs, artifact management) |
| `Explore` | Anthropic built-in | Read-only codebase exploration (invoked directly, no MoAI file) |

### Archived Agent Names (rejected at spawn)

Legacy agent names from the former tiered catalog are **archived** and MUST NOT be spawned. The archived set is: `manager-strategy`, `manager-quality`, `manager-brain`, `manager-project`, `claude-code-guide`, `researcher`, and the six `expert-*` agents (`expert-backend`, `expert-frontend`, `expert-security`, `expert-devops`, `expert-performance`, `expert-refactoring`). When a delegation references an archived agent, the orchestrator rejects the spawn (`ARCHIVED_AGENT_REJECTED`) and routes the work to one of the 11 retained agents above, or to a per-spawn general-purpose agent with a domain whitelist. Domain-specific work (backend / frontend / security / performance / refactoring) is handled in run-phase by `manager-develop` or a per-spawn general-purpose agent, not by a dedicated tiered expert.

---

## Advanced Implementation (10+ minutes)

### Delegation Principles

1. Delegate, don't execute: MoAI orchestrates and delegates implementation to specialist agents; it does not implement complex tasks directly.
2. Natural-language invocation: agents are invoked through prose delegation ("Use the {agent} subagent to {task}"), carrying full context, constraints, and rationale — never a `subagent_type` code literal.
3. Context passing: pass each agent's results as context to the next agent in a sequence.
4. Sequential vs parallel: sequential when dependencies exist; parallel (multiple agents in one turn) when work is independent.
5. Flat hierarchy: MoAI subagents do not nest — they are configured without the `Agent` tool, so a subagent cannot spawn further subagents.

### Orchestration Primitives

Three runtime primitives exist; choose by who holds the plan:

| Primitive | Who decides next step | Scale |
|-----------|----------------------|-------|
| Sub-agents (`Agent()`) | Claude, turn by turn | A few delegated tasks per turn |
| Agent Teams | Claude + teammates via shared TaskList | 3-5 teammates |
| Dynamic Workflows | The orchestration script | Dozens-to-hundreds of agents per run |

For coding-heavy work prefer sequential sub-agents; reserve workflow-scale fan-out for genuinely parallel, high-volume tasks (codebase sweeps, large migrations, cross-checked research).

### Agent Selection Criteria

| Task Complexity | Files | Architecture Impact | Strategy |
|----------------|-------|---------------------|----------|
| Simple | 1 file | None | 1 agent or direct execution |
| Medium | 3-5 files | Moderate | 1-2 agents, sequential |
| Complex | 10+ files | High | Multiple agents, parallel where independent |

---

## Works Well With

Skills:
- [moai-foundation-core](../SKILL.md) - Parent skill (this module is part of it)
- [moai-foundation-cc](../../moai-foundation-cc/SKILL.md) - Claude Code configuration and authoring

Other Modules:
- [delegation-patterns.md](delegation-patterns.md) - Delegation strategies
- [commands-reference.md](commands-reference.md) - Command catalog and command-to-agent mapping

Commands (current `/moai` subcommands → agent):
- `/moai plan` → `manager-spec`
- `/moai run` → `manager-develop`
- `/moai sync` → `manager-docs`
- `/moai project` → `manager-docs` (project-doc scaffolding)

---

Maintained by: MoAI-ADK Team
Status: Production Ready
