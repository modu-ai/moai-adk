---
description: "Agentic AI automation - From SPEC to code with autonomous loop"
argument-hint: '"task description" [--loop] [--max N] [--seq] | resume SPEC-XXX'
type: utility
allowed-tools: Task, AskUserQuestion, TodoWrite, Bash, Read, Write, Edit, Glob, Grep
model: inherit
---

## Pre-execution Context

!git status --porcelain
!git branch --show-current

## Essential Files

@.moai/config/sections/ralph.yaml
@.moai/config/sections/git-strategy.yaml
@.moai/config/sections/quality.yaml
@.moai/config/sections/llm.yaml

---

# /moai:alfred - Agentic AI Autonomous Automation

## Core Principle: Fully Autonomous Automation

User provides a goal, AI autonomously plans, executes, and completes.

```
USER: "Add authentication feature"
  ↓
AI: Explore → Plan → Implement → Verify → Repeat
  ↓
AI: All issues resolved
  ↓
AI: <moai>DONE</moai>
```

## Command Purpose

Autonomously execute the full MoAI workflow:

1. **Parallel Exploration** (Explore + Research simultaneously)
2. **SPEC Generation** (after user approval)
3. **TDD Implementation** (auto iterative fixing)
4. **Documentation Sync**
5. **Completion Marker Detection** (`<moai>DONE</moai>`)

Feature Description: $ARGUMENTS

## Quick Start

```bash
# Default autonomous execution (parallel exploration)
/moai:alfred "Add JWT authentication"

# Enable auto loop (max 50 iterations)
/moai:alfred "JWT auth" --loop --max 50

# Sequential exploration + auto loop
/moai:alfred "JWT auth" --loop --sequential

# Resume previous work
/moai:alfred resume SPEC-AUTH-001
```

## Command Options

| Option | Alias | Description | Default |
|--------|-------|-------------|---------|
| `--loop` | - | Enable auto iterative fixing | ralph.yaml |
| `--max N` | --max-iterations | Maximum iteration count | 100 |
| `--sequential` | --seq | Sequential exploration (for debugging) | Parallel |
| `--branch` | - | Auto-create feature branch | git-strategy |
| `--pr` | - | Auto-create PR | git-strategy |
| `--resume SPEC` | - | Resume previous work | - |

## Completion Promise

AI must add a marker when work is complete:

```markdown
## Complete

All implementation complete, tests passing, docs updated. <moai>DONE</moai>
```

**Marker Types**:
- `<moai>DONE</moai>` - Task complete
- `<moai>COMPLETE</moai>` - Full completion
- `<moai:done />` - XML format

## Agentic Autonomous Flow

```
START: /moai:alfred "task description"

PHASE 0: Parallel Exploration (autonomous)
  ┌── Explore Agent: Codebase analysis
  ├── Research Agent: Documentation/issue search
  └── Quality Agent: Current state diagnosis
  ↓
Integration → Execution plan generation

PHASE 1: SPEC Generation
  └── Written in EARS format
  ↓
User Approval
  ↓
PHASE 2: TDD Implementation (autonomous loop)
  │
  └── WHILE (issues_exist AND iteration < max):
       ├── Diagnostics (LSP + Tests + Coverage)
       ├── TODO generation
       ├── Fix execution
       ├── Verification
       └── Completion marker detected? → BREAK
  ↓
PHASE 3: Documentation Sync
  └── Add marker: <moai>DONE</moai>
  ↓
COMPLETE
```

### Phase 0: Parallel Exploration Implementation

By default, execute three exploration agents simultaneously for optimal performance:

Step 1 - Launch Parallel Agent Tasks:

In a single response, invoke all three Task tools simultaneously (Claude Code executes them in parallel automatically):

1. Explore Agent:
   - subagent_type: Explore
   - description: Codebase analysis for task context
   - prompt: Include user task description, request relevant files, architecture patterns, existing implementations

2. Research Agent:
   - subagent_type: Explore (with WebSearch/WebFetch focus)
   - description: External documentation and best practices research
   - prompt: Include user task description, request API docs, library documentation, similar implementations

3. Quality Agent:
   - subagent_type: manager-quality
   - description: Current project quality assessment
   - prompt: Request test coverage status, lint status, technical debt assessment

All three Task calls in the same response triggers automatic parallel execution (up to 10 concurrent).

Step 2 - Collect and Integrate Results:

After all three agents complete:

1. Collect outputs from each agent response
2. Extract key findings:
   - From Explore: Relevant files, architecture patterns, existing code to reference
   - From Research: External knowledge, best practices, API patterns
   - From Quality: Current test coverage, code quality baseline, known issues

3. Synthesize into unified exploration report

Step 3 - Generate Execution Plan:

Based on integrated findings:

1. Identify implementation approach
2. List files to create/modify
3. Define test strategy
4. Estimate scope and complexity

Step 4 - User Approval Checkpoint:

