# Code Examples: Enterprise Testing Patterns ( )

## Example 1: Complete pytest Test Suite with Fixtures

```python
# conftest.py - Shared test configuration
import pytest
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker, Session
from app.models import Base, User

@pytest.fixture(scope="session")
def db_engine():
    """Create test database engine for entire session."""
    engine = create_engine("sqlite:///:memory:")
    Base.metadata.create_all(engine)
    yield engine
    engine.dispose()

@pytest.fixture
def db_session(db_engine) -> Session:
    """Create isolated database session for each test."""
    connection = db_engine.connect()
    transaction = connection.begin()
    session = sessionmaker(bind=connection)()
    
    yield session
    
    session.close()
    transaction.rollback()
    connection.close()

@pytest.fixture
def user_factory():
    """Factory fixture for creating test users."""
    def _create_user(name: str = "Test", email: str = "test@example.com", **kwargs):
        return User(name=name, email=email, is_active=True, **kwargs)
    return _create_user

# tests/test_user_service.py
import pytest
from app.services import UserService

@pytest.mark.asyncio
class TestUserService:
    """Test suite for UserService with parametrization."""
    
    async def test_create_user(self, db_session, user_factory):
        """Test user creation."""
        user = user_factory(name="John")
        service = UserService(db_session)
        
        created_user = await service.create(user)
        
        assert created_user.id is not None
        assert created_user.name == "John"
    
    @pytest.mark.parametrize("email,is_valid", [
        ("valid@example.com", True),
        ("invalid-email", False),
        ("another@test.org", True),
    ])
    async def test_email_validation(self, db_session, email: str, is_valid: bool):
        """Test email validation with parametrized values."""
        service = UserService(db_session)
        result = await service.validate_email(email)
        assert result == is_valid
```

---

## Example 2: Vitest TypeScript Testing with Type-Aware Hooks

```typescript
// vitest.config.ts
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    browser: {
      provider: 'playwright',
      instances: [
        { browser: 'chromium' },
        { browser: 'firefox' }
      ],
    },
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html'],
      lines: 85,
      functions: 85,
      branches: 80,
    },
    environment: 'jsdom',
    globals: true,
  },
});

// components/Button.test.tsx
import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Button } from './Button';

interface TestContext {
  user: typeof userEvent;
  mockOnClick: (callback: () => void) => void;
}

describe('Button Component', () => {
  let ctx: TestContext;
  const mockFn = vi.fn();

  beforeEach<TestContext>(async () => {
    ctx = {
      user: userEvent.setup(),
      mockOnClick: (callback) => mockFn.mockImplementation(callback),
    };
    return ctx;
  });

  it('should render button text', () => {
    render(<Button label="Click me" onClick={mockFn} />);
    expect(screen.getByRole('button', { name: 'Click me' })).toBeInTheDocument();
  });

  it('should call onClick when clicked', async () => {
    const { user } = ctx;
    render(<Button label="Click" onClick={mockFn} />);
    
    await user.click(screen.getByRole('button'));
    expect(mockFn).toHaveBeenCalledOnce();
  });

  it('should be disabled when disabled prop is true', () => {
    render(<Button label="Click" disabled={true} onClick={mockFn} />);
    expect(screen.getByRole('button')).toBeDisabled();
  });
});
```

---

## Example 3: Playwright E2E Testing with Page Objects

