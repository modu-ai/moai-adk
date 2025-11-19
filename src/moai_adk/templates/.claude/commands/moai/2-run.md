---
name: moai:2-run
description: "Execute TDD implementation cycle"
argument-hint: 'SPEC-ID - All with SPEC ID to implement (e.g. SPEC-001) or all "SPEC Implementation"'
allowed-tools:
  - Task
  - AskUserQuestion
---

# ⚒️ MoAI-ADK Step 2: Execute Implementation (Run) - TDD Implementation

> **Architecture**: Commands → Agents → Skills. This command orchestrates ONLY through `Task()` tool.
>
> **Delegation Model**: Phase-based sequential agent delegation. Command orchestrates 4 phases directly.

**Workflow**: Phase 1 → Analysis & Planning → Phase 2 → TDD Implementation → Phase 3 → Git Operations → Phase 4 → Completion & Guidance.

---

## 🎯 Command Purpose

Execute TDD implementation of SPEC requirements through complete agent delegation.

The `/moai:2-run` command orchestrates the complete implementation workflow:

1. **Phase 1**: SPEC analysis and execution plan creation
2. **Phase 2**: TDD implementation (RED → GREEN → REFACTOR)
3. **Phase 3**: Git commit management
4. **Phase 4**: Completion and next steps guidance

**Run on**: `$ARGUMENTS` (SPEC ID, e.g., SPEC-001)

---

## 💡 Execution Philosophy: "Plan → Run → Sync"

`/moai:2-run` performs SPEC implementation through phase-based sequential agent delegation:

```text
User Command: /moai:2-run SPEC-001
    ↓
Phase 1: Task(subagent_type="implementation-planner")
    → SPEC Analysis & Execution Plan Creation
    ↓
Phase 2: Task(subagent_type="tdd-implementer")
    → RED → GREEN → REFACTOR TDD Cycle
    ↓
Phase 2.5: Task(subagent_type="quality-gate")
    → TRUST 5 Quality Validation
    ↓
Phase 3: Task(subagent_type="git-manager")
    → Commit Creation & Git Operations
    ↓
Phase 4: AskUserQuestion(...)
    → Completion Summary & Next Steps Guidance
    ↓
Output: Implemented feature with passing tests and commits
```

### Key Principle: Zero Direct Tool Usage

**This command uses ONLY these tools:**

- ✅ **Task()** for phase agent delegation (implementation-planner → tdd-implementer → quality-gate → git-manager)
- ✅ **AskUserQuestion()** for user approval and next steps
- ✅ **TodoWrite()** for task tracking (delegated to agent)
- ❌ No Read/Write/Edit/Bash (all delegated to agents)

Command orchestrates phases sequentially; agents handle complexity.

---

## 🧠 Phase Agents & Skills

### Phase Agents (Sequential Execution)

| Agent                      | Purpose                                      | When                            |
| -------------------------- | -------------------------------------------- | ------------------------------- |
| **implementation-planner** | Analyzes SPEC and creates execution strategy | Phase 1                         |
| **tdd-implementer**        | Implements code through TDD cycle            | Phase 2                         |
| **quality-gate**           | Verifies TRUST 5 principles                  | Phase 2 (after tdd-implementer) |
| **git-manager**            | Creates and manages Git commits              | Phase 3                         |

### Skills Used (by agents, not command)

- `Skill("moai-core-workflow")` - Workflow orchestration
- `Skill("moai-core-todowrite-pattern")` - Task tracking
- `Skill("moai-core-ask-user-questions")` - User interaction
- `Skill("moai-core-reporting")` - Result reporting
- Domain-specific skills (selected per language/framework)

---

## 🚀 Phase Execution Details

### Phase 1: Analysis & Planning

Command calls `Task(subagent_type="implementation-planner")`:

1. Agent reads SPEC document
2. Analyzes requirements and creates execution strategy
3. Returns plan for user approval
4. Waits for user confirmation (proceed/modify/postpone)
5. Stores plan context for Phase 2

### Phase 2: TDD Implementation

Command calls `Task(subagent_type="tdd-implementer")`:

1. Agent initializes task tracking (TodoWrite)
2. Checks domain readiness (if multi-domain SPEC)
3. Executes RED → GREEN → REFACTOR cycle
4. Returns implementation results and coverage metrics

### Phase 2.5: Quality Validation

Command calls `Task(subagent_type="quality-gate")`:

1. Agent verifies TRUST 5 principles (Test-first, Readable, Unified, Secured, Trackable)
2. Validates test coverage (>= 85%)
3. Checks security compliance
4. Returns quality assessment (PASS/WARNING/CRITICAL)

### Phase 3: Git Operations

Command calls `Task(subagent_type="git-manager")`:

1. Agent creates feature branch if needed
2. Creates commits with implementation changes
3. Verifies commits were successful
4. Returns commit summary

### Phase 4: Completion & Guidance

Command calls `AskUserQuestion()`:

1. Displays implementation summary
2. Shows next action options
3. Guides user to `/moai:3-sync` or additional features

---

## 📋 Execution Flow (High-Level)

```text
/moai:2-run [SPEC-ID]
    ↓
Parse SPEC ID from $ARGUMENTS
    ↓
✅ Phase 1: Task(subagent_type="implementation-planner")
    → Analyze SPEC → Create execution plan → Get approval
    ↓
✅ Phase 2: Task(subagent_type="tdd-implementer")
    → RED-GREEN-REFACTOR → Tests passing → Coverage verified
    ↓
✅ Phase 2.5: Task(subagent_type="quality-gate")
    → Validate TRUST 5 principles → Return quality status
    ↓
✅ Phase 3: Task(subagent_type="git-manager")
    → Create feature branch → Commit changes → Verify commits
    ↓
✅ Phase 4: AskUserQuestion(...)
    → Display summary → Guide next steps → Offer options
    ↓
Output: "Implementation complete. Next step: /moai:3-sync"
```

