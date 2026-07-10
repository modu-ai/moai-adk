# Context7 Test Patterns and Best Practices

> Module: Testing patterns, Context7 integration, and industry best practices
> Complexity: Advanced
> Time: 15+ minutes
> Dependencies: the project's test runner, Context7 MCP

## Context7 DDD Integration

```text
class Context7DDDIntegration:
    context7
    pattern_cache = {}

    load_ddd_patterns(language = "python"):
        cache_key = "ddd_patterns_" + language
        if cache_key in pattern_cache: return pattern_cache[cache_key]

        patterns = {}
        if context7 available:
            try:
                patterns.ddd_best_practices = context7.get_library_docs(
                    "<testing-lib>", topic="DDD ANALYZE-PRESERVE-IMPROVE patterns best practices",
                    tokens=4000)
                patterns.testing = context7.get_library_docs(
                    "<" + language + "-testing>",
                    topic="advanced testing patterns mocking fixtures",
                    tokens=3000)
                patterns.assertions = context7.get_library_docs(
                    "<assertions>", topic="assertion patterns error messages test design",
                    tokens=2000)
                patterns.mocking = context7.get_library_docs(
                    "<" + language + "-mock>",
                    topic="mocking strategies test doubles isolation patterns",
                    tokens=3000)
            except e:
                log("Failed to load Context7 patterns: " + e)
                patterns = default_patterns()
        else:
            patterns = default_patterns()

        pattern_cache[cache_key] = patterns
        return patterns

    default_patterns():
        return {
          ddd_best_practices: {
            analyze_phase: [
                "Understand existing code structure and patterns",
                "Identify current behavior through code reading",
                "Document dependencies and side effects",
                "Map test coverage gaps"],
            preserve_phase: [
                "Write characterization tests for existing behavior",
                "Capture current behavior as the golden standard",
                "Ensure tests pass with current implementation",
                "Create behavior snapshots for complex outputs"],
            improve_phase: [
                "Refactor code while keeping tests green",
                "Make small, incremental changes",
                "Run tests after each change",
                "Maintain behavior preservation"]
          },
          testing: {
            features: [
                "Parametrized / table-driven tests for multiple scenarios",
                "Fixtures / setup-teardown for test state",
                "Tags / markers for categorizing tests",
                "Plugins for enhanced functionality"],
            assertions: [
                "Use the runner's rich assertion helpers",
                "Provide clear failure messages",
                "Test expected exceptions explicitly",
                "Use approximate comparison for floating point"]
          },
          assertions: {
            best_practices: [
                "One assertion per test when possible",
                "Clear and descriptive assertion messages",
                "Test both positive and negative cases",
                "Use appropriate assertion methods"]
          },
          mocking: {
            strategies: [
                "Mock external dependencies",
                "Use dependency injection for testability",
                "Create test doubles for complex objects",
                "Verify interactions with mocks"]
          }
        }
```

## Testing Patterns

The patterns below are universal; each language's test runner has its own syntax for the same ideas (Go table-driven tests, pytest parametrize, JS it.each, Rust rstest, etc.).

### Given-When-Then Pattern

```text
test_user_authentication_valid_credentials():
    # Given: a registered user with valid credentials
    user = User(email="test@example.com", password="secure_password")
    auth_service = AuthenticationService()

    # When: the user attempts to authenticate
    result = auth_service.authenticate(user.email, user.password)

    # Then: the system returns a valid authentication token
    assert result is not none
    assert result.token is not none
    assert result.expires_at > now()
```

### Arrange-Act-Assert Pattern

```text
test_calculate_total_price_with_discount():
    # Arrange
    cart = ShoppingCart()
    cart.add_item("item1", price=100.0, quantity=2)
    cart.add_item("item2", price=50.0, quantity=1)
    discount_code = "SAVE10"

    # Act
    total = cart.calculate_total(discount_code)

    # Assert
    assert total == 225.0   # (200 + 50) * 0.9
```

