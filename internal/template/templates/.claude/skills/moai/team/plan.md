---
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
  agents: ["general-purpose"]
  phases: ["plan"]
---
# Workflow: Team Plan - Agent Teams SPEC Creation

Purpose: Create comprehensive SPEC documents through parallel team-based research and analysis. Used when plan phase benefits from multi-angle exploration.

Flow: Spawn research teammates (implicit team) -> Parallel Research -> Annotation Cycle -> SPEC Document -> Shutdown

## Prerequisites

See .claude/rules/moai/workflow/spec-workflow.md for team mode prerequisites.

## Phase 0: Team Setup

1. Read configuration:
   - .moai/config/sections/workflow.yaml for team settings
   - .moai/config/sections/quality.yaml for development mode

2. The team forms implicitly on the first teammate spawn (one team per session, no setup step). Teams/tasks are stored under the session-derived name `session-<first8>`.

3. Create shared task list:
   ```
   TaskCreate: "Explore codebase architecture and dependencies"
   TaskCreate: "Analyze requirements, user stories, and edge cases"
   TaskCreate: "Design technical approach and evaluate alternatives"
   TaskCreate: "Synthesize findings into SPEC document" (blocked by above 3)
   ```

## Phase 1: Spawn Research Team

Spawn 3 teammates using `subagent_type: "general-purpose"` with the read-only role_profiles (`researcher` / `analyst` / `architect`) applied via the `name` + `model` + `mode: "plan"` overrides. All spawns MUST use Agent() with the `name` parameter — the team forms implicitly on the first spawn (no setup step; the `team_name` parameter is accepted but ignored as of Claude Code v2.1.178). Launch all three in a single response for parallel execution:

```
Agent(
  subagent_type: "general-purpose",
  name: "researcher",
  model: "haiku",
  mode: "plan",
  prompt: "You are a codebase researcher on team moai-plan-{feature-slug}.
    Explore the codebase for {feature_description}.
    Read target code areas IN DEPTH — understand deeply how each module works, its intricacies and side effects.
    Study cross-module interactions IN GREAT DETAIL — trace data flow through the system.
    Go through related test files to understand expected behavior and edge cases.
    Search for REFERENCE IMPLEMENTATIONS — find similar patterns in the codebase that can guide the new feature.
    Document all findings with specific file paths and line references.
    DO NOT write implementation code — focus exclusively on research and analysis.
    When done, mark your task as completed via TaskUpdate and send findings to the team lead via SendMessage."
)

Agent(
  subagent_type: "general-purpose",
  name: "analyst",
  model: "sonnet",
  mode: "plan",
  prompt: "You are a requirements analyst on team moai-plan-{feature-slug}.
    Analyze requirements for {feature_description}.
    Identify user stories, acceptance criteria, edge cases, risks, and constraints.
    Define acceptance criteria using EARS format.
    Identify risks, constraints, dependencies, and non-functional requirements.
    When done, mark your task as completed via TaskUpdate and send findings to the team lead via SendMessage."
)

Agent(
  subagent_type: "general-purpose",
  name: "architect",
  model: "opus",
  mode: "plan",
  prompt: "You are a technical architect on team moai-plan-{feature-slug}.
    Design the technical approach for {feature_description}.
    Evaluate implementation alternatives, assess trade-offs, propose architecture.
    Consider existing patterns found by the researcher — build on reference implementations rather than designing from scratch.
    When a concrete reference implementation exists in the codebase, use it as the foundation for your design.
    Define file impact analysis, interface contracts, and implementation order.
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

#### [HARD] Pre-Write Frontmatter Checklist

[HARD] Before manager-spec calls Write/MultiEdit for spec.md, it MUST validate the frontmatter contains ALL 12 required fields AND rejects snake_case legacy aliases. This checklist prevents dual-schema drift between the plan workflow and the SPEC frontmatter lint rule. SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md`.

Required 12 fields (canonical order):
- [ ] `id: SPEC-{DOMAIN}-{NUM}` — matches `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$`
- [ ] `title: "Human-readable title"` — quoted string
- [ ] `version: "X.Y.Z"` — quoted semver string (NOT `0.1` unquoted)
- [ ] `status: draft` — enum: draft | planned | in-progress | implemented | completed | superseded | archived | rejected
- [ ] `created: YYYY-MM-DD` — ISO date (NEVER `created_at`, NEVER `date`)
- [ ] `updated: YYYY-MM-DD` — ISO date (NEVER `updated_at`)
- [ ] `author: <name>` — string, not empty
- [ ] `priority: P1` — uppercase Pn style (P0|P1|P2|P3) or High|Medium|Low|Critical
- [ ] `phase: "vX.Y.Z target"` — release phase string
- [ ] `module: "path/to/module"` — affected module path
- [ ] `lifecycle: spec-anchored` — enum: spec-anchored | spec-lite | exploratory
- [ ] `tags: "tag1, tag2, ..."` — comma-separated string (NOT `labels:`, NOT YAML array)

Optional fields (do NOT include unless needed):
- `issue_number: null` — integer or null (omit entirely when not tracking GitHub issue)

Rejected legacy aliases (fail closed — do NOT accept):
- `created_at:` (use `created:`)
- `updated_at:` (use `updated:`)
- `labels:` (use `tags:`)
- `spec_id:` (use `id:`)

Pre-write gate behavior:
1. manager-spec generates frontmatter draft in memory.
2. manager-spec self-audits against the 12-field checklist above.
3. If any required field is missing OR any rejected alias appears: manager-spec HALTS, reports the schema violation, and re-generates. It does NOT call Write.
4. Phase 2.3 plan-auditor independently re-verifies the schema on the written file as a second line of defense.

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
5. Team cleanup is automatic on session exit (no explicit teardown call — the TeamDelete tool was removed in Claude Code v2.1.178)
6. Log any unresponsive teammates for debugging
7. Do NOT wait indefinitely for shutdown_response
8. Execute /clear to free context for next phase

**Timeout Rule**: If a teammate does not respond to shutdown_request within 30 seconds, proceed without their confirmation. This prevents the common issue of teammates hanging during cleanup.

## Fallback

If team creation fails or AGENT_TEAMS not enabled:
- Fall back to sub-agent plan workflow (workflows/plan.md)
- Log warning about team mode unavailability

---

Version: 3.0.0 (Dynamic general-purpose read-only role_profiles + Annotation Cycle)
