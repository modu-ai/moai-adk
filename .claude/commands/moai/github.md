---
description: "GitHub Workflow - Parallel issue fixing and PR review via worktree isolation"
argument-hint: "issues [--all | --label LABEL | NUMBER | --merge] | pr [--all | NUMBER] | --solo"
type: local
allowed-tools: Read, Write, Edit, Grep, Glob, Bash, AskUserQuestion, Task, TeamCreate, SendMessage, TaskCreate, TaskUpdate, TaskList, TaskGet, TeamDelete
model: sonnet
version: 2.0.0
---

## GitHub Workflow Configuration

- **Repository**: Auto-detected from `gh repo view --json nameWithOwner`
- **Default mode**: Worktree-isolated parallel agents (falls back to `--solo` if unavailable)
- **Branch prefix**: `fix/issue-{n}` for bugs, `feat/issue-{n}` for features
- **Git strategy**: Reads `github.git_workflow` from `.moai/config/sections/system.yaml`
- **Max parallel worktrees**: Auto (up to 3, based on selected issue count)

---

## Architecture Overview

```
issues (parallel worktrees):
  [Worktree A] fix/issue-1: analyze + fix + test + push  ─┐
  [Worktree B] fix/issue-2: analyze + fix + test + push  ─┼─ parallel
  [Worktree C] fix/issue-3: analyze + fix + test + push  ─┘
  MoAI: create PR #1, PR #2, PR #3

pr (parallel review + worktree verification):
  [Worktree V] verifier: checkout PR branch, run tests    ─┐
  [No worktree] security-reviewer: analyze diff           ─┼─ parallel
  [No worktree] perf-reviewer: analyze diff               ─┤
  [No worktree] quality-reviewer: analyze diff            ─┘
  MoAI: synthesize + user approval + submit review
```

---

## EXECUTION DIRECTIVE - START IMMEDIATELY

This is the GitHub workflow command. Parse $ARGUMENTS and execute immediately.

### Argument Parsing

First word determines sub-command:

- **issues** (aliases: issue, fix-issues): Issue fixing workflow
- **pr** (aliases: review, pull-request): PR code review workflow
- No sub-command: Use AskUserQuestion to let user choose

Remaining arguments:

- `--all`: Process all open items
- `--label LABEL`: Filter by label
- `--solo`: Force sequential sub-agent mode (disables worktrees and teams)
- `--merge`: Auto-merge PRs after CI passes (issues only)
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

## Mode Selection

```
if --solo flag OR AGENT_TEAMS not available:
  -> Sequential sub-agent mode (no worktrees)
else:
  -> Worktree-isolated parallel mode (default)
     Max parallel: min(3, selected_count)
     Batch if selected_count > 3
```

---

# SUB-COMMAND: issues

Purpose: Fetch GitHub issues, implement fixes in isolated worktrees, push branches, and create PRs.

## Issues Phase 1: Issue Discovery

### Step 1.1: Fetch Open Issues

`gh issue list --state open --limit 50 --json number,title,labels,assignees,body,createdAt`

### Step 1.2: Issue Selection

If NUMBER argument provided:
- Fetch specific issue: `gh issue view {number} --json number,title,labels,body,comments`
- Proceed directly to Phase 2

If --all or no argument:
- Display issue list as formatted table
- AskUserQuestion: Select which issue(s) to fix (options: individual numbers, batches)

If --label LABEL:
- Filter: `gh issue list --state open --label "{LABEL}" --json number,title,labels,body`
- Display filtered list and let user select

### Step 1.3: Issue Classification

For each selected issue:
- **bug** → `fix/issue-{n}` branch
- **feature** → `feat/issue-{n}` branch
- **enhancement** → `improve/issue-{n}` branch
- **docs** → `docs/issue-{n}` branch

Classification based on: labels, title keywords, body content.

Determine agent type from affected domain:
- Frontend files (*.tsx, *.vue, *.css, *.html) → `team-frontend-dev`
- Default → `team-backend-dev`

### Step 1.4: Pre-flight Checks

For each selected issue, verify no branch conflict:
```bash
git ls-remote --heads origin {prefix}/issue-{number}
```

If remote branch exists:
- Warn user, offer: skip issue, force-push, or create alternate branch name.

---

## Issues Phase 2: Complexity Assessment and Optional Deep Analysis

