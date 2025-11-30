---
name: moai:2-run
description: "Execute TDD implementation cycle"
argument-hint: 'SPEC-ID - All with SPEC ID to implement (e.g. SPEC-001) or all "SPEC Implementation"'
allowed-tools: Read, Write, Edit, Grep, Glob, WebFetch, WebSearch, Bash, TodoWrite, AskUserQuestion, Task, Skill
model: inherit
---

##  Pre-execution Context

!git status --porcelain
!git branch --show-current
!git log --oneline -5
!git diff --name-only HEAD

##  Essential Files

@.moai/config/config.json
@.moai/specs/

---

#  MoAI-ADK Step 2: Execute Implementation (Run) - TDD Implementation

> Architecture: Commands → Agents → Skills. This command orchestrates ONLY through `Task()` tool.
>
> Delegation Model: Phase-based sequential agent delegation. Command orchestrates 4 phases directly.

Workflow: Phase 1 → Analysis & Planning → Phase 2 → TDD Implementation → Phase 3 → Git Operations → Phase 4 → Completion & Guidance.

---

##  Command Purpose

Execute TDD implementation of SPEC requirements through complete agent delegation.

The `/moai:2-run` command orchestrates the complete implementation workflow:

1. Phase 1: SPEC analysis and execution plan creation
2. Phase 2: TDD implementation (RED → GREEN → REFACTOR)
3. Phase 3: Git commit management
4. Phase 4: Completion and next steps guidance

Run on: `$ARGUMENTS` (SPEC ID, e.g., SPEC-001)

---

##  Execution Philosophy: "Plan → Run → Sync"

`/moai:2-run` performs SPEC implementation through phase-based sequential agent delegation:

```
User Command: /moai:2-run SPEC-001
    ↓
Phase 1: Task(subagent_type="manager-strategy")
    → SPEC Analysis & Execution Plan Creation
    ↓
Phase 2: Task(subagent_type="manager-tdd")
    → RED → GREEN → REFACTOR TDD Cycle
    ↓
Phase 2.5: Task(subagent_type="manager-quality")
    → TRUST 5 Quality Validation
    ↓
Phase 3: Task(subagent_type="manager-git")
    → Commit Creation & Git Operations
    ↓
Phase 4: AskUserQuestion(...)
    → Completion Summary & Next Steps Guidance
    ↓
Output: Implemented feature with passing tests and commits
```

### Key Principle: Zero Direct Tool Usage

This command uses ONLY these tools:

-  Task() for phase agent delegation (manager-strategy → manager-tdd → manager-quality → manager-git)
-  AskUserQuestion() for user approval and next steps
-  TodoWrite() for task tracking
-  No Read/Write/Edit/Bash (all delegated to agents)

Command orchestrates phases sequentially; agents handle complexity.

---

##  Associated Agents & Skills

| Agent/Skill                    | Purpose                                                |
| ------------------------------ | ------------------------------------------------------ |
| manager-strategy               | Analyzes SPEC and creates execution strategy           |
| manager-tdd                    | Implements code through TDD cycle (RED-GREEN-REFACTOR) |
| manager-quality                | Verifies TRUST 5 principles and validates quality      |
| manager-git                    | Creates and manages Git commits                        |
| moai-alfred-workflow           | Workflow orchestration patterns                        |
| moai-alfred-todowrite-pattern  | Task tracking and progress management                  |
| moai-alfred-ask-user-questions | User interaction patterns                              |
| moai-alfred-reporting          | Result reporting and summaries                         |

---

##  Phase Execution Details

### Phase 1: Analysis & Planning

Use the manager-strategy subagent to:

1. Read SPEC document
2. Analyze requirements and create execution strategy
3. Return plan for user approval
4. Wait for user confirmation (proceed/modify/postpone)
5. Store plan context for Phase 2

### Phase 2: TDD Implementation

Use the manager-tdd subagent to:

1. Initialize task tracking (TodoWrite)
2. Check domain readiness (if multi-domain SPEC)
3. Execute RED → GREEN → REFACTOR cycle
4. Return implementation results and coverage metrics

### Phase 2.5: Quality Validation

Use the manager-quality subagent to:

1. Verify TRUST 5 principles (Test-first, Readable, Unified, Secured, Trackable)
2. Validate test coverage (>= 85%)
3. Check security compliance
4. Return quality assessment (PASS/WARNING/CRITICAL)

### Phase 3: Git Operations

Use the manager-git subagent to:

1. Create feature branch if needed
2. Create commits with implementation changes
3. Verify commits were successful
4. Return commit summary

### Phase 4: Completion & Guidance

Command calls `AskUserQuestion()`:

1. Displays implementation summary
2. Shows next action options
3. Guides user to `/moai:3-sync` or additional features

---

##  Execution Flow (High-Level)

```
/moai:2-run SPEC-XXX
    ↓
Parse SPEC ID from $ARGUMENTS
    ↓
 Phase 1: manager-strategy subagent
    → Analyze SPEC → Create execution plan → Get approval
    ↓
 Phase 2: manager-tdd subagent
    → RED-GREEN-REFACTOR → Tests passing → Coverage verified
    ↓
 Phase 2.5: manager-quality subagent
    → Validate TRUST 5 principles → Return quality status
    ↓
 Phase 3: manager-git subagent
    → Create feature branch → Commit changes → Verify commits
    ↓
 Phase 4: AskUserQuestion(...)
    → Display summary → Guide next steps → Offer options
    ↓
Output: "Implementation complete. Next step: /moai:3-sync"
```

---

##  Command Implementation

### Sequential Phase Execution

Command implementation flow:

```
# Phase 1: SPEC Analysis & Planning
Use the manager-strategy subagent to:
- Analyze SPEC-$ARGUMENTS and create detailed execution plan
- Extract requirements and success criteria
- Identify implementation phases and tasks
- Determine tech stack and dependencies
- Estimate complexity and effort
- Present step-by-step execution strategy

# User approval checkpoint
AskUserQuestion({
    "question": "Does this execution plan look good?",
    "header": "Plan Review",
    "multiSelect": false,
    "options": [
        {"label": "Proceed with plan", "description": "Start implementation"},
        {"label": "Modify plan", "description": "Request changes"},
        {"label": "Postpone", "description": "Stop here, continue later"}
    ]
})

# Phase 2: TDD Implementation (if approved)
if approval == "Proceed with plan":
    Use the manager-tdd subagent to:
    - Execute complete TDD implementation for SPEC-$ARGUMENTS
    - Write failing tests (RED phase)
    - Implement minimal code (GREEN phase)
    - Refactor for quality (REFACTOR phase)
    - Ensure test coverage >= 85%
    - Verify all tests passing

# Phase 2.5: Quality Validation
Use the manager-quality subagent to:
    - Validate implementation against TRUST 5 principles:
      - T: Test-first (tests exist and pass)
      - R: Readable (code is clear and documented)
      - U: Unified (follows project patterns)
      - S: Secured (no security vulnerabilities)
      - T: Trackable (changes are logged and traceable)

Return quality assessment with specific findings.

# Phase 3: Git Operations (Context from previous phases)
if quality_result.status == "PASS" or quality_result.status == "WARNING":
    Use the manager-git subagent to:
    - Create commits for SPEC-$ARGUMENTS implementation
    - The complete context (planning, implementation, quality review) informs meaningful commit messages
    - Create feature branch: feature/SPEC-$ARGUMENTS
    - Stage all relevant files
    - Create meaningful commits (follow conventional commits)
    - Verify commits created successfully
    - Return commit summary with SHA references

# Phase 4: Completion & Guidance
next_steps = AskUserQuestion({
    "question": "Implementation complete. What would you like to do next?",
    "header": "Next Steps",
    "multiSelect": false,
    "options": [
        {"label": "Sync Documentation", "description": "/moai:3-sync"},
        {"label": "Implement Another Feature", "description": "/moai:1-plan"},
        {"label": "Review Results", "description": "Examine the implementation"},
        {"label": "Finish", "description": "Session complete"}
    ]
})
```

### Context Flow

Phase Progression:

```
Phase 1: SPEC Analysis
  → Planning context created

Phase 2: Implementation
  → Builds on planning context
  → Implements without re-reading SPEC

Phase 2.5: Quality Validation
  → Uses planning + implementation context
  → Validates with complete feature knowledge

Phase 3: Git Operations
  → Uses complete feature context
  → Creates commits with full understanding
```

Benefits:

-  Context Continuity: Full knowledge chain across all phases
-  Unified Coding: Phase 1 architectural decisions naturally propagate
-  Better Commits: manager-git understands full context for meaningful messages

---

##  Design Improvements (vs Previous Version)

| Metric                 | Before           | After          | Improvement            |
| ---------------------- | ---------------- | -------------- | ---------------------- |
| Command LOC        | ~420             | ~120           | 71% reduction      |
| allowed-tools      | 14 types         | 1 type         | 93% reduction      |
| Direct tool usage  | Yes (Read/Bash)  | No             | 100% eliminated    |
| Agent count        | 4 separate calls | 1 orchestrator | 100% simplified    |
| User approval flow | In command       | In agent       | Cleaner separation |
| Error handling     | Dispersed        | Centralized    | Better structure   |

---

##  Verification Checklist

After implementation, verify:

- [ ]  Command has ONLY `Task`, `AskUserQuestion`, `TodoWrite` in allowed-tools
- [ ]  Command contains NO `Read`, `Write`, `Edit`, `Bash` usage
- [ ]  Command delegates execution to phase agents sequentially
- [ ]  Phase 1: manager-strategy executes successfully
- [ ]  Phase 2: manager-tdd executes successfully
- [ ]  Phase 2.5: manager-quality validates TRUST 5
- [ ]  Phase 3: manager-git creates commits
- [ ]  Phase 4: User guided to next steps
- [ ]  User approval checkpoints working

---

##  Quick Reference

| Scenario                     | Entry Point                                | Key Phases                                                               | Expected Outcome                            |
| ---------------------------- | ------------------------------------------ | ------------------------------------------------------------------------ | ------------------------------------------- |
| Implement SPEC feature       | `/moai:2-run SPEC-XXX`                     | Phase 1 → Planning → Phase 2 → TDD → Phase 2.5 → Quality → Phase 3 → Git | Implemented feature with ≥85% test coverage |
| Resume failed implementation | `/moai:2-run SPEC-XXX` (retry)             | Resume from last successful phase                                        | Completed implementation                    |
| Implement with modifications | `/moai:2-run SPEC-XXX` (with plan changes) | Modify plan → Execute phases                                             | Modified implementation                     |

Associated Agents:

- `manager-strategy` - SPEC analysis and execution strategy
- `manager-tdd` - TDD implementation (RED-GREEN-REFACTOR)
- `manager-quality` - TRUST 5 validation
- `manager-git` - Git operations and commit management

Implementation Results:

- Code: Implemented feature files
- Tests: Test files with ≥85% coverage
- Commits: Git commits with proper messages
- Quality: PASS/WARNING/CRITICAL status

Version: 3.1.0 (Command-Level Phase Orchestration)
Updated: 2025-11-25
Pattern: Sequential Phase-Based Agent Delegation
Compliance: Claude Code Best Practices + Zero Direct Tool Usage
Architecture: Commands → Agents → Skills (Complete delegation)

---

## Final Step: Next Action Selection

After TDD implementation completes, use AskUserQuestion tool to guide user to next action:

```python
AskUserQuestion({
    "questions": [{
        "question": "Implementation is complete. What would you like to do next?",
        "header": "Next Steps",
        "multiSelect": false,
        "options": [
            {
                "label": "Sync Documentation",
                "description": "Execute /moai:3-sync to organize documentation and create PR"
            },
            {
                "label": "Additional Implementation",
                "description": "Implement more features"
            },
            {
                "label": "Quality Verification",
                "description": "Review tests and code quality"
            }
        ]
    }]
})
```

Important:

- Use conversation language from config
- No emojis in any AskUserQuestion fields
- Always provide clear next step options

##  EXECUTION DIRECTIVE

You must NOW execute the command following the "Execution Philosophy" described above.

1. Start Phase 1: Analysis & Planning immediately.
2. Use the manager-strategy subagent.
3. Do NOT just describe what you will do. DO IT.
