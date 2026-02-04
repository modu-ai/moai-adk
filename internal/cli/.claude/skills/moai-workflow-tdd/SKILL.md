---
name: moai-workflow-tdd
description: >
  Test-Driven Development workflow specialist using RED-GREEN-REFACTOR
  cycle for test-first software development.
  Use when developing new features from scratch, creating isolated modules,
  or when behavior specification drives implementation.
  Do NOT use for refactoring existing code (use moai-workflow-ddd instead)
  or when behavior preservation is the primary goal.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Write Edit Bash(pytest:*) Bash(ruff:*) Bash(npm:*) Bash(npx:*) Bash(node:*) Bash(jest:*) Bash(vitest:*) Bash(go:*) Bash(cargo:*) Bash(mix:*) Bash(uv:*) Bash(bundle:*) Bash(php:*) Bash(phpunit:*) Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-03"
  modularized: "true"
  tags: "workflow, tdd, test-driven, red-green-refactor, test-first"
  author: "MoAI-ADK Team"
  context: "fork"
  agent: "manager-tdd"
  related-skills: "moai-workflow-ddd, moai-workflow-testing, moai-foundation-quality"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["TDD", "test-driven development", "red-green-refactor", "test-first", "new feature", "greenfield"]
  phases: ["run"]
  agents: ["manager-tdd", "expert-backend", "expert-frontend", "expert-testing"]
---

# Test-Driven Development (TDD) Workflow

## Development Mode Configuration (CRITICAL)

[NOTE] This workflow is selected based on `.moai/config/sections/quality.yaml`:

```yaml
constitution:
  development_mode: hybrid    # or ddd, tdd
  hybrid_settings:
    new_features: tdd        # New code → use TDD (this workflow)
    legacy_refactoring: ddd  # Existing code → use DDD
```

**When to use this workflow**:
- `development_mode: tdd` → Always use TDD
- `development_mode: hybrid` + new package/module → Use TDD
- `development_mode: hybrid` + refactoring existing code → Use DDD instead (moai-workflow-ddd)

