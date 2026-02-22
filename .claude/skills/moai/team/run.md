---
name: moai-workflow-team-run
description: >
  Implement SPEC requirements using team-based architecture.
  Supports CG Mode (Claude leader + GLM teammates via tmux) and
  Agent Teams Mode (all same API, parallel teammates).
  CG mode uses tmux pane-level env isolation for API separation.
  Agent Teams mode uses file ownership for parallel coordination.
user-invocable: false
metadata:
  version: "3.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-22"
  tags: "run, team, glm, tmux, implementation, parallel, agent-teams"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["team run", "glm worker", "parallel implementation"]
  agents: ["team-backend-dev", "team-frontend-dev", "team-tester"]
  phases: ["run"]
---
# Workflow: Team Run - Implementation with Agent Teams

Purpose: Implement SPEC requirements using team-based architecture with parallel
teammates. Supports CG Mode (Claude + GLM) and standard Agent Teams Mode.

Flow: Mode Detection -> Plan (Leader) -> Run (Agent Teams) -> Quality (Leader) -> Sync (Leader)

## Mode Selection

Before executing this workflow, check `.moai/config/sections/llm.yaml`:

| team_mode | Execution Mode | Description |
|-----------|---------------|-------------|
| (empty) | Sub-agent | Single session, Task() subagents |
| cg | CG Mode | Claude Leader + GLM Teammates via tmux |
| agent-teams | Agent Teams | All same API, parallel teammates |

- If `team_mode == "cg"`: Use CG Mode section below
- If `team_mode == "agent-teams"`: Use Agent Teams Mode section below
- If `team_mode == ""`: Fall back to sub-agent mode (workflows/run.md)

---

## CG Mode (Claude Leader + GLM Teammates)

### Overview

CG mode uses tmux pane-level environment isolation:
- **Leader (Claude)**: Runs in the original tmux pane with no GLM env vars
- **Teammates (GLM)**: Spawn in new tmux panes that inherit GLM env from tmux session

This is standard Agent Teams with `CLAUDE_CODE_TEAMMATE_DISPLAY=tmux`, where
the tmux session has GLM env vars injected by `moai cg`.

### Prerequisites

- `moai cg` has been run inside tmux (team_mode="cg" in llm.yaml)
- Claude Code started in the SAME pane where `moai cg` was run
- GLM API key saved via `moai glm <key>` or `GLM_API_KEY` env

### Phase 1: Plan (Leader on Claude)

The Leader creates the SPEC document using Claude's reasoning capabilities.

1. **Delegate to manager-spec subagent**:
   ```
   Task(
     subagent_type: "manager-spec",
     prompt: "Create SPEC document for: {user_description}
              Follow EARS format.
              Output to: .moai/specs/SPEC-XXX/spec.md"
   )
   ```

2. **User Approval** via AskUserQuestion:
   - Approve SPEC and proceed to implementation
   - Request modifications
   - Cancel workflow

3. **Output**: `.moai/specs/SPEC-XXX/spec.md`

### Phase 2: Run (Agent Teams — Teammates on GLM)

Teammates execute implementation in parallel using GLM via Z.AI API.

#### 2.1 Team Setup

1. Create team:
   ```
   TeamCreate(team_name: "moai-run-SPEC-XXX")
   ```

2. Create shared task list with dependencies:
   ```
   TaskCreate: "Implement data models and schema" (no deps)
   TaskCreate: "Implement API endpoints" (blocked by data models)
   TaskCreate: "Implement UI components" (blocked by API)
   TaskCreate: "Write unit and integration tests" (blocked by API + UI)
   TaskCreate: "Quality validation - TRUST 5" (blocked by all above)
   ```

#### 2.2 Spawn Teammates

Spawn teammates using Task() with team_name. Because `CLAUDE_CODE_TEAMMATE_DISPLAY=tmux`
is set, each teammate spawns in a new tmux pane. New panes inherit GLM env vars
from the tmux session, routing them through Z.AI API.

```
Task(
  subagent_type: "team-backend-dev",
  team_name: "moai-run-SPEC-XXX",
  name: "backend-dev",
  mode: "acceptEdits",
  prompt: "You are backend-dev on team moai-run-SPEC-XXX.
    Implement backend tasks from the shared task list.
    SPEC: .moai/specs/SPEC-XXX/spec.md
    Follow TDD methodology. Claim tasks via TaskUpdate.
    Mark tasks completed when done. Send results via SendMessage."
)

Task(
  subagent_type: "team-frontend-dev",
  team_name: "moai-run-SPEC-XXX",
  name: "frontend-dev",
  mode: "acceptEdits",
  prompt: "You are frontend-dev on team moai-run-SPEC-XXX.
    Implement frontend tasks from the shared task list.
    SPEC: .moai/specs/SPEC-XXX/spec.md
    Follow TDD methodology. Claim tasks via TaskUpdate.
    Mark tasks completed when done. Send results via SendMessage."
)

Task(
  subagent_type: "team-tester",
  team_name: "moai-run-SPEC-XXX",
  name: "tester",
  mode: "acceptEdits",
  prompt: "You are tester on team moai-run-SPEC-XXX.
    Write tests for implemented features.
    SPEC: .moai/specs/SPEC-XXX/spec.md
    Own all *_test.go files exclusively.
    Mark tasks completed when done. Send results via SendMessage."
)
```

All teammates spawn in parallel in separate tmux panes.

#### 2.3 Monitor and Coordinate

MoAI monitors teammate progress:

1. **Receive messages automatically** (no polling needed)
2. **Handle idle notifications**:
   - Check TaskList to verify work status
   - If complete: Send shutdown_request
   - If work remains: Send new instructions
   - NEVER ignore idle notifications
