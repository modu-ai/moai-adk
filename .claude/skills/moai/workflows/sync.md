---
name: moai-workflow-sync
description: >
  Synchronizes documentation with code changes, verifies project quality,
  and finalizes pull requests. Third step of the Plan-Run-Sync workflow.
  Includes SPEC divergence analysis and project document updates.
  Use when documentation sync, PR creation, or quality verification is needed.
license: Apache-2.0
compatibility: Designed for Claude Code
user-invocable: false
metadata:
  version: "1.1.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-03"
  tags: "sync, documentation, pull-request, quality, verification, pr"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["sync", "docs", "pr", "documentation", "pull request", "changelog", "readme"]
  agents: ["manager-docs", "manager-quality", "manager-git"]
  phases: ["sync"]
---

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

### Phase 0: Deployment Readiness Check

Purpose: Verify the implementation is deployment-ready before quality verification and documentation sync. Catches deployment-blocking issues early.

#### Step 0.1: Test Passage Verification

- Run full test suite for detected project language
- Verify all tests pass (zero failures required)
- If tests fail: Present failure summary and offer options via AskUserQuestion
  - Fix and retry: Delegate to expert-debug subagent
  - Continue anyway: Proceed with warning
  - Abort: Exit sync workflow

#### Step 0.2: Migration Check

- Scan for database schema changes (new models, altered tables, migration files)
- Scan for configuration format changes (new config keys, changed defaults)
- Scan for data format changes (API request/response shape changes)
- If migrations detected: Flag as deployment prerequisite and include in sync report

#### Step 0.3: Environment and Configuration Changes

- Scan for new environment variables referenced in code but not in .env.example or documentation
- Scan for new configuration files or sections added
- Scan for changed default values in existing configuration
- If changes detected: Generate environment change summary for inclusion in PR description

#### Step 0.4: Backward Compatibility Assessment

- Identify public API changes (removed endpoints, changed signatures, removed fields)
- Identify breaking changes in exported functions or types
- Identify dependency version changes that may affect consumers
- Severity classification:
  - Breaking: Must be documented and versioned (semver major bump)
  - Deprecation: Must include migration guide
  - Compatible: No action required
- If breaking changes detected: Require explicit user acknowledgment via AskUserQuestion

Output: deployment_readiness_report with test_status, migrations_needed, env_changes, breaking_changes, and overall readiness status (READY, NEEDS_ATTENTION, or BLOCKED).

If overall status is BLOCKED: Present blocking issues to user and exit unless user overrides.

### Phase 0.5: Quality Verification

Purpose: Detect project language and run language-specific diagnostics (tests, linter, type checker) in parallel, followed by code review.

#### Step 0.5.1: Language Detection

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

#### Step 0.5.2: Execute Diagnostics in Parallel

Launch three background tasks simultaneously:

- Test Runner: Language-specific test command (pytest, npm test, go test, cargo test, etc.)
- Linter: Language-specific lint command (ruff, eslint, golangci-lint, clippy, etc.)
- Type Checker: Language-specific type check (mypy, tsc --noEmit, go vet, etc.)

Collect all results with timeouts (180s for tests, 120s for others). Handle partial failures gracefully.

#### Step 0.5.3: Handle Test Failures

If any tests fail, use AskUserQuestion:

- Continue: Proceed with sync despite failures
- Abort: Stop sync, fix tests first (exit to Phase 4 graceful exit)

#### Step 0.5.4: Code Review

Agent: manager-quality subagent

Invoke regardless of project language. Execute TRUST 5 quality validation and generate comprehensive quality report.

#### LSP Quality Gates

The sync phase enforces LSP-based quality gates as configured in quality.yaml:
- Zero errors required (lsp_quality_gates.sync.max_errors: 0)
- Maximum 10 warnings allowed (lsp_quality_gates.sync.max_warnings: 10)
- Clean LSP state required (lsp_quality_gates.sync.require_clean_lsp: true)

#### Step 0.5.5: Generate Quality Report

Aggregate all results into a quality report showing status for test-runner, linter, type-checker, and code-review. Determine overall status (PASS or WARN).

Status mode early exit: If mode is "status", display quality report and exit. No further phases execute.

### Phase 1: Analysis and Planning

#### Step 1.1: Verify Prerequisites

- .moai/ directory must exist
- .claude/ directory must exist
- Project must be inside a Git repository


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

Create synchronization strategy based on Git changes, mode, project verification results, and deployment readiness report from Phase 0. Output: documents to update, SPECs requiring sync, project improvements needed, estimated scope, deployment notes to include in PR.

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
- Create backup directory: .moai/backups/sync-{timestamp}/
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

### Phase 3: Git Operations and Delivery

#### Step 3.0: Detect Git Workflow Strategy

Read `github.git_workflow` from `.moai/config/sections/system.yaml`. This determines how changes are delivered.

| Strategy | Branch Model | PR Behavior | Best For |
|----------|-------------|-------------|----------|
| github_flow | Feature branches off main | Auto-create PR to main | Team/OSS projects |
| main_direct | Direct commits to main | No PR created | Solo development |
| gitflow | develop/release/hotfix branches | PR to appropriate base | Enterprise projects |

Default strategy (if not configured): `github_flow`

Also read `github.spec_git_workflow` to determine SPEC branch handling:
- `feature_branch`: Each SPEC gets its own branch (recommended for github_flow/gitflow)
- `main_direct`: SPEC changes committed to current branch (only when git_workflow is main_direct)

#### Step 3.1: Commit Changes

Agent: manager-git subagent

