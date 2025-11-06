---
name: moai-alfred-workflow-core
version: 1.0.0
created: 2025-11-06
updated: 2025-11-06
status: active
description: Core 4-step Alfred workflow execution system with intent clarification, task planning, progress tracking, and quality gates. Essential for systematic development with transparency and traceability. Use when executing multi-step tasks, planning complex features, or ensuring quality standards.
keywords: ['workflow', 'execution', 'planning', 'task-tracking', 'quality-gates', 'alfred', '4-step']
allowed-tools:
  - Read
  - TodoWrite
---

# Alfred Core 4-Step Workflow System

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-workflow-core |
| **Version** | 1.0.0 (2025-11-06) |
| **Status** | Active |
| **Tier** | Alfred Core |
| **Purpose** | Systematic 4-step workflow execution |

---

## What It Does

Alfred uses a consistent 4-step workflow for all user requests to ensure clarity, planning, transparency, and traceability.

**Key capabilities**:
- ✅ Intent clarification with structured questions
- ✅ Task planning and decomposition with dependencies
- ✅ Transparent progress tracking with TodoWrite
- ✅ Automated reporting and git commits
- ✅ Quality gate validation and compliance
- ✅ SPEC-First TDD integration

---

## When to Use

**Automatic triggers**:
- User request received → analyze intent clarity
- Multiple interpretation possible → use AskUserQuestion
- Task complexity > 1 step → invoke Plan Agent
- Executing tasks → activate TodoWrite tracking
- Task completion → generate report and commit

**Manual reference**:
- Understanding workflow execution patterns
- Planning multi-step features and projects
- Learning best practices for task tracking
- Ensuring quality standards compliance

---

## The 4-Step Workflow

### Step 1: Intent Understanding

**Goal**: Clarify user intent before any action

**Actions**:
- Evaluate request clarity:
  - **HIGH clarity**: Tech stack, requirements, scope all specified → Skip to Step 2
  - **MEDIUM/LOW clarity**: Multiple interpretations possible → Invoke AskUserQuestion
- Present 3-5 clear options (never open-ended)
- Gather user responses before proceeding

**When to Ask Questions**:
```
Triggers for AskUserQuestion:
├─ Multiple tech stack choices available
├─ Architecture decisions needed
├─ Business/UX decisions unclear
├─ Ambiguous requirements
└─ Existing component impacts unknown
```

**Example**:
```
User says: "Add authentication"
          ↓
Clarity = MEDIUM (multiple approaches possible)
          ↓
Ask: "Which authentication method?"
- Option 1: JWT tokens (stateless, scalable)
- Option 2: OAuth 2.0 (third-party providers)
- Option 3: Session-based (server-side state)
```

---

### Step 2: Plan Creation

**Goal**: Analyze tasks and identify execution strategy

**Actions**:
- Invoke Plan Agent (built-in Claude agent) to:
  - Decompose tasks into structured steps
  - Identify dependencies between tasks
  - Determine single vs parallel execution opportunities
  - Estimate file changes and work scope
- Output structured task breakdown with phases

**Plan Output Format**:
```
Task Breakdown:

Phase 1: Preparation (30 mins)
├─ Task 1: Set up development environment
├─ Task 2: Install required dependencies
└─ Task 3: Create test fixtures and mocks

Phase 2: Implementation (2 hours)
├─ Task 4: Implement core feature (parallel ready)
├─ Task 5: Create API endpoints (parallel ready)
└─ Task 6: Write integration tests (depends on 4, 5)

Phase 3: Verification (30 mins)
├─ Task 7: Integration testing
├─ Task 8: Documentation updates
└─ Task 9: Code review and quality checks
```

---

### Step 3: Task Execution

**Goal**: Execute tasks with transparent progress tracking

**Actions**:
1. Initialize TodoWrite with all tasks (status: pending)
2. For each task:
   - Update TodoWrite: pending → **in_progress**
   - Execute task (call appropriate sub-agent or action)
   - Update TodoWrite: in_progress → **completed**
3. Handle blockers: Keep in_progress, create new blocking task

**TodoWrite Rules**:
- Each task must have:
  - `content`: Imperative form ("Run tests", "Fix bug")
  - `activeForm`: Present continuous ("Running tests", "Fixing bug")  
  - `status`: One of pending/in_progress/completed
- **EXACTLY ONE** task in_progress at a time (unless Plan Agent approved parallel)
- Mark completed ONLY when fully done (tests pass, no errors, implementation complete)

**Example TodoWrite Progression**:

Initial state (all pending):
```
1. [pending] Set up environment
2. [pending] Install dependencies  
3. [pending] Implement core feature
4. [pending] Write tests
5. [pending] Documentation
```

After starting Task 1:
```
1. [in_progress] Set up environment     ← ONE task in progress
2. [pending] Install dependencies
3. [pending] Implement core feature
4. [pending] Write tests
5. [pending] Documentation
```