```typescript
// e2e/pages/LoginPage.ts - Page Object Model
import { Page, expect } from '@playwright/test';

export class LoginPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto('/login');
    await this.page.waitForLoadState('networkidle');
  }

  async login(email: string, password: string) {
    await this.page.fill('input[name="email"]', email);
    await this.page.fill('input[name="password"]', password);
    await this.page.click('button[type="submit"]');
    await this.page.waitForURL('/dashboard');
  }

  async getErrorMessage(): Promise<string | null> {
    return this.page.textContent('[data-testid="error-message"]');
  }

  async waitForLoginForm() {
    await this.page.waitForSelector('form');
  }
}

// e2e/auth.spec.ts - Test using Page Object
import { test, expect } from '@playwright/test';
import { LoginPage } from './pages/LoginPage';

test.describe('Authentication Flow', () => {
  let loginPage: LoginPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    await loginPage.goto();
  });

  test('should login with valid credentials', async ({ page }) => {
    await loginPage.login('user@example.com', 'password123');
    
    // Verify redirection
    expect(page.url()).toContain('/dashboard');
  });

  test('should show error with invalid email', async () => {
    await loginPage.login('invalid-email', 'password123');
    
    const errorMsg = await loginPage.getErrorMessage();
    expect(errorMsg).toContain('Invalid email');
  });

  test('should be accessible (WCAG 2.1 AA)', async ({ page }) => {
    const accessibilityScan = await new AxeBuilder({ page })
      .withTags(['wcag2aa'])
      .analyze();
    
    expect(accessibilityScan.violations).toEqual([]);
  });
});

// Run with: npx playwright test --project=chromium --headed
```

---

## Example 4: Testing Library — User-Centric Component Testing

```typescript
// components/LoginForm.test.tsx - User-centric testing
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { LoginForm } from './LoginForm';

describe('LoginForm - User-Centric Testing', () => {
  it('should log in successfully', async () => {
    const user = userEvent.setup();
    const mockOnSuccess = vi.fn();

    render(<LoginForm onSuccess={mockOnSuccess} />);

    // Use accessible queries (role-based is most accessible)
    const emailInput = screen.getByRole('textbox', { name: /email/i });
    const passwordInput = screen.getByLabelText(/password/i);
    const submitButton = screen.getByRole('button', { name: /sign in/i });

    // User interaction
    await user.type(emailInput, 'user@example.com');
    await user.type(passwordInput, 'MyPassword123!');
    await user.click(submitButton);

    // Verification
    expect(mockOnSuccess).toHaveBeenCalledWith({
      email: 'user@example.com',
    });

    // Wait for success message
    const successMessage = await screen.findByText(/welcome/i);
    expect(successMessage).toBeInTheDocument();
  });

  it('should show validation errors', async () => {
    const user = userEvent.setup();
    render(<LoginForm onSuccess={vi.fn()} />);

    const submitButton = screen.getByRole('button', { name: /sign in/i });
    await user.click(submitButton);

    // Check for specific error messages
    expect(screen.getByText(/email is required/i)).toBeInTheDocument();
    expect(screen.getByText(/password is required/i)).toBeInTheDocument();
  });
});
```

---

## Example 5: API Testing with httpx and pytest

```python
# tests/test_api.py - Async API testing
import pytest
import httpx
from app.api import create_app

@pytest.fixture
async def api_client():
    """Create async API client for testing."""
    async with httpx.AsyncClient(app=create_app()) as client:
        yield client

@pytest.mark.asyncio
class TestUserAPI:
    """API tests with async HTTP client."""
    
    async def test_get_users(self, api_client: httpx.AsyncClient):
        """Test GET /api/users endpoint."""
        response = await api_client.get('/api/users')
        
        assert response.status_code == 200
        data = response.json()
        assert isinstance(data, list)
    
    async def test_create_user(self, api_client: httpx.AsyncClient):
        """Test POST /api/users with request body."""
        payload = {
            'name': 'John Doe',
            'email': 'john@example.com',
        }
        
        response = await api_client.post('/api/users', json=payload)
        
        assert response.status_code == 201
        created_user = response.json()
        assert created_user['id'] is not None
        assert created_user['email'] == payload['email']
    
    async def test_api_with_authentication(self, api_client: httpx.AsyncClient):
        """Test API with bearer token authentication."""
        token = 'test-token-12345'
        headers = {'Authorization': f'Bearer {token}'}
        
        response = await api_client.get('/api/protected', headers=headers)
        
        assert response.status_code == 200

# Using pytest-httpx for mocking
@pytest.mark.asyncio
async def test_api_with_mocked_response(respx_mock):
    """Test with mocked HTTP responses."""
    respx_mock.get('https://api.example.com/users').mock(
        return_value=httpx.Response(
            200,
            json=[
                {'id': 1, 'name': 'User 1'},
                {'id': 2, 'name': 'User 2'},
            ]
        )
    )

    async with httpx.AsyncClient() as client:
        response = await client.get('https://api.example.com/users')
        assert len(response.json()) == 2
```

