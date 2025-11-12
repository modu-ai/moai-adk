# moai-foundation-tags: Practical Examples & Implementation Guide

## Example 1: Complete Feature Implementation with Full TAG Chain

### Scenario: User Authentication Feature

#### Day 1: SPEC Creation

**File**: `.moai/specs/SPEC-001/spec.md`

```markdown
---
title: SPEC-001 User Authentication System
created: 2025-11-12
status: APPROVED
author: spec-builder
@SPEC:AUTH-001
---

# SPEC-001: User Authentication System

## Overview
Implement secure user authentication system with email/password login, password hashing, and rate limiting.

## Requirements

@REQ-AUTH-001-001: Users can register with email and password
- Email format validation
- Password complexity validation (min 8 chars, 1 uppercase, 1 digit)
- Unique email constraint

@REQ-AUTH-001-002: Users can log in with email/password
- Verify email/password combination
- Rate limit login attempts (max 5 per minute)
- Return authentication token on success

@REQ-AUTH-001-003: Passwords must be securely hashed
- Use bcrypt with salt rounds ≥ 10
- Never store plaintext passwords
- Support password verification without reversal

@REQ-AUTH-001-004: Account lockout on multiple failures
- Lock after 5 failed attempts
- Auto-unlock after 30 minutes
- Email notification on lockout

## Acceptance Criteria

- [ ] All password validation tests pass
- [ ] Rate limiting works end-to-end
- [ ] Coverage >= 85%
- [ ] No hardcoded credentials in code
- [ ] SQL injection prevention verified
```

#### Day 2-3: Tests Implementation (RED Phase)

**File**: `tests/test_auth.py`

```python
"""
User Authentication Tests

@TEST:SPEC:AUTH-001
Validates all requirements in SPEC-001
"""

import pytest
from src.auth import (
    register_user,
    authenticate_user,
    hash_password,
    verify_password,
    get_account_lock_status
)


class TestUserRegistration:
    """@TEST:SPEC:AUTH-001-001: User registration"""

    def test_register_with_valid_email_password(self):
        """
        @TEST:SPEC:AUTH-001-001-001: Registration with valid inputs
        Requirements: Email format + password complexity
        """
        user = register_user("test@example.com", "SecurePass123")
        assert user.email == "test@example.com"
        assert user.id is not None
        assert not user.password_plaintext  # Never stored

    def test_register_reject_invalid_email(self):
        """@TEST:SPEC:AUTH-001-001-002: Email format validation"""
        with pytest.raises(ValueError) as exc:
            register_user("invalid-email", "SecurePass123")
        assert "invalid email" in str(exc.value)

    def test_register_reject_weak_password(self):
        """@TEST:SPEC:AUTH-001-001-003: Password complexity validation"""
        with pytest.raises(ValueError) as exc:
            register_user("test@example.com", "weak")
        assert "password" in str(exc.value).lower()

    def test_register_reject_duplicate_email(self):
        """@TEST:SPEC:AUTH-001-001-004: Unique email constraint"""
        register_user("test@example.com", "SecurePass123")
        with pytest.raises(ValueError) as exc:
            register_user("test@example.com", "AnotherPass123")
        assert "already exists" in str(exc.value)


class TestUserLogin:
    """@TEST:SPEC:AUTH-001-002: User login"""

    @pytest.fixture
    def registered_user(self):
        return register_user("user@example.com", "TestPass123")

    def test_login_with_correct_credentials(self, registered_user):
        """
        @TEST:SPEC:AUTH-001-002-001: Successful login
        Requirement: Verify credentials and return token
        """
        token = authenticate_user("user@example.com", "TestPass123")
        assert token is not None
        assert len(token) > 0

    def test_login_reject_incorrect_password(self, registered_user):
        """@TEST:SPEC:AUTH-001-002-002: Reject wrong password"""
        with pytest.raises(ValueError) as exc:
            authenticate_user("user@example.com", "WrongPassword")
        assert "invalid" in str(exc.value).lower()

    def test_login_rate_limiting(self, registered_user):
        """
        @TEST:SPEC:AUTH-001-002-003: Rate limit 5 attempts/minute
        Requirement: Prevent brute force attacks
        """
        # First 5 attempts should fail
        for i in range(5):
            with pytest.raises(ValueError):
                authenticate_user("user@example.com", "WrongPass")
        
        # 6th attempt should be blocked by rate limiter
        with pytest.raises(RateLimitError) as exc:
            authenticate_user("user@example.com", "TestPass123")
        assert "rate limit" in str(exc.value).lower()


class TestPasswordHashing:
    """@TEST:SPEC:AUTH-001-003: Password security"""

    def test_hash_password_non_reversible(self):
        """@TEST:SPEC:AUTH-001-003-001: Bcrypt hashing"""
        hashed = hash_password("SecurePass123")
        # Hash should not contain original password
        assert "SecurePass123" not in hashed
        assert len(hashed) > 20

    def test_hash_password_unique_per_salt(self):
        """@TEST:SPEC:AUTH-001-003-002: Unique salts"""
        hash1 = hash_password("SamePassword")
        hash2 = hash_password("SamePassword")
        # Same password should produce different hashes (different salt)
        assert hash1 != hash2

    def test_verify_password_correct(self):
        """@TEST:SPEC:AUTH-001-003-003: Password verification"""
        hashed = hash_password("TestPass123")
        assert verify_password("TestPass123", hashed) is True

    def test_verify_password_incorrect(self):
        """@TEST:SPEC:AUTH-001-003-004: Reject wrong password"""
        hashed = hash_password("TestPass123")
        assert verify_password("WrongPass123", hashed) is False


class TestAccountLockout:
    """@TEST:SPEC:AUTH-001-004: Account security"""

    def test_account_locked_after_5_failures(self):
        """@TEST:SPEC:AUTH-001-004-001: Lockout after 5 attempts"""
        register_user("secure@example.com", "SecurePass123")
        
        # 5 failed attempts
        for i in range(5):
            with pytest.raises(ValueError):
                authenticate_user("secure@example.com", "Wrong")
        
        # Account should be locked
        status = get_account_lock_status("secure@example.com")
        assert status.is_locked is True

    def test_account_locked_notification(self):
        """@TEST:SPEC:AUTH-001-004-002: Email notification on lockout"""
        # Trigger lockout (implementation detail)
        # Email should be sent
        # Note: Real test would use email mock
        pass

    def test_account_auto_unlock_after_30_minutes(self):
        """@TEST:SPEC:AUTH-001-004-003: Auto-unlock timeout"""
        # This would use mock time travel
        # Test that account unlocks after 30 minutes
        pass
```

