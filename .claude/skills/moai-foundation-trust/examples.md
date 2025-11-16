# moai-foundation-trust: TRUST 5 Implementation Examples

## Example 1: Complete TRUST 5 Implementation - Password Authentication

### Scenario: Secure User Authentication with All TRUST 5 Principles

#### SPEC: Requirements Definition

```markdown
# SPEC-AUTH-001: User Password Authentication


## Requirements

1. Users must securely set and verify passwords
2. Passwords stored as hashes (bcrypt)
3. Login attempts rate limited (5/minute)
4. Account locked after 5 failures
5. MFA support for sensitive operations

## TRUST 5 Compliance

- **T**est First: ≥85% coverage required
- **R**eadable: Clear function names, <50 lines per function
- **U**nified: Consistent with auth module patterns
- **S**ecured: OWASP A02 (cryptographic failures) compliance
```

#### T: Test First - Write Tests (RED Phase)

```python
# tests/test_auth_password.py

"""
Password authentication tests
Coverage target: 96% (exceeds 85% minimum)
"""

import pytest
from src.auth import hash_password, verify_password, authenticate_user
from src.exceptions import RateLimitError


class TestPasswordHashing:

    def test_hash_password_creates_bcrypt_hash(self):
        """
        Expected: Non-plaintext hash in bcrypt format
        """
        plaintext = "SecurePass123"
        hashed = hash_password(plaintext)
        
        assert plaintext not in hashed, "Plaintext leaked in hash"
        assert hashed.startswith("$2"), "Not bcrypt format"
        assert len(hashed) == 60, "Invalid bcrypt length"

    def test_hash_password_unique_per_salt(self):
        """
        Expected: Same password → different hashes
        """
        plaintext = "SamePassword"
        hash1 = hash_password(plaintext)
        hash2 = hash_password(plaintext)
        
        assert hash1 != hash2, "Hashes not unique"

    def test_hash_password_12_rounds(self):
        """
        Expected: Bcrypt with 12 rounds minimum
        """
        plaintext = "TestPassword"
        hashed = hash_password(plaintext)
        
        # Bcrypt format: $2b$rounds$salt$hash
        rounds = int(hashed.split("$")[2])
        assert rounds >= 12, "Rounds less than 12"


class TestPasswordVerification:

    def test_verify_correct_password(self):
        """
        Expected: Return True
        """
        plaintext = "CorrectPass123"
        hashed = hash_password(plaintext)
        
        result = verify_password(plaintext, hashed)
        assert result is True

    def test_verify_incorrect_password(self):
        """
        Expected: Return False
        """
        hashed = hash_password("CorrectPass123")
        
        result = verify_password("WrongPassword", hashed)
        assert result is False

    def test_verify_empty_password(self):
        """
        Expected: Raise ValueError
        """
        with pytest.raises(ValueError):
            verify_password("", hash_password("SomePass"))


class TestLoginRateLimiting:

    def test_login_rate_limit_5_per_minute(self):
        """
        Expected: 6th attempt blocked
        """
        email = "testuser@example.com"
        password = "TestPass123"
        
        # First 5 attempts fail with wrong password
        for attempt in range(5):
            with pytest.raises(ValueError, match="invalid"):
                authenticate_user(email, "WrongPassword")
        
        # 6th attempt blocked by rate limiter (even correct password)
        with pytest.raises(RateLimitError):
            authenticate_user(email, password)

    def test_rate_limit_reset_on_success(self):
        """
        Expected: Counter resets after successful login
        """
        email = "testuser@example.com"
        
        # Fail 3 times
        for _ in range(3):
            with pytest.raises(ValueError):
                authenticate_user(email, "WrongPassword")
        
        # Success resets counter
        token = authenticate_user(email, "CorrectPassword")
        assert token is not None
        
        # Can attempt again
        with pytest.raises(ValueError):
            authenticate_user(email, "WrongPassword")


class TestAccountLockout:

    def test_account_locked_after_5_failures(self):
        """
        Expected: Account locked for 30 minutes
        """
        email = "secure@example.com"
        
        # 5 failed attempts
        for _ in range(5):
            try:
                authenticate_user(email, "WrongPassword")
            except ValueError:
                pass
        
        # Check locked
        from src.auth import get_lock_status
        status = get_lock_status(email)
        assert status.is_locked
        assert status.locked_until is not None

    def test_locked_account_notification(self):
        """
        Expected: Email sent to user on lockout
        """
        email = "notified@example.com"
        
        # Trigger lockout
        for _ in range(5):
            try:
                authenticate_user(email, "WrongPassword")
            except (ValueError, RateLimitError):
                pass
        
        # Verify email was sent
        from src.email import get_sent_emails
        emails = get_sent_emails(to=email)
        assert any("locked" in e.body.lower() for e in emails)
```

