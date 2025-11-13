---
name: "moai-domain-testing"
version: "4.0.0"
description: Enterprise-grade testing framework expertise with pytest 8.4.x, Vitest 4.x TypeScript, Playwright 1.48.x E2E multi-browser automation, Cypress 14.x, Testing Library 15.x (React/Vue/Svelte), httpx 0.28.x async API testing, Mocha 10.x, Coverage.py 7.11.x, k6 1.0+ load testing, performance testing, accessibility testing with axe-core WCAG 2.1 compliance; activates for unit testing, integration testing, E2E automation, API testing, test coverage strategy, test data factories, mock and stub patterns, CI/CD test automation, and production-grade testing architecture.
allowed-tools: 
  - Read
  - Bash
  - WebSearch
  - WebFetch
status: stable
---

# 🧪 Enterprise Testing Framework & Quality Assurance — v4.0

## 🎯 Skill Metadata

| Field | Value |
| ----- | ----- |
| **Version** | **4.0.0 Enterprise** |
| **Created** | 2025-11-12 |
| **Updated** | 2025-11-12 |
| **Allowed tools** | Read, Bash, WebSearch, WebFetch |
| **Auto-load** | On-demand for testing, QA, and quality assurance requests |
| **Trigger cues** | pytest, Vitest, Playwright, Cypress, Testing Library, httpx, test coverage, E2E testing, unit testing, API testing, test data, fixtures, mocks, stubs, load testing, performance testing, accessibility testing, test strategy, CI/CD testing |
| **Tier** | **4 (Enterprise)** |
| **Lines** | 950+ lines |
| **Size** | ~32KB |
| **Stable Versions** | pytest 8.4.x, Vitest 4.x, Playwright 1.48.x, Cypress 14.x, Testing Library 15.x, httpx 0.28.x, Coverage.py 7.11.x, k6 1.0+, axe-core 4.8.x |

---

## Enterprise Testing Stack v4.0 — November 2025 Stable Versions

### Technology Domains

**Unit Testing Frameworks**:
- pytest 8.4.2 (Python, comprehensive fixture system, parametrization)
- Vitest 4.x (JavaScript/TypeScript, Vite-powered, Browser mode stable)
- Jest 30.x (JavaScript, snapshot testing, coverage built-in)
- Mocha 10.x (Node.js, flexible, async/await support)

**E2E & Integration Testing**:
- Playwright 1.48.x (multi-browser automation, trace viewer, stable)
- Cypress 14.x (interactive debugging, real browser testing)
- Selenium 4.x (legacy enterprise support)
- TestCafe 1.23.x (stablemulti-browser, no plugins required)

**API Testing**:
- httpx 0.28.x (async Python, cookies, auth, streaming)
- pytest-asyncio (async test support with pytest)
- Supertest (Node.js HTTP assertions)
- REST Client (VS Code debugging)

**Component & DOM Testing**:
- Testing Library 15.x (React, Vue, Svelte, Solid integration)
- @testing-library/react 15.x (user-centric testing)
- @testing-library/vue 8.x (Vue 3 stable)
- @testing-library/svelte 4.x (Svelte 5 compatible)

**Test Data & Fixtures**:
- polyfactory 2.x (mock data generation, Pydantic/dataclass)
- pytest-factoryboy (factory pattern with pytest fixtures)
- factory-boy 3.x (established Django/Python factory standard)
- Faker 28.x (realistic fake data generation)

**Test Coverage & Quality**:
- Coverage.py 7.11.3 (Python, line/branch coverage, reports)
- c8 0.20.x (JavaScript coverage, nyc-compatible)
- Istanbul 1.x (JavaScript/TypeScript coverage)
- pytest-cov (coverage plugin for pytest)

**Performance & Load Testing**:
- k6 1.0+ (modern load testing with Grafana integration)
- Locust 2.x (Python load testing framework)
- ApacheBench (simple performance baseline)
- Artillery 2.x (cloud load testing)

**Accessibility Testing**:
- axe-core 4.8.x (WCAG 2.1/2.2 automated compliance)
- pa11y 6.x (automated accessibility testing)
- deque/axe-playwright (Playwright integration)
- Lighthouse API (Google performance/accessibility)

**Mocking & Stubbing**:
- pytest-mock (pytest plugin for mock/patch)
- unittest.mock (Python standard library)
- sinon.js 18.x (JavaScript spy/stub/mock library)
- MSW 2.x (Mock Service Worker, API mocking)