#### Day 4-5: Code Implementation (GREEN Phase)

**File**: `src/auth.py`

```python
"""
User Authentication Module

@CODE:SPEC:AUTH-001
Implements user authentication requirements from SPEC-001

References:
- @TEST:SPEC:AUTH-001-001-001 (User registration)
- @TEST:SPEC:AUTH-001-002-001 (User login)
- @TEST:SPEC:AUTH-001-003 (Password hashing)
- @TEST:SPEC:AUTH-001-004 (Account lockout)
"""

import re
import bcrypt
from datetime import datetime, timedelta
from typing import Optional
from src.models import User
from src.rate_limiter import RateLimiter


rate_limiter = RateLimiter(max_attempts=5, window_minutes=1)


@CODE:SPEC:AUTH-001-001
def register_user(email: str, password: str) -> User:
    """
    Register new user with email and password
    
    @CODE:SPEC:AUTH-001-001: User registration
    Validates email format and password complexity
    
    Args:
        email: User email address
        password: User password (min 8 chars, 1 uppercase, 1 digit)
    
    Returns:
        Registered User object
    
    Raises:
        ValueError: If email format invalid or password weak
    """
    # @CODE:SPEC:AUTH-001-001-001: Email validation
    if not _is_valid_email(email):
        raise ValueError("Email format invalid")
    
    # @CODE:SPEC:AUTH-001-001-002: Password complexity
    if not _is_strong_password(password):
        raise ValueError("Password must be ≥8 chars with uppercase and digit")
    
    # @CODE:SPEC:AUTH-001-001-003: Unique constraint
    if User.query.filter_by(email=email).first():
        raise ValueError(f"Email {email} already exists")
    
    # @CODE:SPEC:AUTH-001-001-004: Hash password before storage
    hashed_pwd = hash_password(password)
    
    user = User(email=email, password_hash=hashed_pwd)
    user.save()
    return user


@CODE:SPEC:AUTH-001-002
def authenticate_user(email: str, password: str) -> str:
    """
    Authenticate user and return token
    
    @CODE:SPEC:AUTH-001-002: User login
    Verifies credentials with rate limiting
    
    Args:
        email: User email
        password: User password (plaintext)
    
    Returns:
        Authentication token
    
    Raises:
        ValueError: If credentials invalid
        RateLimitError: If too many attempts
    """
    # @CODE:SPEC:AUTH-001-002-003: Rate limiting
    if not rate_limiter.allow_attempt(email):
        raise RateLimitError(f"Too many login attempts for {email}")
    
    # @CODE:SPEC:AUTH-001-004-001: Check lockout status
    user = User.query.filter_by(email=email).first()
    if not user:
        raise ValueError("Invalid email or password")
    
    if _is_account_locked(user):
        raise ValueError("Account locked. Try again later.")
    
    # @CODE:SPEC:AUTH-001-002-001: Credential verification
    if not verify_password(password, user.password_hash):
        _increment_failed_attempts(user)
        raise ValueError("Invalid email or password")
    
    # Reset failed attempts on success
    _reset_failed_attempts(user)
    
    # Generate and return token
    token = _generate_token(user)
    return token


@CODE:SPEC:AUTH-001-003
def hash_password(plaintext: str) -> str:
    """
    Hash password using bcrypt
    
    @CODE:SPEC:AUTH-001-003: Password security
    Implements secure hashing with salt
    
    Args:
        plaintext: Plain text password
    
    Returns:
        Hashed password (bcrypt format)
    """
    # @CODE:SPEC:AUTH-001-003-001: Bcrypt with salt rounds ≥ 10
    salt = bcrypt.gensalt(rounds=12)
    hashed = bcrypt.hashpw(plaintext.encode('utf-8'), salt)
    return hashed.decode('utf-8')


@CODE:SPEC:AUTH-001-003
def verify_password(plaintext: str, hashed: str) -> bool:
    """
    Verify plaintext password against hash
    
    @CODE:SPEC:AUTH-001-003: Password verification
    
    Args:
        plaintext: Plain text password to verify
        hashed: Bcrypt hash to check against
    
    Returns:
        True if password matches, False otherwise
    """
    return bcrypt.checkpw(
        plaintext.encode('utf-8'),
        hashed.encode('utf-8')
    )


# Helper functions (also tagged)

@CODE:SPEC:AUTH-001-001-001
def _is_valid_email(email: str) -> bool:
    """@CODE:SPEC:AUTH-001-001-001: Email format validation"""
    pattern = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
    return re.match(pattern, email) is not None


@CODE:SPEC:AUTH-001-001-002
def _is_strong_password(password: str) -> bool:
    """@CODE:SPEC:AUTH-001-001-002: Password complexity check"""
    return (
        len(password) >= 8 and
        any(c.isupper() for c in password) and
        any(c.isdigit() for c in password)
    )


@CODE:SPEC:AUTH-001-004
def _is_account_locked(user: User) -> bool:
    """@CODE:SPEC:AUTH-001-004: Check lock status"""
    if not user.locked_until:
        return False
    return datetime.utcnow() < user.locked_until


@CODE:SPEC:AUTH-001-004-001
def _increment_failed_attempts(user: User) -> None:
    """@CODE:SPEC:AUTH-001-004-001: Track failed attempts"""
    user.failed_attempts += 1
    if user.failed_attempts >= 5:
        # @CODE:SPEC:AUTH-001-004-001: Lock for 30 minutes
        user.locked_until = datetime.utcnow() + timedelta(minutes=30)
    user.save()


@CODE:SPEC:AUTH-001-004
def _reset_failed_attempts(user: User) -> None:
    """@CODE:SPEC:AUTH-001-004: Reset attempts on success"""
    user.failed_attempts = 0
    user.locked_until = None
    user.save()


def get_account_lock_status(email: str) -> dict:
    """
    @CODE:SPEC:AUTH-001-004: Get lockout status
    """
    user = User.query.filter_by(email=email).first()
    if not user:
        return {"locked": False}
    
    return {
        "locked": _is_account_locked(user),
        "locked_until": user.locked_until,
        "failed_attempts": user.failed_attempts
    }


def _generate_token(user: User) -> str:
    """@CODE:SPEC:AUTH-001-002: Generate auth token"""
    import jwt
    payload = {
        "user_id": user.id,
        "email": user.email,
        "exp": datetime.utcnow() + timedelta(hours=24)
    }
    return jwt.encode(payload, "secret_key", algorithm="HS256")
```

