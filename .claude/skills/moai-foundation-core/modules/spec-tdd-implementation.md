# TDD Implementation Patterns

Purpose: Advanced RED-GREEN-REFACTOR workflows and test implementation patterns for SPEC-First TDD.

Version: 1.0.0
Last Updated: 2026-01-06
Parent: [spec-first-tdd.md](spec-first-tdd.md)

---

## Complete TDD Example

### User Registration Implementation

```python
# tests/auth/test_registration.py
import pytest
from src.auth.registration import register_user
from src.exceptions import ValidationError

class TestUserRegistration:
    """Test suite for SPEC-001-REQ-01: User Registration."""

    def test_register_user_success(self):
        """Test successful registration with valid data."""
        result = register_user("user@example.com", "SecureP@ssw0rd")
        assert result.success is True
        assert result.user.email == "user@example.com"
        assert result.confirmation_sent is True

    def test_register_user_invalid_email(self):
        """Test registration fails with invalid email format."""
        with pytest.raises(ValidationError, match="Invalid email format"):
            register_user("invalid-email", "SecureP@ssw0rd")

    def test_register_user_duplicate_email(self):
        """Test registration fails with duplicate email."""
        register_user("user@example.com", "SecureP@ssw0rd")
        with pytest.raises(ValidationError, match="Email already registered"):
            register_user("user@example.com", "AnotherP@ssw0rd")

    def test_register_user_weak_password(self):
        """Test registration fails with weak password."""
        weak_passwords = [
            "short",            # Too short
            "alllowercase1",    # No uppercase
            "ALLUPPERCASE1",    # No lowercase
            "NoNumbersHere!",   # No digits
            "NoSymbols123"      # No symbols
        ]

        for weak_pwd in weak_passwords:
            with pytest.raises(ValidationError, match="Password must be"):
                register_user("user@example.com", weak_pwd)

    def test_register_user_password_hashing(self):
        """Test password is properly hashed (not stored plain text)."""
        result = register_user("user@example.com", "SecureP@ssw0rd")
        assert result.user.password_hash != "SecureP@ssw0rd"
        assert result.user.password_hash.startswith("$2b$")
```

---

## Refactored Implementation

```python
# src/auth/registration.py
from dataclasses import dataclass
from typing import Optional
import bcrypt
import re

EMAIL_REGEX = re.compile(r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$')
PASSWORD_MIN_LENGTH = 8

@dataclass
class RegistrationResult:
    """Result of user registration attempt."""
    success: bool
    user: Optional[User]
    confirmation_sent: bool
    error: Optional[str] = None

def register_user(email: str, password: str) -> RegistrationResult:
    """Register new user with email and password.

    Implements SPEC-001-REQ-01: User Registration (Ubiquitous)

    Args:
        email: User email address (must be valid format)
        password: User password (≥8 chars, mixed case, numbers, symbols)

    Returns:
        RegistrationResult with user data or error

    Raises:
        ValidationError: If email or password invalid
    """
    if not EMAIL_REGEX.match(email):
        raise ValidationError("Invalid email format")

    existing_user = User.query.filter_by(email=email).first()
    if existing_user:
        raise ValidationError("Email already registered")

    if not _is_password_strong(password):
        raise ValidationError(
            "Password must be ≥8 characters with mixed case, numbers, and symbols"
        )

    password_hash = bcrypt.hashpw(
        password.encode('utf-8'),
        bcrypt.gensalt(rounds=12)
    ).decode('utf-8')

    user = User(email=email, password_hash=password_hash)
    db_session.add(user)
    db_session.commit()

    confirmation_sent = send_confirmation_email(user.email, user.id)

    return RegistrationResult(
        success=True,
        user=user,
        confirmation_sent=confirmation_sent
    )

def _is_password_strong(password: str) -> bool:
    """Validate password strength."""
    if len(password) < PASSWORD_MIN_LENGTH:
        return False

    has_upper = any(c.isupper() for c in password)
    has_lower = any(c.islower() for c in password)
    has_digit = any(c.isdigit() for c in password)
    has_symbol = any(not c.isalnum() for c in password)

    return all([has_upper, has_lower, has_digit, has_symbol])
```

---

## MFA Implementation Example

```python
# RED PHASE: Complex test scenario
def test_mfa_verification_with_valid_code():
    """SPEC-002-REQ-03: MFA verification happy path."""
    user = create_user_with_mfa_enabled()
    login_attempt = initiate_login(user.email, user.password)
    totp_code = get_pending_totp_code(user.id)

    result = verify_mfa(user.id, code=totp_code, timestamp=datetime.now())

    assert result.success is True
    assert result.token is not None
    assert result.mfa_verified is True

def test_mfa_verification_with_expired_code():
    """SPEC-002-REQ-03: MFA code expiry."""
    user = create_user_with_mfa_enabled()
    totp_code = get_pending_totp_code(user.id)

    result = verify_mfa(
        user.id, code=totp_code,
        timestamp=datetime.now() + timedelta(minutes=6)
    )

    assert result.success is False
    assert result.error == "Code expired"

def test_mfa_rate_limiting():
    """SPEC-002-REQ-03: Rate limiting after failed attempts."""
    user = create_user_with_mfa_enabled()

    for _ in range(3):
        verify_mfa(user.id, code="000000")

    result = verify_mfa(user.id, code="123456")

    assert result.success is False
    assert "Too many failed attempts" in result.error
```

---

## Iterative SPEC Refinement

```python
# Initial SPEC (v1.0.0)
# SPEC-003-REQ-01: File upload with size limit (10MB)

# Implementation reveals edge case
# → User uploads 9.9MB file successfully
# → But total storage exceeds user quota

# Refined SPEC (v1.1.0)
# SPEC-003-REQ-01: File upload with size and quota validation
# - Single file limit: 10MB
# - User quota limit: 100MB total
# - Validation: Check both limits before accepting upload

# Updated test
def test_file_upload_exceeds_quota():
    """SPEC-003-REQ-01 v1.1.0: Quota validation."""
    user = create_user(quota_limit_mb=100)
    upload_files(user, total_size_mb=95)

    result = upload_file(user, file_size_mb=8)

    assert result.success is False
    assert result.error == "Upload would exceed storage quota"
```

---

## CI/CD Pipeline Integration

```yaml
# .github/workflows/spec-tdd-pipeline.yml
name: SPEC-First TDD Pipeline

on:
  push:
    paths:
      - '.moai/specs/**'
      - 'src/**'
      - 'tests/**'

jobs:
  spec-validation:
    name: "Phase 1: SPEC Validation"
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Validate SPEC format
        run: python .moai/scripts/validate_spec.py
      - name: Check requirement traceability
        run: python .moai/scripts/check_traceability.py

  tdd-implementation:
    name: "Phase 2: TDD Implementation"
    needs: spec-validation
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run RED phase tests
        run: pytest tests/ -v --tb=short || true
      - name: Verify tests exist for all requirements
        run: python .moai/scripts/verify_test_coverage_mapping.py

  quality-gates:
    name: "Phase 3: Quality Gates"
    needs: tdd-implementation
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run tests with coverage
        run: pytest --cov=src --cov-fail-under=85
      - name: Validate TRUST 5
        run: python .moai/scripts/validate_trust5.py
```

---

## Works Well With

- [spec-first-tdd.md](spec-first-tdd.md) - Main workflow overview
- [spec-ears-format.md](spec-ears-format.md) - EARS patterns

---

Version: 1.0.0
Last Updated: 2026-01-06
