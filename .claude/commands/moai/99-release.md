---
description: "MoAI-ADK Go release with agent delegation for git operations and quality validation"
argument-hint: "[VERSION] - optional target version (e.g., 2.0.0)"
type: local
allowed-tools: Read, Write, Edit, Grep, Glob, Bash, TodoWrite, AskUserQuestion, Task
model: sonnet
version: 2.0.0
---

## ⚠️ IMPORTANT: Test/Private Release Configuration

**This release command is configured for:**
- **Branch**: `moai-go-v2` (NOT main)
- **Tag prefix**: `go-v` (e.g., `go-v2.0.0`)
- **Purpose**: Test and private deployment for v2.0 development
- **Visibility**: Tags are pushed but GitHub Releases are OPTIONAL

**For public main releases**, use a separate release workflow targeting the `main` branch.

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

@go.mod
@pkg/version/version.go

---

## PHASE 1: Quality Gates (Execute Now)

Create TodoWrite with these items, then run each check:

1. Run all tests: `go test -race ./... -count=1 2>&1 | tail -30`
2. Run golangci-lint: `golangci-lint run 2>&1 | tail -20`
3. Run go vet: `go vet ./... 2>&1 | tail -10`
4. Run go fmt check: `gofumpt -l . | head -10`

If formatting issues found, fix them:
`make fmt`

If lint made changes, commit them:
`git add -A && git commit -m "style: Auto-fix lint and format issues"`

Display quality summary:

- tests: PASS or FAIL (if FAIL, stop and report)
- golangci-lint: PASS or WARNING
- go vet: PASS or WARNING
- gofmt: PASS or FIXED

### Error Handling

If any quality gate FAILS or encounters unexpected errors:

- **Use the expert-debug subagent** to diagnose and resolve the issue
- Example: `Use the expert-debug subagent to investigate why tests are failing`
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

If VERSION argument was provided (e.g., "2.0.0"):

- Use that version directly
- Skip AskUserQuestion

If no VERSION argument:

- Read current version from `pkg/version/version.go`
- Use AskUserQuestion to ask: patch/minor/major

Calculate new version and update ALL version files:

1. Edit `pkg/version/version.go` `Version` variable
2. Edit `.moai/config/sections/system.yaml` `moai.version`
3. Edit `internal/template/templates/.moai/config/sections/system.yaml` `moai.version`
4. Commit: `git add pkg/version/version.go .moai/config/sections/system.yaml internal/template/templates/.moai/config/sections/system.yaml && git commit -m "chore: Bump version to X.Y.Z"`

IMPORTANT: All 3 version files MUST be updated for release workflow to succeed.
The GoReleaser validates version consistency via ldflags.

Version files checklist:
- [ ] pkg/version/version.go: Version = "X.Y.Z"
- [ ] .moai/config/sections/system.yaml: moai.version: "X.Y.Z"
- [ ] internal/template/templates/.moai/config/sections/system.yaml: moai.version: "X.Y.Z"

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
moai update

# Verify version
moai version
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
moai update

# 버전 확인
moai version
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

## Mission: Release Git Operations for Version X.Y.Z (Test/Private Release)

### Context

- Target version: X.Y.Z
- Target branch: moai-go-v2 (test branch)
- Tag format: go-vX.Y.Z (Go edition prefix)
- Current state: [describe current git state]
- Quality gates: All passed
- Commits included: [list commit count and summary]

### Required Actions

1. **Check remote status**: Verify if tag go-vX.Y.Z exists on remote (origin)
2. **Handle tag conflicts**:
   - If remote does NOT have go-vX.Y.Z: Create tag and push
   - If remote already has go-vX.Y.Z: Report situation with options
3. **Execute push**: `git push origin moai-go-v2 --tags`
   - This pushes to moai-go-v2 branch (NOT main)
   - Tags will be available for private use
4. **Optional**: Check if release workflow started (if configured)

### Expected Output

Report back with:

1. Remote tag status
2. Action taken (pushed/recreated/recommended)
3. GitHub Actions workflow status
4. Release links (if successful)

```

Example delegation:
```

Use the manager-git subagent to handle release git operations for version 2.0.0

Context:

- Target branch: moai-go-v2 (test branch)
- Local tag go-v2.0.0 already exists
- 6 commits included since go-v1.9.0
- All quality gates passed

The agent should:

1. Check if go-v2.0.0 exists on remote
2. Push tag to moai-go-v2 branch or handle conflicts
3. Optionally check if GitHub Actions workflow started
4. Report release status with links (if applicable)