3. **Handle plan approval** (if require_plan_approval: true):
   - Respond with plan_approval_response immediately
4. **Forward information** between teammates as needed

#### 2.4 Teammate Completion

When teammates complete:
- All tasks marked completed in shared TaskList
- Tests passing within each teammate's scope
- Changes committed (teammates with `isolation: worktree` commit to their branches)

### Phase 3: Quality (Leader on Claude)

Leader validates quality using Claude's analysis:

1. Run quality gates:
   ```bash
   go test -race ./...
   golangci-lint run
   go test -cover ./...
   ```

2. SPEC verification:
   - Read SPEC acceptance criteria
   - Verify all requirements implemented
   - If gaps found: create follow-up tasks or assign to teammates

3. TRUST 5 validation via manager-quality subagent

### Phase 4: Sync and Cleanup (Leader on Claude)

#### 4.1 Documentation

```
Task(
  subagent_type: "manager-docs",
  prompt: "Generate documentation for SPEC-XXX implementation.
           Update CHANGELOG.md and README.md as needed."
)
```

#### 4.2 Team Shutdown

```
SendMessage(type: "shutdown_request", recipient: "backend-dev", content: "Phase complete")
SendMessage(type: "shutdown_request", recipient: "frontend-dev", content: "Phase complete")
SendMessage(type: "shutdown_request", recipient: "tester", content: "Phase complete")
```

Wait for shutdown_response from each, then:
```
TeamDelete
```

#### 4.3 Report Summary

Present completion report to user:
- SPEC ID and description
- Files modified
- Tests added/modified
- Coverage achieved
- Cost savings estimate (GLM vs Claude)

### CG Mode Error Recovery

| Failure | Recovery |
|---------|----------|
| Teammate spawn failure | Fall back to sub-agent mode |
| tmux pane crash | Check teammate status, respawn if needed |
| Quality gate failure | Leader creates fix task |
| Merge conflicts (worktree) | Leader resolves or user choice |

---

## Agent Teams Mode

When `team_mode == "agent-teams"` in llm.yaml, use parallel teammates all on the same API.

### Phase 1: Team Setup

1. Create team:
   ```
   TeamCreate(team_name: "moai-run-SPEC-XXX")
   ```

2. Create shared task list with dependencies:
   ```
   TaskCreate: "Implement data models and schema" (no deps)
   TaskCreate: "Implement API endpoints" (blocked by data models)
   TaskCreate: "Implement UI components" (blocked by API endpoints)
   TaskCreate: "Write unit and integration tests" (blocked by API + UI)
   TaskCreate: "Quality validation - TRUST 5" (blocked by all above)
   ```

### Phase 2: Spawn Implementation Team

Spawn teammates with file ownership boundaries:

```
Task(subagent_type: "team-backend-dev", team_name: "moai-run-SPEC-XXX", name: "backend-dev", mode: "acceptEdits", ...)
Task(subagent_type: "team-frontend-dev", team_name: "moai-run-SPEC-XXX", name: "frontend-dev", mode: "acceptEdits", ...)
Task(subagent_type: "team-tester", team_name: "moai-run-SPEC-XXX", name: "tester", mode: "acceptEdits", ...)
```

### Phase 3: Handle Idle Notifications

**CRITICAL**: When a teammate goes idle, you MUST respond immediately:

1. **Check TaskList** to verify work status
2. **If all tasks complete**: Send shutdown_request
3. **If work remains**: Send new instructions or wait

Example response to idle notification:
```
# Check tasks
TaskList()

# If work is done, shutdown
SendMessage(type: "shutdown_request", recipient: "backend-dev", content: "Implementation complete, shutting down")

# If work remains, send instructions
SendMessage(type: "message", recipient: "backend-dev", content: "Continue with next task: {instructions}")
```

**FAILURE TO RESPOND TO IDLE NOTIFICATIONS CAUSES INFINITE WAITING**

### Phase 4: Plan Approval (when require_plan_approval: true)

When teammates submit plans, you MUST respond immediately:

```
# Receive plan_approval_request with request_id

# Approve
SendMessage(type: "plan_approval_response", request_id: "{id}", recipient: "{name}", approve: true)

# Reject with feedback
SendMessage(type: "plan_approval_response", request_id: "{id}", recipient: "{name}", approve: false, content: "Revise X")
```

### Phase 5: Quality and Shutdown

1. Assign quality validation task to team-quality (or use manager-quality subagent)
2. After all tasks complete, shutdown teammates:
   ```
   SendMessage(type: "shutdown_request", recipient: "backend-dev", content: "Phase complete")
   SendMessage(type: "shutdown_request", recipient: "frontend-dev", content: "Phase complete")
   SendMessage(type: "shutdown_request", recipient: "tester", content: "Phase complete")
   ```
3. Wait for shutdown_response from each teammate
4. TeamDelete to clean up resources

---

## Comparison

| Aspect | CG Mode | Agent Teams Mode | Sub-agent Mode |
|--------|---------|------------------|----------------|
| APIs | Claude + GLM | Single (all same) | Single |
| Cost | Lowest | Highest | Medium |
| Parallelism | Parallel (tmux panes) | Parallel (in-process/tmux) | Sequential |
| Quality | Highest (Claude reviews) | High | High |
| Requires tmux | Yes | No (optional) | No |
| Isolation | tmux env + optional worktree | File ownership + optional worktree | None |

## Fallback

If team mode fails at any point:
1. Log error details
2. Clean up team (TeamDelete) if created
3. Fall back to sub-agent mode (workflows/run.md)
4. Continue from last successful phase

---

Version: 3.0.0 (tmux Agent Teams CG Mode)
Last Updated: 2026-02-22
