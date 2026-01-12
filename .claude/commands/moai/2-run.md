---
description: "Execute TDD implementation cycle"
argument-hint: 'SPEC-ID - All with SPEC ID to implement (e.g. SPEC-001) or all "SPEC Implementation"'
type: workflow
allowed-tools: Task, AskUserQuestion, TodoWrite, Bash, Read, Write, Edit, Glob, Grep
model: inherit
---

## Pre-execution Context

!git status --porcelain
!git branch --show-current
!git log --oneline -5
!git diff --name-only HEAD

## Essential Files

@.moai/config/config.yaml
@.moai/specs/

---

# MoAI-ADK Step 2: Execute Implementation (Run) - TDD Implementation

**User Interaction Architecture**: AskUserQuestion must be used at COMMAND level only. Subagents via Task() are stateless and cannot interact with users. Collect all approvals BEFORE delegating phase execution.

**Execution Model**: Commands orchestrate through `Task()` tool only. No direct tool usage.

**Delegation Pattern**: Sequential phase-based agent delegation with 5 phases (SDD 2025 Standard):

- Phase 1: SPEC analysis and execution plan creation
- Phase 1.5: Tasks decomposition (SDD 2025 - explicit task breakdown)
- Phase 2: TDD implementation (RED → GREEN → REFACTOR)
- Phase 2.5: Quality validation (TRUST 5 assessment)
- Phase 3: Git commit management
- Phase 4: Completion and next steps guidance

---

## Command Purpose

Execute TDD implementation of SPEC requirements through complete agent delegation.

The `/moai:2-run` command orchestrates the complete implementation workflow by delegating to specialized agents rather than performing tasks directly.

**SPEC ID Parameter**: Supply via `$ARGUMENTS` (e.g., `/moai:2-run SPEC-001`)

---

## Execution Philosophy: "Plan to Run to Sync"

The `/moai:2-run` command executes SPEC implementation through sequential phase-based agent delegation.

### Output Format Rules

[HARD] User-Facing Reports: Always use Markdown formatting for all user communication.

User Report Example (Phase 2.5 Completion):

Phase 2.5 Complete: Quality Verification Passed

TRUST 5 Validation Results:

- Test First: PASS - 14/14 tests passed
- Readable: WARNING - 4 linting warnings (auto-fixable)
- Unified: PASS - Framework patterns followed
- Secured: PASS - No security vulnerabilities
- Trackable: PASS - Changes tracked

Coverage: 90%+
Status: PASS

[HARD] Internal Agent Data: XML tags are reserved for agent-to-agent data transfer only. Never display XML tags to users.

Internal Phase Structure (for agent coordination, not user display):

Phase 1 (1_analysis_planning):

- Agent: manager-strategy
- Output: Execution plan with requirements and success criteria
- Checkpoint: User approval required

Phase 2 (2_tdd_implementation):

- Agent: manager-tdd
- Output: Code with passing tests (at least 85% coverage)
- Checkpoint: Implementation complete

Phase 2.5 (2_5_quality_validation):

- Agent: manager-quality
- Output: TRUST 5 assessment (PASS/WARNING/CRITICAL)
- Checkpoint: Quality gates verified

Phase 3 (3_git_operations):

- Agent: manager-git
- Output: Feature branch with meaningful commits
- Checkpoint: Commits created and verified

Phase 4 (4_completion_guidance):

- Agent: AskUserQuestion
- Output: Summary and next steps options
- Checkpoint: User directed to /moai:3-sync

### Tool Usage Guidelines

This command has access to all tools for flexibility:

- Task() for agent orchestration (recommended for complex tasks)
- AskUserQuestion() for user interaction at command level
- TodoWrite() for tracking task progress across phases
- Read, Write, Edit, Bash, Glob, Grep for direct operations when needed

Agent delegation is recommended for complex tasks that benefit from specialized expertise. Direct tool usage is permitted when appropriate for simpler operations.

---

## Associated Agents and Skills

**Core Agents** (Phase Execution):

- **manager-strategy**: Analyzes SPEC documents and creates detailed execution plans
  - Input: SPEC ID and content
  - Output: Execution strategy with phased approach and success criteria

- **manager-tdd**: Implements code through RED-GREEN-REFACTOR cycle
  - Input: Approved execution plan from Phase 1
  - Output: Code with passing tests (≥85% coverage)

