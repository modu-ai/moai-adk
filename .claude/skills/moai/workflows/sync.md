# Sync Workflow Orchestration

## Purpose

Synchronize documentation with code changes, verify project quality, and finalize pull requests. This is the third step of the Plan-Run-Sync workflow.

## Scope

- Implements Step 4 of MoAI's 4-step workflow (Report and Commit)
- Receives implementation artifacts from /moai run
- Produces synchronized documentation, commits, and PR readiness

## Input

- $ARGUMENTS: Mode and optional path
  - Mode: auto (default), force, status, project
  - Path: Optional synchronization target path (e.g., src/auth/)
  - Flag: --merge

## Supported Modes

- auto (default): Smart selective sync of changed files only. PR Ready conversion. Daily development workflow.
- force: Complete regeneration of all documentation. Error recovery and major refactoring use case.
- status: Read-only health check. Quick project health report with no changes.
- project: Project-wide documentation updates. Milestone completion and periodic sync use case.

## Supported Flags

- --merge: After sync, auto-merge PR and clean up branch. Worktree/branch environment is auto-detected from git context.

## Context Loading

Before execution, load these essential files:

- .moai/config/config.yaml (git strategy, language settings)
- .moai/config/sections/git-strategy.yaml (auto_branch, branch creation policy)
- .moai/config/sections/language.yaml (git_commit_messages setting)
- .moai/specs/ directory listing (SPEC documents for sync)
- .moai/project/ directory listing (project documents for conditional update)
- README.md (current project documentation)

Pre-execution commands: git status, git diff, git branch, git log, find .moai/specs.

---

## Phase Sequence

### Phase 0.5: Quality Verification (Parallel Diagnostics)

Purpose: Validate project quality before synchronization begins. Runs before Phase 1 to catch issues early.

#### Step 1: Detect Project Language

Check indicator files in priority order (first match wins):

- Python: pyproject.toml, setup.py, requirements.txt, .python-version, Pipfile
- TypeScript: tsconfig.json, package.json with typescript dependency
- JavaScript: package.json without tsconfig
- Go: go.mod, go.sum
- Rust: Cargo.toml, Cargo.lock
- Ruby: Gemfile, .ruby-version, Rakefile
- Java: pom.xml, build.gradle, build.gradle.kts
- PHP: composer.json, composer.lock
- Kotlin: build.gradle.kts with kotlin plugin
- Swift: Package.swift, .xcodeproj, .xcworkspace
- C#/.NET: .csproj, .sln, .fsproj
- C++: CMakeLists.txt, Makefile with C++ content
- Elixir: mix.exs
- R: DESCRIPTION (R package), .Rproj, renv.lock
- Flutter/Dart: pubspec.yaml
- Scala: build.sbt, build.sc
- Fallback: unknown (skip language-specific tools, proceed to code review)

#### Step 2: Execute Diagnostics in Parallel

Launch three background tasks simultaneously:

- Test Runner: Language-specific test command (pytest, npm test, go test, cargo test, etc.)
- Linter: Language-specific lint command (ruff, eslint, golangci-lint, clippy, etc.)
- Type Checker: Language-specific type check (mypy, tsc --noEmit, go vet, etc.)

Collect all results with timeouts (180s for tests, 120s for others). Handle partial failures gracefully.

#### Step 3: Handle Test Failures

If any tests fail, use AskUserQuestion:

- Continue: Proceed with sync despite failures
- Abort: Stop sync, fix tests first (exit to Phase 4 graceful exit)

#### Step 4: Code Review

Agent: manager-quality subagent

Invoke regardless of project language. Execute TRUST 5 quality validation and generate comprehensive quality report.

#### Step 5: Generate Quality Report

Aggregate all results into a quality report showing status for test-runner, linter, type-checker, and code-review. Determine overall status (PASS or WARN).

Status mode early exit: If mode is "status", display quality report and exit. No further phases execute.

### Phase 1: Analysis and Planning

#### Step 1.1: Verify Prerequisites

- .moai/ directory must exist
- .claude/ directory must exist
- Project must be inside a Git repository
- Python 3 should be available (soft requirement)

#### Step 1.2: Analyze Project Status

