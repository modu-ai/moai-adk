# GitHub CLI 2.51+ Automation

Comprehensive guide to automating GitHub workflows with gh CLI for AI-powered PR management, bulk operations, issue tracking, and release automation.

---

## Installation & Setup

**Install GitHub CLI**:
```bash
# macOS
brew install gh

# Linux
sudo apt install gh

# Windows
winget install --id GitHub.cli

# Verify installation
gh --version  # Should be ≥2.51.0
```

**Authentication**:
```bash
# Login with interactive auth
gh auth login

# Or use token
gh auth login --with-token < token.txt

# Verify auth status
gh auth status
```

---

## AI-Powered PR Creation

**Purpose**: Leverage GitHub Copilot AI to generate comprehensive PR descriptions automatically

### Basic AI PR Creation

**Auto-generate from commits**:
```bash
# Create PR with AI-generated description
gh pr create \
  --base develop \
  --title "feat(auth): implement JWT authentication" \
  --generate-description

# AI analyzes:
# - All commits in the branch
# - Code changes and diffs
# - Test coverage changes
# - Documentation updates
# - Related issues

# Generates comprehensive description:
# ## Summary
# Implements JWT token-based authentication...
#
# ### Changes
# - JWT validation with signature verification
# - Token expiration handling
# - Comprehensive error handling
#
# ### Testing
# - Added 15 unit tests
# - Integration tests for auth flow
# - Security tests for token manipulation
```

### Advanced AI PR Patterns

**Multi-commit feature PR**:
```bash
# AI consolidates multiple commits into coherent description
git log --oneline develop..HEAD
# abc123 feat(auth): add JWT validation
# def456 test(auth): add JWT tests
# ghi789 docs(auth): document JWT usage
# jkl012 refactor(auth): improve error handling

gh pr create \
  --base develop \
  --title "feat(auth): implement JWT authentication" \
  --generate-description

# AI generates structured description:
# ## Summary
# Complete JWT authentication implementation
#
# ### Implementation Details
# - Token validation with RS256 signatures
# - Expiration checking with clock skew tolerance
# - Comprehensive error handling for edge cases
#
# ### Test Coverage
# - 95% code coverage (up from 87%)
# - All authentication scenarios tested
# - Security vulnerability tests added
#
# ### Documentation
# - API endpoint documentation
# - Security considerations guide
# - Usage examples for frontend integration
```

**Draft PR for early feedback**:
```bash
# Create draft PR with AI description
gh pr create --draft \
  --base develop \
  --title "WIP: Payment gateway integration" \
  --generate-description

# Marks PR as draft (not ready for review)
# Team can see progress and provide early feedback
# Auto-generated description helps communicate current status
```

---

## PR Management Workflows

### PR Status & Review

**View PR details**:
```bash
# List your open PRs
gh pr list --author @me --state open

# View specific PR
gh pr view 123

# Output:
# feat(auth): implement JWT authentication #123
# Open • alice wants to merge 5 commits into develop
#
# Reviewers: bob (approved), charlie (changes requested)
# Assignees: alice
# Labels: enhancement, needs-testing
# Checks: 3/5 passing
```

**Check PR status**:
```bash
# View PR checks
gh pr checks 123

# Output:
# ✓ Tests (pytest) — passed
# ✓ Linting (ruff) — passed
# ✓ Type checking (mypy) — passed
# ✗ Security scan (bandit) — failed
# ⏳ Build (docker) — in progress
```

### PR Review Process

**Request reviewers**:
```bash
# Request specific reviewers
gh pr create \
  --base develop \
  --title "feat(api): add user endpoint" \
  --reviewer alice,bob \
  --reviewer team:security-team

# Add reviewers to existing PR
gh pr edit 123 --add-reviewer charlie
```

**Mark PR ready**:
```bash
# Convert draft to ready for review
gh pr ready 123

# Add labels
gh pr edit 123 --add-label "ready-for-review,high-priority"
```

### PR Merge Strategies

**Squash merge** (recommended for feature branches):
```bash
# Squash all commits into one
gh pr merge 123 --squash \
  --subject "feat(auth): implement JWT authentication" \
  --body "Complete JWT implementation with 95% test coverage" \
  --delete-branch

# Benefits:
# - Clean linear history
# - Single commit per feature
# - Preserves detailed commit history in PR
```

