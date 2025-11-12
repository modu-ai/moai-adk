# Alfred Development Guide - Examples

## Example 1: Complete SPEC-First TDD Workflow for User Authentication

### Phase 1: Write SPEC (30 minutes)

**File**: `.moai/specs/SPEC-AUTH-001/spec.md`

```markdown
---
id: AUTH-001
title: "User Authentication System"
version: 0.1.0
status: active
created: 2025-11-03
updated: 2025-11-03
priority: high
---

# User Authentication SPEC


## Ubiquitous Requirements
- The system shall provide user authentication via email and password.
- The system shall validate email format (RFC 5322) before storage.
- The system shall hash passwords using bcrypt (≥12 rounds).

## Event-driven Requirements
- WHEN a user submits signup form, the system shall create user account.
- WHEN email verification link is clicked, the system shall activate account.
- WHEN login fails 3 times, the system shall lock account for 1 hour.

## State-driven Requirements
- WHILE user is authenticated, the system shall allow access to protected resources.
- WHILE session is active, the system shall maintain JWT token validity.

## Optional Features
- WHERE 2FA is enabled, the system may require additional verification.

## Constraints
- IF password is invalid, the system shall return 401 Unauthorized.
- The system shall NOT store plaintext passwords.
- The system shall enforce 8-character minimum password length.

## HISTORY

### v0.1.0 (2025-11-03)
- Initial draft with core auth requirements
```

### Phase 2: TDD Implementation (2 hours)

**RED: Write failing tests first**

```python
# tests/test_auth.py
import pytest
from src.models import User

class TestSignup:
    """SPEC: User signup and verification workflow"""

        """SPEC: The system shall create user account."""
        response = signup(email="user@example.com", password="securePass123")

        assert response["status"] == "created"
        user = User.find_by_email("user@example.com")
        assert user is not None
        assert user.verified is False

        """SPEC: The system shall validate email format."""
        with pytest.raises(ValueError):
            signup(email="invalid-email", password="securePass123")

        """SPEC: Enforce 8-character minimum password length."""
        with pytest.raises(ValueError):
            signup(email="user@example.com", password="short")

class TestLogin:
    """SPEC: User login and session management"""

        """SPEC: The system shall authenticate valid credentials."""
        user = create_test_user("user@example.com", "securePass123", verified=True)

        token = login(email="user@example.com", password="securePass123")

        assert token is not None
        assert isinstance(token, str)

        """SPEC: WHEN login fails 3 times, lock account."""
        user = create_test_user("user@example.com", "securePass123", verified=True)

        for i in range(3):
            with pytest.raises(Exception):
                login(email="user@example.com", password="wrongPassword")

        assert user.locked is True
        assert user.locked_until is not None
```

**GREEN: Minimal implementation that passes tests**

```python
# src/auth.py
import bcrypt
import jwt
from src.models import User
from src.email import send_verification_email

    """Create new user account with email verification."""
    # Validate inputs
    if "@" not in email:
        raise ValueError("Invalid email")
    if len(password) < 8:
        raise ValueError("Password too short")

    # Hash password
    hashed = bcrypt.hashpw(password.encode(), bcrypt.gensalt(12))

    # Create user
    user = User.create(email=email, password_hash=hashed, verified=False)

    # Send verification email
    token = generate_verification_token(user.id)
    send_verification_email(email, token)

    return {"status": "created", "user_id": user.id}

    """Authenticate user and return JWT token."""
    user = User.find_by_email(email)
    if not user:
        raise ValueError("Invalid credentials")

    if user.locked:
        raise ValueError("Account locked")

    # Verify password
    if not bcrypt.checkpw(password.encode(), user.password_hash):
        user.failed_attempts += 1
        if user.failed_attempts >= 3:
            user.locked = True
            user.locked_until = datetime.now() + timedelta(hours=1)
        user.save()
        raise ValueError("Invalid credentials")

    # Reset attempts and create token
    user.failed_attempts = 0
    token = jwt.encode({"user_id": user.id, "exp": datetime.utcnow() + timedelta(hours=1)}, "SECRET_KEY")
    user.save()

    return token
```

**REFACTOR: Improve code quality**

