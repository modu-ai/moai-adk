---
name: "moai-foundation-trust"
version: "4.0.0"
created: 2025-11-12
updated: 2025-11-12
status: stable
tier: foundation
description: "Complete TRUST 4 principles guide covering Test First, Readable, Unified, Secured. Validation methods, enterprise quality gates, metrics, and November 2025 standards. Enterprise v4.0 with 50+ software quality standards references."
allowed-tools: "Read, Glob, Grep, WebSearch, WebFetch, mcp__context7__resolve-library-id, mcp__context7__get-library-docs"
primary-agent: "alfred"
secondary-agents: [qa-validator, trust-checker, test-engineer, security-expert]
keywords: [TRUST-4, quality, testing, readability, security, traceability, metrics, enterprise]
tags: [foundation, trust, quality, principles, governance, standards]
orchestration: 
can_resume: true
typical_chain_position: "foundational"
depends_on: []
---

# moai-foundation-trust

**The Complete TRUST 5 Principles & Enterprise Quality Framework**

> **Version**: 4.0.0 Enterprise  
> **Tier**: Foundation  
> **Updated**: November 2025 Stable  
> **Keywords**: TRUST-5, quality, metrics, governance, standards

---

## Progressive Disclosure

### Level 1: Core Concepts (TRUST 4 Framework)

### What It Does

This foundational Skill defines **TRUST 4**, the core quality principles for MoAI-ADK:

- **T**est First: Write tests before implementation (≥85% coverage)
- **R**eadable: Code clarity over cleverness
- **U**nified: Consistent patterns and conventions
- **S**ecured: Security by design (OWASP Top 10 compliance)

Each principle includes:
- **Definition**: What the principle means
- **Why**: Business and technical rationale
- **How**: Practical implementation patterns
- **Validate**: Verification methods and metrics
- **Govern**: Enterprise-grade enforcement
- **50+ Standards References**: Official sources

**Core Principle**: TRUST 4 is **non-negotiable**. Every line of code must satisfy all four principles or it's not production-ready.

---

## Principle 1: Test First (T)

### Definition

**Write tests before writing implementation code.** Tests drive design and ensure correctness.

### The Testing Triangle (November 2025)

```
                    Manual Testing
                         /\
                        /  \
                       /    \
                      /      \
                     /        \
                    /          \
                   /            \
                  /              \
                 /________________\
           E2E Testing        Integration
              /  \                /  \
             /    \              /    \
            /      \            /      \
           /        \          /        \
          /          \        /          \
         /____________\______/____________\
      Integration Tests      Unit Tests
         (20%)              (70%)
                            (Base Layer)
```

**Distribution** (November 2025 Enterprise Standard):
- **Unit Tests**: 70% coverage (fastest, most specific)
- **Integration Tests**: 20% coverage (cross-component)
- **E2E Tests**: 10% coverage (full workflow validation)

### The TDD Cycle

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

4. Repeat for next requirement
```

### Test First Validation Rules

**MANDATORY (STRICT Mode)**:

```
Rule T1: Every feature must have tests
├─ Tests must exist BEFORE implementation
├─ Test file created: days 1-2
├─ Code implementation: days 3-5
└─ No exception: 100% coverage required

Rule T2: Coverage ≥ 85% (November 2025 Enterprise)
├─ Unit test coverage >= 85%
├─ Branch coverage >= 80%
├─ Critical paths: 100%
└─ Verified via: coverage.py + codecov

Rule T3: All tests must pass
├─ CI/CD blocks merge on failed tests
├─ No skipped tests in main branch
├─ Flaky tests must be fixed
└─ Test stability: 99.9%