#### Day 6: Documentation

**File**: `docs/authentication.md`

```markdown
# User Authentication

@DOC:SPEC:AUTH-001
Complete guide to user authentication system

## Overview

@DOC:SPEC:AUTH-001-001: User Registration
Users can register with email and password. The system validates email format and enforces password complexity requirements.

## User Registration

@DOC:SPEC:AUTH-001-001
### Creating a New Account

```python
from src.auth import register_user

user = register_user("newuser@example.com", "SecurePass123")
```

**Password Requirements**:
- Minimum 8 characters
- At least 1 uppercase letter
- At least 1 digit

## User Login

@DOC:SPEC:AUTH-001-002
### Authenticating Users

```python
from src.auth import authenticate_user

try:
    token = authenticate_user("user@example.com", "SecurePass123")
    # Use token for authenticated requests
except ValueError:
    print("Invalid credentials")
```

**Rate Limiting**: Maximum 5 login attempts per minute per email address.

**Account Lockout**: After 5 failed attempts, account is locked for 30 minutes.

## Password Security

@DOC:SPEC:AUTH-001-003
### How Passwords Are Stored

Passwords are hashed using bcrypt with 12 rounds of salting. This ensures:
- Passwords cannot be reversed
- Same password creates different hashes
- Resistant to dictionary and brute force attacks

## Account Lockout

@DOC:SPEC:AUTH-001-004
### Security Features

- **Brute Force Protection**: Limited login attempts prevent automated attacks
- **Account Lockout**: Automatic 30-minute lockout after 5 failed attempts
- **Notification**: Email sent when account is locked

See [migration-guide.md] for upgrading from legacy authentication.
```

