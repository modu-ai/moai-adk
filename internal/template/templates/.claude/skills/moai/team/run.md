---
name: moai-workflow-team-run
description: >
  Implement SPEC requirements using team-based architecture with dynamic
  teammate generation. Teammates are spawned as general-purpose agents
  with runtime parameter overrides from workflow.yaml role profiles.
  Supports CG Mode (Claude leader + GLM teammates via tmux) and
  Agent Teams Mode (all same API, parallel teammates).
user-invocable: false
metadata:
  version: "4.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-03-31"
  tags: "run, team, glm, tmux, implementation, parallel, agent-teams, dynamic"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["team run", "glm worker", "parallel implementation"]
  agents: []
  phases: ["run"]
---
# Workflow: Team Run - Dynamic Team Generation

Purpose: Implement SPEC requirements using dynamically generated teammates.
All teammates use `subagent_type: "general-purpose"` with runtime parameter
overrides from `workflow.yaml` role profiles. No static team agent definitions.

Flow: Mode Detection -> Plan (Leader) -> Run (Dynamic Teams) -> Quality (Leader) -> Sync (Leader)

## Architecture: Dynamic Team Generation

Teammates are spawned using the Agent tool with runtime overrides:

| Parameter | Source | Purpose |
|-----------|--------|---------|
| subagent_type | Always "general-purpose" | Full tool access |
| team_name | TeamCreate result | Team coordination |
| name | Pattern role name | Addressable via SendMessage |
| model | workflow.yaml role_profiles | Cost optimization |
| mode | workflow.yaml role_profiles | Permission control |
| isolation | workflow.yaml role_profiles | File safety |
| prompt | Orchestrator-generated | Role, context, instructions |

### Role Profile Reference

From `.moai/config/sections/workflow.yaml` → `team.role_profiles`:

| Profile | mode | model | isolation | Use For |
|---------|------|-------|-----------|---------|
| researcher | plan | haiku | none | Codebase exploration, read-only analysis |
| analyst | plan | sonnet | none | Requirements analysis, validation |
| architect | plan | sonnet | none | Solution design, architecture |
| implementer | acceptEdits | sonnet | worktree | Backend, frontend, full-stack code |
| tester | acceptEdits | sonnet | worktree | Test creation, coverage validation |
| designer | acceptEdits | sonnet | worktree | UI/UX with MCP design tools |
| reviewer | plan | haiku | none | Code review, quality validation |

## Mode Selection

This workflow is loaded ONLY when team mode has been explicitly selected (via `--team` flag or auto-selection). Check `.moai/config/sections/llm.yaml` to determine WHICH team mode to use:

| team_mode | Execution Mode | Description |
|-----------|---------------|-------------|
| (empty) or agent-teams | **Agent Teams** | All same API, parallel teammates (default for `--team` flag) |
| glm | GLM Mode | All GLM, credentials in settings.local.json |
| cg | CG Mode (tmux required) | Claude Leader + GLM Teammates via tmux session env |

[HARD] When this workflow is loaded, team mode is already decided. Empty `team_mode` defaults to Agent Teams, NOT sub-agent fallback. Sub-agent mode uses a different workflow (`workflows/run.md`).

---

## CG Mode (Claude Leader + GLM Teammates)

### Overview

CG mode uses tmux pane-level environment isolation:
- **Leader (Claude)**: Runs in the original tmux pane with no GLM env vars
- **Teammates (GLM)**: Spawn in new tmux panes that inherit GLM env from tmux session

This is standard Agent Teams with `CLAUDE_CODE_TEAMMATE_DISPLAY=tmux`, where
the tmux session has GLM env vars injected by `moai cg`.

### Env Isolation Mechanism (Verified)

`moai cg` executes two complementary steps:
1. `injectTmuxSessionEnv()` → `tmux set-environment` (session-scoped, no -g) injects GLM env vars
2. `removeGLMEnv()` → removes GLM env from `settings.local.json` so leader uses Claude API

When Claude Code Agent Teams spawns teammates via `tmux split-window`:
- New panes inherit tmux session env → teammates get GLM vars → Z.AI API
- Leader process already running → not affected by session env changes → Claude API
- Result: Leader on Claude, Teammates on GLM, within the same tmux session

### Prerequisites

- `moai cg` has been run inside tmux (team_mode="cg" in llm.yaml)
- Claude Code started in the SAME pane where `moai cg` was run
- GLM API key saved via `moai glm setup <key>`

### Phase 1: Plan (Leader on Claude)

The Leader creates the SPEC document using Claude's reasoning capabilities.

1. **Delegate to manager-spec subagent**:
   ```
   Agent(
     subagent_type: "manager-spec",
     prompt: "Create SPEC document for: {user_description}
              Follow EARS format.
              Output to: .moai/specs/SPEC-XXX/spec.md"
   )
   ```

2. **User Approval** via AskUserQuestion

3. **Output**: `.moai/specs/SPEC-XXX/spec.md`

### Phase 2: Run (Dynamic Teams — Teammates on GLM)

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

#### 2.2 Spawn Teammates (Dynamic Generation)

Spawn teammates using `Agent(subagent_type: "general-purpose")` with role profile overrides.

**Path Rules for Worktree Teammates:**
- All file references in teammate prompts MUST use project-root-relative paths
- Do NOT include absolute paths to the main project directory
- See `.claude/rules/moai/workflow/worktree-integration.md` Prompt Path Rules section

