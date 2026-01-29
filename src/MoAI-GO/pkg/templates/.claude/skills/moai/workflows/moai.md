# Workflow: MoAI - Autonomous Development Orchestration

Purpose: Full autonomous workflow. User provides a goal, MoAI autonomously executes plan -> run -> sync pipeline. This is the default workflow when no subcommand is specified.

Flow: Explore -> Plan -> Run -> Sync -> Done

## Supported Flags

- --loop: Enable auto iterative fixing during run phase
- --max N: Maximum iteration count for loop (default 100)
- --branch: Auto-create feature branch
- --pr: Auto-create pull request after completion
- --resume SPEC-XXX: Resume previous work from existing SPEC

## Configuration Files

- ralph.yaml: Loop settings and iteration defaults
- git-strategy.yaml: Branch and PR automation settings
- quality.yaml: TRUST 5 quality thresholds
- llm.yaml: LLM mode routing (claude-only, hybrid, glm-only)

## Phase 0: Parallel Exploration

Launch three agents simultaneously in a single response for 2-3x speedup (15-30s vs 45-90s).

Agent 1 - Explore (subagent_type Explore):
- Codebase analysis for task context
- Relevant files, architecture patterns, existing implementations

Agent 2 - Research (subagent_type Explore with WebSearch/WebFetch focus):
- External documentation and best practices
- API docs, library documentation, similar implementations

Agent 3 - Quality (subagent_type manager-quality):
- Current project quality assessment
- Test coverage status, lint status, technical debt

After all agents complete:
- Collect outputs from each agent response
- Extract key findings from Explore (files, patterns), Research (external knowledge), Quality (coverage baseline)
- Synthesize into unified exploration report
- Generate execution plan with files to create/modify and test strategy

Error handling: If any agent fails, continue with results from successful agents. Note missing information in plan.

If --sequential flag: Run Explore, then Research, then Quality sequentially instead.

## Phase 0 Completion: Routing Decision

Single-domain routing:
- If task is single-domain (e.g., "SQL optimization"): Delegate directly to expert agent, skip SPEC generation
- If task is multi-domain: Proceed to full workflow with SPEC generation

User approval checkpoint via AskUserQuestion:
- Options: Proceed to SPEC creation, Modify approach, Cancel

## Phase 1: SPEC Generation

- Delegate to manager-spec subagent
- Output: EARS-format SPEC document at .moai/specs/SPEC-XXX/spec.md
- Includes requirements, acceptance criteria, technical approach

## Phase 2: DDD Implementation Loop

[HARD] Agent delegation mandate: ALL implementation tasks MUST be delegated to specialized agents. NEVER execute implementation directly, even after auto compact.

Expert agent selection:
- Backend logic: expert-backend subagent
- Frontend components: expert-frontend subagent
- Test creation: expert-testing subagent
- Bug fixing: expert-debug subagent
- Refactoring: expert-refactoring subagent
- Security fixes: expert-security subagent

Loop behavior (when --loop or ralph.yaml loop.enabled is true):
- While issues exist AND iteration less than max:
  - Execute diagnostics (parallel by default)
  - Delegate fix to appropriate expert agent
  - Verify fix results
  - Check for completion marker
  - If marker found: Break loop

## Phase 3: Documentation Sync

- Delegate to manager-docs subagent
- Synchronize documentation with implementation
- Add completion marker on success

## Task Tracking

[HARD] Task management tools mandatory for all task tracking:
- When issues discovered: TaskCreate with pending status
- Before starting work: TaskUpdate with in_progress status
- After completing work: TaskUpdate with completed status
- Never output TODO lists as text

## Completion Markers

AI must add a marker when work is complete:
- `<moai>DONE</moai>` - Task complete
- `<moai>COMPLETE</moai>` - Full completion
- `<moai:done />` - XML format

## LLM Mode Routing

Auto-routing based on llm.yaml settings:

- claude-only: Plan phase uses Claude, Run phase uses Claude
- hybrid: Plan phase uses Claude, Run phase uses GLM (worktree)
- glm-only: Plan phase uses GLM (worktree), Run phase uses GLM (worktree)

## Execution Summary

1. Parse arguments (extract flags: --loop, --max, --sequential, --branch, --pr, --resume)
2. If --resume with SPEC ID: Load existing SPEC and continue from last state
3. Detect LLM mode from llm.yaml
4. Execute Phase 0 (parallel or sequential exploration)
5. Routing decision (single-domain direct delegation vs full workflow)
6. TaskCreate for discovered tasks
7. User confirmation via AskUserQuestion
8. Phase 1: SPEC generation via manager-spec
9. Phase 2: DDD implementation loop via expert agents
10. Phase 3: Documentation sync via manager-docs
11. Terminate with completion marker

---

Version: 1.1.0
Source: Renamed from alfred.md. Unified plan->run->sync pipeline.
