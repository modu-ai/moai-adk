# Workflow: release - Release Management

Purpose: Orchestrate a complete release cycle with quality validation, version management, CHANGELOG generation, and publishing. All git operations MUST be delegated to the manager-git agent.

---

## Phase 1: Quality Gates

Run each quality check sequentially via Bash:

- Smoke tests: uv run pytest tests/ -m "smoke or critical" -v --tb=short --maxfail=5
- Ruff check: uv run ruff check src/ --fix
- Ruff format: uv run ruff format src/
- Mypy: uv run mypy src/moai_adk/ --ignore-missing-imports

If ruff made changes, commit them: git add -A && git commit -m "style: Auto-fix lint and format issues"

Display quality summary:

- smoke tests: PASS or FAIL
- ruff: PASS or FIXED (number of issues corrected)
- mypy: PASS or WARNING (number of type hints missing)

Error Handling:

- If any quality gate FAILS: Delegate to expert-debug subagent for diagnosis
- Resume release workflow only after all gates pass
- If smoke tests FAIL: Stop immediately and report

---

## Phase 2: Code Review

[SOFT] Apply --ultrathink for comprehensive code review analysis.

Get commits since last tag: git log $(git describe --tags --abbrev=0)..HEAD --oneline

Get diff stats: git diff $(git describe --tags --abbrev=0)..HEAD --stat

Analyze changes for:

- Bug potential
- Security issues
- Breaking changes
- Test coverage gaps

Display review report with recommendation: PROCEED or REVIEW_NEEDED.

---

## Phase 3: Version Selection

If VERSION argument was provided in $ARGUMENTS:

- Use that version directly, skip AskUserQuestion

If no VERSION argument:

- Read current version from pyproject.toml
- Use AskUserQuestion: Select version bump type

Options:

- Patch (X.Y.Z+1): Bug fixes, minor changes
- Minor (X.Y+1.0): New features, backward compatible
- Major (X+1.0.0): Breaking changes

### Version File Updates (5 files - ALL required)

- pyproject.toml: version = "X.Y.Z"
- src/moai_adk/version.py: _FALLBACK_VERSION = "X.Y.Z"
- .moai/config/config.yaml: moai.version: "X.Y.Z"
- .moai/config/sections/system.yaml: moai.version: "X.Y.Z"
- src/moai_adk/templates/.moai/config/sections/system.yaml: moai.version: "X.Y.Z"

After updating all files: git add [all 5 files] && git commit -m "chore: Bump version to X.Y.Z"

---

## Phase 4: CHANGELOG Generation (Bilingual EN+KO)

Get commits: git log $(git describe --tags --abbrev=0)..HEAD --pretty=format:"- %s (%h)"

[HARD] Each version MUST have Korean section IMMEDIATELY after English section.

Correct structure (English then Korean per version):

- vX.Y.Z English Title (YYYY-MM-DD) with Summary, Breaking Changes, Added, Changed, Installation sections
- vX.Y.Z Korean Title (YYYY-MM-DD) with corresponding Korean sections
- Then previous version entries follow

Wrong structure: All English versions grouped together, then all Korean versions grouped together.

English section headers: Summary, Breaking Changes, Added, Changed, Installation and Update
Korean section headers: (summary), Breaking Changes, (added), (installation and update)

Prepend both sections to CHANGELOG.md and commit: git add CHANGELOG.md && git commit -m "docs: Update CHANGELOG for vX.Y.Z"

---

## Phase 5: Final Approval

Display release summary via AskUserQuestion:

- Version change: current to target
- Commits included: count and summary
- Quality gate results
- What will happen after approval (tag creation, push, GitHub Actions)

Options:

- Release: Create tag and push
- Abort: Cancel (changes remain local)

---

## Phase 6: Tag and Push

[HARD] ALL git operations MUST be delegated to manager-git subagent.

[HARD] Direct git commands (tag, push) are PROHIBITED at orchestrator level.

Delegate to manager-git subagent with context:

- Target version: X.Y.Z
- Current git state
- Quality gate results
- Commit count and summary

Manager-git agent actions:

- Check remote status: Verify if tag vX.Y.Z exists on remote (origin)
- Handle tag conflicts: Create new tag or report conflict
- Execute push: git push origin main --tags
- Verify GitHub Actions: Check if release workflow started

Tag format: vX.Y.Z (with 'v' prefix).

---

## Phase 7: Release Verification

### Step 1: Verify GitHub Actions

Check workflow status: gh run list --workflow=release.yml --limit 3

Wait for workflow completion (typically 2-5 minutes).

### Step 2: Verify GitHub Release

View release: gh release view vX.Y.Z

### Step 3: Update Release Notes

[HARD] GitHub Release notes MUST include full CHANGELOG content (English + Korean).

Update release notes with full bilingual content from CHANGELOG.md using: gh release edit vX.Y.Z --notes "..."

### Step 4: Final Report

Report release summary with links:

- GitHub Release: https://github.com/modu-ai/moai-adk/releases/tag/vX.Y.Z
- GitHub Actions: https://github.com/modu-ai/moai-adk/actions
- PyPI: https://pypi.org/project/moai-adk/

---

## State Management and Recovery

Release state is saved to .moai/cache/release-snapshots/ for recovery if interrupted.

Snapshot includes: timestamp, target_version, current_phase, quality_results, commits_included, version_files_updated.

Recovery: /moai release --resume (loads latest snapshot and resumes from last completed phase).

---

## Agent Chain Summary

- Phase 1: MoAI orchestrator (Bash for quality checks), expert-debug subagent (on failure)
- Phase 2: MoAI orchestrator (analysis with --ultrathink)
- Phase 3-5: MoAI orchestrator (AskUserQuestion, file edits)
- Phase 6: manager-git subagent (tag creation, push, conflict handling)
- Phase 7: MoAI orchestrator (verification via gh CLI)

---

Version: 1.0.0
Last Updated: 2026-01-28
