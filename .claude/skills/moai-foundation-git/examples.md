# Git & GitHub CLI Working Examples

> Practical workflows for MoAI-ADK GitFlow automation

_Last updated: 2025-10-22_

---

## Example 1: Complete Feature Development (TDD Workflow)

### Scenario
Implementing user authentication with JWT tokens using TDD methodology.

### Step-by-Step Workflow

```bash
# 1. Create feature branch from develop
git checkout develop
git pull origin develop
git checkout -b feature/jwt-authentication

# 2. RED Phase - Write failing test
cat > tests/auth/test_jwt.py << 'TEST'
# @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md

def test_jwt_token_validation():
    """When the user submits valid token, system shall validate it."""
    token = generate_test_token()
    result = validate_jwt_token(token)
    assert result.is_valid == True
    assert result.user_id == "test-user"
TEST

git add tests/auth/test_jwt.py
git commit -m "test: add failing test for JWT validation

RED phase: Test expects JWT validation with HS256.

Refs: @TAG:AUTH-001"

# 3. Run tests (should fail)
pytest tests/auth/test_jwt.py  # ❌ Fails as expected

# 4. GREEN Phase - Implement feature
cat > src/auth/jwt.py << 'CODE'
# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/test_jwt.py

import jwt

def validate_jwt_token(token: str):
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=["HS256"])
        return ValidationResult(is_valid=True, user_id=payload["user_id"])
    except jwt.InvalidTokenError:
        return ValidationResult(is_valid=False, user_id=None)
CODE

git add src/auth/jwt.py
git commit -m "feat: implement JWT token validation

GREEN phase: Validates HS256 tokens with expiration.
Passes all authentication tests.

Refs: @TAG:AUTH-001"

# 5. Run tests (should pass)
pytest tests/auth/test_jwt.py  # ✅ Passes

# 6. REFACTOR Phase - Improve code
git add src/auth/jwt.py
git commit -m "refactor: extract JWT config to constants

REFACTOR phase: Improves maintainability.
Extracts algorithm and expiration to constants.

Refs: @TAG:AUTH-001"

# 7. Push and create Draft PR
git push -u origin feature/jwt-authentication

gh pr create --draft --title "feat: Add JWT authentication" --body "$(cat <<EOF
## Summary
- Implements JWT token validation with HS256 algorithm
- Adds comprehensive test coverage for auth flows
- Follows TDD RED → GREEN → REFACTOR pattern

## Changes
- \`src/auth/jwt.py\`: JWT validation logic
- \`tests/auth/test_jwt.py\`: Comprehensive test suite

## Test Plan
- [x] Unit tests pass (pytest)
- [x] Test coverage ≥85%
- [ ] Manual testing in staging
- [ ] Security review

## TRUST Checklist
- [T] Test coverage: 92%
- [R] Linted with ruff
- [U] Type hints added
- [S] No secrets in code
- [T] TAG references added

Refs: @TAG:AUTH-001
EOF
)"

# 8. After review, mark as Ready
gh pr ready

# 9. Merge after approval
gh pr merge --squash
```

---

## Example 2: Hotfix for Production Bug

### Scenario
Critical security vulnerability discovered in production.

```bash
# 1. Create hotfix branch from main
git checkout main
git pull origin main
git checkout -b hotfix/security-patch-v1.2.3

# 2. Fix the issue
git add src/auth/rate_limit.py
git commit -m "fix: patch rate limit bypass vulnerability

SECURITY: Fixes CVE-2025-XXXX rate limit bypass.
Ensures rate limits apply per IP, not per session.

Impact: Critical
Affected versions: 1.2.0 - 1.2.2

Refs: @TAG:SEC-005"

# 3. Update tests
git add tests/security/test_rate_limit.py
git commit -m "test: add regression test for rate limit bypass

Ensures vulnerability cannot be reintroduced.

Refs: @TAG:SEC-005"

# 4. Push and create urgent PR
git push -u origin hotfix/security-patch-v1.2.3

gh pr create --title "HOTFIX: Rate limit bypass vulnerability" \
  --label "security,urgent" \
  --body "$(cat <<EOF
## 🔴 Security Hotfix

**Vulnerability**: Rate limit bypass allows unlimited requests

**Impact**: Critical (CVSS 9.1)

**Fix**: Apply rate limits per IP address, not session ID

**Testing**:
- [x] Regression test added
- [x] Manual security testing passed
- [x] No performance impact

**Deployment**: Requires immediate production deployment

Refs: @TAG:SEC-005
EOF
)"

# 5. Fast-track merge (after security review)
gh pr merge --merge  # Use merge commit for hotfixes

# 6. Tag release
git checkout main
git pull origin main
git tag -a v1.2.3 -m "Security patch: Rate limit bypass fix"
git push origin v1.2.3

# 7. Backport to develop
git checkout develop
git merge main
git push origin develop
```

---

## Example 3: Interactive PR Review Workflow