---

## Example 6: Test Data Factories with polyfactory

```python
# tests/factories.py - polyfactory for test data generation
from polyfactory.factories.pydantic_factory import ModelFactory
from pydantic import BaseModel, EmailStr

class UserSchema(BaseModel):
    id: int
    name: str
    email: EmailStr
    is_active: bool

class UserFactory(ModelFactory[UserSchema]):
    """Factory for generating test User data."""
    __model__ = UserSchema

# Usage
user = UserFactory.create()  # Generate single instance
users = UserFactory.batch(5)  # Generate batch of 5

# Customized generation
user = UserFactory.create(
    name="John Doe",
    email="john@example.com",
    is_active=True
)

# In tests
def test_user_creation():
    """Test with factory-generated data."""
    user = UserFactory.create()
    assert user.email is not None
    assert user.is_active is True
```

---

## Example 7: Complete E2E Test with Multi-Browser Support

```typescript
// playwright.config.ts - Enterprise configuration
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,

  reporter: [
    ['html', { outputFolder: 'playwright-report' }],
    ['junit', { outputFile: 'junit-results.xml' }],
  ],

  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
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
  ],

  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
  },
});

// e2e/checkout.spec.ts
import { test, expect } from '@playwright/test';

test.describe('E-commerce Checkout Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/login');
    await page.fill('input[name="email"]', 'user@example.com');
    await page.fill('input[name="password"]', 'password');
    await page.click('button:has-text("Sign In")');
    await page.waitForURL('/products');
  });

  test('should complete checkout in all browsers', async ({ page, browserName }) => {
    // Add product to cart
    await page.click('text=Add to Cart');
    
    // Verify cart updated
    const cartBadge = page.locator('[data-testid="cart-count"]');
    await expect(cartBadge).toContainText('1');

    // Navigate to checkout
    await page.click('button:has-text("Checkout")');
    await page.waitForURL('/checkout');

    // Fill shipping address
    await page.fill('input[name="address"]', '123 Main St');
    await page.fill('input[name="city"]', 'New York');

    // Submit order
    await page.click('button:has-text("Place Order")');

    // Verify success
    await expect(page.locator('text=Order Confirmed')).toBeVisible();
    
    console.log(`Checkout completed on ${browserName}`);
  });

  test('should validate form with accessibility', async ({ page }) => {
    await page.goto('/checkout');

    // Check for accessible form elements
    const form = page.locator('form[aria-label="Checkout Form"]');
    expect(form).toBeDefined();

    // Verify labels are associated
    const inputs = await page.locator('input').all();
    for (const input of inputs) {
      const ariaLabel = await input.getAttribute('aria-label');
      const id = await input.getAttribute('id');
      
      // Either aria-label or associated label required
      expect(ariaLabel || id).toBeTruthy();
    }
  });
});
```

---

## Example 8: Accessibility Testing with axe-core

```typescript
// e2e/accessibility.spec.ts - WCAG 2.1 compliance
import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

test.describe('Accessibility Testing', () => {
  test('homepage should comply with WCAG 2.1 Level AA', async ({ page }) => {
    await page.goto('/');

    // Run axe accessibility audit
    const results = await new AxeBuilder({ page })
      .withTags(['wcag2aa'])
      .analyze();

    // Assert no violations
    expect(results.violations).toHaveLength(0);

    // Log results for analysis
    console.log(`Accessibility Audit Results:
      - Passes: ${results.passes.length}
      - Violations: ${results.violations.length}
      - Incomplete: ${results.incomplete.length}
    `);
  });

  test('should have proper heading hierarchy', async ({ page }) => {
    await page.goto('/');

    // Check for h1 (exactly one required)
    const h1s = await page.locator('h1').count();
    expect(h1s).toBe(1);

    // Verify heading order
    const headings = await page.locator('h1, h2, h3, h4, h5, h6').all();
    for (let i = 1; i < headings.length; i++) {
      const prevLevel = parseInt(await headings[i - 1].evaluate((el) => el.tagName[1]));
      const currLevel = parseInt(await headings[i].evaluate((el) => el.tagName[1]));
      
      expect(currLevel - prevLevel).toBeLessThanOrEqual(1);
    }
  });

  test('should have sufficient color contrast', async ({ page }) => {
    await page.goto('/');

    // Check button contrast
    const buttons = await page.locator('button').all();
    for (const button of buttons) {
      const styles = await button.evaluate((el) => {
        const computed = window.getComputedStyle(el);
        return {
          color: computed.color,
          backgroundColor: computed.backgroundColor,
        };
      });
      // Requires additional library for contrast calculation
      // https://www.npmjs.com/package/polished
    }
  });
});
```