**Key distinction**:
- **New file/package** (doesn't exist yet) → TDD (this workflow)
- **Existing code** (file already exists) → DDD (ANALYZE-PRESERVE-IMPROVE)

## Quick Reference

Test-Driven Development provides a disciplined approach for creating new functionality where tests define the expected behavior before implementation.

Core Cycle - RED-GREEN-REFACTOR:

- RED: Write a failing test that defines desired behavior
- GREEN: Write minimal code to make the test pass
- REFACTOR: Improve code structure while keeping tests green

When to Use TDD:

- Creating new functionality from scratch
- Building isolated modules with no existing dependencies
- When behavior specification drives development
- New API endpoints with clear contracts
- New UI components with defined behavior
- Greenfield projects (rare - usually Hybrid is better)

When NOT to Use TDD:

- Refactoring existing code (use DDD instead)
- When behavior preservation is the primary goal
- Legacy codebase without test coverage (use DDD first)
- When modifying existing files (consider Hybrid mode)

---

## Core Philosophy

### TDD vs DDD Comparison

TDD Approach:

- Cycle: RED-GREEN-REFACTOR
- Goal: Create new functionality through tests
- Starting Point: No code exists
- Test Type: Specification tests that define expected behavior
- Outcome: New working code with test coverage

DDD Approach:

- Cycle: ANALYZE-PRESERVE-IMPROVE
- Goal: Improve structure without behavior change
- Starting Point: Existing code with defined behavior
- Test Type: Characterization tests that capture current behavior
- Outcome: Better structured code with identical behavior

### Test-First Principle

The golden rule of TDD is that tests must be written before implementation code:

- Tests define the contract
- Tests document expected behavior
- Tests catch regressions immediately
- Implementation is driven by test requirements

---

## Implementation Guide

### Phase 1: RED - Write a Failing Test

The RED phase focuses on defining the desired behavior through a failing test.

#### Writing Effective Tests

Before writing any implementation code:

- Understand the requirement clearly
- Define the expected behavior in test form
- Write one test at a time
- Keep tests focused and specific
- Use descriptive test names that document behavior

#### Test Structure

Follow the Arrange-Act-Assert pattern:

- Arrange: Set up test data and dependencies
- Act: Execute the code under test
- Assert: Verify the expected outcome

#### Verification

The test must fail initially:

- Confirms the test actually tests something
- Ensures the test is not passing by accident
- Documents the gap between current and desired state

### Phase 2: GREEN - Make the Test Pass

The GREEN phase focuses on writing minimal code to satisfy the test.

#### Minimal Implementation

Write only enough code to make the test pass:

- Do not over-engineer
- Do not add features not required by tests
- Focus on correctness, not perfection
- Hardcode values if necessary (refactor later)

#### Verification

Run the test to confirm it passes:

- All assertions must succeed
- No other tests should break
- Implementation satisfies the test requirements

### Phase 3: REFACTOR - Improve the Code

The REFACTOR phase focuses on improving code quality while maintaining behavior.

#### Safe Refactoring

With passing tests as a safety net:

- Remove duplication
- Improve naming and readability
- Extract methods and classes
- Apply design patterns where appropriate

#### Continuous Verification

After each refactoring step:

- Run all tests
- If any test fails, revert immediately
- Commit when tests pass

---

## TDD Workflow Execution

### Standard TDD Session

When executing TDD through manager-tdd:

Step 1 - Understand Requirements:

- Read SPEC document for feature scope
- Identify test cases from acceptance criteria
- Plan test implementation order

Step 2 - RED Phase:

- Write first failing test
- Verify test fails for the right reason
- Document expected behavior

Step 3 - GREEN Phase:

- Write minimal implementation
- Run test to verify it passes
- Move to next test

Step 4 - REFACTOR Phase:

- Review code for improvements
- Apply refactoring with tests as safety net
- Commit clean code

Step 5 - Repeat:

- Continue RED-GREEN-REFACTOR cycle
- Until all requirements are implemented
- Until all acceptance criteria pass

### TDD Loop Pattern

For features requiring multiple test cases:

- Identify all test cases upfront
- Prioritize by dependency and complexity
- Execute RED-GREEN-REFACTOR for each
- Maintain cumulative test suite

---

## Quality Metrics

### TDD Success Criteria

Test Coverage (Required):

- Minimum 80% coverage per commit
- 90% recommended for new code
- All public interfaces tested

Code Quality (Goals):

- All tests pass
- No test written after implementation
- Clear test names documenting behavior
- Minimal implementation satisfying tests

### TDD-Specific TRUST Validation

Apply TRUST 5 framework with TDD focus:

- Testability: Test-first approach ensures testability
- Readability: Tests document expected behavior
- Understandability: Tests serve as living documentation
- Security: Security tests written before implementation
- Transparency: Test failures provide immediate feedback

---

## Integration Points

### With DDD Workflow

TDD and DDD are complementary:

- TDD for new code
- DDD for existing code refactoring
- Hybrid mode combines both approaches

### With Testing Workflow

TDD integrates with testing workflow:

- Uses specification tests
- Integrates with coverage tools
- Supports mutation testing for test quality

### With Quality Framework

TDD outputs feed into quality assessment:

- Coverage metrics tracked
- TRUST 5 validation for changes
- Quality gates enforce standards

---

## Troubleshooting

### Common Issues

Test is Too Complex:

- Break into smaller, focused tests
- Test one behavior at a time
- Use test fixtures for complex setup

Implementation Grows Too Fast:

- Resist urge to implement untested features
- Return to RED phase for new functionality
- Keep GREEN phase minimal

Refactoring Breaks Tests:

- Revert immediately
- Refactor in smaller steps
- Ensure tests verify behavior, not implementation

### Recovery Procedures

When TDD discipline breaks down:

- Stop and assess current state
- Write characterization tests for existing code
- Resume TDD for remaining features
- Consider switching to Hybrid mode

---

Version: 1.0.0
Status: Active
Last Updated: 2026-02-03
