# Workflow: Team Run - GLM Worker Mode

Purpose: Implement SPEC requirements using Leader (current Claude model) + Worker (GLM-5) architecture with worktree isolation.

Flow: Mode Detection -> Plan (Leader) -> Run (GLM-5 Worker) -> Merge (Leader) -> Sync (Leader)

## Mode Selection

Before executing this workflow, check `.moai/config/sections/llm.yaml`:

| team_mode | Execution Mode | Description |
|-----------|---------------|-------------|
| (empty) | Sub-agent | Single session, Task() subagents |
| glm | GLM Worker | Opus Leader + GLM-5 Worker in worktree |
| agent-teams | Agent Teams | Parallel teammates with file ownership |

- If `team_mode == "glm"`: Use this workflow (GLM Worker Mode)
- If `team_mode == "agent-teams"`: See Agent Teams Mode section below
- If `team_mode == ""`: Fall back to sub-agent mode (run.md)

## Overview

This workflow enables cost-effective development by combining:
- **Leader (current Claude model)**: SPEC creation, merge, review, documentation - uses your selected model (Opus, Sonnet, etc.)
- **Worker (GLM-5)**: Cost-effective implementation in isolated worktree

Cost savings: Run phase typically uses ~70% of total tokens. Using GLM-5 for this phase reduces overall cost by 60-70%.

## Prerequisites

- `moai glm --team` has been run (team_mode="glm" in llm.yaml)
- Claude Code session started with `claude` (runs on Opus)
- GLM API key saved via `moai glm <key>` or set in GLM_API_KEY env

## Phase 0: Mode Detection

Read `.moai/config/sections/llm.yaml` to determine execution mode:

| team_mode | Execution Mode | Description |
|-----------|---------------|-------------|
| (empty) | Sub-agent | Single session, Task() subagents |
| glm | GLM Worker | Opus Leader + GLM-5 Worker in worktree |

Detection steps:
```
1. Read .moai/config/sections/llm.yaml
2. If team_mode == "glm": proceed with GLM Worker mode (this workflow)
3. If team_mode == "": fall back to sub-agent mode (see run.md)
```

## Phase 1: Plan (Leader - Current Session)

The Leader (running on your selected Claude model) creates the SPEC document.

### Steps

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
   - Request modifications (specify which section)
   - Cancel workflow

3. **Output**: `.moai/specs/SPEC-XXX/spec.md`

### Why Leader for Plan

SPEC documents define architecture and requirements. Quality at this stage prevents costly rework. Your selected Claude model (Opus for complex projects, Sonnet for standard work) provides appropriate reasoning for design decisions.

## Phase 2: Run (GLM-5 Worker in Worktree)

The Worker implements the SPEC using GLM-5 in an isolated git worktree.

### 2.1 Prepare Worker Environment

1. **Read GLM configuration** from llm.yaml:
   - base_url: `https://api.z.ai/api/anthropic`
   - models: opus=glm-5, sonnet=glm-4.7, haiku=glm-4.7-flashx
   - env_var: `GLM_API_KEY`

2. **Read API key**:
   - From `~/.moai/.env.glm` (saved by `moai glm`)
   - Or from environment variable `GLM_API_KEY`

3. **Build worker prompt** from SPEC:
   ```
   You are implementing SPEC-XXX.

   SPEC Location: .moai/specs/SPEC-XXX/spec.md

   Requirements:
   {extract_requirements_from_spec}

   Instructions:
   1. Read the SPEC document carefully
   2. Follow TDD methodology: write tests first, then implement
   3. Implement all requirements listed in the SPEC
   4. Run tests to verify: go test -race ./...
   5. Run lint: golangci-lint run
   6. Ensure 85%+ code coverage
   7. Commit all changes with conventional commit format
   8. Reference SPEC-XXX in commit message

   When complete, all tests should pass and the implementation should match the SPEC requirements.
   ```

### 2.2 Spawn Worker via Bash

Execute worker using Bash tool with background execution:

```bash
# Worker command with GLM environment
env \
  ANTHROPIC_AUTH_TOKEN="$GLM_API_KEY" \
  ANTHROPIC_BASE_URL="https://api.z.ai/api/anthropic" \
  ANTHROPIC_DEFAULT_OPUS_MODEL="glm-5" \
  ANTHROPIC_DEFAULT_SONNET_MODEL="glm-4.7" \
  ANTHROPIC_DEFAULT_HAIKU_MODEL="glm-4.7-flashx" \
  API_TIMEOUT_MS="3000000" \
  claude -w worker-SPEC-XXX \
    -p "IMPLEMENTATION_PROMPT_HERE" \
    --output-format text \
    > .moai/team/worker-SPEC-XXX.log 2>&1 &

echo "Worker PID: $!"
```

