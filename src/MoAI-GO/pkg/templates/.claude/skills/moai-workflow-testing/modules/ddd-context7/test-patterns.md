# Context7 Test Patterns and Best Practices

> Module: Testing patterns, Context7 integration, and industry best practices
> Complexity: Advanced
> Time: 15+ minutes
> Dependencies: Python 3.8+, pytest, Context7 MCP, unittest.mock

## Context7 DDD Integration

```python
class Context7DDDIntegration:
    """Integration with Context7 for DDD patterns and best practices."""

    def __init__(self, context7_client=None):
        self.context7 = context7_client
        self.pattern_cache = {}

    async def load_ddd_patterns(self, language: str = "python") -> Dict[str, Any]:
        """Load DDD patterns and best practices from Context7."""

        cache_key = f"ddd_patterns_{language}"
        if cache_key in self.pattern_cache:
            return self.pattern_cache[cache_key]

        patterns = {}

        if self.context7:
            try:
                # Load DDD best practices
                ddd_patterns = await self.context7.get_library_docs(
                    context7_library_id="/testing/pytest",
                    topic="DDD ANALYZE-PRESERVE-IMPROVE patterns best practices 2025",
                    tokens=4000
                )
                patterns['ddd_best_practices'] = ddd_patterns

                # Load test patterns for specific language
                if language == "python":
                    python_patterns = await self.context7.get_library_docs(
                        context7_library_id="/python/pytest",
                        topic="advanced testing patterns mocking fixtures 2025",
                        tokens=3000
                    )
                    patterns['python_testing'] = python_patterns

                # Load assertion patterns
                assertion_patterns = await self.context7.get_library_docs(
                    context7_library_id="/testing/assertions",
                    topic="assertion patterns error messages test design 2025",
                    tokens=2000
                )
                patterns['assertions'] = assertion_patterns

                # Load mocking patterns
                mocking_patterns = await self.context7.get_library_docs(
                    context7_library_id="/python/unittest-mock",
                    topic="mocking strategies test doubles isolation patterns 2025",
                    tokens=3000
                )
                patterns['mocking'] = mocking_patterns

            except Exception as e:
                print(f"Failed to load Context7 patterns: {e}")
                patterns = self._get_default_patterns()
        else:
            patterns = self._get_default_patterns()

        self.pattern_cache[cache_key] = patterns
        return patterns

    def _get_default_patterns(self) -> Dict[str, Any]:
        """Get default DDD patterns when Context7 is unavailable."""
        return {
            'ddd_best_practices': {
                'analyze_phase': [
                    "Understand existing code structure and patterns",
                    "Identify current behavior through code reading",
                    "Document dependencies and side effects",
                    "Map test coverage gaps"
                ],
                'preserve_phase': [
                    "Write characterization tests for existing behavior",
                    "Capture current behavior as the golden standard",
                    "Ensure tests pass with current implementation",
                    "Create behavior snapshots for complex outputs"
                ],
                'improve_phase': [
                    "Refactor code while keeping tests green",
                    "Make small, incremental changes",
                    "Run tests after each change",
                    "Maintain behavior preservation"
                ]
            },
            'python_testing': {
                'pytest_features': [
                    "Parametrized tests for multiple scenarios",
                    "Fixtures for test setup and teardown",
                    "Markers for categorizing tests",
                    "Plugins for enhanced functionality"
                ],
                'assertions': [
                    "Use pytest's assert statements",
                    "Provide clear error messages",
                    "Test expected exceptions with pytest.raises",
                    "Use pytest.approx for floating point comparisons"
                ]
            },
            'assertions': {
                'best_practices': [
                    "One assertion per test when possible",
                    "Clear and descriptive assertion messages",
                    "Test both positive and negative cases",
                    "Use appropriate assertion methods"
                ]
            },
            'mocking': {
                'strategies': [
                    "Mock external dependencies",
                    "Use dependency injection for testability",
                    "Create test doubles for complex objects",
                    "Verify interactions with mocks"
                ]
            }
        }
```

## Testing Patterns

### Given-When-Then Pattern

```python
def test_user_authentication_valid_credentials():
    """
    Test user authentication with valid credentials.

    Given: A registered user with valid credentials
    When: The user attempts to authenticate
    Then: The system should return a valid authentication token
    """
    # Given
    user = User(email="test@example.com", password="secure_password")
    auth_service = AuthenticationService()

    # When
    result = auth_service.authenticate(user.email, user.password)

    # Then
    assert result is not None
    assert result.token is not None
    assert result.expires_at > datetime.now()
```

### Arrange-Act-Assert Pattern

```python
def test_calculate_total_price_with_discount():
    """
    Test total price calculation with discount applied.
    """
    # Arrange
    cart = ShoppingCart()
    cart.add_item("item1", price=100.0, quantity=2)
    cart.add_item("item2", price=50.0, quantity=1)
    discount_code = "SAVE10"

    # Act
    total = cart.calculate_total(discount_code)

    # Assert
    assert total == 225.0  # (200 + 50) * 0.9
```