### Step 2.1: Auto-Classify Issue Complexity

For each selected issue, score complexity signals automatically:

| Signal | Detection | Weight |
|--------|-----------|--------|
| Label | `complexity:high`, `needs-investigation`, `hard`, `research-needed` | +2 |
| Body length | > 800 characters | +1 |
| Comment count | > 5 comments | +1 |
| Title keyword | "investigate", "regression", "intermittent", "flaky", "root cause" | +1 |
| Cross-module | Body mentions 3+ distinct packages or files | +1 |

**Complexity score >= 2 → run deep analysis (Phase 2.2) before implementation**
**Complexity score < 2 → skip directly to Phase 3**

Display complexity assessment to user:
```
Issue #123 (fix login bug): complexity=1 → direct fix
Issue #456 (investigate flaky test): complexity=3 → deep analysis first
```

### Step 2.2: Deep Analysis (Auto-triggered for complex issues)

Create analysis team only for complex issues:
```
TeamCreate(team_name: "github-analysis-{repo-slug}")
```

Spawn one `team-researcher` per complex issue in parallel (max 3):
```
Task(
  subagent_type: "team-researcher",
  team_name: "github-analysis-{repo-slug}",
  name: "analyst-{number}",
  mode: "plan",
  prompt: "Analyze GitHub issue #{number}: {title}.
    Body: {body}
    Comments: {comments}
    Complexity signals detected: {signals}
    Explore the codebase in depth to identify root cause, affected files, and fix approach.
    Do NOT write implementation code.
    Mark task completed via TaskUpdate and send findings via SendMessage with:
    - Root cause analysis
    - Affected file list with line references
    - Recommended fix approach
    - Estimated scope: single-file | multi-file | cross-module"
)
```

After all analysis tasks complete, shutdown team:
```
SendMessage(type: "shutdown_request", recipient: "analyst-{n}", content: "Analysis complete")
TeamDelete
```

Use analysis findings in Phase 3 implementation prompts.

---

## Issues Phase 3: Parallel Implementation (Worktree Isolation)

This is the core phase. Each issue gets its own isolated worktree via `isolation: "worktree"`.

### Batch Strategy

Auto-batch based on selected issue count:
- Up to 3 issues: all parallel in one batch
- 4+ issues: batches of 3, processed sequentially
- Within each batch, all issues execute in parallel

### Team Setup

```
TeamCreate(team_name: "github-issues-{repo-slug}")
```

Create task per issue:
```
TaskCreate: "Fix issue #{number}: {title}" → status: pending
```

### Spawn Fixer Agents (Parallel)

For each issue in the current batch, spawn in a SINGLE message (parallel):

```
Task(
  subagent_type: "team-coder",  // role (backend/frontend) specified in prompt
  team_name: "github-issues-{repo-slug}",
  name: "fixer-{number}",
  isolation: "worktree",
  mode: "acceptEdits",
  prompt: "You are fixing GitHub issue #{number} in an isolated git worktree.

    ISSUE DETAILS:
    Title: {title}
    Body: {body}
    Comments: {comments}
    Type: {classification}
    Target branch name: {prefix}/issue-{number}
    {optional: Analysis findings: {analyst_findings}}

    IMPORTANT: You are working in an ISOLATED WORKTREE. Your changes are completely
    separate from other agents fixing other issues simultaneously. No branch conflicts
    possible.

    EXECUTION STEPS:
    1. Create feature branch (you are on a temp branch in the worktree):
       git checkout -b {prefix}/issue-{number}

    2. Analyze the issue (read affected files, understand the problem):
       - Read files mentioned in the issue
       - Trace the root cause through the code
       - Plan the minimal fix

    3. Write a failing test first (TDD - bug regression test):
       - Write a test that reproduces the bug
       - Verify: go test -run TestName ./... must FAIL

    4. Implement the minimal fix:
       - Make the smallest change that fixes the issue
       - Run: go test -run TestName ./... must PASS

    5. Quality verification:
       - go test -race ./...
       - golangci-lint run
       - go vet ./...
       - If any fail, fix and retry (max 3 attempts)
       - If still failing after 3 attempts, report failure via SendMessage and stop

    6. Stage and commit (specific files only, never git add -A):
       git add {specific_files}
       git commit -m '{type}({scope}): {description}

       Fixes #{number}

       Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>'

    7. Push the branch:
       git push -u origin {prefix}/issue-{number}

    8. Mark task completed and report:
       TaskUpdate: mark task #{taskId} as completed
       SendMessage to team lead with:
       {
         issue_number: {number},
         branch: {prefix}/issue-{number},
         status: success|failed,
         files_modified: [...],
         test_result: pass|fail,
         commit_sha: ...,
         error: ... (if failed)
       }"
)
```

