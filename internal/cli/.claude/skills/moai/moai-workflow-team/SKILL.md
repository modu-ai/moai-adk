---
name: moai-workflow-team
description: >
  Agent Teams workflow management for MoAI-ADK. Handles team creation,
  teammate spawning, task decomposition, inter-agent messaging, and
  graceful shutdown. Integrates with SPEC workflow for team-based
  Plan and Run phases. Supports dual-mode execution with automatic
  fallback to sub-agent mode when teams are unavailable.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Task TeamCreate TeamDelete SendMessage TaskCreate TaskUpdate TaskList TaskGet Read Grep Glob AskUserQuestion
user-invocable: false
metadata:
  version: "1.1.0"
  category: "workflow"
  status: "experimental"
  updated: "2026-02-07"
  modularized: "false"
  tags: "team, agent-teams, collaboration, parallel, dual-mode"
  related-skills: "moai-workflow-spec, moai-workflow-ddd, moai-workflow-tdd"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 8000

# MoAI Extension: Triggers
triggers:
  keywords: ["team", "agent-team", "parallel", "collaborate", "teammates", "--team"]
  agents: ["moai"]
  phases: ["plan", "run"]
---

# MoAI Agent Teams Workflow

## Overview

This skill manages Agent Teams execution for MoAI workflows. When team mode is selected (via --team flag, auto-detection, or configuration), MoAI operates as Team Lead coordinating persistent teammates.

## Prerequisites

Agent Teams requires:
- `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` in settings.json env
- `workflow.team.enabled: true` in workflow.yaml
- Claude Code v2.1.32 or later

## Mode Selection

The mode selector determines execution strategy:

1. Check --team/--solo flags (user override)
2. Check workflow.yaml execution_mode setting
3. If "auto": Analyze complexity score
   - Domain count >= 3: team mode
   - Affected files >= 10: team mode
   - Complexity score >= 7: team mode
   - Otherwise: sub-agent mode
4. Verify AGENT_TEAMS is enabled
5. If not enabled: warn user, fall back to sub-agent

## Team Lifecycle

### Phase 1: Team Creation

```
TeamCreate(team_name: "moai-{workflow}-{timestamp}")
```

Team naming convention:
- Plan phase: `moai-plan-SPEC-XXX`
- Run phase: `moai-run-SPEC-XXX`
- Debug: `moai-debug-{issue}`
- Review: `moai-review-{target}`

### Phase 2: Task Decomposition

Before spawning teammates, create the complete shared task list:

```
TaskCreate(subject: "Task description", description: "Detailed requirements")
```

Rules for task decomposition:
- Each task should be self-contained (one clear deliverable)
- Define dependencies between tasks (addBlockedBy)
- Assign file ownership boundaries per teammate role
- Target 5-6 tasks per teammate for optimal flow
- Tasks should map to SPEC requirements where applicable

### Phase 3: Teammate Spawning

Spawn teammates using Task tool with team_name parameter:

```
Task(
  subagent_type: "team-backend-dev",
  team_name: "moai-run-SPEC-XXX",
  name: "backend-dev",
  prompt: "You are the backend developer for this team. Your file ownership: {detected_ownership}. SPEC context: {spec_summary}",
  mode: "plan"
)
```

Spawning rules:
- Include SPEC context in the spawn prompt
- Assign file ownership boundaries detected from project structure (see File Ownership Detection)
- Use appropriate model per role (haiku for research, sonnet for implementation)

Plan approval (when workflow.yaml `team.require_plan_approval: true`):
- Spawn implementation teammates with `mode: "plan"`
- Teammates must submit a plan before writing any code
- Team lead receives plan_approval_request messages from teammates
- Team lead reviews plan scope, file ownership compliance, and approach
- Approve via: `SendMessage(type: "plan_approval_response", request_id: "{id}", recipient: "{name}", approve: true)`
- Reject with feedback via: `SendMessage(type: "plan_approval_response", request_id: "{id}", recipient: "{name}", approve: false, content: "Feedback here")`
- After approval, teammate exits plan mode and begins implementation
- When `require_plan_approval` is false, spawn with `mode: "acceptEdits"` instead

### Phase 4: Coordination

MoAI as Team Lead monitors and coordinates:

1. Receive automatic messages from teammates (progress, completion, issues)
2. Use SendMessage for direct coordination
3. Broadcast critical updates to all teammates
4. Resolve file ownership conflicts
5. Reassign tasks if a teammate is blocked