---

## Example 9: Load Testing with k6

```javascript
// tests/load.js - k6 load testing script
import http from 'k6/http';
import { check, sleep, group } from 'k6';

export const options = {
  stages: [
    { duration: '1m', target: 50 },    // Ramp to 50 users
    { duration: '3m', target: 50 },    // Stay at 50
    { duration: '1m', target: 100 },   // Ramp to 100
    { duration: '3m', target: 100 },   // Stay at 100
    { duration: '1m', target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed: ['rate<0.1'],
    group_duration: ['p(95)<800'],
  },
};

export default function () {
  const baseURL = 'http://localhost:3000/api';

  group('User CRUD Operations', () => {
    // GET /users
    let getResp = http.get(`${baseURL}/users`);
    check(getResp, {
      'GET /users status 200': (r) => r.status === 200,
      'GET /users duration < 500ms': (r) => r.timings.duration < 500,
    });

    // POST /users (create)
    const createPayload = JSON.stringify({
      name: `User-${__VU}-${__ITER}`,
      email: `user-${__VU}-${__ITER}@example.com`,
    });

    let createResp = http.post(`${baseURL}/users`, createPayload, {
      headers: { 'Content-Type': 'application/json' },
    });

    check(createResp, {
      'POST /users status 201': (r) => r.status === 201,
    });

    const userId = createResp.json('id');

    // GET /users/{id}
    let getOneResp = http.get(`${baseURL}/users/${userId}`);
    check(getOneResp, {
      'GET /users/:id status 200': (r) => r.status === 200,
    });

    sleep(1); // Think time
  });
}

// Run: k6 run tests/load.js
// With Grafana Cloud: k6 run tests/load.js -o cloud
```

---

## Example 10: Mock & Stub Patterns with pytest-mock

```python
# tests/test_services.py - Mocking with pytest-mock
import pytest
from unittest.mock import MagicMock
from app.services import UserService, EmailService

def test_user_registration_with_mock(mocker):
    """Test user registration with mocked email service."""
    # Mock the email service
    mock_email = mocker.patch.object(EmailService, 'send_verification')
    mock_email.return_value = True

    service = UserService()
    user = service.register('john@example.com', 'password123')

    # Assertions
    assert user.is_active is False  # Not active until verified
    mock_email.assert_called_once()

    # Check call arguments
    call_args = mock_email.call_args
    assert 'john@example.com' in call_args[0]

def test_user_service_with_mocked_db(mocker):
    """Test service with mocked database."""
    mock_db = mocker.MagicMock()
    mock_db.query.return_value.filter.return_value.first.return_value = {
        'id': 1,
        'email': 'john@example.com',
        'name': 'John',
    }

    service = UserService(db=mock_db)
    user = service.get_user_by_email('john@example.com')

    assert user['email'] == 'john@example.com'
    mock_db.query.assert_called_once()

def test_retry_logic_with_mock(mocker):
    """Test retry logic using mock side effects."""
    mock_api = mocker.MagicMock()
    
    # Simulate failures then success
    mock_api.fetch.side_effect = [
        TimeoutError('Timeout'),
        TimeoutError('Timeout'),
        {'status': 'ok'},  # Success on third try
    ]

    service = UserService(api=mock_api)
    result = service.fetch_with_retry('endpoint', retries=3)

    assert result['status'] == 'ok'
    assert mock_api.fetch.call_count == 3
```

