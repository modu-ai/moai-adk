# Domain-Driven Development with Context7 Integration

> Module: ANALYZE-PRESERVE-IMPROVE DDD cycle with Context7 patterns and AI-powered testing
> Complexity: Advanced
> Time: 25+ minutes
> Dependencies: test runner for your language, Context7 MCP (optional)

## Overview

DDD Context7 integration provides a comprehensive domain-driven development workflow with AI-powered test generation, Context7-enhanced testing patterns, and automated best practices enforcement.

### Key Features

- AI-Powered Test Generation: Generate comprehensive test suites from specifications
- Context7 Integration: Access latest testing patterns and best practices
- ANALYZE-PRESERVE-IMPROVE Cycle: Complete DDD workflow implementation
- Advanced Testing: Property-based testing, mutation testing, continuous testing
- Test Patterns: Comprehensive library of testing patterns and fixtures

## Quick Start

### Basic DDD Cycle

This module describes a workflow, not an importable library. Drive the
ANALYZE-PRESERVE-IMPROVE cycle yourself using your project's own test runner.
A characterization-test specification is a plain description of the behavior
to preserve — express it as a test in your language's framework.

```
# Conceptual characterization-test specification (language-neutral)
spec = {
    name: "test_user_login_behavior_preservation",
    description: "Preserve existing login behavior during refactoring",
    test_type: CHARACTERIZATION,            # capture current behavior, not aspirational
    requirements: [
        "Existing login flow must continue to work",
        "Error messages should remain consistent"
    ],
    acceptance_criteria: [
        "Valid credentials return user token (existing behavior)",
        "Invalid credentials raise same error messages"
    ],
    edge_cases: [
        "Empty email (existing behavior)",
        "Empty password (existing behavior)"
    ]
}

# Drive the cycle with your own toolchain:
#   ANALYZE   — read authenticate_user, map current behavior + dependencies
#   PRESERVE  — write characterization tests from `spec` (must PASS on current code)
#   IMPROVE   — refactor authenticate_user incrementally, keep tests green
#   REVIEW    — verify all characterization tests still pass, commit
```

If Context7 MCP is available, load up-to-date characterization-testing patterns
for your language/framework (e.g. `/pytest` , `/jest`, `/gotest`) to refine the
specification before writing the test.

## Core Components

### DDD Cycle Phases

1. ANALYZE Phase: Understand existing code
   - Analyze existing code structure and patterns
   - Identify current behavior through code reading
   - Document dependencies and side effects
   - Map test coverage gaps

2. PRESERVE Phase: Create characterization tests
   - Write characterization tests for existing behavior
   - Capture current behavior as the "golden standard"
   - Ensure tests pass with current implementation
   - Create behavior snapshots for complex outputs

3. IMPROVE Phase: Refactor with behavior preservation
   - Refactor code while keeping tests green
   - Make small, incremental changes
   - Run tests after each change
   - Maintain behavior preservation

4. REVIEW Phase: Verify and commit
   - Verify all characterization tests still pass
   - Review code quality and documentation
   - Check for any behavior changes
   - Commit changes with clear messages

### Context7 Integration

The DDD Context7 integration provides:

- Pattern Loading: Access latest testing patterns from Context7
- AI Test Generation: Enhanced test generation with Context7 patterns
- Best Practices: Industry-standard testing practices
- Edge Case Detection: Automatic edge case identification
- Test Suggestions: AI-powered test improvement suggestions

## Module Structure

### Core Modules

**ANALYZE-PRESERVE-IMPROVE Implementation** (`ddd-context7/analyze-preserve-improve.md`)
- DDD cycle implementation
- Test execution and validation
- Coverage analysis
- Session management

**Test Generation** (`ddd-context7/test-generation.md`)
- AI-powered test generation
- Specification-based generation
- Context7-enhanced generation
- Template-based generation

**Test Patterns** (`ddd-context7/test-patterns.md`)
- Testing patterns and best practices
- Pytest fixtures and organization
- Test discovery structure
- Coverage configuration

**Advanced Features** (`ddd-context7/advanced-features.md`)
- Comprehensive test suite generation
- Property-based testing
- Mutation testing
- Continuous testing

## Common Use Cases

Each use case below is a characterization-test specification expressed as a
language-neutral description. Translate it into a concrete test in your
project's framework (pytest, Jest, go test, cargo test, JUnit, etc.).

### Behavior Preservation