- Analyze Git changes: git status, git diff, categorize changed files
- Read project configuration: git_strategy.mode, conversation_language, spec_git_workflow
- Determine synchronization mode from $ARGUMENTS
- Detect worktree context: Check if git directory contains worktrees/ component
- Detect branch context: Check current branch name

#### Step 1.3: Project Status Verification

Scan ALL source files (not just changed files) for:

- Broken references and inconsistencies
- Issues with precise locations
- Severity classification (Critical, High, Medium, Low)

#### Step 1.4: Synchronization Plan

Agent: manager-docs subagent

Create synchronization strategy based on Git changes, mode, and project verification results. Output: documents to update, SPECs requiring sync, project improvements needed, estimated scope.

#### Step 1.5: SPEC-Implementation Divergence Analysis

Purpose: Detect differences between the original SPEC plan and actual implementation to ensure documentation accuracy.

For each SPEC associated with the current sync:

- Step 1.5.1: Load SPEC Documents
  - Read spec.md (requirements), plan.md (implementation plan), acceptance.md (criteria)
  - Extract planned files, planned features, and planned scope

- Step 1.5.2: Analyze Actual Implementation
  - Use git diff and git log to identify all files created, modified, or deleted during the run phase
  - Categorize changes by domain (backend, frontend, tests, config, docs)

- Step 1.5.3: Compare Plan vs Reality
  - Identify files created that were NOT in the original plan.md
  - Identify features or endpoints implemented beyond original spec.md scope
  - Identify planned items that were NOT implemented (deferred or dropped)
  - Identify unplanned refactoring or dependency changes

- Step 1.5.4: Generate Divergence Report
  - Categorize divergences: scope_expansion, unplanned_additions, deferred_items, structural_changes
  - Include: new_directories_created, new_dependencies_added, new_features_implemented
  - This report feeds into Phase 2.2 (SPEC updates) and Phase 2.2.5 (project doc updates)

- Step 1.5.5: Check SPEC Lifecycle Level
  - Read SPEC metadata for lifecycle level (default: spec-first if not specified)
  - Level 1 (spec-first): SPEC will be marked completed with implementation summary appended
  - Level 2 (spec-anchored): SPEC content will be updated to reflect actual implementation
  - Level 3 (spec-as-source): Flag discrepancies as warnings (implementation should match SPEC exactly)

#### Step 1.6: User Approval

Tool: AskUserQuestion

Display sync plan report and present options:

- Proceed with Sync
- Request Modifications (re-run Phase 1)
- Review Details (show full project results, re-ask)
- Abort (exit with no changes)

### Phase 2: Execute Document Synchronization

#### Step 2.1: Create Safety Backup

Before any modifications:

- Generate timestamp identifier
- Create backup directory: .moai-backups/sync-{timestamp}/
- Copy critical files: README.md, docs/, .moai/specs/
- Verify backup integrity (non-empty directory check)

#### Step 2.2: Document Synchronization

Agent: manager-docs subagent

Input: Approved sync plan, project verification results, changed files list, divergence report from Phase 1.5.

Tasks for manager-docs:

- Reflect changed code in Living Documents
- Auto-generate and update API documentation
- Update README if needed
- Synchronize architecture documents
- Fix project issues and restore broken references
- Update SPEC documents based on divergence analysis and lifecycle level (see Step 2.2.1)
- Detect changed domains and generate domain-specific updates
- Generate sync report: .moai/reports/sync-report-{timestamp}.md

All document updates use conversation_language setting.

##### Step 2.2.1: SPEC Document Update (Based on Divergence Report)

Apply updates based on SPEC lifecycle level detected in Phase 1.5.5:

Level 1 (spec-first):
- Append "Implementation Notes" section to spec.md summarizing actual implementation
- Record scope changes: features added beyond plan, items deferred
- Mark SPEC as completed (no ongoing maintenance expected)

Level 2 (spec-anchored):
- Update spec.md requirements to reflect actual implementation
- Add new EARS-format requirements for features implemented beyond original scope
- Update plan.md with actual implementation steps taken
- Update acceptance.md with new acceptance criteria for added features
- Preserve original requirements with "as-implemented" annotations where changed

Level 3 (spec-as-source):
- Do NOT modify SPEC content
- Generate discrepancy report listing implementation deviations from SPEC
- Flag as warnings in sync report for manual review
- Recommend either updating SPEC or adjusting implementation

