---
spec_id: SPEC-AGENT-001
plan_version: "0.1.0"
plan_date: 2026-05-14
plan_author: manager-spec (auto)
plan_status: draft
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-14 | manager-spec | Initial plan |

## 1. Plan Overview

SPEC-AGENT-001 adds anti-trigger ("NOT for:") scope boundaries to agent descriptions and creates 3 agent-extending reference skills. Codebase analysis reveals significant pre-existing implementation:

- **Anti-triggers**: Most agents already have "NOT for:" lines in their description fields (e.g., evaluator-active.md line 11, manager-quality.md line 11, expert-backend.md, expert-frontend.md).
- **Reference skills**: 5 reference skills already exist (`moai-ref-api-patterns`, `moai-ref-react-patterns`, `moai-ref-git-workflow`, `moai-ref-owasp-checklist`, `moai-ref-testing-pyramid`), exceeding the 3 originally scoped.

The remaining work is gap-closure: verifying all 17 agents have anti-triggers, confirming reference skills follow the agent-extending pattern, and closing any coverage gaps.

## 2. Gap Analysis

### Already Implemented

| R | Status | Evidence |
|---|--------|----------|
| R1 | PARTIAL | Anti-triggers present in most agents; verify all 17+ agents |
| R2 | DONE | 5 reference skills exist (3 original + 2 additional) |
| R3 | DONE | Skills follow agent-extending pattern with "Target Agent" declarations |

### Remaining Gaps

1. **Anti-trigger audit**: Need to verify every agent in `.claude/agents/moai/` has a "NOT for:" line. Retired agents (manager-ddd.md, manager-tdd.md) are exempt.
2. **Coverage completeness**: Verify builder-* and claude-code-guide agents have anti-triggers.

## 3. Milestone Breakdown

### M1 -- Anti-Trigger Audit -- Priority P0

Audit all agent definitions for "NOT for:" anti-trigger presence:
- Scan all `.claude/agents/moai/*.md` files
- Identify agents missing anti-triggers
- Add anti-trigger lines where missing
- Exclude retired agents from the audit

Files:
- `internal/template/templates/.claude/agents/moai/*.md` (all agent files)

### M2 -- Reference Skill Verification -- Priority P0

Verify all reference skills follow the agent-extending pattern:
- Confirm each skill has "Target Agent" declaration
- Confirm skills contain pure reference content (tables, checklists, patterns)
- Confirm skills are NOT procedural workflow instructions
- Verify 3 originally scoped skills exist: moai-ref-api-patterns, moai-ref-react-patterns, moai-ref-git-workflow

Files:
- `internal/template/templates/.claude/skills/moai-ref-api-patterns/skill.md`
- `internal/template/templates/.claude/skills/moai-ref-react-patterns/skill.md`
- `internal/template/templates/.claude/skills/moai-ref-git-workflow/skill.md`

### M3 -- Gap Closure -- Priority P1

Fix any identified gaps from M1 and M2:
- Add missing anti-triggers to agent descriptions
- Fix any reference skills that deviate from the agent-extending pattern
- Update agent frontmatter descriptions as needed

Files:
- Individual agent files identified in M1 audit

## 4. File:line Anchors

| File | Line(s) | Action | Purpose |
|------|---------|--------|---------|
| `internal/template/templates/.claude/agents/moai/evaluator-active.md` | 11 | Verify | "NOT for:" anti-trigger example |
| `internal/template/templates/.claude/agents/moai/expert-backend.md` | description | Verify/Audit | Anti-trigger presence |
| `internal/template/templates/.claude/agents/moai/expert-frontend.md` | description | Verify/Audit | Anti-trigger presence |
| `internal/template/templates/.claude/agents/moai/builder-agent.md` | description | Verify/Audit | Anti-trigger presence |
| `internal/template/templates/.claude/skills/moai-ref-api-patterns/skill.md` | frontmatter | Verify | Target Agent declaration |

## 5. Quality Gates

- All active (non-retired) agents have "NOT for:" anti-trigger in description
- All 3 originally scoped reference skills exist and follow agent-extending pattern
- No regression in existing agent functionality
- Agent frontmatter valid per agent-authoring rules

## 6. Dependencies

- SPEC-AGENT-002: Agent body reduction (separate SPEC, out of scope)
- SPEC-V3R3-RETIRED-AGENT-001: Agent retirement already completed
