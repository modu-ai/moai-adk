---
description: "MoAI-ADK release with Claude Code review and tag-based auto deployment"
argument-hint: "[VERSION] - optional target version (e.g., 0.35.0)"
type: local
allowed-tools: Read, Write, Edit, Grep, Glob, Bash, TodoWrite, AskUserQuestion
model: sonnet
---

## EXECUTION DIRECTIVE - START IMMEDIATELY

This is a release command. Execute the workflow below in order. Do NOT just describe the steps - actually run the commands.

Arguments provided: $ARGUMENTS

- If VERSION argument provided: Use it as target version, skip version selection
- If no argument: Ask user to select version type (patch/minor/major)

---

## Pre-execution Context

!git status --porcelain
!git branch --show-current
!git tag --list --sort=-v:refname | head -5
!git log --oneline -10

@pyproject.toml
@src/moai_adk/version.py

---

## PHASE 1: Quality Gates (Execute Now)

Create TodoWrite with these items, then run each check:

1. Run pytest: `uv run pytest tests/ -v --tb=short -q 2>&1 | tail -30`
2. Run ruff check: `uv run ruff check src/ --fix`
3. Run ruff format: `uv run ruff format src/`
4. Run mypy: `uv run mypy src/moai_adk/ --ignore-missing-imports 2>&1 | tail -20`

If ruff made changes, commit them:
`git add -A && git commit -m "style: Auto-fix lint and format issues"`

Display quality summary:

- pytest: PASS or FAIL (if FAIL, stop and report)
- ruff: PASS or FIXED
- mypy: PASS or WARNING

---

## PHASE 2: Code Review (Execute Now)

Get commits since last tag:
`git log $(git describe --tags --abbrev=0 2>/dev/null || echo HEAD~20)..HEAD --oneline`

Get diff stats:
`git diff $(git describe --tags --abbrev=0 2>/dev/null || echo HEAD~20)..HEAD --stat`

Analyze changes for:

- Bug potential
- Security issues
- Breaking changes
- Test coverage

Display review report with recommendation: PROCEED or REVIEW_NEEDED

---

## PHASE 3: Version Selection

If VERSION argument was provided (e.g., "0.35.0"):

- Use that version directly
- Skip AskUserQuestion

If no VERSION argument:

- Read current version from pyproject.toml
- Use AskUserQuestion to ask: patch/minor/major

Calculate new version and update:

1. Edit pyproject.toml version field
2. Edit src/moai_adk/version.py MOAI_VERSION
3. Commit: `git add pyproject.toml src/moai_adk/version.py && git commit -m "chore: Bump version to X.Y.Z"`

---

## PHASE 4: CHANGELOG Generation (Bilingual Required)

IMPORTANT: CHANGELOG MUST always contain actual content!

Step 1: Get commits since last tag
```
COMMITS=$(git log $(git describe --tags --abbrev=0)..HEAD --pretty=format:"- %s (%h)")
```

Step 2: Analyze commits to categorize changes:
- Features (feat): New functionality
- Fixes (fix): Bug fixes
- Changes (chore): Maintenance, refactoring
- Docs (docs): Documentation updates
- Performance (perf): Performance improvements

Step 3: Generate CHANGELOG content

Read the existing CHANGELOG.md first to understand the format.

Then create TWO separate sections in CHANGELOG.md:

Section 1 - English (MUST contain actual changes):

```
# vX.Y.Z - [Title based on main changes] (YYYY-MM-DD)

## Summary
[1-2 sentence summary of what this release delivers]

## What's Changed

### Features
[List new features from commits with "feat:" prefix]

### Bug Fixes
[List bug fixes from commits with "fix:" prefix]

### Improvements
[List other changes from "chore:", "refactor:", etc.]

### Documentation
[List docs changes if significant]

## Installation & Update

```bash
# Upgrade existing installation
pip install --upgrade moai-adk

# Or with uv
uv pip install --upgrade moai-adk

# Or with pipx
pipx upgrade moai-adk
```

For full installation instructions, visit: https://github.com/modu-ai/moai-adk#installation

---
```

Section 2 - Korean (immediately after English section, MUST contain actual changes):

```
# vX.Y.Z - [Korean title based on changes] (YYYY-MM-DD)

## 요약
[이번 릴리스가 제공하는 것을 1-2문장으로 설명]

## 변경 사항

### 새로운 기능
["feat:" 접두사가 있는 커밋 목록]

### 버그 수정
["fix:" 접두사가 있는 커밋 목록]

### 개선 사항
["chore:", "refactor:" 등 기타 변경 사항]

### 문서
[문서 업데이트가 있으면 목록]

## 설치 및 업데이트

```bash
# 기존 설치 업그레이드
pip install --upgrade moai-adk

# uv 사용
uv pip install --upgrade moai-adk

# pipx 사용
pipx upgrade moai-adk
```

전체 설치 방법: https://github.com/modu-ai/moai-adk#installation

---
```

CRITICAL RULES:
1. ALWAYS include actual commit-based changes, NEVER leave placeholder text
2. If no changes in a category, omit that section entirely
3. Both English and Korean sections are MANDATORY
4. Date format: YYYY-MM-DD (today's date)
5. Title should reflect the most significant change

Prepend both sections to CHANGELOG.md and commit:
`git add CHANGELOG.md && git commit -m "docs: Update CHANGELOG for vX.Y.Z"`

VERIFY CHANGELOG content before committing:
```bash
head -60 CHANGELOG.md  # Should show both sections with actual content
```

---

## PHASE 5: Final Approval

Display release summary:

- Version change
- Commits included
- Quality gate results
- What will happen after approval

Use AskUserQuestion:

- Release: Create tag and push
- Abort: Cancel (changes remain local)

---

## PHASE 6: Tag and Push

If approved:

1. Create tag: `git tag -a vX.Y.Z -m "Release vX.Y.Z"`
2. Push: `git push origin main --tags`
3. Wait 5 seconds for GitHub Actions to start
4. Verify GitHub Actions workflow started: `gh run list --limit 3`
5. Display completion message

---

## PHASE 7: Release Verification

After push completes:

1. Check release workflow: `gh run list --workflow=release.yml --limit 1`
2. Verify GitHub Release: `gh release list --limit 3`
3. Display release information: `gh release view vX.Y.Z`

Display final summary with links:

- GitHub Release: https://github.com/modu-ai/moai-adk/releases/tag/vX.Y.Z
- GitHub Actions: https://github.com/modu-ai/moai-adk/actions
- PyPI: https://pypi.org/project/moai-adk/

Note: GitHub Release is created automatically by release.yml workflow.
If the release is not immediately visible, wait 2-3 minutes for the workflow to complete.

---

## Key Rules

- pytest MUST pass to continue
- All version files must be consistent
- Tag format: vX.Y.Z (with 'v' prefix)
- GitHub Actions handles PyPI deployment automatically

---

## BEGIN EXECUTION

Start Phase 1 now. Create TodoWrite and run quality gates immediately.