#### Step 2.2.5: Project Document Update (Conditional)

Purpose: Update .moai/project/ documents when significant structural changes are detected.

Condition: Execute this step ONLY when the divergence report from Phase 1.5 indicates:
- New directories were created in the project
- New dependencies or technologies were added
- New major features or capabilities were implemented
- Significant architectural changes occurred

Skip condition: If .moai/project/ directory does not exist or contains no files, skip this step entirely.

Agent: manager-docs subagent

Tasks for manager-docs:

- If new directories created: Update structure.md with new directory descriptions and purposes
- If new dependencies added: Update tech.md with new technology stack entries and rationale
- If new features implemented: Update product.md with new feature descriptions and use cases
- If architectural changes: Update structure.md with revised architecture patterns

Constraints:
- Only update sections relevant to detected changes (do not regenerate entire files)
- Preserve existing content and append or modify incrementally
- Use conversation_language setting for all updates

#### Step 2.3: Post-Sync Quality Verification

Agent: manager-quality subagent

Verify synchronization quality against TRUST 5:

- All project links complete
- Documents well-formatted
- All documents consistent
- No credentials exposed
- All SPECs properly linked

#### Step 2.4: Update SPEC Status

Update SPEC status based on lifecycle level and implementation completeness:

- Level 1 (spec-first): Set status to "completed". No further maintenance required.
- Level 2 (spec-anchored): Set status to "completed" if all requirements met, or "in-progress" if partial. Schedule next review based on quarterly maintenance policy.
- Level 3 (spec-as-source): Set status based on implementation-SPEC alignment. Flag discrepancies for resolution.

Record version changes, status transitions, and divergence summary. Include in sync report.

### Phase 3: Git Operations and PR

#### Step 3.1: Commit Changes

Agent: manager-git subagent

- Stage all changed document files, reports, README, docs/
- Create single commit with descriptive message listing synchronized documents, project repairs, and SPEC updates
- Verify commit with git log

#### Step 3.2: PR Ready Transition (Team Mode Only)

- Check git_strategy.mode from config
- If Team mode: Transition PR from Draft to Ready via gh pr ready
- Assign reviewers and labels if configured
- If Personal mode: Skip

#### Step 3.3: Auto-Merge (When --merge flag set)

- Check CI/CD status via gh pr checks
- Check merge conflicts via gh pr view --json mergeable
- If passing and mergeable: Execute gh pr merge --squash --delete-branch
- Checkout develop, pull, delete local branch

### Phase 4: Completion and Next Steps

#### Standard Completion Report

Display summary: mode, scope, files updated and created, project improvements, documents updated, reports generated, backup location.

#### Worktree Mode Next Steps (auto-detected from git context)

Tool: AskUserQuestion with options:

- Return to Main Directory
- Continue in Worktree
- Switch to Another Worktree
- Remove This Worktree

#### Branch Mode Next Steps (auto-detected from git context)

Tool: AskUserQuestion with options:

- Commit and Push Changes
- Return to Main Branch
- Create Pull Request
- Continue on Branch

#### Standard Next Steps

Tool: AskUserQuestion with options:

- Create Next SPEC (/moai plan)
- Start New Session (/clear)
- Review PR (Team mode, gh pr view)
- Continue Development (Personal mode)

---

## Graceful Exit

When user aborts at any decision point:

- No changes made to documents, Git history, or branch state
- Project remains in current state
- Display retry command: /moai sync [mode]
- Exit with code 0

---

## Completion Criteria

All of the following must be verified:

- Phase 0.5: Quality verification completed (tests, linter, type checker, code review)
- Phase 1: Prerequisites verified, project analyzed, divergence analysis completed, sync plan approved by user
- Phase 2: Safety backup created and verified, documents synchronized, SPEC documents updated per lifecycle level, project documents updated (if applicable), quality verified, SPEC status updated
- Phase 3: Changes committed, PR transitioned (Team mode), auto-merge executed (if flagged)
- Phase 4: Completion report displayed, appropriate next steps presented based on mode

---

Version: 1.1.0
Source: Extracted from .claude/commands/moai/3-sync.md v3.4.0. Added SPEC divergence analysis, project document updates, and SPEC lifecycle awareness.