- Stage all changed document files, reports, README, docs/
- Create single commit with descriptive message listing synchronized documents, project repairs, and SPEC updates
- Commit message language follows `language.git_commit_messages` setting
- Verify commit with git log

#### Step 3.2: Push and Deliver (Strategy-Aware)

Behavior varies based on `github.git_workflow` setting and current branch context.

##### Strategy: github_flow

Detect current branch:

**Feature branch** (any branch other than main):
1. Push branch to remote: `git push -u origin <branch>`
2. Check if PR already exists: `gh pr list --head <branch> --json number`
3. If no PR exists: Create PR via `gh pr create`
   - Title: Derived from SPEC title or branch name
   - Body: Include sync summary, files changed, quality report, deployment readiness notes (migrations, env changes, breaking changes)
   - Base: main
   - Labels: auto-detected from changed files
4. If PR exists: Update with comment summarizing sync changes
5. Display PR URL to user

**Main branch** (direct commit):
- Push directly: `git push origin main`
- Display push confirmation
- Note: Direct main commits are permitted but feature branches are recommended

**Worktree context** (detected from git directory structure):
- Push worktree branch to remote
- Create PR if not exists (same as feature branch flow)
- Display PR URL and worktree context

##### Strategy: main_direct

All commits go directly to main, no PRs:
1. Push to main: `git push origin main`
2. Display push confirmation
3. No PR created regardless of branch name

##### Strategy: gitflow

Detect current branch type and route accordingly:

**feature/* branch** → PR to `develop`:
1. Push branch: `git push -u origin <branch>`
2. Create or update PR targeting `develop` branch
3. Display PR URL

**release/* branch** → PR to `main`:
1. Push branch: `git push -u origin <branch>`
2. Create or update PR targeting `main` branch
3. Display PR URL

**hotfix/* branch** → PR to `main` (and back-merge to develop):
1. Push branch: `git push -u origin <branch>`
2. Create or update PR targeting `main` branch
3. After merge: Create follow-up PR to `develop` for back-merge
4. Display PR URLs

**develop branch** → Push directly:
1. Push to develop: `git push origin develop`
2. Display push confirmation

**main branch** → Error:
- Direct commits to main are not allowed in gitflow
- Suggest creating a hotfix or release branch instead

#### Step 3.3: PR Ready Transition (Team Mode)

Only applies when a PR was created in Step 3.2:

- If Team mode enabled and PR is draft: Transition to ready via `gh pr ready`
- Assign reviewers and labels if configured
- If Team mode disabled: Do NOT automatically transition (user controls readiness)

#### Step 3.4: Auto-Merge (When --merge flag set)

Only applies when a PR was created in Step 3.2.

Execution conditions [HARD]:
- Flag must be explicitly set: --merge
- All CI/CD checks must pass
- PR must have zero merge conflicts
- Minimum reviewer approvals obtained (if Team mode)

Auto-merge execution:
1. Check CI/CD status via `gh pr checks --watch` (wait for completion)
2. Check merge conflicts via `gh pr view --json mergeable`
3. If passing and mergeable: Execute `gh pr merge --squash --delete-branch`
4. Checkout target branch, fetch latest
5. Verify local is synchronized with remote

Auto-merge failures:
- If CI/CD fails: Report failure, display error details, do NOT merge
- If merge conflicts: Report conflicts, provide manual resolution guidance, do NOT merge
- If approvals missing (Team mode): Report pending approvals, do NOT merge

### Phase 4: Completion and Next Steps

#### Completion Report

Display summary including:
- Git workflow strategy used (github_flow, main_direct, or gitflow)
- Sync mode and scope
- Files updated and created
- Project improvements made
- Documents updated
- Reports generated
- Backup location
- PR URL (if created) or push target (if direct push)

#### Context-Aware Next Steps

Tool: AskUserQuestion with options tailored to delivery result:

**If PR was created (github_flow feature branch, or gitflow):**
- Review PR on GitHub
- Auto-Merge PR (/moai sync --merge)
- Create Next SPEC (/moai plan)
- Start New Session (/clear)

**If direct push (main_direct, or github_flow main branch):**
- Create Next SPEC (/moai plan)
- Continue Development
- Start New Session (/clear)

**If worktree context:**
- Review PR in Browser
- Return to Main Directory
- Remove This Worktree

---

## Team Mode

The sync phase always uses sub-agent mode (manager-docs), even when --team is active for other phases. Documentation synchronization requires sequential consistency and a single authoritative view of project state.

For rationale and details, see workflows/team-sync.md.

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

- Phase 0: Deployment readiness verified (tests, migrations, env changes, backward compatibility)
- Phase 0.5: Quality verification completed (tests, linter, type checker, code review)
- Phase 1: Prerequisites verified, project analyzed, divergence analysis completed, sync plan approved by user
- Phase 2: Safety backup created and verified, documents synchronized, SPEC documents updated per lifecycle level, project documents updated (if applicable), quality verified, SPEC status updated
- Phase 3: Changes committed, delivered per git_workflow strategy (PR created for github_flow/gitflow, direct push for main_direct), auto-merge executed (if flagged and PR exists)
- Phase 4: Completion report displayed with delivery result, appropriate next steps presented based on strategy and context

---

Version: 3.1.0
Updated: 2026-02-13
Source: Extracted from .claude/commands/moai/3-sync.md v3.4.0. Added SPEC divergence analysis, project document updates, SPEC lifecycle awareness, team mode section, LSP quality gates, strategy-aware git delivery, and deployment readiness check (Phase 0) with test verification, migration detection, environment changes, and backward compatibility assessment.
