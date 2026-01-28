---
description: "MoAI-ADK release with agent delegation for git operations and quality validation"
argument-hint: "[VERSION] - optional target version (e.g., 0.35.0)"
type: local
allowed-tools: Read, Write, Edit, Grep, Glob, Bash, TodoWrite, AskUserQuestion, Task
model: sonnet
version: 1.2.0
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

### Error Handling

If any quality gate FAILS or encounters unexpected errors:

- **Use the expert-debug subagent** to diagnose and resolve the issue
- Example: `Use the expert-debug subagent to investigate why smoke tests are failing`
- Resume release workflow only after all gates pass

---

## PHASE 2: Code Review (Execute Now)

[SOFT] Apply --ultrathink keyword for comprehensive code review analysis
WHY: Release requires careful analysis of changes for bugs, security issues, and breaking changes
IMPACT: Sequential thinking ensures thorough risk assessment before version release

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
2. Edit src/moai_adk/version.py \_FALLBACK_VERSION
3. Edit .moai/config/config.yaml moai.version
4. Edit .moai/config/sections/system.yaml moai.version
5. Edit src/moai_adk/templates/.moai/config/sections/system.yaml moai.version
6. Commit: `git add pyproject.toml src/moai_adk/version.py .moai/config/config.yaml .moai/config/sections/system.yaml src/moai_adk/templates/.moai/config/sections/system.yaml && git commit -m "chore: Bump version to X.Y.Z"`

IMPORTANT: All 5 version files MUST be updated for release workflow to succeed.
The Unified Release Pipeline validates version consistency across all config files.

Version files checklist:
- [ ] pyproject.toml: version = "X.Y.Z"
- [ ] src/moai_adk/version.py: _FALLBACK_VERSION = "X.Y.Z"
- [ ] .moai/config/config.yaml: moai.version: "X.Y.Z"
- [ ] .moai/config/sections/system.yaml: moai.version: "X.Y.Z"
- [ ] src/moai_adk/templates/.moai/config/sections/system.yaml: moai.version: "X.Y.Z"

---

## PHASE 4: CHANGELOG Generation (Bilingual Required)

Get commits: `git log $(git describe --tags --abbrev=0)..HEAD --pretty=format:"- %s (%h)"`

### CRITICAL: CHANGELOG Structure Rule

**[HARD] Each version MUST have Korean section IMMEDIATELY after English section.**

Correct structure (English → Korean per version):
```
# vX.Y.Z - English Title (YYYY-MM-DD)
[English content]
---
# vX.Y.Z - Korean Title (YYYY-MM-DD)
[Korean content]
---
# vX-1.Y.Z - Previous English
[Previous English content]
---
# vX-1.Y.Z - Previous Korean
[Previous Korean content]
```

**WRONG structure (all English then all Korean):**
```
# vX.Y.Z - English
# vX-1.Y.Z - English  ← WRONG: Korean should come before this
# vX.Y.Z - Korean
# vX-1.Y.Z - Korean
```

### Section 1 - English:

```markdown
# vX.Y.Z - English Title (YYYY-MM-DD)

## Summary
[English summary with key features as bullet list]

## Breaking Changes
[List breaking changes if any]

## Added
[New features grouped by category]

## Changed
[Modified features]

## Installation & Update

\`\`\`bash
# Update to the latest version
uv tool update moai-adk

# Update project templates in your folder
moai update

# Verify version
moai --version
\`\`\`
```

---

### Section 2 - Korean (IMMEDIATELY after English, BEFORE previous version):

```markdown
# vX.Y.Z - Korean Title (YYYY-MM-DD)

## 요약
[Korean summary]

## Breaking Changes
[Korean breaking changes]

## 추가됨
[Korean additions]

## 설치 및 업데이트

\`\`\`bash
# 최신 버전으로 업데이트
uv tool update moai-adk

# 프로젝트 폴더 템플릿 업데이트
moai update

# 버전 확인
moai --version
\`\`\`
```

---

Both sections are REQUIRED. Verify structure before committing:
- [ ] English vX.Y.Z section exists
- [ ] Korean vX.Y.Z section IMMEDIATELY follows English vX.Y.Z
- [ ] Previous version (vX-1.Y.Z) comes AFTER Korean vX.Y.Z

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

## PHASE 6: Tag and Push (AGENT DELEGATION REQUIRED)