### Handling Idle Notifications

When a fixer agent goes idle:
1. Check TaskList to see if its task is completed
2. If task completed: acknowledge, no action needed
3. If task still pending: SendMessage with clarification or new subtask
4. NEVER ignore idle notifications

### Wait and Collect Results

After spawning all agents in a batch:
- Monitor for SendMessage results from each fixer
- Log successes and failures as they arrive
- After all agents in batch complete, process next batch

---

## Issues Phase 4: PR Creation

After each fixer completes successfully (do NOT wait for entire batch, create PRs as results arrive):

### Step 4.1: Verify Branch on Remote

```bash
git ls-remote --heads origin {prefix}/issue-{number}
```

If not found: retry 3 times (30s intervals), then report error.

### Step 4.2: Create PR

Read `github.git_workflow` from system.yaml:

**github_flow** (default):
```bash
gh pr create \
  --head {prefix}/issue-{number} \
  --base main \
  --title "{type}: {issue_title}" \
  --body "$(cat <<'EOF'
## Summary
{agent_fix_summary}

## Changes
{files_modified}

## Test Results
All tests pass. Coverage maintained.

Fixes #{issue_number}

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

**gitflow**:
- Use `--base develop` instead of `--base main`

**main_direct**:
- Skip PR creation, changes already pushed to main branch

### Step 4.3: Link Issue to PR

```bash
gh issue comment {number} --body "Fix submitted in PR #{pr_number}"
```

### Step 4.4: Worktree Cleanup

After PR is created:
```bash
git worktree prune
```

This cleans up any stale worktree references from agents that have terminated.

---

## Issues Phase 5: Cleanup and Report

### Step 5.1: Shutdown Team

```
SendMessage(type: "shutdown_request", recipient: "fixer-{n}", content: "All issues complete")
// Wait max 30 seconds, then proceed
TeamDelete
```

### Step 5.2: Batch Summary

Display results table:
```markdown
## GitHub Issues: Complete

| Issue | Title | Status | Branch | PR |
|-------|-------|--------|--------|----|
| #123 | Fix login bug | Fixed | fix/issue-123 | #456 |
| #124 | Add dark mode | Fixed | feat/issue-124 | #457 |
| #125 | Memory leak | Failed | - | - |
```

### Step 5.3: Next Steps

AskUserQuestion for next steps:
- Review PRs on GitHub
- Merge All PRs (if --merge flag set)
- Process Failed Issues (retry individually)
- Done

---

## Issues Sub-agent Mode (--solo fallback)

When `--solo` or when AGENT_TEAMS unavailable:

Process issues sequentially, one at a time:
1. `git checkout main && git pull origin main`
2. `git checkout -b {prefix}/issue-{number}`
3. Delegate to appropriate expert agent:
   - Bug fix: expert-debug subagent
   - Feature: expert-backend or expert-frontend subagent
   - Enhancement: expert-refactoring subagent
   - Docs: manager-docs subagent
4. Run tests, commit, push
5. `gh pr create ...`
6. `git checkout main`
7. Proceed to next issue

---

# SUB-COMMAND: pr

Purpose: Fetch PRs, run objective verification in a worktree, perform multi-perspective analysis, and submit review.

## PR Phase 1: PR Discovery

### Step 1.1: Fetch Open PRs

`gh pr list --state open --limit 30 --json number,title,author,labels,additions,deletions,changedFiles,headRefName`

### Step 1.2: PR Selection

If NUMBER argument:
- Fetch: `gh pr view {number} --json number,title,body,files,commits,reviews,headRefName`
- Proceed to Phase 2

If --all or no argument:
- Display PR list as formatted table
- AskUserQuestion: Select PR(s) to review

### Step 1.3: Fetch PR Details

For each selected PR:
```bash
# Full diff for analysis agents
gh pr diff {number}

