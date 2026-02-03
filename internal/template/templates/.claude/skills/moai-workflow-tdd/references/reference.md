# TDD Reference

## RED-GREEN-REFACTOR Cycle

### RED Phase

Write a failing test:

- Test describes desired behavior
- Test fails because code does not exist
- Test name documents the requirement

### GREEN Phase

Make the test pass:

- Write minimal code
- Only satisfy current test
- Do not add extra features

### REFACTOR Phase

Improve the code:

- Remove duplication
- Improve naming
- Extract methods
- Keep tests passing

## Test Quality Guidelines

### Good Test Characteristics

- Fast: Tests run quickly
- Independent: No test depends on another
- Repeatable: Same result every time
- Self-validating: Pass or fail clearly
- Timely: Written before implementation

### Naming Conventions

Test names should document behavior:

- test_should_return_empty_list_when_no_items
- test_creates_user_with_valid_email
- test_throws_error_for_invalid_input

## Coverage Guidelines

| Code Type | Minimum Coverage |
|-----------|-----------------|
| New feature code | 90% |
| Per commit | 80% |
| Critical paths | 100% |

## Integration with Hybrid Mode

In Hybrid mode, TDD applies to:

- New files
- New functions in existing files
- New classes and modules

DDD applies to:

- Modifications to existing code
- Refactoring existing functions
- Behavior-preserving changes
