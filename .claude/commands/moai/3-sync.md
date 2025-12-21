---
name: moai:3-sync
description: "Synchronize documentation with Phase 0.5 quality verification and finalize PR"
argument-hint: "Mode target path - Mode: auto (default)|force|status|project, target path: Synchronization target path"
allowed-tools: Task, AskUserQuestion, TodoWrite
model: inherit
---

##  Pre-execution Context

!git status --porcelain
!git diff --name-only HEAD
!git branch --show-current
!git log --oneline -10
!find .moai/specs -name "spec.md" -type f 2>/dev/null

##  Essential Files

@.moai/config/config.yaml
@.moai/specs/
@README.md

---

#  MoAI-ADK Step 3: Document Synchronization (+Optional PR Ready)

User Interaction Architecture: AskUserQuestion must be used at COMMAND level only. Subagents via Task() are stateless and cannot interact with users. Collect all approvals BEFORE delegating to agents.

Batched Design: All AskUserQuestion calls follow batched design principles (1-4 questions per call, max 4 options per question). See CLAUDE.md section "User Interaction Architecture" for details.

4-Step Workflow Integration: This command implements Step 4 of Alfred's workflow (Report & Commit with conditional report generation). See CLAUDE.md for full workflow details.

---

##  Command Purpose

[HARD] This command orchestrates ONLY - delegates all sync work to manager-docs agent

Document sync target: $ARGUMENTS

Agent Delegation Pattern:

Correct Approach [HARD]:
- Invoke Task() with subagent_type="manager-docs"
- Pass complete context including changed files and verification results
- Let manager-docs agent determine implementation strategy
- WHY: Specialized agent has domain expertise and handles all complexity

Forbidden Approach:
- Direct file manipulation using Edit, Write, Read tools
- Direct bash execution of document updates
- WHY: Bypasses quality controls and loses context-specific error handling

Standard Workflow Sequence:
Step 1 (Analysis & Planning) leads to Step 2 (Document Sync via Agent) leads to Step 3 (Git Commit & PR)

---

##  Execution Modes

This command supports 4 operational modes:

| Mode               | Scope                   | PR Processing         | Use Case                            |
| ------------------ | ----------------------- | --------------------- | ----------------------------------- |
| auto (default) | Smart selective sync    | PR Ready conversion   | Daily development workflow          |
| force          | Full project re-sync    | Full regeneration     | Error recovery, major refactoring   |
| status         | Status check only       | Report only           | Quick health check                  |
| project        | Integrated project-wide | Project-level updates | Milestone completion, periodic sync |

Command usage examples:

- `/moai:3-sync` → Auto-sync (PR Ready only)
- `/moai:3-sync --auto-merge` → PR auto-merge + branch cleanup
- `/moai:3-sync force` → Force full synchronization
- `/moai:3-sync status` → Check synchronization status
- `/moai:3-sync project` → Integrated project synchronization
- `/moai:3-sync auto src/auth/` → Specific path synchronization
- `/moai:3-sync --worktree` → Sync in worktree mode (with worktree exit options)
- `/moai:3-sync --branch` → Sync in branch mode (with branch management options)

---

##  Associated Agents

| Agent           | Purpose                                         |
| --------------- | ----------------------------------------------- |
| manager-docs    | Synchronize Living Documents with code changes  |
| manager-quality | Verify project integrity and TRUST 5 compliance |
| manager-git     | Handle Git operations and commit management     |

---

## Agent Invocation Patterns (CLAUDE.md Compliance)

This command uses agent execution patterns defined in CLAUDE.md (lines 96-120).

### Sequential Phase-Based Chaining PASS

Command implements sequential chaining through 3 core phases:

Phase Flow:
- Phase 1: Analysis & Planning (manager-docs analyzes changed files and sync scope)
- Phase 2: Execute Sync (manager-docs updates documentation, manager-quality validates)
- Phase 3: Git Operations & PR (manager-git creates commits and prepares PR if applicable)

Each phase receives outputs from previous phases as context.

WHY: Sequential execution ensures documentation consistency and validation
- Phase 2 requires analysis results from Phase 1 to determine sync scope
- Phase 3 requires validated documentation from Phase 2 before commit
- PR creation requires successful commit from Phase 3

IMPACT: Skipping phases would create inconsistent documentation or invalid commits

### Parallel Execution WARNING

Limited parallel execution within Phase 2 for independent documentation files

WHY: Some documentation files can be generated simultaneously
- Multiple markdown files without cross-references can be updated in parallel
- Index updates and validation must remain sequential

IMPACT: Full parallel execution would risk broken cross-references and index inconsistencies

### Resumable Agent Support FAIL

Not applicable - command typically completes quickly in single execution

WHY: Documentation sync is fast operation (usually under 2 minutes)
- Most sync operations complete in first attempt
- File system operations are atomic and recoverable
- No long-running processes requiring checkpoints

IMPACT: Resume pattern unnecessary for typical sync workflows

---

Refer to CLAUDE.md "Agent Chaining Patterns" (lines 96-120) for complete pattern architecture.

---

##  Execution Philosophy: "Sync → Verify → Commit"

`/moai:3-sync` performs documentation synchronization through complete agent delegation:

```
User Command: /moai:3-sync [mode] [path]
    ↓
/moai:3-sync Command
    └─ Task(subagent_type="manager-docs" or "manager-quality" or "manager-git")
        ├─ Phase 1: Analysis & Planning (manager-docs)
        ├─ Phase 2: Execute Sync (manager-docs + manager-quality)
        └─ Phase 3: Git Operations & PR (manager-git)
            ↓
        Output: Synchronized docs + commit + PR Ready (conditional)
```

### Key Principle: Zero Direct Tool Usage

[HARD] This command uses ONLY Task(), AskUserQuestion(), and TodoWrite():

Permitted Tools:
- Task() for orchestration [HARD]
- AskUserQuestion() for user interaction AT COMMAND LEVEL ONLY [HARD]
  - WHY: Subagents via Task() are stateless and cannot interact with users
  - CORRECT: Collect approvals before Task() calls, pass choices as parameters
- TodoWrite() for progress tracking [HARD]

Forbidden Tools (All delegated to agents):
- Read (delegated) - All file reading operations must be performed by specialized agents
- Write (delegated) - All file writing operations must be performed by specialized agents
- Edit (delegated) - All file editing operations must be performed by specialized agents
- Bash (delegated) - All bash execution must be performed by specialized agents

WHY: Zero direct tool usage maintains clean separation of concerns. Each tool type has a specialized agent that understands context-specific requirements.

IMPACT: Direct tool usage would bypass quality controls, specialized agent expertise, and error recovery mechanisms. Delegation ensures consistent execution patterns.

All complexity is handled by specialized agents (manager-docs, manager-quality, manager-git).

---

---

##  Output Format

All command execution outputs must use semantic XML sections for clarity and consistency:

XML Structure Format:
```
<analysis>Detailed assessment of project state, identified changes, and validation results</analysis>
<plan>Synchronization strategy including scope, affected documents, and approach rationale</plan>
<execution>Concrete actions taken: files updated, reports generated, status changes recorded</execution>
<verification>Quality validation results: TRUST 5 compliance, link integrity, consistency checks</verification>
<completion>Summary of outcomes, generated reports locations, and next steps for user</completion>
```

Required Elements:
- Analysis must detail all findings from project validation and Git analysis
- Plan must explain strategy choice including WHY and IMPACT of decisions
- Execution must track all agent actions and file modifications
- Verification must report all quality gates and their outcomes
- Completion must guide user toward next meaningful action

WHY: XML sections provide machine-parseable structure and enable audit trails. Clear sections ensure user can understand command progress at any point.

IMPACT: Unstructured output reduces comprehension and prevents automated processing of results.

---

##  OVERALL WORKFLOW STRUCTURE

```
┌──────────────────────────────────────────────────────────┐
│ PHASE 1: Analysis & Planning (tag-agent + manager-docs)│
│  - Verify prerequisites                                  │
│  - Analyze project status (Git + SPEC)                    │
│  - Execute quality verification (Phase 0.5)              │
│  - Request user approval                                 │
└──────────────────────────────────────────────────────────┘
                          ↓
          ┌───────────────┴───────────────┐
          │                               │
     User approves                   User aborts
          │                               │
          ↓                               ↓
┌─────────────────────────┐   ┌──────────────────────┐
│ PHASE 2: Execute Sync   │   │ PHASE 4: Graceful    │
│ (manager-docs+quality) │   │ Exit (no changes)    │
│  - Create backup        │   └──────────────────────┘
│  - Sync documents       │
│  - Verify SPECs          │
└─────────────────────────┘
          ↓
┌──────────────────────────────────────────────────────────┐
│ PHASE 3: Git Operations & PR (manager-git)                  │
│  - Commit document changes                               │
│  - Transition PR (Team mode)                             │
│  - Auto-merge (if requested)                             │
│  - Branch cleanup                                        │
│  - Next steps guidance                                   │
└──────────────────────────────────────────────────────────┘
```

---

##  PHASE 1: Analysis & Planning

Goal: Gather project context, verify project status, and get user approval.

### Step 1.1: Verify Prerequisites & Load Skills

Execute these verification steps:

1. [HARD] TUI System Must Be Ready:

   Requirement: Interactive menus must be available for all user interactions
   WHY: User interaction requires responsive feedback mechanism
   IMPACT: Missing TUI will cause command to fail during user approval phase

2. [HARD] MoAI-ADK Structure Must Exist:

   Requirement: `.moai/` directory must exist in project root
   Requirement: `.claude/` directory must exist in project root
   Failure Action: Print error message with missing directory details and exit immediately
   WHY: These directories contain essential project metadata and configuration
   IMPACT: Missing directories indicates incomplete project initialization

3. [HARD] Git Repository Must Be Initialized:

   Requirement: Project must be inside a Git repository
   Verification: Execute `git rev-parse --is-inside-work-tree` and verify output is "true"
   Failure Action: Print error message and exit immediately
   WHY: Document synchronization requires Git history and commit operations
   IMPACT: Non-Git projects cannot track documentation changes or create commits

4. [SOFT] Python Environment Should Be Available:

   Requirement: Python 3 should be installed and accessible
   Verification: Execute `which python3` and verify command succeeds
   Failure Action: Print warning message but continue with reduced functionality
   WHY: Python environment enables advanced automation and validation
   IMPACT: Missing Python limits automation but does not prevent basic sync operations

Result: All hard prerequisites verified. All soft requirements checked. Command may proceed to next step.

---