### Scenario
Reviewing a colleague's pull request with feedback.

```bash
# 1. List open PRs
gh pr list --author teammate

# 2. Check out PR locally
gh pr checkout 123

# 3. Review changes
gh pr diff 123
gh pr view 123

# 4. Run tests locally
pytest
npm test

# 5. Request changes
gh pr review 123 --request-changes --body "$(cat <<EOF
Great start! A few suggestions:

## Code Quality
- [ ] Add type hints to \`validate_user()\` function
- [ ] Extract magic number \`5\` to constant \`MAX_LOGIN_ATTEMPTS\`

## Testing
- [ ] Add test case for rate limit edge case
- [ ] Increase coverage for error handling paths

## Documentation
- [ ] Update CHANGELOG.md with user-facing changes
- [ ] Add docstring to public API methods

Let me know when these are addressed!
EOF
)"

# 6. After updates, approve PR
gh pr review 123 --approve --body "✅ LGTM! All feedback addressed. Thanks!"

# 7. Enable auto-merge
gh pr merge 123 --auto --squash
```

---

## Example 4: Bulk Operations with gh CLI

### Scenario
Managing multiple PRs and issues across projects.

```bash
# Close stale PRs
for pr in $(gh pr list --state open --json number --jq '.[].number'); do
  gh pr view $pr --json updatedAt --jq '.updatedAt' | \
    xargs -I{} date -d {} +%s | \
    awk -v now=$(date +%s) '{if (now - $1 > 2592000) print "stale"}' | \
    grep -q stale && gh pr close $pr --comment "Closing due to inactivity"
done

# Auto-label PRs by file changes
gh pr list --state open --json number --jq '.[].number' | while read pr; do
  if gh pr diff $pr | grep -q "tests/"; then
    gh pr edit $pr --add-label "tests"
  fi
  if gh pr diff $pr | grep -q "docs/"; then
    gh pr edit $pr --add-label "documentation"
  fi
done

# Generate weekly PR report
gh pr list --state merged --search "merged:>=$(date -d '7 days ago' +%Y-%m-%d)" \
  --json number,title,author,mergedAt \
  --jq '.[] | "- #\(.number): \(.title) by @\(.author.login)"'
```

---

## Example 5: Release Management

### Scenario
Preparing a new release with automated changelog.

```bash
# 1. Create release branch
git checkout develop
git pull origin develop
git checkout -b release/v1.3.0

# 2. Update version and changelog
echo "1.3.0" > VERSION
git add VERSION

# Generate changelog from commit messages
git log v1.2.0..HEAD --pretty=format:"- %s (%h)" --no-merges > CHANGELOG_NEW.md

# 3. Commit release prep
git commit -m "chore: prepare v1.3.0 release

- Update VERSION to 1.3.0
- Generate changelog from commits since v1.2.0"

# 4. Push and create release PR
git push -u origin release/v1.3.0

gh pr create --base main --title "Release v1.3.0" --body "$(cat <<EOF
## Release v1.3.0

**Release Date**: $(date +%Y-%m-%d)

### Features
$(git log v1.2.0..HEAD --grep="feat:" --pretty=format:"- %s (%h)" --no-merges)

### Fixes
$(git log v1.2.0..HEAD --grep="fix:" --pretty=format:"- %s (%h)" --no-merges)

### Breaking Changes
$(git log v1.2.0..HEAD --grep="BREAKING" --pretty=format:"- %s (%h)" --no-merges)

## Checklist
- [x] Version bumped
- [x] CHANGELOG updated
- [ ] Documentation updated
- [ ] Release notes drafted
EOF
)"

# 5. After approval, merge to main
gh pr merge --merge

# 6. Create GitHub release
gh release create v1.3.0 --title "v1.3.0" --notes "$(cat CHANGELOG_NEW.md)"

# 7. Merge back to develop
git checkout develop
git merge main
git push origin develop
```

---

## Common Workflows Cheat Sheet

### Daily Development

```bash
# Start work
git checkout develop && git pull origin develop
git checkout -b feature/my-feature

# Make changes with TDD
# (RED → GREEN → REFACTOR commits)

# Push and create PR
git push -u origin feature/my-feature
gh pr create --draft

# Mark ready after review
gh pr ready
```

### Code Review

```bash
# Check out PR
gh pr checkout 123

# Test locally
pytest && npm test

# Approve or request changes
gh pr review 123 --approve
gh pr review 123 --request-changes --body "Feedback..."
```

### Maintenance

```bash
# Sync fork
gh repo sync owner/repo

# Delete merged branches
git branch --merged | grep -v "\*\|main\|develop" | xargs -n 1 git branch -d

# Update local branches
git fetch --all --prune
```

---

**For complete command reference, see [reference.md](reference.md)**

---

**Last Updated**: 2025-10-22
**Examples**: 5 complete workflows + cheat sheet
**Maintained by**: MoAI-ADK Foundation Team