---

## 1. Unit Testing with pytest 8.4.x — Python

### Fixture Architecture & Parametrization

**Advanced Fixture Patterns**:

```python
# conftest.py - Fixture configuration with parametrization
import pytest
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

# Session-scoped database fixture (expensive resource)
@pytest.fixture(scope="session")
def db_engine():
    """Create database engine for entire test session."""
    engine = create_engine("sqlite:///:memory:")
    yield engine
    engine.dispose()

# Function-scoped database transaction fixture
@pytest.fixture
def db_session(db_engine):
    """Create isolated database session for each test."""
    connection = db_engine.connect()
    transaction = connection.begin()
    session = sessionmaker(bind=connection)()
    
    yield session
    
    session.close()
    transaction.rollback()
    connection.close()

# Parametrized fixture with indirect parametrization
@pytest.fixture(params=[
    {"timeout": 1, "retries": 3},
    {"timeout": 5, "retries": 1},
    {"timeout": 10, "retries": 0},
])
def http_config(request):
    """Parametrized HTTP configuration fixture."""
    return request.param

# Factory fixture for dynamic object creation
@pytest.fixture
def user_factory():
    """Factory fixture for creating user test objects."""
    def _create_user(name="Test User", email="test@example.com", **kwargs):
        return {
            "name": name,
            "email": email,
            "is_active": True,
            **kwargs
        }
    return _create_user
```

**Parametrize with Fixtures & Tests**:

```python
# Test parametrization combining fixtures and values
import pytest

@pytest.mark.parametrize("input_value,expected", [
    ("valid", True),
    ("invalid", False),
    ("edge_case", None),
])
def test_validator_with_params(input_value, expected, db_session):
    """Test with both parametrized values and fixtures."""
    result = validate_input(input_value, db_session)
    assert result == expected

# Indirect parametrization (pass values to fixtures)
@pytest.mark.parametrize("http_config", [
    {"timeout": 1, "retries": 3},
    {"timeout": 10, "retries": 0},
], indirect=True)
def test_http_resilience(http_config):
    """Test HTTP client with parametrized configurations."""
    client = HTTPClient(**http_config)
    response = client.get("https://api.example.com/health")
    assert response.status_code == 200
```

### Test Organization & Markers

```python
# pytest.ini - Test markers configuration
[pytest]
markers =
    unit: Unit tests (fast, isolated)
    integration: Integration tests (medium speed, external services)
    e2e: End-to-end tests (slow, full workflow)
    slow: Slow tests (mark with @pytest.mark.slow)
    skip_ci: Skip in CI environments
    security: Security-focused tests
    performance: Performance tests

# Running tests by marker
# pytest -m "unit and not slow"  # Unit tests excluding slow tests
# pytest -m "integration or e2e"  # Integration or E2E tests
```

**Fixture-Based Test Organization**:

```python
# tests/conftest.py - Shared fixtures across test suite
@pytest.fixture(autouse=True)
def reset_app_state():
    """Auto-reset application state before each test."""
    yield
    clear_cache()
    reset_database()

@pytest.fixture
def authenticated_client(db_session):
    """Client with authenticated user session."""
    user = create_test_user(db_session)
    client = APIClient(session=db_session, user=user)
    return client

@pytest.fixture
def api_base_url():
    """Base URL for API endpoints."""
    return "http://localhost:8000/api/v1"
```

---

## 2. TypeScript Testing with Vitest 4.x

### Browser Mode & Type-Aware Testing

**Setup with TypeScript Support**:

```typescript
// vitest.config.ts - Enterprise TypeScript configuration
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    // Browser mode (stable in v4)
    browser: {
      provider: 'playwright',
      enabled: true,
      instances: [
        { browser: 'chromium' },
        { browser: 'firefox' },
        { browser: 'webkit' }
      ],
    },
    // Coverage configuration
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      lines: 85,
      functions: 85,
      branches: 80,
      statements: 85,
    },
    // Type-aware hooks (v4 feature)
    typecheck: {
      enabled: true,
      only: true,
    },
    // Setup files
    setupFiles: ['./vitest.setup.ts'],
    // Test environment
    environment: 'jsdom',
    // Globals for compatibility
    globals: true,
  },
});
```

**Type-Aware Test Hooks**:

```typescript
// tests/setup.ts - Type-safe beforeEach hook
import { beforeEach, afterEach, describe, it, expect } from 'vitest';

interface TestContext {
  mockFetch: typeof fetch;
  setupDatabase: () => Promise<void>;
  cleanup: () => Promise<void>;
}

describe('Type-aware hooks example', () => {
  let ctx: TestContext;

  beforeEach<TestContext>(async () => {
    ctx = {
      mockFetch: vi.fn(),
      setupDatabase: async () => { /* setup logic */ },
      cleanup: async () => { /* cleanup logic */ },
    };
    await ctx.setupDatabase();
    return ctx; // Type-safe context passing
  });

  afterEach<TestContext>(async (ctx) => {
    await ctx.cleanup();
  });

  it('should handle async operations', async () => {
    const result = await fetchData(ctx.mockFetch);
    expect(result).toBeDefined();
  });
});
```

**Visual Regression Testing (v4 Feature)**:

```typescript
// tests/visual.spec.ts - Visual regression with Playwright
import { test, expect } from 'vitest';
import { preview } from 'vite';

test('visual regression: home page', async ({ page }) => {
  const server = await preview();
  await page.goto(`http://localhost:${server.config.preview?.port || 5173}`);
  
  // Capture screenshot for regression
  await expect(page).toHaveScreenshot('home-page.png', {
    maxDiffPixels: 100, // Allow 100px deviation
  });

  // Trace capture for debugging
  await page.context().tracing.start({ screenshots: true, snapshots: true });
  await page.click('button');
  await page.context().tracing.stop({ path: 'trace.zip' });
});
```

---

## 3. E2E Testing with Playwright 1.48.x

### Multi-Browser Testing & CI/CD Integration

**Playwright Configuration**:

```typescript
// playwright.config.ts - Enterprise multi-browser setup
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  testMatch: '**/*.spec.ts',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  
  reporter: [
    ['html'],
    ['junit', { outputFile: 'junit-results.xml' }],
    ['json', { outputFile: 'test-results.json' }],
  ],

  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },

  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },
    // Mobile device testing
    {
      name: 'Mobile Chrome',
      use: { ...devices['Pixel 5'] },
    },
    {
      name: 'Mobile Safari',
      use: { ...devices['iPhone 12'] },
    },
  ],
});
```

**Test Fixtures & Page Object Pattern**:

```typescript
// e2e/fixtures.ts - Reusable fixtures with page objects
import { test as baseTest } from '@playwright/test';

type TestFixtures = {
  loginPage: LoginPage;
  dashboardPage: DashboardPage;
  apiClient: APIClient;
};

class LoginPage {
  constructor(private page: Page) {}
  
  async login(email: string, password: string) {
    await this.page.fill('[data-testid="email"]', email);
    await this.page.fill('[data-testid="password"]', password);
    await this.page.click('[data-testid="login-btn"]');
    await this.page.waitForURL('/dashboard');
  }
}

class DashboardPage {
  constructor(private page: Page) {}
  
  async navigateToSettings() {
    await this.page.click('[data-testid="settings-link"]');
    await this.page.waitForLoadState('networkidle');
  }
}

export const test = baseTest.extend<TestFixtures>({
  loginPage: async ({ page }, use) => {
    await use(new LoginPage(page));
  },
  dashboardPage: async ({ page }, use) => {
    await use(new DashboardPage(page));
  },
  apiClient: async ({}, use) => {
    await use(new APIClient());
  },
});

// e2e/auth.spec.ts - Usage with fixtures
test('user login flow', async ({ page, loginPage, dashboardPage }) => {
  await page.goto('/login');
  await loginPage.login('user@example.com', 'password123');
  await dashboardPage.navigateToSettings();
});
```

**Multi-Browser Test Execution**:

```bash
# Run tests in all browsers
npx playwright test

# Run specific browser
npx playwright test --project=chromium

# Run with trace viewer
npx playwright test --trace on

# Debug mode with inspector
npx playwright test --debug

# Run tests in headed mode
npx playwright test --headed
```

---

## 4. Testing Library 15.x — Component Testing

### User-Centric Testing with React/Vue/Svelte

**React Testing Library Setup**:

```typescript
// tests/setup.ts - Setup and configuration
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import '@testing-library/jest-dom';

// Custom render with providers
function renderWithProviders(
  ui: React.ReactElement,
  { initialState = {}, ...renderOptions } = {}
) {
  function Wrapper({ children }) {
    return (
      <Provider initialState={initialState}>
        {children}
      </Provider>
    );
  }
  return render(ui, { wrapper: Wrapper, ...renderOptions });
}