### Step 1.2: Analyze Project Status

Gather context for synchronization planning:

1. [HARD] Analyze Git Changes:

   Requirement: Determine all files that have changed since last commit
   Actions:
   - Execute: `git status --porcelain` to identify modified files
   - Execute: `git diff --name-only HEAD` to list changed file paths
   - Categorize: Count Python files, test files, documents, SPEC files separately
   WHY: Git state determines synchronization scope and impact analysis
   IMPACT: Incomplete Git analysis leads to partial synchronization

2. [HARD] Read Project Configuration:

   Requirement: Load essential project settings from configuration file
   Actions:
   - Read: `.moai/config.yaml` file
   - Extract: `git_strategy.mode` value (must be Personal or Team)
   - Extract: `language.conversation_language` value (determines document language)
   - Extract: `git_strategy.spec_git_workflow` value
   WHY: Configuration drives workflow behavior and language output
   IMPACT: Missing configuration values cause workflow misalignment and language errors

3. [HARD] Determine Synchronization Mode:

   Requirement: Parse arguments to establish execution mode
   Actions:
   - Parse $ARGUMENTS for mode value: must be one of auto, force, status, or project
   - If empty: Default to auto mode
   - Parse optional flags: --auto-merge, --skip-pre-check, --skip-quality-check
   - Parse special flags: --worktree, --branch
   WHY: Mode determines scope and processing strategy
   IMPACT: Invalid mode selection produces incorrect synchronization behavior

4. [SOFT] Handle Worktree Detection:

   Requirement: Identify if execution occurs within a Git worktree
   Actions:
   - Execute: `git rev-parse --git-dir` to locate git directory
   - Analyze: Check if git directory path contains `worktrees/` component
   - If in worktree: Extract SPEC ID from current path (format: SPEC-{DOMAIN}-{NUMBER})
   - Alternative: Check worktree registry at `~/worktrees/{PROJECT_NAME}/.moai-worktree-registry.json`
   - Store: `$WORKTREE_MODE=true` and `$CURRENT_SPEC_ID` for later use
   WHY: Worktree context enables specialized cleanup and workflow options
   IMPACT: Missing worktree detection prevents proper exit handling but does not block sync

5. [SOFT] Handle Branch Detection:

   Requirement: Identify if execution occurs on feature branch
   Actions:
   - Check: Is --branch flag present in arguments OR is current branch not main
   - Execute: `git branch --show-current` to get current branch name
   - Store: `$BRANCH_MODE=true` and `$CURRENT_BRANCH` for later use
   WHY: Branch context enables proper merge and cleanup operations
   IMPACT: Missing branch detection reduces workflow automation but does not prevent sync

6. [HARD] Handle Status Mode Early Exit:

   Requirement: For status mode, provide quick health report and exit
   Actions:
   - Check: If mode is status value
   - Report: Current project health assessment
   - Report: Count of changed files
   - Report: Synchronization recommendation
   - If status mode: Execute Phase 0.5 quality verification before reporting
   - Exit: Command completes with success code (no further phases)
   WHY: Status mode serves quick health check without making changes, but still validates project quality
   IMPACT: Skipping quality check in status mode reduces visibility of project health

Result: Project context gathered. Synchronization mode established. Ready for quality verification (Phase 0.5).

---

### Step 1.2.5: Phase 0.5 Quality Verification

[HARD] Execute quality verification pipeline before document synchronization begins.

WHY: Quality checks prevent synchronizing broken states. Early validation catches issues before expensive document updates.
IMPACT: Skipping quality checks risks propagating issues to documentation.

Quality Verification Pipeline:

This phase automatically detects the project language and runs appropriate quality tools:

```
Detect Project Language
    ↓
Language-specific tool execution
    ├── Test Runner (pytest/jest/go test/cargo test/etc.)
    ├── Linter (ruff/eslint/golangci-lint/clippy/etc.)
    └── Type Checker (mypy/tsc/go vet/etc.)
    ↓
code-review invocation (manager-quality)
    ↓
Quality Report
```

---

#### Task 1.2.5.1: Detect Project Language

Automatically detect the primary project language by checking indicator files:

Language Detection Rules (check in order, first match wins):

1. Python Detection:
   - Indicator files: pyproject.toml, setup.py, requirements.txt, .python-version, Pipfile
   - IF found: Store PROJECT_LANGUAGE as python

2. TypeScript/JavaScript Detection:
   - Indicator files: tsconfig.json, package.json (with typescript dependency), .nvmrc
   - IF tsconfig.json exists: Store PROJECT_LANGUAGE as typescript
   - IF package.json exists without tsconfig: Store PROJECT_LANGUAGE as javascript

3. Go Detection:
   - Indicator files: go.mod, go.sum
   - IF found: Store PROJECT_LANGUAGE as go

4. Rust Detection:
   - Indicator files: Cargo.toml, Cargo.lock
   - IF found: Store PROJECT_LANGUAGE as rust

5. Ruby Detection:
   - Indicator files: Gemfile, .ruby-version, Rakefile
   - IF found: Store PROJECT_LANGUAGE as ruby

6. Java Detection:
   - Indicator files: pom.xml, build.gradle, build.gradle.kts
   - IF found: Store PROJECT_LANGUAGE as java

7. PHP Detection:
   - Indicator files: composer.json, composer.lock
   - IF found: Store PROJECT_LANGUAGE as php