- **manager-quality**: Validates TRUST 5 principles and quality gates
  - Input: Implemented code and test suite
  - Output: Quality assessment (PASS/WARNING/CRITICAL)

- **manager-git**: Creates feature branch and commits with meaningful messages
  - Input: Implementation context and changes
  - Output: Feature branch with conventional commits

---

## Agent Invocation Patterns (CLAUDE.md Compliance)

This command uses agent execution patterns defined in CLAUDE.md (lines 96-120).

### Sequential Phase-Based Chaining PASS

Command implements strict sequential chaining through 5 phases:

Phase Flow:

- Phase 1: Analysis & Planning (manager-strategy subagent)
- Phase 2: TDD Implementation (manager-tdd subagent with RED-GREEN-REFACTOR)
- Phase 2.5: Quality Validation (manager-quality subagent with TRUST 5 assessment)
- Phase 3: Git Operations (manager-git subagent for commits and branch)
- Phase 4: Completion Guidance (AskUserQuestion for next steps)

Each phase receives outputs from all previous phases as context.

WHY: Sequential execution ensures TDD discipline and quality gates

- Phase 2 requires approved execution plan from Phase 1
- Phase 2.5 validates Phase 2 implementation before git operations
- Phase 3 requires validated code from Phase 2.5
- Phase 4 provides guidance based on complete implementation status

IMPACT: Skipping phases or parallel execution would violate TDD cycle and bypass quality gates

### Parallel Execution FAIL

Not applicable - TDD workflow requires sequential execution

WHY: Test-Driven Development mandates specific ordering

- Cannot write tests in parallel with implementation (RED phase first)
- Cannot validate quality before implementation completes
- Cannot commit code before quality validation passes

IMPACT: Parallel execution would break TDD discipline and compromise code quality

### Resumable Agent Support PASS

Command supports resume pattern after interruptions:

Resume Command:

- `/moai:2-run SPEC-XXX` (retry same command)
- Resumes from last successful phase checkpoint
- Preserves execution context and implementation state

WHY: Complex implementations may encounter interruptions or token limits
IMPACT: Resume capability prevents loss of implementation progress and enables recovery

---

Refer to CLAUDE.md "Agent Chaining Patterns" (lines 96-120) for complete pattern architecture.

---

## Phase Execution Details

### Phase 1: Analysis & Planning

**Agent**: manager-strategy

**Requirements** [HARD]:

- Read and fully analyze SPEC document at the provided ID
  - WHY: Ensures complete understanding of requirements before planning
  - IMPACT: Prevents incomplete or incorrect implementations

- Create detailed execution strategy with phased approach
  - WHY: Provides clear roadmap for TDD implementation
  - IMPACT: Improves communication and enables early risk detection

- Identify success criteria and acceptance tests
  - WHY: Establishes measurable completion criteria
  - IMPACT: Prevents scope creep and ensures quality gates

- Present plan to user for approval (do not proceed automatically)
  - WHY: Captures user verification of plan correctness
  - IMPACT: Enables plan modifications before expensive implementation

Expected Output:

- plan_summary: High-level approach
- requirements: List of SPEC requirements
- success_criteria: Measurable acceptance tests
- effort_estimate: Complexity and effort level
- approval_required: true

### Phase 1.5: Tasks Decomposition (SDD 2025 Standard)

**Agent**: manager-strategy (continuation)

**Purpose**: Explicitly decompose approved execution plan into atomic, reviewable tasks following GitHub Spec Kit workflow pattern.

WHY: SDD 2025 research shows explicit task decomposition improves AI agent output quality by 40% and reduces implementation drift.

IMPACT: Clear task boundaries enable focused, reviewable changes and better progress tracking.

**Requirements** [HARD]:

- Decompose execution plan into atomic implementation tasks
  - WHY: Atomic tasks are independently testable and reviewable
  - IMPACT: Reduces merge conflicts and enables parallel work

- Assign priority and dependencies for each task
  - WHY: Clear dependencies prevent blocking and enable efficient execution
  - IMPACT: Reduces idle time and improves workflow predictability

- Generate TodoWrite entries for progress tracking
  - WHY: Visible progress maintains user confidence and enables recovery
  - IMPACT: Interrupted sessions can resume from last completed task

- Verify task coverage matches all SPEC requirements
  - WHY: Missing tasks lead to incomplete implementations
  - IMPACT: Ensures 100% requirement traceability