---

## 🎯 Command Implementation

### Sequential Phase Execution

**Command implementation flow:**

```python
# Phase 1: SPEC Analysis & Planning
plan_result = Task(
  subagent_type="implementation-planner",
  description="Analyze SPEC-$ARGUMENTS and create execution plan",
  prompt="""
SPEC ID: $ARGUMENTS

Analyze this SPEC and create detailed execution plan:
1. Extract requirements and success criteria
2. Identify implementation phases and tasks
3. Determine tech stack and dependencies
4. Estimate complexity and effort
5. Create step-by-step execution strategy
6. Present plan for user approval
"""
)

# User approval checkpoint
approval = AskUserQuestion({
    "question": "Does this execution plan look good?",
    "header": "Plan Review",
    "multiSelect": false,
    "options": [
        {"label": "Proceed with plan", "description": "Start implementation"},
        {"label": "Modify plan", "description": "Request changes"},
        {"label": "Postpone", "description": "Stop here, continue later"}
    ]
})

# Phase 2: TDD Implementation
if approval == "Proceed with plan":
    implementation_result = Task(
        subagent_type="tdd-implementer",
        description="Implement SPEC-$ARGUMENTS using TDD cycle",
        prompt="""
SPEC ID: $ARGUMENTS

Execute complete TDD implementation:
1. Write failing tests (RED phase)
2. Implement minimal code (GREEN phase)
3. Refactor for quality (REFACTOR phase)
4. Ensure test coverage >= 85%
5. Verify all tests passing
6. Return implementation summary
"""
    )

    # Phase 2.5: Quality Validation
    quality_result = Task(
        subagent_type="quality-gate",
        description="Validate TRUST 5 compliance for SPEC-$ARGUMENTS",
        prompt="""
SPEC ID: $ARGUMENTS

Validate implementation against TRUST 5 principles:
- T: Test-first (tests exist and pass)
- R: Readable (code is clear and documented)
- U: Unified (follows project patterns)
- S: Secured (no security vulnerabilities)
- T: Trackable (changes are logged and traceable)

Return quality assessment with specific findings.
"""
    )

    # Phase 3: Git Operations
    if quality_result.status == "PASS":
        git_result = Task(
            subagent_type="git-manager",
            description="Create commits for SPEC-$ARGUMENTS implementation",
            prompt="""
SPEC ID: $ARGUMENTS

Create git commits for implementation:
1. Create feature branch: feature/SPEC-$ARGUMENTS
2. Stage all relevant files
3. Create meaningful commits (follow conventional commits)
4. Verify commits created successfully
5. Return commit summary
"""
        )

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

---

## 📊 Design Improvements (vs Previous Version)

| Metric                 | Before           | After          | Improvement            |
| ---------------------- | ---------------- | -------------- | ---------------------- |
| **Command LOC**        | ~420             | ~120           | **71% reduction**      |
| **allowed-tools**      | 14 types         | 1 type         | **93% reduction**      |
| **Direct tool usage**  | Yes (Read/Bash)  | No             | **100% eliminated**    |
| **Agent count**        | 4 separate calls | 1 orchestrator | **100% simplified**    |
| **User approval flow** | In command       | In agent       | **Cleaner separation** |
| **Error handling**     | Dispersed        | Centralized    | **Better structure**   |

---

## 🔍 Verification Checklist

After implementation, verify:

- [ ] ✅ Command has ONLY `Task`, `AskUserQuestion`, `TodoWrite` in allowed-tools
- [ ] ✅ Command contains NO `Read`, `Write`, `Edit`, `Bash` usage
- [ ] ✅ Command delegates execution to phase agents sequentially
- [ ] ✅ Phase 1: implementation-planner executes successfully
- [ ] ✅ Phase 2: tdd-implementer executes successfully
- [ ] ✅ Phase 2.5: quality-gate validates TRUST 5
- [ ] ✅ Phase 3: git-manager creates commits
- [ ] ✅ Phase 4: User guided to next steps
- [ ] ✅ User approval checkpoints working

---

## 📚 Quick Reference

**This command**:

- Accepts SPEC ID: `/moai:2-run SPEC-AUTH-001`
- Orchestrates 4 phases sequentially with user approval checkpoints
- Outputs: Implementation summary with next steps

**For details, see**:

- `.claude/agents/moai/implementation-planner.md` - Phase 1: SPEC analysis
- `.claude/agents/moai/tdd-implementer.md` - Phase 2: TDD implementation
- `.claude/agents/moai/quality-gate.md` - Phase 2.5: Quality validation
- `.claude/agents/moai/git-manager.md` - Phase 3: Git operations

**Architecture Pattern**:

```text
Commands (Orchestration)
    ↓ Task()
Agents (Execution)
    ↓ Skill()
Skills (Knowledge)
```

---

**Version**: 3.1.0 (Command-Level Phase Orchestration)
**Updated**: 2025-11-19
**Pattern**: Sequential Phase-Based Agent Delegation
**Compliance**: Claude Code Best Practices + Zero Direct Tool Usage

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

**Important**:

- Use conversation language from config
- No emojis in any AskUserQuestion fields
- Always provide clear next step options

## ⚡️ EXECUTION DIRECTIVE

**You must NOW execute the command following the "Execution Flow" described above.**

1. Start Phase 1 immediately.
2. Call the `Task` tool with `subagent_type="implementation-planner"`.
3. Do NOT just describe what you will do. DO IT.
