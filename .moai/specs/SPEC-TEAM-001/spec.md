# SPEC-TEAM-001: Agent Teams Dynamic Generation Architecture

**Status**: Draft
**Priority**: High
**Created**: 2026-03-31
**Author**: MoAI + GOOS

## Summary

Replace static `team-*` agent definition files (`.claude/agents/moai/team-*.md`) with dynamic team generation using `Agent(subagent_type: "general-purpose")` and runtime parameter overrides. Team composition is driven by `workflow.yaml` role profiles and orchestrator-generated prompts.

## Background

MoAI currently maintains 5 static team agent definitions:
- team-reader, team-coder, team-tester, team-designer, team-validator

These files provide tools, model, permissionMode, skills, and hooks. However, research shows that most of these can be overridden at Agent() spawn time, making static files unnecessary overhead with misleading documentation ("AGENT TEAMS ONLY", deprecated `maxTurns`).

## Requirements (EARS Format)

### REQ-1: Remove Static Team Agent Definitions
**When** the MoAI template system deploys to a new project,
**the system shall** NOT include any `team-*` agent definition files in `.claude/agents/moai/`.

**Acceptance Criteria:**
- [ ] AC-1.1: `.claude/agents/moai/team-coder.md` deleted (local + template)
- [ ] AC-1.2: `.claude/agents/moai/team-reader.md` deleted (local + template)
- [ ] AC-1.3: `.claude/agents/moai/team-tester.md` deleted (local + template)
- [ ] AC-1.4: `.claude/agents/moai/team-designer.md` deleted (local + template)
- [ ] AC-1.5: `.claude/agents/moai/team-validator.md` deleted (local + template)
- [ ] AC-1.6: No Go code references these deleted agent files

### REQ-2: Add Role Profiles to workflow.yaml
**When** the orchestrator needs to spawn a teammate,
**the system shall** use role profiles from `workflow.yaml` to determine Agent() parameters.

**Acceptance Criteria:**
- [ ] AC-2.1: `workflow.yaml` contains `team.role_profiles` section
- [ ] AC-2.2: Each profile defines: mode, model, isolation
- [ ] AC-2.3: Profiles cover: researcher, analyst, architect, implementer, tester, designer, reviewer

### REQ-3: Dynamic Team Spawning in team/run.md
**When** the team run workflow executes,
**the system shall** spawn teammates using `Agent(subagent_type: "general-purpose")` with runtime parameter overrides from role profiles.

**Acceptance Criteria:**
- [ ] AC-3.1: team/run.md uses `subagent_type: "general-purpose"` for all teammates
- [ ] AC-3.2: model, mode, isolation are set via Agent() parameters (not agent definitions)
- [ ] AC-3.3: Teammate prompts include project-type-specific instructions
- [ ] AC-3.4: Quality verification instructions embedded in prompts (replacing agent-scoped hooks)

### REQ-4: Update Documentation
**When** developers read MoAI documentation,
**the system shall** accurately reflect the dynamic team generation architecture.

**Acceptance Criteria:**
- [ ] AC-4.1: CLAUDE.md Section 4 (Agent Catalog) removes Team Agents (5) section
- [ ] AC-4.2: CLAUDE.md Section 15 (Agent Teams) updated for dynamic generation
- [ ] AC-4.3: agent-authoring.md removes team-* agent section, adds dynamic team section
- [ ] AC-4.4: spec-workflow.md Agent Teams Variant updated

### REQ-5: Template Synchronization
**When** template files are modified,
**the system shall** maintain synchronization between local and template copies.

**Acceptance Criteria:**
- [ ] AC-5.1: `internal/template/templates/` mirrors local changes
- [ ] AC-5.2: `make build` succeeds after all changes
- [ ] AC-5.3: `go test ./...` passes

## Technical Approach

### Dynamic Team Generation Architecture

```
workflow.yaml                         Project Context
  role_profiles:                      (language, framework, MCP servers)
    researcher: {mode: plan, ...}           |
    implementer: {mode: acceptEdits, ...}   |
              |                             |
              +-----------------------------+
              |
              v
    MoAI Orchestrator (team/run.md)
    1. Select pattern from workflow.yaml
    2. Map roles to role_profiles
    3. Generate context-aware prompts
    4. Spawn: Agent(subagent_type: "general-purpose",
                    team_name: "...",
                    name: "...",
                    model: <from profile>,
                    mode: <from profile>,
                    isolation: <from profile>,
                    prompt: <dynamic>)
```

### Role Profile Schema

```yaml
team:
  role_profiles:
    researcher:
      mode: plan
      model: haiku
      isolation: none
      description: "Read-only codebase exploration and analysis"
    analyst:
      mode: plan
      model: sonnet
      isolation: none
      description: "Requirements analysis and validation"
    architect:
      mode: plan
      model: sonnet
      isolation: none
      description: "Solution design and architecture decisions"
    implementer:
      mode: acceptEdits
      model: sonnet
      isolation: worktree
      description: "Code implementation (backend, frontend, full-stack)"
    tester:
      mode: acceptEdits
      model: sonnet
      isolation: worktree
      description: "Test creation and coverage validation"
    designer:
      mode: acceptEdits
      model: sonnet
      isolation: worktree
      description: "UI/UX design with MCP design tools"
    reviewer:
      mode: plan
      model: haiku
      isolation: none
      description: "Code review and quality validation"
```

### Files to Modify

**Delete (10 files):**
- `.claude/agents/moai/team-coder.md`
- `.claude/agents/moai/team-reader.md`
- `.claude/agents/moai/team-tester.md`
- `.claude/agents/moai/team-designer.md`
- `.claude/agents/moai/team-validator.md`
- `internal/template/templates/.claude/agents/moai/team-coder.md`
- `internal/template/templates/.claude/agents/moai/team-reader.md`
- `internal/template/templates/.claude/agents/moai/team-tester.md`
- `internal/template/templates/.claude/agents/moai/team-designer.md`
- `internal/template/templates/.claude/agents/moai/team-validator.md`

**Modify (6+ files):**
- `.moai/config/sections/workflow.yaml` — Add role_profiles
- `internal/template/templates/.moai/config/sections/workflow.yaml` — Same
- `.claude/skills/moai/team/run.md` — Rewrite for dynamic generation
- `internal/template/templates/.claude/skills/moai/team/run.md` — Same
- `CLAUDE.md` — Update Sections 4, 15
- `internal/template/templates/CLAUDE.md` — Same
- `.claude/rules/moai/development/agent-authoring.md` — Update team section
- `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` — Same
- `.claude/rules/moai/workflow/spec-workflow.md` — Update Agent Teams Variant
- `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` — Same

### Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Loss of agent-scoped hooks | Medium | Global TeammateIdle/TaskCompleted hooks + prompt instructions |
| Loss of skill preloading | Low | Teammates use Skill() tool or receive instructions in prompt |
| Loss of tool restrictions | Low | mode: "plan" enforces read-only; write agents need all tools |
| general-purpose has excessive tools | Low | Acceptable trade-off for flexibility |

## Out of Scope

- Go CLI code changes (no Go code references team-* agent files directly)
- Hook handler code changes (agent hooks in internal/hook/agents/ remain for non-team agents)
- CG mode tmux env verification (separate investigation needed)
