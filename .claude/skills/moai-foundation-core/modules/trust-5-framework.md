# TRUST 5 Framework - Quality Assurance System

Purpose: Automated quality gates ensuring code quality, security, maintainability, and traceability through five core principles.

Version: 1.0.0 (Updated from TRUST 4)
Last Updated: 2025-11-25

---

## Quick Reference (30 seconds)

TRUST 5 is MoAI-ADK's comprehensive quality assurance framework enforcing five pillars:

1. Test-first(T) - ≥85% coverage, RED-GREEN-REFACTOR cycle
2. Readable(R) - Clear naming, ≤10 cyclomatic complexity
3. Unified(U) - Consistent patterns, architecture compliance
4. Secured(S) - OWASP Top 10 compliance, security validation
5. Trackable(T) - Clear commits, requirement traceability

Integration Points:
- Pre-commit hooks → Automated validation
- CI/CD pipelines → Quality gate enforcement
- quality-gate agent → TRUST 5 validation
- /moai:2-run → Enforces ≥85% coverage

Quick Validation:
```python
validations = [
 test_coverage >= 85, # T
 complexity <= 10, # R
 consistency > 90, # U
 security_score == 100,# S
 has_clear_commits # T
]
```

---

## Implementation Guide (5 minutes)

### Principle 1: Test-First (T)

RED-GREEN-REFACTOR Cycle:

```
RED Phase: Write failing test
 Test defines requirement
 Code doesn't exist yet
 Test fails as expected 

GREEN Phase: Write minimal code
 Simplest code to pass test
 Focus on making test pass
 Test now passes 

REFACTOR Phase: Improve quality
 Extract functions/classes
 Optimize performance
 Add documentation
 Keep tests passing 
```

Test Coverage Requirements:

| Level    | Coverage | Action                      |
| -------- | -------- | --------------------------- |
| Critical | ≥85%     | Required for merge          |
| Warning  | 70-84%   | Review required             |
| Failing  | <70%     | Block merge, generate tests |

Implementation Pattern:

```python
# RED: Write failing test first
def test_calculate_total_price_with_tax():
 item = ShoppingItem(name="Widget", price=10.00)
 total = calculate_total_with_tax(item, tax_rate=0.10)
 assert total == 11.00 # Fails - function doesn't exist

# GREEN: Minimal implementation
def calculate_total_with_tax(item, tax_rate):
 return item.price * (1 + tax_rate)

# REFACTOR: Improve code quality
def calculate_total_with_tax(item: ShoppingItem, tax_rate: float) -> float:
 """Calculate total price including tax.
 
 Args:
 item: Shopping item with price
 tax_rate: Tax rate as decimal (0.10 = 10%)
 
 Returns:
 Total price including tax
 
 Raises:
 ValueError: If tax_rate not between 0 and 1
 
 Example:
 >>> item = ShoppingItem("Widget", 10.00)
 >>> calculate_total_with_tax(item, 0.10)
 11.0
 """
 if not 0 <= tax_rate <= 1:
 raise ValueError("Tax rate must be between 0 and 1")
 
 return item.price * (1 + tax_rate)
```

Validation Commands:
```bash
# Run tests with coverage
pytest --cov=src --cov-report=html --cov-fail-under=85

# Generate coverage report
coverage report -m

# Show missing lines
coverage html && open htmlcov/index.html
```

---

### Principle 2: Readable (R)

Readability Metrics:

| Metric                | Target    | Tool   | Max Threshold |
| --------------------- | --------- | ------ | ------------- |
| Cyclomatic Complexity | ≤10       | pylint | 15            |
| Function Length       | ≤50 lines | custom | 100           |
| Nesting Depth         | ≤3 levels | pylint | 5             |
| Comment Ratio         | 15-20%    | custom | Min 10%       |
| Type Hint Coverage    | 100%      | mypy   | 90%           |

Readability Checklist:

```
 Clear function/variable names (noun_verb pattern)
 Single responsibility principle
 Type hints on all parameters and returns
 Docstrings with examples (Google style)
 No magic numbers (use named constants)
 DRY principle applied (no code duplication)
 SOLID principles followed
```

Example - Bad vs Good:

```python
# BAD: Unreadable, no types, magic numbers
def calc(x, y):
 if x > 0:
 if y > 0:
 if x + y < 100:
 return x * 1.1 + y * 0.9
 return 0

# GOOD: Readable, typed, constants
TAX_RATE = 0.10
DISCOUNT_RATE = 0.10
MAX_TOTAL = 100.00

def calculate_order_total(
 base_amount: float,
 discount_amount: float
) -> float:
 """Calculate order total with tax and discount.
 
 Args:
 base_amount: Base order amount before tax
 discount_amount: Discount amount to apply
 
 Returns:
 Final order total with tax applied
 
 Raises:
 ValueError: If amounts are negative or exceed max
 """
 if base_amount < 0 or discount_amount < 0:
 raise ValueError("Amounts must be non-negative")
 
 subtotal = base_amount - discount_amount
 
 if subtotal > MAX_TOTAL:
 raise ValueError(f"Total exceeds maximum {MAX_TOTAL}")
 
 return subtotal * (1 + TAX_RATE)
```

Validation Commands:
```bash
# Pylint complexity check
pylint src/ --fail-under=8.0

# Black format check
black --check src/

# MyPy type check
mypy src/ --strict
```

---

### Principle 3: Unified (U)

Consistency Requirements:

```
Architecture Consistency:
 Same pattern across all modules
 Same error handling approach
 Same logging strategy
 Same naming conventions

Testing Consistency:
 Same test structure (Arrange-Act-Assert)
 Same fixtures/factories
 Same assertion patterns
 Same mock strategies

Documentation Consistency:
 Same docstring format (Google style)
 Same README structure
 Same API documentation
 Same changelog format (conventional commits)
```

Pattern Enforcement:

```python
# Standard error handling pattern
class DomainError(Exception):
 """Base error for domain-specific errors."""
 pass

class ValidationError(DomainError):
 """Validation failed."""
 pass

def process_data(data: dict) -> Result:
 """Standard processing pattern."""
 try:
 # 1. Validate input
 validated = validate_input(data)
 
 # 2. Process
 result = perform_processing(validated)
 
 # 3. Return result
 return Result(success=True, data=result)
 
 except ValidationError as e:
 logger.error(f"Validation failed: {e}")
 raise
 except Exception as e:
 logger.exception(f"Processing failed: {e}")
 raise DomainError(f"Processing failed: {e}") from e
```

Validation Tools:
```bash
# Check architecture compliance
python .moai/scripts/validate_architecture.py

# Verify naming conventions
grep -r "def [A-Z]" src/ && echo "Found camelCase functions!"

# Check consistent imports
isort --check-only src/
```

---

### Principle 4: Secured (S)

OWASP Top 10 (2024) Compliance:

| Risk                      | Mitigation                | Validation             |
| ------------------------- | ------------------------- | ---------------------- |
| Broken Access Control     | RBAC, permission checks   | security-expert review |
| Cryptographic Failures    | bcrypt, proper encryption | Bandit scan            |
| Injection                 | Parameterized queries     | SQLMap test            |
| Insecure Design           | Threat modeling           | Architecture review    |
| Security Misconfiguration | Environment variables     | Config audit           |
| Vulnerable Components     | Dependency scanning       | pip-audit, safety      |
| Authentication Failures   | MFA, secure sessions      | Penetration test       |
| Data Integrity            | Checksums, signatures     | Integrity validation   |
| Logging Failures          | Comprehensive logging     | Log analysis           |
| SSRF                      | URL validation            | Security test          |

Security Patterns:

```python
# 1. Broken Access Control - RBAC implementation
from functools import wraps

def require_permission(permission: str):
 """Decorator to enforce permission checks."""
 def decorator(func):
 @wraps(func)
 def wrapper(user: User, *args, kwargs):
 if not user.has_permission(permission):
 raise UnauthorizedError(
 f"User lacks permission: {permission}"
 )
 return func(user, *args, kwargs)
 return wrapper
 return decorator

@require_permission("user:update")
def update_user_profile(user: User, profile_data: dict) -> UserProfile:
 """Update user profile (requires permission)."""
 return user.update_profile(profile_data)

# 2. Cryptographic Failures - Secure password hashing
from bcrypt import hashpw, gensalt, checkpw

def hash_password(plaintext: str) -> str:
 """Hash password securely with bcrypt."""
 salt = gensalt(rounds=12) # Adaptive cost factor
 return hashpw(plaintext.encode('utf-8'), salt).decode('utf-8')

def verify_password(plaintext: str, hashed: str) -> bool:
 """Verify password against hash."""
 return checkpw(
 plaintext.encode('utf-8'),
 hashed.encode('utf-8')
 )

# 3. Injection Prevention - Parameterized queries
from sqlalchemy import text

def safe_user_query(username: str) -> List[User]:
 """Query users safely with parameterized query."""
 query = text("SELECT * FROM users WHERE username = :username")
 return db.session.execute(
 query,
 {"username": username} # Parameterized, not concatenated
 ).fetchall()

# 4. Security Misconfiguration - Environment-based config
import os
from pathlib import Path

def load_secure_config() -> dict:
 """Load configuration from environment variables."""
 config = {
 'DEBUG': os.getenv('DEBUG', 'false').lower() == 'true',
 'DATABASE_URL': os.getenv('DATABASE_URL'),
 'SECRET_KEY': os.getenv('SECRET_KEY'),
 'ALLOWED_HOSTS': os.getenv('ALLOWED_HOSTS', '').split(',')
 }
 
 # Validate required configs
 required = ['DATABASE_URL', 'SECRET_KEY']
 for key in required:
 if not config.get(key):
 raise ValueError(f"Required config missing: {key}")
 
 # Never expose secrets in logs
 logger.info(f"Config loaded (DEBUG={config['DEBUG']})")
 
 return config
```

Security Validation:
```bash
# Bandit security scan
bandit -r src/ -ll

# Dependency audit
pip-audit
safety check

# OWASP ZAP scan (for APIs)
zap-cli quick-scan http://localhost:8000

# Secret scanning
detect-secrets scan
```

---

### Principle 5: Trackable (T)

Traceability Requirements:

```
Commit Traceability:
 Conventional commit format
 Link to SPEC or issue
 Clear description of changes
 Test evidence included

Requirement Traceability:
 SPEC-XXX-REQ-YY mapping
 Implementation → Test linkage
 Test → Acceptance criteria
 Acceptance → User story

Quality Traceability:
 Coverage reports
 Quality metrics dashboard
 CI/CD pipeline results
 Code review history
```

Conventional Commit Format:

```bash
# Format: <type>(<scope>): <subject>
#
# <body>
#
# <footer>

# Examples:
feat(auth): Add OAuth2 integration

Implement OAuth2 authentication flow with Google provider.
Addresses SPEC-001-REQ-02.

- Add OAuth2 client configuration
- Implement callback handling
- Add token validation
- Test coverage: 92%

Closes #42

---

fix(api): Resolve JWT token expiry bug

JWT tokens were expiring 1 hour early due to timezone issue.
Fixes SPEC-001-REQ-03.

- Correct timezone handling in token generation
- Add expiry validation tests
- Update API documentation

Fixes #58

---

refactor(database): Optimize query performance

Improve query performance by adding database indexes.
Related to SPEC-005-REQ-01.

- Add indexes on frequently queried columns
- Reduce query time by 70%
- Add performance benchmarks

Performance improvement from 500ms → 150ms average query time.
```

Traceability Matrix:

```yaml
# .moai/specs/traceability.yaml
requirements:
 SPEC-001-REQ-01:
 description: "User registration with email/password"
 implementation:
 - src/auth/registration.py::register_user
 - src/models/user.py::User
 tests:
 - tests/auth/test_registration.py::test_register_user_success
 - tests/auth/test_registration.py::test_register_user_duplicate
 coverage: 95%
 status: Implemented
 
 SPEC-001-REQ-02:
 description: "OAuth2 authentication"
 implementation:
 - src/auth/oauth2.py::OAuth2Handler
 - src/auth/providers/google.py::GoogleOAuth2Provider
 tests:
 - tests/auth/test_oauth2.py::test_oauth2_flow
 - tests/auth/test_oauth2.py::test_token_validation
 coverage: 92%
 status: Implemented
 
 SPEC-001-REQ-03:
 description: "JWT token generation and validation"
 implementation:
 - src/auth/jwt_manager.py::JWTManager
 tests:
 - tests/auth/test_jwt.py::test_token_generation
 - tests/auth/test_jwt.py::test_token_validation
 - tests/auth/test_jwt.py::test_token_expiry
 coverage: 98%
 status: Implemented
```

