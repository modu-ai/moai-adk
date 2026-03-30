---
id: SPEC-AGENT-001
title: Agent Architecture Modernization - Anti-Triggers and Reference Skills
version: "1.0.0"
status: approved
created: "2026-03-30"
updated: "2026-03-30"
author: GOOS
priority: P1
---

# SPEC-AGENT-001: Agent Architecture Modernization (Phase 1)

## Background

Internal analysis identified key patterns missing from MoAI-ADK agents:
1. Anti-triggers ("NOT for:") - 19 agents lack explicit scope boundaries
2. Agent-extending reference skills - Pure knowledge skills that amplify agent expertise
3. Agent body reduction to ≤5KB (deferred to Phase 2 - separate session)

## Requirements (EARS Format)

### R1: Anti-Triggers for Agent Descriptions
WHERE an agent definition has a description field,
THE system SHALL include explicit "NOT for:" scope boundaries
listing domains and tasks the agent should NOT handle.

### R2: Agent-Extending Reference Skills
WHEN domain-specific knowledge is needed by agents,
THE system SHALL provide it via dedicated reference skills
with a "Target Agent" declaration.

### R3: Reference Skill Structure
WHEN creating agent-extending reference skills,
THE system SHALL follow the agent-extending skill pattern:
- Pure reference content (tables, checklists, patterns)
- "Target Agent" declaration
- NOT procedural workflow instructions

## Scope

### In Scope (This Session)
- Add anti-trigger lines to 17 agent descriptions (managers + experts)
- Create 3 agent-extending reference skills:
  - moai-ref-api-patterns (target: expert-backend)
  - moai-ref-react-patterns (target: expert-frontend)
  - moai-ref-git-workflow (target: manager-git)

### Out of Scope (Deferred to SPEC-AGENT-002)
- Agent body content reduction to ≤5KB
- Deliverable template sections for all agents
- Communication protocol sections for all agents
- Additional reference skills (5+ more)

## Impact Analysis

### Files Modified (17 agent descriptions)
All agents in internal/template/templates/.claude/agents/moai/ with >3 skills

### Files Created (3 new reference skills)
1. .claude/skills/moai-ref-api-patterns/skill.md
2. .claude/skills/moai-ref-react-patterns/skill.md
3. .claude/skills/moai-ref-git-workflow/skill.md
