---
name: moai-workflow-run
description: >
  DDD/TDD implementation workflow for SPEC requirements. Second step
  of the Plan-Run-Sync workflow. Routes to manager-ddd or manager-tdd based
  on quality.yaml development_mode setting.
license: Apache-2.0
compatibility: Designed for Claude Code
user-invocable: false
metadata:
  version: "2.3.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-21"
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

Implement SPEC requirements using the configured development methodology. The methodology is determined by `development_mode` in quality.yaml:

- **ddd**: Domain-Driven Development using ANALYZE-PRESERVE-IMPROVE cycle (for existing projects with < 10% test coverage)
- **tdd**: Test-Driven Development using RED-GREEN-REFACTOR cycle (default for all development work)

This is the second step of the Plan-Run-Sync workflow.

## Scope

- Implements Step 3 of MoAI's 4-step workflow (Task Execution)
- Receives SPEC documents created by /moai plan
- Hands off to /moai sync for documentation and PR

## Input

- $ARGUMENTS: SPEC-ID to implement (e.g., SPEC-AUTH-001)
- Resume: Re-running /moai run SPEC-XXX resumes from last successful phase checkpoint
- --team: Enable team-based implementation (see team-run.md for parallel implementation team)

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

Before presenting options, verify the plan against these criteria:

- Proportionality: Is the plan proportional to the requirements? Flag plans with excessive abstraction layers, unnecessary patterns, or scope creep beyond SPEC requirements.
- Code Reuse: Has the plan identified existing code, libraries, or patterns that can be reused? Flag plans that reinvent existing functionality.
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

### Development Mode Routing

Before Phase 2, determine the development methodology by reading `.moai/config/sections/quality.yaml`:

**If development_mode is "ddd":**
- Route all tasks to manager-ddd subagent
- Use ANALYZE-PRESERVE-IMPROVE cycle
- Focus on behavior preservation with characterization tests

**If development_mode is "tdd":**
- Route all tasks to manager-tdd subagent
- Use RED-GREEN-REFACTOR cycle
- Focus on test-first development with specification tests

### Phase 2: Implementation (Mode-Dependent)

#### Phase 2A: DDD Implementation (for ddd mode)

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

Input: Approved execution plan from Phase 1 plus task decomposition from Phase 1.5.

The TDD cycle executes three stages:

- RED: Write specification tests that define expected behavior. Tests must fail initially (confirms they test something new).
- GREEN: Write minimal implementation code to make tests pass. Focus on correctness, not elegance.
- REFACTOR: Improve code structure while keeping tests green. Apply clean code principles.

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

Additional validation (mode-dependent):

For DDD mode:
- Test coverage at least 85%
- Behavior preservation: All existing tests pass unchanged
- Characterization tests pass: Behavior snapshots match
- Structural improvement: Coupling and cohesion metrics improved

For TDD mode:
- Test coverage at least 80% per commit (85% recommended for new code)
- Test-first discipline: No code written without failing test first
- All specification tests pass
- Clean code principles applied in REFACTOR phase

Output: trust_5_validation results per pillar, coverage percentage, overall status (PASS, WARNING, or CRITICAL), and issues_found list.

#### Extended Quality Checks

In addition to TRUST 5 validation, the following checks are performed:

Code Complexity Analysis:
- Function size: Flag functions exceeding 50 lines (suggest splitting)
- File size: Flag files exceeding 500 lines (suggest decomposition)
- Cyclomatic complexity: Flag functions with complexity > 10
- Nesting depth: Flag code with nesting > 4 levels

Dead Code Detection:
- Unused imports: Identify and flag unused import statements
- Unused functions: Identify exported functions with no callers
- Unused variables: Identify declared but unreferenced variables
- Orphaned files: Identify files not referenced by any other module

Side Effect Analysis:
- Caller impact: For each modified function, identify all callers and assess impact
- Interface changes: Flag signature changes that affect downstream consumers
- State mutations: Identify unexpected state changes in modified code paths
- Dependency chain: Trace changes through dependency graph to detect cascading effects