### Parameterized Testing Pattern

```python
@pytest.mark.parametrize("input,expected", [
    (2, 4),    # 2^2 = 4
    (3, 9),    # 3^2 = 9
    (0, 0),    # 0^2 = 0
    (-1, 1),   # (-1)^2 = 1
    (10, 100)  # 10^2 = 100
])
def test_square_function(input, expected):
    """Test square function with various inputs."""
    result = square(input)
    assert result == expected
```

### Exception Testing Pattern

```python
def test_divide_by_zero_raises_exception():
    """Test that division by zero raises ZeroDivisionError."""
    with pytest.raises(ZeroDivisionError) as exc_info:
        divide(10, 0)

    assert "division by zero" in str(exc_info.value)
```

### Mock Testing Pattern

```python
def test_external_api_call_with_mock():
    """Test external API call with mocked response."""
    # Create mock
    mock_api = Mock()
    mock_api.get_data.return_value = {"status": "success", "data": [1, 2, 3]}

    # Use mock in test
    service = DataService(api_client=mock_api)
    result = service.fetch_data()

    # Verify interaction
    mock_api.get_data.assert_called_once()
    assert result == [1, 2, 3]
```

## Pytest Fixtures

### Basic Fixture

```python
@pytest.fixture
def sample_user():
    """Create a sample user for testing."""
    return User(
        email="test@example.com",
        username="testuser",
        password="secure_password"
    )

def test_user_email(sample_user):
    """Test user email attribute."""
    assert sample_user.email == "test@example.com"
```

### Fixture with Setup and Teardown

```python
@pytest.fixture
def database_connection():
    """Create database connection with cleanup."""
    # Setup
    conn = Database.connect(":memory:")
    conn.create_tables()

    yield conn  # Provide connection to test

    # Teardown
    conn.close()

def test_database_query(database_connection):
    """Test database query with fixture."""
    result = database_connection.query("SELECT * FROM users")
    assert len(result) >= 0
```

### Parametrized Fixture

```python
@pytest.fixture(params=[
    ("valid_email@example.com", True),
    ("invalid_email", False),
    ("", False)
])
def email_validation_data(request):
    """Provide email validation test data."""
    return request.param

def test_email_validation(email_validation_data):
    """Test email validation with various inputs."""
    email, expected_valid = email_validation_data
    result = validate_email(email)
    assert result.is_valid == expected_valid
```

## Test Organization

### Test Discovery Structure

```
tests/
├── unit/              # Unit tests for individual components
│   ├── test_models.py
│   ├── test_services.py
│   └── test_utils.py
├── integration/       # Integration tests for component interaction
│   ├── test_api_integration.py
│   └── test_database_integration.py
├── acceptance/        # Acceptance tests for user scenarios
│   ├── test_user_scenarios.py
│   └── test_business_workflows.py
└── conftest.py        # Shared fixtures and configuration
```

### Test Markers

```python
import pytest

# Define custom markers
pytest.mark.slow = pytest.mark.slow
pytest.mark.integration = pytest.mark.integration
pytest.mark.unit = pytest.mark.unit

# Use markers in tests
@pytest.mark.unit
def test_individual_function():
    """Unit test for individual function."""
    assert calculate(2, 2) == 4

@pytest.mark.integration
def test_database_integration():
    """Integration test for database."""
    result = db.query("SELECT * FROM users")
    assert result is not None

@pytest.mark.slow
def test_performance_benchmark():
    """Slow performance test."""
    result = expensive_operation()
    assert result is not None
```

## Test Coverage

### Running Coverage Analysis

```bash
# Run tests with coverage
pytest --cov=src --cov-report=html --cov-report=term

# Generate coverage report
pytest --cov=src --cov-report=html
```

### Coverage Configuration

```ini
# .coveragerc
[run]
source = src
omit =
    */tests/*
    */__init__.py

[report]
exclude_lines =
    pragma: no cover
    def __repr__
    raise AssertionError
    raise NotImplementedError
    if __name__ == .__main__.:
```

## Best Practices

1. Test Isolation: Each test should be independent and not rely on other tests
2. Descriptive Names: Test names should clearly describe what is being tested
3. One Assertion Per Test: Keep tests focused on a single behavior
4. Arrange-Act-Assert: Structure tests clearly with this pattern
5. Mock External Dependencies: Use mocks for external services and databases
6. Test Edge Cases: Include tests for boundary conditions and error cases
7. Fast Tests: Keep unit tests fast for quick feedback
8. Maintainable Tests: Keep tests simple and easy to understand
9. Context7 Integration: Leverage Context7 for latest testing patterns and best practices
10. Continuous Testing: Run tests automatically with every code change

---

Related: [ANALYZE-PRESERVE-IMPROVE](./analyze-preserve-improve.md) | [Test Generation](./test-generation.md)