#### R: Readable - Implementation Code

```python
# src/auth/password.py

"""
User password authentication module.

This module handles secure password operations including hashing,
verification, and authentication. It implements TRUST 5 principles
for high-quality, secure code.

Security (S): Uses bcrypt for password hashing (OWASP approved)
Readable (R): Clear function names, comprehensive docstrings
"""

import bcrypt
import logging
from datetime import datetime, timedelta
from typing import Tuple, Optional

logger = logging.getLogger(__name__)

BCRYPT_ROUNDS = 12  # November 2025 standard (was 10 in older versions)
RATE_LIMIT_ATTEMPTS = 5
RATE_LIMIT_WINDOW_MINUTES = 1
LOCKOUT_DURATION_MINUTES = 30


def hash_password(plaintext: str) -> str:
    """
    Hash password using bcrypt algorithm.
    
    Implements OWASP A02:2021 Cryptographic Failures prevention
    using bcrypt with 12 salt rounds (November 2025 standard).
    
    
    Security properties:
    - Uses bcrypt (OWASP recommended for passwords)
    - 12 salt rounds (resistant to GPU/ASIC attacks)
    - Auto-unique salt per hash
    - Non-reversible one-way function
    
    Args:
        plaintext: Raw password string
        
    Returns:
        Bcrypt hash string ($2b$12$...)
        
    Raises:
        ValueError: If plaintext is empty or not string
        
    Example:
        >>> hashed = hash_password("MyPassword123")
        >>> hashed.startswith("$2b$12$")
        True
    """
    if not plaintext or not isinstance(plaintext, str):
        raise ValueError("Password must be non-empty string")
    
    if len(plaintext) > 72:
        raise ValueError("Password too long (max 72 characters for bcrypt)")
    
    try:
        salt = bcrypt.gensalt(rounds=BCRYPT_ROUNDS)
        hashed_bytes = bcrypt.hashpw(plaintext.encode('utf-8'), salt)
        hashed_str = hashed_bytes.decode('utf-8')
        
        logger.debug("Password hashed successfully")
        return hashed_str
        
    except Exception as e:
        logger.error(f"Hashing failed: {e}")
        raise ValueError("Password hashing failed") from e


def verify_password(plaintext: str, hashed: str) -> bool:
    """
    Verify plaintext password against bcrypt hash.
    
    
    Safe comparison prevents timing attacks by using
    bcrypt's constant-time comparison.
    
    Args:
        plaintext: Raw password to verify
        hashed: Bcrypt hash to check against
        
    Returns:
        True if password matches, False otherwise
        
    Example:
        >>> hashed = hash_password("MyPassword")
        >>> verify_password("MyPassword", hashed)
        True
        >>> verify_password("WrongPassword", hashed)
        False
    """
    if not plaintext or not hashed:
        return False
    
    try:
        return bcrypt.checkpw(
            plaintext.encode('utf-8'),
            hashed.encode('utf-8')
        )
    except (ValueError, TypeError):
        # Hash format invalid or comparison failed
        return False


def authenticate_user(email: str, password: str) -> str:
    """
    Authenticate user and return JWT token.
    
    
    Implements:
    - Rate limiting (5 attempts per minute)
    - Account lockout (30 minutes after 5 failures)
    - Password verification against bcrypt hash
    
    Args:
        email: User email address
        password: Raw password to verify
        
    Returns:
        JWT authentication token
        
    Raises:
        ValueError: If credentials invalid
        RateLimitError: If rate limited
        AccountLockedError: If account locked
        
    Example:
        >>> token = authenticate_user("user@example.com", "Pass123")
        >>> token.startswith("eyJ")
        True
    """
    from src.models import User
    from src.rate_limiter import RateLimiter
    from src.exceptions import RateLimitError, AccountLockedError
    
    limiter = RateLimiter()
    if not limiter.allow_attempt(email):
        raise RateLimitError(f"Too many login attempts for {email}")
    
    user = User.query.filter_by(email=email).first()
    if not user:
        logger.warning(f"Login attempt for non-existent user: {email}")
        raise ValueError("Invalid email or password")
    
    if _is_account_locked(user):
        raise AccountLockedError("Account locked. Try again later.")
    
    if not verify_password(password, user.password_hash):
        _increment_failed_attempts(user)
        logger.warning(f"Failed login for {email}")
        raise ValueError("Invalid email or password")
    
    _reset_failed_attempts(user)
    logger.info(f"Successful login for {email}")
    
    token = _generate_token(user)
    return token


# Helper functions (also tagged)

def _is_account_locked(user: 'User') -> bool:
    if not user.locked_until:
        return False
    return datetime.utcnow() < user.locked_until


def _increment_failed_attempts(user: 'User') -> None:
    user.failed_attempts += 1
    if user.failed_attempts >= RATE_LIMIT_ATTEMPTS:
        user.locked_until = datetime.utcnow() + timedelta(
            minutes=LOCKOUT_DURATION_MINUTES
        )
        logger.warning(f"Account locked: {user.email}")
        # Send notification email
        from src.email import send_lockout_notification
        send_lockout_notification(user.email)
    
    user.save()


def _reset_failed_attempts(user: 'User') -> None:
    user.failed_attempts = 0
    user.locked_until = None
    user.save()


def _generate_token(user: 'User') -> str:
    import jwt
    from datetime import datetime, timedelta
    
    payload = {
        'user_id': user.id,
        'email': user.email,
        'exp': datetime.utcnow() + timedelta(hours=24)
    }
    
    # Note: Secret should come from environment, not hardcoded
    import os
    secret = os.getenv('JWT_SECRET')
    if not secret:
        raise ValueError("JWT_SECRET not configured")
    
    return jwt.encode(payload, secret, algorithm='HS256')
```

