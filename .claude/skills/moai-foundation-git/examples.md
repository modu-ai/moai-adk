# Git 2.47-2.50 & GitHub CLI 2.51+ Examples

## Complete TDD Workflow Examples

### Example 1: Basic Feature Development

**Scenario**: Implement user authentication system (SPEC-001)

```bash
# Step 1: Create feature branch
git switch -c feature/SPEC-001

# Step 2: RED phase - Write failing test
cat > tests/test_auth.py << 'EOF'
def test_user_login_with_valid_credentials():
    auth_service = AuthService()
    result = auth_service.login("alice", "password123")
    assert result.success is True
    assert result.token is not None
EOF

pytest tests/test_auth.py  # Fails (no implementation)
git add tests/test_auth.py
git commit -m "test(auth): add failing test for user login"

# Step 3: GREEN phase - Minimal implementation
cat > src/auth.py << 'EOF'
class AuthService:
    def login(self, username: str, password: str):
        # Minimal implementation to pass test
        if username and password:
            return LoginResult(success=True, token="fake-token")
        return LoginResult(success=False, token=None)
EOF

pytest tests/test_auth.py  # Passes
git add src/auth.py
git commit -m "feat(auth): implement basic user login"

# Step 4: REFACTOR phase - Improve code quality
cat > src/auth.py << 'EOF'
import jwt
from datetime import datetime, timedelta

class AuthService:
    def login(self, username: str, password: str) -> LoginResult:
        """Authenticate user and generate JWT token."""
        if not username or not password:
            raise ValueError("Username and password required")
        
        # Verify credentials (stub)
        if self._verify_credentials(username, password):
            token = self._generate_token(username)
            return LoginResult(success=True, token=token)
        
        return LoginResult(success=False, token=None)
    
    def _verify_credentials(self, username: str, password: str) -> bool:
        # TODO: Implement actual verification
        return True
    
    def _generate_token(self, username: str) -> str:
        payload = {
            "username": username,
            "exp": datetime.utcnow() + timedelta(hours=1)
        }
        return jwt.encode(payload, "secret-key", algorithm="HS256")
EOF

pytest tests/test_auth.py  # Still passes
git add src/auth.py
git commit -m "refactor(auth): improve login implementation with JWT"

# Step 5: Create PR with AI-generated description
gh pr create \
  --base develop \
  --title "feat(auth): implement user authentication" \
  --generate-description

# Output:
# Created pull request #123
# https://github.com/user/moai-adk/pull/123
```

### Example 2: Multi-SPEC Feature with Worktree

**Scenario**: Work on 3 related SPECs simultaneously

```bash
# Main repository
cd /Users/goos/MoAI/MoAI-ADK

# Create worktrees for parallel work
git worktree add ../moai-adk-auth feature/SPEC-001-auth
git worktree add ../moai-adk-api feature/SPEC-002-api
git worktree add ../moai-adk-db feature/SPEC-003-db

# Work on authentication (SPEC-001)
cd ../moai-adk-auth
/moai:2-run SPEC-001
git commit -m "test(auth): add user login test"
git commit -m "feat(auth): implement user authentication"

# Switch to API work (SPEC-002)
cd ../moai-adk-api
/moai:2-run SPEC-002
git commit -m "test(api): add API endpoint test"
git commit -m "feat(api): add authentication endpoints"

# Switch to database work (SPEC-003)
cd ../moai-adk-db
/moai:2-run SPEC-003
git commit -m "test(db): add user schema test"
git commit -m "feat(db): create user database schema"

# Return to main and check status
cd /Users/goos/MoAI/MoAI-ADK
git worktree list
# /Users/goos/MoAI/MoAI-ADK              (develop)
# /Users/goos/moai-adk-auth              (feature/SPEC-001-auth) [changes: 2 commits]
# /Users/goos/moai-adk-api               (feature/SPEC-002-api) [changes: 2 commits]
# /Users/goos/moai-adk-db                (feature/SPEC-003-db) [changes: 2 commits]

# Create PRs for all
cd ../moai-adk-auth && gh pr create --base develop --title "feat(auth): user authentication"
cd ../moai-adk-api && gh pr create --base develop --title "feat(api): authentication endpoints"
cd ../moai-adk-db && gh pr create --base develop --title "feat(db): user database schema"

# Cleanup after merge
cd /Users/goos/MoAI/MoAI-ADK
git worktree remove ../moai-adk-auth
git worktree remove ../moai-adk-api
git worktree remove ../moai-adk-db
```

