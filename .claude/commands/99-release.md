---
description: "MoAI-ADK v2.x production release via GitHub Flow. Creates release/vX.Y.Z branch, version bump, CHANGELOG, PR to main, squash merge, then tag push. Tag vX.Y.Z triggers GoReleaser. All git operations delegated to manager-git. Quality failures escalate to expert-debug."
argument-hint: "[VERSION] - optional target version (e.g., 2.1.0). If omitted, prompts for patch/minor/major selection."
type: local
allowed-tools: Read, Write, Edit, Grep, Glob, Bash, TodoWrite, AskUserQuestion, Task
disable-model-invocation: true
version: 4.0.0
metadata:
  release_target: "production"
  branch: "main"
  tag_format: "vX.Y.Z"
  changelog_format: "korean_first"
  release_notes_format: "bilingual"
  git_delegation: "required"
  quality_escalation: "expert-debug"
---

## Release Configuration

- **Git workflow**: GitHub Flow with enforce_admins (direct main push blocked)
- **Release branch**: `release/vX.Y.Z` (created from main, merged via PR)
- **Target branch**: `main` (all releases merge to main)
- **Tag format**: `vX.Y.Z` (standard semver, triggers GoReleaser via `.github/workflows/release.yml`)
- **Merge strategy**: Squash merge with branch auto-delete
- **Release URL**: https://github.com/modu-ai/moai-adk/releases/tag/vX.Y.Z
- **Binaries**: darwin-arm64, darwin-amd64, linux-arm64, linux-amd64, windows-amd64

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

## PHASE 0: Pre-flight Checks

Before starting the release process, verify the working directory is clean:

1. **Check for uncommitted changes**:
   ```bash
   git status --porcelain
   ```

2. **Handle uncommitted files**:
   - If untracked files in `.claude/` exist: Check if they should be committed or discarded
   - Clean up command:
   ```bash
   git checkout -- .claude/
   git clean -fd .claude/ internal/template/templates/.claude/
   ```

3. **Verify branch**:
   - Must be on `main` branch
   - If not, checkout main: `git checkout main`

4. **Pull latest changes**:
   ```bash
   git pull origin main
   ```

5. **Create release branch**:
   ```bash
   git checkout -b release/vX.Y.Z
   ```
   All version bump and CHANGELOG commits will be made on this branch.
   The release branch will be merged to main via PR (enforce_admins is enabled).

---

## PHASE 1: Quality Gates

Create TodoWrite with these items, then run each check in parallel where possible:

1. Run all tests: `go test -race ./... -count=1 2>&1 | tail -30`
2. Run go vet: `go vet ./... 2>&1 | tail -10`
3. Run go fmt check: `gofumpt -l . 2>/dev/null | head -10`

If formatting issues found, fix with `make fmt` and commit:
`git add -A && git commit -m "style: auto-fix formatting issues"`

Display quality summary:

- tests: PASS or FAIL (if FAIL, stop and report)
- go vet: PASS or WARNING
- gofmt: PASS or FIXED

### Error Handling

If any quality gate FAILS:

- **Use the expert-debug subagent** to diagnose and resolve the issue
- Resume release workflow only after all gates pass

---

## PHASE 2: Code Review

Get commits since last tag:
`git log $(git describe --tags --abbrev=0 2>/dev/null || echo HEAD~20)..HEAD --oneline`

Get diff stats:
`git diff $(git describe --tags --abbrev=0 2>/dev/null || echo HEAD~20)..HEAD --stat`

Analyze changes for:

- Bug potential
- Security issues
- Breaking changes
- Test coverage gaps

Display review report with recommendation: PROCEED or REVIEW_NEEDED

---

## PHASE 3: Version Selection

If VERSION argument was provided (e.g., "2.1.0"):

- Use that version directly
- Skip AskUserQuestion

If no VERSION argument:

- Read current version from `pkg/version/version.go`
- Use AskUserQuestion to ask: patch/minor/major

Calculate new version and update ALL version files:

1. Edit `pkg/version/version.go` `Version` variable
2. Edit `.moai/config/sections/system.yaml` `moai.version` AND `moai.template_version`
3. Edit `internal/template/templates/.moai/config/sections/system.yaml` `moai.version`
4. Commit: `git add pkg/version/version.go .moai/config/sections/system.yaml internal/template/templates/.moai/config/sections/system.yaml && git commit -m "chore: bump version to vX.Y.Z"`

Version files checklist:
- [ ] pkg/version/version.go: Version = "vX.Y.Z"
- [ ] .moai/config/sections/system.yaml: moai.version: "X.Y.Z"
- [ ] .moai/config/sections/system.yaml: moai.template_version: "vX.Y.Z"
- [ ] internal/template/templates/.moai/config/sections/system.yaml: moai.version: "X.Y.Z"

---

## PHASE 4: CHANGELOG Generation (Bilingual: English First)

### [HARD] English-First Bilingual Format

CHANGELOG.md and GitHub Release notes MUST follow English-first bilingual structure. This ensures international users see English first while maintaining Korean documentation.

Get commits for changelog: `git log $(git describe --tags --abbrev=0)..HEAD --pretty=format:"- %s (%h)"`

### [HARD] Content Filtering: Separate User-Facing vs Internal Changes

CHANGELOG and GitHub Release notes MUST prioritize **user-facing changes** that affect the moai binary behavior or CLI functionality.

**Full description (individual bullet points):**
- Go source code changes (`cmd/`, `internal/`, `pkg/`)
- User-visible CLI commands and flags
- Hook behavior changes that users experience
- Breaking changes to configuration format
- Security fixes

**Abbreviated (single summary line, no individual bullets):**
- Template/rules file optimization (`.claude/rules/`, `.claude/skills/`, `.claude/agents/`)
- Code annotation changes (`@MX` tags, comments-only changes)
- Internal documentation updates (worktree rules, agent authoring guides)
- Skill definition template updates

**Excluded entirely (do not mention):**
- Local development configuration changes (`.gitignore`, IDE settings, editor configs)
- Release workflow and CI pipeline modifications (`.github/workflows/`, release command updates)

### CHANGELOG.md Structure

Prepend new version entry to CHANGELOG.md with this structure:

```
## [X.Y.Z] - YYYY-MM-DD

### Summary
[English: Key features and improvements as 2-3 line summary]

### Breaking Changes
[English: List of breaking changes, or "None" if none]

### Added
- [English addition 1]
- [English addition 2]

### Changed
- [English change 1]
- [English change 2]

### Fixed
- [English fix 1]
- [English fix 2]

### Installation & Update

\`\`\`bash
# Update to the latest version
moai update

# Verify version
moai version
\`\`\`

---

## [X.Y.Z] - YYYY-MM-DD (한국어)

### 요약
[Korean: 핵심 기능과 개선 사항을 2-3줄로 요약]

### 주요 변경 사항 (Breaking Changes)
[Korean: 호환성을 깨는 변경 사항 목록, 없으면 "없음"]

### 추가됨 (Added)
- [Korean addition 1]
- [Korean addition 2]

### 변경됨 (Changed)
- [Korean change 1]
- [Korean change 2]

### 수정됨 (Fixed)
- [Korean fix 1]
- [Korean fix 2]

### 설치 및 업데이트 (Installation & Update)

\`\`\`bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
\`\`\`

---

[Previous version entry comes here]
```

### CHANGELOG Verification Checklist

- [ ] English section appears FIRST in version entry
- [ ] Korean section appears SECOND with `---` separator
- [ ] English uses standard changelog terminology (Added, Changed, Fixed)
- [ ] Korean uses native terminology (추가됨, 변경됨, 수정됨)
- [ ] Installation commands are identical in both sections
- [ ] Previous version entry comes AFTER both sections

Commit CHANGELOG.md:
`git add CHANGELOG.md && git commit -m "docs: update CHANGELOG for vX.Y.Z"`

---

## PHASE 5: Final Approval

Display release summary:

- Version change (current to target)
- Commits included (count and key items)
- Quality gate results
- What will happen after approval

