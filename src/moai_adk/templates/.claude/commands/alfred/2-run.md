---
name: alfred:2-run
description: "Execute TDD implementation cycle"
argument-hint: "SPEC-ID - All with SPEC ID to implement (e.g. SPEC-001) or all \"SPEC Implementation\""
allowed-tools:
  - Task
---

# ⚒️ MoAI-ADK Step 2: Execute Implementation (Run) - TDD Implementation

> **Architecture**: Commands → Agents → Skills. This command orchestrates ONLY through `Task()` tool.
>
> **Delegation Model**: Complete agent-first pattern. All execution delegated to run-orchestrator.


**4-Step Workflow Integration**: This command implements Step 3 of Alfred's workflow (Task Execution with TodoWrite tracking). See CLAUDE.md for full workflow details.

---

## 🎯 Command Purpose

Execute TDD implementation of SPEC requirements through complete agent delegation.

The `/alfred:2-run` command orchestrates the complete implementation workflow:
1. **Phase 1**: SPEC analysis and execution plan creation
2. **Phase 2**: TDD implementation (RED → GREEN → REFACTOR)
3. **Phase 3**: Git commit management
4. **Phase 4**: Completion and next steps guidance

**Run on**: `$ARGUMENTS` (SPEC ID, e.g., SPEC-001)

---

## 💡 Execution Philosophy: "Plan → Run → Sync"

`/alfred:2-run` performs SPEC implementation through a complete agent delegation model:

```
User Command: /alfred:2-run SPEC-001
    ↓
/alfred:2-run Command
    └─ Task(subagent_type="run-orchestrator")
        ├─ Phase 1: Analysis & Planning
        ├─ Phase 2: TDD Implementation
        ├─ Phase 3: Git Operations
        └─ Phase 4: Completion
            ↓
        Output: Implemented feature with passing tests and commits
```

### Key Principle: Zero Direct Tool Usage

**This command uses ONLY Task() tool:**
- ❌ No Read (file operations delegated)
- ❌ No Write (file operations delegated)
- ❌ No Edit (file operations delegated)
- ❌ No Bash (all bash commands delegated)
- ❌ No TodoWrite (delegated to run-orchestrator)
- ✅ **Only Task()** for orchestration

All complexity is handled by the **run-orchestrator** agent.

---

## 🧠 Associated Agents & Skills

### Primary Agent: run-orchestrator

**Orchestrates all 4 phases:**
- Coordinates implementation-planner for SPEC analysis
- Manages tdd-implementer for TDD implementation
- Verifies with quality-gate for TRUST 4 compliance
- Creates commits via git-manager

### Supporting Agents (called by run-orchestrator)

| Agent | Purpose | When |
|-------|---------|------|
| **implementation-planner** | Analyzes SPEC and creates execution strategy | Phase 1 |
| **tdd-implementer** | Implements code through TDD cycle | Phase 2 |
| **quality-gate** | Verifies TRUST 4 principles | Phase 2 (after tdd-implementer) |
| **git-manager** | Creates and manages Git commits | Phase 3 |

### Skills Used (by agents, not command)

- `Skill("moai-alfred-workflow")` - Workflow orchestration
- `Skill("moai-alfred-todowrite-pattern")` - Task tracking
- `Skill("moai-alfred-ask-user-questions")` - User interaction
- `Skill("moai-alfred-reporting")` - Result reporting
- Domain-specific skills (selected per language/framework)

---

## 🚀 Execution (Delegated to run-orchestrator)

### Phase 1: Analysis & Planning

The run-orchestrator:
1. Reads SPEC document
2. Invokes implementation-planner to analyze requirements
3. Presents execution plan to user
4. Handles user approval flow (proceed/modify/postpone)
5. Updates SPEC status if approved

### Phase 2: TDD Implementation

The run-orchestrator:
1. Initializes TodoWrite for task tracking
2. Checks domain readiness (if multi-domain SPEC)
3. Invokes tdd-implementer for RED → GREEN → REFACTOR cycle
4. Invokes quality-gate for TRUST 4 validation
5. Handles quality gate results (PASS/WARNING/CRITICAL)

### Phase 3: Git Operations

The run-orchestrator:
1. Invokes git-manager to create commits
2. Verifies commits were created successfully
3. Displays commit summary

### Phase 4: Completion

The run-orchestrator:
1. Displays implementation summary
2. Asks user for next steps
3. Guides to `/alfred:3-sync` or new feature

---

## 📋 Execution Flow (High-Level)