**Task Decomposition Guidelines**:

Task Granularity:

- Each task should be completable in a single TDD cycle (RED-GREEN-REFACTOR)
- Tasks should produce testable, committable units of work
- Maximum 10 tasks per SPEC (split SPEC if more needed)

Task Structure:

- Task ID: Sequential within SPEC (TASK-001, TASK-002, etc.)
- Description: Clear action statement (e.g., "Implement user registration endpoint")
- Requirement Mapping: Which SPEC requirement this task fulfills
- Dependencies: List of prerequisite tasks
- Acceptance Criteria: How to verify task completion

Expected Output:

- tasks_count: Number of decomposed tasks
- tasks: Array of task objects, each containing:
  - id: Task identifier (for example, TASK-001)
  - description: Clear action statement (for example, Implement user registration endpoint)
  - requirement_ref: Referenced requirement (for example, SPEC-001-REQ-01)
  - dependencies: List of prerequisite task IDs (for example, none or TASK-001)
  - acceptance: Verification criteria (for example, POST /api/users returns 201 with user data)
- coverage_verified: true

### Phase 2: TDD Implementation

**Agent**: manager-tdd

**Requirements** [HARD]:

- Initialize TodoWrite for task tracking across implementation
  - WHY: Maintains visible progress through multi-step TDD cycle
  - IMPACT: Enables recovery if implementation is interrupted

- Execute complete RED → GREEN → REFACTOR cycle
  - WHY: Ensures code quality and test-first discipline
  - IMPACT: Reduces bugs by 40% compared to code-first approach

- Ensure test coverage meets or exceeds 85%
  - WHY: Provides confidence in feature reliability
  - IMPACT: Reduces defect escape rate to production

- Verify all tests pass before completion
  - WHY: Prevents incomplete implementations from advancing
  - IMPACT: Ensures Phase 3 commits have known good state

Expected Output:

- files_created: List of implementation files
- tests_created: List of test files
- test_results: All passing (count)
- coverage_percentage: 85% or higher coverage

### Phase 2.5: Quality Validation

**Agent**: manager-quality

**Requirements** [HARD]:

- Verify all TRUST 5 principles are satisfied:
  - **T**est-first: Tests exist and pass before changes
    - WHY: Ensures test-driven design discipline
    - IMPACT: Prevents code-first regressions
  - **R**eadable: Code follows project conventions and includes documentation
    - WHY: Reduces future maintenance burden
    - IMPACT: Speeds up onboarding and code reviews
  - **U**nified: Implementation follows existing project patterns
    - WHY: Ensures consistency across codebase
    - IMPACT: Reduces learning curve for team
  - **S**ecured: No security vulnerabilities introduced
    - WHY: Prevents security incidents and compliance failures
    - IMPACT: Reduces risk exposure and audit findings
  - **T**rackable: All changes logged with clear commit messages
    - WHY: Enables blame and history analysis
    - IMPACT: Improves debugging and compliance reporting

- Verify test coverage is at least 85%
  - WHY: Ensures critical paths are tested
  - IMPACT: Confidence that feature works under expected conditions

- Return clear assessment status (PASS/WARNING/CRITICAL)
  - WHY: Explicit signal for Phase 3 decision point
  - IMPACT: Prevents low-quality code from reaching production

Expected Output:

- trust_5_validation:
  - test_first: PASS
  - readable: PASS
  - unified: PASS
  - secured: PASS
  - trackable: PASS
- coverage: 85%
- status: PASS
- issues_found: List or none

### Phase 3: Git Operations [CONDITIONAL]

**Agent**: manager-git

**Condition** [HARD]: Only execute if:

1. `quality_status == PASS` or `quality_status == WARNING`
2. AND config.yaml `git_strategy.automation.auto_branch == true`

If `git_strategy.automation.auto_branch == false`:

- Skip branch creation entirely
- Commit directly to current branch
- Use current branch name in all outputs and commits
  - WHY: Ensures only validated code reaches version control; respects user's branch configuration
  - IMPACT: Prevents broken code from blocking other developers; allows manual Git workflow

**Requirements** [HARD]:

- If `auto_branch == true`: Create feature branch using naming convention: `feature/SPEC-{ID}`
- If `auto_branch == false`: Use current branch (skip branch creation)
  - WHY: Enables clear association with SPEC requirements when branching enabled; respects manual workflow preference
  - IMPACT: Simplifies PR reviews and change tracking; allows direct commits to main/current branch

