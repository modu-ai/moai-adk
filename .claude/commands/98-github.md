---
description: "GitHub Workflow - Manage issues and review PRs with Agent Teams"
argument-hint: "issues [--all | --label LABEL | NUMBER | --merge | --solo | --tmux] | pr [--all | NUMBER | --merge | --solo]"
type: local
allowed-tools: Read, Write, Edit, Grep, Glob, Bash, AskUserQuestion, Agent, TeamCreate, SendMessage, TaskCreate, TaskUpdate, TaskList, TaskGet, TeamDelete
version: 2.0.0
---

## GitHub Workflow Configuration

- **Repository**: Auto-detected from `gh repo view --json nameWithOwner`
- **Default mode**: Agent Teams (falls back to sub-agent if AGENT_TEAMS unavailable)
- **Branch prefix**: `fix/issue-{number}` for bugs, `feat/issue-{number}` for features
- **Git strategy**: Reads `github.git_workflow` from `.moai/config/sections/system.yaml`

---

## EXECUTION DIRECTIVE - START IMMEDIATELY

This is the GitHub workflow command. Parse $ARGUMENTS and execute immediately.

### Argument Parsing

First word determines sub-command:

- **issues** (aliases: issue, fix-issues): Issue fixing workflow
- **pr** (aliases: review, pull-request): PR code review workflow
- No sub-command: Use AskUserQuestion to let user choose

Remaining arguments become sub-command arguments:

- `--all`: Process all open items
- `--label LABEL`: Filter by label
- `--solo`: Force sub-agent mode (skip Agent Teams)
- `--merge`: Auto-merge after CI passes (issues: merge created PRs; pr: merge selected PRs)
- `NUMBER`: Target specific issue or PR number

---

## Pre-execution Context

!gh repo view --json nameWithOwner --jq '.nameWithOwner'
!git branch --show-current
!git status --porcelain
!printenv CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS 2>/dev/null && echo "(CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS is set)" || echo "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS not set"

@.moai/config/sections/system.yaml
@.moai/config/sections/language.yaml
@.moai/config/sections/workflow.yaml

---

## Team Mode (Default)

Agent Teams mode is the DEFAULT for this workflow. No `--team` flag required.

Prerequisites check: Read the values injected by the Pre-execution Context above:
1. AGENT_TEAMS status is shown by the `!printenv CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` output above
2. `workflow.team.enabled` is shown in the workflow.yaml loaded above

If both prerequisites met: Use Agent Teams mode
If either prerequisite missing OR `--solo` flag: Fall back to sub-agent mode

---

# SUB-COMMAND: issues

Purpose: Fetch GitHub issues, analyze root cause, implement fixes, and create PRs.

## Issues Phase 1: Issue Discovery

### Step 1.1: Fetch Open Issues

Fetch all open issues from GitHub:
`gh issue list --state open --limit 50 --json number,title,labels,assignees,body,createdAt`

### Step 1.2: Issue Selection

If NUMBER argument provided:
- Fetch specific issue: `gh issue view {number} --json number,title,labels,body,comments`
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

## Issues Phase 1.5: Existing Work Detection

Before starting analysis, check for existing bot work on each selected issue.

### Step 1.5.1: Detect Bot Branches and PRs

```bash
# Check for @claude bot branches
git ls-remote --heads origin | grep -i "claude/issue-{number}"

# Check for other automated fix branches
git ls-remote --heads origin | grep -E "(fix|feat|improve|docs)/issue-{number}"

# Check for PRs referencing this issue
gh pr list --search "#{number}" --state all --json number,title,headRefName,state,author
```

### Step 1.5.2: Decision Matrix

| Bot Branch | PR Exists | PR State | Action |
|-----------|-----------|----------|--------|
| Yes | No | - | AskUser: Create PR from bot branch / Redo fix |
| Yes | Yes | OPEN | AskUser: Review existing PR / Redo fix |
| Yes | Yes | MERGED | Skip issue (already resolved) |
| No | Yes | OPEN | AskUser: Review existing PR / Continue work |
| No | Yes | MERGED | Skip issue (already resolved) |
| No | No | - | Proceed with normal Phase 2 analysis |

When existing bot work is found, use AskUserQuestion:
- Use existing fix (Recommended): Verify tests pass on bot branch, then create PR
- Redo fix from scratch: Ignore existing branches, run full analysis and implementation
- Skip this issue: Move to next selected issue

## Issues Phase 2: Analysis

