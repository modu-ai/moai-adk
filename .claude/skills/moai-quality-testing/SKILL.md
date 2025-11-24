---
name: moai-quality-testing
description: Testing consolidating pytest, Vitest, Playwright for unit, integration, E2E testing with coverage analysis
version: 1.0.0
modularized: true
last_updated: 2025-11-24
allowed-tools:
  - Task
  - AskUserQuestion
  - Skill
  - Read
  - Write
  - Edit
compliance_score: 85
modules:
  - testing-frameworks
  - integration-patterns
  - best-practices
dependencies:
  - moai-foundation-trust
  - moai-lang-python
  - moai-lang-typescript
deprecated: false
successor: null
category_tier: 3
auto_trigger_keywords:
  - test
  - pytest
  - vitest
  - testing
  - unit-test
  - integration-test
  - e2e
  - playwright
  - coverage
  - assertions
agent_coverage:
  - test-engineer
  - tdd-implementer
  - quality-gate
context7_references:
  - pytest
  - vitest
  - playwright
invocation_api_version: "1.0"
---

## Quick Reference (30 seconds)

**Enterprise Testing Consolidated**

Unified testing framework consolidating domain-testing, essentials-testing, and Playwright into single comprehensive skill. Supports unit tests (pytest/Vitest), integration tests, E2E testing (Playwright), performance profiling, and coverage analysis with 85%+ targets.

**Core Capabilities**:
- ✅ Unit testing (pytest 8.3+, Vitest 2.x)
- ✅ Integration testing with fixtures and mocks
- ✅ E2E testing with Playwright 1.48+
- ✅ Performance testing and profiling
- ✅ Coverage analysis and reporting
- ✅ Test organization and strategies

**When to Use**:
- Writing tests (TDD RED-GREEN-REFACTOR cycle)
- Testing backend APIs and services
- Testing frontend components and interactions
- Validating integration points
- Ensuring coverage ≥85% (TRUST 5)

**Core Framework**: PYTEST + VITEST + PLAYWRIGHT
```
1. Unit Tests (fastest, isolated)
   ↓
2. Integration Tests (fixtures, mocks)
   ↓
3. E2E Tests (full system flow)
   ↓
4. Coverage Analysis & Reporting
```

---

## Core Patterns (5-10 minutes each)

### Pattern 1: Test-First TDD Cycle with pytest

**Concept**: Write failing tests first, then implement code to pass tests, then refactor.

```python
import pytest
from myapp.auth import authenticate

class TestAuthentication:
    """User authentication test suite."""

    def test_valid_credentials_returns_token(self):
        """GIVEN valid username/password, WHEN authenticate called, THEN token returned."""
        token = authenticate(username="alice", password="secret123")
        assert token is not None
        assert len(token) > 20  # JWT format

    def test_invalid_credentials_raises_error(self):
        """GIVEN invalid password, WHEN authenticate called, THEN AuthError raised."""
        with pytest.raises(AuthError) as exc_info:
            authenticate(username="alice", password="wrong")
        assert "Invalid credentials" in str(exc_info.value)

    def test_expired_token_rejected(self):
        """GIVEN expired token, WHEN verify_token called, THEN TokenExpiredError raised."""
        expired_token = "eyJ..." # expired JWT
        with pytest.raises(TokenExpiredError):
            verify_token(expired_token)

    @pytest.fixture
    def sample_user(self, db_session):
        """Create test user in temporary database."""
        user = User(username="alice", email="alice@test.com")
        db_session.add(user)
        db_session.commit()
        yield user
        db_session.delete(user)

    def test_register_new_user(self, sample_user):
        """GIVEN new username, WHEN register called, THEN user created."""
        assert sample_user.username == "alice"
```

**Use Case**: Implement authentication module with 100% test coverage.

---

### Pattern 2: Parameterized Testing for Multiple Cases

**Concept**: Run same test with different inputs to cover edge cases efficiently.

```python
import pytest

class TestDataValidation:
    """Input validation across multiple data types."""

    @pytest.mark.parametrize("email,is_valid", [
        ("alice@example.com", True),  # Valid email
        ("bob@test.co.uk", True),     # Valid with country TLD
        ("invalid.email", False),      # No @
        ("@example.com", False),       # No username
        ("", False),                   # Empty
        ("alice+tag@example.com", True),  # Plus addressing
    ])
    def test_email_validation(self, email, is_valid):
        """Validate email format across multiple cases."""
        assert is_valid_email(email) == is_valid

    @pytest.mark.parametrize("age,is_adult", [
        (18, True),
        (17, False),
        (65, True),
        (-1, False),
        (999, True),
    ])
    def test_age_validation(self, age, is_adult):
        """Validate age constraints."""
        assert is_adult_user(age) == is_adult
```

**Use Case**: Test validation logic covering normal, boundary, and invalid cases.

---

### Pattern 3: Async Testing with pytest-asyncio

**Concept**: Test async/await code with proper event loop and timeouts.

```python
import pytest
import asyncio
from httpx import AsyncClient

@pytest.mark.asyncio
async def test_async_api_call():
    """Test async HTTP request with FastAPI."""
    async with AsyncClient(app=app, base_url="http://test") as client:
        response = await client.get("/users/123")
        assert response.status_code == 200
        assert response.json()["id"] == 123

@pytest.mark.asyncio
async def test_concurrent_operations():
    """Test multiple concurrent operations."""
    results = await asyncio.gather(
        fetch_user(1),
        fetch_user(2),
        fetch_user(3),
    )
    assert len(results) == 3
    assert all(r.id for r in results)

@pytest.mark.asyncio
async def test_timeout_handling():
    """Test timeout handling in async code."""
    with pytest.raises(asyncio.TimeoutError):
        await asyncio.wait_for(long_running_task(), timeout=1.0)
```

