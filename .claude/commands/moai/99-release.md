---
name: moai:99-release
description: "Interactive release management for MoAI-ADK packages with automatic quality gates"
argument-hint: "[no arguments - uses interactive menu]"
allowed-tools: Read, Write, Edit, Grep, Glob, Bash, TodoWrite, AskUserQuestion
model: "sonnet"
---

## Pre-execution Context

!git status --porcelain
!git branch --show-current
!git tag --list --sort=-v:refname | head -5
!git log --oneline -5
!git remote -v

## Essential Files

@pyproject.toml
@src/moai_adk/__init__.py
@CHANGELOG.md
@.moai/config/config.json

# MoAI-ADK Interactive Release Management

# MoAI-ADK 인터랙티브 릴리즈 관리

## EXCEPTION: Local-Only Development Tool

This command is exempt from "Zero Direct Tool Usage" principle because:

1. Local development only - Not distributed with package distributions
2. Maintainer-only usage - GoosLab (project owner) exclusive access
3. Direct system access required - PyPI release automation requires direct shell commands
4. Interactive menu system - Uses AskUserQuestion for user-driven workflow

Production commands must strictly adhere to agent delegation principle.

---

> Local-Only Tool: This command is for local development only.
> Never deployed with package distributions.
> Use for PyPI releases, version bumping, changelog generation, and emergency rollbacks.

---

## Command Purpose

Execute automated release management workflow following GitHub Flow:

1. PHASE 1: Automatic Quality Gates (pytest, ruff, mypy)
2. PHASE 2: Auto-fix and Commit (if fixes needed)
3. PHASE 3: Interactive Menu Selection
4. PHASE 4: Execute Selected Operation
5. PHASE 5: PR Creation and Merge (GitHub Flow)
6. PHASE 6: Completion and Next Steps

Run on: Interactive menu (no arguments required)

---

## CRITICAL: GitHub Flow Compliance

This command MUST follow GitHub Flow. Direct commits/pushes to main are PROHIBITED.

### Required Workflow

1. Create or use existing release branch: `release/vX.Y.Z`
2. All changes committed to release branch
3. Push release branch to origin
4. Create PR from release branch to main
5. Wait for CI checks to pass
6. Merge PR (via GitHub UI or gh CLI)
7. Tag the merged commit on main

### Prohibited Actions

- NEVER commit directly to main
- NEVER push directly to main
- NEVER create tags before PR merge
- NEVER use `git merge` locally to main

---

## Execution Philosophy: "Validate → Fix → Select → Execute"

`/moai:99-release` performs release management through automatic quality validation followed by interactive menu:

```
```
User Command: /moai:99-release
    |
Phase 1: Automatic Quality Gates
    |- pytest execution
    |- ruff check and format
    |- mypy type check
    |
Phase 2: Auto-fix and Commit (if needed)
    |- Apply ruff fixes
    |- Commit fixes automatically
    |
Phase 3: AskUserQuestion (Main Menu)
    |- validate: Full quality validation
    |- version: Version management
    |- changelog: Bilingual changelog
    |- prepare: CI/CD preparation
    |- rollback: Emergency rollback
    |
Phase 4: Execute Selected Operation
    |- Run selected workflow
    |- Report results
    |
Phase 5: AskUserQuestion (Next Steps)
    |- Continue with another operation
    |- Exit release management
```

### Key Principle: Automatic Progression

This command automatically executes quality checks on start:

- Phase 1 runs IMMEDIATELY on command invocation
- No waiting for user input before quality checks
- Auto-fix and commit if issues found
- Then present interactive menu

---

## PHASE 1: Automatic Quality Gates

Goal: Run quality checks immediately on command start

### Step 1.1: Initialize Task Tracking

Create TodoWrite to track progress:

```json
todos:
  - content: "Run pytest quality check"
    status: "in_progress"
    activeForm: "Running pytest quality check"
  - content: "Run ruff lint and format check"
    status: "pending"
    activeForm: "Running ruff lint and format check"
  - content: "Run mypy type check"
    status: "pending"
    activeForm: "Running mypy type check"
  - content: "Auto-fix and commit if needed"
    status: "pending"
    activeForm: "Auto-fixing and committing"
  - content: "Present interactive menu"
    status: "pending"
    activeForm: "Presenting interactive menu"