**IMPORTANT: ALL git operations MUST be delegated to manager-git agent.**

If approved:

### DO NOT execute git commands directly

Instead, delegate to manager-git subagent with this prompt:

```

## Mission: Release Git Operations for Version X.Y.Z

### Context

- Target version: X.Y.Z
- Current state: [describe current git state]
- Quality gates: All passed
- Commits included: [list commit count and summary]

### Required Actions

1. **Check remote status**: Verify if tag X.Y.Z exists on remote (origin)
2. **Handle tag conflicts**:
   - If remote does NOT have v{X.Y.Z}: Create tag and push
   - If remote already has v{X.Y.Z}: Report situation with options
3. **Execute push**: `git push origin main --tags`
4. **Verify GitHub Actions**: Check if release workflow started

### Expected Output

Report back with:

1. Remote tag status
2. Action taken (pushed/recreated/recommended)
3. GitHub Actions workflow status
4. Release links (if successful)

```

Example delegation:
```

Use the manager-git subagent to handle release git operations for version 1.5.0

Context:

- Local tag v1.5.0 already exists
- 6 commits included since v1.4.6
- All quality gates passed

The agent should:

1. Check if v1.5.0 exists on remote
2. Push tag to remote or handle conflicts
3. Verify GitHub Actions workflow started
4. Report release status with links

````

---

## PHASE 7: Release Verification & Notes Update

### Step 1: Verify GitHub Actions Workflow

Check if release workflow started:
`gh run list --workflow=release.yml --limit 3`

Wait for workflow completion (typically 2-5 minutes).

### Step 2: Verify GitHub Release Created

`gh release view vX.Y.Z`

If release exists but has minimal notes, proceed to Step 3.

### Step 3: Update GitHub Release Notes with CHANGELOG Content

**[HARD] GitHub Release notes MUST include full CHANGELOG content (English + Korean).**

The automated release workflow creates a basic release. Update it with full CHANGELOG:

```bash
gh release edit vX.Y.Z --notes "$(cat <<'RELEASE_EOF'
# vX.Y.Z - English Title (YYYY-MM-DD)

## Summary
[Copy from CHANGELOG.md English section]

## Breaking Changes
[Copy breaking changes]

## Added
[Copy additions - can be summarized]

## Installation & Update

\`\`\`bash
uv tool update moai-adk
moai update
moai --version
\`\`\`

---

# vX.Y.Z - Korean Title (YYYY-MM-DD)

## 요약
[Copy from CHANGELOG.md Korean section]

## Breaking Changes
[Copy Korean breaking changes]

## 추가됨
[Copy Korean additions]

## 설치 및 업데이트

\`\`\`bash
uv tool update moai-adk
moai update
moai --version
\`\`\`
RELEASE_EOF
)"
```

### Step 4: Final Verification

1. Verify release notes updated: `gh release view vX.Y.Z | head -50`
2. Check PyPI package: https://pypi.org/project/moai-adk/
3. Report final summary with links:
   - GitHub Release: https://github.com/modu-ai/moai-adk/releases/tag/vX.Y.Z
   - GitHub Actions: https://github.com/modu-ai/moai-adk/actions
   - PyPI: https://pypi.org/project/moai-adk/

**Note**: GitHub Release notes should match CHANGELOG structure (English → Korean).

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
````

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
- **[HARD] ALL git operations MUST be delegated to manager-git agent**
  - Direct git commands (tag, push) are PROHIBITED
  - Use Task tool with manager-git subagent for all git operations
- **[HARD] Quality gate failures MUST be delegated to expert-debug agent**
  - Use Task tool with expert-debug subagent for diagnostics
  - Resume only after all gates pass

## Agent Delegation Pattern

**For git operations (Phase 6 & 7):**

```bash
Use the manager-git subagent to handle release git operations for version X.Y.Z

Context:
- [current git state]
- [commit summary]
- [quality gate results]

The agent should:
1. Check remote tag status
2. Handle conflicts appropriately
3. Push tag to remote
4. Verify GitHub Actions workflow
5. Report release status with links
```

**For quality gate failures (Phase 1):**

```bash
Use the expert-debug subagent to diagnose quality gate failures

Issue: [describe failure]
Context: [test/lint/mypy output]

The agent should:
1. Analyze root cause
2. Propose fixes
3. Verify resolution
```

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