# Changed file list
gh pr view {number} --json files --jq '.files[].path'

# Head branch name for verifier
gh pr view {number} --json headRefName --jq '.headRefName'

# Existing reviews
gh pr view {number} --json reviews
```

---

## PR Phase 2: Parallel Verification + Analysis

### Team Setup

```
TeamCreate(team_name: "github-pr-review-{number}")
```

Create tasks:
```
TaskCreate: "Verify PR #{number}: run tests and lint in worktree"
TaskCreate: "Security review of PR #{number} diff"
TaskCreate: "Performance review of PR #{number} diff"
TaskCreate: "Quality review of PR #{number} diff"
```

### Spawn All Agents in ONE Message (Parallel)

**Verifier** (1 agent WITH `isolation: "worktree"` — needs to checkout PR branch):

```
Task(
  subagent_type: "team-tester",
  team_name: "github-pr-review-{number}",
  name: "verifier",
  isolation: "worktree",
  mode: "plan",
  prompt: "You are objectively verifying PR #{number} in an isolated worktree.
    PR head branch: {head_ref_name}

    EXECUTION STEPS:
    1. Fetch and checkout the PR branch:
       git fetch origin {head_ref_name}
       git checkout {head_ref_name}

    2. Run full test suite:
       go test -race -count=1 ./...
       Record: total tests, passed, failed, coverage %

    3. Run linter:
       golangci-lint run
       Record: error count, warning count, specific issues

    4. Run vet:
       go vet ./...

    5. Verify build:
       go build ./...

    IMPORTANT: Do NOT modify any source files. Do NOT commit anything.
    Your role is objective verification only.

    Mark task completed via TaskUpdate.
    Send results via SendMessage with:
    {
      type: 'verification',
      tests: { passed: N, failed: N, coverage: '%' },
      lint: { errors: N, warnings: N, issues: [...] },
      build: { success: bool, error: '...' },
      verdict: 'pass' | 'fail'
    }"
)
```

**Analysis Agents** (3 agents, NO worktree — diff analysis only):

```
Task(
  subagent_type: "team-reader",
  team_name: "github-pr-review-{number}",
  name: "security-reviewer",
  mode: "plan",
  prompt: "Security perspective review for PR #{number}.
    Focus: SQL/XSS/command injection, auth/authz gaps, sensitive data exposure, OWASP Top 10.

    Changed files: {file_list}
    Diff: {diff_content}

    For each finding, provide:
    - Severity: Critical | Important | Suggestion
    - File path and line number
    - Description of the issue
    - Recommended fix

    Mark task completed via TaskUpdate.
    Send findings via SendMessage with severity-classified list."
)

Task(
  subagent_type: "team-reader",
  team_name: "github-pr-review-{number}",
  name: "perf-reviewer",
  mode: "plan",
  prompt: "Performance perspective review for PR #{number}.
    Focus: O(n^2) loops, N+1 queries, memory leaks, missing indexes, race conditions, goroutine leaks.

    Changed files: {file_list}
    Diff: {diff_content}

    For each finding:
    - Severity: Critical | Important | Suggestion
    - File path and line number
    - Description and impact
    - Recommended fix

    Mark task completed via TaskUpdate.
    Send findings via SendMessage."
)

Task(
  subagent_type: "team-reader",
  team_name: "github-pr-review-{number}",
  name: "quality-reviewer",
  mode: "plan",
  prompt: "Code quality perspective review for PR #{number}.
    Focus: correctness, edge cases, test coverage for changes, naming conventions,
    error handling completeness, API contract consistency.

    Changed files: {file_list}
    Diff: {diff_content}

    For each finding:
    - Severity: Critical | Important | Suggestion
    - File path and line number
    - Description
    - Recommended improvement

    Mark task completed via TaskUpdate.
    Send findings via SendMessage."
)
```

### Handling Idle Notifications

When any agent goes idle:
1. Check TaskList for task status
2. If completed: send shutdown_request
3. If pending: investigate and send clarification
4. NEVER ignore idle notifications

---

## PR Phase 3: Synthesize and Submit Review

### Step 3.1: Collect All Results

Wait for all 4 agents (verifier + 3 reviewers) to complete.

### Step 3.2: Classify and Prioritize

Aggregate all findings:

```
Critical: (test failures OR security vulns with severity=Critical)
  - Automatically recommend "Request Changes"
  - Must be resolved before merge