Present integrated findings and proposed plan to user via AskUserQuestion:
- Options: Proceed to SPEC creation, Modify approach, Cancel

Error Handling:

If any agent fails:
- Continue with results from successful agents
- Note missing information in plan
- Offer to retry failed agent or proceed with partial information

WHY: Parallel exploration reduces Phase 0 time from 45-90 seconds to 15-30 seconds (2-3x speedup)
IMPACT: Faster initial context gathering without sacrificing comprehensiveness

## TODO-Obsessive Rule

[HARD] TodoWrite Tool Mandatory Usage:

1. Immediate Creation: When issues are discovered, call TodoWrite tool to add items with pending status
2. Immediate Progress: Before starting work, call TodoWrite tool to change item to in_progress
3. Immediate Completion: After completing work, call TodoWrite tool to change item to completed
4. Prohibited: Output TODO lists as text (MUST use TodoWrite tool)
5. Completion Condition: All TODOs completed OR completion marker detected

WHY: Using TodoWrite tool allows users to track progress in real-time.

## Output Format

### Running

```markdown
## Alfred: Phase 2 (Loop 3/100)

### TODO Status
- [x] Implement JWT token generation
- [x] Implement login endpoint
- [ ] Token validation middleware ← in progress

### Issues
- ERROR: src/auth.py:45 - undefined 'jwt_decode'
- WARNING: tests/test_auth.py:12 - unused 'result'

Fixing...
```

### Complete

```markdown
## Alfred: COMPLETE

### Summary
- SPEC: SPEC-AUTH-001
- Files: 8 files modified
- Tests: 25/25 passing
- Coverage: 88%
- Loops: 7 iterations

### Changes
+ JWT token generation
+ Login endpoint
+ Token validation middleware
+ Unit tests (12 cases)
+ API documentation

<moai>DONE</moai>
```

## LLM Mode

Auto-routing based on `llm.yaml` settings:

| Mode | Plan Phase | Run Phase |
|------|------------|-----------|
| opus-only | Claude (current) | Claude (current) |
| hybrid | Claude (current) | GLM (worktree) |
| glm-only | GLM (worktree) | GLM (worktree) |

## Expert Delegation (Single Domain)

Single domain tasks are delegated directly to expert agents:

```bash
# Alfred auto-determines
/moai:alfred "SQL query optimization"

# → Delegates directly to expert-performance agent
# → Immediate implementation without SPEC
```

## Quick Reference

```bash
# Autonomous execution (default parallel)
/moai:alfred "task"

# Auto loop + sequential
/moai:alfred "task" --loop --sequential

# Specify max iterations
/moai:alfred "task" --loop --max 50

# Branch + PR
/moai:alfred "task" --branch --pr

# Resume
/moai:alfred resume SPEC-XXX
```

---

## EXECUTION DIRECTIVE

1. Parse $ARGUMENTS (extract --loop, --max, --sequential, --branch, --pr, --resume flags)

2. IF --resume flag with SPEC ID: Load existing SPEC and continue from last state

3. Detect LLM mode from llm.yaml (opus-only, hybrid, glm-only)

4. Execute Phase 0 - Parallel Exploration:

   IF --sequential flag is specified:

   4a. Run Explore, then Research, then Quality sequentially

   ELSE (default parallel mode):

   4b. In a single response, invoke three Task tools simultaneously:
       - Task 1 (Explore): Codebase analysis with subagent_type="Explore"
       - Task 2 (Research): Documentation research with subagent_type="Explore" and WebSearch focus
       - Task 3 (Quality): Quality assessment with subagent_type="manager-quality"

   4c. Collect and integrate results from all three agents

   4d. Generate unified exploration report and execution plan

5. Routing decision:
   - IF single-domain task (e.g., "SQL optimization"): Delegate directly to expert agent, skip SPEC
   - IF multi-domain task: Proceed to full workflow with SPEC generation

6. [HARD] Call TodoWrite tool to add discovered tasks with pending status

7. User confirmation via AskUserQuestion (Proceed, Modify, Cancel)

8. Execute Phase 1 - SPEC Generation:
   - Use manager-spec subagent to create EARS-format SPEC document

9. Execute Phase 2 - TDD Implementation Loop:

   IF --loop flag OR ralph.yaml loop.enabled is true:

   9a. WHILE (issues exist AND iteration less than max):
       - [HARD] Before each task, call TodoWrite to change item to in_progress
       - Execute diagnostics (parallel if enabled)
       - Fix discovered issues
       - [HARD] After each fix, call TodoWrite to change item to completed
       - Check for completion marker
       - IF marker found: Break loop

10. Execute Phase 3 - Documentation Sync:
    - Use manager-docs subagent to synchronize documentation

11. Terminate with completion marker: Add marker when all tasks complete successfully

---

Version: 3.1.0
Last Updated: 2026-01-11
Core: Agentic AI Autonomous Automation
