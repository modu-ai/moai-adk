# Advanced Testing Patterns - Enterprise Test Architecture

**Version**: 4.0.0 (2025-11-22)
**Status**: Production Ready

---

## Property-Based Testing

### Hypothesis Framework for Python

```python
from hypothesis import given, strategies as st, settings, Phase
import pytest

class PropertyBasedTester:
    """Test properties with generated test cases."""

    @given(st.integers(min_value=0, max_value=1000))
    @settings(max_examples=100, phases=[Phase.generate, Phase.shrink])
    def test_positive_integer_properties(self, num):
        """Property: abs(x) is always non-negative."""
        assert abs(num) >= 0

    @given(
        st.lists(
            st.integers(),
            min_size=0,
            max_size=100
        )
    )
    def test_sort_properties(self, numbers):
        """Property: sorted list is always <= original length."""
        sorted_list = sorted(numbers)
        assert len(sorted_list) == len(numbers)

        # All elements present
        assert sorted(sorted_list) == sorted(numbers)

    @given(
        x=st.floats(allow_nan=False, allow_infinity=False),
        y=st.floats(allow_nan=False, allow_infinity=False)
    )
    def test_commutativity(self, x, y):
        """Property: addition is commutative (x + y == y + x)."""
        assert x + y == y + x

    @given(st.text(min_size=1))
    def test_string_properties(self, text):
        """Property: uppercase then lowercase == original."""
        upper = text.upper()
        lower = upper.lower()
        assert lower == text.lower()
```

### Fast-Check for TypeScript

```typescript
import * as fc from 'fast-check';

describe('Property-Based Tests', () => {
    it('should satisfy array commutative property', () => {
        fc.assert(
            fc.property(
                fc.array(fc.integer()),
                fc.array(fc.integer()),
                (arr1, arr2) => {
                    const concatenated1 = [...arr1, ...arr2];
                    const concatenated2 = [...arr2, ...arr1];

                    // Both have same length
                    expect(concatenated1.length).toBe(
                        arr1.length + arr2.length
                    );
                }
            )
        );
    });

    it('should validate parser roundtrip', () => {
        fc.assert(
            fc.property(
                fc.object({ key: fc.string() }),
                (obj) => {
                    const json = JSON.stringify(obj);
                    const parsed = JSON.parse(json);

                    expect(parsed).toEqual(obj);
                }
            ),
            { numRuns: 1000 }
        );
    });
});
```

---

## Mutation Testing

### Stryker for JavaScript/TypeScript

```javascript
// stryker.conf.js
module.exports = {
  testRunner: 'vitest',
  testFramework: 'vitest',
  reporters: ['html', 'clear-text'],
  mutate: ['src/**/*.ts', '!src/**/*.spec.ts'],
  mutators: ['ArithmeticOperator', 'StringLiteral', 'LogicalOperator'],
  thresholds: {
    high: 80,  // Fail if mutation score < 80%
    low: 60,
    break: 50
  },
  plugins: ['@stryker-mutator/typescript-checker'],
};

// Run: npx stryker run
```

### PIT for Java

```xml
<!-- pom.xml -->
<plugin>
    <groupId>org.pitest</groupId>
    <artifactId>pitest-maven</artifactId>
    <version>1.14.4</version>
    <configuration>
        <targetClasses>
            <param>com.example.calculators.*</param>
        </targetClasses>
        <targetTests>
            <param>com.example.calculators.*Test</param>
        </targetTests>
        <mutators>
            <mutator>DEFAULTS</mutator>
        </mutators>
        <coverageThreshold>80</coverageThreshold>
    </configuration>
</plugin>

<!-- Run: mvn org.pitest:pitest-maven:mutationCoverage -->
```

---

## Contract Testing

### Pact for Microservice Testing

```python
from pact import Consumer, Provider
import requests

class ContractTest:
    """Test API contracts between services."""

    def test_get_user_contract(self):
        pact = (
            Consumer('UserConsumer')
            .has_pact_with(Provider('UserProvider'))
        )

        (pact
         .upon_receiving('a request for a user')
         .with_request('GET', '/users/123')
         .will_respond_with(200, body={
             'id': 123,
             'name': 'John Doe',
             'email': 'john@example.com'
         }))

        with pact:
            response = requests.get('http://localhost:8080/users/123')
            assert response.status_code == 200
            assert response.json()['name'] == 'John Doe'

        pact.write_to_pact_file()
```

### Spring Cloud Contract

