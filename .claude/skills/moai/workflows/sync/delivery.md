---
description: "Sync Phase 3~4 — Git Operations and Delivery (CI mirror, push/PR, auto-merge), Completion, Team Mode, Graceful Exit, Test Scenarios, and Custom Harness."
user-invocable: false
metadata:
  parent: moai-workflow-sync
  phase: "Phase 3~4: Git Delivery, Completion, and Auxiliary"
---

<!-- TRACE PROBE: workflow-split baseline trace mechanism -->
<!-- Activated by MOAI_TRACE_PHASES=1 environment variable -->
<!-- Emits one line per Phase entry/exit to stderr in format: [trace] /moai sync Phase <N> <enter|exit> -->

### Phase 3: Git Operations and Delivery

#### Step 3.0: Detect Git Workflow Strategy

Read `github.git_workflow` from `.moai/config/sections/system.yaml`. This determines how changes are delivered.

| Strategy | Branch Model | PR Behavior | Best For |
|----------|-------------|-------------|----------|
| github_flow | Feature branches off main | Auto-create PR to main | Team/OSS projects |
| main_direct | Direct commits to main | No PR created | Solo development |
| gitflow | develop/release/hotfix branches | PR to appropriate base | Enterprise projects |

Default strategy (if not configured): `github_flow`

Also read `github.spec_git_workflow` to determine SPEC branch handling:
- `feature_branch`: Each SPEC gets its own branch (recommended for github_flow/gitflow)
- `main_direct`: SPEC changes committed to current branch (only when git_workflow is main_direct)

#### Step 3.1: Commit Changes

Agent: manager-git subagent

- Stage all changed document files, reports, README, docs/
- Create single commit with descriptive message listing synchronized documents, project repairs, and SPEC updates
- Commit message language follows `language.git_commit_messages` setting
- Verify commit with git log

#### Step 3.1.1: Context Memory Generation in Git Commits

Purpose: Embed structured context within git commit operations to enable seamless session resumption across development cycles.

**Context Collection Process:**

1. **Decision Tracking**: Gather all decisions made during the sync phase
   - Documentation choices and rationale
   - SPEC update approach and divergence handling
   - Project improvement selections
   - Quality trade-offs accepted or deferred

2. **Constraint Discovery**: Record any constraints identified
   - Formatting requirements discovered
   - API documentation standards applied
   - Platform-specific considerations
   - Technology limitations encountered

3. **Gotcha Documentation**: Note issues found during documentation review
   - Outdated references in existing documentation
   - Missing API documentation sections
   - Inconsistencies between code and docs
   - Breaking changes requiring user notification

4. **Pattern Usage**: Document patterns applied during sync
   - Documentation templates used
   - Code-to-doc mapping strategies
   - Mermaid diagram patterns for architecture
   - README.md structure improvements

**Commit Format for Sync Phase:**

All sync commits MUST include structured context using this format:

```
docs(sync): [brief description of changes]

## SPEC Reference
SPEC: SPEC-XXX
Phase: SYNC
Timestamp: ISO-8601 timestamp

## Context (AI-Developer Memory)
- Decision: [documentation decision 1]
- Decision: [documentation decision 2]
- Pattern: [pattern 1 applied]
- Pattern: [pattern 2 applied]
- Constraint: [constraint discovered]
- Gotcha: [issue found and how resolved]

## Affected Areas
- Documents Updated: [count]
- SPEC Status: [completed|in-progress]
- Coverage Impact: [change or percentage]
```

#### Step 3.1.5: Local CI Mirror Validation (Pre-PR Gate)

Purpose: Replicate CI checks locally before pushing and creating a PR to catch failures fast, without waiting for slow remote CI. Windows-specific tests are skipped (cannot run locally).

**Trigger condition**: Only run when a PR is about to be created (github_flow feature branch, gitflow feature/release/hotfix). Skip for `main_direct` strategy and direct pushes to main/develop.

##### Step 3.1.5.1: Discover CI Configuration

Read `.github/workflows/` to auto-detect CI jobs:
- If `ci.yml` (or any CI file) exists: parse jobs, steps, and commands
- If no CI config found: skip this phase entirely, log "No CI config detected"

Build a local execution plan mapping each CI job to its local equivalent:

| CI Job | CI Runner | Local Equivalent | Skippable |
|--------|-----------|-----------------|-----------|
| test (ubuntu) | ubuntu-latest | Local OS tests | No (run on current OS) |
| test (macos) | macos-latest | Local OS tests | No (identical on macOS) |
| test (windows) | windows-latest | **SKIP** | Yes — cannot run locally |
| lint | ubuntu-latest | Local lint | No |
| build (cross-compile) | ubuntu-latest | Local cross-compile | No |

##### Step 3.1.5.2: Run Local Equivalents in Parallel

