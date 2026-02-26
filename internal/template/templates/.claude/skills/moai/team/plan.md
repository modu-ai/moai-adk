---
name: moai-workflow-team-plan
description: >
  Create SPEC documents through parallel team-based research and analysis.
  Spawns researcher, analyst, and architect teammates for multi-angle exploration.
  Synthesizes findings into comprehensive SPEC document via manager-spec.
  Use when plan phase benefits from parallel multi-perspective exploration.
user-invocable: false
metadata:
  version: "2.7.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-23"
  tags: "plan, team, research, spec, parallel"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 3000

# MoAI Extension: Triggers
triggers:
  keywords: ["team plan", "parallel research", "team spec"]
  agents: ["team-researcher", "team-analyst", "team-architect"]
  phases: ["plan"]
---
# Workflow: Team Plan - Agent Teams SPEC Creation

Purpose: Create comprehensive SPEC documents through parallel team-based research and analysis. Used when plan phase benefits from multi-angle exploration.

Flow: TeamCreate -> Parallel Research -> Annotation Cycle -> SPEC Document -> Shutdown

## Prerequisites

See @.claude/rules/moai/workflow/spec-workflow.md for team mode prerequisites.

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
    Read target code areas IN DEPTH — understand deeply how each module works, its intricacies and side effects.
    Study cross-module interactions IN GREAT DETAIL — trace data flow through the system.
    Go through related test files to understand expected behavior and edge cases.
    Search for REFERENCE IMPLEMENTATIONS — find similar patterns in the codebase that can guide the new feature.
    Document all findings with specific file paths and line references.
    DO NOT write implementation code — focus exclusively on research and analysis.
    When done, write your findings to .moai/specs/SPEC-{ID}/research.md, mark your task as completed via TaskUpdate, and send findings to the team lead via SendMessage."
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
    Consider existing patterns found by the researcher — build on reference implementations rather than designing from scratch.
    When a concrete reference implementation exists in the codebase, use it as the foundation for your design.
    DO NOT write implementation code — focus exclusively on architectural design and planning.
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

**Handling Idle Notifications:**

When a teammate goes idle, you MUST respond:

1. Check TaskList to verify work status:
   - If task is still pending/in_progress, the teammate may be waiting for input
   - If task is completed, proceed to step 2

2. If all assigned work is complete, send new work or shutdown:
   ```
   SendMessage(type: "message", recipient: "{name}", content: "New task: {instructions}")
   # OR
   SendMessage(type: "shutdown_request", recipient: "{name}", content: "Work complete, shutting down")
   ```

3. NEVER ignore idle notifications - failure to respond causes infinite waiting

## Phase 2.5: Annotation Cycle

After research completes and before SPEC synthesis:

1. MoAI presents consolidated research findings (research.md) and draft plan to user
2. User reviews and adds inline annotations/corrections
3. MoAI delegates annotation processing to manager-spec subagent with guard: "Address all inline notes. DO NOT implement any code."
4. Repeat 1-6 times until user approves

This iterative refinement catches architectural misunderstandings before SPEC creation.

## Phase 3: Synthesis

After all research tasks complete and annotation cycle is approved:
1. Collect findings from all three teammates
2. Include research.md artifact as the foundation for SPEC creation
3. Delegate SPEC creation to manager-spec subagent (NOT a teammate) with all findings
4. Include: codebase analysis (research.md), requirements, technical design, edge cases, reference implementations

SPEC output at: .moai/specs/SPEC-XXX/spec.md

## Phase 4: User Approval

AskUserQuestion with options:
- Approve SPEC and proceed to implementation
- Request modifications (specify which section)
- Cancel workflow

## Phase 5: Cleanup (Timeout-Based)

1. Verify all tasks are completed via TaskList
2. Shutdown all teammates in parallel:
   ```
   SendMessage(type: "shutdown_request", recipient: "researcher", content: "Plan phase complete, shutting down")
   SendMessage(type: "shutdown_request", recipient: "analyst", content: "Plan phase complete, shutting down")
   SendMessage(type: "shutdown_request", recipient: "architect", content: "Plan phase complete, shutting down")
   ```
3. Wait maximum 30 seconds for shutdown_responses
4. Clean up GLM env vars and restore Claude-only operation:
   ```bash
   moai cc
   ```
   This safely removes GLM env vars while preserving ANTHROPIC_AUTH_TOKEN and other settings.
   Do NOT manually Read/Write settings.local.json — use the CLI command which handles JSON merging correctly.
5. TeamDelete to clean up team resources
6. Log any unresponsive teammates for debugging
7. Do NOT wait indefinitely for shutdown_response
8. Execute /clear to free context for next phase

**Timeout Rule**: If a teammate does not respond to shutdown_request within 30 seconds, proceed without their confirmation. This prevents the common issue of teammates hanging during cleanup.

## Fallback

If team creation fails or AGENT_TEAMS not enabled:
- Fall back to sub-agent plan workflow (workflows/plan.md)
- Log warning about team mode unavailability

---

Version: 2.7.0 (Research Phase + Annotation Cycle)