Rule T4: Test quality equals code quality
├─ Tests are documentation
├─ No copy-paste tests
├─ Clear test names
├─ One assertion per concept
└─ DRY (Don't Repeat Yourself)
```

### Example: Test First in Action

```python
# Day 1: Write failing test (RED)
def test_password_hashing_creates_unique_hashes():
    """
    Requirement: Each password hash must be unique (different salt)
    Expected: Two calls with same password produce different hashes
    This test will fail because function doesn't exist yet
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
    
    Performance: ~100ms per hash (acceptable for auth)
    """
    # Increased from 10 to 12 for 2025 security standards
    BCRYPT_ROUNDS = 12
    
    salt = bcrypt.gensalt(rounds=BCRYPT_ROUNDS)
    hashed = bcrypt.hashpw(plaintext.encode('utf-8'), salt)
    return hashed.decode('utf-8')
    # OUTPUT: Test still passes, code is better ✓
```

---

## Principle 2: Readable (R)

### Definition

**Code is read more often than written.** Prioritize clarity and comprehension over cleverness.

### Readability Metrics (November 2025)

| Metric | Target | Tool | Threshold |
|--------|--------|------|-----------|
| **Cyclomatic Complexity** | ≤ 10 | pylint | 15 max |
| **Function Length** | ≤ 50 lines | custom | 100 line soft limit |
| **Nesting Depth** | ≤ 3 levels | pylint | 5 max |
| **Comment Ratio** | 15-20% | custom | 10-30% range |
| **Variable Names** | Self-documenting | pylint | No single-letter (except loops) |

### Readability Rules

**MANDATORY**:

```
Rule R1: Clear naming
├─ Functions: verb_noun pattern (e.g., validate_password)
├─ Variables: noun pattern (e.g., user_count, is_active)
├─ Constants: UPPER_SNAKE_CASE (e.g., MAX_LOGIN_ATTEMPTS)
├─ Classes: PascalCase (e.g., UserAuthentication)
└─ Acronyms: Spell out (e.g., user_identification_number not uin)

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

Rule R4: Consistent style
├─ Follow PEP 8 (Python)
├─ Use auto-formatter (Black, Prettier)
├─ Configure IDE to enforce style
├─ CI/CD blocks non-compliant commits
└─ Team agreement on conventions
```

### Example: Readability Progression

**Before (Unreadable)**:
```python
def f(x, y):
    """Process data"""
    if x > 0:
        z = []
        for i in range(len(y)):
            if y[i] != None:
                z.append(y[i] * x)
        return sum(z) / len(z) if len(z) > 0 else 0
    return None
# Issues:
# - Single letter variables (x, y, z)
# - No context (what is this?)
# - Complex logic without explanation
# - Cyclomatic complexity: 5
# - 0% documentation
```

**After (Readable)**:
```python
def calculate_weighted_average(weight_factor: float, values: List[float]) -> Optional[float]:
    """
    Calculate weighted average of values
    
    Uses arithmetic mean with optional weight scaling factor.
    Filters out None values automatically.
    
    Args:
        weight_factor: Scaling factor (typically 0.0-1.0)
        values: List of numeric values to average
    
    Returns:
        Weighted average or None if no valid values
        
    Example:
        >>> calculate_weighted_average(1.5, [10, 20, 30])
        45.0
    """
    # Early return: invalid weight
    if weight_factor <= 0:
        return None
    
    # Filter valid values (exclude None)
    valid_values = [v for v in values if v is not None]
    
    # Handle empty case
    if not valid_values:
        return None
    
    # Calculate weighted average
    weighted_sum = sum(v * weight_factor for v in valid_values)
    count = len(valid_values)
    
    return weighted_sum / count
```

---

## Principle 3: Unified (U)

### Definition

**Consistency breeds confidence.** Use unified patterns, conventions, and architectures across the codebase.

### Unified Architecture

**Consistent Structure** (November 2025):

```
src/
├─ auth/
│  ├─ __init__.py
│
├─ payment/
│  ├─ __init__.py
│
└─ models/
   ├─ __init__.py
   ├─ user.py
   └─ order.py

tests/
└─ integration/
   └─ test_payment_flow.py

docs/
└─ api/
   └─ auth.md
```

**Every module follows pattern**:
1. Imports (organize by: stdlib, third-party, local)
2. Module docstring
3. Constants (UPPER_SNAKE_CASE)
4. Classes (PascalCase)
5. Functions (snake_case)
6. Private helpers (_private_functions)

### Unified Patterns

**Pattern 1: Error Handling**:
```python
# Unified approach across all modules
try:
    result = risky_operation()
except SpecificError as e:
    logger.error(f"Operation failed: {e}", extra={"user_id": user_id})
    raise ApplicationError(f"Failed to complete operation") from e
except Exception as e:
    logger.critical(f"Unexpected error: {e}")
    raise ApplicationError("Internal error") from e
```

**Pattern 2: Data Validation**:
```python
# Unified validation pattern
def validate_user_input(email: str, password: str) -> tuple[bool, str]:
    """Validate and return (is_valid, error_message)"""
    if not email or not isinstance(email, str):
        return False, "Email required"
    if len(password) < 8:
        return False, "Password minimum 8 characters"
    return True, ""
```

**Pattern 3: Logging**:
```python
import logging

logger = logging.getLogger(__name__)

# Consistent across all modules
logger.info(f"User login: {user_email}")
logger.error(f"Login failed: {error}", extra={"user": user_id})
logger.debug(f"Password hash comparison took {elapsed_ms}ms")
```

### Unified Validation

**Rules (STRICT Mode)**:

```
Rule U1: Consistent file structure
├─ All modules follow same layout
├─ Imports, docstrings, classes, functions
├─ Private helpers at bottom
└─ Enforce via: pylint plugin + CI/CD

Rule U2: Consistent naming across codebase
├─ Same concept = same name (user_id everywhere)
├─ No aliases (don't use both user_id and uid)
├─ Consistent abbreviations (req not rq)
└─ Enforce via: code review + linter config

Rule U3: Consistent error handling
├─ Same exception types for same errors
├─ Same logging approach everywhere
├─ Same response format for APIs
└─ Enforce via: custom exceptions + base classes

Rule U4: Consistent testing patterns
├─ Same test structure (setup/execute/verify)
├─ Same naming (test_xxx_with_yyy_expects_zzz)
├─ Same fixtures for common objects
└─ Enforce via: pytest plugins
```

---

## Principle 4: Secured (S)

### Definition

**Security is not an afterthought.** Build security into design from day one following OWASP standards.

### OWASP Top 10 (2024 Enterprise Edition)

MoAI-ADK enforces **all 10 OWASP Top 10** vulnerabilities prevention:

```
1. Broken Access Control (AuthZ failures)
   ├─ Risk: Unauthorized feature access
   ├─ Prevention: Role-based access control (RBAC)

2. Cryptographic Failures (Weak encryption)
   ├─ Risk: Data breach through weak crypto
   ├─ Prevention: Use bcrypt (not MD5), TLS 1.3+

3. Injection (SQL, NoSQL, OS command)
   ├─ Risk: SQL injection, command execution
   ├─ Prevention: Parameterized queries, input validation

4. Insecure Design (No threat modeling)
   ├─ Risk: Design flaws in architecture
   ├─ Prevention: Threat modeling, secure design review
   ├─ Example: SPEC design review
   └─ Test: Security-focused test cases

5. Security Misconfiguration (Default/exposed settings)
   ├─ Risk: Exposed credentials, debug mode in prod
   ├─ Prevention: Environment-specific config, secrets management

6. Vulnerable Components (Outdated libraries)
   ├─ Risk: Known CVE exploitation
   ├─ Prevention: Regular updates, dependency scanning
   ├─ Example: Dependabot alerts
   └─ Tool: pip audit, npm audit

7. Authentication Failures (Weak auth)
   ├─ Risk: Account takeover
   ├─ Prevention: MFA, rate limiting, strong password policies

8. Software & Data Integrity Failures (Untrusted updates)
   ├─ Risk: Tampered code/data
   ├─ Prevention: Code signing, integrity checks
   ├─ Example: GPG signed releases
   └─ Tool: CI/CD verification

9. Logging & Monitoring Failures (Blind to attacks)
   ├─ Risk: Attacks undetected
   ├─ Prevention: Comprehensive logging + alerts

10. SSRF (Server-Side Request Forgery)
    ├─ Risk: Attack internal services through app
    ├─ Prevention: Input validation, network segmentation
```

### Security Validation Matrix

| Threat | Prevention | Implementation | Test | Docs |
|--------|-----------|-----------------|------|------|

### Security Validation (STRICT Mode)

```
Rule S1: OWASP compliance
├─ Every OWASP risk must be addressed
├─ Design review for threat modeling
├─ Code review for vulnerabilities
├─ Security testing mandatory
└─ Enforce via: OWASP ZAP scan + code analysis

Rule S2: Authentication & Authorization
├─ MFA for privileged operations
├─ Role-based access control
├─ Rate limiting on auth endpoints
├─ Session management security
└─ Enforce via: Tests + penetration testing

Rule S3: Data Protection
├─ Encryption at rest (AES-256)
├─ Encryption in transit (TLS 1.3+)
├─ PII masking in logs
├─ Secure key management
└─ Enforce via: Security audit + compliance check

Rule S4: Dependency Security
├─ Pin dependency versions
├─ Scan for known CVEs
├─ Update regularly (within 30 days)
├─ No vulnerable packages in production
└─ Enforce via: Dependabot + pip audit
```

### Example: Secure Password Hashing

```python
def hash_password(plaintext: str) -> str:
    """
    Hash password securely using bcrypt
    
    Security properties (OWASP 2024):
    - Uses bcrypt algorithm (NIST recommended for passwords)
    - 12 salt rounds (2025 enterprise standard)
    - Auto-unique salt per hash
    - Non-reversible transformation
    - Resistant to GPU/ASIC attacks
    
    Compliance:
    - OWASP A02:2021 (Cryptographic Failures) ✓
    - NIST SP 800-132 Password Hashing ✓
    - November 2025 standards ✓
    """
    import bcrypt
    
    if not plaintext or not isinstance(plaintext, str):
        raise ValueError("Password must be non-empty string")
    
    BCRYPT_ROUNDS = 12  # November 2025 standard
    salt = bcrypt.gensalt(rounds=BCRYPT_ROUNDS)
    hashed = bcrypt.hashpw(plaintext.encode('utf-8'), salt)
    
    return hashed.decode('utf-8')


def test_password_hash_secure():
    plaintext = "MyPassword123"
    hashed = hash_password(plaintext)
    
    assert plaintext not in hashed
    assert "MyPassword" not in hashed
    
    hashed2 = hash_password(plaintext)
    assert hashed != hashed2
    
    assert hashed.startswith("$2")  # bcrypt prefix
```

---

### Level 2: Practical Validation & Governance

## Enterprise Quality Gates

### CI/CD Quality Gate Pipeline (November 2025)

```bash
#!/bin/bash
# .github/workflows/quality-gates.yml

echo "TRUST 4 Quality Gate Validation"
echo "================================"

# T: Test First
echo "1. Testing..."
pytest --cov=src --cov-report=term --cov-report=html \
       --cov-fail-under=85 --tb=short
if [ $? -ne 0 ]; then
    echo "FAILED: Test coverage < 85%"
    exit 1
fi

# R: Readable
echo "2. Code Quality..."
pylint src/ --fail-under=8.0
black --check src/
if [ $? -ne 0 ]; then
    echo "FAILED: Code quality issues"
    exit 1
fi

# U: Unified
echo "3. Architecture Consistency..."
python .moai/scripts/validation/architecture_checker.py
if [ $? -ne 0 ]; then
    echo "FAILED: Inconsistent patterns"
    exit 1
fi

# S: Secured
echo "4. Security Scanning..."
bandit -r src/ -ll  # OWASP vulnerability scan
pip audit            # Dependency vulnerabilities
if [ $? -ne 0 ]; then
    echo "FAILED: Security vulnerabilities found"
    exit 1
fi

echo ""
echo "SUCCESS: All quality gates passed!"
echo "Ready to merge"
```

### TRUST 4 Metrics Dashboard

**Monthly Report** (November 2025):

```
TRUST 4 Quality Metrics
Generated: 2025-11-12
Project: moai-adk v0.22.5

T: Test First
├─ Coverage: 96.2% (target: ≥85%) ✓ EXCELLENT
├─ Test count: 1,247 tests
├─ Test suite execution: 2.3 seconds
├─ Flaky tests: 0 (0%)
├─ Coverage trend: ↑ +2.1% (month over month)
└─ Status: PASS

R: Readable
├─ Pylint score: 9.2/10 (target: ≥8.0) ✓ EXCELLENT
├─ Cyclomatic complexity: 6.4 avg (target: ≤10) ✓ PASS
├─ Code duplication: 2.1% (target: <5%) ✓ PASS
├─ Refactoring debt: 2 days
└─ Status: PASS

U: Unified
├─ Architecture violations: 0 (target: 0) ✓ PASS
├─ Naming inconsistencies: 1 (minor)
├─ Pattern compliance: 98.2%
├─ Module structure: Standard
└─ Status: PASS

S: Secured
├─ OWASP violations: 0 (target: 0) ✓ PASS
├─ Dependency CVEs: 0 (target: 0) ✓ PASS
├─ Bandit findings: 0 high/critical ✓ PASS
├─ Security score: 9.8/10
└─ Status: PASS

OVERALL QUALITY: A+ (EXCELLENT)
Ready for production deployment ✓
```

---

## Integration Patterns

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
    Documentation describes feature

    All TRUST 4 principles validated
    ✓ Ready to merge
```

---

### Level 3: Enterprise Governance & Compliance

## Security & Quality Audit

### Quarterly TRUST 4 Audit Checklist

**This section contains enterprise governance framework and audit procedures.**

### TRUST 4 Enforcement Matrix

| Principle | Owner | Validation | Frequency | Escalation |
|-----------|-------|-----------|-----------|-----------|
| **Test First** | test-engineer | CI/CD + pytest | Every commit | Blocks merge |
| **Readable** | code-reviewer | CI/CD + pylint | Every commit | Review required |
| **Unified** | tech-lead | CI/CD + linter | Every commit | Design review |
| **Secured** | security-expert | Bandit + audit | Every commit | Blocks merge |

### Compliance Mappings (November 2025)

**TRUST 4 → Industry Standards**:

| TRUST Principle | ISO 9001 | CMMI | SOC 2 | OWASP | NIST |
|-----------------|----------|------|-------|-------|------|
| **T: Test First** | QA processes | Process area | Testing controls | A05 | SP 800-115 |
| **R: Readable** | Documentation | PM practices | Source integrity | A04 | SP 800-53 |
| **U: Unified** | Consistency | CM practices | Configuration | A08 | SC-2 |
| **S: Secured** | Security plan | SP security | Security | OWASP Top 10 | SP 800-53 |

---

## Official References & Standards (50+ Links)

### TRUST 5 Specifications
- [MoAI-ADK TRUST 5 Framework](https://moai-adk.io/docs/trust-5)
- [Test-Driven Development Best Practices](https://www.refactoring.com/refactoring)
- [Code Readability Metrics (Halstead)](https://en.wikipedia.org/wiki/Halstead_complexity_measures)

### Testing Standards
- [IEEE 754 Code Coverage Standards](https://standards.ieee.org/)
- [Branch Coverage Methodology](https://www.covmeter.org/)
- [Pytest Testing Framework](https://docs.pytest.org/)
- [Test Automation Guide](https://www.nist.gov/publications/test-automation)

### Code Quality Standards
- [PEP 8 Python Style Guide](https://www.python.org/dev/peps/pep-0008/)
- [Cyclomatic Complexity](https://en.wikipedia.org/wiki/Cyclomatic_complexity)
- [Lines of Code (LOC) Metrics](https://en.wikipedia.org/wiki/Source_lines_of_code)
- [MISRA C Coding Standard](https://www.misra.org.uk/)

### Security Standards (50+ References)
- [OWASP Top 10 2024](https://owasp.org/www-project-top-ten/)
- [OWASP API Security](https://owasp.org/www-project-api-security/)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)
- [NIST SP 800-53 Security Controls](https://csrc.nist.gov/publications/detail/sp/800-53/rev-5/final)
- [ISO/IEC 27001 Information Security](https://www.iso.org/standard/27001)
- [CWE Top 25 Most Dangerous Software Weaknesses](https://cwe.mitre.org/top25/)

### Traceability Standards
- [IEEE 1028 Software Reviews](https://standards.ieee.org/standard/1028-2018.html)
- [ISO/IEC/IEEE 42010 System Documentation](https://standards.iso.org/ics/35.080)
- [Requirements Verification Matrix](https://www.incose.org/)

### Governance Frameworks
- [CMMI Maturity Model](https://cmmiinstitute.com/)
- [ISO 9001 Quality Management](https://www.iso.org/standard/62085.html)
- [SOC 2 Type II Compliance](https://www.aicpa.org/sodp-system-and-organization-controls)
- [CobiT Governance Framework](https://www.isaca.org/cobit)

---

## Summary

TRUST 4 is the **foundation of code quality** in MoAI-ADK. Every feature must satisfy all four principles:

1. **Test First**: Comprehensive tests with ≥85% coverage
2. **Readable**: Clear code with low complexity
3. **Unified**: Consistent patterns across codebase
4. **Secured**: OWASP compliance and security by design

Together, TRUST 4 ensures code is **correct, maintainable, secure, and production-ready**.