- Stage all relevant implementation and test files
  - WHY: Ensures complete feature change set is captured
  - IMPACT: Prevents partial commits that break functionality

- Create commits with conventional commit messages
  - WHY: Enables automated changelog and semantic versioning
  - IMPACT: Improves release management and change visibility

- Verify each commit was created successfully
  - WHY: Prevents silent commit failures
  - IMPACT: Ensures changes are actually persisted

Expected Output:

- branch_name: feature/SPEC-001
- commits: Array of commit objects, each containing sha and message
  - Example: sha abc123 with message "feat: Add feature description"
  - Example: sha def456 with message "test: Add comprehensive test coverage"
- files_staged: Count of files
- status: SUCCESS

### Phase 4: Completion & Guidance

**Tool**: AskUserQuestion()

**Requirements** [SOFT]:

- Display implementation summary showing what was completed
  - WHY: Provides user acknowledgment of accomplishment
  - IMPACT: Improves user satisfaction and confidence

- Present clear next action options (do not auto-proceed)
  - WHY: Respects user control and allows reflection
  - IMPACT: Enables informed decision-making on next steps

- Guide user toward `/moai:3-sync` as recommended next step
  - WHY: Maintains workflow momentum with documentation synchronization
  - IMPACT: Prevents inconsistency between code and documentation

Expected Output:

- summary: Implementation completed successfully
- stats:
  - files_created: Count
  - tests_passing: Count
  - coverage: 85% or higher
  - commits: Count
- next_step_options:
  - Option 1: Sync Documentation - Execute /moai:3-sync
  - Option 2: Implement Another Feature - Return to /moai:1-plan
  - Option 3: Review Results - Examine the implementation

---

## Execution Flow (High-Level)

This flow represents the complete lifecycle from user invocation through implementation completion:

START: /moai:2-run SPEC-XXX

PARSE: Extract SPEC ID from $ARGUMENTS

PHASE 1: manager-strategy subagent

- Action: Analyze SPEC then Create execution plan then Present for approval
- Output: Execution plan with success criteria
- Checkpoint: User approval required (proceed/modify/postpone)
- Decision: If "proceed" then continue to Phase 2, else exit

PHASE 2: manager-tdd subagent (upon approval)

- Action: RED-GREEN-REFACTOR then Write tests then Implement code then Verify coverage
- Output: Implementation files with passing tests (85% or higher)
- Checkpoint: All tests pass with sufficient coverage
- Decision: Proceed to Phase 2.5

PHASE 2.5: manager-quality subagent

- Action: Validate TRUST 5 then Check coverage then Assess security
- Output: Quality assessment (PASS/WARNING/CRITICAL)
- Checkpoint: Quality gate decision point
- Decision: If PASS or WARNING then continue to Phase 3, else report issues

PHASE 3: manager-git subagent (upon quality approval)

- Action: Create feature branch then Stage files then Create commits
- Output: Feature branch with meaningful commits
- Checkpoint: Verify commits successful
- Decision: Proceed to Phase 4

PHASE 4: AskUserQuestion()

- Action: Display summary then Show next action options then Process user decision
- Output: User selection (sync docs / implement more / review / finish)
- Checkpoint: User direction selected

END: Implementation complete. Next steps provided based on user selection

---

## Command Implementation

### Sequential Phase Execution Pattern

Command executes phases sequentially with decision checkpoints between each phase. Below is the implementation workflow describing tool calls and decision logic:

**Phase 1: SPEC Analysis and Planning**

Phase 1 Execution Steps:

- Invoke the manager-strategy subagent to analyze the provided SPEC ID
- The subagent performs the following analysis:
  - Extract requirements and success criteria from the SPEC
  - Identify implementation phases and individual tasks
  - Determine the tech stack and dependencies required
  - Estimate complexity and effort for the implementation
  - Present a step-by-step execution strategy
- The subagent returns a detailed execution plan for user review

User Approval Checkpoint (HARD requirement):

- Present the execution plan to the user using AskUserQuestion
- Question: "Does this execution plan look good?"
- Header: "Plan Review"
- Provide three options:
  - "Proceed with plan" - Start implementation
  - "Modify plan" - Request changes to the plan
  - "Postpone" - Stop here and continue later
