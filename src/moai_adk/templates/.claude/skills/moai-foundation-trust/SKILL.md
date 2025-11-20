---
name: moai-foundation-trust
version: 4.0.0
updated: 2025-11-20
status: stable
tier: foundation
description: TRUST 4 principles - Test First, Readable, Unified, Secured
allowed-tools: [Read, Bash, WebSearch, WebFetch]
---

# TRUST 4 Principles

**Foundation of Code Quality**

> **Principles**: Test First, Readable, Unified, Secured  
> **Standards**: ≥85% Coverage, Pylint ≥8.0, OWASP Top 10

---

## Overview

TRUST 4 defines non-negotiable quality requirements for all code.

### The Four Principles

1.  **T**est First: Write tests before implementation.
2.  **R**eadable: Clarity over cleverness.
3.  **U**nified: Consistent patterns across codebase.
4.  **S**ecured: OWASP compliance by design.

---

## Principle 1: Test First (T)

### TDD Cycle

1.  **RED**: Write failing test.
2.  **GREEN**: Implement minimal code to pass.
3.  **REFACTOR**: Improve without breaking tests.

### Coverage Targets

- **Unit Tests**: 70% (fast, specific)
- **Integration Tests**: 20% (cross-component)
- **E2E Tests**: 10% (full workflow)
- **Total**: ≥85%

### Example

```python
# 1. RED: Write test (fails)
def test_password_hash_unique():
    hash1 = hash_password("TestPass123")
    hash2 = hash_password("TestPass123")
    assert hash1 != hash2

# 2. GREEN: Implement
def hash_password(plaintext: str) -> str:
    import bcrypt
    salt = bcrypt.gensalt(rounds=12)
    return bcrypt.hashpw(plaintext.encode(), salt).decode()

# 3. REFACTOR: Add security docs
```

---

## Principle 2: Readable (R)

### Metrics

- **Cyclomatic Complexity**: ≤10
- **Function Length**: ≤50 lines
- **Nesting Depth**: ≤3 levels

### Rules

- Clear naming (`validate_password` not `vp`)
- Single responsibility per function
- Docstrings for all functions
- Consistent style (PEP 8, Black, Prettier)

### Example

```python
# Bad
def f(x, y):
    return sum([i*x for i in y if i]) / len(y) if y else 0

# Good
def calculate_weighted_average(weight: float, values: List[float]) -> float:
    """Calculate weighted average, excluding None values."""
    valid = [v for v in values if v is not None]
    if not valid:
        return 0.0
    return sum(v * weight for v in valid) / len(valid)
```

---

## Principle 3: Unified (U)

### Consistency Rules

- Same file structure across modules
- Same naming for same concepts (`user_id` everywhere, not `uid`)
- Same error handling patterns
- Same logging approach

### Example Pattern

```python
# Unified error handling
try:
    result = process_payment(order)
except PaymentError as e:
    logger.error(f"Payment failed: {e}", extra={"order_id": order.id})
    raise ApplicationError("Payment processing failed") from e
```

---

## Principle 4: Secured (S)

### OWASP Top 10 (2024)

1.  **Broken Access Control**: Implement RBAC
2.  **Cryptographic Failures**: Use bcrypt (not MD5), TLS 1.3+
3.  **Injection**: Parameterized queries
4.  **Insecure Design**: Threat modeling
5.  **Security Misconfiguration**: Secrets management
6.  **Vulnerable Components**: Regular updates (pip audit)
7.  **Auth Failures**: MFA, rate limiting
8.  **Data Integrity**: Code signing
9.  **Logging Failures**: Comprehensive logs
10. **SSRF**: Input validation

### Example

```python
def hash_password(plaintext: str) -> str:
    """
    OWASP A02:2021 compliant password hashing.
    Uses bcrypt with 12 rounds (2025 standard).
    """
    import bcrypt
    if not plaintext:
        raise ValueError("Password required")

    salt = bcrypt.gensalt(rounds=12)
    return bcrypt.hashpw(plaintext.encode(), salt).decode()
```

---

## Validation Checklist

- [ ] **Test**: Coverage ≥85%?
- [ ] **Readable**: Pylint ≥8.0?
- [ ] **Unified**: Patterns consistent?
- [ ] **Secured**: OWASP scan passed (bandit, pip audit)?

---

## CI/CD Quality Gates

```bash
# Automated quality gate
pytest --cov=src --cov-fail-under=85  # T: Test
pylint src/ --fail-under=8.0           # R: Readable
bandit -r src/ -ll                     # S: Secured
```

---

## Related Skills

- `moai-core-rules`: Core development guidelines
- `moai-security-devsecops`: Security testing
- `moai-project-documentation`: Documentation standards

---

**Last Updated**: 2025-11-20