**Handling Blockers**:
```
Example: Task 4 blocked by missing library

Action:
├─ Keep Task 4 as in_progress
├─ Create new task: "Install library X"
├─ Add to todo list with high priority
└─ Start new task first, then return to Task 4
```

---

### Step 4: Report & Commit

**Goal**: Document work and create git history

**Actions**:
- **Report Generation**: ONLY if user explicitly requested
  - ❌ Don't auto-generate in project root
  - ✅ OK to generate in `.moai/docs/`, `.moai/reports/`, `.moai/analysis/`
- **Git Commit**: ALWAYS create commits (mandatory)
  - Use git-manager for all Git operations
  - Follow TDD commits: RED → GREEN → REFACTOR
  - Include Alfred co-authorship

**Report Conditions**:
```
User says: "Show me a report" → Generate report
User says: "Create analysis" → Generate report  
User says: "I'm done with feature X" → NO auto-report
```

**Commit Message Format**:
```
feat: Add authentication support

- JWT token validation implemented
- Session management added
- Rate limiting configured

🤖 Generated with Claude Code

Co-Authored-By: Alfred <alfred@mo.ai.kr>
```

---

## SPEC-First TDD Integration

This workflow integrates seamlessly with SPEC-First TDD:

### During Step 2 (Planning):
- Reference existing SPEC documents with @SPEC:ID tags
- Plan tasks in TDD order: RED → GREEN → REFACTOR

### During Step 3 (Execution):
- **RED**: Write failing tests with @TEST:ID tags
- **GREEN**: Implement minimal code with @CODE:ID tags  
- **REFACTOR**: Improve code while maintaining SPEC compliance

### During Step 4 (Reporting):
- Verify @TAG chain integrity (SPEC→TEST→CODE→DOC)
- Document completion with @DOC:ID tags
- Generate sync report

---

## Quality Gates

### Pre-Execution Validation:
- [ ] Intent clarified (AskUserQuestion used if needed)
- [ ] Task plan created with dependencies
- [ ] TodoWrite initialized with all tasks
- [ ] Quality criteria defined

### Post-Execution Validation:
- [ ] All tasks completed successfully
- [ ] Tests pass (coverage ≥85%)
- [ ] Code linting and type checking pass
- [ ] Documentation updated
- [ ] Git commit created with proper format
- [ ] @TAG integrity maintained

---

## Decision Trees

### When to Use AskUserQuestion
```
Request clarity unclear?
├─ YES → Use AskUserQuestion
│   ├─ Present 3-5 clear options
│   ├─ Use structured format with header/description
│   └─ Wait for user response before proceeding
└─ NO → Proceed to planning phase
```

### When to Mark Task Completed
```
Task marked in_progress?
├─ Code implemented → Tests pass?
├─ Tests pass → Type checking pass?
├─ Type checking pass → Linting pass?
└─ All quality gates pass → Mark COMPLETED ✅
   └─ Any gate fails → Keep in_progress ⏳
```

### When to Create Blocking Task
```
Task execution blocked?
├─ External dependency missing?
├─ Pre-requisite not completed?
├─ Unknown issue requiring research?
└─ YES → Create blocking task
   ├─ Add to todo list with high priority
   ├─ Execute blocking task first
   └─ Return to original task
```

---

## Key Principles

1. **Clarity First**: Never assume user intent
2. **Systematic**: Follow 4 steps in strict order
3. **Transparent**: Track all progress visually with TodoWrite
4. **Traceable**: Document every decision and action
5. **Quality**: Validate before marking complete
6. **SPEC-First**: Always reference SPEC documents
7. **TDD-Driven**: Follow RED-GREEN-REFACTOR cycle

---

## Integration with Other Skills

### Required Skills:
- `moai-alfred-ask-user-questions`: For intent clarification
- `moai-alfred-todowrite-pattern`: For task tracking
- `moai-foundation-specs`: For SPEC document management
- `moai-foundation-tags`: For @TAG system management

### Complementary Skills:
- `moai-alfred-best-practices`: Quality standards and compliance
- `moai-cc-skill-factory`: For creating new Skills
- `moai-context-optimization`: For managing context budgets

---

## Examples

### Simple Feature Workflow:
```
User: "Add user registration"
↓
Step 1: Clarify (email verification? social login?)
Step 2: Plan (API, DB, tests, docs)
Step 3: Execute (RED → GREEN → REFACTOR)
Step 4: Commit (with proper message and @TAGs)
```

### Complex Multi-Phase Project:
```
User: "Build payment processing system"
↓
Step 1: Clarify (payment providers? compliance?)
Step 2: Plan (Phase 1: Basic, Phase 2: Advanced, Phase 3: Analytics)
Step 3: Execute (with parallel tasks where possible)
Step 4: Report + Commit (per phase + final)
```

---

**End of Skill** | Consolidated from moai-alfred-workflow + moai-alfred-dev-guide
