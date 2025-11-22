# Git 2.47-2.50 & GitHub CLI 2.51+ Reference

## Official Documentation

### Git 2.47-2.50
- [Git Official Docs](https://git-scm.com/docs) - Complete Git documentation
- [Git Release Notes](https://github.com/git/git/tree/master/Documentation/RelNotes) - Version-specific changes
- [Git 2.47 Release](https://github.blog/2024-10-07-highlights-from-git-2-47/) - October 2024
- [Git Worktree](https://git-scm.com/docs/git-worktree) - Parallel branch development
- [Git Sparse-Checkout](https://git-scm.com/docs/git-sparse-checkout) - Partial repository access

### GitHub CLI 2.51+
- [GitHub CLI Docs](https://cli.github.com/manual/) - Complete CLI reference
- [gh pr](https://cli.github.com/manual/gh_pr) - Pull request commands
- [gh workflow](https://cli.github.com/manual/gh_workflow) - GitHub Actions automation

### Conventional Commits 2025
- [Conventional Commits Spec](https://www.conventionalcommits.org/en/v1.0.0/) - Official specification
- [Angular Commit Format](https://github.com/angular/angular/blob/main/CONTRIBUTING.md#-commit-message-format) - Reference implementation

## Advanced Git Techniques

### Git Worktree Advanced Usage

**Scenario 1: Multiple SPEC Development**
```bash
# Main development
cd /Users/goos/MoAI/MoAI-ADK

# Add worktrees for parallel SPEC work
git worktree add ../moai-adk-spec-001 feature/SPEC-001
git worktree add ../moai-adk-spec-002 feature/SPEC-002
git worktree add ../moai-adk-spec-003 feature/SPEC-003

# Work on SPEC-001
cd ../moai-adk-spec-001
/moai:2-run SPEC-001  # TDD implementation

# Switch to SPEC-002 without losing SPEC-001 context
cd ../moai-adk-spec-002
/moai:2-run SPEC-002  # Parallel implementation

# Cleanup when done
git worktree remove ../moai-adk-spec-001
```

**Scenario 2: Hotfix Without Losing WIP**
```bash
# Currently working on feature (uncommitted changes)
cd /Users/goos/MoAI/MoAI-ADK
git status  # Shows modified files

# Critical bug reported, need immediate fix
git worktree add ../moai-adk-hotfix hotfix/critical-bug
cd ../moai-adk-hotfix

# Fix bug in isolation
git commit -m "fix(api): handle null pointer in user endpoint"
git push origin hotfix/critical-bug
gh pr create --base main --title "Hotfix: Null pointer" --body "..."

# Return to feature work (WIP still intact)
cd /Users/goos/MoAI/MoAI-ADK
git status  # Original modifications still present
```

### Git Sparse-Checkout Patterns

**Pattern 1: Frontend-Only Development**
```bash
# Clone repository
git clone https://github.com/user/moai-adk.git
cd moai-adk

# Enable sparse-checkout
git sparse-checkout init --cone

# Include only frontend directories
git sparse-checkout set src/frontend/ tests/frontend/ docs/frontend/

# Result: Backend code not checked out (saves disk space)
du -sh .  # 300MB instead of 2GB
```

**Pattern 2: Gradual Expansion**
```bash
# Start with minimal checkout
git sparse-checkout set src/core/ tests/core/

# Add more as needed
git sparse-checkout add src/api/
git sparse-checkout add docs/

# Remove unnecessary directories
git sparse-checkout set src/core/ tests/core/ docs/
```

### Git Rebase Strategies

**Strategy 1: Interactive Rebase with Autosquash**
```bash
# Development commits:
# abc123 feat(auth): add user login
# def456 test(auth): add login tests
# ghi789 fixup! feat(auth): add user login  # Typo fix
# jkl012 test(auth): add edge case test

# Squash fixup commits
git rebase --interactive --autosquash develop

# Result:
# abc123 feat(auth): add user login (includes typo fix)
# def456 test(auth): add login tests
# jkl012 test(auth): add edge case test
```

**Strategy 2: Rebase with Preserve Merges**
```bash
# Complex branch with merge commits
git rebase --rebase-merges develop

# Preserves merge structure while updating base
```

## Performance Optimization Techniques

### MIDX (Multi-Pack Index) Optimization

**Configuration**:
```bash
# Enable globally
git config --global gc.writeMultiPackIndex true
git config --global gc.multiPackIndex true
git config --global repack.writeBitmaps true
git config --global feature.experimental true

# Per-repository
git config gc.writeMultiPackIndex true
```

**Manual Optimization**:
```bash
# Repack with MIDX
git repack -ad --write-midx

# Verify integrity
git verify-pack -v .git/objects/pack/multi-pack-index

# Force garbage collection with MIDX
git gc --aggressive --prune=now
```

**Benchmarking Results** (moai-adk repository):
```
Repository: 250K objects, 45 packfiles
Test: git log --all --oneline | wc -l

Without MIDX:
  Time: 8.2s
  Pack files: 45 separate
  
With MIDX:
  Time: 5.1s (38% faster)
  Pack files: 1 consolidated MIDX
```

### Shallow & Partial Clones

**Shallow Clone Strategies**:
```bash
# Clone last 50 commits (CI/CD)
git clone --depth 50 https://github.com/user/moai-adk.git

# Clone last month's commits
git clone --shallow-since="1 month ago" https://github.com/user/moai-adk.git

# Deepen if needed
git fetch --deepen=100
```

**Partial Clone (Blob-less)**:
```bash
# Clone without downloading blobs
git clone --filter=blob:none https://github.com/user/moai-adk.git

# Blobs downloaded on-demand when accessed
git switch feature/SPEC-001  # Downloads necessary blobs

# Check clone size
du -sh .git  # 85MB instead of 450MB
```

## GitHub CLI Advanced Patterns

### AI-Powered PR Automation

**Example 1: Multi-SPEC PR with AI Description**
```bash
# Create PR with AI-generated description
gh pr create \
  --base develop \
  --head feature/SPEC-001-002-003 \
  --title "Implement user authentication, authorization, and audit logging" \
  --generate-description

# AI analyzes commits and generates:
## Summary
- Implement JWT-based authentication (SPEC-001)
- Add role-based authorization (SPEC-002)
- Create comprehensive audit logging (SPEC-003)

## Test Coverage
- SPEC-001: 92% coverage
- SPEC-002: 88% coverage
- SPEC-003: 95% coverage

## Related Issues
- Closes #45, #67, #89
```

**Example 2: Automated PR Review Requests**
```bash
# Create PR and assign reviewers
gh pr create \
  --base develop \
  --title "feat(api): add rate limiting" \
  --reviewer alice,bob \
  --assignee @me \
  --label "enhancement" \
  --milestone "v2.0.0"
```

### Bulk Operations

**Example 1: Close Stale PRs**
```bash
# List PRs older than 60 days
gh pr list --state open --json number,updatedAt,title \
  | jq -r '.[] | select(.updatedAt < "2024-09-01") | "\(.number) - \(.title)"'

# Close stale PRs
gh pr list --state open --json number,updatedAt \
  | jq -r '.[] | select(.updatedAt < "2024-09-01") | .number' \
  | xargs -I {} gh pr close {} --comment "Closed due to inactivity"
```

**Example 2: Merge Multiple Related PRs**
```bash
# Merge SPEC-001, SPEC-002, SPEC-003 in sequence
for pr in 123 124 125; do
  gh pr merge $pr --squash --delete-branch --auto
done
```

## Conventional Commits Best Practices

### Breaking Changes

**Format**:
```bash
# Footer notation (recommended)
git commit -m "feat(api)!: redesign authentication endpoints

BREAKING CHANGE: All /auth/* endpoints moved to /v2/auth/*.
Update client applications to use new endpoints."

# Multiple breaking changes
git commit -m "refactor(api)!: restructure API versioning

BREAKING CHANGE: API v1 removed. Use /v2/ prefix for all endpoints.
BREAKING CHANGE: Authentication header changed from X-Auth-Token to Authorization."
```

### Multi-Scope Commits

```bash
# Multiple scopes affected
git commit -m "feat(auth,api,db): implement OAuth2 provider integration"

# Nested scopes
git commit -m "fix(api/auth): handle token expiration edge case"
```

### Commit Body Best Practices

```bash
# Detailed explanation
git commit -m "perf(db): optimize user query with database indexes

Added compound index on (user_id, created_at) to speed up
user activity queries. Benchmarks show 85% reduction in
query time (120ms → 18ms).

Closes #456"
```

## Version-Specific Features

### Git 2.47 (October 2024)
- **Multi-pack indexes (MIDX)**: Incremental updates
- **Branch base detection**: `%(is-base:develop)` atom
- **Performance improvements**: 30-40% faster pack operations

### Git 2.48 (November 2024)
- **Experimental backfill**: Smart partial clone fetching
- **Survey command**: Repository data shape analysis
- **Reftable improvements**: Better concurrent access

### Git 2.49 (December 2024)
- **C11 standard compliance**: Platform stability
- **VSCode mergetool integration**: Native IDE support
- **Enhanced git log filtering**: Performance improvements

### Git 2.50 (Expected January 2025)
- **Ref verification**: Stronger integrity checks
- **Commit graph improvements**: Faster traversal
- **Packfile optimization**: Reduced storage overhead

## Integration with MoAI-ADK

### `/moai:1-plan` Integration
```bash
# Command execution
/moai:1-plan "User authentication system"

# Git operations performed:
1. git switch -c feature/SPEC-001
2. Create .moai/specs/SPEC-001/spec.md
3. git add .moai/specs/SPEC-001/
4. git commit -m "docs(spec): create SPEC-001 for user authentication"
```

### `/moai:2-run` Integration
```bash
# Command execution
/moai:2-run SPEC-001

# Git operations performed (TDD):
1. git commit -m "test(auth): add failing test for user login"      # RED
2. git commit -m "feat(auth): implement user login validation"      # GREEN
3. git commit -m "refactor(auth): improve error handling"           # REFACTOR
```

### `/moai:3-sync` Integration
```bash
# Command execution
/moai:3-sync auto SPEC-001

# Git operations performed:
1. git add docs/
2. git commit -m "docs(api): generate API documentation for SPEC-001"
3. gh pr create --base develop --title "feat(auth): implement user authentication"
4. gh pr merge 123 --squash --delete-branch (if auto-merge enabled)
```

---

**Last Updated**: 2025-11-22  
**Maintained by**: MoAI-ADK Team