Code Reuse Opportunities:
- Duplication detection: Identify code blocks similar to existing utilities or helpers
- Library overlap: Check if implemented functionality exists in project dependencies
- Pattern consolidation: Suggest merging similar patterns across modified files
- Shared abstraction: Identify opportunities to extract common logic

### Quality Gate Decision

If status is CRITICAL:

- Present quality issues to user via AskUserQuestion
- Option to return to implementation phase for fixes
- Exit current execution flow

If status is PASS or WARNING: Continue to Phase 2.7.

### Phase 2.7: Post-Implementation Review (Optional)

Purpose: Multi-dimensional review iteration for high-quality output. Activated when quality status is WARNING or when --review flag is set.

Review dimensions (each executed via manager-quality subagent):

- Purpose alignment: Does the implementation fulfill the SPEC requirements without deviation?
- Improvement safety: Do the improvements introduce any new issues?
- Side effect verification: Are there unintended behavioral changes in related modules?
- Full change review: Comprehensive diff review of all modified files
- Dead code cleanup: Remove code made obsolete during implementation
- User flow validation: Verify end-to-end user workflows remain functional

Iteration behavior:
- Each review dimension generates findings with severity (critical, warning, suggestion)
- Critical findings trigger a fix cycle: delegate to appropriate expert agent, then re-review
- Warning findings are logged for user attention
- Maximum 3 review iterations to prevent infinite loops
- If all dimensions pass with no critical findings: Continue to Phase 3

Output: review_findings per dimension, iterations_completed count, final review status.

### Phase 2.8: MX Tag Update

Purpose: Update @MX code annotations for modified files.

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

## Team Mode Routing

When --team flag is provided or auto-selected, the run phase MUST switch to team orchestration:

1. Verify prerequisites: workflow.team.enabled == true AND CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 env var is set
2. If prerequisites met: Read workflows/team-run.md and execute the team workflow (TeamCreate with backend-dev + frontend-dev + tester + quality)
3. If prerequisites NOT met: Warn user with message "Team mode requires workflow.team.enabled: true in workflow.yaml and CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 env var" then fallback to standard sub-agent mode (manager-ddd/tdd based on development_mode)

Team composition: backend-dev (inherit) + frontend-dev (inherit) + tester (inherit) + quality (inherit, read-only)

For detailed team orchestration steps, see workflows/team-run.md.

---

## Context Propagation

Context flows forward through every phase:

- Phase 1 to Phase 2: Execution plan with architecture decisions guides implementation
- Phase 2 to Phase 2.5: Implementation code plus planning context enables context-aware validation
- Phase 2.5 to Phase 3: Quality findings enable semantically meaningful commit messages
- Phase 2 to /moai sync: Implementation divergence report enables accurate SPEC and project document updates

Benefits: No re-analysis between phases. Architectural decisions propagate naturally. Commits explain both what changed and why. Divergence tracking ensures sync phase can accurately update SPEC and project documents.

---

## Completion Criteria

All of the following must be verified:

- Phase 1: manager-strategy returned execution plan with requirements and success criteria
- User approval checkpoint blocked Phase 2 until user confirmed
- Phase 1.5: Tasks decomposed with requirement traceability
- Phase 2: Implementation completed according to development_mode:
  - DDD mode: manager-ddd executed ANALYZE-PRESERVE-IMPROVE with 85%+ coverage
  - TDD mode: manager-tdd executed RED-GREEN-REFACTOR with 85%+ coverage
- Phase 2.5: manager-quality completed TRUST 5 validation with PASS or WARNING status
- Quality gate blocked Phase 3 if status was CRITICAL
- Phase 3: manager-git created commits (branch or direct) only if quality permitted
- Phase 4: User presented with next step options

---

## Standalone: Code Review (/moai review)

Purpose: Comprehensive quality and security review of uncommitted or staged changes using git diff analysis. Invoked independently via `/moai review`.

### Input

- $ARGUMENTS: Optional flags
- --staged: Review only staged changes (git diff --cached)
- --branch [base]: Compare current branch against base (default: main)
- --security: Focus on security-specific review patterns only

### Phase 1: Collect Changes

