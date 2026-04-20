# TRUST 5 Framework - Quality Assurance System

Purpose: Automated quality gates ensuring code quality, security, maintainability, and traceability through five core principles.

Version: 2.0.0 (Modular Split)
Last Updated: 2026-01-06

---

## Quick Reference (30 seconds)

TRUST 5 is MoAI-ADK's comprehensive quality assurance framework enforcing five pillars:

1. Test-first(T) - ≥85% coverage, RED-GREEN-REFACTOR cycle
2. Readable(R) - Clear naming, ≤10 cyclomatic complexity
3. Unified(U) - Consistent patterns, architecture compliance
4. Secured(S) - OWASP Top 10 compliance, security validation
5. Trackable(T) - Clear commits, requirement traceability

Integration Points:
- Pre-commit hooks - Automated validation
- CI/CD pipelines - Quality gate enforcement
- quality-gate agent - TRUST 5 validation
- /moai:2-run - Enforces ≥85% coverage

Quick Validation:
```python
validations = [
    test_coverage >= 85,    # T
    complexity <= 10,       # R
    consistency > 90,       # U
    security_score == 100,  # S
    has_clear_commits       # T
]
```

Extended Documentation:
- [Implementation Details](trust-5-implementation.md) - Detailed patterns and code examples
- [Validation Framework](trust-5-validation.md) - CI/CD integration and metrics

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

- Critical (≥85%): Required for merge
- Warning (70-84%): Review required
- Failing (<70%): Block merge, generate tests

Validation Commands:
```bash
# Run tests with coverage
pytest --cov=src --cov-report=html --cov-fail-under=85

# Generate coverage report
coverage report -m
```

---

### Principle 2: Readable (R)

Readability Metrics:

- Cyclomatic Complexity: ≤10 (max 15)
- Function Length: ≤50 lines (max 100)
- Nesting Depth: ≤3 levels (max 5)
- Comment Ratio: 15-20% (min 10%)
- Type Hint Coverage: 100% (min 90%)

Readability Checklist:

- Clear function/variable names (noun_verb pattern)
- Single responsibility principle
- Type hints on all parameters and returns
- Docstrings with examples (Google style)
- No magic numbers (use named constants)
- DRY principle applied (no code duplication)
- SOLID principles followed

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

Architecture Consistency:
- Same pattern across all modules
- Same error handling approach
- Same logging strategy
- Same naming conventions

Testing Consistency:
- Same test structure (Arrange-Act-Assert)
- Same fixtures/factories
- Same assertion patterns
- Same mock strategies

Documentation Consistency:
- Same docstring format (Google style)
- Same README structure
- Same API documentation
- Same changelog format (conventional commits)

Validation Tools:
```bash
# Check architecture compliance
python .moai/scripts/validate_architecture.py

# Check consistent imports
isort --check-only src/
```

---

### Principle 4: Secured (S)

OWASP Top 10 (2024) Compliance:

1. Broken Access Control - RBAC, permission checks
2. Cryptographic Failures - bcrypt, proper encryption
3. Injection - Parameterized queries
4. Insecure Design - Threat modeling
5. Security Misconfiguration - Environment variables
6. Vulnerable Components - Dependency scanning
7. Authentication Failures - MFA, secure sessions
8. Data Integrity - Checksums, signatures
9. Logging Failures - Comprehensive logging
10. SSRF - URL validation

Security Validation:
```bash
# Bandit security scan
bandit -r src/ -ll

# Dependency audit
pip-audit
safety check

# Secret scanning
detect-secrets scan
```

---

### Principle 5: Trackable (T)

Traceability Requirements:

Commit Traceability:
- Conventional commit format
- Link to SPEC or issue
- Clear description of changes
- Test evidence included

Requirement Traceability:
- SPEC-XXX-REQ-YY mapping
- Implementation - Test linkage
- Test - Acceptance criteria
- Acceptance - User story

Conventional Commit Format:
```bash
# Format: <type>(<scope>): <subject>
feat(auth): Add OAuth2 integration

Implement OAuth2 authentication flow with Google provider.
Adddesses SPEC-001-REQ-02.

Closes #42
```

---

## Advanced Patterns

For comprehensive implementation patterns including CI/CD integration, validation frameworks, and metrics dashboards, see:

- [TRUST 5 Implementation](trust-5-implementation.md) - Detailed code patterns
- [TRUST 5 Validation](trust-5-validation.md) - CI/CD and metrics

---

## Works Well With

Agents:
- quality-gate - Automated TRUST 5 validation
- ddd-implementer - ANALYZE-PRESERVE-IMPROVE enforcement
- security-expert - OWASP compliance checking
- test-engineer - Test generation and coverage

Skills:
- moai-workflow-testing - Test framework setup
- moai-domain-security - Security patterns

Commands:
- /moai:2-run - Enforces ≥85% coverage requirement
- /moai:9-feedback - Quality improvement suggestions

---

Version: 2.0.0
Last Updated: 2026-01-06
Status: Production Ready
