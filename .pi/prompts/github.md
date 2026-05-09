---
description: "GitHub Workflow - Manage issues and review PRs with Pi teams"
argument-hint: "issues [--all | --label LABEL | NUMBER | --merge | --solo] | pr [--all | NUMBER | --merge | --solo]"
type: local
version: 2.1.0
runtime: pi
---

## Pi GitHub Workflow

Parse `$@` and execute the requested GitHub workflow using Pi tools.

### Runtime tool policy

Use Pi-native tools only:

- Shell and repository inspection: `bash`, `read`
- File changes: `edit`, `write`
- User choices: `ask_user_question`
- Parallel/team work: `teams`
- Solo specialist handoff: `subagent`

Do not use non-Pi tool names. When the workflow implies team task creation, use `teams` tasks/messages. When it implies a one-shot specialist, use `subagent`.

### Pre-flight

1. Detect repository:
   - `gh repo view --json nameWithOwner --jq '.nameWithOwner'`
2. Check working tree and current branch:
   - `git branch --show-current`
   - `git status --porcelain`
3. Load project policy when needed:
   - `.pi/generated/source/moai-config/sections/system.yaml`
   - `.pi/generated/source/moai-config/sections/language.yaml`
   - `.pi/generated/source/moai-config/sections/workflow.yaml`
4. If arguments are missing, ask the user whether to run `issues` or `pr`.

### Argument parsing

First token:

- `issues`, `issue`, `fix-issues`: issue fixing workflow
- `pr`, `review`, `pull-request`: pull request review workflow

Supported options:

- `NUMBER`: target one issue or PR
- `--all`: process all open matching items
- `--label LABEL`: filter issues by label
- `--solo`: use `subagent` instead of `teams`
- `--merge`: merge after validation/CI passes and user policy allows it

## Workflow: issues

Goal: select GitHub issue(s), analyze, implement fixes, validate, and create PRs.

### 1. Discover and select issues

- Specific issue: `gh issue view NUMBER --json number,title,labels,body,comments`
- Label filter: `gh issue list --state open --label LABEL --json number,title,labels,body`
- General list: `gh issue list --state open --limit 50 --json number,title,labels,assignees,body,createdAt`

If no exact issue was provided, present concise choices with `ask_user_question`.

### 2. Classify and plan

Classify each selected issue:

- bug → `fix/issue-NUMBER`
- feature → `feat/issue-NUMBER`
- enhancement → `improve/issue-NUMBER`
- docs → `docs/issue-NUMBER`

Before implementation, check for existing branches/PRs:

```bash
git ls-remote --heads origin | grep -E "(fix|feat|improve|docs|ai|bot)/issue-NUMBER" || true
gh pr list --search "#NUMBER" --state all --json number,title,headRefName,state,author
```

If existing work exists, ask whether to continue, reuse, or skip.

### 3. Execute

Default: use `teams` for independent issue work.

Recommended team task split per issue:

- Analyze issue, code ownership, and reproduction path
- Implement minimal fix on a branch
- Add or update tests
- Validate and prepare PR summary

For `--solo`, use `subagent` with a concrete prompt containing issue URL, branch name, success criteria, and validation commands.

### 4. Validate

Run targeted checks first, then broader checks as appropriate:

```bash
git diff --stat
git status --porcelain
```

Use project-specific tests from the issue context. For Go code prefer:

```bash
go test ./...
go vet ./...
```

If validation fails, diagnose with `subagent` or create a follow-up `teams` task.

### 5. Create PR

Push branch and create PR:

```bash
git push -u origin BRANCH_NAME
gh pr create --title "..." --body "..."
```

PR body must include:

- Issue link
- Summary of changes
- Tests run
- Risks/limitations

If `--merge` is set, verify CI and repository policy first. Ask for confirmation before merging unless the user explicitly requested fully automatic merge.

## Workflow: pr

Goal: select PR(s), review them, comment findings, and optionally merge.

### 1. Discover and select PRs

- Specific PR: `gh pr view NUMBER --json number,title,body,author,headRefName,baseRefName,files,commits,reviews,statusCheckRollup`
- List PRs: `gh pr list --state open --limit 50 --json number,title,author,headRefName,baseRefName,labels,isDraft`

If no exact PR was provided, ask the user to select PR(s).

### 2. Review

Default: use `teams` for parallel review angles:

- Correctness/regression risk
- Security/data handling
- Tests/coverage
- Maintainability/scope control

For `--solo`, use one `subagent` reviewer.

Review evidence directly from:

```bash
gh pr diff NUMBER
gh pr view NUMBER --comments --json comments,reviews,statusCheckRollup
```

### 3. Report or comment

Produce a concise Markdown report:

- Blocking findings
- Non-blocking improvements
- Tests/CI status
- Merge recommendation

Only post GitHub comments after user confirmation unless explicitly requested.

### 4. Merge

If `--merge` is set:

1. Confirm checks are passing.
2. Confirm review conclusion is safe to merge.
3. Ask user confirmation when policy is ambiguous.
4. Merge with repository-preferred strategy, usually:

```bash
gh pr merge NUMBER --squash --delete-branch
```

## Output requirements

Always end with:

- What was selected
- What was changed or reviewed
- Validation commands/results
- PR/comment/merge status
- Remaining risks or user decisions needed