8. Kotlin Detection:
   - Indicator files: build.gradle.kts (with kotlin plugin), settings.gradle.kts
   - IF found with kotlin references: Store PROJECT_LANGUAGE as kotlin

9. Swift Detection:
   - Indicator files: Package.swift, .xcodeproj, .xcworkspace
   - IF found: Store PROJECT_LANGUAGE as swift

10. C#/.NET Detection:
    - Indicator files: .csproj, .sln, .fsproj
    - IF found: Store PROJECT_LANGUAGE as csharp

11. C++ Detection:
    - Indicator files: CMakeLists.txt, .cpp, .hpp, Makefile (with C++ content)
    - IF found: Store PROJECT_LANGUAGE as cpp

12. Elixir Detection:
    - Indicator files: mix.exs, .exs, .ex
    - IF found: Store PROJECT_LANGUAGE as elixir

13. R Detection:
    - Indicator files: DESCRIPTION (R package), .Rproj, renv.lock
    - IF found: Store PROJECT_LANGUAGE as r

14. Flutter/Dart Detection:
    - Indicator files: pubspec.yaml, .dart
    - IF found: Store PROJECT_LANGUAGE as flutter

15. Scala Detection:
    - Indicator files: build.sbt, .scala, build.sc (mill)
    - IF found: Store PROJECT_LANGUAGE as scala

16. Fallback:
    - IF no indicators found: Store PROJECT_LANGUAGE as unknown
    - Skip language-specific tools, proceed directly to code-review

Log detection result:
- Print to user: "Project language detection: [detected_language]"

---

#### Task 1.2.5.2: Execute Language-Specific Test Runner

Execute the appropriate test runner based on detected language:

Test Runner Configuration by Language:

- Python: Execute pytest -v --tb=short or python -m unittest discover
- TypeScript/JavaScript: Execute npm test or npx vitest run or npx jest
- Go: Execute go test ./... -v
- Rust: Execute cargo test
- Ruby: Execute bundle exec rspec or rake test
- Java: Execute mvn test or gradle test
- PHP: Execute vendor/bin/phpunit or composer test
- Kotlin: Execute gradle test
- Swift: Execute swift test
- C#/.NET: Execute dotnet test
- C++: Execute ctest or catch2 or gtest
- Elixir: Execute mix test
- R: Execute testthat or devtools::test()
- Flutter/Dart: Execute flutter test or dart test
- Scala: Execute sbt test or mill test

Execution Steps:

1. Check tool availability:
   - Verify the test runner is installed
   - IF not installed: Print warning "[test runner] not installed - skipping" and proceed to linter
   - IF installed: Continue to execution

2. Execute test runner:
   - Run the appropriate command for detected language
   - Capture all output including test results and failures

3. Parse test results:
   - Count passed tests
   - Count failed tests
   - Identify failed test names

4. Determine test status:
   - IF all tests passed: Store TEST_STATUS as PASS
   - IF any tests failed: Store TEST_STATUS as FAIL
   - IF tool not installed: Store TEST_STATUS as WARN

5. IF test failure occurs:
   - Ask user for decision using AskUserQuestion
   - Question: "Test execution result: X tests failed. How would you like to proceed?"
   - Header: "Quality Verification"
   - Options:
     - "Continue" - Ignore test failures and proceed with sync
     - "Abort" - Stop sync and fix tests first
   - IF user chooses "Continue": Set flag and proceed to linter
   - IF user chooses "Abort": Exit command with graceful exit message (Phase 4)

---

#### Task 1.2.5.3: Execute Language-Specific Linter

Execute the appropriate linter based on detected language:

Linter Configuration by Language:

- Python: Execute ruff check . or flake8 .
- TypeScript/JavaScript: Execute npx eslint . or npm run lint
- Go: Execute golangci-lint run or go vet ./...
- Rust: Execute cargo clippy
- Ruby: Execute bundle exec rubocop or rubocop
- Java: Execute mvn checkstyle:check or gradle checkstyleMain
- PHP: Execute vendor/bin/phpstan analyse or php-cs-fixer fix --dry-run
- Kotlin: Execute gradle ktlintCheck or detekt
- Swift: Execute swiftlint
- C#/.NET: Execute dotnet format --verify-no-changes
- C++: Execute clang-tidy or cppcheck
- Elixir: Execute mix credo
- R: Execute lintr
- Flutter/Dart: Execute dart analyze
- Scala: Execute scalafmt --check or scalafix

Execution Steps:

1. Check tool availability:
   - Verify the linter is installed
   - IF not installed: Print warning "[linter] not installed - skipping" and proceed to type checker
   - IF installed: Continue to execution

2. Execute linter:
   - Run the appropriate command for detected language
   - Capture all output including violations found

3. Parse linter results:
   - Count total violations
   - Categorize by severity if available
   - Identify files with violations

4. Determine linter status:
   - IF no violations: Store LINT_STATUS as PASS
   - IF violations found: Store LINT_STATUS as FAIL
   - IF tool not installed: Store LINT_STATUS as WARN

5. Report findings:
   - Print to user: "[linter name] check: [result] - X violations found"

---

#### Task 1.2.5.4: Execute Language-Specific Type Checker

Execute the appropriate type checker based on detected language:

Type Checker Configuration by Language:

- Python: Execute mypy . or pyright .
- TypeScript: Execute npx tsc --noEmit
- JavaScript: Skip (no native type checking) or use TypeScript in check mode
- Go: Execute go vet ./... (already includes type checking)
- Rust: Skip (type checking is part of cargo build/test)
- Ruby: Execute bundle exec steep check or sorbet tc
- Java: Skip (compilation handles type checking)
- PHP: Execute vendor/bin/phpstan analyse --level max
- Kotlin: Skip (compilation handles type checking)
- Swift: Skip (compilation handles type checking)
- C#/.NET: Execute dotnet build --no-restore (type check via compilation)
- C++: Skip (compilation handles type checking)
- Elixir: Execute mix dialyzer (optional, requires dialyxir dependency)
- R: Skip (no native type checking)
- Flutter/Dart: Execute dart analyze (includes type checking)
- Scala: Skip (compilation handles type checking)

Execution Steps:

1. Check if type checking is applicable:
   - Some languages include type checking in compilation
   - IF not applicable: Print info "[language] uses compilation for type checking - skipping separate check"
   - IF applicable: Continue to tool check

2. Check tool availability:
   - Verify the type checker is installed
   - IF not installed: Print warning "[type checker] not installed - skipping" and proceed to code-review
   - IF installed: Continue to execution

3. Execute type checker:
   - Run the appropriate command for detected language
   - Capture all output including type errors

4. Parse type checker results:
   - Count total type errors
   - Identify files with type issues

5. Determine type checker status:
   - IF no errors: Store TYPE_STATUS as PASS
   - IF errors found: Store TYPE_STATUS as FAIL
   - IF tool not installed or not applicable: Store TYPE_STATUS as SKIP

6. Report findings:
   - Print to user: "[type checker] check: [result] - X type errors found"

---

#### Task 1.2.5.5: Invoke manager-quality for Code Review

Execute code-review agent regardless of project language:

Use the manager-quality subagent to:

Perform automated code review and TRUST 5 validation

Critical Language Configuration:

- Receive instructions in agent_prompt_language from config (default: English)
- Respond in conversation_language from config (user's preferred language)

Task Instructions:

1. Execute TRUST 5 quality validation:
   - Test: Verify test coverage is adequate
   - Readable: Check code readability and standards
   - Understandable: Verify code organization and documentation
   - Secure: Check for security issues and best practices
   - Tagged: Verify TAG annotations are present (if applicable)

2. Generate quality report:
   - Summary of findings
   - List of issues found (if any)
   - Recommendations for improvements

3. Determine overall code review status:
   - IF no issues found: Store CODE_REVIEW_STATUS as PASS
   - IF issues found: Store CODE_REVIEW_STATUS as WARN

Store: Response in CODE_REVIEW_RESULTS

---

#### Task 1.2.5.6: Generate Quality Report

After all checks complete, generate comprehensive quality report:

1. Create quality report output:

   ```
   Phase 0.5 Quality Verification Results
   =====================================

   Project Language: [detected_language]

   Tool Verification Results:
   test-runner:  [PASS|FAIL|WARN|SKIP] ([tool_name])
   linter:       [PASS|FAIL|WARN|SKIP] ([tool_name])
   type-checker: [PASS|FAIL|WARN|SKIP] ([tool_name])
   code-review:  [PASS|WARN]

   Details:
   - test-runner: [summary]
   - linter: [summary]
   - type-checker: [summary]
   - code-review: [summary]

   Overall Status: [PASS|WARN]

   Next Step: Proceeding to document synchronization
   ```

2. Store report:
   - Save complete report in QUALITY_REPORT
   - Print summary to user

3. Determine phase outcome:
   - IF test runner user chose "Abort": Go to PHASE 4 (exit)
   - IF all checks complete: Continue to Step 1.3

Result: Quality verification complete. All issues documented. Ready for project status verification.

---

### Step 1.3: Project Status Verification

[HARD] Your task: Verify project status across entire project.

[HARD] Required Scope: Scan ALL source files, not just changed files.
WHY: Partial scans miss issues in unmodified sections. Comprehensive scanning ensures quality gates pass.

Verification Requirements:
- Project integrity assessment (identify broken references, inconsistencies)
- Complete issues detection with precise locations
- Resolution recommendations for discovered issues

Output Requirements:
- Complete issue list with file locations and line numbers
- Project integrity status: Healthy or Issues Detected
- Severity classification for each issue: Critical, High, Medium, Low

Storage:
- Store complete response in `$PROJECT_VALIDATION_RESULTS`
- Format must be machine-parseable for downstream agent processing

WHY: Complete validation prevents synchronization of broken states. Detailed results enable targeted fixes.

---

### Step 1.4: Invoke Doc-Syncer for Synchronization Plan

Your task: Call manager-docs to analyze Git changes and create synchronization strategy.

Use the manager-docs subagent to:

Establish a document synchronization plan

Critical Language Configuration:

- Receive instructions in agent_prompt_language from config (default: English)
- Respond in conversation_language from config (user's preferred language)
- Example: If agent_prompt_language="en" and conversation_language="ko", receive English instructions but respond in Korean

Task Instructions:

- Analyze Git changes and create a synchronization plan
- Synchronization mode: [auto/force/status/project]
- Changed files: [from git diff]
- Project verification results: [from analysis]

Required output:

1. Summary of documents to update
2. SPEC documents requiring synchronization
3. Project improvements needed
4. Estimated work scope

Ensure all document updates align with conversation_language setting.

Store: Response in `$SYNC_PLAN`

---

### Step 1.5: Request User Approval

Present synchronization plan and get user decision:

1. Display comprehensive plan report:

   ```
   ═══════════════════════════════════════════════════════
    Document Synchronization Plan Report
   ═══════════════════════════════════════════════════════

    Project Analysis:
   - Mode: [mode]
   - Scope: [scope]
   - Changed files: [count]
   - Project mode: [Personal/Team]

    Synchronization Strategy:
   - Living Documents: [list]
   - SPEC documents: [list]
   - Project improvements needed: [count]

    Project Status:
   - Project integrity: [Healthy / Issues]
   - Project issues: [count]
   - Broken references: [count]

   ═══════════════════════════════════════════════════════
   ```

2. Ask for user approval using AskUserQuestion:

   - `question`: "Synchronization plan is ready. How would you like to proceed?"
   - `header`: "Plan Approval"
   - `multiSelect`: false
   - `options`: 4 choices:
     1. "Proceed with Sync" → Execute synchronization
     2. "Request Modifications" → Modify strategy
     3. "Review Details" → See full project results
     4. "Abort" → Cancel (no changes made)

3. Process user response:
   - IF "Proceed" → Go to PHASE 2
   - IF "Modifications" → Ask for changes, re-run PHASE 1
   - IF "Review Details" → Show project results, re-ask approval
   - IF "Abort" → Go to PHASE 4 (graceful exit)

Result: User decision captured. Command proceeds or exits.

---

##  PHASE 2: Execute Document Synchronization

Goal: Synchronize documents with code changes, update SPECs, verify quality.

### Step 2.1: Create Safety Backup

[HARD] Create safety backup before any modifications begin:

[HARD] Backup Creation Must Complete Successfully:
WHY: Backups enable rollback if synchronization fails or produces unexpected results
IMPACT: Missing backup eliminates recovery option if sync produces errors

1. [HARD] Generate Timestamp for Backup Identity:

   Requirement: Create unique timestamp identifier for backup
   Action: Execute `date +%Y-%m-%d-%H%M%S` and store result in `$TIMESTAMP`
   WHY: Timestamp enables multiple backups without overwriting previous ones

2. [HARD] Create Backup Directory:

   Requirement: Establish isolated directory for backup files
   Action: Execute `mkdir -p .moai-backups/sync-$TIMESTAMP/`
   WHY: Isolated directory prevents mixing backup versions

3. [HARD] Backup All Critical Project Files:

   Requirement: Copy all essential project files to backup
   Files to backup:
   - README.md (if exists) - Project documentation
   - docs/ directory (if exists) - Additional documentation
   - .moai/specs/ directory - SPEC definitions
   WHY: Backing up all critical files enables complete state restoration

4. [HARD] Verify Backup Integrity:

   Requirement: Confirm backup was created successfully
   Actions:
   - Execute: `ls -la .moai-backups/sync-$TIMESTAMP/`
   - Verify: Backup directory is not empty
   - If empty: Print error message and exit with failure code
   - If complete: Print success message and continue
   WHY: Verification confirms backup is usable for recovery

Result: Safety backup created and verified. Ready for synchronization phase.

---

### Step 2.2: Invoke Doc-Syncer for Document Synchronization

Your task: Call manager-docs to execute the approved synchronization plan.

Use the manager-docs subagent to:

Execute Living Document synchronization

Critical Language Configuration:

- Receive instructions in agent_prompt_language from config (default: English)
- Respond in conversation_language from config (user's preferred language)
- Example: If agent_prompt_language="en" and conversation_language="ko", receive English instructions but respond in Korean

Execute the approved synchronization plan:

Previous analysis results:

- Project verification: [from analysis]
- Synchronization strategy: [from manager-docs analysis]

Task Instructions:

1. Living Document synchronization:

   - Reflect changed code in documentation
   - Auto-generate/update API documentation
   - Update README (if needed)
   - Synchronize Architecture documents

2. Project improvements:

   - Fix project issues (if possible)
   - Restore broken references

3. SPEC synchronization:

   - Ensure SPEC documents match implementation
   - Update EARS statements if needed

4. Domain-based documentation:

   - Detect changed domains (frontend/backend/devops/database/ml/mobile)
   - Generate domain-specific documentation updates

5. Generate synchronization report:
   - File location: .moai/reports/sync-report-$TIMESTAMP.md
   - Include: Updated file list, Project improvements, results summary

Important: Use conversation_language for all document updates.

Execute the plan precisely and report results in detail.

Store: Response in `$SYNC_RESULTS`

---

### Step 2.3: Invoke Quality-Gate for Verification

Your task: Call manager-quality to verify synchronization quality.

Use the manager-quality subagent to:

Verify document synchronization quality

Critical Language Configuration:

- Receive instructions in agent_prompt_language from config (default: English)
- Respond in conversation_language from config (user's preferred language)
- Example: If agent_prompt_language="en" and conversation_language="ko", receive English instructions but respond in Korean

Task: Verify that document synchronization meets TRUST 5 principles.

Verification checks:

1. Test First: Are all project links complete?
2. Readable: Are documents well-formatted?
3. Unified: Are all documents consistent?
4. Secured: Are no credentials exposed?
5. Trackable: Are all SPECs properly linked?

Output: PASS / FAIL with details

Result: Quality verification complete.

---

### Step 2.4: Update SPEC Status to Completed

After successful synchronization, update SPEC status to completed:

1. Batch update all completed SPECs:

   ```bash
   python3 .claude/hooks/moai/spec_status_hooks.py batch_update
   ```

2. Verify status updates:

   - Check results from batch update
   - Record version changes and status transitions
   - Include status changes in sync report

3. Handle individual SPEC validation (if needed):

   ```bash
   python3 .claude/hooks/moai/spec_status_hooks.py validate_completion <SPEC_ID>
   python3 .claude/hooks/moai/spec_status_hooks.py status_update <SPEC_ID> --status completed --reason "Documentation synchronized successfully"
   ```

4. Generate status update summary:
   - Count of SPECs updated to completed
   - List of any failed updates with reasons
   - Version changes for each SPEC
   - Integration with sync report

Integration: Status updates are included in the Git commit from Phase 3 with detailed commit message.

---

##  PHASE 3: Git Operations & PR

Goal: Commit changes, transition PR (if Team mode), optionally auto-merge.

### Step 3.1: Invoke Git-Manager for Commit

Your task: Call manager-git to commit all document changes.

Use the manager-git subagent to:

Commit document synchronization changes to Git

Critical Language Configuration:

- Receive instructions in agent_prompt_language from config (default: English)
- Respond in conversation_language from config (user's preferred language)
- Example: If agent_prompt_language="en" and conversation_language="ko", receive English instructions but respond in Korean

Task: Commit document synchronization changes to Git.

Commit Scope:

- All changed document files
- .moai/reports/ directory
- README.md (if changed)
- docs/ directory (if changed)

Commit Message Template:

```
docs: sync documentation with code changes

Synchronized Living Documents:
- [list from synchronization results]

Project updates:
- [count] repairs completed
- SPEC index updated

SPEC synchronization:
- [count] SPECs updated

Domain-specific sync:
- [domain list if applicable]

Generated with Claude Code
```

Important:

- Bundle all changes into a single commit
- Report success after commit

Execution Order:

1. git add (changed document files)
2. git commit -m (commit message above)
3. git log -1 (verify commit)

Verify:

- Execute: `git log -1 --oneline`
- Print commit info
- IF commit failed → Exit with error code

---

### Step 3.2: (Optional) PR Ready Transition

For Team mode projects only:

1. Check if Team mode:

   - Read: `git_strategy.mode` from config
   - IF Personal → Skip to next phase

2. Transition PR to Ready:

   - Use Task tool:
     - `subagent_type`: "manager-git"
     - `description`: "Transition PR to Ready for Review"
     - `prompt`: "Transition PR from Draft to Ready. Execute: `gh pr ready`"

3. Assign reviewers and labels (if configured)

---

### Step 3.3: (Optional) PR Auto-Merge

If `--auto-merge` flag is set:

1. Check CI/CD status:

   - Execute: `gh pr checks`
   - IF failing → Print warning and skip merge

2. Check merge conflicts:

   - Execute: `gh pr view --json mergeable`
   - IF conflicts exist → Print warning and skip merge

3. Execute auto-merge:

   - Execute: `gh pr merge --squash --delete-branch`

4. Branch cleanup:
   - Checkout: `git checkout develop`
   - Pull: `git pull origin develop`
   - Delete local branch if merge succeeded

---

##  PHASE 4: Completion & Next Steps

Goal: Report results and guide user to next action.

### Step 4.1: Display Completion Report

Print comprehensive summary:

```
═══════════════════════════════════════════════════════
 Document Synchronization Complete
═══════════════════════════════════════════════════════

 Synchronization Summary:
- Mode: [mode]
- Scope: [scope]
- Files updated: [count]
- Files created: [count]
- Project improvements: [count]

 Documents Updated:
- Living Documents: [list]
- SPEC documents: [list]
- Domain-specific reports: [count]

 Project Status:
- Project integrity: [PASS / WARNING]

 Reports Generated:
- Master sync report: .moai/reports/sync-report-$TIMESTAMP.md
- Domain reports: [list if any]

 Backup Location:
- Safety backup: .moai-backups/sync-$TIMESTAMP/

═══════════════════════════════════════════════════════
```

---

### Step 4.2: Handle Worktree/Branch Workflow Options

If $WORKTREE_MODE is true:

After sync completion, provide worktree-specific options:

1. Ask for worktree next action:

   ```python
   AskUserQuestion({
       "questions": [{
           "question": f"Worktree synchronization for {CURRENT_SPEC_ID} is complete. What would you like to do?",
           "header": "Worktree Next Steps",
           "multiSelect": false,
           "options": [
               {
                   "label": "Return to Main Directory",
                   "description": "Exit worktree and return to main project directory"
               },
               {
                   "label": "Continue in Worktree",
                   "description": "Stay in current worktree for continued development"
               },
               {
                   "label": "Switch to Another Worktree",
                   "description": "Navigate to a different SPEC worktree"
               },
               {
                   "label": "Remove This Worktree",
                   "description": "Clean up and remove the current worktree"
               }
           ]
       }]
   })
   ```

2. Execute user choice:
   - IF "Return to Main Directory" → Execute: `cd ~/MoAI/MoAI-ADK`
   - IF "Continue in Worktree" → Stay in current directory
   - IF "Switch to Another Worktree" → List available worktrees and facilitate switch
   - IF "Remove This Worktree" → Execute: `moai-worktree remove {CURRENT_SPEC_ID}` then return to main

If $BRANCH_MODE is true:

After sync completion, provide branch-specific options:

1. Ask for branch next action:

   ```python
   AskUserQuestion({
       "questions": [{
           "question": f"Branch synchronization for {CURRENT_BRANCH} is complete. What would you like to do?",
           "header": "Branch Next Steps",
           "multiSelect": false,
           "options": [
               {
                   "label": "Commit and Push Changes",
                   "description": "Commit sync changes and push to remote branch"
               },
               {
                   "label": "Return to Main Branch",
                   "description": "Switch back to main branch without pushing"
               },
               {
                   "label": "Create Pull Request",
                   "description": "Create PR for this branch and return to main"
               },
               {
                   "label": "Continue on Branch",
                   "description": "Stay on current branch for continued development"
               }
           ]
       }]
   })
   ```

2. Execute user choice:
   - IF "Commit and Push Changes" → Execute: `git add . && git commit && git push origin {CURRENT_BRANCH}`
   - IF "Return to Main Branch" → Execute: `git checkout main` (warn about uncommitted changes)
   - IF "Create Pull Request" → Execute: `gh pr create` then checkout main
   - IF "Continue on Branch" → Stay on current branch

### Step 4.3: Standard Next Steps (Non-Worktree/Branch Mode)

Use AskUserQuestion to guide next steps:

- `question`: "Documentation synchronization complete. What would you like to do next?"
- `header`: "Next Steps"
- `multiSelect`: false
- `options`: 3-4 choices depending on context:
  - " Create Next SPEC" → /moai:1-plan
  - " Start New Session" → /clear for fresh context
  - "📤 Review PR" (Team mode) → gh pr view --web
  - " Continue Development" (Personal mode)
  - " Project Overview" → Review reports and docs

---

##  Graceful Exit (User Aborts)

If user chooses to abort in PHASE 1:

```
═══════════════════════════════════════════════════════
 Synchronization Aborted
═══════════════════════════════════════════════════════

No changes were made to:
- Documents
- Git history
- Branch state

Your project remains in its current state.

You can retry synchronization anytime with:
/moai:3-sync [mode]

═══════════════════════════════════════════════════════
```

Exit command with code 0.

---

##  Quick Reference

| Scenario             | Mode    | Entry Point                 | Key Phases                                                    | Expected Outcome          |
| -------------------- | ------- | --------------------------- | ------------------------------------------------------------- | ------------------------- |
| Daily development    | auto    | `/moai:3-sync`              | Phase 0.5 (QV) → Phase 1 → Analysis → Phase 2 → Sync → Phase 3 → Git | PR Ready + docs synced    |
| Error recovery       | force   | `/moai:3-sync force`        | Phase 0.5 (QV) → Full project re-sync                         | All docs regenerated      |
| Quick health check   | status  | `/moai:3-sync status`       | Phase 0.5 (QV) → Status check only                            | Health report             |
| Milestone completion | project | `/moai:3-sync project`      | Phase 0.5 (QV) → Integrated sync                              | Project-wide updates      |
| Auto-merge workflow  | auto    | `/moai:3-sync --auto-merge` | Phase 0.5 (QV) → PR auto-merge + cleanup                      | Branch merged and deleted |

Note: Phase 0.5 = Quality Verification (pytest, ruff, mypy, code-review)

Associated Agents:

- `manager-docs` - Living Document synchronization
- `manager-quality` - TRUST 5 validation
- `manager-git` - Git operations and PR management

Documentation Outputs:

- Living Documents: Auto-synchronized with code
- SPEC Documents: Updated to match implementation
- Reports: `.moai/reports/sync-report-{timestamp}.md`
- Backup: `.moai-backups/sync-{timestamp}/` (safety backup)

Version: 3.4.0 (Full Language Support with Config-Based Coverage)
Last Updated: 2025-12-22
Architecture: Commands → Agents → Skills (Complete delegation)
Supported Languages: Python, TypeScript, JavaScript, Go, Rust, Ruby, Java, PHP, Kotlin, Swift, C#, C++, Elixir, R, Flutter/Dart, Scala
Coverage Target: Read from .moai/config/sections/quality.yaml (constitution.test_coverage_target)

---

## Final Step: Next Action Selection

After documentation synchronization completes, use AskUserQuestion tool to guide user to next action:

```python
AskUserQuestion({
    "questions": [{
        "question": "Documentation synchronization is complete. What would you like to do next?",
        "header": "Next Steps",
        "multiSelect": false,
        "options": [
            {
                "label": "Develop New Feature",
                "description": "Execute /moai:1-plan to plan new feature"
            },
            {
                "label": "Process PR Merge",
                "description": "Review and merge Pull Request"
            },
            {
                "label": "Complete Workflow",
                "description": "Complete current work and clean up session"
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

You must NOW execute the command following the "OVERALL WORKFLOW STRUCTURE" described above.

1. Start PHASE 1: Analysis & Planning immediately.
2. Use the manager-docs subagent (or appropriate subagent for the step).
3. Do NOT just describe what you will do. DO IT.