- Default: Run `git diff` for all unstaged and staged changes
- If --staged: Run `git diff --cached` for staged changes only
- If --branch: Run `git diff <base>...HEAD` for branch comparison
- Parse diff output to extract: files changed, lines added/removed, change context

### Phase 2: Code Review Analysis

Agent: manager-quality subagent

Review each change at four severity levels:

- Critical: Security vulnerabilities, data loss risks, authentication bypasses
- High: Logic errors, race conditions, resource leaks, error handling gaps
- Medium: Code style violations, missing validation, incomplete error messages
- Low: Naming conventions, documentation gaps, minor improvements

For each finding, report:
- Severity level
- File path and line number
- Description of the issue
- Suggested fix

### Phase 2.5: @MX Tag Compliance Check

Verify @MX annotations on modified code:

- New exported functions without @MX:NOTE or @MX:ANCHOR → flag as Medium
- Functions with fan_in >= 5 missing @MX:ANCHOR → flag as High
- Dangerous patterns (goroutines, complexity >= 15) without @MX:WARN → flag as Medium
- Untested public functions without @MX:TODO → flag as Low
- Existing @MX tags on modified code that are now stale → flag as Low

Include MX compliance findings in the review report alongside code quality findings.

### Phase 3: Approval Recommendation

Based on findings, generate recommendation:

- Approve: Zero Critical or High findings
- Warning: Medium findings only (proceed with awareness)
- Block: One or more Critical or High findings (must fix before merge)

### Phase 4: Report

Display review summary:
- Total files reviewed
- Findings by severity
- Approval recommendation
- Specific action items for Critical/High findings

---

## Standalone: Test Coverage Analysis (/moai coverage)

Purpose: Analyze test coverage, identify gaps, and generate missing tests to meet coverage targets. Invoked independently via `/moai coverage`.

### Input

- $ARGUMENTS: Optional flags
- --target N: Override coverage threshold (default: from quality.yaml test_coverage_target)
- --file PATH: Analyze specific file or directory only
- --report: Report-only mode, no test generation

### Phase 1: Run Tests with Coverage

Execute language-specific test command with coverage output:
- Go: `go test -coverprofile=coverage.out ./...`
- Python: `pytest --cov --cov-report=json`
- TypeScript: `npx vitest run --coverage` or `npx jest --coverage`
- JavaScript: `npx vitest run --coverage` or `npx jest --coverage` or `npx c8 node`
- Rust: `cargo tarpaulin --out json`
- Java: `mvn jacoco:report` or `gradle jacocoTestReport`
- Kotlin: `gradle koverReport`
- C#: `dotnet test --collect:"XPlat Code Coverage"`
- Ruby: `COVERAGE=true bundle exec rspec` (with simplecov)
- PHP: `XDEBUG_MODE=coverage vendor/bin/phpunit --coverage-text`
- Elixir: `mix test --cover`
- C++: `gcov` / `lcov` or `cmake --build . --target coverage`
- Scala: `sbt coverage test coverageReport`
- R: `covr::package_coverage()` or `Rscript -e 'covr::report()'`
- Flutter/Dart: `flutter test --coverage`
- Swift: `swift test --enable-code-coverage`

Language auto-detection uses indicator files from sync.md Phase 0.5 language detection. If language not detected, ask user via AskUserQuestion.

Parse coverage report to extract per-file and per-function coverage percentages.

### Phase 2: Gap Analysis

Identify uncovered code:
- Functions with 0% coverage (completely untested)
- Functions below target threshold
- Critical paths (error handlers, security checks) without tests
- Recently modified code without corresponding test updates

Sort gaps by priority:
1. Public API functions without tests
2. Error handling paths
3. Complex functions (cyclomatic complexity > 5)
4. Recently changed code

### Phase 3: Test Generation (unless --report)

Agent: expert-testing subagent

For each identified gap (prioritized):
- Generate specification-based tests (not implementation-coupled)
- Include edge cases and error scenarios
- Follow project's existing test patterns and conventions
- Verify generated tests pass

### Phase 4: Before/After Comparison

