---
name: "moai-domain-testing"
version: "4.0.0"
description: "Enterprise testing framework with pytest 8.4.x, Vitest 4.x, Playwright 1.48.x, Testing Library 15.x, httpx 0.28.x, k6 load testing, and accessibility testing"
allowed_tools: 
  - Read
  - Bash
  - WebSearch
  - WebFetch
status: stable
---

# Enterprise Testing Framework & Quality Assurance - v4.0.0

## Skill Overview

**Stable Versions (November 2025)**:
- pytest 8.4.2 (Python comprehensive testing)
- Vitest 4.x (TypeScript browser mode stable)
- Playwright 1.48.x (Multi-browser automation)
- Testing Library 15.x (User-centric component testing)
- httpx 0.28.x (Async HTTP testing)
- k6 1.0+ (Modern load testing)
- axe-core 4.8.x (WCAG 2.1 compliance)

**Focus**: Production-grade testing strategies with 85%+ coverage targets

---

## Level 1: Quick Reference (50-150 lines)

### Essential Testing Patterns

**pytest 8.4.x Foundation**:
```python
# conftest.py - Essential fixtures
import pytest
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

@pytest.fixture(scope="session")
def db_engine():
    """Database engine for entire test session."""
    engine = create_engine("sqlite:///:memory:")
    yield engine
    engine.dispose()

@pytest.fixture
def db_session(db_engine):
    """Isolated database session for each test."""
    connection = db_engine.connect()
    transaction = connection.begin()
    session = sessionmaker(bind=connection)()
    
    yield session
    
    session.close()
    transaction.rollback()
    connection.close()

@pytest.fixture
def user_factory():
    """Factory for creating test users."""
    def _create_user(name="Test User", email="test@example.com", **kwargs):
        return {
            "name": name,
            "email": email,
            "is_active": True,
            **kwargs
        }
    return _create_user

# Test usage
def test_user_creation(db_session, user_factory):
    user = user_factory(email="custom@example.com")
    assert user["email"] == "custom@example.com"
```

**Vitest 4.x TypeScript Setup**:
```typescript
// vitest.config.ts - Enterprise configuration
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    browser: {
      provider: 'playwright',
      enabled: true,
      instances: [{ browser: 'chromium' }, { browser: 'firefox' }],
    },
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      lines: 85,
      functions: 85,
      branches: 80,
    },
    typecheck: { enabled: true },
    setupFiles: ['./vitest.setup.ts'],
  },
});
```

**Playwright 1.48.x E2E Foundation**:
```typescript
// playwright.config.ts - Multi-browser setup
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },

  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
    { name: 'firefox', use: { ...devices['Desktop Firefox'] } },
  ],
});
```

**Testing Library User-Centric Testing**:
```typescript
// React Testing Library example
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

test('user can submit form with valid data', async () => {
  const user = userEvent.setup();
  render(<LoginForm onSubmit={vi.fn()} />);

  await user.type(
    screen.getByRole('textbox', { name: /email/i }),
    'user@example.com'
  );

  const passwordInput = screen.getByLabelText(/password/i);
  await user.type(passwordInput, 'password123');

  await user.click(
    screen.getByRole('button', { name: /submit/i })
  );

  expect(screen.getByText(/success/i)).toBeInTheDocument();
});
```

**Core Testing Principles**:
- ✅ Test isolation (no shared state)
- ✅ User-centric testing (Testing Library)
- ✅ 85%+ coverage targets
- ✅ Multi-browser E2E testing
- ✅ Accessibility compliance testing

---

## Level 2: Practical Implementation (200-300 lines)

### Advanced Testing Architecture

**pytest Advanced Patterns**:
```python
# Async testing with httpx
@pytest.mark.asyncio
async def test_api_endpoint():
    """Test async HTTP client with httpx."""
    async with httpx.AsyncClient() as client:
        response = await client.get(
            "http://localhost:8000/api/users",
            timeout=10.0
        )
        assert response.status_code == 200
        data = response.json()
        assert len(data) > 0

# Parametrized testing
@pytest.mark.parametrize("input_value,expected", [
    ("valid", True),
    ("invalid", False),
    ("edge_case", None),
])
def test_validator_with_params(input_value, expected, db_session):
    result = validate_input(input_value, db_session)
    assert result == expected
```