**Merge commit** (preserve commit history):
```bash
# Preserve all commits
gh pr merge 123 --merge --delete-branch

# Benefits:
# - Full commit history preserved
# - Useful for major features with logical commit grouping
```

**Rebase merge** (linear history):
```bash
# Rebase and merge
gh pr merge 123 --rebase --delete-branch

# Benefits:
# - Linear history
# - Preserves individual commits
# - Clean git log
```

**Auto-merge** (when CI passes):
```bash
# Auto-merge when all checks pass
gh pr merge 123 --auto --squash --delete-branch

# Waits for:
# - All required checks to pass
# - Required reviews to be approved
# - Then automatically merges
```

---

## Bulk Operations

### Managing Multiple PRs

**Close stale PRs**:
```bash
# Find PRs not updated in 30 days
gh pr list --state open --json number,updatedAt --jq \
  '.[] | select(.updatedAt < "2025-10-24") | .number' | \
  xargs -I {} gh pr close {}

# Or with comment
gh pr list --state open --json number,updatedAt --jq \
  '.[] | select(.updatedAt < "2025-10-24") | .number' | \
  xargs -I {} sh -c 'gh pr close {} --comment "Closing due to inactivity"'
```

**Merge multiple approved PRs**:
```bash
# Find approved PRs
gh pr list --state open --json number,reviewDecision --jq \
  '.[] | select(.reviewDecision == "APPROVED") | .number' | \
  xargs -I {} gh pr merge {} --squash --delete-branch

# Or merge sequentially
gh pr list --state open --label "approved" --json number --jq '.[].number' | \
  while read pr; do
    gh pr merge "$pr" --squash --delete-branch
    sleep 2  # Avoid rate limiting
  done
```

**Update PR labels in bulk**:
```bash
# Add label to all open PRs by specific author
gh pr list --author alice --state open --json number --jq '.[].number' | \
  xargs -I {} gh pr edit {} --add-label "needs-update"
```

### Batch PR Creation

**Create PRs from multiple branches**:
```bash
#!/bin/bash
# create-prs.sh

branches=(
  "feature/SPEC-001"
  "feature/SPEC-002"
  "feature/SPEC-003"
)

for branch in "${branches[@]}"; do
  git switch "$branch"
  gh pr create \
    --base develop \
    --title "$(git log -1 --pretty=%s)" \
    --generate-description \
    --draft

  sleep 2  # Avoid rate limiting
done
```

---

## Issue Management

### Creating Issues

**Create issue from template**:
```bash
# Create bug report
gh issue create \
  --title "API endpoint returns 500 error" \
  --body "$(cat .github/ISSUE_TEMPLATE/bug_report.md)" \
  --label bug,high-priority \
  --assignee alice

# Create feature request
gh issue create \
  --title "Add OAuth2 support" \
  --body "Implement OAuth2 authentication for third-party integrations" \
  --label enhancement,needs-discussion
```

**Link issues to PRs**:
```bash
# Create PR that closes issue
gh pr create \
  --base develop \
  --title "fix(api): handle null user gracefully" \
  --body "Fixes #342" \
  --generate-description
```

### Issue Queries

**Advanced issue search**:
```bash
# Find open bugs assigned to you
gh issue list --assignee @me --label bug --state open

# Find issues needing triage
gh issue list --label needs-triage --state open

# Find stale issues (no updates in 60 days)
gh issue list --state open --json number,updatedAt --jq \
  '.[] | select(.updatedAt < "2025-09-24")'
```

---

## Release Automation

### Creating Releases

**Create release from tag**:
```bash
# Create and push tag
git tag -a v2.1.0 -m "Release v2.1.0"
git push origin v2.1.0

# Create GitHub release
gh release create v2.1.0 \
  --title "MoAI-ADK v2.1.0" \
  --generate-notes

# Auto-generates release notes from:
# - All merged PRs since last release
# - Grouped by type (features, bug fixes, etc.)
# - Contributors list
```

**Release with artifacts**:
```bash
# Create release with build artifacts
gh release create v2.1.0 \
  --title "MoAI-ADK v2.1.0" \
  --notes "Major performance improvements and bug fixes" \
  dist/moai-adk-2.1.0.tar.gz \
  dist/moai-adk-2.1.0-py3-none-any.whl

# Upload additional files
gh release upload v2.1.0 docs/changelog.md
```