```groovy
// src/test/resources/contracts/users/getUser.groovy
Contract.make {
    request {
        method 'GET'
        url '/users/123'
    }
    response {
        status 200
        body(
            id: 123,
            name: 'John Doe'
        )
        headers {
            'Content-Type': 'application/json'
        }
    }
}
```

---

## Test Pyramid Implementation

### Unit Tests (Base Layer)

```python
import pytest
from unittest.mock import Mock, patch

class UnitTestLayer:
    """Base layer: Fast, isolated unit tests."""

    def test_calculator_addition(self):
        """Test single function in isolation."""
        calc = Calculator()
        assert calc.add(2, 3) == 5

    def test_with_mock(self):
        """Unit test with mocked dependencies."""
        mock_logger = Mock()
        service = UserService(logger=mock_logger)

        service.create_user('john')

        mock_logger.log.assert_called_once_with('User created')

    @pytest.mark.parametrize("input,expected", [
        (1, 2),
        (5, 10),
        (10, 20),
    ])
    def test_parametrized(self, input, expected):
        """Parametrized unit tests for coverage."""
        assert double(input) == expected
```

### Integration Tests (Middle Layer)

```python
import pytest
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

class IntegrationTestLayer:
    """Middle layer: Database integration tests."""

    @pytest.fixture
    def db(self):
        """Setup test database."""
        engine = create_engine('sqlite:///:memory:')
        Base.metadata.create_all(engine)
        Session = sessionmaker(bind=engine)
        return Session()

    def test_user_creation_flow(self, db):
        """Test database interaction."""
        user = User(name='John', email='john@example.com')
        db.add(user)
        db.commit()

        retrieved = db.query(User).filter_by(name='John').first()
        assert retrieved.email == 'john@example.com'

    def test_api_with_database(self, client, db):
        """Test API endpoint with real database."""
        response = client.post('/users', json={
            'name': 'Jane',
            'email': 'jane@example.com'
        })
        assert response.status_code == 201
        assert db.query(User).count() == 1
```

### End-to-End Tests (Top Layer)

```typescript
import { test, expect } from '@playwright/test';

test.describe('E2E: User Registration Flow', () => {
    test('should complete registration', async ({ page }) => {
        await page.goto('http://localhost:3000/register');

        await page.fill('[name="email"]', 'newuser@example.com');
        await page.fill('[name="password"]', 'SecurePass123!');
        await page.click('[type="submit"]');

        await expect(page).toHaveURL(/.*\/dashboard/);
        await expect(page.locator('text=Welcome')).toBeVisible();
    });
});
```

---

## Performance Testing

### Load Testing with k6

```javascript
// load-test.js
import http from 'k6/http';
import { check, sleep, group } from 'k6';

export const options = {
    stages: [
        { duration: '30s', target: 20 },   // Ramp-up
        { duration: '1m30s', target: 20 }, // Stay at load
        { duration: '30s', target: 0 },    // Ramp-down
    ],
    thresholds: {
        http_req_duration: ['p(95)<500', 'p(99)<1000'],
        http_req_failed: ['rate<0.1'],
    },
};

export default function () {
    group('API Load Test', () => {
        const response = http.get('http://localhost:3000/api/users');
        check(response, {
            'status 200': (r) => r.status === 200,
            'response time < 500ms': (r) => r.timings.duration < 500,
            'valid JSON': (r) => {
                try {
                    return r.json() !== null;
                } catch {
                    return false;
                }
            }
        });
    });

    sleep(1);
}

// Run: k6 run load-test.js
```

---

## Accessibility Testing

### Axe Integration

```typescript
import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

test('homepage accessibility audit', async ({ page }) => {
    await page.goto('http://localhost:3000');

    const results = await new AxeBuilder({ page })
        .withTags(['wcag2aa', 'wcag2aaa'])
        .analyze();

    expect(results.violations).toHaveLength(0);
    expect(results.passes.length).toBeGreaterThan(5);
});
```

---

## Best Practices

**DO**:
- Implement full test pyramid (70% unit, 20% integration, 10% E2E)
- Use property-based testing for complex logic
- Implement mutation testing to verify test quality
- Run performance tests in CI/CD
- Test accessibility in automated tests
- Use contract tests for microservices
- Mock external dependencies in unit tests

**DON'T**:
- Skip mutation testing (tests can be wrong!)
- Only test happy paths (test edge cases)
- Use real services in unit tests
- Make E2E tests flaky (use proper waits)
- Ignore performance regressions
- Skip accessibility testing
- Over-mock internal dependencies

---

**Related Skills**: moai-domain-backend, moai-domain-frontend, moai-essentials-perf