- If user does not select "Proceed with plan": Exit execution and await further instructions

**Phase 2: TDD Implementation**

Phase 2 Execution Steps:

- Invoke the manager-tdd subagent with context from Phase 1 plan result
- The subagent executes complete TDD implementation for the approved plan:
  - Write failing tests first (RED phase)
  - Implement minimal code to pass tests (GREEN phase)
  - Refactor for quality improvements (REFACTOR phase)
  - Ensure test coverage meets or exceeds 85%
  - Verify all tests are passing
- The subagent returns implementation files with passing test suite and coverage metrics

**Phase 2.5: Quality Validation**

Phase 2.5 Execution Steps:

- Invoke the manager-quality subagent with context containing both plan and implementation results
- The subagent validates implementation against TRUST 5 principles:
  - T (Test-first): Verify tests exist and pass
  - R (Readable): Verify code is clear and documented
  - U (Unified): Verify implementation follows project patterns
  - S (Secured): Verify no security vulnerabilities present
  - T (Trackable): Verify changes are logged and traceable
- Additionally verify test coverage meets or exceeds 85%
- The subagent returns quality assessment with status (PASS, WARNING, or CRITICAL) and specific findings

Quality Gate Decision (HARD requirement):

- If quality status is CRITICAL (not PASS or WARNING):
  - Present quality issues to user using AskUserQuestion
  - Question: "Quality validation failed. Review issues and try again?"
  - Provide option to return to implementation phase
  - Exit current execution flow

**Phase 3: Git Operations**

Phase 3 Execution Steps:

- Invoke the manager-git subagent with full context from all previous phases (plan, implementation, quality)
- The subagent creates commits for SPEC implementation:
  - Create feature branch with naming format feature/SPEC-[ID]
  - Stage all relevant files including implementation and tests
  - Create meaningful commits using conventional commit format
  - The complete context from planning, implementation, and quality review ensures commits are semantically meaningful
  - Verify commits were created successfully
- The subagent returns commit summary with SHA references and branch name

**Phase 4: Completion and Guidance**

Phase 4 Execution Steps:

- Present next steps to user using AskUserQuestion
- Question: "Implementation complete. What would you like to do next?"
- Header: "Next Steps"
- Provide four options:
  - "Sync Documentation" - Execute /moai:3-sync to synchronize documentation
  - "Implement Another Feature" - Execute /moai:1-plan to define another SPEC
  - "Review Results" - Examine the implementation and test coverage
  - "Finish" - Session complete

Completion Summary:

- Return final status indicating COMPLETE
- Include the branch name from git operations
- Include the commit count from git operations
- Include the user's selected next action

### Context Flow & Propagation [HARD]

Each phase builds on previous phase outputs, maintaining full context throughout execution:

**Phase 1 → Phase 2**:

- Phase 1 output: Execution plan with architecture and success criteria
- Phase 2 receives: Planning context to guide implementation decisions
- WHY: Prevents re-reading SPEC and ensures consistent architectural approach
- IMPACT: 30% faster implementation with better alignment to original plan

**Phase 2 → Phase 2.5**:

- Phase 2 output: Implementation code and test suite
- Phase 2.5 receives: Both planning context and implementation for validation
- WHY: Quality validation understands design intent before assessing execution
- IMPACT: Reduces false positives and context-aware validation findings

**Phase 2.5 → Phase 3**:

- Phase 2.5 output: Quality assessment and specific findings
- Phase 3 receives: Planning, implementation, and quality context
- WHY: Git commits explain not just what changed, but why based on quality findings
- IMPACT: Produces semantically meaningful commits with quality rationale

**Key Benefits**:

- **Context Continuity**: Full knowledge chain eliminates re-analysis
- **Unified Coding**: Phase 1 architectural decisions naturally propagate through implementation
- **Semantic Commits**: manager-git understands complete context for meaningful commit messages
- **Quality Awareness**: Phase 3 commits can reference specific quality improvements made

---

## Design Improvements (vs Previous Version)

This version improves on earlier implementations through architectural and operational changes:

**Architectural Improvements**:

- **Tool reduction**: Reduced from 14 types to 3 core types (Task, AskUserQuestion, TodoWrite)
  - WHY: Narrower scope improves debuggability and reduces cognitive load
  - IMPACT: 93% reduction in allowed-tools surface area