**Go project** (detected via `go.mod`):

Launch all checks in parallel:

```bash
# Check 1: go vet (mirrors CI step)
go vet ./...

# Check 2: Tests with race detector (mirrors CI test job)
go test -race -coverprofile=coverage.out -covermode=atomic ./...

# Check 3: golangci-lint (mirrors CI lint job)
# Auto-detect if golangci-lint is available
which golangci-lint && golangci-lint run --timeout=5m \
  || echo "SKIP: golangci-lint not installed (install via your project's pinned version)"

# Check 4: Cross-compile all CI targets (mirrors CI build job)
# Replace <your-module> with your main package path (e.g. ./cmd/<your-binary>/).
# Replicate whatever GOOS/GOARCH targets your CI build matrix declares; the
# example below shows the common 5-target matrix — run them in parallel, CGO_ENABLED=0 for all.
GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o /tmp/ci-build-linux-amd64     ./<your-module>/ &
GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o /tmp/ci-build-linux-arm64     ./<your-module>/ &
GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o /tmp/ci-build-darwin-amd64    ./<your-module>/ &
GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o /tmp/ci-build-darwin-arm64    ./<your-module>/ &
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o /tmp/ci-build-windows-amd64.exe ./<your-module>/ &
wait
```

**Python project** (detected via `pyproject.toml`):

```bash
pytest --tb=short
ruff check . && ruff format --check .
mypy . --ignore-missing-imports
```

**TypeScript/JavaScript project** (detected via `package.json`):

```bash
npm test -- --run
npm run lint
npm run build
```

**Other languages**: Run the standard test + lint + build commands discovered from CI config.

**Cross-platform build targets**: If CI config shows `strategy.matrix` with multiple `os` or `GOOS/GOARCH` values, replicate all cross-compile targets using the local toolchain.

##### Step 3.1.5.3: Skipped Checks Report

Always report what was skipped and why:

```
CI Mirror: Skipped checks
- test (windows-latest): Cannot run Windows tests locally — will be verified by remote CI
- lint: golangci-lint not installed — install via your project's pinned version
```

##### Step 3.1.5.4: Evaluate Results

**All checks pass**: Proceed to Step 3.2 automatically. Log "Local CI mirror: PASS".

**Any check fails**: Present failure summary via AskUserQuestion:

- Fix now — delegate to manager-develop subagent with failure details, then re-run CI mirror
- Push anyway — proceed to Step 3.2 with warning embedded in PR description
- Abort — exit sync workflow, preserve commit (allow local fix and re-run)

**golangci-lint not installed**: Treat as warning (not failure). Proceed to Step 3.2 with a note in the PR description: "Local lint check skipped: golangci-lint not installed."

##### Step 3.1.5.5: Embed Results in PR Description

Pass CI mirror results to Step 3.2 for inclusion in the PR body:

```markdown
## Local CI Mirror Results
| Check | Status | Notes |
|-------|--------|-------|
| go vet | ✅ Pass | |
| go test -race (macOS) | ✅ Pass | Coverage: 87% |
| golangci-lint | ✅ Pass | |
| build linux/amd64 | ✅ Pass | |
| build linux/arm64 | ✅ Pass | |
| build darwin/amd64 | ✅ Pass | |
| build darwin/arm64 | ✅ Pass | |
| build windows/amd64 | ✅ Pass | |
| test (windows) | ⏭ Skipped | Cannot run locally |
```

#### Step 3.2: Push and Deliver (Strategy-Aware)

Behavior varies based on `github.git_workflow` setting and current branch context.

**Base Branch Resolution** (applies to all strategies below):
1. Read `git_strategy.mode` from `.moai/config/sections/git-strategy.yaml`
2. Resolve `main_branch`:
   - If `git_strategy.{mode}.main_branch` exists: use that value
   - If missing (e.g., `manual` mode): default to `main`
3. Use `{main_branch}` in all branch checkout and PR creation commands below

##### Strategy: github_flow

Detect current branch:

**Feature branch** (any branch other than main):
1. Push branch to remote: `git push -u origin <branch>`
2. Check if PR already exists: `gh pr list --head <branch> --json number`
3. If no PR exists: Create PR via `gh pr create`
   - Title: Derived from SPEC title or branch name
   - Body: Include sync summary, files changed, quality report, deployment readiness notes (migrations, env changes, breaking changes)
   - If SPEC metadata contains `issue_number` (non-zero): Include `Fixes #{issue_number}` in PR body footer for automatic Issue closure on merge
   - Base: {main_branch}
   - Labels: auto-detected from changed files
4. If PR exists: Update with comment summarizing sync changes
5. Display PR URL to user

**Main branch** (direct commit):
- Push directly: `git push origin {main_branch}`
- Display push confirmation
- Note: Direct main commits are permitted but feature branches are recommended