export { renderWithProviders, userEvent, screen };
```

**User-Centric Testing Patterns**:

```typescript
// components/LoginForm.test.tsx - DOM query best practices
import { renderWithProviders, screen, userEvent } from '../tests/setup';
import { LoginForm } from './LoginForm';

describe('LoginForm', () => {
  it('should submit form with valid credentials', async () => {
    const user = userEvent.setup();
    renderWithProviders(<LoginForm onSubmit={vi.fn()} />);

    // Query by role (most accessible approach)
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

  it('should display validation errors', async () => {
    const user = userEvent.setup();
    renderWithProviders(<LoginForm onSubmit={vi.fn()} />);

    // Submit without filling form
    await user.click(screen.getByRole('button', { name: /submit/i }));

    // Check for error messages
    expect(screen.getByText(/email is required/i)).toBeInTheDocument();
  });
});
```

---

## 5. API Testing with httpx & Pytest

### Async HTTP Testing

```python
# tests/test_api.py - Async API testing with httpx
import pytest
import httpx

@pytest.mark.asyncio
async def test_api_endpoint():
    """Test async HTTP client with pytest-asyncio."""
    async with httpx.AsyncClient() as client:
        response = await client.get(
            "http://localhost:8000/api/users",
            timeout=10.0
        )
        assert response.status_code == 200
        data = response.json()
        assert len(data) > 0

# Using pytest_httpx for mocking
@pytest.fixture
def httpx_mock(respx_mock):
    """Fixture for mocking HTTP responses."""
    return respx_mock

@pytest.mark.asyncio
async def test_api_with_mock(httpx_mock):
    """Test with mocked HTTP responses."""
    httpx_mock.get(
        "http://api.example.com/users/1"
    ).mock(return_value=httpx.Response(
        200,
        json={"id": 1, "name": "John"}
    ))

    async with httpx.AsyncClient() as client:
        response = await client.get("http://api.example.com/users/1")
        assert response.json() == {"id": 1, "name": "John"}
```

---

## 6. Test Data Factories

### polyfactory & pytest-factoryboy

**Polyfactory Pattern**:

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

# Custom field values
user = UserFactory.create(email="custom@example.com")
```

**Factory Boy Pattern**:

```python
# tests/factories.py - factory-boy for Django/SQLAlchemy
import factory
from faker import Faker

fake = Faker()

class UserFactory(factory.Factory):
    class Meta:
        model = User

    id = factory.Sequence(lambda n: n)
    email = factory.LazyFunction(fake.email)
    name = factory.LazyFunction(fake.name)
    is_active = True

    @factory.post_generation
    def permissions(obj, create, extracted):
        if not create:
            return
        if extracted:
            for permission in extracted:
                obj.permissions.add(permission)

# Create instances
user = UserFactory()  # Single instance
users = UserFactory.create_batch(5)  # Batch creation
```

---

## 7. Test Coverage Strategy

### Coverage.py Configuration

```python
# .coveragerc - Coverage.py configuration
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
    @abstract

precision = 2
show_missing = True
skip_covered = False

[html]
directory = htmlcov

[paths]
source =
    src
    */site-packages
```

**Coverage Commands**:

```bash
# Run tests with coverage
pytest --cov=src --cov-report=html --cov-report=term

# Show missing lines
coverage report --skip-covered

# HTML report
coverage html  # Open htmlcov/index.html

# Coverage badges
coverage-badge -o coverage.svg
```

---

## 8. Mock & Stub Patterns

### pytest-mock Integration

```python
# tests/test_external_services.py - Mocking with pytest-mock
def test_email_sending(mocker):
    """Mock external email service."""
    # Patch the email send method
    mock_send = mocker.patch('app.services.EmailService.send')
    mock_send.return_value = True

    # Execute code that calls the service
    send_notification(user_id=1, message="Hello")

    # Assert the mock was called correctly
    mock_send.assert_called_once_with(
        to="user@example.com",
        subject="Notification",
        body="Hello"
    )

def test_database_transaction(mocker):
    """Mock database operations."""
    mock_db = mocker.MagicMock()
    mock_db.query.return_value.filter.return_value.all.return_value = [
        {"id": 1, "name": "User 1"},
        {"id": 2, "name": "User 2"},
    ]

    result = get_users(db=mock_db)
    assert len(result) == 2
```

### JavaScript Mocking with sinon.js

```typescript
// test/mock.test.ts - sinon.js for spies, stubs, mocks
import sinon from 'sinon';

describe('Mocking with sinon', () => {
  it('should stub method calls', () => {
    const user = {
      setName(name: string) {
        this.name = name;
      },
    };

    const stub = sinon.stub(user, 'setName');
    user.setName('John');

    expect(stub.calledOnce).toBe(true);
    expect(stub.calledWith('John')).toBe(true);

    stub.restore();
  });

  it('should track function calls with spies', () => {
    const api = {
      fetch: (url: string) => Promise.resolve({ ok: true }),
    };

    const spy = sinon.spy(api, 'fetch');
    api.fetch('/api/users');

    expect(spy.calledOnce).toBe(true);
    expect(spy.calledWith('/api/users')).toBe(true);

    spy.restore();
  });
});
```

---

## 9. Performance & Load Testing

### k6 Load Testing (v1.0+)

```javascript
// tests/load.js - k6 load testing script
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 20 },   // Ramp up to 20 VUs
    { duration: '1m30s', target: 20 }, // Stay at 20 VUs
    { duration: '30s', target: 0 },    // Ramp down to 0 VUs
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'], // 95% under 500ms
    http_req_failed: ['rate<0.1'],                  // Less than 10% failures
  },
};

export default function () {
  // GET request
  const getResponse = http.get('http://localhost:3000/api/users');
  check(getResponse, {
    'GET status 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });

  // POST request
  const payload = JSON.stringify({
    name: 'Test User',
    email: 'test@example.com',
  });
  
  const postResponse = http.post(
    'http://localhost:3000/api/users',
    payload,
    {
      headers: { 'Content-Type': 'application/json' },
    }
  );

  check(postResponse, {
    'POST status 201': (r) => r.status === 201,
  });

  sleep(1); // Think time between requests
}
```

**Running k6 Tests**:

```bash
# Basic run
k6 run tests/load.js

# With output to JSON
k6 run tests/load.js -o json=results.json

# Grafana Cloud integration
k6 run tests/load.js -o cloud
```

---

## 10. Accessibility Testing with axe-core

### WCAG 2.1 Compliance Automation

```typescript
// tests/accessibility.spec.ts - Accessibility testing with Playwright
import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

test('accessibility audit: homepage', async ({ page }) => {
  await page.goto('http://localhost:3000');

  // Run axe accessibility audit
  const accessibilityScanResults = await new AxeBuilder({ page })
    .withTags(['wcag2aa', 'wcag2aaa'])
    .withStandards(['WCAG 2.1 Level AA'])
    .analyze();

  // Check for violations
  expect(accessibilityScanResults.violations).toEqual([]);

  // Log results for analysis
  console.log('Axe Results:', {
    violations: accessibilityScanResults.violations,
    passes: accessibilityScanResults.passes.length,
    inapplicable: accessibilityScanResults.inapplicable.length,
  });
});

test('accessibility: form labels and ARIA', async ({ page }) => {
  await page.goto('http://localhost:3000/form');

  // Check form label associations
  const inputs = await page.locator('input');
  for (const input of await inputs.all()) {
    const ariaLabel = await input.getAttribute('aria-label');
    const id = await input.getAttribute('id');
    
    // Either aria-label or associated label required
    if (!ariaLabel && id) {
      const label = await page.locator(`label[for="${id}"]`);
      expect(label).toBeDefined();
    }
  }

  // Check for color contrast
  const buttons = await page.locator('button');
  for (const button of await buttons.all()) {
    const computedStyle = await button.evaluate((el) => {
      const style = window.getComputedStyle(el);
      return {
        color: style.color,
        backgroundColor: style.backgroundColor,
      };
    });
    // Validate contrast ratio (requires additional color contrast library)
  }
});
```

---

## 11. Async Test Handling

### pytest-asyncio & JavaScript async/await

**Python Async Tests**:

```python
# tests/test_async.py - pytest-asyncio for async functions
import pytest
import asyncio

@pytest.mark.asyncio
async def test_async_function():
    """Test async functions with pytest-asyncio."""
    result = await async_operation()
    assert result == expected_value

@pytest.mark.asyncio
async def test_concurrent_operations():
    """Test multiple concurrent async operations."""
    results = await asyncio.gather(
        fetch_users(),
        fetch_posts(),
        fetch_comments(),
    )
    assert len(results) == 3
    assert all(r for r in results)

# Auto-use async fixture
@pytest.fixture
async def async_client():
    """Async client fixture."""
    client = AsyncAPIClient()
    await client.connect()
    yield client
    await client.disconnect()
```

**JavaScript Async Tests**:

```typescript
// test/async.test.ts - Async test patterns
import { describe, it, expect, beforeEach, afterEach } from 'vitest';

describe('Async operations', () => {
  it('should handle promises', async () => {
    const result = await fetchData();
    expect(result).toBeDefined();
  });

  it('should handle async/await', async () => {
    const user = await getUserById(1);
    expect(user.id).toBe(1);
  });

  it('should handle concurrent operations', async () => {
    const [users, posts, comments] = await Promise.all([
      fetchUsers(),
      fetchPosts(),
      fetchComments(),
    ]);

    expect(users).toHaveLength(10);
    expect(posts).toHaveLength(50);
  });

  it('should timeout on slow operations', async () => {
    expect.assertions(1);
    try {
      await slowOperation({ timeout: 100 });
    } catch (error) {
      expect(error).toMatch('timeout');
    }
  });
});
```

---

## 12. Database Seeding for Tests

### Test Data Setup Strategies

```python
# tests/fixtures.py - Database seeding fixtures
import pytest
from app.models import User, Post

@pytest.fixture
def seed_users(db_session):
    """Seed database with test users."""
    users = [
        User(name="Admin", email="admin@example.com", role="admin"),
        User(name="User1", email="user1@example.com", role="user"),
        User(name="User2", email="user2@example.com", role="user"),
    ]
    db_session.add_all(users)
    db_session.commit()
    return users

@pytest.fixture
def seed_posts(db_session, seed_users):
    """Seed posts for users (depends on users fixture)."""
    posts = [
        Post(title="Post 1", author_id=seed_users[0].id),
        Post(title="Post 2", author_id=seed_users[1].id),
    ]
    db_session.add_all(posts)
    db_session.commit()
    return posts

def test_user_posts(db_session, seed_posts):
    """Test with seeded data."""
    user = db_session.query(User).filter_by(name="Admin").first()
    posts = user.posts
    assert len(posts) == 1
```

---

## Enterprise Testing Strategy

### Testing Pyramid Approach

```
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

### Test Execution Strategy

```bash
# Local development: Run unit tests on save
pytest --watch

# Pre-commit: Unit + fast integration tests
pytest tests/unit tests/integration -k "not slow"

# CI pipeline:
# 1. Unit tests (pytest + Vitest)
pytest tests/unit --cov --cov-fail-under=85
npm run test:unit

# 2. Integration tests
pytest tests/integration

# 3. E2E tests (parallel across browsers)
npx playwright test
npx cypress run

# 4. Performance baselines
k6 run tests/load.js

# 5. Accessibility audit
npm run test:a11y
```

---

## Best Practices Checklist

- [x] **Test Isolation**: Each test runs independently (no shared state)
- [x] **Fixture Scoping**: Use appropriate fixture scopes (function/module/session)
- [x] **Async Handling**: Proper async/await in tests and fixtures
- [x] **Mock Strategy**: Mock external services, not internal logic
- [x] **Coverage Targets**: Maintain ≥85% code coverage
- [x] **Test Names**: Descriptive test names indicating behavior
- [x] **Parametrization**: Use parametrized tests for multiple scenarios
- [x] **CI/CD Integration**: Tests run in automated pipelines
- [x] **Performance**: Tests complete in <30s (unit), <5m (integration), <10m (E2E)
- [x] **Accessibility**: WCAG 2.1 compliance automated in E2E tests

---

## When to Use This Skill

**Activate for requests involving**:
- ✅ Unit testing architecture (pytest, Vitest, Jest)
- ✅ E2E automation (Playwright, Cypress)
- ✅ API testing (httpx, Supertest)
- ✅ Component testing (Testing Library)
- ✅ Test coverage strategy & measurement
- ✅ Mock/stub patterns and test doubles
- ✅ Performance & load testing (k6, Locust)
- ✅ Accessibility testing (axe-core, pa11y)
- ✅ Test data factories & fixtures
- ✅ CI/CD test automation
- ✅ Async test patterns
- ✅ Production testing strategy

