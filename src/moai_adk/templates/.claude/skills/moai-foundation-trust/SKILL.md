---
name: moai-foundation-trust
description: Complete TRUST 4 principles guide covering Test First, Readable, Unified, Secured
---

## Quick Reference

**TRUST 4** - core quality principles for MoAI-ADK:

- **T**est First: Write tests before implementation (≥85% coverage)
- **R**eadable: Code clarity over cleverness
- **U**nified: Consistent patterns and conventions
- **S**ecured: Security by design (OWASP Top 10 compliance)

**Core Principle**: TRUST 4 is **non-negotiable**. Every line of code must satisfy all four principles or it's not production-ready.

---

## Implementation Guide

### Principle 1: Test First (T)

**The TDD Cycle**:
```
1. RED Phase
   ├─ Write failing test
   ├─ Test defines requirement
   ├─ Code doesn't exist yet
   └─ Test fails with clear error

2. GREEN Phase
   ├─ Write minimal code to pass
   ├─ Don't over-engineer
   ├─ Focus on making test pass
   └─ Test now passes

3. REFACTOR Phase
   ├─ Improve code quality
   ├─ Extract functions/classes
   ├─ Optimize performance
   ├─ Keep tests passing
   └─ No test modification
```

**Test First Validation Rules**:
```
Rule T1: Every feature must have tests
├─ Tests must exist BEFORE implementation
├─ Test file created: days 1-2
├─ Code implementation: days 3-5
└─ No exception: 100% coverage required

Rule T2: Coverage ≥ 85%
├─ Unit test coverage >= 85%
├─ Branch coverage >= 80%
├─ Critical paths: 100%
└─ Verified via: coverage.py + codecov

Rule T3: All tests must pass
├─ CI/CD blocks merge on failed tests
├─ No skipped tests in main branch
├─ Flaky tests must be fixed
└─ Test stability: 99.9%
```

**Example: Test First in Action**:
```python
# Day 1: Write failing test (RED)
def test_password_hashing_creates_unique_hashes():
    """
    Requirement: Each password hash must be unique (different salt)
    Expected: Two calls with same password produce different hashes
    """
    hash1 = hash_password("TestPass123")
    hash2 = hash_password("TestPass123")
    assert hash1 != hash2, "Hashes must be unique"
    # OUTPUT: NameError: hash_password not defined ✓ Expected

# Days 2-3: Write minimal code (GREEN)
def hash_password(plaintext: str) -> str:
    """Hash password using bcrypt"""
    salt = bcrypt.gensalt(rounds=12)
    return bcrypt.hashpw(plaintext.encode('utf-8'), salt).decode('utf-8')
    # OUTPUT: Test passes ✓

# Days 4-5: Refactor for quality
def hash_password(plaintext: str) -> str:
    """
    Hash password using bcrypt with enterprise security settings
    
    Security:
    - Uses bcrypt algorithm (OWASP recommended)
    - Salt rounds: 12 (industry standard 2025)
    - Auto-unique salt per call
    - Non-reversible hash
    """
    BCRYPT_ROUNDS = 12
    salt = bcrypt.gensalt(rounds=BCRYPT_ROUNDS)
    hashed = bcrypt.hashpw(plaintext.encode('utf-8'), salt)
    return hashed.decode('utf-8')
    # OUTPUT: Test still passes, code is better ✓
```

### Principle 2: Readable (R)

**Readability Metrics (November 2025)**:
| Metric | Target | Tool | Threshold |
|--------|--------|------|-----------|
| Cyclomatic Complexity | ≤ 10 | pylint | 15 max |
| Function Length | ≤ 50 lines | custom | 100 line soft limit |
| Nesting Depth | ≤ 3 levels | pylint | 5 max |
| Comment Ratio | 15-20% | custom | 10-30% range |

**Readability Rules**:
```
Rule R1: Clear naming
├─ Functions: verb_noun pattern (e.g., validate_password)
├─ Variables: noun pattern (e.g., user_count, is_active)
├─ Constants: UPPER_SNAKE_CASE (e.g., MAX_LOGIN_ATTEMPTS)
├─ Classes: PascalCase (e.g., UserAuthentication)
└─ Acronyms: Spell out

Rule R2: Single responsibility principle
├─ One function = one job
├─ One class = one reason to change
├─ Extract complexity
├─ Maximum cyclomatic complexity: 10
└─ If complex: split into smaller functions

Rule R3: Documentation
├─ Function docstrings (every function)
├─ Module docstrings (at file top)
├─ Complex logic: inline comments
├─ Why, not what: explain reasoning
└─ Keep docs in sync with code
```

### Principle 3: Unified (U)

**Consistent Structure**:
```
src/
├─ auth/
│  ├─ __init__.py
│  └─ handlers.py
├─ payment/
│  ├─ __init__.py
│  └─ processors.py
└─ models/
   ├─ user.py
   └─ order.py

tests/
└─ integration/
   └─ test_payment_flow.py
```