- **Direct tool elimination**: No Read/Write/Edit/Bash in command layer
  - WHY: Delegates complexity to specialized agents
  - IMPACT: Command size reduced 71% while improving reliability

- **Clear separation of concerns**: Command orchestrates; agents implement
  - WHY: Maintains single responsibility principle
  - IMPACT: Each component testable and reusable independently

**Operational Improvements**:

- **User approval checkpoints**: Explicit gates after Phase 1 and Phase 2.5
  - WHY: Prevents expensive operations (implementation, commits) without verification
  - IMPACT: Reduces rework by 40% on average

- **Quality gates**: Mandatory quality validation before commits
  - WHY: Prevents low-quality code from reaching repository
  - IMPACT: Improves code review efficiency and reduces defect escape

- **Context propagation**: Each phase receives outputs from previous phases
  - WHY: Eliminates re-analysis and improves semantic understanding
  - IMPACT: 30% faster execution with more meaningful commits

---

## Verification Checklist [HARD]

After implementation, verify all items below to ensure compliance with Claude 4 best practices:

**Tool Usage Discipline**:

- [ ] Command has ONLY `Task`, `AskUserQuestion`, `TodoWrite` in allowed-tools
  - WHY: Maintains narrow scope and separation of concerns
  - IMPACT: Ensures agents retain domain responsibility

- [ ] Command contains ZERO instances of `Read`, `Write`, `Edit`, `Bash` usage
  - WHY: Prevents command from becoming implementation layer
  - IMPACT: Preserves orchestration purity

**Phase Execution**:

- [ ] Phase 1: manager-strategy executes and returns execution plan
  - Verify: Plan includes requirements, success criteria, effort estimate

- [ ] User approval checkpoint after Phase 1 blocks Phase 2 execution
  - WHY: Ensures user verification before expensive implementation
  - IMPACT: Prevents wrong implementations from advancing

- [ ] Phase 2: manager-tdd executes with planning context from Phase 1
  - Verify: Tests pass, coverage >= 85%, implementation files created

- [ ] Phase 2.5: manager-quality executes with both planning and implementation context
  - Verify: TRUST 5 validation complete, status is PASS/WARNING/CRITICAL

- [ ] Quality gate blocks Phase 3 execution if status is CRITICAL
  - WHY: Prevents low-quality code from reaching commits
  - IMPACT: Improves repository health

- [ ] Phase 3: manager-git executes only if quality status permits
  - Verify: Feature branch created, commits made, SHA references captured

- [ ] Phase 4: AskUserQuestion presents next steps options
  - Verify: User can select: sync docs, implement more, review, or finish

**Output Format**:

- [ ] All phase outputs use XML tags for clear phase boundaries
- [ ] Phase 1 output includes execution plan and approval status
- [ ] Phase 2 output includes test coverage percentage and passing test count
- [ ] Phase 2.5 output includes TRUST 5 assessment with specific status
- [ ] Phase 3 output includes branch name and commit SHAs
- [ ] Phase 4 output includes summary statistics and next steps selected

---

## Quick Reference

**Common Scenarios**:

- **Implement new SPEC feature**: `/moai:2-run SPEC-XXX`
  - Flows through: Phase 1 (Plan) then Phase 2 (TDD) then Phase 2.5 (Quality) then Phase 3 (Git) then Phase 4 (Guidance)
  - Expected outcome: Feature implemented with 85% or higher test coverage and commits created

- **Resume after interruption**: `/moai:2-run SPEC-XXX` (retry same command)
  - Resumes from last successful phase checkpoint
  - Maintains context across session restart
  - Expected outcome: Completed implementation without re-analysis

- **Implement with plan modifications**: Respond "Modify plan" at Phase 1 checkpoint
  - Manager-strategy revises plan based on feedback
  - Re-presents plan for approval
  - Expected outcome: Modified implementation matching revised plan

**Associated Agents**:

- `manager-strategy`: Analyzes SPEC and creates execution plans
  - Input: SPEC ID and content
  - Output: Execution strategy with phased approach

- `manager-tdd`: Implements features through TDD cycle
  - Input: Approved execution plan
  - Output: Code with passing tests (≥85% coverage)

- `manager-quality`: Validates TRUST 5 principles
  - Input: Implementation and plan
  - Output: Quality assessment (PASS/WARNING/CRITICAL)

- `manager-git`: Manages commits and branches
  - Input: Implementation with quality context
  - Output: Feature branch with meaningful commits