### Example 3: Hotfix While Working on Feature

**Scenario**: Critical bug reported while working on feature

```bash
# Currently working on feature (uncommitted changes)
git status
# On branch feature/SPEC-001
# Changes not staged for commit:
#   modified:   src/auth.py
#   modified:   tests/test_auth.py

# Critical bug reported in production
git worktree add ../moai-adk-hotfix hotfix/api-crash
cd ../moai-adk-hotfix

# Fix bug in isolation
git commit -m "fix(api): handle null pointer in user endpoint

Critical bug causing API crashes when user data is null.
Added null check and default value.

Fixes #789"

# Push and create emergency PR
git push origin hotfix/api-crash
gh pr create \
  --base main \
  --title "fix(api): critical null pointer crash" \
  --body "Emergency fix for production issue #789" \
  --label "critical" \
  --reviewer @team/reviewers

# Merge immediately after approval
gh pr merge --squash --delete-branch

# Return to feature work (WIP preserved)
cd /Users/goos/MoAI/MoAI-ADK
git status
# On branch feature/SPEC-001
# Changes not staged for commit:
#   modified:   src/auth.py
#   modified:   tests/test_auth.py
# (Original work still intact)

# Continue feature development
git add src/auth.py tests/test_auth.py
git commit -m "feat(auth): implement user login validation"
```

## Sparse-Checkout Examples

### Example 4: Frontend-Only Development

**Scenario**: Frontend developer doesn't need backend code

```bash
# Clone repository
git clone https://github.com/user/moai-adk.git
cd moai-adk

# Enable sparse-checkout
git sparse-checkout init --cone

# Check out only frontend directories
git sparse-checkout set \
  src/frontend/ \
  tests/frontend/ \
  docs/frontend/ \
  package.json \
  README.md

# Verify
ls -la
# Only shows:
# - src/frontend/
# - tests/frontend/
# - docs/frontend/
# - package.json
# - README.md

# Backend directories not checked out:
# - src/backend/
# - src/database/
# - tests/backend/

# Check disk usage
du -sh .
# 280MB (instead of 2.1GB full checkout)

# Add more directories as needed
git sparse-checkout add src/shared/
```

### Example 5: Documentation-Only Checkout

**Scenario**: Technical writer only needs documentation

```bash
# Clone with sparse-checkout
git clone --filter=blob:none --sparse https://github.com/user/moai-adk.git
cd moai-adk

# Checkout only documentation
git sparse-checkout set \
  docs/ \
  README.md \
  CONTRIBUTING.md \
  LICENSE

# Verify
ls -la
# Only documentation files visible

# Update documentation
vim docs/api/authentication.md
git add docs/api/authentication.md
git commit -m "docs(api): update authentication guide"
git push origin main
```

## GitHub CLI Advanced Examples

### Example 6: Automated PR Workflow

**Scenario**: Create, review, and merge PR in one workflow

```bash
# Create PR with complete metadata
gh pr create \
  --base develop \
  --title "feat(auth): implement OAuth2 provider" \
  --body "Implements SPEC-005: OAuth2 integration

## Changes
- Add OAuth2 provider configuration
- Implement token exchange
- Add provider callback handlers

## Test Coverage
- Unit tests: 92%
- Integration tests: 88%

## Related
- Closes #123
- Related to #456" \
  --label "enhancement" \
  --milestone "v2.0.0" \
  --assignee @me \
  --reviewer alice,bob

# Wait for CI to pass
gh pr checks 123 --watch

# Merge when approved
gh pr merge 123 --squash --delete-branch --auto

# Result:
# ✓ Pull request #123 merged
# ✓ Deleted branch feature/SPEC-005
```

