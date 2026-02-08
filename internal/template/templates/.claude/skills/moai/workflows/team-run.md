# Workflow: Team Run - Agent Teams Implementation

Purpose: Implement SPEC requirements through parallel team-based development. Each teammate owns specific files/domains to prevent conflicts.

Flow: TeamCreate -> Task Decomposition -> Parallel Implementation -> Quality Validation -> Shutdown

## Prerequisites

- Approved SPEC document at .moai/specs/SPEC-XXX/
- workflow.team.enabled: true
- CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1
- Triggered by: /moai run SPEC-XXX --team OR auto-detected complexity >= threshold

## Phase 0: SPEC Analysis and Task Decomposition

1. Read SPEC document and analyze scope
2. Read quality.yaml for development_mode:
   - hybrid (default for new projects): New code uses TDD, existing code uses DDD
   - ddd (for existing projects): All code uses ANALYZE-PRESERVE-IMPROVE
   - tdd (explicit selection): All code uses RED-GREEN-REFACTOR

3. Decompose SPEC into implementation tasks:
   - Identify domain boundaries (backend, frontend, data, tests)
   - Assign file ownership per domain
   - Create tasks with clear dependencies
   - Target 5-6 tasks per teammate

4. Create team:
   ```
   TeamCreate(team_name: "moai-run-SPEC-XXX")
   ```

5. Create shared task list with dependencies:
   ```
   TaskCreate: "Implement data models and schema" (no deps)
   TaskCreate: "Implement API endpoints" (blocked by data models)
   TaskCreate: "Implement UI components" (blocked by API endpoints)
   TaskCreate: "Write unit and integration tests" (blocked by API + UI)
   TaskCreate: "Quality validation - TRUST 5" (blocked by all above)
   ```

## Phase 1: Spawn Implementation Team

Select team pattern based on SPEC scope and spawn ALL teammates using Task() with `team_name` and `name` parameters. Launch all teammates in a single response for parallel execution.

### Pattern: implementation (cross-layer features)

```
Task(
  subagent_type: "team-backend-dev",
  team_name: "moai-run-SPEC-XXX",
  name: "backend-dev",
  mode: "acceptEdits",
  prompt: "You are the backend developer on team moai-run-SPEC-XXX.
    SPEC Summary: {spec_summary}
    Your requirements: {backend_requirements}
    File ownership: {backend_file_patterns}
    Methodology: TDD for new code, DDD for existing code.
    Quality targets: 85%+ coverage, zero lint errors.
    Coordinate with frontend-dev for API contracts via SendMessage.
    Mark tasks completed via TaskUpdate when done."
)

Task(
  subagent_type: "team-frontend-dev",
  team_name: "moai-run-SPEC-XXX",
  name: "frontend-dev",
  mode: "acceptEdits",
  prompt: "You are the frontend developer on team moai-run-SPEC-XXX.
    SPEC Summary: {spec_summary}
    Your requirements: {frontend_requirements}
    File ownership: {frontend_file_patterns}
    Methodology: TDD for new components, DDD for existing.
    Wait for API contracts from backend-dev before implementing data fetching.
    Mark tasks completed via TaskUpdate when done."
)

Task(
  subagent_type: "team-tester",
  team_name: "moai-run-SPEC-XXX",
  name: "tester",
  mode: "acceptEdits",
  prompt: "You are the testing specialist on team moai-run-SPEC-XXX.
    SPEC Summary: {spec_summary}
    File ownership: all test files (*_test.go, *.test.*, __tests__/)
    Wait for implementation tasks to complete before writing integration tests.
    Coverage targets: 85%+ overall, 90%+ new code.
    Mark tasks completed via TaskUpdate when done."
)
```

### Pattern: design_implementation (with UI/UX design)

Add designer teammate before the implementation pattern teammates:

```
Task(
  subagent_type: "team-designer",
  team_name: "moai-run-SPEC-XXX",
  name: "designer",
  mode: "acceptEdits",
  prompt: "You are the UI/UX designer on team moai-run-SPEC-XXX.
    SPEC Summary: {spec_summary}
    File ownership: *.pen, design tokens, style configs.
    Create designs first, then share specs with frontend-dev via SendMessage.
    Mark tasks completed via TaskUpdate when done."
)
```

### Pattern: full_stack (with quality gate)

Use team-backend-dev for both api-layer and data-layer (unique names):

```
Task(subagent_type: "team-backend-dev", team_name: "moai-run-SPEC-XXX", name: "api-layer", mode: "acceptEdits", ...)
Task(subagent_type: "team-frontend-dev", team_name: "moai-run-SPEC-XXX", name: "ui-layer", mode: "acceptEdits", ...)
Task(subagent_type: "team-backend-dev", team_name: "moai-run-SPEC-XXX", name: "data-layer", mode: "acceptEdits", ...)
Task(subagent_type: "team-quality", team_name: "moai-run-SPEC-XXX", name: "quality", mode: "plan", ...)
```