````

---

## PHASE 7: Release Verification & Notes Update (OPTIONAL for Test Releases)

**NOTE**: For test/private releases on `moai-go-v2` branch, GitHub Release steps are OPTIONAL. The tag is sufficient for internal use.

### Step 1: [OPTIONAL] Verify GitHub Actions Workflow

Check if release workflow started:
`gh run list --workflow=release.yml --limit 3`

Wait for workflow completion (typically 5-10 minutes for Go builds).

### Step 2: [OPTIONAL] Verify GitHub Release Created

`gh release view go-vX.Y.Z`

If release exists but has minimal notes, proceed to Step 3.

If no release exists, you can skip to final verification or create a draft release manually.

### Step 3: [OPTIONAL] Update GitHub Release Notes with CHANGELOG Content

**For test releases**: Consider creating a DRAFT or PRIVATE release instead of public.

If creating/updating a GitHub Release, use CHANGELOG content:

```bash
# For test releases, consider creating as DRAFT
gh release edit go-vX.Y.Z --draft --notes "$(cat <<'RELEASE_EOF'
# go-vX.Y.Z - English Title (YYYY-MM-DD) [TEST RELEASE]

## Summary
[Copy from CHANGELOG.md English section]

## Breaking Changes
[Copy breaking changes]

## Added
[Copy additions - can be summarized]

## Installation & Update

\`\`\`bash
# Test release installation (using specific tag)
curl -sSL https://raw.githubusercontent.com/modu-ai/moai-adk/moai-go-v2/install.sh | bash -s -- --version X.Y.Z
moai version
\`\`\`

---

# go-vX.Y.Z - Korean Title (YYYY-MM-DD) [테스트 릴리스]

## 요약
[Copy from CHANGELOG.md Korean section]

## Breaking Changes
[Copy Korean breaking changes]

## 추가됨
[Copy Korean additions]

## 설치 및 업데이트

\`\`\`bash
# 테스트 릴리스 설치 (특정 태그 사용)
curl -sSL https://raw.githubusercontent.com/modu-ai/moai-adk/moai-go-v2/install.sh | bash -s -- --version X.Y.Z
moai version
\`\`\`
RELEASE_EOF
)"
```

### Step 4: Final Verification

1. Verify release notes updated: `gh release view go-vX.Y.Z | head -50`
2. Check release assets: `gh release view go-vX.Y.Z --json assets`
3. Report final summary with links:
   - GitHub Release: https://github.com/modu-ai/moai-adk/releases/tag/go-vX.Y.Z
   - GitHub Actions: https://github.com/modu-ai/moai-adk/actions
   - Branch: moai-go-v2 (test branch)

**Note**:
- GitHub Release notes should match CHANGELOG structure (English → Korean)
- Test releases are marked as DRAFT by default
- Tag format: `go-vX.Y.Z` (not `vX.Y.Z`)

---

## Output Format

### Phase Progress

```markdown
## Release: Phase 3/7 - Version Selection

### Quality Gates
- tests: PASS (20/20 packages)
- golangci-lint: PASS
- go vet: PASS
- gofmt: CLEAN

### Version Update
- Current: 2.0.0-rc1
- Target: 2.0.0 (patch)

Updating version files...
````

### Complete

```markdown
## Release: COMPLETE

### Summary

- Version: 2.0.0-rc1 → 2.0.0
- Commits: 8 commits included
- Quality: All gates passed

### Links

- GitHub Release: https://github.com/modu-ai/moai-adk/releases/tag/v2.0.0
- Release Assets: darwin-arm64, darwin-amd64, linux-arm64, linux-amd64, windows-amd64

<moai>DONE</moai>
```

---

## Key Rules

- **Target branch**: `moai-go-v2` (test/private releases, NOT main)
- **Tag format**: `go-vX.Y.Z` (Go edition prefix, e.g., `go-v2.0.0`)
- Tests MUST pass to continue (85%+ coverage per package)
- All version files must be consistent (3 files: pkg/version/version.go, .moai/config/sections/system.yaml, internal/template/templates/.moai/config/sections/system.yaml)
- GitHub Releases are OPTIONAL for test releases
- GoReleaser handles binary builds if configured
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
Context: [test/lint output]

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
├── release-20260206-143052.json    # Timestamp-based snapshot
└── latest.json                      # Symlink to most recent

# Snapshot contents
{
  "timestamp": "2026-02-06T14:30:52Z",
  "target_version": "2.0.0",
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