**Key points**:
- `env VAR=val ...` sets environment ONLY for the claude process
- `claude -w worker-SPEC-XXX` creates worktree at `.claude/worktrees/worker-SPEC-XXX/`
- `-p "..."` runs headless with the given prompt
- Background `&` allows Leader to continue

### 2.3 Monitor Worker Progress

1. **Poll for completion**:
   ```bash
   # Check if worktree has new commits
   git log worktree-worker-SPEC-XXX --oneline -1

   # Or check if process is still running
   ps -p $WORKER_PID
   ```

2. **Read worker log**:
   ```bash
   tail -f .moai/team/worker-SPEC-XXX.log
   ```

3. **Timeout handling**:
   - Default timeout: 30 minutes
   - If exceeded, warn user and offer options:
     - Continue waiting
     - Check worker status
     - Abort and merge partial work

### 2.4 Worker Completion

When worker completes:
- All tests pass in worktree
- Changes committed to branch `worktree-worker-SPEC-XXX`
- Log file shows completion message

## Phase 3: Merge (Leader - Current Session)

The Leader integrates worker changes with quality validation.

### 3.1 Merge Worktree Branch

```bash
# Fetch worktree changes
git fetch origin worktree-worker-SPEC-XXX

# Merge with main
git merge worktree-worker-SPEC-XXX --no-ff -m "feat(scope): implement SPEC-XXX

Implemented by GLM-5 Worker in worktree.

SPEC: .moai/specs/SPEC-XXX/spec.md
"
```

### 3.2 Handle Conflicts

If merge conflicts occur:
1. Leader (current model) analyzes conflicts
2. Auto-resolves if straightforward
3. Otherwise, presents to user via AskUserQuestion with options:
   - Accept worker changes
   - Keep main changes
   - Manual resolution

### 3.3 Quality Validation

Run quality gates:

```bash
# Tests
go test -race ./...

# Lint
golangci-lint run

# Coverage
go test -cover ./...
```

**Quality gates must pass**:
- Zero test failures
- Zero lint errors
- 85%+ coverage maintained

### 3.4 SPEC Verification

Verify all SPEC requirements are implemented:
- Read SPEC acceptance criteria
- Check implementation matches requirements
- If gaps found, either:
  - Create follow-up tasks
  - Re-run worker with refined prompt

## Phase 4: Sync & Cleanup (Leader - Current Session)

### 4.1 Documentation

Delegate to manager-docs subagent:
```
Task(
  subagent_type: "manager-docs",
  prompt: "Generate documentation for SPEC-XXX implementation.
           Update CHANGELOG.md with feature description.
           Update README.md if API changes."
)
```

### 4.2 Worktree Cleanup

```bash
# Remove worktree (claude -w auto-cleans on exit, but verify)
git worktree remove worker-SPEC-XXX --force

# Delete branch
git branch -D worktree-worker-SPEC-XXX
```

### 4.3 Report Summary

Present completion report to user:
- SPEC ID and description
- Files modified
- Tests added/modified
- Coverage achieved
- Documentation updated

## Error Recovery

### Worker Fails

1. Check worker log for errors
2. Options:
   - Retry with refined prompt
   - Fall back to sub-agent mode
   - User manual intervention

### Merge Conflicts

1. Leader attempts auto-resolution
2. Complex conflicts: user choice
3. Worst case: abort merge, manual resolution

### Quality Gate Failures

1. Identify failing tests/lint errors
2. Options:
   - Create fix task for worker
   - Leader implements fixes directly
   - User manual intervention

## Multiple Workers (Advanced)

For large SPECs with independent components:

For large SPECs with independent components:

1. Split SPEC into sub-SPECs
2. Spawn multiple workers in parallel:
   - `worker-SPEC-XXX-api`
   - `worker-SPEC-XXX-ui`
   - `worker-SPEC-XXX-tests`
3. Each worker in separate worktree
4. Merge sequentially by dependency order

## Comparison with Agent Teams Mode

| Aspect | GLM Worker Mode | Agent Teams Mode |
|--------|----------------|------------------|
| APIs | Claude + GLM-5 | Single API (all same) |
| Isolation | Git worktree | File ownership convention |
| Coordination | Sequential (spawn -> wait -> merge) | Parallel (SendMessage) |
| Cost | Lower (GLM for implementation) | Higher (all same model) |
| Quality | Leader reviews all changes | All teammates share quality |

## Fallback

If GLM Worker mode fails at any point:
1. Log error details
2. Clean up worktree
3. Fall back to sub-agent mode (run.md)
4. Continue from last successful phase

---

## Agent Teams Mode

When `team_mode == "agent-teams"` in llm.yaml, use parallel teammates instead of GLM Worker.

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

Version: 2.1.0 (Added Agent Teams Mode)
Last Updated: 2026-02-20