---

## Advanced Implementation (10+ minutes)

### TRUST 5 CI/CD Integration

Complete Quality Gate Pipeline:

```yaml
# .github/workflows/trust-5-quality-gates.yml
name: TRUST 5 Quality Gates

on:
 pull_request:
 branches: [main, develop]
 push:
 branches: [main, develop]

jobs:
 test-first:
 name: "T1: Test Coverage ≥85%"
 runs-on: ubuntu-latest
 steps:
 - uses: actions/checkout@v4
 
 - name: Set up Python
 uses: actions/setup-python@v4
 with:
 python-version: '3.13'
 
 - name: Install dependencies
 run: |
 pip install -r requirements.txt
 pip install pytest pytest-cov
 
 - name: Run tests with coverage
 run: |
 pytest --cov=src --cov-report=xml --cov-fail-under=85
 
 - name: Upload coverage to Codecov
 uses: codecov/codecov-action@v3
 with:
 file: ./coverage.xml
 fail_ci_if_error: true
 
 readable:
 name: "R: Code Quality ≥8.0"
 runs-on: ubuntu-latest
 steps:
 - uses: actions/checkout@v4
 
 - name: Pylint check
 run: |
 pip install pylint
 pylint src/ --fail-under=8.0
 
 - name: Black format check
 run: |
 pip install black
 black --check src/
 
 - name: MyPy type check
 run: |
 pip install mypy
 mypy src/ --strict
 
 unified:
 name: "U: Consistency ≥90%"
 runs-on: ubuntu-latest
 steps:
 - uses: actions/checkout@v4
 
 - name: Architecture validation
 run: |
 python .moai/scripts/validate_architecture.py
 
 - name: Import consistency
 run: |
 pip install isort
 isort --check-only src/
 
 - name: Naming convention check
 run: |
 python .moai/scripts/check_naming_conventions.py
 
 secured:
 name: "S: Security Score 100"
 runs-on: ubuntu-latest
 steps:
 - uses: actions/checkout@v4
 
 - name: Bandit security scan
 run: |
 pip install bandit
 bandit -r src/ -ll
 
 - name: Dependency audit
 run: |
 pip install pip-audit safety
 pip-audit
 safety check
 
 - name: Secret scanning
 run: |
 pip install detect-secrets
 detect-secrets scan --baseline .secrets.baseline
 
 trackable:
 name: "T2: Traceability Check"
 runs-on: ubuntu-latest
 steps:
 - uses: actions/checkout@v4
 
 - name: Validate commit messages
 run: |
 python .moai/scripts/validate_commits.py
 
 - name: Check requirement traceability
 run: |
 python .moai/scripts/check_traceability.py
 
 - name: Generate traceability report
 run: |
 python .moai/scripts/generate_traceability_report.py
 
 quality-gate:
 name: "Final Quality Gate"
 needs: [test-first, readable, unified, secured, trackable]
 runs-on: ubuntu-latest
 steps:
 - name: All gates passed
 run: echo " TRUST 5 quality gates passed!"
```

### TRUST 5 Validation Framework

Comprehensive Validator Implementation:

```python
from dataclasses import dataclass
from typing import List, Dict, Any
import subprocess
import re

@dataclass
class ValidationResult:
 """Result of TRUST 5 validation."""
 passed: bool
 test_coverage: float
 code_quality: float
 consistency_score: float
 security_score: int
 traceability_score: float
 issues: List[str]
 warnings: List[str]
 
 def overall_score(self) -> float:
 """Calculate overall TRUST 5 score."""
 weights = {
 'test': 0.20,
 'quality': 0.20,
 'consistency': 0.20,
 'security': 0.20,
 'traceability': 0.20
 }
 return (
 self.test_coverage * weights['test'] +
 self.code_quality * weights['quality'] +
 self.consistency_score * weights['consistency'] +
 self.security_score * weights['security'] +
 self.traceability_score * weights['traceability']
 )

class TRUST5Validator:
 """Comprehensive TRUST 5 validation engine."""
 
 def __init__(self, src_dir: str = "src/"):
 self.src_dir = src_dir
 self.result = ValidationResult(
 passed=False,
 test_coverage=0.0,
 code_quality=0.0,
 consistency_score=0.0,
 security_score=0,
 traceability_score=0.0,
 issues=[],
 warnings=[]
 )
 
 def validate_all(self) -> ValidationResult:
 """Run all TRUST 5 validations."""
 # T1: Test-first
 self._validate_test_coverage()
 
 # R: Readable
 self._validate_readability()
 
 # U: Unified
 self._validate_consistency()
 
 # S: Secured
 self._validate_security()
 
 # T2: Trackable
 self._validate_traceability()
 
 # Final gate
 self.result.passed = all([
 self.result.test_coverage >= 85,
 self.result.code_quality >= 8.0,
 self.result.consistency_score >= 90,
 self.result.security_score == 100,
 self.result.traceability_score >= 80
 ])
 
 return self.result
 
 def _validate_test_coverage(self):
 """Validate test coverage ≥85%."""
 try:
 result = subprocess.run(
 ["pytest", "--cov", self.src_dir, "--cov-report=json"],
 capture_output=True,
 text=True,
 check=False
 )
 
 import json
 with open("coverage.json") as f:
 coverage_data = json.load(f)
 
 self.result.test_coverage = coverage_data['totals']['percent_covered']
 
 if self.result.test_coverage < 85:
 self.result.issues.append(
 f"Test coverage {self.result.test_coverage:.1f}% < 85% required"
 )
 
 except Exception as e:
 self.result.issues.append(f"Coverage validation failed: {e}")
 
 def _validate_readability(self):
 """Validate code quality with pylint."""
 try:
 result = subprocess.run(
 ["pylint", self.src_dir, "--output-format=json"],
 capture_output=True,
 text=True,
 check=False
 )
 
 import json
 pylint_output = json.loads(result.stdout)
 
 # Calculate score from pylint output
 score_match = re.search(r"rated at ([\d.]+)/10", result.stdout)
 if score_match:
 self.result.code_quality = float(score_match.group(1))
 
 if self.result.code_quality < 8.0:
 self.result.issues.append(
 f"Code quality {self.result.code_quality:.1f} < 8.0 required"
 )
 
 except Exception as e:
 self.result.issues.append(f"Readability validation failed: {e}")
 
 def _validate_consistency(self):
 """Validate architectural consistency."""
 try:
 # Check import consistency
 isort_result = subprocess.run(
 ["isort", "--check-only", self.src_dir],
 capture_output=True,
 check=False
 )
 
 # Check formatting consistency
 black_result = subprocess.run(
 ["black", "--check", self.src_dir],
 capture_output=True,
 check=False
 )
 
 # Score based on checks passed
 checks_passed = 0
 if isort_result.returncode == 0:
 checks_passed += 50
 if black_result.returncode == 0:
 checks_passed += 50
 
 self.result.consistency_score = checks_passed
 
 if self.result.consistency_score < 90:
 self.result.issues.append(
 f"Consistency {self.result.consistency_score}% < 90% required"
 )
 
 except Exception as e:
 self.result.issues.append(f"Consistency validation failed: {e}")
 
 def _validate_security(self):
 """Validate security with Bandit."""
 try:
 result = subprocess.run(
 ["bandit", "-r", self.src_dir, "-f", "json"],
 capture_output=True,
 text=True,
 check=False
 )
 
 import json
 bandit_output = json.loads(result.stdout)
 
 # Count high/medium severity issues
 issues = bandit_output.get('results', [])
 critical_issues = [
 i for i in issues 
 if i['issue_severity'] in ['HIGH', 'MEDIUM']
 ]
 
 if critical_issues:
 self.result.security_score = 0
 for issue in critical_issues:
 self.result.issues.append(
 f"Security: {issue['issue_text']} "
 f"({issue['issue_severity']})"
 )
 else:
 self.result.security_score = 100
 
 except Exception as e:
 self.result.issues.append(f"Security validation failed: {e}")
 
 def _validate_traceability(self):
 """Validate requirement traceability."""
 try:
 # Check commit message format
 result = subprocess.run(
 ["git", "log", "-1", "--pretty=%s"],
 capture_output=True,
 text=True,
 check=True
 )
 
 commit_msg = result.stdout.strip()
 
 # Conventional commit pattern
 pattern = r'^(feat|fix|docs|style|refactor|test|chore)(\(.+\))?: .{1,50}'
 
 if re.match(pattern, commit_msg):
 self.result.traceability_score = 100
 else:
 self.result.traceability_score = 50
 self.result.warnings.append(
 f"Commit message doesn't follow conventional format: {commit_msg}"
 )
 
 except Exception as e:
 self.result.warnings.append(f"Traceability validation failed: {e}")
 self.result.traceability_score = 0

# Usage
validator = TRUST5Validator(src_dir="src/")
result = validator.validate_all()

if result.passed:
 print(f" TRUST 5 validation passed! (Score: {result.overall_score():.1f})")
else:
 print(f" TRUST 5 validation failed!")
 for issue in result.issues:
 print(f" - {issue}")
```