**Worktree context** (detected from git directory structure):
- Push worktree branch to remote
- Create PR if not exists (same as feature branch flow)
- Display PR URL and worktree context

##### Strategy: main_direct

All commits go directly to main, no PRs:
1. Push to main: `git push origin main`
2. Display push confirmation
3. No PR created regardless of branch name

##### Strategy: gitflow

Detect current branch type and route accordingly:

**feature/* branch** → PR to `develop`:
1. Push branch: `git push -u origin <branch>`
2. Create or update PR targeting `develop` branch
3. Display PR URL

**release/* branch** → PR to `main`:
1. Push branch: `git push -u origin <branch>`
2. Create or update PR targeting `main` branch
3. Display PR URL

**hotfix/* branch** → PR to `main` (and back-merge to develop):
1. Push branch: `git push -u origin <branch>`
2. Create or update PR targeting `main` branch
3. After merge: Create follow-up PR to `develop` for back-merge
4. Display PR URLs

**develop branch** → Push directly:
1. Push to develop: `git push origin develop`
2. Display push confirmation

**main branch** → Error:
- Direct commits to main are not allowed in gitflow
- Suggest creating a hotfix or release branch instead

#### Step 3.3: PR Ready Transition (Team Mode)

Only applies when a PR was created in Step 3.2:

- If Team mode enabled and PR is draft: Transition to ready via `gh pr ready`
- Assign reviewers and labels if configured
- If Team mode disabled: Do NOT automatically transition (user controls readiness)

#### Step 3.3.5: Return to Base Branch (Post-PR Cleanup)

After PR/MR creation (Step 3.2) and optional ready transition (Step 3.3), return to the base branch to leave the working directory in a clean state:

**github_flow**: `git checkout {main_branch} && git pull origin {main_branch}`
**gitflow**: `git checkout develop && git pull origin develop` (for feature branches), `git checkout main && git pull origin main` (for release/hotfix)
**main_direct**: No branch switch needed (already on main)

This ensures the developer's working directory is on the base branch, ready for the next task. The feature branch remains on the remote for review.

Remote branch cleanup after merge is handled by the hosting platform's auto-delete setting (GitHub: "Automatically delete head branches", GitLab: "Delete source branch when merge request is accepted", Bitbucket: "Close source branch"). Local branch cleanup is left to the developer (`git branch -d <branch>`).

#### Step 3.4: Auto-Merge Behavior

Only applies when a PR was created in Step 3.2.

##### Auto-Merge Trigger Conditions

Auto-merge trigger conditions:
- `is_worktree_context == true` AND `--no-merge` flag NOT set
- OR `--merge` flag explicitly set (deprecated, logged as warning)

When auto-merge is triggered:
1. Verify all CI/CD checks pass (gh pr checks)
2. Verify zero merge conflicts (gh pr view --json mergeable)
3. If all checks pass: Execute `gh pr merge --squash --delete-branch`
4. If checks fail: Report error with recovery command, do NOT merge

##### Flag Behavior

- `--no-merge`: Skip auto-merge even in worktree context. PR is created but not merged.
- `--merge`: Deprecated. Logs warning: "The --merge flag is deprecated. Auto-merge is now the default for worktree contexts."

##### Auto-Merge Execution

1. Check CI/CD status via `gh pr checks --watch` (wait for completion)
2. Check merge conflicts via `gh pr view --json mergeable`
3. If passing and mergeable: Execute `gh pr merge --squash --delete-branch`
4. Checkout target branch, fetch latest
5. Verify local is synchronized with remote

##### Auto-Merge Failures

- If CI/CD fails: Report failure, display error details, do NOT merge
- If merge conflicts: Report conflicts, provide manual resolution guidance, do NOT merge
- If approvals missing (Team mode): Report pending approvals, do NOT merge

##### Post-Merge Automatic Cleanup

Condition: Auto-merge succeeded AND `workflow.worktree.auto_cleanup == true`

Steps:
1. Detect worktree path for current SPEC-ID from registry
2. Execute cleanup equivalent to `moai worktree done SPEC-{ID} --auto --delete-branch`:
   - Remove worktree directory
   - Remove feature branch (already deleted by --delete-branch in merge)
   - Update worktree registry
3. Log cleanup result

Error handling:
- Cleanup failure does NOT block or affect merge result
- On failure: Log warning with manual cleanup command
- Message: "Worktree cleanup warning: {error}. Manual: `moai worktree done SPEC-{ID}`"

### Phase 4: Completion and Next Steps

#### Completion Report

Display summary including:
- Git workflow strategy used (github_flow, main_direct, or gitflow)
- Sync mode and scope
- Files updated and created
- Project improvements made
- Documents updated
- Reports generated
- Backup location
- PR URL (if created) or push target (if direct push)

#### Context-Aware Next Steps

Tool: AskUserQuestion with options tailored to delivery result:

**If PR was created (github_flow feature branch, or gitflow):**
- Review PR on GitHub
- Auto-Merge PR (/moai sync --merge)
- Create Next SPEC (/moai plan)
- Start New Session (/clear)

**If direct push (main_direct, or github_flow main branch):**
- Create Next SPEC (/moai plan)
- Continue Development
- Start New Session (/clear)

**If worktree context:**
- Review PR in Browser
- Return to Main Directory
- Remove This Worktree

---

## Team Mode

The sync phase always uses sub-agent mode (manager-docs), even when --team is active for other phases. Documentation synchronization requires sequential consistency and a single authoritative view of project state.

For rationale and details, see ${CLAUDE_SKILL_DIR}/../team/sync.md.

---

## Graceful Exit

When user aborts at any decision point:

- No changes made to documents, Git history, or branch state
- Project remains in current state
- Display retry command: /moai sync [mode]
- Exit with code 0

---

## Completion Criteria

All of the following must be verified:

- Phase 0: Deployment readiness verified (tests, migrations, env changes, backward compatibility)
- Phase 0.5: Quality verification completed (tests, linter, type checker, deep code review with auto-fix)
- Phase 0.55: Security scan completed (if security-sensitive files changed)
- Phase 0.7: Coverage analysis completed (measurement, gap analysis, test generation, verification)
- Phase 1: Prerequisites verified, project analyzed, divergence analysis completed, sync plan approved by user
- Phase 2: Safety backup created and verified, documents synchronized, SPEC documents updated per lifecycle level, project documents updated (if applicable), quality verified, SPEC status updated
- Phase 3: Changes committed, local CI mirror validated (Step 3.1.5: vet + test-race + lint + cross-compile — Windows skipped), delivered per git_workflow strategy (PR created for github_flow/gitflow, direct push for main_direct), auto-merge executed (if flagged and PR exists)
- Phase 4: Completion report displayed with delivery result, appropriate next steps presented based on strategy and context

---

## Test Scenarios

### Normal Flow
**Prompt**: "/moai sync SPEC-AUTH-001"
**Expected Result**:
- Phase 0: Pre-sync quality gate passes (tests, lint)
- Phase 0.5: Quality verification confirms TRUST 5 compliance
- Phase 1: Divergence analysis shows implementation matches SPEC
- Decision Point: User approves sync plan
- Phase 2: Documentation updated (README, CHANGELOG, API docs)
- Phase 2.2.1: SPEC status updated to "implemented"
- Phase 3: Commits created, PR opened with summary

### Partial Implementation Flow
**Prompt**: "/moai sync SPEC-AUTH-001" (only backend implemented, frontend pending)
**Expected Result**:
- Phase 1.5: Divergence detected - 3/5 acceptance criteria met
- Sync plan notes partial implementation
- SPEC status updated to "in-progress" (not "implemented")
- Documentation reflects completed portions only
- PR description notes remaining work

### Error Flow
**Prompt**: "/moai sync" (no SPEC specified, uncommitted changes exist)
**Expected Result**:
- Auto-detect: Finds uncommitted changes on current branch
- AskUserQuestion: "Sync changes on current branch?"
- If user confirms, syncs based on git diff
- If no changes found, reports "Nothing to sync"

---

## Related Skills

정적 routing:

- **moai-workflow-ci-loop** — Phase 4 (`gh pr create`) 성공 후 CI watch + auto-fix loop을 자동 호출하는 skill. HARD invocation contracts: `.claude/rules/moai/workflow/ci-watch-protocol.md` + `.claude/rules/moai/workflow/ci-autofix-protocol.md`. 30s polling, 30분 hard timeout, required vs auxiliary check 분류 후 ready-to-merge handoff 또는 max 3-iteration auto-fix 시도, semantic 실패는 즉시 escalation.

이 skill은 `auto` 모드 sync에서 PR 생성 직후 무조건 호출되며, invocation contract에 따라 orchestrator가 다음을 보장한다: gh 인증 확인 → `.github/required-checks.yml` 존재 확인 → 양의 정수 PR 번호 → 90s 이내 활성 watch 부재.

---

Version: 3.8.0
Updated: 2026-05-17
Changes: Added test scenarios (3.7.0) + Related Skills section (3.8.0) + consolidated moai-workflow-ci-watch reference to moai-workflow-ci-loop per the skill consolidation policy (3.9.0).

---

## Custom Harness Extension (Optional)

@.moai/harness/sync-extension.md

*(이 파일은 `/moai project --harness`로 생성됩니다. 파일이 없으면 자동으로 skip됩니다.)*