### TAG Chain Summary for Example 1

```
SPEC Layer:
  .moai/specs/SPEC-001/spec.md
  @SPEC:AUTH-001

TEST Layer:
  tests/test_auth.py
  @TEST:SPEC:AUTH-001-001  (Registration)
  @TEST:SPEC:AUTH-001-002  (Login)
  @TEST:SPEC:AUTH-001-003  (Hashing)
  @TEST:SPEC:AUTH-001-004  (Lockout)

CODE Layer:
  src/auth.py
  @CODE:SPEC:AUTH-001       (Main module)
  @CODE:SPEC:AUTH-001-001   (register_user)
  @CODE:SPEC:AUTH-001-002   (authenticate_user)
  @CODE:SPEC:AUTH-001-003   (Password functions)
  @CODE:SPEC:AUTH-001-004   (Lockout logic)

DOC Layer:
  docs/authentication.md
  @DOC:SPEC:AUTH-001        (Main doc)
  @DOC:SPEC:AUTH-001-001    (Registration section)
  @DOC:SPEC:AUTH-001-002    (Login section)
  @DOC:SPEC:AUTH-001-003    (Security section)
  @DOC:SPEC:AUTH-001-004    (Lockout section)

Chain Verification:
  ✓ 1 SPEC with 4 sub-requirements
  ✓ 4 TEST groups (12 individual tests)
  ✓ 4 CODE functions (+ 5 helpers)
  ✓ 4 DOC sections
  ✓ Coverage: 96% (well above 85% minimum)
  ✓ All chains linked and valid
```

---

## Example 2: Bug Fix with TAG Chain

### Scenario: Fix Password Reset Token Expiration

#### Original SPEC Reference

```
@SPEC:AUTH-001
  Requirement: "Passwords must be securely hashed"
  Status: APPROVED
```

#### Issue Reported

```
BUG-042: Password reset tokens expire after 12 hours
Expected: 24 hours
Current: 12 hours
Impact: Users unable to reset password after 12 hours
Linked To: @SPEC:AUTH-001 (password security requirement)
```

#### Fix Process

**1. Create Bug Fix SPEC**

```markdown
@SPEC:BUG-FIX-AUTH-TOKEN-001
title: Fix password reset token expiration
relates_to: @SPEC:AUTH-001
severity: HIGH
```

**2. Write Tests (RED)**