Coordination patterns:
- When backend completes API: notify frontend-dev of available endpoints
- When implementation completes: assign quality validation tasks
- When quality finds issues: direct fix messages to responsible teammate
- When all tasks complete: begin shutdown sequence

Delegate mode (when workflow.yaml `team.delegate_mode: true`):
- MoAI operates in coordination-only mode
- Focus on task assignment, message routing, progress monitoring, and conflict resolution
- Do NOT directly implement code or modify files (no Write, Edit, or Bash for implementation)
- Delegate ALL implementation work to teammates via task assignment and SendMessage
- Read and Grep are permitted for understanding context and reviewing teammate output
- If a task has no suitable teammate, spawn a new one rather than implementing directly
- When delegate_mode is false, team lead may implement small tasks directly alongside teammates

### Phase 5: Shutdown

Graceful shutdown sequence:

1. Verify all tasks are completed via TaskList
2. Send shutdown_request to each teammate:
   ```
   SendMessage(type: "shutdown_request", recipient: "backend-dev")
   ```
3. Wait for shutdown approval from each
4. Clean up team resources:
   ```
   TeamDelete()
   ```

## File Ownership Strategy

Prevent write conflicts by assigning exclusive file ownership.

[HARD] Team lead MUST analyze project structure before assigning ownership. Use Explore agent or Glob/Grep to map directory structure and assign ownership boundaries that match the actual project layout. Never use hardcoded patterns from a different project type.

### File Ownership Detection

Ownership patterns depend on the project type. Detect the project structure first, then assign accordingly:

**Go projects:**

| Role | Ownership |
|------|-----------|
| backend-dev | internal/**, pkg/**, cmd/** |
| tester | *_test.go, testdata/**, test/** |
| quality | (read-only, no file ownership) |

**Web projects (React, Vue, Angular):**

| Role | Ownership |
|------|-----------|
| backend-dev | src/api/**, src/models/**, src/services/** |
| frontend-dev | src/ui/**, src/components/**, src/pages/** |
| tester | tests/**, __tests__/**, *.test.*, *.spec.* |
| quality | (read-only, no file ownership) |

**Full-stack projects (separate client/server):**

| Role | Ownership |
|------|-----------|
| backend-dev | server/**, api/**, src/server/** |
| frontend-dev | client/**, app/**, src/client/** |
| data-layer | db/**, migrations/**, schema/** |
| tester | tests/**, __tests__/**, *_test.go, *.test.*, *.spec.* |
| quality | (read-only, no file ownership) |

**Monorepo projects:**

| Role | Ownership |
|------|-----------|
| Per-domain teammate | packages/<domain-name>/**, apps/<domain-name>/** |
| tester | **/tests/**, **/__tests__/**, **/*_test.go, **/*.test.*, **/*.spec.* |
| quality | (read-only, no file ownership) |

**Python projects:**

| Role | Ownership |
|------|-----------|
| backend-dev | src/<package>/**, <package>/** |
| tester | tests/**, **/test_*.py, **/*_test.py |
| quality | (read-only, no file ownership) |

### Ownership Rules

- No two teammates own the same file
- Shared types/interfaces: owned by the creating teammate, shared via message
- Config files: owned by team lead or explicitly assigned
- If ownership conflict: team lead resolves via SendMessage
- Test files always belong to the tester role regardless of location

## Team Patterns Reference

### Plan Research Team
- Roles: researcher (haiku), analyst (sonnet), architect (sonnet)
- Use: Complex SPEC creation requiring multi-angle exploration
- Duration: Short-lived (exploration phase only)

### Implementation Team
- Roles: backend-dev (sonnet), frontend-dev (sonnet), tester (sonnet)
- Use: Cross-layer feature implementation
- Duration: Medium (full run phase)

### Full-Stack Team
- Roles: api-layer, ui-layer, data-layer, quality (all sonnet)
- Use: Large-scale full-stack features
- Duration: Medium-long

### Investigation Team
- Roles: hypothesis-1, hypothesis-2, hypothesis-3 (all haiku)
- Use: Complex debugging with competing theories
- Duration: Short

### Review Team
- Roles: security-reviewer, perf-reviewer, quality-reviewer (all sonnet)
- Use: Multi-perspective code review
- Duration: Short

## Error Recovery

- Teammate crash: Spawn replacement with same role and resume context
- Task stuck: Team lead reassigns to different teammate
- File conflict: Team lead mediates via SendMessage, adjusts ownership
- All teammates idle: Check if tasks remain, assign or shutdown
- Token limit: Shutdown team gracefully, fall back to sub-agent for remaining work

---

Version: 1.1.0
Last Updated: 2026-02-07
