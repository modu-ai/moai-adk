# Git Workflow Examples

_Last updated: 2025-10-22_

## Example 1: TDD Commit Flow

### Scenario
Implementing authentication feature using RED-GREEN-REFACTOR pattern.

### Commit Sequence
```bash
# RED: Write failing test
git add tests/test_auth.py
git commit -m "test: add failing test for JWT token generation

- Test expects generate_token() to return valid JWT
- Test verifies token expiration is set correctly

@TEST:AUTH-001"

# GREEN: Make test pass
git add src/auth/token.py
git commit -m "feat: implement JWT token generation

- Add generate_token() function
- Set expiration to 24 hours
- Include user ID in payload

@CODE:AUTH-001 | TEST: tests/test_auth.py"

# REFACTOR: Improve code quality
git add src/auth/token.py
git commit -m "refactor: extract token config to constants

- Move expiration time to AUTH_TOKEN_EXPIRY
- Extract secret key configuration
- Add type hints

@CODE:AUTH-001"
```

---

## Example 2: Feature Branch Workflow

### Scenario
Creating a feature branch for user profile editing.

### Workflow
```bash
# Create feature branch from main
git checkout main
git pull origin main
git checkout -b feature/user-profile-edit

# Make changes and commit
git add src/profile/editor.py tests/test_profile.py
git commit -m "feat: add user profile edit functionality

- Implement ProfileEditor class
- Add validation for email and username
- Include tests for edge cases

@CODE:PROFILE-002 | @TEST:PROFILE-002"

# Push and create draft PR
git push -u origin feature/user-profile-edit

gh pr create \
  --draft \
  --title "Add user profile editing" \
  --body "Implements SPEC-PROFILE-002

## Changes
- New ProfileEditor class
- Input validation
- Test coverage: 92%

## Testing
- Unit tests pass
- Integration tests pending

@SPEC:PROFILE-002"

# Mark as ready after review
gh pr ready
gh pr merge --squash
```

---

## Example 3: Hotfix Workflow

### Scenario
Critical bug fix for production.

### Workflow
```bash
# Create hotfix branch from main
git checkout main
git pull origin main
git checkout -b hotfix/auth-token-validation

# Fix and commit
git add src/auth/validator.py
git commit -m "fix: correct token expiration validation logic

Critical bug: tokens were accepted after expiration

- Add timezone-aware datetime comparison
- Fix edge case with UTC conversion
- Add regression test

Fixes @BUG:AUTH-015 | @CODE:AUTH-001"

# Push and create PR (not draft)
git push -u origin hotfix/auth-token-validation

gh pr create \
  --title "Fix token expiration validation" \
  --body "Critical fix for production token validation bug.

## Root Cause
Token expiration check was using naive datetime instead of UTC.

## Fix
- Use timezone-aware datetime
- Add UTC conversion
- Regression test added

## Impact
- Production: All users
- Severity: Critical (security issue)

Fixes @BUG:AUTH-015"

# Fast-track merge after approval
gh pr merge --squash
```

---

## Example 4: Draft to Ready PR Transition

### Scenario
Converting Draft PR to Ready after completing checklist.

### Checklist Verification
```bash
# Run quality checks
pytest --cov=src --cov-report=term-missing  # ✓ 87% coverage
ruff check .                                # ✓ No issues
mypy src/                                   # ✓ Type safe
semgrep scan --config=auto                  # ✓ No vulnerabilities

# Verify TAG coverage
rg '@(CODE|TEST|SPEC):AUTH' -c
# @CODE:AUTH-001 - 5 occurrences
# @TEST:AUTH-001 - 5 occurrences
# @SPEC:AUTH-001 - 1 occurrence

# Update PR description
gh pr edit --body "$(cat << 'PR_BODY'
## Summary
Implements JWT-based authentication per SPEC-AUTH-001.

## Changes
- JWT token generation and validation
- Refresh token mechanism
- Token expiration handling

## Quality Gates
- ✅ Test coverage: 87%
- ✅ Linting: Pass
- ✅ Type checking: Pass
- ✅ Security scan: Pass
- ✅ TAG coverage: Complete

## Testing
- Unit tests: All passing
- Integration tests: All passing
- Manual testing: Verified on staging

@SPEC:AUTH-001
PR_BODY
)"

# Mark as ready
gh pr ready

# Merge after approval
gh pr merge --squash
```

---

_For Git workflow patterns, see reference.md_