```

### Step 1.2: Run pytest

Execute pytest and capture results:

```bash
uv run pytest tests/ -v --tb=short 2>&1 | head -100
```

Record result:

- PASS: All tests passed
- FAIL: Tests failed (note count and names)

### Step 1.3: Run ruff check

Execute ruff lint check:

```bash
uv run ruff check src/ --output-format=concise 2>&1 | head -50
```

Record result:

- PASS: No lint issues
- FIXABLE: Issues found, can be auto-fixed
- FAIL: Critical issues requiring manual fix

### Step 1.4: Run ruff format check

Execute ruff format check:

```bash
uv run ruff format --check src/ 2>&1 | head -20
```

Record result:

- PASS: No formatting issues
- FIXABLE: Formatting needed

### Step 1.5: Run mypy

Execute mypy type check:

```bash
uv run mypy src/moai_adk/ --ignore-missing-imports 2>&1 | head -50
```

Record result:

- PASS: No type errors
- WARNING: Type hints missing
- FAIL: Type errors found

### Step 1.6: Quality Summary

Present quality gate summary to user:

```text
Quality Gate Results:
- pytest: [PASS/FAIL]
- ruff lint: [PASS/FIXABLE/FAIL]
- ruff format: [PASS/FIXABLE]
- mypy: [PASS/WARNING/FAIL]

Overall Status: [READY/NEEDS_FIX/BLOCKED]
```

---

## PHASE 2: Auto-fix and Commit

Goal: Automatically fix issues and commit if needed

### Step 2.1: Check if Fixes Needed

Condition: Execute only if ruff lint or format has FIXABLE status

### Step 2.2: Apply ruff fixes

```bash
uv run ruff check src/ --fix
uv run ruff format src/
```

### Step 2.3: Check for Changes

```bash
git status --porcelain
```

### Step 2.4: Auto-commit Fixes

If changes exist, create commit:

```bash
git add -A
git commit -m "fix: Auto-fix lint and format issues for release preparation"
```

### Step 2.5: Update Quality Status

Re-run quick validation to confirm fixes:

```bash
uv run ruff check src/ --output-format=concise 2>&1 | head -10
```

Report: "Auto-fix applied and committed successfully"

---

## PHASE 3: Interactive Main Menu

Goal: Present release management options via AskUserQuestion

### Step 3.1: Present Main Menu

Use AskUserQuestion to display options:

```yaml
question: "Quality gates completed. Select release management operation:"
header: "MoAI-ADK Release Management"
multiSelect: false
options:
  - label: "validate"
    description: "Run full pre-release quality validation (pytest, ruff, mypy, bandit, pip-audit)"
  - label: "version"
    description: "Version management and synchronization (patch/minor/major)"
  - label: "changelog"
    description: "Generate bilingual changelog (Korean/English)"
  - label: "prepare"
    description: "Prepare CI/CD deployment (TestPyPI or PyPI)"
  - label: "rollback"
    description: "Emergency rollback procedures"
  - label: "exit"
    description: "Exit release management"