**Use Case**: Test FastAPI endpoints, async database operations, concurrent tasks.

---

### Pattern 4: Frontend Testing with Vitest + Playwright

**Concept**: Unit test components with Vitest, E2E test user flows with Playwright.

```typescript
// Vitest component test
import { render, screen } from '@testing-library/react';
import { Button } from './Button';

describe('Button Component', () => {
    it('renders button with text', () => {
        render(<Button label="Click me" />);
        expect(screen.getByText('Click me')).toBeInTheDocument();
    });

    it('calls onClick handler when clicked', () => {
        const handleClick = vi.fn();
        render(<Button label="Click" onClick={handleClick} />);

        screen.getByText('Click').click();
        expect(handleClick).toHaveBeenCalledOnce();
    });
});
```

```typescript
// Playwright E2E test
import { test, expect } from '@playwright/test';

test('user login flow', async ({ page }) => {
    await page.goto('http://localhost:3000');

    // Fill login form
    await page.fill('input[name="email"]', 'alice@example.com');
    await page.fill('input[name="password"]', 'secret123');

    // Submit and verify redirect
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL('/dashboard');

    // Verify user is logged in
    await expect(page.locator('text=Welcome Alice')).toBeVisible();
});
```

**Use Case**: Test React components and user interactions end-to-end.

---

### Pattern 5: Coverage Analysis & Reporting

**Concept**: Measure test coverage, identify gaps, enforce minimum thresholds.

```bash
# Run tests with coverage
pytest tests/ --cov=src --cov-report=html --cov-report=term-missing

# Output:
# src/auth.py: 89% coverage (missing lines 45-47)
# src/database.py: 92% coverage
# TOTAL: 90% coverage ✅

# Enforce minimum coverage
pytest tests/ --cov=src --cov-fail-under=85
```

```python
# pytest.ini configuration
[pytest]
addopts =
    --cov=src
    --cov-fail-under=85
    --cov-report=html
    --cov-report=term-missing
    -v
```

**Use Case**: Enforce 85%+ coverage requirement in TRUST 5 quality gate.

---

## Advanced Documentation

For detailed testing patterns and implementation strategies:

- **[modules/testing-frameworks.md](modules/testing-frameworks.md)** - pytest, Vitest, Playwright detailed guides
- **[modules/integration-patterns.md](modules/integration-patterns.md)** - Database fixtures, API mocking, async patterns
- **[modules/best-practices.md](modules/best-practices.md)** - Test organization, naming, coverage strategies

---

## Best Practices

### ✅ DO
- Write tests FIRST (RED-GREEN-REFACTOR cycle)
- Use descriptive test names (`test_valid_email_accepted` not `test_email`)
- Test behavior, not implementation details
- Use fixtures for setup/teardown
- Mock external dependencies
- Parameterize tests for multiple cases
- Keep tests fast (< 100ms each)
- Organize by test pyramid (70% unit, 20% integration, 10% E2E)

### ❌ DON'T
- Skip tests (every change needs tests)
- Use `time.sleep()` in tests (use fixtures/mocking)
- Test multiple behaviors in one test
- Hardcode test data (use fixtures)
- Ignore test failures
- Skip edge cases
- Mix unit and E2E tests in same file
- Assume infrastructure (mock everything)

---

## Success Metrics

- **Coverage**: ≥85% (TRUST 5 requirement)
- **Test Speed**: <100ms average per test
- **Failure Clarity**: Test name explains failure
- **Mock Usage**: 100% of external dependencies mocked
- **Parameterization**: 3+ cases per input validation
- **Organization**: Clear separation (unit/integration/E2E)

---

## Context7 Integration

### Related Libraries & Tools
- **[pytest](/pytest-dev/pytest)**: Python testing framework with fixtures and parametrization
- **[Vitest](/vitest-dev/vitest)**: Blazing fast unit testing for JavaScript/TypeScript
- **[Playwright](/microsoft/playwright)**: Cross-browser E2E testing framework
- **[@testing-library/react](/testing-library/react)**: DOM testing utilities for React
- **[pytest-asyncio](/pytest-dev/pytest-asyncio)**: Async/await testing support

### Official Documentation
- [pytest Documentation](https://docs.pytest.org/) - Comprehensive testing guide
- [Vitest Guide](https://vitest.dev/) - Fast unit testing
- [Playwright Testing](https://playwright.dev/docs/intro) - E2E testing
- [pytest-asyncio](https://pytest-asyncio.readthedocs.io/) - Async test support

---

## Related Skills

- `moai-foundation-trust` (TRUST 5 coverage requirements)
- `moai-lang-python` (pytest patterns and libraries)
- `moai-lang-typescript` (Vitest and Playwright)
- `moai-quality-review` (Code review for test quality)

---

## Consolidation Source Skills

This skill consolidates:
- domain-testing (testing strategies)
- essentials-testing (core testing patterns)
- playwright (E2E testing)

All functionality preserved in unified architecture.

---

## Workflow Integration

**Typical Testing Workflow**:
```
1. RED: Write failing test
   ↓
2. GREEN: Implement code to pass
   ↓
3. REFACTOR: Clean up implementation
   ↓
4. Coverage Check: ≥85% required
   ↓
5. Quality Gate: All tests must pass
```

---

## Changelog

- **v1.0.0** (2025-11-24): Consolidated domain-testing, essentials-testing, playwright into unified moai-quality-testing

---

**Status**: Production Ready (Enterprise)
**Generated with**: MoAI-ADK Skill Factory
**Modular Architecture**: SKILL.md + 3 modules (testing-frameworks, integration-patterns, best-practices)