### Team Mode (Default)

Create a team for parallel issue analysis:

```
TeamCreate(team_name: "github-issues-{repo-slug}")
```

For each selected issue, create tasks:
```
TaskCreate: "Analyze issue #{number}: {title}"
TaskCreate: "Implement fix for issue #{number}" (blocked by analysis task)
TaskCreate: "Verify fix for issue #{number}" (blocked by implementation task)
```

Spawn analysis teammates in parallel (one per issue, max 3 concurrent):

```
Agent(
  subagent_type: "team-reader",
  team_name: "github-issues-{repo-slug}",
  name: "analyst-{number}",
  mode: "plan",
  prompt: "Analyze GitHub issue #{number}.
    Title: {title}
    Body: {body}
    Comments: {comments}
    Explore the codebase to identify root cause, affected files, and fix approach.
    Mark your task completed via TaskUpdate and send findings via SendMessage."
)
```

After analysis completes, spawn implementation teammates:

```
Agent(
  subagent_type: "team-coder",  // role (backend/frontend) specified in prompt
  team_name: "github-issues-{repo-slug}",
  name: "fixer-{number}",
  mode: "acceptEdits",
  isolation: "worktree",
  prompt: "Fix GitHub issue #{number} based on analysis findings.
    Analysis: {analyst_findings}
    Affected files: {file_list} (use project-root-relative paths only)
    Create feature branch: {prefix}/issue-{number}
    Write tests, implement fix, verify tests pass.
    [HARD] Use relative paths only — do NOT use absolute paths or cd to specific directories.
    Your CWD is already set to the project root.
    Mark your task completed via TaskUpdate and send results via SendMessage."
)
```

### Sub-agent Mode (--solo or fallback)

Delegate to appropriate expert agent based on classification:
- Bug fix: expert-debug subagent
- Feature: expert-backend or expert-frontend subagent
- Enhancement: expert-refactoring subagent
- Docs: manager-docs subagent

## Issues Phase 3: Branch and Fix

### Step 3.1: Create Feature Branch

Read `github.git_workflow` from system.yaml:

**github_flow or gitflow**:
1. Ensure on main (or develop for gitflow): `git checkout main && git pull origin main`
2. Create branch: `git checkout -b {prefix}/issue-{number}`

**main_direct**:
- Stay on main, no branch creation

### Step 3.2: Verify Fix

After implementation:
1. Run tests: Language-specific test command
2. Run linter: Language-specific lint command
3. If tests fail: Retry with error context (max 3 attempts)
4. If still failing: AskUserQuestion (retry, skip, abort)

### Step 3.3: Commit Changes

Delegate to manager-git subagent.

Commit message format:
```
fix(scope): description

Fixes #{issue_number}

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
```

## Issues Phase 4: Create PR

Read `github.git_workflow` from system.yaml:

**github_flow**:
1. Push: `git push -u origin {prefix}/issue-{number}`
2. Create PR: `gh pr create --title "fix: {issue title}" --body "$(body)"`
   - Body includes: Fix summary, test plan, `Fixes #{number}` reference

**gitflow**:
1. Push and create PR targeting develop

**main_direct**:
1. Push to main directly

Link issue to PR:
- `gh issue comment {number} --body "Fix submitted in PR #{pr_number}"`

After PR: `git checkout main` to prepare for next issue.

## Issues Phase 5: Issue Closure

After PRs are merged (manually or via --merge flag), close issues with multilingual comments.

### Step 5.1: Detect Language

Read user's `conversation_language` from `.moai/config/sections/language.yaml`.
Supported languages: `en`, `ko`, `ja`, `zh`. Unsupported codes fall back to English.

### Step 5.2: Generate Success Comment

Use `internal/i18n.CommentGenerator` to produce a multilingual comment:
- Implementation summary from SPEC or commit messages
- PR link: `#<pr_number>`
- Merge timestamp with timezone
- Test coverage percentage (if available)

Comment templates are language-aware. Example (Korean):
```
Issue resolved successfully!

Implementation: Added user authentication
Related PR: #456
Merge time: 2026-02-16 16:30 KST
```

### Step 5.3: Close Issue

Use `internal/github.IssueCloser` to execute the 3-step closure:
1. Post comment: `gh issue comment {number} --body "{comment}"`
2. Add label: `gh issue edit {number} --add-label resolved`
3. Close issue: `gh issue close {number}`

Each step retries up to 3 times with exponential backoff (2s, 4s, 8s).
Label failure is non-critical (closure continues even if labeling fails).

