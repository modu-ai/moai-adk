---
description: "MoAI-ADK production release workflow via GitHub Flow"
argument-hint: "[VERSION] | patch | minor | major"
type: local
version: 4.2.0
runtime: pi
metadata:
  release_target: "production"
  branch: "main"
  tag_format: "vVERSION"
  changelog_format: "english_first_bilingual"
  release_notes_format: "english_first_bilingual"
---

## Pi Release Workflow

Execute the release workflow using Pi-native tools. Parse `$ARGUMENTS`.

### Runtime tool policy

Use Pi-native tools only:

- Shell and repository inspection: `bash`, `read`
- File changes: `edit`, `write`
- User choices: `ask_user_question`
- Specialist diagnosis/review: `subagent`
- Parallel validation when useful: `teams`

Represent task tracking as a Markdown checklist in your response or as `teams` tasks when parallel execution is useful. Never execute commands containing literal placeholder text such as `VERSION`, `TAG`, `RELEASE_BRANCH`, or `PREVIOUS_TAG`; resolve variables first and show resolved values before running destructive or remote commands.

### Release policy

- Git workflow: GitHub Flow with direct `main` push blocked by branch protection
- Release branch: `release/VERSION`, where `VERSION` is the resolved semantic version without the `v` prefix
- Target branch: `main`
- Tag format: `vVERSION`, where `VERSION` is the resolved semantic version
- Merge strategy: GitHub PR, preferably squash merge with branch deletion
- Tag push triggers release automation through `.github/workflows/release.yml`
- Expected binary assets: darwin arm64/amd64, linux arm64/amd64, windows amd64/arm64, plus checksums

## Phase 0: Resolve target version

1. Read arguments:
   - If `$ARGUMENTS` is a semantic version, use it as `VERSION` after removing a leading `v` if present.
   - If `$ARGUMENTS` is `patch`, `minor`, or `major`, derive the next version from the latest tag or version file.
   - If `$ARGUMENTS` is empty, ask the user to choose patch/minor/major with `ask_user_question`.
2. Inspect current version sources before deciding:

```bash
git tag --list --sort=-v:refname | head -5
git describe --tags --abbrev=0 2>/dev/null || true
grep -n "Version" pkg/version/version.go 2>/dev/null || true
grep -n "moai.version\|template_version" .moai/config/sections/system.yaml 2>/dev/null || true
grep -n "moai.version\|template_version" .pi/generated/source/moai-config/sections/system.yaml 2>/dev/null || true
```

Report the resolved values:

- `VERSION`
- `TAG=vVERSION`
- `RELEASE_BRANCH=release/VERSION`
- previous tag used for changelog

## Phase 1: Pre-flight checks

1. Check working tree and branch:

```bash
git status --porcelain
git branch --show-current
git remote -v
```

2. If uncommitted changes exist, stop and ask whether to commit, stash, or abort. Do not discard changes without explicit approval.

3. Ensure release starts from `main` unless the user explicitly approves otherwise:

```bash
git checkout main
git pull origin main
```

4. Create the release branch after resolving `RELEASE_BRANCH`:

```bash
git checkout -b RELEASE_BRANCH
```

5. Optional cleanup for generated/source test artifacts: inspect first, then ask before deleting anything.

## Phase 2: Quality gates

Run quality checks before changing version files. Use `teams` only for independent read-only validation tasks; avoid concurrent writes.

Recommended Go checks:

```bash
go test -race ./... -count=1
go vet ./...
gofumpt -l . 2>/dev/null | head -20
```

If formatting issues are found, ask before applying broad formatting unless the user explicitly requested automatic release execution. If a gate fails, use `subagent` with the `expert-debug` role for focused diagnosis and resume only after the gate passes.

Report:

- tests: PASS/FAIL
- vet: PASS/WARN/FAIL
- formatting: PASS/FIXED/NEEDS ACTION

## Phase 3: Code review and release scope

Collect commits and diff since the previous tag:

```bash
git log PREVIOUS_TAG..HEAD --oneline
git diff PREVIOUS_TAG..HEAD --stat
```

Analyze and report:

- Bug potential
- Security issues
- Breaking changes
- Test coverage gaps
- User-facing changes affecting the `moai` binary or CLI behavior
- Internal-only changes that should be summarized briefly or omitted from release notes