Display coverage improvement:
- Before: N% overall, N% for target files
- After: N% overall, N% for target files
- Gap: N% remaining to target
- Files still needing attention

---

## Standalone: E2E Testing (/moai e2e)

Purpose: Create and execute end-to-end tests for critical user journeys using one of three supported testing approaches. Invoked independently via `/moai e2e`.

### Input

- $ARGUMENTS: Optional flags and test scope
- --record: Record test execution as GIF or video
- --url URL: Override base URL for testing
- --journey NAME: Run specific named journey only

### Phase 1: Testing Tool Selection

Tool: AskUserQuestion (at orchestrator level)

Present three E2E testing approaches to the user:

Option 1 - Claude-in-Chrome (Recommended):
Browser automation via Claude Code's built-in Chrome extension. No installation needed. Best for interactive testing, visual verification, and GIF recording. Uses MCP tools (navigate, form_input, read_page, get_page_text, gif_creator).

Option 2 - Playwright CLI:
Industry-standard headless browser testing via CLI. Fastest execution and lowest token consumption. Best for CI/CD integration and repeatable test suites. If not installed, offer to install: `npm init playwright@latest`.

Option 3 - Agent Browser:
AI-powered browser automation from Vercel Labs (https://github.com/vercel-labs/agent-browser). Best for complex multi-step user journeys with AI reasoning. If not installed, offer to install: `npm install agent-browser`.

After user selects a tool:
- Check if the selected tool is installed (for Playwright and Agent Browser)
- If not installed, ask user permission to install via AskUserQuestion
- If user approves, install using appropriate package manager
- Verify installation success before proceeding

### Phase 2: Skill Verification

Before proceeding, check if a corresponding skill exists for the selected tool:

- Claude-in-Chrome: Check for `moai-platform-chrome-e2e` or similar skill
- Playwright: Check for `moai-platform-playwright` or similar skill
- Agent Browser: Check for `moai-platform-agent-browser` or similar skill

If the skill does NOT exist:
1. Research the tool's official documentation (WebSearch + WebFetch)
2. Create a new skill definition following the agent skills standard (using builder-skill subagent)
3. The skill should include: setup instructions, test writing patterns, assertion methods, best practices

### Phase 3: Journey Identification

Agent: Explore subagent (or expert-testing subagent)

Identify critical user journeys from:
- SPEC acceptance criteria (if SPEC-ID provided)
- Application routes and navigation structure
- Form submissions and data flows
- Authentication and authorization flows

### Phase 4: Test Plan and Creation

Agent: expert-testing subagent

Create a test plan via AskUserQuestion showing:
- Identified user journeys (numbered list)
- Estimated test count per journey
- Recommended assertion strategy

For each approved journey, generate tests using the selected tool:

**Claude-in-Chrome tests:**
- mcp__claude-in-chrome__navigate for page navigation
- mcp__claude-in-chrome__form_input for form interactions
- mcp__claude-in-chrome__read_page for content verification
- mcp__claude-in-chrome__get_page_text for text assertions
- mcp__claude-in-chrome__gif_creator for recording (if --record)

**Playwright CLI tests:**
- Generate .spec.ts test files with Page Object Model pattern
- Use playwright test runner configuration
- Include screenshot and trace on failure
- Support headed and headless modes

**Agent Browser tests:**
- Generate test scripts using agent-browser API
- Leverage AI reasoning for dynamic content handling
- Include retry logic for flaky elements

### Phase 5: Test Execution and Results

Execute tests using the selected tool and report:
- Journeys tested: N passed, N failed
- Recordings/screenshots (if --record or on failure)
- Failed step details with browser state
- Recommendations for flaky test prevention
- Coverage of acceptance criteria (if SPEC-ID provided)

---

Version: 2.3.0
Updated: 2026-02-21
Source: Extracted from .claude/commands/moai/2-run.md v5.0.0. Added implementation divergence tracking, development_mode routing (ddd/tdd), team mode support, LSP quality gates, extended quality checks (code complexity, dead code, side effects, code reuse), plan proportionality validation, post-implementation review loop, and standalone workflows for code review, test coverage analysis, and E2E testing.