Use AskUserQuestion:

- Release: Push release branch, create PR, merge, and tag
- Abort: Cancel (changes remain on local release branch)

---

## PHASE 6: Release Branch PR and Tag (AGENT DELEGATION REQUIRED)

**[HARD] ALL git operations MUST be delegated to manager-git agent.**
**[HARD] Branch protection with enforce_admins is enabled. Direct push to main is blocked.**
**[HARD] manager-git subagent MUST be invoked with `isolation: "worktree"` to enforce branch isolation.**

If approved, delegate to manager-git subagent with this context and Task invocation:

```bash
# Example Task invocation with worktree isolation
Task(
  subagent_type: "manager-git",
  isolation: "worktree",
  prompt: "[Mission context below]"
)
```

```
## Mission: Release Git Operations for Version X.Y.Z

### Context

- Target version: X.Y.Z
- Current branch: release/vX.Y.Z (created in Phase 0)
- Tag format: vX.Y.Z (standard semver)
- Branch protection: enforce_admins enabled, direct main push blocked
- Quality gates: All passed
- Commits on release branch: version bump + CHANGELOG

### Required Actions (GitHub Flow Release)

1. Push release branch to remote:
   `git push -u origin release/vX.Y.Z`

2. Create PR from release branch to main:
   ```bash
   gh pr create \
     --head release/vX.Y.Z \
     --base main \
     --title "release: vX.Y.Z" \
     --body "$(cat <<'EOF'
   ## Release vX.Y.Z

   ### Changes
   - Version bump to vX.Y.Z
   - CHANGELOG updated (bilingual)
   - [list key changes from commits since last tag]

   ### Quality Gates
   - Tests: PASS
   - Lint: PASS
   - Build: PASS

   This PR is auto-generated by the MoAI release workflow.
   EOF
   )"
   ```

3. Wait for CI checks to pass:
   `gh pr checks --watch`

4. Merge PR with squash:
   `gh pr merge --squash --delete-branch`

5. Sync local main with merged result:
   ```bash
   git checkout main
   git pull origin main
   ```

6. Check remote status: Verify if tag vX.Y.Z exists on remote (origin)
7. Handle tag conflicts:
   - If remote does NOT have vX.Y.Z: Create tag on current main HEAD and push
   - If remote already has vX.Y.Z: Report situation with options
8. Create and push tag (only if tag does NOT exist on remote):
   ```bash
   # Only execute this block if vX.Y.Z tag does NOT exist on origin
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```
9. Verify GoReleaser workflow triggered (tags bypass branch protection)

### Expected Output

Report back with:
1. PR number and merge status
2. Remote tag status
3. Action taken
4. GitHub Actions workflow status
5. Release URL: https://github.com/modu-ai/moai-adk/releases/tag/vX.Y.Z
```

---

## PHASE 7: GitHub Release Notes (Bilingual: English First)

### Step 1: Wait for GoReleaser

**Check workflow status with retry logic:**

```bash
# Check if workflow was triggered
gh run list --workflow=release.yml --limit 3

# Wait for workflow completion (retry up to 10 times)
for i in {1..10}; do
  STATUS=$(gh run list --workflow=release.yml --limit 1 --json status --jq '.[0].status')
  if [[ "$STATUS" == "completed" ]]; then
    echo "GoReleaser workflow completed"
    break
  fi
  echo "Waiting for GoReleaser... (attempt $i/10)"
  sleep 30
done

# Verify workflow success
CONCLUSION=$(gh run list --workflow=release.yml --limit 1 --json conclusion --jq '.[0].conclusion')
if [[ "$CONCLUSION" != "success" ]]; then
  echo "GoReleaser workflow failed with: $CONCLUSION"
  echo "Check logs: gh run view --log"
  exit 1
fi
```

**Verify release was created:**

```bash
# Check if release exists
gh release view vX.Y.Z --json tagName,assets

# If release doesn't exist, check workflow logs
gh run view --log | grep -A 20 "GoReleaser"
```

GoReleaser creates an initial release with auto-generated notes and binary assets.

