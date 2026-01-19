---
description: "MoAI-ADK release with Claude Code review and tag-based auto deployment"
argument-hint: "[VERSION] - optional target version (e.g., 0.35.0)"
type: local
allowed-tools: Read, Write, Edit, Grep, Glob, Bash, TodoWrite, AskUserQuestion
model: sonnet
version: 1.0.0
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

1. Run smoke tests: `uv run pytest tests/ -m "smoke or critical" -v --tb=short --maxfail=5 2>&1 | tail -30`
2. Run ruff check: `uv run ruff check src/ --fix`
3. Run ruff format: `uv run ruff format src/`
4. Run mypy: `uv run mypy src/moai_adk/ --ignore-missing-imports 2>&1 | tail -20`

If ruff made changes, commit them:
`git add -A && git commit -m "style: Auto-fix lint and format issues"`

Display quality summary:

- smoke tests: PASS or FAIL (if FAIL, stop and report)
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

Calculate new version and update ALL version files:

1. Edit pyproject.toml version field
2. Edit src/moai_adk/version.py _FALLBACK_VERSION
3. Edit .moai/config/config.yaml moai.version
4. Edit .moai/config/sections/system.yaml moai.version
5. Commit: `git add pyproject.toml src/moai_adk/version.py .moai/config/config.yaml .moai/config/sections/system.yaml && git commit -m "chore: Bump version to X.Y.Z"`

IMPORTANT: All 4 version files MUST be updated for release workflow to succeed.
The Unified Release Pipeline validates version consistency across all config files.

---

## PHASE 4: CHANGELOG Generation (Bilingual Required)

Get commits: `git log $(git describe --tags --abbrev=0)..HEAD --pretty=format:"- %s (%h)"`

IMPORTANT: Create TWO separate sections in CHANGELOG.md

Section 1 - English:

```
# vX.Y.Z - English Title (YYYY-MM-DD)
## Summary
[English summary]
## Changes
[English changes]
## Installation & Update

```bash
# Update to the latest version
uv tool update moai-adk

# Update project templates in your folder
moai update

# Verify version
moai --version
```
---
```

Section 2 - Korean (immediately after English section):

```
# vX.Y.Z - Korean Title (YYYY-MM-DD)
## 요약
[Korean summary]
## 변경 사항
[Korean changes]
## 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
uv tool update moai-adk

# 프로젝트 폴더 템플릿 업데이트
moai update

# 버전 확인
moai --version
```
---
```

Both sections are REQUIRED for proper GitHub Release generation.

Prepend both sections to CHANGELOG.md and commit:
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

## Output Format

### Phase Progress

```markdown
## Release: Phase 3/7 - Version Selection

### Quality Gates
- smoke tests: PASS (25/25)
- ruff: FIXED (3 issues auto-corrected)
- mypy: WARNING (2 type hints missing)

### Version Update
- Current: 1.4.0
- Target: 1.5.0 (minor)

Updating version files...
```

### Complete

```markdown
## Release: COMPLETE

### Summary
- Version: 1.4.0 → 1.5.0
- Commits: 12 commits included
- Quality: All gates passed

### Links
- GitHub Release: https://github.com/modu-ai/moai-adk/releases/tag/v1.5.0
- PyPI: https://pypi.org/project/moai-adk/1.5.0/

<moai>DONE</moai>
```

---

## Key Rules

- Smoke tests MUST pass to continue (tests/test_smoke.py)
- All version files must be consistent
- Tag format: vX.Y.Z (with 'v' prefix)
- GitHub Actions handles PyPI deployment automatically

---

## State Management & Recovery

Release state is saved for recovery if interrupted:

```
# Snapshot location
.moai/cache/release-snapshots/
├── release-20260119-143052.json    # Timestamp-based snapshot
└── latest.json                      # Symlink to most recent

# Snapshot contents
{
  "timestamp": "2026-01-19T14:30:52Z",
  "target_version": "1.5.0",
  "current_phase": 3,
  "quality_results": {...},
  "commits_included": [...],
  "version_files_updated": [...]
}
```

Recovery Commands:

```bash
# Resume from latest snapshot (if release was interrupted)
/moai:99-release --resume

# Check release status
/moai:99-release --status
```

WHY: Release process involves multiple steps; recovery prevents partial releases.

---

## BEGIN EXECUTION

Start Phase 1 now. Create TodoWrite and run quality gates immediately.
