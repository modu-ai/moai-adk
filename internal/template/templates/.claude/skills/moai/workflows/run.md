---
name: moai-workflow-run
description: >
  DDD/TDD implementation workflow for SPEC requirements. Second step
  of the Plan-Run-Sync workflow. Routes to manager-ddd or manager-tdd based
  on quality.yaml development_mode setting.
user-invocable: false
metadata:
  version: "2.6.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-23"
  tags: "run, implementation, ddd, tdd, spec"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["run", "implement", "build", "create", "develop", "code"]
  agents: ["manager-ddd", "manager-tdd", "manager-strategy", "manager-quality", "manager-git"]
  phases: ["run"]
---

# Run Workflow Orchestration

## Purpose

Implement SPEC requirements using the configured development methodology.

For methodology details (DDD ANALYZE-PRESERVE-IMPROVE and TDD RED-GREEN-REFACTOR cycles, success criteria, brownfield enhancement), see: @.claude/rules/moai/workflow/workflow-modes.md

## Scope

- Implements Step 3 of MoAI's 4-step workflow (Task Execution)
- Receives SPEC documents created by /moai plan
- Hands off to /moai sync for documentation and PR

## Input

- $ARGUMENTS: SPEC-ID to implement (e.g., SPEC-AUTH-001)
- Resume: Re-running /moai run SPEC-XXX resumes from last successful phase checkpoint
- --team: Enable team-based implementation (see team/run.md for parallel implementation team)

## Context Loading

Before execution, load these essential files:

- .moai/config/config.yaml (git strategy, automation settings)
- .moai/config/sections/quality.yaml (coverage targets, TRUST 5 settings, development_mode)
- .moai/config/sections/git-strategy.yaml (auto_branch, branch creation policy)
- .moai/config/sections/language.yaml (git_commit_messages setting)
- .moai/specs/SPEC-{ID}/ directory (spec.md, plan.md, acceptance.md)
- .moai/specs/SPEC-{ID}/progress.md (session resume context: if exists, load to identify completed phases and skip them; if absent, will be created at Phase 1 start)
- .moai/project/structure.md (architecture context for implementation decisions)
- .moai/project/tech.md (technology stack context)
- .moai/project/codemaps/ directory listing (architecture maps for dependency and module understanding)

Pre-execution commands: git status, git branch, git log, git diff.

### Resume Check

Before Phase 1, check if `.moai/specs/SPEC-{ID}/progress.md` exists:
- If it exists: Load content, identify last completed phase checkpoint, skip all completed phases, resume from the next pending phase. Log: "Resuming SPEC-{ID} from Phase {N}"
- If it does not exist: Create the file now with initial entry:
  ```
  ## SPEC-{ID} Progress

  - Started: {current timestamp}
  ```
- The progress.md file persists across sessions and enables seamless resume after interruption.

---

## Phase Sequence

All phases execute sequentially. Each phase receives outputs from all previous phases as context.

### Phase 1: Analysis and Planning

Agent: manager-strategy subagent

Input: SPEC document content from the provided SPEC-ID. If research.md exists in the SPEC directory (.moai/specs/SPEC-{ID}/research.md), include it as additional context for deeper understanding of the codebase architecture, reference implementations, and identified risks.

Tasks for manager-strategy:

- Read and fully analyze the SPEC document
- Extract requirements and success criteria
- Identify implementation phases and individual tasks
- Determine tech stack and dependencies required
- Estimate complexity and effort
- Create detailed execution strategy with phased approach

Output: Execution plan containing plan_summary, requirements list, success_criteria, and effort_estimate.

Implementation guard: [HARD] During Phase 1 (Analysis and Planning), the manager-strategy subagent MUST NOT write any implementation code. The explicit instruction "DO NOT implement any code — focus exclusively on analysis and planning" MUST be included in the agent prompt. This separation of thinking and execution prevents premature implementation and ensures the plan is reviewed before any code is written.

### Decision Point 1: Plan Approval

Tool: AskUserQuestion (at orchestrator level)

Before presenting options, verify the plan against these criteria:

