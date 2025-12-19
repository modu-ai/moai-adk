---
name: moai:99-release
description: "MoAI-ADK release with Claude Code review and tag-based auto deployment"
argument-hint: "[no arguments - uses interactive workflow]"
allowed-tools: Read, Write, Edit, Grep, Glob, Bash, TodoWrite, AskUserQuestion
model: "sonnet"
---

## Pre-execution Context

!git status --porcelain
!git branch --show-current
!git tag --list --sort=-v:refname | head -5
!git log --oneline -10
!git remote -v

## Essential Files

@pyproject.toml
@src/moai_adk/version.py
@CHANGELOG.md

# MoAI-ADK Release Management

## Local-Only Development Tool

This command is for local development only.
- Not distributed with package
- Maintainer-only access
- Direct system access required

---

## Execution Philosophy

Simple tag-based release with Claude Code review:

```
/moai:99-release
    |
[Phase 1] Quality Gates (auto)
    |- pytest, ruff, mypy
    |
[Phase 2] Claude Code Review (auto)
    |- Analyze recent changes
    |- Check for issues
    |
[Phase 3] Version Selection (interactive)
    |- patch / minor / major
    |
[Phase 4] CHANGELOG Generation (auto)
    |- Bilingual (EN/KO)
    |
[Phase 5] Final Approval (interactive)
    |- Show summary
    |- Confirm release
    |
[Phase 6] Tag & Push (auto)
    |- git tag vX.Y.Z
    |- git push --tags
    |
[GitHub Actions] Auto Deploy
    |- CI tests
    |- PyPI publish
    |- GitHub Release
```

Key Principles:
- Work directly on main branch (1-person development)
- No release branches or PRs required
- Tag push triggers GitHub Actions deployment
- Claude Code provides local code review

---

## PHASE 1: Quality Gates

Goal: Ensure code quality before release

### Step 1.1: Initialize Tracking

Create TodoWrite:
- Run pytest
- Run ruff check and format
- Run mypy type check
- Claude Code review
- Version selection
- CHANGELOG generation
- Final approval and release

### Step 1.2: Run pytest

Execute and check results:

```bash
uv run pytest tests/ -v --tb=short -q 2>&1 | tail -20
```

Result: PASS (all tests passed) or FAIL (note failures)

### Step 1.3: Run ruff

Check and auto-fix:

```bash
uv run ruff check src/ --fix
uv run ruff format src/
```

If changes made, commit:

```bash
git add -A
git commit -m "style: Auto-fix lint and format issues"
```

### Step 1.4: Run mypy

Type check:

```bash
uv run mypy src/moai_adk/ --ignore-missing-imports 2>&1 | tail -20
```

Result: PASS or WARNING (note issues)

### Step 1.5: Quality Summary

Display results:

```
Quality Gate Results:
- pytest: [PASS/FAIL]
- ruff: [PASS/FIXED]
- mypy: [PASS/WARNING]

Overall: [READY/BLOCKED]
```

If BLOCKED (pytest failed), stop and report.

---

## PHASE 2: Claude Code Review

Goal: AI-powered code review of recent changes

### Step 2.1: Collect Changes

Get commits since last tag:

```bash
git log $(git describe --tags --abbrev=0 2>/dev/null || echo HEAD~20)..HEAD --oneline
```

Get diff summary:

```bash
git diff $(git describe --tags --abbrev=0 2>/dev/null || echo HEAD~20)..HEAD --stat
```

### Step 2.2: Analyze Changes

Review the changes for:
- Bug potential: Logic errors, edge cases
- Security issues: Hardcoded secrets, injection vulnerabilities
- Breaking changes: API modifications, removed features
- Test coverage: New code without tests
- Documentation: Missing docstrings or comments

### Step 2.3: Review Report

Present findings:

```
Claude Code Review Report:

Changes Summary:
- X commits since last release
- Y files changed (+A/-B lines)

Review Findings:
- [OK/WARNING/ISSUE] Bug potential
- [OK/WARNING/ISSUE] Security
- [OK/WARNING/ISSUE] Breaking changes
- [OK/WARNING/ISSUE] Test coverage

Recommendation: [PROCEED/REVIEW_NEEDED]
```

If critical issues found, ask user whether to proceed or abort.

---

## PHASE 3: Version Selection

Goal: Determine new version number

### Step 3.1: Show Current Version

Read from pyproject.toml and display:

```
Current Version: 0.32.10
Last Tag: v0.32.10
```

### Step 3.2: Ask Version Type

Use AskUserQuestion:

```yaml
question: "Select version bump type for this release:"
header: "Version"
multiSelect: false
options:
  - label: "patch"
    description: "Bug fixes (0.32.10 -> 0.32.11)"
  - label: "minor"
    description: "New features (0.32.10 -> 0.33.0)"
  - label: "major"
    description: "Breaking changes (0.32.10 -> 1.0.0)"
```

### Step 3.3: Calculate New Version

Based on selection:
- patch: increment PATCH (X.Y.Z+1)
- minor: increment MINOR, reset PATCH (X.Y+1.0)
- major: increment MAJOR, reset others (X+1.0.0)

### Step 3.4: Update Version Files

Update pyproject.toml:

```bash
# Use sed or Edit tool to update version
```

Update src/moai_adk/version.py:

```python
MOAI_VERSION = "X.Y.Z"
```

Commit version bump:

```bash
git add pyproject.toml src/moai_adk/version.py
git commit -m "chore: Bump version to X.Y.Z"
```

