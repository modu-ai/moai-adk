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

Select team pattern based on SPEC scope:

For cross-layer features (implementation pattern):
- backend-dev (team-backend-dev, sonnet): Server-side implementation
- frontend-dev (team-frontend-dev, sonnet): Client-side implementation
- tester (team-tester, sonnet): Test creation and coverage

For cross-layer features with design (design_implementation pattern):
- designer (team-designer, sonnet): UI/UX design with Pencil/Figma MCP
- backend-dev (team-backend-dev, sonnet): Server-side implementation
- frontend-dev (team-frontend-dev, sonnet): Client-side implementation
- tester (team-tester, sonnet): Test creation and coverage

For full-stack features (full_stack pattern):
- api-layer (team-backend-dev, sonnet): API and business logic
- ui-layer (team-frontend-dev, sonnet): UI and components
- data-layer (team-backend-dev, sonnet): Database and schema
- quality (team-quality, sonnet): Quality validation

Spawn prompt must include:
- SPEC summary and their specific requirements
- File ownership boundaries (detected from project structure, see SKILL.md File Ownership Detection)
- Development methodology (TDD for new code, DDD for existing)
- Quality targets (coverage, lint, type checking)

### Plan Approval Mode

When workflow.yaml `team.require_plan_approval: true`:
- Spawn implementation teammates with `mode: "plan"` instead of `mode: "acceptEdits"`
- Each teammate must submit a plan before implementing any code
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

1. Shutdown all teammates gracefully
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