```python
# src/auth.py (refactored)
import bcrypt
import jwt
from typing import Dict, Optional
from src.models import User
from src.email import send_verification_email
from src.config import Settings


def _validate_email(email: str) -> None:
    """Validate email format."""
    if "@" not in email or "." not in email.split("@")[1]:
        raise ValueError("Invalid email format")

def _validate_password(password: str) -> None:
    """Validate password strength."""
    if len(password) < MIN_PASSWORD_LENGTH:
        raise ValueError(f"Password must be at least {MIN_PASSWORD_LENGTH} characters")

    """Create new user account with email verification.

    Args:
        email: User email address
        password: User password (min 8 chars)

    Returns:
        Dictionary with status and user_id

    Raises:
        ValueError: If email or password invalid
    """
    _validate_email(email)
    _validate_password(password)

    hashed = bcrypt.hashpw(password.encode(), bcrypt.gensalt(12))
    user = User.create(email=email, password_hash=hashed, verified=False)

    token = jwt.encode({"user_id": user.id, "type": "verify"}, Settings.SECRET_KEY)
    send_verification_email(email, token)

    return {"status": "created", "user_id": user.id}

    """Authenticate user and return JWT token."""
    user = User.find_by_email(email)
    if not user:
        raise ValueError("Invalid credentials")

    if user.is_locked():
        raise ValueError("Account temporarily locked")

    if not bcrypt.checkpw(password.encode(), user.password_hash):
        _handle_failed_login(user)
        raise ValueError("Invalid credentials")

    _reset_login_attempts(user)
    token = jwt.encode({
        "user_id": user.id,
        "exp": datetime.utcnow() + timedelta(hours=1)
    }, Settings.SECRET_KEY)

    return token

def _handle_failed_login(user: User) -> None:
    """Track failed login attempts."""
    user.failed_attempts += 1
    if user.failed_attempts >= MAX_FAILED_ATTEMPTS:
        user.locked_until = datetime.now() + timedelta(hours=LOCKOUT_DURATION_HOURS)
    user.save()

def _reset_login_attempts(user: User) -> None:
    """Reset failed attempt counter on successful login."""
    user.failed_attempts = 0
    user.locked_until = None
    user.save()
```

### Phase 3: Sync Documentation (1 hour)

**Update SPEC status**:

```yaml
status: completed
version: 1.0.0
updated: 2025-11-03
```

**Update README.md**:

```markdown
## Authentication

User authentication system with email verification and brute-force protection.

- Email signup with verification
- JWT-based sessions
- Automatic lockout after 3 failed attempts
- See SPEC-AUTH-001 for requirements
```

**Generate sync report**:

```bash
rg '@(SPEC|TEST|CODE|DOC):AUTH-001' -n .moai/specs/ tests/ src/ docs/

# Output:
# README.md:SPEC-AUTH-001

✅ TAG chain verified: SPEC→TEST→CODE→DOC
```

## Example 2: Context Engineering in Action

### `/alfred:1-plan` Command

Only load:
```
✅ .moai/project/product.md (150 KB)
✅ .moai/project/structure.md (50 KB)

❌ Skip: All SPEC files, analysis, reports
```

Context saved: ~5 MB

### `/alfred:2-run` Command

Load only:
```
✅ .moai/specs/SPEC-AUTH-001/spec.md (20 KB)
✅ development-guide.md (14 KB) ← This Skill!

❌ Skip: Other SPECs, unrelated docs, analysis
```

Context saved: ~200 MB

## Example 3: TRUST 5 Validation Checklist

```markdown
# AUTH-001 Implementation Verification

## T – Test-Driven (SPEC-Aligned)
- [x] SPEC written with EARS format (5 patterns)
- [x] RED: Tests written first, all fail
- [x] GREEN: Minimal code passes all tests
- [x] REFACTOR: Code improved, tests still pass

## R – Readable
- [x] Function names: signup(), login(), _validate_email()
- [x] Variables: email, password, hashed, token
- [x] Comments explain WHY: "bcrypt 12 rounds for security"
- [x] Docstrings for public functions
- [x] README updated

## U – Unified
- [x] Naming: snake_case for functions, UPPER_CASE for constants
- [x] Error handling: consistent ValueError usage
- [x] Test structure: Arrange→Act→Assert pattern
- [x] SPEC format: 5 EARS patterns used
- [x] Documentation: SPEC→README→docstrings

## S – Secured
- [x] No hardcoded secrets (use Settings.SECRET_KEY)
- [x] Input validation: email, password format
- [x] Error handling: Generic "Invalid credentials"
- [x] Brute-force protection: 3 attempts → lockout
- [x] Password hashing: bcrypt with 12 rounds

## E – Evaluated (≥85% Coverage)
- [x] Test coverage: 92%
  - signup(): 95% (valid, invalid email, weak password)
  - login(): 89% (success, invalid, locked, failed attempts)
  - helper functions: 100%
- [x] Edge cases: lockout expiration, token expiration
- [x] Integration tests: signup→verify→login

✅ All TRUST 5 principles met!
```