**Expected assets:**
- moai-adk_X.Y.Z_darwin_arm64.tar.gz
- moai-adk_X.Y.Z_darwin_amd64.tar.gz
- moai-adk_X.Y.Z_linux_arm64.tar.gz
- moai-adk_X.Y.Z_linux_amd64.tar.gz
- moai-adk_X.Y.Z_windows_amd64.zip
- moai-adk_X.Y.Z_windows_arm64.zip
- checksums.txt

### Step 2: Replace Release Notes with English-First Bilingual Content

**[HARD] English section FIRST, Korean section SECOND. Use bilingual format identical to CHANGELOG.**

Use `gh release edit` to replace the auto-generated notes:

```bash
gh release edit vX.Y.Z --notes "$(cat <<'RELEASE_EOF'
# vX.Y.Z - [English Title] (YYYY-MM-DD)

## Summary

[English: Key features and improvements summary]

## Breaking Changes

[English: List of breaking changes, or "None" if none]

## Added

[English additions grouped by category]

## Changed

[English modifications]

## Fixed

[English bug fixes]

## Installation & Update

\`\`\`bash
# Fresh install
curl -sSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
moai version

# Existing users update
moai update
\`\`\`

**Migrating from Python Version (v1.x)**:
1. Uninstall Python version: `uv tool uninstall moai-adk`
2. Install Go Edition (use commands above)
3. Update project templates: `moai init`

---

# vX.Y.Z - [Korean Title] (YYYY-MM-DD)

## 요약

[Korean: 핵심 기능과 개선 사항을 포함한 요약]

## 주요 변경 사항 (Breaking Changes)

[Korean: 호환성을 깨는 변경 사항, 없으면 "없음"]

## 추가됨 (Added)

- [Korean additions grouped by category]

## 변경됨 (Changed)

- [Korean modifications]

## 수정됨 (Fixed)

- [Korean bug fixes]

## 설치 및 업데이트 (Installation & Update)

\`\`\`bash
# 신규 설치
curl -sSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
moai version

# 기존 사용자 업데이트
moai update
\`\`\`

**Python 버전(v1.x)에서 마이그레이션**:
1. Python 버전 제거: `uv tool uninstall moai-adk`
2. Go 에디션 설치 (위 명령어 사용)
3. 프로젝트 템플릿 업데이트: `moai init`
RELEASE_EOF
)"
```

### Step 3: Final Verification

**1. Verify release notes format:**

```bash
# Check release notes structure
gh release view vX.Y.Z | head -80
```

Checklist:
- [ ] English section appears first
- [ ] Separator `---` present between sections
- [ ] Korean section appears second
- [ ] Installation commands present in both sections
- [ ] Breaking changes section present (or "None")

**2. Verify release assets:**

```bash
# List all assets
gh release view vX.Y.Z --json assets --jq '.assets[].name'

# Expected count: 7 files (6 binaries + checksums.txt)
```

Required assets:
- [ ] moai-adk_X.Y.Z_darwin_arm64.tar.gz
- [ ] moai-adk_X.Y.Z_darwin_amd64.tar.gz
- [ ] moai-adk_X.Y.Z_linux_arm64.tar.gz
- [ ] moai-adk_X.Y.Z_linux_amd64.tar.gz
- [ ] moai-adk_X.Y.Z_windows_amd64.zip
- [ ] moai-adk_X.Y.Z_windows_arm64.zip
- [ ] checksums.txt

**3. Manual binary download test:**

```bash
# Test darwin-arm64 binary download
DOWNLOAD_URL=$(gh release view vX.Y.Z --json assets --jq '.assets[] | select(.name | contains("darwin_arm64")) | .url')
curl -L -o /tmp/test-binary.tar.gz "$DOWNLOAD_URL"

# Verify archive contents
tar -tzf /tmp/test-binary.tar.gz | head -5

# Clean up
rm /tmp/test-binary.tar.gz
```

Expected archive structure:
```
moai-adk_X.Y.Z_darwin_arm64/
moai-adk_X.Y.Z_darwin_arm64/moai
moai-adk_X.Y.Z_darwin_arm64/README.md
moai-adk_X.Y.Z_darwin_arm64/LICENSE
```

