# Traceability System Complete Guide

Building complete SPEC-TEST-CODE-DOC traceability through TAG chains.

## Traceability Principles

All features must include all 4 layers:

```
1. SPEC    (Requirements) ← Written by Alfred
2. @TEST   (Tests)         ← Written by tdd-implementer
3. @CODE   (Implementation) ← Written by tdd-implementer
4. @DOC    (Documentation) ← Written by doc-syncer
```

## TAG Chain Diagram

```
SPEC-001 ────┐
             │
             ├─→ @TEST:SPEC-001:*
             │
             ├─→ @CODE:SPEC-001:*
             │
             ├─→ @DOC:SPEC-001:*
             │
             └─→ All cross-referenced
```

## Complete Traceability Example

### Step 1: SPEC Writing

```markdown
# SPEC-001: User Authentication

## Requirements
- Login feature
- Registration
- Password reset
```

### Step 2: @TEST Writing

```python
# tests/test_auth.py

# @TEST:SPEC-001:login_success
def test_login_success():
    user = login("user@example.com", "pass123")
    assert user is not None

# @TEST:SPEC-001:register_success
def test_register_success():
    user = register("new@example.com", "pass123")
    assert user.email == "new@example.com"

# @TEST:SPEC-001:password_reset_success
def test_password_reset():
    token = request_reset("user@example.com")
    assert token is not None
```

### Step 3: @CODE Writing

```python
# src/auth.py

# @CODE:SPEC-001:login
def login(email, password):
    """Tested by @TEST:SPEC-001:login_success"""
    user = User.query.filter_by(email=email).first()
    if user and user.verify_password(password):
        return user
    return None

# @CODE:SPEC-001:register
def register(email, password):
    """Tested by @TEST:SPEC-001:register_success"""
    if User.query.filter_by(email=email).first():
        return None
    user = User(email=email)
    user.set_password(password)
    db.session.add(user)
    db.session.commit()
    return user

# @CODE:SPEC-001:password_reset
def request_password_reset(email):
    """Tested by @TEST:SPEC-001:password_reset_success"""
    user = User.query.filter_by(email=email).first()
    if user:
        token = generate_reset_token(user)
        return token
    return None
```

### Step 4: @DOC Writing

````markdown
# User Authentication API @DOC:SPEC-001:api_docs

This document explains the implementation of @SPEC-001.

## Login Endpoint

**Implementation**: @CODE:SPEC-001:login
**Test**: @TEST:SPEC-001:login_success

```bash
POST /api/auth/login
````

## Registration Endpoint

**Implementation**: @CODE:SPEC-001:register **Test**: @TEST:SPEC-001:register_success

```bash
POST /api/auth/register
```

## Password Reset

**Implementation**: @CODE:SPEC-001:password_reset **Test**: @TEST:SPEC-001:password_reset_success

```bash
POST /api/auth/reset-password
```

````

## Traceability Validation

### Automatic Validation

```bash
# Automatically executed in /alfred:3-sync
moai-adk tag-agent

1. SPEC detected: SPEC-001 ✓
2. @TEST verified: 3 ✓
3. @CODE verified: 3 ✓
4. @DOC verified: 1 ✓
5. Chain complete: Done ✅

Result: SPEC-001 traceability 100%
```

### Manual Validation

```bash
# Check specific SPEC status
moai-adk status --spec SPEC-001

# Result:
SPEC-001: User Authentication
├─ @TEST:SPEC-001:* (3)
├─ @CODE:SPEC-001:* (3)
├─ @DOC:SPEC-001:* (1)
└─ ✅ Completeness: 100%
```

## Traceability Problem Diagnosis

### Problem 1: Orphaned TAG

```python
# Problem: TAG without SPEC
@CODE:SPEC-999:orphan_function
def some_function():
    pass

# Solution: Create SPEC-999 or remove TAG
```

### Problem 2: Incomplete Chain

```python
# Problem: TEST and CODE exist but no DOC
@TEST:SPEC-001:test
def test_feature():
    pass

@CODE:SPEC-001:feature
def feature():
    pass

# Solution: Create @DOC:SPEC-001
```

### Problem 3: TAG Duplication

```python
# Problem: Same TAG in multiple files
# file1.py:
@CODE:SPEC-001:register

# file2.py:
@CODE:SPEC-001:register  # Duplicate!

# Solution: Rename tag or consolidate
```

## TAG Deduplication Process

```bash
# 1. Scan for duplicates
/alfred:tag-dedup --scan-only

# 2. Create plan
/alfred:tag-dedup --dry-run

# 3. Apply after backup
/alfred:tag-dedup --apply --backup

# Result: All TAGs unique, chain complete
```

## Traceability Metrics

### Calculation Method

```
Traceability = (TEST + CODE + DOC) / (SPEC × 3) × 100%

Example:
- SPEC: 5
- @TEST: 15 (5 × 3)
- @CODE: 15 (5 × 3)
- @DOC: 5 (5 × 1)

Traceability = (15 + 15 + 5) / (5 × 3) × 100% = 100%
```

### Goals

| Level     | Traceability | Status       |
| --------- | ------------ | ------------ |
| Excellent | 100%         | ✅ Deployable |
| Good      | 90%+         | ⚠️ Review needed |
| Fair      | 70%+         | 🚨 Improvement needed |
| Poor      | <70%         | :x: Not deployable |

______________________________________________________________________

**Next**: [TAG Types](types.md) or [TAG Overview](index.md)



