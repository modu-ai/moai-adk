---
name: moai:99-release
description: "MoAI-ADK release with Claude Code review and tag-based auto deployment"
argument-hint: "[VERSION] - optional target version (e.g., 0.35.0)"
allowed-tools: Read, Write, Edit, Grep, Glob, Bash, TodoWrite, AskUserQuestion
model: "sonnet"
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

## PHASE 4: CHANGELOG Generation

Get commits: `git log $(git describe --tags --abbrev=0)..HEAD --pretty=format:"- %s (%h)"`

Create bilingual CHANGELOG entry (English + Korean) with:
- Summary
- New Features (feat commits)
- Bug Fixes (fix commits)
- Other changes
- Installation instructions

Prepend to CHANGELOG.md and commit:
`git add CHANGELOG.md && git commit -m "docs: Update CHANGELOG for vX.Y.Z"`

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
3. Display completion message with GitHub Actions links

---

## Key Rules

- pytest MUST pass to continue
- All version files must be consistent
- Tag format: vX.Y.Z (with 'v' prefix)
- GitHub Actions handles PyPI deployment automatically

---

## BEGIN EXECUTION

Start Phase 1 now. Create TodoWrite and run quality gates immediately.