---

## PHASE 4: CHANGELOG Generation

Goal: Create bilingual changelog entry

### Step 4.1: Collect Commit History

Get commits since last tag:

```bash
git log $(git describe --tags --abbrev=0)..HEAD --pretty=format:"- %s (%h)"
```

### Step 4.2: Categorize Commits

Group by conventional commit type:
- feat: New Features
- fix: Bug Fixes
- docs: Documentation
- refactor: Code Refactoring
- test: Testing
- chore: Maintenance

### Step 4.3: Generate Bilingual Entry

Create new CHANGELOG entry with format:

```markdown
# vX.Y.Z - [Title] (YYYY-MM-DD)

## Summary
[English summary of changes]

## Changes

### New Features
- [feat commits in English]

### Bug Fixes
- [fix commits in English]

[other sections as needed]

## Installation & Update

### Fresh Install (uv tool - Recommended)
uv tool install moai-adk

### Update Existing Installation
uv tool upgrade moai-adk

### Alternative Methods
# Using uvx (no install needed)
uvx moai-adk --help

# Using pip
pip install moai-adk==X.Y.Z

---

# vX.Y.Z - [Korean Title] (YYYY-MM-DD)

## 요약
[Korean summary - translation of English]

## 변경 사항

### 신규 기능
- [feat commits in Korean]

### 버그 수정
- [fix commits in Korean]

[other sections as needed]

## 설치 및 업데이트

### 신규 설치 (uv tool - 권장)
uv tool install moai-adk

### 기존 설치 업데이트
uv tool upgrade moai-adk

### 대체 방법
# uvx 사용 (설치 없이)
uvx moai-adk --help

# pip 사용
pip install moai-adk==X.Y.Z

---
```

### Step 4.4: Update CHANGELOG.md

Prepend new entry to CHANGELOG.md using Edit tool.

Commit:

```bash
git add CHANGELOG.md
git commit -m "docs: Update CHANGELOG for vX.Y.Z"
```

---

## PHASE 5: Final Approval

Goal: Confirm release before pushing

### Step 5.1: Show Release Summary

Display comprehensive summary:

```
Release Summary for vX.Y.Z
==========================

Version: 0.32.10 -> X.Y.Z
Commits: N commits included
Files Changed: M files

Quality Gates:
- pytest: PASS
- ruff: PASS
- mypy: PASS

Claude Code Review: PROCEED

New Commits:
- commit 1
- commit 2
- ...

After approval:
1. Tag vX.Y.Z will be created
2. Push to origin (main + tag)
3. GitHub Actions will:
   - Run CI tests
   - Build package
   - Publish to PyPI
   - Create GitHub Release
```

### Step 5.2: Ask for Approval

Use AskUserQuestion:

```yaml
question: "Ready to release vX.Y.Z to PyPI?"
header: "Release"
multiSelect: false
options:
  - label: "Release"
    description: "Create tag and push to trigger deployment"
  - label: "Abort"
    description: "Cancel release (changes remain committed locally)"
```

If Abort selected, display message and exit.

---

## PHASE 6: Tag and Push

Goal: Create tag and push to trigger GitHub Actions

### Step 6.1: Create Tag

```bash
git tag -a vX.Y.Z -m "Release vX.Y.Z"
```

### Step 6.2: Push to Origin

Push main branch and tag:

```bash
git push origin main --tags
```

### Step 6.3: Completion Message

Display:

```
Release vX.Y.Z initiated successfully!

Local Actions Completed:
- Version bumped to X.Y.Z
- CHANGELOG updated
- Tag vX.Y.Z created
- Pushed to origin

GitHub Actions (automatic):
- CI tests running
- PyPI deployment pending
- GitHub Release creation pending

Monitor progress:
- GitHub Actions: https://github.com/modu-ai/moai-adk/actions
- PyPI: https://pypi.org/project/moai-adk/

Next release command: /moai:99-release
```

---

## Quality Standards

### Required for Release
- pytest: All tests pass
- ruff: No lint errors
- mypy: No critical errors
- Version: Consistent across files

### Version Files
- pyproject.toml (master source)
- src/moai_adk/version.py

### Tag Format
- Format: vX.Y.Z (e.g., v0.32.11)
- Annotated tag with message

---

## GitHub Actions Integration

Tag push triggers `.github/workflows/release.yml`:

1. Extract version from tag
2. Validate version consistency
3. Run full test suite
4. Build package (wheel + sdist)
5. Publish to PyPI (or TestPyPI for test environment)
6. Create GitHub Release with artifacts

No manual PyPI upload required.

---

## Quick Reference

| Phase | Action | Type |
|-------|--------|------|
| 1 | Quality Gates | Auto |
| 2 | Claude Code Review | Auto |
| 3 | Version Selection | Interactive |
| 4 | CHANGELOG Generation | Auto |
| 5 | Final Approval | Interactive |
| 6 | Tag & Push | Auto |
| 7 | GitHub Actions Deploy | Auto (remote) |

---

## EXECUTION DIRECTIVE

Execute the release workflow now:

1. Run Phase 1 quality gates immediately (pytest, ruff, mypy)
2. Display quality results
3. If quality passes, run Phase 2 Claude Code review
4. Present Phase 3 version selection via AskUserQuestion
5. Generate Phase 4 CHANGELOG
6. Show Phase 5 summary and ask for approval
7. Execute Phase 6 tag and push if approved

Begin with TodoWrite to track progress, then execute quality gates.
