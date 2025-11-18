# Git & GitHub Automation - Enterprise Reference ( .0)

**Last Updated**: 2025-11-12 | Version: 4.0.0 Enterprise | Git 2.47-2.50, GitHub CLI 2.83.0

---

## Git 2.47+ Commands Reference

### Incremental Multi-Pack Index (MIDX) Commands

```bash
# Check if MIDX is enabled
git config gc.writeMultiPackIndex

# Enable MIDX globally
git config --global gc.writeMultiPackIndex true

# Verify MIDX structure
git verify-pack -v .git/objects/pack/multi-pack-index

# Trigger MIDX creation
git gc --aggressive

# Check MIDX performance
git rev-parse --all | time git cat-file --batch-check > /dev/null
```

### Branch Base Detection (Git 2.47+)

```bash
# Find branches based on develop
git for-each-ref \
  --format='%(if)%(is-base:develop)%(then)✓ %(refname:short)%(else)✗ %(refname:short)%(end)' \
  refs/heads/

# Output:
# ✓ feature/SPEC-001
# ✓ feature/SPEC-002
# ✗ hotfix/critical-bug
```

### Experimental Commands (Git 2.48+)

```bash
# Enable experimental features
git config feature.experimental true

# Use backfill for sparse clone
git backfill --lazy

# Run survey on repository
git survey
# Output:
# Repository efficiency: 87%
# Recommendation: Enable sparse checkout for docs/
```

---

## GitHub CLI 2.83.0 Commands

### PR Management

```bash
# Create feature branch PR (draft)
gh pr create \
  --draft \
  --title "WIP: Feature description" \
  --body "Detailed description" \
  --base develop \
  --head feature/SPEC-001

# List PRs with filters
gh pr list --author @me --state open --label "type:feature"

# View PR with JSON output
gh pr view 123 --json title,state,createdAt,reviews

# Add reviewer
gh pr review 123 --request-changes

# Merge PR with squash and delete branch
gh pr merge 123 --squash --delete-branch --auto

# Reopen closed PR
gh pr reopen 123
```

### Issue Management

```bash
# Create issue
gh issue create --title "Bug: Login fails" --body "Steps to reproduce..."

# List issues assigned to user
gh issue list --assignee @me

# Close issue
gh issue close 456

# Link issue to PR
gh pr view 123 --json body  # Check if mentions issue
```

### Release Management

```bash
# Create release (draft)
gh release create v1.0.0 \
  --title "Version 1.0.0" \
  --notes "Release notes" \
  --draft

# Publish release
gh release edit v1.0.0 --draft=false

# Upload assets
gh release upload v1.0.0 ./build/app.tar.gz

# List releases
gh release list --limit 5
```

### Workflow Management

```bash
# List workflows
gh workflow list

# Run workflow
gh workflow run ci.yml --ref develop

# View workflow runs
gh run list --workflow=ci.yml --status completed

# Cancel run
gh run cancel 123456

# View run logs
gh run view 123456 --log
```

---

## TDD Commit Message Examples

### RED Phase Example

```
🔴 RED: test_user_login_fails_with_invalid_password

Added test case to verify that login fails gracefully when
user provides invalid password. Test uses mocked AuthService
to ensure proper error handling.

Test file: tests/auth/test_login.py
Test function: test_login_invalid_password

```

### GREEN Phase Example

```
🟢 GREEN: implement_user_authentication_service

Implemented AuthService class with login() method that:
- Validates email format
- Checks password strength
- Returns authenticated user or error

Implementation follows SOLID principles and includes
comprehensive error handling.

Files: src/services/auth_service.py
Lines: 150 (+)

```

### REFACTOR Phase Example

```
♻️ REFACTOR: improve_authentication_error_messages

Enhanced error messages in AuthService to provide
more helpful feedback to users:
- "Invalid credentials" → "Email or password incorrect"
- Added error codes for API responses
- Improved logging for debugging

Performance: No changes
Coverage: 87% → 89%

```

---

## GitHub Actions Integration

### Example: TDD CI/CD Pipeline