#### U: Unified - Architecture Consistency

```python
# src/auth/__init__.py
"""
Authentication module exports.

Follows unified module structure:
- Imports grouped by category
- Clear public API
- Private implementation hidden
"""

# Unified import structure (enforced across all modules)
# 1. Standard library
import logging
from typing import Optional

# 2. Third-party
import bcrypt

# 3. Local application
from .password import hash_password, verify_password, authenticate_user
from .mfa import enable_mfa, verify_mfa

logger = logging.getLogger(__name__)

# Public API
__all__ = [
    'hash_password',
    'verify_password',
    'authenticate_user',
    'enable_mfa',
    'verify_mfa',
]

# Consistent naming pattern across auth module
# authenticate_user() - public function
# _is_account_locked() - private helper
# AuthenticationManager - public class (if used)
```

#### S: Secured - Security Implementation

```python

# ✓ OWASP A02: Cryptographic Failures - MITIGATED
#   Using bcrypt instead of MD5/SHA1
#   Bcrypt with 12 rounds (2025 standard)

# ✓ OWASP A07: Authentication Failures - MITIGATED
#   Rate limiting (5 attempts/minute)
#   Account lockout (30 minutes)
#   Password strength requirements

# ✓ OWASP A05: Access Control - MITIGATED
#   Token validation on protected endpoints
#   Role-based access control (RBAC)

# Security Testing Integration
def test_password_not_logged():
    """Verify passwords never logged"""
    import logging
    
    with patch('src.auth.logger') as mock_logger:
        authenticate_user("test@example.com", "SecretPass123")
        
        # Verify password not in any log message
        for call in mock_logger.mock_calls:
            assert "SecretPass123" not in str(call)
            assert "SecretPass" not in str(call)

def test_timing_attack_resistance():
    """Verify timing-attack resistant comparison"""
    import time
    
    hashed = hash_password("CorrectPassword")
    
    # Wrong password vs non-existent hash
    # Both should take similar time (prevents brute force)
    
    start = time.time()
    verify_password("WrongPassword", hashed)
    wrong_time = time.time() - start
    
    start = time.time()
    verify_password("WrongPassword", "invalid_hash")
    invalid_time = time.time() - start
    
    # Times should be similar (within 10ms)
    assert abs(wrong_time - invalid_time) < 0.01
```