```
# Characterization test specification
char_spec = {
    name: "test_calculate_sum_existing_behavior",
    description: "Preserve existing sum calculation behavior",
    test_type: CHARACTERIZATION,
    requirements: ["Function should sum two numbers (existing behavior)"],
    acceptance_criteria: ["Returns correct sum as currently implemented"],
    edge_cases: ["Zero values", "Negative numbers", "Large numbers"]
}
# Author the test from char_spec; it MUST pass against the current implementation.
```

### Refactoring with Tests

```
# Integration test specification for refactoring
refactor_spec = {
    name: "test_database_integration_refactor",
    description: "Preserve database behavior during refactoring",
    test_type: INTEGRATION,
    requirements: ["Database connection", "Query execution"],
    acceptance_criteria: ["Connection succeeds as before", "Query returns same data"],
    edge_cases: ["Connection failure handling", "Empty results", "Large datasets"]
}
```

### Exception Behavior Preservation

```
# Exception test specification
exception_spec = {
    name: "test_divide_by_zero_existing_behavior",
    description: "Preserve division by zero exception handling",
    test_type: CHARACTERIZATION,
    requirements: ["Division function", "Error handling"],
    acceptance_criteria: ["Raises the same error/exception type as before"],
    edge_cases: ["Divisor is zero", "Dividend is zero"]
}
```

## Best Practices

### Test Design

1. Characterization First: Write tests that capture existing behavior before changing code
2. Descriptive Names: Test names should clearly describe what behavior is being preserved
3. Arrange-Act-Assert: Structure tests with this pattern for clarity
4. Independent Tests: Tests should not depend on each other
5. Fast Execution: Keep tests fast for quick feedback

### Context7 Integration

1. Pattern Loading: Load Context7 patterns for latest best practices
2. Edge Case Detection: Use Context7 to identify missing edge cases
3. Test Suggestions: Leverage AI suggestions for test improvements
4. Quality Analysis: Use Context7 for test quality analysis

### DDD Workflow

1. Analyze First: Always understand existing behavior before changing code
2. Preserve with Tests: Create characterization tests before refactoring
3. Keep Tests Green: Never commit failing tests
4. Small Increments: Make small, incremental changes
5. Continuous Testing: Run tests after every change

## Advanced Features

### Property-Based Testing

Use Hypothesis for property-based testing to verify code properties across many random inputs.

### Mutation Testing

Use mutation testing to verify test suite quality by introducing code mutations and checking if tests catch them.

### Continuous Testing

Implement watch mode for automatic test execution on file changes.

### AI-Powered Generation

Leverage Context7 for intelligent test generation and suggestions.

## Performance Considerations

- Test Execution: Use parallel test execution for faster feedback
- Test Isolation: Ensure tests are isolated to prevent interference
- Mock External Dependencies: Mock external services for fast, reliable tests
- Optimize Setup: Use fixtures and test factories for efficient test setup

## Troubleshooting

### Common Issues

1. Tests Failing Intermittently
   - Check for shared state between tests
   - Verify test isolation
   - Add proper cleanup in fixtures

2. Slow Test Execution
   - Use parallel test execution
   - Mock external dependencies
   - Optimize test setup

3. Context7 Integration Issues
   - Verify Context7 client configuration
   - Check network connectivity
   - Use default patterns as fallback

## Resources

### Detailed Modules

- [ANALYZE-PRESERVE-IMPROVE Implementation](./ddd-context7/analyze-preserve-improve.md) - Core DDD cycle
- [Test Generation](./ddd-context7/test-generation.md) - AI-powered generation
- [Test Patterns](./ddd-context7/test-patterns.md) - Patterns and best practices
- [Advanced Features](./ddd-context7/advanced-features.md) - Advanced testing techniques

### Related Modules

- [AI Debugging](./ai-debugging.md) - Debugging techniques
- [Performance Optimization](./performance-optimization.md) - Performance testing
- [Smart Refactoring](./smart-refactoring.md) - Refactoring with tests

### External Resources

- [Pytest Documentation](https://docs.pytest.org/)
- [Python Testing Best Practices](https://docs.python-guide.org/writing/tests/)
- [Hypothesis Property-Based Testing](https://hypothesis.works/)
- [Context7 MCP Documentation](https://context7.io/docs)

---

Module: `modules/ddd-context7.md`
Version: 2.0.0 (DDD Migration)
Last Updated: 2026-01-17