```yaml
# .github/workflows/tdd-quality-gate.yml

name: TDD Quality Gate

on:
  pull_request:
    branches: [develop, main]
  push:
    branches: [develop]

jobs:
  test-coverage:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-python@v4
        with:
          python-version: '3.12'
      
      - name: Install dependencies
        run: pip install -r requirements.txt pytest pytest-cov
      
      - name: Run tests
        run: pytest tests/ --cov=src --cov-report=xml
      
      - name: Verify coverage >= 85%
        run: |
          coverage report --fail-under=85
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.xml
  
  lint-type-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-python@v4
      
      - name: Lint with ruff
        run: pip install ruff && ruff check src/ tests/
      
      - name: Type check with mypy
        run: pip install mypy && mypy src/
  
  security-scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run Bandit
        run: pip install bandit && bandit -r src/
```

---

## Complete Workflow Example

### End-to-End Feature Implementation

```bash
# Step 1: Create SPEC
/alfred:1-plan "Implement user registration"
# Creates feature/SPEC-005
# Asks: Feature Branch + PR or Direct Commit?
# Choose: Feature Branch + PR

# Step 2: Start feature branch
git checkout feature/SPEC-005
git pull origin feature/SPEC-005

# Step 3: RED phase - Write tests
echo 'def test_register_user_with_valid_email():
    user = register_user("user@example.com", "secure123")
    assert user.email == "user@example.com"' > tests/auth/test_register.py

git add tests/auth/test_register.py
git commit -m "🔴 RED: test_user_registration

Tests basic user registration flow.


git push origin feature/SPEC-005

# Step 4: GREEN phase - Implement
echo 'def register_user(email, password):
    validate_email(email)
    validate_password(password)
    user = User(email=email, password=hash_password(password))
    db.save(user)
    return user' > src/services/user_service.py

git add src/services/user_service.py
git commit -m "🟢 GREEN: implement_user_registration

Implemented registration service with validation.


git push origin feature/SPEC-005

# Step 5: REFACTOR phase - Improve
# (Add error handling, logging, etc.)
git commit -m "♻️ REFACTOR: improve_registration_error_handling

Added comprehensive error handling and logging.


git push origin feature/SPEC-005

# Step 6: Create PR
gh pr create \
  --base develop \
  --head feature/SPEC-005 \
  --title "feat: Implement user registration (SPEC-005)" \
  --body "## Summary
  
- Implements user registration with email/password
- Includes validation for both fields
- Comprehensive error handling
- Test coverage: 87%

## Testing
- All tests passing ✓
- Manual testing completed ✓

## Deployment
Ready for staging environment"

# Step 7: Verify quality gates pass
# (GitHub Actions runs automatically)
# - Tests: PASS
# - Lint: PASS
# - Type check: PASS
# - Coverage: 87% (>= 85%) PASS

# Step 8: Get review and merge
gh pr merge feature/SPEC-005 \
  --squash \
  --delete-branch

# Step 9: Sync documentation
/alfred:3-sync auto SPEC-005
```

---

## Git Performance Tuning (Enterprise)

### Large Repository Optimization

```bash
# Enable all performance features
git config --global gc.writeMultiPackIndex true
git config --global gc.multiPackIndex true
git config --global repack.writeBitmaps true
git config --global core.commitGraph true
git config --global core.preloadIndex true

# Run aggressive cleanup
git gc --aggressive --prune=now

# Verify improvement
time git rev-parse --all
```

### Monorepo Optimization

```bash
# Enable sparse checkout for large monorepos
git sparse-checkout init --cone

# Add paths to check out
git sparse-checkout set \
  src/moai_adk \
  tests \
  docs \
  .github

# Result: Only checkout ~20% of files, 70% faster clone
```

---

## Troubleshooting Guide

### Merge Conflict Resolution

```bash
# Start interactive rebase
git rebase -i develop

# Resolve conflicts
git status  # See conflicts
# Edit conflicting files
git add <resolved-file>
git rebase --continue

# If rebase fails, abort
git rebase --abort
```

### Large Commit Recovery

```bash
# If accidentally pushed large file
git log --all --full-history -- <file>

# Remove from history
git filter-branch --tree-filter 'rm -f <file>' --all

# Force push (only if team notified)
git push origin --force --all
```

### Session Recovery

```bash
# Check .moai/sessions/ for saved state
ls -la .moai/sessions/

# View checkpoint
cat .moai/sessions/checkpoints/<checkpoint-id>.json

# Resume from checkpoint
git checkout feature/SPEC-001
git reset --hard <checkpoint-commit-hash>
```

---

**Reference Last Updated**: 2025-11-12 | **Version**: 4.0.0 Enterprise | **Git**: 2.47-2.50, **GitHub CLI**: 2.83.0