Important: (performance issues, missing error handling, no test for changed code)
  - Recommend "Request Changes"
  - Should be resolved

Suggestion: (naming, style, minor improvements)
  - Recommend "Approve with comments" or "Comment Only"
  - Optional improvement
```

### Step 3.3: Format Review Body

```markdown
## Code Review: PR #{number} - {title}

### Objective Verification
- Tests: {passed}/{total} passing | Coverage: {coverage}%
- Lint: {errors} errors, {warnings} warnings
- Build: ✅ Success / ❌ Failed

### Security Analysis
{security_findings_by_severity}

### Performance Analysis
{perf_findings_by_severity}

### Code Quality Analysis
{quality_findings_by_severity}

### Summary
- Critical issues: {n}
- Important issues: {n}
- Suggestions: {n}
- Recommendation: Approve / Request Changes

---
Reviewed by MoAI agent team (security, performance, quality perspectives)
```

### Step 3.4: User Approval

AskUserQuestion (4 options max):
- Approve — submit approval with inline comments
- Request Changes — submit with required changes listed
- Comment Only — submit findings as comment without approval decision
- Skip — do not submit review

### Step 3.5: Submit Review

```bash
# Approve:
gh pr review {number} --approve --body "$(review_body)"

# Request changes:
gh pr review {number} --request-changes --body "$(review_body)"

# Comment only:
gh pr review {number} --comment --body "$(review_body)"
```

For file-level inline comments (Critical and Important issues):
```bash
gh api repos/{owner}/{repo}/pulls/{number}/reviews \
  --method POST \
  --field body="Review summary" \
  --field event="COMMENT" \
  --field "comments=[{\"path\":\"{file}\",\"line\":{line},\"body\":\"{finding}\"}]"
```

---

## PR Phase 4: Cleanup and Report

### Step 4.1: Shutdown Team

```
SendMessage(type: "shutdown_request", recipient: "verifier", content: "Review complete")
SendMessage(type: "shutdown_request", recipient: "security-reviewer", content: "Review complete")
SendMessage(type: "shutdown_request", recipient: "perf-reviewer", content: "Review complete")
SendMessage(type: "shutdown_request", recipient: "quality-reviewer", content: "Review complete")
// Wait max 30 seconds
TeamDelete
```

Verifier worktree auto-cleans (no commits were made).

### Step 4.2: Review Summary

```markdown
## PR Review: Complete

| PR | Title | Decision | Tests | Issues |
|----|-------|----------|-------|--------|
| #456 | Add auth middleware | Request Changes | ✅ 98% | 2 Critical, 3 Important |

### Critical Issues Requiring Fix
- auth.go:45: Missing token expiration check (security)
- handler.go:78: No error handling on db.Query (quality)

### Important Issues
- service.go:123: O(n²) loop in user lookup (performance)
```

AskUserQuestion:
- Review Next PR
- Done

---

## PR Sub-agent Mode (--solo fallback)

When `--solo` or AGENT_TEAMS unavailable:

1. Fetch PR diff
2. expert-security subagent: Security analysis
3. expert-performance subagent: Performance analysis
4. manager-quality subagent: Code quality review
5. Synthesize and present findings
6. AskUserQuestion for review action

No worktree verification in solo mode (no team-tester available).

---

## Common Rules

- **[HARD] Agent delegation**: All analysis and fixes MUST be delegated to agents
- **[HARD] User approval required**: Issue fixes (PR review submission) require user confirmation
- **[HARD] worktree isolation**: All implementation agents MUST use `isolation: "worktree"`
- **[HARD] Specific staging**: Agents must stage specific files (`git add <file>`), never `git add -A`
- **[HARD] No direct commits to main**: Always create feature branch in worktree
- **Git strategy aware**: Reads `github.git_workflow` from system.yaml
- **Issue linking**: Always include `Fixes #{number}` in commits/PRs
- **Test verification**: All fixes must pass `go test -race ./...` before push
- **Batch safe**: Max 3 parallel worktrees (configurable); batch larger sets
- **Idle response**: ALWAYS respond to TeammateIdle events

---

## BEGIN EXECUTION

Parse $ARGUMENTS to determine sub-command (issues or pr), then execute the corresponding workflow phases immediately.