### Parameterized / Table-Driven Testing Pattern

```text
# A table of (input, expected) rows — most runners support this idiom
# (pytest parametrize, Go table-driven tests, JS it.each, Rust rstest).
cases = [
    (2,  4),    # 2^2
    (3,  9),    # 3^2
    (0,  0),    # 0^2
    (-1, 1),    # (-1)^2
    (10, 100)   # 10^2
]
for (input, expected) in cases:
    test "square(" + input + ") == " + expected:
        assert square(input) == expected
```

### Exception Testing Pattern

```text
test_divide_by_zero_raises_error():
    # Assert that dividing by zero raises an error, and check its message.
    # Idiom per runner: pytest.raises, Go require.ErrorIs, JS expect.toThrow,
    # Rust #[should_panic], JUnit assertThrows.
    assert raises(DivideByZero, divide(10, 0)):
        message contains "division by zero"
```

### Mock Testing Pattern

```text
test_external_api_call_with_mock():
    # Create a test double for the external API
    mock_api = Mock()
    mock_api.get_data.returns({ status: "success", data: [1, 2, 3] })

    service = DataService(api_client=mock_api)
    result = service.fetch_data()

    # Verify the interaction and the result
    assert mock_api.get_data was called once
    assert result == [1, 2, 3]
```

## Test Fixtures

### Basic Fixture

```text
# A factory/fixture that builds a sample object for reuse across tests.
# Idiom: pytest @fixture, Go t.Helper sub-tests, JS beforeEach, Rust helper fn.
fixture sample_user():
    return User(email="test@example.com", username="testuser", password="secure_password")

test_user_email(sample_user):
    assert sample_user.email == "test@example.com"
```

### Fixture with Setup and Teardown

```text
fixture database_connection():
    # Setup
    conn = Database.connect(":memory:")
    conn.create_tables()
    yield conn        # provide connection to the test body
    # Teardown (runs after the test)
    conn.close()

test_database_query(database_connection):
    result = database_connection.query("SELECT * FROM users")
    assert len(result) >= 0
```

### Parametrized Fixture

```text
fixture email_validation_data(params=[
    ("valid_email@example.com", true),
    ("invalid_email",           false),
    ("",                        false)
]):
    return current_param   # each row is injected as a separate test case

for (email, expected_valid) in email_validation_data:
    test "validate_email(" + email + ")":
        assert validate_email(email).is_valid == expected_valid
```

## Test Organization

### Test Discovery Structure

```text
tests/
  unit/             # unit tests for individual components
  integration/      # integration tests for component interaction
  acceptance/       # acceptance tests for user scenarios
  shared.<ext>      # shared fixtures and configuration (conftest, helper, etc.)
```

### Test Markers / Tags

```text
# Tag tests for selective runs. Idiom per runner: pytest markers,
# Go build tags, JS describe/it labels, JUnit @Tag, Rust test attributes.

@unit
test individual_function():
    assert calculate(2, 2) == 4

@integration
test database_integration():
    result = db.query("SELECT * FROM users")
    assert result is not none

@slow
test performance_benchmark():
    result = expensive_operation()
    assert result is not none
```

## Test Coverage

### Running Coverage Analysis

```text
# Run the project's test runner with coverage. Examples per language:
#   Go:       go test -coverprofile=cover.out ./... ; go tool cover -html=cover.out
#   Python:   pytest --cov=src --cov-report=html --cov-report=term
#   Node:     npm test -- --coverage
#   Rust:     cargo tarpaulin --out Html
#   .NET:     dotnet test --collect:"XPlat Code Coverage"
run_with_coverage()
generate_html_report()
```

### Coverage Configuration

```text
# Coverage tool config (shape is tool-specific; the intent is language-neutral):
#   source = src
#   exclude unit-test directories and generated/empty entry points
#   exclude_lines = no-cover pragmas, debug __repr__-style helpers,
#                   unreachable assertion/not-implemented raises, the __main__ guard
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
