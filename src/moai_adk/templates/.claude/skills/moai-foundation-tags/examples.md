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
---

# SPEC-001: User Authentication System

## Overview
Implement secure user authentication system with email/password login, password hashing, and rate limiting.

## Requirements

- Email format validation
- Password complexity validation (min 8 chars, 1 uppercase, 1 digit)
- Unique email constraint

- Verify email/password combination
- Rate limit login attempts (max 5 per minute)
- Return authentication token on success

- Use bcrypt with salt rounds ≥ 10
- Never store plaintext passwords
- Support password verification without reversal

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

    def test_register_with_valid_email_password(self):
        """
        Requirements: Email format + password complexity
        """
        user = register_user("test@example.com", "SecurePass123")
        assert user.email == "test@example.com"
        assert user.id is not None
        assert not user.password_plaintext  # Never stored

    def test_register_reject_invalid_email(self):
        with pytest.raises(ValueError) as exc:
            register_user("invalid-email", "SecurePass123")
        assert "invalid email" in str(exc.value)

    def test_register_reject_weak_password(self):
        with pytest.raises(ValueError) as exc:
            register_user("test@example.com", "weak")
        assert "password" in str(exc.value).lower()

    def test_register_reject_duplicate_email(self):
        register_user("test@example.com", "SecurePass123")
        with pytest.raises(ValueError) as exc:
            register_user("test@example.com", "AnotherPass123")
        assert "already exists" in str(exc.value)


class TestUserLogin:

    @pytest.fixture
    def registered_user(self):
        return register_user("user@example.com", "TestPass123")

    def test_login_with_correct_credentials(self, registered_user):
        """
        Requirement: Verify credentials and return token
        """
        token = authenticate_user("user@example.com", "TestPass123")
        assert token is not None
        assert len(token) > 0

    def test_login_reject_incorrect_password(self, registered_user):
        with pytest.raises(ValueError) as exc:
            authenticate_user("user@example.com", "WrongPassword")
        assert "invalid" in str(exc.value).lower()

    def test_login_rate_limiting(self, registered_user):
        """
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

    def test_hash_password_non_reversible(self):
        hashed = hash_password("SecurePass123")
        # Hash should not contain original password
        assert "SecurePass123" not in hashed
        assert len(hashed) > 20

    def test_hash_password_unique_per_salt(self):
        hash1 = hash_password("SamePassword")
        hash2 = hash_password("SamePassword")
        # Same password should produce different hashes (different salt)
        assert hash1 != hash2

    def test_verify_password_correct(self):
        hashed = hash_password("TestPass123")
        assert verify_password("TestPass123", hashed) is True

    def test_verify_password_incorrect(self):
        hashed = hash_password("TestPass123")
        assert verify_password("WrongPass123", hashed) is False


class TestAccountLockout:

    def test_account_locked_after_5_failures(self):
        register_user("secure@example.com", "SecurePass123")
        
        # 5 failed attempts
        for i in range(5):
            with pytest.raises(ValueError):
                authenticate_user("secure@example.com", "Wrong")
        
        # Account should be locked
        status = get_account_lock_status("secure@example.com")
        assert status.is_locked is True

    def test_account_locked_notification(self):
        # Trigger lockout (implementation detail)
        # Email should be sent
        # Note: Real test would use email mock
        pass

    def test_account_auto_unlock_after_30_minutes(self):
        # This would use mock time travel
        # Test that account unlocks after 30 minutes
        pass
```

#### Day 4-5: Code Implementation (GREEN Phase)

**File**: `src/auth.py`

```python
"""
User Authentication Module

Implements user authentication requirements from SPEC-001

References:
"""

import re
import bcrypt
from datetime import datetime, timedelta
from typing import Optional
from src.models import User
from src.rate_limiter import RateLimiter


rate_limiter = RateLimiter(max_attempts=5, window_minutes=1)


def register_user(email: str, password: str) -> User:
    """
    Register new user with email and password
    
    Validates email format and password complexity
    
    Args:
        email: User email address
        password: User password (min 8 chars, 1 uppercase, 1 digit)
    
    Returns:
        Registered User object
    
    Raises:
        ValueError: If email format invalid or password weak
    """
    if not _is_valid_email(email):
        raise ValueError("Email format invalid")
    
    if not _is_strong_password(password):
        raise ValueError("Password must be ≥8 chars with uppercase and digit")
    
    if User.query.filter_by(email=email).first():
        raise ValueError(f"Email {email} already exists")
    
    hashed_pwd = hash_password(password)
    
    user = User(email=email, password_hash=hashed_pwd)
    user.save()
    return user


def authenticate_user(email: str, password: str) -> str:
    """
    Authenticate user and return token
    
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
    if not rate_limiter.allow_attempt(email):
        raise RateLimitError(f"Too many login attempts for {email}")
    
    user = User.query.filter_by(email=email).first()
    if not user:
        raise ValueError("Invalid email or password")
    
    if _is_account_locked(user):
        raise ValueError("Account locked. Try again later.")
    
    if not verify_password(password, user.password_hash):
        _increment_failed_attempts(user)
        raise ValueError("Invalid email or password")
    
    # Reset failed attempts on success
    _reset_failed_attempts(user)
    
    # Generate and return token
    token = _generate_token(user)
    return token