### Step 5.4: Update SPEC Status

If a SPEC document exists for this issue (`SPEC-ISSUE-{number}`):
- Update SPEC metadata `status` to `completed`

## Issues Phase 6: Cleanup and Report

If team mode was used:
1. Shutdown all teammates via SendMessage(type: "shutdown_request")
2. TeamDelete to clean up resources

Display batch summary:
```markdown
## GitHub Issues: Complete

| Issue | Title | Status | PR | Closed |
|-------|-------|--------|-----|--------|
| #123 | Fix login bug | Merged | #456 | Yes |
| #124 | Add dark mode | Skipped | - | No |
```

AskUserQuestion for next steps:
- Review PRs on GitHub
- Merge All PRs (if --merge flag)
- Process More Issues
- Done

## Issues Phase 7 (Optional): tmux Parallel Development

When `--tmux` flag is provided and tmux is available on the system:

### Step 7.1: Detect tmux

Use `internal/tmux.Detector` to check:
- `tmux.IsAvailable()` - verify tmux binary exists
- `tmux.Version()` - ensure compatible version
- If unavailable, fall back to sequential execution with warning

### Step 7.2: Create Session

Use `internal/tmux.SessionManager` to create a multi-pane session:
- Session name: `github-issues-{timestamp}`
- One pane per issue worktree (max 3 visible via vertical splits)
- Additional panes overflow to horizontal splits
- Each pane auto-executes: `moai worktree go SPEC-ISSUE-{number}`

Layout algorithm:
- Panes 1-3: vertical splits (`tmux split-window -v`)
- Panes 4+: horizontal splits (`tmux split-window -h`)
- Focus returns to first pane after creation

---

## Go Package Integration Reference

### internal/i18n (Multilingual Comments)

```go
// Create generator (one-time setup)
gen := i18n.NewCommentGenerator()

// Generate comment in user's language
comment, err := gen.Generate(langCode, &i18n.CommentData{
    Summary:         "Added user authentication",
    PRNumber:        456,
    IssueNumber:     123,
    MergedAt:        time.Now(),
    TimeZone:        "KST",
    CoveragePercent: 92,
})
```

Supported languages: en, ko, ja, zh. Unknown codes fall back to English.

### internal/github (Issue Closure)

```go
// Create closer with retry configuration
closer := github.NewIssueCloser(repoRoot,
    github.WithMaxRetries(3),
    github.WithRetryDelay(2 * time.Second),
)

// Close issue with generated comment
result, err := closer.Close(ctx, issueNumber, comment)
// result.CommentPosted, result.LabelAdded, result.IssueClosed
```

### internal/tmux (Session Management)

```go
// Check availability
detector := tmux.NewDetector()
if !detector.IsAvailable() {
    // Fall back to sequential mode
}

// Create session
mgr := tmux.NewSessionManager()
result, err := mgr.Create(ctx, &tmux.SessionConfig{
    Name:       "github-issues-20260216-1630",
    Panes:      panes,  // []tmux.PaneConfig
    MaxVisible: 3,
})
```

---

# SUB-COMMAND: pr

Purpose: Fetch PRs, perform multi-perspective code review, and submit review comments.

## PR Phase 1: PR Discovery

### Step 1.1: Fetch Open PRs

`gh pr list --state open --limit 30 --json number,title,author,labels,additions,deletions,changedFiles,headRefName`

### Step 1.2: PR Selection

If NUMBER argument provided:
- Fetch specific PR: `gh pr view {number} --json number,title,body,files,commits,reviews`
- Proceed to Phase 2

If --all or no argument:
- Display PR list as formatted table (number, title, author, +/- lines, files changed)
- Use AskUserQuestion to let user select PR(s) to review

### Step 1.3: Fetch PR Details

For each selected PR:
- Get full diff: `gh pr diff {number}`
- Get changed files: `gh pr view {number} --json files --jq '.files[].path'`
- Get existing reviews: `gh pr view {number} --json reviews`

## PR Phase 2: Code Review

### Team Mode (Default)

Create a review team for parallel multi-perspective analysis:

```
TeamCreate(team_name: "github-pr-review-{number}")
```

Create review tasks:
```
TaskCreate: "Security review of PR #{number}"
TaskCreate: "Performance review of PR #{number}"
TaskCreate: "Quality and correctness review of PR #{number}"
```

Spawn 3 reviewers in parallel:

```
Agent(
  subagent_type: "team-reader",
  team_name: "github-pr-review-{number}",
  name: "security-reviewer",
  mode: "plan",
  prompt: "You are a security reviewer for PR #{number} in {repo}.
    Review the following diff for security vulnerabilities:
    - Injection risks (SQL, XSS, command injection)
    - Authentication/authorization issues
    - Sensitive data exposure
    - OWASP Top 10 compliance
    Changed files: {file_list}
    Diff: {diff_content}
    Mark task completed and send findings via SendMessage."
)

Agent(
  subagent_type: "team-reader",
  team_name: "github-pr-review-{number}",
  name: "perf-reviewer",
  mode: "plan",
  prompt: "You are a performance reviewer for PR #{number} in {repo}.
    Review the following diff for performance issues:
    - Algorithm complexity (O(n^2) loops, unnecessary allocations)
    - Database query patterns (N+1, missing indexes)
    - Memory leaks and resource management
    - Concurrency issues (race conditions, deadlocks)
    Changed files: {file_list}
    Diff: {diff_content}
    Mark task completed and send findings via SendMessage."
)

Agent(
  subagent_type: "team-reader",
  team_name: "github-pr-review-{number}",
  name: "quality-reviewer",
  mode: "plan",
  prompt: "You are a code quality reviewer for PR #{number} in {repo}.
    Review the following diff for quality issues:
    - Code correctness and edge cases
    - Test coverage for changes
    - Naming conventions and readability
    - Error handling completeness
    - API contract consistency
    Changed files: {file_list}
    Diff: {diff_content}
    Mark task completed and send findings via SendMessage."
)
```

### Sub-agent Mode (--solo or fallback)

Delegate sequentially:
1. expert-security subagent: Security analysis of PR diff
2. expert-performance subagent: Performance analysis
3. manager-quality subagent: Code quality review

## PR Phase 3: Synthesize and Submit Review

After all reviewers complete:

1. Collect findings from all perspectives
2. Classify issues by severity:
   - **Critical**: Must fix before merge (security vulnerabilities, data loss risks)
   - **Important**: Should fix (performance issues, missing error handling)
   - **Suggestion**: Nice to have (naming, style, minor improvements)
3. Format as GitHub review

### Submit Review

Use AskUserQuestion to confirm review action:
- Approve: Submit approval with summary
- Request Changes: Submit with required changes
- Comment Only: Submit as comment without approval decision
- Skip: Do not submit review

If approved, submit via:
```bash
gh pr review {number} --approve --body "$(review_body)"
# OR
gh pr review {number} --request-changes --body "$(review_body)"
# OR
gh pr review {number} --comment --body "$(review_body)"
```

For inline comments on specific lines:
```bash
gh api repos/{owner}/{repo}/pulls/{number}/reviews \
  --method POST \
  --field body="Review summary" \
  --field event="COMMENT" \
  --field comments="[{\"path\":\"file.go\",\"line\":42,\"body\":\"Issue description\"}]"
```

## PR Phase 4: Cleanup and Report

If team mode was used:
1. Shutdown all reviewers via SendMessage(type: "shutdown_request")
2. TeamDelete to clean up resources

Display review summary:
```markdown
## PR Review: Complete

| PR | Title | Decision | Issues Found |
|----|-------|----------|-------------|
| #456 | Add auth middleware | Request Changes | 2 Critical, 3 Important |

### Critical Issues
- [file.go:42] SQL injection risk in query builder
- [auth.go:15] Missing token expiration check

### Important Issues
- [handler.go:88] O(n^2) loop in user lookup
```

AskUserQuestion for next steps:
- Review Next PR
- Done

## PR Phase 5: Merge (--merge flag only)

When `--merge` flag is provided, attempt to merge reviewed PRs:

1. Run Pre-Merge Validation (see Auto-Merge Safety Protocol below)
2. Run Bot Review Integration checks (see below)
3. If all checks pass:
   - Try direct merge: `gh pr merge {number} --squash --delete-branch`
   - If blocked by branch protection: `gh pr merge {number} --squash --delete-branch --auto`
4. Run Post-Auto-Merge Monitoring if `--auto` was used
5. Report merge status

---

## Bot Review Integration

When merging PRs (via `--merge` flag in either issues or pr sub-command), check bot review status before merge.

### Step 1: Check Bot Reviews

