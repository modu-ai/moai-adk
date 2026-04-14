---
id: SPEC-TOKEN-001
title: Agent Skill Binding Reduction for Token Budget Optimization
version: "1.0.0"
status: planned
created: "2026-03-30"
updated: "2026-03-30"
author: GOOS
priority: P0
---

# SPEC-TOKEN-001: Agent Skill Binding Reduction

## Background

Internal analysis revealed that MoAI-ADK agents have excessive skill bindings:
- expert-frontend: 27 skills
- expert-backend: 27 skills
- manager-spec: 13 skills
- Total system: ~305K tokens allocated (152% of 200K context budget)

Best-practice agent definitions use minimal skills in frontmatter, relying on orchestrator-injected domain knowledge.

## Requirements (EARS Format)

### R1: Agent Skill Reduction
WHEN an agent definition has more than 4 skills in frontmatter,
THE system SHALL reduce skills to a maximum of 4, keeping only:
1. moai-foundation-core (universal)
2. 1 domain skill (most relevant to role)
3. 1-2 workflow/tool skills (specific to agent's phase)

### R2: Language Skill Removal
WHERE language-specific skills (moai-lang-*) are bound in agent frontmatter,
THE system SHALL remove ALL language skills from agent frontmatter.

### R3: JIT Language Detection
WHEN MoAI orchestrator spawns an expert or manager agent,
THE system SHALL detect the project's primary language from root indicator files
AND include the appropriate language skill reference in the agent spawn prompt.

### R4: Foundation Skill Rationalization
WHERE moai-foundation-claude is bound to non-builder agents,
THE system SHALL remove it, as it is only needed by builder-* agents.
WHERE moai-foundation-philosopher is bound to agents,
THE system SHALL remove it unless the agent requires strategic thinking (manager-strategy, manager-project).

### R5: Backward Compatibility
WHILE reducing skill bindings,
THE system SHALL preserve all agent functionality by ensuring
language and domain knowledge is available via JIT loading.

## Scope

### In Scope
- Modify 17 agent definition files (those with >4 skills)
- Add JIT language detection guidance to run.md workflow
- Update CLAUDE.md Section 4 with JIT skill loading note

### Out of Scope
- Agent body content reduction (deferred to SPEC-AGENT-001)
- Workflow file modularization (deferred to SPEC-INFRA-001)
- New agent-extending reference skills (deferred to SPEC-AGENT-001)
- Anti-trigger additions (deferred to SPEC-AGENT-001)

## Impact Analysis

### Files Modified (17 agent definitions)
1. expert-frontend.md (27 -> 4 skills)
2. expert-backend.md (27 -> 4 skills)
3. manager-spec.md (13 -> 4 skills)
4. expert-testing.md (12 -> 3 skills)
5. expert-debug.md (11 -> 3 skills)
6. expert-devops.md (10 -> 3 skills)
7. manager-strategy.md (9 -> 4 skills)
8. manager-project.md (9 -> 4 skills)
9. manager-docs.md (9 -> 3 skills)
10. manager-ddd.md (9 -> 3 skills)
11. expert-performance.md (9 -> 3 skills)
12. manager-tdd.md (7 -> 3 skills)
13. manager-quality.md (7 -> 3 skills)
14. expert-security.md (7 -> 3 skills)
15. manager-git.md (6 -> 3 skills)
16. expert-refactoring.md (6 -> 3 skills)
17. builder-skill.md (4 -> 3 skills)

### Files Added/Modified (JIT loading)
18. workflows/run.md (add JIT language detection section)

### Unchanged (7 agents already at target)
- builder-plugin.md (3 skills)
- builder-agent.md (3 skills)
- team-designer.md (3 skills)
- team-tester.md (2 skills)
- team-validator.md (1 skill)
- team-reader.md (1 skill)
- team-coder.md (0 skills)

## Token Savings Estimate

| Category | Before | After | Savings |
|----------|--------|-------|---------|
| Total skill bindings | ~190 | ~55 | ~135 bindings |
| Estimated token reduction | ~200K | ~80K | ~120K tokens |
| Budget utilization | 152% | ~60% | 92% improvement |