### TRUST 5 Metrics Dashboard

```python
class TRUST5Metrics:
 """Real-time TRUST 5 quality metrics."""
 
 def __init__(self):
 self.test_coverage = 0.0 # Target: ≥85%
 self.code_quality = 0.0 # Target: ≥8.0
 self.consistency_score = 0.0 # Target: ≥90%
 self.security_score = 0 # Target: 100
 self.traceability_score = 0.0 # Target: ≥80%
 
 def update_metrics(self, validation_result: ValidationResult):
 """Update metrics from validation result."""
 self.test_coverage = validation_result.test_coverage
 self.code_quality = validation_result.code_quality
 self.consistency_score = validation_result.consistency_score
 self.security_score = validation_result.security_score
 self.traceability_score = validation_result.traceability_score
 
 def get_dashboard_data(self) -> dict:
 """Get metrics for dashboard display."""
 return {
 'overall_score': self.get_overall_score(),
 'production_ready': self.is_production_ready(),
 'metrics': {
 'test_coverage': {
 'value': self.test_coverage,
 'target': 85,
 'status': 'pass' if self.test_coverage >= 85 else 'fail'
 },
 'code_quality': {
 'value': self.code_quality,
 'target': 8.0,
 'status': 'pass' if self.code_quality >= 8.0 else 'fail'
 },
 'consistency': {
 'value': self.consistency_score,
 'target': 90,
 'status': 'pass' if self.consistency_score >= 90 else 'fail'
 },
 'security': {
 'value': self.security_score,
 'target': 100,
 'status': 'pass' if self.security_score == 100 else 'fail'
 },
 'traceability': {
 'value': self.traceability_score,
 'target': 80,
 'status': 'pass' if self.traceability_score >= 80 else 'fail'
 }
 }
 }
 
 def get_overall_score(self) -> float:
 """Calculate overall TRUST 5 score (0-100)."""
 weights = {
 'test': 0.20,
 'quality': 0.20,
 'consistency': 0.20,
 'security': 0.20,
 'traceability': 0.20
 }
 
 return (
 self.test_coverage * weights['test'] +
 (self.code_quality * 10) * weights['quality'] +
 self.consistency_score * weights['consistency'] +
 self.security_score * weights['security'] +
 self.traceability_score * weights['traceability']
 )
 
 def is_production_ready(self) -> bool:
 """Check if code meets production standards."""
 return (
 self.test_coverage >= 85 and
 self.code_quality >= 8.0 and
 self.consistency_score >= 90 and
 self.security_score == 100 and
 self.traceability_score >= 80
 )
```

---

## Works Well With

Agents:
- quality-gate - Automated TRUST 5 validation
- tdd-implementer - RED-GREEN-REFACTOR enforcement
- security-expert - OWASP compliance checking
- test-engineer - Test generation and coverage

Skills:
- moai-essentials-testing-integration - Test framework setup
- moai-domain-security - Security patterns
- moai-core-code-reviewer - Code quality analysis

Commands:
- /moai:2-run - Enforces ≥85% coverage requirement
- /moai:9-feedback - Quality improvement suggestions

Memory:
- Skill("moai-foundation-core") modules/execution-rules.md - Quality gates
- @.moai/config/config.yaml - constitution.test_coverage_target

---

Version: 1.0.0
Last Updated: 2025-11-25
Status: Production Ready
