# MoAI Agent Cross-References — Full Inventory

This skill orchestrates but does NOT replace existing agents. All agents referenced below are static MoAI agents — no new agents are introduced.

## Planning & Strategy

- `manager-spec` — Discovery (Phase 1) and Synthesis (Phase 3)
- `manager-strategy` — Analysis codebase scan (Phase 2)
- `plan-auditor` — EARS compliance check on the SPEC produced in Phase 3

## Implementation

- `expert-backend` — Backend domain harness templates
- `expert-frontend` — Frontend domain harness templates
- `expert-devops` — DevOps/platform domain harness templates
- `expert-security` — Security review of generated permissions
- `manager-develop` (cycle_type=tdd) — Test harness pattern generation (former expert-testing capability)
- `manager-quality` (diagnostic-mode) — Debug agent patterns (former expert-debug capability)
- `expert-refactoring` — Refactoring workflow patterns
- `expert-performance` — Performance profiling patterns

## Builders

- `builder-harness` (artifact_type=agent) — Generates `.claude/agents/harness/*.md` content
- `builder-harness` (artifact_type=skill) — Generates `.claude/skills/moai-harness-*/SKILL.md` content
- `builder-harness` (artifact_type=plugin) — Optional plugin bundling of generated artifacts

## Workflow Managers

- `manager-develop` (`cycle_type=ddd` or `cycle_type=tdd` per `.moai/config/sections/quality.yaml` `development_mode`) — DDD or TDD-flavored harness workflow templates (the retired-DDD policy M3 consolidated the prior DDD and TDD specialist managers into the unified `manager-develop` agent with cycle-type dispatch)
- `manager-quality` — Quality gate configuration in generated harnesses
- `manager-docs` — Documentation generation patterns
- `manager-git` — Git workflow patterns for generated harnesses

## Quality

- `sync-auditor` — Sprint Contract evaluation (Phase 6)
