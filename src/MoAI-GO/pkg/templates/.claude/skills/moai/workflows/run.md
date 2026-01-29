# Run Workflow Orchestration

## Purpose

Implement SPEC requirements through Domain-Driven Development (DDD) methodology using the ANALYZE-PRESERVE-IMPROVE cycle. This is the second step of the Plan-Run-Sync workflow.

## Scope

- Implements Step 3 of MoAI's 4-step workflow (Task Execution)
- Receives SPEC documents created by /moai plan
- Hands off to /moai sync for documentation and PR

## Input

- $ARGUMENTS: SPEC-ID to implement (e.g., SPEC-AUTH-001)
- Resume: Re-running /moai run SPEC-XXX resumes from last successful phase checkpoint

## Context Loading

Before execution, load these essential files:

- .moai/config/config.yaml (git strategy, automation settings)
- .moai/config/sections/quality.yaml (coverage targets, TRUST 5 settings)
- .moai/config/sections/git-strategy.yaml (auto_branch, branch creation policy)
- .moai/config/sections/language.yaml (git_commit_messages setting)
- .moai/specs/SPEC-{ID}/ directory (spec.md, plan.md, acceptance.md)

Pre-execution commands: git status, git branch, git log, git diff.

---

## Phase Sequence

All phases execute sequentially. Each phase receives outputs from all previous phases as context. Parallel execution is not permitted because DDD methodology mandates specific ordering.

### Phase 1: Analysis and Planning

Agent: manager-strategy subagent

Input: SPEC document content from the provided SPEC-ID.

Tasks for manager-strategy:

- Read and fully analyze the SPEC document
- Extract requirements and success criteria
- Identify implementation phases and individual tasks
- Determine tech stack and dependencies required
- Estimate complexity and effort
- Create detailed execution strategy with phased approach

Output: Execution plan containing plan_summary, requirements list, success_criteria, and effort_estimate.

### Decision Point 1: Plan Approval

Tool: AskUserQuestion (at orchestrator level)

Options:

- Proceed with plan (continue to Phase 1.5)
- Modify plan (collect feedback, re-run Phase 1)
- Postpone (exit, continue later)

If user does not select "Proceed": Exit execution.

### Phase 1.5: Task Decomposition

Agent: manager-strategy subagent (continuation)

Purpose: Decompose the approved execution plan into atomic, reviewable tasks following SDD 2025 standard.

Tasks for manager-strategy:

- Decompose plan into atomic implementation tasks
- Each task must be completable in a single DDD cycle
- Assign priority and dependencies for each task
- Generate task tracking entries for progress visibility
- Verify task coverage matches all SPEC requirements

Task structure for each decomposed task:

- Task ID: Sequential within SPEC (TASK-001, TASK-002, etc.)
- Description: Clear action statement
- Requirement Mapping: Which SPEC requirement it fulfills
- Dependencies: List of prerequisite tasks
- Acceptance Criteria: How to verify completion

Constraints: Maximum 10 tasks per SPEC. If more needed, the SPEC should be split.

Output: Task list with coverage_verified flag set to true.

### Phase 2: DDD Implementation

Agent: manager-ddd subagent

Input: Approved execution plan from Phase 1 plus task decomposition from Phase 1.5.

The DDD cycle executes three stages:

- ANALYZE: Identify domain boundaries, coupling metrics, and refactoring targets. Read existing code and map dependencies.
- PRESERVE: Verify existing tests. Create characterization tests for uncovered code paths to establish a safety net before changes.
- IMPROVE: Apply incremental transformations with continuous verification. Run all tests after each transformation.

Requirements:

- Initialize task tracking for progress across refactoring steps
- Execute the complete ANALYZE-PRESERVE-IMPROVE cycle
- Verify all existing tests pass after each transformation
- Create characterization tests for uncovered code paths
- Ensure test coverage meets or exceeds 85%

Output: files_modified list, characterization_tests_created list, test_results (all passing), behavior_preserved flag, structural_metrics comparison.

### Phase 2.5: Quality Validation

Agent: manager-quality subagent

Input: Both Phase 1 planning context and Phase 2 implementation results.

TRUST 5 validation checks:

- Tested: Tests exist and pass before changes. Test-driven design discipline maintained.
- Readable: Code follows project conventions and includes documentation.
- Unified: Implementation follows existing project patterns.
- Secured: No security vulnerabilities introduced. OWASP compliance verified.
- Trackable: All changes logged with clear commit messages. History analysis supported.

Additional validation:

- Test coverage at least 85%
- Behavior preservation: All existing tests pass unchanged
- Characterization tests pass: Behavior snapshots match
- Structural improvement: Coupling and cohesion metrics improved

Output: trust_5_validation results per pillar, coverage percentage, overall status (PASS, WARNING, or CRITICAL), and issues_found list.

### Quality Gate Decision

If status is CRITICAL:

- Present quality issues to user via AskUserQuestion
- Option to return to implementation phase for fixes
- Exit current execution flow

If status is PASS or WARNING: Continue to Phase 3.

### Phase 3: Git Operations (Conditional)

Agent: manager-git subagent

Input: Full context from Phases 1, 2, and 2.5.

Execution conditions:

- quality_status is PASS or WARNING
- If config git_strategy.automation.auto_branch is true: Create feature branch feature/SPEC-{ID}
- If auto_branch is false: Commit directly to current branch

Tasks for manager-git:

- Create feature branch (if auto_branch enabled)
- Stage all relevant implementation and test files
- Create commits with conventional commit messages
- Verify each commit was created successfully

Output: branch_name, commits array (sha and message), files_staged count, status.

### Phase 4: Completion and Guidance

Tool: AskUserQuestion (at orchestrator level)

Display implementation summary:

- Files created count
- Tests passing count
- Coverage percentage
- Commits count

Options:

- Sync Documentation (recommended): Execute /moai sync to synchronize docs and create PR
- Implement Another Feature: Return to /moai plan for additional SPEC
- Review Results: Examine implementation and test coverage locally
- Finish: Session complete

---

## Context Propagation

Context flows forward through every phase:

- Phase 1 to Phase 2: Execution plan with architecture decisions guides implementation
- Phase 2 to Phase 2.5: Implementation code plus planning context enables context-aware validation
- Phase 2.5 to Phase 3: Quality findings enable semantically meaningful commit messages

Benefits: No re-analysis between phases. Architectural decisions propagate naturally. Commits explain both what changed and why.

---

## Completion Criteria

All of the following must be verified:

- Phase 1: manager-strategy returned execution plan with requirements and success criteria
- User approval checkpoint blocked Phase 2 until user confirmed
- Phase 1.5: Tasks decomposed with requirement traceability
- Phase 2: manager-ddd executed complete DDD cycle with 85%+ coverage and behavior preserved
- Phase 2.5: manager-quality completed TRUST 5 validation with PASS or WARNING status
- Quality gate blocked Phase 3 if status was CRITICAL
- Phase 3: manager-git created commits (branch or direct) only if quality permitted
- Phase 4: User presented with next step options

---

Version: 1.0.0
Source: Extracted from .claude/commands/moai/2-run.md v5.0.0
