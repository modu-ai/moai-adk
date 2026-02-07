---
description: "Community PR - Fetch GitHub issues, analyze, fix, and create PRs automatically"
argument-hint: "[ISSUE_NUMBER | --all | --label LABEL] - target specific issue or batch"
type: local
allowed-tools: Read, Write, Edit, Grep, Glob, Bash, TodoWrite, AskUserQuestion, Task
model: sonnet
version: 1.0.0
---

## Community PR Configuration

- **Repository**: Auto-detected from `gh repo view --json nameWithOwner`
- **Branch prefix**: `fix/issue-{number}` for bug fixes, `feat/issue-{number}` for features
- **PR template**: Links to original issue, includes fix summary and test plan
- **Git strategy**: Reads `github.git_workflow` from `.moai/config/sections/system.yaml`

---

## EXECUTION DIRECTIVE - START IMMEDIATELY

This is a community PR command. Execute the workflow below in order. Do NOT just describe the steps - actually run the commands.

Arguments provided: $ARGUMENTS

- If ISSUE_NUMBER provided: Fix that specific issue
- If --all provided: List all open issues and let user select
- If --label LABEL provided: Filter issues by label
- If no argument: List all open issues and let user select

---

## Pre-execution Context

!gh repo view --json nameWithOwner --jq '.nameWithOwner'
!gh issue list --state open --limit 30 --json number,title,labels,body
!git branch --show-current
!git status --porcelain

@.moai/config/sections/system.yaml
@.moai/config/sections/language.yaml

---

## PHASE 1: Issue Discovery

### Step 1.1: Fetch Open Issues

Fetch all open issues from GitHub:
`gh issue list --state open --limit 50 --json number,title,labels,assignees,body,createdAt`

### Step 1.2: Issue Selection

If ISSUE_NUMBER argument provided:
- Fetch that specific issue: `gh issue view {number} --json number,title,labels,body,comments`
- Proceed directly to Phase 2

If --all or no argument:
- Display issue list as formatted table
- Use AskUserQuestion to let user select which issue(s) to fix
- Options: Individual issue numbers, or "All" for batch mode

If --label LABEL:
- Filter: `gh issue list --state open --label "{LABEL}" --json number,title,labels,body`
- Display filtered list and let user select

### Step 1.3: Issue Classification

For each selected issue, classify by type:
- **bug**: Fix existing behavior (branch prefix: `fix/issue-{number}`)
- **feature**: New functionality (branch prefix: `feat/issue-{number}`)
- **enhancement**: Improve existing feature (branch prefix: `improve/issue-{number}`)
- **docs**: Documentation only (branch prefix: `docs/issue-{number}`)

Classification based on: labels, title keywords, body content analysis.

---

## PHASE 2: Issue Analysis (Per Issue)

For each selected issue, execute Phases 2-4 sequentially.

### Step 2.1: Read Issue Details

Fetch full issue context:
`gh issue view {number} --json number,title,body,labels,comments,reactions`

### Step 2.2: Analyze Issue

Agent: expert-debug subagent (or team mode if --team flag)

Provide issue context to agent:
- Issue title and body
- Issue comments (if any)
- Relevant code files mentioned in issue
- Error messages or stack traces from issue body

Agent tasks:
- Identify root cause or scope of work
- List affected files
- Propose fix approach
- Estimate complexity (low/medium/high)

### Step 2.3: User Approval

Tool: AskUserQuestion

Display analysis summary and present options:
- Fix This Issue: Proceed to implementation
- Skip This Issue: Move to next issue in batch
- Modify Approach: Provide different fix direction
- Abort: Stop processing all issues

---

## PHASE 3: Branch and Fix

### Step 3.1: Create Feature Branch

Read `github.git_workflow` from system.yaml to determine branching:

**github_flow or gitflow**:
1. Ensure on main (or develop for gitflow): `git checkout main && git pull origin main`
2. Create branch: `git checkout -b {prefix}/issue-{number}`
3. Verify: `git branch --show-current`

**main_direct**:
- Stay on main branch, no branch creation
- All fixes committed directly