**4. Verify archive naming matches checker logic:**

The `internal/update/checker.go` expects archives WITHOUT "v" prefix:
- Correct: `moai-adk_2.1.0_darwin_arm64.tar.gz`
- Wrong: `moai-adk_v2.1.0_darwin_arm64.tar.gz`

```bash
# Verify naming convention
gh release view vX.Y.Z --json assets --jq '.assets[].name' | grep -v "^moai-adk_v"
```

If any assets have "v" prefix, the update checker will fail.

**5. Report final summary:**

Display completion report with:
- GitHub Release: https://github.com/modu-ai/moai-adk/releases/tag/vX.Y.Z
- GitHub Actions: https://github.com/modu-ai/moai-adk/actions
- Full release: `gh release view vX.Y.Z --web` (opens in browser)
- Asset count: 7 files verified
- Manual download test: PASS

---

## PHASE 8: Local Environment Update

After release is verified, update the local development environment to use the new binary and sync templates.

**1. Update local binary to released version:**

```bash
moai update --binary
```

This downloads and installs the released binary from GitHub Releases. The `--binary` flag skips template sync since the local project already has the latest templates (they were the source for the release).

**2. Sync local project templates (if needed):**

If local `.claude/` or `.moai/` files are out of sync with the templates:

```bash
moai update --templates-only
```

**3. Verify local environment:**

```bash
moai version
```

Confirm the version matches the released `vX.Y.Z`.

---

## Output Format

### Phase Progress

```markdown
## Release: Phase 3/7 - Version Selection

### Quality Gates
- tests: PASS (33 packages)
- go vet: PASS
- gofmt: CLEAN

### Version Update
- Current: 2.0.0
- Target: 2.1.0 (minor)

Updating version files...
```

### Complete

```markdown
## Release: COMPLETE

### Summary
- Version: 2.0.0 -> 2.1.0
- Commits: 8 commits included
- Quality: All gates passed

### Links
- GitHub Release: https://github.com/modu-ai/moai-adk/releases/tag/v2.1.0
- Release Assets: darwin-arm64, darwin-amd64, linux-arm64, linux-amd64, windows-amd64

<moai>DONE</moai>
```

---

## Key Rules

- **Target branch**: `main` (production releases)
- **Git workflow**: GitHub Flow with enforce_admins (direct main push blocked)
- **Release flow**: release/vX.Y.Z branch → PR → squash merge → tag → push tag
- **Tag format**: `vX.Y.Z` (triggers GoReleaser via release.yml)
- Tags bypass branch protection (only branch pushes are blocked)
- Tests MUST pass to continue (85%+ coverage per package)
- All 3 version files must be consistent
- **[HARD] CHANGELOG and GitHub Release: English FIRST, Korean SECOND**
- **[HARD] ALL git operations MUST be delegated to manager-git agent**
- **[HARD] Quality gate failures MUST be delegated to expert-debug agent**
- **[HARD] Never `git push origin main` — always use PR merge flow**

---

## Troubleshooting

### Issue 1: Go Build Cache Permission Errors

**Symptoms:**
```
permission denied: operation not permitted
cache access errors during test execution
```

**Solution:**
```bash
# Clear Go build and test caches with sandbox bypass
go clean -cache -testcache

# Verify cache was cleared
go clean -cache -testcache && echo "Cache cleared successfully"
```

**Prevention:** Run cache cleanup in Phase 0 pre-flight checks.

---

### Issue 2: Unused Import Errors in Tests

**Symptoms:**
```
internal/shell/config_test.go:6: "runtime" imported and not used
internal/template/deployer_test.go:8: "runtime" imported and not used
```

**Solution:**
```bash
# Delegate to expert-debug agent to identify and remove unused imports
# Or manually remove unused imports and run gofmt

# Fix formatting
make fmt

# Verify fix
go test ./...
```

**Prevention:** Run `go vet` and `gofumpt -l .` in Phase 1 quality gates.

---