```python
@TEST:SPEC:BUG-FIX-AUTH-TOKEN-001-001
def test_reset_token_valid_24_hours():
    """Token should remain valid for 24 hours"""
    user = register_user("test@example.com", "Pass123")
    token = create_password_reset_token(user.email)
    
    # Check valid at 23 hours
    assert token_is_valid(token, hours=23) is True
    
    # Check invalid at 25 hours
    assert token_is_valid(token, hours=25) is False
```

**3. Implement Fix (GREEN)**

```python
@CODE:SPEC:BUG-FIX-AUTH-TOKEN-001
def create_password_reset_token(email: str) -> str:
    """
    Create password reset token
    
    Previously: 12 hour expiration (BUG)
    Now: 24 hour expiration (FIXED)
    """
    user = User.query.filter_by(email=email).first()
    payload = {
        "email": email,
        # @CODE:SPEC:BUG-FIX-AUTH-TOKEN-001-001: Fixed to 24 hours
        "exp": datetime.utcnow() + timedelta(hours=24)  # Was: hours=12
    }
    return jwt.encode(payload, "secret", algorithm="HS256")
```

**4. Document Fix (DOC)**

```markdown
@DOC:SPEC:BUG-FIX-AUTH-TOKEN-001
## Password Reset Token Fix

**Issue**: Reset tokens expired after 12 hours  
**Fix**: Extended expiration to 24 hours  
**Date**: 2025-11-13  
**Status**: RESOLVED

See @DOC:SPEC:AUTH-001-003 for password security details
```

**5. TAG Chain for Bug Fix**

```
Bug Fix Chain:
├─ Issue: BUG-042
├─ Related SPEC: @SPEC:AUTH-001
├─ Fix SPEC: @SPEC:BUG-FIX-AUTH-TOKEN-001
├─ Test: @TEST:SPEC:BUG-FIX-AUTH-TOKEN-001-001
├─ Code: @CODE:SPEC:BUG-FIX-AUTH-TOKEN-001
└─ Doc: @DOC:SPEC:BUG-FIX-AUTH-TOKEN-001
```

---

## Example 3: Detecting and Fixing an Orphan TAG

### Scenario: Orphan @CODE TAG Without SPEC

**Found During Scan**:
```bash
$ rg '@CODE:' src/
src/payment.py:15: @CODE:PAYMENT-PROCESSOR-001
# But no @SPEC:PAYMENT-PROCESSOR-001 exists!
```

**Analysis**:
```python
# File: src/payment.py
@CODE:PAYMENT-PROCESSOR-001
def process_payment(order_id, amount):
    """
    Process payment through Stripe
    This was copied from legacy code but has no SPEC
    """
    pass
```

**Resolution**:

**Step 1: Create Missing SPEC**
```markdown
@SPEC:PAYMENT-PROCESSOR-001
title: Payment Processing Module
description: Process payments via Stripe API
```

**Step 2: Create Tests**
```python
@TEST:SPEC:PAYMENT-PROCESSOR-001-001
def test_process_valid_payment():
    """Process payment with valid inputs"""
    pass

@TEST:SPEC:PAYMENT-PROCESSOR-001-002
def test_process_payment_rate_limit():
    """Respect Stripe rate limits"""
    pass
```

**Step 3: Link Code**
```python
@CODE:SPEC:PAYMENT-PROCESSOR-001
def process_payment(order_id, amount):
    """Now properly linked to SPEC"""
    pass
```

**Step 4: Document**
```markdown
@DOC:SPEC:PAYMENT-PROCESSOR-001
## Payment Processing
Handles order payment processing via Stripe...
```

---

## Validation Workflow

### November 2025 Validation Pipeline

```bash
#!/bin/bash
# Run full TAG validation before commit

echo "1. Scanning TAGs..."
rg '@(SPEC|TEST|CODE|DOC):' --count-matches

echo "2. Checking chains..."
python .moai/scripts/validation/tag_chain_validator.py

echo "3. Finding orphans..."
python .moai/scripts/validation/orphan_detector.py

echo "4. Verifying coverage..."
coverage report --minimum=85

if [ $? -eq 0 ]; then
    echo "✓ All validations passed"
    git commit -m "feat: Implementation with complete TAG chains"
else
    echo "✗ Validation failed. Fix issues before committing."
    exit 1
fi
```