```
/alfred:2-run SPEC-XXX
    ↓
Parse SPEC ID from $ARGUMENTS
    ↓
Task(
  subagent_type="run-orchestrator",
  description="Orchestrate SPEC-XXX implementation",
  prompt="Execute all 4 phases for SPEC-XXX"
)
    ↓
Run-Orchestrator handles:
├─ Phase 1: Planning (calls implementation-planner)
├─ Phase 2: Implementation (calls tdd-implementer + quality-gate)
├─ Phase 3: Git (calls git-manager)
└─ Phase 4: Completion (user interaction)
    ↓
Output: Summary and next steps
```

---

## 🎯 Command Implementation

### Step 1: Orchestrator Invocation

**Use Task tool to invoke run-orchestrator:**

```
Task(
  subagent_type="run-orchestrator",
  description="Orchestrate SPEC-$ARGUMENTS implementation workflow",
  prompt="""
You are the run-orchestrator agent.

Task: Execute SPEC-$ARGUMENTS implementation through 4 phases.

SPEC ID: $ARGUMENTS

Execute:
1. PHASE 1: Analysis & Planning
   - Analyze SPEC requirements
   - Create execution strategy
   - Get user approval

2. PHASE 2: TDD Implementation
   - Initialize task tracking
   - Execute TDD cycle (RED → GREEN → REFACTOR)
   - Validate quality gates

3. PHASE 3: Git Operations
   - Create commits for all work
   - Verify commits

4. PHASE 4: Completion
   - Display summary
   - Ask next steps
   - Guide user

Use your tools:
- Task: Delegate to specialist agents (implementation-planner, tdd-implementer, quality-gate, git-manager)
- AskUserQuestion: User interaction (approval, quality gate decisions, next steps)
- TodoWrite: Task progress tracking
- Read: Config file access only

Report final status and guide user to next action.
"""
)
```

---

## 📊 Design Improvements (vs Previous Version)

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| **Command LOC** | ~420 | ~120 | **71% reduction** |
| **allowed-tools** | 14 types | 1 type | **93% reduction** |
| **Direct tool usage** | Yes (Read/Bash) | No | **100% eliminated** |
| **Agent count** | 4 separate calls | 1 orchestrator | **100% simplified** |
| **User approval flow** | In command | In agent | **Cleaner separation** |
| **Error handling** | Dispersed | Centralized | **Better structure** |

---

## 🔍 Verification Checklist

After implementation, verify:

- [ ] ✅ Command has ONLY `Task` in allowed-tools
- [ ] ✅ Command contains NO `Read`, `Write`, `Edit`, `Bash` usage
- [ ] ✅ Command delegates ALL execution to run-orchestrator
- [ ] ✅ Command file is ~120 lines (vs ~420 before)
- [ ] ✅ run-orchestrator agent exists and works
- [ ] ✅ All 4 phases execute through orchestrator

---

## 📚 Quick Reference

**This command**:
- Accepts SPEC ID: `/alfred:2-run SPEC-AUTH-001`
- Delegates to: run-orchestrator agent
- Outputs: Implementation summary with next steps

**For details, see**:
- `.claude/agents/run-orchestrator.md` - Orchestrator responsibilities
- `.claude/agents/alfred/implementation-planner.md` - SPEC analysis
- `.claude/agents/alfred/tdd-implementer.md` - TDD implementation
- `.claude/agents/alfred/quality-gate.md` - Quality validation
- `.claude/agents/alfred/git-manager.md` - Git operations

**Architecture Pattern**:
```
Commands (Orchestration)
    ↓ Task()
Agents (Execution)
    ↓ Skill()
Skills (Knowledge)
```

---

**Version**: 3.0.0 (Agent-First Orchestration)
**Updated**: 2025-11-12
**Pattern**: Complete Agent Delegation
**Compliance**: Claude Code Best Practices

---

## Final Step: Next Action Selection

After TDD implementation completes, use AskUserQuestion tool to guide user to next action:

```python
AskUserQuestion({
    "questions": [{
        "question": "구현이 완료되었습니다. 다음으로 무엇을 하시겠습니까?",
        "header": "다음 단계",
        "multiSelect": false,
        "options": [
            {
                "label": "문서 동기화",
                "description": "/alfred:3-sync 실행하여 문서 정리 및 PR 생성"
            },
            {
                "label": "추가 구현",
                "description": "더 많은 기능 구현"
            },
            {
                "label": "품질 검증",
                "description": "테스트 및 코드 품질 검토"
            }
        ]
    }]
})
```

**Important**: 
- Use conversation language from config (ko)
- No emojis in any AskUserQuestion fields
- Always provide clear next step options
