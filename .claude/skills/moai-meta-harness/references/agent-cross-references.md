# MoAI Agent Cross-References — Full Inventory

This skill orchestrates but does NOT replace existing agents. All named agents below are retained MoAI agents — no new agents are introduced; domain-specific work is delegated via per-spawn Agent(general-purpose) with domain instructions.

## Planning & Strategy

- `manager-spec` — Discovery (Phase 1), Analysis codebase scan (Phase 2), and Synthesis (Phase 3); strategic analysis is absorbed into manager-spec
- `plan-auditor` — EARS compliance check on the SPEC produced in Phase 3

## Implementation

- Backend / Frontend / DevOps domain harness templates — per-spawn `Agent(general-purpose)` with the matching domain instructions injected at delegation time (per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C)
- Security review of generated permissions — per-spawn `Agent(general-purpose)` with security-review instructions (OWASP checks); dependency-manifest audit is mechanically enforced by the `sync-phase-quality-gate.sh` Stop hook
- `manager-develop` (cycle_type=tdd) — Test harness pattern generation (former expert-testing capability)
- `manager-develop` (cycle_type=autofix) — Debug / failure-diagnosis agent patterns (former expert-debug / manager-quality diagnostic capability)
- Refactoring / Performance domain patterns — per-spawn `Agent(general-purpose)` with refactoring / performance instructions

## Builders

- `builder-harness` (artifact_type=agent) — Generates `.claude/agents/harness/*.md` content
- `builder-harness` (artifact_type=skill) — Generates `.claude/skills/moai-harness-*/SKILL.md` content
- `builder-harness` (artifact_type=plugin) — Optional plugin bundling of generated artifacts

## Workflow Managers

- `manager-develop` (`cycle_type=ddd` or `cycle_type=tdd` per `.moai/config/sections/quality.yaml` `development_mode`) — DDD or TDD-flavored harness workflow templates (the retired-DDD policy M3 consolidated the prior DDD and TDD specialist managers into the unified `manager-develop` agent with cycle-type dispatch)
- Quality gate configuration in generated harnesses — the `/moai gate` skill or the `sync-phase-quality-gate.sh` Stop hook (former manager-quality quality-gate role)
- `manager-docs` — Documentation generation patterns
- `manager-git` — Git workflow patterns for generated harnesses

## Quality

- `sync-auditor` — Sprint Contract evaluation (Phase 6)