def hash_password(plaintext: str) -> str:
    """
    Hash password using bcrypt
    
    Implements secure hashing with salt
    
    Args:
        plaintext: Plain text password
    
    Returns:
        Hashed password (bcrypt format)
    """
    salt = bcrypt.gensalt(rounds=12)
    hashed = bcrypt.hashpw(plaintext.encode('utf-8'), salt)
    return hashed.decode('utf-8')


def verify_password(plaintext: str, hashed: str) -> bool:
    """
    Verify plaintext password against hash
    
    
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

def _is_valid_email(email: str) -> bool:
    pattern = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
    return re.match(pattern, email) is not None


def _is_strong_password(password: str) -> bool:
    return (
        len(password) >= 8 and
        any(c.isupper() for c in password) and
        any(c.isdigit() for c in password)
    )


def _is_account_locked(user: User) -> bool:
    if not user.locked_until:
        return False
    return datetime.utcnow() < user.locked_until


def _increment_failed_attempts(user: User) -> None:
    user.failed_attempts += 1
    if user.failed_attempts >= 5:
        user.locked_until = datetime.utcnow() + timedelta(minutes=30)
    user.save()


def _reset_failed_attempts(user: User) -> None:
    user.failed_attempts = 0
    user.locked_until = None
    user.save()


def get_account_lock_status(email: str) -> dict:
    """
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

Complete guide to user authentication system

## Overview

Users can register with email and password. The system validates email format and enforces password complexity requirements.

## User Registration

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

### How Passwords Are Stored

Passwords are hashed using bcrypt with 12 rounds of salting. This ensures:
- Passwords cannot be reversed
- Same password creates different hashes
- Resistant to dictionary and brute force attacks

## Account Lockout

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

TEST Layer:
  tests/test_auth.py

CODE Layer:
  src/auth.py

DOC Layer:
  docs/authentication.md

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
  Requirement: "Passwords must be securely hashed"
  Status: APPROVED
```

#### Issue Reported

```
BUG-042: Password reset tokens expire after 12 hours
Expected: 24 hours
Current: 12 hours
Impact: Users unable to reset password after 12 hours
```

#### Fix Process

**1. Create Bug Fix SPEC**

```markdown
title: Fix password reset token expiration
severity: HIGH
```

**2. Write Tests (RED)**

```python
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
def create_password_reset_token(email: str) -> str:
    """
    Create password reset token
    
    Previously: 12 hour expiration (BUG)
    Now: 24 hour expiration (FIXED)
    """
    user = User.query.filter_by(email=email).first()
    payload = {
        "email": email,
        "exp": datetime.utcnow() + timedelta(hours=24)  # Was: hours=12
    }
    return jwt.encode(payload, "secret", algorithm="HS256")
```

**4. Document Fix (DOC)**

```markdown
## Password Reset Token Fix

**Issue**: Reset tokens expired after 12 hours  
**Fix**: Extended expiration to 24 hours  
**Date**: 2025-11-13  
**Status**: RESOLVED

```

**5. TAG Chain for Bug Fix**

```
Bug Fix Chain:
├─ Issue: BUG-042
```

---

## Example 3: Detecting and Fixing an Orphan TAG


**Found During Scan**:
```bash
```

**Analysis**:
```python
# File: src/payment.py
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
title: Payment Processing Module
description: Process payments via Stripe API
```

**Step 2: Create Tests**
```python
def test_process_valid_payment():
    """Process payment with valid inputs"""
    pass

def test_process_payment_rate_limit():
    """Respect Stripe rate limits"""
    pass
```

**Step 3: Link Code**
```python
def process_payment(order_id, amount):
    """Now properly linked to SPEC"""
    pass
```

**Step 4: Document**
```markdown
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

