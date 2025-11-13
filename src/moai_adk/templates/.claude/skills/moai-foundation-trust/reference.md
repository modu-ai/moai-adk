# moai-foundation-trust: Reference & Implementation Guide

## TRUST 5 Quick Validation Checklist

### Before Every Commit

```bash
# T: Test First
pytest --cov=src --cov-report=term --cov-fail-under=85

# R: Readable
pylint src/ --fail-under=8.0
black --check src/

# U: Unified
python .moai/scripts/validation/architecture_checker.py

# S: Secured
bandit -r src/ -ll
pip audit

# T: Trackable
python .moai/scripts/validation/tag_validator.py

# Result: All checks pass → Ready to commit ✓
```

## Principle 1: Test First - Metrics & Validation

### Coverage Targets (November 2025)

| Metric | Minimum | Target | Excellent |
|--------|---------|--------|-----------|
| **Unit Test Coverage** | 75% | 85% | 95% |
| **Branch Coverage** | 70% | 80% | 90% |
| **Critical Path** | 95% | 100% | 100% |
| **Line Coverage** | 75% | 85% | 95% |

### Test Execution

```bash
# Run all tests with coverage
pytest tests/ --cov=src --cov-report=html --cov-report=term

# Generate HTML report
coverage html
# View: htmlcov/index.html

# Check specific module coverage
coverage report --include=src/auth.py

# Find uncovered lines
coverage report --show-missing
```

### Test Quality Metrics

```python
# Test file structure (enforce via code review)
def test_function_behavior_with_inputs_expects_output():
    """
    Test name format: test_[unit]_[condition]_[expected_result]
    
    Structure:
    1. Setup: Create test fixtures
    2. Execute: Call function under test
    3. Verify: Assert expected behavior
    """
    # Setup
    user = create_test_user("test@example.com", "TestPass123")
    
    # Execute
    result = authenticate_user(user.email, "TestPass123")
    
    # Verify
    assert result is not None
    assert result.startswith("eyJ")  # JWT token format
```

## Principle 2: Readable - Code Quality Tools

### Pylint Configuration

```ini
# .pylintrc
[MASTER]
fail-under = 8.0

[DESIGN]
max-locals = 15
max-branches = 12
max-attributes = 7

[COMPLEXITY]
max-complexity = 10

[FORMAT]
max-line-length = 100

[MESSAGES CONTROL]
disable = missing-docstring  # Enforce in code review
```

### Black Auto-Formatting

```bash
# Format all Python files
black src/ tests/

# Check without modifying
black --check src/ tests/

# Configure in pyproject.toml
[tool.black]
line-length = 100
target-version = ['py312']
```

### Cyclomatic Complexity Rules

```
CC = 1: Low complexity ✓ Excellent
CC = 2-5: Moderate ✓ Good
CC = 6-10: High ⚠ Accept with review
CC = 11+: Very high ✗ Refactor required

Calculation:
  CC = number of independent paths through function
  IF/ELSE = +1 path
  LOOPS = +1 path
  BOOLEAN_AND/OR = +1 path
```

### Readability Verification Script

```python
# .moai/scripts/validation/readability_checker.py
import ast
import subprocess

def check_readability(src_dir):
    """Verify readability metrics"""
    
    # 1. Check naming conventions
    # 2. Check function length
    # 3. Check comment ratio
    # 4. Check variable clarity
    # 5. Report violations
    
    issues = []
    for file in glob(f"{src_dir}/**/*.py"):
        # ... implementation
        pass
    
    return issues
```

## Principle 3: Unified - Architecture Validation

### Directory Structure Validation

```bash
# Verify consistent structure
find src/ -name "__init__.py" | wc -l   # Every package has __init__.py
find src/ -name "*.py" | grep -c "^src/[a-z_]/[a-z_]*.py$"  # Flat structure

# Check naming consistency
find src/ -name "*util*.py"    # Find any util.py (use domain-specific names)
find src/ -name "*helper*.py"  # Find any helper.py (split into domain)
```

### Pattern Consistency Checker

```python
# src/auth/authenticate.py
# Consistent pattern: imports → docstring → classes → functions

"""
User authentication module.

Standard structure enforced by architecture validator.
"""

# 1. Imports
import logging
from typing import Optional
from src.models import User
from src.exceptions import AuthenticationError

# 2. Module constants
logger = logging.getLogger(__name__)
MAX_LOGIN_ATTEMPTS = 5

# 3. Classes
class AuthenticationManager:
    """Unified authentication handler"""
    pass

# 4. Public functions
def authenticate_user(email: str, password: str) -> str:
    """Public API"""
    pass

# 5. Private helpers
def _validate_email(email: str) -> bool:
    """Private helper"""
    pass
```

