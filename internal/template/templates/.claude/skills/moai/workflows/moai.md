---
name: moai-workflow-moai
description: >
  Full autonomous plan-run-sync pipeline. Default workflow when no subcommand
  is specified. Handles parallel exploration, SPEC generation, DDD/TDD
  implementation with optional auto-fix loop, and documentation sync.
license: Apache-2.0
compatibility: Designed for Claude Code
user-invocable: false
metadata:
  version: "2.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-07"
  tags: "moai, autonomous, pipeline, plan-run-sync, default"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["moai", "autonomous", "pipeline", "build", "implement", "create"]
  agents: ["moai"]
  phases: ["plan", "run", "sync"]
---

# Workflow: MoAI - Autonomous Development Orchestration

Purpose: Full autonomous workflow. User provides a goal, MoAI autonomously executes plan -> run -> sync pipeline. This is the default workflow when no subcommand is specified.

Flow: Explore -> Plan -> Run -> Sync -> Done

## Supported Flags

- --loop: Enable auto iterative fixing during run phase
- --max N: Maximum iteration count for loop (default 100)
- --branch: Auto-create feature branch
- --pr: Auto-create pull request after completion
- --resume SPEC-XXX: Resume previous work from existing SPEC
- --team: Force Agent Teams mode for plan and run phases
- --solo: Force sub-agent mode (single agent per phase)

**Default Behavior (no flag)**: System auto-selects based on complexity:
- Team mode: Multi-domain tasks (>=3 domains), many files (>=10), or high complexity (>=7)
- Sub-agent mode: Focused, single-domain tasks

## Configuration Files

- quality.yaml: TRUST 5 quality thresholds AND development_mode routing
- workflow.yaml: Execution mode, team settings, loop prevention, completion markers

## Development Mode Routing (CRITICAL)

[HARD] Before Phase 2 implementation, ALWAYS check `.moai/config/sections/quality.yaml`:

```yaml
constitution:
  development_mode: hybrid    # ddd, tdd, or hybrid
  hybrid_settings:
    new_features: tdd        # New code uses TDD
    legacy_refactoring: ddd  # Existing code uses DDD
```

**Routing Logic**:

| Feature Type | Mode: ddd | Mode: tdd | Mode: hybrid |
|--------------|-----------|-----------|--------------|
| **New package/module** (no existing file) | DDD* | TDD | TDD |
| **New feature in existing file** | DDD | TDD | TDD |
| **Refactoring existing code** | DDD | Use DDD for this part | DDD |
| **Bug fix in existing code** | DDD | TDD | DDD |

*DDD adapts for greenfield (ANALYZE requirements → PRESERVE with spec tests → IMPROVE)

**Agent Selection**:
- **TDD cycle**: `manager-tdd` subagent (RED-GREEN-REFACTOR)
- **DDD cycle**: `manager-ddd` subagent (ANALYZE-PRESERVE-IMPROVE)

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

## Phase 2: Implementation (TDD or DDD based on development_mode)

[HARD] Agent delegation mandate: ALL implementation tasks MUST be delegated to specialized agents. NEVER execute implementation directly, even after auto compact.

[HARD] Methodology selection based on `.moai/config/sections/quality.yaml`:

- **New features** (per hybrid_settings.new_features): Use `manager-tdd` (RED-GREEN-REFACTOR)
- **Legacy refactoring** (per hybrid_settings.legacy_refactoring): Use `manager-ddd` (ANALYZE-PRESERVE-IMPROVE)

Expert agent selection (for domain-specific work):
- Backend logic: expert-backend subagent
- Frontend components: expert-frontend subagent
- Test creation: expert-testing subagent
- Bug fixing: expert-debug subagent
- Refactoring: expert-refactoring subagent
- Security fixes: expert-security subagent

Loop behavior (when --loop flag or workflow.yaml loop_prevention settings enabled):
- While issues exist AND iteration less than max:
  - Execute diagnostics (parallel by default)
  - Delegate fix to appropriate expert agent
  - Verify fix results
  - Check for completion marker
  - If marker found: Break loop

## Phase 3: Documentation Sync

- Delegate to manager-docs subagent
- Synchronize documentation with implementation
- Detect SPEC-implementation divergence and update SPEC documents accordingly
- Conditionally update project documents (.moai/project/) when structural changes detected
- Respect SPEC lifecycle level for update strategy (spec-first, spec-anchored, spec-as-source)
- Add completion marker on success

## Team Mode

When --team flag is provided or auto-selected (based on complexity thresholds in workflow.yaml):

- Phase 0 exploration: Parallel research team (researcher + analyst + architect)
- Phase 2 implementation: Parallel implementation team (backend-dev + frontend-dev + tester)
- Phase 3 sync: Always sub-agent mode (manager-docs)

For team orchestration details:
- Plan phase: See workflows/team-plan.md
- Run phase: See workflows/team-run.md
- Sync rationale: See workflows/team-sync.md

Mode selection:
- --team: Force team mode for all applicable phases
- --solo: Force sub-agent mode
- No flag (default): System auto-selects based on complexity thresholds (domains >= 3, files >= 10, or score >= 7)

## Task Tracking

[HARD] Task management tools mandatory for all task tracking:
- When issues discovered: TaskCreate with pending status
- Before starting work: TaskUpdate with in_progress status
- After completing work: TaskUpdate with completed status
- Never output TODO lists as text

## Safe Development Protocol

All implementation follows CLAUDE.md Section 7 Safe Development Protocol:
- Approach-first: Explain approach before writing code
- Multi-file decomposition: Split work when modifying 3+ files
- Post-implementation review: List potential issues and suggest tests
- Reproduction-first: Write failing test before fixing bugs

## Completion Markers

AI must add a marker when work is complete:
- `<moai>DONE</moai>` - Task complete
- `<moai>COMPLETE</moai>` - Full completion

## Execution Summary

1. Parse arguments (extract flags: --loop, --max, --sequential, --branch, --pr, --resume, --team, --solo)
2. If --resume with SPEC ID: Load existing SPEC and continue from last state
3. Detect development_mode from quality.yaml (hybrid/ddd/tdd)
4. **Team mode decision**: Read workflow.yaml team settings and determine execution mode
   - If `--team` flag: Force team mode (requires workflow.team.enabled: true AND CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 env var)
   - If `--solo` flag: Force sub-agent mode (skip team mode entirely)
   - If no flag (default): Check complexity thresholds from workflow.yaml auto_selection (domains >= 3, files >= 10, or score >= 7)
   - If team mode selected but prerequisites not met: Warn user and fallback to sub-agent mode
5. Execute Phase 0 (parallel or sequential exploration)
6. Routing decision (single-domain direct delegation vs full workflow)
7. TaskCreate for discovered tasks
8. User confirmation via AskUserQuestion
9. **Phase 1 (Plan)**: If team mode → Read workflows/team-plan.md and follow team orchestration. Else → manager-spec sub-agent
10. **Phase 2 (Run)**: If team mode → Read workflows/team-run.md and follow team orchestration. Else → manager-tdd (new features) OR manager-ddd (legacy refactoring) sub-agent
11. **Phase 3 (Sync)**: Always manager-docs sub-agent (sync phase never uses team mode)
12. Terminate with completion marker

---

Version: 2.0.0
Source: Renamed from alfred.md. Unified plan->run->sync pipeline. Added SPEC/project document update in sync phase.
