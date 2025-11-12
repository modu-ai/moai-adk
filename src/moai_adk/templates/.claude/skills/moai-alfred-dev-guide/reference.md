# Alfred Development Guide - Reference

## SPEC-First TDD Workflow Details

### Phase 1: SPEC (`/alfred:1-plan`)

**Duration**: 30-60 minutes per feature
**Output**: `.moai/specs/SPEC-{ID}/spec.md`

#### EARS Requirement Patterns

```markdown
### Ubiquitous Requirements (Baseline)
- The system shall provide user authentication via email.
- The system shall validate email format before storage.

### Event-driven Requirements
- WHEN a user clicks 'Sign Up', the system shall display the signup form.
- WHEN signup is complete, the system shall send verification email.
- WHEN verification email expires, the system shall invalidate token.

### State-driven Requirements
- WHILE the user is unauthenticated, the system shall deny access to protected resources.
- WHILE the session is active, the system shall allow request processing.

### Optional Features
- WHERE the user enables 2FA, the system may require additional verification.
- WHERE API quota remains, the system may allow batch operations.

### Constraints
- IF password is invalid 3 times, the system shall lock the account.
- IF rate limit exceeded, the system shall return 429 error.
```

#### SPEC Metadata Required

```yaml
id: AUTH-001
title: "User Authentication System"
version: 0.1.0
status: active
created: 2025-11-03
updated: 2025-11-03
priority: high
```

### Phase 2: TDD Implementation (`/alfred:2-run`)

**Duration**: 2-4 hours per SPEC

#### RED Phase (Failing Tests)

```python
# tests/test_auth.py
import pytest

    """SPEC: IF password invalid 3 times, lock account"""
    user = create_test_user()

    with pytest.raises(AuthenticationError):  # RED: Will fail
        for _ in range(3):
            authenticate(user.email, "wrong_password")

    assert user.account_locked  # Then passes (GREEN)
```

#### GREEN Phase (Minimal Implementation)

```python
# src/auth.py

    """Authenticate user with email and password."""
    user = User.find_by_email(email)

    if not user.verify_password(password):
        user.failed_attempts += 1
        if user.failed_attempts >= 3:
            user.account_locked = True
        return None

    user.failed_attempts = 0
    return user
```

#### REFACTOR Phase (Improve Code)

```python
# Refactored: Better error handling and type hints
from src.models import User
from src.exceptions import AuthenticationError


    """Authenticate user and manage attempt tracking."""
    user = User.find_by_email(email)
    if not user:
        raise AuthenticationError("User not found")

    if not user.verify_password(password):
        user.increment_failed_attempts(MAX_FAILED_ATTEMPTS)
        raise AuthenticationError("Invalid credentials")

    user.reset_failed_attempts()
    return user
```

### Phase 3: Documentation Sync (`/alfred:3-sync`)

**Duration**: 30-60 minutes per SPEC
**Actions**:
2. Update README, CHANGELOG
3. Generate sync report
4. Create PR to develop

## Context Engineering: JIT Loading

### When to Load What

**`/alfred:1-plan` command**:
```
Always load:
- .moai/project/product.md (project overview)

Optional load:
- .moai/project/structure.md (codebase layout)
- .moai/project/tech.md (technology choices)

Never load:
- Individual SPEC files (load only candidates)
```

**`/alfred:2-run` command**:
```
Always load:
- .moai/specs/SPEC-{ID}/spec.md (target SPEC)

Optionally load:
- development-guide.md (if testing TDD patterns)
- SPEC-related files (if chained features)

Never load:
- Unrelated SPECs, docs, analysis files
```

**`/alfred:3-sync` command**:
```
Always load:
- .moai/reports/sync-report.md (previous sync state)

Optional load:
- Changed SPEC files (only ones modified)
- TAG validation results

Never load:
- Old sync reports (keep recent only)
```

## TRUST 5 Principles Checklist

### T – Test-Driven (SPEC-Aligned)
- [ ] SPEC written with EARS format
- [ ] RED phase: failing tests written first
- [ ] GREEN phase: code passes all tests
- [ ] REFACTOR: improved code still passes

### R – Readable
- [ ] Variable names are descriptive
- [ ] Functions do one thing
- [ ] Comments explain WHY, not WHAT
- [ ] Code follows language style guide
- [ ] Documentation is current

### U – Unified
- [ ] Consistent naming conventions
- [ ] Same patterns across codebase
- [ ] Shared language (EARS for specs)
- [ ] Common test structure
- [ ] Standardized documentation

### S – Secured
- [ ] No hardcoded secrets
- [ ] Input validation present
- [ ] Error handling included
- [ ] OWASP top 10 considered
- [ ] Security review completed

### E – Evaluated (85%+ Coverage)
- [ ] Test coverage ≥85%
- [ ] All functions tested
- [ ] Error paths covered
- [ ] Edge cases included
- [ ] Integration tests exist


```bash
# Check all TAGs
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/


# Find orphan CODE tags (SPEC missing)
```

## Common Commands

```bash
# Start new SPEC
mkdir -p .moai/specs/SPEC-AUTH-001

# Check for duplicates

# List all SPECs by status
rg "^status:" .moai/specs/SPEC-*/spec.md

# Validate all YAML frontmatter
for f in .moai/specs/SPEC-*/spec.md; do
  echo "Checking $f..."
  head -20 "$f" | grep -E "^(id|version|status|created|updated|author|priority):"
done

# Generate SPEC index
echo "# SPEC Inventory" && \
rg "^id:" .moai/specs/SPEC-*/spec.md | sed 's/:id:/: /'
```