**Unified Patterns**:
```python
# Pattern 1: Error Handling (Unified across all modules)
try:
    result = risky_operation()
except SpecificError as e:
    logger.error(f"Operation failed: {e}", extra={"user_id": user_id})
    raise ApplicationError(f"Failed to complete operation") from e
except Exception as e:
    logger.critical(f"Unexpected error: {e}")
    raise ApplicationError("Internal error") from e

# Pattern 2: Data Validation (Unified)
def validate_user_input(email: str, password: str) -> tuple[bool, str]:
    """Validate and return (is_valid, error_message)"""
    if not email or not isinstance(email, str):
        return False, "Email required"
    if len(password) < 8:
        return False, "Password minimum 8 characters"
    return True, ""
```

**Unified Validation**:
```
Rule U1: Consistent file structure
├─ All modules follow same layout
├─ Imports, docstrings, classes, functions
└─ Enforce via: pylint plugin + CI/CD

Rule U2: Consistent naming across codebase
├─ Same concept = same name (user_id everywhere)
├─ No aliases (don't use both user_id and uid)
└─ Enforce via: code review + linter config

Rule U3: Consistent error handling
├─ Same exception types for same errors
├─ Same logging approach everywhere
└─ Enforce via: custom exceptions + base classes
```

### Principle 4: Secured (S)

**OWASP Top 10 (2024 Enterprise Edition)**:
```
1. Broken Access Control
   ├─ Risk: Unauthorized feature access
   └─ Prevention: Role-based access control (RBAC)

2. Cryptographic Failures
   ├─ Risk: Data breach through weak crypto
   └─ Prevention: Use bcrypt (not MD5), TLS 1.3+

3. Injection
   ├─ Risk: SQL injection, command execution
   └─ Prevention: Parameterized queries, input validation

4. Insecure Design
   ├─ Risk: Design flaws in architecture
   └─ Prevention: Threat modeling, secure design review

5. Security Misconfiguration
   ├─ Risk: Exposed credentials, debug mode in prod
   └─ Prevention: Environment-specific config, secrets management
```

**Security Validation (STRICT Mode)**:
```
Rule S1: OWASP compliance
├─ Every OWASP risk must be addressed
├─ Design review for threat modeling
├─ Code review for vulnerabilities
└─ Enforce via: OWASP ZAP scan + code analysis

Rule S2: Authentication & Authorization
├─ MFA for privileged operations
├─ Role-based access control
├─ Rate limiting on auth endpoints
└─ Enforce via: Tests + penetration testing

Rule S3: Data Protection
├─ Encryption at rest (AES-256)
├─ Encryption in transit (TLS 1.3+)
├─ PII masking in logs
└─ Enforce via: Security audit + compliance check

Rule S4: Dependency Security
├─ Pin dependency versions
├─ Scan for known CVEs
├─ Update regularly (within 30 days)
└─ Enforce via: Dependabot + pip audit
```

---

## Advanced Patterns

### CI/CD Quality Gate Pipeline

```bash
#!/bin/bash
# .github/workflows/quality-gates.yml

echo "TRUST 4 Quality Gate Validation"

# T: Test First
pytest --cov=src --cov-report=term --cov-fail-under=85
if [ $? -ne 0 ]; then
    echo "FAILED: Test coverage < 85%"
    exit 1
fi

# R: Readable
pylint src/ --fail-under=8.0
black --check src/
if [ $? -ne 0 ]; then
    echo "FAILED: Code quality issues"
    exit 1
fi

# U: Unified
python .moai/scripts/validation/architecture_checker.py
if [ $? -ne 0 ]; then
    echo "FAILED: Inconsistent patterns"
    exit 1
fi

# S: Secured
bandit -r src/ -ll
pip audit
if [ $? -ne 0 ]; then
    echo "FAILED: Security vulnerabilities found"
    exit 1
fi

echo "SUCCESS: All quality gates passed!"
```

### TRUST 4 in Workflow

```
/alfred:1-plan "New Feature"
  ↓
  Status: DRAFT

/alfred:2-run SPEC-001
  ↓
  RED Phase: Write tests
  └─ Tests fail (no code yet)

  GREEN Phase: Write code
  ├─ Implement minimum for tests to pass
  └─ All tests pass

  REFACTOR Phase: Improve code
  ├─ Apply TRUST 4 validation
  ├─ Improve readability (R)
  ├─ Ensure unified patterns (U)
  └─ Add security checks (S)

  Quality Gates
  ├─ Test coverage: 96% >= 85% ✓
  ├─ Pylint: 9.2 >= 8.0 ✓
  ├─ Security scan: 0 vulnerabilities ✓
  └─ Status: PASS

/alfred:3-sync auto SPEC-001
  ↓
  All TRUST 4 principles validated
  ✓ Ready to merge
```

**Summary**: TRUST 4 ensures code is **correct, maintainable, secure, and production-ready**.