### Step 3.2: Implement Fix

Agent: Delegate to appropriate agent based on issue classification.

- Bug fix: expert-debug subagent
- Feature: expert-backend or expert-frontend subagent (based on affected files)
- Enhancement: expert-refactoring subagent
- Docs: manager-docs subagent

If --team flag is set:
- Use team mode with multiple agents working in parallel
- See fix.md --team and team-debug.md for competing hypothesis investigation

Agent receives:
- Full issue context (title, body, comments)
- Analysis from Phase 2
- Affected files list
- Branch name and git strategy context

### Step 3.3: Verify Fix

After implementation:
1. Run tests: Language-specific test command (go test, npm test, pytest, etc.)
2. Run linter: Language-specific lint command
3. Verify no regressions: `go vet ./...` (for Go projects)

If tests fail:
- Retry fix with error context (max 3 attempts)
- If still failing: Report to user with AskUserQuestion (retry, skip, abort)

### Step 3.4: Commit Changes

Agent: manager-git subagent

Commit message format:
```
fix(scope): description of fix

Fixes #{issue_number}

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
```

The `Fixes #N` keyword auto-closes the issue when PR is merged.

---

## PHASE 4: Create PR (Strategy-Aware)

### Step 4.1: Push and Create PR

Read `github.git_workflow` from system.yaml:

**github_flow**:
1. Push branch: `git push -u origin {prefix}/issue-{number}`
2. Create PR: `gh pr create --title "fix: {issue title}" --body "$(body)"`
   - Body includes: Fix summary, test plan, `Fixes #{number}` reference
   - Base: main
   - Labels: auto-detected + issue labels

**gitflow**:
1. Push branch: `git push -u origin {prefix}/issue-{number}`
2. Create PR targeting develop: `gh pr create --base develop --title "fix: {issue title}" --body "$(body)"`

**main_direct**:
1. Push to main: `git push origin main`
2. No PR created
3. Issue is auto-closed if commit message contains `Fixes #{number}`

### Step 4.2: Link Issue to PR

If PR was created:
- Add comment to issue: `gh issue comment {number} --body "Fix submitted in PR #{pr_number}"`
- Add labels to PR matching issue labels

### Step 4.3: Return to Main

After PR creation:
1. Checkout main: `git checkout main`
2. Ready for next issue in batch mode

---

## PHASE 5: Batch Completion Report

After all issues are processed, display summary:

```markdown
## Community PR: Complete

### Issues Processed
| Issue | Title | Status | PR |
|-------|-------|--------|-----|
| #123 | Fix login bug | PR Created | #456 |
| #124 | Add dark mode | Skipped | - |
| #125 | Update docs | PR Created | #457 |

### Next Steps
- Review PRs on GitHub
- After PR merge: /moai release to publish new version
- Check CI status: gh pr checks {number}
```

### Next Steps Options

Tool: AskUserQuestion

- Review PRs (open GitHub)
- Merge All PRs (squash merge each)
- Create Release (/moai release)
- Process More Issues (/moai cpr --all)

---

## BATCH AUTO-MERGE (When --merge flag set)

If --merge flag is provided, after all PRs are created:

1. For each PR:
   a. Wait for CI: `gh pr checks {number} --watch`
   b. If passing: `gh pr merge {number} --squash --delete-branch`
   c. If failing: Skip and report

2. After all merges:
   - Checkout main: `git checkout main && git pull origin main`
   - Suggest release: "All PRs merged. Run /moai release to publish."

---

## Key Rules

- **Git strategy aware**: Reads `github.git_workflow` from config
- **Issue linking**: Always include `Fixes #{number}` in commit/PR
- **Branch per issue**: Each issue gets its own branch (except main_direct)
- **Test verification**: All fixes must pass tests before PR creation
- **Batch safe**: Process multiple issues sequentially without conflicts
- **[HARD] Agent delegation**: Implementation MUST be delegated to specialized agents
- **[HARD] User approval**: Each issue fix requires user approval before implementation

---

## BEGIN EXECUTION

Start Phase 1 now. Fetch open issues and display to user.