### Example 7: Bulk PR Management

**Scenario**: Close multiple stale PRs

```bash
# List all open PRs older than 60 days
gh pr list --state open --json number,updatedAt,title \
  | jq -r '.[] | select(.updatedAt < "2024-09-01") | "\(.number) - \(.title)"'

# Output:
# 45 - feat(api): add rate limiting
# 67 - fix(db): optimize queries
# 89 - docs(readme): update installation

# Close stale PRs with comment
for pr in 45 67 89; do
  gh pr close $pr --comment "Closing due to inactivity. Please reopen if still relevant."
done

# Verify closure
gh pr list --state closed --json number,title | head -n 3
```

### Example 8: PR Review Automation

**Scenario**: Review multiple PRs in batch

```bash
# List my assigned PRs
gh pr list --assignee @me

# Review first PR
gh pr view 123

# View changes
gh pr diff 123

# Add review comment
gh pr review 123 --approve --body "LGTM! Great implementation."

# Merge approved PR
gh pr merge 123 --squash --delete-branch

# Repeat for next PR
gh pr view 124
gh pr diff 124
gh pr review 124 --request-changes --body "Please add tests for edge cases."
```

## Performance Optimization Examples

### Example 9: Large Repository Optimization

**Scenario**: Optimize moai-adk repository (250K objects)

```bash
# Before optimization
git gc
# Counting objects: 250000, done.
# Compressing objects: 100%
# Duration: 45 seconds

# Enable MIDX
git config gc.writeMultiPackIndex true
git config gc.multiPackIndex true

# Repack with MIDX
git repack -ad --write-midx

# After optimization
git gc
# Counting objects: 250000, done.
# Compressing objects: 100%
# Duration: 28 seconds (38% faster)

# Verify MIDX
git verify-pack -v .git/objects/pack/multi-pack-index
# multi-pack-index contains 45 packs
# Total objects: 250000
```

### Example 10: CI/CD Shallow Clone

**Scenario**: Fast checkout for CI/CD pipeline

```bash
# GitHub Actions workflow
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout (shallow)
        uses: actions/checkout@v4
        with:
          fetch-depth: 50  # Last 50 commits only
      
      - name: Run tests
        run: pytest tests/

# Result:
# Full clone: 450MB, 35 seconds
# Shallow:    120MB, 8 seconds (77% faster)
```

## Conventional Commits Examples

### Example 11: Breaking Change

**Scenario**: API endpoint restructuring (breaking change)

```bash
# Breaking change with footer notation
git commit -m "feat(api)!: restructure authentication endpoints

Moved all /auth/* endpoints to /v2/auth/* for better versioning.
Old endpoints deprecated and will be removed in v3.0.0.

BREAKING CHANGE: All /auth/* endpoints moved to /v2/auth/*.
Update client applications to use new endpoint structure.

Migration guide: docs/migration/v2-auth.md

Closes #456
Related #789"
```

### Example 12: Multi-Scope Feature

**Scenario**: Feature affecting multiple components

```bash
# Multiple scopes
git commit -m "feat(auth,api,db): implement OAuth2 provider integration

Added OAuth2 provider support across entire stack:
- auth: OAuth2 client configuration
- api: Provider callback endpoints
- db: Provider token storage

Test coverage: 92%

Closes #123, #456, #789"
```

### Example 13: Performance Improvement

**Scenario**: Database query optimization

```bash
# Performance improvement with benchmarks
git commit -m "perf(db): optimize user activity queries

Added compound index on (user_id, created_at) to improve
user activity query performance.

Benchmarks:
- Before: 120ms average query time
- After:  18ms average query time
- Improvement: 85% faster

Index size overhead: +2.5MB (acceptable)

Closes #234"
```

---

**Last Updated**: 2025-11-22  
**Maintained by**: MoAI-ADK Team