```

### Step 3.2: Route to Selected Operation

Based on user selection, proceed to Phase 4 with appropriate workflow.

---

## PHASE 4: Execute Selected Operation

Goal: Execute the selected release management operation

### Operation: validate

Full pre-release quality validation workflow:

Sub-options (via AskUserQuestion):

- Quick (5min): pytest, ruff, mypy, version consistency
- Full (15min): Quick + bandit, pip-audit, full coverage
- Custom: Select specific checks

Execution steps:

1. Quick validation:

```bash
uv run pytest tests/ -v --tb=short
uv run ruff check src/
uv run mypy src/moai_adk/ --ignore-missing-imports
```

2. Full validation (additional):

```bash
uv run bandit -r src/ -ll
uv run pip-audit
uv run pytest tests/ --cov=src/moai_adk --cov-report=term-missing
```

3. Version consistency check:

- Read version from pyproject.toml
- Compare with .moai/config/config.json
- Compare with src/moai_adk/__init__.py
- Report any mismatches

Report validation results with PASS/FAIL for each item.

### Operation: version

Version management workflow (GitHub Flow compliant):

Sub-options (via AskUserQuestion):

- patch: Bug fix (0.31.3 -> 0.31.4)
- minor: Feature addition (0.31.3 -> 0.32.0)
- major: Breaking change (0.31.3 -> 1.0.0)
- custom: Enter specific version
- check: Version consistency check only
- sync: Synchronize all version files

Execution steps:

1. Verify current branch is release/vX.Y.Z (NOT main)
2. Read current version from pyproject.toml
3. Calculate new version based on selection
4. Update files on release branch:
   - pyproject.toml (master source)
   - .moai/config/config.json
   - src/moai_adk/__init__.py (if applicable)
5. Create Git commit: "chore: Bump version to X.Y.Z"
6. Push release branch to origin
7. IMPORTANT: Tags are created AFTER PR merge (see Phase 5)
8. Report all changes made

NOTE: Do NOT create tags on release branch. Tags are created on main after PR merge.

### Operation: changelog

Bilingual changelog generation:

Sub-options (via AskUserQuestion):

- auto: Generate from Git history since last tag
- manual: Provide changelog template for editing
- edit: Modify existing CHANGELOG.md

Execution steps:

1. Get commits since last tag:

```bash
git log $(git describe --tags --abbrev=0)..HEAD --oneline
```

2. Categorize commits:

   - feat: New features
   - fix: Bug fixes
   - docs: Documentation
   - refactor: Code refactoring
   - test: Test additions
   - chore: Maintenance

3. Generate bilingual changelog format:

   - First section: Complete changelog content in English
   - Separator: "---" line
   - Second section: Complete changelog content in Korean (full translation of English section)

   Example format:
   ```
   # v0.31.4 - Session Start Hook Fixes (2025-12-03)

   ## Summary
   [English content...]

   ## Bug Fixes
   [English content...]

   ---

   # v0.31.4 - 세션 시작 훅 수정 (2025-12-03)

   ## 요약
   [Korean content - full translation...]

   ## 버그 수정
   [Korean content - full translation...]
   ```

4. Update CHANGELOG.md with new bilingual section
5. Create commit: "docs: Update changelog for vX.Y.Z"

### Operation: prepare

CI/CD deployment preparation (GitHub Flow compliant):

Sub-options (via AskUserQuestion):

- pr-create: Create PR from release branch to main
- pr-merge: Merge approved PR and create tag
- test-deploy: TestPyPI deployment (after PR merge)
- prod-deploy: PyPI deployment (after PR merge)

Execution steps:

1. Verify current branch is release/vX.Y.Z
2. Verify all quality gates passed
3. Verify version consistency
4. Push release branch if needed:

```bash
git push origin release/vX.Y.Z
```

5. Create or update PR:

```bash
gh pr create --base main --head release/vX.Y.Z --title "Release vX.Y.Z" --body "..."
# OR if PR exists:
gh pr view release/vX.Y.Z --json url
```

6. Wait for CI checks to pass
7. After PR approval, merge via:

```bash
gh pr merge --squash --delete-branch
```

8. After merge, checkout main and create tag:

```bash
git checkout main
git pull origin main
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

9. Build and deploy package:

```bash
uv run python -m build
uv run twine check dist/*
uv run twine upload dist/*  # or --repository testpypi
```

10. Report deployment status

### Operation: rollback

Emergency rollback procedures:

Sub-options (via AskUserQuestion):

- pypi-only: Hide PyPI version
- full: PyPI + GitHub Release + Tag
- specific: Rollback to specific version

Execution steps:

1. Confirm rollback with user (double confirmation for production)
2. Execute selected rollback:
   - PyPI: Use yank command
   - GitHub: Delete release via gh CLI
   - Git: Delete tag locally and remotely

3. Report rollback status
4. Create incident log entry

---

## PHASE 5: PR Creation and Merge (GitHub Flow)

Goal: Create PR from release branch to main and manage merge process

### Step 5.1: Verify Branch State

```bash
git branch --show-current  # Must be release/vX.Y.Z
git status --porcelain     # Must be clean
```

### Step 5.2: Push Release Branch

```bash
git push origin release/vX.Y.Z
```

### Step 5.3: Create or Check PR

Check if PR exists:

```bash
gh pr list --head release/vX.Y.Z --json number,state
```

If no PR exists, create one:

```bash
gh pr create --base main --head release/vX.Y.Z \
  --title "Release vX.Y.Z: [Brief Description]" \
  --body "[PR body with summary and test plan]"
```

### Step 5.4: Wait for CI

```bash
gh pr checks release/vX.Y.Z --watch
```

### Step 5.5: Merge PR (After Approval)

Use AskUserQuestion to confirm merge:

```yaml
question: "PR checks passed. Ready to merge to main?"
options:
  - label: "Merge PR"
    description: "Merge release branch to main via squash merge"
  - label: "Wait"
    description: "Wait for more review or changes"
```

If merge selected:

```bash
gh pr merge --squash --delete-branch
```

### Step 5.6: Create Tag on Main

After merge:

```bash
git checkout main
git pull origin main
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

---

## PHASE 6: Completion and Next Steps

Goal: Guide user to next action after operation completes

### Step 6.1: Report Operation Results

Display summary of completed operation:
- Operation type
- Actions performed
- Files modified
- Git commits created
- PR URL (if created)
- Tag created (if applicable)
- Current status

### Step 6.2: Present Next Steps

Use AskUserQuestion:

```yaml
question: "[Operation] completed successfully. What would you like to do next?"
header: "Next Steps"
multiSelect: false
options:
  - label: "Continue with another operation"
    description: "Return to main menu"
  - label: "View current status"
    description: "Show project and release status"
  - label: "Deploy to PyPI"
    description: "Build and upload package"
  - label: "Exit"
    description: "Complete release management session"
```

### Step 6.3: Loop or Exit

If user selects "Continue", return to Phase 3 (Main Menu).
If user selects "Exit", display final summary and end session.

---

## Quality Gate Standards

### Minimum Requirements for Release

- pytest: All tests must pass
- ruff lint: No errors (warnings acceptable)
- ruff format: Must be formatted
- mypy: No critical errors
- Version: All files synchronized

### Security Requirements (Full validation)

- bandit: No high-severity issues
- pip-audit: No known vulnerabilities

### Coverage Requirements

- Minimum: 80% code coverage
- Target: 85% code coverage

---

## Version File Locations

Master source: pyproject.toml
Synchronized files:
- .moai/config/config.json (project.version)
- src/moai_adk/__init__.py (__version__)

Git tag format: vX.Y.Z (e.g., v0.31.3)

---

## Logging and Monitoring

Execution logs: .moai/logs/release-YYYY-MM-DD.log

Log format:
- Timestamp (ISO8601)
- Operation type
- Status (start/success/fail)
- Details

---

## Related Documentation

- CI/CD Workflow: .github/workflows/release-secure.yml
- Version Management: .moai/docs/version-management-guide.md
- Core Module: src/moai_adk/core/version_sync.py

---

## Quick Reference

| Operation | Purpose | Duration | Key Steps |
|-----------|---------|----------|-----------|
| validate | Quality check | 5-15 min | pytest, ruff, mypy, bandit |
| version | Version bump | 2 min | Update files, commit, tag |
| changelog | Update log | 5 min | Parse commits, generate bilingual |
| prepare | Build & deploy | 10 min | Build, verify, upload |
| rollback | Emergency revert | 5 min | Yank, delete, log |

Status: Enhanced with GitHub Flow Compliance and PR-based Workflow
Python Version: 3.14
MoAI-ADK Version: 0.32.0+
Last Updated: 2025-12-04

---

## EXECUTION DIRECTIVE

You must NOW execute the command following the "Execution Philosophy" described above.

1. IMMEDIATELY verify current branch (must be release/vX.Y.Z, NOT main)
2. Run Phase 1 quality gates (pytest, ruff, mypy)
3. Display quality gate results to user
4. If fixes needed, apply auto-fixes and commit on release branch (Phase 2)
5. Present interactive main menu via AskUserQuestion (Phase 3)
6. Execute selected operation (Phase 4)
7. Create/update PR for release (Phase 5)
8. Guide to next steps (Phase 6)

CRITICAL RULES:
- NEVER commit or push directly to main
- ALWAYS work on release/vX.Y.Z branch
- ALWAYS create PR for main merge
- ONLY create tags AFTER PR is merged to main

Begin with TodoWrite to track progress, then verify branch and execute pytest immediately.