**Pre-release**:
```bash
# Create pre-release (beta/alpha)
gh release create v2.1.0-beta.1 \
  --title "MoAI-ADK v2.1.0 Beta 1" \
  --prerelease \
  --generate-notes
```

### Release Management

**List releases**:
```bash
# View all releases
gh release list

# View specific release
gh release view v2.1.0
```

**Download release assets**:
```bash
# Download all assets for release
gh release download v2.1.0

# Download specific file
gh release download v2.1.0 --pattern "*.whl"
```

---

## GitHub Actions Integration

### Workflow Runs

**View workflow runs**:
```bash
# List recent workflow runs
gh run list --limit 10

# View specific run
gh run view 12345

# Watch run in real-time
gh run watch 12345
```

**Trigger workflows**:
```bash
# Trigger workflow_dispatch
gh workflow run ci.yml \
  --ref develop \
  --field environment=staging

# Re-run failed jobs
gh run rerun 12345 --failed
```

**Download artifacts**:
```bash
# Download artifacts from run
gh run download 12345

# Download specific artifact
gh run download 12345 --name test-results
```

---

## Advanced Automation Patterns

### PR Comment Commands

**Add PR review**:
```bash
# Approve PR
gh pr review 123 --approve

# Request changes
gh pr review 123 --request-changes \
  --body "Please add more tests for edge cases"

# Add comment
gh pr comment 123 \
  --body "LGTM! Excellent work on the error handling."
```

**PR diff operations**:
```bash
# View PR diff
gh pr diff 123

# View specific file diff
gh pr diff 123 --name-only

# Export diff to file
gh pr diff 123 > feature-changes.diff
```

### Notifications & Monitoring

**List notifications**:
```bash
# View unread notifications
gh api notifications

# Mark notifications as read
gh api notifications --method PUT
```

**Watch repository events**:
```bash
# Subscribe to repository
gh repo set-default owner/repo

# View repository activity
gh repo view --json watchers,stargazers
```

---

## CI/CD Integration Examples

### Automated PR Workflow

**GitHub Actions workflow**:
```yaml
# .github/workflows/pr-automation.yml
name: PR Automation

on:
  pull_request:
    types: [opened, synchronize]

jobs:
  auto-label:
    runs-on: ubuntu-latest
    steps:
      - name: Add labels based on files
        run: |
          if gh pr view ${{ github.event.pull_request.number }} --json files \
             --jq '.files[].path' | grep -q "tests/"; then
            gh pr edit ${{ github.event.pull_request.number }} --add-label "has-tests"
          fi

          if gh pr view ${{ github.event.pull_request.number }} --json files \
             --jq '.files[].path' | grep -q "docs/"; then
            gh pr edit ${{ github.event.pull_request.number }} --add-label "has-docs"
          fi
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}

  size-label:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Label by size
        run: |
          additions=$(gh pr view ${{ github.event.pull_request.number }} \
                      --json additions --jq '.additions')

          if [ "$additions" -lt 100 ]; then
            gh pr edit ${{ github.event.pull_request.number }} --add-label "size/small"
          elif [ "$additions" -lt 500 ]; then
            gh pr edit ${{ github.event.pull_request.number }} --add-label "size/medium"
          else
            gh pr edit ${{ github.event.pull_request.number }} --add-label "size/large"
          fi
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

## Performance Metrics

**Efficiency gains with gh CLI**:
```
Manual PR creation (web): ~3-5 minutes
gh CLI + AI description: ~30 seconds (83-90% faster)

Bulk operations (10 PRs):
- Manual: ~30 minutes
- gh CLI batch: ~2 minutes (93% faster)

Release creation:
- Manual: ~10 minutes
- gh CLI auto-notes: ~1 minute (90% faster)
```

---

## Best Practices

### ✅ DO

- Use `--generate-description` for AI-powered PR descriptions
- Set up repo defaults with `gh repo set-default`
- Use `--auto` for merge automation when CI passes
- Batch operations with shell scripts for efficiency
- Use `--draft` for early feedback PRs
- Delete branches after merge with `--delete-branch`

### ❌ DON'T

- Skip `--generate-description` (manual descriptions take 5x longer)
- Force push without `--force-with-lease`
- Merge PRs without review (unless solo project)
- Create PRs without CI passing
- Leave stale PRs open (clean up monthly)

---

**Last Updated**: 2025-11-23
**Lines**: 298