- Proportionality: Is the plan proportional to the requirements? Flag plans with excessive abstraction layers, unnecessary patterns, or scope creep beyond SPEC requirements.
- Code Reuse: Has the plan identified existing code, libraries, or patterns that can be reused? Flag plans that reinvent existing functionality.
- Reference Implementations: Has the plan leveraged reference implementations from research.md? Flag plans that ignore available reference code in the codebase.
- Simplicity: Does the plan follow YAGNI (You Aren't Gonna Need It)? Flag speculative features not in the SPEC.

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
- Each task must be completable in a single DDD/TDD cycle
- Assign priority and dependencies for each task
- Generate task tracking entries for progress visibility
- Verify task coverage matches all SPEC requirements

Task structure for each decomposed task:

- Task ID: Sequential within SPEC (TASK-001, TASK-002, etc.)
- Description: Clear action statement
- Requirement Mapping: Which SPEC requirement it fulfills
- Dependencies: List of prerequisite tasks
- Acceptance Criteria: How to verify completion

Constraints: Decompose into atomic tasks where each task completes in a single DDD/TDD cycle. No artificial limit on task count. If the SPEC itself is too complex, consider splitting the SPEC.

Output: Task list with coverage_verified flag set to true.

### Phase 1.6: Acceptance Criteria Initialization (Failing Checklist)

Purpose: Convert all SPEC acceptance criteria into explicit pending TaskList entries. This creates a visible "failing checklist" — all items start as pending and are marked completed (passing) as implementation progresses, following the Harness Engineering pattern.

Action:
- Read spec.md acceptance criteria for SPEC-{ID}
- For each acceptance criterion, execute TaskCreate:
  - subject: `[AC-N] <acceptance criterion statement>`
  - description: Requirement reference, expected behavior, verification method
  - status: pending (starts as "failing")
- Verify all SPEC requirements are covered by at least one task

Output: TaskList populated with all acceptance criteria as pending items.

Progress update: Append to `.moai/specs/SPEC-{ID}/progress.md`:
```
- Phase 1.6 complete: {N} acceptance criteria registered as pending tasks
```

### Phase 1.7: File Structure Scaffolding

Purpose: Create empty file stubs for all planned new files before implementation begins. This prevents entropy by establishing structure before coding, following the Harness Engineering "Blueprint" pattern.

Condition: Execute only for planned new files (files that do not yet exist in the codebase). Skip if all planned files already exist (modification-only SPEC).

Action:
- Identify all planned new files from Phase 1.5 task decomposition
- For each planned new file that does not yet exist:
  - Create empty stub with minimal required structure matching the project's language conventions (e.g., package declaration for Go, module header for Python, empty class for TypeScript)
  - Do NOT add any implementation logic — stubs only
- After stub creation: Capture LSP baseline (this is the clean baseline before any implementation)

Output: List of stub files created, LSP baseline diagnostics captured.

Progress update: Append to `.moai/specs/SPEC-{ID}/progress.md`:
```
- Phase 1.7 complete: {N} stub files created, LSP baseline captured
```

### Phase 1.8: Pre-Implementation MX Context Scan

Purpose: Scan files that will be modified during implementation to build an MX context map for implementation agents.

**Scan Target:** All existing files listed in the task decomposition (from Phase 1.5).

**MX Context Extraction:**
- @MX:ANCHOR: Identify invariant contracts. Pass to implementation agents as "do not break" constraints with fan_in counts.
- @MX:WARN: Identify danger zones. Alert agents to approach these areas with extra caution.
- @MX:NOTE: Collect business logic context. Include in agent prompts for informed implementation.
- @MX:TODO: Match against SPEC requirements. If a TODO aligns with a task, the implementation resolves it.
- @MX:LEGACY: Identify legacy code without SPEC. Flag for careful handling during modifications.

**Output:** MX context map included in Phase 2 agent prompts. The map is structured per-file:
- file_path: list of tags with type, line, description, and constraints

**Skip Condition:** If target files do not exist (greenfield implementation), skip this phase.

See @.claude/rules/moai/workflow/mx-tag-protocol.md for tag type definitions.

### Batch Mode Decision [MANDATORY EVALUATION]

Before routing to Phase 2, MoAI MUST evaluate whether to use Skill("batch") for parallel implementation.

Evaluate ALL of the following conditions:

- Condition A: task_count >= 5 (from Phase 1.5 decomposition)
- Condition B: predicted_file_changes >= 10 (estimated from SPEC scope)
- Condition C: independent_tasks >= 3 (tasks with no inter-dependencies)

Decision:

- If ANY condition is met: Execute Skill("batch") directly. Batch mode replaces sequential Phase 2. Each batch unit executes one atomic task in an isolated git worktree, runs tests, and creates a PR. MoAI collects all PRs and merges in dependency order via manager-git.
- If NO condition is met: Continue to standard sequential Phase 2 below.

Batch execution instructions when triggered:
1. Provide Skill("batch") with the full task list from Phase 1.5, the SPEC document content, and the development_mode from quality.yaml
2. Each batch agent MUST follow the same DDD/TDD cycle defined in the Development Mode Routing section
3. After all batch agents complete, collect results and jump directly to Phase 2.5 (Quality Validation)

### Development Mode Routing

Before Phase 2, determine the development methodology by reading `.moai/config/sections/quality.yaml`:

**If development_mode is "ddd":**
- Route all tasks to manager-ddd subagent
- Use ANALYZE-PRESERVE-IMPROVE cycle (see @workflow-modes.md for details)

**If development_mode is "tdd":**
- Route all tasks to manager-tdd subagent
- Use RED-GREEN-REFACTOR cycle (see @workflow-modes.md for details)

### Phase 2: Implementation (Mode-Dependent)

#### Phase 2A: DDD Implementation (for ddd mode)

Agent: manager-ddd subagent

Input: Approved execution plan from Phase 1 plus task decomposition from Phase 1.5. Include `.moai/project/structure.md` and `.moai/project/tech.md` as onboarding context in the agent prompt so the implementation agent understands the project's architecture conventions before writing code.

Requirements:

- Initialize task tracking for progress across refactoring steps
- Execute the complete ANALYZE-PRESERVE-IMPROVE cycle
- Verify all existing tests pass after each transformation
- Create characterization tests for uncovered code paths
- Ensure test coverage meets or exceeds 85%

Output: files_modified list, characterization_tests_created list, test_results (all passing), behavior_preserved flag, structural_metrics comparison, implementation_divergence report.

Implementation Divergence Tracking:

The manager-ddd subagent must track deviations from the original SPEC plan during implementation:

- planned_files: Files listed in plan.md that were expected to be created or modified
- actual_files: Files actually created or modified during the DDD cycle
- additional_features: Features or capabilities implemented beyond the original SPEC scope (with rationale)
- scope_changes: Description of any scope adjustments made during implementation (expansions, deferrals, or substitutions)
- new_dependencies: Any new libraries, packages, or external dependencies introduced
- new_directories: Any new directory structures created

This divergence data is consumed by /moai sync for SPEC document updates and project document synchronization.

#### Phase 2B: TDD Implementation (for tdd mode)

Agent: manager-tdd subagent

Input: Approved execution plan from Phase 1 plus task decomposition from Phase 1.5. Include `.moai/project/structure.md` and `.moai/project/tech.md` as onboarding context in the agent prompt so the implementation agent understands the project's architecture conventions before writing code.

Requirements:

- Initialize task tracking for progress across TDD cycles
- Execute the complete RED-GREEN-REFACTOR cycle for each feature
- Write tests before implementation (test-first discipline)
- Ensure minimum 80% coverage per commit (85% recommended for new code)

Output: files_created list, specification_tests_created list, test_results (all passing), coverage percentage, refactoring_improvements list, implementation_divergence report.

Implementation Divergence Tracking:

The manager-tdd subagent must track deviations from the original SPEC plan during implementation:

- planned_files: Files listed in plan.md that were expected to be created
- actual_files: Files actually created during the TDD cycle
- additional_features: Features or capabilities implemented beyond the original SPEC scope (with rationale)
- scope_changes: Description of any scope adjustments made during implementation
- new_dependencies: Any new libraries, packages, or external dependencies introduced
- new_directories: Any new directory structures created

This divergence data is consumed by /moai sync for SPEC document updates and project document synchronization.

### Phase 2.5: Quality Validation

Agent: manager-quality subagent

Input: Both Phase 1 planning context and Phase 2 implementation results.

TRUST 5 validation checks:

- Tested: Tests exist and pass before changes. Test-driven design discipline maintained.
- Readable: Code follows project conventions and includes documentation.
- Unified: Implementation follows existing project patterns.
- Secured: No security vulnerabilities introduced. OWASP compliance verified.
- Trackable: All changes logged with clear commit messages. History analysis supported.

Output: trust_5_validation results per pillar, coverage percentage, overall status (PASS, WARNING, or CRITICAL), and issues_found list.

#### Extended Quality Checks

Code Complexity Analysis:
- Function size: Flag functions exceeding 50 lines (suggest splitting)
- File size: Flag files exceeding 500 lines (suggest decomposition)
- Cyclomatic complexity: Flag functions with complexity > 10
- Nesting depth: Flag code with nesting > 4 levels

Dead Code Detection:
- Unused imports, functions, variables, and orphaned files
- Auto-removal: When confirmed, delegate to clean workflow (workflows/clean.md)

Side Effect Analysis:
- Caller impact: For each modified function, identify all callers and assess impact
- Interface changes: Flag signature changes that affect downstream consumers
- State mutations: Identify unexpected state changes in modified code paths
- Dependency chain: Trace changes through dependency graph to detect cascading effects

Code Reuse Opportunities:
- Duplication detection, library overlap, pattern consolidation, shared abstraction

### Quality Gate Decision

If status is CRITICAL:
- Present quality issues to user via AskUserQuestion
- Option to return to implementation phase for fixes
- Exit current execution flow

If coverage is below target (quality.yaml test_coverage_target):
- Auto-route to coverage workflow (workflows/coverage.md)
- Re-run quality validation after coverage improvement

If status is PASS or WARNING: Continue to Phase 2.7.

### Phase 2.7: Post-Implementation Review (Optional)

Purpose: Multi-dimensional review iteration for high-quality output. Activated when quality status is WARNING or when --review flag is set.

Review dimensions (each executed via manager-quality subagent):
- Purpose alignment, improvement safety, side effect verification, full change review, dead code cleanup, user flow validation

Extended review (when --review flag is set or changes affect security/performance/UX domains):
- Delegate to review workflow (workflows/review.md) for multi-perspective analysis

Iteration behavior:
- Each review dimension generates findings with severity (critical, warning, suggestion)
- Critical findings trigger a fix cycle: delegate to appropriate expert agent, then re-review
- Maximum 3 review iterations to prevent infinite loops
- If all dimensions pass with no critical findings: Continue to Phase 3

Output: review_findings per dimension, iterations_completed count, final review status.

### Phase 2.8: MX Tag Update

Purpose: Update @MX code annotations for modified files. See @.claude/rules/moai/workflow/mx-tag-protocol.md for tag rules.

**TDD Mode:**
- Remove `@MX:TODO` tags for tests that now pass
- Add `@MX:NOTE` for complex logic added during GREEN phase
- Review `@MX:WARN` tags if dangerous patterns were improved

**DDD Mode:**
- Run 3-Pass scan if codebase has zero @MX tags
- Update `@MX:ANCHOR` tags if fan_in changed
- Add `@MX:NOTE` for business rules discovered during ANALYZE
- Convert `@MX:LEGACY` to `@MX:SPEC` if SPEC retroactively created

Output: MX_TAG_REPORT with tags added, updated, removed by type.

### Phase 2.9: Simplify Pass [MANDATORY]

Purpose: Apply a parallel quality pass to all files modified during implementation. This phase is ALWAYS executed after Phase 2.8 — it is not optional.

Action: MoAI MUST call Skill("simplify") at this phase. Do not delegate to a subagent. Call it directly.

Skill("simplify") will:
- Use parallel agents to review all modified files for reuse opportunities, quality issues, and efficiency improvements
- Enforce CLAUDE.md coding standards compliance
- Fix discovered issues automatically

Scope: Only files listed in the implementation output (files_created + files_modified). Do not run on unrelated files.

Output: simplify_report with files_improved count, issues_fixed list, and compliance_status.

If simplify_report contains remaining unfixed issues:
- Include them in the quality findings passed to Phase 2.5 re-evaluation
- Do NOT block progress for suggestion-level issues; only block for critical issues

### LSP Quality Gates

The run phase enforces LSP-based quality gates as configured in quality.yaml:
- Zero LSP errors required (lsp_quality_gates.run.max_errors: 0)
- Zero type errors required (lsp_quality_gates.run.max_type_errors: 0)
- Zero lint errors required (lsp_quality_gates.run.max_lint_errors: 0)
- No regression from baseline allowed (lsp_quality_gates.run.allow_regression: false)

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
- Files created count, tests passing count, coverage percentage, commits count

Options:
- Sync Documentation (recommended): Execute /moai sync to synchronize docs and create PR
- Implement Another Feature: Return to /moai plan for additional SPEC
- Review Results: Examine implementation and test coverage locally
- Finish: Session complete

---

## Execution Mode Gate Integration

When the run phase is invoked from plan.md Decision Point 3.5 or moai.md Phase 11.5, the gate passes these parameters:
- `execution_mode`: worktree | team | sub-agent
- `active_mode`: cc | glm | cg
- `tmux_available`: true | false

**If execution_mode == "worktree":**
This run invocation is already inside the isolated tmux session and worktree.
Proceed with standard sub-agent run phase in the current environment.
No additional routing needed — CC/GLM/CG env is already configured by the Gate.

**If execution_mode == "team":**
Apply Team Mode Routing below. The active_mode determines worker model selection:
- CC: All teammates use Claude (default behavior)
- GLM: All teammates inherit GLM env from tmux session
- CG: Leader=Claude (clean session), Workers=GLM (tmux env injected)

**If execution_mode == "sub-agent":**
Skip Team Mode Routing. Proceed directly to Phase 1 (Strategy).

**If no execution_mode provided (direct `/moai run` invocation):**
Apply existing --team/--solo flag logic in Team Mode Routing below.

---

## Team Mode Routing

When --team flag is provided or auto-selected, the run phase MUST switch to team orchestration:

1. Verify prerequisites: workflow.team.enabled == true AND CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 env var is set
2. If prerequisites met: Read team/run.md and execute the team workflow (TeamCreate with backend-dev + frontend-dev + tester + quality)
3. If prerequisites NOT met: Warn user then fallback to standard sub-agent mode

Team composition: backend-dev (inherit) + frontend-dev (inherit) + tester (inherit) + quality (inherit, read-only)

### Worktree Isolation [HARD]

- [HARD] Implementation teammates (backend-dev, frontend-dev, tester) MUST use `isolation: "worktree"` when spawned via Task()
- [HARD] Read-only teammates (quality) MUST NOT use `isolation: "worktree"` — permissionMode: plan is sufficient
- After team shutdown, run `git worktree prune` to clean up stale worktree references

See @.claude/rules/moai/workflow/worktree-integration.md for the complete worktree decision tree.

For detailed team orchestration steps, see team/run.md.

---

## Context Propagation

Context flows forward through every phase:

- Phase 1 to Phase 2: Execution plan with architecture decisions guides implementation
- Phase 2 to Phase 2.5: Implementation code plus planning context enables context-aware validation
- Phase 2.5 to Phase 3: Quality findings enable semantically meaningful commit messages
- Phase 2 to /moai sync: Implementation divergence report enables accurate SPEC and project document updates

---

## Completion Criteria

All of the following must be verified:

- Phase 1: manager-strategy returned execution plan with requirements and success criteria
- User approval checkpoint blocked Phase 2 until user confirmed
- Phase 1.5: Tasks decomposed with requirement traceability
- Phase 1.8: MX context map built for target files (skipped for greenfield)
- Phase 2: Implementation completed according to development_mode (with MX context)
- Phase 2.5: manager-quality completed TRUST 5 validation with PASS or WARNING status
- Quality gate blocked Phase 3 if status was CRITICAL
- Phase 3: manager-git created commits (branch or direct) only if quality permitted
- Phase 4: User presented with next step options

---

Version: 2.9.0
Updated: 2026-03-02. Added Harness Engineering improvements: progress.md resume (P1), failing checklist Phase 1.6 (P2), file scaffolding Phase 1.7 (P5), onboarding context propagation to implementation agents (P3).