### Issue 3: Binary Download Failure "No Go binary available"

**Root Cause:** Archive naming mismatch between GoReleaser output and update checker.

**Symptoms:**
```
$ moai update
Error: No Go binary available for this platform (darwin/arm64)
```

**Diagnosis:**
```bash
# Check actual asset names
gh release view v2.1.0 --json assets --jq '.assets[].name'

# Expected: moai-adk_2.1.0_darwin_arm64.tar.gz (without "v")
# If you see: moai-adk_v2.1.0_darwin_arm64.tar.gz (with "v"), checker will fail
```

**Solution:** Archive names MUST NOT include "v" prefix. GoReleaser's `{{ .Version }}` strips "v" automatically.

**Code Fix Location:** `internal/update/checker.go` line 121-126
```go
// Strip "v" and "go-v" prefixes from tag name to match GoReleaser's {{ .Version }}
version := strings.TrimPrefix(release.TagName, "go-v")
version = strings.TrimPrefix(version, "v")
archiveName := fmt.Sprintf("moai-adk_%s_%s_%s.%s", version, runtime.GOOS, runtime.GOARCH, ext)
```

**Prevention:** Verify archive naming in Phase 7 Step 4.

---

### Issue 4: GoReleaser Workflow Failed

**Symptoms:**
```bash
$ gh run list --workflow=release.yml --limit 1
# Shows: conclusion = "failure"
```

**Diagnosis:**
```bash
# View workflow logs
gh run view --log

# Common causes:
# 1. Tag format incorrect (must be vX.Y.Z)
# 2. GoReleaser config syntax error
# 3. Missing GitHub token permissions
# 4. Binary build failure on specific platform
```

**Solution:**
```bash
# Check tag format
git tag -l "v*" | tail -5

# If tag is wrong format, delete and recreate
git tag -d vX.Y.Z
git push origin :refs/tags/vX.Y.Z
git tag vX.Y.Z
git push origin vX.Y.Z --tags
```

**Prevention:** Verify tag format matches `vX.Y.Z` in Phase 6.

---

### Issue 5: Test Artifacts in Working Directory

**Symptoms:**
```bash
$ git status
modified:   internal/cli/.claude/agents/moai/manager-spec.md
modified:   internal/cli/.claude/commands/moai/01-plan.md
# ... many more files in internal/cli/.claude/
```

**Root Cause:** Test files created by Claude Code during testing.

**Solution:**
```bash
# Discard test artifacts (run in Phase 0)
git checkout -- internal/cli/ .claude/
git clean -fd .claude/ internal/template/templates/.claude/

# Verify clean
git status --porcelain
```

**Prevention:** Always run Phase 0 pre-flight checks before starting release.

---

### Issue 6: Version File Inconsistency

**Symptoms:**
- `pkg/version/version.go` shows v2.1.0
- `.moai/config/sections/system.yaml` shows 2.0.5
- `internal/template/templates/.moai/config/sections/system.yaml` shows 2.0.5
- Statusline shows old `template_version` instead of new release version

**Solution:** All 4 version fields MUST be updated together in Phase 3.

**Verification:**
```bash
# Check consistency
grep -n "Version" pkg/version/version.go
grep -n "moai.version\|template_version" .moai/config/sections/system.yaml
grep -n "moai.version" internal/template/templates/.moai/config/sections/system.yaml

# All should show same version (with/without "v" prefix as appropriate)
```

**Prevention:** Phase 3 includes version files checklist.

---

### Issue 7: CHANGELOG Format Violation

**Symptoms:** English and Korean sections in wrong order or missing separator.

**Correct Format:**
```markdown
## [2.1.0] - 2026-02-09

### Summary
[English text here]

### Added
- [English items]

---

## [2.1.0] - 2026-02-09 (한국어)

### 요약
[Korean text here]

### 추가됨 (Added)
- [Korean items]

---

[Previous version entry]
```

**Prevention:** Use CHANGELOG verification checklist in Phase 4.

---

## BEGIN EXECUTION

Start Phase 1 now. Create TodoWrite and run quality gates immediately.
