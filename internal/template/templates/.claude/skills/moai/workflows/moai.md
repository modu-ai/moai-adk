---
name: moai-workflow-moai
description: >
  Full autonomous plan-run-sync pipeline. Default workflow when no subcommand
  is specified. Handles parallel exploration, SPEC generation, DDD/TDD
  implementation with optional auto-fix loop, and documentation sync.
user-invocable: false
metadata:
  version: "2.6.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-23"
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

For phase overview, token budgets, and phase transitions, see: .claude/rules/moai/workflow/spec-workflow.md

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
  development_mode: tdd    # or ddd
```

**Routing Logic**:

| Feature Type | Mode: ddd | Mode: tdd |
|--------------|-----------|-----------|
| **New package/module** (no existing file) | DDD* | TDD |
| **New feature in existing file** | DDD | TDD |
| **Refactoring existing code** | DDD | TDD (with brownfield pre-RED analysis) |
| **Bug fix in existing code** | DDD | TDD |

*DDD adapts for greenfield (ANALYZE requirements -> PRESERVE with spec tests -> IMPROVE)

**Agent Selection**:
- **TDD cycle**: `manager-tdd` subagent (RED-GREEN-REFACTOR)
- **DDD cycle**: `manager-ddd` subagent (ANALYZE-PRESERVE-IMPROVE)

For methodology details, see: .claude/rules/moai/workflow/workflow-modes.md

## Phase 0: Parallel Exploration

Launch three agents simultaneously in a single response for 2-3x speedup (15-30s vs 45-90s).

Agent 1 - Explore (subagent_type Explore, produces research.md):
- If .moai/project/codemaps/ exists: Use as architecture baseline to accelerate exploration (skip redundant scanning)
- Read target code areas IN DEPTH — understand deeply how each module works, its intricacies and side effects
- Study cross-module interactions IN GREAT DETAIL — trace data flow, identify implicit contracts
- Search for REFERENCE IMPLEMENTATIONS — find similar patterns in the codebase that can guide the new feature
- Document findings with specific file paths and line references
- Output: research.md artifact with architecture analysis, reference implementations, risks, and constraints

Agent 2 - Research (subagent_type Explore with WebSearch/WebFetch focus):
- External documentation and best practices
- API docs, library documentation, similar implementations
- Reference implementations from open-source projects that align with project conventions
- Documented design patterns relevant to the feature being implemented

Agent 3 - Quality (subagent_type manager-quality):
- Current project quality assessment
- Test coverage status, lint status, technical debt

After all agents complete:
- Collect outputs from each agent response
- Extract key findings from Explore (research.md with files, patterns, reference implementations), Research (external knowledge, documented patterns), Quality (coverage baseline)
- Synthesize into unified exploration report including research.md artifact
- Save research.md to .moai/specs/SPEC-{ID}/research.md when SPEC ID is determined
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

## Phase 1.5: Plan Annotation Cycle (1-6 iterations)

After SPEC generation and before implementation:
1. Present SPEC document and research.md to user for review
2. User adds inline annotations/corrections to plan
3. MoAI delegates to manager-spec: "Address all inline notes. DO NOT implement any code."
4. Repeat until user approves (maximum 6 iterations)
5. Track iteration count: "Annotation cycle {N}/6"

This iterative refinement catches architectural misunderstandings before implementation begins.

## Phase 2: Implementation (TDD or DDD based on development_mode)

[HARD] Agent delegation mandate: ALL implementation tasks MUST be delegated to specialized agents. NEVER execute implementation directly, even after auto compact.

[HARD] Methodology selection based on `.moai/config/sections/quality.yaml`:

- **development_mode: tdd** (default): Use `manager-tdd` (RED-GREEN-REFACTOR)
- **development_mode: ddd**: Use `manager-ddd` (ANALYZE-PRESERVE-IMPROVE)

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
- Plan phase: See team/plan.md
- Run phase: See team/run.md
- Sync rationale: See team/sync.md

Mode selection:
- --team: Force team mode for all applicable phases
- --solo: Force sub-agent mode
- No flag (default): System auto-selects based on complexity thresholds (domains >= 3, files >= 10, or score >= 7)

## Execution Summary

1. Parse arguments (extract flags: --loop, --max, --sequential, --branch, --pr, --resume, --team, --solo)
2. If --resume with SPEC ID: Load existing SPEC and continue from last state
3. Detect development_mode from quality.yaml (ddd/tdd)
4. **Team mode decision**: Read workflow.yaml team settings and determine execution mode
   - If `--team` flag: Force team mode (requires workflow.team.enabled: true AND CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 env var)
   - If `--solo` flag: Force sub-agent mode (skip team mode entirely)
   - If no flag (default): Check complexity thresholds from workflow.yaml auto_selection (domains >= 3, files >= 10, or score >= 7)
   - If team mode selected but prerequisites not met: Warn user and fallback to sub-agent mode
5. Execute Phase 0 (parallel or sequential exploration)
6. Routing decision (single-domain direct delegation vs full workflow)
7. TaskCreate for discovered tasks
8. User confirmation via AskUserQuestion
9. **Phase 0.5 (Research)**: Save research.md from Phase 0 Explore findings to SPEC directory
10. **Phase 1 (Plan)**: If team mode -> Read team/plan.md and follow team orchestration. Else -> manager-spec sub-agent
11. **Phase 1.5 (Annotate)**: Run annotation cycle (1-6 iterations) until user approves plan
11.5. **Execution Mode Selection Gate**: After Phase 1.5 approval, before Phase 2
   - Read .moai/config/sections/llm.yaml → team_mode ("" = cc, "glm" = glm, "cg" = cg)
   - Bash: test -n "$TMUX" && echo "tmux" || echo "no-tmux"
   - AskUserQuestion: worktree+{mode} (Recommended if tmux available) | team | sub-agent
   - Worktree selected: Launch new tmux session in worktree dir, terminate current pipeline
   - Team/Sub-agent selected: Pass execution_mode + active_mode to Phase 2
   - See plan.md Decision Point 3.5 for full option details
12. **Phase 2 (Run)**: Route based on Gate result (execution_mode parameter)
   - worktree: Already running in isolated tmux+worktree session (Gate handled transition)
   - team: Read team/run.md and follow team orchestration
   - sub-agent: manager-tdd or manager-ddd (per quality.yaml development_mode)
13. **Phase 3 (Sync)**: Always manager-docs sub-agent (sync phase never uses team mode)
14. Terminate with completion marker

---

Version: 2.6.0
Source: SPEC-MOAI-001. Integrated research pattern with deep codebase analysis, reference implementations, and annotation cycle for plan refinement.