## Principle 4: Secured - Security Validation Tools

### Bandit Configuration

```bash
# Scan for security issues
bandit -r src/ -ll  # Report high/medium only

# Custom config
[bandit]
exclude_dirs = ['tests', 'docs']
skips = ['B101']  # Skip assert_used in tests
```

### Dependency Security

```bash
# Check for known vulnerabilities
pip audit

# Fix vulnerabilities
pip install --upgrade vulnerable-package

# Pin versions in requirements.txt
cryptography==41.0.7  # Pin exact version (not cryptography>=41.0)
```

### Security Test Examples

```python
def test_password_hash_not_reversible():
    """Verify passwords cannot be reversed"""
    plaintext = "SecurePassword123"
    hashed = hash_password(plaintext)
    
    # Hash must not contain original
    assert plaintext not in hashed
    assert len(hashed) > 20  # Bcrypt format

def test_query_safe_against_sql_injection():
    """Verify parameterized queries prevent injection"""
    # Dangerous: f"SELECT * FROM users WHERE email='{email}'"
    # Safe: execute("SELECT * FROM users WHERE email=?", [email])
    
    result = find_user_by_email("test@example.com' OR '1'='1")
    # Query safe because parameterized
    assert result is None
```

## Principle 5: Trackable - TAG Validation

### TAG Scanning Command

```bash
# Find all TAGs
rg '@(SPEC|TEST|CODE|DOC):' --no-filename -o | sort | uniq -c

# Validate chains
python .moai/scripts/validation/tag_validator.py

# Report by type
```

### TAG Linking Rules

```
Valid links:

Invalid patterns (will fail validation):
└─ Circular references
```

## TRUST 5 Integration with CI/CD

### GitHub Actions Example

```yaml
# .github/workflows/trust-5-gate.yml
name: TRUST 5 Quality Gates

on: [push, pull_request]

jobs:
  trust-5:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.12'
      
      - name: Install dependencies
        run: |
          pip install -r requirements.txt
          pip install pytest coverage pylint black bandit
      
      - name: T - Test First
        run: |
          pytest --cov=src --cov-fail-under=85
          
      - name: R - Readable
        run: |
          pylint src/ --fail-under=8.0
          black --check src/ tests/
          
      - name: U - Unified
        run: |
          python .moai/scripts/validation/architecture_checker.py
          
      - name: S - Secured
        run: |
          bandit -r src/ -ll
          pip audit
          
      - name: T - Trackable
        run: |
          python .moai/scripts/validation/tag_validator.py
      
      - name: Report Results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: trust-5-report
          path: .moai/reports/
```

## Troubleshooting Common Issues

### Issue: Coverage < 85%
```bash
# Find uncovered lines
coverage report --show-missing

# Add tests for missing coverage
# Rerun: pytest --cov --cov-report=html

# If unreachable code, mark as such:
if False:  # pragma: no cover
    unreachable_code()
```

### Issue: Pylint score < 8.0
```bash
# See violations
pylint src/

# Fix common issues
black src/  # Auto-format
pylint --generate-rcfile > .pylintrc  # Generate config

# Discuss with team if score acceptable
```

### Issue: Bandit finds vulnerability
```bash
# Review finding
bandit -r src/ -ll -v

# If false positive:
# @bandit: disable=B101
code_with_assert()

# If real vulnerability: Fix code before merge
```

### Issue: TAG validation fails
```bash
# Check TAG syntax
rg '@(SPEC|TEST|CODE|DOC):' -n

# Find orphans
python .moai/scripts/validation/tag_validator.py --report

# Link or mark deprecated
```

## TRUST 5 Team Responsibilities

| Role | TRUST Principle | Responsibility |
|------|-----------------|-----------------|
| **Developer** | All 5 | Implement with TRUST 5 in mind |
| **Test Engineer** | Test First | Maintain ≥85% coverage |
| **Code Reviewer** | Readable, Unified | Review code quality + consistency |
| **Security Engineer** | Secured | Review security implementation |
| **DevOps/QA** | Trackable | Maintain CI/CD validation |

## TRUST 5 Onboarding Checklist

- [ ] Read TRUST 5 principles (Level 1)
- [ ] Review validation tools installed
- [ ] Configure IDE with linters (pylint, black)
- [ ] Review TRUST 5 code examples
- [ ] Run first commit with all gates passing
- [ ] Understand TAG system (see moai-foundation-tags)
- [ ] Join TRUST 5 review meetings
- [ ] Mentor new team members