Recommendation must be either `PROCEED` or `REVIEW_NEEDED`. If review is needed, use `subagent` for focused review before continuing.

## Phase 4: Version bump

Update all applicable version files consistently. Inspect files first; only edit files that exist in this checkout.

Primary candidates:

- `pkg/version/version.go`: version constant should use `vVERSION` when the file expects a tag string.
- `.moai/config/sections/system.yaml`: `moai.version` should use `VERSION`; `template_version` should use `vVERSION` if present.
- `.pi/generated/source/moai-config/sections/system.yaml`: keep Pi-local snapshot release metadata consistent when this repository tracks it.
- `internal/template/templates/.moai/config/sections/system.yaml`: update if present.

Verification checklist:

- `pkg/version/version.go` shows resolved release version.
- `.moai/config/sections/system.yaml` and Pi-local snapshot do not disagree.
- `template_version`, if present, uses the tag form expected by the file.
- No file still contains the previous version where release metadata should be updated.

Commit version changes separately when possible:

```bash
git diff --stat
git add pkg/version/version.go .moai/config/sections/system.yaml .pi/generated/source/moai-config/sections/system.yaml internal/template/templates/.moai/config/sections/system.yaml 2>/dev/null || true
git commit -m "chore: bump version to vVERSION"
```

## Phase 5: CHANGELOG generation

CHANGELOG and GitHub Release notes must be English-first bilingual: English section first, Korean section second, separated by `---`.

### Content filtering

Prioritize user-facing changes:

- Go source code changes under `cmd/`, `internal/`, `pkg/`
- User-visible CLI commands and flags
- Hook behavior changes users experience
- Breaking configuration changes
- Security fixes

Summarize or omit internal-only changes:

- Template/rules/agent/skill optimization
- Comments-only changes
- Local development configuration
- Release workflow or CI pipeline modifications unless they affect users

### Required CHANGELOG entry structure

Use this structure after resolving `VERSION` and date:

````markdown
## [VERSION] - YYYY-MM-DD

### Summary
English summary in 2-3 lines.

### Breaking Changes
None, or list breaking changes.

### Added
- English additions.

### Changed
- English changes.

### Fixed
- English fixes.

### Installation & Update

```bash
moai update
moai version
```

---

## [VERSION] - YYYY-MM-DD (한국어)

### 요약
한국어 요약 2-3줄.

### 주요 변경 사항 (Breaking Changes)
없음, 또는 호환성을 깨는 변경 사항.

### 추가됨 (Added)
- 한국어 추가 사항.

### 변경됨 (Changed)
- 한국어 변경 사항.

### 수정됨 (Fixed)
- 한국어 수정 사항.

### 설치 및 업데이트 (Installation & Update)

```bash
moai update
moai version
```
````

Verification checklist:

- English section appears first.
- Korean section appears second.
- Separator is present.
- Installation commands are present in both sections.
- Previous version entry remains after the new bilingual entry.

Commit changelog separately when possible:

```bash
git add CHANGELOG.md
git commit -m "docs: update changelog for vVERSION"
```

## Phase 6: Final approval before remote operations

Display release summary:

- Current version to target version
- Commits included and key user-facing changes
- Quality gate results
- Version files changed
- Changelog status
- What remote operations will happen next

Ask the user to approve one of:

- Proceed: push release branch, create PR, merge when safe, and tag
- Pause: leave changes on local release branch
- Abort: stop without remote operations

## Phase 7: PR, merge, and tag

All remote git operations must follow GitHub Flow. Prefer delegating this phase to `subagent` with the `manager-git` role, or use a worktree-isolated `teams` task if multiple release checks run in parallel.

Required remote sequence after approval:

1. Push release branch:

```bash
git push -u origin RELEASE_BRANCH
```

2. Create PR:

```bash
gh pr create --head RELEASE_BRANCH --base main --title "release: vVERSION" --body "..."
```

PR body must include:

- Version bump
- CHANGELOG updated
- Key changes since previous tag
- Quality gates summary
- Rollback/risk notes

3. Wait for PR checks:

```bash
gh pr checks --watch
```

4. Merge with squash and delete branch:

```bash
gh pr merge --squash --delete-branch
```

5. Sync local main:

```bash
git checkout main
git pull origin main
```

6. Check whether remote tag already exists:

```bash
git ls-remote --tags origin TAG
```

7. If the tag does not exist, create and push it:

```bash
git tag TAG
git push origin TAG
```

If the tag already exists, stop and ask the user how to proceed.

## Phase 8: Release automation and GitHub release notes

### Wait for release workflow

```bash
gh run list --workflow=release.yml --limit 3
```

Retry status checks until the latest run completes or a reasonable retry limit is reached:

```bash
for i in {1..10}; do
  STATUS=$(gh run list --workflow=release.yml --limit 1 --json status --jq '.[0].status')
  if [[ "$STATUS" == "completed" ]]; then
    break
  fi
  echo "Waiting for release workflow... attempt $i/10"
  sleep 30
done
```

Verify success:

```bash
gh run list --workflow=release.yml --limit 1 --json conclusion --jq '.[0].conclusion'
gh release view TAG --json tagName,assets
```

### Replace auto-generated release notes

Use `gh release edit TAG --notes ...` with English-first bilingual content matching the CHANGELOG structure. Include:

- English title and summary
- English breaking changes/added/changed/fixed
- English install/update commands
- Korean title and summary
- Korean breaking changes/added/changed/fixed
- Korean install/update commands
- Migration note from Python version when relevant

### Verify release notes and assets

Checklist:

- English section first, Korean section second.
- Breaking Changes section present or explicitly says None/없음.
- Installation commands present in both languages.
- Required assets present:
  - `moai-adk_VERSION_darwin_arm64.tar.gz`
  - `moai-adk_VERSION_darwin_amd64.tar.gz`
  - `moai-adk_VERSION_linux_arm64.tar.gz`
  - `moai-adk_VERSION_linux_amd64.tar.gz`
  - `moai-adk_VERSION_windows_amd64.zip`
  - `moai-adk_VERSION_windows_arm64.zip`
  - `checksums.txt`
- Asset names do not include `v` before `VERSION`; update checker expects `moai-adk_VERSION_OS_ARCH`.

Manual binary download smoke test:

```bash
DOWNLOAD_URL=$(gh release view TAG --json assets --jq '.assets[] | select(.name | contains("darwin_arm64")) | .url')
curl -L -o /tmp/test-binary.tar.gz "$DOWNLOAD_URL"
tar -tzf /tmp/test-binary.tar.gz | head -5
rm /tmp/test-binary.tar.gz
```

Expected archive contents include the binary, README, and LICENSE.

## Phase 9: Local environment update

After release is verified, update local development environment if requested:

```bash
moai update --binary
moai version
```

If local templates need synchronization, ask before running:

```bash
moai update --templates-only
```

## Troubleshooting

### Go build cache permission errors

```bash
go clean -cache -testcache
go clean -cache -testcache && echo "Cache cleared successfully"
```

### Unused imports or formatting failures

Use `subagent` with the `expert-debug` role or apply minimal fixes, then run:

```bash
make fmt
go test ./...
```

### Binary download failure

Check asset naming:

```bash
gh release view TAG --json assets --jq '.assets[].name'
```

Expected archive names omit the `v` prefix before `VERSION`. If names are wrong, inspect `internal/update/checker.go` and GoReleaser config.

### Release workflow failed

```bash
gh run view --log
git tag -l "v*" | tail -5
```

Common causes: wrong tag format, GoReleaser config error, token permission issue, platform-specific build failure.

### Test artifacts in working directory

Inspect first, then ask before cleanup:

```bash
git status --porcelain
```

Do not delete generated or template files without explicit approval.

### Version inconsistency

```bash
grep -n "Version" pkg/version/version.go 2>/dev/null || true
grep -n "moai.version\|template_version" .moai/config/sections/system.yaml 2>/dev/null || true
grep -n "moai.version\|template_version" .pi/generated/source/moai-config/sections/system.yaml 2>/dev/null || true
```

All release metadata should agree after Phase 4.

## Output requirements

During execution, report phase progress in Markdown. Final output must include:

- Target version and previous version
- Branch, PR, merge, and tag status
- Quality gate results
- Version files and changelog files changed
- GitHub release URL and automation status
- Asset verification result
- Local update status
- Remaining risks or manual follow-up
