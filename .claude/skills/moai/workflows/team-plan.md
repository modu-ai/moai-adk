# Workflow: Team Plan - Agent Teams SPEC Creation

Purpose: Create comprehensive SPEC documents through parallel team-based research and analysis. Used when plan phase benefits from multi-angle exploration.

Flow: TeamCreate -> Parallel Research -> Synthesis -> SPEC Document -> Shutdown

## Prerequisites

- workflow.team.enabled: true
- CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1
- Triggered by: /moai plan --team OR auto-detected complexity >= threshold

## Phase 0: Team Setup

1. Read configuration:
   - .moai/config/sections/workflow.yaml for team settings
   - .moai/config/sections/quality.yaml for development mode

2. Create team:
   ```
   TeamCreate(team_name: "moai-plan-{feature-slug}")
   ```

3. Create shared task list:
   ```
   TaskCreate: "Explore codebase architecture and dependencies"
   TaskCreate: "Analyze requirements, user stories, and edge cases"
   TaskCreate: "Design technical approach and evaluate alternatives"
   TaskCreate: "Synthesize findings into SPEC document" (blocked by above 3)
   ```

## Phase 1: Spawn Research Team

Spawn 3 teammates from the plan_research pattern. All spawns MUST use Task() with `team_name` and `name` parameters. Launch all three in a single response for parallel execution:

```
Task(
  subagent_type: "team-researcher",
  team_name: "moai-plan-{feature-slug}",
  name: "researcher",
  mode: "plan",
  prompt: "You are a codebase researcher on team moai-plan-{feature-slug}.
    Explore the codebase for {feature_description}.
    Map architecture, find relevant files, identify dependencies and patterns.
    When done, mark your task as completed via TaskUpdate and send findings to the team lead via SendMessage."
)

Task(
  subagent_type: "team-analyst",
  team_name: "moai-plan-{feature-slug}",
  name: "analyst",
  mode: "plan",
  prompt: "You are a requirements analyst on team moai-plan-{feature-slug}.
    Analyze requirements for {feature_description}.
    Identify user stories, acceptance criteria, edge cases, risks, and constraints.
    When done, mark your task as completed via TaskUpdate and send findings to the team lead via SendMessage."
)

Task(
  subagent_type: "team-architect",
  team_name: "moai-plan-{feature-slug}",
  name: "architect",
  mode: "plan",
  prompt: "You are a technical architect on team moai-plan-{feature-slug}.
    Design the technical approach for {feature_description}.
    Evaluate implementation alternatives, assess trade-offs, propose architecture.
    Consider existing patterns found by the researcher.
    When done, mark your task as completed via TaskUpdate and send findings to the team lead via SendMessage."
)
```

All three teammates run in parallel. Messages from teammates are delivered automatically to MoAI.

## Phase 2: Parallel Research

Teammates work independently:
- researcher explores codebase (fastest, haiku)
- analyst defines requirements (medium)
- architect designs solution (waits for researcher findings)

MoAI monitors:
- Receive progress messages automatically
- Forward researcher findings to architect when available
- Resolve any questions from teammates

## Phase 3: Synthesis

After all research tasks complete:
1. Collect findings from all three teammates
2. Delegate SPEC creation to manager-spec subagent (NOT a teammate) with all findings
3. Include: codebase analysis, requirements, technical design, edge cases

SPEC output at: .moai/specs/SPEC-XXX/spec.md

## Phase 4: User Approval

AskUserQuestion with options:
- Approve SPEC and proceed to implementation
- Request modifications (specify which section)
- Cancel workflow

## Phase 5: Cleanup

1. Shutdown all teammates:
   ```
   SendMessage(type: "shutdown_request", recipient: "researcher", content: "Plan phase complete, shutting down")
   SendMessage(type: "shutdown_request", recipient: "analyst", content: "Plan phase complete, shutting down")
   SendMessage(type: "shutdown_request", recipient: "architect", content: "Plan phase complete, shutting down")
   ```
2. TeamDelete to clean up resources
3. Execute /clear to free context for next phase

## Fallback

If team creation fails or AGENT_TEAMS not enabled:
- Fall back to sub-agent plan workflow (workflows/plan.md)
- Log warning about team mode unavailability

---

Version: 1.1.0