**Playwright Page Object Pattern**:
```typescript
// e2e/fixtures.ts - Reusable fixtures
import { test as baseTest } from '@playwright/test';

class LoginPage {
  constructor(private page: Page) {}
  
  async login(email: string, password: string) {
    await this.page.fill('[data-testid="email"]', email);
    await this.page.fill('[data-testid="password"]', password);
    await this.page.click('[data-testid="login-btn"]');
    await this.page.waitForURL('/dashboard');
  }
}

export const test = baseTest.extend({
  loginPage: async ({ page }, use) => {
    await use(new LoginPage(page));
  },
});

// Usage
test('user login flow', async ({ page, loginPage }) => {
  await page.goto('/login');
  await loginPage.login('user@example.com', 'password123');
  await expect(page).toHaveURL('/dashboard');
});
```

**Test Data Factories**:
```python
# tests/factories.py - Polyfactory for Pydantic models
from polyfactory.factories.pydantic_factory import ModelFactory
from pydantic import BaseModel

class User(BaseModel):
    id: int
    email: str
    name: str
    is_active: bool

class UserFactory(ModelFactory[User]):
    __model__ = User

# Generate test data
user = UserFactory.create()  # Single instance
users = UserFactory.batch(5)  # Batch of 5
```

**Accessibility Testing**:
```typescript
// tests/accessibility.spec.ts - WCAG 2.1 compliance
import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

test('accessibility audit: homepage', async ({ page }) => {
  await page.goto('http://localhost:3000');

  const accessibilityScanResults = await new AxeBuilder({ page })
    .withTags(['wcag2aa', 'wcag2aaa'])
    .analyze();

  expect(accessibilityScanResults.violations).toEqual([]);
});
```

**Mock & Stub Patterns**:
```python
# tests/test_external_services.py - pytest-mock integration
def test_email_sending(mocker):
    """Mock external email service."""
    mock_send = mocker.patch('app.services.EmailService.send')
    mock_send.return_value = True

    send_notification(user_id=1, message="Hello")

    mock_send.assert_called_once_with(
        to="user@example.com",
        subject="Notification",
        body="Hello"
    )
```

**k6 Load Testing**:
```javascript
// tests/load.js - Modern load testing with k6
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 20 },
    { duration: '1m30s', target: 20 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed: ['rate<0.1'],
  },
};

export default function () {
  const response = http.get('http://localhost:3000/api/users');
  check(response, {
    'GET status 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });

  sleep(1);
}
```

---

## Level 3: Advanced Integration (50-150 lines)

### Enterprise Testing Strategy

**Testing Pyramid Implementation**:
```bash
# Test execution strategy
# Local development
pytest --watch

# Pre-commit: Unit + fast integration
pytest tests/unit tests/integration -k "not slow"

# CI pipeline:
# 1. Unit tests (pytest + Vitest)
pytest tests/unit --cov --cov-fail-under=85
npm run test:unit

# 2. Integration tests
pytest tests/integration

# 3. E2E tests
npx playwright test

# 4. Performance baselines
k6 run tests/load.js

# 5. Accessibility audit
npm run test:a11y
```

**Coverage.py Configuration**:
```python
# .coveragerc - Enterprise coverage configuration
[run]
source = src
omit =
    */tests/*
    */test_*.py
    */__pycache__/*
    */.venv/*

[report]
exclude_lines =
    pragma: no cover
    def __repr__
    if __name__ == .__main__.:
    raise AssertionError
    raise NotImplementedError
    if TYPE_CHECKING:
    @abstractmethod

precision = 2
show_missing = True
skip_covered = False

[html]
directory = htmlcov
```

**Enterprise Test Strategy**:
```
Testing Pyramid Approach:

        /\
       /  \  E2E Tests (10-15%)
      /    \  - Playwright
     /______\  - Cypress
     
    /        \
   /  API     \  Integration Tests (25-35%)
  /  Tests    \  - httpx
 /____________\  - Supertest

/              \
/   Unit Tests   \  Unit Tests (50-60%)
\   (pytest,     /  - pytest
 \   Vitest)    /   - Vitest
  \____________/
```

**Best Practices Checklist**:
- [x] **Test Isolation**: Each test runs independently
- [x] **Fixture Scoping**: Use appropriate scopes (function/module/session)
- [x] **Async Handling**: Proper async/await in tests and fixtures
- [x] **Mock Strategy**: Mock external services, not internal logic
- [x] **Coverage Targets**: Maintain ≥85% code coverage
- [x] **CI/CD Integration**: Tests run in automated pipelines
- [x] **Performance**: Tests complete in <30s (unit), <5m (integration), <10m (E2E)
- [x] **Accessibility**: WCAG 2.1 compliance automated in E2E tests

---

**Version**: 4.0.0 Enterprise  
**Last Updated**: 2025-11-13  
**Status**: Production Ready  
**Coverage Target**: 85%+  
**Testing Pyramid**: Unit (50-60%), Integration (25-35%), E2E (10-15%)
