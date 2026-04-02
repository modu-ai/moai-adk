---
name: moai-workflow-github
description: >
  GitHub issue fixing and PR code review workflow using gh CLI.
  Analyzes issues, implements fixes with test verification, creates PRs,
  and performs multi-perspective code reviews.
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-04-02"
  tags: "github, issues, pr, review, gh"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 4000

# MoAI Extension: Triggers
triggers:
  keywords: ["github", "issue", "pr", "pull request"]
  agents: ["expert-debug", "expert-backend", "manager-quality"]
  phases: ["github"]
---

# Workflow: GitHub - Issue Fix and PR Review

Purpose: Fix GitHub issues and review PRs using `gh` CLI directly. No custom Go wrappers - leverages `gh` for all GitHub operations.

Flow: Discovery -> Analysis -> Implementation -> PR Creation -> Report

## Prerequisites

- `gh` CLI installed and authenticated (`gh auth status`)
- Git repository with GitHub remote

## Supported Flags

- --all: Process all open items (issues or PRs)
- --label LABEL: Filter issues by label

## Sub-commands

First argument determines the workflow:
- **issues** (aliases: issue, fix): Fix GitHub issues
- **pr** (aliases: review, pull-request): Review PRs
- No argument: AskUserQuestion to choose

---

## Sub-command: issues

### Phase 1: Issue Discovery

Step 1.1: Fetch open issues
```bash
gh issue list --state open --limit 30 --json number,title,labels,body,assignees
```

Step 1.2: Issue selection
- If NUMBER provided: `gh issue view {number} --json number,title,body,labels,comments`
- If --all: Process all open issues sequentially
- If --label LABEL: `gh issue list --state open --label "{LABEL}" --json number,title,labels,body`
- Otherwise: Display list via AskUserQuestion, let user select

Step 1.3: Classification
Classify by title/labels/body:
- **bug** → branch prefix `fix/issue-{number}`
- **feature** → branch prefix `feat/issue-{number}`
- **enhancement** → branch prefix `improve/issue-{number}`
- **docs** → branch prefix `docs/issue-{number}`

### Phase 2: Analysis and Implementation

[HARD] Delegate all implementation to specialized agents.

Step 2.1: Analyze root cause
- Delegate to expert-debug subagent (bugs) or expert-backend subagent (features)
- Agent reads issue body, explores codebase, identifies affected files and fix approach

Step 2.2: Create branch and implement
```bash
git checkout main && git pull origin main
git checkout -b {prefix}/issue-{number}
```
- Agent implements fix using Edit tool
- Agent writes/updates tests for the fix
- Run test suite to verify (language-specific: `go test ./...`, `npm test`, etc.)
- If tests fail: retry with error context (max 3 attempts)

Step 2.3: Commit
```bash
git add {modified_files}
git commit -m "{type}({scope}): {description}

Fixes #{number}"
```

### Phase 3: PR Creation

```bash
git push -u origin {prefix}/issue-{number}
gh pr create --title "{type}: {title}" --body "## Summary
{fix_summary}

## Test Plan
- {test_descriptions}

Fixes #{number}"
```

After PR creation:
```bash
git checkout main
```

### Phase 4: Report

Display result:
```markdown
## Issue #{number} Fixed

- Branch: {prefix}/issue-{number}
- PR: #{pr_number} ({pr_url})
- Files modified: {count}
- Tests: {pass_count}/{total_count} passing
```

AskUserQuestion for next steps:
- Fix another issue: Continue to next issue
- Done: End workflow

---

## Sub-command: pr

### Phase 1: PR Discovery

Step 1.1: Fetch open PRs
```bash
gh pr list --state open --limit 20 --json number,title,author,additions,deletions,changedFiles,headRefName
```

Step 1.2: PR selection
- If NUMBER provided: Fetch specific PR
- If --all: Review all sequentially
- Otherwise: AskUserQuestion to select

Step 1.3: Fetch details
```bash
gh pr diff {number}
gh pr view {number} --json files --jq '.files[].path'
```

### Phase 2: Code Review

Delegate to two sub-agents in parallel:

Agent 1 - expert-security:
- Injection risks (SQL, XSS, command injection)
- Authentication/authorization issues
- Sensitive data exposure
- OWASP Top 10 compliance

Agent 2 - manager-quality:
- Code correctness and edge cases
- Test coverage for changes
- Error handling completeness
- Naming conventions and readability

Synthesize findings:
- **Critical**: Must fix before merge
- **Important**: Should fix
- **Suggestion**: Nice to have

### Phase 3: Submit Review

Present review summary to user via AskUserQuestion:
- Approve (Recommended if no Critical issues): Submit approval
- Request Changes: Submit with required changes
- Comment Only: Submit observations without decision
- Skip: Do not submit review

Submit via:
```bash
gh pr review {number} --approve --body "{review_body}"
# OR
gh pr review {number} --request-changes --body "{review_body}"
# OR
gh pr review {number} --comment --body "{review_body}"
```

### Phase 4: Report

Display review summary:
```markdown
## PR #{number} Review Complete

- Decision: {APPROVE|REQUEST_CHANGES|COMMENT}
- Critical: {count}
- Important: {count}
- Suggestions: {count}
```

AskUserQuestion for next steps:
- Review another PR
- Done

---

## Common Rules

- [HARD] All GitHub operations use `gh` CLI directly - no custom wrappers
- [HARD] All implementation delegated to specialized agents
- [HARD] User confirmation required before PR creation and review submission
- Branch per issue: each fix gets its own branch
- Test verification: all fixes must pass tests before PR
- Conventional Commits: commit messages follow project convention
- Language detection: read language.yaml for conversation_language in reports

---

Version: 1.0.0
Updated: 2026-04-02