#### T: Trackable - Documentation

```markdown
# docs/authentication.md

Authentication System Documentation

## Overview

User authentication using secure password hashing and token-based access.

### References

## User Authentication

### Login Flow

1. User submits email + password
2. System validates rate limits
3. System checks account lockout status
4. System verifies password against bcrypt hash
5. System generates JWT token
6. System returns token to client

## Security Features

- **Password Hashing**: bcrypt with 12 rounds (OWASP standard)
- **Rate Limiting**: 5 login attempts per minute per email
- **Account Lockout**: 30-minute lockout after 5 failures
- **Notifications**: Email alerts on lockout events

```

#### Complete TRUST 5 Chain for Example 1

```
SPEC Layer: SPEC-AUTH-001
  Location: .moai/specs/SPEC-001/spec.md
  Status: APPROVED
  
TEST Layer: 15+ Tests, 96% Coverage
  ... (10+ more)
  
CODE Layer: 6 Functions + 4 Helpers
  
DOC Layer: Complete Guide
  
TRUST 5 Validation:
  ✓ Test First: 96% coverage (exceeds 85% minimum)
  ✓ Readable: Pylint 9.2/10, CC=4 avg
  ✓ Unified: Consistent patterns across auth module
  ✓ Secured: OWASP Top 10 compliance verified
  ✓ Trackable: SPEC→TEST→CODE→DOC chain complete
  
Status: PRODUCTION READY ✓
```

---

## Example 2: TRUST 5 Quality Gate in CI/CD

### Before Commit

```bash
$ git diff  # Review changes
$ pytest --cov=src --cov-fail-under=85  # Test First
$ pylint src/ --fail-under=8.0           # Readable
$ bandit -r src/ -ll                     # Secured
$ python .moai/scripts/validation/tag_validator.py  # Trackable

All gates passed! ✓
Ready to commit
```

### After Commit (CI/CD Pipeline)

```
commit: "feat: Add secure password authentication (SPEC-AUTH-001)"
  ↓
GitHub Actions triggered
  ├─ T: Test First
  │  ├─ Run pytest --cov
  │  └─ Coverage: 96.2% ✓ PASS (≥85%)
  │
  ├─ R: Readable
  │  ├─ Run pylint
  │  ├─ Pylint: 9.2/10 ✓ PASS (≥8.0)
  │  └─ Black: 100% compliant ✓
  │
  ├─ U: Unified
  │  ├─ Architecture check
  │  └─ Consistency: 100% ✓ PASS
  │
  ├─ S: Secured
  │  ├─ Bandit scan
  │  ├─ No vulnerabilities ✓ PASS
  │  └─ Dependency audit: OK ✓
  │
  ├─ T: Trackable
  │  ├─ TAG validation
  │  ├─ Chains: SPEC→TEST→CODE→DOC ✓
  │  └─ No orphans ✓ PASS
  │
  └─ Overall: SUCCESS ✓
     ✓ All gates passed
     ✓ Ready to merge
```