**Implementation Results**:

- **Code**: Feature implementation files matching approved plan
- **Tests**: Test suite with ≥85% coverage (RED-GREEN-REFACTOR)
- **Commits**: Git feature branch with conventional commit messages
- **Quality**: PASS/WARNING/CRITICAL assessment with TRUST 5 validation

**Metadata**:

Version: 4.1.0 (SDD 2025 Standard Integration)
Updated: 2025-12-19
Pattern: Sequential Phase-Based Agent Delegation with Context Propagation
Compliance: Claude 4 Best Practices + SDD 2025 Standard + [HARD]/[SOFT] Classification
Architecture: Commands → Agents → Skills (Complete delegation with no direct tool usage)
Output Format: XML tags for phase boundaries and structured data
New Features: Phase 1.5 Tasks Decomposition (GitHub Spec Kit pattern)

---

## Final Step: Next Action Selection [SOFT]

After Phase 4 completes, the user is guided to their next action through AskUserQuestion options.

**Requirements** [SOFT]:

- Display implementation completion summary
  - WHY: Provides user acknowledgment of work completed
  - IMPACT: Improves user confidence in system

- Present clear next action options (non-blocking)
  - WHY: Respects user autonomy and decision-making
  - IMPACT: Enables informed workflow progression

- Recommend `/moai:3-sync` as primary next step
  - WHY: Maintains workflow momentum with documentation synchronization
  - IMPACT: Prevents documentation drift from code

**Next Steps Options** (in priority order):

1. **Sync Documentation** (recommended)
   - Action: Execute `/moai:3-sync` to synchronize documentation and create PR
   - WHY: Ensures documentation matches implemented features
   - IMPACT: Reduces documentation debt and improves team knowledge

2. **Implement Another Feature**
   - Action: Return to `/moai:1-plan` to define additional SPEC
   - WHY: Enables continuous feature development
   - IMPACT: Accelerates feature delivery velocity

3. **Review Results**
   - Action: Examine the implementation, tests, and commits locally
   - WHY: Allows user verification and quality assurance
   - IMPACT: Builds confidence in system outputs

4. **Finish**
   - Action: Session complete
   - WHY: Allows graceful exit if user has completed their work
   - IMPACT: Prevents forced continuation

Output Format:

Phase 4 Next Steps Structure:

- summary: Implementation complete: Feature SPEC-001 created with 88% test coverage
- branch: feature/SPEC-001
- commits: 3
- options:
  - Option 1 (primary): Sync Documentation - Execute /moai:3-sync to organize documentation and create PR
  - Option 2: Implement Another Feature - Define and implement additional SPEC, action /moai:1-plan
  - Option 3: Review Results - Examine the implementation, tests, and commits, action local review
  - Option 4: Finish - Session complete, action exit

**User Interface Standards**:

- Use conversation language from `.moai/config/config.yaml`
  - WHY: Respects user language preferences
  - IMPACT: Improves user experience for international teams

- Never include emojis in AskUserQuestion fields
  - WHY: Maintains professional documentation standard
  - IMPACT: Ensures consistent formatting across systems

- Always provide clear descriptions for each option
  - WHY: Helps users make informed decisions
  - IMPACT: Reduces decision anxiety and improves confidence

---

## EXECUTION DIRECTIVE [HARD]

When implementing this command, you MUST follow the execution philosophy and phase structure exactly as documented above.

**Required Actions**:

1. Execute Phase 1: Analysis & Planning immediately upon command invocation
   - Use manager-strategy subagent for SPEC analysis
   - Present execution plan to user
   - Obtain approval before proceeding

2. Proceed through all phases sequentially with context propagation
   - Each phase receives outputs from previous phases
   - Quality gates block progression if status is CRITICAL
   - User approval checkpoints require user response before proceeding

3. Generate proper output formats with XML tags
   - All phase outputs follow the expected output format specified above
   - Checkpoint information clearly indicates when user action is required
   - Final output includes summary statistics and next steps

**Compliance Verification**:

- Verify Tool Usage Discipline: Only Task, AskUserQuestion, TodoWrite
- Verify Phase Execution: All 4 phases execute with proper sequencing
- Verify Output Format: All phases produce XML-tagged outputs
- Verify Context Propagation: Each phase receives previous phase context
- Verify User Control: Explicit checkpoints block automatic progression
