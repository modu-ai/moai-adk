# TRUST 5 Validation Framework

Purpose: CI/CD integration, validation engine, and metrics dashboard for TRUST 5 quality gates.

Version: 1.0.0
Last Updated: 2026-01-06
Parent: [trust-5-framework.md](trust-5-framework.md)

---

## CI/CD Integration

### Complete Quality Gate Pipeline

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
        run: pip install -r requirements.txt pytest pytest-cov
      - name: Run tests with coverage
        run: pytest --cov=src --cov-report=xml --cov-fail-under=85
      - name: Upload coverage
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
        run: pip install pylint && pylint src/ --fail-under=8.0
      - name: Black format check
        run: pip install black && black --check src/
      - name: MyPy type check
        run: pip install mypy && mypy src/ --strict

  unified:
    name: "U: Consistency ≥90%"
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Architecture validation
        run: python .moai/scripts/validate_architecture.py
      - name: Import consistency
        run: pip install isort && isort --check-only src/

  secured:
    name: "S: Security Score 100"
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Bandit security scan
        run: pip install bandit && bandit -r src/ -ll
      - name: Dependency audit
        run: pip install pip-audit safety && pip-audit && safety check
      - name: Secret scanning
        run: pip install detect-secrets && detect-secrets scan

  trackable:
    name: "T2: Traceability Check"
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Validate commit messages
        run: python .moai/scripts/validate_commits.py
      - name: Check traceability
        run: python .moai/scripts/check_traceability.py

  quality-gate:
    name: "Final Quality Gate"
    needs: [test-first, readable, unified, secured, trackable]
    runs-on: ubuntu-latest
    steps:
      - name: All gates passed
        run: echo "TRUST 5 quality gates passed!"
```

---

## Validation Engine

### TRUST5Validator Class

```python
from dataclasses import dataclass
from typing import List
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
        weights = {'test': 0.20, 'quality': 0.20, 'consistency': 0.20,
                   'security': 0.20, 'traceability': 0.20}
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
            passed=False, test_coverage=0.0, code_quality=0.0,
            consistency_score=0.0, security_score=0, traceability_score=0.0,
            issues=[], warnings=[]
        )

    def validate_all(self) -> ValidationResult:
        """Run all TRUST 5 validations."""
        self._validate_test_coverage()
        self._validate_readability()
        self._validate_consistency()
        self._validate_security()
        self._validate_traceability()

        self.result.passed = all([
            self.result.test_coverage >= 85,
            self.result.code_quality >= 8.0,
            self.result.consistency_score >= 90,
            self.result.security_score == 100,
            self.result.traceability_score >= 80
        ])
        return self.result
```

---

## Metrics Dashboard

```python
class TRUST5Metrics:
    """Real-time TRUST 5 quality metrics."""

    def __init__(self):
        self.test_coverage = 0.0      # Target: ≥85%
        self.code_quality = 0.0       # Target: ≥8.0
        self.consistency_score = 0.0  # Target: ≥90%
        self.security_score = 0       # Target: 100
        self.traceability_score = 0.0 # Target: ≥80%

    def get_dashboard_data(self) -> dict:
        """Get metrics for dashboard display."""
        return {
            'overall_score': self.get_overall_score(),
            'production_ready': self.is_production_ready(),
            'metrics': {
                'test_coverage': {'value': self.test_coverage, 'target': 85},
                'code_quality': {'value': self.code_quality, 'target': 8.0},
                'consistency': {'value': self.consistency_score, 'target': 90},
                'security': {'value': self.security_score, 'target': 100},
                'traceability': {'value': self.traceability_score, 'target': 80}
            }
        }

    def get_overall_score(self) -> float:
        """Calculate overall TRUST 5 score (0-100)."""
        return (
            self.test_coverage * 0.20 +
            (self.code_quality * 10) * 0.20 +
            self.consistency_score * 0.20 +
            self.security_score * 0.20 +
            self.traceability_score * 0.20
        )

    def is_production_ready(self) -> bool:
        """Check if code meets production standards."""
        return (self.test_coverage >= 85 and self.code_quality >= 8.0 and
                self.consistency_score >= 90 and self.security_score == 100 and
                self.traceability_score >= 80)
```

---

## Works Well With

- [trust-5-framework.md](trust-5-framework.md) - Overview and principles
- [trust-5-implementation.md](trust-5-implementation.md) - Code patterns

---

Version: 1.0.0
Last Updated: 2026-01-06