```
Agent(
  subagent_type: "general-purpose",
  team_name: "moai-run-SPEC-XXX",
  name: "backend-dev",
  model: "sonnet",
  mode: "acceptEdits",
  isolation: "worktree",
  prompt: "You are backend-dev on team moai-run-SPEC-XXX.

    SPEC: .moai/specs/SPEC-XXX/spec.md
    Project type: {detected_language} ({detected_framework})
    Methodology: {development_mode} (from quality.yaml)

    File ownership: server-side files (*.go excluding *_test.go), API handlers, models, database code.

    Quality requirements:
    - Run tests after each significant change
    - Run linter before marking tasks complete
    - Follow project conventions

    Claim tasks via TaskUpdate. Mark tasks completed when done.
    Send results to team lead via SendMessage.
    Report blockers immediately via SendMessage."
)

Agent(
  subagent_type: "general-purpose",
  team_name: "moai-run-SPEC-XXX",
  name: "frontend-dev",
  model: "sonnet",
  mode: "acceptEdits",
  isolation: "worktree",
  prompt: "You are frontend-dev on team moai-run-SPEC-XXX.

    SPEC: .moai/specs/SPEC-XXX/spec.md
    Project type: {detected_language} ({detected_framework})

    File ownership: client-side files (components, pages, styles, assets).

    Quality requirements:
    - Run tests after each significant change
    - Follow project conventions

    Claim tasks via TaskUpdate. Mark tasks completed when done.
    Send results to team lead via SendMessage."
)

Agent(
  subagent_type: "general-purpose",
  team_name: "moai-run-SPEC-XXX",
  name: "tester",
  model: "sonnet",
  mode: "acceptEdits",
  isolation: "worktree",
  prompt: "You are tester on team moai-run-SPEC-XXX.

    SPEC: .moai/specs/SPEC-XXX/spec.md
    Project type: {detected_language}

    File ownership: test files exclusively (*_test.go, *.test.*, __tests__/).
    Read implementation files but do NOT modify them.
    If implementation has bugs, report to relevant teammate via SendMessage.

    Quality standards:
    - Meet 85%+ overall coverage, 90%+ for new code
    - Tests should be specification-based, not implementation-coupled
    - Include edge cases, error scenarios, boundary conditions

    Claim tasks via TaskUpdate. Mark tasks completed when done.
    Send results to team lead via SendMessage."
)
```

All teammates spawn in parallel. Implementation teammates use `isolation: "worktree"` for file safety.

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

1. Run language-appropriate quality gates (auto-detected)
2. SPEC verification against acceptance criteria
3. TRUST 5 validation via manager-quality subagent

### Phase 4: Sync and Cleanup (Leader on Claude)

#### 4.1 Documentation

```
Agent(
  subagent_type: "manager-docs",
  prompt: "Generate documentation for SPEC-XXX implementation.
           Update CHANGELOG.md and README.md as needed."
)
```

#### 4.2 Team Shutdown

1. Shutdown all teammates:
   ```
   SendMessage(type: "shutdown_request", recipient: "backend-dev", content: "Phase complete")
   SendMessage(type: "shutdown_request", recipient: "frontend-dev", content: "Phase complete")
   SendMessage(type: "shutdown_request", recipient: "tester", content: "Phase complete")
   ```

2. Wait for shutdown_response from each teammate

3. Clean up GLM env (CG mode only):
   ```bash
   moai cc
   ```

4. TeamDelete to clean up team resources

---

## Agent Teams Mode

When `team_mode` is empty or `"agent-teams"` in llm.yaml, use parallel teammates all on the same API. This is the default team execution mode when `--team` flag is used.

### Phase 1: Team Setup

1. Create team:
   ```
   TeamCreate(team_name: "moai-run-SPEC-XXX")
   ```

2. Create shared task list with dependencies (same as CG mode)

### Phase 2: Spawn Implementation Team

Spawn teammates with role profile overrides and worktree isolation:

```
Agent(subagent_type: "general-purpose", team_name: "moai-run-SPEC-XXX", name: "backend-dev", model: "sonnet", mode: "acceptEdits", isolation: "worktree", prompt: "Backend role. File ownership: server-side code. ...")
Agent(subagent_type: "general-purpose", team_name: "moai-run-SPEC-XXX", name: "frontend-dev", model: "sonnet", mode: "acceptEdits", isolation: "worktree", prompt: "Frontend role. File ownership: client-side code. ...")
Agent(subagent_type: "general-purpose", team_name: "moai-run-SPEC-XXX", name: "tester", model: "sonnet", mode: "acceptEdits", isolation: "worktree", prompt: "Testing role. File ownership: test files exclusively. ...")
```

[HARD] All implementation teammates MUST use `isolation: "worktree"` for parallel file safety.

### Phase 3: Handle Idle Notifications

**CRITICAL**: When a teammate goes idle, you MUST respond immediately:

1. **Check TaskList** to verify work status
2. **If all tasks complete**: Send shutdown_request
3. **If work remains**: Send new instructions or wait

**FAILURE TO RESPOND TO IDLE NOTIFICATIONS CAUSES INFINITE WAITING**

### Phase 4: Plan Approval (when require_plan_approval: true)

When teammates submit plans, respond immediately with plan_approval_response.

### Phase 5: Quality and Shutdown

1. Quality validation via manager-quality subagent (or reviewer teammate)
2. Shutdown all teammates via SendMessage shutdown_request
3. Wait for shutdown_response from each
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
| Isolation | tmux env + worktree (HARD) | File ownership + worktree (HARD) | None |
| Agent definitions | None (dynamic) | None (dynamic) | Static (.claude/agents/) |

## Fallback

If team mode fails at any point:
1. Log error details
2. Clean up team (TeamDelete) if created
3. Fall back to sub-agent mode (workflows/run.md)
4. Continue from last successful phase

---

Version: 4.1.0 (Fix --team flag routing: empty team_mode defaults to Agent Teams)
Last Updated: 2026-04-09
Source: SPEC-TEAM-001