```bash
gh pr view {number} --json reviews --jq '
  .reviews[] | select(
    .author.login == "coderabbitai" or
    .author.login == "github-actions" or
    .author.login == "copilot"
  ) | {author: .author.login, state: .state}'
```

### Step 2: Bot Review Decision Matrix

| Bot | Review State | Action |
|-----|-------------|--------|
| CodeRabbit | CHANGES_REQUESTED | Fix feedback, then post `@coderabbitai resolve` |
| CodeRabbit | APPROVED | Proceed with merge |
| CodeRabbit | COMMENTED | Review comments, fix if Critical/Important severity |
| Other bots | CHANGES_REQUESTED | AskUser: fix feedback or dismiss |
| No bot reviews | - | Proceed with merge |

### Step 3: CodeRabbit Feedback Resolution

When CodeRabbit has CHANGES_REQUESTED:

1. Parse review comments from `gh api repos/{owner}/{repo}/pulls/{number}/reviews`
2. Delegate fixes to expert agent (expert-backend or expert-frontend based on file types)
3. Push fixes to PR branch
4. Post resolution comment: `gh pr comment {number} --body "@coderabbitai resolve"`
5. Wait for re-review (poll `gh pr view {number} --json reviews` every 30s, max 5 min)
6. Verify review state is no longer CHANGES_REQUESTED

---

## Auto-Merge Safety Protocol

### Pre-Merge Validation

Before attempting any merge operation:

```bash
# 1. Check PR mergeability
MERGE_STATE=$(gh pr view {number} --json mergeStateStatus --jq '.mergeStateStatus')
# CLEAN → mergeable, BEHIND → needs update, BLOCKED → requirements not met, DIRTY → conflicts

# 2. Check review decision
REVIEW_DECISION=$(gh pr view {number} --json reviewDecision --jq '.reviewDecision')
# APPROVED → ok, CHANGES_REQUESTED → fix first, "" → no required reviews

# 3. Check CI status
gh pr checks {number} --required
```

### Merge State Resolution

| mergeStateStatus | Action |
|-----------------|--------|
| CLEAN | Proceed with merge |
| BEHIND | Run `gh pr update-branch {number}`, wait for CI, retry |
| BLOCKED | Check reviews and CI, resolve blockers |
| DIRTY | Report conflict to user, cannot auto-merge |

### Branch Protection Handling

When direct merge fails with "base branch policy prohibits the merge":
- Use `gh pr merge {number} --squash --delete-branch --auto` for deferred merge
- Monitor auto-merge status with polling (30s intervals, max 10 min)

### Post-Auto-Merge Monitoring

After enabling auto-merge (`--auto`):

```bash
for i in {1..20}; do
  STATE=$(gh pr view {number} --json state --jq '.state')
  if [[ "$STATE" == "MERGED" ]]; then break; fi
  MERGE_STATE=$(gh pr view {number} --json mergeStateStatus --jq '.mergeStateStatus')
  if [[ "$MERGE_STATE" == "BEHIND" ]]; then
    gh pr update-branch {number}
  fi
  sleep 30
done
```

If not merged after timeout: report blocking reasons and AskUserQuestion (Retry / Manual merge / Skip).

### Multi-PR Merge Orchestration

When merging multiple PRs sequentially:

1. Sort PRs: independent (no file overlap) first, dependent later
2. Merge first PR
3. For each remaining PR:
   a. Check `mergeStateStatus` — if BEHIND, run `gh pr update-branch`
   b. Wait for CI to pass on updated branch
   c. Merge PR
4. Report final status for all PRs

---

## Common Rules

- **[HARD] Agent delegation**: All analysis and fixes MUST be delegated to agents
- **[HARD] User approval**: Issue fixes and review submissions require user confirmation
- **Team mode default**: Agent Teams used by default, `--solo` to override
- **Git strategy aware**: Reads `github.git_workflow` from system.yaml
- **Issue linking**: Always include `Fixes #{number}` in commits/PRs
- **Branch per issue**: Each issue gets its own branch (except main_direct)
- **Test verification**: All fixes must pass tests before PR creation
- **Batch safe**: Process multiple items sequentially to avoid branch conflicts
- **Bot work reuse**: Check for existing @claude bot branches before creating new fixes
- **Bot review aware**: Check CodeRabbit/bot review status before merge attempts
- **Auto-merge safe**: Validate mergeability, handle BEHIND state, monitor --auto merges

---

## BEGIN EXECUTION

Parse $ARGUMENTS to determine sub-command (issues or pr), then execute the corresponding workflow phases immediately.