### Spawn Prompt Requirements

Every spawn prompt MUST include:
- SPEC summary and teammate-specific requirements
- File ownership boundaries (detected from project structure)
- Development methodology (TDD for new code, DDD for existing)
- Quality targets (coverage, lint, type checking)
- Instructions to use TaskUpdate and SendMessage for coordination

### Plan Approval Mode

When workflow.yaml `team.require_plan_approval: true`:
- Spawn implementation teammates with `mode: "plan"` instead of `mode: "acceptEdits"`
- Each teammate must submit a plan via ExitPlanMode before implementing any code
- Team lead receives `plan_approval_request` messages with the proposed approach
- Team lead reviews: file ownership compliance, approach alignment with SPEC, scope correctness
- Approve: `SendMessage(type: "plan_approval_response", request_id: "{id}", recipient: "{name}", approve: true)`
- Reject with feedback: `SendMessage(type: "plan_approval_response", request_id: "{id}", recipient: "{name}", approve: false, content: "Revise X")`
- After approval, teammate automatically exits plan mode and begins implementation
- When `require_plan_approval` is false, spawn directly with `mode: "acceptEdits"`

## Phase 2: Parallel Implementation

Teammates self-claim tasks from the shared list and work independently:

Design (when team-designer is included):
- Creates UI/UX designs using Pencil MCP or Figma MCP (Task 0)
- Produces design tokens, style guides, and component specifications
- Shares design specs with frontend-dev via SendMessage
- Owns design files (.pen, design tokens, style configs)

Backend development:
- Creates data models and schema (Task 1)
- Implements API endpoints and business logic (Task 2)
- Follows TDD for new code: write test -> implement -> refactor
- Follows DDD for existing code: analyze -> preserve with tests -> improve
- Notifies frontend-dev when API contracts are ready

Frontend development:
- Waits for API contracts from backend-dev and design specs from designer
- Implements UI components and pages (Task 3)
- Follows TDD for new components
- Coordinates with backend on data shapes and designer on visual specs via SendMessage

Testing:
- Waits for implementation tasks to complete
- Writes integration tests spanning API and UI (Task 4)
- Validates coverage targets
- Reports test failures to responsible teammates

MoAI coordination:
- Forward API contract info from backend to frontend
- Resolve any blocking issues
- Monitor task progress via TaskList
- Redirect teammates if approach isn't working

### Delegate Mode

When workflow.yaml `team.delegate_mode: true`:
- MoAI operates in coordination-only mode during the entire run phase
- Focus on: task assignment, message routing, progress monitoring, conflict resolution
- Do NOT directly implement code or modify files (no Write, Edit, or Bash for implementation)
- Delegate ALL implementation to teammates via task assignment and SendMessage
- Read and Grep are permitted for context understanding and reviewing teammate output
- If a task has no suitable teammate, spawn a new one rather than implementing directly
- When delegate_mode is false, team lead may implement small tasks directly alongside teammates

## Phase 3: Quality Validation

After implementation and testing tasks complete:

Option A (with quality teammate):
- Assign Task 5 to quality teammate
- Quality runs TRUST 5 validation
- Reports findings to team lead
- Team lead directs fixes to responsible teammates

Option B (with sub-agent, for smaller teams):
- Delegate quality validation to manager-quality subagent
- Review findings and create fix tasks if needed
- Assign fixes to existing teammates

Quality gates (must all pass):
- Zero lint errors
- Zero type errors
- Coverage targets met (85%+ overall, 90%+ new code)
- No critical security issues
- All SPEC acceptance criteria verified

## Phase 4: Git Operations

After quality validation passes:
- Delegate to manager-git subagent (NOT a teammate)
- Create meaningful commit with conventional commit format
- Reference SPEC ID in commit message

## Phase 5: Cleanup

1. Shutdown all teammates gracefully (send to each active teammate):
   ```
   SendMessage(type: "shutdown_request", recipient: "backend-dev", content: "Implementation complete, shutting down")
   SendMessage(type: "shutdown_request", recipient: "frontend-dev", content: "Implementation complete, shutting down")
   SendMessage(type: "shutdown_request", recipient: "tester", content: "Implementation complete, shutting down")
   ```
   Wait for each teammate to respond with shutdown_response before proceeding.
2. TeamDelete to clean up resources
3. Report implementation summary to user

## Fallback

If team mode fails at any point:
- Shutdown remaining teammates gracefully
- Fall back to sub-agent run workflow (workflows/run.md)
- Continue from the last completed task
- Log warning about team mode failure

## Task Tracking

[HARD] All task status changes via TaskUpdate:
- pending -> in_progress: When teammate claims task
- in_progress -> completed: When task work is verified
- Never use plain text TODO lists

---

Version: 1.2.0
